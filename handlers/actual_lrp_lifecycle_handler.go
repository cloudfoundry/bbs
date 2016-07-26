package handlers

import (
	"net/http"

	"code.cloudfoundry.org/auctioneer"
	"code.cloudfoundry.org/bbs/controllers"
	"code.cloudfoundry.org/bbs/db"
	"code.cloudfoundry.org/bbs/events"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

type ActualLRPLifecycleHandler struct {
	db               db.ActualLRPDB
	desiredLRPDB     db.DesiredLRPDB
	actualHub        events.Hub
	auctioneerClient auctioneer.Client
	retirer          controllers.ActualLRPRetirer
	exitChan         chan<- struct{}
}

func NewActualLRPLifecycleHandler(
	db db.ActualLRPDB,
	desiredLRPDB db.DesiredLRPDB,
	actualHub events.Hub,
	auctioneerClient auctioneer.Client,
	retirer controllers.ActualLRPRetirer,
	exitChan chan<- struct{},
) *ActualLRPLifecycleHandler {
	return &ActualLRPLifecycleHandler{
		db:               db,
		desiredLRPDB:     desiredLRPDB,
		actualHub:        actualHub,
		auctioneerClient: auctioneerClient,
		retirer:          retirer,
		exitChan:         exitChan,
	}
}

func (h *ActualLRPLifecycleHandler) ClaimActualLRP(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	var err error
	logger = logger.Session("claim-actual-lrp")

	request := &models.ClaimActualLRPRequest{}
	response := &models.ActualLRPLifecycleResponse{}
	defer func() { exitIfUnrecoverable(logger, h.exitChan, response.Error) }()
	defer writeResponse(w, response)

	err = parseRequest(logger, req, request)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	before, after, err := h.db.ClaimActualLRP(logger, request.ProcessGuid, request.Index, request.ActualLrpInstanceKey)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	if !after.Equal(before) {
		go h.actualHub.Emit(models.NewActualLRPChangedEvent(before, after))
	}
}

func (h *ActualLRPLifecycleHandler) StartActualLRP(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	var err error

	logger = logger.Session("start-actual-lrp")

	request := &models.StartActualLRPRequest{}
	response := &models.ActualLRPLifecycleResponse{}

	defer func() { exitIfUnrecoverable(logger, h.exitChan, response.Error) }()
	defer writeResponse(w, response)

	err = parseRequest(logger, req, request)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	before, after, err := h.db.StartActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey, request.ActualLrpNetInfo)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	if before == nil {
		go h.actualHub.Emit(models.NewActualLRPCreatedEvent(after))
	} else if !before.Equal(after) {
		go h.actualHub.Emit(models.NewActualLRPChangedEvent(before, after))
	}
}

func (h *ActualLRPLifecycleHandler) CrashActualLRP(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	logger = logger.Session("crash-actual-lrp")

	request := &models.CrashActualLRPRequest{}
	response := &models.ActualLRPLifecycleResponse{}
	defer func() { exitIfUnrecoverable(logger, h.exitChan, response.Error) }()
	defer writeResponse(w, response)

	err := parseRequest(logger, req, request)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	actualLRPKey := request.ActualLrpKey
	actualLRPInstanceKey := request.ActualLrpInstanceKey

	before, after, shouldRestart, err := h.db.CrashActualLRP(logger, actualLRPKey, actualLRPInstanceKey, request.ErrorMessage)
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
		logger.Info("start-lrp-auction-request", lager.Data{"app_guid": schedInfo.ProcessGuid, "index": int(actualLRPKey.Index)})
		err = h.auctioneerClient.RequestLRPAuctions([]*auctioneer.LRPStartRequest{&startRequest})
		logger.Info("finished-lrp-auction-request", lager.Data{"app_guid": schedInfo.ProcessGuid, "index": int(actualLRPKey.Index)})
		if err != nil {
			logger.Error("failed-requesting-auction", err)
			response.Error = models.ConvertError(err)
			return
		}
	}

	actualLRP, _ := after.Resolve()
	go h.actualHub.Emit(models.NewActualLRPCrashedEvent(actualLRP))
	go h.actualHub.Emit(models.NewActualLRPChangedEvent(before, after))
}

func (h *ActualLRPLifecycleHandler) FailActualLRP(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	var err error
	logger = logger.Session("fail-actual-lrp")

	request := &models.FailActualLRPRequest{}
	response := &models.ActualLRPLifecycleResponse{}

	defer func() { exitIfUnrecoverable(logger, h.exitChan, response.Error) }()
	defer writeResponse(w, response)

	err = parseRequest(logger, req, request)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	before, after, err := h.db.FailActualLRP(logger, request.ActualLrpKey, request.ErrorMessage)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	go h.actualHub.Emit(models.NewActualLRPChangedEvent(before, after))
}

func (h *ActualLRPLifecycleHandler) RemoveActualLRP(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	var err error
	logger = logger.Session("remove-actual-lrp")

	request := &models.RemoveActualLRPRequest{}
	response := &models.ActualLRPLifecycleResponse{}

	defer func() { exitIfUnrecoverable(logger, h.exitChan, response.Error) }()
	defer writeResponse(w, response)

	err = parseRequest(logger, req, request)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	beforeActualLRPGroup, err := h.db.ActualLRPGroupByProcessGuidAndIndex(logger, request.ProcessGuid, request.Index)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	err = h.db.RemoveActualLRP(logger, request.ProcessGuid, request.Index, request.ActualLrpInstanceKey)
	if err != nil {
		response.Error = models.ConvertError(err)
		return

	}
	go h.actualHub.Emit(models.NewActualLRPRemovedEvent(beforeActualLRPGroup))
}

func (h *ActualLRPLifecycleHandler) RetireActualLRP(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	logger = logger.Session("retire-actual-lrp")
	request := &models.RetireActualLRPRequest{}
	response := &models.ActualLRPLifecycleResponse{}

	var err error
	defer func() { exitIfUnrecoverable(logger, h.exitChan, response.Error) }()
	defer writeResponse(w, response)

	err = parseRequest(logger, req, request)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	err = h.retirer.RetireActualLRP(logger, request.ActualLrpKey.ProcessGuid, request.ActualLrpKey.Index)
	response.Error = models.ConvertError(err)
}
