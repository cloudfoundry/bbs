package metrics_test

import (
	"time"

	"code.cloudfoundry.org/bbs/metrics"
	"code.cloudfoundry.org/clock/fakeclock"
	"code.cloudfoundry.org/diego-logging-client/testhelpers"
	"code.cloudfoundry.org/go-loggregator"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("LRPStatMetronNotifier", func() {
	type metric struct {
		Name  string
		Value int
	}

	type counter struct {
		Name  string
		Value uint64
	}

	var (
		metricsCh    chan metric
		counterCh    chan counter
		fakeClock    *fakeclock.FakeClock
		metronClient *testhelpers.FakeIngressClient

		notifier metrics.LRPStatMetronNotifier

		process ifrit.Process
	)

	BeforeEach(func() {
		metricsCh = make(chan metric, 100)
		counterCh = make(chan counter, 100)

		fakeClock = fakeclock.NewFakeClock(time.Now())
		metronClient = new(testhelpers.FakeIngressClient)
		metronClient.SendMetricStub = func(name string, value int, opts ...loggregator.EmitGaugeOption) error {
			metricsCh <- metric{name, value}
			return nil
		}
		metronClient.IncrementCounterWithDeltaStub = func(name string, value uint64) error {
			counterCh <- counter{name, value}
			return nil
		}

		notifier = metrics.NewLRPStatMetronNotifier(fakeClock, metronClient)
		Expect(notifier).NotTo(BeNil())

		process = ginkgomon.Invoke(notifier)
	})

	AfterEach(func() {
		ginkgomon.Kill(process)
	})

	Describe("convergence runs metrics", func() {
		Context("when metrics are emitted for the first time and convergence has happened", func() {
			BeforeEach(func() {
				notifier.RecordLRPConvergenceDuration(time.Second)
				notifier.RecordLRPConvergenceDuration(3 * time.Second)

				fakeClock.Increment(metrics.DefaultEmitMetricsFrequency)
			})

			It("emits the number of convergence runs", func() {
				Eventually(counterCh).Should(Receive(Equal(counter{
					Name:  "ConvergenceLRPRuns",
					Value: 2,
				})))
			})

			It("emits the duration of the most recent convergence run", func() {
				Eventually(metronClient.SendDurationCallCount).Should(Equal(1))

				metricName, duration, _ := metronClient.SendDurationArgsForCall(0)
				Expect(metricName).To(Equal("ConvergenceLRPDuration"))
				Expect(duration).To(BeEquivalentTo(3 * time.Second))
			})

			Context("when metrics are emitted a second time and convergence has happened again", func() {
				BeforeEach(func() {
					// wait for previous set of metrics to be emitted then emit the next set
					Eventually(counterCh).Should(Receive())

					notifier.RecordLRPConvergenceDuration(5 * time.Second)

					fakeClock.Increment(metrics.DefaultEmitMetricsFrequency)
				})

				It("emits the number of runs between the first and second metric emissions", func() {
					Eventually(counterCh).Should(Receive(Equal(counter{
						Name:  "ConvergenceLRPRuns",
						Value: 1,
					})))
				})

				It("emits the duration of the most recent convergence run", func() {
					Eventually(metronClient.SendDurationCallCount).Should(Equal(2))

					metricName, duration, _ := metronClient.SendDurationArgsForCall(1)
					Expect(metricName).To(Equal("ConvergenceLRPDuration"))
					Expect(duration).To(BeEquivalentTo(5 * time.Second))
				})
			})

			Context("when metrics are emitted a second time and convergence has NOT happened again", func() {
				BeforeEach(func() {
					Eventually(counterCh).Should(Receive())

					fakeClock.Increment(metrics.DefaultEmitMetricsFrequency)
				})

				It("does not emit convergence runs", func() {
					Consistently(counterCh).ShouldNot(Receive())
				})

				It("emits the cached duration of the last convergence run", func() {
					Eventually(metronClient.SendDurationCallCount).Should(Equal(2))

					metricName, duration, _ := metronClient.SendDurationArgsForCall(1)
					Expect(metricName).To(Equal("ConvergenceLRPDuration"))
					Expect(duration).To(BeEquivalentTo(3 * time.Second))
				})
			})
		})
	})

	Describe("all other metrics", func() {
		BeforeEach(func() {
			notifier.RecordFreshDomains([]string{"domain-1", "domain-2"})
			notifier.RecordStateOfLRPs(1, 2, 3, 4, 5)
			notifier.RecordDesiredLRPs(6)
			notifier.RecordMissingLRPs(7)
			notifier.RecordExtraLRPs(8)

			fakeClock.Increment(metrics.DefaultEmitMetricsFrequency)
		})

		Context("when metrics are emitted for the first time and convergence has happened", func() {
			It("emits metrics", func() {
				Eventually(metricsCh).Should(Receive(Equal(metric{
					Name:  "Domain.domain-1",
					Value: 1,
				})))
				Eventually(metricsCh).Should(Receive(Equal(metric{
					Name:  "Domain.domain-2",
					Value: 1,
				})))

				Eventually(metricsCh).Should(Receive(Equal(metric{
					Name:  "LRPsUnclaimed",
					Value: 1,
				})))
				Eventually(metricsCh).Should(Receive(Equal(metric{
					Name:  "LRPsClaimed",
					Value: 2,
				})))
				Eventually(metricsCh).Should(Receive(Equal(metric{
					Name:  "LRPsRunning",
					Value: 3,
				})))
				Eventually(metricsCh).Should(Receive(Equal(metric{
					Name:  "CrashedActualLRPs",
					Value: 4,
				})))
				Eventually(metricsCh).Should(Receive(Equal(metric{
					Name:  "CrashingDesiredLRPs",
					Value: 5,
				})))
				Eventually(metricsCh).Should(Receive(Equal(metric{
					Name:  "LRPsDesired",
					Value: 6,
				})))
				Eventually(metricsCh).Should(Receive(Equal(metric{
					Name:  "LRPsMissing",
					Value: 7,
				})))
				Eventually(metricsCh).Should(Receive(Equal(metric{
					Name:  "LRPsExtra",
					Value: 8,
				})))
			})
		})

		Context("when metrics are emitted a second time and convergence has happened again", func() {
			BeforeEach(func() {
				notifier.RecordFreshDomains([]string{"domain-11", "domain-12"})
				notifier.RecordStateOfLRPs(11, 12, 13, 14, 15)
				notifier.RecordDesiredLRPs(16)
				notifier.RecordMissingLRPs(17)
				notifier.RecordExtraLRPs(18)

				fakeClock.Increment(metrics.DefaultEmitMetricsFrequency)
			})

			It("emits the most recent values of these metrics", func() {
				Eventually(metricsCh).Should(Receive(Equal(metric{
					Name:  "Domain.domain-11",
					Value: 1,
				})))
				Eventually(metricsCh).Should(Receive(Equal(metric{
					Name:  "Domain.domain-12",
					Value: 1,
				})))

				Eventually(metricsCh).Should(Receive(Equal(metric{
					Name:  "LRPsUnclaimed",
					Value: 11,
				})))
				Eventually(metricsCh).Should(Receive(Equal(metric{
					Name:  "LRPsClaimed",
					Value: 12,
				})))
				Eventually(metricsCh).Should(Receive(Equal(metric{
					Name:  "LRPsRunning",
					Value: 13,
				})))
				Eventually(metricsCh).Should(Receive(Equal(metric{
					Name:  "CrashedActualLRPs",
					Value: 14,
				})))
				Eventually(metricsCh).Should(Receive(Equal(metric{
					Name:  "CrashingDesiredLRPs",
					Value: 15,
				})))
				Eventually(metricsCh).Should(Receive(Equal(metric{
					Name:  "LRPsDesired",
					Value: 16,
				})))
				Eventually(metricsCh).Should(Receive(Equal(metric{
					Name:  "LRPsMissing",
					Value: 17,
				})))
				Eventually(metricsCh).Should(Receive(Equal(metric{
					Name:  "LRPsExtra",
					Value: 18,
				})))
			})
		})

		Context("when metrics are emitted a second time and convergence has NOT happened", func() {
			BeforeEach(func() {
				fakeClock.Increment(metrics.DefaultEmitMetricsFrequency)
			})

			It("emits the cached values of these metrics", func() {
				Eventually(metricsCh).Should(Receive(Equal(metric{
					Name:  "Domain.domain-1",
					Value: 1,
				})))
				Eventually(metricsCh).Should(Receive(Equal(metric{
					Name:  "Domain.domain-2",
					Value: 1,
				})))

				Eventually(metricsCh).Should(Receive(Equal(metric{
					Name:  "LRPsUnclaimed",
					Value: 1,
				})))
				Eventually(metricsCh).Should(Receive(Equal(metric{
					Name:  "LRPsClaimed",
					Value: 2,
				})))
				Eventually(metricsCh).Should(Receive(Equal(metric{
					Name:  "LRPsRunning",
					Value: 3,
				})))
				Eventually(metricsCh).Should(Receive(Equal(metric{
					Name:  "CrashedActualLRPs",
					Value: 4,
				})))
				Eventually(metricsCh).Should(Receive(Equal(metric{
					Name:  "CrashingDesiredLRPs",
					Value: 5,
				})))
				Eventually(metricsCh).Should(Receive(Equal(metric{
					Name:  "LRPsDesired",
					Value: 6,
				})))
				Eventually(metricsCh).Should(Receive(Equal(metric{
					Name:  "LRPsMissing",
					Value: 7,
				})))
				Eventually(metricsCh).Should(Receive(Equal(metric{
					Name:  "LRPsExtra",
					Value: 8,
				})))
			})
		})
	})
})
