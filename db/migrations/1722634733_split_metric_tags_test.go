package migrations_test

import (
	"encoding/json"

	"code.cloudfoundry.org/bbs/db/migrations"
	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/format"
	"code.cloudfoundry.org/bbs/migration"
	"code.cloudfoundry.org/bbs/models"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SplitMetricTags", func() {
	var (
		migration  migration.Migration
		serializer format.Serializer
	)

	BeforeEach(func() {
		rawSQLDB.Exec("DROP TABLE desired_lrps;")

		serializer = format.NewSerializer(cryptor)

		initialMigration := migrations.NewInitSQL()
		initialMigration.SetDBFlavor(flavor)
		initialMigration.SetClock(fakeClock)
		testUpInTransaction(rawSQLDB, initialMigration, logger)

		migration = migrations.NewSplitMetricTags()
	})

	It("appends itself to the migration list", func() {
		Expect(migrations.AllMigrations()).To(ContainElement(migration))
	})

	Describe("Version", func() {
		It("returns the timestamp from which it was created", func() {
			Expect(migration.Version()).To(BeEquivalentTo(1722634733))
		})
	})

	Describe("Up", func() {
		var runInfo *models.DesiredLRPRunInfo

		BeforeEach(func() {
			migration.SetCryptor(cryptor)
			migration.SetDBFlavor(flavor)

			runInfo = &models.DesiredLRPRunInfo{}
		})

		Context("when there is desired lrp", func() {
			JustBeforeEach(func() {
				runInfoData, err := serializer.Marshal(logger, runInfo)
				Expect(err).NotTo(HaveOccurred())

				_, err = rawSQLDB.Exec(
					helpers.RebindForFlavor(
						`INSERT INTO desired_lrps
					  (process_guid, domain, log_guid, instances, memory_mb,
						  disk_mb, rootfs, routes, volume_placement, modification_tag_epoch, run_info)
					  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
						flavor,
					),
					"guid", "domain", "log guid", 2, 1, 1, "rootfs", "routes", "volumes yo", "1", runInfoData,
				)
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when run info has metric tags", func() {
				BeforeEach(func() {
					runInfo.MetricTags = models.MetricTags{"foo": &models.MetricTagValue{Static: "some-value"}}
				})

				It("updates metric_tags to run_info metric_tags", func() {
					testUpInTransaction(rawSQLDB, migration, logger)

					query := helpers.RebindForFlavor("select metric_tags from desired_lrps limit 1", flavor)
					row := rawSQLDB.QueryRow(query)
					var metricTagsData []byte
					Expect(row.Scan(&metricTagsData)).To(Succeed())

					encoder := format.NewEncoder(cryptor)
					decodedMetricTags, err := encoder.Decode(metricTagsData)
					Expect(err).NotTo(HaveOccurred())

					var metricTags models.MetricTags
					err = json.Unmarshal(decodedMetricTags, &metricTags)
					Expect(err).NotTo(HaveOccurred())

					Expect(metricTags).To(Equal(models.MetricTags{"foo": &models.MetricTagValue{Static: "some-value"}}))
				})
			})

			Context("when run info does not have metric tags", func() {
				It("updates metric_tags to be empty", func() {
					testUpInTransaction(rawSQLDB, migration, logger)

					query := helpers.RebindForFlavor("select metric_tags from desired_lrps limit 1", flavor)
					row := rawSQLDB.QueryRow(query)
					var metricTagsData []byte
					Expect(row.Scan(&metricTagsData)).To(Succeed())

					encoder := format.NewEncoder(cryptor)
					decodedMetricTags, err := encoder.Decode(metricTagsData)
					Expect(err).NotTo(HaveOccurred())

					var metricTags models.MetricTags
					err = json.Unmarshal(decodedMetricTags, &metricTags)
					Expect(err).NotTo(HaveOccurred())

					Expect(metricTags).To(BeEmpty())
				})
			})
		})

		It("is idempotent", func() {
			testIdempotency(rawSQLDB, migration, logger)
		})
	})
})
