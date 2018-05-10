package handlers

import (
	"net/http"

	"code.cloudfoundry.org/auctioneer"
	"code.cloudfoundry.org/bbs/db"
	"code.cloudfoundry.org/bbs/events"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
	"github.com/gogo/protobuf/proto"
)

type EvacuationHandler struct {
	db               db.EvacuationDB
	actualLRPDB      db.ActualLRPDB
	desiredLRPDB     db.DesiredLRPDB
	actualHub        events.Hub
	auctioneerClient auctioneer.Client
	exitChan         chan<- struct{}
}

func NewEvacuationHandler(
	db db.EvacuationDB,
	actualLRPDB db.ActualLRPDB,
	desiredLRPDB db.DesiredLRPDB,
	actualHub events.Hub,
	auctioneerClient auctioneer.Client,
	exitChan chan<- struct{},
) *EvacuationHandler {
	return &EvacuationHandler{
		db:               db,
		actualLRPDB:      actualLRPDB,
		desiredLRPDB:     desiredLRPDB,
		actualHub:        actualHub,
		auctioneerClient: auctioneerClient,
		exitChan:         exitChan,
	}
}

type MessageValidator interface {
	proto.Message
	Validate() error
	Unmarshal(data []byte) error
}

func (h *EvacuationHandler) RemoveEvacuatingActualLRP(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	var err error
	logger = logger.Session("remove-evacuating-actual-lrp")
	logger.Info("started")
	defer logger.Info("completed")

	request := &models.RemoveEvacuatingActualLRPRequest{}
	response := &models.RemoveEvacuatingActualLRPResponse{}

	defer func() { exitIfUnrecoverable(logger, h.exitChan, response.Error) }()
	defer writeResponse(w, response)

	err = parseRequest(logger, req, request)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	beforeActualLRP, err := h.actualLRPDB.ActualLRPByProcessGuidAndIndex(logger, request.ActualLrpKey.ProcessGuid, request.ActualLrpKey.Index)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	evacuatingLRPLogData := lager.Data{
		"process-guid": request.ActualLrpKey.ProcessGuid,
		"index":        request.ActualLrpKey.Index,
		"instance-key": request.ActualLrpInstanceKey,
	}
	if beforeActualLRP.PlacementState != models.PlacementStateType_Normal {
		evacuatingLRPLogData["replacement-lrp-instance-key"] = beforeActualLRP.ActualLRPInstanceKey
		evacuatingLRPLogData["replacement-state"] = beforeActualLRP.State
		evacuatingLRPLogData["replacement-lrp-placement-error"] = beforeActualLRP.PlacementError
	}

	logger.Info("removing-stranded-evacuating-actual-lrp", evacuatingLRPLogData)

	err = h.db.RemoveEvacuatingActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	if beforeActualLRP.PlacementState != models.PlacementStateType_Evacuating {
		logger.Info("evacuating-lrp-is-emtpy")
		response.Error = models.ConvertError(models.ErrResourceNotFound)
		return
	}

	go h.actualHub.Emit(models.NewFlattenedActualLRPRemovedEvent(beforeActualLRP))
}

func (h *EvacuationHandler) EvacuateClaimedActualLRP(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	logger = logger.Session("evacuate-claimed-actual-lrp")
	logger.Info("started")
	defer logger.Info("completed")

	request := &models.EvacuateClaimedActualLRPRequest{}
	response := &models.EvacuationResponse{}
	defer func() { exitIfUnrecoverable(logger, h.exitChan, response.Error) }()
	defer writeResponse(w, response)

	err := parseRequest(logger, req, request)
	if err != nil {
		logger.Error("failed-parsing-request", err)
		response.Error = models.ConvertError(err)
		response.KeepContainer = true
		return
	}

	beforeActualLRP, err := h.actualLRPDB.ActualLRPByProcessGuidAndIndex(logger, request.ActualLrpKey.ProcessGuid, request.ActualLrpKey.Index)
	if err == nil {
		err = h.db.RemoveEvacuatingActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey)
		if err != nil {
			logger.Error("failed-removing-evacuating-actual-lrp", err)
			exitIfUnrecoverable(logger, h.exitChan, models.ConvertError(err))
		} else {
			go h.actualHub.Emit(models.NewFlattenedActualLRPRemovedEvent(beforeActualLRP))
		}
	}

	err = h.unclaimAndRequestAuction(logger, request.ActualLrpKey)
	bbsErr := models.ConvertError(err)
	if bbsErr != nil && bbsErr.Type != models.Error_ResourceNotFound {
		response.Error = bbsErr
		response.KeepContainer = true
		return
	}
}

