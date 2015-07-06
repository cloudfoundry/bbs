package db_test

import (
	"github.com/cloudfoundry-incubator/bbs/db"
	. "github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/gogo/protobuf/proto"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ActualLRPDB", func() {
	const (
		cellID          = "cell-id"
		noExpirationTTL = 0

		baseProcessGuid   = "base-process-guid"
		baseDomain        = "base-domain"
		baseInstanceGuid  = "base-instance-guid"
		otherInstanceGuid = "other-instance-guid"

		baseIndex  = 1
		otherIndex = 2

		evacuatingInstanceGuid = "evacuating-instance-guid"

		otherDomainProcessGuid = "other-domain-process-guid"

		otherDomain = "other-domain"
		otherCellID = "other-cell-id"
	)

	var (
		etcdDB db.ActualLRPDB

		baseLRP        models.ActualLRP
		otherIndexLRP  models.ActualLRP
		evacuatingLRP  models.ActualLRP
		otherDomainLRP models.ActualLRP
		otherCellIdLRP models.ActualLRP

		baseLRPKey          models.ActualLRPKey
		baseLRPInstanceKey  models.ActualLRPInstanceKey
		otherLRPInstanceKey models.ActualLRPInstanceKey
		netInfo             models.ActualLRPNetInfo
	)

	BeforeEach(func() {
		baseLRPKey = models.NewActualLRPKey(baseProcessGuid, baseIndex, baseDomain)
		baseLRPInstanceKey = models.NewActualLRPInstanceKey(baseInstanceGuid, cellID)
		otherLRPInstanceKey = models.NewActualLRPInstanceKey(otherInstanceGuid, otherCellID)

		netInfo = models.NewActualLRPNetInfo("127.0.0.1", []*models.PortMapping{{proto.Uint32(8080), proto.Uint32(80)}})

		baseLRP = models.ActualLRP{
			ActualLRPKey:         baseLRPKey,
			ActualLRPInstanceKey: baseLRPInstanceKey,
			ActualLRPNetInfo:     netInfo,
			State:                proto.String(models.ActualLRPStateRunning),
			Since:                proto.Int64(clock.Now().UnixNano()),
		}

		evacuatingLRP = models.ActualLRP{
			ActualLRPKey:         baseLRPKey,
			ActualLRPInstanceKey: models.NewActualLRPInstanceKey(evacuatingInstanceGuid, cellID),
			ActualLRPNetInfo:     netInfo,
			State:                proto.String(models.ActualLRPStateRunning),
			Since:                proto.Int64(clock.Now().UnixNano() - 1000),
		}

		otherIndexLRP = models.ActualLRP{
			ActualLRPKey:         models.NewActualLRPKey(baseProcessGuid, otherIndex, baseDomain),
			ActualLRPInstanceKey: baseLRPInstanceKey,
			State:                proto.String(models.ActualLRPStateClaimed),
			Since:                proto.Int64(clock.Now().UnixNano()),
		}

		otherDomainLRP = models.ActualLRP{
			ActualLRPKey:         models.NewActualLRPKey(otherDomainProcessGuid, baseIndex, otherDomain),
			ActualLRPInstanceKey: baseLRPInstanceKey,
			ActualLRPNetInfo:     netInfo,
			State:                proto.String(models.ActualLRPStateRunning),
			Since:                proto.Int64(clock.Now().UnixNano()),
		}

		otherCellIdLRP = models.ActualLRP{
			ActualLRPKey:         models.NewActualLRPKey(otherDomainProcessGuid, otherIndex, otherDomain),
			ActualLRPInstanceKey: otherLRPInstanceKey,
			ActualLRPNetInfo:     netInfo,
			State:                proto.String(models.ActualLRPStateRunning),
			Since:                proto.Int64(clock.Now().UnixNano()),
		}
		etcdDB = NewETCD(etcdClient)
	})

	Describe("ActualLRPGroups", func() {
		var filter models.ActualLRPFilter

		Context("when there are both /instance and /evacuating LRPs", func() {
			BeforeEach(func() {
				testHelper.SetRawActualLRP(baseLRP)
				testHelper.SetRawEvacuatingActualLRP(evacuatingLRP, noExpirationTTL)
				testHelper.SetRawActualLRP(otherDomainLRP)
				testHelper.SetRawEvacuatingActualLRP(otherIndexLRP, noExpirationTTL)
				testHelper.SetRawActualLRP(otherCellIdLRP)
			})

			It("returns all the /instance LRPs and /evacuating LRPs in groups", func() {
				actualLRPGroups, err := etcdDB.ActualLRPGroups(filter, logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(actualLRPGroups.GetActualLrpGroups()).To(ConsistOf(
					&models.ActualLRPGroup{Instance: &baseLRP, Evacuating: &evacuatingLRP},
					&models.ActualLRPGroup{Instance: &otherDomainLRP, Evacuating: nil},
					&models.ActualLRPGroup{Instance: nil, Evacuating: &otherIndexLRP},
					&models.ActualLRPGroup{Instance: &otherCellIdLRP, Evacuating: nil},
				))
			})

			It("can filter by domain", func() {
				filter.Domain = otherDomain
				actualLRPGroups, err := etcdDB.ActualLRPGroups(filter, logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(actualLRPGroups.GetActualLrpGroups()).To(ConsistOf(
					&models.ActualLRPGroup{Instance: &otherDomainLRP, Evacuating: nil},
					&models.ActualLRPGroup{Instance: &otherCellIdLRP, Evacuating: nil},
				))
			})

			It("can filter by cell id", func() {
				filter.CellID = otherCellID
				actualLRPGroups, err := etcdDB.ActualLRPGroups(filter, logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(actualLRPGroups.GetActualLrpGroups()).To(ConsistOf(
					&models.ActualLRPGroup{Instance: &otherCellIdLRP, Evacuating: nil},
				))
			})
		})

		Context("when there are no LRPs", func() {
			It("returns an empty list", func() {
				actualLRPGroups, err := etcdDB.ActualLRPGroups(filter, logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(actualLRPGroups).NotTo(BeNil())
				Expect(actualLRPGroups.GetActualLrpGroups()).To(BeEmpty())
			})
		})

		Context("when the root node exists with no child nodes", func() {
			BeforeEach(func() {
				testHelper.SetRawActualLRP(baseLRP)

				processGuid := baseLRP.ActualLRPKey.GetProcessGuid()
				_, err := etcdClient.Delete(ActualLRPProcessDir(processGuid), true)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an empty list", func() {
				actualLRPGroups, err := etcdDB.ActualLRPGroups(filter, logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(actualLRPGroups).NotTo(BeNil())
				Expect(actualLRPGroups.GetActualLrpGroups()).To(BeEmpty())
			})
		})

		Context("when there is invalid data", func() {
			BeforeEach(func() {
				testHelper.CreateValidActualLRP("some-guid", 0)
				testHelper.CreateMalformedActualLRP("some-other-guid", 0)
				testHelper.CreateValidActualLRP("some-third-guid", 0)
			})

			It("errors", func() {
				_, err := etcdDB.ActualLRPGroups(filter, logger)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
