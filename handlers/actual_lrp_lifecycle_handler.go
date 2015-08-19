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
	logger := h.logger.Session("claim-actual-lrp")

	request := &models.ClaimActualLRPRequest{}
	response := &models.ActualLRPLifecycleResponse{}

	response.Error = parseRequest(logger, req, request)
	if response.Error == nil {
		response.Error = h.db.ClaimActualLRP(h.logger, request.ProcessGuid, request.Index, request.ActualLrpInstanceKey)
	}

	writeResponse(w, response)
}

func (h *ActualLRPLifecycleHandler) StartActualLRP(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("start-actual-lrp")

	request := &models.StartActualLRPRequest{}
	response := &models.ActualLRPLifecycleResponse{}

	response.Error = parseRequest(logger, req, request)
	if response.Error == nil {
		response.Error = h.db.StartActualLRP(h.logger, request.ActualLrpKey, request.ActualLrpInstanceKey, request.ActualLrpNetInfo)
	}

	writeResponse(w, response)
}

func (h *ActualLRPLifecycleHandler) CrashActualLRP(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("crash-actual-lrp")

	request := &models.CrashActualLRPRequest{}
	response := &models.ActualLRPLifecycleResponse{}

	response.Error = parseRequest(logger, req, request)
	if response.Error == nil {
		response.Error = h.db.CrashActualLRP(h.logger, request.ActualLrpKey, request.ActualLrpInstanceKey, request.ErrorMessage)
	}

	writeResponse(w, response)
}

func (h *ActualLRPLifecycleHandler) FailActualLRP(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("fail-actual-lrp")

	request := &models.FailActualLRPRequest{}
	response := &models.ActualLRPLifecycleResponse{}

	response.Error = parseRequest(logger, req, request)
	if response.Error == nil {
		response.Error = h.db.FailActualLRP(h.logger, request.ActualLrpKey, request.ErrorMessage)
	}

	writeResponse(w, response)
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

	request := &models.RetireActualLRPRequest{}
	response := &models.ActualLRPLifecycleResponse{}

	response.Error = parseRequest(logger, req, request)
	if response.Error == nil {
		response.Error = h.db.RetireActualLRP(h.logger, request.ActualLrpKey)
	}

	writeResponse(w, response)
}
