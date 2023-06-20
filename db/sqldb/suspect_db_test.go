package sqldb_test

import (
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/models/test/model_helpers"
	"code.cloudfoundry.org/bbs/test_helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Suspect ActualLRPs", func() {
	var (
		actualLRP *models.ActualLRP
		guid      string
		index     int32
	)

	BeforeEach(func() {
		guid = "some-guid"
		index = int32(1)
		actualLRP = model_helpers.NewValidActualLRP(guid, index)
		actualLRP.CrashCount = 0
		actualLRP.CrashReason = ""
		actualLRP.Since = fakeClock.Now().UnixNano()
		actualLRP.ModificationTag = models.ModificationTag{}
		actualLRP.ModificationTag.Increment()
		actualLRP.ModificationTag.Increment()

		_, err := sqlDB.CreateUnclaimedActualLRP(ctx, logger, &actualLRP.ActualLRPKey)
		Expect(err).NotTo(HaveOccurred())
		_, _, err = sqlDB.ClaimActualLRP(ctx, logger, guid, index, &actualLRP.ActualLRPInstanceKey)
		Expect(err).NotTo(HaveOccurred())
		_, _, err = sqlDB.StartActualLRP(ctx, logger, &actualLRP.ActualLRPKey, &actualLRP.ActualLRPInstanceKey, &actualLRP.ActualLRPNetInfo, model_helpers.NewActualLRPInternalRoutes(), nil)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("PromoteSuspectActualLRP", func() {
		Context("when gettting suspect LRP fails", func() {
			It("returns an error", func() {
				_, _, _, err := sqlDB.PromoteSuspectActualLRP(ctx, logger, actualLRP.ProcessGuid, actualLRP.Index)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})

			Context("when there is an ordinary actualLRP with the same key", func() {
				var (
					replacementLRP   *models.ActualLRP
					replacementGuid  string
					replacementIndex int32
				)

				BeforeEach(func() {
					replacementGuid = "some-other-guid"
					replacementIndex = int32(0)
					replacementLRP = model_helpers.NewValidActualLRP(replacementGuid, replacementIndex)
					replacementLRP.CrashCount = 0
					replacementLRP.CrashReason = ""
					replacementLRP.Since = fakeClock.Now().UnixNano()
					replacementLRP.ModificationTag = models.ModificationTag{}
					replacementLRP.ModificationTag.Increment()
					replacementLRP.ModificationTag.Increment()
					replacementLRP.MetricTags = nil
					_, err := sqlDB.CreateUnclaimedActualLRP(ctx, logger, &replacementLRP.ActualLRPKey)
					Expect(err).NotTo(HaveOccurred())
					_, _, err = sqlDB.ClaimActualLRP(ctx, logger, replacementGuid, replacementIndex, &replacementLRP.ActualLRPInstanceKey)
					Expect(err).NotTo(HaveOccurred())
					_, _, err = sqlDB.StartActualLRP(ctx, logger, &replacementLRP.ActualLRPKey, &replacementLRP.ActualLRPInstanceKey, &replacementLRP.ActualLRPNetInfo, model_helpers.NewActualLRPInternalRoutes(), nil)
					Expect(err).NotTo(HaveOccurred())
				})

				It("does not remove the ordinary actual LRP", func() {
					beforeLRPs, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: replacementGuid, Index: &replacementIndex})
					Expect(err).NotTo(HaveOccurred())
					Expect(beforeLRPs).To(ConsistOf(replacementLRP))

					_, _, removedLRP, err := sqlDB.PromoteSuspectActualLRP(ctx, logger, replacementLRP.ProcessGuid, replacementLRP.Index)
					Expect(err).To(HaveOccurred())
					Expect(removedLRP).To(BeNil())

					afterLRPs, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: replacementGuid, Index: &replacementIndex})
					Expect(err).NotTo(HaveOccurred())
					Expect(afterLRPs).To(ConsistOf(replacementLRP))
				})
			})
		})

		Context("when there is a suspect actualLRP", func() {
			BeforeEach(func() {
				queryStr := "UPDATE actual_lrps SET presence = ? WHERE process_guid = ? AND instance_index = ? AND presence = ?"
				if test_helpers.UsePostgres() {
					queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
				}
				_, err := db.ExecContext(ctx, queryStr, models.ActualLRP_Suspect, actualLRP.ProcessGuid, actualLRP.Index, models.ActualLRP_Ordinary)
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when there is an ordinary actualLRP with the same key", func() {
				BeforeEach(func() {
					replacementLRP := model_helpers.NewValidActualLRP(guid, index)
					_, err := sqlDB.CreateUnclaimedActualLRP(ctx, logger, &replacementLRP.ActualLRPKey)
					Expect(err).NotTo(HaveOccurred())
					_, _, err = sqlDB.ClaimActualLRP(ctx, logger, guid, index, &replacementLRP.ActualLRPInstanceKey)
					Expect(err).NotTo(HaveOccurred())
					_, _, err = sqlDB.StartActualLRP(ctx, logger, &replacementLRP.ActualLRPKey, &replacementLRP.ActualLRPInstanceKey, &replacementLRP.ActualLRPNetInfo, model_helpers.NewActualLRPInternalRoutes(), nil)
					Expect(err).NotTo(HaveOccurred())
				})

				It("removes the ordinary actual LRP", func() {
					beforeLRPs, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: guid, Index: &index})
					Expect(err).NotTo(HaveOccurred())

					beforeLRP, afterLRP, removedLRP, err := sqlDB.PromoteSuspectActualLRP(ctx, logger, actualLRP.ProcessGuid, actualLRP.Index)
					Expect(err).NotTo(HaveOccurred())
					Expect(beforeLRPs).To(ConsistOf(removedLRP, beforeLRP))

					afterLRPs, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: guid, Index: &index})
					Expect(err).NotTo(HaveOccurred())
					Expect(afterLRPs).To(ConsistOf(afterLRP))
				})
			})

			It("promotes suspect LRP to ordinary", func() {
				_, afterLRP, _, err := sqlDB.PromoteSuspectActualLRP(ctx, logger, actualLRP.ProcessGuid, actualLRP.Index)
				Expect(err).NotTo(HaveOccurred())

				afterLRPs, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: guid, Index: &index})
				Expect(err).NotTo(HaveOccurred())
				Expect(afterLRPs).To(ConsistOf(afterLRP))
				Expect(afterLRP.Presence).To(Equal(models.ActualLRP_Ordinary))
			})
		})
	})

	Describe("RemoveSuspectActualLRP", func() {
		Context("when there is a suspect actualLRP", func() {
			BeforeEach(func() {
				queryStr := "UPDATE actual_lrps SET presence = ? WHERE process_guid = ? AND instance_index = ? AND presence = ?"
				if test_helpers.UsePostgres() {
					queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
				}
				_, err := db.ExecContext(ctx, queryStr, models.ActualLRP_Suspect, actualLRP.ProcessGuid, actualLRP.Index, models.ActualLRP_Ordinary)
				Expect(err).NotTo(HaveOccurred())
			})

			It("removes the suspect actual LRP", func() {
				beforeLRPs, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: guid, Index: &index})
				Expect(err).NotTo(HaveOccurred())

				lrp, err := sqlDB.RemoveSuspectActualLRP(ctx, logger, &actualLRP.ActualLRPKey)
				Expect(err).ToNot(HaveOccurred())
				Expect(beforeLRPs).To(ConsistOf(lrp))

				afterLRPs, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: guid, Index: &index})
				Expect(err).NotTo(HaveOccurred())
				Expect(afterLRPs).To(BeEmpty())
			})
		})

		Context("when the actualLRP does not exist", func() {
			// the only LRP in the database is the Ordinary one created in the
			// BeforeEach
			It("does not return an error", func() {
				before, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: "some-guid"})
				Expect(err).NotTo(HaveOccurred())

				_, err = sqlDB.RemoveSuspectActualLRP(ctx, logger, &actualLRP.ActualLRPKey)
				Expect(err).NotTo(HaveOccurred())

				after, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: "some-guid"})
				Expect(err).NotTo(HaveOccurred())

				Expect(after).To(Equal(before))
			})
		})
	})
})