func (h *EvacuationHandler) EvacuateCrashedActualLRP(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	logger = logger.Session("evacuate-crashed-actual-lrp")
	logger.Info("started")
	defer logger.Info("completed")

	request := &models.EvacuateCrashedActualLRPRequest{}
	response := &models.EvacuationResponse{}
	defer func() { exitIfUnrecoverable(logger, h.exitChan, response.Error) }()
	defer writeResponse(w, response)

	err := parseRequest(logger, req, request)
	if err != nil {
		logger.Error("failed-parsing-request", err)
		response.Error = models.ConvertError(err)
		return
	}

	beforeActualLRP, err := h.actualLRPDB.ActualLRPByProcessGuidAndIndex(logger, request.ActualLrpKey.ProcessGuid, request.ActualLrpKey.Index)
	if err == nil {
		err = h.db.RemoveEvacuatingActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey)
		if err != nil {
			logger.Error("failed-removing-evacuating-actual-lrp", err)
			exitIfUnrecoverable(logger, h.exitChan, models.ConvertError(err))
		} else {
			go h.actualHub.Emit(models.NewFlattenedActualLRPRemovedEvent(beforeActualLRP))
		}
	}

	_, _, _, err = h.actualLRPDB.CrashActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey, request.ErrorMessage)
	bbsErr := models.ConvertError(err)
	if bbsErr != nil && bbsErr.Type != models.Error_ResourceNotFound {
		logger.Error("failed-crashing-actual-lrp", err)
		response.Error = bbsErr
		return
	}
}

func (h *EvacuationHandler) EvacuateRunningActualLRP(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	logger = logger.Session("evacuate-running-actual-lrp")
	logger.Info("starting")
	defer logger.Info("completed")

	response := &models.EvacuationResponse{}
	response.KeepContainer = true
	defer func() { exitIfUnrecoverable(logger, h.exitChan, response.Error) }()
	defer writeResponse(w, response)

	request := &models.EvacuateRunningActualLRPRequest{}
	err := parseRequest(logger, req, request)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	guid := request.ActualLrpKey.ProcessGuid
	index := request.ActualLrpKey.Index
	actualLRP, err := h.actualLRPDB.ActualLRPByProcessGuidAndIndex(logger, guid, index)
	if err != nil {
		if err == models.ErrResourceNotFound {
			response.KeepContainer = false
			return
		}
		logger.Error("failed-fetching-lrp-group", err)
		response.Error = models.ConvertError(err)
		return
	}

	// If the instance is not there, clean up the corresponding evacuating LRP, if one exists.
	if actualLRP.PlacementState != models.PlacementStateType_Normal {
		err = h.db.RemoveEvacuatingActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey)
		if err != nil {
			if err == models.ErrActualLRPCannotBeRemoved {
				logger.Debug("remove-evacuating-actual-lrp-failed")
				response.KeepContainer = false
				return
			}
			logger.Error("failed-removing-evacuating-actual-lrp", err)
			response.Error = models.ConvertError(err)
			return
		}

		go h.actualHub.Emit(models.NewFlattenedActualLRPRemovedEvent(actualLRP))
		response.KeepContainer = false
		return
	}

	if (actualLRP.State == models.ActualLRPStateUnclaimed && actualLRP.PlacementError == "") ||
		(actualLRP.State == models.ActualLRPStateClaimed && !actualLRP.ActualLRPInstanceKey.Equal(request.ActualLrpInstanceKey)) {
		if actualLRP.PlacementState == models.PlacementStateType_Evacuating && !actualLRP.ActualLRPInstanceKey.Equal(request.ActualLrpInstanceKey) {
			logger.Error("already-evacuated-by-different-cell", err)
			response.KeepContainer = false
			return
		}

		evacuatingLRP, err := h.db.EvacuateActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey, request.ActualLrpNetInfo, request.Ttl)
		if err == models.ErrActualLRPCannotBeEvacuated {
			logger.Error("cannot-evacuate-actual-lrp", err)
			response.KeepContainer = false
			return
		}

		response.KeepContainer = true

		if err != nil {
			logger.Error("failed-evacuating-actual-lrp", err)
			response.Error = models.ConvertError(err)
		} else {
			go h.actualHub.Emit(models.NewFlattenedActualLRPCreatedEvent(evacuatingLRP))
		}

		return
	}

	if (actualLRP.State == models.ActualLRPStateRunning && !actualLRP.ActualLRPInstanceKey.Equal(request.ActualLrpInstanceKey)) ||
		actualLRP.State == models.ActualLRPStateCrashed {
		response.KeepContainer = false

		// if there is not evacuating instance, it probably got removed when the
		// new instance transitioned to a Running state
		if actualLRP.PlacementState != models.PlacementStateType_Evacuating {
			return
		}

		err = h.db.RemoveEvacuatingActualLRP(logger, &actualLRP.ActualLRPKey, &actualLRP.ActualLRPInstanceKey)
		if err == nil {
			go h.actualHub.Emit(models.NewFlattenedActualLRPRemovedEvent(actualLRP))
		}
		if err != nil && err != models.ErrActualLRPCannotBeRemoved {
			response.KeepContainer = true
			response.Error = models.ConvertError(err)
		}
		return
	}

	if (actualLRP.State == models.ActualLRPStateClaimed || actualLRP.State == models.ActualLRPStateRunning) &&
		actualLRP.ActualLRPInstanceKey.Equal(request.ActualLrpInstanceKey) {
		evacuatingLRP, err := h.db.EvacuateActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey, request.ActualLrpNetInfo, request.Ttl)
		if err != nil {
			response.Error = models.ConvertError(err)
			return
		}

		go h.actualHub.Emit(models.NewFlattenedActualLRPCreatedEvent(evacuatingLRP))

		err = h.unclaimAndRequestAuction(logger, request.ActualLrpKey)
		if err != nil {
			response.Error = models.ConvertError(err)
			return
		}
	}
}

