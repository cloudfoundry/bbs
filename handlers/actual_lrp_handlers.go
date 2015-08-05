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
	domain := req.FormValue("domain")
	cellId := req.FormValue("cell_id")
	logger := h.logger.Session("actual-lrp-groups", lager.Data{
		"domain": domain, "cell_id": cellId,
	})

	filter := models.ActualLRPFilter{Domain: domain, CellID: cellId}
	actualLRPGroups, err := h.db.ActualLRPGroups(h.logger, filter)
	if err != nil {
		logger.Error("failed-to-fetch-actual-lrp-groups", err)
		writeInternalServerErrorResponse(w, err)
		return
	}

	writeProtoResponse(w, http.StatusOK, actualLRPGroups)
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
