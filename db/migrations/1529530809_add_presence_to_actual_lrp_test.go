package migrations_test

import (
	"fmt"

	"code.cloudfoundry.org/bbs/db/migrations"
	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/migration"
	"code.cloudfoundry.org/bbs/models"

	. "github.com/onsi/ginkgo/v2"
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
			initialMigration.SetDBFlavor(flavor)
			initialMigration.SetClock(fakeClock)
			testUpInTransaction(rawSQLDB, initialMigration, logger)

			migration.SetDBFlavor(flavor)

		})

		It("add presence to the actual_lrps and defaults it to ordinary", func() {
			testUpInTransaction(rawSQLDB, migration, logger)

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
			query := helpers.RebindForFlavor("SELECT presence FROM actual_lrps LIMIT 1", flavor)
			row := rawSQLDB.QueryRow(query)
			Expect(row.Scan(&presence)).To(Succeed())
			Expect(presence).To(Equal(fmt.Sprintf("%d", models.ActualLRP_ORDINARY)))
		})

		It("adds presence as a primary key so that duplicate entries with different presence do not violate the unique constraint", func() {
			testUpInTransaction(rawSQLDB, migration, logger)

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

			_, err = rawSQLDB.Exec(
				helpers.RebindForFlavor(
					`INSERT INTO actual_lrps
						(process_guid, instance_index, domain, state, net_info,
						modification_tag_epoch, modification_tag_index, presence)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
					flavor,
				),
				"guid", 10, "cfapps", "RUNNING", "", "epoch", 0, models.ActualLRP_EVACUATING,
			)
			Expect(err).NotTo(HaveOccurred())

		})

		Context("with preexisting data with evacuating set to true", func() {
			BeforeEach(func() {
				_, err := rawSQLDB.Exec(
					helpers.RebindForFlavor(
						`INSERT INTO actual_lrps
						(process_guid, instance_index, domain, state, net_info,
						modification_tag_epoch, modification_tag_index, evacuating)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
						flavor,
					),
					"guid", 1, "cfapps", "RUNNING", "", "epoch", 0, true,
				)
				Expect(err).NotTo(HaveOccurred())

			})

			It("does not error on LRPs with the same process_guid + index when changing the primary key", func() {
				_, err := rawSQLDB.Exec(
					helpers.RebindForFlavor(
						`INSERT INTO actual_lrps
						(process_guid, instance_index, domain, state, net_info,
						modification_tag_epoch, modification_tag_index, evacuating)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
						flavor,
					),
					"guid", 1, "cfapps", "RUNNING", "", "epoch", 0, false,
				)
				Expect(err).NotTo(HaveOccurred())
				testUpInTransaction(rawSQLDB, migration, logger)
			})

			It("sets the presence of the evacuating row to evacuating", func() {
				testUpInTransaction(rawSQLDB, migration, logger)

				var presence string
				query := helpers.RebindForFlavor("SELECT presence FROM actual_lrps WHERE evacuating = true LIMIT 1", flavor)
				row := rawSQLDB.QueryRow(query)
				Expect(row.Scan(&presence)).To(Succeed())
				Expect(presence).To(Equal(fmt.Sprintf("%d", models.ActualLRP_EVACUATING)))
			})

			It("is idempotent even with preexisting data", func() {
				testIdempotency(rawSQLDB, migration, logger)
			})
		})
	})
})
