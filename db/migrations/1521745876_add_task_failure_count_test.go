package migrations_test

import (
	"code.cloudfoundry.org/bbs/db/migrations"
	"code.cloudfoundry.org/bbs/migration"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AddTaskFailureCount", func() {
	var (
		migration migration.Migration
	)

	BeforeEach(func() {
		// TODO: db cleanup here

		migration = migrations.NewAddTaskFailureCount()
	})

	It("appends itself to the migration list", func() {
		Expect(migrations.AllMigrations()).To(ContainElement(migration))
	})

	Describe("Version", func() {
		It("returns the timestamp from which it was created", func() {
			Expect(migration.Version()).To(BeEquivalentTo(1521745876))
		})
	})

	Describe("Up", func() {
		BeforeEach(func() {
			_, err := rawSQLDB.Exec(`
CREATE TABLE tasks(
	task_definition TEXT NOT NULL
);`)
			Expect(err).NotTo(HaveOccurred())

			migration.SetRawSQLDB(rawSQLDB)
			migration.SetDBFlavor(flavor)
		})

		It("is idempotent", func() {
			testIdempotency(rawSQLDB, migration, logger)
		})
	})
})
