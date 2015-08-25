package handlers_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry-incubator/bbs/db/fakes"
	"github.com/cloudfoundry-incubator/bbs/handlers"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/models/test/model_helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
)

var _ = Describe("DesiredLRP Handlers", func() {
	var (
		logger           lager.Logger
		fakeDesiredLRPDB *fakes.FakeDesiredLRPDB
		responseRecorder *httptest.ResponseRecorder
		handler          *handlers.DesiredLRPHandler

		desiredLRP1 models.DesiredLRP
		desiredLRP2 models.DesiredLRP
	)

	BeforeEach(func() {
		fakeDesiredLRPDB = new(fakes.FakeDesiredLRPDB)
		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		responseRecorder = httptest.NewRecorder()
		handler = handlers.NewDesiredLRPHandler(logger, fakeDesiredLRPDB)
	})

	Describe("DesiredLRPs", func() {
		var requestBody interface{}

		BeforeEach(func() {
			requestBody = &models.DesiredLRPsRequest{}
			desiredLRP1 = models.DesiredLRP{}
			desiredLRP2 = models.DesiredLRP{}
		})

		JustBeforeEach(func() {
			request := newTestRequest(requestBody)
			handler.DesiredLRPs(responseRecorder, request)
		})

		Context("when reading desired lrps from DB succeeds", func() {
			var desiredLRPs []*models.DesiredLRP

			BeforeEach(func() {
				desiredLRPs = []*models.DesiredLRP{&desiredLRP1, &desiredLRP2}
				fakeDesiredLRPDB.DesiredLRPsReturns(desiredLRPs, nil)
			})

			It("returns a list of desired lrp groups", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := models.DesiredLRPsResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(BeNil())
				Expect(response.DesiredLrps).To(Equal(desiredLRPs))
			})

			Context("and no filter is provided", func() {
				It("call the DB with no filters to retrieve the desired lrps", func() {
					Expect(fakeDesiredLRPDB.DesiredLRPsCallCount()).To(Equal(1))
					_, filter := fakeDesiredLRPDB.DesiredLRPsArgsForCall(0)
					Expect(filter).To(Equal(models.DesiredLRPFilter{}))
				})
			})

			Context("and filtering by domain", func() {
				BeforeEach(func() {
					requestBody = &models.DesiredLRPsRequest{Domain: "domain-1"}
				})

				It("call the DB with the domain filter to retrieve the desired lrps", func() {
					Expect(fakeDesiredLRPDB.DesiredLRPsCallCount()).To(Equal(1))
					_, filter := fakeDesiredLRPDB.DesiredLRPsArgsForCall(0)
					Expect(filter.Domain).To(Equal("domain-1"))
				})
			})
		})

		Context("when the DB returns no desired lrp groups", func() {
			BeforeEach(func() {
				fakeDesiredLRPDB.DesiredLRPsReturns([]*models.DesiredLRP{}, nil)
			})

			It("returns an empty list", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := models.DesiredLRPsResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(BeNil())
				Expect(response.DesiredLrps).To(BeEmpty())
			})
		})

		Context("when the DB errors out", func() {
			BeforeEach(func() {
				fakeDesiredLRPDB.DesiredLRPsReturns([]*models.DesiredLRP{}, models.ErrUnknownError)
			})

			It("provides relevant error information", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := models.DesiredLRPsResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrUnknownError))
			})
		})
	})

	Describe("DesiredLRPByProcessGuid", func() {
		var (
			processGuid = "process-guid"

			requestBody interface{}
		)

		BeforeEach(func() {
			requestBody = &models.DesiredLRPByProcessGuidRequest{
				ProcessGuid: processGuid,
			}
		})

		JustBeforeEach(func() {
			request := newTestRequest(requestBody)
			handler.DesiredLRPByProcessGuid(responseRecorder, request)
		})

		Context("when reading desired lrp from DB succeeds", func() {
			var desiredLRP *models.DesiredLRP

			BeforeEach(func() {
				desiredLRP = &models.DesiredLRP{ProcessGuid: processGuid}
				fakeDesiredLRPDB.DesiredLRPByProcessGuidReturns(desiredLRP, nil)
			})

			It("fetches desired lrp by process guid", func() {
				Expect(fakeDesiredLRPDB.DesiredLRPByProcessGuidCallCount()).To(Equal(1))
				_, actualProcessGuid := fakeDesiredLRPDB.DesiredLRPByProcessGuidArgsForCall(0)
				Expect(actualProcessGuid).To(Equal(processGuid))

				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := models.DesiredLRPResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(BeNil())
				Expect(response.DesiredLrp).To(Equal(desiredLRP))
			})
		})

		Context("when the DB returns no desired lrp", func() {
			BeforeEach(func() {
				fakeDesiredLRPDB.DesiredLRPByProcessGuidReturns(nil, models.ErrResourceNotFound)
			})

			It("returns a resource not found error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := models.DesiredLRPResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrResourceNotFound))
			})
		})

		Context("when the DB errors out", func() {
			BeforeEach(func() {
				fakeDesiredLRPDB.DesiredLRPByProcessGuidReturns(nil, models.ErrUnknownError)
			})

			It("provides relevant error information", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := models.DesiredLRPResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrUnknownError))
			})
		})
	})

	Describe("DesireDesiredLRP", func() {
		var (
			desiredLRP *models.DesiredLRP

			requestBody interface{}
		)

		BeforeEach(func() {
			desiredLRP = model_helpers.NewValidDesiredLRP("some-guid")
			requestBody = &models.DesireLRPRequest{
				DesiredLrp: desiredLRP,
			}
		})

		JustBeforeEach(func() {
			request := newTestRequest(requestBody)
			handler.DesireDesiredLRP(responseRecorder, request)
		})

		Context("when creating desired lrp in DB succeeds", func() {
			BeforeEach(func() {
				fakeDesiredLRPDB.DesireLRPReturns(nil)
			})

			It("creates desired lrp", func() {
				Expect(fakeDesiredLRPDB.DesireLRPCallCount()).To(Equal(1))
				_, actualDesiredLRP := fakeDesiredLRPDB.DesireLRPArgsForCall(0)
				Expect(actualDesiredLRP).To(Equal(desiredLRP))

				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := models.DesiredLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(BeNil())
			})
		})

		Context("when the DB errors out", func() {
			BeforeEach(func() {
				fakeDesiredLRPDB.DesireLRPReturns(models.ErrUnknownError)
			})

			It("provides relevant error information", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := models.DesiredLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrUnknownError))
			})
		})
	})

	Describe("UpdateDesiredLRP", func() {
		var (
			processGuid string
			update      *models.DesiredLRPUpdate

			requestBody interface{}
		)

		BeforeEach(func() {
			processGuid = "some-guid"
			someText := "some-text"
			update = &models.DesiredLRPUpdate{
				Annotation: &someText,
			}
			requestBody = &models.UpdateDesiredLRPRequest{
				ProcessGuid: processGuid,
				Update:      update,
			}
		})

		JustBeforeEach(func() {
			request := newTestRequest(requestBody)
			handler.UpdateDesiredLRP(responseRecorder, request)
		})

		Context("when updating desired lrp in DB succeeds", func() {
			BeforeEach(func() {
				fakeDesiredLRPDB.UpdateDesiredLRPReturns(nil)
			})

			It("updates the desired lrp", func() {
				Expect(fakeDesiredLRPDB.UpdateDesiredLRPCallCount()).To(Equal(1))
				_, actualProcessGuid, actualUpdate := fakeDesiredLRPDB.UpdateDesiredLRPArgsForCall(0)
				Expect(actualProcessGuid).To(Equal(processGuid))
				Expect(actualUpdate).To(Equal(update))

				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := models.DesiredLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(BeNil())
			})
		})

		Context("when the DB errors out", func() {
			BeforeEach(func() {
				fakeDesiredLRPDB.UpdateDesiredLRPReturns(models.ErrUnknownError)
			})

			It("provides relevant error information", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := models.DesiredLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrUnknownError))
			})
		})
	})

	Describe("RemoveDesiredLRP", func() {
		var (
			processGuid string

			requestBody interface{}
		)

		BeforeEach(func() {
			processGuid = "some-guid"
			requestBody = &models.RemoveDesiredLRPRequest{
				ProcessGuid: processGuid,
			}
		})

		JustBeforeEach(func() {
			request := newTestRequest(requestBody)
			handler.RemoveDesiredLRP(responseRecorder, request)
		})

		Context("when updating desired lrp in DB succeeds", func() {
			BeforeEach(func() {
				fakeDesiredLRPDB.RemoveDesiredLRPReturns(nil)
			})

			It("updates the desired lrp", func() {
				Expect(fakeDesiredLRPDB.RemoveDesiredLRPCallCount()).To(Equal(1))
				_, actualProcessGuid := fakeDesiredLRPDB.RemoveDesiredLRPArgsForCall(0)
				Expect(actualProcessGuid).To(Equal(processGuid))

				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := models.DesiredLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(BeNil())
			})
		})

		Context("when the DB errors out", func() {
			BeforeEach(func() {
				fakeDesiredLRPDB.RemoveDesiredLRPReturns(models.ErrUnknownError)
			})

			It("provides relevant error information", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := models.DesiredLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrUnknownError))
			})
		})
	})
})
