package helpers_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/test_helpers"
	"code.cloudfoundry.org/lager"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestHelpers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Helpers Suite")
}

var (
	db                     *sql.DB
	ctx                    context.Context
	dbName                 string
	dbDriverName           string
	dbBaseConnectionString string
	dbFlavor               string
	tableName              string
)

var _ = BeforeEach(func() {
	dbName = fmt.Sprintf("diego_%d", GinkgoParallelNode())

	if test_helpers.UsePostgres() {
		dbDriverName = "postgres"
		user, ok := os.LookupEnv("POSTGRES_USER")
		if !ok {
			user = "diego"
		}
		password, ok := os.LookupEnv("POSTGRES_PASSWORD")
		if !ok {
			password = "diego_pw"
		}
		dbBaseConnectionString = fmt.Sprintf("postgres://%s:%s@localhost/", user, password)
		dbFlavor = helpers.Postgres
	} else if test_helpers.UseMySQL() {
		dbDriverName = "mysql"
		user, ok := os.LookupEnv("MYSQL_USER")
		if !ok {
			user = "diego"
		}
		password, ok := os.LookupEnv("MYSQL_PASSWORD")
		if !ok {
			password = "diego_password"
		}
		dbBaseConnectionString = fmt.Sprintf("%s:%s@/", user, password)
		dbFlavor = helpers.MySQL
	} else {
		panic("Unsupported driver")
	}

	logger := lager.NewLogger("helper-suite-test")

	// mysql must be set up on localhost as described in the CONTRIBUTING.md doc
	// in diego-release.
	var err error
	db, err = helpers.Connect(logger, dbDriverName, dbBaseConnectionString, "", false)
	Expect(err).NotTo(HaveOccurred())
	Expect(db.Ping()).NotTo(HaveOccurred())

	ctx = context.Background()

	// Ensure that if another test failed to clean up we can still proceed
	db.ExecContext(ctx, fmt.Sprintf("DROP DATABASE %s", dbName))

	_, err = db.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", dbName))
	Expect(err).NotTo(HaveOccurred())

	Expect(db.Close()).To(Succeed())

	connStringWithDB := fmt.Sprintf("%s%s", dbBaseConnectionString, dbName)
	db, err = helpers.Connect(logger, dbDriverName, connStringWithDB, "", false)
	Expect(err).NotTo(HaveOccurred())
	Expect(db.Ping()).NotTo(HaveOccurred())
})

var _ = AfterEach(func() {
	logger := lager.NewLogger("helper-suite-test")

	Expect(db.Close()).NotTo(HaveOccurred())
	db, err := helpers.Connect(logger, dbDriverName, dbBaseConnectionString, "", false)
	Expect(err).NotTo(HaveOccurred())
	Expect(db.Ping()).NotTo(HaveOccurred())
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE diego_%d", GinkgoParallelNode()))
	Expect(err).NotTo(HaveOccurred())
	Expect(db.Close()).NotTo(HaveOccurred())
})
