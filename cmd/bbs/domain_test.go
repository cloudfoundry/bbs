package main_test

import (
	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Domain API", func() {
	BeforeEach(func() {
		bbsProcess = ginkgomon.Invoke(bbsRunner)
	})

	AfterEach(func() {
		ginkgomon.Kill(bbsProcess)
	})

	Describe("GET /v1/domains", func() {
		var expectedDomains []string
		var actualDomains []string
		var getErr error

		BeforeEach(func() {
			expectedDomains = []string{"domain-0", "domain-1"}
			for i, d := range expectedDomains {
				_, err := etcdClient.Set(db.DomainSchemaPath(d), "", uint64(100*(i+1)))
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
