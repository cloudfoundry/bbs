package sqlrunner

import (
	"database/sql"
	"fmt"
	"os"

	"code.cloudfoundry.org/bbs/test_helpers/tool_helpers"
	"github.com/denisenkom/go-mssqldb"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// MSSQLRunner is responsible for creating and tearing down a test database in
// a Microsoft SQL instance. This runner assumes mssql is already running
// on localhost or Azure and has firewall set properly.
// You can follow this guide https://docs.microsoft.com/en-us/azure/sql-database/sql-database-configure-firewall-settings to create firewall on Azure.
// To run the test, you need to specific MSSQL_BASE_CONNECTION_STRING in env.
// example: SQL_FLAVOR="mssql" MSSQL_BASE_CONNECTION_STRING="server=<server>.database.windows.net;user id=<username>;password=<password>;port=1433"
// Be noted that you should not set a database in MSSQL_BASE_CONNECTION_STRING, the test will create one for you.
type MSSQLRunner struct {
	sqlDBName string
	db        *sql.DB
}

func NewMSSQLRunner(sqlDBName string) *MSSQLRunner {
	return &MSSQLRunner{
		sqlDBName: sqlDBName,
	}
}

func (m *MSSQLRunner) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	defer GinkgoRecover()

	db_connection_string := os.Getenv("MSSQL_BASE_CONNECTION_STRING")
	if db_connection_string == "" {
		panic(fmt.Sprintf("You must specify MSSQL_BASE_CONNECTION_STRING when running test for mssql"))
	}

	var err error
	m.db, err = sql.Open("mssql", db_connection_string)
	Expect(err).NotTo(HaveOccurred())
	Expect(m.db.Ping()).NotTo(HaveOccurred())

	_, err = m.db.Exec(fmt.Sprintf("CREATE DATABASE %s", m.sqlDBName))

	err = tool_helpers.Retry(5, func() error {
		var err error
		m.db, err = sql.Open("mssql", m.ConnectionString())
		err = m.db.Ping()
		return err
	})
	Expect(err).NotTo(HaveOccurred())

	close(ready)

	<-signals

	m.db.Exec(fmt.Sprintf("DROP DATABASE %s", m.sqlDBName))
	m.db = nil

	return nil
}

func (m *MSSQLRunner) ConnectionString() string {
	return fmt.Sprintf("%s;database=%s", os.Getenv("MSSQL_BASE_CONNECTION_STRING"), m.sqlDBName)
}

func (m *MSSQLRunner) DriverName() string {
	return "mssql"
}

func (m *MSSQLRunner) DB() *sql.DB {
	return m.db
}

func (m *MSSQLRunner) ResetTables(tables []string) {
	for _, name := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s", name)
		result, err := m.db.Exec(query)
		switch err := err.(type) {
		case mssql.Error:
			if err.Number == 4701 {
				// missing table error, it's fine because we're trying to truncate it
				continue
			}
		}

		Expect(err).NotTo(HaveOccurred())
		Expect(result.RowsAffected()).To(BeEquivalentTo(0))
	}
}

func (m *MSSQLRunner) Reset() {
	m.ResetTables([]string{"domains", "configurations", "tasks", "desired_lrps", "actual_lrps", "locks"})
}
