package main_test

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/models/test/model_helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Evacuation API", func() {
	Describe("RemoveEvacuatingActualLRP", func() {
		var actual *models.ActualLRP
		const noExpirationTTL = 0

		BeforeEach(func() {
			actual = model_helpers.NewValidActualLRP("some-process-guid", 1)
			actual.State = models.ActualLRPStateRunning
			etcdHelper.SetRawEvacuatingActualLRP(actual, noExpirationTTL)
			etcdHelper.SetRawActualLRP(actual)
			etcdHelper.CreateValidDesiredLRP(actual.ProcessGuid)
		})

		It("removes the evacuating actual_lrp", func() {
			err := client.RemoveEvacuatingActualLRP(&actual.ActualLRPKey, &actual.ActualLRPInstanceKey)
			Expect(err).NotTo(HaveOccurred())

			group, err := client.ActualLRPGroupByProcessGuidAndIndex(actual.ProcessGuid, int(actual.Index))
			Expect(err).ToNot(HaveOccurred())
			Expect(group.Evacuating).To(BeNil())
		})
	})

	Describe("EvacuateClaimedActualLRP", func() {
		var actual *models.ActualLRP

		BeforeEach(func() {
			actual = model_helpers.NewValidActualLRP("some-process-guid", 1)
			actual.State = models.ActualLRPStateClaimed
			etcdHelper.SetRawActualLRP(actual)
			etcdHelper.CreateValidDesiredLRP(actual.ProcessGuid)
		})

		It("removes the claimed actual_lrp without evacuating", func() {
			keepContainer, evacuateErr := client.EvacuateClaimedActualLRP(&actual.ActualLRPKey, &actual.ActualLRPInstanceKey)
			Expect(keepContainer).To(BeFalse())
			Expect(evacuateErr).NotTo(HaveOccurred())

			actualLRPGroup, err := client.ActualLRPGroupByProcessGuidAndIndex(actual.ProcessGuid, int(actual.Index))
			Expect(err).NotTo(HaveOccurred())
			Expect(actualLRPGroup.Evacuating).To(BeNil())
			Expect(actualLRPGroup.Instance).NotTo(BeNil())
			Expect(actualLRPGroup.Instance.State).To(Equal(models.ActualLRPStateUnclaimed))
		})
	})

	Describe("EvacuateRunningActualLRP", func() {
		var actual *models.ActualLRP

		BeforeEach(func() {
			actual = model_helpers.NewValidActualLRP("some-process-guid", 1)
			actual.State = models.ActualLRPStateRunning
			etcdHelper.SetRawActualLRP(actual)
			etcdHelper.CreateValidDesiredLRP(actual.ProcessGuid)
		})

		It("runs the evacuating ActualLRP and unclaims the instance ActualLRP", func() {
			keepContainer, err := client.EvacuateRunningActualLRP(&actual.ActualLRPKey, &actual.ActualLRPInstanceKey, &actual.ActualLRPNetInfo, uint64(10000))
			Expect(keepContainer).To(BeTrue())
			Expect(err).NotTo(HaveOccurred())

			actualLRPGroup, err := client.ActualLRPGroupByProcessGuidAndIndex(actual.ProcessGuid, int(actual.Index))
			Expect(err).NotTo(HaveOccurred())
			Expect(actualLRPGroup.Evacuating).NotTo(BeNil())
			Expect(actualLRPGroup.Instance).NotTo(BeNil())
			Expect(actualLRPGroup.Evacuating.State).To(Equal(models.ActualLRPStateRunning))
			Expect(actualLRPGroup.Instance.State).To(Equal(models.ActualLRPStateUnclaimed))
		})
	})

	Describe("EvacuateStoppedActualLRP", func() {
		var actual *models.ActualLRP
		const noExpirationTTL = 0

		BeforeEach(func() {
			actual = model_helpers.NewValidActualLRP("some-process-guid", 1)
			actual.State = models.ActualLRPStateRunning
			etcdHelper.SetRawEvacuatingActualLRP(actual, noExpirationTTL)
			etcdHelper.SetRawActualLRP(actual)
			etcdHelper.CreateValidDesiredLRP(actual.ProcessGuid)
		})

		It("deletes the container and both actualLRPs", func() {
			keepContainer, err := client.EvacuateStoppedActualLRP(&actual.ActualLRPKey, &actual.ActualLRPInstanceKey)
			Expect(keepContainer).To(BeFalse())
			Expect(err).NotTo(HaveOccurred())
			_, err = client.ActualLRPGroupByProcessGuidAndIndex(actual.ProcessGuid, int(actual.Index))
			Expect(err).To(Equal(models.ErrResourceNotFound))
		})
	})

	Describe("EvacuateCrashedActualLRP", func() {
		var actual *models.ActualLRP

		BeforeEach(func() {
			actual = model_helpers.NewValidActualLRP("some-process-guid", 1)
			actual.State = models.ActualLRPStateRunning
			etcdHelper.SetRawActualLRP(actual)
			etcdHelper.CreateValidDesiredLRP(actual.ProcessGuid)
		})
		It("removes the crashed evacuating LRP and unclaims the instance ActualLRP", func() {
			keepContainer, evacuateErr := client.EvacuateCrashedActualLRP(&actual.ActualLRPKey, &actual.ActualLRPInstanceKey, "some-reason")
			Expect(keepContainer).Should(BeFalse())
			Expect(evacuateErr).NotTo(HaveOccurred())

			actualLRPGroup, err := client.ActualLRPGroupByProcessGuidAndIndex(actual.ProcessGuid, int(actual.Index))
			Expect(err).NotTo(HaveOccurred())
			Expect(actualLRPGroup.Evacuating).To(BeNil())
			Expect(actualLRPGroup.Instance).ToNot(BeNil())
			Expect(actualLRPGroup.Instance.State).To(Equal(models.ActualLRPStateUnclaimed))
		})
	})
})
