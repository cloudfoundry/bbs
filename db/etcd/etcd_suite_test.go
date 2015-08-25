package etcd_test

import (
	"fmt"
	"time"

	fakeauctioneer "github.com/cloudfoundry-incubator/bbs/auctionhandlers/fakes"
	fakecellhandlers "github.com/cloudfoundry-incubator/bbs/cellhandlers/fakes"
	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/db/consul"
	"github.com/cloudfoundry-incubator/bbs/db/consul/test/consul_helpers"
	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/db/etcd/test/etcd_helpers"
	fakeHelpers "github.com/cloudfoundry-incubator/bbs/db/etcd/test/etcd_helpers/fakes"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/consuladapter"
	"github.com/cloudfoundry-incubator/consuladapter/consulrunner"
	"github.com/cloudfoundry/gunk/workpool"
	"github.com/cloudfoundry/storeadapter/storerunner/etcdstorerunner"
	etcdclient "github.com/coreos/go-etcd/etcd"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/clock/fakeclock"
	"github.com/pivotal-golang/lager/lagertest"

	"testing"
)

const receptorURL = "http://some-receptor-url"

var etcdPort int
var etcdUrl string
var etcdRunner *etcdstorerunner.ETCDClusterRunner
var etcdClient *etcdclient.Client

var consulRunner *consulrunner.ClusterRunner
var consulSession *consuladapter.Session

var auctioneerClient *fakeauctioneer.FakeClient
var cellClient *fakecellhandlers.FakeClient

var logger *lagertest.TestLogger
var clock *fakeclock.FakeClock
var etcdHelper *etcd_helpers.ETCDHelper
var consulHelper *consul_helpers.ConsulHelper

var cellDB db.CellDB
var etcdDB db.DB
var taskCBWorkPool *workpool.WorkPool
var workPoolCreateError error
var fakeTaskCBFactory *fakeHelpers.FakeTaskCallbackFactory

func TestDB(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ETCD DB Suite")
}

var _ = BeforeSuite(func() {
	clock = fakeclock.NewFakeClock(time.Unix(0, 1138))

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

	taskCBWorkPool, workPoolCreateError = workpool.NewWorkPool(1)
	Expect(workPoolCreateError).ToNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	etcdRunner.Stop()
	consulRunner.Stop()
	taskCBWorkPool.Stop()
})

var _ = BeforeEach(func() {
	logger = lagertest.NewTestLogger("test")

	auctioneerClient = new(fakeauctioneer.FakeClient)
	cellClient = new(fakecellhandlers.FakeClient)
	etcdRunner.Reset()

	consulRunner.Reset()
	consulSession = consulRunner.NewSession("a-session")

	etcdClient = etcdRunner.Client()
	etcdClient.SetConsistency(etcdclient.STRONG_CONSISTENCY)
	etcdHelper = etcd_helpers.NewETCDHelper(etcdClient)
	consulHelper = consul_helpers.NewConsulHelper(consulSession)
	cellDB = consul.NewConsul(consulSession)
	fakeTaskCBFactory = new(fakeHelpers.FakeTaskCallbackFactory)
	fakeTaskCBFactory.TaskCallbackWorkReturns(func() {})

	etcdDB = etcd.NewETCD(etcdClient, auctioneerClient, cellClient, cellDB, clock, taskCBWorkPool, fakeTaskCBFactory.TaskCallbackWork)
})

func registerCell(cell models.CellPresence) {
	var err error
	jsonBytes, err := models.ToJSON(cell)
	Expect(err).NotTo(HaveOccurred())

	_, err = consulSession.SetPresence(consul.CellSchemaPath(cell.CellID), jsonBytes)
	Expect(err).NotTo(HaveOccurred())
}
