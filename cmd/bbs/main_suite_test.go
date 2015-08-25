package main_test

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/cloudfoundry-incubator/bbs"
	"github.com/cloudfoundry-incubator/bbs/cmd/bbs/testrunner"
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
var etcdRunner *etcdstorerunner.ETCDClusterRunner
var etcdClient *etcdclient.Client

var logger lager.Logger

var client bbs.Client
var bbsBinPath string
var bbsAddress string
var bbsArgs testrunner.Args
var bbsRunner *ginkgomon.Runner
var bbsProcess ifrit.Process
var consulSession *consuladapter.Session
var consulRunner *consulrunner.ClusterRunner
var etcdHelper *etcd_helpers.ETCDHelper
var auctioneerServer *ghttp.Server

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

		consulRunner = consulrunner.NewClusterRunner(
			9001+config.GinkgoConfig.ParallelNode*consulrunner.PortOffsetLength,
			1,
			"http",
		)

		etcdRunner.Start()
		consulRunner.Start()
	},
)

var _ = SynchronizedAfterSuite(func() {
	etcdRunner.Stop()
	consulRunner.Stop()
}, func() {
	gexec.CleanupBuildArtifacts()
})

var _ = BeforeEach(func() {
	logger = lagertest.NewTestLogger("test")

	etcdRunner.Reset()
	etcdClient = etcdRunner.Client()
	etcdClient.SetConsistency(etcdclient.STRONG_CONSISTENCY)

	consulRunner.Reset()

	auctioneerServer = ghttp.NewServer()
	auctioneerServer.UnhandledRequestStatusCode = http.StatusAccepted
	auctioneerServer.AllowUnhandledRequests = true

	bbsAddress = fmt.Sprintf("127.0.0.1:%d", 6700+GinkgoParallelNode())

	bbsURL := &url.URL{
		Scheme: "http",
		Host:   bbsAddress,
	}

	client = bbs.NewClient(bbsURL.String())

	bbsArgs = testrunner.Args{
		Address:           bbsAddress,
		AuctioneerAddress: auctioneerServer.URL(),
		EtcdCluster:       etcdUrl,
		ConsulCluster:     consulRunner.ConsulCluster(),
	}
	bbsRunner = testrunner.New(bbsBinPath, bbsArgs)

	bbsProcess = ginkgomon.Invoke(bbsRunner)
	etcdHelper = etcd_helpers.NewETCDHelper(etcdClient)
})

var _ = AfterEach(func() {
	ginkgomon.Kill(bbsProcess)
	auctioneerServer.Close()
})
