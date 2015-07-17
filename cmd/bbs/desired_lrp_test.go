package main_test

import (
	"github.com/cloudfoundry-incubator/bbs/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DesiredLRP API", func() {
	var (
		desiredLRPs         map[string][]*models.DesiredLRP
		expectedDesiredLRPs []*models.DesiredLRP
		actualDesiredLRPs   []*models.DesiredLRP

		filter models.DesiredLRPFilter

		getErr error
	)

	BeforeEach(func() {
		filter = models.DesiredLRPFilter{}
		expectedDesiredLRPs = []*models.DesiredLRP{}
		actualDesiredLRPs = []*models.DesiredLRP{}
		desiredLRPs = testHelper.CreateDesiredLRPsInDomains(map[string]int{
			"domain-1": 2,
			"domain-2": 3,
		})
	})

	Describe("GET /v1/desired_lrps", func() {
		JustBeforeEach(func() {
			actualDesiredLRPs, getErr = client.DesiredLRPs(filter)
		})

		It("responds without error", func() {
			Expect(getErr).NotTo(HaveOccurred())
		})

		It("has the correct number of responses", func() {
			Expect(actualDesiredLRPs).To(HaveLen(5))
		})

		Context("when not filtering", func() {
			It("returns all desired lrps from the bbs", func() {
				for _, domainLRPs := range desiredLRPs {
					for _, lrp := range domainLRPs {
						expectedDesiredLRPs = append(expectedDesiredLRPs, lrp)
					}
				}
				Expect(actualDesiredLRPs).To(ConsistOf(expectedDesiredLRPs))
			})
		})

		Context("when filtering by domain", func() {
			var domain string
			BeforeEach(func() {
				domain = "domain-1"
				filter = models.DesiredLRPFilter{Domain: domain}
			})

			It("has the correct number of responses", func() {
				Expect(actualDesiredLRPs).To(HaveLen(2))
			})

			It("returns only the desired lrps in the requested domain", func() {
				for _, lrp := range desiredLRPs[domain] {
					expectedDesiredLRPs = append(expectedDesiredLRPs, lrp)
				}
				Expect(actualDesiredLRPs).To(ConsistOf(expectedDesiredLRPs))
			})
		})
	})

	Describe("GET /v1/desired_lrps/:process_guid", func() {
		var (
			desiredLRP         *models.DesiredLRP
			expectedDesiredLRP *models.DesiredLRP
		)

		JustBeforeEach(func() {
			expectedDesiredLRP = desiredLRPs["domain-1"][0]
			desiredLRP, getErr = client.DesiredLRPByProcessGuid(expectedDesiredLRP.GetProcessGuid())
		})

		It("responds without error", func() {
			Expect(getErr).NotTo(HaveOccurred())
		})

		It("returns all desired lrps from the bbs", func() {
			Expect(desiredLRP).To(Equal(expectedDesiredLRP))
		})
	})
})
