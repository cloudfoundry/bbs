package watcher_test

import (
	"errors"
	"os"
	"time"

	"github.com/cloudfoundry-incubator/bbs/db/fakes"
	"github.com/cloudfoundry-incubator/bbs/events/eventfakes"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/watcher"
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/clock/fakeclock"
	"github.com/pivotal-golang/lager/lagertest"
	"github.com/tedsuo/ifrit"
)

var _ = Describe("Watcher", func() {
	const (
		expectedProcessGuid  = "some-process-guid"
		expectedInstanceGuid = "some-instance-guid"
		retryWaitDuration    = 50 * time.Millisecond
	)

	var (
		db         *fakes.FakeEventDB
		hub        *eventfakes.FakeHub
		clock      *fakeclock.FakeClock
		bbsWatcher watcher.Watcher
		process    ifrit.Process

		desiredLRPStop   chan bool
		desiredLRPErrors chan error

		actualLRPStop   chan bool
		actualLRPErrors chan error
	)

	BeforeEach(func() {
		db = new(fakes.FakeEventDB)
		hub = new(eventfakes.FakeHub)
		clock = fakeclock.NewFakeClock(time.Now())
		logger := lagertest.NewTestLogger("test")

		desiredLRPStop = make(chan bool, 1)
		desiredLRPErrors = make(chan error)

		actualLRPStop = make(chan bool, 1)
		actualLRPErrors = make(chan error)

		db.WatchForDesiredLRPChangesReturns(desiredLRPStop, desiredLRPErrors)
		db.WatchForActualLRPChangesReturns(actualLRPStop, actualLRPErrors)

		bbsWatcher = watcher.NewWatcher(db, hub, clock, retryWaitDuration, logger)
	})

	AfterEach(func() {
		process.Signal(os.Interrupt)
		Eventually(process.Wait()).Should(Receive())
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
				Consistently(db.WatchForDesiredLRPChangesCallCount).Should(BeZero())
				Consistently(db.WatchForActualLRPChangesCallCount).Should(BeZero())
			})

			Context("and then the hub reports a subscriber", func() {
				var callback func(int)

				BeforeEach(func() {
					Expect(hub.RegisterCallbackCallCount()).To(Equal(1))
					callback = hub.RegisterCallbackArgsForCall(0)
					callback(1)
				})

				It("requests watches", func() {
					Eventually(db.WatchForDesiredLRPChangesCallCount).Should(Equal(1))
					Eventually(db.WatchForActualLRPChangesCallCount).Should(Equal(1))
				})

				Context("and then the hub reports two subscribers", func() {
					BeforeEach(func() {
						callback(2)
					})

					It("does not request more watches", func() {
						Eventually(db.WatchForDesiredLRPChangesCallCount).Should(Equal(1))
						Consistently(db.WatchForDesiredLRPChangesCallCount).Should(Equal(1))

						Eventually(db.WatchForActualLRPChangesCallCount).Should(Equal(1))
						Consistently(db.WatchForActualLRPChangesCallCount).Should(Equal(1))
					})
				})

				Context("and then the hub reports no subscribers", func() {
					BeforeEach(func() {
						callback(0)
					})

					It("stops the watches", func() {
						Eventually(desiredLRPStop).Should(Receive())
						Eventually(actualLRPStop).Should(Receive())
					})
				})

				Context("when the desired watch reports an error", func() {
					BeforeEach(func() {
						desiredLRPErrors <- errors.New("oh no!")
					})

					It("requests a new desired watch after the retry interval", func() {
						clock.Increment(retryWaitDuration / 2)
						Consistently(db.WatchForDesiredLRPChangesCallCount).Should(Equal(1))
						clock.Increment(retryWaitDuration * 2)
						Eventually(db.WatchForDesiredLRPChangesCallCount).Should(Equal(2))
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
							Consistently(db.WatchForDesiredLRPChangesCallCount).Should(Equal(1))
						})
					})
				})

				Context("when the actual watch reports an error", func() {
					BeforeEach(func() {
						actualLRPErrors <- errors.New("oh no!")
					})

					It("requests a new actual watch after the retry interval", func() {
						clock.Increment(retryWaitDuration / 2)
						Consistently(db.WatchForActualLRPChangesCallCount).Should(Equal(1))
						clock.Increment(retryWaitDuration * 2)
						Eventually(db.WatchForActualLRPChangesCallCount).Should(Equal(2))
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
							Consistently(db.WatchForActualLRPChangesCallCount).Should(Equal(1))
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
				Eventually(db.WatchForDesiredLRPChangesCallCount).Should(Equal(1))
				Eventually(db.WatchForActualLRPChangesCallCount).Should(Equal(1))
			})

			Context("and then the watcher is signaled to stop", func() {
				It("stops the watches", func() {
					process.Signal(os.Interrupt)
					Eventually(desiredLRPStop).Should(Receive())
					Eventually(actualLRPStop).Should(Receive())
					Eventually(process.Wait()).Should(Receive())
				})
			})

			Context("when the watcher receives several desired watch errors in a retry interval", func() {
				It("uses only one active timer", func() {
					Expect(hub.RegisterCallbackCallCount()).To(Equal(1))
					callback := hub.RegisterCallbackArgsForCall(0)

					Eventually(db.WatchForDesiredLRPChangesCallCount).Should(Equal(1))

					desiredLRPErrors <- errors.New("first error")

					callback(1)

					Eventually(db.WatchForDesiredLRPChangesCallCount).Should(Equal(2))
					desiredLRPErrors <- errors.New("second error")

					Consistently(clock.WatcherCount).Should(Equal(1))
				})
			})

			Context("when the watcher receives several actual watch errors in a retry interval", func() {
				It("uses only one active timer", func() {
					Expect(hub.RegisterCallbackCallCount()).To(Equal(1))
					callback := hub.RegisterCallbackArgsForCall(0)

					Eventually(db.WatchForActualLRPChangesCallCount).Should(Equal(1))

					actualLRPErrors <- errors.New("first error")

					callback(1)

					Eventually(db.WatchForActualLRPChangesCallCount).Should(Equal(2))
					actualLRPErrors <- errors.New("second error")

					Consistently(clock.WatcherCount).Should(Equal(1))
				})
			})
		})
	})

	Describe("when watching the bbs", func() {
		var (
			desiredCreateCB func(*models.DesiredLRP)
			desiredChangeCB func(*models.DesiredLRPChange)
			desiredDeleteCB func(*models.DesiredLRP)
			actualCreateCB  func(*models.ActualLRPGroup)
			actualChangeCB  func(*models.ActualLRPChange)
			actualDeleteCB  func(*models.ActualLRPGroup)
		)

		BeforeEach(func() {
			hub.RegisterCallbackStub = func(cb func(int)) {
				cb(1)
			}
			process = ifrit.Invoke(bbsWatcher)
			Eventually(db.WatchForDesiredLRPChangesCallCount).Should(Equal(1))
			Eventually(db.WatchForActualLRPChangesCallCount).Should(Equal(1))

			_, desiredCreateCB, desiredChangeCB, desiredDeleteCB = db.WatchForDesiredLRPChangesArgsForCall(0)
			_, actualCreateCB, actualChangeCB, actualDeleteCB = db.WatchForActualLRPChangesArgsForCall(0)
		})

		Describe("Desired LRP changes", func() {
			var desiredLRP *models.DesiredLRP

			BeforeEach(func() {
				desiredLRP = &models.DesiredLRP{
					Action: models.WrapAction(&models.RunAction{
						User: proto.String("me"),
						Path: proto.String("ls"),
					}),
					Domain:      proto.String("tests"),
					ProcessGuid: proto.String(expectedProcessGuid),
				}
			})

			Context("when a create arrives", func() {
				BeforeEach(func() {
					desiredCreateCB(desiredLRP)
				})

				It("emits a DesiredLRPCreatedEvent to the hub", func() {
					Expect(hub.EmitCallCount()).To(Equal(1))
					event := hub.EmitArgsForCall(0)

					desiredLRPCreatedEvent, ok := event.(*models.DesiredLRPCreatedEvent)
					Expect(ok).To(BeTrue())
					Expect(desiredLRPCreatedEvent.DesiredLrp).To(Equal(desiredLRP))
				})
			})

			Context("when a change arrives", func() {
				BeforeEach(func() {
					desiredChangeCB(&models.DesiredLRPChange{Before: desiredLRP, After: desiredLRP})
				})

				It("emits a DesiredLRPChangedEvent to the hub", func() {
					Expect(hub.EmitCallCount()).To(Equal(1))
					event := hub.EmitArgsForCall(0)

					desiredLRPChangedEvent, ok := event.(*models.DesiredLRPChangedEvent)
					Expect(ok).To(BeTrue())
					Expect(desiredLRPChangedEvent.Before).To(Equal(desiredLRP))
					Expect(desiredLRPChangedEvent.After).To(Equal(desiredLRP))
				})
			})

			Context("when a delete arrives", func() {
				BeforeEach(func() {
					desiredDeleteCB(desiredLRP)
				})

				It("emits a DesiredLRPRemovedEvent to the hub", func() {
					Expect(hub.EmitCallCount()).To(Equal(1))
					event := hub.EmitArgsForCall(0)

					desiredLRPRemovedEvent, ok := event.(*models.DesiredLRPRemovedEvent)
					Expect(ok).To(BeTrue())
					Expect(desiredLRPRemovedEvent.DesiredLrp).To(Equal(desiredLRP))
				})
			})
		})

		Describe("Actual LRP changes", func() {
			var actualLRPGroup *models.ActualLRPGroup

			BeforeEach(func() {
				actualLRPGroup = &models.ActualLRPGroup{
					Instance: &models.ActualLRP{
						ActualLRPKey:         models.NewActualLRPKey(expectedProcessGuid, 1, "domain"),
						ActualLRPInstanceKey: models.NewActualLRPInstanceKey(expectedInstanceGuid, "cell-id"),
					},
				}
			})

			Context("when a create arrives", func() {
				BeforeEach(func() {
					actualCreateCB(actualLRPGroup)
				})

				It("emits an ActualLRPCreatedEvent to the hub", func() {
					Expect(hub.EmitCallCount()).To(Equal(1))
					event := hub.EmitArgsForCall(0)
					Expect(event).To(BeAssignableToTypeOf(&models.ActualLRPCreatedEvent{}))

					actualLRPCreatedEvent := event.(*models.ActualLRPCreatedEvent)
					Expect(actualLRPCreatedEvent.ActualLrpGroup).To(Equal(actualLRPGroup))
				})
			})

			Context("when a change arrives", func() {
				BeforeEach(func() {
					actualChangeCB(&models.ActualLRPChange{Before: actualLRPGroup, After: actualLRPGroup})
				})

				It("emits an ActualLRPChangedEvent to the hub", func() {
					Expect(hub.EmitCallCount()).To(Equal(1))
					event := hub.EmitArgsForCall(0)
					Expect(event).To(BeAssignableToTypeOf(&models.ActualLRPChangedEvent{}))

					actualLRPChangedEvent := event.(*models.ActualLRPChangedEvent)
					Expect(actualLRPChangedEvent.Before).To(Equal(actualLRPGroup))
					Expect(actualLRPChangedEvent.After).To(Equal(actualLRPGroup))
				})
			})

			Context("when a delete arrives", func() {
				BeforeEach(func() {
					actualDeleteCB(actualLRPGroup)
				})

				It("emits an ActualLRPRemovedEvent to the hub", func() {
					Expect(hub.EmitCallCount()).To(Equal(1))
					event := hub.EmitArgsForCall(0)
					Expect(event).To(BeAssignableToTypeOf(&models.ActualLRPRemovedEvent{}))

					actualLRPRemovedEvent := event.(*models.ActualLRPRemovedEvent)
					Expect(actualLRPRemovedEvent.ActualLrpGroup).To(Equal(actualLRPGroup))
				})
			})
		})
	})
})
