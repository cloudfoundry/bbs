package main_test

import (
	"time"

	etcddb "github.com/cloudfoundry-incubator/bbs/db/etcd"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Domain API", func() {
	Describe("PUT /v1/domains/:domain", func() {
		var existingDomain string

		BeforeEach(func() {
			existingDomain = "existing-domain"
			_, err := etcdClient.Set(etcddb.DomainSchemaPath(existingDomain), "", 100)
			Expect(err).NotTo(HaveOccurred())
		})

		It("updates the TTL when updating an existing domain", func() {
			err := client.UpsertDomain(existingDomain, 200*time.Second)
			Expect(err).ToNot(HaveOccurred())

			etcdEntry, err := etcdClient.Get(etcddb.DomainSchemaPath(existingDomain), false, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(etcdEntry.Node.TTL).To(BeNumerically(">", 100))
		})

		It("creates a domain with the desired TTL", func() {
			err := client.UpsertDomain("new-domain", 54*time.Second)
			Expect(err).ToNot(HaveOccurred())

			etcdEntry, err := etcdClient.Get(etcddb.DomainSchemaPath("new-domain"), false, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(etcdEntry.Node.TTL).To(BeNumerically("<=", 54))
		})
	})

	Describe("GET /v1/domains", func() {
		var expectedDomains []string
		var actualDomains []string
		var getErr error

		BeforeEach(func() {
			expectedDomains = []string{"domain-0", "domain-1"}
			for i, d := range expectedDomains {
				_, err := etcdClient.Set(etcddb.DomainSchemaPath(d), "", uint64(100*(i+1)))
				Expect(err).NotTo(HaveOccurred())
			}

			actualDomains, getErr = client.Domains()
		})

		It("responds without error", func() {
			Expect(getErr).NotTo(HaveOccurred())
		})

		It("has the correct number of responses", func() {
			Expect(actualDomains).To(HaveLen(2))
		})

		It("has the correct domains from the bbs", func() {
			Expect(expectedDomains).To(ConsistOf(actualDomains))
		})
	})
})
