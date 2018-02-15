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
