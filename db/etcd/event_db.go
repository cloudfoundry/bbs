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

type DesiredEventCache map[string]DesiredComponents

type DesiredComponents struct {
	*models.DesiredLRPSchedulingInfo
	*models.DesiredLRPRunInfo
}

func NewDesiredEventCache() DesiredEventCache {
	return DesiredEventCache{}
}

func (d DesiredEventCache) AddSchedulingInfo(logger lager.Logger, schedulingInfo *models.DesiredLRPSchedulingInfo) (*models.DesiredLRP, bool) {
	logger.Info("adding-scheduling-info", lager.Data{"process-guid": schedulingInfo.ProcessGuid})
	components, exists := d[schedulingInfo.ProcessGuid]
	if !exists {
		components = DesiredComponents{}
	}
	components.DesiredLRPSchedulingInfo = schedulingInfo
	if !exists {
		d[schedulingInfo.ProcessGuid] = components
	}

	if components.DesiredLRPRunInfo != nil {
		desiredLRP := models.NewDesiredLRP(*components.DesiredLRPSchedulingInfo, *components.DesiredLRPRunInfo)
		delete(d, schedulingInfo.ProcessGuid)
		return &desiredLRP, true
	}

	return nil, false
}

func (d DesiredEventCache) AddRunInfo(logger lager.Logger, runInfo *models.DesiredLRPRunInfo) (*models.DesiredLRP, bool) {
	logger.Info("adding-run-info", lager.Data{"process-guid": runInfo.ProcessGuid})
	components, exists := d[runInfo.ProcessGuid]
	if !exists {
		components = DesiredComponents{}
	}
	components.DesiredLRPRunInfo = runInfo
	if !exists {
		d[runInfo.ProcessGuid] = components
	}
	if components.DesiredLRPSchedulingInfo != nil {
		desiredLRP := models.NewDesiredLRP(*components.DesiredLRPSchedulingInfo, *components.DesiredLRPRunInfo)
		delete(d, runInfo.ProcessGuid)
		return &desiredLRP, true
	}

	return nil, false
}

func (db *ETCDDB) handleDesiredLRPSchedulingInfoEvent(
	logger lager.Logger,
	event watchEvent,
	created func(*models.DesiredLRP),
	changed func(*models.DesiredLRPChange),
	deleted func(*models.DesiredLRP),
	createsEventCache,
	deletesEventCache DesiredEventCache,
) {
	logger = logger.Session("scheduling-info")
	switch {
	case event.Node != nil && event.PrevNode == nil:
		logger.Debug("received-create")

		schedulingInfo := new(models.DesiredLRPSchedulingInfo)
		err := db.deserializeModel(logger, event.Node, schedulingInfo)
		if err != nil {
			logger.Error("failed-to-unmarshal-desired-lrp-scheduling-info", err)
			return
		}

		desiredLRP, complete := createsEventCache.AddSchedulingInfo(logger, schedulingInfo)
		if complete {
			logger.Debug("sending-create", lager.Data{"process-guid": schedulingInfo.ProcessGuid})
			created(desiredLRP)
		}

	case event.Node != nil && event.PrevNode != nil: // update
		logger.Debug("received-update")

		beforeSchedulingInfo := new(models.DesiredLRPSchedulingInfo)
		err := db.deserializeModel(logger, event.PrevNode, beforeSchedulingInfo)
		if err != nil {
			logger.Error("failed-to-unmarshal-desired-lrp-scheduling-info", err)
			return
		}

		afterSchedulingInfo := new(models.DesiredLRPSchedulingInfo)
		err = db.deserializeModel(logger, event.Node, afterSchedulingInfo)
		if err != nil {
			logger.Error("failed-to-unmarshal-desired-lrp-scheduling-info", err)
			return
		}

		runInfo, err := db.rawDesiredLRPRunInfo(logger, beforeSchedulingInfo.ProcessGuid)
		if err != nil {
			logger.Error("failed-to-fetch-run-info", err, lager.Data{"process-guid": beforeSchedulingInfo.ProcessGuid})
			return
		}

		before := models.NewDesiredLRP(*beforeSchedulingInfo, *runInfo)
		after := models.NewDesiredLRP(*afterSchedulingInfo, *runInfo)

		changed(&models.DesiredLRPChange{Before: &before, After: &after})

	case event.Node == nil && event.PrevNode != nil: // delete
		logger.Debug("received-delete")

		schedulingInfo := new(models.DesiredLRPSchedulingInfo)
		err := db.deserializeModel(logger, event.PrevNode, schedulingInfo)
		if err != nil {
			logger.Error("failed-to-unmarshal-desired-lrp-scheduling-info", err)
			return
		}

		logger.Debug("sending-delete", lager.Data{"process-guid": schedulingInfo.ProcessGuid})
		desiredLRP, complete := deletesEventCache.AddSchedulingInfo(logger, schedulingInfo)
		if complete {
			deleted(desiredLRP)
		}

	default:
		logger.Debug("received-event-with-both-nodes-nil")
	}
}

