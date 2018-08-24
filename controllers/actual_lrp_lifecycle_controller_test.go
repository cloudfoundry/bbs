package controllers_test

import (
	"errors"

	"code.cloudfoundry.org/auctioneer"
	"code.cloudfoundry.org/auctioneer/auctioneerfakes"
	"code.cloudfoundry.org/bbs/controllers"
	"code.cloudfoundry.org/bbs/db/dbfakes"
	"code.cloudfoundry.org/bbs/events/eventfakes"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/serviceclient/serviceclientfakes"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/rep/repfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("ActualLRP Lifecycle Controller", func() {
	var (
		logger               *lagertest.TestLogger
		fakeActualLRPDB      *dbfakes.FakeActualLRPDB
		fakeDesiredLRPDB     *dbfakes.FakeDesiredLRPDB
		fakeEvacuationDB     *dbfakes.FakeEvacuationDB
		fakeSuspectDB        *dbfakes.FakeSuspectDB
		fakeAuctioneerClient *auctioneerfakes.FakeClient
		actualHub            *eventfakes.FakeHub

		actualLRP      models.ActualLRP
		afterActualLRP models.ActualLRP
		controller     *controllers.ActualLRPLifecycleController
		err            error
	)

	BeforeEach(func() {
		fakeActualLRPDB = new(dbfakes.FakeActualLRPDB)
		fakeSuspectDB = new(dbfakes.FakeSuspectDB)
		fakeDesiredLRPDB = new(dbfakes.FakeDesiredLRPDB)
		fakeEvacuationDB = new(dbfakes.FakeEvacuationDB)
		fakeAuctioneerClient = new(auctioneerfakes.FakeClient)
		logger = lagertest.NewTestLogger("test")

		fakeServiceClient = new(serviceclientfakes.FakeServiceClient)
		fakeRepClientFactory = new(repfakes.FakeClientFactory)
		fakeRepClient = new(repfakes.FakeClient)
		fakeRepClientFactory.CreateClientReturns(fakeRepClient, nil)

		actualHub = &eventfakes.FakeHub{}
		controller = controllers.NewActualLRPLifecycleController(
			fakeActualLRPDB,
			fakeSuspectDB,
			fakeEvacuationDB,
			fakeDesiredLRPDB,
			fakeAuctioneerClient,
			fakeServiceClient,
			fakeRepClientFactory,
			actualHub,
		)
	})

	Describe("ClaimActualLRP", func() {
		var (
			processGuid       = "process-guid"
			index       int32 = 1
			instanceKey models.ActualLRPInstanceKey
		)

		BeforeEach(func() {
			instanceKey = models.NewActualLRPInstanceKey(
				"instance-guid-0",
				"cell-id-0",
			)
			actualLRPKey := models.NewActualLRPKey(
				processGuid,
				1,
				"domain-0",
			)
			actualLRP = models.ActualLRP{
				ActualLRPKey: actualLRPKey,
				State:        models.ActualLRPStateUnclaimed,
				Since:        1138,
			}
			afterActualLRP = models.ActualLRP{
				ActualLRPKey:         actualLRPKey,
				ActualLRPInstanceKey: instanceKey,
				State:                models.ActualLRPStateClaimed,
				Since:                1140,
			}

			fakeActualLRPDB.ActualLRPsReturns([]*models.ActualLRP{&afterActualLRP}, nil)
		})

		JustBeforeEach(func() {
			err = controller.ClaimActualLRP(logger, processGuid, index, &instanceKey)
		})

		Context("when there is a running Suspect LRP", func() {
			BeforeEach(func() {
				suspect := &models.ActualLRP{
					Presence:     models.ActualLRP_Suspect,
					ActualLRPKey: actualLRP.ActualLRPKey,
					ActualLRPInstanceKey: models.ActualLRPInstanceKey{
						InstanceGuid: "suspect-ig",
						CellId:       "suspect-cell-id",
					},
				}
				fakeActualLRPDB.ActualLRPsReturns([]*models.ActualLRP{suspect}, nil)
			})

			It("does not emit ActualLRPChangedEvent", func() {
				Consistently(actualHub.EmitCallCount).Should(BeZero())
			})
		})

		Context("when claiming the actual lrp in the DB succeeds", func() {
			BeforeEach(func() {
				fakeActualLRPDB.ClaimActualLRPReturns(&actualLRP, &afterActualLRP, nil)
			})

			It("calls the DB successfully", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeActualLRPDB.ClaimActualLRPCallCount()).To(Equal(1))
			})

			It("emits a change to the hub", func() {
				Eventually(actualHub.EmitCallCount).Should(Equal(1))
				event := actualHub.EmitArgsForCall(0)
				changedEvent, ok := event.(*models.ActualLRPChangedEvent)
				Expect(ok).To(BeTrue())
				Expect(changedEvent.Before).To(Equal(actualLRP.ToActualLRPGroup()))
				Expect(changedEvent.After).To(Equal(afterActualLRP.ToActualLRPGroup()))
			})

			Context("when the actual lrp did not actually change", func() {
				BeforeEach(func() {
					fakeActualLRPDB.ClaimActualLRPReturns(
						&afterActualLRP,
						&afterActualLRP,
						nil,
					)
				})

				It("does not emit a change event to the hub", func() {
					Eventually(actualHub.EmitCallCount).Should(Equal(0))
				})
			})
		})
	})

	Describe("StartActualLRP", func() {
		var (
			processGuid = "process-guid"
			index       = int32(1)

			key         models.ActualLRPKey
			instanceKey models.ActualLRPInstanceKey
			netInfo     models.ActualLRPNetInfo
			err         error
		)

		BeforeEach(func() {
			key = models.NewActualLRPKey(
				processGuid,
				index,
				"domain-0",
			)
			instanceKey = models.NewActualLRPInstanceKey(
				"instance-guid-0",
				"cell-id-0",
			)
			netInfo = models.NewActualLRPNetInfo("1.1.1.1", "2.2.2.2", models.NewPortMapping(10, 20))

			actualLRP = models.ActualLRP{
				ActualLRPKey: key,
				State:        models.ActualLRPStateUnclaimed,
				Since:        1138,
			}

			afterActualLRP = models.ActualLRP{
				ActualLRPKey:         key,
				ActualLRPInstanceKey: instanceKey,
				ActualLRPNetInfo:     netInfo,
				State:                models.ActualLRPStateRunning,
				Since:                1139,
			}
		})

		JustBeforeEach(func() {
			err = controller.StartActualLRP(logger, &key, &instanceKey, &netInfo)
		})

		Context("when there is a Suspect LRP running", func() {
			BeforeEach(func() {
				suspect := &models.ActualLRP{
					Presence: models.ActualLRP_Suspect,
					ActualLRPInstanceKey: models.ActualLRPInstanceKey{
						InstanceGuid: "suspect-instance-guid",
						CellId:       "cell-id-1",
					},
					ActualLRPKey: models.ActualLRPKey{
						ProcessGuid: processGuid,
						Index:       index,
						Domain:      "domain-0",
					},
				}
				fakeActualLRPDB.ActualLRPsReturns([]*models.ActualLRP{suspect}, nil)
				fakeSuspectDB.RemoveSuspectActualLRPReturns(suspect, nil)
			})

			It("removes the suspect lrp", func() {
				Eventually(fakeSuspectDB.RemoveSuspectActualLRPCallCount).Should(Equal(1))
				_, lrpKey := fakeSuspectDB.RemoveSuspectActualLRPArgsForCall(0)
				Expect(lrpKey).To(Equal(&models.ActualLRPKey{
					ProcessGuid: processGuid,
					Index:       index,
					Domain:      "domain-0",
				}))
			})

			It("emits ActualLRPCreatedEvent", func() {
				Eventually(actualHub.EmitCallCount).Should(Equal(2))
				var e *models.ActualLRPCreatedEvent
				Expect(actualHub.EmitArgsForCall(0)).To(BeAssignableToTypeOf(e))
			})

			Context("when RemoveSuspectActualLRP returns an error", func() {
				BeforeEach(func() {
					fakeSuspectDB.RemoveSuspectActualLRPReturns(nil, errors.New("boooom!"))
				})

				It("logs the error", func() {
					Expect(logger.Buffer()).Should(gbytes.Say("boooom!"))
				})

				It("does not emit the ActualLRPRemovedEvent", func() {
					Eventually(actualHub.EmitCallCount).Should(Equal(1))

					event := actualHub.EmitArgsForCall(0)
					Expect(event).To(BeAssignableToTypeOf(&models.ActualLRPCreatedEvent{}))

					Consistently(actualHub.EmitCallCount).Should(Equal(1))
				})
			})
		})

		Context("when the LRP being started is Suspect", func() {
			BeforeEach(func() {
				instanceKey = models.NewActualLRPInstanceKey(
					"instance-guid-0",
					"cell-id-0",
				)
				actualLRPKey := models.NewActualLRPKey(
					processGuid,
					1,
					"domain-0",
				)
				actualLRP = models.ActualLRP{
					ActualLRPKey:         actualLRPKey,
					ActualLRPInstanceKey: instanceKey,
				}
				fakeActualLRPDB.ActualLRPsReturns([]*models.ActualLRP{&actualLRP}, nil)
			})

			Context("when there is a Running Ordinary LRP", func() {
				BeforeEach(func() {
					// the db layer resolution logic will return the Ordinary LRP
					actualLRP.Presence = models.ActualLRP_Ordinary
					fakeActualLRPDB.StartActualLRPReturns(nil, nil, models.ErrActualLRPCannotBeStarted)
				})

				It("returns ErrActualLRPCannotBeStarted", func() {
					Expect(err).To(MatchError(models.ErrActualLRPCannotBeStarted))
				})
			})

			Context("and the Ordinary LRP is not running", func() {
				BeforeEach(func() {
					// the db layer resolution logic will return the Suspect LRP
					actualLRP.Presence = models.ActualLRP_Suspect
					fakeActualLRPDB.ActualLRPsReturns([]*models.ActualLRP{&actualLRP}, nil)
				})

				It("don't do anything", func() {
					Expect(fakeActualLRPDB.StartActualLRPCallCount()).To(BeZero())
				})
			})
		})

		Context("when starting the actual lrp in the DB succeeds", func() {
			BeforeEach(func() {
				newActualLRP := actualLRP
				newAfterActualLRP := afterActualLRP

				fakeActualLRPDB.StartActualLRPReturns(&newActualLRP, &newAfterActualLRP, nil)
				fakeActualLRPDB.ActualLRPsReturns([]*models.ActualLRP{&newAfterActualLRP}, nil)
			})

			It("calls DB successfully", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeActualLRPDB.StartActualLRPCallCount()).To(Equal(1))
				Expect(fakeActualLRPDB.ActualLRPsCallCount()).To(Equal(1))
			})

			Context("when a non-ResourceNotFound error occurs while fetching the lrp", func() {
				BeforeEach(func() {
					fakeActualLRPDB.ActualLRPsReturns(nil, errors.New("BOOM!!!"))
				})

				It("should return the error", func() {
					Expect(err).To(MatchError("BOOM!!!"))
				})
			})

			Context("when a ResourceNotFound error occurs while fetching the lrp", func() {
				BeforeEach(func() {
					fakeActualLRPDB.ActualLRPsReturns(nil, models.ErrResourceNotFound)
				})

				It("should continue to start the LRP", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeActualLRPDB.StartActualLRPCallCount()).To(Equal(1))
				})
			})

			Context("when the lrp is evacuating", func() {
				var evacuatingLRP models.ActualLRP

				BeforeEach(func() {
					evacuatingLRP = actualLRP
					newAfterActualLRP := afterActualLRP

					evacuatingLRP.ActualLRPInstanceKey = models.NewActualLRPInstanceKey(
						"instance-guid-1",
						"cell-id-1",
					)
					evacuatingLRP.Presence = models.ActualLRP_Evacuating
					fakeActualLRPDB.ActualLRPsReturns([]*models.ActualLRP{&evacuatingLRP, &afterActualLRP}, nil)
					fakeActualLRPDB.StartActualLRPReturns(&evacuatingLRP, &newAfterActualLRP, nil)
				})

				It("removes the evacuating lrp", func() {
					Expect(fakeEvacuationDB.RemoveEvacuatingActualLRPCallCount()).To(Equal(1))
				})

				It("should emit an ActualLRPChanged event and an ActualLRPRemoved event", func() {
					Eventually(actualHub.EmitCallCount).Should(Equal(2))

					event1 := actualHub.EmitArgsForCall(0)
					changedEvent, ok := event1.(*models.ActualLRPChangedEvent)
					Expect(ok).To(BeTrue())
					Expect(changedEvent.Before).To(Equal(evacuatingLRP.ToActualLRPGroup()))
					Expect(changedEvent.After).To(Equal(afterActualLRP.ToActualLRPGroup()))

					event2 := actualHub.EmitArgsForCall(1)
					removedEvent, ok := event2.(*models.ActualLRPRemovedEvent)
					Expect(ok).To(BeTrue())
					Expect(removedEvent.ActualLrpGroup).To(Equal(evacuatingLRP.ToActualLRPGroup()))
				})
			})

			Context("when the actual lrp was created", func() {
				BeforeEach(func() {
					newAfterActualLRP := afterActualLRP
					fakeActualLRPDB.StartActualLRPReturns(nil, &newAfterActualLRP, nil)
				})

				It("emits a created event to the hub", func() {
					Eventually(actualHub.EmitCallCount).Should(Equal(1))
					event := actualHub.EmitArgsForCall(0)
					createdEvent, ok := event.(*models.ActualLRPCreatedEvent)
					Expect(ok).To(BeTrue())
					Expect(createdEvent.ActualLrpGroup).To(Equal(afterActualLRP.ToActualLRPGroup()))
				})
			})

			Context("when the actual lrp was updated", func() {
				It("emits a change event to the hub", func() {
					Eventually(actualHub.EmitCallCount).Should(Equal(1))
					event := actualHub.EmitArgsForCall(0)
					changedEvent, ok := event.(*models.ActualLRPChangedEvent)
					Expect(ok).To(BeTrue())
					Expect(changedEvent.Before).To(Equal(actualLRP.ToActualLRPGroup()))
					Expect(changedEvent.After).To(Equal(afterActualLRP.ToActualLRPGroup()))
				})
			})

			Context("when the actual lrp wasn't updated", func() {
				BeforeEach(func() {
					newActualLRP := actualLRP
					fakeActualLRPDB.StartActualLRPReturns(&newActualLRP, &newActualLRP, nil)
				})

				It("does not emit a change event to the hub", func() {
					Consistently(actualHub.EmitCallCount).Should(Equal(0))
				})
			})
		})

		Context("when starting the actual lrp fails", func() {
			BeforeEach(func() {
				fakeActualLRPDB.ActualLRPsReturns([]*models.ActualLRP{&actualLRP}, nil)
				fakeActualLRPDB.StartActualLRPReturns(nil, nil, models.ErrUnknownError)
			})

			It("responds with an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(models.ErrUnknownError))
			})

			It("does not emit a change event to the hub", func() {
				Consistently(actualHub.EmitCallCount).Should(Equal(0))
			})
		})
	})

	Describe("CrashActualLRP", func() {
		var (
			processGuid  = "process-guid"
			index        = int32(1)
			instanceGuid = "instance-guid"
			cellId       = "cell-id"

			key          models.ActualLRPKey
			instanceKey  models.ActualLRPInstanceKey
			errorMessage string
			lrps         []*models.ActualLRP
		)

		BeforeEach(func() {
			key = models.NewActualLRPKey(
				processGuid,
				index,
				"domain-0",
			)
			instanceKey = models.NewActualLRPInstanceKey(instanceGuid, cellId)
			errorMessage = "something went wrong"
			actualLRP = models.ActualLRP{
				ActualLRPKey: models.NewActualLRPKey(
					processGuid,
					1,
					"domain-0",
				),
				ActualLRPInstanceKey: instanceKey,
				State:                models.ActualLRPStateUnclaimed,
				Since:                1138,
			}
			afterActualLRP = models.ActualLRP{
				ActualLRPKey: models.NewActualLRPKey(
					processGuid,
					1,
					"domain-0",
				),
				State:       models.ActualLRPStateUnclaimed,
				Since:       1138,
				CrashCount:  1,
				CrashReason: errorMessage,
			}

			lrps = []*models.ActualLRP{&actualLRP}
		})

		BeforeEach(func() {
			fakeActualLRPDB.CrashActualLRPReturns(
				&actualLRP,
				&afterActualLRP,
				false,
				nil,
			)
		})

		JustBeforeEach(func() {
			fakeActualLRPDB.ActualLRPsReturns(lrps, nil)
			err = controller.CrashActualLRP(logger, &key, &instanceKey, errorMessage)
		})

		Context("when the LRP being crashed is a Suspect LRP", func() {
			BeforeEach(func() {
				fakeSuspectDB.RemoveSuspectActualLRPReturns(&actualLRP, nil)
				actualLRP.Presence = models.ActualLRP_Suspect
			})

			It("removes the Suspsect LRP", func() {
				Expect(fakeSuspectDB.RemoveSuspectActualLRPCallCount()).To(Equal(1))
			})

			It("emits ActualLRPRemovedEvent", func() {
				Eventually(actualHub.EmitCallCount).Should(Equal(1))
				event := actualHub.EmitArgsForCall(0)
				removedEvent, ok := event.(*models.ActualLRPRemovedEvent)
				Expect(ok).To(BeTrue())
				Expect(removedEvent.ActualLrpGroup).To(Equal(actualLRP.ToActualLRPGroup()))
			})

			Context("when RemoveSuspectActualLRP returns an error", func() {
				BeforeEach(func() {
					fakeSuspectDB.RemoveSuspectActualLRPReturns(nil, errors.New("boooom!"))
				})

				It("returns the error to the caller", func() {
					Expect(err).To(MatchError("boooom!"))
				})

				It("does not emit any events", func() {
					Consistently(actualHub.EmitCallCount).Should(BeZero())
				})
			})
		})

		Context("when crashing the actual lrp in the DB succeeds", func() {
			var desiredLRP *models.DesiredLRP

			itEmitsCrashAndChangedEvents := func() {
				var eventChan chan models.Event

				BeforeEach(func() {
					ch := make(chan models.Event)
					eventChan = ch

					actualHub.EmitStub = func(event models.Event) {
						ch <- event
					}
				})

				It("emits a crash and change event to the hub", func() {
					Eventually(eventChan).Should(Receive(Equal(&models.ActualLRPCrashedEvent{
						ActualLRPKey:         actualLRP.ActualLRPKey,
						ActualLRPInstanceKey: actualLRP.ActualLRPInstanceKey,
						Since:                1138,
						CrashCount:           1,
						CrashReason:          errorMessage,
					})))
				})

				It("emits a change event to the hub", func() {
					Eventually(eventChan).Should(Receive(Equal(&models.ActualLRPChangedEvent{
						Before: actualLRP.ToActualLRPGroup(),
						After:  afterActualLRP.ToActualLRPGroup(),
					})))
				})
			}

			BeforeEach(func() {
				desiredLRP = &models.DesiredLRP{
					ProcessGuid: "process-guid",
					Domain:      "some-domain",
					RootFs:      "some-stack",
					MemoryMb:    128,
					DiskMb:      512,
					MaxPids:     100,
				}

				fakeDesiredLRPDB.DesiredLRPByProcessGuidReturns(desiredLRP, nil)
				fakeActualLRPDB.CrashActualLRPReturns(&actualLRP, &afterActualLRP, true, nil)
			})

			It("response with no error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("crashes the actual lrp by process guid and index", func() {
				Expect(fakeActualLRPDB.CrashActualLRPCallCount()).To(Equal(1))
				_, actualKey, actualInstanceKey, actualErrorMessage := fakeActualLRPDB.CrashActualLRPArgsForCall(0)
				Expect(*actualKey).To(Equal(key))
				Expect(*actualInstanceKey).To(Equal(instanceKey))
				Expect(actualErrorMessage).To(Equal(errorMessage))
			})

			itEmitsCrashAndChangedEvents()

			Describe("restarting the instance", func() {
				Context("when the actual LRP should be restarted", func() {
					It("request an auction", func() {
						Expect(err).NotTo(HaveOccurred())

						Expect(fakeDesiredLRPDB.DesiredLRPByProcessGuidCallCount()).To(Equal(1))
						_, processGuid := fakeDesiredLRPDB.DesiredLRPByProcessGuidArgsForCall(0)
						Expect(processGuid).To(Equal("process-guid"))

						Expect(fakeAuctioneerClient.RequestLRPAuctionsCallCount()).To(Equal(1))
						_, startRequests := fakeAuctioneerClient.RequestLRPAuctionsArgsForCall(0)
						Expect(startRequests).To(HaveLen(1))
						schedulingInfo := desiredLRP.DesiredLRPSchedulingInfo()
						expectedStartRequest := auctioneer.NewLRPStartRequestFromSchedulingInfo(&schedulingInfo, 1)
						Expect(startRequests[0]).To(BeEquivalentTo(&expectedStartRequest))
					})
				})

				Context("when the actual lrp should not be restarted (e.g., crashed)", func() {
					BeforeEach(func() {
						fakeActualLRPDB.CrashActualLRPReturns(&actualLRP, &actualLRP, false, nil)
					})

					It("does not request an auction", func() {
						Expect(err).NotTo(HaveOccurred())
						Expect(fakeAuctioneerClient.RequestLRPAuctionsCallCount()).To(Equal(0))
					})
				})

				Context("when fetching the desired lrp fails", func() {
					BeforeEach(func() {
						fakeDesiredLRPDB.DesiredLRPByProcessGuidReturns(nil, errors.New("error occured"))
					})

					It("fails and does not request an auction", func() {
						Expect(err).To(MatchError("error occured"))
						Expect(fakeAuctioneerClient.RequestLRPAuctionsCallCount()).To(Equal(0))
					})
				})

				Context("when requesting the auction fails", func() {
					BeforeEach(func() {
						fakeAuctioneerClient.RequestLRPAuctionsReturns(errors.New("some else bid higher"))
					})

					It("should not return an error", func() {
						Expect(err).NotTo(HaveOccurred())
					})

					itEmitsCrashAndChangedEvents()
				})
			})
		})

		Context("when crashing the actual lrp fails", func() {
			BeforeEach(func() {
				fakeActualLRPDB.CrashActualLRPReturns(nil, nil, false, models.ErrUnknownError)
			})

			It("responds with an error", func() {
				Expect(err).To(MatchError(models.ErrUnknownError))
			})

			It("does not emit a change event to the hub", func() {
				Consistently(actualHub.EmitCallCount).Should(Equal(0))
			})
		})

		Context("when the instance key matches the suspect LRP instance key", func() {
			BeforeEach(func() {
				actualLRP.Presence = models.ActualLRP_Suspect
				fakeSuspectDB.RemoveSuspectActualLRPReturns(&actualLRP, nil)
				fakeActualLRPDB.ActualLRPsReturns([]*models.ActualLRP{&actualLRP}, nil)
			})

			It("removes the suspect LRP", func() {
				Expect(fakeSuspectDB.RemoveSuspectActualLRPCallCount()).To(Equal(1))

				_, lrpKey := fakeSuspectDB.RemoveSuspectActualLRPArgsForCall(0)
				Expect(lrpKey.ProcessGuid).To(Equal(processGuid))
				Expect(lrpKey.Index).To(BeEquivalentTo(index))
			})

			It("emits an actual LRP removed event", func() {
				Eventually(actualHub.EmitCallCount).Should(Equal(1))
				event := actualHub.EmitArgsForCall(0)
				removedEvent, ok := event.(*models.ActualLRPRemovedEvent)
				Expect(ok).To(BeTrue())
				Expect(removedEvent.ActualLrpGroup).To(Equal(actualLRP.ToActualLRPGroup()))
			})
		})
	})

	Describe("FailActualLRP", func() {
		var (
			processGuid = "process-guid"
			index       = int32(1)

			key          models.ActualLRPKey
			errorMessage string
		)

		BeforeEach(func() {
			key = models.NewActualLRPKey(
				processGuid,
				index,
				"domain-0",
			)
			errorMessage = "something went wrong"

			actualLRP = models.ActualLRP{
				ActualLRPKey: models.NewActualLRPKey(
					processGuid,
					1,
					"domain-0",
				),
				State: models.ActualLRPStateUnclaimed,
				Since: 1138,
			}

			afterActualLRP = models.ActualLRP{
				ActualLRPKey: models.NewActualLRPKey(
					processGuid,
					1,
					"domain-0",
				),
				State: models.ActualLRPStateUnclaimed,
				Since: 1138,
			}

			fakeActualLRPDB.ActualLRPsReturns([]*models.ActualLRP{}, nil)
		})

		JustBeforeEach(func() {
			err = controller.FailActualLRP(logger, &key, errorMessage)
		})

		Context("when failing the actual lrp in the DB succeeds", func() {
			BeforeEach(func() {
				fakeActualLRPDB.FailActualLRPReturns(&actualLRP, &afterActualLRP, nil)
			})

			It("fails the actual lrp by process guid and index", func() {
				Expect(fakeActualLRPDB.FailActualLRPCallCount()).To(Equal(1))
				_, actualKey, actualErrorMessage := fakeActualLRPDB.FailActualLRPArgsForCall(0)
				Expect(*actualKey).To(Equal(key))
				Expect(actualErrorMessage).To(Equal(errorMessage))

				Expect(err).NotTo(HaveOccurred())
			})

			It("emits a change event to the hub", func() {
				Eventually(actualHub.EmitCallCount).Should(Equal(1))
				event := actualHub.EmitArgsForCall(0)
				changedEvent, ok := event.(*models.ActualLRPChangedEvent)
				Expect(ok).To(BeTrue())
				Expect(changedEvent.Before).To(Equal(&models.ActualLRPGroup{Instance: &actualLRP}))
				Expect(changedEvent.After).To(Equal(&models.ActualLRPGroup{Instance: &afterActualLRP}))
			})
		})

		Context("when there is a Suspect LRP running", func() {
			BeforeEach(func() {
				fakeActualLRPDB.ActualLRPsReturns([]*models.ActualLRP{
					&models.ActualLRP{
						Presence: models.ActualLRP_Suspect,
					},
				}, nil)
			})

			It("does not emit a ActualLRPChangedEventChanged", func() {
				Consistently(actualHub.EmitCallCount).Should(BeZero())
			})

			Context("when there is no non-suspect instance", func() {
				BeforeEach(func() {
					fakeActualLRPDB.FailActualLRPReturns(nil, nil, models.ErrResourceNotFound)
				})

				It("does not error", func() {
					Expect(err).To(BeNil())
				})

				It("does not emit a ActualLRPChangedEventChanged", func() {
					Consistently(actualHub.EmitCallCount).Should(BeZero())
				})
			})
		})

		Context("when failing the actual lrp fails with a non-ResourceNotFound error", func() {
			BeforeEach(func() {
				fakeActualLRPDB.FailActualLRPReturns(nil, nil, models.ErrUnknownError)
			})

			It("responds with an error", func() {
				Expect(err).To(MatchError(models.ErrUnknownError))
			})

			It("does not emit a change event to the hub", func() {
				Consistently(actualHub.EmitCallCount).Should(Equal(0))
			})
		})

		Context("when there is no LRP", func() {
			BeforeEach(func() {
				fakeActualLRPDB.FailActualLRPReturns(nil, nil, models.ErrResourceNotFound)
				fakeActualLRPDB.ActualLRPsReturns(nil, models.ErrResourceNotFound)
			})

			It("responds with a ResourceNotFound error", func() {
				Expect(err).To(MatchError(models.ErrResourceNotFound))
			})

			It("does not emit a change event to the hub", func() {
				Consistently(actualHub.EmitCallCount).Should(Equal(0))
			})
		})
	})

	Describe("RemoveActualLRP", func() {
		var (
			processGuid       = "process-guid"
			index       int32 = 1

			instanceKey models.ActualLRPInstanceKey
		)

		BeforeEach(func() {
			instanceKey = models.NewActualLRPInstanceKey(
				"instance-guid-0",
				"cell-id-0",
			)
			actualLRP = models.ActualLRP{
				ActualLRPKey: models.NewActualLRPKey(
					processGuid,
					1,
					"domain-0",
				),
				State: models.ActualLRPStateUnclaimed,
				Since: 1138,
			}

			fakeActualLRPDB.ActualLRPsReturns([]*models.ActualLRP{&actualLRP}, nil)
		})

		JustBeforeEach(func() {
			err = controller.RemoveActualLRP(logger, processGuid, index, &instanceKey)
		})

		Context("when removing the actual lrp in the DB succeeds", func() {
			var removedActualLRP models.ActualLRP

			BeforeEach(func() {
				removedActualLRP = actualLRP
				removedActualLRP.ActualLRPInstanceKey = instanceKey
				fakeActualLRPDB.RemoveActualLRPReturns(nil)
			})

			It("removes the actual lrp by process guid and index", func() {
				Expect(fakeActualLRPDB.RemoveActualLRPCallCount()).To(Equal(1))

				_, actualProcessGuid, idx, actualInstanceKey := fakeActualLRPDB.RemoveActualLRPArgsForCall(0)
				Expect(actualProcessGuid).To(Equal(processGuid))
				Expect(idx).To(BeEquivalentTo(index))
				Expect(actualInstanceKey).To(Equal(&instanceKey))
			})

			It("response with no error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("emits a removed event to the hub", func() {
				Eventually(actualHub.EmitCallCount).Should(Equal(1))
				event := actualHub.EmitArgsForCall(0)
				removedEvent, ok := event.(*models.ActualLRPRemovedEvent)
				Expect(ok).To(BeTrue())
				Expect(removedEvent.ActualLrpGroup).To(Equal(actualLRP.ToActualLRPGroup()))
			})
		})

		Context("when the DB returns an error", func() {
			Context("when doing the actual LRP lookup", func() {
				BeforeEach(func() {
					fakeActualLRPDB.ActualLRPsReturns(nil, models.ErrUnknownError)
				})

				It("returns the error", func() {
					Expect(err).To(MatchError(models.ErrUnknownError))
				})

				It("does not emit a change event to the hub", func() {
					Consistently(actualHub.EmitCallCount).Should(Equal(0))
				})
			})

			Context("when doing the actual LRP removal", func() {
				BeforeEach(func() {
					fakeActualLRPDB.RemoveActualLRPReturns(models.ErrUnknownError)
				})

				It("returns the error", func() {
					Expect(err).To(MatchError(models.ErrUnknownError))
				})

				It("does not emit a change event to the hub", func() {
					Consistently(actualHub.EmitCallCount).Should(Equal(0))
				})
			})
		})
	})

	Describe("RetireActualLRP", func() {
		var (
			processGuid = "process-guid"
			index       = int32(1)

			key              models.ActualLRPKey
			retiredActualLRP models.ActualLRP
		)

		BeforeEach(func() {
			key = models.NewActualLRPKey(
				processGuid,
				index,
				"domain-0",
			)

			retiredActualLRP = models.ActualLRP{
				ActualLRPKey: key,
				State:        models.ActualLRPStateUnclaimed,
				Since:        1138,
			}
		})

		JustBeforeEach(func() {
			err = controller.RetireActualLRP(logger, &key)
		})

		Context("when finding the actualLRP fails", func() {
			BeforeEach(func() {
				fakeActualLRPDB.ActualLRPsReturns(nil, models.ErrUnknownError)
			})

			It("returns an error and does not retry", func() {
				Expect(err).To(MatchError(models.ErrUnknownError))
				Expect(fakeActualLRPDB.ActualLRPsCallCount()).To(Equal(1))
			})

			It("does not emit a change event to the hub", func() {
				Consistently(actualHub.EmitCallCount).Should(Equal(0))
			})
		})

		Context("when there is no matching actual lrp", func() {
			BeforeEach(func() {
				fakeActualLRPDB.ActualLRPsReturns([]*models.ActualLRP{}, nil)
			})

			It("returns an error and does not retry", func() {
				Expect(err).To(Equal(models.ErrResourceNotFound))
				Expect(fakeActualLRPDB.ActualLRPsCallCount()).To(Equal(1))
			})

			It("does not emit a change event to the hub", func() {
				Consistently(actualHub.EmitCallCount).Should(Equal(0))
			})
		})

		Context("with an Unclaimed LRP", func() {
			BeforeEach(func() {
				fakeActualLRPDB.ActualLRPsReturns([]*models.ActualLRP{&retiredActualLRP}, nil)
			})

			It("removes the LRP", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeActualLRPDB.RemoveActualLRPCallCount()).To(Equal(1))

				_, deletedLRPGuid, deletedLRPIndex, deletedLRPInstanceKey := fakeActualLRPDB.RemoveActualLRPArgsForCall(0)
				Expect(deletedLRPGuid).To(Equal(processGuid))
				Expect(deletedLRPIndex).To(Equal(index))
				Expect(deletedLRPInstanceKey).To(Equal(&retiredActualLRP.ActualLRPInstanceKey))
			})

			It("emits a removed event to the hub", func() {
				Eventually(actualHub.EmitCallCount).Should(Equal(1))
				event := actualHub.EmitArgsForCall(0)
				removedEvent, ok := event.(*models.ActualLRPRemovedEvent)
				Expect(ok).To(BeTrue())
				Expect(removedEvent.ActualLrpGroup).To(Equal(retiredActualLRP.ToActualLRPGroup()))
			})

			Context("when removing the actual lrp fails", func() {
				BeforeEach(func() {
					fakeActualLRPDB.RemoveActualLRPReturns(errors.New("boom!"))
				})

				It("retries removing up to RetireActualLRPRetryAttempts times", func() {
					Expect(err).To(MatchError("boom!"))
					Expect(fakeActualLRPDB.RemoveActualLRPCallCount()).To(Equal(5))
					Expect(fakeActualLRPDB.ActualLRPsCallCount()).To(Equal(5))
				})

				It("does not emit a change event to the hub", func() {
					Consistently(actualHub.EmitCallCount).Should(Equal(0))
				})
			})
		})

		Context("when the LRP is crashed", func() {
			BeforeEach(func() {
				retiredActualLRP.State = models.ActualLRPStateCrashed
				fakeActualLRPDB.ActualLRPsReturns([]*models.ActualLRP{&retiredActualLRP}, nil)
			})

			It("removes the LRP", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeActualLRPDB.RemoveActualLRPCallCount()).To(Equal(1))

				_, deletedLRPGuid, deletedLRPIndex, deletedLRPInstanceKey := fakeActualLRPDB.RemoveActualLRPArgsForCall(0)
				Expect(deletedLRPGuid).To(Equal(processGuid))
				Expect(deletedLRPIndex).To(Equal(index))
				Expect(deletedLRPInstanceKey).To(Equal(&retiredActualLRP.ActualLRPInstanceKey))
			})

			It("emits a removed event to the hub", func() {
				Eventually(actualHub.EmitCallCount).Should(Equal(1))
				event := actualHub.EmitArgsForCall(0)
				removedEvent, ok := event.(*models.ActualLRPRemovedEvent)
				Expect(ok).To(BeTrue())
				Expect(removedEvent.ActualLrpGroup).To(Equal(retiredActualLRP.ToActualLRPGroup()))
			})

			Context("when removing the actual lrp fails", func() {
				BeforeEach(func() {
					fakeActualLRPDB.RemoveActualLRPReturns(errors.New("boom!"))
				})

				It("retries removing up to RetireActualLRPRetryAttempts times", func() {
					Expect(err).To(MatchError("boom!"))
					Expect(fakeActualLRPDB.RemoveActualLRPCallCount()).To(Equal(5))
					Expect(fakeActualLRPDB.ActualLRPsCallCount()).To(Equal(5))
				})

				It("does not emit a change event to the hub", func() {
					Consistently(actualHub.EmitCallCount).Should(Equal(0))
				})
			})
		})

		Context("when the LRP is Claimed or Running", func() {
			var (
				cellID       string
				cellPresence models.CellPresence
				instanceKey  models.ActualLRPInstanceKey
			)

			BeforeEach(func() {
				cellID = "cell-id"
				instanceKey = models.NewActualLRPInstanceKey("instance-guid", cellID)

				retiredActualLRP.CellId = cellID
				retiredActualLRP.ActualLRPInstanceKey = instanceKey
				retiredActualLRP.State = models.ActualLRPStateClaimed
				fakeActualLRPDB.ActualLRPsReturns([]*models.ActualLRP{&retiredActualLRP}, nil)
			})

			Context("when the cell", func() {
				Context("is present", func() {
					BeforeEach(func() {
						cellPresence = models.NewCellPresence(
							cellID,
							"cell1.addr",
							"",
							"the-zone",
							models.NewCellCapacity(128, 1024, 6),
							[]string{},
							[]string{},
							[]string{},
							[]string{},
						)

						fakeServiceClient.CellByIdReturns(&cellPresence, nil)
					})

					It("stops the LRPs", func() {
						Expect(fakeRepClientFactory.CreateClientCallCount()).To(Equal(1))
						Expect(fakeRepClientFactory.CreateClientArgsForCall(0)).To(Equal(cellPresence.RepAddress))

						Expect(fakeServiceClient.CellByIdCallCount()).To(Equal(1))
						_, fetchedCellID := fakeServiceClient.CellByIdArgsForCall(0)
						Expect(fetchedCellID).To(Equal(cellID))

						Expect(fakeRepClient.StopLRPInstanceCallCount()).Should(Equal(1))
						_, stoppedKey, stoppedInstanceKey := fakeRepClient.StopLRPInstanceArgsForCall(0)
						Expect(stoppedKey).To(Equal(key))
						Expect(stoppedInstanceKey).To(Equal(instanceKey))
					})

					Context("when the rep announces a rep url", func() {
						BeforeEach(func() {
							cellPresence = models.NewCellPresence(
								cellID,
								"cell1.addr",
								"http://cell1.addr",
								"the-zone",
								models.NewCellCapacity(128, 1024, 6),
								[]string{},
								[]string{},
								[]string{},
								[]string{},
							)

							fakeServiceClient.CellByIdReturns(&cellPresence, nil)
						})

						It("passes the url when creating a rep client", func() {
							Expect(fakeRepClientFactory.CreateClientCallCount()).To(Equal(1))
							repAddr, repURL := fakeRepClientFactory.CreateClientArgsForCall(0)
							Expect(repAddr).To(Equal(cellPresence.RepAddress))
							Expect(repURL).To(Equal(cellPresence.RepUrl))
						})
					})

					Context("when creating a rep client fails", func() {
						BeforeEach(func() {
							err := errors.New("BOOM!!!")
							fakeRepClientFactory.CreateClientReturns(nil, err)
						})

						It("should return the error", func() {
							Expect(err).To(MatchError("BOOM!!!"))
						})
					})

					Context("Stopping the LRP fails", func() {
						BeforeEach(func() {
							fakeRepClient.StopLRPInstanceReturns(errors.New("Failed to stop app"))
						})

						It("retries to stop the app", func() {
							Expect(err).To(MatchError("Failed to stop app"))
							Expect(fakeRepClient.StopLRPInstanceCallCount()).Should(Equal(5))
						})
					})
				})

				Context("is not present", func() {
					BeforeEach(func() {
						fakeServiceClient.CellByIdReturns(nil,
							&models.Error{
								Type:    models.Error_ResourceNotFound,
								Message: "cell not found",
							})
					})

					Context("removing the actualLRP succeeds", func() {
						It("removes the LRPs", func() {
							Expect(err).NotTo(HaveOccurred())
							Expect(fakeActualLRPDB.RemoveActualLRPCallCount()).To(Equal(1))

							_, deletedLRPGuid, deletedLRPIndex, deletedLRPInstanceKey := fakeActualLRPDB.RemoveActualLRPArgsForCall(0)
							Expect(deletedLRPGuid).To(Equal(processGuid))
							Expect(deletedLRPIndex).To(Equal(index))
							Expect(deletedLRPInstanceKey).To(Equal(&instanceKey))
						})

						It("emits a removed event to the hub", func() {
							Eventually(actualHub.EmitCallCount).Should(Equal(1))
							event := actualHub.EmitArgsForCall(0)
							removedEvent, ok := event.(*models.ActualLRPRemovedEvent)
							Expect(ok).To(BeTrue())
							Expect(removedEvent.ActualLrpGroup).To(Equal(retiredActualLRP.ToActualLRPGroup()))
						})
					})

					Context("removing the actualLRP fails", func() {
						BeforeEach(func() {
							fakeActualLRPDB.RemoveActualLRPReturns(errors.New("failed to delete actual LRP"))
						})

						It("returns an error and does not retry", func() {
							Expect(err).To(MatchError("failed to delete actual LRP"))
							Expect(fakeActualLRPDB.RemoveActualLRPCallCount()).To(Equal(1))
						})

						It("does not emit a change event to the hub", func() {
							Consistently(actualHub.EmitCallCount).Should(Equal(0))
						})
					})
				})

				Context("is present, but returns an error on lookup", func() {
					BeforeEach(func() {
						fakeServiceClient.CellByIdReturns(nil, errors.New("cell error"))
					})

					It("returns an error and retries", func() {
						Expect(err).To(MatchError("cell error"))
						Expect(fakeActualLRPDB.RemoveActualLRPCallCount()).To(Equal(0))
						Expect(fakeServiceClient.CellByIdCallCount()).To(Equal(1))
					})

					It("does not emit a change event to the hub", func() {
						Consistently(actualHub.EmitCallCount).Should(Equal(0))
					})
				})
			})
		})
	})
})
