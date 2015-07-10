package handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry-incubator/bbs"
	"github.com/cloudfoundry-incubator/bbs/db/fakes"
	"github.com/cloudfoundry-incubator/bbs/handlers"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/gogo/protobuf/proto"
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
		handler = handlers.NewDesiredLRPHandler(fakeDesiredLRPDB, logger)
	})

	Describe("DesiredLRPs", func() {
		var request *http.Request

		BeforeEach(func() {
			request = newTestRequest("")

			desiredLRP1 = models.DesiredLRP{}
			desiredLRP2 = models.DesiredLRP{}
		})

		JustBeforeEach(func() {
			handler.DesiredLRPs(responseRecorder, request)
		})

		Context("when reading desired lrps from DB succeeds", func() {
			var desiredLRPs *models.DesiredLRPs

			BeforeEach(func() {
				desiredLRPs = &models.DesiredLRPs{
					[]*models.DesiredLRP{&desiredLRP1, &desiredLRP2},
				}
				fakeDesiredLRPDB.DesiredLRPsReturns(desiredLRPs, nil)
			})

			It("responds with 200 Status OK", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})

			It("returns a list of desired lrp groups", func() {
				response := &models.DesiredLRPs{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response).To(Equal(desiredLRPs))
			})

			Context("and no filter is provided", func() {
				It("call the DB with no filters to retrieve the desired lrps", func() {
					Expect(fakeDesiredLRPDB.DesiredLRPsCallCount()).To(Equal(1))
					filter, _ := fakeDesiredLRPDB.DesiredLRPsArgsForCall(0)
					Expect(filter).To(Equal(models.DesiredLRPFilter{}))
				})
			})

			Context("and filtering by domain", func() {
				BeforeEach(func() {
					var err error
					request, err = http.NewRequest("", "http://example.com?domain=domain-1", nil)
					Expect(err).NotTo(HaveOccurred())
				})

				It("call the DB with the domain filter to retrieve the desired lrps", func() {
					Expect(fakeDesiredLRPDB.DesiredLRPsCallCount()).To(Equal(1))
					filter, _ := fakeDesiredLRPDB.DesiredLRPsArgsForCall(0)
					Expect(filter.Domain).To(Equal("domain-1"))
				})
			})
		})

		Context("when the DB returns no desired lrp groups", func() {
			BeforeEach(func() {
				fakeDesiredLRPDB.DesiredLRPsReturns(&models.DesiredLRPs{}, nil)
			})

			It("responds with 200 Status OK", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})

			It("returns an empty list", func() {
				response := &models.DesiredLRPs{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response).To(Equal(&models.DesiredLRPs{}))
			})
		})

		Context("when the DB errors out", func() {
			BeforeEach(func() {
				fakeDesiredLRPDB.DesiredLRPsReturns(&models.DesiredLRPs{}, errors.New("Something went wrong"))
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})

			It("provides relevant error information", func() {
				var bbsError bbs.Error
				err := bbsError.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(bbsError).To(Equal(bbs.Error{
					Type:    proto.String(bbs.UnknownError),
					Message: proto.String("Something went wrong"),
				}))
			})
		})
	})
})
