package main_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/tedsuo/ifrit"
	ginkgomon "github.com/tedsuo/ifrit/ginkgomon_v2"

	bbs "code.cloudfoundry.org/bbs/cmd/bbs"
	"code.cloudfoundry.org/bbs/db/dbfakes"
	"code.cloudfoundry.org/clock/fakeclock"
	"code.cloudfoundry.org/lager/v3"
	"code.cloudfoundry.org/lager/v3/lagertest"
)

var _ = Describe("DBHealthCheckRunner", func() {
	var (
		fakeClock  *fakeclock.FakeClock
		fakeDB     *dbfakes.FakeBBSHealthCheckDB
		fakeLogger *lagertest.TestLogger
		runner     *bbs.DBHealthCheckRunner
		process    ifrit.Process
		readyChan  chan struct{}
	)
	BeforeEach(func() {
		readyChan = make(chan struct{})
		fakeClock = fakeclock.NewFakeClock(time.Now())
		fakeLogger = lagertest.NewTestLogger("test")
		fakeDB = &dbfakes.FakeBBSHealthCheckDB{}
		readyChan = make(chan struct{})
		close(readyChan)
		runner = bbs.NewDBHealthCheckRunner(fakeLogger, fakeDB, fakeClock, 4, 100*time.Millisecond, 1, readyChan)
	})
	JustBeforeEach(func() {
		process = ginkgomon.Invoke(runner)
	})
	AfterEach(func() {
		ginkgomon.Kill(process)
	})

	Context("when using empty values for healthcheck settings", func() {
		BeforeEach(func() {
			runner = bbs.NewDBHealthCheckRunner(fakeLogger, fakeDB, fakeClock, 0, 0, 0, readyChan)
		})
		It("sets default values", func() {
			Expect(runner.HealthCheckFailureThreshold).To(Equal(3))
			Expect(runner.HealthCheckTimeout).To(Equal(5 * time.Second))
			Expect(runner.HealthCheckInterval).To(Equal(10 * time.Second))
		})
		It("queries PerformBBSHealthCheck() every interval", func() {
			callCount := fakeDB.PerformBBSHealthCheckCallCount()
			fakeClock.IncrementBySeconds(11)
			Eventually(fakeDB.PerformBBSHealthCheckCallCount).Should(BeNumerically(">", callCount))
			fakeClock.IncrementBySeconds(11)
			Eventually(fakeDB.PerformBBSHealthCheckCallCount).Should(BeNumerically(">", callCount+1))
		})
	})

	It("queries PerformBBSHealthCheck() every interval", func() {
		callCount := fakeDB.PerformBBSHealthCheckCallCount()
		fakeClock.IncrementBySeconds(2)
		Eventually(fakeDB.PerformBBSHealthCheckCallCount).Should(BeNumerically(">", callCount))
		fakeClock.IncrementBySeconds(2)
		Eventually(fakeDB.PerformBBSHealthCheckCallCount).Should(BeNumerically(">", callCount+1))
	})
	Context("when signaled", func() {
		It("exits without an error", func() {
			ginkgomon.Interrupt(process)
			waitCh := process.Wait()
			Eventually(func(g Gomega) {
				err := <-waitCh
				g.Expect(err).ToNot(HaveOccurred())
			}).Should(Succeed())
			Consistently(fakeLogger).ShouldNot(gbytes.Say("catastrophic-database-failure-detected"))
		})
	})
	Context("ExecuteTimedHealthCheckWithRetries()", func() {
		Context("when PerformBBSHealthCheck() fails", func() {
			Context("the entire time", func() {
				BeforeEach(func() {
					fakeDB.PerformBBSHealthCheckReturns(fmt.Errorf("meow"))
				})
				It("returns an error", func() {
					fakeClock.IncrementBySeconds(2)
					waitCh := process.Wait()
					Eventually(waitCh).Should(Receive(MatchError("meow\nmeow\nmeow\nmeow")))
					Eventually(fakeLogger).Should(gbytes.Say("catastrophic-database-failure-detected"))
					Expect(fakeDB.PerformBBSHealthCheckCallCount()).To(Equal(4))

				})
			})
			Context("only twice and then succeeds", func() {
				BeforeEach(func() {
					fakeDB.PerformBBSHealthCheckReturnsOnCall(0, fmt.Errorf("meow"))
					fakeDB.PerformBBSHealthCheckReturnsOnCall(1, fmt.Errorf("meow"))
					fakeDB.PerformBBSHealthCheckReturnsOnCall(2, nil)
				})
				It("returns a success", func() {
					fakeClock.IncrementBySeconds(15)
					waitCh := process.Wait()
					Consistently(waitCh).ShouldNot(Receive())
					Eventually(fakeLogger).Should(gbytes.Say("health-check-succeeded"))
				})
			})
		})
	})
	Describe("ExecuteTimedHealthCheck()", func() {
		Context("when PerformBBSHealthCheck returns an error", func() {
			BeforeEach(func() {
				fakeDB.PerformBBSHealthCheckReturns(fmt.Errorf("meow"))
			})
			It("fails", func() {
				err := runner.ExecuteTimedHealthCheck()
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("meow"))
			})
		})
		Context("when PerformBBSHealthCheck succeeds", func() {
			It("succeeds", func() {
				err := runner.ExecuteTimedHealthCheck()
				Expect(err).NotTo(HaveOccurred())
			})
		})
		Context("when PerformBBSHealthCheck takes > the timeout", func() {
			BeforeEach(func() {
				fakeDB.PerformBBSHealthCheckCalls(func(ctx context.Context, logger lager.Logger, t time.Time) error {
					fakeClock.Increment(200 * time.Millisecond)
					time.Sleep(200 * time.Millisecond)
					return nil
				})
			})

			It("returns an error", func() {
				err := runner.ExecuteTimedHealthCheck()
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("timed out after 100ms while executing DB health check"))

			})
			Context("when using the default timeout", func() {
				BeforeEach(func() {
					runner = bbs.NewDBHealthCheckRunner(fakeLogger, fakeDB, fakeClock, 0, 0, 0, readyChan)
					fakeDB.PerformBBSHealthCheckCalls(func(ctx context.Context, logger lager.Logger, t time.Time) error {
						fakeClock.Increment(6 * time.Second)
						time.Sleep(200 * time.Millisecond)
						return nil
					})
				})
				It("returns an error", func() {
					err := runner.ExecuteTimedHealthCheck()
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("timed out after 5s while executing DB health check"))
				})
			})
		})
	})
})
