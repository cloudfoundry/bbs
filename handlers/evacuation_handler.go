package handlers

import (
	"net/http"

	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/gogo/protobuf/proto"
	"github.com/pivotal-golang/lager"
)

type EvacuationHandler struct {
	db     db.EvacuationDB
	logger lager.Logger
}

func NewEvacuationHandler(logger lager.Logger, db db.EvacuationDB) *EvacuationHandler {
	return &EvacuationHandler{
		db:     db,
		logger: logger.Session("evacuation-handler"),
	}
}

type MessageValidator interface {
	proto.Message
	Validate() error
	Unmarshal(data []byte) error
}

func (h *EvacuationHandler) RemoveEvacuatingActualLRP(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("remove-evacuating-actual-lrp")

	request := &models.RemoveEvacuatingActualLRPRequest{}
	response := &models.RemoveEvacuatingActualLRPResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		err = h.db.RemoveEvacuatingActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey)
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}

func (h *EvacuationHandler) EvacuateClaimedActualLRP(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("evacuate-claimed-actual-lrp")

	request := &models.EvacuateClaimedActualLRPRequest{}
	response := &models.EvacuationResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		response.KeepContainer, err = h.db.EvacuateClaimedActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey)
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}

func (h *EvacuationHandler) EvacuateCrashedActualLRP(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("evacuate-crashed-actual-lrp")

	request := &models.EvacuateCrashedActualLRPRequest{}
	response := &models.EvacuationResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		response.KeepContainer, err = h.db.EvacuateCrashedActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey, request.ErrorMessage)
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}

func (h *EvacuationHandler) EvacuateRunningActualLRP(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("evacuate-running-actual-lrp")

	request := &models.EvacuateRunningActualLRPRequest{}
	response := &models.EvacuationResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		response.KeepContainer, err = h.db.EvacuateRunningActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey, request.ActualLrpNetInfo, request.Ttl)
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}

func (h *EvacuationHandler) EvacuateStoppedActualLRP(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("evacuate-stopped-actual-lrp")
	request := &models.EvacuateStoppedActualLRPRequest{}
	response := &models.EvacuationResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		response.KeepContainer, err = h.db.EvacuateStoppedActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey)
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}
