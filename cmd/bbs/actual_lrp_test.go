package main_test

import (
	"time"

	"github.com/cloudfoundry-incubator/bbs/db/etcd/internal/test_helpers"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/gogo/protobuf/proto"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ActualLRP API", func() {
	var testHelper *test_helpers.TestHelper

	BeforeEach(func() {
		bbsProcess = ginkgomon.Invoke(bbsRunner)
		testHelper = test_helpers.NewTestHelper(etcdClient)
	})

	AfterEach(func() {
		ginkgomon.Kill(bbsProcess)
	})

	Describe("GET /v1/actual_lrps_groups", func() {
		const (
			cellID          = "cell-id"
			noExpirationTTL = 0

			baseProcessGuid  = "base-process-guid"
			baseDomain       = "base-domain"
			baseInstanceGuid = "base-instance-guid"

			baseIndex = 1

			evacuatingInstanceGuid = "evacuating-instance-guid"
		)

		var (
			expectedActualLRPGroups []*models.ActualLRPGroup
			actualActualLRPGroups   models.ActualLRPGroups

			baseLRP       models.ActualLRP
			evacuatingLRP models.ActualLRP

			baseLRPKey         models.ActualLRPKey
			baseLRPInstanceKey models.ActualLRPInstanceKey
			netInfo            models.ActualLRPNetInfo

			getErr error
		)

		BeforeEach(func() {
			baseLRPKey = models.NewActualLRPKey(baseProcessGuid, baseIndex, baseDomain)
			baseLRPInstanceKey = models.NewActualLRPInstanceKey(baseInstanceGuid, cellID)
			netInfo = models.NewActualLRPNetInfo("127.0.0.1", []*models.PortMapping{{proto.Uint32(8080), proto.Uint32(80)}})

			baseLRP = models.ActualLRP{
				ActualLRPKey:         baseLRPKey,
				ActualLRPInstanceKey: baseLRPInstanceKey,
				ActualLRPNetInfo:     netInfo,
				State:                proto.String(models.ActualLRPStateRunning),
				Since:                proto.Int64(time.Now().UnixNano()),
			}
			evacuatingLRP = models.ActualLRP{
				ActualLRPKey:         baseLRPKey,
				ActualLRPInstanceKey: models.NewActualLRPInstanceKey(evacuatingInstanceGuid, cellID),
				ActualLRPNetInfo:     netInfo,
				State:                proto.String(models.ActualLRPStateRunning),
				Since:                proto.Int64(time.Now().UnixNano() - 1000),
			}

			testHelper.SetRawActualLRP(baseLRP)
			testHelper.SetRawEvacuatingActualLRP(evacuatingLRP, noExpirationTTL)
			expectedActualLRPGroups = []*models.ActualLRPGroup{{Instance: &baseLRP, Evacuating: &evacuatingLRP}}

			actualActualLRPGroups, getErr = client.ActualLRPGroups()
		})

		It("responds without error", func() {
			Expect(getErr).NotTo(HaveOccurred())
		})

		It("has the correct number of responses", func() {
			Expect(actualActualLRPGroups.GetActualLrpGroups()).To(HaveLen(1))
		})

		It("has the correct actuallrps from the bbs", func() {
			Expect(actualActualLRPGroups.GetActualLrpGroups()).To(ConsistOf(expectedActualLRPGroups))
		})
	})
})
