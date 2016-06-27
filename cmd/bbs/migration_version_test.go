package main_test

import (
	"database/sql"
	"encoding/json"
	"io"
	"os"

	"code.cloudfoundry.org/bbs/cmd/bbs/testrunner"
	"code.cloudfoundry.org/bbs/db/etcd"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/test_helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("Migration Version", func() {
	var migrationFixtureFilePath, migrationFilePath string

	BeforeEach(func() {
		migrationFixtureFilePath = "fixtures/9999999999_sql_test_migration.go.bs"
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

		bbsConfig, err := gexec.Build("code.cloudfoundry.org/bbs/cmd/bbs", "-race")
		Expect(err).NotTo(HaveOccurred())
		bbsBinPath = string(bbsConfig)

		value, err := json.Marshal(models.Version{CurrentVersion: 100, TargetVersion: 100})
		// write initial version
		_, err = storeClient.Set(etcd.VersionKey, value, etcd.NO_TTL)
		Expect(err).NotTo(HaveOccurred())

		bbsRunner = testrunner.WaitForMigration(bbsBinPath, bbsArgs)
		bbsProcess = ginkgomon.Invoke(bbsRunner)
	})

	AfterEach(func() {
		err := os.RemoveAll(migrationFilePath)
		Expect(err).NotTo(HaveOccurred())

		bbsConfig, err := gexec.Build("code.cloudfoundry.org/bbs/cmd/bbs", "-race")
		Expect(err).NotTo(HaveOccurred())
		bbsBinPath = string(bbsConfig)
	})

	if !test_helpers.UseSQL() {
		Context("Running Migrations Without SQL", func() {
			It("loads and runs the given migrations up to the last etcd migration", func() {
				response, err := storeClient.Get(etcd.VersionKey, false, false)
				Expect(err).NotTo(HaveOccurred())

				var version models.Version
				err = json.Unmarshal([]byte(response.Node.Value), &version)
				Expect(err).NotTo(HaveOccurred())

				// the final etcd migration
				Expect(version.CurrentVersion).To(BeEquivalentTo(1451635200))
				Expect(version.TargetVersion).To(BeEquivalentTo(1451635200))
			})
		})
	}

	if test_helpers.UseSQL() {
		Context("Running Migrations With SQL", func() {
			var (
				sqlConn *sql.DB
				err     error
			)

			BeforeEach(func() {
				sqlConn, err = sql.Open(sqlRunner.DriverName(), sqlRunner.ConnectionString())
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				sqlConn.Close()
			})

			It("loads and runs all the migrations", func() {
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

				var count int
				err = sqlConn.QueryRow(`SELECT count(*) FROM information_schema.tables WHERE table_name = 'sweet_table'`).Scan(&count)
				Expect(err).NotTo(HaveOccurred())
				Expect(count).To(Equal(1))
			})
		})
	}
})
