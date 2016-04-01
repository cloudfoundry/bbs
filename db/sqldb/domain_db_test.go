package sqldb_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DomainDB", func() {
	Describe("Domains", func() {
		Context("when there are domains in the DB", func() {
			BeforeEach(func() {
				futureTime := fakeClock.Now().Add(5 * time.Second).UnixNano()
				_, err := db.Exec("INSERT INTO domains VALUES (?, ?)", "jims-domain", futureTime)
				Expect(err).NotTo(HaveOccurred())

				_, err = db.Exec("INSERT INTO domains VALUES (?, ?)", "amelias-domain", futureTime)
				Expect(err).NotTo(HaveOccurred())

				pastTime := fakeClock.Now().Add(-5 * time.Second).UnixNano()
				_, err = db.Exec("INSERT INTO domains VALUES (?, ?)", "past-domain", pastTime)
				Expect(err).NotTo(HaveOccurred())

				_, err = db.Exec("INSERT INTO domains VALUES (?, ?)", "current-domain", fakeClock.Now().Round(time.Second).UnixNano())
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns all the non-expired domains in the DB", func() {
				domains, err := sqlDB.Domains(logger)
				Expect(err).NotTo(HaveOccurred())

				Expect(domains).To(HaveLen(2))
				Expect(domains).To(ConsistOf([]string{"jims-domain", "amelias-domain"}))
			})
		})

		Context("when there are no domains in the DB", func() {
			It("returns no domains", func() {
				domains, err := sqlDB.Domains(logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(domains).To(HaveLen(0))
			})
		})
	})

	Describe("UpsertDomain", func() {
		Context("when the domain is not present in the DB", func() {
			It("inserts a new domain with the requested TTL", func() {
				domain := "my-awesome-domain"

				bbsErr := sqlDB.UpsertDomain(logger, domain, 5432)
				Expect(bbsErr).NotTo(HaveOccurred())

				rows, err := db.Query("SELECT * FROM domains;")
				Expect(err).NotTo(HaveOccurred())
				defer rows.Close()

				var domainName string
				var expireTime int64

				Expect(rows.Next()).To(BeTrue())
				err = rows.Scan(&domainName, &expireTime)
				Expect(err).NotTo(HaveOccurred())
				Expect(domainName).To(Equal(domain))
				expectedExpireTime := fakeClock.Now().UTC().Add(time.Duration(5432) * time.Second).UnixNano()
				Expect(expireTime).To(BeEquivalentTo(expectedExpireTime))
			})

			Context("when the domain is too long", func() {
				It("returns an error", func() {
					domain := randStr(256)
					bbsErr := sqlDB.UpsertDomain(logger, domain, 5432)
					Expect(bbsErr).To(HaveOccurred())
				})
			})
		})

		Context("when the domain is already present in the DB", func() {
			var existingDomain = "the-domain-that-was-already-there"

			BeforeEach(func() {
				bbsErr := sqlDB.UpsertDomain(logger, existingDomain, 1)
				Expect(bbsErr).NotTo(HaveOccurred())
			})

			It("updates the TTL on the existing record", func() {
				fakeClock.Increment(10 * time.Second)

				bbsErr := sqlDB.UpsertDomain(logger, existingDomain, 1)
				Expect(bbsErr).NotTo(HaveOccurred())

				rowsCount, err := db.Query("SELECT COUNT(*) FROM domains;")
				Expect(err).NotTo(HaveOccurred())
				Expect(rowsCount.Next()).To(BeTrue())
				var domainCount int
				err = rowsCount.Scan(&domainCount)
				Expect(err).NotTo(HaveOccurred())
				Expect(domainCount).To(Equal(1))

				rows, err := db.Query("SELECT * FROM domains;")
				Expect(err).NotTo(HaveOccurred())
				defer rows.Close()

				var domainName string
				var expireTime int64

				Expect(rows.Next()).To(BeTrue())
				err = rows.Scan(&domainName, &expireTime)
				Expect(err).NotTo(HaveOccurred())
				Expect(domainName).To(Equal(existingDomain))
				expectedExpireTime := fakeClock.Now().UTC().Add(time.Duration(1) * time.Second).UnixNano()
				Expect(expireTime).To(BeEquivalentTo(expectedExpireTime))
			})
		})
	})
})
