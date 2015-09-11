package etcd_test

import (
	"encoding/json"
	"fmt"
	"time"

	fakeauctioneer "github.com/cloudfoundry-incubator/bbs/auctionhandlers/fakes"
	fakecellhandlers "github.com/cloudfoundry-incubator/bbs/cellhandlers/fakes"
	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/db/consul"
	"github.com/cloudfoundry-incubator/bbs/db/consul/test/consul_helpers"
	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/db/etcd/test/etcd_helpers"
	"github.com/cloudfoundry-incubator/bbs/format"
	"github.com/cloudfoundry-incubator/bbs/models"
	faketaskworkpool "github.com/cloudfoundry-incubator/bbs/taskworkpool/fakes"
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

const receptorURL = "http://some-receptor-url"

var etcdPort int
var etcdUrl string
var etcdRunner *etcdstorerunner.ETCDClusterRunner
var storeClient etcd.StoreClient
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
var workPoolCreateError error
var fakeTaskCompletionClient *faketaskworkpool.FakeTaskCompletionClient

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

	Expect(workPoolCreateError).ToNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	etcdRunner.Stop()
	consulRunner.Stop()
})

var _ = BeforeEach(func() {
	logger = lagertest.NewTestLogger("test")

	auctioneerClient = new(fakeauctioneer.FakeClient)
	cellClient = new(fakecellhandlers.FakeClient)
	etcdRunner.Reset()

	consulRunner.Reset()
	consulSession = consulRunner.NewSession("a-session")

	etcdClient := etcdRunner.Client()
	etcdClient.SetConsistency(etcdclient.STRONG_CONSISTENCY)
	storeClient = etcd.NewStoreClient(etcdClient)
	etcdHelper = etcd_helpers.NewETCDHelper(format.ENCODED_PROTO, storeClient)
	consulHelper = consul_helpers.NewConsulHelper(consulSession)
	cellDB = consul.NewConsul(consulSession)
	fakeTaskCompletionClient = new(faketaskworkpool.FakeTaskCompletionClient)

	etcdDB = etcd.NewETCD(format.ENCODED_PROTO, storeClient, auctioneerClient, cellClient, cellDB, clock, fakeTaskCompletionClient)
})

func registerCell(cell models.CellPresence) {
	var err error
	jsonBytes, err := json.Marshal(cell)
	Expect(err).NotTo(HaveOccurred())

	_, err = consulSession.SetPresence(consul.CellSchemaPath(cell.CellID), jsonBytes)
	Expect(err).NotTo(HaveOccurred())
}
