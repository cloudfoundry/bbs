package metrics_test

import (
	"time"

	"code.cloudfoundry.org/bbs/metrics"
	"code.cloudfoundry.org/clock/fakeclock"
	"code.cloudfoundry.org/diego-logging-client/testhelpers"
	"code.cloudfoundry.org/go-loggregator"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("TaskStatMetronNotifier", func() {
	type metric struct {
		Name  string
		Value int
		Opts  []loggregator.EmitGaugeOption
	}

	type counter struct {
		Name  string
		Delta uint64
	}

	var (
		taskStatMetronNotifier metrics.TaskStatMetronNotifier
		fakeClock              *fakeclock.FakeClock
		logger                 lager.Logger
		metronClient           *testhelpers.FakeIngressClient
		metricsCh              chan metric
		process                ifrit.Process
	)

	BeforeEach(func() {
		metricsCh = make(chan metric, 100)

		metronClient = &testhelpers.FakeIngressClient{}
		metronClient.SendMetricStub = func(name string, value int, opts ...loggregator.EmitGaugeOption) error {
			metricsCh <- metric{name, value, opts}
			return nil
		}

		logger = lagertest.NewTestLogger("test")
		fakeClock = fakeclock.NewFakeClock(time.Now())
		taskStatMetronNotifier = metrics.NewTaskStatMetronNotifier(logger, fakeClock, metronClient)
		Expect(taskStatMetronNotifier).NotTo(BeNil())
	})

	JustBeforeEach(func() {
		process = ginkgomon.Invoke(taskStatMetronNotifier)
		fakeClock.WaitForWatcherAndIncrement(60 * time.Second)
	})

	AfterEach(func() {
		ginkgomon.Kill(process)
	})

	Context("when a task is started", func() {
		BeforeEach(func() {
			taskStatMetronNotifier.TaskStarted("cell-1")
		})

		It("emits the metric", func() {
			Eventually(metricsCh).Should(Receive(gstruct.MatchAllFields(gstruct.Fields{
				"Name":  Equal("TasksStarted"),
				"Value": Equal(1),
				"Opts":  haveTag("cell-id", "cell-1"),
			})))
		})

		Context("when 60 seconds have elapsed", func() {
			JustBeforeEach(func() {
				fakeClock.WaitForWatcherAndIncrement(60 * time.Second)
			})

			It("emits the metric again", func() {
				Eventually(metricsCh).Should(Receive(gstruct.MatchAllFields(gstruct.Fields{
					"Name":  Equal("TasksStarted"),
					"Value": Equal(1),
					"Opts":  haveTag("cell-id", "cell-1"),
				})))
			})
		})
	})

	Context("when a task succeeds", func() {
		BeforeEach(func() {
			taskStatMetronNotifier.TaskSucceeded("cell-1")
		})

		It("emits the metric with the proper tag", func() {
			Eventually(metricsCh).Should(Receive(gstruct.MatchAllFields(gstruct.Fields{
				"Name":  Equal("TasksSucceeded"),
				"Value": Equal(1),
				"Opts":  haveTag("cell-id", "cell-1"),
			})))
		})
	})

	Context("when a task fails", func() {
		BeforeEach(func() {
			taskStatMetronNotifier.TaskFailed("cell-1")
			taskStatMetronNotifier.TaskFailed("cell-1")
		})

		It("emits the metric with the proper tag", func() {
			Eventually(metricsCh).Should(Receive(gstruct.MatchAllFields(gstruct.Fields{
				"Name":  Equal("TasksFailed"),
				"Value": Equal(2),
				"Opts":  haveTag("cell-id", "cell-1"),
			})))
		})
	})

	Context("when tasks on multiple cells are started", func() {
		BeforeEach(func() {
			taskStatMetronNotifier.TaskFailed("cell-1")
			taskStatMetronNotifier.TaskFailed("cell-2")
		})

		It("emits the metric for the second cell with the proper tag", func() {
			Eventually(metricsCh).Should(Receive(gstruct.MatchAllFields(gstruct.Fields{
				"Name":  Equal("TasksFailed"),
				"Value": Equal(1),
				"Opts":  haveTag("cell-id", "cell-2"),
			})))
		})

		It("emits the metric for the first cell with the proper tag", func() {
			Eventually(metricsCh).Should(Receive(gstruct.MatchAllFields(gstruct.Fields{
				"Name":  Equal("TasksFailed"),
				"Value": Equal(1),
				"Opts":  haveTag("cell-id", "cell-1"),
			})))
		})
	})

	Describe("metrics about convergence", func() {
		BeforeEach(func() {
			taskStatMetronNotifier.TaskConvergenceStarted()
			taskStatMetronNotifier.TaskConvergenceStarted()
			taskStatMetronNotifier.TaskConvergenceStarted()

			taskStatMetronNotifier.TaskConvergenceDuration(1 * time.Second)
		})

		It("emits the number of convergence runs since the last time metrics were emitted", func() {
			fakeClock.WaitForWatcherAndIncrement(0 * time.Second)
			Eventually(metronClient.IncrementCounterWithDeltaCallCount).Should(Equal(1))

			counterName, delta := metronClient.IncrementCounterWithDeltaArgsForCall(0)
			Expect(counterName).To(Equal("ConvergenceTaskRuns"))
			Expect(delta).To(BeEquivalentTo(3))
		})

		It("emits the duration of the last convergence run", func() {
			Eventually(metronClient.SendDurationCallCount).Should(Equal(1))

			metricName, duration, _ := metronClient.SendDurationArgsForCall(0)
			Expect(metricName).To(Equal(metrics.ConvergeTaskDuration))
			Expect(duration).To(BeEquivalentTo(1 * time.Second))
		})

		Context("after metrics have been emitted, and then another convergence loop starts", func() {
			JustBeforeEach(func() {
				// This causes the ticker loop to stop, and gives the test a chance to update the fixtures
				fakeClock.WaitForWatcherAndIncrement(0 * time.Second)
				taskStatMetronNotifier.TaskConvergenceStarted()
				taskStatMetronNotifier.TaskConvergenceDuration(2 * time.Second)

				fakeClock.Increment(60 * time.Second)
			})

			It("resets the value and emits the number of runs since the last time metrics were emitted", func() {
				fakeClock.WaitForWatcherAndIncrement(0 * time.Second)
				Eventually(metronClient.IncrementCounterWithDeltaCallCount).Should(Equal(2))

				counterName, delta := metronClient.IncrementCounterWithDeltaArgsForCall(1)
				Expect(counterName).To(Equal(metrics.ConvergeTaskRunsCounter))
				Expect(delta).To(BeEquivalentTo(1))
			})

			It("emits the duration of the new convergence run", func() {
				Eventually(metronClient.SendDurationCallCount).Should(Equal(2))

				metricName, duration, _ := metronClient.SendDurationArgsForCall(1)
				Expect(metricName).To(Equal(metrics.ConvergeTaskDuration))
				Expect(duration).To(BeEquivalentTo(2 * time.Second))
			})
		})

		Context("after metrics have been emitted, but another convergence loop has not yet started", func() {
			JustBeforeEach(func() {
				fakeClock.WaitForWatcherAndIncrement(60 * time.Second)
			})

			It("doesn't update the converge runs counter", func() {
				fakeClock.WaitForWatcherAndIncrement(0 * time.Second)
				Eventually(metronClient.IncrementCounterWithDeltaCallCount).Should(Equal(1))
			})

			It("emits the duration of the last convergence run", func() {
				fakeClock.WaitForWatcherAndIncrement(0 * time.Second)
				Eventually(metronClient.SendDurationCallCount).Should(Equal(2))

				metricName, duration, _ := metronClient.SendDurationArgsForCall(1)
				Expect(metricName).To(Equal(metrics.ConvergeTaskDuration))
				Expect(duration).To(BeEquivalentTo(1 * time.Second))
			})
		})
	})

	Describe("metrics about tasks resulting from convergence", func() {
		BeforeEach(func() {
			taskStatMetronNotifier.TaskConvergenceResults(1, 2, 3, 4)
		})

		It("emits the number of pending, running, completed, and resolving tasks", func() {
			Eventually(metricsCh).Should(Receive(gstruct.MatchAllFields(gstruct.Fields{
				"Name":  Equal(metrics.PendingTasksMetric),
				"Value": Equal(1),
				"Opts":  BeEmpty(),
			})))

			Eventually(metricsCh).Should(Receive(gstruct.MatchAllFields(gstruct.Fields{
				"Name":  Equal(metrics.RunningTasksMetric),
				"Value": Equal(2),
				"Opts":  BeEmpty(),
			})))

			Eventually(metricsCh).Should(Receive(gstruct.MatchAllFields(gstruct.Fields{
				"Name":  Equal(metrics.CompletedTasksMetric),
				"Value": Equal(3),
				"Opts":  BeEmpty(),
			})))

			Eventually(metricsCh).Should(Receive(gstruct.MatchAllFields(gstruct.Fields{
				"Name":  Equal(metrics.ResolvingTasksMetric),
				"Value": Equal(4),
				"Opts":  BeEmpty(),
			})))
		})

		Context("after 60 seconds have elapsed", func() {
			JustBeforeEach(func() {
				fakeClock.WaitForWatcherAndIncrement(60 * time.Second)
			})

			Context("and a convergence loop has also occurred", func() {
				BeforeEach(func() {
					taskStatMetronNotifier.TaskConvergenceResults(5, 6, 7, 8)
				})

				It("emits the new value for the metric", func() {
					Eventually(metricsCh).Should(Receive(gstruct.MatchAllFields(gstruct.Fields{
						"Name":  Equal(metrics.PendingTasksMetric),
						"Value": Equal(5),
						"Opts":  BeEmpty(),
					})))

					Eventually(metricsCh).Should(Receive(gstruct.MatchAllFields(gstruct.Fields{
						"Name":  Equal(metrics.RunningTasksMetric),
						"Value": Equal(6),
						"Opts":  BeEmpty(),
					})))

					Eventually(metricsCh).Should(Receive(gstruct.MatchAllFields(gstruct.Fields{
						"Name":  Equal(metrics.CompletedTasksMetric),
						"Value": Equal(7),
						"Opts":  BeEmpty(),
					})))

					Eventually(metricsCh).Should(Receive(gstruct.MatchAllFields(gstruct.Fields{
						"Name":  Equal(metrics.ResolvingTasksMetric),
						"Value": Equal(8),
						"Opts":  BeEmpty(),
					})))

				})
			})

			Context("and a convergence loop has not occurred in the meantime", func() {
				It("emits the last value of the metric", func() {
					Eventually(metricsCh).Should(Receive(gstruct.MatchAllFields(gstruct.Fields{
						"Name":  Equal(metrics.PendingTasksMetric),
						"Value": Equal(1),
						"Opts":  BeEmpty(),
					})))

					Eventually(metricsCh).Should(Receive(gstruct.MatchAllFields(gstruct.Fields{
						"Name":  Equal(metrics.RunningTasksMetric),
						"Value": Equal(2),
						"Opts":  BeEmpty(),
					})))

					Eventually(metricsCh).Should(Receive(gstruct.MatchAllFields(gstruct.Fields{
						"Name":  Equal(metrics.CompletedTasksMetric),
						"Value": Equal(3),
						"Opts":  BeEmpty(),
					})))

					Eventually(metricsCh).Should(Receive(gstruct.MatchAllFields(gstruct.Fields{
						"Name":  Equal(metrics.ResolvingTasksMetric),
						"Value": Equal(4),
						"Opts":  BeEmpty(),
					})))

				})
			})
		})
	})
})

func haveTag(name, value string) types.GomegaMatcher {
	return WithTransform(func(opts []loggregator.EmitGaugeOption) map[string]string {
		envelope := &loggregator_v2.Envelope{
			Tags: make(map[string]string),
		}
		for _, opt := range opts {
			opt(envelope)
		}
		return envelope.Tags
	}, Equal(map[string]string{name: value}))
}
