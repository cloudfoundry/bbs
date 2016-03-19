package handlers

import (
	"net/http"

	"github.com/cloudfoundry-incubator/auctioneer"
	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

type ActualLRPLifecycleHandler struct {
	db               db.ActualLRPDB
	desiredLRPDB     db.DesiredLRPDB
	auctioneerClient auctioneer.Client
	retirer          ActualLRPRetirer
	logger           lager.Logger
}

func NewActualLRPLifecycleHandler(logger lager.Logger, db db.ActualLRPDB, desiredLRPDB db.DesiredLRPDB, auctioneerClient auctioneer.Client, retirer ActualLRPRetirer) *ActualLRPLifecycleHandler {
	return &ActualLRPLifecycleHandler{
		db:               db,
		desiredLRPDB:     desiredLRPDB,
		auctioneerClient: auctioneerClient,
		retirer:          retirer,
		logger:           logger.Session("actual-lrp-handler"),
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
	logger := h.logger.Session("crash-actual-lrp")

	request := &models.CrashActualLRPRequest{}
	response := &models.ActualLRPLifecycleResponse{}
	defer writeResponse(w, response)

	err := parseRequest(logger, req, request)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	actualLRPKey := request.ActualLrpKey
	actualLRPInstanceKey := request.ActualLrpInstanceKey

	shouldRestart, err := h.db.CrashActualLRP(logger, actualLRPKey, actualLRPInstanceKey, request.ErrorMessage)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	if shouldRestart {
		desiredLRP, err := h.desiredLRPDB.DesiredLRPByProcessGuid(logger, actualLRPKey.ProcessGuid)
		if err != nil {
			logger.Error("failed-fetching-desired-lrp", err)
			response.Error = models.ConvertError(err)
			return
		}

		schedInfo := desiredLRP.DesiredLRPSchedulingInfo()
		startRequest := auctioneer.NewLRPStartRequestFromSchedulingInfo(&schedInfo, int(actualLRPKey.Index))
		err = h.auctioneerClient.RequestLRPAuctions([]*auctioneer.LRPStartRequest{&startRequest})
		if err != nil {
			logger.Error("failed-requesting-auction", err)
			response.Error = models.ConvertError(err)
			return
		}
	}
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
	logger := h.logger.Session("retire-actual-lrp")
	request := &models.RetireActualLRPRequest{}
	response := &models.ActualLRPLifecycleResponse{}

	var err error
	defer writeResponse(w, response)

	err = parseRequest(logger, req, request)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	err = h.retirer.RetireActualLRP(logger, request.ActualLrpKey.ProcessGuid, request.ActualLrpKey.Index)
	response.Error = models.ConvertError(err)
}
