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

var _ = Describe("Domain Handlers", func() {
	var (
		logger           lager.Logger
		fakeDomainDB     *fakes.FakeDomainDB
		responseRecorder *httptest.ResponseRecorder
		handler          *handlers.DomainHandler
	)

	BeforeEach(func() {
		fakeDomainDB = new(fakes.FakeDomainDB)
		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		responseRecorder = httptest.NewRecorder()
		handler = handlers.NewDomainHandler(fakeDomainDB, logger)
	})

	Describe("Upsert", func() {
		var domain string
		var ttl int

		BeforeEach(func() {
			domain = "domain-to-add"
			ttl = 12345
		})

		JustBeforeEach(func() {
			req := newTestRequest("")
			req.URL.RawQuery = url.Values{":domain": []string{domain}}.Encode()
			req.Header["Cache-Control"] = []string{"public", "max-age=" + strconv.Itoa(ttl)}
			handler.Upsert(responseRecorder, req)
		})

		Context("when upserting domain to DB succeeds", func() {
			BeforeEach(func() {
				fakeDomainDB.UpsertDomainReturns(nil)
			})

			It("call the DB to upsert the domain", func() {
				Expect(fakeDomainDB.UpsertDomainCallCount()).To(Equal(1))
				domainUpserted, ttlUpserted, _ := fakeDomainDB.UpsertDomainArgsForCall(0)
				Expect(domainUpserted).To(Equal(domain))
				Expect(ttlUpserted).To(Equal(ttl))
			})

			It("responds with 204 Status No Content", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusNoContent))
			})
		})

		Context("when the DB errors out", func() {
			BeforeEach(func() {
				fakeDomainDB.UpsertDomainReturns(models.ErrUnknownError)
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
