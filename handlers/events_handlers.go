package handlers

import (
	"net/http"

	"code.cloudfoundry.org/bbs/events"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

type EventController interface {
	Subscribe(logger lager.Logger, w http.ResponseWriter, req *http.Request)
	Subscribe_r0(logger lager.Logger, w http.ResponseWriter, req *http.Request)
}

type EventHandler struct {
	desiredHub events.Hub
	actualHub  events.Hub
}

type TaskEventHandler struct {
	taskHub events.Hub
}

func NewEventHandler(desiredHub, actualHub events.Hub) *EventHandler {
	return &EventHandler{
		desiredHub: desiredHub,
		actualHub:  actualHub,
	}
}

func NewTaskEventHandler(taskHub events.Hub) *TaskEventHandler {
	return &TaskEventHandler{
		taskHub: taskHub,
	}
}

func streamEventsToResponse(logger lager.Logger, w http.ResponseWriter, eventChan <-chan models.Event, errorChan <-chan error) {
	w.Header().Add("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Add("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "identity")

	w.WriteHeader(http.StatusOK)

	conn, rw, err := w.(http.Hijacker).Hijack()
	if err != nil {
		return
	}

	defer func() {
		err := conn.Close()
		if err != nil {
			logger.Error("failed-to-close-connection", err)
		}
	}()

	if err := rw.Flush(); err != nil {
		logger.Error("failed-to-flush", err)
		return
	}

	var event models.Event
	eventID := 0
	closeNotifier := w.(http.CloseNotifier).CloseNotify()

	for {
		select {
		case event = <-eventChan:
		case err := <-errorChan:
			logger.Error("failed-to-get-next-event", err)
			return
		case <-closeNotifier:
			logger.Debug("received-close-notify")
			return
		}

		sseEvent, err := events.NewEventFromModelEvent(eventID, event)
		if err != nil {
			logger.Error("failed-to-marshal-event", err)
			return
		}

		err = sseEvent.Write(conn)
		if err != nil {
			logger.Error("failed-to-write-event", err)
			return
		}

		eventID++
	}
}

type EventFetcher func() (models.Event, error)

func (h *EventHandler) Subscribe(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	logger = logger.Session("subscribe")
	h.subscribe(logger, w, req, false)
}

func (h *EventHandler) Subscribe_r0(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	logger = logger.Session("subscribe-r0")
	h.subscribe(logger, w, req, true)
}

func (h *EventHandler) subscribe(logger lager.Logger, w http.ResponseWriter, req *http.Request, useV0FormatForDesiredLRPs bool) {
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

				if filterByCellID(request.CellId, event, err) {
					return event, nil
				}
			}
		}
	}

	desiredEventsFetcher := func() (models.Event, error) {
		event, err := desiredSource.Next()
		if err != nil {
			return nil, err
		}
		if useV0FormatForDesiredLRPs {
			event = models.VersionDesiredLRPsToV0(event)
		}
		return event, nil
	}

	go streamSource(eventChan, errorChan, closeChan, desiredEventsFetcher)
	go streamSource(eventChan, errorChan, closeChan, actualEventsFetcher)

	streamEventsToResponse(logger, w, eventChan, errorChan)
}

func (h *TaskEventHandler) Subscribe(logger lager.Logger, w http.ResponseWriter, req *http.Request) {} // TODO come back to this!!

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

func streamSource(eventChan chan<- models.Event, errorChan chan<- error, closeChan chan struct{}, fetchEvent EventFetcher) {
	for {
		event, err := fetchEvent()
		if err != nil {
			select {
			case errorChan <- err:
			case <-closeChan:
			}
			return
		}
		select {
		case eventChan <- event:
		case <-closeChan:
			return
		}
	}
}
