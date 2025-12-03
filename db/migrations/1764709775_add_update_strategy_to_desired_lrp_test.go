package migrations_test

import (
	"fmt"
	"time"

	"code.cloudfoundry.org/bbs/db/migrations"
	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/migration"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/clock/fakeclock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Add Update Strategy to Desired LRPs", func() {
	var (
		mig migration.Migration
	)

	BeforeEach(func() {
		fakeClock = fakeclock.NewFakeClock(time.Now())
		rawSQLDB.Exec("DROP TABLE domains;")
		rawSQLDB.Exec("DROP TABLE tasks;")
		rawSQLDB.Exec("DROP TABLE desired_lrps;")
		rawSQLDB.Exec("DROP TABLE actual_lrps;")

		mig = migrations.NewAddUpdateStrategyToDesiredLRPs()
	})

	It("appends itself to the migration list", func() {
		Expect(migrations.AllMigrations()).To(ContainElement(mig))
	})

	Describe("Version", func() {
		It("returns the timestamp from which it was created", func() {
			Expect(mig.Version()).To(BeEquivalentTo(1764709775))
		})
	})

	Describe("Up", func() {
		BeforeEach(func() {
			initialMigrations := []migration.Migration{
				migrations.NewInitSQL(),
				migrations.NewIncreaseRunInfoColumnSize(),
			}

			for _, m := range initialMigrations {
				m.SetDBFlavor(flavor)
				m.SetClock(fakeClock)
				testUpInTransaction(rawSQLDB, m, logger)
			}

			mig.SetDBFlavor(flavor)
			mig.SetClock(fakeClock)
		})

		It("sets rolling update_strategy by default on new desired lrps", func() {
			testUpInTransaction(rawSQLDB, mig, logger)
			_, err := rawSQLDB.Exec(
				helpers.RebindForFlavor(
					`INSERT INTO desired_lrps
						  (process_guid, domain, log_guid, instances, memory_mb,
							  disk_mb, rootfs, routes, volume_placement, modification_tag_epoch, run_info)
						  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
					flavor,
				),
				"guid", "domain",
				"log guid", 2, 1, 1, "rootfs", "routes", "volumes yo", "1", "run info",
			)
			Expect(err).NotTo(HaveOccurred())

			var fetchedUpdateStrategy string
			query := helpers.RebindForFlavor("select update_strategy from desired_lrps limit 1", flavor)
			row := rawSQLDB.QueryRow(query)
			Expect(row.Scan(&fetchedUpdateStrategy)).NotTo(HaveOccurred())
			Expect(fetchedUpdateStrategy).To(Equal(fmt.Sprintf("%d", models.DesiredLRP_Rolling)))
		})

		Context("when there are desiredLRPs already in the db", func() {
			BeforeEach(func() {
				_, err := rawSQLDB.Exec(
					helpers.RebindForFlavor(
						`INSERT INTO desired_lrps
						  (process_guid, domain, log_guid, instances, memory_mb,
							  disk_mb, rootfs, routes, volume_placement, modification_tag_epoch, run_info)
						  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
						flavor,
					),
					"meow-guid", "domain",
					"log guid", 2, 1, 1, "rootfs", "routes", "volumes yo", "1", "run info",
				)
				Expect(err).NotTo(HaveOccurred())
			})

			It("sets the update strategy to rolling", func() {
				testUpInTransaction(rawSQLDB, mig, logger)

				var fetchedUpdateStrategy string
				query := helpers.RebindForFlavor("select update_strategy from desired_lrps where process_guid = 'meow-guid' limit 1", flavor)
				row := rawSQLDB.QueryRow(query)
				Expect(row.Scan(&fetchedUpdateStrategy)).NotTo(HaveOccurred())
				Expect(fetchedUpdateStrategy).To(Equal(fmt.Sprintf("%d", models.DesiredLRP_Rolling)))
			})
		})

		It("is idempotent", func() {
			testIdempotency(rawSQLDB, mig, logger)
		})
	})
})
