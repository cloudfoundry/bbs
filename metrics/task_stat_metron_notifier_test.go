package metrics_test

import (
	"time"

	"code.cloudfoundry.org/bbs/metrics"
	"code.cloudfoundry.org/clock/fakeclock"
	"code.cloudfoundry.org/diego-logging-client/testhelpers"
	"code.cloudfoundry.org/go-loggregator"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
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
		Value uint64
	}

	var (
		taskStatMetronNotifier metrics.TaskStatMetronNotifier
		fakeClock              *fakeclock.FakeClock
		metronClient           *testhelpers.FakeIngressClient
		metricsCh              chan metric
		counterCh              chan counter
		process                ifrit.Process
	)

	BeforeEach(func() {
		metricsCh = make(chan metric, 100)
		counterCh = make(chan counter, 100)

		metronClient = &testhelpers.FakeIngressClient{}
		metronClient.SendMetricStub = func(name string, value int, opts ...loggregator.EmitGaugeOption) error {
			metricsCh <- metric{name, value, opts}
			return nil
		}
		metronClient.IncrementCounterWithDeltaStub = func(name string, value uint64) error {
			counterCh <- counter{name, value}
			return nil
		}

		fakeClock = fakeclock.NewFakeClock(time.Now())
		taskStatMetronNotifier = metrics.NewTaskStatMetronNotifier(fakeClock, metronClient)
		Expect(taskStatMetronNotifier).NotTo(BeNil())

		process = ginkgomon.Invoke(taskStatMetronNotifier)
	})

	AfterEach(func() {
		ginkgomon.Kill(process)
	})

	Context("when a task is started", func() {
		BeforeEach(func() {
			taskStatMetronNotifier.TaskStarted("cell-1")
			fakeClock.Increment(60 * time.Second)
		})

		It("emits the metric", func() {
			Eventually(metricsCh).Should(Receive(gstruct.MatchAllFields(gstruct.Fields{
				"Name":  Equal("TasksStarted"),
				"Value": Equal(1),
				"Opts":  haveTag("cell-id", "cell-1"),
			})))
		})

		Context("when metrics were emitted and another 60 seconds have elapsed", func() {
			BeforeEach(func() {
				Eventually(metricsCh).Should(Receive())
				fakeClock.Increment(60 * time.Second)
			})

			It("emits the same metric again", func() {
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
			fakeClock.Increment(60 * time.Second)
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
			fakeClock.Increment(60 * time.Second)
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
			fakeClock.Increment(60 * time.Second)
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

			taskStatMetronNotifier.TaskConvergenceDuration(5 * time.Second)
			taskStatMetronNotifier.TaskConvergenceDuration(1 * time.Second)

			fakeClock.Increment(60 * time.Second)
		})

		It("emits the number of convergence runs since the last time metrics were emitted", func() {
			Eventually(counterCh).Should(Receive(Equal(counter{
				Name:  "ConvergenceTaskRuns",
				Value: 3,
			})))
		})

		It("emits the duration of the last convergence run", func() {
			Eventually(metronClient.SendDurationCallCount).Should(Equal(1))

			metricName, duration, _ := metronClient.SendDurationArgsForCall(0)
			Expect(metricName).To(Equal(metrics.ConvergeTaskDuration))
			Expect(duration).To(BeEquivalentTo(1 * time.Second))
		})

		Context("after metrics have been emitted, and then another convergence loop starts", func() {
			BeforeEach(func() {
				// wait for previous set of metrics to be emitted then emit the next set
				Eventually(counterCh).Should(Receive())

				taskStatMetronNotifier.TaskConvergenceStarted()
				taskStatMetronNotifier.TaskConvergenceDuration(2 * time.Second)

				fakeClock.Increment(60 * time.Second)
			})

			It("resets the value and emits the number of runs since the last time metrics were emitted", func() {
				Eventually(counterCh).Should(Receive(Equal(counter{
					Name:  "ConvergenceTaskRuns",
					Value: 1,
				})))
			})

			It("emits the duration of the new convergence run", func() {
				Eventually(metronClient.SendDurationCallCount).Should(Equal(2))

				metricName, duration, _ := metronClient.SendDurationArgsForCall(1)
				Expect(metricName).To(Equal(metrics.ConvergeTaskDuration))
				Expect(duration).To(BeEquivalentTo(2 * time.Second))
			})
		})

		Context("after metrics have been emitted, but another convergence loop has not yet started", func() {
			BeforeEach(func() {
				Eventually(counterCh).Should(Receive())
				fakeClock.Increment(60 * time.Second)
			})

			It("doesn't update the converge runs counter", func() {
				Consistently(counterCh).ShouldNot(Receive(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"Name": Equal("TasksFailed"),
				})))
			})

			It("emits the duration of the last convergence run", func() {
				Eventually(metronClient.SendDurationCallCount).Should(Equal(2))

				metricName, duration, _ := metronClient.SendDurationArgsForCall(1)
				Expect(metricName).To(Equal(metrics.ConvergeTaskDuration))
				Expect(duration).To(BeEquivalentTo(1 * time.Second))
			})
		})
	})

	Describe("metrics about tasks resulting from convergence", func() {
		BeforeEach(func() {
			taskStatMetronNotifier.SnapshotTaskStats(1, 2, 3, 4, 5, 6)
			fakeClock.Increment(60 * time.Second)
		})

		It("emits the number of pending, running, completed, resolving, pruned, and kicked tasks", func() {
			Eventually(metricsCh).Should(Receive(Equal(metric{
				Name:  "TasksPending",
				Value: 1,
			})))

			Eventually(metricsCh).Should(Receive(Equal(metric{
				Name:  "TasksRunning",
				Value: 2,
			})))

			Eventually(metricsCh).Should(Receive(Equal(metric{
				Name:  "TasksCompleted",
				Value: 3,
			})))

			Eventually(metricsCh).Should(Receive(Equal(metric{
				Name:  "TasksResolving",
				Value: 4,
			})))

			Eventually(counterCh).Should(Receive(Equal(counter{
				Name:  "ConvergenceTasksPruned",
				Value: uint64(5),
			})))

			Eventually(counterCh).Should(Receive(Equal(counter{
				Name:  "ConvergenceTasksKicked",
				Value: uint64(6),
			})))
		})

		Context("after 60 seconds have elapsed", func() {
			Context("and task state metrics have been updated", func() {
				BeforeEach(func() {
					taskStatMetronNotifier.SnapshotTaskStats(5, 6, 7, 8, 9, 10)
					fakeClock.Increment(60 * time.Second)
				})

				It("emits the new value for the metric", func() {
					Eventually(metricsCh).Should(Receive(Equal(metric{
						Name:  "TasksPending",
						Value: 5,
					})))

					Eventually(metricsCh).Should(Receive(Equal(metric{
						Name:  "TasksRunning",
						Value: 6,
					})))

					Eventually(metricsCh).Should(Receive(Equal(metric{
						Name:  "TasksCompleted",
						Value: 7,
					})))

					Eventually(metricsCh).Should(Receive(Equal(metric{
						Name:  "TasksResolving",
						Value: 8,
					})))

					Eventually(counterCh).Should(Receive(Equal(counter{
						Name:  "ConvergenceTasksPruned",
						Value: uint64(9),
					})))

					Eventually(counterCh).Should(Receive(Equal(counter{
						Name:  "ConvergenceTasksKicked",
						Value: uint64(10),
					})))
				})
			})

			Context("and task state metrics have not been updated", func() {
				BeforeEach(func() {
					fakeClock.Increment(60 * time.Second)
				})

				It("emits the last value of the metric", func() {
					Eventually(metricsCh).Should(Receive(Equal(metric{
						Name:  "TasksPending",
						Value: 1,
					})))

					Eventually(metricsCh).Should(Receive(Equal(metric{
						Name:  "TasksRunning",
						Value: 2,
					})))

					Eventually(metricsCh).Should(Receive(Equal(metric{
						Name:  "TasksCompleted",
						Value: 3,
					})))

					Eventually(metricsCh).Should(Receive(Equal(metric{
						Name:  "TasksResolving",
						Value: 4,
					})))

					Eventually(counterCh).Should(Receive(Equal(counter{
						Name:  "ConvergenceTasksPruned",
						Value: uint64(5),
					})))

					Eventually(counterCh).Should(Receive(Equal(counter{
						Name:  "ConvergenceTasksKicked",
						Value: uint64(6),
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
