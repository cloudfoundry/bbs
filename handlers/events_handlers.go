package handlers

import (
	"encoding/base64"
	"net/http"
	"strconv"

	"github.com/cloudfoundry-incubator/bbs/events"
	"github.com/gogo/protobuf/proto"
	"github.com/pivotal-golang/lager"
	"github.com/vito/go-sse/sse"
)

type EventHandler struct {
	hub    events.Hub
	logger lager.Logger
}

var ()

func NewEventHandler(logger lager.Logger, hub events.Hub) *EventHandler {
	return &EventHandler{
		hub:    hub,
		logger: logger.Session("domain-handler"),
	}
}

func (h *EventHandler) Subscribe(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("event-handler")

	closeNotifier := w.(http.CloseNotifier).CloseNotify()

	flusher := w.(http.Flusher)

	source, err := h.hub.Subscribe()
	if err != nil {
		logger.Error("failed-to-subscribe-to-event-hub", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer source.Close()

	go func() {
		<-closeNotifier
		source.Close()
	}()

	w.Header().Add("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Add("Connection", "keep-alive")

	w.WriteHeader(http.StatusOK)

	flusher.Flush()

	eventID := 0
	for {
		event, err := source.Next()
		if err != nil {
			logger.Error("failed-to-get-next-event", err)
			return
		}

		payload, err := proto.Marshal(event)
		if err != nil {
			logger.Error("failed-to-marshal-event", err)
			return
		}

		encodedPayload := base64.StdEncoding.EncodeToString(payload)
		err = sse.Event{
			ID:   strconv.Itoa(eventID),
			Name: string(event.EventType()),
			Data: []byte(encodedPayload),
		}.Write(w)
		if err != nil {
			break
		}

		flusher.Flush()

		eventID++
	}
}
