package metrics_test

import (
	"os"
	"sync"
	"time"

	"code.cloudfoundry.org/bbs/cmd/bbs/config"
	"code.cloudfoundry.org/bbs/metrics"
	"code.cloudfoundry.org/clock/fakeclock"
	mfakes "code.cloudfoundry.org/diego-logging-client/testhelpers"
	loggregator "code.cloudfoundry.org/go-loggregator/v9"
	"code.cloudfoundry.org/lager/v3/lagertest"
	"github.com/tedsuo/ifrit"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PeriodicMetronNotifier", func() {
	var (
		fakeMetronClient *mfakes.FakeIngressClient
		counterMap       map[string]uint64
		durationMap      map[string]time.Duration
		metricsLock      sync.Mutex

		reportInterval time.Duration
		fakeClock      *fakeclock.FakeClock

		mn                    *metrics.RequestStatMetronNotifier
		mnp                   ifrit.Process
		advancedMetricsConfig config.AdvancedMetrics
	)

	BeforeEach(func() {
		counterMap = make(map[string]uint64)
		durationMap = make(map[string]time.Duration)
		fakeMetronClient = new(mfakes.FakeIngressClient)
		fakeMetronClient.IncrementCounterWithDeltaStub = func(name string, delta uint64) error {
			metricsLock.Lock()
			defer metricsLock.Unlock()
			counterMap[name] = delta
			return nil
		}

		fakeMetronClient.SendDurationStub = func(name string, value time.Duration, opts ...loggregator.EmitGaugeOption) error {
			metricsLock.Lock()
			defer metricsLock.Unlock()
			durationMap[name] = value
			return nil
		}

		reportInterval = 100 * time.Millisecond

		fakeClock = fakeclock.NewFakeClock(time.Unix(123, 456))
	})

	JustBeforeEach(func() {
		ticker := fakeClock.NewTicker(reportInterval)
		mn = metrics.NewRequestStatMetronNotifier(lagertest.NewTestLogger("test"), ticker, fakeMetronClient, advancedMetricsConfig)
		mnp = ifrit.Invoke(mn)
	})

	AfterEach(func() {
		mnp.Signal(os.Interrupt)
		Eventually(mnp.Wait(), 2*time.Second).Should(Receive())
	})

	It("should emit a request count event periodically", func() {
		mn.IncrementRequestCounter(1, "")
		mn.UpdateLatency(5*time.Second, "")
		fakeClock.WaitForWatcherAndIncrement(reportInterval)

		Eventually(func() uint64 {
			metricsLock.Lock()
			defer metricsLock.Unlock()
			return counterMap["RequestCount"]
		}).Should(Equal(uint64(1)))

		Eventually(func() time.Duration {
			metricsLock.Lock()
			defer metricsLock.Unlock()
			return durationMap["RequestLatency"]
		}).Should(Equal(5 * time.Second))

		mn.IncrementRequestCounter(1, "")
		mn.UpdateLatency(3*time.Second, "")

		mn.IncrementRequestCounter(1, "")
		mn.UpdateLatency(2*time.Second, "")
		fakeClock.WaitForWatcherAndIncrement(reportInterval)

		Eventually(func() uint64 {
			metricsLock.Lock()
			defer metricsLock.Unlock()
			return counterMap["RequestCount"]
		}).Should(Equal(uint64(2)))

		Eventually(func() time.Duration {
			metricsLock.Lock()
			defer metricsLock.Unlock()
			return durationMap["RequestLatency"]
		}).Should(Equal(3 * time.Second))
	})

	Context("Advanced Metrics", func() {
		When("Advanced Metrics are disabled", func() {
			BeforeEach(func() {
				advancedMetricsConfig = config.AdvancedMetrics{
					Enabled: false,
					RouteConfig: config.RouteConfiguration{
						RequestCountRoutes:   []string{"TEST_ROUTE"},
						RequestLatencyRoutes: []string{"TEST_ROUTE"},
					},
				}
			})

			It("should panic on any route", func() {
				Expect(func() {
					mn.IncrementRequestCounter(1, "TEST_ROUTE")
				}).To(Panic())
				Expect(func() {
					mn.UpdateLatency(5*time.Second, "TEST_ROUTE")
				}).To(Panic())
			})
		})

		When("Advanced Metrics are enabled", func() {
			BeforeEach(func() {
				advancedMetricsConfig = config.AdvancedMetrics{
					Enabled: true,
					RouteConfig: config.RouteConfiguration{
						RequestCountRoutes:   []string{"TEST_ROUTE", "TEST_ROUTE_1"},
						RequestLatencyRoutes: []string{"TEST_ROUTE", "TEST_ROUTE_2"},
					},
				}
			})

			It("should emit advanced metrics", func() {
				mn.IncrementRequestCounter(1, "TEST_ROUTE")
				mn.UpdateLatency(5*time.Second, "TEST_ROUTE")
				fakeClock.WaitForWatcherAndIncrement(reportInterval)

				Eventually(func() uint64 {
					metricsLock.Lock()
					defer metricsLock.Unlock()
					return counterMap["RequestCount.TEST_ROUTE"]
				}).Should(Equal(uint64(1)))

				Eventually(func() time.Duration {
					metricsLock.Lock()
					defer metricsLock.Unlock()
					return durationMap["RequestLatency.TEST_ROUTE"]
				}).Should(Equal(5 * time.Second))

				mn.IncrementRequestCounter(1, "TEST_ROUTE")
				mn.UpdateLatency(3*time.Second, "TEST_ROUTE")

				mn.IncrementRequestCounter(1, "TEST_ROUTE_1")
				mn.UpdateLatency(2*time.Second, "TEST_ROUTE_1")
				fakeClock.WaitForWatcherAndIncrement(reportInterval)

				Eventually(func() uint64 {
					metricsLock.Lock()
					defer metricsLock.Unlock()
					return counterMap["RequestCount.TEST_ROUTE"]
				}).Should(Equal(uint64(1)))

				Eventually(func() time.Duration {
					metricsLock.Lock()
					defer metricsLock.Unlock()
					return durationMap["RequestLatency.TEST_ROUTE"]
				}).Should(Equal(3 * time.Second))

				Eventually(func() uint64 {
					metricsLock.Lock()
					defer metricsLock.Unlock()
					return counterMap["RequestCount.TEST_ROUTE_1"]
				}).Should(Equal(uint64(1)))
			})

			It("should not emit advanced metrics for routes not in the config", func() {
				mn.IncrementRequestCounter(1, "TEST_ROUTE_2")
				mn.UpdateLatency(5*time.Second, "TEST_ROUTE_2")
				fakeClock.WaitForWatcherAndIncrement(reportInterval)

				Consistently(func() bool {
					metricsLock.Lock()
					defer metricsLock.Unlock()
					_, exists := durationMap["RequestLatency.TEST_ROUTE_1"]

					return exists
				}).Should(BeFalse())

				Consistently(func() bool {
					metricsLock.Lock()
					defer metricsLock.Unlock()
					_, exists := counterMap["RequestCount.TEST_ROUTE_2"]

					return exists
				}).Should(BeFalse())
			})
		})
	})
})
