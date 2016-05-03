package sqldb_test

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"os"
	"time"

	thepackagedb "github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/db/migrations"
	"github.com/cloudfoundry-incubator/bbs/db/sqldb"
	"github.com/cloudfoundry-incubator/bbs/encryption"
	"github.com/cloudfoundry-incubator/bbs/format"
	"github.com/cloudfoundry-incubator/bbs/guidprovider/fakes"
	"github.com/cloudfoundry-incubator/bbs/migration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/clock/fakeclock"
	"github.com/pivotal-golang/lager/lagertest"
	"github.com/tedsuo/ifrit"

	_ "github.com/go-sql-driver/mysql"

	"testing"
)

var (
	db               *sql.DB
	sqlDB            thepackagedb.DB
	fakeClock        *fakeclock.FakeClock
	fakeGUIDProvider *fakes.FakeGUIDProvider
	logger           *lagertest.TestLogger
	cryptor          encryption.Cryptor
	serializer       format.Serializer
	migrationProcess ifrit.Process
)

func TestSql(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "SQL DB Suite")
}

var _ = BeforeSuite(func() {
	var err error
	fakeClock = fakeclock.NewFakeClock(time.Now())
	fakeGUIDProvider = &fakes.FakeGUIDProvider{}
	logger = lagertest.NewTestLogger("sql-db")

	// mysql must be set up on localhost as described in the CONTRIBUTING.md doc
	// in diego-release.
	db, err = sql.Open("mysql", "diego:diego_password@/")
	Expect(err).NotTo(HaveOccurred())
	Expect(db.Ping()).NotTo(HaveOccurred())

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE diego_%d", GinkgoParallelNode()))
	Expect(err).NotTo(HaveOccurred())

	db, err = sql.Open("mysql", fmt.Sprintf("diego:diego_password@/diego_%d", GinkgoParallelNode()))
	Expect(err).NotTo(HaveOccurred())
	Expect(db.Ping()).NotTo(HaveOccurred())

	encryptionKey, err := encryption.NewKey("label", "passphrase")
	Expect(err).NotTo(HaveOccurred())
	keyManager, err := encryption.NewKeyManager(encryptionKey, nil)
	Expect(err).NotTo(HaveOccurred())
	cryptor = encryption.NewCryptor(keyManager, rand.Reader)
	serializer = format.NewSerializer(cryptor)

	internalSQLDB := sqldb.NewSQLDB(db, 5, 5, format.ENCRYPTED_PROTO, cryptor, fakeGUIDProvider, fakeClock)
	err = internalSQLDB.CreateConfigurationsTable(logger)
	if err != nil {
		logger.Fatal("sql-failed-create-configurations-table", err)
	}

	sqlDB = internalSQLDB

	migrationsDone := make(chan struct{})

	migrationManager := migration.NewManager(logger,
		nil,
		nil,
		sqlDB,
		db,
		cryptor,
		migrations.Migrations,
		migrationsDone,
		fakeClock,
	)

	migrationProcess = ifrit.Invoke(migrationManager)
})

var _ = AfterEach(func() {
	truncateTables(db)
	fakeGUIDProvider.NextGUIDReturns("", nil)
})

var _ = AfterSuite(func() {
	migrationProcess.Signal(os.Kill)
	_, err := db.Exec(fmt.Sprintf("DROP DATABASE diego_%d", GinkgoParallelNode()))
	Expect(err).NotTo(HaveOccurred())

	Expect(db.Close()).NotTo(HaveOccurred())
})

func truncateTables(db *sql.DB) {
	for _, query := range truncateTablesSQL {
		result, err := db.Exec(query)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.RowsAffected()).To(BeEquivalentTo(0))
	}
}

var truncateTablesSQL = []string{
	"TRUNCATE TABLE domains",
	"TRUNCATE TABLE configurations",
	"TRUNCATE TABLE tasks",
	"TRUNCATE TABLE desired_lrps",
	"TRUNCATE TABLE actual_lrps",
}

func randStr(strSize int) string {
	alphanum := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, strSize)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}
