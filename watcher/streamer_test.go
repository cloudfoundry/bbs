package watcher_test

import (
	"github.com/cloudfoundry-incubator/bbs/db/fakes"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/models/test/model_helpers"
	"github.com/cloudfoundry-incubator/bbs/watcher"
	"github.com/pivotal-golang/lager/lagertest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	expectedProcessGuid  = "some-process-guid"
	expectedInstanceGuid = "some-instance-guid"
)

var _ = Describe("DesiredStreamer", func() {

	var (
		desiredCreateCB func(*models.DesiredLRP)
		desiredChangeCB func(*models.DesiredLRPChange)
		desiredDeleteCB func(*models.DesiredLRP)
		eventDB         *fakes.FakeEventDB
		eventChan       chan models.Event
		streamer        *watcher.DesiredStreamer
	)

	BeforeEach(func() {
		logger := lagertest.NewTestLogger("test")

		eventDB = &fakes.FakeEventDB{}
		streamer = watcher.NewDesiredStreamer(eventDB)

		eventChan = make(chan models.Event, 1)
		streamer.Stream(logger, eventChan)

		Eventually(eventDB.WatchForDesiredLRPChangesCallCount).Should(Equal(1))
		_, desiredCreateCB, desiredChangeCB, desiredDeleteCB = eventDB.WatchForDesiredLRPChangesArgsForCall(0)
	})

	Describe("Desired LRP changes", func() {
		var desiredLRP *models.DesiredLRP

		BeforeEach(func() {
			desiredLRP = &models.DesiredLRP{
				Action: models.WrapAction(&models.RunAction{
					User: "me",
					Path: "ls",
				}),
				Domain:      "tests",
				ProcessGuid: expectedProcessGuid,
			}
		})

		Context("when a create arrives", func() {
			BeforeEach(func() {
				desiredCreateCB(desiredLRP)
			})

			It("emits a DesiredLRPCreatedEvent", func() {
				var event models.Event
				Eventually(eventChan).Should(Receive(&event))

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
				var event models.Event
				Eventually(eventChan).Should(Receive(&event))

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
				var event models.Event
				Eventually(eventChan).Should(Receive(&event))

				desiredLRPRemovedEvent, ok := event.(*models.DesiredLRPRemovedEvent)
				Expect(ok).To(BeTrue())
				Expect(desiredLRPRemovedEvent.DesiredLrp).To(Equal(desiredLRP))
			})
		})
	})

})

var _ = Describe("ActualStreamer", func() {
	var (
		actualCreateCB func(*models.ActualLRPGroup)
		actualChangeCB func(*models.ActualLRPChange)
		actualDeleteCB func(*models.ActualLRPGroup)
		eventDB        *fakes.FakeEventDB
		streamer       *watcher.ActualStreamer
		eventChan      chan models.Event
	)

	BeforeEach(func() {
		logger := lagertest.NewTestLogger("test")

		eventDB = &fakes.FakeEventDB{}
		streamer = watcher.NewActualStreamer(eventDB)

		eventChan = make(chan models.Event, 1)
		streamer.Stream(logger, eventChan)

		Eventually(eventDB.WatchForActualLRPChangesCallCount).Should(Equal(1))
		_, actualCreateCB, actualChangeCB, actualDeleteCB = eventDB.WatchForActualLRPChangesArgsForCall(0)
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
				var event models.Event
				Eventually(eventChan).Should(Receive(&event))
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
				var event models.Event
				Eventually(eventChan).Should(Receive(&event))
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
				var event models.Event
				Eventually(eventChan).Should(Receive(&event))
				Expect(event).To(BeAssignableToTypeOf(&models.ActualLRPRemovedEvent{}))

				actualLRPRemovedEvent := event.(*models.ActualLRPRemovedEvent)
				Expect(actualLRPRemovedEvent.ActualLrpGroup).To(Equal(actualLRPGroup))
			})
		})
	})
})

var _ = Describe("TaskStreamer", func() {
	var (
		taskCreateCB func(*models.Task)
		taskChangeCB func(*models.TaskChange)
		taskDeleteCB func(*models.Task)
		eventDB      *fakes.FakeEventDB
		streamer     *watcher.TaskStreamer
		eventChan    chan models.Event
	)

	BeforeEach(func() {
		logger := lagertest.NewTestLogger("test")

		eventDB = &fakes.FakeEventDB{}
		streamer = watcher.NewTaskStreamer(eventDB)

		eventChan = make(chan models.Event, 1)
		streamer.Stream(logger, eventChan)

		Eventually(eventDB.WatchForTaskChangesCallCount).Should(Equal(1))
		_, taskCreateCB, taskChangeCB, taskDeleteCB = eventDB.WatchForTaskChangesArgsForCall(0)
	})

	Describe("Task changes", func() {
		var task *models.Task

		BeforeEach(func() {
			task = model_helpers.NewValidTask("some-task-guid")
		})

		Context("when a create arrives", func() {
			BeforeEach(func() {
				taskCreateCB(task)
			})

			It("emits an TaskCreatedEvent to the hub", func() {
				var event models.Event
				Eventually(eventChan).Should(Receive(&event))
				Expect(event).To(BeAssignableToTypeOf(&models.TaskCreatedEvent{}))

				taskCreatedEvent := event.(*models.TaskCreatedEvent)
				Expect(taskCreatedEvent.Task).To(Equal(task))
			})
		})

		Context("when a change arrives", func() {
			BeforeEach(func() {
				taskChangeCB(&models.TaskChange{Before: task, After: task})
			})

			It("emits a TaskChangedEvent to the hub", func() {
				var event models.Event
				Eventually(eventChan).Should(Receive(&event))
				Expect(event).To(BeAssignableToTypeOf(&models.TaskChangedEvent{}))

				taskChangedEvent := event.(*models.TaskChangedEvent)
				Expect(taskChangedEvent.Before).To(Equal(task))
				Expect(taskChangedEvent.After).To(Equal(task))
			})
		})

		Context("when a delete arrives", func() {
			BeforeEach(func() {
				taskDeleteCB(task)
			})

			It("emits an ActualLRPRemovedEvent to the hub", func() {
				var event models.Event
				Eventually(eventChan).Should(Receive(&event))
				Expect(event).To(BeAssignableToTypeOf(&models.TaskRemovedEvent{}))

				taskRemovedEvent := event.(*models.TaskRemovedEvent)
				Expect(taskRemovedEvent.Task).To(Equal(task))
			})
		})
	})
})
