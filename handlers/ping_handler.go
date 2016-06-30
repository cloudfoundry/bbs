package handlers

import (
	"net/http"

	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

type PingHandler struct {
	logger lager.Logger
}

func NewPingHandler(logger lager.Logger) *PingHandler {
	return &PingHandler{
		logger: logger.Session("ping-handler"),
	}
}

func (h *PingHandler) Ping(w http.ResponseWriter, req *http.Request) {
	response := &models.PingResponse{}
	response.Available = true
	writeResponse(w, response)
}
