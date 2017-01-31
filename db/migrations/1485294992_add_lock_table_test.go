package migrations_test

import (
	"code.cloudfoundry.org/bbs/db/migrations"
	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/migration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Add Lock Table", func() {
	var (
		mig    migration.Migration
		migErr error
	)

	BeforeEach(func() {
		rawSQLDB.Exec("DROP TABLE locks;")
		mig = migrations.NewAddLockTable()
	})

	It("appends itself to the migration list", func() {
		Expect(migrations.Migrations).To(ContainElement(mig))
	})

	Describe("Version", func() {
		It("returns the timestamp from which it was created", func() {
			Expect(mig.Version()).To(BeEquivalentTo(1485294992))
		})
	})

	Describe("Up", func() {
		BeforeEach(func() {
			// Can't do this in the Describe BeforeEach
			// as the migration list-appending test will panic
			mig.SetRawSQLDB(rawSQLDB)
			mig.SetDBFlavor(flavor)
		})

		JustBeforeEach(func() {
			migErr = mig.Up(logger)
		})

		It("does not error out", func() {
			Expect(migErr).NotTo(HaveOccurred())
		})

		It("creates a locks table", func() {
			query := helpers.RebindForFlavor(`SELECT table_name FROM information_schema.tables WHERE table_name = ?`, flavor)
			row := rawSQLDB.QueryRow(query, "locks")

			var tableName string
			err := row.Scan(&tableName)
			Expect(err).NotTo(HaveOccurred())
			Expect(tableName).To(Equal("locks"))
		})

		Context("when the locks table already exists", func() {
			BeforeEach(func() {
				_, err := rawSQLDB.Exec(`CREATE TABLE locks (
					key VARCHAR(255) PRIMARY KEY,
					foobar VARCHAR(255),
					value VARCHAR(255)
				);`)

				Expect(err).NotTo(HaveOccurred())
			})

			It("drops the existing table and creates a new one", func() {
				Expect(migErr).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Down", func() {
		It("returns a not implemented error", func() {
			Expect(mig.Down(logger)).To(HaveOccurred())
		})
	})
})
