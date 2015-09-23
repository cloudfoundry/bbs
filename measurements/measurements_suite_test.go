package measurements_test

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"

	"github.com/cloudfoundry-incubator/bbs"
	"github.com/cloudfoundry-incubator/bbs/cmd/bbs/testrunner"
	"github.com/cloudfoundry-incubator/bbs/db/consul/test/consul_helpers"
	"github.com/cloudfoundry-incubator/bbs/db/etcd/test/etcd_helpers"
	"github.com/cloudfoundry-incubator/consuladapter"
	"github.com/cloudfoundry-incubator/consuladapter/consulrunner"
	"github.com/cloudfoundry/storeadapter/storerunner/etcdstorerunner"
	etcdclient "github.com/coreos/go-etcd/etcd"
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

var etcdPort int
var etcdUrl string
var etcdSSLConfig *etcdstorerunner.SSLConfig
var etcdRunner *etcdstorerunner.ETCDClusterRunner
var etcdClient *etcdclient.Client

var logger lager.Logger

var client bbs.Client
var bbsBinPath string
var bbsAddress string
var bbsURL *url.URL
var bbsArgs testrunner.Args
var bbsRunner *ginkgomon.Runner
var bbsProcess ifrit.Process
var consulSession *consuladapter.Session
var consulRunner *consulrunner.ClusterRunner
var etcdHelper *etcd_helpers.ETCDHelper
var consulHelper *consul_helpers.ConsulHelper
var auctioneerServer *ghttp.Server

func TestMeasurements(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Measurements Suite")
}

var _ = SynchronizedBeforeSuite(
	func() []byte {
		os.Setenv("GOMAXPROCS", strconv.Itoa(runtime.NumCPU()))
		bbsConfig, err := gexec.Build("github.com/cloudfoundry-incubator/bbs/cmd/bbs", "-race")
		Expect(err).NotTo(HaveOccurred())
		return []byte(bbsConfig)
	},
	func(bbsConfig []byte) {
		bbsBinPath = string(bbsConfig)
		SetDefaultEventuallyTimeout(15 * time.Second)
	},
)

var _ = SynchronizedAfterSuite(func() {
}, func() {
	gexec.CleanupBuildArtifacts()
})

var _ = AfterEach(func() {
	ginkgomon.Kill(bbsProcess)
	etcdRunner.Stop()
	consulRunner.Stop()
	auctioneerServer.Close()
})

var _ = JustBeforeEach(func() {
	etcdPort = 4001 + GinkgoParallelNode()
	etcdScheme := "http"
	if etcdSSLConfig != nil {
		etcdScheme = "https"
	}
	etcdUrl = fmt.Sprintf(etcdScheme+"://127.0.0.1:%d", etcdPort)
	etcdRunner = etcdstorerunner.NewETCDClusterRunner(etcdPort, 1, etcdSSLConfig)

	consulRunner = consulrunner.NewClusterRunner(
		9001+config.GinkgoConfig.ParallelNode*consulrunner.PortOffsetLength,
		1,
		"http",
	)

	consulRunner.Start()
	consulRunner.WaitUntilReady()
	consulRunner.Reset()

	etcdRunner.Start()
	etcdRunner.Reset()

	bbsArgs.ConsulCluster = consulRunner.ConsulCluster()
	bbsArgs.EtcdCluster = etcdUrl

	bbsRunner = testrunner.New(bbsBinPath, bbsArgs)
	bbsProcess = ginkgomon.Invoke(bbsRunner)
})

var _ = BeforeEach(func() {
	logger = lagertest.NewTestLogger("test")

	auctioneerServer = ghttp.NewServer()
	auctioneerServer.UnhandledRequestStatusCode = http.StatusAccepted
	auctioneerServer.AllowUnhandledRequests = true

	bbsAddress = fmt.Sprintf("127.0.0.1:%d", 6700+GinkgoParallelNode())

	bbsURL = &url.URL{
		Scheme: "http",
		Host:   bbsAddress,
	}

	bbsArgs = testrunner.Args{
		Address:               bbsAddress,
		AdvertiseURL:          bbsURL.String(),
		AuctioneerAddress:     auctioneerServer.URL(),
		MetricsReportInterval: 10 * time.Millisecond,

		EncryptionKeys: []string{"label:key"},
		ActiveKeyLabel: "label",
	}
})
