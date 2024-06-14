package handlers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"code.cloudfoundry.org/bbs/db/dbfakes"
	"code.cloudfoundry.org/bbs/handlers"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager/v3"
	"code.cloudfoundry.org/lager/v3/lagertest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"google.golang.org/protobuf/proto"
)

var _ = Describe("Domain Handlers", func() {
	var (
		logger           *lagertest.TestLogger
		fakeDomainDB     *dbfakes.FakeDomainDB
		responseRecorder *httptest.ResponseRecorder
		handler          *handlers.DomainHandler
		requestBody      interface{}
		exitCh           chan struct{}

		requestIdHeader   string
		b3RequestIdHeader string
	)

	BeforeEach(func() {
		fakeDomainDB = new(dbfakes.FakeDomainDB)
		logger = lagertest.NewTestLogger("test")
		responseRecorder = httptest.NewRecorder()
		requestIdHeader = "93f2374a-c0ad-455a-98bc-aafd4e4a1dc4"
		b3RequestIdHeader = fmt.Sprintf(`"trace-id":"%s"`, strings.Replace(requestIdHeader, "-", "", -1))
		exitCh = make(chan struct{}, 1)
		handler = handlers.NewDomainHandler(fakeDomainDB, exitCh)
	})

	Describe("Upsert", func() {
		var (
			domain string
			ttl    uint32
		)

		BeforeEach(func() {
			domain = "domain-to-add"
			ttl = 12345

			requestBody = &models.ProtoUpsertDomainRequest{
				Domain: domain,
				Ttl:    ttl,
			}
		})

		JustBeforeEach(func() {
			request := newTestRequest(requestBody)
			request.Header.Set(lager.RequestIdHeader, requestIdHeader)
			handler.Upsert(logger, responseRecorder, request)
		})

		Context("when upserting domain to DB succeeds", func() {
			BeforeEach(func() {
				fakeDomainDB.UpsertDomainReturns(nil)
			})

			It("call the DB to upsert the domain", func() {
				Expect(fakeDomainDB.UpsertDomainCallCount()).To(Equal(1))
				_, _, domainUpserted, ttlUpserted := fakeDomainDB.UpsertDomainArgsForCall(0)
				Expect(domainUpserted).To(Equal(domain))
				Expect(ttlUpserted).To(Equal(ttl))
			})

			It("responds with 200 OK", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})

			It("responds with no error", func() {
				var response models.UpsertDomainResponse
				var protoResponse models.ProtoUpsertDomainResponse
				err := proto.Unmarshal(responseRecorder.Body.Bytes(), &protoResponse)
				response = *protoResponse.FromProto()
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(BeNil())
			})
		})

		Context("when the request is invalid", func() {
			BeforeEach(func() {
				requestBody = &models.ProtoUpsertDomainRequest{}
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))

				var response models.UpsertDomainResponse
				var protoResponse models.ProtoUpsertDomainResponse
				err := proto.Unmarshal(responseRecorder.Body.Bytes(), &protoResponse)
				response = *protoResponse.FromProto()
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).NotTo(BeNil())
				Expect(response.Error.Type).To(Equal(models.Error_InvalidRequest))
			})
		})

		Context("when parsing the body crashs", func() {
			BeforeEach(func() {
				requestBody = "beep boop beep boop -- i am a robot"
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))

				var response models.UpsertDomainResponse
				var protoResponse models.ProtoUpsertDomainResponse
				err := proto.Unmarshal(responseRecorder.Body.Bytes(), &protoResponse)
				response = *protoResponse.FromProto()
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).NotTo(BeNil())
				Expect(response.Error).To(Equal(models.ErrBadRequest))
			})
		})

		Context("when the DB returns an unrecoverable error", func() {
			BeforeEach(func() {
				fakeDomainDB.UpsertDomainReturns(models.NewUnrecoverableError(nil))
			})

			It("logs and writes to the exit channel", func() {
				Eventually(logger).Should(gbytes.Say("unrecoverable-error"))
				Eventually(logger).Should(gbytes.Say(b3RequestIdHeader))
				Eventually(exitCh).Should(Receive())
			})
		})

		Context("when the DB errors out", func() {
			BeforeEach(func() {
				fakeDomainDB.UpsertDomainReturns(models.ErrUnknownError)
			})

			It("provides relevant error information", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))

				var response models.UpsertDomainResponse
				var protoResponse models.ProtoUpsertDomainResponse
				err := proto.Unmarshal(responseRecorder.Body.Bytes(), &protoResponse)
				response = *protoResponse.FromProto()
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).NotTo(BeNil())
				Expect(response.Error).To(Equal(models.ErrUnknownError))
			})
		})
	})

	Describe("Domains", func() {
		var domains []string

		BeforeEach(func() {
			domains = []string{"domain-a", "domain-b"}
		})

		JustBeforeEach(func() {
			request := newTestRequest("")
			request.Header.Set(lager.RequestIdHeader, requestIdHeader)
			handler.Domains(logger, responseRecorder, request)
		})

		Context("when reading domains from DB succeeds", func() {
			BeforeEach(func() {
				fakeDomainDB.FreshDomainsReturns(domains, nil)
			})

			It("call the DB to retrieve the domains", func() {
				Expect(fakeDomainDB.FreshDomainsCallCount()).To(Equal(1))
			})

			It("returns a list of domains", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))

				var response models.DomainsResponse
				var protoResponse models.ProtoDomainsResponse
				err := proto.Unmarshal(responseRecorder.Body.Bytes(), &protoResponse)
				response = *protoResponse.FromProto()
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(BeNil())
				Expect(response.Domains).To(ConsistOf(domains))
			})
		})

		Context("when the DB returns no domains", func() {
			BeforeEach(func() {
				fakeDomainDB.FreshDomainsReturns([]string{}, nil)
			})

			It("returns an empty list", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))

				var response models.DomainsResponse
				var protoResponse models.ProtoDomainsResponse
				err := proto.Unmarshal(responseRecorder.Body.Bytes(), &protoResponse)
				response = *protoResponse.FromProto()
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(BeNil())
				Expect(response.Domains).To(BeNil())
			})
		})

		Context("when the DB returns an unrecoverable error", func() {
			BeforeEach(func() {
				fakeDomainDB.FreshDomainsReturns([]string{}, models.NewUnrecoverableError(nil))
			})

			It("logs and writes to the exit channel", func() {
				Eventually(logger).Should(gbytes.Say("unrecoverable-error"))
				Eventually(logger).Should(gbytes.Say(b3RequestIdHeader))
				Eventually(exitCh).Should(Receive())
			})
		})

		Context("when the DB errors out", func() {
			BeforeEach(func() {
				fakeDomainDB.FreshDomainsReturns([]string{}, models.ErrUnknownError)
			})

			It("provides relevant error information", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))

				var response models.DomainsResponse
				var protoResponse models.ProtoDomainsResponse
				err := proto.Unmarshal(responseRecorder.Body.Bytes(), &protoResponse)
				response = *protoResponse.FromProto()
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrUnknownError))
				Expect(response.Domains).To(BeNil())
			})
		})
	})
})
