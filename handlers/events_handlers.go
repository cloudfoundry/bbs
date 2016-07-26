package handlers

import (
	"net/http"

	"code.cloudfoundry.org/bbs/events"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

type EventHandler struct {
	desiredHub events.Hub
	actualHub  events.Hub
}

func NewEventHandler(desiredHub, actualHub events.Hub) *EventHandler {
	return &EventHandler{
		desiredHub: desiredHub,
		actualHub:  actualHub,
	}
}

func streamEventsToResponse(logger lager.Logger, w http.ResponseWriter, eventChan <-chan models.Event, errorChan <-chan error) {
	w.Header().Add("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Add("Connection", "keep-alive")

	w.WriteHeader(http.StatusOK)

	flusher := w.(http.Flusher)
	flusher.Flush()
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
			return
		}

		sseEvent, err := events.NewEventFromModelEvent(eventID, event)
		if err != nil {
			logger.Error("failed-to-marshal-event", err)
			return
		}

		err = sseEvent.Write(w)
		if err != nil {
			return
		}

		flusher.Flush()

		eventID++
	}
}

type EventFetcher func() (models.Event, error)

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
