package cellhandlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/cloudfoundry-incubator/rep/evacuation/evacuation_context"
	"github.com/pivotal-golang/lager"
)

type EvacuationHandler struct {
	evacuatable evacuation_context.Evacuatable
	logger      lager.Logger
}

func NewEvacuationHandler(
	logger lager.Logger,
	evacuatable evacuation_context.Evacuatable,
) *EvacuationHandler {
	return &EvacuationHandler{
		evacuatable: evacuatable,
		logger:      logger,
	}
}

func (h *EvacuationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.Session("handling-evacuation")
	logger.Info("starting")
	defer logger.Info("finished")

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
