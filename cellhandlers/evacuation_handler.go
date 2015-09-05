package cellhandlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/cloudfoundry-incubator/bbs"
	"github.com/cloudfoundry-incubator/rep/evacuation/evacuation_context"
	"github.com/pivotal-golang/lager"
)

type EvacuationHandler struct {
	evacuatable evacuation_context.Evacuatable
	bbsClient   bbs.Client
	pingTimeout time.Duration
	logger      lager.Logger
}

func NewEvacuationHandler(
	logger lager.Logger,
	bbsClient bbs.Client,
	pingTimeout time.Duration,
	evacuatable evacuation_context.Evacuatable,
) *EvacuationHandler {
	return &EvacuationHandler{
		evacuatable: evacuatable,
		bbsClient:   bbsClient,
		pingTimeout: pingTimeout,
		logger:      logger,
	}
}

func (h *EvacuationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.Session("handling-evacuation")
	logger.Info("starting")
	defer logger.Info("finished")

	start := time.Now()

	for !h.bbsClient.Ping() {
		time.Sleep(100 * time.Millisecond)
		now := time.Now()
		if now.Sub(start) > h.pingTimeout {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
	}

	h.evacuatable.Evacuate()

	jsonBytes, err := json.Marshal(map[string]string{"ping_path": "/ping"})
	if err != nil {
		logger.Error("failed-to-marshal-response-payload", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Length", strconv.Itoa(len(jsonBytes)))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write(jsonBytes)
}
