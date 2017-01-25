package lock_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/bbs/lock"
	"code.cloudfoundry.org/bbs/lock/lockfakes"
	"code.cloudfoundry.org/bbs/models"
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

		fakeLocker *lockfakes.FakeLocker
		fakeClock  *fakeclock.FakeClock

		expectedLock      models.Lock
		lockRetryInterval time.Duration

		lockRunner  ifrit.Runner
		lockProcess ifrit.Process
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("lock")

		fakeLocker = &lockfakes.FakeLocker{}
		fakeClock = fakeclock.NewFakeClock(time.Now())

		lockRetryInterval = locket.RetryInterval
		expectedLock = models.Lock{Key: "test", Owner: "jim", Value: "is pretty sweet."}

		lockRunner = lock.NewLockRunner(
			logger,
			fakeLocker,
			expectedLock,
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
		Eventually(fakeLocker.LockCallCount).Should(Equal(1))
		_, lock := fakeLocker.LockArgsForCall(0)
		Expect(lock).To(Equal(expectedLock))
	})

	Context("when the lock cannot be acquired", func() {
		BeforeEach(func() {
			fakeLocker.LockReturns(errors.New("no-lock-for-you"))
		})

		It("retries locking after the lock retry interval", func() {
			Eventually(fakeLocker.LockCallCount).Should(Equal(1))
			_, lock := fakeLocker.LockArgsForCall(0)
			Expect(lock).To(Equal(expectedLock))

			fakeClock.WaitForWatcherAndIncrement(lockRetryInterval)

			Eventually(fakeLocker.LockCallCount).Should(Equal(2))
			_, lock = fakeLocker.LockArgsForCall(1)
			Expect(lock).To(Equal(expectedLock))
		})

		Context("and the lock becomes available", func() {
			It("stops retrying to grab the lock", func() {
				Eventually(fakeLocker.LockCallCount).Should(Equal(1))
				_, lock := fakeLocker.LockArgsForCall(0)
				Expect(lock).To(Equal(expectedLock))

				fakeLocker.LockReturns(nil)
				fakeClock.WaitForWatcherAndIncrement(lockRetryInterval)

				Eventually(fakeLocker.LockCallCount).Should(Equal(2))
				_, lock = fakeLocker.LockArgsForCall(1)
				Expect(lock).To(Equal(expectedLock))

				Consistently(fakeClock.WatcherCount()).Should(Equal(0))
				fakeClock.Increment(lockRetryInterval)
				Consistently(fakeLocker.LockCallCount).Should(Equal(2))
			})
		})
	})

	Context("when the lock can be acquired", func() {
		It("grabs the lock and then stops trying to grab it", func() {
			Eventually(fakeLocker.LockCallCount).Should(Equal(1))
			_, lock := fakeLocker.LockArgsForCall(0)
			Expect(lock).To(Equal(expectedLock))

			Consistently(fakeClock.WatcherCount()).Should(Equal(0))
			fakeClock.Increment(lockRetryInterval)
			Consistently(fakeLocker.LockCallCount).Should(Equal(1))
		})
	})

	Context("when the lock process receives a signal", func() {
		It("releases the lock", func() {
			ginkgomon.Interrupt(lockProcess)
			Eventually(fakeLocker.ReleaseLockCallCount).Should(Equal(1))
			_, lock := fakeLocker.ReleaseLockArgsForCall(0)
			Expect(lock).To(Equal(expectedLock))
		})
	})
})
