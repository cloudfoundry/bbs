package db_test

import (
	. "github.com/cloudfoundry-incubator/bbs/db"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DomainDB", func() {
	var db DomainDB

	BeforeEach(func() {
		db = NewETCD(etcdClient)
	})

	Describe("GetAllDomains", func() {
		Context("when there are domains in the DB", func() {
			BeforeEach(func() {
				var err error
				_, err = etcdClient.Set(DomainSchemaPath("domain-1"), "", 100)
				Expect(err).NotTo(HaveOccurred())
				_, err = etcdClient.Set(DomainSchemaPath("domain-2"), "", 100)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns all the existing domains in the DB", func() {
				domains, err := db.GetAllDomains()
				Expect(err).NotTo(HaveOccurred())

				Expect(domains).To(HaveLen(2))
				Expect(domains).To(ConsistOf([]string{"domain-1", "domain-2"}))
			})
		})

		Context("when there are no domains in the DB", func() {
			It("returns no domains", func() {
				domains, err := db.GetAllDomains()
				Expect(err).NotTo(HaveOccurred())
				Expect(domains).To(HaveLen(0))
			})
		})
	})
})
