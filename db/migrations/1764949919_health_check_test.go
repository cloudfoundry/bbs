package migrations_test

import (
	"time"

	"code.cloudfoundry.org/bbs/db/migrations"
	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/migration"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("BBSHealthCheckTable", func() {
	var (
		migration migration.Migration
	)

	BeforeEach(func() {
		rawSQLDB.Exec("DROP TABLE bbs_health_check;")

		initialMigration := migrations.NewInitSQL()
		initialMigration.SetDBFlavor(flavor)
		initialMigration.SetClock(fakeClock)
		testUpInTransaction(rawSQLDB, initialMigration, logger)

		migration = migrations.NewBBSHealthCheckTable()
	})

	It("appends itself to the migration list", func() {
		Expect(migrations.AllMigrations()).To(ContainElement(migration))
	})

	Describe("Version", func() {
		It("returns the timestamp from which it was created", func() {
			Expect(migration.Version()).To(BeEquivalentTo(1764949919))
		})
	})

	Describe("Up", func() {
		BeforeEach(func() {
			migration.SetCryptor(cryptor)
			migration.SetDBFlavor(flavor)
		})

		It("adds the table", func() {
			testUpInTransaction(rawSQLDB, migration, logger)
			t := time.Now().UnixNano()
			insertSQL := "INSERT INTO bbs_health_check (id, time) VALUES (?, ?)"
			_, err := rawSQLDB.Exec(helpers.RebindForFlavor(insertSQL, flavor), 1, t)
			Expect(err).ToNot(HaveOccurred())

			querySQL := "SELECT time FROM bbs_health_check WHERE id = ?"
			row := rawSQLDB.QueryRow(helpers.RebindForFlavor(querySQL, flavor), 1)
			var newT int64
			Expect(row.Scan(&newT)).To(Succeed())

			Expect(newT).To(Equal(t))
		})

		It("is idempotent", func() {
			testIdempotency(rawSQLDB, migration, logger)
		})
	})
})
