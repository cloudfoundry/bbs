package etcd

import (
	"fmt"
	"time"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/coreos/go-etcd/etcd"
	"github.com/pivotal-golang/lager"
)

type watchEvent struct {
	Type     int
	Node     *etcd.Node
	PrevNode *etcd.Node
}

const (
	invalidEvent = iota
	createEvent
	deleteEvent
	expireEvent
	updateEvent
)

func (db *ETCDDB) WatchForDesiredLRPChanges(logger lager.Logger,
	created func(*models.DesiredLRP),
	changed func(*models.DesiredLRPChange),
	deleted func(*models.DesiredLRP),
) (chan<- bool, <-chan error) {
	logger = logger.Session("watching-for-desired-lrp-changes")

	events, stop, err := db.watch(DesiredLRPSchemaRoot)

	go func() {
		for event := range events {
			switch {
			case event.Node != nil && event.PrevNode == nil:
				logger.Debug("received-create")

				var desiredLRP models.DesiredLRP
				err := models.FromJSON([]byte(event.Node.Value), &desiredLRP)
				if err != nil {
					logger.Error("failed-to-unmarshal-desired-lrp", err, lager.Data{"value": event.Node.Value})
					continue
				}

				logger.Debug("sending-create", lager.Data{"desired-lrp": &desiredLRP})
				created(&desiredLRP)

			case event.Node != nil && event.PrevNode != nil: // update
				logger.Debug("received-update")

				var before models.DesiredLRP
				err := models.FromJSON([]byte(event.PrevNode.Value), &before)
				if err != nil {
					logger.Error("failed-to-unmarshal-desired-lrp", err, lager.Data{"value": event.PrevNode.Value})
					continue
				}

				var after models.DesiredLRP
				err = models.FromJSON([]byte(event.Node.Value), &after)
				if err != nil {
					logger.Error("failed-to-unmarshal-desired-lrp", err, lager.Data{"value": event.Node.Value})
					continue
				}

				logger.Debug("sending-update", lager.Data{"before": &before, "after": &after})
				changed(&models.DesiredLRPChange{Before: &before, After: &after})

			case event.Node == nil && event.PrevNode != nil: // delete
				logger.Debug("received-delete")

				var desiredLRP models.DesiredLRP
				err := models.FromJSON([]byte(event.PrevNode.Value), &desiredLRP)
				if err != nil {
					logger.Error("failed-to-unmarshal-desired-lrp", err, lager.Data{"value": event.PrevNode.Value})
					continue
				}

				logger.Debug("sending-delete", lager.Data{"desired-lrp": &desiredLRP})
				deleted(&desiredLRP)

			default:
				logger.Debug("received-event-with-both-nodes-nil")
			}
		}
	}()

	return stop, err
}

