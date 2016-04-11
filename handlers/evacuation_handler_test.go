package handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry-incubator/auctioneer"
	"github.com/cloudfoundry-incubator/auctioneer/auctioneerfakes"
	"github.com/cloudfoundry-incubator/bbs/db/fakes"
	"github.com/cloudfoundry-incubator/bbs/handlers"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/models/test/model_helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("Evacuation Handlers", func() {
	var (
		logger               lager.Logger
		fakeEvacuationDB     *fakes.FakeEvacuationDB
		fakeActualLRPDB      *fakes.FakeActualLRPDB
		fakeDesiredLRPDB     *fakes.FakeDesiredLRPDB
		fakeAuctioneerClient *auctioneerfakes.FakeClient
		responseRecorder     *httptest.ResponseRecorder
		handler              *handlers.EvacuationHandler
	)

	BeforeEach(func() {
		fakeEvacuationDB = new(fakes.FakeEvacuationDB)
		fakeActualLRPDB = new(fakes.FakeActualLRPDB)
		fakeDesiredLRPDB = new(fakes.FakeDesiredLRPDB)
		fakeAuctioneerClient = new(auctioneerfakes.FakeClient)
		logger = lagertest.NewTestLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		responseRecorder = httptest.NewRecorder()
		handler = handlers.NewEvacuationHandler(logger, fakeEvacuationDB, fakeActualLRPDB, fakeDesiredLRPDB, fakeAuctioneerClient)
	})

	Describe("RemoveEvacuatingActualLRP", func() {
		var (
			processGuid = "process-guid"
			index       = int32(1)

			key         models.ActualLRPKey
			instanceKey models.ActualLRPInstanceKey

			requestBody interface{}
		)

		BeforeEach(func() {
			key = models.NewActualLRPKey(
				processGuid,
				index,
				"domain-0",
			)
			instanceKey = models.NewActualLRPInstanceKey("instance-guid", "cell-id")
			requestBody = &models.RemoveEvacuatingActualLRPRequest{
				ActualLrpKey:         &key,
				ActualLrpInstanceKey: &instanceKey,
			}
		})

		JustBeforeEach(func() {
			request := newTestRequest(requestBody)
			handler.RemoveEvacuatingActualLRP(responseRecorder, request)
		})

		Context("when removeEvacuatinging the actual lrp in the DB succeeds", func() {
			BeforeEach(func() {
				fakeEvacuationDB.RemoveEvacuatingActualLRPReturns(nil)
			})

			It("removeEvacuatings the actual lrp by process guid and index", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))

				Expect(fakeEvacuationDB.RemoveEvacuatingActualLRPCallCount()).To(Equal(1))
				_, actualKey, actualInstanceKey := fakeEvacuationDB.RemoveEvacuatingActualLRPArgsForCall(0)
				Expect(*actualKey).To(Equal(key))
				Expect(*actualInstanceKey).To(Equal(instanceKey))
			})
		})

		Context("when the request is invalid", func() {
			BeforeEach(func() {
				requestBody = &models.RemoveEvacuatingActualLRPRequest{}
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))

				var response models.RemoveEvacuatingActualLRPResponse
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).NotTo(BeNil())
				Expect(response.Error.Type).To(Equal(models.Error_InvalidRequest))
			})
		})

		Context("when parsing the body fails", func() {
			BeforeEach(func() {
				requestBody = "beep boop beep boop -- i am a robot"
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))

				var response models.RemoveEvacuatingActualLRPResponse
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).NotTo(BeNil())
				Expect(response.Error).To(Equal(models.ErrBadRequest))
			})
		})

		Context("when DB errors out", func() {
			BeforeEach(func() {
				fakeEvacuationDB.RemoveEvacuatingActualLRPReturns(models.ErrUnknownError)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))

				var response models.RemoveEvacuatingActualLRPResponse
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).NotTo(BeNil())
				Expect(response.Error).To(Equal(models.ErrUnknownError))
			})
		})

		Context("when we cannot find the resource", func() {
			BeforeEach(func() {
				fakeEvacuationDB.RemoveEvacuatingActualLRPReturns(models.ErrResourceNotFound)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))

				var response models.RemoveEvacuatingActualLRPResponse
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).NotTo(BeNil())
				Expect(response.Error).To(Equal(models.ErrResourceNotFound))
			})
		})
	})

	Describe("EvacuateClaimedActualLRP", func() {
		var (
			request     *http.Request
			requestBody *models.EvacuateClaimedActualLRPRequest
			actual      *models.ActualLRP
			desiredLRP  *models.DesiredLRP
		)

		BeforeEach(func() {
			desiredLRP = model_helpers.NewValidDesiredLRP("the-guid")
			fakeDesiredLRPDB.DesiredLRPByProcessGuidReturns(desiredLRP, nil)

			actual = model_helpers.NewValidActualLRP("process-guid", 1)
			requestBody = &models.EvacuateClaimedActualLRPRequest{
				ActualLrpKey:         &actual.ActualLRPKey,
				ActualLrpInstanceKey: &actual.ActualLRPInstanceKey,
			}

			request = newTestRequest(requestBody)
		})

		JustBeforeEach(func() {
			handler.EvacuateClaimedActualLRP(responseRecorder, request)
			Expect(responseRecorder.Code).To(Equal(http.StatusOK))
		})

		It("removes the evacuating actual lrp", func() {
			response := models.EvacuationResponse{}
			err := response.Unmarshal(responseRecorder.Body.Bytes())
			Expect(err).NotTo(HaveOccurred())
			Expect(response.KeepContainer).To(BeFalse())
			Expect(response.Error).To(BeNil())

			Expect(fakeEvacuationDB.RemoveEvacuatingActualLRPCallCount()).To(Equal(1))
			_, key, instanceKey := fakeEvacuationDB.RemoveEvacuatingActualLRPArgsForCall(0)
			Expect(*key).To(Equal(actual.ActualLRPKey))
			Expect(*instanceKey).To(Equal(actual.ActualLRPInstanceKey))
		})

		It("unclaims the actual lrp instance and requests an auction", func() {
			response := models.EvacuationResponse{}
			err := response.Unmarshal(responseRecorder.Body.Bytes())
			Expect(err).NotTo(HaveOccurred())
			Expect(response.KeepContainer).To(BeFalse())
			Expect(response.Error).To(BeNil())

			Expect(fakeActualLRPDB.UnclaimActualLRPCallCount()).To(Equal(1))
			_, lrpKey := fakeActualLRPDB.UnclaimActualLRPArgsForCall(0)
			Expect(lrpKey.ProcessGuid).To(Equal("process-guid"))
			Expect(lrpKey.Index).To(BeEquivalentTo(1))

			Expect(fakeDesiredLRPDB.DesiredLRPByProcessGuidCallCount()).To(Equal(1))
			_, guid := fakeDesiredLRPDB.DesiredLRPByProcessGuidArgsForCall(0)
			Expect(guid).To(Equal("process-guid"))

			expectedStartRequest := auctioneer.NewLRPStartRequestFromModel(desiredLRP, int(actual.Index))
			Expect(fakeAuctioneerClient.RequestLRPAuctionsCallCount()).To(Equal(1))
			startRequests := fakeAuctioneerClient.RequestLRPAuctionsArgsForCall(0)
			Expect(startRequests).To(Equal([]*auctioneer.LRPStartRequest{&expectedStartRequest}))
		})

		Context("when removing the evacuating lrp fails", func() {
			BeforeEach(func() {
				fakeEvacuationDB.RemoveEvacuatingActualLRPReturns(errors.New("i failed"))
			})

			It("logs the error and continues", func() {
				response := models.EvacuationResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())
				Expect(response.KeepContainer).To(BeFalse())
				Expect(response.Error).To(BeNil())
				Expect(logger).To(gbytes.Say("failed-removing-evacuating-actual-lrp"))
			})
		})

		Context("when unclaiming the lrp instance fails", func() {
			Context("because the instance does not exist", func() {
				BeforeEach(func() {
					fakeActualLRPDB.UnclaimActualLRPReturns(models.ErrResourceNotFound)
				})

				It("does not keep the container and does not return an error", func() {
					response := models.EvacuationResponse{}
					err := response.Unmarshal(responseRecorder.Body.Bytes())
					Expect(err).NotTo(HaveOccurred())
					Expect(response.KeepContainer).To(BeFalse())
					Expect(response.Error).To(BeNil())
				})
			})

			Context("for another reason", func() {
				BeforeEach(func() {
					fakeActualLRPDB.UnclaimActualLRPReturns(errors.New("can't unclaim this"))
				})

				It("returns the error and keeps the container", func() {
					response := models.EvacuationResponse{}
					err := response.Unmarshal(responseRecorder.Body.Bytes())
					Expect(err).NotTo(HaveOccurred())
					Expect(response.KeepContainer).To(BeTrue())
					Expect(response.Error).NotTo(BeNil())
					Expect(response.Error.Error()).To(Equal("can't unclaim this"))
				})
			})
		})

		Context("when requesting the lrp auction fails", func() {
			BeforeEach(func() {
				fakeAuctioneerClient.RequestLRPAuctionsReturns(errors.New("boom!"))
			})

			It("does not return the error or keep the container", func() {
				response := models.EvacuationResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())
				Expect(response.KeepContainer).To(BeFalse())
				Expect(response.Error).To(BeNil())
			})
		})

		Context("when the request is invalid", func() {
			BeforeEach(func() {
				request = newTestRequest("{{")
			})

			It("returns an error and keeps the container", func() {
				response := models.EvacuationResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())
				Expect(response.KeepContainer).To(BeTrue())
				Expect(response.Error).NotTo(BeNil())
				Expect(response.Error).To(Equal(models.ErrBadRequest))
			})
		})
	})

	Describe("EvacuateCrashedActualLRP", func() {
		var (
			request     *http.Request
			requestBody *models.EvacuateCrashedActualLRPRequest
			actual      *models.ActualLRP
		)

		BeforeEach(func() {
			actual = model_helpers.NewValidActualLRP("process-guid", 1)
			requestBody = &models.EvacuateCrashedActualLRPRequest{
				ActualLrpKey:         &actual.ActualLRPKey,
				ActualLrpInstanceKey: &actual.ActualLRPInstanceKey,
				ErrorMessage:         "i failed",
			}

			request = newTestRequest(requestBody)
		})

		JustBeforeEach(func() {
			handler.EvacuateCrashedActualLRP(responseRecorder, request)
			Expect(responseRecorder.Code).To(Equal(http.StatusOK))
		})

		It("does not return an error or keep the container", func() {
			response := models.EvacuationResponse{}
			err := response.Unmarshal(responseRecorder.Body.Bytes())
			Expect(err).NotTo(HaveOccurred())
			Expect(response.KeepContainer).To(BeFalse())
			Expect(response.Error).To(BeNil())
		})

		It("removes the evacuating actual lrp", func() {
			Expect(fakeEvacuationDB.RemoveEvacuatingActualLRPCallCount()).To(Equal(1))
			_, key, instanceKey := fakeEvacuationDB.RemoveEvacuatingActualLRPArgsForCall(0)
			Expect(*key).To(Equal(actual.ActualLRPKey))
			Expect(*instanceKey).To(Equal(actual.ActualLRPInstanceKey))
		})

		It("crashes the actual lrp instance", func() {
			Expect(fakeActualLRPDB.CrashActualLRPCallCount()).To(Equal(1))
			_, key, instanceKey, errorMessage := fakeActualLRPDB.CrashActualLRPArgsForCall(0)
			Expect(*key).To(Equal(actual.ActualLRPKey))
			Expect(*instanceKey).To(Equal(actual.ActualLRPInstanceKey))
			Expect(errorMessage).To(Equal("i failed"))
		})

		Context("when removing the evacuating actual lrp fails", func() {
			BeforeEach(func() {
				fakeEvacuationDB.RemoveEvacuatingActualLRPReturns(errors.New("oh no!"))
			})

			It("logs the error and continues", func() {
				response := models.EvacuationResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())
				Expect(response.KeepContainer).To(BeFalse())
				Expect(response.Error).To(BeNil())

				Expect(logger).To(gbytes.Say("failed-removing-evacuating-actual-lrp"))
			})
		})

		Context("when crashing the actual lrp fails", func() {
			Context("because the error does not exist", func() {
				BeforeEach(func() {
					fakeActualLRPDB.CrashActualLRPReturns(false, models.ErrResourceNotFound)
				})

				It("does not return an error or keep the container", func() {
					response := models.EvacuationResponse{}
					err := response.Unmarshal(responseRecorder.Body.Bytes())
					Expect(err).NotTo(HaveOccurred())
					Expect(response.KeepContainer).To(BeFalse())
					Expect(response.Error).To(BeNil())
				})
			})

			Context("for another reason", func() {
				BeforeEach(func() {
					fakeActualLRPDB.CrashActualLRPReturns(false, errors.New("failed-crashing-dawg"))
				})

				It("returns an error and does not keep the container", func() {
					response := models.EvacuationResponse{}
					err := response.Unmarshal(responseRecorder.Body.Bytes())
					Expect(err).NotTo(HaveOccurred())
					Expect(response.KeepContainer).To(BeFalse())
					Expect(response.Error).NotTo(BeNil())
					Expect(response.Error.Error()).To(Equal("failed-crashing-dawg"))
				})
			})
		})

		Context("when the request is invalid", func() {
			BeforeEach(func() {
				request = newTestRequest("{{")
			})

			It("returns an error", func() {
				response := models.EvacuationResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())
				Expect(response.KeepContainer).To(BeFalse())
				Expect(response.Error).To(Equal(models.ErrBadRequest))
			})
		})
	})

	Describe("EvacuateRunningActualLRP", func() {
		var (
			request        *http.Request
			requestBody    *models.EvacuateRunningActualLRPRequest
			actualLRP      *models.ActualLRP
			actualLRPGroup *models.ActualLRPGroup
			desiredLRP     *models.DesiredLRP
		)

		BeforeEach(func() {
			request = nil
			desiredLRP = model_helpers.NewValidDesiredLRP("the-guid")
			fakeDesiredLRPDB.DesiredLRPByProcessGuidReturns(desiredLRP, nil)

			actualLRP = model_helpers.NewValidActualLRP("the-guid", 1)

			key := actualLRP.ActualLRPKey
			instanceKey := actualLRP.ActualLRPInstanceKey
			netInfo := actualLRP.ActualLRPNetInfo
			requestBody = &models.EvacuateRunningActualLRPRequest{
				ActualLrpKey:         &key,
				ActualLrpInstanceKey: &instanceKey,
				ActualLrpNetInfo:     &netInfo,
				Ttl:                  60,
			}

			actualLRPGroup = &models.ActualLRPGroup{
				Instance: actualLRP,
			}

			fakeActualLRPDB.ActualLRPGroupByProcessGuidAndIndexReturns(actualLRPGroup, nil)
		})

		JustBeforeEach(func() {
			if request == nil {
				request = newTestRequest(requestBody)
			}
			handler.EvacuateRunningActualLRP(responseRecorder, request)
			Expect(responseRecorder.Code).To(Equal(http.StatusOK))
		})

		Context("when the actual lrp group exists", func() {
			Context("when the actual lrp instance does not exist", func() {
				BeforeEach(func() {
					actualLRPGroup.Instance = nil
				})

				It("removes the evacuating lrp and does not keep the container", func() {
					response := models.EvacuationResponse{}
					err := response.Unmarshal(responseRecorder.Body.Bytes())
					Expect(err).NotTo(HaveOccurred())
					Expect(response.KeepContainer).To(BeFalse())
					Expect(response.Error).To(BeNil())

					Expect(fakeEvacuationDB.RemoveEvacuatingActualLRPCallCount()).To(Equal(1))
					_, actualLRPKey, actualLRPInstanceKey := fakeEvacuationDB.RemoveEvacuatingActualLRPArgsForCall(0)
					Expect(*actualLRPKey).To(Equal(actualLRP.ActualLRPKey))
					Expect(*actualLRPInstanceKey).To(Equal(actualLRP.ActualLRPInstanceKey))
				})

				Context("when the evacuating lrp cannot be removed", func() {
					BeforeEach(func() {
						fakeEvacuationDB.RemoveEvacuatingActualLRPReturns(models.ErrActualLRPCannotBeRemoved)
					})

					It("returns no error and removes the container", func() {
						response := models.EvacuationResponse{}
						err := response.Unmarshal(responseRecorder.Body.Bytes())
						Expect(err).NotTo(HaveOccurred())
						Expect(response.KeepContainer).To(BeFalse())
						Expect(response.Error).To(BeNil())
					})
				})

				Context("when removing the evacuating lrp fails for a different reason", func() {
					BeforeEach(func() {
						fakeEvacuationDB.RemoveEvacuatingActualLRPReturns(errors.New("didnt work"))
					})

					It("returns an error and keeps the container", func() {
						response := models.EvacuationResponse{}
						err := response.Unmarshal(responseRecorder.Body.Bytes())
						Expect(err).NotTo(HaveOccurred())
						Expect(response.KeepContainer).To(BeTrue())
						Expect(response.Error).NotTo(BeNil())
						Expect(response.Error.Error()).To(Equal("didnt work"))
					})
				})
			})

			Context("when the instance is unclaimed", func() {
				BeforeEach(func() {
					actualLRP.State = models.ActualLRPStateUnclaimed
				})

				Context("without a placement error", func() {
					BeforeEach(func() {
						actualLRP.PlacementError = ""
					})

					It("evacuates the LRP", func() {
						response := models.EvacuationResponse{}
						err := response.Unmarshal(responseRecorder.Body.Bytes())
						Expect(err).NotTo(HaveOccurred())
						Expect(response.KeepContainer).To(BeTrue())
						Expect(response.Error).To(BeNil())

						Expect(fakeEvacuationDB.EvacuateActualLRPCallCount()).To(Equal(1))
						_, actualLRPKey, actualLRPInstanceKey, actualLrpNetInfo, ttl := fakeEvacuationDB.EvacuateActualLRPArgsForCall(0)
						Expect(*actualLRPKey).To(Equal(actualLRP.ActualLRPKey))
						Expect(*actualLRPInstanceKey).To(Equal(actualLRP.ActualLRPInstanceKey))
						Expect(*actualLrpNetInfo).To(Equal(actualLRP.ActualLRPNetInfo))
						Expect(ttl).To(BeEquivalentTo(60))
					})

					Context("when there's an existing evacuating on another cell", func() {
						BeforeEach(func() {
							evacuatingLRP := model_helpers.NewValidActualLRP("the-guid", 1)
							evacuatingLRP.CellId = "some-other-cell"
							actualLRPGroup.Evacuating = evacuatingLRP
						})

						It("does not error and does not keep the container", func() {
							response := models.EvacuationResponse{}
							err := response.Unmarshal(responseRecorder.Body.Bytes())
							Expect(err).NotTo(HaveOccurred())
							Expect(response.KeepContainer).To(BeFalse())
							Expect(response.Error).To(BeNil())
						})
					})

					Context("when evacuating the actual lrp fails", func() {
						BeforeEach(func() {
							fakeEvacuationDB.EvacuateActualLRPReturns(errors.New("didnt work"))
						})

						It("returns an error and keeps the container", func() {
							response := models.EvacuationResponse{}
							err := response.Unmarshal(responseRecorder.Body.Bytes())
							Expect(err).NotTo(HaveOccurred())
							Expect(response.KeepContainer).To(BeTrue())
							Expect(response.Error).NotTo(BeNil())
							Expect(response.Error.Error()).To(Equal("didnt work"))
						})
					})

					Context("when the lrp cannot be evacuated", func() {
						BeforeEach(func() {
							fakeEvacuationDB.EvacuateActualLRPReturns(models.ErrActualLRPCannotBeEvacuated)
						})

						It("does not error and does not keep the container", func() {
							response := models.EvacuationResponse{}
							err := response.Unmarshal(responseRecorder.Body.Bytes())
							Expect(err).NotTo(HaveOccurred())
							Expect(response.KeepContainer).To(BeFalse())
							Expect(response.Error).To(BeNil())
						})
					})
				})

				Context("with a placement error", func() {
					BeforeEach(func() {
						actualLRP.PlacementError = "jim kinda likes cats, but loves kittens"
					})

					It("removes the evacuating LRP", func() {
						response := models.EvacuationResponse{}
						err := response.Unmarshal(responseRecorder.Body.Bytes())
						Expect(err).NotTo(HaveOccurred())
						Expect(response.KeepContainer).To(BeFalse())
						Expect(response.Error).To(BeNil())

						Expect(fakeEvacuationDB.RemoveEvacuatingActualLRPCallCount()).To(Equal(1))
						_, actualLRPKey, actualLRPInstanceKey := fakeEvacuationDB.RemoveEvacuatingActualLRPArgsForCall(0)
						Expect(*actualLRPKey).To(Equal(actualLRP.ActualLRPKey))
						Expect(*actualLRPInstanceKey).To(Equal(actualLRP.ActualLRPInstanceKey))
					})

					Context("when removing the evacuating LRP fails", func() {
						BeforeEach(func() {
							fakeEvacuationDB.RemoveEvacuatingActualLRPReturns(errors.New("oh nooo!!!"))
						})

						It("returns an error and does keep the container", func() {
							response := models.EvacuationResponse{}
							err := response.Unmarshal(responseRecorder.Body.Bytes())
							Expect(err).NotTo(HaveOccurred())
							Expect(response.KeepContainer).To(BeTrue())
							Expect(response.Error).NotTo(BeNil())
							Expect(response.Error.Error()).To(Equal("oh nooo!!!"))

							Expect(fakeEvacuationDB.RemoveEvacuatingActualLRPCallCount()).To(Equal(1))
						})

						Context("when the error is a ErrActualLRPCannotBeRemoved", func() {
							BeforeEach(func() {
								fakeEvacuationDB.RemoveEvacuatingActualLRPReturns(models.ErrActualLRPCannotBeRemoved)
							})

							It("does not return an error or keep the container", func() {
								response := models.EvacuationResponse{}
								err := response.Unmarshal(responseRecorder.Body.Bytes())
								Expect(err).NotTo(HaveOccurred())
								Expect(response.KeepContainer).To(BeFalse())
								Expect(response.Error).To(BeNil())
							})
						})
					})
				})
			})

			Context("when the instance is claimed", func() {
				BeforeEach(func() {
					actualLRP.State = models.ActualLRPStateClaimed
				})

				Context("by another cell", func() {
					BeforeEach(func() {
						actualLRP.CellId = "some-other-cell"
					})

					It("evacuates the LRP", func() {
						response := models.EvacuationResponse{}
						err := response.Unmarshal(responseRecorder.Body.Bytes())
						Expect(err).NotTo(HaveOccurred())
						Expect(response.KeepContainer).To(BeTrue())
						Expect(response.Error).To(BeNil())

						Expect(fakeEvacuationDB.EvacuateActualLRPCallCount()).To(Equal(1))
						_, actualLRPKey, actualLRPInstanceKey, actualLrpNetInfo, ttl := fakeEvacuationDB.EvacuateActualLRPArgsForCall(0)
						Expect(*actualLRPKey).To(Equal(actualLRP.ActualLRPKey))
						Expect(*actualLRPInstanceKey).To(Equal(*requestBody.ActualLrpInstanceKey))
						Expect(*actualLrpNetInfo).To(Equal(actualLRP.ActualLRPNetInfo))
						Expect(ttl).To(BeEquivalentTo(60))
					})

					Context("when there's an existing evacuating on another cell", func() {
						BeforeEach(func() {
							evacuatingLRP := model_helpers.NewValidActualLRP("the-guid", 1)
							evacuatingLRP.CellId = "some-other-cell"
							actualLRPGroup.Evacuating = evacuatingLRP
						})

						It("does not error and does not keep the container", func() {
							response := models.EvacuationResponse{}
							err := response.Unmarshal(responseRecorder.Body.Bytes())
							Expect(err).NotTo(HaveOccurred())
							Expect(response.KeepContainer).To(BeFalse())
							Expect(response.Error).To(BeNil())
						})
					})

					Context("when evacuating the actual lrp fails", func() {
						BeforeEach(func() {
							fakeEvacuationDB.EvacuateActualLRPReturns(errors.New("didnt work"))
						})

						It("returns an error and keeps the container", func() {
							response := models.EvacuationResponse{}
							err := response.Unmarshal(responseRecorder.Body.Bytes())
							Expect(err).NotTo(HaveOccurred())
							Expect(response.KeepContainer).To(BeTrue())
							Expect(response.Error).NotTo(BeNil())
							Expect(response.Error.Error()).To(Equal("didnt work"))
						})
					})

					Context("when the lrp cannot be evacuated", func() {
						BeforeEach(func() {
							fakeEvacuationDB.EvacuateActualLRPReturns(models.ErrActualLRPCannotBeEvacuated)
						})

						It("does not error and does not keep the container", func() {
							response := models.EvacuationResponse{}
							err := response.Unmarshal(responseRecorder.Body.Bytes())
							Expect(err).NotTo(HaveOccurred())
							Expect(response.KeepContainer).To(BeFalse())
							Expect(response.Error).To(BeNil())
						})
					})
				})

				Context("by the same cell", func() {
					It("evacuates the lrp", func() {
						response := models.EvacuationResponse{}
						err := response.Unmarshal(responseRecorder.Body.Bytes())
						Expect(err).NotTo(HaveOccurred())
						Expect(response.KeepContainer).To(BeTrue())
						Expect(response.Error).To(BeNil())

						Expect(fakeEvacuationDB.EvacuateActualLRPCallCount()).To(Equal(1))
						_, actualLRPKey, actualLRPInstanceKey, actualLrpNetInfo, ttl := fakeEvacuationDB.EvacuateActualLRPArgsForCall(0)
						Expect(*actualLRPKey).To(Equal(actualLRP.ActualLRPKey))
						Expect(*actualLRPInstanceKey).To(Equal(actualLRP.ActualLRPInstanceKey))
						Expect(*actualLrpNetInfo).To(Equal(actualLRP.ActualLRPNetInfo))
						Expect(ttl).To(BeEquivalentTo(60))
					})

					It("unclaims the lrp and requests an auction", func() {
						Expect(fakeActualLRPDB.UnclaimActualLRPCallCount()).To(Equal(1))
						_, actualLRPKey, actualLRPInstanceKey, actualLrpNetInfo, ttl := fakeEvacuationDB.EvacuateActualLRPArgsForCall(0)
						Expect(*actualLRPKey).To(Equal(actualLRP.ActualLRPKey))
						Expect(*actualLRPInstanceKey).To(Equal(actualLRP.ActualLRPInstanceKey))
						Expect(*actualLrpNetInfo).To(Equal(actualLRP.ActualLRPNetInfo))
						Expect(ttl).To(BeEquivalentTo(60))

						schedulingInfo := desiredLRP.DesiredLRPSchedulingInfo()
						expectedStartRequest := auctioneer.NewLRPStartRequestFromSchedulingInfo(&schedulingInfo, int(actualLRP.Index))

						Expect(fakeAuctioneerClient.RequestLRPAuctionsCallCount()).To(Equal(1))
						startRequests := fakeAuctioneerClient.RequestLRPAuctionsArgsForCall(0)
						Expect(startRequests).To(Equal([]*auctioneer.LRPStartRequest{&expectedStartRequest}))
					})

					Context("when evacuating fails", func() {
						BeforeEach(func() {
							fakeEvacuationDB.EvacuateActualLRPReturns(errors.New("this is a disaster"))
						})

						It("returns an error and keep the container", func() {
							response := models.EvacuationResponse{}
							err := response.Unmarshal(responseRecorder.Body.Bytes())
							Expect(err).NotTo(HaveOccurred())
							Expect(response.KeepContainer).To(BeTrue())
							Expect(response.Error).NotTo(BeNil())
							Expect(response.Error.Error()).To(Equal("this is a disaster"))
						})
					})

					Context("when unclaiming fails", func() {
						BeforeEach(func() {
							fakeActualLRPDB.UnclaimActualLRPReturns(errors.New("unclaiming failed"))
						})

						It("returns an error and keeps the contianer", func() {
							response := models.EvacuationResponse{}
							err := response.Unmarshal(responseRecorder.Body.Bytes())
							Expect(err).NotTo(HaveOccurred())
							Expect(response.KeepContainer).To(BeTrue())
							Expect(response.Error).NotTo(BeNil())
							Expect(response.Error.Error()).To(Equal("unclaiming failed"))
						})
					})
				})
			})

			Context("when the instance is running", func() {
				BeforeEach(func() {
					actualLRP.State = models.ActualLRPStateRunning
				})

				Context("on this cell", func() {
					It("evacuates the lrp and keeps the container", func() {
						response := models.EvacuationResponse{}
						err := response.Unmarshal(responseRecorder.Body.Bytes())
						Expect(err).NotTo(HaveOccurred())
						Expect(response.KeepContainer).To(BeTrue())
						Expect(response.Error).To(BeNil())

						Expect(fakeEvacuationDB.EvacuateActualLRPCallCount()).To(Equal(1))
						_, actualLRPKey, actualLRPInstanceKey, actualLrpNetInfo, ttl := fakeEvacuationDB.EvacuateActualLRPArgsForCall(0)
						Expect(*actualLRPKey).To(Equal(actualLRP.ActualLRPKey))
						Expect(*actualLRPInstanceKey).To(Equal(actualLRP.ActualLRPInstanceKey))
						Expect(*actualLrpNetInfo).To(Equal(actualLRP.ActualLRPNetInfo))
						Expect(ttl).To(BeEquivalentTo(60))
					})

					It("unclaims the lrp and requests an auction", func() {
						Expect(fakeActualLRPDB.UnclaimActualLRPCallCount()).To(Equal(1))
						_, lrpKey := fakeActualLRPDB.UnclaimActualLRPArgsForCall(0)
						Expect(lrpKey.ProcessGuid).To(Equal(actualLRP.ProcessGuid))
						Expect(lrpKey.Index).To(Equal(actualLRP.Index))

						schedulingInfo := desiredLRP.DesiredLRPSchedulingInfo()
						expectedStartRequest := auctioneer.NewLRPStartRequestFromSchedulingInfo(&schedulingInfo, int(actualLRP.Index))

						Expect(fakeAuctioneerClient.RequestLRPAuctionsCallCount()).To(Equal(1))
						startRequests := fakeAuctioneerClient.RequestLRPAuctionsArgsForCall(0)
						Expect(startRequests).To(Equal([]*auctioneer.LRPStartRequest{&expectedStartRequest}))
					})

					Context("when evacuating fails", func() {
						BeforeEach(func() {
							fakeEvacuationDB.EvacuateActualLRPReturns(errors.New("this is a disaster"))
						})

						It("returns an error and keep the container", func() {
							response := models.EvacuationResponse{}
							err := response.Unmarshal(responseRecorder.Body.Bytes())
							Expect(err).NotTo(HaveOccurred())
							Expect(response.KeepContainer).To(BeTrue())
							Expect(response.Error).NotTo(BeNil())
							Expect(response.Error.Error()).To(Equal("this is a disaster"))
						})
					})

					Context("when unclaiming fails", func() {
						BeforeEach(func() {
							fakeActualLRPDB.UnclaimActualLRPReturns(errors.New("unclaiming failed"))
						})

						It("returns an error and keeps the contianer", func() {
							response := models.EvacuationResponse{}
							err := response.Unmarshal(responseRecorder.Body.Bytes())
							Expect(err).NotTo(HaveOccurred())
							Expect(response.KeepContainer).To(BeTrue())
							Expect(response.Error).NotTo(BeNil())
							Expect(response.Error.Error()).To(Equal("unclaiming failed"))
						})
					})

					Context("when fetching the desired lrp fails", func() {
						BeforeEach(func() {
							fakeDesiredLRPDB.DesiredLRPByProcessGuidReturns(nil, errors.New("jolly rancher beer :/"))
						})

						It("does not return an error and keeps the container", func() {
							response := models.EvacuationResponse{}
							err := response.Unmarshal(responseRecorder.Body.Bytes())
							Expect(err).NotTo(HaveOccurred())
							Expect(response.KeepContainer).To(BeTrue())
							Expect(response.Error).To(BeNil())
						})
					})
				})

				Context("on another cell", func() {
					BeforeEach(func() {
						actualLRP.CellId = "some-other-cell"
					})

					It("removes the evacuating LRP", func() {
						response := models.EvacuationResponse{}
						err := response.Unmarshal(responseRecorder.Body.Bytes())
						Expect(err).NotTo(HaveOccurred())
						Expect(response.KeepContainer).To(BeFalse())
						Expect(response.Error).To(BeNil())

						Expect(fakeEvacuationDB.RemoveEvacuatingActualLRPCallCount()).To(Equal(1))
						_, actualLRPKey, actualLRPInstanceKey := fakeEvacuationDB.RemoveEvacuatingActualLRPArgsForCall(0)
						Expect(*actualLRPKey).To(Equal(actualLRP.ActualLRPKey))
						Expect(*actualLRPInstanceKey).To(Equal(actualLRP.ActualLRPInstanceKey))
					})

					Context("when removing the evacuating LRP fails", func() {
						BeforeEach(func() {
							fakeEvacuationDB.RemoveEvacuatingActualLRPReturns(errors.New("boom!"))
						})

						It("returns an error and does keep the container", func() {
							response := models.EvacuationResponse{}
							err := response.Unmarshal(responseRecorder.Body.Bytes())
							Expect(err).NotTo(HaveOccurred())
							Expect(response.KeepContainer).To(BeTrue())
							Expect(response.Error).NotTo(BeNil())
							Expect(response.Error.Error()).To(Equal("boom!"))
						})

						Context("when the error is a ErrActualLRPCannotBeRemoved", func() {
							BeforeEach(func() {
								fakeEvacuationDB.RemoveEvacuatingActualLRPReturns(models.ErrActualLRPCannotBeRemoved)
							})

							It("does not return an error or keep the container", func() {
								response := models.EvacuationResponse{}
								err := response.Unmarshal(responseRecorder.Body.Bytes())
								Expect(err).NotTo(HaveOccurred())
								Expect(response.KeepContainer).To(BeFalse())
								Expect(response.Error).To(BeNil())
							})
						})
					})
				})
			})

			Context("when the instance is crashed", func() {
				BeforeEach(func() {
					actualLRP.State = models.ActualLRPStateCrashed
				})

				It("removes the evacuating LRP", func() {
					response := models.EvacuationResponse{}
					err := response.Unmarshal(responseRecorder.Body.Bytes())
					Expect(err).NotTo(HaveOccurred())
					Expect(response.KeepContainer).To(BeFalse())
					Expect(response.Error).To(BeNil())

					Expect(fakeEvacuationDB.RemoveEvacuatingActualLRPCallCount()).To(Equal(1))
					_, actualLRPKey, actualLRPInstanceKey := fakeEvacuationDB.RemoveEvacuatingActualLRPArgsForCall(0)
					Expect(*actualLRPKey).To(Equal(actualLRP.ActualLRPKey))
					Expect(*actualLRPInstanceKey).To(Equal(actualLRP.ActualLRPInstanceKey))
				})

				Context("when removing the evacuating LRP fails", func() {
					BeforeEach(func() {
						fakeEvacuationDB.RemoveEvacuatingActualLRPReturns(errors.New("boom!"))
					})

					It("returns an error and does keep the container", func() {
						response := models.EvacuationResponse{}
						err := response.Unmarshal(responseRecorder.Body.Bytes())
						Expect(err).NotTo(HaveOccurred())
						Expect(response.KeepContainer).To(BeTrue())
						Expect(response.Error).NotTo(BeNil())
						Expect(response.Error.Error()).To(Equal("boom!"))
					})

					Context("when the error is a ErrActualLRPCannotBeRemoved", func() {
						BeforeEach(func() {
							fakeEvacuationDB.RemoveEvacuatingActualLRPReturns(models.ErrActualLRPCannotBeRemoved)
						})

						It("does not return an error or keep the container", func() {
							response := models.EvacuationResponse{}
							err := response.Unmarshal(responseRecorder.Body.Bytes())
							Expect(err).NotTo(HaveOccurred())
							Expect(response.KeepContainer).To(BeFalse())
							Expect(response.Error).To(BeNil())
						})
					})
				})
			})
		})

		Context("when the actual lrp group does not exist", func() {
			BeforeEach(func() {
				fakeActualLRPDB.ActualLRPGroupByProcessGuidAndIndexReturns(nil, models.ErrResourceNotFound)
			})

			It("does not return an error or keep the container", func() {
				response := models.EvacuationResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())
				Expect(response.KeepContainer).To(BeFalse())
				Expect(response.Error).To(BeNil())
			})
		})

		Context("when the request body is invalid", func() {
			BeforeEach(func() {
				request = newTestRequest("{{bad: stuff}")
			})

			It("returns an error and keeps the container", func() {
				response := models.EvacuationResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())
				Expect(response.KeepContainer).To(BeTrue())
				Expect(response.Error).NotTo(BeNil())
				Expect(response.Error).To(Equal(models.ErrBadRequest))
			})
		})
	})

	Describe("EvacuateStoppedActualLRP", func() {
		var (
			request *http.Request
			actual  *models.ActualLRP
		)

		BeforeEach(func() {
			actual = model_helpers.NewValidActualLRP("process-guid", 1)
			requestBody := &models.EvacuateStoppedActualLRPRequest{
				ActualLrpKey:         &actual.ActualLRPKey,
				ActualLrpInstanceKey: &actual.ActualLRPInstanceKey,
			}

			fakeActualLRPDB.ActualLRPGroupByProcessGuidAndIndexReturns(&models.ActualLRPGroup{
				Instance: actual,
			}, nil)

			request = newTestRequest(requestBody)
		})

		JustBeforeEach(func() {
			handler.EvacuateStoppedActualLRP(responseRecorder, request)
		})

		It("does not error and does not keep the container", func() {
			response := models.EvacuationResponse{}
			err := response.Unmarshal(responseRecorder.Body.Bytes())
			Expect(err).NotTo(HaveOccurred())
			Expect(response.KeepContainer).To(BeFalse())
			Expect(response.Error).To(BeNil())
		})

		It("removes the actual lrp", func() {
			Expect(fakeActualLRPDB.RemoveActualLRPCallCount()).To(Equal(1))
			_, guid, index := fakeActualLRPDB.RemoveActualLRPArgsForCall(0)
			Expect(guid).To(Equal("process-guid"))
			Expect(index).To(BeEquivalentTo(1))
		})

		It("removes the evacuating actual lrp", func() {
			Expect(fakeEvacuationDB.RemoveEvacuatingActualLRPCallCount()).To(Equal(1))
			_, lrpKey, lrpInstanceKey := fakeEvacuationDB.RemoveEvacuatingActualLRPArgsForCall(0)
			Expect(*lrpKey).To(Equal(actual.ActualLRPKey))
			Expect(*lrpInstanceKey).To(Equal(actual.ActualLRPInstanceKey))
		})

		Context("when the actual lrp is on a different cell", func() {
			BeforeEach(func() {
				actual.ActualLRPInstanceKey.CellId = "different-cell"
			})

			It("returns an error but does not keep the container", func() {
				response := models.EvacuationResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())
				Expect(response.KeepContainer).To(BeFalse())
				Expect(response.Error).To(Equal(models.ErrActualLRPCannotBeRemoved))
			})
		})

		Context("when removing the actual lrp fails", func() {
			BeforeEach(func() {
				fakeActualLRPDB.RemoveActualLRPReturns(errors.New("boom!"))
			})

			It("returns an error but does not keep the container", func() {
				response := models.EvacuationResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())
				Expect(response.KeepContainer).To(BeFalse())
				Expect(response.Error).NotTo(BeNil())
				Expect(response.Error.Error()).To(Equal("boom!"))
			})
		})

		Context("when fetching the actual lrp group fails", func() {
			BeforeEach(func() {
				fakeActualLRPDB.ActualLRPGroupByProcessGuidAndIndexReturns(nil, errors.New("i failed"))
			})

			It("returns an error but does not keep the container", func() {
				response := models.EvacuationResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())
				Expect(response.KeepContainer).To(BeFalse())
				Expect(response.Error).NotTo(BeNil())
				Expect(response.Error.Error()).To(Equal("i failed"))
			})
		})

		Context("when the request is invalid", func() {
			BeforeEach(func() {
				request = newTestRequest("{{")
			})

			It("returns an error but does not keep the container", func() {
				response := models.EvacuationResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())
				Expect(response.KeepContainer).To(BeFalse())
				Expect(response.Error).To(Equal(models.ErrBadRequest))
			})
		})
	})
})
