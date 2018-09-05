package controllers_test

import (
	"errors"

	"code.cloudfoundry.org/auctioneer"
	"code.cloudfoundry.org/auctioneer/auctioneerfakes"
	"code.cloudfoundry.org/bbs/controllers"
	"code.cloudfoundry.org/bbs/controllers/fakes"
	"code.cloudfoundry.org/bbs/db"
	"code.cloudfoundry.org/bbs/db/dbfakes"
	"code.cloudfoundry.org/bbs/events/eventfakes"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/models/test/model_helpers"
	"code.cloudfoundry.org/bbs/serviceclient/serviceclientfakes"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/rep/repfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("LRP Convergence Controllers", func() {
	var (
		logger               *lagertest.TestLogger
		fakeLRPDB            *dbfakes.FakeLRPDB
		fakeSuspectDB        *dbfakes.FakeSuspectDB
		actualHub            *eventfakes.FakeHub
		actualLRPInstanceHub *eventfakes.FakeHub
		retirer              *fakes.FakeRetirer
		fakeAuctioneerClient *auctioneerfakes.FakeClient

		keysToRetire         []*models.ActualLRPKey
		keysWithMissingCells []*models.ActualLRPKeyWithSchedulingInfo

		retiringActualLRP1 *models.ActualLRP
		retiringActualLRP2 *models.ActualLRP

		desiredLRP1 models.DesiredLRPSchedulingInfo
		cellSet     models.CellSet

		generateSuspectActualLRPs bool

		controller *controllers.LRPConvergenceController
	)

	BeforeEach(func() {
		fakeLRPDB = new(dbfakes.FakeLRPDB)
		fakeSuspectDB = new(dbfakes.FakeSuspectDB)
		fakeAuctioneerClient = new(auctioneerfakes.FakeClient)
		logger = lagertest.NewTestLogger("test")

		desiredLRP1 = model_helpers.NewValidDesiredLRP("to-unclaim-1").DesiredLRPSchedulingInfo()

		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))

		fakeServiceClient = new(serviceclientfakes.FakeServiceClient)
		fakeRepClientFactory = new(repfakes.FakeClientFactory)
		fakeRepClient = new(repfakes.FakeClient)
		fakeRepClientFactory.CreateClientReturns(fakeRepClient, nil)
		fakeServiceClient.CellByIdReturns(nil, errors.New("hi"))

		cellPresence := models.NewCellPresence("cell-id", "1.1.1.1", "", "z1", models.CellCapacity{}, nil, nil, nil, nil)
		cellSet = models.CellSet{"cell-id": &cellPresence}
		fakeServiceClient.CellsReturns(cellSet, nil)

		actualHub = &eventfakes.FakeHub{}
		actualLRPInstanceHub = &eventfakes.FakeHub{}
		retirer = &fakes.FakeRetirer{}

		generateSuspectActualLRPs = false
	})

	JustBeforeEach(func() {
		controller = controllers.NewLRPConvergenceController(
			logger,
			fakeLRPDB,
			fakeSuspectDB,
			actualHub,
			actualLRPInstanceHub,
			fakeAuctioneerClient,
			fakeServiceClient,
			retirer,
			2,
			generateSuspectActualLRPs,
		)
		controller.ConvergeLRPs(logger)
	})

	It("calls ConvergeLRPs", func() {
		Expect(fakeLRPDB.ConvergeLRPsCallCount()).To(Equal(1))
		_, actualCellSet := fakeLRPDB.ConvergeLRPsArgsForCall(0)
		Expect(actualCellSet).To(BeEquivalentTo(cellSet))
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

			_, startAuctions := fakeAuctioneerClient.RequestLRPAuctionsArgsForCall(0)
			Expect(startAuctions).To(HaveLen(1))
			request := auctioneer.NewLRPStartRequestFromModel(model_helpers.NewValidDesiredLRP("some-guid"), 0)
			Expect(startAuctions).To(ContainElement(&request))
		})

		It("transition the LRP to UNCLAIMED state", func() {
			Eventually(fakeLRPDB.UnclaimActualLRPCallCount).Should(Equal(1))
			_, actualKey := fakeLRPDB.UnclaimActualLRPArgsForCall(0)
			Expect(actualKey).To(Equal(key))
		})

		It("emits an LRPChanged event", func() {
			Eventually(actualHub.EmitCallCount).Should(Equal(1))
			event := actualHub.EmitArgsForCall(0)
			Expect(event).To(Equal(models.NewActualLRPChangedEvent(before.ToActualLRPGroup(), after.ToActualLRPGroup())))
		})

		It("emits an LRPInstanceChanged event", func() {
			Eventually(actualLRPInstanceHub.EmitCallCount).Should(Equal(1))
			event := actualLRPInstanceHub.EmitArgsForCall(0)
			Expect(event).To(Equal(models.NewActualLRPInstanceChangedEvent(before, after)))
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

				_, startAuctions := fakeAuctioneerClient.RequestLRPAuctionsArgsForCall(0)
				Expect(startAuctions).To(HaveLen(1))
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

			_, startAuctions := fakeAuctioneerClient.RequestLRPAuctionsArgsForCall(0)
			Expect(startAuctions).To(HaveLen(1))
			request := auctioneer.NewLRPStartRequestFromModel(model_helpers.NewValidDesiredLRP("some-guid"), 0)
			Expect(startAuctions).To(ContainElement(&request))
		})

		It("creates the LPR record in the database", func() {
			Eventually(fakeLRPDB.CreateUnclaimedActualLRPCallCount).Should(Equal(1))
			_, actualKey := fakeLRPDB.CreateUnclaimedActualLRPArgsForCall(0)
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
			Expect(event).To(Equal(models.NewActualLRPInstanceCreatedEvent(after)))
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
			_, actualCellSet := fakeLRPDB.ConvergeLRPsArgsForCall(0)
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
			suspectActualLRP *models.ActualLRP
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

		Context("and generateSuspectActualLRPs is disabled", func() {
			It("updates the actual lrp to be unclaimed", func() {
				Expect(fakeLRPDB.UnclaimActualLRPCallCount()).To(Equal(1))
				_, lrpKey := fakeLRPDB.UnclaimActualLRPArgsForCall(0)
				Expect(lrpKey).To(Equal(&suspectActualLRP.ActualLRPKey))
			})

			It("emites change events", func() {
				Eventually(actualHub.EmitCallCount).Should(Equal(1))
				event := actualHub.EmitArgsForCall(0)
				Expect(event.EventType()).To(Equal(models.EventTypeActualLRPChanged))
			})

			It("emites instance change events", func() {
				Eventually(actualLRPInstanceHub.EmitCallCount).Should(Equal(1))
				event := actualLRPInstanceHub.EmitArgsForCall(0)
				Expect(event.EventType()).To(Equal(models.EventTypeActualLRPInstanceChanged))
			})

			It("auctions the unclaimed lrp", func() {
				Expect(fakeAuctioneerClient.RequestLRPAuctionsCallCount()).To(Equal(1))

				unclaimedStartRequest := auctioneer.NewLRPStartRequestFromSchedulingInfo(&desiredLRP1, 0)

				keysToAuction := []*auctioneer.LRPStartRequest{&unclaimedStartRequest}

				_, startAuctions := fakeAuctioneerClient.RequestLRPAuctionsArgsForCall(0)
				Expect(startAuctions).To(ConsistOf(keysToAuction))
			})

		})

		Context("and generateSuspectActualLRPs is enabled", func() {
			BeforeEach(func() {
				generateSuspectActualLRPs = true
			})

			It("change the LRP presence to 'Suspect'", func() {
				Eventually(fakeLRPDB.ChangeActualLRPPresenceCallCount).Should(Equal(1))

				_, key, from, to := fakeLRPDB.ChangeActualLRPPresenceArgsForCall(0)
				Expect(from).To(Equal(models.ActualLRP_Ordinary))
				Expect(to).To(Equal(models.ActualLRP_Suspect))
				Expect(key).To(Equal(&suspectActualLRP.ActualLRPKey))
			})

			It("creates a new unclaimed LRP", func() {
				Expect(fakeLRPDB.CreateUnclaimedActualLRPCallCount()).To(Equal(1))
				_, lrpKey := fakeLRPDB.CreateUnclaimedActualLRPArgsForCall(0)

				Expect(lrpKey).To(Equal(&suspectActualLRP.ActualLRPKey))
			})

			It("auctions new lrps", func() {
				Expect(fakeAuctioneerClient.RequestLRPAuctionsCallCount()).To(Equal(1))

				unclaimedStartRequest := auctioneer.NewLRPStartRequestFromSchedulingInfo(&desiredLRP1, 0)

				keysToAuction := []*auctioneer.LRPStartRequest{&unclaimedStartRequest}

				_, startAuctions := fakeAuctioneerClient.RequestLRPAuctionsArgsForCall(0)
				Expect(startAuctions).To(ConsistOf(keysToAuction))
			})

			It("emits no events", func() {
				Consistently(actualHub.EmitCallCount).Should(Equal(0))
			})

			It("emits no instance  events", func() {
				Consistently(actualLRPInstanceHub.EmitCallCount).Should(Equal(0))
			})

			Context("when there already is a Suspect LRP", func() {
				var (
					before, after *models.ActualLRP
				)

				BeforeEach(func() {
					suspectLRPKeys := []*models.ActualLRPKey{
						&suspectActualLRP.ActualLRPKey,
					}
					fakeLRPDB.ConvergeLRPsReturns(db.ConvergenceResult{
						KeysWithMissingCells: keysWithMissingCells,
						SuspectKeys:          suspectLRPKeys,
					})
					before = &models.ActualLRP{State: models.ActualLRPStateClaimed}
					after = &models.ActualLRP{State: models.ActualLRPStateUnclaimed}
					fakeLRPDB.UnclaimActualLRPReturns(before, after, nil)
				})

				It("unclaims the lrp", func() {
					Expect(fakeLRPDB.UnclaimActualLRPCallCount()).To(Equal(1))
				})

				It("does not emit change events", func() {
					Consistently(actualHub.EmitCallCount).Should(Equal(0))
				})

				It("does not emit instance change events", func() {
					Consistently(actualLRPInstanceHub.EmitCallCount).Should(Equal(0))
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
	})

	Context("when there are suspect LRPs with existing cells", func() {
		var (
			suspectActualLRP             *models.ActualLRP
			ordinaryActualLRP            *models.ActualLRP
			suspectKeysWithExistingCells []*models.ActualLRPKey
		)

		BeforeEach(func() {
			ordinaryActualLRP = model_helpers.NewValidActualLRP("suspect-1", 0)
			suspectActualLRP = model_helpers.NewValidActualLRP("suspect-1", 0)
			suspectActualLRP.Presence = models.ActualLRP_Suspect

			suspectKeysWithExistingCells = []*models.ActualLRPKey{&suspectActualLRP.ActualLRPKey}

			fakeLRPDB.ActualLRPsReturns([]*models.ActualLRP{ordinaryActualLRP}, nil)
			fakeLRPDB.ChangeActualLRPPresenceReturns(suspectActualLRP, ordinaryActualLRP, nil)
			fakeLRPDB.ConvergeLRPsReturns(db.ConvergenceResult{
				SuspectKeysWithExistingCells: suspectKeysWithExistingCells,
			})
		})

		It("remove the Ordinary LRP", func() {
			Eventually(fakeLRPDB.RemoveActualLRPCallCount).Should(Equal(1))

			_, guid, index, key := fakeLRPDB.RemoveActualLRPArgsForCall(0)

			Expect(guid).To(Equal(ordinaryActualLRP.ProcessGuid))
			Expect(index).To(Equal(ordinaryActualLRP.Index))
			Expect(key).To(BeNil())
		})

		It("changes the suspect LRP presence to Ordinary", func() {
			Eventually(fakeLRPDB.ChangeActualLRPPresenceCallCount).Should(Equal(1))
			_, lrpKey, from, to := fakeLRPDB.ChangeActualLRPPresenceArgsForCall(0)

			Expect(lrpKey).To(Equal(&suspectActualLRP.ActualLRPKey))
			Expect(from).To(Equal(models.ActualLRP_Suspect))
			Expect(to).To(Equal(models.ActualLRP_Ordinary))
		})

		It("does not emit any events", func() {
			Consistently(actualHub.EmitCallCount).Should(Equal(0))
		})

		It("does not emit any instance events", func() {
			Consistently(actualLRPInstanceHub.EmitCallCount).Should(Equal(0))
		})

		Context("when the ordinary lrp cannot be removed", func() {
			BeforeEach(func() {
				fakeLRPDB.RemoveActualLRPReturns(errors.New("booom!"))
			})

			It("does not change the suspect LRP presence", func() {
				Consistently(fakeLRPDB.ChangeActualLRPPresenceCallCount).Should(BeZero())
			})
		})
	})

	Context("there are extra suspect LRPs", func() {
		var (
			key                    *models.ActualLRPKey
			runningLRP, suspectLRP *models.ActualLRP
		)

		BeforeEach(func() {
			lrp := model_helpers.NewValidDesiredLRP("some-guid")
			key = &models.ActualLRPKey{
				ProcessGuid: lrp.ProcessGuid,
				Index:       0,
				Domain:      lrp.Domain,
			}
			fakeLRPDB.ConvergeLRPsReturns(db.ConvergenceResult{
				SuspectLRPKeysToRetire: []*models.ActualLRPKey{key},
			})

			runningLRP = model_helpers.NewValidActualLRP("some-guid", 1)
			runningLRP.State = models.ActualLRPStateClaimed
			runningLRP.Presence = models.ActualLRP_Ordinary
			suspectLRP = model_helpers.NewValidActualLRP("some-guid", 0)
			suspectLRP.State = models.ActualLRPStateClaimed
			suspectLRP.Presence = models.ActualLRP_Suspect
		})

		Context("when removing a suspect lrp", func() {
			BeforeEach(func() {
				fakeSuspectDB.RemoveSuspectActualLRPReturns(suspectLRP, nil)
			})

			It("removes the suspect LRP", func() {
				Eventually(fakeSuspectDB.RemoveSuspectActualLRPCallCount).Should(Equal(1))
				_, lrpKey := fakeSuspectDB.RemoveSuspectActualLRPArgsForCall(0)
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
				Expect(event).To(Equal(models.NewActualLRPInstanceRemovedEvent(suspectLRP)))
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
				_, key := retirer.RetireActualLRPArgsForCall(i)
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
			expectedInstanceRemovedEvent = models.NewActualLRPInstanceRemovedEvent(lrp)

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
})
