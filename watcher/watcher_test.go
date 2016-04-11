package watcher_test

import (
	"errors"
	"os"
	"time"

	"github.com/cloudfoundry-incubator/bbs/db/fakes"
	"github.com/cloudfoundry-incubator/bbs/events/eventfakes"
	"github.com/cloudfoundry-incubator/bbs/watcher"
	"github.com/cloudfoundry-incubator/bbs/watcher/watcherfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/clock/fakeclock"
	"github.com/pivotal-golang/lager/lagertest"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("Watcher", func() {
	const (
		retryWaitDuration = 50 * time.Millisecond
	)

	var (
		eventDB    *fakes.FakeEventDB
		hub        *eventfakes.FakeHub
		clock      *fakeclock.FakeClock
		streamer   *watcherfakes.FakeEventStreamer
		bbsWatcher watcher.Watcher
		process    ifrit.Process

		stopChan  chan bool
		errorChan chan error
	)

	BeforeEach(func() {
		logger := lagertest.NewTestLogger("test")

		eventDB = new(fakes.FakeEventDB)
		hub = new(eventfakes.FakeHub)
		clock = fakeclock.NewFakeClock(time.Now())

		stopChan = make(chan bool, 1)
		errorChan = make(chan error)
		streamer = new(watcherfakes.FakeEventStreamer)
		streamer.StreamReturns(stopChan, errorChan)

		bbsWatcher = watcher.NewWatcher(logger, "test", retryWaitDuration, streamer, hub, clock)
	})

	AfterEach(func() {
		ginkgomon.Interrupt(process)
	})

	Describe("starting", func() {
		Context("when the hub initially reports no subscribers", func() {
			BeforeEach(func() {
				hub.RegisterCallbackStub = func(cb func(int)) {
					cb(0)
				}
				process = ifrit.Invoke(bbsWatcher)
			})

			It("does not request a watch", func() {
				Consistently(streamer.StreamCallCount).Should(BeZero())
			})

			Context("and then the hub reports a subscriber", func() {
				var callback func(int)

				BeforeEach(func() {
					Expect(hub.RegisterCallbackCallCount()).To(Equal(1))
					callback = hub.RegisterCallbackArgsForCall(0)
					callback(1)
				})

				It("requests watches", func() {
					Eventually(streamer.StreamCallCount).Should(Equal(1))
				})

				Context("and then the hub reports two subscribers", func() {
					BeforeEach(func() {
						callback(2)
					})

					It("does not request more watches", func() {
						Eventually(streamer.StreamCallCount).Should(Equal(1))
						Consistently(streamer.StreamCallCount).Should(Equal(1))
					})
				})

				Context("and then the hub reports no subscribers", func() {
					BeforeEach(func() {
						callback(0)
					})

					It("stops the watches", func() {
						Eventually(stopChan).Should(Receive())
					})
				})

				Context("when the desired watch reports an error", func() {
					BeforeEach(func() {
						errorChan <- errors.New("oh no!")
					})

					It("requests a new desired watch after the retry interval", func() {
						clock.WaitForWatcherAndIncrement(retryWaitDuration / 2)
						Eventually(streamer.StreamCallCount).Should(Equal(1))
						clock.WaitForWatcherAndIncrement(retryWaitDuration * 2)
						Eventually(streamer.StreamCallCount).Should(Equal(2))
					})

					Context("and the hub reports no subscribers before the retry interval elapses", func() {
						BeforeEach(func() {
							clock.Increment(retryWaitDuration / 2)
							callback(0)
							// give watcher time to clear out event loop
							time.Sleep(10 * time.Millisecond)
						})

						It("does not request new watches", func() {
							clock.Increment(retryWaitDuration * 2)
							Consistently(streamer.StreamCallCount).Should(Equal(1))
						})
					})
				})
			})
		})

		Context("when the hub initially reports a subscriber", func() {
			BeforeEach(func() {
				hub.RegisterCallbackStub = func(cb func(int)) {
					cb(1)
				}
				process = ifrit.Invoke(bbsWatcher)
			})

			It("requests watches", func() {
				Eventually(streamer.StreamCallCount).Should(Equal(1))
			})

			Context("and then the watcher is signaled to stop", func() {
				It("closes the hub", func() {
					process.Signal(os.Interrupt)
					Eventually(hub.CloseCallCount).Should(Equal(1))
					Eventually(process.Wait()).Should(Receive())
				})
			})

			Context("when the watcher receives several desired watch errors in a retry interval", func() {
				It("uses only one active timer", func() {
					Expect(hub.RegisterCallbackCallCount()).To(Equal(1))
					callback := hub.RegisterCallbackArgsForCall(0)

					Eventually(streamer.StreamCallCount).Should(Equal(1))

					errorChan <- errors.New("first error")

					callback(1)

					Eventually(streamer.StreamCallCount).Should(Equal(2))
					errorChan <- errors.New("second error")

					Consistently(clock.WatcherCount).Should(Equal(1))
				})
			})
		})
	})
})
