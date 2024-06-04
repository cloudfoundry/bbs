package main_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"google.golang.org/grpc/grpclog"

	"code.cloudfoundry.org/bbs"
	bbsconfig "code.cloudfoundry.org/bbs/cmd/bbs/config"
	"code.cloudfoundry.org/bbs/encryption"
	"code.cloudfoundry.org/bbs/test_helpers"
	"code.cloudfoundry.org/bbs/test_helpers/sqlrunner"
	diego_logging_client "code.cloudfoundry.org/diego-logging-client"
	"code.cloudfoundry.org/diego-logging-client/testhelpers"
	"code.cloudfoundry.org/durationjson"
	"code.cloudfoundry.org/go-loggregator/v9/rpc/loggregator_v2"
	"code.cloudfoundry.org/inigo/helpers/portauthority"
	"code.cloudfoundry.org/lager/v3"
	"code.cloudfoundry.org/lager/v3/lagerflags"
	"code.cloudfoundry.org/lager/v3/lagertest"
	"code.cloudfoundry.org/locket"
	locketconfig "code.cloudfoundry.org/locket/cmd/locket/config"
	locketrunner "code.cloudfoundry.org/locket/cmd/locket/testrunner"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
	"github.com/tedsuo/ifrit"
	ginkgomon "github.com/tedsuo/ifrit/ginkgomon_v2"

	"testing"
	"time"
)

var (
	logger        lager.Logger
	ctx           context.Context
	portAllocator portauthority.PortAllocator

	client            bbs.InternalClient
	bbsBinPath        string
	bbsAddress        string
	bbsHealthAddress  string
	bbsPort           uint16
	bbsURL            *url.URL
	bbsConfig         bbsconfig.BBSConfig
	bbsRunner         *ginkgomon.Runner
	bbsProcess        ifrit.Process
	locketHelper      *test_helpers.LocketHelper
	auctioneerServer  *ghttp.Server
	testMetricsChan   chan *loggregator_v2.Envelope
	locketBinPath     string
	locketProcess     ifrit.Process
	testIngressServer *testhelpers.TestIngressServer

	signalMetricsChan chan struct{}
	sqlProcess        ifrit.Process
	sqlRunner         sqlrunner.SQLRunner
)

func TestBBS(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BBS Cmd Suite")
}

var _ = SynchronizedBeforeSuite(
	func() []byte {
		bbsPath, err := gexec.Build("code.cloudfoundry.org/bbs/cmd/bbs", "-race")
		Expect(err).NotTo(HaveOccurred())

		locketPath, err := gexec.Build("code.cloudfoundry.org/locket/cmd/locket", "-race")
		Expect(err).NotTo(HaveOccurred())

		return []byte(strings.Join([]string{bbsPath, locketPath}, ","))
	},
	func(binPaths []byte) {
		grpclog.SetLoggerV2(grpclog.NewLoggerV2(io.Discard, io.Discard, io.Discard))
		startPort := 1050 * GinkgoParallelProcess()
		portRange := 1000
		var err error
		portAllocator, err = portauthority.New(startPort, startPort+portRange)
		Expect(err).NotTo(HaveOccurred())

		path := string(binPaths)
		bbsBinPath = strings.Split(path, ",")[0]
		locketBinPath = strings.Split(path, ",")[1]

		SetDefaultEventuallyTimeout(15 * time.Second)

		dbName := fmt.Sprintf("diego_%d", GinkgoParallelProcess())
		sqlRunner = test_helpers.NewSQLRunner(dbName)
		sqlProcess = ginkgomon.Invoke(sqlRunner)
	},
)

var _ = SynchronizedAfterSuite(func() {
	ginkgomon.Kill(sqlProcess)
}, func() {
	gexec.CleanupBuildArtifacts()
})

