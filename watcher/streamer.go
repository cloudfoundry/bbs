package watcher

import (
	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter -o watcherfakes/fake_event_streamer.go . EventStreamer

type EventStreamer interface {
	Stream(logger lager.Logger, eventChan chan<- models.Event) (stop chan<- bool, error <-chan error)
}

type DesiredStreamer struct {
	eventDB db.EventDB
}

func NewDesiredStreamer(eventDB db.EventDB) *DesiredStreamer {
	return &DesiredStreamer{
		eventDB: eventDB,
	}
}

func (s *DesiredStreamer) Stream(logger lager.Logger, eventChan chan<- models.Event) (chan<- bool, <-chan error) {
	return s.eventDB.WatchForDesiredLRPChanges(logger,
		func(created *models.DesiredLRP) {
			logger.Debug("handling-desired-create")
			eventChan <- models.NewDesiredLRPCreatedEvent(created)
		},
		func(changed *models.DesiredLRPChange) {
			logger.Debug("handling-desired-change")
			eventChan <- models.NewDesiredLRPChangedEvent(
				changed.Before,
				changed.After,
			)
		},
		func(deleted *models.DesiredLRP) {
			logger.Debug("handling-desired-delete")
			eventChan <- models.NewDesiredLRPRemovedEvent(deleted)
		})
}

type ActualStreamer struct {
	eventDB db.EventDB
}

func NewActualStreamer(eventDB db.EventDB) *ActualStreamer {
	return &ActualStreamer{
		eventDB: eventDB,
	}
}

func (s *ActualStreamer) Stream(logger lager.Logger, eventChan chan<- models.Event) (chan<- bool, <-chan error) {
	return s.eventDB.WatchForActualLRPChanges(logger,
		func(created *models.ActualLRPGroup) {
			logger.Debug("handling-actual-create")
			eventChan <- models.NewActualLRPCreatedEvent(created)
		},
		func(changed *models.ActualLRPChange) {
			logger.Debug("handling-actual-change")
			eventChan <- models.NewActualLRPChangedEvent(
				changed.Before,
				changed.After,
			)
		},
		func(deleted *models.ActualLRPGroup) {
			logger.Debug("handling-actual-delete")
			eventChan <- models.NewActualLRPRemovedEvent(deleted)
		})
}

type TaskStreamer struct {
	eventDB db.EventDB
}

func NewTaskStreamer(eventDB db.EventDB) *TaskStreamer {
	return &TaskStreamer{
		eventDB: eventDB,
	}
}

func (s *TaskStreamer) Stream(logger lager.Logger, eventChan chan<- models.Event) (chan<- bool, <-chan error) {
	return s.eventDB.WatchForTaskChanges(logger,
		func(created *models.Task) {
			logger.Debug("handling-task-create")
			eventChan <- models.NewTaskCreatedEvent(created)
		},
		func(changed *models.TaskChange) {
			logger.Debug("handling-task-change")
			eventChan <- models.NewTaskChangedEvent(
				changed.Before,
				changed.After,
			)
		},
		func(deleted *models.Task) {
			logger.Debug("handling-task-delete")
			eventChan <- models.NewTaskRemovedEvent(deleted)
		})
}
