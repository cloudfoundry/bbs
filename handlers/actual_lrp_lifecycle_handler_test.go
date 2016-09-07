package handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"

	"code.cloudfoundry.org/auctioneer"
	"code.cloudfoundry.org/auctioneer/auctioneerfakes"
	"code.cloudfoundry.org/bbs/controllers"
	"code.cloudfoundry.org/bbs/db/dbfakes"
	"code.cloudfoundry.org/bbs/events/eventfakes"
	"code.cloudfoundry.org/bbs/fake_bbs"
	"code.cloudfoundry.org/bbs/handlers"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/rep/repfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("ActualLRP Lifecycle Handlers", func() {
	var (
		logger               *lagertest.TestLogger
		fakeActualLRPDB      *dbfakes.FakeActualLRPDB
		fakeDesiredLRPDB     *dbfakes.FakeDesiredLRPDB
		fakeAuctioneerClient *auctioneerfakes.FakeClient
		actualHub            *eventfakes.FakeHub
		responseRecorder     *httptest.ResponseRecorder
		handler              *handlers.ActualLRPLifecycleHandler
		exitCh               chan struct{}

		actualLRP      models.ActualLRP
		afterActualLRP models.ActualLRP

		fakeServiceClient    *fake_bbs.FakeServiceClient
		fakeRepClientFactory *repfakes.FakeClientFactory
		fakeRepClient        *repfakes.FakeClient
	)

	BeforeEach(func() {
		fakeActualLRPDB = new(dbfakes.FakeActualLRPDB)
		fakeAuctioneerClient = new(auctioneerfakes.FakeClient)
		fakeDesiredLRPDB = new(dbfakes.FakeDesiredLRPDB)
		logger = lagertest.NewTestLogger("test")
		responseRecorder = httptest.NewRecorder()

		fakeServiceClient = new(fake_bbs.FakeServiceClient)
		fakeRepClientFactory = new(repfakes.FakeClientFactory)
		fakeRepClient = new(repfakes.FakeClient)
		fakeRepClientFactory.CreateClientReturns(fakeRepClient)

		actualHub = &eventfakes.FakeHub{}
		exitCh = make(chan struct{}, 1)
		retirer := controllers.NewActualLRPRetirer(fakeActualLRPDB, actualHub, fakeRepClientFactory, fakeServiceClient)
		handler = handlers.NewActualLRPLifecycleHandler(fakeActualLRPDB, fakeDesiredLRPDB, actualHub, fakeAuctioneerClient, retirer, exitCh)
	})

	Describe("ClaimActualLRP", func() {
		var (
			processGuid       = "process-guid"
			index       int32 = 1
			instanceKey models.ActualLRPInstanceKey
			requestBody interface{}
		)

		BeforeEach(func() {
			instanceKey = models.NewActualLRPInstanceKey(
				"instance-guid-0",
				"cell-id-0",
			)
			requestBody = &instanceKey
			requestBody = &models.ClaimActualLRPRequest{
				ProcessGuid:          processGuid,
				Index:                index,
				ActualLrpInstanceKey: &instanceKey,
			}
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
				State: models.ActualLRPStateClaimed,
				Since: 1140,
			}
		})

		JustBeforeEach(func() {
			request := newTestRequest(requestBody)
			handler.ClaimActualLRP(logger, responseRecorder, request)
		})

		Context("when claiming the actual lrp in the DB succeeds", func() {
			BeforeEach(func() {
				fakeActualLRPDB.ClaimActualLRPReturns(&models.ActualLRPGroup{Instance: &actualLRP}, &models.ActualLRPGroup{Instance: &afterActualLRP}, nil)
			})

			It("response with no error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := &models.ActualLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(BeNil())
			})

			It("claims the actual lrp by process guid and index", func() {
				Expect(fakeActualLRPDB.ClaimActualLRPCallCount()).To(Equal(1))
				_, actualProcessGuid, actualIndex, actualInstanceKey := fakeActualLRPDB.ClaimActualLRPArgsForCall(0)
				Expect(actualProcessGuid).To(Equal(processGuid))
				Expect(actualIndex).To(BeEquivalentTo(index))
				Expect(*actualInstanceKey).To(Equal(instanceKey))
			})

			It("emits a change event to the hub", func() {
				Eventually(actualHub.EmitCallCount).Should(Equal(1))
				event := actualHub.EmitArgsForCall(0)
				changedEvent, ok := event.(*models.ActualLRPChangedEvent)
				Expect(ok).To(BeTrue())
				Expect(changedEvent.Before).To(Equal(&models.ActualLRPGroup{Instance: &actualLRP}))
				Expect(changedEvent.After).To(Equal(&models.ActualLRPGroup{Instance: &afterActualLRP}))
			})

			Context("when the actual lrp did not actually change", func() {
				BeforeEach(func() {
					fakeActualLRPDB.ClaimActualLRPReturns(
						&models.ActualLRPGroup{Instance: &afterActualLRP},
						&models.ActualLRPGroup{Instance: &afterActualLRP},
						nil,
					)
				})

				It("does not emit a change event to the hub", func() {
					Eventually(actualHub.EmitCallCount).Should(Equal(0))
				})
			})
		})

		Context("when the DB returns an unrecoverable error", func() {
			BeforeEach(func() {
				fakeActualLRPDB.ClaimActualLRPReturns(nil, nil, models.NewUnrecoverableError(nil))
			})

			It("logs and writes to the exit channel", func() {
				Eventually(logger).Should(gbytes.Say("unrecoverable-error"))
				Eventually(exitCh).Should(Receive())
			})
		})

		Context("when claiming the actual lrp fails", func() {
			BeforeEach(func() {
				fakeActualLRPDB.ClaimActualLRPReturns(nil, nil, models.ErrUnknownError)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := &models.ActualLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrUnknownError))
			})

			It("does not emits a change event to the hub", func() {
				Consistently(actualHub.EmitCallCount).Should(Equal(0))
			})
		})

		Context("when we cannot find the resource", func() {
			BeforeEach(func() {
				fakeActualLRPDB.ClaimActualLRPReturns(nil, nil, models.ErrResourceNotFound)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := &models.ActualLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrResourceNotFound))
			})

			It("does not emits a change event to the hub", func() {
				Consistently(actualHub.EmitCallCount).Should(Equal(0))
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

			requestBody interface{}
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
			netInfo = models.NewActualLRPNetInfo("1.1.1.1", models.NewPortMapping(10, 20))
			requestBody = &models.StartActualLRPRequest{
				ActualLrpKey:         &key,
				ActualLrpInstanceKey: &instanceKey,
				ActualLrpNetInfo:     &netInfo,
			}

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
			request := newTestRequest(requestBody)
			handler.StartActualLRP(logger, responseRecorder, request)
		})

		Context("when starting the actual lrp in the DB succeeds", func() {
			BeforeEach(func() {
				fakeActualLRPDB.StartActualLRPReturns(&models.ActualLRPGroup{Instance: &actualLRP}, &models.ActualLRPGroup{Instance: &afterActualLRP}, nil)
			})

			It("responds with no error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := &models.ActualLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(BeNil())
			})

			It("starts the actual lrp by process guid and index", func() {
				Expect(fakeActualLRPDB.StartActualLRPCallCount()).To(Equal(1))
				_, actualKey, actualInstanceKey, actualNetInfo := fakeActualLRPDB.StartActualLRPArgsForCall(0)
				Expect(*actualKey).To(Equal(key))
				Expect(*actualInstanceKey).To(Equal(instanceKey))
				Expect(*actualNetInfo).To(Equal(netInfo))
			})

			Context("when the actual lrp was created", func() {
				BeforeEach(func() {
					fakeActualLRPDB.StartActualLRPReturns(nil, &models.ActualLRPGroup{Instance: &afterActualLRP}, nil)
				})

				It("emits a created event to the hub", func() {
					Eventually(actualHub.EmitCallCount).Should(Equal(1))
					event := actualHub.EmitArgsForCall(0)
					createdEvent, ok := event.(*models.ActualLRPCreatedEvent)
					Expect(ok).To(BeTrue())
					Expect(createdEvent.ActualLrpGroup).To(Equal(&models.ActualLRPGroup{Instance: &afterActualLRP}))
				})
			})

			Context("when the actual lrp was updated", func() {
				It("emits a change event to the hub", func() {
					Eventually(actualHub.EmitCallCount).Should(Equal(1))
					event := actualHub.EmitArgsForCall(0)
					changedEvent, ok := event.(*models.ActualLRPChangedEvent)
					Expect(ok).To(BeTrue())
					Expect(changedEvent.Before).To(Equal(&models.ActualLRPGroup{Instance: &actualLRP}))
					Expect(changedEvent.After).To(Equal(&models.ActualLRPGroup{Instance: &afterActualLRP}))
				})
			})

			Context("when the actual lrp wasn't updated", func() {
				BeforeEach(func() {
					fakeActualLRPDB.StartActualLRPReturns(&models.ActualLRPGroup{Instance: &actualLRP}, &models.ActualLRPGroup{Instance: &actualLRP}, nil)
				})

				It("does not emit a change event to the hub", func() {
					Consistently(actualHub.EmitCallCount).Should(Equal(0))
				})
			})
		})

		Context("when the DB returns an unrecoverable error when doing the actual LRP lookup", func() {
			BeforeEach(func() {
				fakeActualLRPDB.StartActualLRPReturns(nil, nil, models.NewUnrecoverableError(nil))
			})

			It("logs and writes to the exit channel", func() {
				Eventually(logger).Should(gbytes.Say("unrecoverable-error"))
				Eventually(exitCh).Should(Receive())
			})
		})

		Context("when starting the actual lrp fails", func() {
			BeforeEach(func() {
				fakeActualLRPDB.StartActualLRPReturns(nil, nil, models.ErrUnknownError)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := &models.ActualLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrUnknownError))
			})

			It("does not emit a change event to the hub", func() {
				Consistently(actualHub.EmitCallCount).Should(Equal(0))
			})
		})

		Context("when we cannot find the resource", func() {
			BeforeEach(func() {
				fakeActualLRPDB.StartActualLRPReturns(nil, nil, models.ErrResourceNotFound)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := &models.ActualLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrResourceNotFound))
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

			requestBody interface{}
		)

		BeforeEach(func() {
			key = models.NewActualLRPKey(
				processGuid,
				index,
				"domain-0",
			)
			instanceKey = models.NewActualLRPInstanceKey(instanceGuid, cellId)
			errorMessage = "something went wrong"
			requestBody = &models.CrashActualLRPRequest{
				ActualLrpKey:         &key,
				ActualLrpInstanceKey: &instanceKey,
				ErrorMessage:         errorMessage,
			}
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
				State:       models.ActualLRPStateUnclaimed,
				Since:       1138,
				CrashCount:  1,
				CrashReason: errorMessage,
			}
		})

		JustBeforeEach(func() {
			request := newTestRequest(requestBody)
			handler.CrashActualLRP(logger, responseRecorder, request)
		})

		Context("when crashing the actual lrp in the DB succeeds", func() {
			var desiredLRP *models.DesiredLRP

			BeforeEach(func() {
				desiredLRP = &models.DesiredLRP{
					ProcessGuid: "process-guid",
					Domain:      "some-domain",
					RootFs:      "some-stack",
					MemoryMb:    128,
					DiskMb:      512,
				}

				fakeDesiredLRPDB.DesiredLRPByProcessGuidReturns(desiredLRP, nil)
				fakeActualLRPDB.CrashActualLRPReturns(&models.ActualLRPGroup{Instance: &actualLRP}, &models.ActualLRPGroup{Instance: &afterActualLRP}, true, nil)
			})

			It("response with no error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := &models.ActualLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(BeNil())
			})

			It("crashes the actual lrp by process guid and index", func() {
				Expect(fakeActualLRPDB.CrashActualLRPCallCount()).To(Equal(1))
				_, actualKey, actualInstanceKey, actualErrorMessage := fakeActualLRPDB.CrashActualLRPArgsForCall(0)
				Expect(*actualKey).To(Equal(key))
				Expect(*actualInstanceKey).To(Equal(instanceKey))
				Expect(actualErrorMessage).To(Equal(errorMessage))
			})

			It("emits a crash and change event to the hub", func() {
				Eventually(actualHub.EmitCallCount).Should(Equal(2))
				event1 := actualHub.EmitArgsForCall(0)
				event2 := actualHub.EmitArgsForCall(1)
				crashEvent, ok := event1.(*models.ActualLRPCrashedEvent)
				if !ok {
					crashEvent, ok = event2.(*models.ActualLRPCrashedEvent)
				}

				Expect(ok).To(BeTrue())
				Expect(crashEvent.ActualLRPKey).To(Equal(actualLRP.ActualLRPKey))
				Expect(crashEvent.ActualLRPInstanceKey).To(Equal(actualLRP.ActualLRPInstanceKey))
				Expect(crashEvent.Since).To(Equal(int64(1138)))
				Expect(crashEvent.CrashCount).To(Equal(int32(1)))
				Expect(crashEvent.CrashReason).To(Equal(errorMessage))

				changedEvent, ok := event1.(*models.ActualLRPChangedEvent)
				if !ok {
					changedEvent, ok = event2.(*models.ActualLRPChangedEvent)
				}

				Expect(changedEvent.Before).To(Equal(&models.ActualLRPGroup{Instance: &actualLRP}))
				Expect(changedEvent.After).To(Equal(&models.ActualLRPGroup{Instance: &afterActualLRP}))
			})

			Describe("restarting the instance", func() {
				Context("when the actual LRP should be restarted", func() {
					It("request an auction", func() {
						Expect(responseRecorder.Code).To(Equal(http.StatusOK))
						response := &models.ActualLRPLifecycleResponse{}
						err := response.Unmarshal(responseRecorder.Body.Bytes())
						Expect(err).NotTo(HaveOccurred())

						Expect(response.Error).To(BeNil())

						Expect(fakeDesiredLRPDB.DesiredLRPByProcessGuidCallCount()).To(Equal(1))
						_, processGuid := fakeDesiredLRPDB.DesiredLRPByProcessGuidArgsForCall(0)
						Expect(processGuid).To(Equal("process-guid"))

						Expect(fakeAuctioneerClient.RequestLRPAuctionsCallCount()).To(Equal(1))
						startRequests := fakeAuctioneerClient.RequestLRPAuctionsArgsForCall(0)
						Expect(startRequests).To(HaveLen(1))
						schedulingInfo := desiredLRP.DesiredLRPSchedulingInfo()
						expectedStartRequest := auctioneer.NewLRPStartRequestFromSchedulingInfo(&schedulingInfo, 1)
						Expect(startRequests[0]).To(BeEquivalentTo(&expectedStartRequest))
					})
				})

				Context("when the actual lrp should not be restarted (e.g., crashed)", func() {
					BeforeEach(func() {
						fakeActualLRPDB.CrashActualLRPReturns(&models.ActualLRPGroup{Instance: &actualLRP}, &models.ActualLRPGroup{Instance: &actualLRP}, false, nil)
					})

					It("does not request an auction", func() {
						Expect(responseRecorder.Code).To(Equal(http.StatusOK))
						response := &models.ActualLRPLifecycleResponse{}
						err := response.Unmarshal(responseRecorder.Body.Bytes())
						Expect(err).NotTo(HaveOccurred())
						Expect(response.Error).To(BeNil())

						Expect(fakeAuctioneerClient.RequestLRPAuctionsCallCount()).To(Equal(0))
					})
				})

				Context("when the DB returns an unrecoverable error", func() {
					BeforeEach(func() {
						fakeDesiredLRPDB.DesiredLRPByProcessGuidReturns(nil, models.NewUnrecoverableError(nil))
					})

					It("logs and writes to the exit channel", func() {
						Eventually(logger).Should(gbytes.Say("unrecoverable-error"))
						Eventually(exitCh).Should(Receive())
					})
				})

				Context("when fetching the desired lrp fails", func() {
					BeforeEach(func() {
						fakeDesiredLRPDB.DesiredLRPByProcessGuidReturns(nil, errors.New("error occured"))
					})

					It("fails and does not request an auction", func() {
						Expect(responseRecorder.Code).To(Equal(http.StatusOK))
						response := &models.ActualLRPLifecycleResponse{}
						err := response.Unmarshal(responseRecorder.Body.Bytes())
						Expect(err).NotTo(HaveOccurred())
						Expect(response.Error.Error()).To(Equal("error occured"))

						Expect(fakeAuctioneerClient.RequestLRPAuctionsCallCount()).To(Equal(0))
					})
				})

				Context("when requesting the auction fails", func() {
					BeforeEach(func() {
						fakeAuctioneerClient.RequestLRPAuctionsReturns(errors.New("some else bid higher"))
					})

					It("returns an error", func() {
						Expect(responseRecorder.Code).To(Equal(http.StatusOK))
						response := &models.ActualLRPLifecycleResponse{}
						err := response.Unmarshal(responseRecorder.Body.Bytes())
						Expect(err).NotTo(HaveOccurred())
						Expect(response.Error.Error()).To(Equal("some else bid higher"))
					})
				})
			})
		})

		Context("when the DB returns an unrecoverable error", func() {
			BeforeEach(func() {
				fakeActualLRPDB.CrashActualLRPReturns(nil, nil, false, models.NewUnrecoverableError(nil))
			})

			It("logs and writes to the exit channel", func() {
				Eventually(logger).Should(gbytes.Say("unrecoverable-error"))
				Eventually(exitCh).Should(Receive())
			})
		})

		Context("when crashing the actual lrp fails", func() {
			BeforeEach(func() {
				fakeActualLRPDB.CrashActualLRPReturns(nil, nil, false, models.ErrUnknownError)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := &models.ActualLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrUnknownError))
			})

			It("does not emit a change event to the hub", func() {
				Consistently(actualHub.EmitCallCount).Should(Equal(0))
			})
		})

		Context("when we cannot find the resource", func() {
			BeforeEach(func() {
				fakeActualLRPDB.CrashActualLRPReturns(nil, nil, false, models.ErrResourceNotFound)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := &models.ActualLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrResourceNotFound))
			})

			It("does not emit a change event to the hub", func() {
				Consistently(actualHub.EmitCallCount).Should(Equal(0))
			})
		})
	})

	Describe("RetireActualLRP", func() {
		var (
			request     *http.Request
			response    *models.ActualLRPLifecycleResponse
			processGuid = "process-guid"
			index       = int32(1)

			key models.ActualLRPKey

			requestBody interface{}

			actualLRPGroup *models.ActualLRPGroup
		)

		BeforeEach(func() {
			key = models.NewActualLRPKey(
				processGuid,
				index,
				"domain-0",
			)

			requestBody = &models.RetireActualLRPRequest{
				ActualLrpKey: &key,
			}

			actualLRP = models.ActualLRP{
				ActualLRPKey: key,
				State:        models.ActualLRPStateUnclaimed,
				Since:        1138,
			}

			actualLRPGroup = &models.ActualLRPGroup{
				Instance: &actualLRP,
			}
		})

		JustBeforeEach(func() {
			request = newTestRequest(requestBody)
			handler.RetireActualLRP(logger, responseRecorder, request)

			response = &models.ActualLRPLifecycleResponse{}
			err := response.Unmarshal(responseRecorder.Body.Bytes())
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the DB returns an unrecoverable error", func() {
			BeforeEach(func() {
				fakeActualLRPDB.ActualLRPGroupByProcessGuidAndIndexReturns(nil, models.NewUnrecoverableError(nil))
			})

			It("logs and writes to the exit channel", func() {
				Eventually(logger).Should(gbytes.Say("unrecoverable-error"))
				Eventually(exitCh).Should(Receive())
			})
		})

		Context("when finding the actualLRP fails", func() {
			BeforeEach(func() {
				fakeActualLRPDB.ActualLRPGroupByProcessGuidAndIndexReturns(nil, errors.New("could not find lrp"))
			})

			It("returns an error and does not retry", func() {
				Expect(response.Error.Message).To(Equal("could not find lrp"))
				Expect(fakeActualLRPDB.ActualLRPGroupByProcessGuidAndIndexCallCount()).To(Equal(1))
			})

			It("does not emit a change event to the hub", func() {
				Consistently(actualHub.EmitCallCount).Should(Equal(0))
			})
		})

		Context("when there is no instance in the actual lrp group", func() {
			BeforeEach(func() {
				actualLRPGroup.Instance = nil
				fakeActualLRPDB.ActualLRPGroupByProcessGuidAndIndexReturns(actualLRPGroup, nil)
			})

			It("returns an error and does not retry", func() {
				Expect(response.Error).To(Equal(models.ErrResourceNotFound))
				Expect(fakeActualLRPDB.ActualLRPGroupByProcessGuidAndIndexCallCount()).To(Equal(1))
			})

			It("does not emit a change event to the hub", func() {
				Consistently(actualHub.EmitCallCount).Should(Equal(0))
			})
		})

		Context("with an Unclaimed LRP", func() {
			BeforeEach(func() {
				fakeActualLRPDB.ActualLRPGroupByProcessGuidAndIndexReturns(actualLRPGroup, nil)
			})

			It("removes the LRP", func() {
				Expect(response.Error).To(BeNil())
				Expect(fakeActualLRPDB.RemoveActualLRPCallCount()).To(Equal(1))

				_, deletedLRPGuid, deletedLRPIndex, deletedLRPInstanceKey := fakeActualLRPDB.RemoveActualLRPArgsForCall(0)
				Expect(deletedLRPGuid).To(Equal(processGuid))
				Expect(deletedLRPIndex).To(Equal(index))
				Expect(deletedLRPInstanceKey).To(Equal(&actualLRPGroup.Instance.ActualLRPInstanceKey))
			})

			It("emits a removed event to the hub", func() {
				Eventually(actualHub.EmitCallCount).Should(Equal(1))
				event := actualHub.EmitArgsForCall(0)
				removedEvent, ok := event.(*models.ActualLRPRemovedEvent)
				Expect(ok).To(BeTrue())
				Expect(removedEvent.ActualLrpGroup).To(Equal(&models.ActualLRPGroup{Instance: &actualLRP}))
			})

			Context("when removing the actual lrp fails", func() {
				BeforeEach(func() {
					fakeActualLRPDB.RemoveActualLRPReturns(errors.New("boom!"))
				})

				It("retries removing up to RetireActualLRPRetryAttempts times", func() {
					Expect(response.Error.Message).To(Equal("boom!"))
					Expect(fakeActualLRPDB.RemoveActualLRPCallCount()).To(Equal(5))
					Expect(fakeActualLRPDB.ActualLRPGroupByProcessGuidAndIndexCallCount()).To(Equal(5))
				})

				It("does not emit a change event to the hub", func() {
					Consistently(actualHub.EmitCallCount).Should(Equal(0))
				})
			})
		})

		Context("when the LRP is crashed", func() {
			BeforeEach(func() {
				actualLRPGroup.Instance.State = models.ActualLRPStateCrashed
				fakeActualLRPDB.ActualLRPGroupByProcessGuidAndIndexReturns(actualLRPGroup, nil)
			})

			It("removes the LRP", func() {
				Expect(response.Error).To(BeNil())
				Expect(fakeActualLRPDB.RemoveActualLRPCallCount()).To(Equal(1))

				_, deletedLRPGuid, deletedLRPIndex, deletedLRPInstanceKey := fakeActualLRPDB.RemoveActualLRPArgsForCall(0)
				Expect(deletedLRPGuid).To(Equal(processGuid))
				Expect(deletedLRPIndex).To(Equal(index))
				Expect(deletedLRPInstanceKey).To(Equal(&actualLRPGroup.Instance.ActualLRPInstanceKey))
			})

			It("emits a removed event to the hub", func() {
				Eventually(actualHub.EmitCallCount).Should(Equal(1))
				event := actualHub.EmitArgsForCall(0)
				removedEvent, ok := event.(*models.ActualLRPRemovedEvent)
				Expect(ok).To(BeTrue())
				Expect(removedEvent.ActualLrpGroup).To(Equal(&models.ActualLRPGroup{Instance: &actualLRP}))
			})

			Context("when removing the actual lrp fails", func() {
				BeforeEach(func() {
					fakeActualLRPDB.RemoveActualLRPReturns(errors.New("boom!"))
				})

				It("retries removing up to RetireActualLRPRetryAttempts times", func() {
					Expect(response.Error.Message).To(Equal("boom!"))
					Expect(fakeActualLRPDB.RemoveActualLRPCallCount()).To(Equal(5))
					Expect(fakeActualLRPDB.ActualLRPGroupByProcessGuidAndIndexCallCount()).To(Equal(5))
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

				actualLRP.CellId = cellID
				actualLRP.ActualLRPInstanceKey = instanceKey
				actualLRPGroup.Instance.State = models.ActualLRPStateClaimed
				fakeActualLRPDB.ActualLRPGroupByProcessGuidAndIndexReturns(actualLRPGroup, nil)
			})

			Context("when the cell", func() {
				Context("is present", func() {
					BeforeEach(func() {
						cellPresence = models.NewCellPresence(
							cellID,
							"cell1.addr",
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
						stoppedKey, stoppedInstanceKey := fakeRepClient.StopLRPInstanceArgsForCall(0)
						Expect(stoppedKey).To(Equal(key))
						Expect(stoppedInstanceKey).To(Equal(instanceKey))
					})

					Context("Stopping the LRP fails", func() {
						BeforeEach(func() {
							fakeRepClient.StopLRPInstanceReturns(errors.New("Failed to stop app"))
						})

						It("retries to stop the app", func() {
							Expect(response.Error.Error()).To(Equal("Failed to stop app"))
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
							Expect(response.Error).To(BeNil())
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
							Expect(removedEvent.ActualLrpGroup).To(Equal(&models.ActualLRPGroup{Instance: &actualLRP}))
						})
					})

					Context("removing the actualLRP fails", func() {
						BeforeEach(func() {
							fakeActualLRPDB.RemoveActualLRPReturns(errors.New("failed to delete actual LRP"))
						})

						It("returns an error and does not retry", func() {
							Expect(response.Error.Message).To(Equal("failed to delete actual LRP"))
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
						Expect(response.Error.Error()).To(Equal("cell error"))
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

	Describe("FailActualLRP", func() {
		var (
			request     *http.Request
			processGuid = "process-guid"
			index       = int32(1)

			key          models.ActualLRPKey
			errorMessage string

			requestBody interface{}
		)

		BeforeEach(func() {
			key = models.NewActualLRPKey(
				processGuid,
				index,
				"domain-0",
			)
			errorMessage = "something went wrong"
			requestBody = &models.FailActualLRPRequest{
				ActualLrpKey: &key,
				ErrorMessage: errorMessage,
			}

			actualLRP = models.ActualLRP{
				ActualLRPKey: key,
				State:        models.ActualLRPStateUnclaimed,
				Since:        1138,
			}

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
		})

		JustBeforeEach(func() {
			request = newTestRequest(requestBody)
			handler.FailActualLRP(logger, responseRecorder, request)
		})

		Context("when failing the actual lrp in the DB succeeds", func() {
			BeforeEach(func() {
				fakeActualLRPDB.FailActualLRPReturns(&models.ActualLRPGroup{Instance: &actualLRP}, &models.ActualLRPGroup{Instance: &afterActualLRP}, nil)
			})

			It("fails the actual lrp by process guid and index", func() {
				Expect(fakeActualLRPDB.FailActualLRPCallCount()).To(Equal(1))
				_, actualKey, actualErrorMessage := fakeActualLRPDB.FailActualLRPArgsForCall(0)
				Expect(*actualKey).To(Equal(key))
				Expect(actualErrorMessage).To(Equal(errorMessage))

				Expect(responseRecorder.Code).To(Equal(http.StatusOK))

				response := &models.ActualLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(BeNil())
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

		Context("when the DB returns an unrecoverable error", func() {
			BeforeEach(func() {
				fakeActualLRPDB.FailActualLRPReturns(nil, nil, models.NewUnrecoverableError(nil))
			})

			It("logs and writes to the exit channel", func() {
				Eventually(logger).Should(gbytes.Say("unrecoverable-error"))
				Eventually(exitCh).Should(Receive())
			})
		})

		Context("when failing the actual lrp fails", func() {
			BeforeEach(func() {
				fakeActualLRPDB.FailActualLRPReturns(nil, nil, models.ErrUnknownError)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := &models.ActualLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrUnknownError))
			})

			It("does not emit a change event to the hub", func() {
				Consistently(actualHub.EmitCallCount).Should(Equal(0))
			})
		})

		Context("when we cannot find the resource", func() {
			BeforeEach(func() {
				fakeActualLRPDB.FailActualLRPReturns(nil, nil, models.ErrResourceNotFound)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := &models.ActualLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrResourceNotFound))
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

			requestBody interface{}
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

			requestBody = &models.RemoveActualLRPRequest{
				ProcessGuid:          processGuid,
				Index:                index,
				ActualLrpInstanceKey: &instanceKey,
			}

			fakeActualLRPDB.ActualLRPGroupByProcessGuidAndIndexReturns(&models.ActualLRPGroup{Instance: &actualLRP}, nil)
		})

		JustBeforeEach(func() {
			request := newTestRequest(requestBody)
			handler.RemoveActualLRP(logger, responseRecorder, request)
		})

		Context("when removing the actual lrp in the DB succeeds", func() {
			var removedActualLRP models.ActualLRP

			BeforeEach(func() {
				removedActualLRP = actualLRP
				removedActualLRP.ActualLRPInstanceKey = instanceKey
				fakeActualLRPDB.RemoveActualLRPReturns(nil)
			})

			It("removes the actual lrp by process guid and index", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				Expect(fakeActualLRPDB.RemoveActualLRPCallCount()).To(Equal(1))

				_, actualProcessGuid, idx, actualInstanceKey := fakeActualLRPDB.RemoveActualLRPArgsForCall(0)
				Expect(actualProcessGuid).To(Equal(processGuid))
				Expect(idx).To(BeEquivalentTo(index))
				Expect(actualInstanceKey).To(Equal(&instanceKey))
			})

			It("response with no error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := &models.ActualLRPLifecycleResponse{}

				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())
				Expect(response.Error).To(BeNil())
			})

			It("emits a removed event to the hub", func() {
				Eventually(actualHub.EmitCallCount).Should(Equal(1))
				event := actualHub.EmitArgsForCall(0)
				removedEvent, ok := event.(*models.ActualLRPRemovedEvent)
				Expect(ok).To(BeTrue())
				Expect(removedEvent.ActualLrpGroup).To(Equal(&models.ActualLRPGroup{Instance: &actualLRP}))
			})
		})

		Context("when the DB returns an unrecoverable error", func() {
			Context("when doing the actual LRP lookup", func() {
				BeforeEach(func() {
					fakeActualLRPDB.ActualLRPGroupByProcessGuidAndIndexReturns(nil, models.NewUnrecoverableError(nil))
				})

				It("logs and writes to the exit channel", func() {
					Eventually(logger).Should(gbytes.Say("unrecoverable-error"))
					Eventually(exitCh).Should(Receive())
				})
			})

			Context("when doing the actual LRP removal", func() {
				BeforeEach(func() {
					fakeActualLRPDB.RemoveActualLRPReturns(models.NewUnrecoverableError(nil))
				})

				It("logs and writes to the exit channel", func() {
					Eventually(logger).Should(gbytes.Say("unrecoverable-error"))
					Eventually(exitCh).Should(Receive())
				})
			})
		})

		Context("when removing the actual lrp fails", func() {
			BeforeEach(func() {
				fakeActualLRPDB.RemoveActualLRPReturns(models.ErrUnknownError)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := &models.ActualLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrUnknownError))
			})

			It("does not emit a change event to the hub", func() {
				Consistently(actualHub.EmitCallCount).Should(Equal(0))
			})
		})

		Context("when we cannot find the resource", func() {
			BeforeEach(func() {
				fakeActualLRPDB.RemoveActualLRPReturns(models.ErrResourceNotFound)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := &models.ActualLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrResourceNotFound))
			})

			It("does not emit a change event to the hub", func() {
				Consistently(actualHub.EmitCallCount).Should(Equal(0))
			})
		})
	})
})
