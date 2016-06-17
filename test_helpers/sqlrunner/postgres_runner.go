package sqlrunner

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/lib/pq"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// PostgresRunner is responsible for creating and tearing down a test database in
// a local Postgres instance. This runner assumes mysql is already running
// locally, and does not start or stop the mysql service.  mysql must be set up
// on localhost as described in the CONTRIBUTING.md doc in diego-release.
type PostgresRunner struct {
	db        *sql.DB
	sqlDBName string
}

func NewPostgresRunner(sqlDBName string) *PostgresRunner {
	return &PostgresRunner{
		sqlDBName: sqlDBName,
	}
}

func (p *PostgresRunner) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	defer GinkgoRecover()

	var err error
	p.db, err = sql.Open("postgres", "postgres://diego:diego_pw@localhost")
	Expect(err).NotTo(HaveOccurred())
	Expect(p.db.Ping()).NotTo(HaveOccurred())

	p.db.Exec(fmt.Sprintf("DROP DATABASE %s", p.sqlDBName))
	_, err = p.db.Exec(fmt.Sprintf("CREATE DATABASE %s", p.sqlDBName))
	Expect(err).NotTo(HaveOccurred())

	p.db, err = sql.Open("postgres", fmt.Sprintf("postgres://diego:diego_pw@localhost/%s", p.sqlDBName))
	Expect(err).NotTo(HaveOccurred())
	Expect(p.db.Ping()).NotTo(HaveOccurred())

	close(ready)

	<-signals

	// We need to close the connection to the database we want to drop before dropping it.
	p.db.Close()
	p.db, err = sql.Open("postgres", "postgres://diego:diego_pw@localhost")
	Expect(err).NotTo(HaveOccurred())

	_, err = p.db.Exec(fmt.Sprintf("DROP DATABASE %s", p.sqlDBName))
	Expect(err).NotTo(HaveOccurred())
	Expect(p.db.Close()).To(Succeed())

	return nil
}

func (p *PostgresRunner) ConnectionString() string {
	return fmt.Sprintf("postgres://diego:diego_pw@localhost/%s", p.sqlDBName)
}

func (p *PostgresRunner) DriverName() string {
	return "postgres"
}

func (p *PostgresRunner) DB() *sql.DB {
	return p.db
}

func (p *PostgresRunner) Reset() {
	var truncateTablesSQL = []string{
		"TRUNCATE TABLE domains",
		"TRUNCATE TABLE configurations",
		"TRUNCATE TABLE tasks",
		"TRUNCATE TABLE desired_lrps",
		"TRUNCATE TABLE actual_lrps",
	}
	for _, query := range truncateTablesSQL {
		result, err := p.db.Exec(query)

		switch err := err.(type) {
		case *pq.Error:
			if err.Code == "42P01" {
				// missing table error, it's fine because we're trying to truncate it
				continue
			}
		}

		Expect(err).NotTo(HaveOccurred())
		Expect(result.RowsAffected()).To(BeEquivalentTo(0))
	}
}
