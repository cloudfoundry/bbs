package handlers

import (
	"net/http"

	"code.cloudfoundry.org/bbs/db"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

type ActualLRPHandler struct {
	db       db.ActualLRPDB
	exitChan chan<- struct{}
}

func NewActualLRPHandler(db db.ActualLRPDB, exitChan chan<- struct{}) *ActualLRPHandler {
	return &ActualLRPHandler{
		db:       db,
		exitChan: exitChan,
	}
}

func (h *ActualLRPHandler) ActualLRPs(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	//TODO
	var err error
	logger = logger.Session("actual-lrps")

	request := &models.ActualLRPsRequest{}
	response := &models.ActualLRPsResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		filter := models.ActualLRPFilter{Domain: request.Domain, CellID: request.CellId}
		response.ActualLrps, err = h.db.ActualLRPs(logger, filter)
	}

	response.Error = models.ConvertError(err)

	writeResponse(w, response)
	exitIfUnrecoverable(logger, h.exitChan, response.Error)
}

func (h *ActualLRPHandler) ActualLRPsByProcessGuid(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	//TODO
	var err error
	logger = logger.Session("actual-lrps-by-process-guid")

	request := &models.ActualLRPsByProcessGuidRequest{}
	response := &models.ActualLRPsResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		response.ActualLrps, err = h.db.ActualLRPs(logger, models.ActualLRPFilter{
			ProcessGUID: &request.ProcessGuid,
		})
	}

	response.Error = models.ConvertError(err)

	writeResponse(w, response)
	exitIfUnrecoverable(logger, h.exitChan, response.Error)
}

func (h *ActualLRPHandler) ActualLRPByProcessGuidAndIndex(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	//TODO
	var err error
	logger = logger.Session("actual-lrp-by-process-guid-and-index")

	request := &models.ActualLRPByProcessGuidAndIndexRequest{}
	response := &models.ActualLRPsResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		response.ActualLrps, err = h.db.ActualLRPs(logger, models.ActualLRPFilter{
			ProcessGUID: &request.ProcessGuid,
			Index:       &request.Index,
		})
	}

	response.Error = models.ConvertError(err)

	writeResponse(w, response)
	exitIfUnrecoverable(logger, h.exitChan, response.Error)
}

func (h *ActualLRPHandler) ActualLRPGroups(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	var err error
	logger = logger.Session("actual-lrp-groups")

	request := &models.ActualLRPGroupsRequest{}
	response := &models.ActualLRPGroupsResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		filter := models.ActualLRPFilter{Domain: request.Domain, CellID: request.CellId}
		response.ActualLrpGroups, err = h.db.ActualLRPGroups(logger, filter)
	}

	response.Error = models.ConvertError(err)

	writeResponse(w, response)
	exitIfUnrecoverable(logger, h.exitChan, response.Error)
}

func (h *ActualLRPHandler) ActualLRPGroupsByProcessGuid(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	var err error
	logger = logger.Session("actual-lrp-groups-by-process-guid")

	request := &models.ActualLRPGroupsByProcessGuidRequest{}
	response := &models.ActualLRPGroupsResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		response.ActualLrpGroups, err = h.db.ActualLRPGroupsByProcessGuid(logger, request.ProcessGuid)
	}

	response.Error = models.ConvertError(err)

	writeResponse(w, response)
	exitIfUnrecoverable(logger, h.exitChan, response.Error)
}

func (h *ActualLRPHandler) ActualLRPGroupByProcessGuidAndIndex(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	var err error
	logger = logger.Session("actual-lrp-group-by-process-guid-and-index")

	request := &models.ActualLRPGroupByProcessGuidAndIndexRequest{}
	response := &models.ActualLRPGroupResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		response.ActualLrpGroup, err = h.db.ActualLRPGroupByProcessGuidAndIndex(logger, request.ProcessGuid, request.Index)
	}

	response.Error = models.ConvertError(err)

	writeResponse(w, response)
	exitIfUnrecoverable(logger, h.exitChan, response.Error)
}
