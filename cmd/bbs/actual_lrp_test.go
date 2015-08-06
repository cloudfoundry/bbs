package main_test

import (
	"time"

	"github.com/cloudfoundry-incubator/bbs/models"

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

		otherProcessGuid  = "other-process-guid"
		otherDomain       = "other-domain"
		otherInstanceGuid = "other-instance-guid"

		unclaimedProcessGuid = "unclaimed-process-guid"
		unclaimedDomain      = "unclaimed-domain"

		crashingProcessGuid  = "crashing-process-guid"
		crashingDomain       = "crashing-domain"
		crashingInstanceGuid = "crashing-instance-guid"

		baseIndex      = 1
		otherIndex     = 1
		unclaimedIndex = 2
		crashingIndex  = 1

		evacuatingInstanceGuid = "evacuating-instance-guid"
	)

	var (
		expectedActualLRPGroups []*models.ActualLRPGroup
		actualActualLRPGroups   []*models.ActualLRPGroup

		baseLRP       *models.ActualLRP
		otherLRP      *models.ActualLRP
		evacuatingLRP *models.ActualLRP
		unclaimedLRP  *models.ActualLRP
		crashingLRP   *models.ActualLRP

		baseLRPKey         models.ActualLRPKey
		baseLRPInstanceKey models.ActualLRPInstanceKey

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
		filter = models.ActualLRPFilter{}
		expectedActualLRPGroups = []*models.ActualLRPGroup{}
		actualActualLRPGroups = []*models.ActualLRPGroup{}

		baseLRPKey = models.NewActualLRPKey(baseProcessGuid, baseIndex, baseDomain)
		baseLRPInstanceKey = models.NewActualLRPInstanceKey(baseInstanceGuid, cellID)

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
			Since:                time.Now().UnixNano(),
		}

		evacuatingLRP = &models.ActualLRP{
			ActualLRPKey:         baseLRPKey,
			ActualLRPInstanceKey: evacuatingLRPInstanceKey,
			ActualLRPNetInfo:     netInfo,
			State:                models.ActualLRPStateRunning,
			Since:                time.Now().UnixNano() - 1000,
		}

		otherLRP = &models.ActualLRP{
			ActualLRPKey:         otherLRPKey,
			ActualLRPInstanceKey: otherLRPInstanceKey,
			ActualLRPNetInfo:     netInfo,
			State:                models.ActualLRPStateRunning,
			Since:                time.Now().UnixNano(),
		}

		unclaimedLRP = &models.ActualLRP{
			ActualLRPKey: unclaimedLRPKey,
			State:        models.ActualLRPStateUnclaimed,
			Since:        time.Now().UnixNano(),
		}

		crashingLRP = &models.ActualLRP{
			ActualLRPKey:         crashingLRPKey,
			ActualLRPInstanceKey: crashingLRPInstanceKey,
			ActualLRPNetInfo:     netInfo,
			State:                models.ActualLRPStateRunning,
			CrashCount:           100,
			Since:                time.Now().UnixNano(),
		}

		etcdHelper.CreateValidDesiredLRP(baseLRP.ActualLRPKey.ProcessGuid)
		etcdHelper.CreateValidDesiredLRP(otherLRP.ActualLRPKey.ProcessGuid)

		etcdHelper.SetRawActualLRP(baseLRP)
		etcdHelper.SetRawActualLRP(otherLRP)
		etcdHelper.SetRawEvacuatingActualLRP(evacuatingLRP, noExpirationTTL)
		etcdHelper.SetRawActualLRP(unclaimedLRP)
		etcdHelper.SetRawActualLRP(crashingLRP)
	})

	Describe("GET /v1/actual_lrps_groups", func() {
		JustBeforeEach(func() {
			actualActualLRPGroups, getErr = client.ActualLRPGroups(filter)
		})

		It("responds without error", func() {
			Expect(getErr).NotTo(HaveOccurred())
		})

		Context("when not filtering", func() {
			It("returns all actual lrps from the bbs", func() {
				Expect(actualActualLRPGroups).To(HaveLen(4))
				expectedActualLRPGroups = []*models.ActualLRPGroup{
					{Instance: baseLRP, Evacuating: evacuatingLRP},
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
				expectedActualLRPGroups = []*models.ActualLRPGroup{{Instance: baseLRP, Evacuating: evacuatingLRP}}
				Expect(actualActualLRPGroups).To(ConsistOf(expectedActualLRPGroups))
			})
		})

		Context("when filtering by cell", func() {
			BeforeEach(func() {
				filter = models.ActualLRPFilter{CellID: cellID}
			})

			It("returns actual lrps from the requested cell", func() {
				expectedActualLRPGroups = []*models.ActualLRPGroup{{Instance: baseLRP, Evacuating: evacuatingLRP}}
				Expect(actualActualLRPGroups).To(ConsistOf(expectedActualLRPGroups))
			})
		})
	})

	Describe("GET /v1/actual_lrps_groups/:process_guid", func() {
		JustBeforeEach(func() {
			actualActualLRPGroups, getErr = client.ActualLRPGroupsByProcessGuid(baseProcessGuid)
		})

		It("returns all actual lrps from the bbs", func() {
			Expect(getErr).NotTo(HaveOccurred())
			Expect(actualActualLRPGroups).To(HaveLen(1))
			expectedActualLRPGroups = []*models.ActualLRPGroup{{Instance: baseLRP, Evacuating: evacuatingLRP}}
			Expect(actualActualLRPGroups).To(ConsistOf(expectedActualLRPGroups))
		})
	})

	Describe("GET /v1/actual_lrps_groups/:process_guid/index/:index", func() {
		var (
			actualLRPGroup         *models.ActualLRPGroup
			expectedActualLRPGroup *models.ActualLRPGroup
		)

		JustBeforeEach(func() {
			actualLRPGroup, getErr = client.ActualLRPGroupByProcessGuidAndIndex(baseProcessGuid, baseIndex)
		})

		It("responds without error", func() {
			Expect(getErr).NotTo(HaveOccurred())
		})

		It("returns all actual lrps from the bbs", func() {
			expectedActualLRPGroup = &models.ActualLRPGroup{Instance: baseLRP, Evacuating: evacuatingLRP}
			Expect(actualActualLRPGroups).To(Equal(expectedActualLRPGroups))
		})
	})

	Describe("POST /v1/actual_lrps/claim", func() {
		var (
			actualLRP   *models.ActualLRP
			instanceKey models.ActualLRPInstanceKey
			claimErr    error
		)

		JustBeforeEach(func() {
			instanceKey = models.ActualLRPInstanceKey{
				CellId:       "my-cell-id",
				InstanceGuid: "my-instance-guid",
			}
			actualLRP, claimErr = client.ClaimActualLRP(unclaimedProcessGuid, unclaimedIndex, &instanceKey)
		})

		It("claims the actual_lrp", func() {
			Expect(claimErr).NotTo(HaveOccurred())

			expectedActualLRP := *unclaimedLRP
			expectedActualLRP.State = models.ActualLRPStateClaimed
			expectedActualLRP.ActualLRPInstanceKey = instanceKey
			expectedActualLRP.ModificationTag.Increment()
			Expect(*actualLRP).To(Equal(expectedActualLRP))

			fetchedActualLRPGroup, err := client.ActualLRPGroupByProcessGuidAndIndex(unclaimedProcessGuid, unclaimedIndex)
			Expect(err).NotTo(HaveOccurred())

			fetchedActualLRP, evacuating := fetchedActualLRPGroup.Resolve()
			Expect(evacuating).To(BeFalse())

			Expect(*fetchedActualLRP).To(Equal(expectedActualLRP))
		})
	})

	Describe("POST /v1/actual_lrps/start", func() {
		var (
			actualLRP   *models.ActualLRP
			instanceKey models.ActualLRPInstanceKey
			startErr    error
		)

		JustBeforeEach(func() {
			instanceKey = models.ActualLRPInstanceKey{
				CellId:       "my-cell-id",
				InstanceGuid: "my-instance-guid",
			}
			actualLRP, startErr = client.StartActualLRP(&unclaimedLRPKey, &instanceKey, &netInfo)
		})

		It("starts the actual_lrp", func() {
			Expect(startErr).NotTo(HaveOccurred())

			expectedActualLRP := *unclaimedLRP
			expectedActualLRP.State = models.ActualLRPStateRunning
			expectedActualLRP.ActualLRPInstanceKey = instanceKey
			expectedActualLRP.ActualLRPNetInfo = netInfo
			expectedActualLRP.ModificationTag.Increment()
			expectedActualLRP.Since = 0
			actualLRP.Since = 0
			Expect(*actualLRP).To(Equal(expectedActualLRP))

			fetchedActualLRPGroup, err := client.ActualLRPGroupByProcessGuidAndIndex(unclaimedProcessGuid, unclaimedIndex)
			Expect(err).NotTo(HaveOccurred())

			fetchedActualLRP, evacuating := fetchedActualLRPGroup.Resolve()
			Expect(evacuating).To(BeFalse())
			fetchedActualLRP.Since = 0

			Expect(*fetchedActualLRP).To(Equal(expectedActualLRP))
		})
	})

	Describe("POST /v1/actual_lrps/fail", func() {
		var (
			instanceKey  models.ActualLRPInstanceKey
			errorMessage string
			failErr      error
		)

		JustBeforeEach(func() {
			errorMessage = "some bad ocurred"
			instanceKey = models.ActualLRPInstanceKey{
				CellId:       "my-cell-id",
				InstanceGuid: "my-instance-guid",
			}
			failErr = client.FailActualLRP(&unclaimedLRPKey, errorMessage)
		})

		It("fails the actual_lrp", func() {
			Expect(failErr).NotTo(HaveOccurred())

			fetchedActualLRPGroup, err := client.ActualLRPGroupByProcessGuidAndIndex(unclaimedProcessGuid, unclaimedIndex)
			Expect(err).NotTo(HaveOccurred())

			fetchedActualLRP, _ := fetchedActualLRPGroup.Resolve()
			Expect(fetchedActualLRP.PlacementError).To(Equal(errorMessage))
		})
	})

	Describe("POST /v1/actual_lrps/crash", func() {
		var (
			errorMessage string
			crashErr     error
		)

		JustBeforeEach(func() {
			errorMessage = "some bad ocurred"
			crashErr = client.CrashActualLRP(&crashingLRPKey, &crashingLRPInstanceKey, errorMessage)
		})

		It("crashes the actual_lrp", func() {
			Expect(crashErr).NotTo(HaveOccurred())

			fetchedActualLRPGroup, err := client.ActualLRPGroupByProcessGuidAndIndex(crashingProcessGuid, crashingIndex)
			Expect(err).NotTo(HaveOccurred())

			fetchedActualLRP, _ := fetchedActualLRPGroup.Resolve()
			Expect(fetchedActualLRP.State).To(Equal(models.ActualLRPStateCrashed))
			Expect(fetchedActualLRP.CrashCount).To(Equal(int32(101)))
			Expect(fetchedActualLRP.CrashReason).To(Equal(errorMessage))
		})
	})

	Describe("POST /v1/actual_lrps/retire", func() {
		var (
			retireErr error
		)

		JustBeforeEach(func() {
			retireErr = client.RetireActualLRP(&unclaimedLRPKey)
		})

		It("retires the actual_lrp", func() {
			Expect(retireErr).NotTo(HaveOccurred())

			_, err := client.ActualLRPGroupByProcessGuidAndIndex(unclaimedProcessGuid, unclaimedIndex)
			Expect(err).To(Equal(models.ErrResourceNotFound))
		})
	})

	Describe("DELETE /v1/actual_lrps/:process_guid/index/:index", func() {
		var (
			instanceKey models.ActualLRPInstanceKey
			removeErr   error
		)

		JustBeforeEach(func() {
			instanceKey = models.ActualLRPInstanceKey{
				CellId:       "my-cell-id",
				InstanceGuid: "my-instance-guid",
			}
			removeErr = client.RemoveActualLRP(otherProcessGuid, otherIndex)
		})

		It("removes the actual_lrp", func() {
			Expect(removeErr).NotTo(HaveOccurred())

			_, err := client.ActualLRPGroupByProcessGuidAndIndex(otherProcessGuid, otherIndex)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(models.ErrResourceNotFound))
		})
	})
})
