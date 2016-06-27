package main_test

import (
	"code.cloudfoundry.org/bbs/cmd/bbs/testrunner"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/models/test/model_helpers"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ActualLRP API", func() {
	const (
		cellID          = "cell-id"
		otherCellID     = "other-cell-id"
		noExpirationTTL = 0

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

		baseIndex       = 0
		otherIndex      = 0
		evacuatingIndex = 0
		unclaimedIndex  = 0
		crashingIndex   = 0
	)

	var (
		expectedActualLRPGroups []*models.ActualLRPGroup
		actualActualLRPGroups   []*models.ActualLRPGroup

		baseLRP               *models.ActualLRP
		otherLRP              *models.ActualLRP
		evacuatingLRP         *models.ActualLRP
		evacuatingInstanceLRP *models.ActualLRP
		unclaimedLRP          *models.ActualLRP
		crashingLRP           *models.ActualLRP

		baseLRPKey         models.ActualLRPKey
		baseLRPInstanceKey models.ActualLRPInstanceKey

		evacuatingLRPKey         models.ActualLRPKey
		evacuatingLRPInstanceKey models.ActualLRPInstanceKey

		otherLRPKey         models.ActualLRPKey
		otherLRPInstanceKey models.ActualLRPInstanceKey

		crashingLRPKey         models.ActualLRPKey
		crashingLRPInstanceKey models.ActualLRPInstanceKey

		netInfo         models.ActualLRPNetInfo
		unclaimedLRPKey models.ActualLRPKey

		filter models.ActualLRPFilter

		getErr error
	)

	BeforeEach(func() {
		bbsRunner = testrunner.New(bbsBinPath, bbsArgs)
		bbsProcess = ginkgomon.Invoke(bbsRunner)

		filter = models.ActualLRPFilter{}
		expectedActualLRPGroups = []*models.ActualLRPGroup{}
		actualActualLRPGroups = []*models.ActualLRPGroup{}

		baseLRPKey = models.NewActualLRPKey(baseProcessGuid, baseIndex, baseDomain)
		baseLRPInstanceKey = models.NewActualLRPInstanceKey(baseInstanceGuid, cellID)

		evacuatingLRPKey = models.NewActualLRPKey(evacuatingProcessGuid, evacuatingIndex, evacuatingDomain)
		evacuatingLRPInstanceKey = models.NewActualLRPInstanceKey(evacuatingInstanceGuid, cellID)
		otherLRPKey = models.NewActualLRPKey(otherProcessGuid, otherIndex, otherDomain)
		otherLRPInstanceKey = models.NewActualLRPInstanceKey(otherInstanceGuid, otherCellID)

		netInfo = models.NewActualLRPNetInfo("127.0.0.1", models.NewPortMapping(8080, 80))

		unclaimedLRPKey = models.NewActualLRPKey(unclaimedProcessGuid, unclaimedIndex, unclaimedDomain)

		crashingLRPKey = models.NewActualLRPKey(crashingProcessGuid, crashingIndex, crashingDomain)
		crashingLRPInstanceKey = models.NewActualLRPInstanceKey(crashingInstanceGuid, otherCellID)

		baseLRP = &models.ActualLRP{
			ActualLRPKey:         baseLRPKey,
			ActualLRPInstanceKey: baseLRPInstanceKey,
			ActualLRPNetInfo:     netInfo,
			State:                models.ActualLRPStateRunning,
		}

		evacuatingLRP = &models.ActualLRP{
			ActualLRPKey:         evacuatingLRPKey,
			ActualLRPInstanceKey: evacuatingLRPInstanceKey,
			ActualLRPNetInfo:     netInfo,
			State:                models.ActualLRPStateRunning,
		}

		evacuatingInstanceLRP = &models.ActualLRP{
			ActualLRPKey: evacuatingLRPKey,
			State:        models.ActualLRPStateUnclaimed,
		}

		otherLRP = &models.ActualLRP{
			ActualLRPKey:         otherLRPKey,
			ActualLRPInstanceKey: otherLRPInstanceKey,
			ActualLRPNetInfo:     netInfo,
			State:                models.ActualLRPStateRunning,
		}

		unclaimedLRP = &models.ActualLRP{
			ActualLRPKey: unclaimedLRPKey,
			State:        models.ActualLRPStateUnclaimed,
		}

		crashingLRP = &models.ActualLRP{
			ActualLRPKey: crashingLRPKey,
			State:        models.ActualLRPStateCrashed,
			CrashReason:  "crash",
			CrashCount:   3,
		}

		var err error

		baseDesiredLRP := model_helpers.NewValidDesiredLRP(baseLRP.ProcessGuid)
		baseDesiredLRP.Domain = baseDomain
		err = client.DesireLRP(logger, baseDesiredLRP)
		Expect(err).NotTo(HaveOccurred())
		err = client.StartActualLRP(logger, &baseLRPKey, &baseLRPInstanceKey, &netInfo)
		Expect(err).NotTo(HaveOccurred())

		otherDesiredLRP := model_helpers.NewValidDesiredLRP(otherLRP.ProcessGuid)
		otherDesiredLRP.Domain = otherDomain
		Expect(client.DesireLRP(logger, otherDesiredLRP)).To(Succeed())
		err = client.StartActualLRP(logger, &otherLRPKey, &otherLRPInstanceKey, &netInfo)
		Expect(err).NotTo(HaveOccurred())

		evacuatingDesiredLRP := model_helpers.NewValidDesiredLRP(evacuatingLRP.ProcessGuid)
		evacuatingDesiredLRP.Domain = evacuatingDomain
		err = client.DesireLRP(logger, evacuatingDesiredLRP)
		Expect(err).NotTo(HaveOccurred())
		err = client.StartActualLRP(logger, &evacuatingLRPKey, &evacuatingLRPInstanceKey, &netInfo)
		Expect(err).NotTo(HaveOccurred())
		_, err = client.EvacuateRunningActualLRP(logger, &evacuatingLRPKey, &evacuatingLRPInstanceKey, &netInfo, noExpirationTTL)
		Expect(err).NotTo(HaveOccurred())

		unclaimedDesiredLRP := model_helpers.NewValidDesiredLRP(unclaimedLRP.ProcessGuid)
		unclaimedDesiredLRP.Domain = unclaimedDomain
		err = client.DesireLRP(logger, unclaimedDesiredLRP)
		Expect(err).NotTo(HaveOccurred())

		crashingDesiredLRP := model_helpers.NewValidDesiredLRP(crashingLRP.ProcessGuid)
		crashingDesiredLRP.Domain = crashingDomain
		Expect(client.DesireLRP(logger, crashingDesiredLRP)).To(Succeed())
		for i := 0; i < 3; i++ {
			err = client.StartActualLRP(logger, &crashingLRPKey, &crashingLRPInstanceKey, &netInfo)
			Expect(err).NotTo(HaveOccurred())
			err = client.CrashActualLRP(logger, &crashingLRPKey, &crashingLRPInstanceKey, "crash")
			Expect(err).NotTo(HaveOccurred())
		}
	})

	Describe("ActualLRPGroups", func() {
		JustBeforeEach(func() {
			actualActualLRPGroups, getErr = client.ActualLRPGroups(logger, filter)
			for _, group := range actualActualLRPGroups {
				if group.Instance != nil {
					group.Instance.Since = 0
					group.Instance.ModificationTag = models.ModificationTag{}
				}

				if group.Evacuating != nil {
					group.Evacuating.Since = 0
					group.Evacuating.ModificationTag = models.ModificationTag{}
				}
			}
		})

		It("responds without error", func() {
			Expect(getErr).NotTo(HaveOccurred())
		})

		Context("when not filtering", func() {
			It("returns all actual lrps from the bbs", func() {
				expectedActualLRPGroups = []*models.ActualLRPGroup{
					{Instance: baseLRP},
					{Instance: evacuatingInstanceLRP, Evacuating: evacuatingLRP},
					{Instance: otherLRP},
					{Instance: unclaimedLRP},
					{Instance: crashingLRP},
				}

				Expect(actualActualLRPGroups).To(ConsistOf(expectedActualLRPGroups))
			})
		})

		Context("when filtering by domain", func() {
			BeforeEach(func() {
				filter = models.ActualLRPFilter{Domain: baseDomain}
			})

			It("returns actual lrps from the requested domain", func() {
				expectedActualLRPGroups = []*models.ActualLRPGroup{{Instance: baseLRP}}
				Expect(actualActualLRPGroups).To(ConsistOf(expectedActualLRPGroups))
			})
		})

		Context("when filtering by cell", func() {
			BeforeEach(func() {
				filter = models.ActualLRPFilter{CellID: cellID}
			})

			It("returns actual lrps from the requested cell", func() {
				expectedActualLRPGroups = []*models.ActualLRPGroup{
					{Instance: baseLRP},
					{Evacuating: evacuatingLRP},
				}
				Expect(actualActualLRPGroups).To(ConsistOf(expectedActualLRPGroups))
			})
		})
	})

	Describe("ActualLRPGroupsByProcessGuid", func() {
		JustBeforeEach(func() {
			actualActualLRPGroups, getErr = client.ActualLRPGroupsByProcessGuid(logger, baseProcessGuid)
		})

		It("returns all actual lrps from the bbs", func() {
			Expect(getErr).NotTo(HaveOccurred())
			Expect(actualActualLRPGroups).To(HaveLen(1))
			baseLRP.ModificationTag.Increment()

			fetchedActualLRPGroup := actualActualLRPGroups[0]
			fetchedActualLRPGroup.Instance.Since = 0
			fetchedActualLRPGroup.Instance.ModificationTag.Epoch = ""
			Expect(fetchedActualLRPGroup.Instance).To(Equal(baseLRP))
		})
	})

	Describe("ActualLRPGroupByProcessGuidAndIndex", func() {
		var (
			actualLRPGroup         *models.ActualLRPGroup
			expectedActualLRPGroup *models.ActualLRPGroup
		)

		JustBeforeEach(func() {
			actualLRPGroup, getErr = client.ActualLRPGroupByProcessGuidAndIndex(logger, baseProcessGuid, baseIndex)
		})

		It("responds without error", func() {
			Expect(getErr).NotTo(HaveOccurred())
		})

		It("returns all actual lrps from the bbs", func() {
			actualLRPGroup.Instance.Since = 0
			actualLRPGroup.Instance.ModificationTag = models.ModificationTag{}
			expectedActualLRPGroup = &models.ActualLRPGroup{Instance: baseLRP}
			Expect(actualLRPGroup).To(Equal(expectedActualLRPGroup))
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
			claimErr = client.ClaimActualLRP(logger, unclaimedProcessGuid, unclaimedIndex, &instanceKey)
		})

		It("claims the actual_lrp", func() {
			Expect(claimErr).NotTo(HaveOccurred())

			expectedActualLRP := *unclaimedLRP
			expectedActualLRP.State = models.ActualLRPStateClaimed
			expectedActualLRP.ActualLRPInstanceKey = instanceKey
			expectedActualLRP.ModificationTag.Increment()

			fetchedActualLRPGroup, err := client.ActualLRPGroupByProcessGuidAndIndex(logger, unclaimedProcessGuid, unclaimedIndex)
			Expect(err).NotTo(HaveOccurred())

			fetchedActualLRP, evacuating := fetchedActualLRPGroup.Resolve()
			Expect(evacuating).To(BeFalse())
			fetchedActualLRP.ModificationTag.Epoch = ""
			fetchedActualLRP.Since = 0

			Expect(*fetchedActualLRP).To(Equal(expectedActualLRP))
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
			startErr = client.StartActualLRP(logger, &unclaimedLRPKey, &instanceKey, &netInfo)
		})

		It("starts the actual_lrp", func() {
			Expect(startErr).NotTo(HaveOccurred())

			expectedActualLRP := *unclaimedLRP
			expectedActualLRP.State = models.ActualLRPStateRunning
			expectedActualLRP.ActualLRPInstanceKey = instanceKey
			expectedActualLRP.ActualLRPNetInfo = netInfo
			expectedActualLRP.ModificationTag.Increment()
			expectedActualLRP.Since = 0

			fetchedActualLRPGroup, err := client.ActualLRPGroupByProcessGuidAndIndex(logger, unclaimedProcessGuid, unclaimedIndex)
			Expect(err).NotTo(HaveOccurred())

			fetchedActualLRP, evacuating := fetchedActualLRPGroup.Resolve()
			Expect(evacuating).To(BeFalse())
			fetchedActualLRP.ModificationTag.Epoch = ""
			fetchedActualLRP.Since = 0

			Expect(*fetchedActualLRP).To(Equal(expectedActualLRP))
		})
	})

	Describe("FailActualLRP", func() {
		var (
			errorMessage string
			failErr      error
		)

		JustBeforeEach(func() {
			errorMessage = "some bad ocurred"
			failErr = client.FailActualLRP(logger, &unclaimedLRPKey, errorMessage)
		})

		It("fails the actual_lrp", func() {
			Expect(failErr).NotTo(HaveOccurred())

			fetchedActualLRPGroup, err := client.ActualLRPGroupByProcessGuidAndIndex(logger, unclaimedProcessGuid, unclaimedIndex)
			Expect(err).NotTo(HaveOccurred())

			fetchedActualLRP, _ := fetchedActualLRPGroup.Resolve()
			Expect(fetchedActualLRP.PlacementError).To(Equal(errorMessage))
		})
	})

	Describe("CrashActualLRP", func() {
		var (
			errorMessage string
			crashErr     error
		)

		JustBeforeEach(func() {
			errorMessage = "some bad ocurred"
			crashErr = client.CrashActualLRP(logger, &baseLRPKey, &baseLRPInstanceKey, errorMessage)
		})

		It("crashes the actual_lrp", func() {
			Expect(crashErr).NotTo(HaveOccurred())

			fetchedActualLRPGroup, err := client.ActualLRPGroupByProcessGuidAndIndex(logger, baseProcessGuid, baseIndex)
			Expect(err).NotTo(HaveOccurred())

			fetchedActualLRP, _ := fetchedActualLRPGroup.Resolve()
			Expect(fetchedActualLRP.State).To(Equal(models.ActualLRPStateUnclaimed))
			Expect(fetchedActualLRP.CrashCount).To(Equal(int32(1)))
			Expect(fetchedActualLRP.CrashReason).To(Equal(errorMessage))
		})
	})

	Describe("RetireActualLRP", func() {
		var (
			retireErr error
		)

		JustBeforeEach(func() {
			retireErr = client.RetireActualLRP(logger, &unclaimedLRPKey)
		})

		It("retires the actual_lrp", func() {
			Expect(retireErr).NotTo(HaveOccurred())

			_, err := client.ActualLRPGroupByProcessGuidAndIndex(logger, unclaimedProcessGuid, unclaimedIndex)
			Expect(err).To(Equal(models.ErrResourceNotFound))
		})
	})

	Describe("RemoveActualLRP", func() {
		var (
			removeErr   error
			instanceKey *models.ActualLRPInstanceKey
		)

		JustBeforeEach(func() {
			removeErr = client.RemoveActualLRP(logger, otherProcessGuid, otherIndex, instanceKey)
		})

		Describe("when the instance key isn't preset", func() {
			BeforeEach(func() {
				instanceKey = nil
			})

			It("removes the actual_lrp", func() {
				Expect(removeErr).NotTo(HaveOccurred())

				_, err := client.ActualLRPGroupByProcessGuidAndIndex(logger, otherProcessGuid, otherIndex)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})
		})

		Describe("when the instance key is equal to the current instance key", func() {
			BeforeEach(func() {
				instanceKey = &otherLRPInstanceKey
			})

			It("removes the actual_lrp", func() {
				Expect(removeErr).NotTo(HaveOccurred())

				_, err := client.ActualLRPGroupByProcessGuidAndIndex(logger, otherProcessGuid, otherIndex)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(models.ErrResourceNotFound))
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
