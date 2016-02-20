package sqldb_test

import (
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

	_, err = db.Query(fmt.Sprintf("CREATE DATABASE diego_%d", GinkgoParallelNode()))
	Expect(err).NotTo(HaveOccurred())

	db, err = sql.Open("mysql", fmt.Sprintf("root:password@/diego_%d?parseTime=true", GinkgoParallelNode()))
	Expect(err).NotTo(HaveOccurred())
	Expect(db.Ping()).NotTo(HaveOccurred())

	createDomains(db)

	sqlDB = sqldb.NewSQLDB(db, fakeClock)
})

var _ = AfterEach(func() {
	truncateDB(db)
})

var _ = AfterSuite(func() {
	_, err := db.Query(fmt.Sprintf("DROP DATABASE diego_%d", GinkgoParallelNode()))
	Expect(err).NotTo(HaveOccurred())

	Expect(db.Close()).NotTo(HaveOccurred())
})

func truncateDB(db *sql.DB) {
	_, err := db.Query(truncateTablesQuery)
	Expect(err).NotTo(HaveOccurred())
}

func createDomains(db *sql.DB) {
	_, err := db.Query(createDomainQuery)
	Expect(err).NotTo(HaveOccurred())
}

const truncateTablesQuery = "TRUNCATE TABLE domains;"

const createDomainQuery = `CREATE TABLE domains(
	domain varchar(255) PRIMARY KEY,
	expireTime timestamp
);`
