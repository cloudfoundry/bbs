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
		err                  error
		logger               *lagertest.TestLogger
		fakeLRPDB            *dbfakes.FakeLRPDB
		fakeSuspectDB        *dbfakes.FakeSuspectDB
		actualHub            *eventfakes.FakeHub
		retirer              *fakes.FakeRetirer
		fakeAuctioneerClient *auctioneerfakes.FakeClient

		keysToAuction        []*auctioneer.LRPStartRequest
		keysToRetire         []*models.ActualLRPKey
		keysWithMissingCells []*models.ActualLRPKeyWithSchedulingInfo

		retiringActualLRP1 *models.ActualLRP
		retiringActualLRP2 *models.ActualLRP

		desiredLRP1, desiredLRP2 models.DesiredLRPSchedulingInfo
		cellSet                  models.CellSet

		controller *controllers.LRPConvergenceController
	)

	BeforeEach(func() {
		fakeLRPDB = new(dbfakes.FakeLRPDB)
		fakeSuspectDB = new(dbfakes.FakeSuspectDB)
		fakeAuctioneerClient = new(auctioneerfakes.FakeClient)
		logger = lagertest.NewTestLogger("test")

		// request1 := auctioneer.NewLRPStartRequestFromModel(model_helpers.NewValidDesiredLRP("to-auction-1"), 1, 2)
		// request2 := auctioneer.NewLRPStartRequestFromModel(model_helpers.NewValidDesiredLRP("to-auction-2"), 0, 4)
		// keysToAuction = []*auctioneer.LRPStartRequest{&request1, &request2}

		desiredLRP1 = model_helpers.NewValidDesiredLRP("to-unclaim-1").DesiredLRPSchedulingInfo()
		desiredLRP2 = model_helpers.NewValidDesiredLRP("to-unclaim-2").DesiredLRPSchedulingInfo()

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
		retirer = &fakes.FakeRetirer{}
		controller = controllers.NewLRPConvergenceController(
			logger,
			fakeLRPDB,
			fakeSuspectDB,
			actualHub,
			fakeAuctioneerClient,
			fakeServiceClient,
			retirer,
			2,
		)
	})

	JustBeforeEach(func() {
		err = controller.ConvergeLRPs(logger)
	})

	It("calls ConvergeLRPs", func() {
		Expect(err).NotTo(HaveOccurred())
		Expect(fakeLRPDB.ConvergeLRPsCallCount()).To(Equal(1))
		_, actualCellSet := fakeLRPDB.ConvergeLRPsArgsForCall(0)
		Expect(actualCellSet).To(BeEquivalentTo(cellSet))
	})

	Context("when there are unstarted ActualLRPs", func() {
		var (
			key            *models.ActualLRPKey
			schedulingInfo models.DesiredLRPSchedulingInfo
			before, after  *models.ActualLRPGroup
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

			crashedLRP := model_helpers.NewValidActualLRP("some-guid", 0)
			crashedLRP.State = models.ActualLRPStateCrashed
			before = &models.ActualLRPGroup{Instance: crashedLRP}
			unclaimedLRP := *crashedLRP
			unclaimedLRP.State = models.ActualLRPStateUnclaimed
			after = &models.ActualLRPGroup{Instance: &unclaimedLRP}
			fakeLRPDB.UnclaimActualLRPReturns(before, after, nil)
		})

		It("auctions off the returned keys", func() {
			Expect(fakeAuctioneerClient.RequestLRPAuctionsCallCount()).To(Equal(1))

			_, startAuctions := fakeAuctioneerClient.RequestLRPAuctionsArgsForCall(0)
			Expect(startAuctions).To(HaveLen(1))
			request := auctioneer.NewLRPStartRequestFromModel(model_helpers.NewValidDesiredLRP("some-guid"), 0)
			Expect(startAuctions).To(ContainElement(&request))
		})

		It("transition the LPR to UNCLAIMED state", func() {
			Eventually(fakeLRPDB.UnclaimActualLRPCallCount).Should(Equal(1))
			_, actualKey := fakeLRPDB.UnclaimActualLRPArgsForCall(0)
			Expect(actualKey).To(Equal(key))
		})

		It("emits an LRPChanged event", func() {
			Eventually(actualHub.EmitCallCount).Should(Equal(1))
			event := actualHub.EmitArgsForCall(0)
			Expect(event).To(Equal(models.NewActualLRPChangedEvent(before, after)))
		})
	})

	Context("when there are missing ActualLRPs", func() {
		var (
			key            *models.ActualLRPKey
			schedulingInfo models.DesiredLRPSchedulingInfo
			after          *models.ActualLRPGroup
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

			unclaimedLRP := model_helpers.NewValidActualLRP("some-guid", 0)
			unclaimedLRP.State = models.ActualLRPStateUnclaimed
			after = &models.ActualLRPGroup{Instance: unclaimedLRP}
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
			Expect(event).To(Equal(models.NewActualLRPCreatedEvent(after)))
		})
	})

	Context("when fetching the cells fails", func() {
		BeforeEach(func() {
			fakeServiceClient.CellsReturns(nil, errors.New("kaboom"))
		})

		It("does not return an error", func() {
			Expect(err).NotTo(HaveOccurred())
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
			Expect(err).NotTo(HaveOccurred())
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

	// Row 1 in https://docs.google.com/document/d/19880DjH4nJKzsDP8BT09m28jBlFfSiVx64skbvilbnA
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

		It("change the LRP presence to 'Suspect'", func() {
			Eventually(fakeLRPDB.ChangeActualLRPPresenceCallCount).Should(Equal(1))

			_, key, from, to := fakeLRPDB.ChangeActualLRPPresenceArgsForCall(0)
			Expect(from).To(Equal(models.ActualLRP_Ordinary))
			Expect(to).To(Equal(models.ActualLRP_Suspect))
			Expect(key).To(Equal(&suspectActualLRP.ActualLRPKey))
		})

		FIt("creates a new unclaimed LRP", func() {
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

		// Row 5 in
		// https://docs.google.com/document/d/19880DjH4nJKzsDP8BT09m28jBlFfSiVx64skbvilbnA. This
		// is implemented by always unclaiming LRPs that cannot be transitioned to
		// Suspect
		Context("when change the actual lrp presence fail because there is already a Suspect LRP", func() {
			var (
				before, after *models.ActualLRPGroup
			)

			BeforeEach(func() {
				fakeLRPDB.ChangeActualLRPPresenceReturns(nil, nil, models.ErrResourceExists)
				fakeLRPDB.ConvergeLRPsReturns(db.ConvergenceResult{
					KeysWithMissingCells: keysWithMissingCells,
				})
				before = &models.ActualLRPGroup{Instance: &models.ActualLRP{State: models.ActualLRPStateClaimed}}
				after = &models.ActualLRPGroup{Instance: &models.ActualLRP{State: models.ActualLRPStateUnclaimed}}
				fakeLRPDB.UnclaimActualLRPReturns(before, after, nil)
			})

			It("unclaims the lrp", func() {
				Expect(fakeLRPDB.UnclaimActualLRPCallCount()).To(Equal(1))
			})

			It("emits a ActualLRPChangedEvent", func() {
				Eventually(actualHub.EmitCallCount).Should(Equal(1))
				event := actualHub.EmitArgsForCall(0)
				Expect(event).To(Equal(models.NewActualLRPChangedEvent(before, after)))
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

			Context("when there are missing LRPs that need to be auctioned", func() {
				BeforeEach(func() {
					lrp := model_helpers.NewValidDesiredLRP("missing-lrp")
					schedulingInfo := lrp.DesiredLRPSchedulingInfo()
					missingLRPKeys := []*models.ActualLRPKeyWithSchedulingInfo{
						{Key: &models.ActualLRPKey{ProcessGuid: lrp.ProcessGuid, Index: 0, Domain: lrp.Domain}, SchedulingInfo: &schedulingInfo},
					}
					unclaimedStartRequest := auctioneer.NewLRPStartRequestFromSchedulingInfo(&schedulingInfo, 0)
					keysToAuction = []*auctioneer.LRPStartRequest{&unclaimedStartRequest}

					fakeLRPDB.ConvergeLRPsReturns(db.ConvergenceResult{
						KeysWithMissingCells: keysWithMissingCells,
						MissingLRPKeys:       missingLRPKeys,
					})
				})

				It("auctions off any missing lrp and does not return early", func() {
					// how can it auction off keys if it returns before adding anything to the startRequests?
					Expect(fakeAuctioneerClient.RequestLRPAuctionsCallCount()).To(Equal(1))

					_, startAuctions := fakeAuctioneerClient.RequestLRPAuctionsArgsForCall(0)
					Expect(startAuctions).To(HaveLen(1))
					Expect(startAuctions).To(ConsistOf(keysToAuction))
				})
			})
		})

		Context("when the DB returns an unrecoverable error", func() {
			BeforeEach(func() {
				fakeLRPDB.ChangeActualLRPPresenceReturns(nil, nil, models.NewUnrecoverableError(nil))
			})

			It("logs the error", func() {
				Eventually(logger).Should(gbytes.Say("unrecoverable-error"))
			})

			It("returns the error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).Should(Equal(models.NewUnrecoverableError(nil)))
			})
		})
	})

	// Row 2 in https://docs.google.com/document/d/19880DjH4nJKzsDP8BT09m28jBlFfSiVx64skbvilbnA
	Context("when there are suspect LRPs with existing cells", func() {
		var (
			suspectActualLRP             *models.ActualLRP
			ordinaryActualLRP            *models.ActualLRP
			suspectKeysWithExistingCells []*models.ActualLRPKey
		)

		BeforeEach(func() {
			ordinaryActualLRP = model_helpers.NewValidActualLRP("suspect-1", 0)
			suspectActualLRP = model_helpers.NewValidActualLRP("suspect-1", 0)
			group := &models.ActualLRPGroup{Instance: suspectActualLRP}

			suspectKeysWithExistingCells = []*models.ActualLRPKey{&suspectActualLRP.ActualLRPKey}

			fakeLRPDB.ActualLRPGroupByProcessGuidAndIndexReturns(&models.ActualLRPGroup{Instance: ordinaryActualLRP}, nil)
			fakeLRPDB.ChangeActualLRPPresenceReturns(group, group, nil)
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

		Context("when the suspect lrp cannot be changed to ordinary lrp", func() {
			Context("and the error is unrecoverable", func() {
				BeforeEach(func() {
					fakeLRPDB.ChangeActualLRPPresenceReturns(nil, nil, models.NewUnrecoverableError(nil))
				})

				It("logs the error", func() {
					Eventually(logger).Should(gbytes.Say("unrecoverable-error"))
				})

				It("returns the error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).Should(Equal(models.NewUnrecoverableError(nil)))
				})
			})
		})

		Context("when the ordinary lrp cannot be removed", func() {
			BeforeEach(func() {
				fakeLRPDB.RemoveActualLRPReturns(errors.New("booom!"))
			})

			It("does not change the suspect LRP presence", func() {
				Consistently(fakeLRPDB.ChangeActualLRPPresenceCallCount).Should(BeZero())
			})

			Context("and the error is unrecoverable", func() {
				BeforeEach(func() {
					fakeLRPDB.RemoveActualLRPReturns(models.NewUnrecoverableError(nil))
				})

				It("logs the error", func() {
					Eventually(logger).Should(gbytes.Say("unrecoverable-error"))
				})

				It("returns the error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).Should(Equal(models.NewUnrecoverableError(nil)))
				})
			})
		})
	})

	// Row 3 in https://docs.google.com/document/d/19880DjH4nJKzsDP8BT09m28jBlFfSiVx64skbvilbnA
	Context("there are extra suspect LRPs", func() {
		var (
			key                    *models.ActualLRPKey
			after                  *models.ActualLRPGroup
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

		It("removes the suspect LRP", func() {
			Eventually(fakeSuspectDB.RemoveSuspectActualLRPCallCount).Should(Equal(1))
			_, lrpKey := fakeSuspectDB.RemoveSuspectActualLRPArgsForCall(0)
			Expect(lrpKey).To(Equal(key))
		})

		It("emits an ActualLRPRemovedEvent containing the suspect LRP", func() {
			Eventually(actualHub.EmitCallCount).Should(Equal(1))
			event := actualHub.EmitArgsForCall(0)
			Expect(event).To(Equal(models.NewActualLRPRemovedEvent(after)))
		})
	})

	// NOP: Row 4 in https://docs.google.com/document/d/19880DjH4nJKzsDP8BT09m28jBlFfSiVx64skbvilbnA

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

			It("should return the error", func() {
				Expect(err).NotTo(HaveOccurred())
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
		var expectedRemovedEvent *models.ActualLRPRemovedEvent
		BeforeEach(func() {
			group1 := &models.ActualLRPGroup{Instance: model_helpers.NewValidActualLRP("evacuating-lrp", 0)}
			expectedRemovedEvent = models.NewActualLRPRemovedEvent(group1)

			events := []models.Event{expectedRemovedEvent}
			fakeLRPDB.ConvergeLRPsReturns(db.ConvergenceResult{
				Events: events,
			})
		})

		It("emits those events", func() {

			Eventually(actualHub.EmitCallCount).Should(Equal(1))
			event := actualHub.EmitArgsForCall(0)
			removedEvent, ok := event.(*models.ActualLRPRemovedEvent)
			Expect(ok).To(BeTrue())
			Expect(removedEvent).To(Equal(expectedRemovedEvent))
		})
	})

	// FContext("Given two containers i1 on cell1 and i2 on cell2", func() {
	// 	// things to define for the test
	// 	// before {Instance:..., Suspect: ..., Evacuating: ...}
	// 	// lrp for i1
	// 	// lrp for i2
	// 	// cell presence for cell1
	// 	// cell presence for cell2
	// 	// func (convergence ||
	// 	// -> expected:
	// 	//	aftter {Instance, Suspect: ,Evcuating}
	// 	//	event emission ( route emitter and ohers )

	// 	var (
	// 		key            *models.ActualLRPKey
	// 		schedulingInfo models.DesiredLRPSchedulingInfo
	// 		before, after  *models.ActualLRPGroup
	// 		i1LRP/*, i2LRP*/ *models.ActualLRP
	// 	)

	// 	Context("when cell2 is Present and cell1 is Missing", func() {
	// 		// Row 1
	// 		Context("when container i1 is running and there is no container i2", func() {
	// 			BeforeEach(func() {
	// 				lrp := model_helpers.NewValidDesiredLRP("some-guid")
	// 				schedulingInfo = lrp.DesiredLRPSchedulingInfo()
	// 				key = &models.ActualLRPKey{
	// 					ProcessGuid: lrp.ProcessGuid,
	// 					Index:       0,
	// 					Domain:      lrp.Domain,
	// 				}

	// 				lrpKeyWithSchedulingInfo := &models.ActualLRPKeyWithSchedulingInfo{
	// 					Key:            key,
	// 					SchedulingInfo: &schedulingInfo,
	// 				}
	// 				i1LRP = &models.ActualLRP{
	// 					ActualLRPKey: *key,
	// 					State:        models.ActualLRPStateRunning,
	// 					Presence:     models.ActualLRP_Ordinary,
	// 				}
	// 				before = &models.ActualLRPGroup{Instance: i1LRP}
	// 				after = &models.ActualLRPGroup{Instance: nil}

	// 				fakeLRPDB.ConvergeLRPsReturns(db.ConvergenceResult{

	// 					KeysWithMissingCells: []*models.ActualLRPKeyWithSchedulingInfo{lrpKeyWithSchedulingInfo},
	// 					//SuspectKeysWithExistingCells: []*models.ActualLRPKey{key},
	// 				})
	// 			})
	// 			It("Move instance i1 to suspect in the after ActualLRPGroup", func() {
	// 				// Expect(i1LRP.Presence).To(Equal(models.ActualLRP_Suspect))
	// 				_, calledKey, calledPresence := fakeLRPDB.ChangeActualLRPPresenceArgsForCall(0)
	// 				// Expect(calledLogger).To(Equal(logger.Session("converge-lrps")))
	// 				Expect(calledKey).To(Equal(key))
	// 				Expect(calledPresence).To(Equal(models.ActualLRP_Suspect))
	// 			})
	// 			It("Creates an Unclaimed Actual LRP  instance i2", func() {

	// 			})

	// 			It("emits an actualLRPChanged event", func() {
	// 				Eventually(actualHub.EmitCallCount).Should(Equal(1))
	// 				event := actualHub.EmitArgsForCall(0)
	// 				Expect(event).To(Equal(models.NewActualLRPChangedEvent(before, after)))
	// 			})

	// 			It("emits an ActualLRPCreated event", func() {
	// 				Eventually(actualHub.EmitCallCount).Should(Equal(1))
	// 				event := actualHub.EmitArgsForCall(0)
	// 				Expect(event).To(Equal(models.NewActualLRPCreatedEvent(after)))
	// 			})
	// 		})
	// 		// Row 4
	// 		Context("when container i1 is Running on a still missing cell and i2 is Created and clainmed", func() {
	// 			It("does nothing, the before and after ActualLRPGroup are identical", func() {})
	// 		})
	// 	})

	// 	Context("when cell2 is Present and cell1 is Present", func() {
	// 		// Row 2
	// 		Context("when container i1 is Running and was suspect, and i2 is CLAIMED and CREATED ", func() {
	// 			It("Removes instance for i2", func() {})
	// 			It("moves suspect i1 back to instance", func() {})
	// 			It("emits an actualLRPChanged event", func() {})
	// 			It("emits an ActualLRPRemovedEvent for i2", func() {})
	// 		})
	// 		// Row 3
	// 		Context("when container i1 is Running and was suspected and container i2 is Running", func() {
	// 			It("removes suspect", func() {})
	// 			It("emits an ActualLRPRemovedEvent containing i1", func() {})
	// 		})
	// 	})
	// 	Context("when cell2 is Missing and cell1 is also Missing", func() {
	// 		// Row 5
	// 		Context("and container i1 is Running and suspect, and container i2 is Created and Claimed", func() {
	// 			It("Removes instance i2", func() {})
	// 			It("Creates and reauction a new actual LRP and keep the suspect", func() {})
	// 			It("keeps the suspect i1", func() {})
	// 			It("Emits an actualLRPChanged event for i2", func() {})
	// 		})
	// 	})
	// })
})
