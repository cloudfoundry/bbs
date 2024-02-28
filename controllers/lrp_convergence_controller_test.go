package controllers_test

import (
	"context"
	"errors"
	"time"

	"code.cloudfoundry.org/auctioneer"
	"code.cloudfoundry.org/auctioneer/auctioneerfakes"
	"code.cloudfoundry.org/bbs/controllers"
	"code.cloudfoundry.org/bbs/controllers/fakes"
	"code.cloudfoundry.org/bbs/db"
	"code.cloudfoundry.org/bbs/db/dbfakes"
	"code.cloudfoundry.org/bbs/events/eventfakes"
	mfakes "code.cloudfoundry.org/bbs/metrics/fakes"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/models/test/model_helpers"
	"code.cloudfoundry.org/bbs/serviceclient/serviceclientfakes"
	"code.cloudfoundry.org/bbs/trace"
	"code.cloudfoundry.org/clock/fakeclock"
	"code.cloudfoundry.org/lager/v3"
	"code.cloudfoundry.org/lager/v3/lagertest"
	"code.cloudfoundry.org/rep"
	"code.cloudfoundry.org/routing-info/internalroutes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("LRP Convergence Controllers", func() {
	const traceId = "some-trace-id"

	var (
		logger                    *lagertest.TestLogger
		fakeClock                 *fakeclock.FakeClock
		fakeLRPDB                 *dbfakes.FakeLRPDB
		fakeSuspectDB             *dbfakes.FakeSuspectDB
		fakeDomainDB              *dbfakes.FakeDomainDB
		actualHub                 *eventfakes.FakeHub
		actualLRPInstanceHub      *eventfakes.FakeHub
		retirer                   *fakes.FakeRetirer
		fakeAuctioneerClient      *auctioneerfakes.FakeClient
		fakeLRPStatMetronNotifier *mfakes.FakeLRPStatMetronNotifier

		keysToRetire         []*models.ActualLRPKey
		keysWithMissingCells []*models.ActualLRPKeyWithSchedulingInfo

		retiringActualLRP1 *models.ActualLRP
		retiringActualLRP2 *models.ActualLRP

		desiredLRP1 models.DesiredLRPSchedulingInfo
		cellSet     models.CellSet

		controller *controllers.LRPConvergenceController
	)

	BeforeEach(func() {
		fakeClock = fakeclock.NewFakeClock(time.Now())
		fakeLRPDB = new(dbfakes.FakeLRPDB)
		fakeSuspectDB = new(dbfakes.FakeSuspectDB)
		fakeDomainDB = new(dbfakes.FakeDomainDB)
		fakeAuctioneerClient = new(auctioneerfakes.FakeClient)
		logger = lagertest.NewTestLogger("test")

		desiredLRP1 = model_helpers.NewValidDesiredLRP("to-unclaim-1").DesiredLRPSchedulingInfo()

		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))

		fakeServiceClient = new(serviceclientfakes.FakeServiceClient)
		fakeServiceClient.CellByIdReturns(nil, errors.New("hi"))
		fakeLRPStatMetronNotifier = new(mfakes.FakeLRPStatMetronNotifier)

		cellPresence := models.NewCellPresence("cell-id", "1.1.1.1", "", "z1", models.CellCapacity{}, nil, nil, nil, nil)
		cellSet = models.CellSet{"cell-id": &cellPresence}
		fakeServiceClient.CellsReturns(cellSet, nil)

		actualHub = &eventfakes.FakeHub{}
		actualLRPInstanceHub = &eventfakes.FakeHub{}
		retirer = &fakes.FakeRetirer{}
	})

	JustBeforeEach(func() {
		controller = controllers.NewLRPConvergenceController(
			logger,
			fakeClock,
			fakeLRPDB,
			fakeSuspectDB,
			fakeDomainDB,
			actualHub,
			actualLRPInstanceHub,
			fakeAuctioneerClient,
			fakeServiceClient,
			fakeRepClientFactory,
			retirer,
			2,
			fakeLRPStatMetronNotifier,
		)
		controller.ConvergeLRPs(context.WithValue(ctx, trace.RequestIdHeader, traceId))
	})

	It("calls ConvergeLRPs", func() {
		Expect(fakeLRPDB.ConvergeLRPsCallCount()).To(Equal(1))
		_, _, actualCellSet := fakeLRPDB.ConvergeLRPsArgsForCall(0)
		Expect(actualCellSet).To(BeEquivalentTo(cellSet))
	})

	Describe("metrics", func() {
		Context("when convergence occurs", func() {
			BeforeEach(func() {
				fakeLRPDB.ConvergeLRPsStub = func(context.Context, lager.Logger, models.CellSet) db.ConvergenceResult {
					fakeClock.Increment(50 * time.Second)
					return db.ConvergenceResult{}
				}
			})

			It("records convergence duration", func() {
				Expect(fakeLRPStatMetronNotifier.RecordConvergenceDurationCallCount()).To(Equal(1))
				Expect(fakeLRPStatMetronNotifier.RecordConvergenceDurationArgsForCall(0)).To(Equal(50 * time.Second))
			})
		})

		Context("when there are fresh domains", func() {
			var domains []string

			BeforeEach(func() {
				domains = []string{"domain-1", "domain-2"}
				fakeDomainDB.FreshDomainsReturns(domains, nil)
			})

			It("records domain freshness metric for each domain", func() {
				Expect(fakeLRPStatMetronNotifier.RecordFreshDomainsCallCount()).To(Equal(1))
				Expect(fakeLRPStatMetronNotifier.RecordFreshDomainsArgsForCall(0)).To(ConsistOf(domains))
			})
		})

		Context("when there are LRPs", func() {
			BeforeEach(func() {
				fakeLRPDB.CountActualLRPsByStateReturns(2, 1, 3, 4, 5)
				fakeLRPDB.CountDesiredInstancesReturns(6)

				fakeLRPDB.ConvergeLRPsReturns(db.ConvergenceResult{
					MissingLRPKeys: []*models.ActualLRPKeyWithSchedulingInfo{
						{
							Key: &models.ActualLRPKey{
								ProcessGuid: "some-lrp",
								Index:       0,
								Domain:      "some-domain",
							},
							SchedulingInfo: &desiredLRP1,
						},
						{
							Key: &models.ActualLRPKey{
								ProcessGuid: "some-other-lrp",
								Index:       1,
								Domain:      "some-other-domain",
							},
							SchedulingInfo: &desiredLRP1,
						},
					},
					KeysToRetire: []*models.ActualLRPKey{
						{ProcessGuid: "some-lrp"},
						{ProcessGuid: "some-other-lrp"},
					},
					SuspectRunningKeys: []*models.ActualLRPKey{
						{ProcessGuid: "some-suspect-lrp"},
						{ProcessGuid: "some-other-suspect-lrp"},
					},
					SuspectClaimedKeys: []*models.ActualLRPKey{
						{ProcessGuid: "some-suspect-claimed-lrp"},
					},
				})
			})

			It("records LRP counts", func() {
				Expect(fakeLRPStatMetronNotifier.RecordLRPCountsCallCount()).To(Equal(1))
				unclaimed, claimed, running, crashed, missing, extra, suspectRunning, suspectClaimed, desired, crashingDesired := fakeLRPStatMetronNotifier.RecordLRPCountsArgsForCall(0)
				Expect(unclaimed).To(Equal(1))
				Expect(claimed).To(Equal(2))
				Expect(running).To(Equal(3))
				Expect(crashed).To(Equal(4))
				Expect(missing).To(Equal(2))
				Expect(extra).To(Equal(2))
				Expect(suspectRunning).To(Equal(2))
				Expect(suspectClaimed).To(Equal(1))
				Expect(desired).To(Equal(6))
				Expect(crashingDesired).To(Equal(5))
			})
		})

		Context("when there are multiple cells", func() {
			BeforeEach(func() {
				presentCellPresence1 := models.NewCellPresence("cell-id-1", "1.1.1.1", "", "z1", models.CellCapacity{}, nil, nil, nil, nil)
				presentCellPresence2 := models.NewCellPresence("cell-id-2", "1.1.1.2", "", "z1", models.CellCapacity{}, nil, nil, nil, nil)
				cellSet = models.CellSet{"cell-id-1": &presentCellPresence1, "cell-id-2": &presentCellPresence2}
				fakeServiceClient.CellsReturns(cellSet, nil)
			})

			It("records present cells count", func() {
				Expect(fakeLRPStatMetronNotifier.RecordCellCountsCallCount()).To(Equal(1))
				presentCellsCount, suspectCellsCount := fakeLRPStatMetronNotifier.RecordCellCountsArgsForCall(0)
				Expect(presentCellsCount).To(Equal(2))
				Expect(suspectCellsCount).To(Equal(0))
			})

			Context("when there are missing cells", func() {
				BeforeEach(func() {
					fakeLRPDB.ConvergeLRPsReturns(db.ConvergenceResult{
						MissingCellIds: []string{"cell-id-3", "cell-id-4", "cell-id-5"},
					})
				})

				It("records suspect cells count", func() {
					Expect(fakeLRPStatMetronNotifier.RecordCellCountsCallCount()).To(Equal(1))
					presentCellsCount, suspectCellsCount := fakeLRPStatMetronNotifier.RecordCellCountsArgsForCall(0)
					Expect(presentCellsCount).To(Equal(2))
					Expect(suspectCellsCount).To(Equal(3))
				})
			})
		})
	})

	Context("when there are unstarted ActualLRPs", func() {
		var (
			key            *models.ActualLRPKey
			schedulingInfo models.DesiredLRPSchedulingInfo
			before, after  *models.ActualLRP
		)

		BeforeEach(func() {
			lrp := model_helpers.NewValidDesiredLRP("some-guid")
			schedulingInfo = lrp.DesiredLRPSchedulingInfo()
			key = &models.ActualLRPKey{
				ProcessGuid: lrp.ProcessGuid,
				Index:       0,
				Domain:      lrp.Domain,
			}
			lrpKeyWithSchedulingInfo := &models.ActualLRPKeyWithSchedulingInfo{
				Key:            key,
				SchedulingInfo: &schedulingInfo,
			}
			fakeLRPDB.ConvergeLRPsReturns(db.ConvergenceResult{
				UnstartedLRPKeys: []*models.ActualLRPKeyWithSchedulingInfo{lrpKeyWithSchedulingInfo},
			})

			before = model_helpers.NewValidActualLRP("some-guid", 0)
			before.State = models.ActualLRPStateCrashed
			unclaimedLRP := *before
			unclaimedLRP.State = models.ActualLRPStateUnclaimed
			after = &unclaimedLRP
			fakeLRPDB.UnclaimActualLRPReturns(before, after, nil)
		})

		It("auctions off the returned keys", func() {
			Expect(fakeAuctioneerClient.RequestLRPAuctionsCallCount()).To(Equal(1))

			_, actualTraceId, startAuctions := fakeAuctioneerClient.RequestLRPAuctionsArgsForCall(0)
			Expect(startAuctions).To(HaveLen(1))
			Expect(actualTraceId).To(Equal(traceId))
			request := auctioneer.NewLRPStartRequestFromModel(model_helpers.NewValidDesiredLRP("some-guid"), 0)
			Expect(startAuctions).To(ContainElement(&request))
		})

		It("transition the LRP to UNCLAIMED state", func() {
			Eventually(fakeLRPDB.UnclaimActualLRPCallCount).Should(Equal(1))
			_, _, actualKey := fakeLRPDB.UnclaimActualLRPArgsForCall(0)
			Expect(actualKey).To(Equal(key))
		})

		It("emits an LRPChanged event", func() {
			Eventually(actualHub.EmitCallCount).Should(Equal(1))
			event := actualHub.EmitArgsForCall(0)
			Expect(event).To(Equal(models.NewActualLRPChangedEvent(before.ToActualLRPGroup(), after.ToActualLRPGroup())))
		})

		It("emits an LRPInstanceCreate event followed by an LRPInstanceRemoved event", func() {
			Eventually(actualLRPInstanceHub.EmitCallCount).Should(Equal(2))
			event := actualLRPInstanceHub.EmitArgsForCall(0)
			Expect(event).To(Equal(models.NewActualLRPInstanceCreatedEvent(after, traceId)))
			event = actualLRPInstanceHub.EmitArgsForCall(1)
			Expect(event).To(Equal(models.NewActualLRPInstanceRemovedEvent(before, traceId)))
		})

		Context("and the LRP isn't changed", func() {
			BeforeEach(func() {
				*before = *after
				before.State = models.ActualLRPStateUnclaimed
				fakeLRPDB.UnclaimActualLRPReturns(before, after, nil)
			})

			It("does not emit any events", func() {
				Consistently(actualHub.EmitCallCount).Should(BeZero())
			})

			It("does not emit any instance events", func() {
				Consistently(actualLRPInstanceHub.EmitCallCount).Should(BeZero())
			})
		})

		Context("when the LRP cannot be unclaimed because it is already unclaimed", func() {
			BeforeEach(func() {
				fakeLRPDB.UnclaimActualLRPReturns(nil, nil, models.ErrActualLRPCannotBeUnclaimed)
			})

			It("auctions off the returned keys", func() {
				Expect(fakeAuctioneerClient.RequestLRPAuctionsCallCount()).To(Equal(1))

				_, actualTraceId, startAuctions := fakeAuctioneerClient.RequestLRPAuctionsArgsForCall(0)
				Expect(startAuctions).To(HaveLen(1))
				Expect(actualTraceId).To(Equal(traceId))
				request := auctioneer.NewLRPStartRequestFromModel(model_helpers.NewValidDesiredLRP("some-guid"), 0)
				Expect(startAuctions).To(ContainElement(&request))
			})

			It("does not emit LRP changed event", func() {
				Consistently(actualHub.EmitCallCount).Should(BeZero())
			})

			It("does not emit any instance events", func() {
				Consistently(actualLRPInstanceHub.EmitCallCount).Should(BeZero())
			})
		})
	})

	Context("when there are missing ActualLRPs", func() {
		var (
			key            *models.ActualLRPKey
			schedulingInfo models.DesiredLRPSchedulingInfo
			after          *models.ActualLRP
		)

		BeforeEach(func() {
			lrp := model_helpers.NewValidDesiredLRP("some-guid")
			schedulingInfo = lrp.DesiredLRPSchedulingInfo()
			key = &models.ActualLRPKey{
				ProcessGuid: lrp.ProcessGuid,
				Index:       0,
				Domain:      lrp.Domain,
			}
			lrpKeyWithSchedulingInfo := &models.ActualLRPKeyWithSchedulingInfo{
				Key:            key,
				SchedulingInfo: &schedulingInfo,
			}
			fakeLRPDB.ConvergeLRPsReturns(db.ConvergenceResult{
				MissingLRPKeys: []*models.ActualLRPKeyWithSchedulingInfo{lrpKeyWithSchedulingInfo},
			})

			after = model_helpers.NewValidActualLRP("some-guid", 0)
			after.State = models.ActualLRPStateUnclaimed
			fakeLRPDB.CreateUnclaimedActualLRPReturns(after, nil)
		})

		It("auctions off the returned keys", func() {
			Expect(fakeAuctioneerClient.RequestLRPAuctionsCallCount()).To(Equal(1))

			_, actualTraceId, startAuctions := fakeAuctioneerClient.RequestLRPAuctionsArgsForCall(0)
			Expect(startAuctions).To(HaveLen(1))
			Expect(actualTraceId).To(Equal(traceId))
			request := auctioneer.NewLRPStartRequestFromModel(model_helpers.NewValidDesiredLRP("some-guid"), 0)
			Expect(startAuctions).To(ContainElement(&request))
		})

		It("creates the LPR record in the database", func() {
			Eventually(fakeLRPDB.CreateUnclaimedActualLRPCallCount).Should(Equal(1))
			_, _, actualKey := fakeLRPDB.CreateUnclaimedActualLRPArgsForCall(0)
			Expect(actualKey).To(Equal(key))
		})

		It("emits a LPRCreated event", func() {
			Eventually(actualHub.EmitCallCount).Should(Equal(1))
			event := actualHub.EmitArgsForCall(0)
			Expect(event).To(Equal(models.NewActualLRPCreatedEvent(after.ToActualLRPGroup())))
		})

		It("emits a LPRInstanceCreated event", func() {
			Eventually(actualLRPInstanceHub.EmitCallCount).Should(Equal(1))
			event := actualLRPInstanceHub.EmitArgsForCall(0)
			Expect(event).To(Equal(models.NewActualLRPInstanceCreatedEvent(after, traceId)))
		})
	})

	Context("when fetching the cells fails", func() {
		BeforeEach(func() {
			fakeServiceClient.CellsReturns(nil, errors.New("kaboom"))
		})

		It("does not call ConvergeLRPs", func() {
			Expect(fakeLRPDB.ConvergeLRPsCallCount()).To(Equal(0))
		})

		It("logs the error", func() {
			Eventually(logger).Should(gbytes.Say("failed-listing-cells"))
		})
	})

	Context("when fetching the cells returns ErrResourceNotFound", func() {
		BeforeEach(func() {
			fakeServiceClient.CellsReturns(nil, models.ErrResourceNotFound)
		})

		It("calls ConvergeLRPs with an empty CellSet", func() {
			Expect(fakeLRPDB.ConvergeLRPsCallCount()).To(Equal(1))
			_, _, actualCellSet := fakeLRPDB.ConvergeLRPsArgsForCall(0)
			Expect(actualCellSet).To(BeEquivalentTo(models.CellSet{}))
		})
	})

	Context("when no lrps to auction", func() {
		BeforeEach(func() {
			fakeLRPDB.ConvergeLRPsReturns(db.ConvergenceResult{})
		})

		It("doesn't start the auctions", func() {
			Consistently(fakeAuctioneerClient.RequestLRPAuctionsCallCount).Should(Equal(0))
		})
	})

	Context("when there is an LRP with missing cell", func() {
		var (
			before, after, suspectActualLRP *models.ActualLRP
		)

		BeforeEach(func() {
			suspectActualLRP = model_helpers.NewValidActualLRP("to-unclaim-1", 0)

			keysWithMissingCells = []*models.ActualLRPKeyWithSchedulingInfo{
				{Key: &suspectActualLRP.ActualLRPKey, SchedulingInfo: &desiredLRP1},
			}
			fakeLRPDB.ConvergeLRPsReturns(db.ConvergenceResult{
				KeysWithMissingCells: keysWithMissingCells,
			})
		})

		Context("when there is an ordinary LRP", func() {
			var unclaimed *models.ActualLRP

			BeforeEach(func() {
				before = &models.ActualLRP{Presence: models.ActualLRP_Ordinary}
				after = &models.ActualLRP{Presence: models.ActualLRP_Suspect}
				fakeLRPDB.ChangeActualLRPPresenceReturns(before, after, nil)

				unclaimed = &models.ActualLRP{State: models.ActualLRPStateUnclaimed}
				fakeLRPDB.CreateUnclaimedActualLRPReturns(unclaimed, nil)
			})

			It("changes the LRP presence to 'Suspect'", func() {
				Eventually(fakeLRPDB.ChangeActualLRPPresenceCallCount).Should(Equal(1))

				_, _, key, from, to := fakeLRPDB.ChangeActualLRPPresenceArgsForCall(0)
				Expect(from).To(Equal(models.ActualLRP_Ordinary))
				Expect(to).To(Equal(models.ActualLRP_Suspect))
				Expect(key).To(Equal(&suspectActualLRP.ActualLRPKey))
			})

			It("creates a new unclaimed LRP", func() {
				Expect(fakeLRPDB.CreateUnclaimedActualLRPCallCount()).To(Equal(1))
				_, _, lrpKey := fakeLRPDB.CreateUnclaimedActualLRPArgsForCall(0)

				Expect(lrpKey).To(Equal(&suspectActualLRP.ActualLRPKey))
			})

			It("auctions new lrps", func() {
				Expect(fakeAuctioneerClient.RequestLRPAuctionsCallCount()).To(Equal(1))

				unclaimedStartRequest := auctioneer.NewLRPStartRequestFromSchedulingInfo(&desiredLRP1, 0)

				keysToAuction := []*auctioneer.LRPStartRequest{&unclaimedStartRequest}

				_, actualTraceId, startAuctions := fakeAuctioneerClient.RequestLRPAuctionsArgsForCall(0)
				Expect(startAuctions).To(ConsistOf(keysToAuction))
				Expect(actualTraceId).To(Equal(traceId))
			})

			It("emits no group events", func() {
				Consistently(actualHub.EmitCallCount).Should(Equal(0))
			})

			It("emits instance change events for the suspect", func() {
				Eventually(fakeLRPDB.ChangeActualLRPPresenceCallCount).Should(Equal(1))
				Eventually(actualLRPInstanceHub.EmitCallCount).Should(Equal(2))
				Consistently(actualLRPInstanceHub.EmitCallCount).Should(Equal(2))

				events := []models.Event{
					actualLRPInstanceHub.EmitArgsForCall(0),
					actualLRPInstanceHub.EmitArgsForCall(1),
				}

				Expect(events).To(
					ConsistOf(
						models.NewActualLRPInstanceChangedEvent(before, after, traceId),
						models.NewActualLRPInstanceCreatedEvent(unclaimed, traceId),
					),
				)
			})
		})

		Context("when there already is a Suspect LRP", func() {
			BeforeEach(func() {
				suspectLRPKeys := []*models.ActualLRPKey{
					&suspectActualLRP.ActualLRPKey,
				}
				fakeLRPDB.ConvergeLRPsReturns(db.ConvergenceResult{
					KeysWithMissingCells: keysWithMissingCells,
					SuspectRunningKeys:   suspectLRPKeys,
				})
				before = &models.ActualLRP{State: models.ActualLRPStateClaimed}
				after = &models.ActualLRP{State: models.ActualLRPStateUnclaimed}
				fakeLRPDB.UnclaimActualLRPReturns(before, after, nil)
			})

			It("unclaims the suspect replacement lrp", func() {
				Expect(fakeLRPDB.UnclaimActualLRPCallCount()).To(Equal(1))
			})

			It("does not emit change events", func() {
				Consistently(actualHub.EmitCallCount).Should(Equal(0))
			})

			It("emits ActualLRPInstanceCreatedEvent for unclaimed and ActualLRPInstanceRemovedEvent for suspect", func() {
				Eventually(actualLRPInstanceHub.EmitCallCount).Should(Equal(2))
				event := actualLRPInstanceHub.EmitArgsForCall(0)
				Expect(event).To(BeAssignableToTypeOf(&models.ActualLRPInstanceCreatedEvent{}))
				Expect(event.(*models.ActualLRPInstanceCreatedEvent).ActualLrp).To(Equal(after))
				event = actualLRPInstanceHub.EmitArgsForCall(1)
				Expect(event).To(BeAssignableToTypeOf(&models.ActualLRPInstanceRemovedEvent{}))
				Expect(event.(*models.ActualLRPInstanceRemovedEvent).ActualLrp).To(Equal(before))
			})

			It("does not try to change the LRP presence or create a new unclaimed LRP", func() {
				Consistently(fakeLRPDB.ChangeActualLRPPresenceCallCount).Should(Equal(0))
				Consistently(fakeLRPDB.CreateUnclaimedActualLRPCallCount).Should(Equal(0))
			})
		})

		Context("when changing the actual lrp presence fails", func() {
			BeforeEach(func() {
				fakeLRPDB.ChangeActualLRPPresenceReturns(nil, nil, errors.New("terrrible"))
				fakeLRPDB.ConvergeLRPsReturns(db.ConvergenceResult{
					KeysWithMissingCells: keysWithMissingCells,
				})
			})

			It("does not emit change events", func() {
				Eventually(fakeLRPDB.ChangeActualLRPPresenceCallCount).Should(Equal(1))
				Consistently(actualHub.EmitCallCount).Should(Equal(0))
			})

			It("does not emit instance change events", func() {
				Eventually(fakeLRPDB.ChangeActualLRPPresenceCallCount).Should(Equal(1))
				Consistently(actualLRPInstanceHub.EmitCallCount).Should(Equal(0))
			})
		})
	})

	Context("when there are suspect LRPs with existing cells", func() {
		var (
			suspectActualLRP             *models.ActualLRP
			ordinaryActualLRP            *models.ActualLRP
			removedActualLRP             *models.ActualLRP
			suspectKeysWithExistingCells []*models.ActualLRPKey
		)

		BeforeEach(func() {
			ordinaryActualLRP = model_helpers.NewValidActualLRP("suspect-1", 0)
			suspectActualLRP = model_helpers.NewValidActualLRP("suspect-1", 0)
			removedActualLRP = model_helpers.NewValidActualLRP("removed-1", 0)
			suspectActualLRP.Presence = models.ActualLRP_Suspect

			suspectKeysWithExistingCells = []*models.ActualLRPKey{&suspectActualLRP.ActualLRPKey}

			fakeLRPDB.ActualLRPsReturns([]*models.ActualLRP{ordinaryActualLRP}, nil)
			fakeLRPDB.ConvergeLRPsReturns(db.ConvergenceResult{
				SuspectKeysWithExistingCells: suspectKeysWithExistingCells,
			})
		})

		Context("suspect LRP can not be promoted to Ordinary", func() {
			BeforeEach(func() {
				fakeSuspectDB.PromoteSuspectActualLRPReturns(nil, nil, nil, errors.New("failed-to-promote-suspect-lrp"))
			})

			It("does not emit any events", func() {
				Eventually(actualLRPInstanceHub.EmitCallCount).Should(Equal(0))
				Consistently(actualLRPInstanceHub.EmitCallCount).Should(Equal(0))
			})
		})

		Context("when the suspect LRP promoted to Ordinary and Ordinary is removed", func() {
			BeforeEach(func() {
				fakeSuspectDB.PromoteSuspectActualLRPReturns(suspectActualLRP, ordinaryActualLRP, removedActualLRP, nil)
			})

			It("emits event for removed LRP", func() {
				Eventually(actualLRPInstanceHub.EmitCallCount).Should(Equal(2))
				Consistently(actualLRPInstanceHub.EmitCallCount).Should(Equal(2))

				Expect(actualLRPInstanceHub.EmitArgsForCall(0)).To(Or(Equal(
					models.NewActualLRPInstanceChangedEvent(suspectActualLRP, ordinaryActualLRP, traceId),
				), Equal(
					models.NewActualLRPInstanceRemovedEvent(removedActualLRP, traceId),
				)))
				Expect(actualLRPInstanceHub.EmitArgsForCall(1)).To(Or(Equal(
					models.NewActualLRPInstanceChangedEvent(suspectActualLRP, ordinaryActualLRP, traceId),
				), Equal(
					models.NewActualLRPInstanceRemovedEvent(removedActualLRP, traceId),
				)))
				Expect(actualLRPInstanceHub.EmitArgsForCall(1)).ToNot(Equal(actualLRPInstanceHub.EmitArgsForCall(0)))
			})
		})

		Context("when the suspect LRP promoted to Ordinary and Ordinary is not removed", func() {
			BeforeEach(func() {
				fakeSuspectDB.PromoteSuspectActualLRPReturns(suspectActualLRP, ordinaryActualLRP, nil, nil)
			})

			It("emits event for changed LRP", func() {
				Eventually(actualLRPInstanceHub.EmitCallCount).Should(Equal(1))
				Consistently(actualLRPInstanceHub.EmitCallCount).Should(Equal(1))

				Expect(actualLRPInstanceHub.EmitArgsForCall(0)).To(Equal(
					models.NewActualLRPInstanceChangedEvent(suspectActualLRP, ordinaryActualLRP, traceId),
				))
			})
		})
	})

	Context("there are extra suspect LRPs", func() {
		var (
			key                    *models.ActualLRPKey
			runningLRP, suspectLRP *models.ActualLRP
		)

		BeforeEach(func() {
			key = &models.ActualLRPKey{
				ProcessGuid: "some-guid",
				Index:       0,
				Domain:      "some-domain",
			}
			fakeLRPDB.ConvergeLRPsReturns(db.ConvergenceResult{
				SuspectLRPKeysToRetire: []*models.ActualLRPKey{key},
			})

			runningLRP = model_helpers.NewValidActualLRP("some-guid", 1)
			runningLRP.State = models.ActualLRPStateRunning
			runningLRP.Presence = models.ActualLRP_Ordinary
			suspectLRP = model_helpers.NewValidActualLRP("some-guid", 0)
			suspectLRP.State = models.ActualLRPStateRunning
			suspectLRP.Presence = models.ActualLRP_Suspect
		})

		Context("when removing a suspect lrp", func() {
			BeforeEach(func() {
				fakeSuspectDB.RemoveSuspectActualLRPReturns(suspectLRP, nil)
			})

			It("removes the suspect LRP", func() {
				Eventually(fakeSuspectDB.RemoveSuspectActualLRPCallCount).Should(Equal(1))
				_, _, lrpKey := fakeSuspectDB.RemoveSuspectActualLRPArgsForCall(0)
				Expect(lrpKey).To(Equal(key))
			})

			It("emits an ActualLRPRemovedEvent containing the suspect LRP", func() {
				Eventually(actualHub.EmitCallCount).Should(Equal(1))
				event := actualHub.EmitArgsForCall(0)
				Expect(event).To(Equal(models.NewActualLRPRemovedEvent(suspectLRP.ToActualLRPGroup())))
			})

			It("emits an ActualLRPInstanceRemovedEvent containing the suspect LRP", func() {
				Eventually(actualLRPInstanceHub.EmitCallCount).Should(Equal(1))
				event := actualLRPInstanceHub.EmitArgsForCall(0)
				Expect(event).To(Equal(models.NewActualLRPInstanceRemovedEvent(suspectLRP, traceId)))
			})
		})

		Context("when RemoveSuspectActualLRP returns an error", func() {
			BeforeEach(func() {
				fakeSuspectDB.RemoveSuspectActualLRPReturns(nil, errors.New("boooom!"))
			})

			It("does not emit an ActualLRPRemovedEvent", func() {
				Consistently(actualHub.EmitCallCount).Should(BeZero())
			})

			It("does not emit an ActualLRPInstanceRemovedEvent", func() {
				Consistently(actualLRPInstanceHub.EmitCallCount).Should(BeZero())
			})
		})
	})

	Context("when there are extra ordinary LRPs", func() {
		BeforeEach(func() {
			retiringActualLRP1 = model_helpers.NewValidActualLRP("to-retire-1", 0)
			retiringActualLRP2 = model_helpers.NewValidActualLRP("to-retire-2", 1)
			keysToRetire = []*models.ActualLRPKey{&retiringActualLRP1.ActualLRPKey, &retiringActualLRP2.ActualLRPKey}

			result := db.ConvergenceResult{
				KeysToRetire: keysToRetire,
			}
			fakeLRPDB.ConvergeLRPsReturns(result)
		})

		Context("when the retirer returns an error", func() {
			BeforeEach(func() {
				retirer.RetireActualLRPReturns(errors.New("BOOM!!!"))
			})

			It("should log the error", func() {
				Expect(logger.Buffer()).To(gbytes.Say("BOOM!!!"))
			})
		})

		It("stops the LRPs", func() {
			Eventually(retirer.RetireActualLRPCallCount()).Should(Equal(2))

			stoppedKeys := make([]*models.ActualLRPKey, 2)

			for i := 0; i < 2; i++ {
				_, _, key := retirer.RetireActualLRPArgsForCall(i)
				stoppedKeys[i] = key
			}

			Expect(stoppedKeys).To(ContainElement(&retiringActualLRP1.ActualLRPKey))
			Expect(stoppedKeys).To(ContainElement(&retiringActualLRP2.ActualLRPKey))
		})
	})

	Context("when the db returns events", func() {
		var (
			expectedRemovedEvent         *models.ActualLRPRemovedEvent
			expectedInstanceRemovedEvent *models.ActualLRPInstanceRemovedEvent
		)

		BeforeEach(func() {
			group1 := &models.ActualLRPGroup{Instance: model_helpers.NewValidActualLRP("evacuating-lrp", 0)}
			expectedRemovedEvent = models.NewActualLRPRemovedEvent(group1)

			lrp := model_helpers.NewValidActualLRP("evacuating-lrp", 0)
			expectedInstanceRemovedEvent = models.NewActualLRPInstanceRemovedEvent(lrp, traceId)

			events := []models.Event{expectedRemovedEvent}
			instanceEvents := []models.Event{expectedInstanceRemovedEvent}
			fakeLRPDB.ConvergeLRPsReturns(db.ConvergenceResult{
				Events:         events,
				InstanceEvents: instanceEvents,
			})
		})

		It("emits those events", func() {
			Eventually(actualHub.EmitCallCount).Should(Equal(1))
			event := actualHub.EmitArgsForCall(0)
			removedEvent, ok := event.(*models.ActualLRPRemovedEvent)
			Expect(ok).To(BeTrue())
			Expect(removedEvent).To(Equal(expectedRemovedEvent))
		})

		It("emits those instance events", func() {
			Eventually(actualLRPInstanceHub.EmitCallCount).Should(Equal(1))
			event := actualLRPInstanceHub.EmitArgsForCall(0)
			removedEvent, ok := event.(*models.ActualLRPInstanceRemovedEvent)
			Expect(ok).To(BeTrue())
			Expect(removedEvent).To(Equal(expectedInstanceRemovedEvent))
		})
	})

	Context("lrps with internal routes that needs updated", func() {
		var (
			actualLRPKeyWithInternalRoutes1, actualLRPKeyWithInternalRoutes2, actualLRPKeyWithInternalRoutes3 db.ActualLRPKeyWithInternalRoutes
			cell1Presence, cell2Presence, cell3Presence                                                       models.CellPresence
		)

		BeforeEach(func() {
			actualLRPKey1 := models.NewActualLRPKey("guid1", 1, "some-domain")
			actualLRPInstanceKey1 := models.ActualLRPInstanceKey{InstanceGuid: "ig-1", CellId: "cell-id-1"}
			actualLRPKey2 := models.NewActualLRPKey("guid2", 1, "some-domain")
			actualLRPInstanceKey2 := models.ActualLRPInstanceKey{InstanceGuid: "ig-2", CellId: "cell-id-2"}
			actualLRPKey3 := models.NewActualLRPKey("guid3", 1, "some-domain")
			actualLRPInstanceKey3 := models.ActualLRPInstanceKey{InstanceGuid: "ig-3", CellId: "cell-id-3"}

			desiredInternalRoutes := internalroutes.InternalRoutes{
				{Hostname: "some-internal-route.apps.internal"},
			}
			actualLRPKeyWithInternalRoutes1 = db.ActualLRPKeyWithInternalRoutes{Key: &actualLRPKey1, InstanceKey: &actualLRPInstanceKey1, DesiredInternalRoutes: desiredInternalRoutes}
			actualLRPKeyWithInternalRoutes2 = db.ActualLRPKeyWithInternalRoutes{Key: &actualLRPKey2, InstanceKey: &actualLRPInstanceKey2, DesiredInternalRoutes: desiredInternalRoutes}
			actualLRPKeyWithInternalRoutes3 = db.ActualLRPKeyWithInternalRoutes{Key: &actualLRPKey3, InstanceKey: &actualLRPInstanceKey3, DesiredInternalRoutes: desiredInternalRoutes}

			fakeLRPDB.ConvergeLRPsReturns(db.ConvergenceResult{
				KeysWithInternalRouteChanges: []*db.ActualLRPKeyWithInternalRoutes{&actualLRPKeyWithInternalRoutes1, &actualLRPKeyWithInternalRoutes2, &actualLRPKeyWithInternalRoutes3},
			})
			cell1Presence = models.NewCellPresence(actualLRPKeyWithInternalRoutes1.InstanceKey.CellId, "1.1.1.1", "rep-1.service.internal", "z1", models.CellCapacity{}, nil, nil, nil, nil)
			cell2Presence = models.NewCellPresence(actualLRPKeyWithInternalRoutes2.InstanceKey.CellId, "1.1.1.2", "rep-2.service.internal", "z2", models.CellCapacity{}, nil, nil, nil, nil)
			cell3Presence = models.NewCellPresence(actualLRPKeyWithInternalRoutes3.InstanceKey.CellId, "1.1.1.3", "rep-3.service.internal", "z3", models.CellCapacity{}, nil, nil, nil, nil)
			fakeServiceClient.CellByIdCalls(func(logger lager.Logger, cellId string) (*models.CellPresence, error) {
				switch cellId {
				case "cell-id-1":
					return &cell1Presence, nil
				case "cell-id-2":
					return &cell2Presence, nil
				case "cell-id-3":
					return &cell3Presence, nil
				}

				return nil, errors.New("wat")
			})
		})

		It("gets each actuallrp's cell presence", func() {
			Eventually(fakeServiceClient.CellByIdCallCount()).Should(Equal(3))

			cellIds := make([]string, 3)

			for i := 0; i < 3; i++ {
				_, id := fakeServiceClient.CellByIdArgsForCall(i)
				cellIds[i] = id
			}

			Expect(cellIds).To(ContainElement(actualLRPKeyWithInternalRoutes1.InstanceKey.CellId))
			Expect(cellIds).To(ContainElement(actualLRPKeyWithInternalRoutes2.InstanceKey.CellId))
			Expect(cellIds).To(ContainElement(actualLRPKeyWithInternalRoutes3.InstanceKey.CellId))
		})

		It("creates a rep client for each actuallrp's cell", func() {
			Eventually(fakeRepClientFactory.CreateClientCallCount()).Should(Equal(3))

			repAddresses := make([]string, 3)
			repURLs := make([]string, 3)
			traceIDs := make([]string, 3)

			for i := 0; i < 3; i++ {
				address, url, traceID := fakeRepClientFactory.CreateClientArgsForCall(i)
				repAddresses[i] = address
				repURLs[i] = url
				traceIDs[i] = traceID
			}

			Expect(repAddresses).To(ContainElement(cell1Presence.RepAddress))
			Expect(repAddresses).To(ContainElement(cell2Presence.RepAddress))
			Expect(repAddresses).To(ContainElement(cell3Presence.RepAddress))

			Expect(repURLs).To(ContainElement(cell1Presence.RepUrl))
			Expect(repURLs).To(ContainElement(cell2Presence.RepUrl))
			Expect(repURLs).To(ContainElement(cell3Presence.RepUrl))

			Expect(traceIDs[0]).To(BeEmpty())
			Expect(traceIDs[1]).To(BeEmpty())
			Expect(traceIDs[2]).To(BeEmpty())
		})

		It("calls UpdateLRPInstance on the rep client", func() {
			Eventually(fakeRepClient.UpdateLRPInstanceCallCount()).Should(Equal(3))

			updates := make([]rep.LRPUpdate, 3)

			for i := 0; i < 3; i++ {
				_, update := fakeRepClient.UpdateLRPInstanceArgsForCall(i)
				updates[i] = update
			}

			internalRoutes := internalroutes.InternalRoutes{internalroutes.InternalRoute{Hostname: "some-internal-route.apps.internal"}}
			expectedLRP1Update := rep.NewLRPUpdate(actualLRPKeyWithInternalRoutes1.InstanceKey.InstanceGuid, *actualLRPKeyWithInternalRoutes1.Key, internalRoutes, nil)
			expectedLRP2Update := rep.NewLRPUpdate(actualLRPKeyWithInternalRoutes2.InstanceKey.InstanceGuid, *actualLRPKeyWithInternalRoutes2.Key, internalRoutes, nil)
			expectedLRP3Update := rep.NewLRPUpdate(actualLRPKeyWithInternalRoutes3.InstanceKey.InstanceGuid, *actualLRPKeyWithInternalRoutes3.Key, internalRoutes, nil)

			Expect(updates).To(ContainElement(expectedLRP1Update))
			Expect(updates).To(ContainElement(expectedLRP2Update))
			Expect(updates).To(ContainElement(expectedLRP3Update))
		})

		Context("when fetching cell presence fails", func() {
			BeforeEach(func() {
				fakeServiceClient.CellByIdReturns(nil, errors.New("kaboom"))
			})

			It("does not call CreateClient", func() {
				Expect(fakeRepClientFactory.CreateClientCallCount()).To(Equal(0))
			})

			It("does not call UpdateLRPInstance", func() {
				Expect(fakeRepClient.UpdateLRPInstanceCallCount()).To(Equal(0))
			})

			It("logs the error", func() {
				Eventually(logger).Should(gbytes.Say("failed-fetching-cell-presence"))
				Eventually(logger).Should(gbytes.Say("kaboom"))
			})
		})

		Context("when creating rep client fails", func() {
			BeforeEach(func() {
				fakeRepClientFactory.CreateClientReturns(nil, errors.New("kaboom"))
			})

			It("does not call UpdateLRPInstance", func() {
				Expect(fakeRepClient.UpdateLRPInstanceCallCount()).To(Equal(0))
			})

			It("logs the error", func() {
				Eventually(logger).Should(gbytes.Say("create-rep-client-failed"))
				Eventually(logger).Should(gbytes.Say("kaboom"))
			})
		})

		Context("when NewLRPUpdate fails", func() {
			BeforeEach(func() {
				fakeRepClient.UpdateLRPInstanceReturns(errors.New("kaboom"))
			})

			It("logs the error", func() {
				Eventually(logger).Should(gbytes.Say("updating-lrp-instance"))
				Eventually(logger).Should(gbytes.Say("kaboom"))
			})
		})
	})

	Context("lrps with metric tags that need to be updated", func() {
		var (
			actualLRPKeyWithMetricTags1, actualLRPKeyWithMetricTags2, actualLRPKeyWithMetricTags3 db.ActualLRPKeyWithMetricTags
			cell1Presence, cell2Presence, cell3Presence                                           models.CellPresence
		)

		BeforeEach(func() {
			actualLRPKey1 := models.NewActualLRPKey("guid1", 1, "some-domain")
			actualLRPInstanceKey1 := models.ActualLRPInstanceKey{InstanceGuid: "ig-1", CellId: "cell-id-1"}
			actualLRPKey2 := models.NewActualLRPKey("guid2", 1, "some-domain")
			actualLRPInstanceKey2 := models.ActualLRPInstanceKey{InstanceGuid: "ig-2", CellId: "cell-id-2"}
			actualLRPKey3 := models.NewActualLRPKey("guid3", 1, "some-domain")
			actualLRPInstanceKey3 := models.ActualLRPInstanceKey{InstanceGuid: "ig-3", CellId: "cell-id-3"}

			desiredMetricTags := map[string]*models.MetricTagValue{"app_name": {Static: "some-app-name"}}
			actualLRPKeyWithMetricTags1 = db.ActualLRPKeyWithMetricTags{Key: &actualLRPKey1, InstanceKey: &actualLRPInstanceKey1, DesiredMetricTags: desiredMetricTags}
			actualLRPKeyWithMetricTags2 = db.ActualLRPKeyWithMetricTags{Key: &actualLRPKey2, InstanceKey: &actualLRPInstanceKey2, DesiredMetricTags: desiredMetricTags}
			actualLRPKeyWithMetricTags3 = db.ActualLRPKeyWithMetricTags{Key: &actualLRPKey3, InstanceKey: &actualLRPInstanceKey3, DesiredMetricTags: desiredMetricTags}

			fakeLRPDB.ConvergeLRPsReturns(db.ConvergenceResult{
				KeysWithMetricTagChanges: []*db.ActualLRPKeyWithMetricTags{&actualLRPKeyWithMetricTags1, &actualLRPKeyWithMetricTags2, &actualLRPKeyWithMetricTags3},
			})
			cell1Presence = models.NewCellPresence(actualLRPKeyWithMetricTags1.InstanceKey.CellId, "1.1.1.1", "rep-1.service.internal", "z1", models.CellCapacity{}, nil, nil, nil, nil)
			cell2Presence = models.NewCellPresence(actualLRPKeyWithMetricTags2.InstanceKey.CellId, "1.1.1.2", "rep-2.service.internal", "z2", models.CellCapacity{}, nil, nil, nil, nil)
			cell3Presence = models.NewCellPresence(actualLRPKeyWithMetricTags3.InstanceKey.CellId, "1.1.1.3", "rep-3.service.internal", "z3", models.CellCapacity{}, nil, nil, nil, nil)
			fakeServiceClient.CellByIdCalls(func(logger lager.Logger, cellId string) (*models.CellPresence, error) {
				switch cellId {
				case "cell-id-1":
					return &cell1Presence, nil
				case "cell-id-2":
					return &cell2Presence, nil
				case "cell-id-3":
					return &cell3Presence, nil
				}

				return nil, errors.New("wat")
			})
		})

		It("gets each actuallrp's cell presence", func() {
			Eventually(fakeServiceClient.CellByIdCallCount()).Should(Equal(3))

			cellIds := make([]string, 3)

			for i := 0; i < 3; i++ {
				_, id := fakeServiceClient.CellByIdArgsForCall(i)
				cellIds[i] = id
			}

			Expect(cellIds).To(ContainElement(actualLRPKeyWithMetricTags1.InstanceKey.CellId))
			Expect(cellIds).To(ContainElement(actualLRPKeyWithMetricTags2.InstanceKey.CellId))
			Expect(cellIds).To(ContainElement(actualLRPKeyWithMetricTags3.InstanceKey.CellId))
		})

		It("creates a rep client for each actuallrp's cell", func() {
			Eventually(fakeRepClientFactory.CreateClientCallCount()).Should(Equal(3))

			repAddresses := make([]string, 3)
			repURLs := make([]string, 3)
			traceIDs := make([]string, 3)

			for i := 0; i < 3; i++ {
				address, url, traceID := fakeRepClientFactory.CreateClientArgsForCall(i)
				repAddresses[i] = address
				repURLs[i] = url
				traceIDs[i] = traceID
			}

			Expect(repAddresses).To(ContainElement(cell1Presence.RepAddress))
			Expect(repAddresses).To(ContainElement(cell2Presence.RepAddress))
			Expect(repAddresses).To(ContainElement(cell3Presence.RepAddress))

			Expect(repURLs).To(ContainElement(cell1Presence.RepUrl))
			Expect(repURLs).To(ContainElement(cell2Presence.RepUrl))
			Expect(repURLs).To(ContainElement(cell3Presence.RepUrl))

			Expect(traceIDs[0]).To(BeEmpty())
			Expect(traceIDs[1]).To(BeEmpty())
			Expect(traceIDs[2]).To(BeEmpty())
		})

		It("calls UpdateLRPInstance on the rep client", func() {
			Eventually(fakeRepClient.UpdateLRPInstanceCallCount()).Should(Equal(3))

			updates := make([]rep.LRPUpdate, 3)

			for i := 0; i < 3; i++ {
				_, update := fakeRepClient.UpdateLRPInstanceArgsForCall(i)
				updates[i] = update
			}

			metricTags := map[string]string{"app_name": "some-app-name"}
			expectedLRP1Update := rep.NewLRPUpdate(actualLRPKeyWithMetricTags1.InstanceKey.InstanceGuid, *actualLRPKeyWithMetricTags1.Key, nil, metricTags)
			expectedLRP2Update := rep.NewLRPUpdate(actualLRPKeyWithMetricTags2.InstanceKey.InstanceGuid, *actualLRPKeyWithMetricTags2.Key, nil, metricTags)
			expectedLRP3Update := rep.NewLRPUpdate(actualLRPKeyWithMetricTags3.InstanceKey.InstanceGuid, *actualLRPKeyWithMetricTags3.Key, nil, metricTags)

			Expect(updates).To(ContainElement(expectedLRP1Update))
			Expect(updates).To(ContainElement(expectedLRP2Update))
			Expect(updates).To(ContainElement(expectedLRP3Update))
		})

		Context("when fetching cell presence fails", func() {
			BeforeEach(func() {
				fakeServiceClient.CellByIdReturns(nil, errors.New("kaboom"))
			})

			It("does not call CreateClient", func() {
				Expect(fakeRepClientFactory.CreateClientCallCount()).To(Equal(0))
			})

			It("does not call UpdateLRPInstance", func() {
				Expect(fakeRepClient.UpdateLRPInstanceCallCount()).To(Equal(0))
			})

			It("logs the error", func() {
				Eventually(logger).Should(gbytes.Say("failed-fetching-cell-presence"))
				Eventually(logger).Should(gbytes.Say("kaboom"))
			})
		})

		Context("when creating rep client fails", func() {
			BeforeEach(func() {
				fakeRepClientFactory.CreateClientReturns(nil, errors.New("kaboom"))
			})

			It("does not call UpdateLRPInstance", func() {
				Expect(fakeRepClient.UpdateLRPInstanceCallCount()).To(Equal(0))
			})

			It("logs the error", func() {
				Eventually(logger).Should(gbytes.Say("create-rep-client-failed"))
				Eventually(logger).Should(gbytes.Say("kaboom"))
			})
		})

		Context("when NewLRPUpdate fails", func() {
			BeforeEach(func() {
				fakeRepClient.UpdateLRPInstanceReturns(errors.New("kaboom"))
			})

			It("logs the error", func() {
				Eventually(logger).Should(gbytes.Say("updating-lrp-instance"))
				Eventually(logger).Should(gbytes.Say("kaboom"))
			})
		})
	})
})
