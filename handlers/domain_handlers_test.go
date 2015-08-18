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

var _ = Describe("Domain Handlers", func() {
	var (
		logger           lager.Logger
		fakeDomainDB     *fakes.FakeDomainDB
		responseRecorder *httptest.ResponseRecorder
		handler          *handlers.DomainHandler
		requestBody      interface{}
	)

	BeforeEach(func() {
		fakeDomainDB = new(fakes.FakeDomainDB)
		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		responseRecorder = httptest.NewRecorder()
		handler = handlers.NewDomainHandler(logger, fakeDomainDB)
	})

	Describe("Upsert", func() {
		var (
			domain string
			ttl    uint32
		)

		BeforeEach(func() {
			domain = "domain-to-add"
			ttl = 12345

			requestBody = &models.UpsertDomainRequest{
				Domain: domain,
				Ttl:    ttl,
			}
		})

		JustBeforeEach(func() {
			request := newTestRequest(requestBody)
			handler.Upsert(responseRecorder, request)
		})

		Context("when upserting domain to DB succeeds", func() {
			BeforeEach(func() {
				fakeDomainDB.UpsertDomainReturns(nil)
			})

			It("call the DB to upsert the domain", func() {
				Expect(fakeDomainDB.UpsertDomainCallCount()).To(Equal(1))
				_, domainUpserted, ttlUpserted := fakeDomainDB.UpsertDomainArgsForCall(0)
				Expect(domainUpserted).To(Equal(domain))
				Expect(ttlUpserted).To(Equal(ttl))
			})

			It("responds with 200 OK", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})

			It("responds with no error", func() {
				var upsertDomainResponse models.UpsertDomainResponse
				err := upsertDomainResponse.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(upsertDomainResponse.Error).To(BeNil())
			})
		})

		Context("when the request is invalid", func() {
			BeforeEach(func() {
				requestBody = &models.UpsertDomainRequest{}
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))

				var upsertDomainResponse models.UpsertDomainResponse
				err := upsertDomainResponse.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(upsertDomainResponse.Error).NotTo(BeNil())
				Expect(upsertDomainResponse.Error.Type).To(Equal(models.InvalidRequest))
			})
		})

		Context("when parsing the body crashs", func() {
			BeforeEach(func() {
				requestBody = "beep boop beep boop -- i am a robot"
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))

				var upsertDomainResponse models.UpsertDomainResponse
				err := upsertDomainResponse.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(upsertDomainResponse.Error).NotTo(BeNil())
				Expect(upsertDomainResponse.Error).To(Equal(models.ErrBadRequest))
			})
		})

		Context("when the DB errors out", func() {
			BeforeEach(func() {
				fakeDomainDB.UpsertDomainReturns(models.ErrUnknownError)
			})

			It("provides relevant error information", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))

				var upsertDomainResponse models.UpsertDomainResponse
				err := upsertDomainResponse.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(upsertDomainResponse.Error).NotTo(BeNil())
				Expect(upsertDomainResponse.Error).To(Equal(models.ErrUnknownError))
			})
		})
	})

	Describe("GetAll", func() {
		var domains []string

		BeforeEach(func() {
			domains = []string{"domain-a", "domain-b"}
		})

		JustBeforeEach(func() {
			handler.GetAll(responseRecorder, newTestRequest(""))
		})

		Context("when reading domains from DB succeeds", func() {
			BeforeEach(func() {
				fakeDomainDB.GetAllDomainsReturns(&models.Domains{Domains: domains}, nil)
			})

			It("call the DB to retrieve the domains", func() {
				Expect(fakeDomainDB.GetAllDomainsCallCount()).To(Equal(1))
			})

			It("responds with 200 Status OK", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})

			It("returns a list of domains", func() {
				response := models.Domains{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.GetDomains()).To(ConsistOf(domains))
			})
		})

		Context("when the DB returns no domains", func() {
			BeforeEach(func() {
				fakeDomainDB.GetAllDomainsReturns(&models.Domains{}, nil)
			})

			It("responds with 200 Status OK", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})

			It("returns an empty list", func() {
				response := &models.Domains{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response).To(Equal(&models.Domains{}))
			})
		})

		Context("when the DB errors out", func() {
			BeforeEach(func() {
				fakeDomainDB.GetAllDomainsReturns(&models.Domains{}, models.ErrUnknownError)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})

			It("provides relevant error information", func() {
				var bbsError models.Error
				err := bbsError.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(bbsError.Equal(models.ErrUnknownError)).To(BeTrue())
			})
		})
	})
})
