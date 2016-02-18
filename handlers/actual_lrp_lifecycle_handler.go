package handlers

import (
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
	var err error
	logger := h.logger.Session("claim-actual-lrp")

	request := &models.ClaimActualLRPRequest{}
	response := &models.ActualLRPLifecycleResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		err = h.db.ClaimActualLRP(logger, request.ProcessGuid, request.Index, request.ActualLrpInstanceKey)
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}

func (h *ActualLRPLifecycleHandler) StartActualLRP(w http.ResponseWriter, req *http.Request) {
	var err error

	logger := h.logger.Session("start-actual-lrp")

	request := &models.StartActualLRPRequest{}
	response := &models.ActualLRPLifecycleResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		err = h.db.StartActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey, request.ActualLrpNetInfo)
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}

func (h *ActualLRPLifecycleHandler) CrashActualLRP(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("crash-actual-lrp")

	request := &models.CrashActualLRPRequest{}
	response := &models.ActualLRPLifecycleResponse{}
	err = parseRequest(logger, req, request)
	if err == nil {
		err = h.db.CrashActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey, request.ErrorMessage)
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}

func (h *ActualLRPLifecycleHandler) FailActualLRP(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("fail-actual-lrp")

	request := &models.FailActualLRPRequest{}
	response := &models.ActualLRPLifecycleResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		err = h.db.FailActualLRP(logger, request.ActualLrpKey, request.ErrorMessage)
	}
	response.Error = models.ConvertError(err)

	writeResponse(w, response)
}

func (h *ActualLRPLifecycleHandler) RemoveActualLRP(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("remove-actual-lrp")

	request := &models.RemoveActualLRPRequest{}
	response := &models.ActualLRPLifecycleResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		err = h.db.RemoveActualLRP(logger, request.ProcessGuid, request.Index)
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}

func (h *ActualLRPLifecycleHandler) RetireActualLRP(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("retire-actual-lrp")

	request := &models.RetireActualLRPRequest{}
	response := &models.ActualLRPLifecycleResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		err = h.db.RetireActualLRP(logger, request.ActualLrpKey)
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}
