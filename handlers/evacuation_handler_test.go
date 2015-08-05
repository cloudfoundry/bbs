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
})
