package main_test

import (
	"database/sql"
	"encoding/json"
	"io"
	"os"

	"github.com/cloudfoundry-incubator/bbs/cmd/bbs/testrunner"
	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("Migration Version", func() {
	var migrationFixtureFilePath, migrationFilePath string

	BeforeEach(func() {
		migrationFixtureFilePath = "fixtures/9999999999_sql_test_migration.go"
		migrationFilePath = "../../db/migrations/9999999999_sql_test_migration.go"
		migrationFixtureFile, err := os.Open(migrationFixtureFilePath)
		Expect(err).NotTo(HaveOccurred())

		migrationFile, err := os.Create(migrationFilePath)
		Expect(err).NotTo(HaveOccurred())

		_, err = io.Copy(migrationFile, migrationFixtureFile)
		Expect(err).NotTo(HaveOccurred())

		err = migrationFixtureFile.Close()
		Expect(err).NotTo(HaveOccurred())

		err = migrationFile.Close()
		Expect(err).NotTo(HaveOccurred())

		bbsConfig, err := gexec.Build("github.com/cloudfoundry-incubator/bbs/cmd/bbs", "-race")
		Expect(err).NotTo(HaveOccurred())
		bbsBinPath = string(bbsConfig)

		value, err := json.Marshal(models.Version{CurrentVersion: 100, TargetVersion: 100})
		// write initial version
		_, err = storeClient.Set(etcd.VersionKey, value, etcd.NO_TTL)
		Expect(err).NotTo(HaveOccurred())

		bbsRunner = testrunner.New(bbsBinPath, bbsArgs)
		bbsProcess = ginkgomon.Invoke(bbsRunner)
	})

	AfterEach(func() {
		ginkgomon.Kill(bbsProcess)
		err := os.RemoveAll(migrationFilePath)
		Expect(err).NotTo(HaveOccurred())

		bbsConfig, err := gexec.Build("github.com/cloudfoundry-incubator/bbs/cmd/bbs", "-race")
		Expect(err).NotTo(HaveOccurred())
		bbsBinPath = string(bbsConfig)
	})

	Context("Running Migrations Without SQL", func() {
		It("loads and runs the given migrations up to the last etcd migration", func() {
			if !useSQL {
				response, err := storeClient.Get(etcd.VersionKey, false, false)
				Expect(err).NotTo(HaveOccurred())

				var version models.Version
				err = json.Unmarshal([]byte(response.Node.Value), &version)
				Expect(err).NotTo(HaveOccurred())

				// the final etcd migration
				Expect(version.CurrentVersion).To(BeEquivalentTo(1451635200))
				Expect(version.TargetVersion).To(BeEquivalentTo(1451635200))
			}

		})
	})

	Context("Running Migrations With SQL", func() {
		var (
			sqlConn *sql.DB
			err     error
		)

		BeforeEach(func() {
			if useSQL {
				sqlConn, err = sql.Open("mysql", mySQLRunner.ConnectionString())
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("loads and runs all the migrations", func() {
			if useSQL {
				var versionJSON string
				err := sqlConn.QueryRow(
					`SELECT value FROM configurations WHERE id = 'version'`,
				).Scan(&versionJSON)
				Expect(err).NotTo(HaveOccurred())

				var version models.Version
				err = json.Unmarshal([]byte(versionJSON), &version)
				Expect(err).NotTo(HaveOccurred())

				// the sql test migration
				Expect(version.CurrentVersion).To(BeEquivalentTo(9999999999))
				Expect(version.TargetVersion).To(BeEquivalentTo(9999999999))

				var table interface{}
				err = sqlConn.QueryRow(`SHOW TABLES LIKE 'sweet_table'`).Scan(&table)
				Expect(err).NotTo(HaveOccurred())
			}
		})
	})
})
