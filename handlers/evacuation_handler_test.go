package handlers_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry-incubator/bbs/db/fakes"
	"github.com/cloudfoundry-incubator/bbs/handlers"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/models/internal/model_helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
)

var _ = Describe("Evacuation Handlers", func() {
	var (
		logger           lager.Logger
		fakeEvacuationDB *fakes.FakeEvacuationDB
		responseRecorder *httptest.ResponseRecorder
		handler          *handlers.EvacuationHandler
	)

	BeforeEach(func() {
		fakeEvacuationDB = new(fakes.FakeEvacuationDB)
		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		responseRecorder = httptest.NewRecorder()
		handler = handlers.NewEvacuationHandler(logger, fakeEvacuationDB)
	})

	Describe("RemoveEvacuatingActualLRP", func() {
		var (
			request     *http.Request
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
			request = newTestRequest(requestBody)
			handler.RemoveEvacuatingActualLRP(responseRecorder, request)
		})

		Context("when removeEvacuatinging the actual lrp in the DB succeeds", func() {
			BeforeEach(func() {
				fakeEvacuationDB.RemoveEvacuatingActualLRPReturns(nil)
			})

			It("responds with 204 No Content", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusNoContent))
			})

			It("removeEvacuatings the actual lrp by process guid and index", func() {
				Expect(fakeEvacuationDB.RemoveEvacuatingActualLRPCallCount()).To(Equal(1))
				_, actualRequest := fakeEvacuationDB.RemoveEvacuatingActualLRPArgsForCall(0)
				Expect(actualRequest).To(Equal(requestBody))
			})
		})

		Context("when the request is invalid", func() {
			BeforeEach(func() {
				requestBody = &models.RemoveEvacuatingActualLRPRequest{}
			})

			It("responds with 400 Bad Request", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("when parsing the body fails", func() {
			BeforeEach(func() {
				requestBody = "beep boop beep boop -- i am a robot"
			})

			It("responds with 400 Bad Request", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("when retiring the actual lrp fails", func() {
			BeforeEach(func() {
				fakeEvacuationDB.RemoveEvacuatingActualLRPReturns(models.ErrUnknownError)
			})

			It("responds with 500 Internal Server Error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})
		})

		Context("when we cannot find the resource", func() {
			BeforeEach(func() {
				fakeEvacuationDB.RemoveEvacuatingActualLRPReturns(models.ErrResourceNotFound)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusNotFound))
			})
		})
	})

	Describe("EvacuateClaimedActualLRP", func() {
		var request *http.Request
		var requestBody *models.EvacuateClaimedActualLRPRequest
		var actual *models.ActualLRP

		BeforeEach(func() {
			actual = model_helpers.NewValidActualLRP("process-guid", 1)
			requestBody = &models.EvacuateClaimedActualLRPRequest{
				ActualLrpKey:         &actual.ActualLRPKey,
				ActualLrpInstanceKey: &actual.ActualLRPInstanceKey,
			}

			request = newTestRequest(requestBody)
		})

		JustBeforeEach(func() {
			handler.EvacuateClaimedActualLRP(responseRecorder, request)
		})

		It("sends the request to the db", func() {
			Expect(fakeEvacuationDB.EvacuateClaimedActualLRPCallCount()).To(Equal(1))
			_, req := fakeEvacuationDB.EvacuateClaimedActualLRPArgsForCall(0)
			Expect(*req.ActualLrpKey).To(Equal(actual.ActualLRPKey))
			Expect(*req.ActualLrpInstanceKey).To(Equal(actual.ActualLRPInstanceKey))
		})

		Context("when the db returns an error", func() {
			BeforeEach(func() {
				fakeEvacuationDB.EvacuateClaimedActualLRPReturns(false, models.ErrUnknownError)
			})

			It("writes the error in the http response", func() {
				response := models.EvacuationResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())
				Expect(response.KeepContainer).To(BeFalse())
				Expect(response.Error).NotTo(BeNil())
				Expect(response.Error).To(Equal(models.ErrUnknownError))
			})

			It("responds with 200 OK", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})
		})

		Context("when the db returns keepContianer", func() {
			BeforeEach(func() {
				fakeEvacuationDB.EvacuateClaimedActualLRPReturns(true, nil)
			})

			It("writes the correct http response", func() {
				response := models.EvacuationResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())
				Expect(response.KeepContainer).To(BeTrue())
				Expect(response.Error).To(BeNil())
			})

			It("responds with 200 OK", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})
		})
	})

	Describe("EvacuateCrashedActualLRP", func() {
		var request *http.Request
		var actual *models.ActualLRP
		var expectedErrorMessage string

		BeforeEach(func() {
			expectedErrorMessage = "error-message"

			actual = model_helpers.NewValidActualLRP("process-guid", 1)
			requestBody := &models.EvacuateCrashedActualLRPRequest{
				ActualLrpKey:         &actual.ActualLRPKey,
				ActualLrpInstanceKey: &actual.ActualLRPInstanceKey,
				ErrorMessage:         expectedErrorMessage,
			}

			request = newTestRequest(requestBody)
		})

		JustBeforeEach(func() {
			handler.EvacuateCrashedActualLRP(responseRecorder, request)
		})

		It("sends the request to the db", func() {
			Expect(fakeEvacuationDB.EvacuateCrashedActualLRPCallCount()).To(Equal(1))
			_, req := fakeEvacuationDB.EvacuateCrashedActualLRPArgsForCall(0)
			Expect(*req.ActualLrpKey).To(Equal(actual.ActualLRPKey))
			Expect(*req.ActualLrpInstanceKey).To(Equal(actual.ActualLRPInstanceKey))
			Expect(req.ErrorMessage).To(Equal(expectedErrorMessage))
		})

		Context("when the db returns an error", func() {
			BeforeEach(func() {
				fakeEvacuationDB.EvacuateCrashedActualLRPReturns(false, models.ErrUnknownError)
			})

			It("writes the error in the http response", func() {
				response := models.EvacuationResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())
				Expect(response.KeepContainer).To(BeFalse())
				Expect(response.Error).NotTo(BeNil())
				Expect(response.Error).To(Equal(models.ErrUnknownError))
			})

			It("responds with 200 OK", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})
		})

		Context("when the db returns keepContianer", func() {
			BeforeEach(func() {
				fakeEvacuationDB.EvacuateCrashedActualLRPReturns(true, nil)
			})

			It("writes the correct http response", func() {
				response := models.EvacuationResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())
				Expect(response.KeepContainer).To(BeTrue())
				Expect(response.Error).To(BeNil())
			})

			It("responds with 200 OK", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})
		})
	})

	Describe("EvacuateRunningActualLRP", func() {
		var request *http.Request
		var requestBody *models.EvacuateRunningActualLRPRequest
		var actual *models.ActualLRP
		const expectedTTL = 123

		BeforeEach(func() {
			actual = model_helpers.NewValidActualLRP("process-guid", 1)
			requestBody = &models.EvacuateRunningActualLRPRequest{
				ActualLrpKey:         &actual.ActualLRPKey,
				ActualLrpInstanceKey: &actual.ActualLRPInstanceKey,
				ActualLrpNetInfo:     &actual.ActualLRPNetInfo,
				Ttl:                  expectedTTL,
			}

			request = newTestRequest(requestBody)
		})

		JustBeforeEach(func() {
			handler.EvacuateRunningActualLRP(responseRecorder, request)
		})

		It("sends the request to the db", func() {
			Expect(fakeEvacuationDB.EvacuateRunningActualLRPCallCount()).To(Equal(1))
			_, req := fakeEvacuationDB.EvacuateRunningActualLRPArgsForCall(0)
			Expect(*req.ActualLrpKey).To(Equal(actual.ActualLRPKey))
			Expect(*req.ActualLrpInstanceKey).To(Equal(actual.ActualLRPInstanceKey))
			Expect(*req.ActualLrpNetInfo).To(Equal(actual.ActualLRPNetInfo))
		})

		Context("when the db returns an error", func() {
			BeforeEach(func() {
				fakeEvacuationDB.EvacuateRunningActualLRPReturns(false, models.ErrUnknownError)
			})

			It("writes the error in the http response", func() {
				response := models.EvacuationResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())
				Expect(response.KeepContainer).To(BeFalse())
				Expect(response.Error).NotTo(BeNil())
				Expect(response.Error).To(Equal(models.ErrUnknownError))
			})

			It("responds with 200 OK", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})
		})

		Context("when the db returns keepContianer", func() {
			BeforeEach(func() {
				fakeEvacuationDB.EvacuateRunningActualLRPReturns(true, nil)
			})

			It("writes the correct http response", func() {
				response := models.EvacuationResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())
				Expect(response.KeepContainer).To(BeTrue())
				Expect(response.Error).To(BeNil())
			})

			It("responds with 200 OK", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})
		})
	})

	Describe("EvacuateStoppedActualLRP", func() {
		var request *http.Request
		var actual *models.ActualLRP

		BeforeEach(func() {
			actual = model_helpers.NewValidActualLRP("process-guid", 1)
			requestBody := &models.EvacuateStoppedActualLRPRequest{
				ActualLrpKey:         &actual.ActualLRPKey,
				ActualLrpInstanceKey: &actual.ActualLRPInstanceKey,
			}

			request = newTestRequest(requestBody)
		})

		JustBeforeEach(func() {
			handler.EvacuateStoppedActualLRP(responseRecorder, request)
		})

		It("sends the request to the db", func() {
			Expect(fakeEvacuationDB.EvacuateStoppedActualLRPCallCount()).To(Equal(1))
			_, req := fakeEvacuationDB.EvacuateStoppedActualLRPArgsForCall(0)
			Expect(*req.ActualLrpKey).To(Equal(actual.ActualLRPKey))
			Expect(*req.ActualLrpInstanceKey).To(Equal(actual.ActualLRPInstanceKey))
		})

		Context("when the db returns an error", func() {
			BeforeEach(func() {
				fakeEvacuationDB.EvacuateStoppedActualLRPReturns(false, models.ErrUnknownError)
			})

			It("writes the error in the http response", func() {
				response := models.EvacuationResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())
				Expect(response.KeepContainer).To(BeFalse())
				Expect(response.Error).NotTo(BeNil())
				Expect(response.Error).To(Equal(models.ErrUnknownError))
			})

			It("responds with 200 OK", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})
		})

		Context("when the db returns keepContianer", func() {
			BeforeEach(func() {
				fakeEvacuationDB.EvacuateStoppedActualLRPReturns(true, nil)
			})

			It("writes the correct http response", func() {
				response := models.EvacuationResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())
				Expect(response.KeepContainer).To(BeTrue())
				Expect(response.Error).To(BeNil())
			})

			It("responds with 200 OK", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})
		})
	})
})
