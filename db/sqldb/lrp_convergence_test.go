package sqldb_test

import (
	"fmt"
	"time"

	"code.cloudfoundry.org/bbs/db/sqldb"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/models/test/model_helpers"
	"code.cloudfoundry.org/bbs/test_helpers"
	"code.cloudfoundry.org/lager/lagertest"

	mfakes "code.cloudfoundry.org/diego-logging-client/testhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = FDescribe("New LRPConvergence", func() {
	type event struct {
		name  string
		value int
	}

	getMetricsEmitted := func(metronClient *mfakes.FakeIngressClient) func() []event {
		return func() []event {
			var events []event
			for i := 0; i < metronClient.SendMetricCallCount(); i++ {
				name, value, _ := metronClient.SendMetricArgsForCall(i)
				events = append(events, event{
					name:  name,
					value: value,
				})
			}
			return events
		}
	}

	actualLRPKeyWithSchedulingInfo := func(desiredLRP *models.DesiredLRP, index int) *models.ActualLRPKeyWithSchedulingInfo {
		schedulingInfo := desiredLRP.DesiredLRPSchedulingInfo()
		lrpKey := models.NewActualLRPKey(desiredLRP.ProcessGuid, int32(index), desiredLRP.Domain)

		lrp := &models.ActualLRPKeyWithSchedulingInfo{
			Key:            &lrpKey,
			SchedulingInfo: &schedulingInfo,
		}
		return lrp
	}

	FContext("when there are fresh domains", func() {
		BeforeEach(func() {
			Expect(sqlDB.UpsertDomain(logger, "some-domain", 5)).To(Succeed())
			Expect(sqlDB.UpsertDomain(logger, "other-domain", 5)).To(Succeed())
		})

		It("emits domain freshness metric for each domain", func() {
			sqlDB.ConvergeLRPs(logger, models.NewCellSet())

			Eventually(getMetricsEmitted(fakeMetronClient)).Should(ContainElement(event{
				name:  "Domain.some-domain",
				value: 1,
			}))

			Eventually(getMetricsEmitted(fakeMetronClient)).Should(ContainElement(event{
				name:  "Domain.other-domain",
				value: 1,
			}))
		})
	})

	FContext("when there are unclaimed LRPs", func() {
		var (
			domain      string
			processGuid string
			cellSet     models.CellSet
		)

		BeforeEach(func() {
			domain = "some-domain"
			processGuid = "desired-with-unclaimed-actuals"
			cellSet = models.NewCellSetFromList([]*models.CellPresence{
				{CellId: "existing-cell"},
			})
			desiredLRPWithStaleActuals := model_helpers.NewValidDesiredLRP(processGuid)
			desiredLRPWithStaleActuals.Domain = domain
			desiredLRPWithStaleActuals.Instances = 1
			err := sqlDB.DesireLRP(logger, desiredLRPWithStaleActuals)
			Expect(err).NotTo(HaveOccurred())
			_, err = sqlDB.CreateUnclaimedActualLRP(logger, &models.ActualLRPKey{ProcessGuid: processGuid, Index: 0, Domain: domain})
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the domain is fresh", func() {
			BeforeEach(func() {
				Expect(sqlDB.UpsertDomain(logger, domain, 5)).To(Succeed())
			})

			It("does not touch the ActualLRPs in the database", func() {
				groupsBefore, err := sqlDB.ActualLRPGroupsByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				sqlDB.ConvergeLRPs(logger, cellSet)

				groupsAfter, err := sqlDB.ActualLRPGroupsByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				Expect(groupsAfter).To(Equal(groupsBefore))
			})
		})
	})

	FContext("when there are claimed LRPs", func() {
		var (
			domain      string
			processGuid string
			cellSet     models.CellSet
		)

		BeforeEach(func() {
			domain = "some-domain"
			processGuid = "desired-with-claimed-actuals"
			cellSet = models.NewCellSetFromList([]*models.CellPresence{
				{CellId: "existing-cell"},
			})
			desiredLRPWithStaleActuals := model_helpers.NewValidDesiredLRP(processGuid)
			desiredLRPWithStaleActuals.Domain = domain
			desiredLRPWithStaleActuals.Instances = 1
			err := sqlDB.DesireLRP(logger, desiredLRPWithStaleActuals)
			Expect(err).NotTo(HaveOccurred())
			_, err = sqlDB.CreateUnclaimedActualLRP(logger, &models.ActualLRPKey{ProcessGuid: processGuid, Index: 0, Domain: domain})
			Expect(err).NotTo(HaveOccurred())
			_, _, err = sqlDB.ClaimActualLRP(logger, processGuid, 0, &models.ActualLRPInstanceKey{InstanceGuid: "instance-guid", CellId: "existing-cell"})
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the domain is fresh", func() {
			BeforeEach(func() {
				Expect(sqlDB.UpsertDomain(logger, domain, 5)).To(Succeed())
			})

			It("does not touch the ActualLRPs in the database", func() {
				groupsBefore, err := sqlDB.ActualLRPGroupsByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				sqlDB.ConvergeLRPs(logger, cellSet)

				groupsAfter, err := sqlDB.ActualLRPGroupsByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				Expect(groupsAfter).To(Equal(groupsBefore))
			})
		})
	})

	Context("when there are stale unclaimed LRPs", func() {
		var (
			domain      string
			processGuid string
			cellSet     models.CellSet
		)

		BeforeEach(func() {
			domain = "some-domain"
			processGuid = "desired-with-stale-actuals"
			cellSet = models.NewCellSetFromList([]*models.CellPresence{
				{CellId: "existing-cell"},
			})
			desiredLRPWithStaleActuals := model_helpers.NewValidDesiredLRP(processGuid)
			desiredLRPWithStaleActuals.Domain = domain
			desiredLRPWithStaleActuals.Instances = 2
			err := sqlDB.DesireLRP(logger, desiredLRPWithStaleActuals)
			Expect(err).NotTo(HaveOccurred())
			fakeClock.Increment(-models.StaleUnclaimedActualLRPDuration)
			_, err = sqlDB.CreateUnclaimedActualLRP(logger, &models.ActualLRPKey{ProcessGuid: processGuid, Index: 0, Domain: domain})
			Expect(err).NotTo(HaveOccurred())
			_, err = sqlDB.CreateUnclaimedActualLRP(logger, &models.ActualLRPKey{ProcessGuid: processGuid, Index: 1, Domain: domain})
			Expect(err).NotTo(HaveOccurred())
			fakeClock.Increment(models.StaleUnclaimedActualLRPDuration + 2)
		})

		Context("when the domain is fresh", func() {
			BeforeEach(func() {
				Expect(sqlDB.UpsertDomain(logger, domain, 5)).To(Succeed())
			})

			It("returns start requests", func() {
				result := sqlDB.ConvergeLRPs(logger, cellSet)
				unstartedLRPKeys := result.UnstartedLRPKeys
				Expect(unstartedLRPKeys).NotTo(BeEmpty())
				Expect(logger).To(gbytes.Say("creating-start-request.*reason\":\"stale-unclaimed-lrp"))

				desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				Expect(unstartedLRPKeys).To(ContainElement(actualLRPKeyWithSchedulingInfo(desiredLRP, 0)))
				Expect(unstartedLRPKeys).To(ContainElement(actualLRPKeyWithSchedulingInfo(desiredLRP, 1)))
			})

			It("does not touch the ActualLRPs in the database", func() {
				groupsBefore, err := sqlDB.ActualLRPGroupsByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				sqlDB.ConvergeLRPs(logger, cellSet)

				groupsAfter, err := sqlDB.ActualLRPGroupsByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				Expect(groupsAfter).To(Equal(groupsBefore))
			})

			It("emits stale unclaimed LRP metrics", func() {
				sqlDB.ConvergeLRPs(logger, models.NewCellSetFromList(nil))

				Eventually(getMetricsEmitted(fakeMetronClient)).Should(ContainElement(event{
					name:  "LRPsUnclaimed",
					value: 2,
				}))
			})
		})

		Context("when the domain is expired", func() {
			BeforeEach(func() {
				fakeClock.Increment(-10 * time.Second)
				Expect(sqlDB.UpsertDomain(logger, domain, 5)).To(Succeed())
				fakeClock.Increment(10 * time.Second)
			})

			It("returns start requests", func() {
				result := sqlDB.ConvergeLRPs(logger, cellSet)
				unstartedLRPKeys := result.UnstartedLRPKeys
				Expect(unstartedLRPKeys).NotTo(BeEmpty())
				Expect(logger).To(gbytes.Say("creating-start-request.*reason\":\"stale-unclaimed-lrp"))

				desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				Expect(unstartedLRPKeys).To(ContainElement(actualLRPKeyWithSchedulingInfo(desiredLRP, 0)))
				Expect(unstartedLRPKeys).To(ContainElement(actualLRPKeyWithSchedulingInfo(desiredLRP, 1)))
			})

			It("does not touch the ActualLRPs in the database", func() {
				groupsBefore, err := sqlDB.ActualLRPGroupsByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				sqlDB.ConvergeLRPs(logger, cellSet)

				groupsAfter, err := sqlDB.ActualLRPGroupsByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				Expect(groupsAfter).To(Equal(groupsBefore))
			})

			It("emits stale unclaimed LRP metrics", func() {
				sqlDB.ConvergeLRPs(logger, models.NewCellSetFromList(nil))

				Eventually(getMetricsEmitted(fakeMetronClient)).Should(ContainElement(event{
					name:  "LRPsUnclaimed",
					value: 2,
				}))
			})
		})

		Context("when the ActualLRPs presence is set to evacuating", func() {
			BeforeEach(func() {
				Expect(sqlDB.UpsertDomain(logger, domain, 5)).To(Succeed())

				queryStr := `UPDATE actual_lrps SET evacuating = ?`
				if test_helpers.UsePostgres() {
					queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
				}
				_, err := db.Exec(queryStr, true)
				Expect(err).NotTo(HaveOccurred())
			})

			It("ignores the evacuating LRPs and emits LRP missing metric", func() {
				sqlDB.ConvergeLRPs(logger, models.NewCellSetFromList(nil))

				Eventually(getMetricsEmitted(fakeMetronClient)).Should(ContainElement(event{
					name:  "LRPsMissing",
					value: 2,
				}))
			})

			// it is the responsibility of the caller to create new LRPs
			It("prune the evacuating LRPs and does not create new ones", func() {
				sqlDB.ConvergeLRPs(logger, models.NewCellSetFromList(nil))

				groups, err := sqlDB.ActualLRPGroupsByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())
				Expect(groups).To(BeEmpty())
			})

			It("emits ActualLRPRemovedEvent for the removed evacuating LRPs", func() {
				groups, err := sqlDB.ActualLRPGroupsByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())
				Expect(groups).To(HaveLen(2))

				result := sqlDB.ConvergeLRPs(logger, models.NewCellSetFromList(nil))
				Expect(result.Events).To(ConsistOf(
					models.NewActualLRPRemovedEvent(groups[0]),
					models.NewActualLRPRemovedEvent(groups[1]),
				))
			})
		})
	})

	Context("when there is an ActualLRP on a missing cell", func() {
		var (
			domain      string
			processGuid string
			cellSet     models.CellSet
		)

		BeforeEach(func() {
			domain = "some-domain"
			processGuid = "desired-with-missing-cell-actuals"
			cellSet = models.NewCellSetFromList([]*models.CellPresence{
				{CellId: "existing-cell"},
			})
			desiredLRPWithMissingCellActuals := model_helpers.NewValidDesiredLRP(processGuid)
			desiredLRPWithMissingCellActuals.Domain = domain
			err := sqlDB.DesireLRP(logger, desiredLRPWithMissingCellActuals)
			Expect(err).NotTo(HaveOccurred())
			_, err = sqlDB.CreateUnclaimedActualLRP(logger, &models.ActualLRPKey{ProcessGuid: processGuid, Index: 0, Domain: domain})
			Expect(err).NotTo(HaveOccurred())
			_, _, err = sqlDB.ClaimActualLRP(logger, processGuid, 0, &models.ActualLRPInstanceKey{InstanceGuid: "actual-with-missing-cell", CellId: "other-cell"})
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the domain is fresh", func() {
			BeforeEach(func() {
				Expect(sqlDB.UpsertDomain(logger, domain, 5)).To(Succeed())
			})

			It("returns the start requests and actual lrp keys for actuals with missing cells", func() {
				result := sqlDB.ConvergeLRPs(logger, cellSet)
				keysWithMissingCells := result.KeysWithMissingCells

				desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				actualLRPGroup, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, processGuid, 0)
				Expect(err).NotTo(HaveOccurred())
				expectedSched := desiredLRP.DesiredLRPSchedulingInfo()
				Expect(keysWithMissingCells).To(ContainElement(&models.ActualLRPKeyWithSchedulingInfo{
					Key:            &actualLRPGroup.Instance.ActualLRPKey,
					SchedulingInfo: &expectedSched,
				}))
			})

			It("does not touch the ActualLRPs in the database", func() {
				groupsBefore, err := sqlDB.ActualLRPGroupsByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				sqlDB.ConvergeLRPs(logger, cellSet)

				groupsAfter, err := sqlDB.ActualLRPGroupsByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				Expect(groupsAfter).To(Equal(groupsBefore))
			})
		})

		Context("when the domain is expired", func() {
			BeforeEach(func() {
				fakeClock.Increment(-10 * time.Second)
				Expect(sqlDB.UpsertDomain(logger, domain, 5)).To(Succeed())
				fakeClock.Increment(10 * time.Second)
			})

			It("return ActualLRPKeys for actuals with missing cells", func() {
				result := sqlDB.ConvergeLRPs(logger, cellSet)
				keysWithMissingCells := result.KeysWithMissingCells

				desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				actualLRPGroup, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, processGuid, 0)
				Expect(err).NotTo(HaveOccurred())
				expectedSched := desiredLRP.DesiredLRPSchedulingInfo()
				Expect(keysWithMissingCells).To(ContainElement(&models.ActualLRPKeyWithSchedulingInfo{
					Key:            &actualLRPGroup.Instance.ActualLRPKey,
					SchedulingInfo: &expectedSched,
				}))
			})

			It("does not touch the ActualLRPs in the database", func() {
				groupsBefore, err := sqlDB.ActualLRPGroupsByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				sqlDB.ConvergeLRPs(logger, cellSet)

				groupsAfter, err := sqlDB.ActualLRPGroupsByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				Expect(groupsAfter).To(Equal(groupsBefore))
			})
		})

		Context("when the ActualLRP's presence is set to evacuating", func() {
			BeforeEach(func() {
				Expect(sqlDB.UpsertDomain(logger, domain, 5)).To(Succeed())

				queryStr := `UPDATE actual_lrps SET evacuating = ? WHERE process_guid = ?`
				if test_helpers.UsePostgres() {
					queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
				}
				_, err := db.Exec(queryStr, true, processGuid)
				Expect(err).NotTo(HaveOccurred())
			})

			It("ignores the evacuating LRPs and emits LRP missing metric", func() {
				sqlDB.ConvergeLRPs(logger, models.NewCellSet())

				Eventually(getMetricsEmitted(fakeMetronClient)).Should(ContainElement(event{
					name:  "LRPsMissing",
					value: 1,
				}))
			})

			It("removes the evacuating lrps", func() {
				sqlDB.ConvergeLRPs(logger, models.NewCellSet())

				groups, err := sqlDB.ActualLRPGroupsByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())
				Expect(groups).To(BeEmpty())
			})

			It("emits ActualLRPRemoveEvent", func() {
				groups, err := sqlDB.ActualLRPGroupsByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				Expect(groups).To(HaveLen(1))

				result := sqlDB.ConvergeLRPs(logger, models.NewCellSet())
				Expect(result.Events).To(ConsistOf(models.NewActualLRPRemovedEvent(groups[0])))
			})
		})

		It("logs the missing cells", func() {
			sqlDB.ConvergeLRPs(logger, cellSet)
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
				sqlDB.ConvergeLRPs(logger, cellSet)
				Expect(logger).ToNot(gbytes.Say("detected-missing-cells"))
			})
		})
	})

	Context("when there are extra ActualLRPs for a DesiredLRP", func() {
		var (
			domain      string
			processGuid string
			cellSet     models.CellSet
		)

		BeforeEach(func() {
			domain = "some-domain"
			processGuid = "desired-with-extra-actuals"
			cellSet = models.NewCellSetFromList([]*models.CellPresence{
				{CellId: "existing-cell"},
			})
			desiredLRPWithExtraActuals := model_helpers.NewValidDesiredLRP(processGuid)
			desiredLRPWithExtraActuals.Domain = domain
			desiredLRPWithExtraActuals.Instances = 1
			err := sqlDB.DesireLRP(logger, desiredLRPWithExtraActuals)
			Expect(err).NotTo(HaveOccurred())
			_, err = sqlDB.CreateUnclaimedActualLRP(logger, &models.ActualLRPKey{ProcessGuid: processGuid, Index: 0, Domain: domain})
			Expect(err).NotTo(HaveOccurred())
			_, err = sqlDB.CreateUnclaimedActualLRP(logger, &models.ActualLRPKey{ProcessGuid: processGuid, Index: 4, Domain: domain})
			Expect(err).NotTo(HaveOccurred())
			_, _, err = sqlDB.ClaimActualLRP(logger, processGuid, 0, &models.ActualLRPInstanceKey{InstanceGuid: "not-extra-actual", CellId: "existing-cell"})
			Expect(err).NotTo(HaveOccurred())
			_, _, err = sqlDB.ClaimActualLRP(logger, processGuid, 4, &models.ActualLRPInstanceKey{InstanceGuid: "extra-actual", CellId: "existing-cell"})
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the domain is fresh", func() {
			BeforeEach(func() {
				Expect(sqlDB.UpsertDomain(logger, domain, 5)).To(Succeed())
			})

			It("returns extra ActualLRPs to be retired", func() {
				result := sqlDB.ConvergeLRPs(logger, cellSet)
				keysToRetire := result.KeysToRetire

				actualLRPKey := models.ActualLRPKey{ProcessGuid: processGuid, Index: 4, Domain: domain}
				Expect(keysToRetire).To(ContainElement(&actualLRPKey))
			})

			It("does not touch the ActualLRPs in the database", func() {
				groupsBefore, err := sqlDB.ActualLRPGroupsByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				sqlDB.ConvergeLRPs(logger, cellSet)

				groupsAfter, err := sqlDB.ActualLRPGroupsByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				Expect(groupsAfter).To(Equal(groupsBefore))
			})

			It("emits LRPsExtra metric", func() {
				sqlDB.ConvergeLRPs(logger, models.NewCellSetFromList(nil))

				Eventually(getMetricsEmitted(fakeMetronClient)).Should(ContainElement(event{
					name:  "LRPsExtra",
					value: 1,
				}))
			})
		})

		Context("when the domain is expired", func() {
			BeforeEach(func() {
				fakeClock.Increment(-10 * time.Second)
				Expect(sqlDB.UpsertDomain(logger, domain, 5)).To(Succeed())
				fakeClock.Increment(10 * time.Second)
			})

			It("does not retire the extra lrps", func() {
				result := sqlDB.ConvergeLRPs(logger, cellSet)
				Expect(result.KeysToRetire).To(BeEmpty())
			})

			It("does not touch the ActualLRPs in the database", func() {
				groupsBefore, err := sqlDB.ActualLRPGroupsByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				sqlDB.ConvergeLRPs(logger, cellSet)

				groupsAfter, err := sqlDB.ActualLRPGroupsByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				Expect(groupsAfter).To(Equal(groupsBefore))
			})

			It("emits a zero for the LRPsExtra metric", func() {
				sqlDB.ConvergeLRPs(logger, models.NewCellSetFromList(nil))

				Eventually(getMetricsEmitted(fakeMetronClient)).Should(ContainElement(event{
					name:  "LRPsExtra",
					value: 0,
				}))
			})
		})

		Context("when the ActualLRP's presence is set to evacuating", func() {
			BeforeEach(func() {
				Expect(sqlDB.UpsertDomain(logger, domain, 5)).To(Succeed())

				queryStr := `UPDATE actual_lrps SET evacuating = ? WHERE process_guid = ?`
				if test_helpers.UsePostgres() {
					queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
				}
				_, err := db.Exec(queryStr, true, processGuid)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns the lrp keys to be started", func() {
				schedulingInfos, err := sqlDB.DesiredLRPSchedulingInfos(logger, models.DesiredLRPFilter{ProcessGuids: []string{processGuid}})
				Expect(err).NotTo(HaveOccurred())

				Expect(schedulingInfos).To(HaveLen(1))

				result := sqlDB.ConvergeLRPs(logger, models.NewCellSetFromList(nil))
				Expect(result.MissingLRPKeys).To(ConsistOf(&models.ActualLRPKeyWithSchedulingInfo{
					Key:            &models.ActualLRPKey{ProcessGuid: processGuid, Index: 0, Domain: domain},
					SchedulingInfo: schedulingInfos[0],
				}))
			})

			It("removes the evacuating LRPs", func() {
				sqlDB.ConvergeLRPs(logger, models.NewCellSetFromList(nil))

				groups, err := sqlDB.ActualLRPGroupsByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())
				Expect(groups).To(BeEmpty())
			})

			It("emits an ActualLRPRemoved Event", func() {
				groups, err := sqlDB.ActualLRPGroupsByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				Expect(groups).To(HaveLen(2))

				result := sqlDB.ConvergeLRPs(logger, models.NewCellSetFromList(nil))
				Expect(result.Events).To(ConsistOf(
					models.NewActualLRPRemovedEvent(groups[0]),
					models.NewActualLRPRemovedEvent(groups[1]),
				))
			})

			It("emits a LRPsMissing metric", func() {
				sqlDB.ConvergeLRPs(logger, models.NewCellSetFromList(nil))

				Eventually(getMetricsEmitted(fakeMetronClient)).Should(ContainElement(event{
					name:  "LRPsMissing",
					value: 1,
				}))
			})
		})
	})

	FContext("when there are no ActualLRPs for a DesiredLRP", func() {
		var (
			domain      string
			processGuid string
		)

		BeforeEach(func() {
			processGuid = "desired-with-missing-all-actuals" + "-" + domain
			desiredLRPWithMissingAllActuals := model_helpers.NewValidDesiredLRP(processGuid)
			desiredLRPWithMissingAllActuals.Domain = domain
			desiredLRPWithMissingAllActuals.Instances = 1
			err := sqlDB.DesireLRP(logger, desiredLRPWithMissingAllActuals)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("and the domain is fresh", func() {
			BeforeEach(func() {
				Expect(sqlDB.UpsertDomain(logger, domain, 5)).To(Succeed())
			})

			It("emits a LRPsMissing metric", func() {
				sqlDB.ConvergeLRPs(logger, models.NewCellSetFromList(nil))

				Eventually(getMetricsEmitted(fakeMetronClient)).Should(ContainElement(event{
					name:  "LRPsMissing",
					value: 1,
				}))
			})

			It("return ActualLRPKeys for actuals with missing cells", func() {
				result := sqlDB.ConvergeLRPs(logger, models.NewCellSet())

				desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(logger, processGuid)
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
				Expect(sqlDB.UpsertDomain(logger, domain, 5)).To(Succeed())
				fakeClock.Increment(10 * time.Second)
			})

			It("emits a LRPsMissing metric", func() {
				sqlDB.ConvergeLRPs(logger, models.NewCellSetFromList(nil))

				Eventually(getMetricsEmitted(fakeMetronClient)).Should(ContainElement(event{
					name:  "LRPsMissing",
					value: 1,
				}))
			})

			It("return ActualLRPKeys for actuals with missing cells", func() {
				result := sqlDB.ConvergeLRPs(logger, models.NewCellSet())

				desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(logger, processGuid)
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
			err := sqlDB.DesireLRP(logger, desiredLRPWithRestartableCrashedActuals)
			Expect(err).NotTo(HaveOccurred())

			for i := int32(0); i < 2; i++ {
				crashedActualLRPKey := models.NewActualLRPKey(processGuid, i, domain)
				_, err = sqlDB.CreateUnclaimedActualLRP(logger, &crashedActualLRPKey)
				Expect(err).NotTo(HaveOccurred())
				instanceGuid := "restartable-crashed-actual" + "-" + domain
				_, _, err = sqlDB.ClaimActualLRP(logger, processGuid, i, &models.ActualLRPInstanceKey{InstanceGuid: instanceGuid, CellId: "existing-cell"})
				Expect(err).NotTo(HaveOccurred())
				actualLRPNetInfo := models.NewActualLRPNetInfo("some-address", "container-address", models.NewPortMapping(2222, 4444))
				_, _, err = sqlDB.StartActualLRP(logger, &crashedActualLRPKey, &models.ActualLRPInstanceKey{InstanceGuid: instanceGuid, CellId: "existing-cell"}, &actualLRPNetInfo)
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
			_, err = db.Exec(queryStr, models.ActualLRPStateCrashed)
			Expect(err).NotTo(HaveOccurred())
		})

		FContext("when the domain is fresh", func() {
			BeforeEach(func() {
				Expect(sqlDB.UpsertDomain(logger, domain, 5)).To(Succeed())
			})

			It("emit CrashedActualLRPs", func() {
				sqlDB.ConvergeLRPs(logger, models.NewCellSetFromList(nil))

				Eventually(getMetricsEmitted(fakeMetronClient)).Should(ContainElement(event{
					name:  "CrashedActualLRPs",
					value: 2,
				}))
			})

			It("add the keys to UnstartedLRPKeys", func() {
				result := sqlDB.ConvergeLRPs(logger, models.NewCellSetFromList(nil))

				desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				expectedSched := desiredLRP.DesiredLRPSchedulingInfo()
				Expect(result.UnstartedLRPKeys).To(ContainElement(&models.ActualLRPKeyWithSchedulingInfo{
					Key:            &models.ActualLRPKey{ProcessGuid: processGuid, Index: 0, Domain: domain},
					SchedulingInfo: &expectedSched,
				}))

			})
		})

		FContext("when the domain is expired", func() {
			BeforeEach(func() {
				fakeClock.Increment(-10 * time.Second)
				Expect(sqlDB.UpsertDomain(logger, domain, 5)).To(Succeed())
				fakeClock.Increment(10 * time.Second)
			})

			It("emit CrashedActualLRPs", func() {
				sqlDB.ConvergeLRPs(logger, models.NewCellSetFromList(nil))

				Eventually(getMetricsEmitted(fakeMetronClient)).Should(ContainElement(event{
					name:  "CrashedActualLRPs",
					value: 2,
				}))
			})

			It("add the keys to UnstartedLRPKeys", func() {
				result := sqlDB.ConvergeLRPs(logger, models.NewCellSetFromList(nil))

				desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				expectedSched := desiredLRP.DesiredLRPSchedulingInfo()
				Expect(result.UnstartedLRPKeys).To(ContainElement(&models.ActualLRPKeyWithSchedulingInfo{
					Key:            &models.ActualLRPKey{ProcessGuid: processGuid, Index: 0, Domain: domain},
					SchedulingInfo: &expectedSched,
				}))
			})
		})

		FContext("when the the lrps are evacuating", func() {
			BeforeEach(func() {
				Expect(sqlDB.UpsertDomain(logger, domain, 5)).To(Succeed())

				queryStr := `UPDATE actual_lrps SET evacuating = ?`
				if test_helpers.UsePostgres() {
					queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
				}
				_, err := db.Exec(queryStr, true)
				Expect(err).NotTo(HaveOccurred())
			})

			It("emits a LRPsMissing metric", func() {
				sqlDB.ConvergeLRPs(logger, models.NewCellSetFromList(nil))

				Eventually(getMetricsEmitted(fakeMetronClient)).Should(ContainElement(event{
					name:  "LRPsMissing",
					value: 2,
				}))
			})

			// it is the responsibility of the caller to create new LRPs
			It("prune the evacuating LRPs and does not create new ones", func() {
				sqlDB.ConvergeLRPs(logger, models.NewCellSetFromList(nil))

				groups, err := sqlDB.ActualLRPGroupsByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())
				Expect(groups).To(BeEmpty())
			})

			It("emits ActualLRPRemovedEvent for the removed evacuating LRPs", func() {
				groups, err := sqlDB.ActualLRPGroupsByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())
				Expect(groups).To(HaveLen(2))

				result := sqlDB.ConvergeLRPs(logger, models.NewCellSetFromList(nil))
				Expect(result.Events).To(ConsistOf(
					models.NewActualLRPRemovedEvent(groups[0]),
					models.NewActualLRPRemovedEvent(groups[1]),
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
			err := sqlDB.DesireLRP(logger, desiredLRPWithRestartableCrashedActuals)
			Expect(err).NotTo(HaveOccurred())

			for i := int32(0); i < 2; i++ {
				crashedActualLRPKey := models.NewActualLRPKey(processGuid, i, domain)
				_, err = sqlDB.CreateUnclaimedActualLRP(logger, &crashedActualLRPKey)
				Expect(err).NotTo(HaveOccurred())
				instanceGuid := "restartable-crashed-actual" + "-" + domain
				_, _, err = sqlDB.ClaimActualLRP(logger, processGuid, i, &models.ActualLRPInstanceKey{InstanceGuid: instanceGuid, CellId: "existing-cell"})
				Expect(err).NotTo(HaveOccurred())
				actualLRPNetInfo := models.NewActualLRPNetInfo("some-address", "container-address", models.NewPortMapping(2222, 4444))
				_, _, err = sqlDB.StartActualLRP(logger, &crashedActualLRPKey, &models.ActualLRPInstanceKey{InstanceGuid: instanceGuid, CellId: "existing-cell"}, &actualLRPNetInfo)
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
			_, err = db.Exec(queryStr, models.DefaultMaxRestarts+1, models.ActualLRPStateCrashed)
			Expect(err).NotTo(HaveOccurred())
		})

		FContext("when the domain is fresh", func() {
			BeforeEach(func() {
				Expect(sqlDB.UpsertDomain(logger, domain, 5)).To(Succeed())
			})

			It("emit CrashedActualLRPs", func() {
				sqlDB.ConvergeLRPs(logger, models.NewCellSetFromList(nil))

				Eventually(getMetricsEmitted(fakeMetronClient)).Should(ContainElement(event{
					name:  "CrashedActualLRPs",
					value: 2,
				}))
			})

			It("does not add the keys to UnstartedLRPKeys", func() {
				result := sqlDB.ConvergeLRPs(logger, models.NewCellSetFromList(nil))
				Expect(result.UnstartedLRPKeys).To(BeEmpty())
			})
		})

		FContext("when the domain is expired", func() {
			BeforeEach(func() {
				fakeClock.Increment(-10 * time.Second)
				Expect(sqlDB.UpsertDomain(logger, domain, 5)).To(Succeed())
				fakeClock.Increment(10 * time.Second)
			})

			It("emit CrashedActualLRPs", func() {
				sqlDB.ConvergeLRPs(logger, models.NewCellSetFromList(nil))

				Eventually(getMetricsEmitted(fakeMetronClient)).Should(ContainElement(event{
					name:  "CrashedActualLRPs",
					value: 2,
				}))
			})

			It("does not add the keys to UnstartedLRPKeys", func() {
				result := sqlDB.ConvergeLRPs(logger, models.NewCellSetFromList(nil))
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
			_, err := sqlDB.CreateUnclaimedActualLRP(logger, actualLRPWithNoDesired)
			Expect(err).NotTo(HaveOccurred())
		})

		FContext("when the domain is fresh", func() {
			BeforeEach(func() {
				Expect(sqlDB.UpsertDomain(logger, domain, 5)).To(Succeed())
			})

			It("returns extra ActualLRPs to be retired", func() {
				result := sqlDB.ConvergeLRPs(logger, models.NewCellSet())
				keysToRetire := result.KeysToRetire

				actualLRPKey := models.ActualLRPKey{ProcessGuid: processGuid, Index: 0, Domain: domain}
				Expect(keysToRetire).To(ContainElement(&actualLRPKey))
			})

			It("returns the no lrp keys to be started", func() {
				result := sqlDB.ConvergeLRPs(logger, models.NewCellSetFromList(nil))
				Expect(result.UnstartedLRPKeys).To(BeEmpty())
				Expect(result.MissingLRPKeys).To(BeEmpty())
			})

			It("does not touch the ActualLRPs in the database", func() {
				groupsBefore, err := sqlDB.ActualLRPGroupsByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				sqlDB.ConvergeLRPs(logger, models.NewCellSet())

				groupsAfter, err := sqlDB.ActualLRPGroupsByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				Expect(groupsAfter).To(Equal(groupsBefore))
			})

			It("emits LRPsExtra metric", func() {
				sqlDB.ConvergeLRPs(logger, models.NewCellSetFromList(nil))

				Eventually(getMetricsEmitted(fakeMetronClient)).Should(ContainElement(event{
					name:  "LRPsExtra",
					value: 1,
				}))
			})
		})

		FContext("when the domain is expired", func() {
			BeforeEach(func() {
				fakeClock.Increment(-10 * time.Second)
				Expect(sqlDB.UpsertDomain(logger, domain, 5)).To(Succeed())
				fakeClock.Increment(10 * time.Second)
			})

			It("does not return extra ActualLRPs to be retired", func() {
				result := sqlDB.ConvergeLRPs(logger, models.NewCellSet())
				Expect(result.KeysToRetire).To(BeEmpty())
			})

			It("returns the no lrp keys to be started", func() {
				result := sqlDB.ConvergeLRPs(logger, models.NewCellSetFromList(nil))
				Expect(result.UnstartedLRPKeys).To(BeEmpty())
				Expect(result.MissingLRPKeys).To(BeEmpty())
			})

			It("does not touch the ActualLRPs in the database", func() {
				groupsBefore, err := sqlDB.ActualLRPGroupsByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				sqlDB.ConvergeLRPs(logger, models.NewCellSet())

				groupsAfter, err := sqlDB.ActualLRPGroupsByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				Expect(groupsAfter).To(Equal(groupsBefore))
			})

			It("emits zero value for LRPsExtra metric", func() {
				sqlDB.ConvergeLRPs(logger, models.NewCellSetFromList(nil))

				Eventually(getMetricsEmitted(fakeMetronClient)).Should(ContainElement(event{
					name:  "LRPsExtra",
					value: 0,
				}))
			})
		})

		FContext("when the the lrps are evacuating", func() {
			BeforeEach(func() {
				Expect(sqlDB.UpsertDomain(logger, domain, 5)).To(Succeed())

				queryStr := `UPDATE actual_lrps SET evacuating = ?`
				if test_helpers.UsePostgres() {
					queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
				}
				_, err := db.Exec(queryStr, true)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns the no lrp keys to be started", func() {
				result := sqlDB.ConvergeLRPs(logger, models.NewCellSetFromList(nil))
				Expect(result.UnstartedLRPKeys).To(BeEmpty())
				Expect(result.MissingLRPKeys).To(BeEmpty())
			})

			It("removes the evacuating LRPs", func() {
				sqlDB.ConvergeLRPs(logger, models.NewCellSetFromList(nil))

				groups, err := sqlDB.ActualLRPGroupsByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())
				Expect(groups).To(BeEmpty())
			})

			It("emits an ActualLRPRemoved Event", func() {
				groups, err := sqlDB.ActualLRPGroupsByProcessGuid(logger, processGuid)
				Expect(err).NotTo(HaveOccurred())

				Expect(groups).To(HaveLen(1))

				result := sqlDB.ConvergeLRPs(logger, models.NewCellSetFromList(nil))
				Expect(result.Events).To(ConsistOf(
					models.NewActualLRPRemovedEvent(groups[0]),
				))
			})

			It("emits LRPsExtra metric", func() {
				sqlDB.ConvergeLRPs(logger, models.NewCellSetFromList(nil))

				Eventually(getMetricsEmitted(fakeMetronClient)).Should(ContainElement(event{
					name:  "LRPsExtra",
					value: 0,
				}))
			})
		})
	})
})

var _ = Describe("LRPConvergence", func() {
	var (
		freshDomain      string
		expiredDomain    string
		evacuatingDomain string
		cellSet          models.CellSet
		sqlDB            *sqldb.SQLDB

		fakeMetronClient *mfakes.FakeIngressClient
	)

	fetchActuals := func() []string {
		rows, err := db.Query("SELECT process_guid FROM actual_lrps")
		Expect(err).NotTo(HaveOccurred())
		defer rows.Close()

		var processGuid string
		var results []string
		for rows.Next() {
			err = rows.Scan(&processGuid)
			Expect(err).NotTo(HaveOccurred())
			results = append(results, processGuid)
		}
		return results
	}

	Describe("general metrics", func() {
		It("emits missing LRP metrics", func() {
			sqlDB.ConvergeLRPs(logger, cellSet)

			Expect(fakeMetronClient.SendMetricCallCount()).To(Equal(10))
			name, value, _ := fakeMetronClient.SendMetricArgsForCall(2)
			Expect(name).To(Equal("LRPsMissing"))
			Expect(value).To(BeNumerically("==", 17))
		})

		It("emits extra LRP metrics", func() {
			sqlDB.ConvergeLRPs(logger, cellSet)
			Expect(fakeMetronClient.SendMetricCallCount()).To(Equal(10))
			name, value, _ := fakeMetronClient.SendMetricArgsForCall(3)
			Expect(name).To(Equal("LRPsExtra"))
			Expect(value).To(BeNumerically("==", 2))
		})

		It("emits metrics for lrps", func() {
			convergenceLogger := lagertest.NewTestLogger("convergence")
			sqlDB.ConvergeLRPs(convergenceLogger, cellSet)
			Expect(fakeMetronClient.SendMetricCallCount()).To(Equal(10))
			name, value, _ := fakeMetronClient.SendMetricArgsForCall(4)
			Expect(name).To(Equal("LRPsUnclaimed"))
			Expect(value).To(Equal(32)) // 16 fresh + 5 expired + 11 evac
			name, value, _ = fakeMetronClient.SendMetricArgsForCall(5)
			Expect(name).To(Equal("LRPsClaimed"))
			Expect(value).To(Equal(7))
			name, value, _ = fakeMetronClient.SendMetricArgsForCall(6)
			Expect(name).To(Equal("LRPsRunning"))
			Expect(value).To(Equal(1))
			name, value, _ = fakeMetronClient.SendMetricArgsForCall(7)
			Expect(name).To(Equal("CrashedActualLRPs"))
			Expect(value).To(Equal(2))
			name, value, _ = fakeMetronClient.SendMetricArgsForCall(8)
			Expect(name).To(Equal("CrashingDesiredLRPs"))
			Expect(value).To(Equal(1))
			name, value, _ = fakeMetronClient.SendMetricArgsForCall(9)
			Expect(name).To(Equal("LRPsDesired"))
			Expect(value).To(Equal(38))
			Consistently(convergenceLogger).ShouldNot(gbytes.Say("failed-.*"))
		})
	})

	Describe("convergence counters", func() {
		It("bumps the convergence counter", func() {
			Expect(fakeMetronClient.IncrementCounterCallCount()).To(Equal(0))
			sqlDB.ConvergeLRPs(logger, models.CellSet{})
			Expect(fakeMetronClient.IncrementCounterCallCount()).To(Equal(1))
			Expect(fakeMetronClient.IncrementCounterArgsForCall(0)).To(Equal("ConvergenceLRPRuns"))
			sqlDB.ConvergeLRPs(logger, models.CellSet{})
			Expect(fakeMetronClient.IncrementCounterCallCount()).To(Equal(2))
			Expect(fakeMetronClient.IncrementCounterArgsForCall(1)).To(Equal("ConvergenceLRPRuns"))
		})

		It("reports the duration that it took to converge", func() {
			sqlDB.ConvergeLRPs(logger, models.CellSet{})

			Eventually(fakeMetronClient.SendDurationCallCount).Should(Equal(1))
			name, value, _ := fakeMetronClient.SendDurationArgsForCall(0)
			Expect(name).To(Equal("ConvergenceLRPDuration"))
			Expect(value).NotTo(BeZero())
		})
	})

	actualLRPKeyWithSchedulingInfo := func(desiredLRP *models.DesiredLRP, index int) *models.ActualLRPKeyWithSchedulingInfo {
		schedulingInfo := desiredLRP.DesiredLRPSchedulingInfo()
		lrpKey := models.NewActualLRPKey(desiredLRP.ProcessGuid, int32(index), desiredLRP.Domain)

		lrp := &models.ActualLRPKeyWithSchedulingInfo{
			Key:            &lrpKey,
			SchedulingInfo: &schedulingInfo,
		}
		return lrp
	}

	// MOVED
	XIt("returns start requests for stale unclaimed actual LRPs", func() {
		result := sqlDB.ConvergeLRPs(logger, cellSet)
		unstartedLRPKeys := result.UnstartedLRPKeys
		Expect(logger).To(gbytes.Say("creating-start-request.*reason\":\"stale-unclaimed-lrp"))

		By("fresh domain", func() {
			Expect(unstartedLRPKeys).NotTo(BeEmpty())

			processGuid := "desired-with-stale-actuals" + "-" + freshDomain
			desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(logger, processGuid)
			Expect(err).NotTo(HaveOccurred())

			Expect(unstartedLRPKeys).To(ContainElement(actualLRPKeyWithSchedulingInfo(desiredLRP, 1)))
			Expect(unstartedLRPKeys).To(ContainElement(actualLRPKeyWithSchedulingInfo(desiredLRP, 1)))
		})

		By("expired domain", func() {
			Expect(unstartedLRPKeys).NotTo(BeEmpty())

			processGuid := "desired-with-stale-actuals" + "-" + expiredDomain
			desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(logger, processGuid)
			Expect(err).NotTo(HaveOccurred())

			Expect(unstartedLRPKeys).To(ContainElement(actualLRPKeyWithSchedulingInfo(desiredLRP, 0)))
			Expect(unstartedLRPKeys).To(ContainElement(actualLRPKeyWithSchedulingInfo(desiredLRP, 1)))
		})
	})

	// MOVED
	XIt("returns the start requests and actual lrp keys for actuals with missing cells", func() {
		result := sqlDB.ConvergeLRPs(logger, cellSet)
		keysWithMissingCells := result.KeysWithMissingCells

		By("fresh domain", func() {
			processGuid := "desired-with-missing-cell-actuals" + "-" + freshDomain
			desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(logger, processGuid)
			Expect(err).NotTo(HaveOccurred())

			actualLRPGroup, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, processGuid, 0)
			Expect(err).NotTo(HaveOccurred())
			expectedSched := desiredLRP.DesiredLRPSchedulingInfo()
			Expect(keysWithMissingCells).To(ContainElement(&models.ActualLRPKeyWithSchedulingInfo{
				Key:            &actualLRPGroup.Instance.ActualLRPKey,
				SchedulingInfo: &expectedSched,
			}))
		})

		By("expired domain", func() {
			processGuid := "desired-with-missing-cell-actuals" + "-" + expiredDomain
			desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(logger, processGuid)
			Expect(err).NotTo(HaveOccurred())

			actualLRPGroup, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, processGuid, 0)
			Expect(err).NotTo(HaveOccurred())
			expectedSched := desiredLRP.DesiredLRPSchedulingInfo()
			Expect(keysWithMissingCells).To(ContainElement(&models.ActualLRPKeyWithSchedulingInfo{
				Key:            &actualLRPGroup.Instance.ActualLRPKey,
				SchedulingInfo: &expectedSched,
			}))
		})
	})

	// MOVED
	XIt("logs the missing cells", func() {
		sqlDB.ConvergeLRPs(logger, cellSet)
		Expect(logger).To(gbytes.Say(`detected-missing-cells.*cell_ids":\["other-cell"\]`))
	})

	// MOVED
	XContext("when there are no missing cells", func() {
		BeforeEach(func() {
			cellSet = models.NewCellSetFromList([]*models.CellPresence{
				{CellId: "existing-cell"},
				{CellId: "other-cell"},
			})
		})
		It("does not log missing cells", func() {
			sqlDB.ConvergeLRPs(logger, cellSet)
			Expect(logger).ToNot(gbytes.Say("detected-missing-cells"))
		})
	})

	It("creates actual LRPs with missing indices, and returns it to be started", func() {
		result := sqlDB.ConvergeLRPs(logger, cellSet)
		missingLRPKeys := result.MissingLRPKeys
		Expect(missingLRPKeys).NotTo(BeEmpty())

		Expect(logger).To(gbytes.Say("creating-start-request.*reason\":\"missing-instance"))

		By("missing all actuals, fresh domain", func() {
			processGuid := "desired-with-missing-all-actuals" + "-" + freshDomain
			desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(logger, processGuid)
			Expect(err).NotTo(HaveOccurred())

			Expect(missingLRPKeys).To(ContainElement(actualLRPKeyWithSchedulingInfo(desiredLRP, 0)))
		})

		By("missing some actuals, fresh domain", func() {
			processGuid := "desired-with-missing-some-actuals" + "-" + freshDomain
			desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(logger, processGuid)
			Expect(err).NotTo(HaveOccurred())

			Expect(missingLRPKeys).To(ContainElement(actualLRPKeyWithSchedulingInfo(desiredLRP, 1)))
			Expect(missingLRPKeys).To(ContainElement(actualLRPKeyWithSchedulingInfo(desiredLRP, 3)))
		})

		By("missing all actuals, expired domain", func() {
			processGuid := "desired-with-missing-all-actuals" + "-" + expiredDomain
			desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(logger, processGuid)
			Expect(err).NotTo(HaveOccurred())

			Expect(missingLRPKeys).To(ContainElement(actualLRPKeyWithSchedulingInfo(desiredLRP, 0)))
		})

		By("missing some actuals, expired domain", func() {
			processGuid := "desired-with-missing-some-actuals" + "-" + expiredDomain
			desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(logger, processGuid)
			Expect(err).NotTo(HaveOccurred())

			Expect(missingLRPKeys).To(ContainElement(actualLRPKeyWithSchedulingInfo(desiredLRP, 1)))
			Expect(missingLRPKeys).To(ContainElement(actualLRPKeyWithSchedulingInfo(desiredLRP, 3)))
		})
	})

	It("unclaims actual LRPs that are crashed and restartable, and returns it to be started", func() {
		result := sqlDB.ConvergeLRPs(logger, cellSet)
		unstartedLRPKeys := result.UnstartedLRPKeys
		Expect(unstartedLRPKeys).NotTo(BeEmpty())

		Expect(logger).To(gbytes.Say("creating-start-request.*reason\":\"crashed-instance"))

		By("fresh domain", func() {
			processGuid := "desired-with-restartable-crashed-actuals" + "-" + freshDomain
			desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(logger, processGuid)
			Expect(err).NotTo(HaveOccurred())

			Expect(unstartedLRPKeys).To(ContainElement(actualLRPKeyWithSchedulingInfo(desiredLRP, 0)))
			Expect(unstartedLRPKeys).To(ContainElement(actualLRPKeyWithSchedulingInfo(desiredLRP, 1)))
		})

		By("expired domain", func() {
			processGuid := "desired-with-restartable-crashed-actuals" + "-" + expiredDomain
			desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(logger, processGuid)
			Expect(err).NotTo(HaveOccurred())

			Expect(unstartedLRPKeys).To(ContainElement(actualLRPKeyWithSchedulingInfo(desiredLRP, 0)))
			Expect(unstartedLRPKeys).To(ContainElement(actualLRPKeyWithSchedulingInfo(desiredLRP, 1)))
		})
	})

	It("returns extra actual LRPs to be retired", func() {
		result := sqlDB.ConvergeLRPs(logger, cellSet)
		keysToRetire := result.KeysToRetire
		Expect(keysToRetire).NotTo(BeEmpty())

		processGuid := "desired-with-extra-actuals" + "-" + freshDomain
		actualLRPKey := models.ActualLRPKey{ProcessGuid: processGuid, Index: 4, Domain: freshDomain}
		Expect(keysToRetire).To(ContainElement(&actualLRPKey))

		processGuid = "actual-with-no-desired" + "-" + freshDomain
		actualLRPKey = models.ActualLRPKey{ProcessGuid: processGuid, Index: 0, Domain: freshDomain}
		Expect(keysToRetire).To(ContainElement(&actualLRPKey))
	})

	It("creates unclaimed for evacuating instances that are missing the running record", func() {
		result := sqlDB.ConvergeLRPs(logger, cellSet)

		processGuids := []string{
			"desired-with-stale-actuals" + "-" + evacuatingDomain,
			"desired-with-missing-cell-actuals" + "-" + evacuatingDomain,
			"desired-with-extra-actuals" + "-" + evacuatingDomain,
			"desired-with-missing-all-actuals" + "-" + evacuatingDomain,
			"desired-with-missing-some-actuals" + "-" + evacuatingDomain,
			"desired-with-restartable-crashed-actuals" + "-" + evacuatingDomain,
		}

		for _, processGuid := range processGuids {
			desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(logger, processGuid)
			Expect(err).NotTo(HaveOccurred())

			for i := 0; i < int(desiredLRP.Instances); i++ {
				Expect(result.MissingLRPKeys).To(ContainElement(actualLRPKeyWithSchedulingInfo(desiredLRP, i)))
			}
		}
	})

	It("clears out expired domains", func() {
		fetchDomains := func() []string {
			rows, err := db.Query("SELECT domain FROM domains")
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

		sqlDB.ConvergeLRPs(logger, cellSet)

		Expect(fetchDomains()).NotTo(ContainElement(expiredDomain))
	})

	Context("with evacuating actual lrps", func() {

		BeforeEach(func() {
			Expect(fetchActuals()).To(ContainElement("evacuating-actual-lrp"))
			Expect(fetchActuals()).To(ContainElement("missing-evacuating-actual-lrp"))
		})

		It("returns an ActualLRPRemovedEvent", func() {
			lrps, err := sqlDB.ActualLRPGroupsByProcessGuid(logger, "missing-evacuating-actual-lrp")
			Expect(err).NotTo(HaveOccurred())
			Expect(len(lrps)).To(Equal(1))

			result := sqlDB.ConvergeLRPs(logger, cellSet)
			events := result.Events

			Expect(len(events)).NotTo(BeZero())
			event := models.NewActualLRPRemovedEvent(lrps[0])
			Expect(events).To(ContainElement(event))
		})

		It("keeps evacuating actual lrps with available cells", func() {
			sqlDB.ConvergeLRPs(logger, cellSet)

			Expect(fetchActuals()).To(ContainElement("evacuating-actual-lrp"))
		})

		It("clears out evacuating actual lrps with missing cells", func() {
			sqlDB.ConvergeLRPs(logger, cellSet)

			Expect(fetchActuals()).NotTo(ContainElement("missing-evacuating-actual-lrp"))
		})
	})

	It("ignores LRPs that don't need convergence", func() {
		processGuids := []string{
			"normal-desired-lrp" + "-" + freshDomain,
			"normal-desired-lrp-with-unclaimed-actuals" + "-" + freshDomain,
			"desired-with-non-restartable-crashed-actuals" + "-" + freshDomain,
			"desired-with-extra-actuals" + "-" + expiredDomain,
		}

		fetch := func(processGuid string) (*models.DesiredLRP, []*models.ActualLRPGroup) {
			desired, err := sqlDB.DesiredLRPByProcessGuid(logger, processGuid)
			Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("should've found desired lrp with guid: %s", processGuid))
			actuals, err := sqlDB.ActualLRPGroupsByProcessGuid(logger, processGuid)
			Expect(err).NotTo(HaveOccurred())
			return desired, actuals
		}

		beforeDesireds := make([]*models.DesiredLRP, 0, len(processGuids))
		beforeActuals := make([][]*models.ActualLRPGroup, 0, len(processGuids))
		for _, processGuid := range processGuids {
			desired, actuals := fetch(processGuid)
			beforeDesireds = append(beforeDesireds, desired)
			beforeActuals = append(beforeActuals, actuals)
		}

		result := sqlDB.ConvergeLRPs(logger, cellSet)
		missingLRPKeys := result.MissingLRPKeys
		keysWithMissingCells := result.KeysWithMissingCells
		keysToRetire := result.KeysToRetire

		startGuids := make([]string, 0, len(missingLRPKeys))
		for _, lrpKey := range missingLRPKeys {
			startGuids = append(startGuids, lrpKey.Key.ProcessGuid)
		}

		for _, processGuid := range processGuids {
			Expect(startGuids).NotTo(ContainElement(processGuid))
		}

		retiredGuids := make([]string, 0, len(keysToRetire))
		for _, keyToRetire := range keysToRetire {
			retiredGuids = append(retiredGuids, keyToRetire.ProcessGuid)
		}
		for _, processGuid := range processGuids {
			Expect(retiredGuids).NotTo(ContainElement(processGuid))
		}

		guidsToUnclaim := make([]string, 0, len(keysWithMissingCells))
		for _, keyWithMissingCell := range keysWithMissingCells {
			guidsToUnclaim = append(guidsToUnclaim, keyWithMissingCell.Key.ProcessGuid)
		}
		for _, processGuid := range processGuids {
			Expect(guidsToUnclaim).NotTo(ContainElement(processGuid))
		}

		afterDesireds := make([]*models.DesiredLRP, 0, len(processGuids))
		afterActuals := make([][]*models.ActualLRPGroup, 0, len(processGuids))
		for _, processGuid := range processGuids {
			desired, actuals := fetch(processGuid)
			afterDesireds = append(afterDesireds, desired)
			afterActuals = append(afterActuals, actuals)
		}

		Expect(beforeDesireds).To(Equal(afterDesireds))
		Expect(beforeActuals).To(Equal(afterActuals))
	})

	Context("when the cell set is empty", func() {
		BeforeEach(func() {
			cellSet = models.NewCellSetFromList([]*models.CellPresence{})
		})

		It("reports all non evacuating actual lrps as missing cells", func() {
			result := sqlDB.ConvergeLRPs(logger, models.CellSet{})
			actualsWithMissingCells := result.KeysWithMissingCells
			Expect(len(actualsWithMissingCells)).To(Equal(23))
		})

		It("clears out all evacuating actual lrps", func() {
			sqlDB.ConvergeLRPs(logger, cellSet)

			Expect(fetchActuals()).NotTo(ContainElement("missing-evacuating-actual-lrp"))
			Expect(fetchActuals()).NotTo(ContainElement("evacuating-actual-lrp"))
		})
	})
})
