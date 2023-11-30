package migrations_test

import (
	"code.cloudfoundry.org/bbs/db/migrations"
	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/migration"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AddAvailabilityZoneToActualLrps", func() {
	var (
		migration migration.Migration
	)

	BeforeEach(func() {
		rawSQLDB.Exec("DROP TABLE actual_lrps;")

		migration = migrations.NewAddAvailabilityZoneToActualLrps()
	})

	It("appends itself to the migration list", func() {
		Expect(migrations.AllMigrations()).To(ContainElement(migration))
	})

	Describe("Version", func() {
		It("returns the timestamp from which it was created", func() {
			Expect(migration.Version()).To(BeEquivalentTo(1698182853))
		})
	})

	Describe("Up", func() {
		BeforeEach(func() {
			initialMigration := migrations.NewInitSQL()
			initialMigration.SetDBFlavor(flavor)
			initialMigration.SetClock(fakeClock)
			testUpInTransaction(rawSQLDB, initialMigration, logger)

			migration.SetDBFlavor(flavor)
		})

		It("adds the availability_zone propety to actual lrps", func() {
			testUpInTransaction(rawSQLDB, migration, logger)

			_, err := rawSQLDB.Exec(
				helpers.RebindForFlavor(
					`insert into actual_lrps
						(process_guid, instance_index, domain, state, net_info,
						modification_tag_epoch, modification_tag_index, availability_zone)
					values (?, ?, ?, ?, ?, ?, ?, ?)`,
					flavor,
				),
				"guid", 10, "cfapps", "running", "", "epoch", 0, "some-zone",
			)
			Expect(err).NotTo(HaveOccurred())

			var availabilityZone string
			query := helpers.RebindForFlavor("select availability_zone from actual_lrps limit 1", flavor)
			row := rawSQLDB.QueryRow(query)
			Expect(row.Scan(&availabilityZone)).To(Succeed())
			Expect(availabilityZone).To(Equal("some-zone"))
		})

		It("is idempotent", func() {
			testIdempotency(rawSQLDB, migration, logger)
		})
	})
})
