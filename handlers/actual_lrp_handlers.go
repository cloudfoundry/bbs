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
		response.ActualLrpGroups = ResolveActualLRPGroups(lrps)
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
		response.ActualLrpGroups = ResolveActualLRPGroups(lrps)
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

func getHigherPriorityActualLRP(lrp1, lrp2 *models.ActualLRP) *models.ActualLRP {
	if hasHigherPriority(lrp1, lrp2) {
		return lrp1
	}
	return lrp2
}

// hasHigherPriority returns true if lrp1 takes precendence over lrp2
func hasHigherPriority(lrp1, lrp2 *models.ActualLRP) bool {
	if lrp1 == nil {
		return false
	}

	if lrp2 == nil {
		return true
	}

	if lrp1.Presence == models.ActualLRP_Ordinary {
		switch lrp1.State {
		case models.ActualLRPStateRunning:
			return true
		case models.ActualLRPStateClaimed:
			return lrp2.State != models.ActualLRPStateRunning && lrp2.State != models.ActualLRPStateClaimed
		}
	} else if lrp1.Presence == models.ActualLRP_Suspect {
		switch lrp1.State {
		case models.ActualLRPStateRunning:
			return lrp2.State != models.ActualLRPStateRunning
		case models.ActualLRPStateClaimed:
			return lrp2.State != models.ActualLRPStateRunning
		}
	}
	// Cases where we are comparing two LRPs with the same presence have undefined behavior since it shouldn't happen
	// with the way they're stored in the database
	return false
}

func ResolveActualLRPGroups(lrps []*models.ActualLRP) []*models.ActualLRPGroup {
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
		if actualLRP.Presence == models.ActualLRP_Evacuating {
			mapOfGroups[actualLRP.ActualLRPKey].Evacuating = actualLRP
		} else if hasHigherPriority(actualLRP, mapOfGroups[actualLRP.ActualLRPKey].Instance) {
			mapOfGroups[actualLRP.ActualLRPKey].Instance = actualLRP
		}
	}

	return result
}

func resolveToActualLRPGroup(lrps []*models.ActualLRP) *models.ActualLRPGroup {
	actualLRPGroups := ResolveActualLRPGroups(lrps)
	switch len(actualLRPGroups) {
	case 0:
		return &models.ActualLRPGroup{}
	case 1:
		return actualLRPGroups[0]
	default:
		panic("shouldn't get here")
	}
}
