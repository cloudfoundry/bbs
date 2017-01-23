package lock_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/bbs/lock"
	"code.cloudfoundry.org/bbs/lock/lockfakes"
	"code.cloudfoundry.org/clock/fakeclock"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/locket"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("Lock", func() {
	var (
		logger *lagertest.TestLogger

		fakeLock  *lockfakes.FakeLock
		fakeClock *fakeclock.FakeClock

		lockKey           string
		lockRetryInterval time.Duration

		lockRunner  ifrit.Runner
		lockProcess ifrit.Process
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("lock")

		fakeLock = &lockfakes.FakeLock{}
		fakeClock = fakeclock.NewFakeClock(time.Now())

		lockRetryInterval = locket.RetryInterval
		lockKey = "test"

		lockRunner = lock.NewLockRunner(
			logger,
			fakeLock,
			lockKey,
			fakeClock,
			lockRetryInterval,
		)
	})

	JustBeforeEach(func() {
		lockProcess = ginkgomon.Invoke(lockRunner)
	})

	AfterEach(func() {
		ginkgomon.Kill(lockProcess)
	})

	It("locks the key", func() {
		Eventually(fakeLock.LockCallCount).Should(Equal(1))
		_, key := fakeLock.LockArgsForCall(0)
		Expect(key).To(Equal(lockKey))
	})

	Context("when the lock cannot be acquired", func() {
		BeforeEach(func() {
			fakeLock.LockReturns(errors.New("no-lock-for-you"))
		})

		It("retries locking after the lock retry interval", func() {
			Eventually(fakeLock.LockCallCount).Should(Equal(1))
			_, key := fakeLock.LockArgsForCall(0)
			Expect(key).To(Equal(lockKey))

			fakeClock.WaitForWatcherAndIncrement(lockRetryInterval)

			Eventually(fakeLock.LockCallCount).Should(Equal(2))
			_, key = fakeLock.LockArgsForCall(1)
			Expect(key).To(Equal(lockKey))
		})

		Context("and the lock becomes available", func() {
			It("stops retrying to grab the lock", func() {
				Eventually(fakeLock.LockCallCount).Should(Equal(1))
				_, key := fakeLock.LockArgsForCall(0)
				Expect(key).To(Equal(lockKey))

				fakeLock.LockReturns(nil)
				fakeClock.WaitForWatcherAndIncrement(lockRetryInterval)

				Eventually(fakeLock.LockCallCount).Should(Equal(2))
				_, key = fakeLock.LockArgsForCall(1)
				Expect(key).To(Equal(lockKey))

				Consistently(fakeClock.WatcherCount()).Should(Equal(0))
				fakeClock.Increment(lockRetryInterval)
				Consistently(fakeLock.LockCallCount).Should(Equal(2))
			})
		})
	})

	Context("when the lock can be acquired", func() {
		It("grabs the lock and then stops trying to grab it", func() {
			Eventually(fakeLock.LockCallCount).Should(Equal(1))
			_, key := fakeLock.LockArgsForCall(0)
			Expect(key).To(Equal(lockKey))

			Consistently(fakeClock.WatcherCount()).Should(Equal(0))
			fakeClock.Increment(lockRetryInterval)
			Consistently(fakeLock.LockCallCount).Should(Equal(1))
		})
	})

	Context("when the lock process receives a signal", func() {
		It("releases the lock", func() {
			ginkgomon.Interrupt(lockProcess)
			Eventually(fakeLock.ReleaseCallCount).Should(Equal(1))
			_, key := fakeLock.ReleaseArgsForCall(0)
			Expect(key).To(Equal(lockKey))
		})
	})
})
