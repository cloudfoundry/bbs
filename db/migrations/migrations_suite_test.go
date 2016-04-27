package migrations_test

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"time"

	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/encryption"
	"github.com/cloudfoundry/storeadapter/storerunner/etcdstorerunner"
	etcdclient "github.com/coreos/go-etcd/etcd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/clock/fakeclock"
	"github.com/pivotal-golang/lager/lagertest"

	_ "github.com/go-sql-driver/mysql"

	"testing"
)

var (
	etcdPort    int
	etcdUrl     string
	etcdRunner  *etcdstorerunner.ETCDClusterRunner
	etcdClient  *etcdclient.Client
	storeClient etcd.StoreClient

	rawSQLDB *sql.DB

	cryptor   encryption.Cryptor
	fakeClock *fakeclock.FakeClock
	logger    *lagertest.TestLogger
)

func TestMigrations(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Migrations Suite")
}

var _ = BeforeSuite(func() {
	logger = lagertest.NewTestLogger("test")

	etcdPort = 4001 + GinkgoParallelNode()
	etcdUrl = fmt.Sprintf("http://127.0.0.1:%d", etcdPort)
	etcdRunner = etcdstorerunner.NewETCDClusterRunner(etcdPort, 1, nil)

	etcdRunner.Start()

	// mysql must be set up on localhost as described in the CONTRIBUTING.md doc
	// in diego-release.
	var err error
	rawSQLDB, err = sql.Open("mysql", "diego:diego_password@/")
	Expect(err).NotTo(HaveOccurred())
	Expect(rawSQLDB.Ping()).NotTo(HaveOccurred())

	_, err = rawSQLDB.Exec(fmt.Sprintf("CREATE DATABASE diego_%d", GinkgoParallelNode()))
	Expect(err).NotTo(HaveOccurred())

	rawSQLDB, err = sql.Open("mysql", fmt.Sprintf("diego:diego_password@/diego_%d", GinkgoParallelNode()))
	Expect(err).NotTo(HaveOccurred())
	Expect(rawSQLDB.Ping()).NotTo(HaveOccurred())

	encryptionKey, err := encryption.NewKey("label", "passphrase")
	Expect(err).NotTo(HaveOccurred())
	keyManager, err := encryption.NewKeyManager(encryptionKey, nil)
	Expect(err).NotTo(HaveOccurred())
	cryptor = encryption.NewCryptor(keyManager, rand.Reader)

	fakeClock = fakeclock.NewFakeClock(time.Now())
})

var _ = AfterSuite(func() {
	etcdRunner.Stop()

	_, err := rawSQLDB.Exec(fmt.Sprintf("DROP DATABASE diego_%d", GinkgoParallelNode()))
	Expect(err).NotTo(HaveOccurred())

	Expect(rawSQLDB.Close()).NotTo(HaveOccurred())
})

var _ = BeforeEach(func() {
	etcdRunner.Reset()

	etcdClient = etcdRunner.Client()
	etcdClient.SetConsistency(etcdclient.STRONG_CONSISTENCY)

	storeClient = etcd.NewStoreClient(etcdClient)
})
