package handlers_test

import (
	"net/http"
	"net/http/httptest"

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
			request := newTestRequest(requestBody)
			handler.ClaimActualLRP(responseRecorder, request)
		})

		Context("when claiming the actual lrp in the DB succeeds", func() {
			BeforeEach(func() {
				fakeActualLRPDB.ClaimActualLRPReturns(nil)
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
		})

		Context("when claiming the actual lrp fails", func() {
			BeforeEach(func() {
				fakeActualLRPDB.ClaimActualLRPReturns(models.ErrUnknownError)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := &models.ActualLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrUnknownError))
			})
		})

		Context("when we cannot find the resource", func() {
			BeforeEach(func() {
				fakeActualLRPDB.ClaimActualLRPReturns(models.ErrResourceNotFound)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := &models.ActualLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrResourceNotFound))
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
		})

		JustBeforeEach(func() {
			request := newTestRequest(requestBody)
			handler.StartActualLRP(responseRecorder, request)
		})

		Context("when starting the actual lrp in the DB succeeds", func() {
			BeforeEach(func() {
				fakeActualLRPDB.StartActualLRPReturns(nil)
			})

			It("response with no error", func() {
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
		})

		Context("when starting the actual lrp fails", func() {
			BeforeEach(func() {
				fakeActualLRPDB.StartActualLRPReturns(models.ErrUnknownError)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := &models.ActualLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrUnknownError))
			})
		})

		Context("when we cannot find the resource", func() {
			BeforeEach(func() {
				fakeActualLRPDB.StartActualLRPReturns(models.ErrResourceNotFound)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := &models.ActualLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrResourceNotFound))
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
				ActualLRPKey: key,
				State:        models.ActualLRPStateUnclaimed,
				Since:        1138,
			}
		})

		JustBeforeEach(func() {
			request := newTestRequest(requestBody)
			handler.CrashActualLRP(responseRecorder, request)
		})

		Context("when crashing the actual lrp in the DB succeeds", func() {
			BeforeEach(func() {
				fakeActualLRPDB.CrashActualLRPReturns(nil)
			})

			It("response with no error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := &models.ActualLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(BeNil())
			})

			It("crashs the actual lrp by process guid and index", func() {
				Expect(fakeActualLRPDB.CrashActualLRPCallCount()).To(Equal(1))
				_, actualKey, actualInstanceKey, actualErrorMessage := fakeActualLRPDB.CrashActualLRPArgsForCall(0)
				Expect(*actualKey).To(Equal(key))
				Expect(*actualInstanceKey).To(Equal(instanceKey))
				Expect(actualErrorMessage).To(Equal(errorMessage))
			})
		})

		Context("when crashing the actual lrp fails", func() {
			BeforeEach(func() {
				fakeActualLRPDB.CrashActualLRPReturns(models.ErrUnknownError)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := &models.ActualLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrUnknownError))
			})
		})

		Context("when we cannot find the resource", func() {
			BeforeEach(func() {
				fakeActualLRPDB.CrashActualLRPReturns(models.ErrResourceNotFound)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := &models.ActualLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrResourceNotFound))
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

			It("retires the actual lrp by process guid and index", func() {
				Expect(fakeActualLRPDB.RetireActualLRPCallCount()).To(Equal(1))
				_, actualKey := fakeActualLRPDB.RetireActualLRPArgsForCall(0)
				Expect(*actualKey).To(Equal(key))

				Expect(responseRecorder.Code).To(Equal(http.StatusOK))

				response := &models.ActualLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(BeNil())
			})
		})

		Context("when retiring the actual lrp fails", func() {
			BeforeEach(func() {
				fakeActualLRPDB.RetireActualLRPReturns(models.ErrUnknownError)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := &models.ActualLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrUnknownError))
			})
		})

		Context("when we cannot find the resource", func() {
			BeforeEach(func() {
				fakeActualLRPDB.RetireActualLRPReturns(models.ErrResourceNotFound)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := &models.ActualLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrResourceNotFound))
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
		})

		Context("when failing the actual lrp fails", func() {
			BeforeEach(func() {
				fakeActualLRPDB.FailActualLRPReturns(models.ErrUnknownError)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := &models.ActualLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrUnknownError))
			})
		})

		Context("when we cannot find the resource", func() {
			BeforeEach(func() {
				fakeActualLRPDB.FailActualLRPReturns(models.ErrResourceNotFound)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := &models.ActualLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrResourceNotFound))
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
			requestBody = &models.RemoveActualLRPRequest{
				ProcessGuid: processGuid,
				Index:       index,
			}

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
			request := newTestRequest(requestBody)
			handler.RemoveActualLRP(responseRecorder, request)
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
				_, actualProcessGuid, idx := fakeActualLRPDB.RemoveActualLRPArgsForCall(0)
				Expect(actualProcessGuid).To(Equal(processGuid))
				Expect(idx).To(BeEquivalentTo(index))
			})

			It("response with no error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := &models.ActualLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(BeNil())
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
		})
	})
})
