package handlers_test

import (
	"net/http"

	"github.com/cloudfoundry-incubator/bbs/handlers"
	"github.com/cloudfoundry/dropsonde/metric_sender/fake"
	dropsonde_metrics "github.com/cloudfoundry/dropsonde/metrics"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Middleware", func() {
	Describe("MeasureWrap", func() {
		var (
			sender  *fake.FakeMetricSender
			handler http.HandlerFunc
		)

		BeforeEach(func() {
			sender = fake.NewFakeMetricSender()
			dropsonde_metrics.Initialize(sender, nil)

			handler = func(w http.ResponseWriter, r *http.Request) {}
			handler = handlers.MeasureWrap(handler)
		})

		It("reports call count", func() {
			handler.ServeHTTP(nil, nil)
			handler.ServeHTTP(nil, nil)
			handler.ServeHTTP(nil, nil)

			Expect(sender.GetCounter("RequestCount")).To(Equal(uint64(3)))
		})

		It("reports latency", func() {
			handler.ServeHTTP(nil, nil)

			latency := sender.GetValue("RequestLatency")
			Expect(latency.Value).NotTo(BeZero())
			Expect(latency.Unit).To(Equal("nanos"))
		})
	})
})