func (db *ETCDDB) handleDesiredLRPRunInfoEvent(
	logger lager.Logger,
	event watchEvent,
	created func(*models.DesiredLRP),
	deleted func(*models.DesiredLRP),
	createsEventCache,
	deletesEventCache DesiredEventCache,
) {
	logger = logger.Session("run-info")
	switch {
	case event.Node != nil && event.PrevNode == nil:
		logger.Debug("received-create")

		runInfo := new(models.DesiredLRPRunInfo)
		err := db.deserializeModel(logger, event.Node, runInfo)
		if err != nil {
			logger.Error("failed-to-unmarshal-desired-lrp-run-info", err)
			return
		}

		desiredLRP, complete := createsEventCache.AddRunInfo(logger, runInfo)
		if complete {
			logger.Debug("sending-create", lager.Data{"process-guid": runInfo.ProcessGuid})
			created(desiredLRP)
		}

	case event.Node == nil && event.PrevNode != nil: // delete
		logger.Debug("received-delete")

		runInfo := new(models.DesiredLRPRunInfo)
		err := db.deserializeModel(logger, event.PrevNode, runInfo)
		if err != nil {
			logger.Error("failed-to-unmarshal-desired-lrp-run-info", err)
			return
		}

		logger.Debug("sending-delete", lager.Data{"process-guid": runInfo.ProcessGuid})
		desiredLRP, complete := deletesEventCache.AddRunInfo(logger, runInfo)
		if complete {
			deleted(desiredLRP)
		}

	default:
		logger.Debug("received-event-with-both-nodes-nil")
	}
}

func (db *ETCDDB) WatchForDesiredLRPChanges(logger lager.Logger,
	created func(*models.DesiredLRP),
	changed func(*models.DesiredLRPChange),
	deleted func(*models.DesiredLRP),
) (chan<- bool, <-chan error) {
	logger = logger.Session("watching-for-desired-lrp-changes")

	createsEventCache := NewDesiredEventCache()
	deletesEventCache := NewDesiredEventCache()

	schedInfoEvents, stop, err := db.watch(DesiredLRPSchedulingInfoSchemaRoot)
	runInfoEvents, stop, err := db.watch(DesiredLRPRunInfoSchemaRoot)

	go func() {
		for schedInfoEvents != nil && runInfoEvents != nil {
			select {
			case event, ok := <-schedInfoEvents:
				if !ok {
					schedInfoEvents = nil
					break
				}

				db.handleDesiredLRPSchedulingInfoEvent(
					logger,
					event,
					created,
					changed,
					deleted,
					createsEventCache,
					deletesEventCache,
				)

			case event, ok := <-runInfoEvents:
				if !ok {
					runInfoEvents = nil
					break
				}

				db.handleDesiredLRPRunInfoEvent(
					logger,
					event,
					created,
					deleted,
					createsEventCache,
					deletesEventCache,
				)
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

				actualLRP := new(models.ActualLRP)
				err := db.deserializeModel(logger, event.Node, actualLRP)
				if err != nil {
					logger.Error("failed-to-unmarshal-actual-lrp-on-create", err, lager.Data{"key": event.Node.Key, "value": event.Node.Value})
					continue
				}

				evacuating := isEvacuatingActualLRPNode(event.Node)
				actualLRPGroup := &models.ActualLRPGroup{}
				if evacuating {
					actualLRPGroup.Evacuating = actualLRP
				} else {
					actualLRPGroup.Instance = actualLRP
				}

				logger.Debug("sending-create", lager.Data{"actual-lrp": &actualLRP, "evacuating": evacuating})
				created(actualLRPGroup)

			case event.Node != nil && event.PrevNode != nil:
				logger.Debug("received-change")

				before := new(models.ActualLRP)
				err := db.deserializeModel(logger, event.PrevNode, before)
				if err != nil {
					logger.Error("failed-to-unmarshal-prev-actual-lrp-on-change", err, lager.Data{"key": event.PrevNode.Key, "value": event.PrevNode.Value})
					continue
				}

				after := new(models.ActualLRP)
				err = db.deserializeModel(logger, event.Node, after)
				if err != nil {
					logger.Error("failed-to-unmarshal-actual-lrp-on-change", err, lager.Data{"key": event.Node.Key, "value": event.Node.Value})
					continue
				}

				evacuating := isEvacuatingActualLRPNode(event.Node)
				beforeGroup := &models.ActualLRPGroup{}
				afterGroup := &models.ActualLRPGroup{}
				if evacuating {
					afterGroup.Evacuating = after
					beforeGroup.Evacuating = before
				} else {
					afterGroup.Instance = after
					beforeGroup.Instance = before
				}

				logger.Debug("sending-change", lager.Data{"before": before, "after": after, "evacuating": evacuating})
				changed(&models.ActualLRPChange{Before: beforeGroup, After: afterGroup})

			case event.PrevNode != nil && event.Node == nil:
				logger.Debug("received-delete")
				if event.PrevNode.Dir {
					continue
				}

				actualLRP := new(models.ActualLRP)
				err := db.deserializeModel(logger, event.PrevNode, actualLRP)
				if err != nil {
					logger.Error("failed-to-unmarshal-prev-actual-lrp-on-delete", err, lager.Data{"key": event.PrevNode.Key, "value": event.PrevNode.Value})
				} else {
					evacuating := isEvacuatingActualLRPNode(event.PrevNode)
					actualLRPGroup := &models.ActualLRPGroup{}
					if evacuating {
						actualLRPGroup.Evacuating = actualLRP
					} else {
						actualLRPGroup.Instance = actualLRP
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
