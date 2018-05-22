package migrations_test

import (
	"code.cloudfoundry.org/bbs/db/migrations"
	"code.cloudfoundry.org/bbs/migration"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SpikeAddSuspectToActualLRP", func() {
	var (
		migration migration.Migration
	)

	BeforeEach(func() {
		// TODO: db cleanup here

		migration = migrations.NewSpikeAddSuspectToActualLRP()
	})

	It("appends itself to the migration list", func() {
		Expect(migrations.AllMigrations()).To(ContainElement(migration))
	})

	Describe("Version", func() {
		It("returns the timestamp from which it was created", func() {
			Expect(migration.Version()).To(BeEquivalentTo(1526935932))
		})
	})

	Describe("Up", func() {
		BeforeEach(func() {
			migration.SetRawSQLDB(rawSQLDB)
			migration.SetDBFlavor(flavor)

			// TODO: db setup here
		})

		It("TODO: CHANGE ME", func() {
			// TODO: test here
		})

		It("is idempotent", func() {
			testIdempotency(rawSQLDB, migration, logger)
		})
	})
})
