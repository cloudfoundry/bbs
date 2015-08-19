package handlers

import (
	"net/http"

	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

type ActualLRPHandler struct {
	db     db.ActualLRPDB
	logger lager.Logger
}

func NewActualLRPHandler(logger lager.Logger, db db.ActualLRPDB) *ActualLRPHandler {
	return &ActualLRPHandler{
		db:     db,
		logger: logger.Session("actuallrp-handler"),
	}
}

func (h *ActualLRPHandler) ActualLRPGroups(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("actual-lrp-groups")

	request := &models.ActualLRPGroupsRequest{}
	response := &models.ActualLRPGroupsResponse{}

	response.Error = parseRequest(logger, req, request)
	if response.Error == nil {
		filter := models.ActualLRPFilter{Domain: request.Domain, CellID: request.CellId}
		response.ActualLrpGroups, response.Error = h.db.ActualLRPGroups(h.logger, filter)
	}

	writeResponse(w, response)
}

func (h *ActualLRPHandler) ActualLRPGroupsByProcessGuid(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("actual-lrp-groups-by-process-guid")

	request := &models.ActualLRPGroupsByProcessGuidRequest{}
	response := &models.ActualLRPGroupsResponse{}

	response.Error = parseRequest(logger, req, request)
	if response.Error == nil {
		response.ActualLrpGroups, response.Error = h.db.ActualLRPGroupsByProcessGuid(h.logger, request.ProcessGuid)
	}

	writeResponse(w, response)
}

func (h *ActualLRPHandler) ActualLRPGroupByProcessGuidAndIndex(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("actual-lrp-group-by-process-guid-and-index")

	request := &models.ActualLRPGroupByProcessGuidAndIndexRequest{}
	response := &models.ActualLRPGroupResponse{}

	response.Error = parseRequest(logger, req, request)
	if response.Error == nil {
		response.ActualLrpGroup, response.Error = h.db.ActualLRPGroupByProcessGuidAndIndex(h.logger, request.ProcessGuid, request.Index)
	}

	writeResponse(w, response)
}
