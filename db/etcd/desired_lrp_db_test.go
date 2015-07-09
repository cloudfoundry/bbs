package db_test

import (
	"github.com/cloudfoundry-incubator/bbs/db"
	. "github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DesiredLRPDB", func() {
	var (
		etcdDB db.DesiredLRPDB
	)

	BeforeEach(func() {
		etcdDB = NewETCD(etcdClient)
	})

	Describe("DesiredLRPs", func() {
		Context("when there are desired LRPs", func() {
			var expectedDesiredLRPs []*models.DesiredLRP
			BeforeEach(func() {
				desiredLRPsInDomains := testHelper.CreateDesiredLRPsInDomains(map[string]int{
					"domain-1": 1,
					"domain-2": 2,
				})
				for _, domainLRPs := range desiredLRPsInDomains {
					for _, lrp := range domainLRPs {
						expectedDesiredLRPs = append(expectedDesiredLRPs, lrp)
					}
				}
			})

			It("returns all the desired LRPs", func() {
				desiredLRPGroups, err := etcdDB.DesiredLRPs(logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(desiredLRPGroups.GetDesiredLrps()).To(ConsistOf(expectedDesiredLRPs))
			})
		})

		Context("when there are no LRPs", func() {
			It("returns an empty list", func() {
				desiredLRPGroups, err := etcdDB.DesiredLRPs(logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(desiredLRPGroups).NotTo(BeNil())
				Expect(desiredLRPGroups.GetDesiredLrps()).To(BeEmpty())
			})
		})

		Context("when there is invalid data", func() {
			BeforeEach(func() {
				testHelper.CreateValidDesiredLRP("some-guid")
				testHelper.CreateMalformedDesiredLRP("some-other-guid")
				testHelper.CreateValidDesiredLRP("some-third-guid")
			})

			It("errors", func() {
				_, err := etcdDB.DesiredLRPs(logger)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
