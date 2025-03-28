package middleware_test

import (
	"code.cloudfoundry.org/bbs/cmd/bbs/config"
	"net/http"
	"time"

	"code.cloudfoundry.org/bbs/handlers/middleware"
	"code.cloudfoundry.org/bbs/handlers/middleware/fakes"
	"code.cloudfoundry.org/lager/v3"
	"code.cloudfoundry.org/lager/v3/lagertest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

const (
	waitTime  = 100 * time.Millisecond
	waitDelay = 10 * time.Millisecond
)

var _ = Describe("Test Middleware", func() {
	Context("Record Metrics", func() {
		var (
			handler http.HandlerFunc
			emitter *fakes.FakeEmitter
		)

		BeforeEach(func() {
			emitter = &fakes.FakeEmitter{}
			handler = func(w http.ResponseWriter, r *http.Request) { time.Sleep(waitTime) }
		})

		Context("Default Metrics", func() {
			const callsNum = 3
			When("RecordRequestCount is called", func() {
				JustBeforeEach(func() {
					handler = middleware.RecordRequestCount(handler, emitter)
				})

				It("should report call count", func() {
					for i := 0; i < callsNum; i++ {
						handler.ServeHTTP(nil, nil)
					}

					Expect(emitter.IncrementRequestCounterCallCount()).To(Equal(callsNum))
					for i := 0; i < callsNum; i++ {
						actualDelta, route := emitter.IncrementRequestCounterArgsForCall(i)
						Expect(actualDelta).To(Equal(1))
						Expect(route).To(Equal(""))
					}
				})
			})

			When("RecordLatency is called", func() {
				JustBeforeEach(func() {
					handler = middleware.RecordLatency(handler, emitter)
				})

				It("should report latency", func() {
					handler.ServeHTTP(nil, nil)
					Expect(emitter.UpdateLatencyCallCount()).To(Equal(1))
					actualDuration, route := emitter.UpdateLatencyArgsForCall(0)
					Expect(actualDuration).To(BeNumerically(">=", waitTime))
					Expect(actualDuration).To(BeNumerically("<=", waitTime+waitDelay))
					Expect(route).To(Equal(""))
				})
			})
		})

		Context("Advanced metrics", func() {
			var (
				advancedMetricsConfig  config.AdvancedMetrics
				route                  string
				didJustBeforeEachPanic bool
			)

			JustBeforeEach(func() {
				defer func() {
					if r := recover(); r != nil {
						didJustBeforeEachPanic = true
					}
				}()

				didJustBeforeEachPanic = false
				handler = middleware.RecordMetrics(handler, emitter, advancedMetricsConfig, route)
			})

			When("disabled", func() {
				BeforeEach(func() {
					route = ""
					advancedMetricsConfig = config.AdvancedMetrics{
						Enabled: false,
						RouteConfig: config.RouteConfiguration{
							RequestCountRoutes:   []string{},
							RequestLatencyRoutes: []string{},
						},
					}
				})

				When("no route is passed", func() {
					It("should not emit advanced metrics per route", func() {
						validateRecordDefaultMetricsWithOneRequest(handler, emitter)
					})
				})

				When("route is passed", func() {
					BeforeEach(func() {
						route = "TEST_ROUTE"
					})

					It("should not call the emitter with any passed route", func() {
						validateRecordDefaultMetricsWithOneRequest(handler, emitter)
					})
				})
			})

			When("enabled", func() {
				BeforeEach(func() {
					route = "TEST_ROUTE"
					advancedMetricsConfig = config.AdvancedMetrics{
						Enabled: true,
						RouteConfig: config.RouteConfiguration{
							RequestCountRoutes:   []string{},
							RequestLatencyRoutes: []string{},
						},
					}
				})

				When("no route is passed", func() {
					BeforeEach(func() {
						route = ""
					})

					It("should panic", func() {
						Expect(didJustBeforeEachPanic).To(BeTrue())
					})
				})

				When("route is passed", func() {
					When("and NOT IN RequestCountRoutes and NOT IN RequestLatencyRoutes", func() {
						It("should emit only Default Metrics", func() {
							validateRecordDefaultMetricsWithOneRequest(handler, emitter)
						})
					})
					When("and NOT IN RequestCountRoutes and IN RequestLatencyRoutes", func() {
						BeforeEach(func() {
							advancedMetricsConfig.RouteConfig.RequestLatencyRoutes = []string{route}
						})

						It("should emit Default metrics and (Advanced metrics for RequestLatencyRoutes)", func() {
							handler.ServeHTTP(nil, nil)
							Expect(emitter.IncrementRequestCounterCallCount()).To(Equal(1))

							actualIncrementCount, actualIncrementRoute := emitter.IncrementRequestCounterArgsForCall(0)
							Expect(actualIncrementCount).To(Equal(1))
							Expect(actualIncrementRoute).To(Equal(""))

							Expect(emitter.UpdateLatencyCallCount()).To(Equal(2))

							actualLatencyDuration, actualLatencyRoute := emitter.UpdateLatencyArgsForCall(0)
							Expect(actualLatencyDuration).To(BeNumerically(">=", waitTime))
							Expect(actualLatencyDuration).To(BeNumerically("<=", waitTime+waitDelay))
							Expect(actualLatencyRoute).To(Equal(""))

							actualLatencyDurationSecondCall, actualLatencyRouteSecondCall := emitter.UpdateLatencyArgsForCall(1)
							Expect(actualLatencyDurationSecondCall).To(Equal(actualLatencyDuration))
							Expect(actualLatencyRouteSecondCall).To(Equal(route))
						})
					})
					When("and IN RequestCountRoutes and NOT IN RequestLatencyRoutes", func() {
						BeforeEach(func() {
							advancedMetricsConfig.RouteConfig.RequestCountRoutes = []string{route}
						})

						It("should emit Default metrics and (Advanced metrics for RequestCountRoutes)", func() {
							handler.ServeHTTP(nil, nil)
							Expect(emitter.IncrementRequestCounterCallCount()).To(Equal(2))

							actualIncrementCount, actualIncrementRoute := emitter.IncrementRequestCounterArgsForCall(0)
							Expect(actualIncrementCount).To(Equal(1))
							Expect(actualIncrementRoute).To(Equal(""))
							actualIncrementCountSecondCall, actualIncrementRouteSecondCall := emitter.IncrementRequestCounterArgsForCall(1)
							Expect(actualIncrementCountSecondCall).To(Equal(1))
							Expect(actualIncrementRouteSecondCall).To(Equal(route))

							Expect(emitter.UpdateLatencyCallCount()).To(Equal(1))

							actualLatencyDuration, actualLatencyRoute := emitter.UpdateLatencyArgsForCall(0)
							Expect(actualLatencyDuration).To(BeNumerically(">=", waitTime))
							Expect(actualLatencyDuration).To(BeNumerically("<=", waitTime+waitDelay))
							Expect(actualLatencyRoute).To(Equal(""))
						})
					})

					When("and IN RequestCountRoutes and IN RequestLatencyRoutes", func() {
						BeforeEach(func() {
							advancedMetricsConfig.RouteConfig.RequestCountRoutes = []string{route}
							advancedMetricsConfig.RouteConfig.RequestLatencyRoutes = []string{route}
						})

						It("should emit Default and Advanced metrics for RequestCountRoutes and RequestLatencyRoutes", func() {
							handler.ServeHTTP(nil, nil)
							Expect(emitter.IncrementRequestCounterCallCount()).To(Equal(2))

							actualIncrementCount, actualIncrementRoute := emitter.IncrementRequestCounterArgsForCall(0)
							Expect(actualIncrementCount).To(Equal(1))
							Expect(actualIncrementRoute).To(Equal(""))
							actualIncrementCountSecondCall, actualIncrementRouteSecondCall := emitter.IncrementRequestCounterArgsForCall(1)
							Expect(actualIncrementCountSecondCall).To(Equal(1))
							Expect(actualIncrementRouteSecondCall).To(Equal(route))

							Expect(emitter.UpdateLatencyCallCount()).To(Equal(2))

							actualLatencyDuration, actualLatencyRoute := emitter.UpdateLatencyArgsForCall(0)
							Expect(actualLatencyDuration).To(BeNumerically(">=", waitTime))
							Expect(actualLatencyDuration).To(BeNumerically("<=", waitTime+waitDelay))
							Expect(actualLatencyRoute).To(Equal(""))

							actualLatencyDurationSecondCall, actualLatencyRouteSecondCall := emitter.UpdateLatencyArgsForCall(1)
							Expect(actualLatencyDurationSecondCall).To(Equal(actualLatencyDuration))
							Expect(actualLatencyRouteSecondCall).To(Equal(route))
						})
					})

				})
			})
		})
	})

	Context("LogWrap", func() {
		var (
			logger              *lagertest.TestLogger
			loggableHandlerFunc middleware.LoggableHandlerFunc
		)

		BeforeEach(func() {
			logger = lagertest.NewTestLogger("test-session")
			logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
			loggableHandlerFunc = func(logger lager.Logger, w http.ResponseWriter, r *http.Request) {
				logger = logger.Session("logger-group")
				logger.Info("written-in-loggable-handler")
			}
		})

		It("creates \"request\" session and passes it to LoggableHandlerFunc", func() {
			handler := middleware.LogWrap(logger, nil, loggableHandlerFunc)
			req, err := http.NewRequest("GET", "http://example.com", nil)
			Expect(err).NotTo(HaveOccurred())
			handler.ServeHTTP(nil, req)
			Expect(logger.Buffer()).To(gbytes.Say("test-session.request.serving"))
			Expect(logger.Buffer()).To(gbytes.Say("\"session\":\"1\""))
			Expect(logger.Buffer()).To(gbytes.Say("test-session.request.logger-group.written-in-loggable-handler"))
			Expect(logger.Buffer()).To(gbytes.Say("\"session\":\"1.1\""))
			Expect(logger.Buffer()).To(gbytes.Say("test-session.request.done"))
			Expect(logger.Buffer()).To(gbytes.Say("\"session\":\"1\""))
		})

		Context("with access loggger", func() {
			var accessLogger *lagertest.TestLogger

			BeforeEach(func() {
				accessLogger = lagertest.NewTestLogger("test-access-session")
				accessLogger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))

				handler := middleware.LogWrap(logger, accessLogger, loggableHandlerFunc)
				req, err := http.NewRequest("GET", "http://example.com", nil)
				Expect(err).NotTo(HaveOccurred())
				req.RemoteAddr = "127.0.0.1:8080"

				handler.ServeHTTP(nil, req)
			})

			It("creates \"request\" session and passes it to LoggableHandlerFunc", func() {
				Expect(logger.Buffer()).To(gbytes.Say("test-session.request.serving"))
				Expect(logger.Buffer()).To(gbytes.Say("\"session\":\"1\""))
				Expect(accessLogger.Buffer()).To(gbytes.Say("test-access-session.request.serving"))
				Expect(accessLogger.Buffer()).To(gbytes.Say("\"session\":\"1\""))

				Expect(logger.Buffer()).To(gbytes.Say("test-session.request.logger-group.written-in-loggable-handler"))
				Expect(logger.Buffer()).To(gbytes.Say("\"session\":\"1.1\""))

				Expect(accessLogger.Buffer()).To(gbytes.Say("test-access-session.request.done"))
				Expect(accessLogger.Buffer()).To(gbytes.Say("\"session\":\"1\""))
				Expect(logger.Buffer()).To(gbytes.Say("test-session.request.done"))
				Expect(logger.Buffer()).To(gbytes.Say("\"session\":\"1\""))
			})

			It("logs method, request, ip, and port to serving and done logs", func() {
				Expect(logger.Buffer()).To(gbytes.Say("test-session.request.serving"))
				Expect(logger.Buffer()).To(gbytes.Say("method\":\"GET\""))
				Expect(logger.Buffer()).To(gbytes.Say("remote_addr\":\"127.0.0.1:8080\""))
				Expect(logger.Buffer()).To(gbytes.Say("request\":\"http://example.com\""))

				Expect(accessLogger.Buffer()).To(gbytes.Say("test-access-session.request.serving"))
				Expect(accessLogger.Buffer()).To(gbytes.Say("method\":\"GET\""))
				Expect(accessLogger.Buffer()).To(gbytes.Say("remote_addr\":\"127.0.0.1:8080\""))
				Expect(accessLogger.Buffer()).To(gbytes.Say("request\":\"http://example.com\""))

				Expect(logger.Buffer()).To(gbytes.Say("test-session.request.done"))
				Expect(logger.Buffer()).To(gbytes.Say("method\":\"GET\""))
				Expect(logger.Buffer()).To(gbytes.Say("remote_addr\":\"127.0.0.1:8080\""))
				Expect(logger.Buffer()).To(gbytes.Say("request\":\"http://example.com\""))

				Expect(accessLogger.Buffer()).To(gbytes.Say("test-access-session.request.done"))
				Expect(accessLogger.Buffer()).To(gbytes.Say("method\":\"GET\""))
				Expect(accessLogger.Buffer()).To(gbytes.Say("remote_addr\":\"127.0.0.1:8080\""))
				Expect(accessLogger.Buffer()).To(gbytes.Say("request\":\"http://example.com\""))
			})
		})
	})
})

func validateRecordDefaultMetricsWithOneRequest(handler http.HandlerFunc, emitter *fakes.FakeEmitter) {

	handler.ServeHTTP(nil, nil)
	Expect(emitter.IncrementRequestCounterCallCount()).To(Equal(1))

	actualIncrementCount, actualIncrementRoute := emitter.IncrementRequestCounterArgsForCall(0)
	Expect(actualIncrementCount).To(Equal(1))
	Expect(actualIncrementRoute).To(Equal(""))

	Expect(emitter.UpdateLatencyCallCount()).To(Equal(1))

	actualLatencyDuration, actualLatencyRoute := emitter.UpdateLatencyArgsForCall(0)
	Expect(actualLatencyDuration).To(BeNumerically(">=", waitTime))
	Expect(actualLatencyDuration).To(BeNumerically("<=", waitTime+waitDelay))
	Expect(actualLatencyRoute).To(Equal(""))
}
