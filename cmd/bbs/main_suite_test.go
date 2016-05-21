package main_test

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/cloudfoundry-incubator/bbs"
	"github.com/cloudfoundry-incubator/bbs/cmd/bbs/testrunner"
	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/test_helpers"
	"github.com/cloudfoundry-incubator/consuladapter"
	"github.com/cloudfoundry-incubator/consuladapter/consulrunner"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/cloudfoundry/storeadapter/storerunner/etcdstorerunner"
	"github.com/cloudfoundry/storeadapter/storerunner/mysqlrunner"
	etcdclient "github.com/coreos/go-etcd/etcd"
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
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

	mySQLProcess ifrit.Process
	mySQLRunner  *mysqlrunner.MySQLRunner
)

func TestBBS(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BBS Cmd Suite")
}

var _ = SynchronizedBeforeSuite(
	func() []byte {
		bbsConfig, err := gexec.Build("github.com/cloudfoundry-incubator/bbs/cmd/bbs", "-race")
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
			mySQLRunner = mysqlrunner.NewMySQLRunner(fmt.Sprintf("diego_%d", GinkgoParallelNode()))
			mySQLProcess = ginkgomon.Invoke(mySQLRunner)
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
	ginkgomon.Kill(mySQLProcess)

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

		EncryptionKeys: []string{"label:key"},
		ActiveKeyLabel: "label",
	}
	if test_helpers.UseSQL() {
		bbsArgs.DatabaseDriver = "mysql"
		bbsArgs.DatabaseConnectionString = mySQLRunner.ConnectionString()
	}
	storeClient = etcd.NewStoreClient(etcdClient)
	consulHelper = test_helpers.NewConsulHelper(logger, consulClient)
})

var _ = AfterEach(func() {
	auctioneerServer.Close()
	testMetricsListener.Close()
	Eventually(testMetricsChan).Should(BeClosed())

	if test_helpers.UseSQL() {
		mySQLRunner.Reset()
	}
})
