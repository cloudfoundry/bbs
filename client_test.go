package bbs_test

import (
	"net/http"
	"time"

	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/cfhttp"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Client", func() {
	var (
		bbsServer *ghttp.Server
		client    bbs.Client
		cfg       bbs.ClientConfig
		logger    lager.Logger
	)

	BeforeEach(func() {
		bbsServer = ghttp.NewServer()
		cfg.URL = bbsServer.URL()
		cfg.Retries = 1

		logger = lagertest.NewTestLogger("bbs-client")
	})

	AfterEach(func() {
		bbsServer.CloseClientConnections()
		bbsServer.Close()
	})

	JustBeforeEach(func() {
		var err error
		client, err = bbs.NewClientWithConfig(cfg)
		Expect(err).ToNot(HaveOccurred())
	})

	Context("when the server responds successfully after some time", func() {
		var serverTimeout time.Duration

		BeforeEach(func() {
			serverTimeout = 30 * time.Millisecond
		})
		JustBeforeEach(func() {
			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/actual_lrp_groups/list"),
					func(w http.ResponseWriter, req *http.Request) {
						time.Sleep(serverTimeout)
					},
					ghttp.RespondWithProto(200, &models.ActualLRPGroupsResponse{
						ActualLrpGroups: []*models.ActualLRPGroup{
							{
								Instance: &models.ActualLRP{
									ActualLRPInfo: models.ActualLRPInfo{
										State: "running",
									},
								},
							},
						},
					}),
				),
			)
		})

		It("returns the successful response", func() {
			lrps, err := client.ActualLRPGroups(logger, models.ActualLRPFilter{})
			Expect(err).ToNot(HaveOccurred())
			Expect(lrps).To(ConsistOf(&models.ActualLRPGroup{
				Instance: &models.ActualLRP{
					ActualLRPInfo: models.ActualLRPInfo{
						State: "running",
					},
				},
			}))
		})

		Context("when the client is configured with a small timeout", func() {
			BeforeEach(func() {
				cfhttp.Initialize(20 * time.Millisecond)
			})

			It("fails the request with a timeout error", func() {
				_, err := client.ActualLRPGroups(logger, models.ActualLRPFilter{})
				var apiError *models.Error
				Expect(err).To(HaveOccurred())
				Expect(err).To(BeAssignableToTypeOf(apiError))
				apiError = err.(*models.Error)
				Expect(apiError.Type).To(Equal(models.Error_Timeout))
			})

		})

	})

	Context("when the server responds with a 500", func() {
		JustBeforeEach(func() {
			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/actual_lrp_groups/list"),
					ghttp.RespondWith(500, nil),
				),
			)
		})

		It("returns the error", func() {
			_, err := client.ActualLRPGroups(logger, models.ActualLRPFilter{})
			Expect(err).To(HaveOccurred())
			responseError := err.(*models.Error)
			Expect(responseError.Type).To(Equal(models.Error_InvalidResponse))
		})
	})

	Context("when an http URL is provided to the secure client", func() {
		It("creating the client returns an error", func() {
			_, err := bbs.NewClient(bbsServer.URL(), "", "", "", 1, 1)
			Expect(err).To(MatchError("Expected https URL"))
		})
	})
})
