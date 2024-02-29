package main_test

import (
	"fmt"
	"time"

	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/bbs/cmd/bbs/testrunner"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/models/test/model_helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ginkgomon "github.com/tedsuo/ifrit/ginkgomon_v2"
)

var _ = Describe("DesiredLRP API", func() {
	var (
		desiredLRPs            map[string][]*models.DesiredLRP
		schedulingInfos        []*models.DesiredLRPSchedulingInfo
		routingInfos           []*models.DesiredLRP
		expectedSchedulingInfo models.DesiredLRPSchedulingInfo
		expectedDesiredLRPs    []*models.DesiredLRP
		actualDesiredLRPs      []*models.DesiredLRP

		filter models.DesiredLRPFilter

		getErr error
	)

	BeforeEach(func() {
		bbsRunner = testrunner.New(bbsBinPath, bbsConfig)
		bbsProcess = ginkgomon.Invoke(bbsRunner)
		filter = models.DesiredLRPFilter{}
		expectedDesiredLRPs = []*models.DesiredLRP{}
		actualDesiredLRPs = []*models.DesiredLRP{}
		desiredLRPs = createDesiredLRPsInDomains(client, map[string]int{
			"domain-1": 2,
			"domain-2": 3,
		})
	})

	Describe("DesiredLRPs", func() {
		JustBeforeEach(func() {
			actualDesiredLRPs, getErr = client.DesiredLRPs(logger, "some-trace-id", filter)
			for _, lrp := range actualDesiredLRPs {
				lrp.ModificationTag.Epoch = "epoch"
			}
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

		Context("when filtering by process guids", func() {
			BeforeEach(func() {
				guids := []string{
					"guid-1-for-domain-1",
					"guid-2-for-domain-2",
				}

				filter = models.DesiredLRPFilter{ProcessGuids: guids}
			})

			It("returns only the scheduling infos in the requested process guids", func() {
				Expect(actualDesiredLRPs).To(HaveLen(2))

				expectedDesiredLRPs := []*models.DesiredLRP{
					desiredLRPs["domain-1"][1],
					desiredLRPs["domain-2"][2],
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
			desiredLRP, getErr = client.DesiredLRPByProcessGuid(logger, "some-trace-id", expectedDesiredLRP.GetProcessGuid())
			desiredLRP.ModificationTag.Epoch = "epoch"
		})

		It("responds without error", func() {
			Expect(getErr).NotTo(HaveOccurred())
		})

		It("returns all desired lrps from the bbs", func() {
			Expect(desiredLRP).To(Equal(expectedDesiredLRP))
		})
	})

	Describe("DesiredLRPSchedulingInfos", func() {
		JustBeforeEach(func() {
			schedulingInfos, getErr = client.DesiredLRPSchedulingInfos(logger, "some-trace-id", filter)
			for _, schedulingInfo := range schedulingInfos {
				schedulingInfo.ModificationTag.Epoch = "epoch"
			}
		})

		It("responds without error", func() {
			Expect(getErr).NotTo(HaveOccurred())
		})

		It("has the correct number of responses", func() {
			Expect(schedulingInfos).To(HaveLen(5))
		})

		Context("when not filtering", func() {
			It("returns all scheduling infos from the bbs", func() {
				expectedSchedulingInfos := []*models.DesiredLRPSchedulingInfo{}
				for _, domainLRPs := range desiredLRPs {
					for _, lrp := range domainLRPs {
						schedulingInfo := lrp.DesiredLRPSchedulingInfo()
						expectedSchedulingInfos = append(expectedSchedulingInfos, &schedulingInfo)
					}
				}
				Expect(schedulingInfos).To(ConsistOf(expectedSchedulingInfos))
			})
		})

		Context("when filtering by domain", func() {
			var domain string
			BeforeEach(func() {
				domain = "domain-1"
				filter = models.DesiredLRPFilter{Domain: domain}
			})

			It("has the correct number of responses", func() {
				Expect(schedulingInfos).To(HaveLen(2))
			})

			It("returns only the scheduling infos in the requested domain", func() {
				expectedSchedulingInfos := []*models.DesiredLRPSchedulingInfo{}
				for _, lrp := range desiredLRPs[domain] {
					schedulingInfo := lrp.DesiredLRPSchedulingInfo()
					expectedSchedulingInfos = append(expectedSchedulingInfos, &schedulingInfo)
				}
				Expect(schedulingInfos).To(ConsistOf(expectedSchedulingInfos))
			})
		})

		Context("when filtering by process guids", func() {
			BeforeEach(func() {
				guids := []string{
					"guid-1-for-domain-1",
					"guid-2-for-domain-2",
				}

				filter = models.DesiredLRPFilter{ProcessGuids: guids}
			})

			It("returns only the scheduling infos in the requested process guids", func() {
				Expect(schedulingInfos).To(HaveLen(2))

				desiredLRP1 := desiredLRPs["domain-1"][1].DesiredLRPSchedulingInfo()
				desiredLRP2 := desiredLRPs["domain-2"][2].DesiredLRPSchedulingInfo()
				expectedSchedulingInfos := []*models.DesiredLRPSchedulingInfo{
					&desiredLRP1,
					&desiredLRP2,
				}
				Expect(schedulingInfos).To(ConsistOf(expectedSchedulingInfos))
			})
		})
	})

	Describe("DesiredLRPSchedulingInfoByProcessGuid", func() {
		var schedulingInfoByProcessGuid *models.DesiredLRPSchedulingInfo

		JustBeforeEach(func() {
			expectedSchedulingInfo = desiredLRPs["domain-1"][0].DesiredLRPSchedulingInfo()
			schedulingInfoByProcessGuid, getErr = client.DesiredLRPSchedulingInfoByProcessGuid(logger, "some-trace-id", expectedSchedulingInfo.GetProcessGuid())
			schedulingInfoByProcessGuid.ModificationTag.Epoch = "epoch"
		})

		It("responds without error", func() {
			Expect(getErr).ToNot(HaveOccurred())
		})

		It("returns the correct desired lrp scheduling info", func() {
			Expect(*schedulingInfoByProcessGuid).To(Equal(expectedSchedulingInfo))
		})
	})

	Describe("DesiredLRPRoutingInfos", func() {
		JustBeforeEach(func() {
			routingInfos, getErr = client.DesiredLRPRoutingInfos(logger, "some-trace-id", filter)
			for _, routingInfo := range routingInfos {
				routingInfo.ModificationTag.Epoch = "epoch"
			}
		})

		It("responds without error", func() {
			Expect(getErr).NotTo(HaveOccurred())
		})

		It("has the correct number of responses", func() {
			Expect(routingInfos).To(HaveLen(5))
		})

		Context("when not filtering", func() {
			It("returns all routing infos from the bbs", func() {
				expectedRoutingInfos := []*models.DesiredLRP{}
				for _, domainLRPs := range desiredLRPs {
					for _, lrp := range domainLRPs {
						routingInfo := lrp.DesiredLRPRoutingInfo()
						expectedRoutingInfos = append(expectedRoutingInfos, &routingInfo)
					}
				}
				Expect(routingInfos).To(ConsistOf(expectedRoutingInfos))
			})
		})

		Context("when filtering by process guids", func() {
			BeforeEach(func() {
				guids := []string{
					"guid-1-for-domain-1",
					"guid-2-for-domain-2",
				}

				filter = models.DesiredLRPFilter{ProcessGuids: guids}
			})

			It("returns only the routing infos in the requested process guids", func() {
				Expect(routingInfos).To(HaveLen(2))

				routingInfo1 := desiredLRPs["domain-1"][1].DesiredLRPRoutingInfo()
				routingInfo2 := desiredLRPs["domain-2"][2].DesiredLRPRoutingInfo()
				expectedRoutingInfos := []*models.DesiredLRP{
					&routingInfo1,
					&routingInfo2,
				}
				Expect(routingInfos).To(ConsistOf(expectedRoutingInfos))
			})
		})
	})

	Describe("DesireLRP", func() {
		var (
			desiredLRP         *models.DesiredLRP
			expectedDesiredLRP *models.DesiredLRP
			desireErr          error
		)

		BeforeEach(func() {
			desiredLRP = model_helpers.NewValidDesiredLRP("super-lrp")
			expectedDesiredLRP = desiredLRP
		})

		JustBeforeEach(func() {
			desireErr = client.DesireLRP(logger, "some-trace-id", desiredLRP)
		})

		It("creates the desired LRP in the system", func() {
			Expect(desireErr).NotTo(HaveOccurred())
			persistedDesiredLRP, err := client.DesiredLRPByProcessGuid(logger, "some-trace-id", "super-lrp")
			Expect(err).NotTo(HaveOccurred())
			Expect(persistedDesiredLRP.DesiredLRPKey()).To(Equal(expectedDesiredLRP.DesiredLRPKey()))
			Expect(persistedDesiredLRP.DesiredLRPResource()).To(Equal(expectedDesiredLRP.DesiredLRPResource()))
			Expect(persistedDesiredLRP.Annotation).To(Equal(expectedDesiredLRP.Annotation))
			Expect(persistedDesiredLRP.Instances).To(Equal(expectedDesiredLRP.Instances))
			Expect(persistedDesiredLRP.DesiredLRPRunInfo(time.Unix(42, 0))).To(Equal(expectedDesiredLRP.DesiredLRPRunInfo(time.Unix(42, 0))))
			Expect(persistedDesiredLRP.Action.RunAction.SuppressLogOutput).To(BeFalse())
			Expect(persistedDesiredLRP.CertificateProperties).NotTo(BeNil())
			Expect(persistedDesiredLRP.CertificateProperties.OrganizationalUnit).NotTo(BeEmpty())
			Expect(persistedDesiredLRP.CertificateProperties.OrganizationalUnit).To(Equal(expectedDesiredLRP.CertificateProperties.OrganizationalUnit))
			Expect(persistedDesiredLRP.ImageUsername).To(Equal(expectedDesiredLRP.ImageUsername))
			Expect(persistedDesiredLRP.ImagePassword).To(Equal(expectedDesiredLRP.ImagePassword))
		})

		Context("when suppressing log output", func() {
			BeforeEach(func() {
				desiredLRP.Action.RunAction.SuppressLogOutput = true
			})

			It("has an action with SuppressLogOutput set to true", func() {
				Expect(desireErr).NotTo(HaveOccurred())
				persistedDesiredLRP, err := client.DesiredLRPByProcessGuid(logger, "some-trace-id", "super-lrp")
				Expect(err).NotTo(HaveOccurred())
				Expect(persistedDesiredLRP.Action.RunAction.SuppressLogOutput).To(BeTrue())
			})
		})

		Context("when not suppressing log output", func() {
			BeforeEach(func() {
				desiredLRP.Action.RunAction.SuppressLogOutput = false
			})

			It("has an action with SuppressLogOutput set to false", func() {
				Expect(desireErr).NotTo(HaveOccurred())
				persistedDesiredLRP, err := client.DesiredLRPByProcessGuid(logger, "some-trace-id", "super-lrp")
				Expect(err).NotTo(HaveOccurred())
				Expect(persistedDesiredLRP.Action.RunAction.SuppressLogOutput).To(BeFalse())
			})
		})
	})

	Describe("RemoveDesiredLRP", func() {
		var (
			desiredLRP *models.DesiredLRP

			removeErr error
		)

		JustBeforeEach(func() {
			desiredLRP = model_helpers.NewValidDesiredLRP("super-lrp")
			err := client.DesireLRP(logger, "some-trace-id", desiredLRP)
			Expect(err).NotTo(HaveOccurred())
			removeErr = client.RemoveDesiredLRP(logger, "some-trace-id", "super-lrp")
		})

		It("creates the desired LRP in the system", func() {
			Expect(removeErr).NotTo(HaveOccurred())
			_, err := client.DesiredLRPByProcessGuid(logger, "some-trace-id", "super-lrp")
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
			err := client.DesireLRP(logger, "some-trace-id", desiredLRP)
			Expect(err).NotTo(HaveOccurred())
			update := &models.DesiredLRPUpdate{}
			update.SetInstances(3)
			updateErr = client.UpdateDesiredLRP(logger, "some-trace-id", "super-lrp", update)
		})

		It("creates the desired LRP in the system", func() {
			Expect(updateErr).NotTo(HaveOccurred())
			persistedDesiredLRP, err := client.DesiredLRPByProcessGuid(logger, "some-trace-id", "super-lrp")
			Expect(err).NotTo(HaveOccurred())
			Expect(persistedDesiredLRP.Instances).To(Equal(int32(3)))
		})
	})
})

func createDesiredLRPsInDomains(client bbs.InternalClient, domainCounts map[string]int) map[string][]*models.DesiredLRP {
	createdDesiredLRPs := map[string][]*models.DesiredLRP{}

	for domain, count := range domainCounts {
		createdDesiredLRPs[domain] = []*models.DesiredLRP{}

		for i := 0; i < count; i++ {
			guid := fmt.Sprintf("guid-%d-for-%s", i, domain)
			desiredLRP := model_helpers.NewValidDesiredLRP(guid)
			desiredLRP.Domain = domain
			err := client.DesireLRP(logger, "some-trace-id", desiredLRP)
			Expect(err).NotTo(HaveOccurred())

			createdDesiredLRPs[domain] = append(createdDesiredLRPs[domain], desiredLRP)
		}
	}

	return createdDesiredLRPs
}
