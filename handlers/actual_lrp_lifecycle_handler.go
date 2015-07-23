package handlers

import (
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

type ActualLRPLifecycleHandler struct {
	db     db.ActualLRPDB
	logger lager.Logger
}

func NewActualLRPLifecycleHandler(logger lager.Logger, db db.ActualLRPDB) *ActualLRPLifecycleHandler {
	return &ActualLRPLifecycleHandler{
		db:     db,
		logger: logger.Session("actuallrp-handler"),
	}
}

func (h *ActualLRPLifecycleHandler) ClaimActualLRP(w http.ResponseWriter, req *http.Request) {
	processGuid := req.FormValue(":process_guid")
	index := req.FormValue(":index")
	logger := h.logger.Session("claim-actual-lrp", lager.Data{
		"process_guid": processGuid,
		"index":        index,
	})

	idx, err := strconv.ParseInt(index, 10, 32)
	if err != nil {
		logger.Error("failed-to-parse-index", err)
		writeBadRequestResponse(w, models.InvalidRequest, err)
		return
	}

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logger.Error("failed-to-read-body", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	instanceKey := &models.ActualLRPInstanceKey{}
	err = instanceKey.Unmarshal(data)
	if err != nil {
		logger.Error("failed-to-parse-request-body", err)
		writeBadRequestResponse(w, models.InvalidRequest, err)
		return
	}

	actualLRP, bbsErr := h.db.ClaimActualLRP(logger, processGuid, int32(idx), *instanceKey)
	if bbsErr != nil {
		logger.Error("failed-to-claim-actual-lrp", bbsErr)
		if bbsErr.Equal(models.ErrResourceNotFound) {
			writeNotFoundResponse(w, bbsErr)
		} else {
			writeUnknownErrorResponse(w, bbsErr)
		}
		return
	}

	writeProtoResponse(w, http.StatusOK, actualLRP)
}

func (h *ActualLRPLifecycleHandler) RemoveActualLRP(w http.ResponseWriter, req *http.Request) {
	processGuid := req.FormValue(":process_guid")
	index := req.FormValue(":index")
	logger := h.logger.Session("remove-actual-lrp", lager.Data{
		"process_guid": processGuid,
		"index":        index,
	})

	idx, err := strconv.ParseInt(index, 10, 32)
	if err != nil {
		logger.Error("failed-to-parse-index", err)
		writeBadRequestResponse(w, models.InvalidRequest, err)
		return
	}

	bbsErr := h.db.RemoveActualLRP(logger, processGuid, int32(idx))
	if bbsErr != nil {
		logger.Error("failed-to-remove-actual-lrp", bbsErr)
		if bbsErr.Equal(models.ErrResourceNotFound) {
			writeNotFoundResponse(w, bbsErr)
		} else {
			writeUnknownErrorResponse(w, bbsErr)
		}
		return
	}

	writeEmptyResponse(w, http.StatusNoContent)
}
