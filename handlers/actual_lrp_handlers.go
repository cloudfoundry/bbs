package handlers

import (
	"net/http"
	"strconv"

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
	processGuid := req.FormValue(":process_guid")
	logger := h.logger.Session("actual-lrp-groups-by-process-guid", lager.Data{
		"process_guid": processGuid,
	})

	actualLRPGroups, err := h.db.ActualLRPGroupsByProcessGuid(h.logger, processGuid)
	if err != nil {
		logger.Error("failed-to-fetch-actual-lrp-groups", err)
		switch err {
		case models.ErrResourceNotFound:
			writeNotFoundResponse(w, err)
		default:
			writeInternalServerErrorResponse(w, err)
		}
		return
	}

	writeProtoResponse(w, http.StatusOK, actualLRPGroups)
}

func (h *ActualLRPHandler) ActualLRPGroupByProcessGuidAndIndex(w http.ResponseWriter, req *http.Request) {
	processGuid := req.FormValue(":process_guid")
	index := req.FormValue(":index")
	logger := h.logger.Session("actual-lrp-group-by-process-guid-and-index", lager.Data{
		"process_guid": processGuid,
		"index":        index,
	})

	idx, err := strconv.ParseInt(index, 10, 32)
	if err != nil {
		logger.Error("failed-to-parse-index", err)
		writeInternalServerErrorResponse(w, err)
		return
	}

	actualLRPGroup, bbsErr := h.db.ActualLRPGroupByProcessGuidAndIndex(h.logger, processGuid, int32(idx))
	if bbsErr != nil {
		logger.Error("failed-to-fetch-actual-lrp-group-by-process-guid-and-index", bbsErr)
		if bbsErr.Equal(models.ErrResourceNotFound) {
			writeNotFoundResponse(w, bbsErr)
		} else {
			writeInternalServerErrorResponse(w, bbsErr)
		}
		return
	}

	writeProtoResponse(w, http.StatusOK, actualLRPGroup)
}