var _ = BeforeEach(func() {
	var err error
	logger = lagertest.NewTestLogger("test")
	ctx = context.Background()
	fixturesPath := "fixtures"

	locketPort, err := portAllocator.ClaimPorts(1)
	Expect(err).NotTo(HaveOccurred())

	locketAddress := fmt.Sprintf("localhost:%d", locketPort)

	locketRunner := locketrunner.NewLocketRunner(locketBinPath, func(cfg *locketconfig.LocketConfig) {
		cfg.DatabaseConnectionString = sqlRunner.ConnectionString()
		cfg.DatabaseDriver = sqlRunner.DriverName()
		cfg.ListenAddress = locketAddress
	})
	locketProcess = ginkgomon.Invoke(locketRunner)

	metronCAFile := path.Join(fixturesPath, "metron", "CA.crt")
	metronClientCertFile := path.Join(fixturesPath, "metron", "client.crt")
	metronClientKeyFile := path.Join(fixturesPath, "metron", "client.key")
	metronServerCertFile := path.Join(fixturesPath, "metron", "metron.crt")
	metronServerKeyFile := path.Join(fixturesPath, "metron", "metron.key")

	auctioneerServer = ghttp.NewServer()
	auctioneerServer.UnhandledRequestStatusCode = http.StatusAccepted
	auctioneerServer.AllowUnhandledRequests = true

	bbsPort, err = portAllocator.ClaimPorts(1)
	Expect(err).NotTo(HaveOccurred())
	bbsAddress = fmt.Sprintf("127.0.0.1:%d", bbsPort)

	bbsHealthPort, err := portAllocator.ClaimPorts(1)
	Expect(err).NotTo(HaveOccurred())
	bbsHealthAddress = fmt.Sprintf("127.0.0.1:%d", bbsHealthPort)

	bbsURL = &url.URL{
		Scheme: "https",
		Host:   bbsAddress,
	}

	testIngressServer, err = testhelpers.NewTestIngressServer(metronServerCertFile, metronServerKeyFile, metronCAFile)
	Expect(err).NotTo(HaveOccurred())

	testIngressServer.Start()

	metricsPort, err := strconv.Atoi(strings.TrimPrefix(testIngressServer.Addr(), "127.0.0.1:"))
	Expect(err).NotTo(HaveOccurred())

	receiversChan := testIngressServer.Receivers()

	testMetricsChan, signalMetricsChan = testhelpers.TestMetricChan(receiversChan)

	serverCaFile := path.Join(fixturesPath, "green-certs", "server-ca.crt")
	clientCertFile := path.Join(fixturesPath, "green-certs", "client.crt")
	clientKeyFile := path.Join(fixturesPath, "green-certs", "client.key")
	client, err = bbs.NewClient(bbsURL.String(), serverCaFile, clientCertFile, clientKeyFile, 0, 0)
	Expect(err).ToNot(HaveOccurred())

	bbsConfig = bbsconfig.BBSConfig{
		SessionName:                 "bbs",
		CommunicationTimeout:        durationjson.Duration(10 * time.Second),
		RequireSSL:                  true,
		DesiredLRPCreationTimeout:   durationjson.Duration(1 * time.Minute),
		ExpireCompletedTaskDuration: durationjson.Duration(2 * time.Minute),
		ExpirePendingTaskDuration:   durationjson.Duration(30 * time.Minute),
		KickTaskDuration:            durationjson.Duration(30 * time.Second),
		LockTTL:                     durationjson.Duration(1 * time.Second),
		LockRetryInterval:           durationjson.Duration(locket.RetryInterval),
		ConvergenceWorkers:          20,
		UpdateWorkers:               1000,
		TaskCallbackWorkers:         1000,
		MaxOpenDatabaseConnections:  200,
		MaxIdleDatabaseConnections:  200,
		AuctioneerRequireTLS:        false,
		RepClientSessionCacheSize:   0,
		RepRequireTLS:               false,

		ListenAddress:     bbsAddress,
		AdvertiseURL:      bbsURL.String(),
		AuctioneerAddress: auctioneerServer.URL(),

		DatabaseDriver:           sqlRunner.DriverName(),
		DatabaseConnectionString: sqlRunner.ConnectionString(),
		ReportInterval:           durationjson.Duration(time.Second / 2),
		HealthAddress:            bbsHealthAddress,

		ClientLocketConfig: locketrunner.ClientLocketConfig(),

		EncryptionConfig: encryption.EncryptionConfig{
			EncryptionKeys: map[string]string{"label": "key"},
			ActiveKeyLabel: "label",
		},
		ConvergeRepeatInterval: durationjson.Duration(time.Hour),
		UUID:                   "bbs-bosh-boshy-bosh-bosh",

		CaFile:   serverCaFile,
		CertFile: path.Join(fixturesPath, "green-certs", "server.crt"),
		KeyFile:  path.Join(fixturesPath, "green-certs", "server.key"),

		LagerConfig: lagerflags.LagerConfig{
			LogLevel: lagerflags.DEBUG,
		},

		LoggregatorConfig: diego_logging_client.Config{
			BatchFlushInterval: 10 * time.Millisecond,
			BatchMaxSize:       1,
			UseV2API:           true,
			APIPort:            metricsPort,
			CACertPath:         metronCAFile,
			KeyPath:            metronClientKeyFile,
			CertPath:           metronClientCertFile,
		},
	}

	bbsConfig.ClientLocketConfig.LocketAddress = locketAddress
	locketClient, err := locket.NewClient(logger, bbsConfig.ClientLocketConfig)
	Expect(err).NotTo(HaveOccurred())
	locketHelper = test_helpers.NewLocketHelper(logger, locketClient)
})

var _ = AfterEach(func() {
	ginkgomon.Kill(bbsProcess)
	ginkgomon.Kill(locketProcess)

	// Make sure the healthcheck server is really gone before trying to start up
	// the bbs again in another test.
	Eventually(func() error {
		conn, err := net.Dial("tcp", bbsHealthAddress)
		if err == nil {
			conn.Close()
			return nil
		}

		return err
	}).Should(HaveOccurred())

	auctioneerServer.Close()
	testIngressServer.Stop()
	close(signalMetricsChan)

	sqlRunner.Reset()
})
