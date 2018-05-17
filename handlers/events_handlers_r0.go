package handlers

import (
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

	actualEventsFetcher := actualSource.Next
	if request.CellId != "" {
		actualEventsFetcher = func() (models.Event, error) {
			for {
				event, err := actualSource.Next()
				if err != nil {
					return event, err
				}

				if matches, err := filterByCellID(request.CellId, event, err); err != nil {
					return nil, err
				} else if matches {
					return event, nil
				}
			}
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

func filterByCellID(cellID string, bbsEvent models.Event, err error) (bool, error) {
	switch x := bbsEvent.(type) {
	case *models.ActualLRPCreatedEvent:
		lrp, _, resolveError := x.ActualLrpGroup.Resolve()
		if resolveError != nil {
			return false, resolveError
		}

		if lrp.CellId != cellID {
			return false, nil
		}

	case *models.ActualLRPChangedEvent:
		beforeLRP, _, beforeResolveError := x.Before.Resolve()
		if beforeResolveError != nil {
			return false, beforeResolveError
		}
		afterLRP, _, afterResolveError := x.After.Resolve()
		if afterResolveError != nil {
			return false, afterResolveError
		}
		if afterLRP.CellId != cellID && beforeLRP.CellId != cellID {
			return false, nil
		}

	case *models.ActualLRPRemovedEvent:
		lrp, _, resolveError := x.ActualLrpGroup.Resolve()
		if resolveError != nil {
			return false, resolveError
		}
		if lrp.CellId != cellID {
			return false, nil
		}

	case *models.ActualLRPCrashedEvent:
		if x.ActualLRPInstanceKey.CellId != cellID {
			return false, nil
		}
	}

	return true, nil
}
