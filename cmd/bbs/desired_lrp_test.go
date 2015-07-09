package main_test

import (
	"github.com/cloudfoundry-incubator/bbs/db/etcd/internal/test_helpers"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DesiredLRP API", func() {
	var testHelper *test_helpers.TestHelper

	BeforeEach(func() {
		bbsProcess = ginkgomon.Invoke(bbsRunner)
		testHelper = test_helpers.NewTestHelper(etcdClient)
	})

	AfterEach(func() {
		ginkgomon.Kill(bbsProcess)
	})

	var (
		desiredLRPs         map[string][]*models.DesiredLRP
		expectedDesiredLRPs []*models.DesiredLRP
		actualDesiredLRPs   []*models.DesiredLRP

		getErr error
	)

	BeforeEach(func() {
		desiredLRPs = testHelper.CreateDesiredLRPsInDomains(map[string]int{
			"domain-1": 2,
			"domain-2": 3,
		})
	})

	Describe("GET /v1/desired_lrps", func() {
		JustBeforeEach(func() {
			actualDesiredLRPs, getErr = client.DesiredLRPs()
		})

		It("responds without error", func() {
			Expect(getErr).NotTo(HaveOccurred())
		})

		It("has the correct number of responses", func() {
			Expect(actualDesiredLRPs).To(HaveLen(5))
		})

		It("returns all desired lrps from the bbs", func() {
			for _, domainLRPs := range desiredLRPs {
				for _, lrp := range domainLRPs {
					expectedDesiredLRPs = append(expectedDesiredLRPs, lrp)
				}
			}
			Expect(actualDesiredLRPs).To(ConsistOf(expectedDesiredLRPs))
		})
	})
})
