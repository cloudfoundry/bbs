package sqldb_test

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	bbsdb "code.cloudfoundry.org/bbs/db"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/models/test/model_helpers"
	"code.cloudfoundry.org/bbs/test_helpers"
	"code.cloudfoundry.org/bbs/trace"
	"code.cloudfoundry.org/routing-info/internalroutes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("LRPConvergence", func() {
	const traceId = "some-trace-id"

	actualLRPKeyWithSchedulingInfo := func(desiredLRP *models.DesiredLRP, index int) *models.ActualLRPKeyWithSchedulingInfo {
		schedulingInfo := desiredLRP.DesiredLRPSchedulingInfo()
		lrpKey := models.NewActualLRPKey(desiredLRP.ProcessGuid, int32(index), desiredLRP.Domain)

		lrp := &models.ActualLRPKeyWithSchedulingInfo{
			Key:            &lrpKey,
			SchedulingInfo: &schedulingInfo,
		}
		return lrp
	}

	var (
		cellSet models.CellSet
	)

	BeforeEach(func() {
		cellSet = models.NewCellSetFromList([]*models.CellPresence{
			{CellId: "existing-cell"},
		})
	})

	Describe("pruning evacuating lrps", func() {
		var (
			processGuid, domain string
		)

		BeforeEach(func() {
			domain = "some-domain"
			processGuid = "desired-with-evacuating-actual"
			desiredLRP := model_helpers.NewValidDesiredLRP(processGuid)
			desiredLRP.Domain = domain
			desiredLRP.Instances = 2
			err := sqlDB.DesireLRP(ctx, logger, desiredLRP)
			Expect(err).NotTo(HaveOccurred())
			_, err = sqlDB.CreateUnclaimedActualLRP(context.WithValue(ctx, trace.RequestIdHeaderCtxKey, traceId), logger, &models.ActualLRPKey{ProcessGuid: processGuid, Index: 0, Domain: domain})
			Expect(err).NotTo(HaveOccurred())
			_, _, err = sqlDB.ClaimActualLRP(context.WithValue(ctx, trace.RequestIdHeaderCtxKey, traceId), logger, processGuid, 0, &models.ActualLRPInstanceKey{InstanceGuid: "ig-1", CellId: "existing-cell"})
			Expect(err).NotTo(HaveOccurred())

			_, err = db.ExecContext(ctx, fmt.Sprintf(`UPDATE actual_lrps SET presence = %d`, models.ActualLRP_Evacuating))
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the cell is present", func() {
			It("keeps evacuating actual lrps with available cells", func() {
				sqlDB.ConvergeLRPs(ctx, logger, cellSet)

				lrps, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: processGuid})
				Expect(err).NotTo(HaveOccurred())

				Expect(lrps).To(HaveLen(1))
			})
		})

		Context("when the cell isn't present", func() {
			BeforeEach(func() {
				cellSet = models.NewCellSet()
			})

			It("clears out evacuating actual lrps with missing cells", func() {
				sqlDB.ConvergeLRPs(ctx, logger, cellSet)

				lrps, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: processGuid})
				Expect(err).NotTo(HaveOccurred())
				Expect(lrps).To(BeEmpty())
			})

			It("return an ActualLRPRemovedEvent", func() {
				actualLRPs, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: processGuid})
				Expect(err).NotTo(HaveOccurred())

				Expect(actualLRPs).To(HaveLen(1))

				result := sqlDB.ConvergeLRPs(context.WithValue(ctx, trace.RequestIdHeaderCtxKey, traceId), logger, cellSet)
				//lint:ignore SA1019 - still need to emit these events until the ActaulLRPGroup api is deleted
				Expect(result.Events).To(ContainElement(models.NewActualLRPRemovedEvent(actualLRPs[0].ToActualLRPGroup())))
				Expect(result.InstanceEvents).To(ContainElement(models.NewActualLRPInstanceRemovedEvent(actualLRPs[0], traceId)))
			})
		})
	})

	Context("when there are expired domains", func() {
		var (
			expiredDomain = "expired-domain"
		)

		BeforeEach(func() {
			fakeClock.Increment(-10 * time.Second)
			sqlDB.UpsertDomain(ctx, logger, expiredDomain, 5)
			fakeClock.Increment(10 * time.Second)
		})

		It("clears out expired domains", func() {
			fetchDomains := func() []string {
				rows, err := db.QueryContext(ctx, "SELECT domain FROM domains")
				Expect(err).NotTo(HaveOccurred())
				defer rows.Close()

				var domain string
				var results []string
				for rows.Next() {
					err = rows.Scan(&domain)
					Expect(err).NotTo(HaveOccurred())
					results = append(results, domain)
				}
				return results
			}

			Expect(fetchDomains()).To(ContainElement(expiredDomain))

			sqlDB.ConvergeLRPs(ctx, logger, cellSet)

			Expect(fetchDomains()).NotTo(ContainElement(expiredDomain))
		})

		It("logs the expired domains", func() {
			sqlDB.ConvergeLRPs(ctx, logger, cellSet)
			Eventually(logger).Should(gbytes.Say("pruning-domain.*expired-domain"))
		})
	})

	Context("when there are unclaimed LRPs", func() {
		var (
			domain      string
			processGuid string
		)

		BeforeEach(func() {
			domain = "some-domain"
			processGuid = "desired-with-unclaimed-actuals"
			desiredLRPWithStaleActuals := model_helpers.NewValidDesiredLRP(processGuid)
			desiredLRPWithStaleActuals.Domain = domain
			desiredLRPWithStaleActuals.Instances = 1
			err := sqlDB.DesireLRP(ctx, logger, desiredLRPWithStaleActuals)
			Expect(err).NotTo(HaveOccurred())
			_, err = sqlDB.CreateUnclaimedActualLRP(ctx, logger, &models.ActualLRPKey{ProcessGuid: processGuid, Index: 0, Domain: domain})
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the domain is fresh", func() {
			BeforeEach(func() {
				Expect(sqlDB.UpsertDomain(ctx, logger, domain, 5)).To(Succeed())
			})

			It("does not touch the ActualLRPs in the database", func() {
				lrpsBefore, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: processGuid})
				Expect(err).NotTo(HaveOccurred())

				sqlDB.ConvergeLRPs(ctx, logger, cellSet)

				lrpsAfter, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: processGuid})
				Expect(err).NotTo(HaveOccurred())

				Expect(lrpsAfter).To(Equal(lrpsBefore))
			})

			It("returns an empty convergence result", func() {
				result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				Expect(result).To(BeZero())
			})
		})

		Context("when the ActualLRP's presence is set to evacuating", func() {
			BeforeEach(func() {
				Expect(sqlDB.UpsertDomain(ctx, logger, domain, 5)).To(Succeed())

				queryStr := `UPDATE actual_lrps SET presence = ? WHERE process_guid = ?`
				if test_helpers.UsePostgres() {
					queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
				}
				_, err := db.ExecContext(ctx, queryStr, models.ActualLRP_Evacuating, processGuid)
				Expect(err).NotTo(HaveOccurred())
			})

			It("ignores the evacuating LRPs and sets missing LRPs to the correct value", func() {
				schedulingInfos, err := sqlDB.DesiredLRPSchedulingInfos(ctx, logger, models.DesiredLRPFilter{ProcessGuids: []string{processGuid}})
				Expect(err).NotTo(HaveOccurred())

				results := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				Expect(results.MissingLRPKeys).To(ConsistOf(
					&models.ActualLRPKeyWithSchedulingInfo{
						Key:            &models.ActualLRPKey{ProcessGuid: processGuid, Index: 0, Domain: domain},
						SchedulingInfo: schedulingInfos[0],
					},
				))
			})

			It("removes the evacuating lrps", func() {
				sqlDB.ConvergeLRPs(ctx, logger, cellSet)

				actualLRPs, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: processGuid})
				Expect(err).NotTo(HaveOccurred())
				Expect(actualLRPs).To(BeEmpty())
			})

			It("return ActualLRPRemoveEvent", func() {
				actualLRPs, err := sqlDB.ActualLRPs(context.WithValue(ctx, trace.RequestIdHeaderCtxKey, traceId), logger, models.ActualLRPFilter{ProcessGuid: processGuid})
				Expect(err).NotTo(HaveOccurred())

				Expect(actualLRPs).To(HaveLen(1))

				result := sqlDB.ConvergeLRPs(context.WithValue(ctx, trace.RequestIdHeaderCtxKey, traceId), logger, cellSet)
				//lint:ignore SA1019 - still need to emit these events until the ActaulLRPGroup api is deleted
				Expect(result.Events).To(ConsistOf(models.NewActualLRPRemovedEvent(actualLRPs[0].ToActualLRPGroup())))
				Expect(result.InstanceEvents).To(ConsistOf(models.NewActualLRPInstanceRemovedEvent(actualLRPs[0], traceId)))
			})
		})
	})

	Context("when the cellset is empty", func() {
		var (
			processGuid, domain string
			lrpKey              models.ActualLRPKey
		)

		BeforeEach(func() {
			// add suspect and ordinary lrps that are running on different cells
			domain = "some-domain"
			processGuid = "desired-with-suspect-and-running-actual"
			desiredLRP := model_helpers.NewValidDesiredLRP(processGuid)
			desiredLRP.Domain = domain
			err := sqlDB.DesireLRP(ctx, logger, desiredLRP)
			Expect(err).NotTo(HaveOccurred())

			// create the suspect lrp
			actualLRPNetInfo := models.NewActualLRPNetInfo("some-address", "container-address", models.ActualLRPNetInfo_PreferredAddressUnknown, models.NewPortMapping(2222, 4444))
			lrpKey = models.NewActualLRPKey(processGuid, 0, domain)
			_, _, err = sqlDB.StartActualLRP(ctx, logger, &lrpKey, &models.ActualLRPInstanceKey{InstanceGuid: "ig-1", CellId: "suspect-cell"}, &actualLRPNetInfo, model_helpers.NewActualLRPInternalRoutes(), model_helpers.NewActualLRPMetricTags(), false, "some-zone")
			Expect(err).NotTo(HaveOccurred())
			_, _, err = sqlDB.ChangeActualLRPPresence(ctx, logger, &lrpKey, models.ActualLRP_Ordinary, models.ActualLRP_Suspect)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns no LRP key in the SuspectKeysWithExistingCells", func() {
			result := sqlDB.ConvergeLRPs(ctx, logger, models.NewCellSet())
			Expect(result.SuspectKeysWithExistingCells).To(BeEmpty())
		})

		Context("and there is an unclaimed Ordinary LRP", func() {
			BeforeEach(func() {
				_, err := sqlDB.CreateUnclaimedActualLRP(ctx, logger, &lrpKey)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns no KeysWithMissingCells", func() {
				result := sqlDB.ConvergeLRPs(ctx, logger, models.NewCellSet())
				Expect(result.KeysWithMissingCells).To(BeEmpty())
			})
		})
	})

	Context("when there is a suspect LRP with an existing cell", func() {
		var (
			processGuid, domain string
			lrpKey              models.ActualLRPKey
			lrpKey2             models.ActualLRPKey
		)

		BeforeEach(func() {
			// add suspect and ordinary lrps that are running on different cells
			domain = "some-domain"
			processGuid = "desired-with-suspect-and-running-actual"
			desiredLRP := model_helpers.NewValidDesiredLRP(processGuid)
			desiredLRP.Domain = domain
			desiredLRP.Instances = 2
			err := sqlDB.DesireLRP(ctx, logger, desiredLRP)
			Expect(err).NotTo(HaveOccurred())

			// create the suspect lrp
			actualLRPNetInfo := models.NewActualLRPNetInfo("some-address", "container-address", models.ActualLRPNetInfo_PreferredAddressUnknown, models.NewPortMapping(2222, 4444))
			lrpKey = models.NewActualLRPKey(processGuid, 0, domain)
			_, _, err = sqlDB.StartActualLRP(ctx, logger, &lrpKey, &models.ActualLRPInstanceKey{InstanceGuid: "ig-1", CellId: "existing-cell"}, &actualLRPNetInfo, model_helpers.NewActualLRPInternalRoutes(), model_helpers.NewActualLRPMetricTags(), false, "some-zone")
			Expect(err).NotTo(HaveOccurred())
			_, _, err = sqlDB.ChangeActualLRPPresence(ctx, logger, &lrpKey, models.ActualLRP_Ordinary, models.ActualLRP_Suspect)
			Expect(err).NotTo(HaveOccurred())

			// create the second suspect lrp
			lrpKey2 = models.NewActualLRPKey(processGuid, 1, domain)
			_, _, err = sqlDB.StartActualLRP(ctx, logger, &lrpKey2, &models.ActualLRPInstanceKey{InstanceGuid: "ig-2", CellId: "suspect-cell"}, &actualLRPNetInfo, model_helpers.NewActualLRPInternalRoutes(), model_helpers.NewActualLRPMetricTags(), false, "some-zone")
			Expect(err).NotTo(HaveOccurred())
			_, _, err = sqlDB.ChangeActualLRPPresence(ctx, logger, &lrpKey2, models.ActualLRP_Ordinary, models.ActualLRP_Suspect)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the LRP key with the existing cell in the SuspectKeysWithExistingCells", func() {
			result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
			Expect(result.SuspectKeysWithExistingCells).To(ConsistOf(&lrpKey))
		})

		It("returns all suspect running LRP keys in the SuspectRunningKeys", func() {
			result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
			Expect(result.SuspectRunningKeys).To(ConsistOf(&lrpKey, &lrpKey2))
		})
	})

	Context("when there is a suspect LRP and ordinary LRP present", func() {
		var (
			processGuid, domain string
			lrpKey              models.ActualLRPKey
			lrpKey2             models.ActualLRPKey
		)

		BeforeEach(func() {
			cellSet = models.NewCellSetFromList([]*models.CellPresence{
				{CellId: "existing-cell"},
			})

			// add suspect and ordinary lrps that are running on different cells
			domain = "some-domain"
			processGuid = "desired-with-suspect-and-running-actual"
			desiredLRP := model_helpers.NewValidDesiredLRP(processGuid)
			desiredLRP.Domain = domain
			err := sqlDB.DesireLRP(ctx, logger, desiredLRP)
			Expect(err).NotTo(HaveOccurred())

			// create the suspect lrp
			actualLRPNetInfo := models.NewActualLRPNetInfo("some-address", "container-address", models.ActualLRPNetInfo_PreferredAddressUnknown, models.NewPortMapping(2222, 4444))
			lrpKey = models.NewActualLRPKey(processGuid, 0, domain)
			_, _, err = sqlDB.StartActualLRP(ctx, logger, &lrpKey, &models.ActualLRPInstanceKey{InstanceGuid: "ig-1", CellId: "suspect-cell"}, &actualLRPNetInfo, model_helpers.NewActualLRPInternalRoutes(), model_helpers.NewActualLRPMetricTags(), false, "some-zone")
			Expect(err).NotTo(HaveOccurred())
			_, err = db.ExecContext(ctx, fmt.Sprintf(`UPDATE actual_lrps SET presence = %d`, models.ActualLRP_Suspect))
			Expect(err).NotTo(HaveOccurred())

			// create the ordinary lrp
			_, _, err = sqlDB.StartActualLRP(ctx, logger, &lrpKey, &models.ActualLRPInstanceKey{InstanceGuid: "ig-2", CellId: "existing-cell"}, &actualLRPNetInfo, model_helpers.NewActualLRPInternalRoutes(), model_helpers.NewActualLRPMetricTags(), false, "some-zone")
			Expect(err).NotTo(HaveOccurred())

			// create the unrelated suspect lrp
			processGuid2 := "other-process-guid"
			desiredLRP2 := model_helpers.NewValidDesiredLRP(processGuid2)
			desiredLRP.Domain = domain
			err = sqlDB.DesireLRP(ctx, logger, desiredLRP2)
			Expect(err).NotTo(HaveOccurred())
			lrpKey2 = models.NewActualLRPKey(processGuid2, 1, domain)
			_, _, err = sqlDB.StartActualLRP(ctx, logger, &lrpKey2, &models.ActualLRPInstanceKey{InstanceGuid: "ig-2", CellId: "suspect-cell"}, &actualLRPNetInfo, model_helpers.NewActualLRPInternalRoutes(), model_helpers.NewActualLRPMetricTags(), false, "some-zone")
			Expect(err).NotTo(HaveOccurred())
			_, _, err = sqlDB.ChangeActualLRPPresence(ctx, logger, &lrpKey2, models.ActualLRP_Ordinary, models.ActualLRP_Suspect)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return the suspect lrp key in the SuspectLRPKeysToRetire", func() {
			result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
			Expect(result.SuspectLRPKeysToRetire).To(ConsistOf(&lrpKey))
		})

		It("includes the suspect lrp's cell id in the MissingCellIds", func() {
			result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
			Expect(result.MissingCellIds).To(ContainElement("suspect-cell"))
		})

		It("logs the missing cell", func() {
			sqlDB.ConvergeLRPs(ctx, logger, cellSet)
			Expect(logger).To(gbytes.Say(`detected-missing-cells.*cell_ids":\["suspect-cell"\]`))
		})

		It("returns all suspect running LRP keys in the SuspectKeys", func() {
			result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
			Expect(result.SuspectRunningKeys).To(ConsistOf(&lrpKey, &lrpKey2))
		})

		Context("if the ordinary lrp is not running", func() {
			BeforeEach(func() {
				_, _, _, err := sqlDB.CrashActualLRP(ctx, logger, &lrpKey, &models.ActualLRPInstanceKey{CellId: "existing-cell", InstanceGuid: "ig-2"}, "booooom!")
				Expect(err).NotTo(HaveOccurred())
			})

			It("does not retire the Suspect LRP", func() {
				result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				Expect(result.SuspectLRPKeysToRetire).To(BeEmpty())
			})
		})
	})

	Context("when there are orphaned suspect LRPs", func() {
		var (
			lrpKey, lrpKey2, lrpKey3 models.ActualLRPKey
		)

		BeforeEach(func() {
			cellSet = models.NewCellSetFromList([]*models.CellPresence{
				{CellId: "suspect-cell"},
			})

			domain := "some-domain"
			Expect(sqlDB.UpsertDomain(ctx, logger, domain, 5)).To(Succeed())

			var err error

			// create the suspect LRP
			actualLRPNetInfo := models.NewActualLRPNetInfo("some-address", "container-address", models.ActualLRPNetInfo_PreferredAddressUnknown, models.NewPortMapping(2222, 4444))
			processGuid := "orphaned-suspect-lrp-1"
			lrpKey = models.NewActualLRPKey(processGuid, 0, domain)
			_, _, err = sqlDB.StartActualLRP(ctx, logger, &lrpKey, &models.ActualLRPInstanceKey{InstanceGuid: "ig-1", CellId: "suspect-cell"}, &actualLRPNetInfo, model_helpers.NewActualLRPInternalRoutes(), model_helpers.NewActualLRPMetricTags(), false, "some-zone")
			Expect(err).NotTo(HaveOccurred())

			otherProcessGuid := "orphaned-suspect-lrp-2"
			lrpKey2 = models.NewActualLRPKey(otherProcessGuid, 0, domain)
			_, _, err = sqlDB.StartActualLRP(ctx, logger, &lrpKey2, &models.ActualLRPInstanceKey{InstanceGuid: "ig-2", CellId: "suspect-cell"}, &actualLRPNetInfo, model_helpers.NewActualLRPInternalRoutes(), model_helpers.NewActualLRPMetricTags(), false, "some-zone")
			Expect(err).NotTo(HaveOccurred())

			_, err = db.ExecContext(ctx, fmt.Sprintf(`UPDATE actual_lrps SET presence = %d`, models.ActualLRP_Suspect))
			Expect(err).NotTo(HaveOccurred())

			// create suspect LRP that is not orphaned
			notOrphanedProcessGuid := "suspect-lrp-that-is-not-orphaned"
			desiredLRP2 := model_helpers.NewValidDesiredLRP(notOrphanedProcessGuid)
			err = sqlDB.DesireLRP(ctx, logger, desiredLRP2)
			Expect(err).NotTo(HaveOccurred())
			lrpKey3 = models.NewActualLRPKey(notOrphanedProcessGuid, 0, domain)
			_, _, err = sqlDB.StartActualLRP(ctx, logger, &lrpKey3, &models.ActualLRPInstanceKey{InstanceGuid: "ig-3", CellId: "suspect-cell"}, &actualLRPNetInfo, model_helpers.NewActualLRPInternalRoutes(), model_helpers.NewActualLRPMetricTags(), false, "some-zone")
			Expect(err).NotTo(HaveOccurred())
			_, _, err = sqlDB.ChangeActualLRPPresence(ctx, logger, &lrpKey3, models.ActualLRP_Ordinary, models.ActualLRP_Suspect)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should only return the orphaned suspect lrp key in the SuspectLRPKeysToRetire", func() {
			result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
			Expect(result.SuspectLRPKeysToRetire).To(ConsistOf(&lrpKey, &lrpKey2))
		})
	})

	Context("when there are claimed LRPs", func() {
		var (
			domain string
			lrpKey models.ActualLRPKey
		)

		BeforeEach(func() {
			domain = "some-domain"
			lrpKey = models.ActualLRPKey{ProcessGuid: "desired-with-claimed-actuals", Index: 0, Domain: domain}
			desiredLRPWithStaleActuals := model_helpers.NewValidDesiredLRP(lrpKey.ProcessGuid)
			desiredLRPWithStaleActuals.Domain = domain
			desiredLRPWithStaleActuals.Instances = 1
			err := sqlDB.DesireLRP(ctx, logger, desiredLRPWithStaleActuals)
			Expect(err).NotTo(HaveOccurred())
			_, err = sqlDB.CreateUnclaimedActualLRP(ctx, logger, &lrpKey)
			Expect(err).NotTo(HaveOccurred())
			_, _, err = sqlDB.ClaimActualLRP(ctx, logger, lrpKey.ProcessGuid, lrpKey.Index, &models.ActualLRPInstanceKey{InstanceGuid: "instance-guid", CellId: "existing-cell"})
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the domain is fresh", func() {
			BeforeEach(func() {
				Expect(sqlDB.UpsertDomain(ctx, logger, domain, 5)).To(Succeed())
			})

			It("does not retire the extra lrps", func() {
				result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				Expect(result.KeysToRetire).To(BeEmpty())
			})

			It("does not touch the ActualLRPs in the database", func() {
				lrpsBefore, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: lrpKey.ProcessGuid})
				Expect(err).NotTo(HaveOccurred())

				sqlDB.ConvergeLRPs(ctx, logger, cellSet)

				lrpsAfter, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: lrpKey.ProcessGuid})
				Expect(err).NotTo(HaveOccurred())

				Expect(lrpsAfter).To(Equal(lrpsBefore))
			})

			Context("when the LRP is suspect", func() {
				BeforeEach(func() {
					_, _, err := sqlDB.ChangeActualLRPPresence(ctx, logger, &lrpKey, models.ActualLRP_Ordinary, models.ActualLRP_Suspect)
					Expect(err).NotTo(HaveOccurred())
				})
				It("returns the suspect claimed ActualLRP in SuspectClaimedKeys", func() {
					result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
					Expect(result.SuspectClaimedKeys).To(ConsistOf(&lrpKey))
				})
			})
		})
	})

	Context("when there are stale unclaimed LRPs", func() {
		var (
			domain      string
			processGuid string
		)

		BeforeEach(func() {
			domain = "some-domain"
			processGuid = "desired-with-stale-actuals"
			desiredLRPWithStaleActuals := model_helpers.NewValidDesiredLRP(processGuid)
			desiredLRPWithStaleActuals.Domain = domain
			desiredLRPWithStaleActuals.Instances = 2
			err := sqlDB.DesireLRP(ctx, logger, desiredLRPWithStaleActuals)
			Expect(err).NotTo(HaveOccurred())
			fakeClock.Increment(-models.StaleUnclaimedActualLRPDuration)
			_, err = sqlDB.CreateUnclaimedActualLRP(ctx, logger, &models.ActualLRPKey{ProcessGuid: processGuid, Index: 0, Domain: domain})
			Expect(err).NotTo(HaveOccurred())
			_, err = sqlDB.CreateUnclaimedActualLRP(ctx, logger, &models.ActualLRPKey{ProcessGuid: processGuid, Index: 1, Domain: domain})
			Expect(err).NotTo(HaveOccurred())
			fakeClock.Increment(models.StaleUnclaimedActualLRPDuration + 2)
		})

		Context("when the domain is fresh", func() {
			BeforeEach(func() {
				Expect(sqlDB.UpsertDomain(ctx, logger, domain, 5)).To(Succeed())
			})

			It("returns start requests", func() {
				result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				unstartedLRPKeys := result.UnstartedLRPKeys
				Expect(unstartedLRPKeys).NotTo(BeEmpty())
				Expect(logger).To(gbytes.Say("creating-start-request.*reason\":\"stale-unclaimed-lrp"))

				desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(ctx, logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				Expect(unstartedLRPKeys).To(ContainElement(actualLRPKeyWithSchedulingInfo(desiredLRP, 0)))
				Expect(unstartedLRPKeys).To(ContainElement(actualLRPKeyWithSchedulingInfo(desiredLRP, 1)))
			})

			It("does not touch the ActualLRPs in the database", func() {
				lrpsBefore, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: processGuid})
				Expect(err).NotTo(HaveOccurred())

				sqlDB.ConvergeLRPs(ctx, logger, cellSet)

				lrpsAfter, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: processGuid})
				Expect(err).NotTo(HaveOccurred())

				Expect(lrpsAfter).To(Equal(lrpsBefore))
			})

			It("should have the correct number of unclaimed LRP instances", func() {
				sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				_, unclaimed, _, _, _ := sqlDB.CountActualLRPsByState(ctx, logger)
				Expect(unclaimed).To(Equal(2))
			})
		})

		Context("when the domain is expired", func() {
			BeforeEach(func() {
				fakeClock.Increment(-10 * time.Second)
				Expect(sqlDB.UpsertDomain(ctx, logger, domain, 5)).To(Succeed())
				fakeClock.Increment(10 * time.Second)
			})

			It("returns start requests", func() {
				result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				unstartedLRPKeys := result.UnstartedLRPKeys
				Expect(unstartedLRPKeys).NotTo(BeEmpty())
				Expect(logger).To(gbytes.Say("creating-start-request.*reason\":\"stale-unclaimed-lrp"))

				desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(ctx, logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				Expect(unstartedLRPKeys).To(ContainElement(actualLRPKeyWithSchedulingInfo(desiredLRP, 0)))
				Expect(unstartedLRPKeys).To(ContainElement(actualLRPKeyWithSchedulingInfo(desiredLRP, 1)))
			})

			It("does not touch the ActualLRPs in the database", func() {
				lrpsBefore, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: processGuid})
				Expect(err).NotTo(HaveOccurred())

				sqlDB.ConvergeLRPs(ctx, logger, cellSet)

				lrpsAfter, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: processGuid})
				Expect(err).NotTo(HaveOccurred())

				Expect(lrpsAfter).To(Equal(lrpsBefore))
			})

			It("should have the correct number of unclaimed LRP instances", func() {
				sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				_, unclaimed, _, _, _ := sqlDB.CountActualLRPsByState(ctx, logger)
				Expect(unclaimed).To(Equal(2))
			})
		})

		Context("when the ActualLRPs presence is set to evacuating", func() {
			BeforeEach(func() {
				Expect(sqlDB.UpsertDomain(ctx, logger, domain, 5)).To(Succeed())

				queryStr := `UPDATE actual_lrps SET presence = ?`
				if test_helpers.UsePostgres() {
					queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
				}
				_, err := db.ExecContext(ctx, queryStr, models.ActualLRP_Evacuating)
				Expect(err).NotTo(HaveOccurred())
			})

			It("ignores the evacuating LRPs and should have the correct number of missing LRPs", func() {
				schedulingInfos, err := sqlDB.DesiredLRPSchedulingInfos(ctx, logger, models.DesiredLRPFilter{ProcessGuids: []string{processGuid}})
				Expect(err).NotTo(HaveOccurred())

				results := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				Expect(results.MissingLRPKeys).To(ConsistOf(
					&models.ActualLRPKeyWithSchedulingInfo{
						Key:            &models.ActualLRPKey{ProcessGuid: processGuid, Index: 0, Domain: domain},
						SchedulingInfo: schedulingInfos[0],
					},
					&models.ActualLRPKeyWithSchedulingInfo{
						Key:            &models.ActualLRPKey{ProcessGuid: processGuid, Index: 1, Domain: domain},
						SchedulingInfo: schedulingInfos[0],
					},
				))
			})

			It("returns the lrp keys in the MissingLRPKeys", func() {
				result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)

				desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(ctx, logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				expectedSched := desiredLRP.DesiredLRPSchedulingInfo()
				Expect(result.MissingLRPKeys).To(ContainElement(&models.ActualLRPKeyWithSchedulingInfo{
					Key:            &models.ActualLRPKey{ProcessGuid: processGuid, Index: 0, Domain: domain},
					SchedulingInfo: &expectedSched,
				}))
			})

			// it is the responsibility of the caller to create new LRPs
			It("prune the evacuating LRPs and does not create new ones", func() {
				sqlDB.ConvergeLRPs(ctx, logger, cellSet)

				lrps, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: processGuid})
				Expect(err).NotTo(HaveOccurred())
				Expect(lrps).To(BeEmpty())
			})

			It("return ActualLRPRemovedEvent for the removed evacuating LRPs", func() {
				actualLRPs, err := sqlDB.ActualLRPs(context.WithValue(ctx, trace.RequestIdHeaderCtxKey, traceId), logger, models.ActualLRPFilter{ProcessGuid: processGuid})
				Expect(err).NotTo(HaveOccurred())
				Expect(actualLRPs).To(HaveLen(2))

				result := sqlDB.ConvergeLRPs(context.WithValue(ctx, trace.RequestIdHeaderCtxKey, traceId), logger, cellSet)
				Expect(result.Events).To(ConsistOf(
					//lint:ignore SA1019 - still need to emit these events until the ActaulLRPGroup api is deleted
					models.NewActualLRPRemovedEvent(actualLRPs[0].ToActualLRPGroup()),
					//lint:ignore SA1019 - still need to emit these events until the ActaulLRPGroup api is deleted
					models.NewActualLRPRemovedEvent(actualLRPs[1].ToActualLRPGroup()),
				))
				Expect(result.InstanceEvents).To(ConsistOf(
					models.NewActualLRPInstanceRemovedEvent(actualLRPs[0], traceId),
					models.NewActualLRPInstanceRemovedEvent(actualLRPs[1], traceId),
				))
			})
		})
	})

	Context("when there is an ActualLRP on a missing cell", func() {
		var (
			domain      string
			processGuid string
		)

		BeforeEach(func() {
			domain = "some-domain"
			processGuid = "desired-with-missing-cell-actuals"
			desiredLRPWithMissingCellActuals := model_helpers.NewValidDesiredLRP(processGuid)
			desiredLRPWithMissingCellActuals.Domain = domain
			err := sqlDB.DesireLRP(ctx, logger, desiredLRPWithMissingCellActuals)
			Expect(err).NotTo(HaveOccurred())
			_, err = sqlDB.CreateUnclaimedActualLRP(ctx, logger, &models.ActualLRPKey{ProcessGuid: processGuid, Index: 0, Domain: domain})
			Expect(err).NotTo(HaveOccurred())
			_, _, err = sqlDB.ClaimActualLRP(ctx, logger, processGuid, 0, &models.ActualLRPInstanceKey{InstanceGuid: "actual-with-missing-cell", CellId: "other-cell"})
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the domain is fresh", func() {
			BeforeEach(func() {
				Expect(sqlDB.UpsertDomain(ctx, logger, domain, 5)).To(Succeed())
			})

			It("returns the start requests, actual lrp keys for actuals with missing cells and missing cell ids", func() {
				result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				keysWithMissingCells := result.KeysWithMissingCells
				missingCellIds := result.MissingCellIds

				desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(ctx, logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				index := int32(0)
				actualLRPs, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: processGuid, Index: &index})
				Expect(err).NotTo(HaveOccurred())
				expectedSched := desiredLRP.DesiredLRPSchedulingInfo()
				Expect(actualLRPs).To(HaveLen(1))
				Expect(keysWithMissingCells).To(ContainElement(&models.ActualLRPKeyWithSchedulingInfo{
					Key:            &actualLRPs[0].ActualLRPKey,
					SchedulingInfo: &expectedSched,
				}))
				Expect(missingCellIds).To(Equal([]string{"other-cell"}))
			})

			It("does not touch the ActualLRPs in the database", func() {
				lrpsBefore, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: processGuid})
				Expect(err).NotTo(HaveOccurred())

				sqlDB.ConvergeLRPs(ctx, logger, cellSet)

				lrpsAfter, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: processGuid})
				Expect(err).NotTo(HaveOccurred())

				Expect(lrpsAfter).To(Equal(lrpsBefore))
			})
		})

		Context("when the domain is expired", func() {
			BeforeEach(func() {
				fakeClock.Increment(-10 * time.Second)
				Expect(sqlDB.UpsertDomain(ctx, logger, domain, 5)).To(Succeed())
				fakeClock.Increment(10 * time.Second)
			})

			It("return ActualLRPKeys for actuals with missing cells", func() {
				result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				keysWithMissingCells := result.KeysWithMissingCells

				desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(ctx, logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				index := int32(0)
				actualLRPs, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: processGuid, Index: &index})
				Expect(err).NotTo(HaveOccurred())
				expectedSched := desiredLRP.DesiredLRPSchedulingInfo()
				Expect(actualLRPs).To(HaveLen(1))
				Expect(keysWithMissingCells).To(ContainElement(&models.ActualLRPKeyWithSchedulingInfo{
					Key:            &actualLRPs[0].ActualLRPKey,
					SchedulingInfo: &expectedSched,
				}))
			})

			It("does not touch the ActualLRPs in the database", func() {
				lrpsBefore, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: processGuid})
				Expect(err).NotTo(HaveOccurred())

				sqlDB.ConvergeLRPs(ctx, logger, cellSet)

				lrpsAfter, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: processGuid})
				Expect(err).NotTo(HaveOccurred())

				Expect(lrpsAfter).To(Equal(lrpsBefore))
			})
		})

		Context("when the lrp is evacuating", func() {
			BeforeEach(func() {
				Expect(sqlDB.UpsertDomain(ctx, logger, domain, 5)).To(Succeed())

				queryStr := `UPDATE actual_lrps SET presence = ?`
				if test_helpers.UsePostgres() {
					queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
				}
				_, err := db.ExecContext(ctx, queryStr, models.ActualLRP_Evacuating)
				Expect(err).NotTo(HaveOccurred())
			})

			It("ignores the evacuating LRPs and should have the correct number of missing LRPs", func() {
				schedulingInfos, err := sqlDB.DesiredLRPSchedulingInfos(ctx, logger, models.DesiredLRPFilter{ProcessGuids: []string{processGuid}})
				Expect(err).NotTo(HaveOccurred())

				results := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				Expect(results.MissingLRPKeys).To(ConsistOf(&models.ActualLRPKeyWithSchedulingInfo{
					Key:            &models.ActualLRPKey{ProcessGuid: processGuid, Index: 0, Domain: domain},
					SchedulingInfo: schedulingInfos[0],
				},
				))
			})

			It("returns the start requests and actual lrp keys for actuals with missing cells", func() {
				result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)

				desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(ctx, logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				expectedSched := desiredLRP.DesiredLRPSchedulingInfo()
				Expect(result.MissingLRPKeys).To(ContainElement(&models.ActualLRPKeyWithSchedulingInfo{
					Key:            &models.ActualLRPKey{ProcessGuid: processGuid, Index: 0, Domain: domain},
					SchedulingInfo: &expectedSched,
				}))
			})

			It("removes the evacuating lrp", func() {
				sqlDB.ConvergeLRPs(ctx, logger, cellSet)

				lrps, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: processGuid})
				Expect(err).NotTo(HaveOccurred())
				Expect(lrps).To(BeEmpty())
			})
		})

		It("logs the missing cells", func() {
			sqlDB.ConvergeLRPs(ctx, logger, cellSet)
			Expect(logger).To(gbytes.Say(`detected-missing-cells.*cell_ids":\["other-cell"\]`))
		})

		Context("when there are no missing cells", func() {
			BeforeEach(func() {
				cellSet = models.NewCellSetFromList([]*models.CellPresence{
					{CellId: "existing-cell"},
					{CellId: "other-cell"},
				})
			})

			It("does not log missing cells", func() {
				sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				Expect(logger).ToNot(gbytes.Say("detected-missing-cells"))
			})
		})
	})

	Context("when there are extra ActualLRPs for a DesiredLRP", func() {
		var (
			domain      string
			processGuid string
		)

		BeforeEach(func() {
			domain = "some-domain"
			processGuid = "desired-with-extra-actuals"
			desiredLRPWithExtraActuals := model_helpers.NewValidDesiredLRP(processGuid)
			desiredLRPWithExtraActuals.Domain = domain
			desiredLRPWithExtraActuals.Instances = 1
			err := sqlDB.DesireLRP(ctx, logger, desiredLRPWithExtraActuals)
			Expect(err).NotTo(HaveOccurred())
			_, err = sqlDB.CreateUnclaimedActualLRP(ctx, logger, &models.ActualLRPKey{ProcessGuid: processGuid, Index: 0, Domain: domain})
			Expect(err).NotTo(HaveOccurred())
			_, err = sqlDB.CreateUnclaimedActualLRP(ctx, logger, &models.ActualLRPKey{ProcessGuid: processGuid, Index: 4, Domain: domain})
			Expect(err).NotTo(HaveOccurred())
			_, _, err = sqlDB.ClaimActualLRP(ctx, logger, processGuid, 0, &models.ActualLRPInstanceKey{InstanceGuid: "not-extra-actual", CellId: "existing-cell"})
			Expect(err).NotTo(HaveOccurred())
			_, _, err = sqlDB.ClaimActualLRP(ctx, logger, processGuid, 4, &models.ActualLRPInstanceKey{InstanceGuid: "extra-actual", CellId: "existing-cell"})
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the domain is fresh", func() {
			BeforeEach(func() {
				Expect(sqlDB.UpsertDomain(ctx, logger, domain, 5)).To(Succeed())
			})

			It("returns extra ActualLRPs to be retired", func() {
				result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				keysToRetire := result.KeysToRetire

				actualLRPKey := models.ActualLRPKey{ProcessGuid: processGuid, Index: 4, Domain: domain}
				Expect(keysToRetire).To(ContainElement(&actualLRPKey))
			})

			It("does not touch the ActualLRPs in the database", func() {
				lrpsBefore, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: processGuid})
				Expect(err).NotTo(HaveOccurred())

				sqlDB.ConvergeLRPs(ctx, logger, cellSet)

				lrpsAfter, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: processGuid})
				Expect(err).NotTo(HaveOccurred())

				Expect(lrpsAfter).To(Equal(lrpsBefore))
			})

			It("should have the correct number of extra LRPs instances", func() {
				results := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				Expect(results.KeysToRetire).To(ConsistOf(&models.ActualLRPKey{ProcessGuid: processGuid, Index: 4, Domain: domain}))
			})
		})

		Context("when the domain is expired", func() {
			BeforeEach(func() {
				fakeClock.Increment(-10 * time.Second)
				Expect(sqlDB.UpsertDomain(ctx, logger, domain, 5)).To(Succeed())
				fakeClock.Increment(10 * time.Second)
			})

			It("returns an empty convergence result", func() {
				result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				Expect(result).To(BeZero())
			})

			It("does not touch the ActualLRPs in the database", func() {
				lrpsBefore, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: processGuid})
				Expect(err).NotTo(HaveOccurred())

				sqlDB.ConvergeLRPs(ctx, logger, cellSet)

				lrpsAfter, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: processGuid})
				Expect(err).NotTo(HaveOccurred())

				Expect(lrpsAfter).To(Equal(lrpsBefore))
			})

			It("should not have any extra LRP instances", func() {
				results := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				Expect(results.KeysToRetire).To(BeEmpty())
			})
		})

		Context("when the ActualLRP's presence is set to evacuating", func() {
			BeforeEach(func() {
				Expect(sqlDB.UpsertDomain(ctx, logger, domain, 5)).To(Succeed())

				queryStr := `UPDATE actual_lrps SET presence = ? WHERE process_guid = ?`
				if test_helpers.UsePostgres() {
					queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
				}
				_, err := db.ExecContext(ctx, queryStr, models.ActualLRP_Evacuating, processGuid)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns the lrp key to be started", func() {
				schedulingInfos, err := sqlDB.DesiredLRPSchedulingInfos(ctx, logger, models.DesiredLRPFilter{ProcessGuids: []string{processGuid}})
				Expect(err).NotTo(HaveOccurred())

				Expect(schedulingInfos).To(HaveLen(1))

				result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				Expect(result.MissingLRPKeys).To(ConsistOf(&models.ActualLRPKeyWithSchedulingInfo{
					Key:            &models.ActualLRPKey{ProcessGuid: processGuid, Index: 0, Domain: domain},
					SchedulingInfo: schedulingInfos[0],
				}))
			})
		})
	})

	Context("when there are no ActualLRPs for a DesiredLRP", func() {
		var (
			domain      string
			processGuid string
		)

		BeforeEach(func() {
			processGuid = "desired-with-missing-all-actuals" + "-" + domain
			desiredLRPWithMissingAllActuals := model_helpers.NewValidDesiredLRP(processGuid)
			desiredLRPWithMissingAllActuals.Domain = domain
			desiredLRPWithMissingAllActuals.Instances = 1
			err := sqlDB.DesireLRP(ctx, logger, desiredLRPWithMissingAllActuals)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("and the domain is fresh", func() {
			BeforeEach(func() {
				Expect(sqlDB.UpsertDomain(ctx, logger, domain, 5)).To(Succeed())
			})

			It("should have the correct number of missing LRP instances", func() {
				schedulingInfos, err := sqlDB.DesiredLRPSchedulingInfos(ctx, logger, models.DesiredLRPFilter{ProcessGuids: []string{processGuid}})
				Expect(err).NotTo(HaveOccurred())

				Expect(schedulingInfos).To(HaveLen(1))

				result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				Expect(result.MissingLRPKeys).To(ConsistOf(&models.ActualLRPKeyWithSchedulingInfo{
					Key:            &models.ActualLRPKey{ProcessGuid: processGuid, Index: 0, Domain: domain},
					SchedulingInfo: schedulingInfos[0],
				}))
			})

			It("return ActualLRPKeys for missing actuals", func() {
				result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)

				desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(ctx, logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				expectedSched := desiredLRP.DesiredLRPSchedulingInfo()
				Expect(result.MissingLRPKeys).To(ContainElement(&models.ActualLRPKeyWithSchedulingInfo{
					Key:            &models.ActualLRPKey{ProcessGuid: processGuid, Index: 0, Domain: domain},
					SchedulingInfo: &expectedSched,
				}))
			})
		})

		Context("and the domain is expired", func() {
			BeforeEach(func() {
				fakeClock.Increment(-10 * time.Second)
				Expect(sqlDB.UpsertDomain(ctx, logger, domain, 5)).To(Succeed())
				fakeClock.Increment(10 * time.Second)
			})

			It("should have the correct number of missing LRP instances", func() {
				schedulingInfos, err := sqlDB.DesiredLRPSchedulingInfos(ctx, logger, models.DesiredLRPFilter{ProcessGuids: []string{processGuid}})
				Expect(err).NotTo(HaveOccurred())

				Expect(schedulingInfos).To(HaveLen(1))

				result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				Expect(result.MissingLRPKeys).To(ConsistOf(&models.ActualLRPKeyWithSchedulingInfo{
					Key:            &models.ActualLRPKey{ProcessGuid: processGuid, Index: 0, Domain: domain},
					SchedulingInfo: schedulingInfos[0],
				}))
			})

			It("return ActualLRPKeys for missing actuals", func() {
				result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)

				desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(ctx, logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				expectedSched := desiredLRP.DesiredLRPSchedulingInfo()
				Expect(result.MissingLRPKeys).To(ContainElement(&models.ActualLRPKeyWithSchedulingInfo{
					Key:            &models.ActualLRPKey{ProcessGuid: processGuid, Index: 0, Domain: domain},
					SchedulingInfo: &expectedSched,
				}))
			})
		})
	})

	Context("when the ActualLRPs are crashed and restartable", func() {
		var (
			domain      string
			processGuid string
		)

		BeforeEach(func() {
			processGuid = "desired-with-restartable-crashed-actuals" + "-" + domain
			desiredLRPWithRestartableCrashedActuals := model_helpers.NewValidDesiredLRP(processGuid)
			desiredLRPWithRestartableCrashedActuals.Domain = domain
			desiredLRPWithRestartableCrashedActuals.Instances = 2
			err := sqlDB.DesireLRP(ctx, logger, desiredLRPWithRestartableCrashedActuals)
			Expect(err).NotTo(HaveOccurred())

			for i := int32(0); i < 2; i++ {
				crashedActualLRPKey := models.NewActualLRPKey(processGuid, i, domain)
				_, err = sqlDB.CreateUnclaimedActualLRP(ctx, logger, &crashedActualLRPKey)
				Expect(err).NotTo(HaveOccurred())
				instanceGuid := "restartable-crashed-actual" + "-" + domain
				_, _, err = sqlDB.ClaimActualLRP(ctx, logger, processGuid, i, &models.ActualLRPInstanceKey{InstanceGuid: instanceGuid, CellId: "existing-cell"})
				Expect(err).NotTo(HaveOccurred())
				actualLRPNetInfo := models.NewActualLRPNetInfo("some-address", "container-address", models.ActualLRPNetInfo_PreferredAddressUnknown, models.NewPortMapping(2222, 4444))
				_, _, err = sqlDB.StartActualLRP(ctx, logger, &crashedActualLRPKey, &models.ActualLRPInstanceKey{InstanceGuid: instanceGuid, CellId: "existing-cell"}, &actualLRPNetInfo, model_helpers.NewActualLRPInternalRoutes(), model_helpers.NewActualLRPMetricTags(), false, "some-zone")
				Expect(err).NotTo(HaveOccurred())
				_, _, _, err = sqlDB.CrashActualLRP(ctx, logger, &crashedActualLRPKey, &models.ActualLRPInstanceKey{InstanceGuid: instanceGuid, CellId: "existing-cell"}, "whatever")
				Expect(err).NotTo(HaveOccurred())
			}

			// we cannot use CrashedActualLRPs, otherwise it will transition the LRP
			// to unclaimed since ShouldRestartCrash will return true on the first
			// crash
			queryStr := `
				UPDATE actual_lrps
				SET state = ?
			`
			if test_helpers.UsePostgres() {
				queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
			}
			_, err = db.ExecContext(ctx, queryStr, models.ActualLRPStateCrashed)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the domain is fresh", func() {
			BeforeEach(func() {
				Expect(sqlDB.UpsertDomain(ctx, logger, domain, 5)).To(Succeed())
			})

			It("should have the correct number of crashed LRP instances", func() {
				sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				_, _, _, crashed, _ := sqlDB.CountActualLRPsByState(ctx, logger)
				Expect(crashed).To(Equal(2))
			})

			It("add the keys to UnstartedLRPKeys", func() {
				result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)

				desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(ctx, logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				expectedSched := desiredLRP.DesiredLRPSchedulingInfo()
				Expect(result.UnstartedLRPKeys).To(ContainElement(&models.ActualLRPKeyWithSchedulingInfo{
					Key:            &models.ActualLRPKey{ProcessGuid: processGuid, Index: 0, Domain: domain},
					SchedulingInfo: &expectedSched,
				}))
			})
		})

		Context("when the domain is expired", func() {
			BeforeEach(func() {
				fakeClock.Increment(-10 * time.Second)
				Expect(sqlDB.UpsertDomain(ctx, logger, domain, 5)).To(Succeed())
				fakeClock.Increment(10 * time.Second)
			})

			It("should have the correct number of crashed LRP instances", func() {
				sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				_, _, _, crashed, _ := sqlDB.CountActualLRPsByState(ctx, logger)
				Expect(crashed).To(Equal(2))
			})

			It("add the keys to UnstartedLRPKeys", func() {
				result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)

				desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(ctx, logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				expectedSched := desiredLRP.DesiredLRPSchedulingInfo()
				Expect(result.UnstartedLRPKeys).To(ContainElement(&models.ActualLRPKeyWithSchedulingInfo{
					Key:            &models.ActualLRPKey{ProcessGuid: processGuid, Index: 0, Domain: domain},
					SchedulingInfo: &expectedSched,
				}))
			})
		})

		Context("when the the lrps are evacuating", func() {
			BeforeEach(func() {
				Expect(sqlDB.UpsertDomain(ctx, logger, domain, 5)).To(Succeed())

				queryStr := `UPDATE actual_lrps SET presence = ?`
				if test_helpers.UsePostgres() {
					queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
				}
				_, err := db.ExecContext(ctx, queryStr, models.ActualLRP_Evacuating)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should have the correct number of missing LRP instances", func() {
				schedulingInfos, err := sqlDB.DesiredLRPSchedulingInfos(ctx, logger, models.DesiredLRPFilter{ProcessGuids: []string{processGuid}})
				Expect(err).NotTo(HaveOccurred())

				Expect(schedulingInfos).To(HaveLen(1))

				result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				Expect(result.MissingLRPKeys).To(ConsistOf(
					&models.ActualLRPKeyWithSchedulingInfo{
						Key:            &models.ActualLRPKey{ProcessGuid: processGuid, Index: 0, Domain: domain},
						SchedulingInfo: schedulingInfos[0],
					},
					&models.ActualLRPKeyWithSchedulingInfo{
						Key:            &models.ActualLRPKey{ProcessGuid: processGuid, Index: 1, Domain: domain},
						SchedulingInfo: schedulingInfos[0],
					},
				))
			})

			// it is the responsibility of the caller to create new LRPs
			It("prune the evacuating LRPs and does not create new ones", func() {
				sqlDB.ConvergeLRPs(ctx, logger, cellSet)

				lrps, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: processGuid})
				Expect(err).NotTo(HaveOccurred())
				Expect(lrps).To(BeEmpty())
			})

			It("return ActualLRPKeys for missing actuals", func() {
				result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)

				desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(ctx, logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				expectedSched := desiredLRP.DesiredLRPSchedulingInfo()
				Expect(result.MissingLRPKeys).To(ContainElement(&models.ActualLRPKeyWithSchedulingInfo{
					Key:            &models.ActualLRPKey{ProcessGuid: processGuid, Index: 0, Domain: domain},
					SchedulingInfo: &expectedSched,
				}))
			})

			It("return ActualLRPRemovedEvent for the removed evacuating LRPs", func() {
				actualLRPs, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: processGuid})
				Expect(err).NotTo(HaveOccurred())
				Expect(actualLRPs).To(HaveLen(2))

				result := sqlDB.ConvergeLRPs(context.WithValue(ctx, trace.RequestIdHeaderCtxKey, traceId), logger, cellSet)
				Expect(result.Events).To(ConsistOf(
					//lint:ignore SA1019 - still need to emit these events until the ActaulLRPGroup api is deleted
					models.NewActualLRPRemovedEvent(actualLRPs[0].ToActualLRPGroup()),
					//lint:ignore SA1019 - still need to emit these events until the ActaulLRPGroup api is deleted
					models.NewActualLRPRemovedEvent(actualLRPs[1].ToActualLRPGroup()),
				))
				Expect(result.InstanceEvents).To(ConsistOf(
					models.NewActualLRPInstanceRemovedEvent(actualLRPs[0], traceId),
					models.NewActualLRPInstanceRemovedEvent(actualLRPs[1], traceId),
				))
			})
		})
	})

	Context("when the ActualLRPs are crashed and non-restartable", func() {
		var (
			domain      string
			processGuid string
		)

		BeforeEach(func() {
			processGuid = "desired-with-non-restartable-crashed-actuals" + "-" + domain
			desiredLRPWithRestartableCrashedActuals := model_helpers.NewValidDesiredLRP(processGuid)
			desiredLRPWithRestartableCrashedActuals.Domain = domain
			desiredLRPWithRestartableCrashedActuals.Instances = 2
			err := sqlDB.DesireLRP(ctx, logger, desiredLRPWithRestartableCrashedActuals)
			Expect(err).NotTo(HaveOccurred())

			for i := int32(0); i < 2; i++ {
				crashedActualLRPKey := models.NewActualLRPKey(processGuid, i, domain)
				_, err = sqlDB.CreateUnclaimedActualLRP(ctx, logger, &crashedActualLRPKey)
				Expect(err).NotTo(HaveOccurred())
				instanceGuid := "restartable-crashed-actual" + "-" + domain
				_, _, err = sqlDB.ClaimActualLRP(ctx, logger, processGuid, i, &models.ActualLRPInstanceKey{InstanceGuid: instanceGuid, CellId: "existing-cell"})
				Expect(err).NotTo(HaveOccurred())
				actualLRPNetInfo := models.NewActualLRPNetInfo("some-address", "container-address", models.ActualLRPNetInfo_PreferredAddressUnknown, models.NewPortMapping(2222, 4444))
				_, _, err = sqlDB.StartActualLRP(ctx, logger, &crashedActualLRPKey, &models.ActualLRPInstanceKey{InstanceGuid: instanceGuid, CellId: "existing-cell"}, &actualLRPNetInfo, model_helpers.NewActualLRPInternalRoutes(), model_helpers.NewActualLRPMetricTags(), false, "some-zone")
				Expect(err).NotTo(HaveOccurred())
			}

			// we cannot use CrashedActualLRPs, otherwise it will transition the LRP
			// to unclaimed since ShouldRestartCrash will return true on the first
			// crash
			queryStr := `
			UPDATE actual_lrps
			SET crash_count = ?, state = ?
			`
			if test_helpers.UsePostgres() {
				queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
			}
			_, err = db.ExecContext(ctx, queryStr, models.DefaultMaxRestarts+1, models.ActualLRPStateCrashed)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the domain is fresh", func() {
			BeforeEach(func() {
				Expect(sqlDB.UpsertDomain(ctx, logger, domain, 5)).To(Succeed())
			})

			It("should have the correct number of crashed LRP instances", func() {
				sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				_, _, _, crashed, _ := sqlDB.CountActualLRPsByState(ctx, logger)
				Expect(crashed).To(Equal(2))
			})

			It("returns an empty convergence result", func() {
				result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				Expect(result).To(BeZero())
			})
		})

		Context("when the domain is expired", func() {
			BeforeEach(func() {
				fakeClock.Increment(-10 * time.Second)
				Expect(sqlDB.UpsertDomain(ctx, logger, domain, 5)).To(Succeed())
				fakeClock.Increment(10 * time.Second)
			})

			It("should have the correct number of crashed LRP instances", func() {
				sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				_, _, _, crashed, _ := sqlDB.CountActualLRPsByState(ctx, logger)
				Expect(crashed).To(Equal(2))
			})

			It("does not add the keys to UnstartedLRPKeys", func() {
				result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				Expect(result.UnstartedLRPKeys).To(BeEmpty())
			})
		})
	})

	Context("there is an ActualLRP without a corresponding DesiredLRP", func() {
		var (
			processGuid, domain string
		)

		BeforeEach(func() {
			domain = "some-domain"
			processGuid = "actual-with-no-desired"
			actualLRPWithNoDesired := &models.ActualLRPKey{ProcessGuid: processGuid, Index: 0, Domain: domain}
			_, err := sqlDB.CreateUnclaimedActualLRP(ctx, logger, actualLRPWithNoDesired)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the domain is fresh", func() {
			BeforeEach(func() {
				Expect(sqlDB.UpsertDomain(ctx, logger, domain, 5)).To(Succeed())
			})

			It("returns extra ActualLRPs to be retired", func() {
				result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				keysToRetire := result.KeysToRetire

				actualLRPKey := models.ActualLRPKey{ProcessGuid: processGuid, Index: 0, Domain: domain}
				Expect(keysToRetire).To(ContainElement(&actualLRPKey))
			})

			It("returns the no lrp keys to be started", func() {
				result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				Expect(result.UnstartedLRPKeys).To(BeEmpty())
				Expect(result.MissingLRPKeys).To(BeEmpty())
			})

			It("does not touch the ActualLRPs in the database", func() {
				lrpsBefore, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: processGuid})
				Expect(err).NotTo(HaveOccurred())

				sqlDB.ConvergeLRPs(ctx, logger, cellSet)

				lrpsAfter, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: processGuid})
				Expect(err).NotTo(HaveOccurred())

				Expect(lrpsAfter).To(Equal(lrpsBefore))
			})

			It("should have the correct number of extra LRP instances", func() {
				results := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				Expect(results.KeysToRetire).To(ConsistOf(&models.ActualLRPKey{ProcessGuid: processGuid, Index: 0, Domain: domain}))
			})
		})

		Context("when the domain is expired", func() {
			BeforeEach(func() {
				fakeClock.Increment(-10 * time.Second)
				Expect(sqlDB.UpsertDomain(ctx, logger, domain, 5)).To(Succeed())
				fakeClock.Increment(10 * time.Second)
			})

			It("does not return extra ActualLRPs to be retired", func() {
				result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				Expect(result.KeysToRetire).To(BeEmpty())
			})

			It("returns the no lrp keys to be started", func() {
				result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				Expect(result.UnstartedLRPKeys).To(BeEmpty())
				Expect(result.MissingLRPKeys).To(BeEmpty())
			})

			It("does not touch the ActualLRPs in the database", func() {
				lrpsBefore, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: processGuid})
				Expect(err).NotTo(HaveOccurred())

				sqlDB.ConvergeLRPs(ctx, logger, cellSet)

				lrpsAfter, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: processGuid})
				Expect(err).NotTo(HaveOccurred())

				Expect(lrpsAfter).To(Equal(lrpsBefore))
			})

			It("should not have any extra LRP instances", func() {
				results := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				Expect(results.KeysToRetire).To(BeEmpty())
			})
		})

		Context("when the the lrps are evacuating", func() {
			BeforeEach(func() {
				Expect(sqlDB.UpsertDomain(ctx, logger, domain, 5)).To(Succeed())

				queryStr := `UPDATE actual_lrps SET presence = ?`
				if test_helpers.UsePostgres() {
					queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
				}
				_, err := db.ExecContext(ctx, queryStr, models.ActualLRP_Evacuating)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns the no lrp keys to be started", func() {
				result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				Expect(result.UnstartedLRPKeys).To(BeEmpty())
				Expect(result.MissingLRPKeys).To(BeEmpty())
			})

			It("removes the evacuating LRPs", func() {
				sqlDB.ConvergeLRPs(ctx, logger, cellSet)

				actualLRPs, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: processGuid})
				Expect(err).NotTo(HaveOccurred())
				Expect(actualLRPs).To(BeEmpty())
			})

			It("return an ActualLRPRemoved Event", func() {
				actualLRPs, err := sqlDB.ActualLRPs(ctx, logger, models.ActualLRPFilter{ProcessGuid: processGuid})
				Expect(err).NotTo(HaveOccurred())

				Expect(actualLRPs).To(HaveLen(1))

				result := sqlDB.ConvergeLRPs(context.WithValue(ctx, trace.RequestIdHeaderCtxKey, traceId), logger, cellSet)
				//lint:ignore SA1019 - still need to emit these events until the ActaulLRPGroup api is deleted
				Expect(result.Events).To(ConsistOf(models.NewActualLRPRemovedEvent(actualLRPs[0].ToActualLRPGroup())))
				Expect(result.InstanceEvents).To(ConsistOf(models.NewActualLRPInstanceRemovedEvent(actualLRPs[0], traceId)))
			})

			It("should not have any extra LRP instances", func() {
				results := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
				Expect(results.KeysToRetire).To(BeEmpty())
			})
		})
	})

	Context("when there are actual LRPs with internal routes different from desired LRP internal routes", func() {
		var (
			processGuid, domain                               string
			lrpKey1, lrpKey2, lrpKey3                         models.ActualLRPKey
			lrpInstanceKey1, lrpInstanceKey2, lrpInstanceKey3 models.ActualLRPInstanceKey
		)

		BeforeEach(func() {
			domain = "some-domain"
			processGuid = "desired-with-different-internal-routes"
			desiredLRP := model_helpers.NewValidDesiredLRP(processGuid)
			desiredLRP.Domain = domain
			desiredLRP.Instances = 5
			err := sqlDB.DesireLRP(ctx, logger, desiredLRP)
			Expect(err).NotTo(HaveOccurred())

			actualLRPNetInfo := models.NewActualLRPNetInfo("some-address", "container-address", models.ActualLRPNetInfo_PreferredAddressUnknown, models.NewPortMapping(2222, 4444))
			lrpKey1 = models.NewActualLRPKey(processGuid, 0, domain)
			lrpInstanceKey1 = models.ActualLRPInstanceKey{InstanceGuid: "ig-1", CellId: "existing-cell"}
			_, _, err = sqlDB.StartActualLRP(ctx, logger, &lrpKey1, &lrpInstanceKey1, &actualLRPNetInfo, model_helpers.NewActualLRPInternalRoutes(), model_helpers.NewActualLRPMetricTags(), false, "some-zone")
			Expect(err).NotTo(HaveOccurred())

			lrpKey2 = models.NewActualLRPKey(processGuid, 1, domain)
			lrpInstanceKey2 = models.ActualLRPInstanceKey{InstanceGuid: "ig-2", CellId: "existing-cell"}
			_, _, err = sqlDB.StartActualLRP(ctx, logger, &lrpKey2, &lrpInstanceKey2, &actualLRPNetInfo, model_helpers.NewActualLRPInternalRoutes(), model_helpers.NewActualLRPMetricTags(), false, "some-zone")
			Expect(err).NotTo(HaveOccurred())

			lrpKey3 = models.NewActualLRPKey(processGuid, 2, domain)
			lrpInstanceKey3 = models.ActualLRPInstanceKey{InstanceGuid: "ig-3", CellId: "existing-cell"}
			_, _, err = sqlDB.StartActualLRP(ctx, logger, &lrpKey3, &lrpInstanceKey3, &actualLRPNetInfo, nil, nil, false, "some-zone")
			Expect(err).NotTo(HaveOccurred())

			queryStr := "UPDATE actual_lrps SET internal_routes = ? WHERE process_guid = ? AND instance_index = ?;"
			if test_helpers.UsePostgres() {
				queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
			}
			_, err = db.ExecContext(ctx, queryStr, nil, lrpKey3.ProcessGuid, lrpKey3.Index)
			Expect(err).NotTo(HaveOccurred())

			lrpKey4 := models.NewActualLRPKey(processGuid, 3, domain)
			lrpInstanceKey4 := models.ActualLRPInstanceKey{InstanceGuid: "ig-4", CellId: "existing-cell"}
			sameInternalRoutes := []*models.ActualLRPInternalRoute{
				{Hostname: "some-internal-route.apps.internal"},
				{Hostname: "some-other-internal-route.apps.internal"},
			}
			_, _, err = sqlDB.StartActualLRP(ctx, logger, &lrpKey4, &lrpInstanceKey4, &actualLRPNetInfo, sameInternalRoutes, model_helpers.NewActualLRPMetricTags(), false, "some-zone")
			Expect(err).NotTo(HaveOccurred())

			lrpKey5 := models.NewActualLRPKey(processGuid, 4, domain)
			lrpInstanceKey5 := models.ActualLRPInstanceKey{InstanceGuid: "ig-5", CellId: "existing-cell"}
			internalRoutesInDifferentOrder := []*models.ActualLRPInternalRoute{
				{Hostname: "some-other-internal-route.apps.internal"},
				{Hostname: "some-internal-route.apps.internal"},
			}
			_, _, err = sqlDB.StartActualLRP(ctx, logger, &lrpKey5, &lrpInstanceKey5, &actualLRPNetInfo, internalRoutesInDifferentOrder, model_helpers.NewActualLRPMetricTags(), false, "some-zone")
			Expect(err).NotTo(HaveOccurred())

			rawInternalRoutes := json.RawMessage(`[{"hostname":"some-internal-route.apps.internal"},{"hostname":"some-other-internal-route.apps.internal"}]`)
			update := models.DesiredLRPUpdate{
				Routes: &models.Routes{internalroutes.INTERNAL_ROUTER: &rawInternalRoutes},
			}
			_, err = sqlDB.UpdateDesiredLRP(ctx, logger, processGuid, &update)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the LRP keys with changed internal routes", func() {
			result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
			desiredInternalRoutes := internalroutes.InternalRoutes{
				{Hostname: "some-internal-route.apps.internal"},
				{Hostname: "some-other-internal-route.apps.internal"},
			}
			lrpKeyWithInternalRoutes1 := bbsdb.ActualLRPKeyWithInternalRoutes{Key: &lrpKey1, InstanceKey: &lrpInstanceKey1, DesiredInternalRoutes: desiredInternalRoutes}
			lrpKeyWithInternalRoutes2 := bbsdb.ActualLRPKeyWithInternalRoutes{Key: &lrpKey2, InstanceKey: &lrpInstanceKey2, DesiredInternalRoutes: desiredInternalRoutes}
			lrpKeyWithInternalRoutes3 := bbsdb.ActualLRPKeyWithInternalRoutes{Key: &lrpKey3, InstanceKey: &lrpInstanceKey3, DesiredInternalRoutes: desiredInternalRoutes}

			Expect(result.KeysWithInternalRouteChanges).To(ConsistOf(&lrpKeyWithInternalRoutes1, &lrpKeyWithInternalRoutes2, &lrpKeyWithInternalRoutes3))
		})
	})

	Context("when there are actual LRPs with non-nil metrics tags different from desired LRP metric tags", func() {
		var (
			processGuid, domain                               string
			lrpKey1, lrpKey2, lrpKey3                         models.ActualLRPKey
			lrpInstanceKey1, lrpInstanceKey2, lrpInstanceKey3 models.ActualLRPInstanceKey
		)

		BeforeEach(func() {
			domain = "some-domain"
			processGuid = "desired-with-different-metric-tags"
			desiredLRP := model_helpers.NewValidDesiredLRP(processGuid)
			desiredLRP.Domain = domain
			desiredLRP.Instances = 4
			err := sqlDB.DesireLRP(ctx, logger, desiredLRP)
			Expect(err).NotTo(HaveOccurred())

			actualLRPNetInfo := models.NewActualLRPNetInfo("some-address", "container-address", models.ActualLRPNetInfo_PreferredAddressUnknown, models.NewPortMapping(2222, 4444))
			lrpKey1 = models.NewActualLRPKey(processGuid, 0, domain)
			lrpInstanceKey1 = models.ActualLRPInstanceKey{InstanceGuid: "ig-1", CellId: "existing-cell"}
			_, _, err = sqlDB.StartActualLRP(ctx, logger, &lrpKey1, &lrpInstanceKey1, &actualLRPNetInfo, model_helpers.NewActualLRPInternalRoutes(), model_helpers.NewActualLRPMetricTags(), false, "some-zone")
			Expect(err).NotTo(HaveOccurred())

			lrpKey2 = models.NewActualLRPKey(processGuid, 1, domain)
			lrpInstanceKey2 = models.ActualLRPInstanceKey{InstanceGuid: "ig-2", CellId: "existing-cell"}
			_, _, err = sqlDB.StartActualLRP(ctx, logger, &lrpKey2, &lrpInstanceKey2, &actualLRPNetInfo, model_helpers.NewActualLRPInternalRoutes(), model_helpers.NewActualLRPMetricTags(), false, "some-zone")
			Expect(err).NotTo(HaveOccurred())

			lrpKey3 = models.NewActualLRPKey(processGuid, 2, domain)
			lrpInstanceKey3 = models.ActualLRPInstanceKey{InstanceGuid: "ig-3", CellId: "existing-cell"}
			_, _, err = sqlDB.StartActualLRP(ctx, logger, &lrpKey3, &lrpInstanceKey3, &actualLRPNetInfo, model_helpers.NewActualLRPInternalRoutes(), nil, false, "some-zone")
			Expect(err).NotTo(HaveOccurred())

			lrpKey4 := models.NewActualLRPKey(processGuid, 3, domain)
			lrpInstanceKey4 := models.ActualLRPInstanceKey{InstanceGuid: "ig-4", CellId: "existing-cell"}
			sameMetricTags := map[string]string{
				"app_name": "some-app-renamed",
			}

			_, _, err = sqlDB.StartActualLRP(ctx, logger, &lrpKey4, &lrpInstanceKey4, &actualLRPNetInfo, model_helpers.NewActualLRPInternalRoutes(), sameMetricTags, false, "some-zone")
			Expect(err).NotTo(HaveOccurred())

			update := models.DesiredLRPUpdate{
				MetricTags: map[string]*models.MetricTagValue{
					"app_name": {Static: "some-app-renamed"},
				},
			}
			_, err = sqlDB.UpdateDesiredLRP(ctx, logger, processGuid, &update)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the LRP keys with changed metric tags", func() {
			result := sqlDB.ConvergeLRPs(ctx, logger, cellSet)
			desiredMetricTags := map[string]*models.MetricTagValue{
				"app_name": {Static: "some-app-renamed"},
			}
			lrpKeyWithMetricTags1 := bbsdb.ActualLRPKeyWithMetricTags{Key: &lrpKey1, InstanceKey: &lrpInstanceKey1, DesiredMetricTags: desiredMetricTags}
			lrpKeyWithMetricTags2 := bbsdb.ActualLRPKeyWithMetricTags{Key: &lrpKey2, InstanceKey: &lrpInstanceKey2, DesiredMetricTags: desiredMetricTags}

			Expect(result.KeysWithMetricTagChanges).To(ConsistOf(&lrpKeyWithMetricTags1, &lrpKeyWithMetricTags2))
		})
	})
})
