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
	var err error
	logger = logger.Session("actual-lrps")

	request := &models.ActualLRPsRequest{}
	response := &models.ActualLRPsResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		filter := models.ActualLRPFilter{Domain: request.Domain, CellID: request.CellId, Index: request.Index, ProcessGuid: request.ProcessGuid}
		response.ActualLrps, err = h.db.ActualLRPs(logger, filter)
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
		lrps, err := h.db.ActualLRPs(logger, filter)
		if err != nil {
			response.Error = models.ConvertError(err)
			writeResponse(w, response)
			exitIfUnrecoverable(logger, h.exitChan, response.Error)
		}
		response.ActualLrpGroups = actualLRPCleanup(lrps)
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
		filter := models.ActualLRPFilter{ProcessGuid: request.ProcessGuid}
		lrps, err := h.db.ActualLRPs(logger, filter)
		if err != nil {
			response.Error = models.ConvertError(err)
			writeResponse(w, response)
			exitIfUnrecoverable(logger, h.exitChan, response.Error)
		}
		response.ActualLrpGroups = actualLRPCleanup(lrps)
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
		filter := models.ActualLRPFilter{ProcessGuid: request.ProcessGuid, Index: &request.Index}
		lrps, err := h.db.ActualLRPs(logger, filter)

		if err == nil && len(lrps) == 0 {
			err = models.ErrResourceNotFound
		}

		if err != nil {
			response.Error = models.ConvertError(err)
			writeResponse(w, response)
			exitIfUnrecoverable(logger, h.exitChan, response.Error)
		}
		response.ActualLrpGroup = resolveToActualLRPGroup(lrps)
	}
	response.Error = models.ConvertError(err)

	writeResponse(w, response)
	exitIfUnrecoverable(logger, h.exitChan, response.Error)
}

func actualLRPCleanup(lrps []*models.ActualLRP) []*models.ActualLRPGroup {
	mapOfGroups := map[models.ActualLRPKey]*models.ActualLRPGroup{}
	result := []*models.ActualLRPGroup{}
	for _, actualLRP := range lrps {
		// Every actual LRP has potentially 2 rows in the database: one for the instance
		// one for the evacuating.  When building the list of actual LRP groups (where
		// a group is the instance and corresponding evacuating), make sure we don't add the same
		// actual lrp twice.
		if mapOfGroups[actualLRP.ActualLRPKey] == nil {
			mapOfGroups[actualLRP.ActualLRPKey] = &models.ActualLRPGroup{}
			result = append(result, mapOfGroups[actualLRP.ActualLRPKey])
		}
		switch actualLRP.Presence {
		case models.ActualLRP_Evacuating:
			mapOfGroups[actualLRP.ActualLRPKey].Evacuating = actualLRP
		case models.ActualLRP_Suspect:
			// only resolve to the Suspect if the Ordinary instance is missing or not running
			if mapOfGroups[actualLRP.ActualLRPKey].Instance == nil || mapOfGroups[actualLRP.ActualLRPKey].Instance.State != models.ActualLRPStateRunning {
				mapOfGroups[actualLRP.ActualLRPKey].Instance = actualLRP
			}
		case models.ActualLRP_Ordinary:
			// only resolve to the Suspect if the Ordinary instance is missing or not running
			if mapOfGroups[actualLRP.ActualLRPKey].Instance == nil || actualLRP.State == models.ActualLRPStateRunning {
				mapOfGroups[actualLRP.ActualLRPKey].Instance = actualLRP
			}
		default:
		}
	}
	return result
}

func resolveToActualLRPGroup(lrps []*models.ActualLRP) *models.ActualLRPGroup {
	actualLRPGroups := actualLRPCleanup(lrps)
	switch len(actualLRPGroups) {
	case 0:
		return &models.ActualLRPGroup{}
	case 1:
		return actualLRPGroups[0]
	default:
		panic("shouldn't get here")
	}
}
