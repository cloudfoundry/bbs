package migrations_test

import (
	"code.cloudfoundry.org/bbs/db/migrations"
	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/migration"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AddPresenceToActualLrp", func() {
	var (
		migration migration.Migration
	)

	BeforeEach(func() {
		rawSQLDB.Exec("DROP TABLE actual_lrps;")

		migration = migrations.NewAddPresenceToActualLrp()
	})

	It("appends itself to the migration list", func() {
		Expect(migrations.AllMigrations()).To(ContainElement(migration))
	})

	Describe("Version", func() {
		It("returns the timestamp from which it was created", func() {
			Expect(migration.Version()).To(BeEquivalentTo(1529530809))
		})
	})

	Describe("Up", func() {
		BeforeEach(func() {
			initialMigration := migrations.NewInitSQL()
			initialMigration.SetRawSQLDB(rawSQLDB)
			initialMigration.SetDBFlavor(flavor)
			initialMigration.SetClock(fakeClock)
			Expect(initialMigration.Up(logger)).To(Succeed())

			migration.SetRawSQLDB(rawSQLDB)
			migration.SetDBFlavor(flavor)
		})

		It("add presence to the actual_lrps and defaults it to \"\"", func() {
			Expect(migration.Up(logger)).To(Succeed())

			_, err := rawSQLDB.Exec(
				helpers.RebindForFlavor(
					`INSERT INTO actual_lrps
						(process_guid, instance_index, domain, state, net_info,
						modification_tag_epoch, modification_tag_index)
					VALUES (?, ?, ?, ?, ?, ?, ?)`,
					flavor,
				),
				"guid", 10, "cfapps", "RUNNING", "", "epoch", 0,
			)
			Expect(err).NotTo(HaveOccurred())

			var presence string
			query := helpers.RebindForFlavor("select presence from actual_lrps limit 1", flavor)
			row := rawSQLDB.QueryRow(query)
			Expect(row.Scan(&presence)).To(Succeed())
			Expect(presence).To(Equal(""))
		})

		It("is idempotent", func() {
			testIdempotency(rawSQLDB, migration, logger)
		})
	})
})
