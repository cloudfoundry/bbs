package migrations_test

import (
	"database/sql"

	"code.cloudfoundry.org/bbs/db/migrations"
	"code.cloudfoundry.org/bbs/format"
	"code.cloudfoundry.org/bbs/migration"
	"code.cloudfoundry.org/bbs/test_helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = FDescribe("Additional RunInfos SQL Migration", func() {
	if test_helpers.UseSQL() {
		var (
			migration    migration.Migration
			serializer   format.Serializer
			migrationErr error
		)

		BeforeEach(func() {
			migration = migrations.NewAdditionalRunInfos()
			serializer = format.NewSerializer(cryptor)
		})

		It("appends itself to the migration list", func() {
			Expect(migrations.Migrations).To(ContainElement(migration))
		})

		Describe("Version", func() {
			It("returns the timestamp from which it was created", func() {
				Expect(migration.Version()).To(BeEquivalentTo(1470245948))
			})
		})

		Describe("Up", func() {
			JustBeforeEach(func() {
				migration.SetStoreClient(storeClient)
				migration.SetRawSQLDB(rawSQLDB)
				migration.SetCryptor(cryptor)
				migration.SetClock(fakeClock)
				migration.SetDBFlavor(sqlRunner.DriverName())
				migrationErr = migration.Up(logger)
			})

			Context("when there is existing data in the database", func() {
				BeforeEach(func() {
					var err error
					_, err = rawSQLDB.Exec(migrations.CreateDesiredLRPsSQL)
					Expect(err).NotTo(HaveOccurred())

					_, err = rawSQLDB.Exec(migrations.CreateActualLRPsSQL)
					Expect(err).NotTo(HaveOccurred())
				})

				It("should update the tables start migration", func() {
					var value string
					err := rawSQLDB.QueryRow(`SELECT run_info_guid FROM desired_lrps`).Scan(&value)
					Expect(err).To(MatchError(sql.ErrNoRows))
					err = rawSQLDB.QueryRow(`SELECT run_info_guid FROM actual_lrps`).Scan(&value)
					Expect(err).To(MatchError(sql.ErrNoRows))
					err = rawSQLDB.QueryRow(`SELECT * FROM run_infos`).Scan(&value)
					Expect(err).To(MatchError(sql.ErrNoRows))
				})
			})
		})

		Describe("Down", func() {
			It("returns a not implemented error", func() {
				Expect(migration.Down(logger)).To(HaveOccurred())
			})
		})
	}
})
