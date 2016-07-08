package converger_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/lager/lagertest"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"

	"code.cloudfoundry.org/bbs/fake_bbs"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/clock/fakeclock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/bbs/converger"
	"code.cloudfoundry.org/bbs/converger/fake_handlers"
)

const aBit = 100 * time.Millisecond

var _ = Describe("ConvergerProcess", func() {
	var (
		fakeLrpConvergenceHandler   *fake_handlers.FakeLrpConvergenceHandler
		fakeTaskConvergenceHandler  *fake_handlers.FakeTaskConvergenceHandler
		fakeBBSServiceClient        *fake_bbs.FakeServiceClient
		logger                      *lagertest.TestLogger
		fakeClock                   *fakeclock.FakeClock
		convergeRepeatInterval      time.Duration
		kickTaskDuration            time.Duration
		expirePendingTaskDuration   time.Duration
		expireCompletedTaskDuration time.Duration

		process ifrit.Process

		waitEvents chan<- models.CellEvent
		waitErrs   chan<- error
	)

	BeforeEach(func() {
		fakeLrpConvergenceHandler = new(fake_handlers.FakeLrpConvergenceHandler)
		fakeTaskConvergenceHandler = new(fake_handlers.FakeTaskConvergenceHandler)
		fakeBBSServiceClient = new(fake_bbs.FakeServiceClient)
		logger = lagertest.NewTestLogger("test")
		fakeClock = fakeclock.NewFakeClock(time.Now())

		convergeRepeatInterval = 1 * time.Second

		kickTaskDuration = 10 * time.Millisecond
		expirePendingTaskDuration = 30 * time.Second
		expireCompletedTaskDuration = 60 * time.Minute

		cellEvents := make(chan models.CellEvent, 100)
		errs := make(chan error, 100)

		waitEvents = cellEvents
		waitErrs = errs

		fakeBBSServiceClient.CellEventsReturns(cellEvents)
	})

	JustBeforeEach(func() {
		process = ifrit.Invoke(
			converger.New(
				fakeLrpConvergenceHandler,
				fakeTaskConvergenceHandler,
				fakeBBSServiceClient,
				logger,
				fakeClock,
				convergeRepeatInterval,
				kickTaskDuration,
				expirePendingTaskDuration,
				expireCompletedTaskDuration,
			),
		)
	})

	AfterEach(func() {
		ginkgomon.Interrupt(process)
		Eventually(process.Wait()).Should(Receive())
	})

	Describe("converging over time", func() {
		It("converges tasks, LRPs, and auctions when the lock is periodically reestablished", func() {
			fakeClock.Increment(convergeRepeatInterval + aBit)

			Eventually(fakeTaskConvergenceHandler.ConvergeTasksCallCount, aBit).Should(Equal(1))
			Eventually(fakeLrpConvergenceHandler.ConvergeLRPsCallCount, aBit).Should(Equal(1))

			actualKickTaskDuration, actualExpirePendingTaskDuration, actualExpireCompletedTaskDuration := fakeTaskConvergenceHandler.ConvergeTasksArgsForCall(0)
			Expect(actualKickTaskDuration).To(Equal(kickTaskDuration))
			Expect(actualExpirePendingTaskDuration).To(Equal(expirePendingTaskDuration))
			Expect(actualExpireCompletedTaskDuration).To(Equal(expireCompletedTaskDuration))

			fakeClock.Increment(convergeRepeatInterval + aBit)

			Eventually(fakeTaskConvergenceHandler.ConvergeTasksCallCount, aBit).Should(Equal(2))
			Eventually(fakeLrpConvergenceHandler.ConvergeLRPsCallCount, aBit).Should(Equal(2))

			actualKickTaskDuration, actualExpirePendingTaskDuration, actualExpireCompletedTaskDuration = fakeTaskConvergenceHandler.ConvergeTasksArgsForCall(1)
			Expect(actualKickTaskDuration).To(Equal(kickTaskDuration))
			Expect(actualExpirePendingTaskDuration).To(Equal(expirePendingTaskDuration))
			Expect(actualExpireCompletedTaskDuration).To(Equal(expireCompletedTaskDuration))
		})
	})

	Describe("converging when cells disappear", func() {
		It("converges tasks and LRPs immediately", func() {
			Consistently(fakeTaskConvergenceHandler.ConvergeTasksCallCount).Should(Equal(0))
			Consistently(fakeLrpConvergenceHandler.ConvergeLRPsCallCount).Should(Equal(0))

			waitEvents <- models.CellDisappearedEvent{
				IDs: []string{"some-cell-id"},
			}

			Eventually(fakeTaskConvergenceHandler.ConvergeTasksCallCount, aBit).Should(Equal(1))
			Eventually(fakeLrpConvergenceHandler.ConvergeLRPsCallCount, aBit).Should(Equal(1))

			actualKickTaskDuration, actualExpirePendingTaskDuration, actualExpireCompletedTaskDuration := fakeTaskConvergenceHandler.ConvergeTasksArgsForCall(0)
			Expect(actualKickTaskDuration).To(Equal(kickTaskDuration))
			Expect(actualExpirePendingTaskDuration).To(Equal(expirePendingTaskDuration))
			Expect(actualExpireCompletedTaskDuration).To(Equal(expireCompletedTaskDuration))

			waitErrs <- errors.New("whoopsie")

			waitEvents <- models.CellDisappearedEvent{
				IDs: []string{"some-cell-id"},
			}

			Eventually(fakeTaskConvergenceHandler.ConvergeTasksCallCount, aBit).Should(Equal(2))
			Eventually(fakeLrpConvergenceHandler.ConvergeLRPsCallCount, aBit).Should(Equal(2))
		})

		It("defers convergence to one full interval later", func() {
			fakeClock.Increment(convergeRepeatInterval - aBit)

			waitEvents <- models.CellDisappearedEvent{
				IDs: []string{"some-cell-id"},
			}

			Eventually(fakeTaskConvergenceHandler.ConvergeTasksCallCount, aBit).Should(Equal(1))
			Eventually(fakeLrpConvergenceHandler.ConvergeLRPsCallCount, aBit).Should(Equal(1))

			fakeClock.Increment(2 * aBit)

			Consistently(fakeTaskConvergenceHandler.ConvergeTasksCallCount, aBit).Should(Equal(1))
			Consistently(fakeLrpConvergenceHandler.ConvergeLRPsCallCount, aBit).Should(Equal(1))

			fakeClock.Increment(convergeRepeatInterval + aBit)
			Eventually(fakeTaskConvergenceHandler.ConvergeTasksCallCount, aBit).Should(Equal(2))
			Eventually(fakeLrpConvergenceHandler.ConvergeLRPsCallCount, aBit).Should(Equal(2))
		})
	})
})
