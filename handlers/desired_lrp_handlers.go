package handlers

import (
	"net/http"

	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

type DesiredLRPHandler struct {
	db     db.DesiredLRPDB
	logger lager.Logger
}

func NewDesiredLRPHandler(logger lager.Logger, db db.DesiredLRPDB) *DesiredLRPHandler {
	return &DesiredLRPHandler{
		db:     db,
		logger: logger.Session("desiredlrp-handler"),
	}
}

func (h *DesiredLRPHandler) DesiredLRPs(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("desired-lrps")

	request := &models.DesiredLRPsRequest{}
	response := &models.DesiredLRPsResponse{}

	response.Error = parseRequest(logger, req, request)
	if response.Error == nil {
		filter := models.DesiredLRPFilter{Domain: request.Domain}
		response.DesiredLrps, response.Error = h.db.DesiredLRPs(logger, filter)
	}

	writeResponse(w, response)
}

func (h *DesiredLRPHandler) DesiredLRPByProcessGuid(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("desired-lrp-by-process-guid")

	request := &models.DesiredLRPByProcessGuidRequest{}
	response := &models.DesiredLRPResponse{}

	response.Error = parseRequest(logger, req, request)
	if response.Error == nil {
		response.DesiredLrp, response.Error = h.db.DesiredLRPByProcessGuid(logger, request.ProcessGuid)
	}

	writeResponse(w, response)
}
