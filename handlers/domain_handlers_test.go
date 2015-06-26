package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"

	"github.com/cloudfoundry-incubator/bbs"
	"github.com/cloudfoundry-incubator/bbs/db/fakes"
	"github.com/cloudfoundry-incubator/bbs/handlers"
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
				domainUpserted, ttlUpserted := fakeDomainDB.UpsertDomainArgsForCall(0)
				Expect(domainUpserted).To(Equal(domain))
				Expect(ttlUpserted).To(Equal(ttl))
			})

			It("responds with 204 Status No Content", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusNoContent))
			})
		})

		Context("when the DB errors out", func() {
			BeforeEach(func() {
				fakeDomainDB.UpsertDomainReturns(errors.New("Something went wrong"))
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})

			It("provides relevant error information", func() {
				var bbsError bbs.Error
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &bbsError)
				Expect(err).NotTo(HaveOccurred())

				Expect(bbsError).To(Equal(bbs.Error{
					Type:    bbs.UnknownError,
					Message: "Something went wrong",
				}))
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
				fakeDomainDB.GetAllDomainsReturns(domains, nil)
			})

			It("call the DB to retrieve the domains", func() {
				Expect(fakeDomainDB.GetAllDomainsCallCount()).To(Equal(1))
			})

			It("responds with 200 Status OK", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})

			It("returns a list of domains", func() {
				response := []string{}
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())

				Expect(response).To(ConsistOf(domains))
			})
		})

		Context("when the DB returns no domains", func() {
			BeforeEach(func() {
				fakeDomainDB.GetAllDomainsReturns([]string{}, nil)
			})

			It("responds with 200 Status OK", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})

			It("returns an empty list", func() {
				Expect(responseRecorder.Body.String()).To(Equal("[]"))
			})
		})

		Context("when the DB errors out", func() {
			BeforeEach(func() {
				fakeDomainDB.GetAllDomainsReturns([]string{}, errors.New("Something went wrong"))
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})

			It("provides relevant error information", func() {
				var bbsError bbs.Error
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &bbsError)
				Expect(err).NotTo(HaveOccurred())

				Expect(bbsError).To(Equal(bbs.Error{
					Type:    bbs.UnknownError,
					Message: "Something went wrong",
				}))
			})
		})
	})
})
