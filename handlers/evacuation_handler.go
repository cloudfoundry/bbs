package handlers

import (
	"io/ioutil"
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

func parseRequest(logger lager.Logger, w http.ResponseWriter, req *http.Request, request MessageValidator) bool {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logger.Error("failed-to-read-body", err)
		writeInternalServerErrorResponse(w, err)
		return false
	}

	err = request.Unmarshal(data)
	if err != nil {
		logger.Error("failed-to-parse-request-body", err)
		writeBadRequestResponse(w, models.InvalidRequest, err)
		return false
	}

	logger.Debug("parsed-request-body", lager.Data{"request": request})
	if err := request.Validate(); err != nil {
		logger.Error("invalid-request", err)
		writeBadRequestResponse(w, models.InvalidRequest, err)
		return false
	}
	return true
}

func (h *EvacuationHandler) RemoveEvacuatingActualLRP(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("remove-evacuating-actual-lrp")

	request := &models.RemoveEvacuatingActualLRPRequest{}
	if !parseRequest(logger, w, req, request) {
		return
	}

	bbsErr := h.db.RemoveEvacuatingActualLRP(logger, request)
	if bbsErr != nil {
		logger.Error("failed-to-remove-evacuating-actual-lrp", bbsErr)
		if bbsErr.Equal(models.ErrResourceNotFound) {
			writeNotFoundResponse(w, bbsErr)
		} else {
			writeInternalServerErrorResponse(w, bbsErr)
		}
		return
	}

	writeEmptyResponse(w, http.StatusNoContent)
}

func (h *EvacuationHandler) EvacuateClaimedActualLRP(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("evacuate-claimed-actual-lrp")

	request := &models.EvacuateClaimedActualLRPRequest{}
	if !parseRequest(logger, w, req, request) {
		return
	}

	keepContainer, bbsErr := h.db.EvacuateClaimedActualLRP(logger, request)

	writeProtoResponse(w, http.StatusOK, &models.EvacuationResponse{
		KeepContainer: keepContainer,
		Error:         bbsErr,
	})
}

func (h *EvacuationHandler) EvacuateCrashedActualLRP(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("evacuate-crashed-actual-lrp")

	request := &models.EvacuateCrashedActualLRPRequest{}
	if !parseRequest(logger, w, req, request) {
		return
	}

	keepContainer, bbsErr := h.db.EvacuateCrashedActualLRP(logger, request)

	writeProtoResponse(w, http.StatusOK, &models.EvacuationResponse{
		KeepContainer: keepContainer,
		Error:         bbsErr,
	})
}

func (h *EvacuationHandler) EvacuateRunningActualLRP(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("evacuate-running-actual-lrp")

	request := &models.EvacuateRunningActualLRPRequest{}
	if !parseRequest(logger, w, req, request) {
		return
	}

	keepContainer, bbsErr := h.db.EvacuateRunningActualLRP(logger, request)

	writeProtoResponse(w, http.StatusOK, &models.EvacuationResponse{
		KeepContainer: keepContainer,
		Error:         bbsErr,
	})
}

func (h *EvacuationHandler) EvacuateStoppedActualLRP(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("evacuate-stopped-actual-lrp")

	request := &models.EvacuateStoppedActualLRPRequest{}
	if !parseRequest(logger, w, req, request) {
		return
	}

	keepContainer, bbsErr := h.db.EvacuateStoppedActualLRP(logger, request)

	writeProtoResponse(w, http.StatusOK, &models.EvacuationResponse{
		KeepContainer: keepContainer,
		Error:         bbsErr,
	})
}