func (h *EvacuationHandler) EvacuateStoppedActualLRP(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	logger = logger.Session("evacuate-stopped-actual-lrp")

	request := &models.EvacuateStoppedActualLRPRequest{}
	response := &models.EvacuationResponse{}

	var bbsErr *models.Error

	defer func() { exitIfUnrecoverable(logger, h.exitChan, bbsErr) }()
	defer writeResponse(w, response)

	err := parseRequest(logger, req, request)
	if err != nil {
		logger.Error("failed-to-parse-request", err)
		bbsErr = models.ConvertError(err)
		response.Error = bbsErr
		return
	}

	guid := request.ActualLrpKey.ProcessGuid
	index := request.ActualLrpKey.Index

	actualLRP, err := h.actualLRPDB.ActualLRPByProcessGuidAndIndex(logger, guid, index)
	if err != nil {
		logger.Error("failed-fetching-actual-lrp-group", err)
		bbsErr = models.ConvertError(err)
		response.Error = bbsErr
		return
	}

	err = h.db.RemoveEvacuatingActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey)
	if err != nil {
		logger.Error("failed-removing-evacuating-actual-lrp", err)
		bbsErr = models.ConvertError(err)
	} else if actualLRP.PlacementState != models.PlacementStateType_Evacuating {
		go h.actualHub.Emit(models.NewFlattenedActualLRPRemovedEvent(actualLRP))
	}

	if actualLRP.PlacementState != models.PlacementStateType_Normal || !actualLRP.ActualLRPInstanceKey.Equal(request.ActualLrpInstanceKey) {
		logger.Debug("cannot-remove-actual-lrp")
		response.Error = models.ErrActualLRPCannotBeRemoved
		return
	}

	err = h.actualLRPDB.RemoveActualLRP(logger, guid, index, request.ActualLrpInstanceKey)
	if err != nil {
		logger.Error("failed-to-remove-actual-lrp", err)
		bbsErr = models.ConvertError(err)
		response.Error = bbsErr
		return
	} else {
		go h.actualHub.Emit(models.NewFlattenedActualLRPRemovedEvent(actualLRP))
	}
}

func (h *EvacuationHandler) unclaimAndRequestAuction(logger lager.Logger, lrpKey *models.ActualLRPKey) error {
	before, after, err := h.actualLRPDB.UnclaimActualLRP(logger, lrpKey)
	if err != nil {
		return err
	}

	event, err := models.NewFlattenedActualLRPChangedEvent(before, after)
	if err != nil {
		go h.actualHub.Emit(models.NewFlattenedActualLRPRemovedEvent(before))
		go h.actualHub.Emit(models.NewFlattenedActualLRPCreatedEvent(after))
	}
	go h.actualHub.Emit(event)

	desiredLRP, err := h.desiredLRPDB.DesiredLRPByProcessGuid(logger, lrpKey.ProcessGuid)
	if err != nil {
		logger.Error("failed-fetching-desired-lrp", err)
		return nil
	}

	schedInfo := desiredLRP.DesiredLRPSchedulingInfo()
	startRequest := auctioneer.NewLRPStartRequestFromSchedulingInfo(&schedInfo, int(lrpKey.Index))
	err = h.auctioneerClient.RequestLRPAuctions(logger, []*auctioneer.LRPStartRequest{&startRequest})
	if err != nil {
		logger.Error("failed-requesting-auction", err)
	}

	return nil
}
