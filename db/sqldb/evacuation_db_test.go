package sqldb_test

import (
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/models/test/model_helpers"
	"code.cloudfoundry.org/bbs/test_helpers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Evacuation", func() {
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
		actualLRP.SetRoutable(true)

		_, err := sqlDB.CreateUnclaimedActualLRP(ctx, logger, &actualLRP.ActualLRPKey)
		Expect(err).NotTo(HaveOccurred())
		_, _, err = sqlDB.ClaimActualLRP(ctx, logger, guid, index, &actualLRP.ActualLRPInstanceKey)
		Expect(err).NotTo(HaveOccurred())
		_, _, err = sqlDB.StartActualLRP(ctx, logger, &actualLRP.ActualLRPKey, &actualLRP.ActualLRPInstanceKey, &actualLRP.ActualLRPNetInfo, model_helpers.NewActualLRPInternalRoutes(), model_helpers.NewActualLRPMetricTags(), actualLRP.GetRoutable(), actualLRP.AvailabilityZone)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("EvacuateActualLRP", func() {

		BeforeEach(func() {
			queryStr := "UPDATE actual_lrps SET presence = ? WHERE process_guid = ? AND instance_index = ? AND presence = ?"
			if test_helpers.UsePostgres() {
				queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
			}
			_, err := db.ExecContext(ctx, queryStr,
				models.ActualLRP_Evacuating,
				actualLRP.ProcessGuid,
				actualLRP.Index,
				models.ActualLRP_Ordinary,
			)
			Expect(err).NotTo(HaveOccurred())

			actualLRP.Presence = models.ActualLRP_Evacuating
		})

		Context("when the something about the actual LRP has changed", func() {
			BeforeEach(func() {
				fakeClock.IncrementBySeconds(5)
				actualLRP.Since = fakeClock.Now().UnixNano()
				actualLRP.ModificationTag.Increment()
			})

			Context("when the lrp key changes", func() {
				BeforeEach(func() {
					actualLRP.Domain = "some-other-domain"
				})

				It("persists the evacuating lrp in sqldb", func() {
					evacuating, err := sqlDB.EvacuateActualLRP(ctx, logger, &actualLRP.ActualLRPKey, &actualLRP.ActualLRPInstanceKey, &actualLRP.ActualLRPNetInfo, model_helpers.NewActualLRPInternalRoutes(), model_helpers.NewActualLRPMetricTags(), true, actualLRP.AvailabilityZone)
					Expect(err).NotTo(HaveOccurred())

					actualLRPs, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: guid, Index: &index})
					Expect(err).NotTo(HaveOccurred())
					Expect(actualLRPs).To(ConsistOf(evacuating))
					Expect(actualLRPs).To(ConsistOf(actualLRP))
				})
			})

			Context("when the instance key changes", func() {
				BeforeEach(func() {
					actualLRP.ActualLRPInstanceKey.InstanceGuid = "i am different here me roar"
				})

				It("persists the evacuating lrp", func() {
					evacuating, err := sqlDB.EvacuateActualLRP(ctx, logger, &actualLRP.ActualLRPKey, &actualLRP.ActualLRPInstanceKey, &actualLRP.ActualLRPNetInfo, model_helpers.NewActualLRPInternalRoutes(), model_helpers.NewActualLRPMetricTags(), true, actualLRP.AvailabilityZone)
					Expect(err).NotTo(HaveOccurred())

					actualLRPs, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: guid, Index: &index})
					Expect(err).NotTo(HaveOccurred())
					Expect(actualLRPs).To(ConsistOf(evacuating))
					Expect(actualLRPs).To(ConsistOf(actualLRP))
				})
			})

			Context("when the netinfo changes", func() {
				BeforeEach(func() {
					actualLRP.ActualLRPNetInfo.Ports = []*models.PortMapping{
						models.NewPortMapping(6666, 7777),
					}
				})

				It("persists the evacuating lrp", func() {
					evacuating, err := sqlDB.EvacuateActualLRP(ctx, logger, &actualLRP.ActualLRPKey, &actualLRP.ActualLRPInstanceKey, &actualLRP.ActualLRPNetInfo, model_helpers.NewActualLRPInternalRoutes(), model_helpers.NewActualLRPMetricTags(), true, actualLRP.AvailabilityZone)
					Expect(err).NotTo(HaveOccurred())

					actualLRPs, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: guid, Index: &index})
					Expect(err).NotTo(HaveOccurred())
					Expect(actualLRPs).To(ConsistOf(actualLRP))
					Expect(actualLRPs).To(ConsistOf(evacuating))
				})
			})
		})

		Context("when the evacuating actual lrp already exists", func() {
			It("returns an ErrResourceExists", func() {
				_, err := sqlDB.EvacuateActualLRP(ctx, logger, &actualLRP.ActualLRPKey, &actualLRP.ActualLRPInstanceKey, &actualLRP.ActualLRPNetInfo, model_helpers.NewActualLRPInternalRoutes(), model_helpers.NewActualLRPMetricTags(), true, actualLRP.AvailabilityZone)
				Expect(err).To(Equal(models.ErrResourceExists))
			})
		})

		Context("when the evacuating actual lrp does not exist", func() {
			Context("because the record is deleted", func() {
				BeforeEach(func() {
					queryStr := "DELETE FROM actual_lrps WHERE process_guid = ? AND instance_index = ? AND presence = ?"
					if test_helpers.UsePostgres() {
						queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
					}
					_, err := db.ExecContext(ctx, queryStr, actualLRP.ProcessGuid, actualLRP.Index, models.ActualLRP_Evacuating)
					Expect(err).NotTo(HaveOccurred())

					actualLRP.CrashCount = 0
					actualLRP.CrashReason = ""
					actualLRP.Since = fakeClock.Now().UnixNano()
				})

				It("creates the evacuating actual lrp", func() {
					internalRoutes := model_helpers.NewActualLRPInternalRoutes()
					metricTags := model_helpers.NewActualLRPMetricTags()
					evacuating, err := sqlDB.EvacuateActualLRP(ctx, logger, &actualLRP.ActualLRPKey, &actualLRP.ActualLRPInstanceKey, &actualLRP.ActualLRPNetInfo, internalRoutes, metricTags, true, actualLRP.AvailabilityZone)
					Expect(err).NotTo(HaveOccurred())

					actualLRPs, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: guid, Index: &index})
					Expect(err).NotTo(HaveOccurred())
					Expect(actualLRPs).To(ConsistOf(evacuating))
					Expect(actualLRPs).To(HaveLen(1))

					Expect(actualLRPs[0].ModificationTag.Epoch).NotTo(BeNil())
					Expect(actualLRPs[0].ModificationTag.Index).To(BeEquivalentTo((0)))
					Expect(actualLRPs[0].ActualLrpInternalRoutes).To(Equal(internalRoutes))
					Expect(actualLRPs[0].MetricTags).To(Equal(metricTags))
					Expect(actualLRPs[0].GetRoutable()).To(Equal(true))
					Expect(actualLRPs[0].AvailabilityZone).To(Equal(actualLRP.AvailabilityZone))

					actualLRPs[0].ModificationTag = actualLRP.ModificationTag
					Expect(actualLRPs).To(ConsistOf(actualLRP))
				})
			})
		})

		Context("when deserializing the data fails", func() {
			BeforeEach(func() {
				queryStr := "UPDATE actual_lrps SET net_info = ? WHERE process_guid = ? AND instance_index = ? AND presence = ?"
				if test_helpers.UsePostgres() {
					queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
				}
				_, err := db.ExecContext(ctx, queryStr,
					"garbage", actualLRP.ProcessGuid, actualLRP.Index, models.ActualLRP_Evacuating)
				Expect(err).NotTo(HaveOccurred())
			})

			It("removes the invalid record and inserts a replacement", func() {
				evacuating, err := sqlDB.EvacuateActualLRP(ctx, logger, &actualLRP.ActualLRPKey, &actualLRP.ActualLRPInstanceKey, &actualLRP.ActualLRPNetInfo, model_helpers.NewActualLRPInternalRoutes(), model_helpers.NewActualLRPMetricTags(), true, actualLRP.AvailabilityZone)
				Expect(err).NotTo(HaveOccurred())

				actualLRPs, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: guid, Index: &index})
				Expect(err).NotTo(HaveOccurred())
				Expect(actualLRPs).To(ConsistOf(evacuating))

				Expect(actualLRPs).To(HaveLen(1))
				Expect(actualLRPs[0].ModificationTag.Epoch).NotTo(BeNil())
				Expect(actualLRPs[0].ModificationTag.Index).To(BeEquivalentTo((0)))

				actualLRPs[0].ModificationTag = actualLRP.ModificationTag
				Expect(actualLRPs).To(ConsistOf(actualLRP))
			})
		})
	})

	Describe("RemoveEvacuatingActualLRP", func() {
		Context("when there is an evacuating actualLRP", func() {
			BeforeEach(func() {
				queryStr := "UPDATE actual_lrps SET presence = ? WHERE process_guid = ? AND instance_index = ? AND presence = ?"
				if test_helpers.UsePostgres() {
					queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
				}
				_, err := db.ExecContext(ctx, queryStr, models.ActualLRP_Evacuating, actualLRP.ProcessGuid, actualLRP.Index, models.ActualLRP_Ordinary)
				Expect(err).NotTo(HaveOccurred())
			})

			It("removes the evacuating actual LRP", func() {
				err := sqlDB.RemoveEvacuatingActualLRP(ctx, logger, &actualLRP.ActualLRPKey, &actualLRP.ActualLRPInstanceKey)
				Expect(err).ToNot(HaveOccurred())

				lrps, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: guid, Index: &index})
				Expect(err).NotTo(HaveOccurred())
				Expect(lrps).To(BeEmpty())
			})

			Context("when the actual lrp instance key is not the same", func() {
				BeforeEach(func() {
					actualLRP.CellId = "a different cell"
				})

				It("returns a ErrActualLRPCannotBeRemoved error", func() {
					err := sqlDB.RemoveEvacuatingActualLRP(ctx, logger, &actualLRP.ActualLRPKey, &actualLRP.ActualLRPInstanceKey)
					Expect(err).To(Equal(models.ErrActualLRPCannotBeRemoved))
				})
			})

			Context("when the actualLRP is expired", func() {
				It("does not return an error", func() {
					err := sqlDB.RemoveEvacuatingActualLRP(ctx, logger, &actualLRP.ActualLRPKey, &actualLRP.ActualLRPInstanceKey)
					Expect(err).NotTo(HaveOccurred())

					lrps, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: guid, Index: &index})
					Expect(err).NotTo(HaveOccurred())
					Expect(lrps).To(BeEmpty())
				})
			})
		})

		Context("when the actualLRP does not exist", func() {
			It("does not return an error", func() {
				err := sqlDB.RemoveEvacuatingActualLRP(ctx, logger, &actualLRP.ActualLRPKey, &actualLRP.ActualLRPInstanceKey)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
