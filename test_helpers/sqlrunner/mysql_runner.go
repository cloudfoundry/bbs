package sqlrunner

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/go-sql-driver/mysql"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// MySQLRunner is responsible for creating and tearing down a test database in
// a local MySQL instance. This runner assumes mysql is already running
// locally, and does not start or stop the mysql service.  mysql must be set up
// on localhost as described in the CONTRIBUTING.md doc in diego-release.
type MySQLRunner struct {
	sqlDBName string
	db        *sql.DB
}

func NewMySQLRunner(sqlDBName string) *MySQLRunner {
	return &MySQLRunner{
		sqlDBName: sqlDBName,
	}
}

func (m *MySQLRunner) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	defer GinkgoRecover()

	var err error
	m.db, err = sql.Open("mysql", "diego:diego_password@/")
	Expect(err).NotTo(HaveOccurred())
	Expect(m.db.Ping()).NotTo(HaveOccurred())

	m.db.Exec(fmt.Sprintf("DROP DATABASE %s", m.sqlDBName))
	_, err = m.db.Exec(fmt.Sprintf("CREATE DATABASE %s", m.sqlDBName))
	Expect(err).NotTo(HaveOccurred())

	m.db, err = sql.Open("mysql", fmt.Sprintf("diego:diego_password@/%s", m.sqlDBName))
	Expect(err).NotTo(HaveOccurred())
	Expect(m.db.Ping()).NotTo(HaveOccurred())

	close(ready)

	<-signals

	_, err = m.db.Exec(fmt.Sprintf("DROP DATABASE %s", m.sqlDBName))
	Expect(err).NotTo(HaveOccurred())
	Expect(m.db.Close()).To(Succeed())

	return nil
}

func (m *MySQLRunner) ConnectionString() string {
	return fmt.Sprintf("diego:diego_password@/%s", m.sqlDBName)
}

func (m *MySQLRunner) DriverName() string {
	return "mysql"
}

func (m *MySQLRunner) DB() *sql.DB {
	return m.db
}

func (m *MySQLRunner) Reset() {
	var truncateTablesSQL = []string{
		"TRUNCATE TABLE domains",
		"TRUNCATE TABLE configurations",
		"TRUNCATE TABLE tasks",
		"TRUNCATE TABLE desired_lrps",
		"TRUNCATE TABLE actual_lrps",
	}
	for _, query := range truncateTablesSQL {
		result, err := m.db.Exec(query)
		switch err := err.(type) {
		case *mysql.MySQLError:
			if err.Number == 1146 {
				// missing table error, it's fine because we're trying to truncate it
				continue
			}
		}

		Expect(err).NotTo(HaveOccurred())
		Expect(result.RowsAffected()).To(BeEquivalentTo(0))
	}
}
