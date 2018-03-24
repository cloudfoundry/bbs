package migrations_test

import (
	"code.cloudfoundry.org/bbs/db/migrations"
	"code.cloudfoundry.org/bbs/migration"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SpikeAddSuspectFieldToActualLrp", func() {
	var (
		migration migration.Migration
	)

	BeforeEach(func() {
		rawSQLDB.Exec("DROP TABLE domains;")
		rawSQLDB.Exec("DROP TABLE tasks;")
		rawSQLDB.Exec("DROP TABLE desired_lrps;")
		rawSQLDB.Exec("DROP TABLE actual_lrps;")

		migration = migrations.NewSpikeAddSuspectFieldToActualLrp()
	})

	It("appends itself to the migration list", func() {
		Expect(migrations.AllMigrations()).To(ContainElement(migration))
	})

	Describe("Version", func() {
		It("returns the timestamp from which it was created", func() {
			Expect(migration.Version()).To(BeEquivalentTo(1521738802))
		})
	})

	Describe("Up", func() {
		BeforeEach(func() {
			migration.SetRawSQLDB(rawSQLDB)
			migration.SetDBFlavor(flavor)

			initialMigration := migrations.NewInitSQL()

			initialMigration.SetRawSQLDB(rawSQLDB)
			initialMigration.SetDBFlavor(flavor)
			initialMigration.SetClock(fakeClock)
			err := initialMigration.Up(logger)
			Expect(err).NotTo(HaveOccurred())
		})

		It("is idempotent", func() {
			testIdempotency(rawSQLDB, migration, logger)
		})
	})
})
