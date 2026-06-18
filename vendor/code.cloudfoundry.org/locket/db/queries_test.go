package db_test

import (
	"code.cloudfoundry.org/diego-db-helpers/sqldb/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CreateHealthCheckTable", func() {
	It("creates the health check table successfully", func() {
		err := sqlDB.CreateHealthCheckTable(ctx, logger)
		Expect(err).NotTo(HaveOccurred())

		// Verify table exists by querying it
		var count int
		scanner := rawDB.QueryRowContext(ctx, helpers.RebindForFlavor("SELECT COUNT(*) FROM locket_health_check", dbFlavor))
		err = scanner.Scan(&count)
		Expect(err).NotTo(HaveOccurred())
	})

	It("is idempotent and can be called multiple times", func() {
		err := sqlDB.CreateHealthCheckTable(ctx, logger)
		Expect(err).NotTo(HaveOccurred())

		err = sqlDB.CreateHealthCheckTable(ctx, logger)
		Expect(err).NotTo(HaveOccurred())

		// Verify table still exists and is functional
		var count int
		scanner := rawDB.QueryRowContext(ctx, helpers.RebindForFlavor("SELECT COUNT(*) FROM locket_health_check", dbFlavor))
		err = scanner.Scan(&count)
		Expect(err).NotTo(HaveOccurred())
	})
})