func (db *ETCDDB) WatchForActualLRPChanges(logger lager.Logger,
	created func(*models.ActualLRPGroup),
	changed func(*models.ActualLRPChange),
	deleted func(*models.ActualLRPGroup),
) (chan<- bool, <-chan error) {
	logger = logger.Session("watching-for-actual-lrp-changes")

	events, stop, err := db.watch(ActualLRPSchemaRoot)

	go func() {
		logger.Info("started-watching")
		defer logger.Info("finished-watching")

		for event := range events {
			switch {
			case event.Node != nil && event.PrevNode == nil:
				logger.Debug("received-create")

				var actualLRP models.ActualLRP
				err := models.FromJSON([]byte(event.Node.Value), &actualLRP)
				if err != nil {
					logger.Error("failed-to-unmarshal-actual-lrp-on-create", err, lager.Data{"key": event.Node.Key, "value": event.Node.Value})
					continue
				}

				evacuating := isEvacuatingActualLRPNode(event.Node)
				actualLRPGroup := &models.ActualLRPGroup{}
				if evacuating {
					actualLRPGroup.Evacuating = &actualLRP
				} else {
					actualLRPGroup.Instance = &actualLRP
				}

				logger.Debug("sending-create", lager.Data{"actual-lrp": &actualLRP, "evacuating": evacuating})
				created(actualLRPGroup)

			case event.Node != nil && event.PrevNode != nil:
				logger.Debug("received-change")

				var before models.ActualLRP
				err := models.FromJSON([]byte(event.PrevNode.Value), &before)
				if err != nil {
					logger.Error("failed-to-unmarshal-prev-actual-lrp-on-change", err, lager.Data{"key": event.PrevNode.Key, "value": event.PrevNode.Value})
					continue
				}

				var after models.ActualLRP
				err = models.FromJSON([]byte(event.Node.Value), &after)
				if err != nil {
					logger.Error("failed-to-unmarshal-actual-lrp-on-change", err, lager.Data{"key": event.Node.Key, "value": event.Node.Value})
					continue
				}

				evacuating := isEvacuatingActualLRPNode(event.Node)
				beforeGroup := &models.ActualLRPGroup{}
				afterGroup := &models.ActualLRPGroup{}
				if evacuating {
					afterGroup.Evacuating = &after
					beforeGroup.Evacuating = &before
				} else {
					afterGroup.Instance = &after
					beforeGroup.Instance = &before
				}

				logger.Debug("sending-change", lager.Data{"before": &before, "after": &after, "evacuating": evacuating})
				changed(&models.ActualLRPChange{Before: beforeGroup, After: afterGroup})

			case event.PrevNode != nil && event.Node == nil:
				logger.Debug("received-delete")

				var actualLRP models.ActualLRP
				if event.PrevNode.Dir {
					continue
				}
				err := models.FromJSON([]byte(event.PrevNode.Value), &actualLRP)
				if err != nil {
					logger.Error("failed-to-unmarshal-prev-actual-lrp-on-delete", err, lager.Data{"key": event.PrevNode.Key, "value": event.PrevNode.Value})
				} else {
					evacuating := isEvacuatingActualLRPNode(event.PrevNode)
					actualLRPGroup := &models.ActualLRPGroup{}
					if evacuating {
						actualLRPGroup.Evacuating = &actualLRP
					} else {
						actualLRPGroup.Instance = &actualLRP
					}

					logger.Debug("sending-delete", lager.Data{"actual-lrp": &actualLRP, "evacuating": evacuating})
					deleted(actualLRPGroup)
				}

			default:
				logger.Debug("received-event-with-both-nodes-nil")
			}
		}
	}()

	return stop, err
}

func (db *ETCDDB) watch(key string) (<-chan watchEvent, chan<- bool, <-chan error) {
	events := make(chan watchEvent)
	errors := make(chan error)
	stop := make(chan bool, 1)

	go db.dispatchWatchEvents(key, events, stop, errors)

	time.Sleep(100 * time.Millisecond) //give the watcher a chance to connect

	return events, stop, errors
}

func (db *ETCDDB) dispatchWatchEvents(key string, events chan<- watchEvent, stop chan bool, errors chan<- error) {
	var index uint64
	db.registerInflightWatch(stop)

	defer close(events)
	defer close(errors)
	defer db.unregisterInflightWatch(stop)

	for {
		response, err := db.client.Watch(key, index, true, nil, stop)
		if err != nil {
			if etcdErrCode(err) == ETCDErrIndexCleared {
				index = 0
				continue
			} else if err == etcd.ErrWatchStoppedByUser {
				return
			} else {
				errors <- err
				return
			}
		}

		event, err := db.makeWatchEvent(response)
		if err != nil {
			errors <- err
			return
		} else {
			events <- event
		}

		index = response.Node.ModifiedIndex + 1
	}
}

func (db *ETCDDB) registerInflightWatch(stop chan bool) {
	db.inflightWatchLock.Lock()
	defer db.inflightWatchLock.Unlock()
	db.inflightWatches[stop] = true
}

func (db *ETCDDB) unregisterInflightWatch(stop chan bool) {
	db.inflightWatchLock.Lock()
	defer db.inflightWatchLock.Unlock()
	delete(db.inflightWatches, stop)
}

func (db *ETCDDB) cancelInflightWatches() {
	db.inflightWatchLock.Lock()
	defer db.inflightWatchLock.Unlock()

	for stop := range db.inflightWatches {
		select {
		case _, ok := <-stop:
			if ok {
				close(stop)
			}
		default:
			close(stop)
		}
	}
}

func (db *ETCDDB) makeWatchEvent(event *etcd.Response) (watchEvent, error) {
	var eventType int

	node := event.Node
	switch event.Action {
	case "delete", "compareAndDelete":
		eventType = deleteEvent
		node = nil
	case "create":
		eventType = createEvent
	case "set", "update", "compareAndSwap":
		eventType = updateEvent
	case "expire":
		eventType = expireEvent
		node = nil
	default:
		return watchEvent{}, fmt.Errorf("unknown event: %s", event.Action)
	}

	return watchEvent{
		Type:     eventType,
		Node:     node,
		PrevNode: event.PrevNode,
	}, nil
}
