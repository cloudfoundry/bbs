package middleware_test

import (
	"net/http"
	"time"

	"code.cloudfoundry.org/bbs/handlers/middleware"
	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/dropsonde/metric_sender/fake"
	dropsonde_metrics "github.com/cloudfoundry/dropsonde/metrics"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Middleware", func() {
	Describe("EmitLatency", func() {
		var (
			sender  *fake.FakeMetricSender
			handler http.HandlerFunc
		)

		BeforeEach(func() {
			sender = fake.NewFakeMetricSender()
			dropsonde_metrics.Initialize(sender, nil)
			logger := lager.NewLogger("test-session")

			handler = func(w http.ResponseWriter, r *http.Request) { time.Sleep(10) }
			handler = middleware.NewLatencyEmitter(logger).EmitLatency(handler)
		})

		It("reports latency", func() {
			handler.ServeHTTP(nil, nil)

			latency := sender.GetValue("RequestLatency")
			Expect(latency.Value).NotTo(BeZero())
			Expect(latency.Unit).To(Equal("nanos"))
		})
	})

	Describe("RequestCountWrap", func() {
		var (
			sender  *fake.FakeMetricSender
			handler http.HandlerFunc
		)

		BeforeEach(func() {
			sender = fake.NewFakeMetricSender()
			dropsonde_metrics.Initialize(sender, nil)

			handler = func(w http.ResponseWriter, r *http.Request) { time.Sleep(10) }
			handler = middleware.RequestCountWrap(handler)
		})

		It("reports call count", func() {
			handler.ServeHTTP(nil, nil)
			handler.ServeHTTP(nil, nil)
			handler.ServeHTTP(nil, nil)

			Expect(sender.GetCounter("RequestCount")).To(Equal(uint64(3)))
		})
	})
})
