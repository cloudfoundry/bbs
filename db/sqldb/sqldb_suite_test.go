package sqldb_test

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"time"

	"github.com/cloudfoundry-incubator/bbs/db/sqldb"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/clock/fakeclock"
	"github.com/pivotal-golang/lager/lagertest"

	_ "github.com/go-sql-driver/mysql"

	"testing"
)

var (
	db        *sql.DB
	sqlDB     *sqldb.SQLDB
	fakeClock *fakeclock.FakeClock
	logger    *lagertest.TestLogger
)

func TestSql(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Sql Suite")
}

var _ = BeforeSuite(func() {
	var err error
	fakeClock = fakeclock.NewFakeClock(time.Now())
	logger = lagertest.NewTestLogger("sql-db")

	db, err = sql.Open("mysql", "root:password@/")
	Expect(err).NotTo(HaveOccurred())
	Expect(db.Ping()).NotTo(HaveOccurred())

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE diego_%d", GinkgoParallelNode()))
	Expect(err).NotTo(HaveOccurred())

	db, err = sql.Open("mysql", fmt.Sprintf("root:password@/diego_%d?parseTime=true", GinkgoParallelNode()))
	Expect(err).NotTo(HaveOccurred())
	Expect(db.Ping()).NotTo(HaveOccurred())

	createTables(db)

	sqlDB = sqldb.NewSQLDB(db, fakeClock)
})

var _ = AfterEach(func() {
	truncateTables(db)
})

var _ = AfterSuite(func() {
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

func createTables(db *sql.DB) {
	for _, query := range createTablesSQL {
		result, err := db.Exec(query)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.RowsAffected()).To(BeEquivalentTo(0))
	}
}

var truncateTablesSQL = []string{
	"TRUNCATE TABLE domains;",
	"TRUNCATE TABLE configurations;",
}

var createTablesSQL = []string{
	createDomainSQL,
	createEncryptionKeyLabelsSQL,
}

const createDomainSQL = `CREATE TABLE domains(
	domain varchar(255) PRIMARY KEY,
	expireTime timestamp
);`

const createEncryptionKeyLabelsSQL = `CREATE TABLE configurations(
	id varchar(255) PRIMARY KEY,
	value varchar(255)
);`

func randStr(str_size int) string {
	alphanum := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, str_size)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}
