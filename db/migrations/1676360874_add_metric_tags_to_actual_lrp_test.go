package migrations_test

import (
	"code.cloudfoundry.org/bbs/db/migrations"
	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/migration"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AddMetricTagsToActualLrp", func() {
	var (
		migration migration.Migration
	)

	BeforeEach(func() {
		rawSQLDB.Exec("DROP TABLE actual_lrps;")

		migration = migrations.NewAddMetricTagsToActualLrp()
	})

	It("appends itself to the migration list", func() {
		Expect(migrations.AllMigrations()).To(ContainElement(migration))
	})

	Describe("Version", func() {
		It("returns the timestamp from which it was created", func() {
			Expect(migration.Version()).To(BeEquivalentTo(1676360874))
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

		It("adds metric tags to actual_lrps", func() {
			Expect(migration.Up(logger)).To(Succeed())

			_, err := rawSQLDB.Exec(
				helpers.RebindForFlavor(
					`INSERT INTO actual_lrps
						(process_guid, instance_index, domain, state, net_info,
						modification_tag_epoch, modification_tag_index, metric_tags)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
					flavor,
				),
				"guid", 10, "cfapps", "RUNNING", "", "epoch", 0, "{}",
			)
			Expect(err).NotTo(HaveOccurred())

			var metricTags string
			query := helpers.RebindForFlavor("SELECT metric_tags FROM actual_lrps LIMIT 1", flavor)
			row := rawSQLDB.QueryRow(query)
			Expect(row.Scan(&metricTags)).To(Succeed())
			Expect(metricTags).To(Equal("{}"))
		})

		It("adds empty metric tags to all existing actual_lrp rows", func() {
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

			var metricTagsNullCount int
			query := helpers.RebindForFlavor("SELECT COUNT(*) FROM actual_lrps WHERE metric_tags IS NULL;", flavor)
			row := rawSQLDB.QueryRow(query)
			Expect(row.Scan(&metricTagsNullCount)).To(Succeed())
			Expect(metricTagsNullCount).To(Equal(2))
		})

		It("is idempotent", func() {
			testIdempotency(rawSQLDB, migration, logger)
		})
	})
})
