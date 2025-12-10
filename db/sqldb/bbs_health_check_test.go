package sqldb_test

import (
	"time"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/test_helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("BBSHealthCheckDB", func() {
	var flavor string
	BeforeEach(func() {
		if test_helpers.UseMySQL() {
			flavor = "mysql"
		} else {
			flavor = "postgres"
		}
	})
	Context("when no previous healthcheck time exists", func() {
		BeforeEach(func() {
			db.ExecContext(ctx, "DELETE FROM bbs_health_check")
		})
		It("adds it", func() {
			now := time.Now()
			err := sqlDB.PerformBBSHealthCheck(ctx, logger, now)
			Expect(err).NotTo(HaveOccurred())

			scanner := db.QueryRowContext(ctx, helpers.RebindForFlavor("SELECT * FROM bbs_health_check WHERE id = ?", flavor), 1)
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
			db.ExecContext(ctx, "DELETE FROM bbs_health_check")
			_, err := db.ExecContext(ctx, helpers.RebindForFlavor("INSERT INTO bbs_health_check (id, time) VALUES(1, ?)", flavor), time.Time{}.UnixNano())
			Expect(err).ToNot(HaveOccurred())
		})
		It("updates it", func() {
			now := time.Now()
			err := sqlDB.PerformBBSHealthCheck(ctx, logger, now)
			Expect(err).NotTo(HaveOccurred())

			scanner := db.QueryRowContext(ctx, helpers.RebindForFlavor("SELECT * FROM bbs_health_check WHERE id = ?", flavor), 1)
			var i int
			var t int64
			err = scanner.Scan(&i, &t)
			Expect(err).ToNot(HaveOccurred())
			Expect(i).To(Equal(1))
			Expect(t).To(Equal(now.UnixNano()))
		})
	})
})
