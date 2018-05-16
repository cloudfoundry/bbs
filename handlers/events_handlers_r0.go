package handlers

import (
	"errors"
	"net/http"

	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

func (h *EventHandler) Subscribe_r0(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	logger = logger.Session("subscribe-r0")

	request := &models.EventsByCellId{}
	err := parseRequest(logger, req, request)
	if err != nil {
		logger.Error("failed-parsing-request", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger.Info("subscribed-to-event-stream", lager.Data{"cell_id": request.CellId})

	desiredSource, err := h.desiredHub.Subscribe()
	if err != nil {
		logger.Error("failed-to-subscribe-to-desired-event-hub", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer desiredSource.Close()

	actualSource, err := h.actualHub.Subscribe()
	if err != nil {
		logger.Error("failed-to-subscribe-to-actual-event-hub", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer actualSource.Close()

	eventChan := make(chan models.Event)
	errorChan := make(chan error)
	closeChan := make(chan struct{})
	defer close(closeChan)

	actualEventsFetcher := func() (models.Event, error) {
		for {
			event, err := actualSource.Next()
			if err != nil {
				return nil, err
			}

			// convert flattened to ALRG
			event, err = convertFromFlattened(event)
			if err != nil {
				return nil, err
			}
			if request.CellId != "" && !filterByCellID(request.CellId, event, err) {
				continue
			}
			return event, nil
		}
	}

	desiredEventsFetcher := func() (models.Event, error) {
		event, err := desiredSource.Next()
		if err != nil {
			return event, err
		}
		event = models.VersionDesiredLRPsToV0(event)
		return event, err
	}

	go streamSource(eventChan, errorChan, closeChan, desiredEventsFetcher)
	go streamSource(eventChan, errorChan, closeChan, actualEventsFetcher)

	streamEventsToResponse(logger, w, eventChan, errorChan)
}

func (h *TaskEventHandler) Subscribe_r0(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	logger = logger.Session("tasks-subscribe-r0")
	logger.Info("subscribed-to-tasks-event-stream")

	taskSource, err := h.taskHub.Subscribe()
	if err != nil {
		logger.Error("failed-to-subscribe-to-task-event-hub", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer taskSource.Close()

	eventChan := make(chan models.Event)
	errorChan := make(chan error)
	closeChan := make(chan struct{})
	defer close(closeChan)

	go streamSource(eventChan, errorChan, closeChan, taskSource.Next)

	streamEventsToResponse(logger, w, eventChan, errorChan)
}

func filterByCellID(cellID string, bbsEvent models.Event, err error) bool {
	switch x := bbsEvent.(type) {
	case *models.ActualLRPCreatedEvent:
		lrp, _ := x.ActualLrpGroup.Resolve()
		if lrp.CellId != cellID {
			return false
		}

	case *models.ActualLRPChangedEvent:
		beforeLRP, _ := x.Before.Resolve()
		afterLRP, _ := x.After.Resolve()
		if afterLRP.CellId != cellID && beforeLRP.CellId != cellID {
			return false
		}

	case *models.ActualLRPRemovedEvent:
		lrp, _ := x.ActualLrpGroup.Resolve()
		if lrp.CellId != cellID {
			return false
		}

	case *models.ActualLRPCrashedEvent:
		if x.ActualLRPInstanceKey.CellId != cellID {
			return false
		}
	}

	return true
}

func convertFromFlattened(event models.Event) (models.Event, error) {
	convertLRP2Group := func(falrp *models.FlattenedActualLRP) models.ActualLRPGroup {
		alrpg := models.ActualLRPGroup{}
		alrp := models.ActualLRP{
			ActualLRPKey:         falrp.ActualLRPKey,
			ActualLRPInstanceKey: falrp.ActualLRPInstanceKey,
			State:                falrp.State,
			Since:                falrp.Since,
			ActualLRPNetInfo:     falrp.ActualLRPNetInfo,
			CrashCount:           falrp.CrashCount,
			CrashReason:          falrp.CrashReason,
			PlacementError:       falrp.PlacementError,
			ModificationTag:      falrp.ModificationTag,
		}
		evacuating := falrp.ActualLRPInfo.PlacementState == models.PlacementStateType_Evacuating
		if evacuating {
			alrpg.Evacuating = &alrp
		} else {
			alrpg.Instance = &alrp
		}
		return alrpg
	}

	switch x := event.(type) {
	case *models.FlattenedActualLRPCreatedEvent:
		alrpg := convertLRP2Group(x.ActualLrp)
		return &models.ActualLRPCreatedEvent{
			ActualLrpGroup: &alrpg,
		}, nil

	case *models.FlattenedActualLRPChangedEvent:
		before, after := x.BeforeAndAfter()
		beforeAlrpg := convertLRP2Group(before)
		afterAlrpg := convertLRP2Group(after)
		return &models.ActualLRPChangedEvent{
			Before: &beforeAlrpg,
			After:  &afterAlrpg,
		}, nil

	case *models.FlattenedActualLRPRemovedEvent:
		alrpg := convertLRP2Group(x.ActualLrp)
		return &models.ActualLRPRemovedEvent{
			ActualLrpGroup: &alrpg,
		}, nil
	}
	return nil, errors.New("not possible")
}
