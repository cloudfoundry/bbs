package migrations_test

import (
	"strings"

	"code.cloudfoundry.org/bbs/db/migrations"
	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/migration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Set Run Info LONGTEXT Column Migration", func() {
	var (
		migration migration.Migration
	)

	BeforeEach(func() {
		rawSQLDB.Exec("DROP TABLE domains;")
		rawSQLDB.Exec("DROP TABLE tasks;")
		rawSQLDB.Exec("DROP TABLE desired_lrps;")
		rawSQLDB.Exec("DROP TABLE actual_lrps;")

		migration = migrations.NewSetRunInfoLongtext()
	})

	It("appends itself to the migration list", func() {
		Expect(migrations.AllMigrations()).To(ContainElement(migration))
	})

	Describe("Version", func() {
		It("returns the timestamp from which it was created", func() {
			Expect(migration.Version()).To(BeEquivalentTo(1674146125))
		})
	})

	Describe("Up from MEDIUMTEXT", func() {
		BeforeEach(func() {
			// Can't do this in the Describe BeforeEach
			// as the test on line 37 will cause ginkgo to panic
			migration.SetRawSQLDB(rawSQLDB)
			migration.SetDBFlavor(flavor)
		})

		BeforeEach(func() {
			createStatements := []string{
				`CREATE TABLE actual_lrps(
	net_info MEDIUMTEXT NOT NULL
);`,
				`CREATE TABLE tasks(
	result MEDIUMTEXT,
	task_definition MEDIUMTEXT NOT NULL
);`,

				`CREATE TABLE desired_lrps(
	annotation MEDIUMTEXT,
	routes MEDIUMTEXT NOT NULL,
	volume_placement MEDIUMTEXT NOT NULL,
	run_info MEDIUMTEXT NOT NULL
);`,
			}
			for _, st := range createStatements {
				_, err := rawSQLDB.Exec(st)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("should change the size of all text columns ", func() {
			Expect(migration.Up(logger)).To(Succeed())
			value := strings.Repeat("x", 16777215*2)
			query := helpers.RebindForFlavor("insert into desired_lrps(annotation, routes, volume_placement, run_info) values('', '', '', ?)", flavor)
			_, err := rawSQLDB.Exec(query, value)
			Expect(err).NotTo(HaveOccurred())
		})

		It("is idempotent", func() {
			testIdempotency(rawSQLDB, migration, logger)
		})
	})

	Describe("Up from LONGTEXT", func() {
		BeforeEach(func() {
			// Can't do this in the Describe BeforeEach
			// as the test on line 37 will cause ginkgo to panic
			migration.SetRawSQLDB(rawSQLDB)
			migration.SetDBFlavor(flavor)
		})

		BeforeEach(func() {
			createStatements := []string{
				`CREATE TABLE actual_lrps(
	net_info LONGTEXT NOT NULL
);`,
				`CREATE TABLE tasks(
	result LONGTEXT,
	task_definition LONGTEXT NOT NULL
);`,

				`CREATE TABLE desired_lrps(
	annotation LONGTEXT,
	routes LONGTEXT NOT NULL,
	volume_placement LONGTEXT NOT NULL,
	run_info LONGTEXT NOT NULL
);`,
			}
			for _, st := range createStatements {
				_, err := rawSQLDB.Exec(st)
				Expect(err).NotTo(HaveOccurred())
			}

			value := strings.Repeat("x", 16777215*2)
			query := helpers.RebindForFlavor("insert into desired_lrps(annotation, routes, volume_placement, run_info) values('', '', '', ?)", flavor)
			_, err := rawSQLDB.Exec(query, value)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should change the size of all text columns ", func() {
			Expect(migration.Up(logger)).To(Succeed())
			value := strings.Repeat("x", 16777215*2)
			rows, err := rawSQLDB.Query("select run_info from desired_lrps;")
			rows.Next()
			var runInfo string
			rows.Scan(&runInfo)

			Expect(runInfo).To(Equal(value))
			Expect(err).NotTo(HaveOccurred())
		})

		It("is idempotent", func() {
			testIdempotency(rawSQLDB, migration, logger)
		})
	})
})
