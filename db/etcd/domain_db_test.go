package db_test

import (
	"github.com/cloudfoundry-incubator/bbs/db"
	. "github.com/cloudfoundry-incubator/bbs/db/etcd"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DomainDB", func() {
	var db db.DomainDB

	BeforeEach(func() {
		db = NewETCD(etcdClient)
	})

	Describe("UpsertDomain", func() {
		Context("when the domain is not present in the DB", func() {
			It("inserts a new domain with the requested TTL", func() {
				domain := "my-awesome-domain"
				err := db.UpsertDomain(domain, 5432)
				Expect(err).NotTo(HaveOccurred())

				etcdEntry, err := etcdClient.Get(DomainSchemaPath(domain), false, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(etcdEntry.Node.TTL).To(BeNumerically("<=", 5432))
			})
		})

		Context("when the domain is already present in the DB", func() {
			var existingDomain = "the-domain-that-was-already-there"

			BeforeEach(func() {
				var err error
				_, err = etcdClient.Set(DomainSchemaPath(existingDomain), "", 100)
				Expect(err).NotTo(HaveOccurred())
			})

			It("updates the TTL on the existing record", func() {
				err := db.UpsertDomain(existingDomain, 1337)
				Expect(err).NotTo(HaveOccurred())

				etcdEntry, err := etcdClient.Get(DomainSchemaPath(existingDomain), false, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(etcdEntry.Node.TTL).To(BeNumerically("<=", 1337))
				Expect(etcdEntry.Node.TTL).To(BeNumerically(">", 100))
			})
		})
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
