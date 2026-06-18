package db_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"code.cloudfoundry.org/diego-db-helpers/guidprovider/guidproviderfakes"
	"code.cloudfoundry.org/diego-db-helpers/sqldb/helpers"
	"code.cloudfoundry.org/diego-db-helpers/sqldb/helpers/monitor"
	"code.cloudfoundry.org/diego-db-helpers/testhelpers"
	"code.cloudfoundry.org/lager/v3/lagertest"
	sqldb "code.cloudfoundry.org/locket/db"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

var (
	rawDB                                *sql.DB
	sqlDB                                *sqldb.SQLDB
	logger                               *lagertest.TestLogger
	ctx                                  context.Context
	fakeGUIDProvider                     *guidproviderfakes.FakeGUIDProvider
	dbDriverName, dbBaseConnectionString string
	dbFlavor                             string
	sqlHelper                            helpers.SQLHelper
	dbParams                             *helpers.ConnectParams
)

func TestSql(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "SQL DB Suite")
}

var _ = BeforeSuite(func() {
	var err error
	logger = lagertest.NewTestLogger("sql-db")
	ctx = context.Background()

	if testhelpers.UsePostgres() {
		dbDriverName = "postgres"
		user, ok := os.LookupEnv("DB_USER")
		if !ok {
			user = "diego"
		}
		password, ok := os.LookupEnv("DB_PASSWORD")
		if !ok {
			password = "diego_pw"
		}
		dbBaseConnectionString = fmt.Sprintf("postgres://%s:%s@localhost/", user, password)
		dbFlavor = helpers.Postgres
	} else if testhelpers.UseMySQL() {
		dbDriverName = "mysql"
		user, ok := os.LookupEnv("DB_USER")
		if !ok {
			user = "diego"
		}
		password, ok := os.LookupEnv("DB_PASSWORD")
		if !ok {
			password = "diego_password"
		}
		dbBaseConnectionString = fmt.Sprintf("%s:%s@/", user, password)
		dbFlavor = helpers.MySQL
	} else {
		panic("Unsupported driver")
	}

	// mysql must be set up on localhost as described in the CONTRIBUTING.md doc
	// in diego-release.
	dbParams = &helpers.ConnectParams{
		DriverName:                    dbDriverName,
		DatabaseConnectionString:      dbBaseConnectionString,
		SqlCACertFile:                 "",
		SqlEnableIdentityVerification: false,
	}
	rawDB, err = helpers.Connect(logger, dbParams)

	Expect(err).NotTo(HaveOccurred())
	Expect(rawDB.Ping()).NotTo(HaveOccurred())

	_, err = rawDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS diego_%d", GinkgoParallelProcess()))
	Expect(err).ToNot(HaveOccurred())
	_, err = rawDB.Exec(fmt.Sprintf("CREATE DATABASE diego_%d", GinkgoParallelProcess()))
	Expect(err).NotTo(HaveOccurred())

	connStringWithDB := fmt.Sprintf("%sdiego_%d", dbBaseConnectionString, GinkgoParallelProcess())
	dbParams.DatabaseConnectionString = connStringWithDB
	rawDB, err = helpers.Connect(logger, dbParams)
	Expect(err).NotTo(HaveOccurred())
	Expect(rawDB.Ping()).NotTo(HaveOccurred())

	fakeGUIDProvider = &guidproviderfakes.FakeGUIDProvider{}
	db := helpers.NewMonitoredDB(rawDB, monitor.New())
	sqlDB = sqldb.NewSQLDB(db, dbFlavor, fakeGUIDProvider)
	err = sqlDB.CreateLockTable(ctx, logger)
	Expect(err).NotTo(HaveOccurred())
	err = sqlDB.CreateHealthCheckTable(ctx, logger)
	Expect(err).NotTo(HaveOccurred())

	sqlHelper = helpers.NewSQLHelper(dbFlavor)

	// ensures sqlDB matches the db.DB interface
	var _ sqldb.LockDB = sqlDB
})

var _ = BeforeEach(func() {

	// ensure that all sqldb functions being tested only require one connection
	// to operate, otherwise a deadlock can be caused in bbs. For more
	// information see https://www.pivotaltracker.com/story/show/136754083
	rawDB.SetMaxOpenConns(1)
})

var _ = AfterEach(func() {
	truncateTables(rawDB)
})

var _ = AfterSuite(func() {
	Expect(rawDB.Close()).NotTo(HaveOccurred())
	dbParams.DatabaseConnectionString = dbBaseConnectionString
	rawDB, err := helpers.Connect(logger, dbParams)
	Expect(err).NotTo(HaveOccurred())
	Expect(rawDB.Ping()).NotTo(HaveOccurred())
	_, err = rawDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS diego_%d", GinkgoParallelProcess()))
	Expect(err).NotTo(HaveOccurred())
	Expect(rawDB.Close()).NotTo(HaveOccurred())
})

func truncateTables(db *sql.DB) {
	for _, query := range truncateTablesSQL {
		result, err := db.Exec(query)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.RowsAffected()).To(BeEquivalentTo(0))
	}
}

var truncateTablesSQL = []string{
	"TRUNCATE TABLE locks",
	"TRUNCATE TABLE locket_health_check",
}
