package main_test

import (
	"fmt"

	"code.cloudfoundry.org/bbs/cmd/bbs/testrunner"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/models/test/model_helpers"
	"code.cloudfoundry.org/bbs/test_helpers"
	"github.com/tedsuo/ifrit"
	ginkgomon "github.com/tedsuo/ifrit/ginkgomon_v2"

	locketconfig "code.cloudfoundry.org/locket/cmd/locket/config"
	locketrunner "code.cloudfoundry.org/locket/cmd/locket/testrunner"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ActualLRP API", func() {
	const (
		cellID      = "cell-id"
		otherCellID = "other-cell-id"

		baseProcessGuid  = "base-process-guid"
		baseDomain       = "base-domain"
		baseInstanceGuid = "base-instance-guid"

		evacuatingProcessGuid  = "evacuating-process-guid"
		evacuatingDomain       = "evacuating-domain"
		evacuatingInstanceGuid = "evacuating-instance-guid"

		otherProcessGuid  = "other-process-guid"
		otherDomain       = "other-domain"
		otherInstanceGuid = "other-instance-guid"

		unclaimedProcessGuid = "unclaimed-process-guid"
		unclaimedDomain      = "unclaimed-domain"

		crashingProcessGuid  = "crashing-process-guid"
		crashingDomain       = "crashing-domain"
		crashingInstanceGuid = "crashing-instance-guid"

		retiredProcessGuid  = "retired-process-guid"
		retiredDomain       = "retired-domain"
		retiredInstanceGuid = "retired-instance-guid"
	)

	var (
		actualActualLRPGroups []*models.ActualLRP

		baseLRP               *models.ActualLRP
		otherLRP0             *models.ActualLRP
		otherLRP1             *models.ActualLRP
		evacuatingLRP         *models.ActualLRP
		evacuatingInstanceLRP *models.ActualLRP
		unclaimedLRP          *models.ActualLRP
		crashingLRP           *models.ActualLRP
		retiredLRP            *models.ActualLRP

		baseLRPKey         models.ActualLRPKey
		baseLRPInstanceKey models.ActualLRPInstanceKey

		evacuatingLRPKey         models.ActualLRPKey
		evacuatingLRPInstanceKey models.ActualLRPInstanceKey

		otherLRP0Key        models.ActualLRPKey
		otherLRP1Key        models.ActualLRPKey
		otherLRPInstanceKey models.ActualLRPInstanceKey

		crashingLRPKey         models.ActualLRPKey
		crashingLRPInstanceKey models.ActualLRPInstanceKey

		retiredLRPKey         models.ActualLRPKey
		retiredLRPInstanceKey models.ActualLRPInstanceKey

		netInfo          models.ActualLRPNetInfo
		unclaimedLRPKey  models.ActualLRPKey
		internalRoutes   []*models.ActualLRPInternalRoute
		metricTags       map[string]string
		availabilityZone string

		filter models.ActualLRPFilter

		getErr error

		baseIndex       = int32(0)
		otherIndex0     = int32(0)
		otherIndex1     = int32(1)
		evacuatingIndex = int32(0)
		unclaimedIndex  = int32(0)
		crashingIndex   = int32(0)
		retiredIndex    = int32(0)
	)

	BeforeEach(func() {
		filter = models.ActualLRPFilter{}
	})

	JustBeforeEach(func() {
		bbsRunner = testrunner.New(bbsBinPath, bbsConfig)
		bbsProcess = ginkgomon.Invoke(bbsRunner)

		actualActualLRPGroups = []*models.ActualLRP{}

		baseLRPKey = models.NewActualLRPKey(baseProcessGuid, baseIndex, baseDomain)
		baseLRPInstanceKey = models.NewActualLRPInstanceKey(baseInstanceGuid, cellID)

		evacuatingLRPKey = models.NewActualLRPKey(evacuatingProcessGuid, evacuatingIndex, evacuatingDomain)
		evacuatingLRPInstanceKey = models.NewActualLRPInstanceKey(evacuatingInstanceGuid, cellID)

		retiredLRPKey = models.NewActualLRPKey(retiredProcessGuid, retiredIndex, retiredDomain)
		retiredLRPInstanceKey = models.NewActualLRPInstanceKey(retiredInstanceGuid, cellID)

		otherLRP0Key = models.NewActualLRPKey(otherProcessGuid, otherIndex0, otherDomain)
		otherLRP1Key = models.NewActualLRPKey(otherProcessGuid, otherIndex1, otherDomain)
		otherLRPInstanceKey = models.NewActualLRPInstanceKey(otherInstanceGuid, otherCellID)

		netInfo = models.NewActualLRPNetInfo("127.0.0.1", "10.10.10.10", models.ActualLRPNetInfo_PreferredAddressHost, models.NewPortMapping(8080, 80))
		internalRoutes = model_helpers.NewActualLRPInternalRoutes()
		metricTags = model_helpers.NewActualLRPMetricTags()
		availabilityZone = "some-zone"

		unclaimedLRPKey = models.NewActualLRPKey(unclaimedProcessGuid, unclaimedIndex, unclaimedDomain)

		crashingLRPKey = models.NewActualLRPKey(crashingProcessGuid, crashingIndex, crashingDomain)
		crashingLRPInstanceKey = models.NewActualLRPInstanceKey(crashingInstanceGuid, otherCellID)

		baseLRP = &models.ActualLRP{
			ActualLRPKey:            baseLRPKey,
			ActualLRPInstanceKey:    baseLRPInstanceKey,
			ActualLRPNetInfo:        netInfo,
			State:                   models.ActualLRPStateRunning,
			ActualLrpInternalRoutes: internalRoutes,
			MetricTags:              metricTags,
			AvailabilityZone:        availabilityZone,
		}
		baseLRP.SetRoutable(true)

		evacuatingLRP = &models.ActualLRP{
			ActualLRPKey:            evacuatingLRPKey,
			ActualLRPInstanceKey:    evacuatingLRPInstanceKey,
			ActualLRPNetInfo:        netInfo,
			State:                   models.ActualLRPStateRunning,
			Presence:                models.ActualLRP_Evacuating,
			ActualLrpInternalRoutes: internalRoutes,
			MetricTags:              metricTags,
			AvailabilityZone:        availabilityZone,
		}
		evacuatingLRP.SetRoutable(true)

		evacuatingInstanceLRP = &models.ActualLRP{
			ActualLRPKey:            evacuatingLRPKey,
			State:                   models.ActualLRPStateUnclaimed,
			ActualLrpInternalRoutes: internalRoutes,
			MetricTags:              metricTags,
			AvailabilityZone:        availabilityZone,
		}
		evacuatingInstanceLRP.SetRoutable(true)

		otherLRP0 = &models.ActualLRP{
			ActualLRPKey:            otherLRP0Key,
			ActualLRPInstanceKey:    otherLRPInstanceKey,
			ActualLRPNetInfo:        netInfo,
			State:                   models.ActualLRPStateRunning,
			ActualLrpInternalRoutes: internalRoutes,
			MetricTags:              metricTags,
			AvailabilityZone:        availabilityZone,
		}
		otherLRP0.SetRoutable(true)

		otherLRP1 = &models.ActualLRP{
			ActualLRPKey:            otherLRP1Key,
			ActualLRPInstanceKey:    otherLRPInstanceKey,
			ActualLRPNetInfo:        netInfo,
			State:                   models.ActualLRPStateRunning,
			ActualLrpInternalRoutes: internalRoutes,
			MetricTags:              metricTags,
			AvailabilityZone:        availabilityZone,
		}
		otherLRP1.SetRoutable(false)

		unclaimedLRP = &models.ActualLRP{
			ActualLRPKey: unclaimedLRPKey,
			State:        models.ActualLRPStateUnclaimed,
		}
		unclaimedLRP.SetRoutable(false)

		crashingLRP = &models.ActualLRP{
			ActualLRPKey:            crashingLRPKey,
			State:                   models.ActualLRPStateCrashed,
			CrashReason:             "crash",
			CrashCount:              3,
			ActualLrpInternalRoutes: internalRoutes,
			MetricTags:              metricTags,
			AvailabilityZone:        availabilityZone,
		}
		crashingLRP.SetRoutable(false)

		retiredLRP = &models.ActualLRP{
			ActualLRPKey:            retiredLRPKey,
			State:                   models.ActualLRPStateRunning,
			ActualLrpInternalRoutes: internalRoutes,
			MetricTags:              metricTags,
			AvailabilityZone:        availabilityZone,
		}
		retiredLRP.SetRoutable(false)

		var err error

		baseDesiredLRP := model_helpers.NewValidDesiredLRP(baseLRP.ProcessGuid)
		baseDesiredLRP.Domain = baseDomain
		err = client.DesireLRP(logger, "some-trace-id", baseDesiredLRP)
		Expect(err).NotTo(HaveOccurred())
		err = client.StartActualLRP(logger, "some-trace-id", &baseLRPKey, &baseLRPInstanceKey, &netInfo, internalRoutes, metricTags, baseLRP.GetRoutable(), availabilityZone)
		Expect(err).NotTo(HaveOccurred())

		otherDesiredLRP := model_helpers.NewValidDesiredLRP(otherLRP0.ProcessGuid)
		otherDesiredLRP.Domain = otherDomain
		Expect(client.DesireLRP(logger, "some-trace-id", otherDesiredLRP)).To(Succeed())
		err = client.StartActualLRP(logger, "some-trace-id", &otherLRP0Key, &otherLRPInstanceKey, &netInfo, internalRoutes, metricTags, otherLRP0.GetRoutable(), availabilityZone)
		Expect(err).NotTo(HaveOccurred())
		err = client.StartActualLRP(logger, "some-trace-id", &otherLRP1Key, &otherLRPInstanceKey, &netInfo, internalRoutes, metricTags, otherLRP1.GetRoutable(), availabilityZone)
		Expect(err).NotTo(HaveOccurred())

		evacuatingDesiredLRP := model_helpers.NewValidDesiredLRP(evacuatingLRP.ProcessGuid)
		evacuatingDesiredLRP.Domain = evacuatingDomain
		err = client.DesireLRP(logger, "some-trace-id", evacuatingDesiredLRP)
		Expect(err).NotTo(HaveOccurred())
		err = client.StartActualLRP(logger, "some-trace-id", &evacuatingLRPKey, &evacuatingLRPInstanceKey, &netInfo, internalRoutes, metricTags, evacuatingLRP.GetRoutable(), availabilityZone)
		Expect(err).NotTo(HaveOccurred())
		_, err = client.EvacuateRunningActualLRP(logger, "some-trace-id", &evacuatingLRPKey, &evacuatingLRPInstanceKey, &netInfo, internalRoutes, metricTags, true, availabilityZone)
		Expect(err).NotTo(HaveOccurred())

		unclaimedDesiredLRP := model_helpers.NewValidDesiredLRP(unclaimedLRP.ProcessGuid)
		unclaimedDesiredLRP.Domain = unclaimedDomain
		err = client.DesireLRP(logger, "some-trace-id", unclaimedDesiredLRP)
		Expect(err).NotTo(HaveOccurred())

		crashingDesiredLRP := model_helpers.NewValidDesiredLRP(crashingLRP.ProcessGuid)
		crashingDesiredLRP.Domain = crashingDomain
		Expect(client.DesireLRP(logger, "some-trace-id", crashingDesiredLRP)).To(Succeed())
		for i := 0; i < 3; i++ {
			err = client.StartActualLRP(logger, "some-trace-id", &crashingLRPKey, &crashingLRPInstanceKey, &netInfo, internalRoutes, metricTags, crashingLRP.GetRoutable(), availabilityZone)
			Expect(err).NotTo(HaveOccurred())
			err = client.CrashActualLRP(logger, "some-trace-id", &crashingLRPKey, &crashingLRPInstanceKey, "crash")
			Expect(err).NotTo(HaveOccurred())
		}

		retiredDesiredLRP := model_helpers.NewValidDesiredLRP(retiredLRP.ProcessGuid)
		retiredDesiredLRP.Domain = retiredDomain
		err = client.DesireLRP(logger, "some-trace-id", retiredDesiredLRP)
		Expect(err).NotTo(HaveOccurred())
		err = client.StartActualLRP(logger, "some-trace-id", &retiredLRPKey, &retiredLRPInstanceKey, &netInfo, internalRoutes, metricTags, retiredLRP.GetRoutable(), availabilityZone)
		Expect(err).NotTo(HaveOccurred())
		retireErr := client.RetireActualLRP(logger, "some-trace-id", &retiredLRPKey)
		Expect(retireErr).NotTo(HaveOccurred())
	})

	Describe("ActualLRPs", func() {
		var actualActualLRPs []*models.ActualLRP

		It("responds without error", func() {
			actualActualLRPs, getErr = client.ActualLRPs(logger, "some-trace-id", filter)
			Expect(getErr).NotTo(HaveOccurred())
		})

		Context("when not filtering", func() {
			It("returns all actual lrps from the bbs", func() {
				actualActualLRPs, getErr = client.ActualLRPs(logger, "some-trace-id", filter)
				Expect(getErr).NotTo(HaveOccurred())

				Expect(actualActualLRPs).To(ConsistOf(
					test_helpers.MatchActualLRP(baseLRP),
					test_helpers.MatchActualLRP(evacuatingInstanceLRP),
					test_helpers.MatchActualLRP(evacuatingLRP),
					test_helpers.MatchActualLRP(otherLRP0),
					test_helpers.MatchActualLRP(otherLRP1),
					test_helpers.MatchActualLRP(unclaimedLRP),
					test_helpers.MatchActualLRP(crashingLRP),
				))
			})
		})

		Context("when filtering by domain", func() {
			BeforeEach(func() {
				filter = models.ActualLRPFilter{Domain: baseDomain}
			})

			It("returns actual lrps from the requested domain", func() {
				actualActualLRPs, getErr = client.ActualLRPs(logger, "some-trace-id", filter)
				Expect(getErr).NotTo(HaveOccurred())

				Expect(actualActualLRPs).To(ConsistOf(test_helpers.MatchActualLRP(baseLRP)))
			})
		})

		Context("when filtering by cell", func() {
			BeforeEach(func() {
				filter = models.ActualLRPFilter{CellID: cellID}
			})

			It("returns actual lrps from the requested cell", func() {
				actualActualLRPs, getErr = client.ActualLRPs(logger, "some-trace-id", filter)
				Expect(getErr).NotTo(HaveOccurred())
				Expect(actualActualLRPs).To(ConsistOf(
					test_helpers.MatchActualLRP(baseLRP),
					test_helpers.MatchActualLRP(evacuatingLRP),
				))
			})
		})

		Context("when filtering by process GUID", func() {
			BeforeEach(func() {
				filter = models.ActualLRPFilter{ProcessGuid: otherProcessGuid}
			})

			It("returns the actual lrps with the requested process GUID", func() {
				actualActualLRPs, getErr = client.ActualLRPs(logger, "some-trace-id", filter)
				Expect(getErr).NotTo(HaveOccurred())
				Expect(actualActualLRPs).To(ConsistOf(
					test_helpers.MatchActualLRP(otherLRP0),
					test_helpers.MatchActualLRP(otherLRP1),
				))
			})
		})

		Context("when filtering by index", func() {
			BeforeEach(func() {
				Expect(otherIndex1).NotTo(Equal(baseIndex))
				filterIdx := int32(otherIndex1)
				filter = models.ActualLRPFilter{Index: &filterIdx}
			})

			It("returns the actual lrps with the requested index", func() {
				actualActualLRPs, getErr = client.ActualLRPs(logger, "some-trace-id", filter)
				Expect(getErr).NotTo(HaveOccurred())
				Expect(actualActualLRPs).To(ConsistOf(
					test_helpers.MatchActualLRP(otherLRP1),
				))
			})
		})

		Context("when filtering by cell ID and index", func() {
			BeforeEach(func() {
				Expect(otherIndex1).NotTo(Equal(baseIndex))
				filterIdx := int32(otherIndex1)
				filter = models.ActualLRPFilter{
					CellID: cellID,
					Index:  &filterIdx,
				}
			})

			It("returns the actual lrps that matches the filter combination", func() {
				actualActualLRPs, getErr = client.ActualLRPs(logger, "some-trace-id", filter)
				Expect(getErr).NotTo(HaveOccurred())
				Expect(actualActualLRPs).To(BeEmpty())
			})
		})

		Context("with a TLS-enabled actual LRP", func() {
			const (
				tlsEnabledProcessGuid  = "tlsEnabled-process-guid"
				tlsEnabledDomain       = "tlsEnabled-domain"
				tlsEnabledInstanceGuid = "tlsEnabled-instance-guid"
				tlsEnabledIndex        = 0
			)
			var (
				tlsEnabledLRP            *models.ActualLRP
				tlsEnabledLRPKey         models.ActualLRPKey
				tlsEnabledLRPInstanceKey models.ActualLRPInstanceKey
				tlsNetInfo               models.ActualLRPNetInfo
			)

			JustBeforeEach(func() {
				tlsEnabledLRPKey = models.NewActualLRPKey(tlsEnabledProcessGuid, tlsEnabledIndex, tlsEnabledDomain)
				tlsEnabledLRPInstanceKey = models.NewActualLRPInstanceKey(tlsEnabledInstanceGuid, cellID)
				tlsNetInfo = models.NewActualLRPNetInfo("127.0.0.1", "10.10.10.10", models.ActualLRPNetInfo_PreferredAddressHost, models.NewPortMappingWithTLSProxy(8080, 80, 60042, 443))

				tlsEnabledLRP = &models.ActualLRP{
					ActualLRPKey:            tlsEnabledLRPKey,
					ActualLRPInstanceKey:    tlsEnabledLRPInstanceKey,
					ActualLRPNetInfo:        tlsNetInfo,
					State:                   models.ActualLRPStateRunning,
					ActualLrpInternalRoutes: internalRoutes,
					MetricTags:              metricTags,
					AvailabilityZone:        availabilityZone,
				}
				tlsEnabledLRP.SetRoutable(true)

				tlsEnabledDesiredLRP := model_helpers.NewValidDesiredLRP(tlsEnabledLRP.ProcessGuid)
				tlsEnabledDesiredLRP.Domain = tlsEnabledDomain

				err := client.DesireLRP(logger, "some-trace-id", tlsEnabledDesiredLRP)
				Expect(err).NotTo(HaveOccurred())
				err = client.StartActualLRP(logger, "some-trace-id", &tlsEnabledLRPKey, &tlsEnabledLRPInstanceKey, &tlsNetInfo, internalRoutes, metricTags, tlsEnabledLRP.GetRoutable(), availabilityZone)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns the TLS host and container port", func() {
				actualLRPs, err := client.ActualLRPs(logger, "some-trace-id", filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(actualLRPs).To(ContainElement(test_helpers.MatchActualLRP(tlsEnabledLRP)))
			})
		})
	})

	Describe("ActualLRPGroups", func() {
		It("responds without error", func() {
			actualActualLRPGroups, getErr = client.ActualLRPs(logger, "some-trace-id", filter)
			Expect(getErr).NotTo(HaveOccurred())
		})

		Context("when not filtering", func() {
			It("returns all actual lrps from the bbs", func() {
				actualActualLRPGroups, getErr = client.ActualLRPs(logger, "some-trace-id", filter)
				Expect(getErr).NotTo(HaveOccurred())

				Expect(actualActualLRPGroups).To(ConsistOf(
					test_helpers.MatchActualLRP(baseLRP),
					test_helpers.MatchActualLRP(evacuatingLRP),
					test_helpers.MatchActualLRP(evacuatingInstanceLRP),
					test_helpers.MatchActualLRP(otherLRP0),
					test_helpers.MatchActualLRP(otherLRP1),
					test_helpers.MatchActualLRP(unclaimedLRP),
					test_helpers.MatchActualLRP(crashingLRP),
				))
			})
		})

		Context("when filtering by domain", func() {
			BeforeEach(func() {
				filter = models.ActualLRPFilter{Domain: baseDomain}
			})

			It("returns actual lrps from the requested domain", func() {
				actualActualLRPGroups, getErr = client.ActualLRPs(logger, "some-trace-id", filter)
				Expect(getErr).NotTo(HaveOccurred())

				Expect(actualActualLRPGroups).To(ConsistOf(test_helpers.MatchActualLRP(baseLRP)))
			})
		})

		Context("when filtering by cell", func() {
			BeforeEach(func() {
				filter = models.ActualLRPFilter{CellID: cellID}
			})

			It("returns actual lrps from the requested cell", func() {
				actualActualLRPGroups, getErr = client.ActualLRPs(logger, "some-trace-id", filter)
				Expect(getErr).NotTo(HaveOccurred())
				Expect(actualActualLRPGroups).To(ConsistOf(
					test_helpers.MatchActualLRP(baseLRP),
					test_helpers.MatchActualLRP(evacuatingLRP),
				))
			})
		})

		Context("with a TLS-enabled actual LRP", func() {
			const (
				tlsEnabledProcessGuid  = "tlsEnabled-process-guid"
				tlsEnabledDomain       = "tlsEnabled-domain"
				tlsEnabledInstanceGuid = "tlsEnabled-instance-guid"
				tlsEnabledIndex        = 0
			)
			var (
				tlsEnabledLRP            *models.ActualLRP
				tlsEnabledLRPKey         models.ActualLRPKey
				tlsEnabledLRPInstanceKey models.ActualLRPInstanceKey
				tlsNetInfo               models.ActualLRPNetInfo
			)

			JustBeforeEach(func() {
				tlsEnabledLRPKey = models.NewActualLRPKey(tlsEnabledProcessGuid, tlsEnabledIndex, tlsEnabledDomain)
				tlsEnabledLRPInstanceKey = models.NewActualLRPInstanceKey(tlsEnabledInstanceGuid, cellID)
				tlsNetInfo = models.NewActualLRPNetInfo("127.0.0.1", "10.10.10.10", models.ActualLRPNetInfo_PreferredAddressHost, models.NewPortMappingWithTLSProxy(8080, 80, 60042, 443))

				tlsEnabledLRP = &models.ActualLRP{
					ActualLRPKey:            tlsEnabledLRPKey,
					ActualLRPInstanceKey:    tlsEnabledLRPInstanceKey,
					ActualLRPNetInfo:        tlsNetInfo,
					State:                   models.ActualLRPStateRunning,
					ActualLrpInternalRoutes: internalRoutes,
					MetricTags:              metricTags,
					AvailabilityZone:        availabilityZone,
				}
				tlsEnabledLRP.SetRoutable(true)

				tlsEnabledDesiredLRP := model_helpers.NewValidDesiredLRP(tlsEnabledLRP.ProcessGuid)
				tlsEnabledDesiredLRP.Domain = tlsEnabledDomain

				err := client.DesireLRP(logger, "some-trace-id", tlsEnabledDesiredLRP)
				Expect(err).NotTo(HaveOccurred())
				err = client.StartActualLRP(logger, "some-trace-id", &tlsEnabledLRPKey, &tlsEnabledLRPInstanceKey, &tlsNetInfo, internalRoutes, metricTags, tlsEnabledLRP.GetRoutable(), availabilityZone)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns the TLS host and container port", func() {
				actualLRPGroups, err := client.ActualLRPs(logger, "some-trace-id", filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(actualLRPGroups).To(ContainElement(test_helpers.MatchActualLRP(tlsEnabledLRP)))
			})
		})
	})

	Describe("ActualLRPGroupsByProcessGuid", func() {
		JustBeforeEach(func() {
			actualActualLRPGroups, getErr = client.ActualLRPs(logger, "some-trace-id", models.ActualLRPFilter{ProcessGuid: baseProcessGuid})
		})

		It("returns the specific actual lrp from the bbs", func() {
			Expect(getErr).NotTo(HaveOccurred())
			Expect(actualActualLRPGroups).To(HaveLen(1))

			fetchedActualLRPGroup := actualActualLRPGroups[0]
			Expect(fetchedActualLRPGroup).To(
				test_helpers.MatchActualLRP(baseLRP),
			)
		})
	})

	Describe("ActualLRPGroupByProcessGuidAndIndex", func() {
		var (
			actualLRPs []*models.ActualLRP
		)

		It("responds without error", func() {
			actualLRPs, getErr = client.ActualLRPs(logger, "some-trace-id", models.ActualLRPFilter{ProcessGuid: baseProcessGuid, Index: &baseIndex})
			Expect(getErr).NotTo(HaveOccurred())
		})

		It("returns all actual lrps from the bbs", func() {
			actualLRPs, getErr = client.ActualLRPs(logger, "some-trace-id", models.ActualLRPFilter{ProcessGuid: baseProcessGuid, Index: &baseIndex})
			Expect(actualLRPs).To(ConsistOf(test_helpers.MatchActualLRP(baseLRP)))
		})

		Context("when no ActualLRP group matches the process guid and index", func() {
			It("returns an error", func() {
				actualLRPs, err := client.ActualLRPs(logger, "some-trace-id", models.ActualLRPFilter{ProcessGuid: retiredProcessGuid, Index: &retiredIndex})
				Expect(err).ToNot(HaveOccurred())
				Expect(actualLRPs).To(BeEmpty())
			})
		})
	})

	Describe("ClaimActualLRP", func() {
		var (
			instanceKey models.ActualLRPInstanceKey
			claimErr    error
		)

		JustBeforeEach(func() {
			instanceKey = models.ActualLRPInstanceKey{
				CellId:       "my-cell-id",
				InstanceGuid: "my-instance-guid",
			}
			claimErr = client.ClaimActualLRP(logger, "some-trace-id", &unclaimedLRPKey, &instanceKey)
		})

		It("claims the actual_lrp", func() {
			Expect(claimErr).NotTo(HaveOccurred())

			expectedActualLRP := *unclaimedLRP
			expectedActualLRP.State = models.ActualLRPStateClaimed
			expectedActualLRP.ActualLRPInstanceKey = instanceKey

			fetchedActualLRPGroup, err := client.ActualLRPs(logger, "some-trace-id", models.ActualLRPFilter{ProcessGuid: unclaimedProcessGuid, Index: &unclaimedIndex})
			Expect(err).NotTo(HaveOccurred())

			Expect(fetchedActualLRPGroup).To(ConsistOf(test_helpers.MatchActualLRP(&expectedActualLRP)))
		})
	})

	Describe("StartActualLRP", func() {
		var (
			instanceKey models.ActualLRPInstanceKey
			startErr    error
		)

		JustBeforeEach(func() {
			instanceKey = models.ActualLRPInstanceKey{
				CellId:       "my-cell-id",
				InstanceGuid: "my-instance-guid",
			}
			startErr = client.StartActualLRP(logger, "some-trace-id", &unclaimedLRPKey, &instanceKey, &netInfo, internalRoutes, metricTags, true, availabilityZone)
		})

		It("starts the actual_lrp", func() {
			Expect(startErr).NotTo(HaveOccurred())

			expectedActualLRP := *unclaimedLRP
			expectedActualLRP.State = models.ActualLRPStateRunning
			expectedActualLRP.ActualLRPInstanceKey = instanceKey
			expectedActualLRP.ActualLRPNetInfo = netInfo
			expectedActualLRP.ActualLrpInternalRoutes = internalRoutes
			expectedActualLRP.MetricTags = metricTags
			expectedActualLRP.SetRoutable(true)
			expectedActualLRP.AvailabilityZone = availabilityZone

			fetchedActualLRPGroup, err := client.ActualLRPs(logger, "some-trace-id", models.ActualLRPFilter{ProcessGuid: unclaimedProcessGuid, Index: &unclaimedIndex})
			Expect(err).NotTo(HaveOccurred())

			Expect(fetchedActualLRPGroup).To(ConsistOf(test_helpers.MatchActualLRP(&expectedActualLRP)))
		})
	})

	Describe("FailActualLRP", func() {
		var (
			errorMessage string
			failErr      error
		)

		JustBeforeEach(func() {
			errorMessage = "some bad ocurred"
			failErr = client.FailActualLRP(logger, "some-trace-id", &unclaimedLRPKey, errorMessage)
		})

		It("fails the actual_lrp", func() {
			Expect(failErr).NotTo(HaveOccurred())

			fetchedActualLRPGroup, err := client.ActualLRPs(logger, "some-trace-id", models.ActualLRPFilter{ProcessGuid: unclaimedProcessGuid, Index: &unclaimedIndex})
			Expect(err).NotTo(HaveOccurred())

			Expect(fetchedActualLRPGroup[0].PlacementError).To(Equal(errorMessage))
		})
	})

	Describe("CrashActualLRP", func() {
		var (
			errorMessage string
			crashErr     error
		)

		JustBeforeEach(func() {
			errorMessage = "some bad ocurred"
			crashErr = client.CrashActualLRP(logger, "some-trace-id", &baseLRPKey, &baseLRPInstanceKey, errorMessage)
		})

		It("crashes the actual_lrp", func() {
			Expect(crashErr).NotTo(HaveOccurred())

			fetchedActualLRPGroup, err := client.ActualLRPs(logger, "some-trace-id", models.ActualLRPFilter{ProcessGuid: baseProcessGuid, Index: &baseIndex})
			Expect(err).NotTo(HaveOccurred())

			Expect(fetchedActualLRPGroup[0].State).To(Equal(models.ActualLRPStateUnclaimed))
			Expect(fetchedActualLRPGroup[0].CrashCount).To(Equal(int32(1)))
			Expect(fetchedActualLRPGroup[0].CrashReason).To(Equal(errorMessage))
		})
	})

	Describe("RetireActualLRP", func() {
		var (
			retireErr error
		)

		JustBeforeEach(func() {
			retireErr = client.RetireActualLRP(logger, "some-trace-id", &unclaimedLRPKey)
		})

		It("retires the actual_lrp", func() {
			Expect(retireErr).NotTo(HaveOccurred())

			actualLRPs, err := client.ActualLRPs(logger, "some-trace-id", models.ActualLRPFilter{ProcessGuid: unclaimedProcessGuid, Index: &unclaimedIndex})
			Expect(err).ToNot(HaveOccurred())
			Expect(actualLRPs).To(BeEmpty())
		})

		Context("when using locket cell presences", func() {
			var (
				locketProcess ifrit.Process
			)

			BeforeEach(func() {
				locketPort, err := portAllocator.ClaimPorts(1)
				Expect(err).NotTo(HaveOccurred())

				locketAddress := fmt.Sprintf("localhost:%d", locketPort)

				locketRunner := locketrunner.NewLocketRunner(locketBinPath, func(cfg *locketconfig.LocketConfig) {
					cfg.DatabaseConnectionString = sqlRunner.ConnectionString()
					cfg.DatabaseDriver = sqlRunner.DriverName()
					cfg.ListenAddress = locketAddress
				})

				locketProcess = ginkgomon.Invoke(locketRunner)
				bbsConfig.ClientLocketConfig = locketrunner.ClientLocketConfig()
				bbsConfig.LocketAddress = locketAddress
			})

			AfterEach(func() {
				ginkgomon.Interrupt(locketProcess)
			})

			It("retires an actual LRP when not found in locket", func() {
				retireErr = client.RetireActualLRP(logger, "some-trace-id", &baseLRPKey)
				Expect(retireErr).NotTo(HaveOccurred())

				actualLRPs, err := client.ActualLRPs(logger, "some-trace-id", models.ActualLRPFilter{ProcessGuid: baseProcessGuid, Index: &baseIndex})
				Expect(err).ToNot(HaveOccurred())
				Expect(actualLRPs).To(BeEmpty())
			})
		})
	})

	Describe("RemoveActualLRP", func() {
		var (
			removeErr   error
			instanceKey *models.ActualLRPInstanceKey
		)

		JustBeforeEach(func() {
			removeErr = client.RemoveActualLRP(logger, "some-trace-id", &otherLRP0Key, instanceKey)
		})

		Describe("when the instance key isn't preset", func() {
			BeforeEach(func() {
				instanceKey = nil
			})

			It("removes the actual_lrp", func() {
				Expect(removeErr).NotTo(HaveOccurred())

				actualLRPs, err := client.ActualLRPs(logger, "some-trace-id", models.ActualLRPFilter{ProcessGuid: otherProcessGuid, Index: &otherIndex0})
				Expect(err).ToNot(HaveOccurred())
				Expect(actualLRPs).To(BeEmpty())
			})
		})

		Describe("when the instance key is equal to the current instance key", func() {
			BeforeEach(func() {
				instanceKey = &otherLRPInstanceKey
			})

			It("removes the actual_lrp", func() {
				Expect(removeErr).NotTo(HaveOccurred())

				actualLRPs, err := client.ActualLRPs(logger, "some-trace-id", models.ActualLRPFilter{ProcessGuid: otherProcessGuid, Index: &otherIndex0})
				Expect(err).ToNot(HaveOccurred())
				Expect(actualLRPs).To(BeEmpty())
			})
		})

		Describe("when the instance key is not equal to the current instance key", func() {
			BeforeEach(func() {
				instanceKey = &baseLRPInstanceKey
			})

			It("returns an error", func() {
				Expect(removeErr).To(Equal(models.ErrResourceNotFound))
			})
		})
	})
})
