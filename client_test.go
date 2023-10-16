package bbs_test

import (
	"context"
	"net/http"
	"path"
	"time"

	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/models/test/model_helpers"
	"code.cloudfoundry.org/lager/v3"
	"code.cloudfoundry.org/lager/v3/lagertest"
	"code.cloudfoundry.org/tlsconfig"

	. "github.com/onsi/ginkgo/v2"
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
		cfg = bbs.ClientConfig{
			URL:           bbsServer.URL(),
			Retries:       1,
			RetryInterval: time.Millisecond,
		}

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

	Context("internal endpoints with different versions", func() {
		var (
			internalClient bbs.InternalClient
		)
		JustBeforeEach(func() {
			var err error
			internalClient, err = bbs.NewClientWithConfig(cfg)
			Expect(err).ToNot(HaveOccurred())
		})
		Context("StartActualLRP", func() {
			It("populates the request", func() {
				actualLRP := model_helpers.NewValidActualLRP("some-guid", 0)
				request := models.StartActualLRPRequest{
					ActualLrpKey:            &actualLRP.ActualLRPKey,
					ActualLrpInstanceKey:    &actualLRP.ActualLRPInstanceKey,
					ActualLrpNetInfo:        &actualLRP.ActualLRPNetInfo,
					ActualLrpInternalRoutes: actualLRP.ActualLrpInternalRoutes,
					MetricTags:              actualLRP.MetricTags,
				}
				request.SetRoutable(false)
				bbsServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v1/actual_lrps/start.r1"),
						ghttp.VerifyHeader(http.Header{"X-Vcap-Request-Id": []string{"some-trace-id"}}),
						ghttp.VerifyProtoRepresenting(&request),
						ghttp.RespondWithProto(200, &models.ActualLRPLifecycleResponse{Error: nil}),
					),
				)

				err := internalClient.StartActualLRP(logger, "some-trace-id", &actualLRP.ActualLRPKey, &actualLRP.ActualLRPInstanceKey, &actualLRP.ActualLRPNetInfo, actualLRP.ActualLrpInternalRoutes, actualLRP.MetricTags, false)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Calls the current endpoint", func() {
				bbsServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v1/actual_lrps/start.r1"),
						ghttp.VerifyHeader(http.Header{"X-Vcap-Request-Id": []string{"some-trace-id"}}),
						ghttp.RespondWithProto(200, &models.ActualLRPLifecycleResponse{Error: nil}),
					),
				)

				err := internalClient.StartActualLRP(logger, "some-trace-id", &models.ActualLRPKey{}, &models.ActualLRPInstanceKey{}, &models.ActualLRPNetInfo{}, []*models.ActualLRPInternalRoute{}, map[string]string{}, false)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Falls back to the deprecated endpoint if the current endpoint returns a 404", func() {
				actualLRP := model_helpers.NewValidActualLRP("some-guid", 0)
				bbsServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v1/actual_lrps/start.r1"),
						ghttp.VerifyHeader(http.Header{"X-Vcap-Request-Id": []string{"some-trace-id"}}),
						ghttp.RespondWith(http.StatusNotFound, nil),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v1/actual_lrps/start"),
						ghttp.VerifyHeader(http.Header{"X-Vcap-Request-Id": []string{"some-trace-id"}}),
						ghttp.VerifyProtoRepresenting(&models.StartActualLRPRequest{
							ActualLrpKey:            &actualLRP.ActualLRPKey,
							ActualLrpInstanceKey:    &actualLRP.ActualLRPInstanceKey,
							ActualLrpNetInfo:        &actualLRP.ActualLRPNetInfo,
							ActualLrpInternalRoutes: nil,
							MetricTags:              nil,
						}),
						ghttp.RespondWithProto(200, &models.ActualLRPLifecycleResponse{Error: nil}),
					),
				)

				err := internalClient.StartActualLRP(logger, "some-trace-id", &actualLRP.ActualLRPKey, &actualLRP.ActualLRPInstanceKey, &actualLRP.ActualLRPNetInfo, actualLRP.ActualLrpInternalRoutes, actualLRP.MetricTags, actualLRP.GetRoutable())
				Expect(err).NotTo(HaveOccurred())
			})

			It("Returns an error if the current call returns a non-successful non-404 status code", func() {
				bbsServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v1/actual_lrps/start.r1"),
						ghttp.VerifyHeader(http.Header{"X-Vcap-Request-Id": []string{"some-trace-id"}}),
						ghttp.RespondWith(http.StatusForbidden, nil),
					),
				)

				err := internalClient.StartActualLRP(logger, "some-trace-id", &models.ActualLRPKey{}, &models.ActualLRPInstanceKey{}, &models.ActualLRPNetInfo{}, []*models.ActualLRPInternalRoute{}, map[string]string{}, false)
				Expect(err).To(MatchError("Invalid Response with status code: 403"))
			})

			It("Still returns an error if the fallback call fails", func() {
				bbsServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v1/actual_lrps/start.r1"),
						ghttp.VerifyHeader(http.Header{"X-Vcap-Request-Id": []string{"some-trace-id"}}),
						ghttp.RespondWith(http.StatusNotFound, nil),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v1/actual_lrps/start"),
						ghttp.VerifyHeader(http.Header{"X-Vcap-Request-Id": []string{"some-trace-id"}}),
						ghttp.RespondWith(http.StatusForbidden, nil),
					),
				)

				err := internalClient.StartActualLRP(logger, "some-trace-id", &models.ActualLRPKey{}, &models.ActualLRPInstanceKey{}, &models.ActualLRPNetInfo{}, []*models.ActualLRPInternalRoute{}, map[string]string{}, false)
				Expect(err).To(MatchError("Invalid Response with status code: 403"))
			})
		})

		Context("evacuateRunningActualLrp", func() {
			It("populates the request", func() {
				actualLRP := model_helpers.NewValidActualLRP("some-guid", 0)
				bbsServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v1/actual_lrps/evacuate_running.r1"),
						ghttp.VerifyHeader(http.Header{"X-Vcap-Request-Id": []string{"some-trace-id"}}),
						ghttp.VerifyProtoRepresenting(&models.EvacuateRunningActualLRPRequest{
							ActualLrpKey:            &actualLRP.ActualLRPKey,
							ActualLrpInstanceKey:    &actualLRP.ActualLRPInstanceKey,
							ActualLrpNetInfo:        &actualLRP.ActualLRPNetInfo,
							ActualLrpInternalRoutes: actualLRP.ActualLrpInternalRoutes,
							MetricTags:              actualLRP.MetricTags,
							OptionalRoutable:        &models.EvacuateRunningActualLRPRequest_Routable{Routable: actualLRP.GetRoutable()},
						}),
						ghttp.RespondWithProto(200, &models.EvacuationResponse{KeepContainer: true, Error: nil}),
					),
				)

				_, err := internalClient.EvacuateRunningActualLRP(logger, "some-trace-id", &actualLRP.ActualLRPKey, &actualLRP.ActualLRPInstanceKey, &actualLRP.ActualLRPNetInfo, actualLRP.ActualLrpInternalRoutes, actualLRP.MetricTags, actualLRP.GetRoutable())
				Expect(err).NotTo(HaveOccurred())
			})
			It("Calls the current endpoint", func() {
				bbsServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v1/actual_lrps/evacuate_running.r1"),
						ghttp.VerifyHeader(http.Header{"X-Vcap-Request-Id": []string{"some-trace-id"}}),
						ghttp.RespondWithProto(200, &models.EvacuationResponse{KeepContainer: true, Error: nil}),
					),
				)

				_, err := internalClient.EvacuateRunningActualLRP(logger, "some-trace-id", &models.ActualLRPKey{}, &models.ActualLRPInstanceKey{}, &models.ActualLRPNetInfo{}, []*models.ActualLRPInternalRoute{}, map[string]string{}, false)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Falls back to the deprecated endpoint if the current endpoint returns a 404", func() {
				actualLRP := model_helpers.NewValidActualLRP("some-guid", 0)
				bbsServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v1/actual_lrps/evacuate_running.r1"),
						ghttp.VerifyHeader(http.Header{"X-Vcap-Request-Id": []string{"some-trace-id"}}),
						ghttp.RespondWith(http.StatusNotFound, nil),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v1/actual_lrps/evacuate_running"),
						ghttp.VerifyHeader(http.Header{"X-Vcap-Request-Id": []string{"some-trace-id"}}),
						ghttp.VerifyProtoRepresenting(&models.EvacuateRunningActualLRPRequest{
							ActualLrpKey:            &actualLRP.ActualLRPKey,
							ActualLrpInstanceKey:    &actualLRP.ActualLRPInstanceKey,
							ActualLrpNetInfo:        &actualLRP.ActualLRPNetInfo,
							ActualLrpInternalRoutes: nil,
							MetricTags:              nil,
							OptionalRoutable:        nil,
						}),
						ghttp.RespondWithProto(200, &models.EvacuationResponse{KeepContainer: true, Error: nil}),
					),
				)

				_, err := internalClient.EvacuateRunningActualLRP(logger, "some-trace-id", &actualLRP.ActualLRPKey, &actualLRP.ActualLRPInstanceKey, &actualLRP.ActualLRPNetInfo, actualLRP.ActualLrpInternalRoutes, actualLRP.MetricTags, false)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Returns an error if the current call returns a non-successful non-404 status code", func() {
				bbsServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v1/actual_lrps/evacuate_running.r1"),
						ghttp.VerifyHeader(http.Header{"X-Vcap-Request-Id": []string{"some-trace-id"}}),
						ghttp.RespondWith(http.StatusForbidden, nil),
					),
				)

				_, err := internalClient.EvacuateRunningActualLRP(logger, "some-trace-id", &models.ActualLRPKey{}, &models.ActualLRPInstanceKey{}, &models.ActualLRPNetInfo{}, []*models.ActualLRPInternalRoute{}, map[string]string{}, false)
				Expect(err).To(MatchError("Invalid Response with status code: 403"))
			})

			It("Still returns an error if the fallback call fails", func() {
				bbsServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v1/actual_lrps/evacuate_running.r1"),
						ghttp.VerifyHeader(http.Header{"X-Vcap-Request-Id": []string{"some-trace-id"}}),
						ghttp.RespondWith(http.StatusNotFound, nil),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v1/actual_lrps/evacuate_running"),
						ghttp.VerifyHeader(http.Header{"X-Vcap-Request-Id": []string{"some-trace-id"}}),
						ghttp.RespondWith(http.StatusForbidden, nil),
					),
				)

				_, err := internalClient.EvacuateRunningActualLRP(logger, "some-trace-id", &models.ActualLRPKey{}, &models.ActualLRPInstanceKey{}, &models.ActualLRPNetInfo{}, []*models.ActualLRPInternalRoute{}, map[string]string{}, false)
				Expect(err).To(MatchError("Invalid Response with status code: 403"))
			})
		})
	})

	Context("when the request timeout is explicitly set", func() {
		Context("when the client is not configured to use TLS", func() {
			BeforeEach(func() {
				cfg.RequestTimeout = 2 * time.Second

				bbsServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v1/actual_lrp_groups/list"),
						ghttp.VerifyHeader(http.Header{"X-Vcap-Request-Id": []string{"some-trace-id"}}),
						func(w http.ResponseWriter, req *http.Request) {
							time.Sleep(cfg.RequestTimeout * 2)
						},
						ghttp.RespondWith(418, nil),
					),
				)
			})

			It("respects the request timeout", func() {
				_, err := client.ActualLRPGroups(logger, "some-trace-id", models.ActualLRPFilter{})
				Expect(err.Error()).To(ContainSubstring(context.DeadlineExceeded.Error()))
			})
		})

		Context("when the client is configured to use TLS", func() {
			var tlsServer *ghttp.Server

			BeforeEach(func() {
				basePath := path.Join("cmd", "bbs", "fixtures")
				caFile := path.Join(basePath, "green-certs", "server-ca.crt")

				cfg.IsTLS = true
				cfg.CAFile = caFile
				cfg.CertFile = path.Join(basePath, "green-certs", "client.crt")
				cfg.KeyFile = path.Join(basePath, "green-certs", "client.key")
				cfg.RequestTimeout = 2 * time.Second

				tlsServer = ghttp.NewUnstartedServer()

				tlsConfig, err := tlsconfig.Build(
					tlsconfig.WithInternalServiceDefaults(),
					tlsconfig.WithIdentityFromFile(
						path.Join(basePath, "green-certs", "server.crt"),
						path.Join(basePath, "green-certs", "server.key"),
					),
				).Server(tlsconfig.WithClientAuthenticationFromFile(caFile))
				Expect(err).NotTo(HaveOccurred())

				tlsServer.HTTPTestServer.TLS = tlsConfig
				tlsServer.HTTPTestServer.StartTLS()
				cfg.URL = tlsServer.URL()

				tlsServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v1/actual_lrp_groups/list"),
						ghttp.VerifyHeader(http.Header{"X-Vcap-Request-Id": []string{"some-trace-id"}}),
						func(w http.ResponseWriter, req *http.Request) {
							time.Sleep(cfg.RequestTimeout * 2)
						},
						ghttp.RespondWith(418, nil),
					),
				)
			})

			AfterEach(func() {
				tlsServer.CloseClientConnections()
				tlsServer.Close()
			})

			It("respects the request timeout", func() {
				_, err := client.ActualLRPGroups(logger, "some-trace-id", models.ActualLRPFilter{})
				Expect(err.Error()).To(ContainSubstring(context.DeadlineExceeded.Error()))
			})
		})
	})

	Context("when the server responds successfully after some time", func() {
		var (
			serverTimeout time.Duration
			blockCh       chan struct{}
		)

		BeforeEach(func() {
			serverTimeout = 30 * time.Millisecond
			blockCh = make(chan struct{}, 1)
		})

		AfterEach(func() {
			close(blockCh)
		})

		JustBeforeEach(func() {
			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/actual_lrp_groups/list"),
					ghttp.VerifyHeader(http.Header{"X-Vcap-Request-Id": []string{"some-trace-id"}}),
					func(w http.ResponseWriter, req *http.Request) {
						<-blockCh
					},
					ghttp.RespondWithProto(200, &models.ActualLRPGroupsResponse{
						ActualLrpGroups: []*models.ActualLRPGroup{
							{
								Instance: &models.ActualLRP{
									State: "running",
								},
							},
						},
					}),
				),
			)
		})

		It("returns the successful response", func() {
			go func() {
				defer GinkgoRecover()

				time.Sleep(serverTimeout)
				Eventually(blockCh).Should(BeSent(struct{}{}))
			}()

			lrps, err := client.ActualLRPGroups(logger, "some-trace-id", models.ActualLRPFilter{})
			Expect(err).ToNot(HaveOccurred())
			Expect(lrps).To(ConsistOf(&models.ActualLRPGroup{
				Instance: &models.ActualLRP{
					State: "running",
				},
			}))
		})

		Context("when the client is configured with a small timeout", func() {
			BeforeEach(func() {
				cfg.RequestTimeout = 20 * time.Millisecond
			})

			It("fails the request with a timeout error", func() {
				_, err := client.ActualLRPGroups(logger, "some-trace-id", models.ActualLRPFilter{})
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
					ghttp.VerifyHeader(http.Header{"X-Vcap-Request-Id": []string{"some-trace-id"}}),
					ghttp.RespondWith(500, nil),
				),
			)
		})

		It("returns the error", func() {
			_, err := client.ActualLRPGroups(logger, "some-trace-id", models.ActualLRPFilter{})
			Expect(err).To(HaveOccurred())
			responseError := err.(*models.Error)
			Expect(responseError.Type).To(Equal(models.Error_InvalidResponse))
		})

	})

	Context("when subscribing to an event stream that fails", func() {
		JustBeforeEach(func() {
			bbsServer.HTTPTestServer.Listener.Close()
		})

		It("an error is returned", func() {
			errCh := make(chan error)
			go func(errCh chan error) {
				_, err := client.SubscribeToInstanceEventsByCellID(logger, "cell-uuid")
				if err != nil {
					errCh <- err
				}
			}(errCh)
			Eventually(errCh).Should(Receive())
		})

	})
	Context("when an http URL is provided to the secure client", func() {
		It("creating the client returns an error", func() {
			_, err := bbs.NewClient(bbsServer.URL(), "", "", "", 1, 1)
			Expect(err).To(MatchError("Expected https URL"))
		})
	})
})
