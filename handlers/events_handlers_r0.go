package handlers

import (
	"net/http"

	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

func (h *EventHandler) Subscribe_r0(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	var cellId string
	logger = logger.Session("subscribe-r0")

	request := &models.ActualLRPGroupsRequest{}

	err := parseRequest(logger, req, request)
	if err == nil {
		cellId = request.CellId
	}

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

	desiredEventsFetcher := func() (models.Event, error) {
		event, err := desiredSource.Next()
		if err != nil {
			return event, err
		}
		event = models.VersionDesiredLRPsToV0(event)
		return event, err
	}

	actualEventsFetcher := func() (models.Event, error) {
		for {
			event, err := actualSource.Next()
			if err != nil || cellId == "" {
				return event, err
			}

			switch event := event.(type) {
			case *models.ActualLRPCreatedEvent:
				lrp, _ := event.ActualLrpGroup.Resolve()
				if lrp.CellId == cellId {
					return event, nil
				}
			case *models.ActualLRPChangedEvent:
				lrp, _ := event.Before.Resolve()
				if lrp.CellId == cellId {
					return event, nil
				}

				lrp, _ = event.After.Resolve()
				if lrp.CellId == cellId {
					return event, nil
				}
			case *models.ActualLRPRemovedEvent:
				lrp, _ := event.ActualLrpGroup.Resolve()
				if lrp.CellId == cellId {
					return event, nil
				}
			case *models.ActualLRPCrashedEvent:
				if event.CellId == cellId {
					return event, nil
				}
			default:
				return event, nil
			}
		}
	}

	go streamSource(eventChan, errorChan, closeChan, desiredEventsFetcher)
	go streamSource(eventChan, errorChan, closeChan, actualEventsFetcher)

	streamEventsToResponse(logger, w, eventChan, errorChan)
}
