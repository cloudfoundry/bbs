package test_helpers

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"code.cloudfoundry.org/bbs/db/migrations"
	"code.cloudfoundry.org/bbs/db/sqldb"
	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/db/sqldb/helpers/monitor"
	"code.cloudfoundry.org/bbs/encryption"
	"code.cloudfoundry.org/bbs/guidprovider"
	"code.cloudfoundry.org/bbs/migration"
	"code.cloudfoundry.org/bbs/test_helpers/sqlrunner"
	"code.cloudfoundry.org/clock/fakeclock"
	"code.cloudfoundry.org/diego-logging-client/testhelpers"
	"code.cloudfoundry.org/lager/v3"
	"github.com/tedsuo/ifrit"
	ginkgomon "github.com/tedsuo/ifrit/ginkgomon_v2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	mysqlFlavor    = "mysql"
	mysql8Flavor   = "mysql8"
	postgresFlavor = "postgres"
)

func UseSQL() bool {
	return true
}

func driver() string {
	flavor := os.Getenv("DB")
	if flavor == "" {
		flavor = postgresFlavor
	}
	return flavor
}

func UseMySQL() bool {
	return driver() == mysqlFlavor || driver() == mysql8Flavor
}

func UsePostgres() bool {
	return driver() == postgresFlavor
}

func NewSQLRunner(dbName string) sqlrunner.SQLRunner {
	var sqlRunner sqlrunner.SQLRunner

	if UseMySQL() {
		sqlRunner = sqlrunner.NewMySQLRunner(dbName)
	} else if UsePostgres() {
		sqlRunner = sqlrunner.NewPostgresRunner(dbName)
	} else {
		panic(fmt.Sprintf("driver '%s' is not supported", driver()))
	}

	return sqlRunner
}

func ReplaceQuestionMarks(queryString string) string {
	strParts := strings.Split(queryString, "?")
	for i := 1; i < len(strParts); i++ {
		strParts[i-1] = fmt.Sprintf("%s$%d", strParts[i-1], i)
	}
	return strings.Join(strParts, "")
}

type TestDatabase struct {
	DB               *sqldb.SQLDB
	sqlConn          *sql.DB
	sqlProcess       ifrit.Process
	migrationProcess ifrit.Process
}

func SetupTestDatabase(logger lager.Logger) *TestDatabase {
	d := &TestDatabase{}
	dbName := fmt.Sprintf("diego_%d", GinkgoParallelProcess())
	sqlRunner := NewSQLRunner(dbName)
	d.sqlProcess = ginkgomon.Invoke(sqlRunner)

	var err error
	d.sqlConn, err = helpers.Connect(
		logger,
		sqlRunner.DriverName(),
		sqlRunner.ConnectionString(),
		"",
		false,
	)
	Expect(err).NotTo(HaveOccurred())

	dbMonitor := monitor.New()
	monitoredDB := helpers.NewMonitoredDB(d.sqlConn, dbMonitor)

	convergenceWorkers := 20
	updateWorkers := 1000
	encryptionKey, err := encryption.NewKey("label", "passphrase")
	Expect(err).NotTo(HaveOccurred())
	keyManager, err := encryption.NewKeyManager(encryptionKey, nil)
	Expect(err).NotTo(HaveOccurred())
	cryptor := encryption.NewCryptor(keyManager, rand.Reader)

	fakeClock := fakeclock.NewFakeClock(time.Now())
	fakeMetronClient := &testhelpers.FakeIngressClient{}
	d.DB = sqldb.NewSQLDB(
		monitoredDB,
		convergenceWorkers,
		updateWorkers,
		cryptor,
		guidprovider.DefaultGuidProvider,
		fakeClock,
		sqlRunner.DriverName(),
		fakeMetronClient,
	)
	err = d.DB.CreateConfigurationsTable(context.Background(), logger)
	Expect(err).NotTo(HaveOccurred())

	migrationsDone := make(chan struct{})

	migrationManager := migration.NewManager(
		logger,
		d.DB,
		d.sqlConn,
		cryptor,
		migrations.AllMigrations(),
		migrationsDone,
		fakeClock,
		sqlRunner.DriverName(),
		fakeMetronClient,
	)
	d.migrationProcess = ifrit.Invoke(migrationManager)
	Eventually(migrationsDone).Should(BeClosed())

	return d
}

func (d *TestDatabase) Stop() {
	Expect(d.sqlConn.Close()).To(Succeed())
	ginkgomon.Kill(d.sqlProcess)
	ginkgomon.Kill(d.migrationProcess)
}
