package main_test

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/bbs/cmd/bbs/testrunner"
	"code.cloudfoundry.org/bbs/db/etcd"
	"code.cloudfoundry.org/bbs/test_helpers"
	"code.cloudfoundry.org/bbs/test_helpers/sqlrunner"
	"code.cloudfoundry.org/consuladapter"
	"code.cloudfoundry.org/consuladapter/consulrunner"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/cloudfoundry/storeadapter/storerunner/etcdstorerunner"
	etcdclient "github.com/coreos/go-etcd/etcd"
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"

	"testing"
	"time"
)

var (
	etcdPort    int
	etcdUrl     string
	etcdRunner  *etcdstorerunner.ETCDClusterRunner
	etcdClient  *etcdclient.Client
	storeClient etcd.StoreClient

	logger lager.Logger

	client              bbs.InternalClient
	bbsBinPath          string
	bbsAddress          string
	bbsHealthAddress    string
	bbsPort             int
	bbsURL              *url.URL
	bbsArgs             testrunner.Args
	bbsRunner           *ginkgomon.Runner
	bbsProcess          ifrit.Process
	consulRunner        *consulrunner.ClusterRunner
	consulClient        consuladapter.Client
	consulHelper        *test_helpers.ConsulHelper
	auctioneerServer    *ghttp.Server
	testMetricsListener net.PacketConn
	testMetricsChan     chan *events.Envelope

	sqlProcess ifrit.Process
	sqlRunner  sqlrunner.SQLRunner
)

func TestBBS(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BBS Cmd Suite")
}

var _ = SynchronizedBeforeSuite(
	func() []byte {
		bbsConfig, err := gexec.Build("code.cloudfoundry.org/bbs/cmd/bbs", "-race")
		Expect(err).NotTo(HaveOccurred())
		return []byte(bbsConfig)
	},
	func(bbsConfig []byte) {
		bbsBinPath = string(bbsConfig)
		SetDefaultEventuallyTimeout(15 * time.Second)

		etcdPort = 4001 + GinkgoParallelNode()
		etcdUrl = fmt.Sprintf("http://127.0.0.1:%d", etcdPort)
		etcdRunner = etcdstorerunner.NewETCDClusterRunner(etcdPort, 1, nil)

		if test_helpers.UseSQL() {
			dbName := fmt.Sprintf("diego_%d", GinkgoParallelNode())
			sqlRunner = test_helpers.NewSQLRunner(dbName)
			sqlProcess = ginkgomon.Invoke(sqlRunner)
		}

		consulRunner = consulrunner.NewClusterRunner(
			9001+config.GinkgoConfig.ParallelNode*consulrunner.PortOffsetLength,
			1,
			"http",
		)

		consulRunner.Start()
		consulRunner.WaitUntilReady()

		etcdRunner.Start()
	},
)

var _ = SynchronizedAfterSuite(func() {
	ginkgomon.Kill(sqlProcess)

	etcdRunner.Stop()
	consulRunner.Stop()
}, func() {
	gexec.CleanupBuildArtifacts()
})

var _ = BeforeEach(func() {
	logger = lagertest.NewTestLogger("test")

	etcdRunner.Reset()

	consulRunner.Reset()
	consulClient = consulRunner.NewClient()

	etcdClient = etcdRunner.Client()
	etcdClient.SetConsistency(etcdclient.STRONG_CONSISTENCY)

	auctioneerServer = ghttp.NewServer()
	auctioneerServer.UnhandledRequestStatusCode = http.StatusAccepted
	auctioneerServer.AllowUnhandledRequests = true

	bbsPort = 6700 + GinkgoParallelNode()*2
	bbsAddress = fmt.Sprintf("127.0.0.1:%d", bbsPort)
	bbsHealthAddress = fmt.Sprintf("127.0.0.1:%d", bbsPort+1)

	bbsURL = &url.URL{
		Scheme: "http",
		Host:   bbsAddress,
	}

	testMetricsListener, _ = net.ListenPacket("udp", "127.0.0.1:0")
	testMetricsChan = make(chan *events.Envelope, 1)
	go func() {
		defer GinkgoRecover()
		for {
			buffer := make([]byte, 1024)
			n, _, err := testMetricsListener.ReadFrom(buffer)
			if err != nil {
				close(testMetricsChan)
				return
			}

			var envelope events.Envelope
			err = proto.Unmarshal(buffer[:n], &envelope)
			Expect(err).NotTo(HaveOccurred())
			testMetricsChan <- &envelope
		}
	}()

	port, err := strconv.Atoi(strings.TrimPrefix(testMetricsListener.LocalAddr().String(), "127.0.0.1:"))
	Expect(err).NotTo(HaveOccurred())

	client = bbs.NewClient(bbsURL.String())

	bbsArgs = testrunner.Args{
		Address:               bbsAddress,
		AdvertiseURL:          bbsURL.String(),
		AuctioneerAddress:     auctioneerServer.URL(),
		ConsulCluster:         consulRunner.ConsulCluster(),
		DropsondePort:         port,
		EtcdCluster:           etcdUrl,
		MetricsReportInterval: 10 * time.Millisecond,
		HealthAddress:         bbsHealthAddress,

		EncryptionKeys:         []string{"label:key"},
		ActiveKeyLabel:         "label",
		ConvergeRepeatInterval: time.Hour,
	}
	if test_helpers.UseSQL() {
		bbsArgs.DatabaseDriver = sqlRunner.DriverName()
		bbsArgs.DatabaseConnectionString = sqlRunner.ConnectionString()
	}
	storeClient = etcd.NewStoreClient(etcdClient)
	consulHelper = test_helpers.NewConsulHelper(logger, consulClient)
})

var _ = AfterEach(func() {
	ginkgomon.Kill(bbsProcess)

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
	testMetricsListener.Close()
	Eventually(testMetricsChan).Should(BeClosed())

	if test_helpers.UseSQL() {
		sqlRunner.Reset()
	}
})
