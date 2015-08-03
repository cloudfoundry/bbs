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
	logger := h.logger.Session("claim-actual-lrp")

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logger.Error("failed-to-read-body", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	request := &models.ClaimActualLRPRequest{}
	err = request.Unmarshal(data)
	if err != nil {
		logger.Error("failed-to-parse-request-body", err)
		writeBadRequestResponse(w, models.InvalidRequest, err)
		return
	}
	logger.Debug("parsed-request-body", lager.Data{"request": request})

	actualLRP, bbsErr := h.db.ClaimActualLRP(logger, request.ProcessGuid, request.Index, request.ActualLrpInstanceKey)
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

func (h *ActualLRPLifecycleHandler) StartActualLRP(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("start-actual-lrp")

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logger.Error("failed-to-read-body", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	request := &models.StartActualLRPRequest{}
	err = request.Unmarshal(data)
	if err != nil {
		logger.Error("failed-to-parse-request-body", err)
		writeBadRequestResponse(w, models.InvalidRequest, err)
		return
	}
	logger.Debug("parsed-request-body", lager.Data{"request": request})
	if err := request.Validate(); err != nil {
		logger.Error("invalid-request", err)
		writeBadRequestResponse(w, models.InvalidRequest, err)
		return
	}

	actualLRP, bbsErr := h.db.StartActualLRP(logger, request)
	if bbsErr != nil {
		logger.Error("failed-to-start-actual-lrp", bbsErr)
		if bbsErr.Equal(models.ErrResourceNotFound) {
			writeNotFoundResponse(w, bbsErr)
		} else {
			writeUnknownErrorResponse(w, bbsErr)
		}
		return
	}

	writeProtoResponse(w, http.StatusOK, actualLRP)
}

func (h *ActualLRPLifecycleHandler) CrashActualLRP(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("crash-actual-lrp")

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logger.Error("crashed-to-read-body", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	request := &models.CrashActualLRPRequest{}
	err = request.Unmarshal(data)
	if err != nil {
		logger.Error("crashed-to-parse-request-body", err)
		writeBadRequestResponse(w, models.InvalidRequest, err)
		return
	}
	logger.Debug("parsed-request-body", lager.Data{"request": request})
	if err := request.Validate(); err != nil {
		logger.Error("invalid-request", err)
		writeBadRequestResponse(w, models.InvalidRequest, err)
		return
	}

	bbsErr := h.db.CrashActualLRP(logger, request)
	if bbsErr != nil {
		logger.Error("crashed-to-crash-actual-lrp", bbsErr)
		if bbsErr.Equal(models.ErrResourceNotFound) {
			writeNotFoundResponse(w, bbsErr)
		} else {
			writeUnknownErrorResponse(w, bbsErr)
		}
		return
	}

	writeEmptyResponse(w, http.StatusNoContent)
}

func (h *ActualLRPLifecycleHandler) FailActualLRP(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("fail-actual-lrp")

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logger.Error("failed-to-read-body", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	request := &models.FailActualLRPRequest{}
	err = request.Unmarshal(data)
	if err != nil {
		logger.Error("failed-to-parse-request-body", err)
		writeBadRequestResponse(w, models.InvalidRequest, err)
		return
	}
	logger.Debug("parsed-request-body", lager.Data{"request": request})
	if err := request.Validate(); err != nil {
		logger.Error("invalid-request", err)
		writeBadRequestResponse(w, models.InvalidRequest, err)
		return
	}

	bbsErr := h.db.FailActualLRP(logger, request)
	if bbsErr != nil {
		logger.Error("failed-to-fail-actual-lrp", bbsErr)
		if bbsErr.Equal(models.ErrResourceNotFound) {
			writeNotFoundResponse(w, bbsErr)
		} else {
			writeUnknownErrorResponse(w, bbsErr)
		}
		return
	}

	writeEmptyResponse(w, http.StatusNoContent)
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

func (h *ActualLRPLifecycleHandler) RetireActualLRP(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("retire-actual-lrp")

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logger.Error("failed-to-read-body", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	request := &models.RetireActualLRPRequest{}
	err = request.Unmarshal(data)
	if err != nil {
		logger.Error("failed-to-parse-request-body", err)
		writeBadRequestResponse(w, models.InvalidRequest, err)
		return
	}
	logger.Debug("parsed-request-body", lager.Data{"request": request})
	if err := request.Validate(); err != nil {
		logger.Error("invalid-request", err)
		writeBadRequestResponse(w, models.InvalidRequest, err)
		return
	}

	bbsErr := h.db.RetireActualLRP(logger, request)
	if bbsErr != nil {
		logger.Error("failed-to-retire-actual-lrp", bbsErr)
		if bbsErr.Equal(models.ErrResourceNotFound) {
			writeNotFoundResponse(w, bbsErr)
		} else {
			writeUnknownErrorResponse(w, bbsErr)
		}
		return
	}

	writeEmptyResponse(w, http.StatusNoContent)
}
