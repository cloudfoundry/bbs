package db_test

import (
	"time"

	"code.cloudfoundry.org/diego-db-helpers/sqldb/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("LocketHealthCheckDB", func() {
	Context("when no previous healthcheck time exists", func() {
		BeforeEach(func() {
			rawDB.ExecContext(ctx, "DELETE FROM locket_health_check")
		})
		It("adds it", func() {
			now := time.Now()
			err := sqlDB.PerformLocketHealthCheck(ctx, logger, now)
			Expect(err).NotTo(HaveOccurred())

			scanner := rawDB.QueryRowContext(ctx, helpers.RebindForFlavor("SELECT * FROM locket_health_check WHERE id = ?", dbFlavor), 1)
			var i int
			var t int64
			err = scanner.Scan(&i, &t)
			Expect(err).ToNot(HaveOccurred())
			Expect(i).To(Equal(1))
			Expect(t).To(Equal(now.UnixNano()))

		})
	})
	Context("when a previous healthcheck time exists", func() {
		BeforeEach(func() {
			rawDB.ExecContext(ctx, "DELETE FROM locket_health_check")
			_, err := rawDB.ExecContext(ctx, helpers.RebindForFlavor("INSERT INTO locket_health_check (id, time) VALUES(1, ?)", dbFlavor), time.Time{}.UnixNano())
			Expect(err).ToNot(HaveOccurred())
		})
		It("updates it", func() {
			now := time.Now()
			err := sqlDB.PerformLocketHealthCheck(ctx, logger, now)
			Expect(err).NotTo(HaveOccurred())

			scanner := rawDB.QueryRowContext(ctx, helpers.RebindForFlavor("SELECT * FROM locket_health_check WHERE id = ?", dbFlavor), 1)
			var i int
			var t int64
			err = scanner.Scan(&i, &t)
			Expect(err).ToNot(HaveOccurred())
			Expect(i).To(Equal(1))
			Expect(t).To(Equal(now.UnixNano()))
		})
	})
})
