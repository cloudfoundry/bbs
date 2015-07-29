package etcd_test

import (
	"fmt"
	"time"

	fakeauctioneer "github.com/cloudfoundry-incubator/auctioneer/fakes"
	"github.com/cloudfoundry-incubator/bbs/db/etcd/internal/test_helpers"
	"github.com/cloudfoundry-incubator/consuladapter"
	"github.com/cloudfoundry-incubator/consuladapter/consulrunner"
	"github.com/cloudfoundry/storeadapter/storerunner/etcdstorerunner"
	etcdclient "github.com/coreos/go-etcd/etcd"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/clock/fakeclock"
	"github.com/pivotal-golang/lager/lagertest"

	"testing"
)

var etcdPort int
var etcdUrl string
var etcdRunner *etcdstorerunner.ETCDClusterRunner
var etcdClient *etcdclient.Client

var consulRunner *consulrunner.ClusterRunner
var consulSession *consuladapter.Session

var auctioneerClient *fakeauctioneer.FakeClient

var logger *lagertest.TestLogger
var clock *fakeclock.FakeClock
var testHelper *test_helpers.TestHelper

func TestDB(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ETCD DB Suite")
}

var _ = BeforeSuite(func() {
	logger = lagertest.NewTestLogger("test")

	clock = fakeclock.NewFakeClock(time.Unix(0, 1138))

	etcdPort = 4001 + GinkgoParallelNode()
	etcdUrl = fmt.Sprintf("http://127.0.0.1:%d", etcdPort)
	etcdRunner = etcdstorerunner.NewETCDClusterRunner(etcdPort, 1, nil)

	consulRunner = consulrunner.NewClusterRunner(
		9001+config.GinkgoConfig.ParallelNode*consulrunner.PortOffsetLength,
		1,
		"http",
	)

	logger = lagertest.NewTestLogger("test")

	consulRunner.Start()
	consulRunner.WaitUntilReady()

	etcdRunner.Start()
})

var _ = AfterSuite(func() {
	etcdRunner.Stop()
	consulRunner.Stop()
})

var _ = BeforeEach(func() {
	auctioneerClient = new(fakeauctioneer.FakeClient)
	etcdRunner.Reset()

	etcdClient = etcdRunner.Client()
	etcdClient.SetConsistency(etcdclient.STRONG_CONSISTENCY)
	testHelper = test_helpers.NewTestHelper(etcdClient)
})
