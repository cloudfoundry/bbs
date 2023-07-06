package migrations_test

import (
	"code.cloudfoundry.org/bbs/db/migrations"
	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/migration"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AddRoutableToActualLrps", func() {
	var (
		migration migration.Migration
	)

	BeforeEach(func() {
		rawSQLDB.Exec("DROP TABLE actual_lrps;")

		migration = migrations.NewAddRoutableToActualLrps()
	})

	It("appends itself to the migration list", func() {
		Expect(migrations.AllMigrations()).To(ContainElement(migration))
	})

	Describe("Version", func() {
		It("returns the timestamp from which it was created", func() {
			Expect(migration.Version()).To(BeEquivalentTo(1686692176))
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

		It("adds the routable propety to actual lrps", func() {
			Expect(migration.Up(logger)).To(Succeed())

			_, err := rawSQLDB.Exec(
				helpers.RebindForFlavor(
					`insert into actual_lrps
						(process_guid, instance_index, domain, state, net_info,
						modification_tag_epoch, modification_tag_index, routable)
					values (?, ?, ?, ?, ?, ?, ?, ?)`,
					flavor,
				),
				"guid", 10, "cfapps", "running", "", "epoch", 0, true,
			)
			Expect(err).NotTo(HaveOccurred())

			var routable bool
			query := helpers.RebindForFlavor("select routable from actual_lrps limit 1", flavor)
			row := rawSQLDB.QueryRow(query)
			Expect(row.Scan(&routable)).To(Succeed())
			Expect(routable).To(BeTrue())
		})

		It("sets routable true for all existing actual_lrp rows in RUNNING state", func() {
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

			Expect(migration.Up(logger)).To(Succeed())

			var routable bool
			query := helpers.RebindForFlavor("select routable from actual_lrps limit 1", flavor)
			row := rawSQLDB.QueryRow(query)
			Expect(row.Scan(&routable)).To(Succeed())
			Expect(routable).To(BeTrue())
		})

		It("sets routable false for all existing actual_lrp rows not in RUNNING state", func() {
			_, err := rawSQLDB.Exec(
				helpers.RebindForFlavor(
					`INSERT INTO actual_lrps
						(process_guid, instance_index, domain, state, net_info,
						modification_tag_epoch, modification_tag_index)
					VALUES (?, ?, ?, ?, ?, ?, ?)`,
					flavor,
				),
				"guid-1", 10, "cfapps", "CLAIMED", "", "epoch", 0,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(migration.Up(logger)).To(Succeed())

			var routable bool
			query := helpers.RebindForFlavor("select routable from actual_lrps limit 1", flavor)
			row := rawSQLDB.QueryRow(query)
			Expect(row.Scan(&routable)).To(Succeed())
			Expect(routable).To(BeFalse())
		})

		It("is idempotent", func() {
			testIdempotency(rawSQLDB, migration, logger)
		})
	})
})
