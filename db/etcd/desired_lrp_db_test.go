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
		var filter models.DesiredLRPFilter
		var desiredLRPsInDomains map[string][]*models.DesiredLRP

		Context("when there are desired LRPs", func() {
			var expectedDesiredLRPs []*models.DesiredLRP

			BeforeEach(func() {
				filter = models.DesiredLRPFilter{}
				expectedDesiredLRPs = []*models.DesiredLRP{}

				desiredLRPsInDomains = testHelper.CreateDesiredLRPsInDomains(map[string]int{
					"domain-1": 1,
					"domain-2": 2,
				})
			})

			It("returns all the desired LRPs", func() {
				for _, domainLRPs := range desiredLRPsInDomains {
					for _, lrp := range domainLRPs {
						expectedDesiredLRPs = append(expectedDesiredLRPs, lrp)
					}
				}
				desiredLRPs, err := etcdDB.DesiredLRPs(filter, logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(desiredLRPs.GetDesiredLrps()).To(ConsistOf(expectedDesiredLRPs))
			})

			It("can filter by domain", func() {
				for _, lrp := range desiredLRPsInDomains["domain-2"] {
					expectedDesiredLRPs = append(expectedDesiredLRPs, lrp)
				}
				filter.Domain = "domain-2"
				desiredLRPs, err := etcdDB.DesiredLRPs(filter, logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(desiredLRPs.GetDesiredLrps()).To(ConsistOf(expectedDesiredLRPs))
			})
		})

		Context("when there are no LRPs", func() {
			It("returns an empty list", func() {
				desiredLRPs, err := etcdDB.DesiredLRPs(filter, logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(desiredLRPs).NotTo(BeNil())
				Expect(desiredLRPs.GetDesiredLrps()).To(BeEmpty())
			})
		})

		Context("when there is invalid data", func() {
			BeforeEach(func() {
				testHelper.CreateValidDesiredLRP("some-guid")
				testHelper.CreateMalformedDesiredLRP("some-other-guid")
				testHelper.CreateValidDesiredLRP("some-third-guid")
			})

			It("errors", func() {
				_, err := etcdDB.DesiredLRPs(filter, logger)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when etcd is not there", func() {
			BeforeEach(func() {
				etcdRunner.Stop()
			})

			AfterEach(func() {
				etcdRunner.Start()
			})

			It("errors", func() {
				_, err := etcdDB.DesiredLRPs(filter, logger)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
