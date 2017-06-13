package metrics_test

import (
	"os"
	"time"

	"code.cloudfoundry.org/bbs/metrics"
	"code.cloudfoundry.org/clock/fakeclock"
	"code.cloudfoundry.org/lager/lagertest"
	"github.com/cloudfoundry/dropsonde/metric_sender/fake"
	dropsonde_metrics "github.com/cloudfoundry/dropsonde/metrics"
	"github.com/tedsuo/ifrit"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PeriodicMetronCountNotifier", func() {
	var (
		sender *fake.FakeMetricSender

		reportInterval time.Duration
		fakeClock      *fakeclock.FakeClock

		mn  *metrics.RequestStatMetronNotifier
		mnp ifrit.Process
	)

	BeforeEach(func() {
		reportInterval = 100 * time.Millisecond

		fakeClock = fakeclock.NewFakeClock(time.Unix(123, 456))

		sender = fake.NewFakeMetricSender()
		dropsonde_metrics.Initialize(sender, nil)
	})

	JustBeforeEach(func() {
		ticker := fakeClock.NewTicker(reportInterval)
		mn = metrics.NewRequestStatMetronNotifier(lagertest.NewTestLogger("test"), ticker)
		mnp = ifrit.Invoke(mn)
	})

	AfterEach(func() {
		mnp.Signal(os.Interrupt)
		Eventually(mnp.Wait(), 2*time.Second).Should(Receive())
	})

	It("should emit a request count event periodically", func() {
		mn.IncrementCounter(1)
		mn.UpdateLatency(time.Second)
		fakeClock.WaitForWatcherAndIncrement(reportInterval)

		Eventually(func() uint64 {
			return sender.GetCounter("RequestCount")
		}).Should(Equal(uint64(1)))

		Eventually(func() fake.Metric {
			return sender.GetValue("RequestLatency")
		}).Should(Equal(fake.Metric{Value: float64(1 * time.Second), Unit: "nanos"}))

		mn.IncrementCounter(1)
		mn.UpdateLatency(3 * time.Second)

		mn.IncrementCounter(1)
		mn.UpdateLatency(2 * time.Second)
		fakeClock.WaitForWatcherAndIncrement(reportInterval)

		Eventually(func() uint64 {
			return sender.GetCounter("RequestCount")
		}).Should(Equal(uint64(3)))

		Eventually(func() fake.Metric {
			return sender.GetValue("RequestLatency")
		}).Should(Equal(fake.Metric{Value: float64(3 * time.Second), Unit: "nanos"}))
	})
})
