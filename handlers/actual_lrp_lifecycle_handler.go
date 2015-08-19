package handlers

import (
	"io/ioutil"
	"net/http"

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
		writeInternalServerErrorResponse(w, err)
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
	if err := request.Validate(); err != nil {
		logger.Error("invalid-request", err)
		writeBadRequestResponse(w, models.InvalidRequest, err)
		return
	}

	bbsErr := h.db.ClaimActualLRP(logger, request)
	if bbsErr != nil {
		logger.Error("failed-to-claim-actual-lrp", bbsErr)
		if bbsErr.Equal(models.ErrResourceNotFound) {
			writeNotFoundResponse(w, bbsErr)
		} else {
			writeInternalServerErrorResponse(w, bbsErr)
		}
		return
	}

	writeEmptyResponse(w, http.StatusOK)
}

func (h *ActualLRPLifecycleHandler) StartActualLRP(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("start-actual-lrp")

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logger.Error("failed-to-read-body", err)
		writeInternalServerErrorResponse(w, err)
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

	bbsErr := h.db.StartActualLRP(logger, request)
	if bbsErr != nil {
		logger.Error("failed-to-start-actual-lrp", bbsErr)
		if bbsErr.Equal(models.ErrResourceNotFound) {
			writeNotFoundResponse(w, bbsErr)
		} else {
			writeInternalServerErrorResponse(w, bbsErr)
		}
		return
	}

	writeEmptyResponse(w, http.StatusOK)
}

func (h *ActualLRPLifecycleHandler) CrashActualLRP(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("crash-actual-lrp")

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logger.Error("crashed-to-read-body", err)
		writeInternalServerErrorResponse(w, err)
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
			writeInternalServerErrorResponse(w, bbsErr)
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
		writeInternalServerErrorResponse(w, err)
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
			writeInternalServerErrorResponse(w, bbsErr)
		}
		return
	}

	writeEmptyResponse(w, http.StatusNoContent)
}

func (h *ActualLRPLifecycleHandler) RemoveActualLRP(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("remove-actual-lrp")

	request := &models.RemoveActualLRPRequest{}
	response := &models.ActualLRPLifecycleResponse{}

	response.Error = parseRequest(logger, req, request)
	if response.Error == nil {
		response.Error = h.db.RemoveActualLRP(h.logger, request.ProcessGuid, request.Index)
	}

	writeResponse(w, response)
}

func (h *ActualLRPLifecycleHandler) RetireActualLRP(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("retire-actual-lrp")

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logger.Error("failed-to-read-body", err)
		writeInternalServerErrorResponse(w, err)
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
			writeInternalServerErrorResponse(w, bbsErr)
		}
		return
	}

	writeEmptyResponse(w, http.StatusNoContent)
}
