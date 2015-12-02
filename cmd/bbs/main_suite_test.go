package main_test

import (
	"crypto/rand"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/cloudfoundry-incubator/bbs"
	"github.com/cloudfoundry-incubator/bbs/cmd/bbs/testrunner"
	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/db/etcd/test/etcd_helpers"
	"github.com/cloudfoundry-incubator/bbs/encryption"
	"github.com/cloudfoundry-incubator/bbs/format"
	"github.com/cloudfoundry-incubator/bbs/test_helpers"
	"github.com/cloudfoundry-incubator/consuladapter"
	"github.com/cloudfoundry-incubator/consuladapter/consulrunner"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/cloudfoundry/storeadapter/storerunner/etcdstorerunner"
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

	"github.com/pivotal-golang/clock/fakeclock"
	"testing"
	"time"
)

var etcdPort int
var etcdUrl string
var etcdRunner *etcdstorerunner.ETCDClusterRunner
var etcdClient *etcdclient.Client
var storeClient etcd.StoreClient

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
var consulHelper *test_helpers.ConsulHelper
var auctioneerServer *ghttp.Server
var testMetricsListener net.PacketConn
var testMetricsChan chan *events.Envelope

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

		consulRunner.Start()
		consulRunner.WaitUntilReady()

		etcdRunner.Start()
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

	consulRunner.Reset()
	consulSession = consulRunner.NewSession("a-session")

	etcdClient = etcdRunner.Client()
	etcdClient.SetConsistency(etcdclient.STRONG_CONSISTENCY)

	auctioneerServer = ghttp.NewServer()
	auctioneerServer.UnhandledRequestStatusCode = http.StatusAccepted
	auctioneerServer.AllowUnhandledRequests = true

	bbsAddress = fmt.Sprintf("127.0.0.1:%d", 6700+GinkgoParallelNode())

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

		EncryptionKeys: []string{"label:key"},
		ActiveKeyLabel: "label",
	}
	storeClient = etcd.NewStoreClient(etcdClient)
	encryptionKey, err := encryption.NewKey("label", "key")
	Expect(err).NotTo(HaveOccurred())
	keyManager, err := encryption.NewKeyManager(encryptionKey, nil)
	Expect(err).NotTo(HaveOccurred())
	cryptor := encryption.NewCryptor(keyManager, rand.Reader)
	etcdHelper = etcd_helpers.NewETCDHelper(format.ENCRYPTED_PROTO, cryptor, storeClient, &fakeclock.FakeClock{})
	consulHelper = test_helpers.NewConsulHelper(consulSession)
})

var _ = AfterEach(func() {
	auctioneerServer.Close()
	testMetricsListener.Close()
	Eventually(testMetricsChan).Should(BeClosed())
})
