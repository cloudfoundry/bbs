package handlers

import (
	"net/http"

	"code.cloudfoundry.org/bbs/models"
)

func (h *EventHandler) SubscribeToDesiredLRPEvents(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("subscribe-desired")

	source, err := h.desiredHub.Subscribe()
	if err != nil {
		logger.Error("failed-to-subscribe-to-event-hub", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer source.Close()

	eventChan := make(chan models.Event)
	errorChan := make(chan error)
	closeChan := make(chan struct{})
	defer close(closeChan)

	go streamSource(eventChan, errorChan, closeChan, source.Next)

	streamEventsToResponse(logger, w, eventChan, errorChan)
}
