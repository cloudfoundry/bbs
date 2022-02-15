package migrations_test

import (
	"code.cloudfoundry.org/bbs/db/migrations"
	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/migration"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AddInternalRoutesToActualLrp", func() {
	var (
		migration migration.Migration
	)

	BeforeEach(func() {
		rawSQLDB.Exec("DROP TABLE actual_lrps;")

		migration = migrations.NewAddInternalRoutesToActualLrp()
	})

	It("appends itself to the migration list", func() {
		Expect(migrations.AllMigrations()).To(ContainElement(migration))
	})

	Describe("Version", func() {
		It("returns the timestamp from which it was created", func() {
			Expect(migration.Version()).To(BeEquivalentTo(1643660541))
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

		It("add internal routes to actual_lrps", func() {
			Expect(migration.Up(logger)).To(Succeed())

			_, err := rawSQLDB.Exec(
				helpers.RebindForFlavor(
					`INSERT INTO actual_lrps
						(process_guid, instance_index, domain, state, net_info,
						modification_tag_epoch, modification_tag_index, internal_routes)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
					flavor,
				),
				"guid", 10, "cfapps", "RUNNING", "", "epoch", 0, "{}",
			)
			Expect(err).NotTo(HaveOccurred())

			var internalRoutes string
			query := helpers.RebindForFlavor("SELECT internal_routes FROM actual_lrps LIMIT 1", flavor)
			row := rawSQLDB.QueryRow(query)
			Expect(row.Scan(&internalRoutes)).To(Succeed())
			Expect(internalRoutes).To(Equal("{}"))
		})

		It("adds empty internal routes to all existing actual_lrp rows", func() {
			_, err := rawSQLDB.Exec(
				helpers.RebindForFlavor(
					`INSERT INTO actual_lrps
						(process_guid, instance_index, domain, state, net_info,
						modification_tag_epoch, modification_tag_index)
					VALUES (?, ?, ?, ?, ?, ?, ?)`,
					flavor,
				),
				"guid-1", 10, "cfapps", "RUNNING", "", "epoch", 0,
			)
			Expect(err).NotTo(HaveOccurred())

			_, err = rawSQLDB.Exec(
				helpers.RebindForFlavor(
					`INSERT INTO actual_lrps
						(process_guid, instance_index, domain, state, net_info,
						modification_tag_epoch, modification_tag_index)
					VALUES (?, ?, ?, ?, ?, ?, ?)`,
					flavor,
				),
				"guid-2", 11, "cfapps", "RUNNING", "", "epoch", 0,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(migration.Up(logger)).To(Succeed())

			var internalRoutesNullCount int
			query := helpers.RebindForFlavor("SELECT COUNT(*) FROM actual_lrps WHERE internal_routes IS NULL;", flavor)
			row := rawSQLDB.QueryRow(query)
			Expect(row.Scan(&internalRoutesNullCount)).To(Succeed())
			Expect(internalRoutesNullCount).To(Equal(2))
		})

		It("is idempotent", func() {
			testIdempotency(rawSQLDB, migration, logger)
		})
	})
})
