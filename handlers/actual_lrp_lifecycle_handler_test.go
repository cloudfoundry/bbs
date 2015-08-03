package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"

	"github.com/cloudfoundry-incubator/bbs/db/fakes"
	"github.com/cloudfoundry-incubator/bbs/handlers"
	"github.com/cloudfoundry-incubator/bbs/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
)

var _ = Describe("ActualLRP Lifecycle Handlers", func() {
	var (
		logger           lager.Logger
		fakeActualLRPDB  *fakes.FakeActualLRPDB
		responseRecorder *httptest.ResponseRecorder
		handler          *handlers.ActualLRPLifecycleHandler

		actualLRP models.ActualLRP
	)

	BeforeEach(func() {
		fakeActualLRPDB = new(fakes.FakeActualLRPDB)
		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		responseRecorder = httptest.NewRecorder()
		handler = handlers.NewActualLRPLifecycleHandler(logger, fakeActualLRPDB)
	})

	Describe("ClaimActualLRP", func() {
		var (
			request     *http.Request
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
		})

		JustBeforeEach(func() {
			request = newTestRequest(requestBody)
			handler.ClaimActualLRP(responseRecorder, request)
		})

		Context("when claiming the actual lrp in the DB succeeds", func() {
			var claimedActualLRP models.ActualLRP

			BeforeEach(func() {
				claimedActualLRP = actualLRP
				claimedActualLRP.ActualLRPInstanceKey = instanceKey
				fakeActualLRPDB.ClaimActualLRPReturns(&claimedActualLRP, nil)
			})

			It("responds with 200 Status OK", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})

			It("claims the actual lrp by process guid and index", func() {
				Expect(fakeActualLRPDB.ClaimActualLRPCallCount()).To(Equal(1))
				_, actualProcessGuid, idx, actualInstanceKey := fakeActualLRPDB.ClaimActualLRPArgsForCall(0)
				Expect(actualProcessGuid).To(Equal(processGuid))
				Expect(idx).To(BeEquivalentTo(index))
				Expect(*actualInstanceKey).To(Equal(instanceKey))
			})

			It("returns the claimed actual lrp", func() {
				response := &models.ActualLRP{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(*response).To(Equal(claimedActualLRP))
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

		Context("when claiming the actual lrp fails", func() {
			BeforeEach(func() {
				fakeActualLRPDB.ClaimActualLRPReturns(nil, models.ErrUnknownError)
			})

			It("responds with 500 Internal Server Error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})
		})

		Context("when we cannot find the resource", func() {
			BeforeEach(func() {
				fakeActualLRPDB.ClaimActualLRPReturns(nil, models.ErrResourceNotFound)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusNotFound))
			})
		})
	})

	Describe("StartActualLRP", func() {
		var (
			request     *http.Request
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
		})

		JustBeforeEach(func() {
			request = newTestRequest(requestBody)
			handler.StartActualLRP(responseRecorder, request)
		})

		Context("when starting the actual lrp in the DB succeeds", func() {
			var startedActualLRP models.ActualLRP

			BeforeEach(func() {
				fakeActualLRPDB.StartActualLRPReturns(&startedActualLRP, nil)
			})

			It("responds with 200 Status OK", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})

			It("starts the actual lrp by process guid and index", func() {
				Expect(fakeActualLRPDB.StartActualLRPCallCount()).To(Equal(1))
				_, actualRequest := fakeActualLRPDB.StartActualLRPArgsForCall(0)
				Expect(actualRequest).To(Equal(requestBody))
			})

			It("returns the started actual lrp", func() {
				response := &models.ActualLRP{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(*response).To(Equal(startedActualLRP))
			})
		})

		Context("when the request is invalid", func() {
			BeforeEach(func() {
				requestBody = &models.StartActualLRPRequest{}
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

		Context("when starting the actual lrp fails", func() {
			BeforeEach(func() {
				fakeActualLRPDB.StartActualLRPReturns(nil, models.ErrUnknownError)
			})

			It("responds with 500 Internal Server Error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})
		})

		Context("when we cannot find the resource", func() {
			BeforeEach(func() {
				fakeActualLRPDB.StartActualLRPReturns(nil, models.ErrResourceNotFound)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusNotFound))
			})
		})
	})

	Describe("CrashActualLRP", func() {
		var (
			request      *http.Request
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
				ActualLRPKey: key,
				State:        models.ActualLRPStateUnclaimed,
				Since:        1138,
			}
		})

		JustBeforeEach(func() {
			request = newTestRequest(requestBody)
			handler.CrashActualLRP(responseRecorder, request)
		})

		Context("when crashing the actual lrp in the DB succeeds", func() {
			BeforeEach(func() {
				fakeActualLRPDB.CrashActualLRPReturns(nil)
			})

			It("responds with 204 No Content", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusNoContent))
			})

			It("crashs the actual lrp by process guid and index", func() {
				Expect(fakeActualLRPDB.CrashActualLRPCallCount()).To(Equal(1))
				_, actualRequest := fakeActualLRPDB.CrashActualLRPArgsForCall(0)
				Expect(actualRequest).To(Equal(requestBody))
			})
		})

		Context("when the request is invalid", func() {
			BeforeEach(func() {
				requestBody = &models.CrashActualLRPRequest{}
			})

			It("responds with 400 Bad Request", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("when parsing the body crashs", func() {
			BeforeEach(func() {
				requestBody = "beep boop beep boop -- i am a robot"
			})

			It("responds with 400 Bad Request", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("when crashing the actual lrp crashs", func() {
			BeforeEach(func() {
				fakeActualLRPDB.CrashActualLRPReturns(models.ErrUnknownError)
			})

			It("responds with 500 Internal Server Error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})
		})

		Context("when we cannot find the resource", func() {
			BeforeEach(func() {
				fakeActualLRPDB.CrashActualLRPReturns(models.ErrResourceNotFound)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusNotFound))
			})
		})
	})

	Describe("RetireActualLRP", func() {
		var (
			request     *http.Request
			processGuid = "process-guid"
			index       = int32(1)

			key models.ActualLRPKey

			requestBody interface{}
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
		})

		JustBeforeEach(func() {
			request = newTestRequest(requestBody)
			handler.RetireActualLRP(responseRecorder, request)
		})

		Context("when retireing the actual lrp in the DB succeeds", func() {
			BeforeEach(func() {
				fakeActualLRPDB.RetireActualLRPReturns(nil)
			})

			It("responds with 204 No Content", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusNoContent))
			})

			It("retires the actual lrp by process guid and index", func() {
				Expect(fakeActualLRPDB.RetireActualLRPCallCount()).To(Equal(1))
				_, actualRequest := fakeActualLRPDB.RetireActualLRPArgsForCall(0)
				Expect(actualRequest).To(Equal(requestBody))
			})
		})

		Context("when the request is invalid", func() {
			BeforeEach(func() {
				requestBody = &models.RetireActualLRPRequest{}
			})

			It("responds with 400 Bad Request", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("when parsing the body retires", func() {
			BeforeEach(func() {
				requestBody = "beep boop beep boop -- i am a robot"
			})

			It("responds with 400 Bad Request", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("when retiring the actual lrp fails", func() {
			BeforeEach(func() {
				fakeActualLRPDB.RetireActualLRPReturns(models.ErrUnknownError)
			})

			It("responds with 500 Internal Server Error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})
		})

		Context("when we cannot find the resource", func() {
			BeforeEach(func() {
				fakeActualLRPDB.RetireActualLRPReturns(models.ErrResourceNotFound)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusNotFound))
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
		})

		JustBeforeEach(func() {
			request = newTestRequest(requestBody)
			handler.FailActualLRP(responseRecorder, request)
		})

		Context("when failing the actual lrp in the DB succeeds", func() {
			BeforeEach(func() {
				fakeActualLRPDB.FailActualLRPReturns(nil)
			})

			It("responds with 204 No Content", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusNoContent))
			})

			It("fails the actual lrp by process guid and index", func() {
				Expect(fakeActualLRPDB.FailActualLRPCallCount()).To(Equal(1))
				_, actualRequest := fakeActualLRPDB.FailActualLRPArgsForCall(0)
				Expect(actualRequest).To(Equal(requestBody))
			})
		})

		Context("when the request is invalid", func() {
			BeforeEach(func() {
				requestBody = &models.FailActualLRPRequest{}
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

		Context("when failing the actual lrp fails", func() {
			BeforeEach(func() {
				fakeActualLRPDB.FailActualLRPReturns(models.ErrUnknownError)
			})

			It("responds with 500 Internal Server Error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})
		})

		Context("when we cannot find the resource", func() {
			BeforeEach(func() {
				fakeActualLRPDB.FailActualLRPReturns(models.ErrResourceNotFound)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusNotFound))
			})
		})
	})

	Describe("RemoveActualLRP", func() {
		var (
			request     *http.Request
			processGuid = "process-guid"
			index       = 1
			instanceKey models.ActualLRPInstanceKey
			indexParam  string
		)

		BeforeEach(func() {
			indexParam = strconv.Itoa(index)
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
		})

		JustBeforeEach(func() {
			request = newTestRequest("")
			request.URL.RawQuery = url.Values{
				":process_guid": []string{processGuid},
				":index":        []string{indexParam},
			}.Encode()

			handler.RemoveActualLRP(responseRecorder, request)
		})

		Context("when removing the actual lrp in the DB succeeds", func() {
			var removedActualLRP models.ActualLRP

			BeforeEach(func() {
				removedActualLRP = actualLRP
				removedActualLRP.ActualLRPInstanceKey = instanceKey
				fakeActualLRPDB.RemoveActualLRPReturns(nil)
			})

			It("responds with 204 Status No Content", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusNoContent))
			})

			It("removes the actual lrp by process guid and index", func() {
				Expect(fakeActualLRPDB.RemoveActualLRPCallCount()).To(Equal(1))
				_, actualProcessGuid, idx := fakeActualLRPDB.RemoveActualLRPArgsForCall(0)
				Expect(actualProcessGuid).To(Equal(processGuid))
				Expect(idx).To(BeEquivalentTo(index))
			})
		})

		Context("when parsing the index fails", func() {
			BeforeEach(func() {
				indexParam = "this is not an index?"
			})

			It("responds with 400 Bad Request", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("when removing the actual lrp fails", func() {
			BeforeEach(func() {
				fakeActualLRPDB.RemoveActualLRPReturns(models.ErrUnknownError)
			})

			It("responds with 500 Internal Server Error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})
		})

		Context("when we cannot find the resource", func() {
			BeforeEach(func() {
				fakeActualLRPDB.RemoveActualLRPReturns(models.ErrResourceNotFound)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusNotFound))
			})
		})
	})
})
