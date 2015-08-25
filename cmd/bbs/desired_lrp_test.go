package main_test

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/models/test/model_helpers"

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
		desiredLRPs = etcdHelper.CreateDesiredLRPsInDomains(map[string]int{
			"domain-1": 2,
			"domain-2": 3,
		})
	})

	Describe("DesiredLRPs", func() {
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

	Describe("DesiredLRPByProcessGuid", func() {
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

	Describe("DesireLRP", func() {
		var (
			desiredLRP *models.DesiredLRP

			desireErr error
		)

		JustBeforeEach(func() {
			desiredLRP = model_helpers.NewValidDesiredLRP("super-lrp")
			desireErr = client.DesireLRP(desiredLRP)
		})

		It("creates the desired LRP in the system", func() {
			Expect(desireErr).NotTo(HaveOccurred())
			persistedDesiredLRP, err := client.DesiredLRPByProcessGuid("super-lrp")
			Expect(err).NotTo(HaveOccurred())
			persistedDesiredLRP.ModificationTag = nil
			Expect(persistedDesiredLRP).To(Equal(desiredLRP))
		})
	})

	Describe("RemoveDesiredLRP", func() {
		var (
			desiredLRP *models.DesiredLRP

			removeErr error
		)

		JustBeforeEach(func() {
			desiredLRP = model_helpers.NewValidDesiredLRP("super-lrp")
			err := client.DesireLRP(desiredLRP)
			Expect(err).NotTo(HaveOccurred())
			removeErr = client.RemoveDesiredLRP("super-lrp")
		})

		It("creates the desired LRP in the system", func() {
			Expect(removeErr).NotTo(HaveOccurred())
			_, err := client.DesiredLRPByProcessGuid("super-lrp")
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(models.ErrResourceNotFound))
		})
	})

	Describe("UpdateDesiredLRP", func() {
		var (
			desiredLRP *models.DesiredLRP

			updateErr error
		)

		JustBeforeEach(func() {
			desiredLRP = model_helpers.NewValidDesiredLRP("super-lrp")
			err := client.DesireLRP(desiredLRP)
			Expect(err).NotTo(HaveOccurred())
			three := int32(3)
			updateErr = client.UpdateDesiredLRP("super-lrp", &models.DesiredLRPUpdate{Instances: &three})
		})

		It("creates the desired LRP in the system", func() {
			Expect(updateErr).NotTo(HaveOccurred())
			persistedDesiredLRP, err := client.DesiredLRPByProcessGuid("super-lrp")
			Expect(err).NotTo(HaveOccurred())
			Expect(persistedDesiredLRP.Instances).To(Equal(int32(3)))
		})
	})
})
