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

func normalLRP(lrps []*models.ActualLRP) *models.ActualLRP {
	for _, lrp := range lrps {
		if lrp.PlacementState == models.PlacementStateType_Normal || lrp.PlacementState == models.PlacementStateType_Suspect {
			return lrp
		}
	}
	return nil
}
func evacuatingLRP(lrps []*models.ActualLRP) *models.ActualLRP {
	for _, lrp := range lrps {
		if lrp.PlacementState == models.PlacementStateType_Evacuating {
			return lrp
		}
	}
	return nil
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

	beforeActualLRPs, err := h.actualLRPDB.ActualLRPs(logger, models.ActualLRPFilter{
		ProcessGUID: &request.ActualLrpKey.ProcessGuid,
		Index:       &request.ActualLrpKey.Index,
	})
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	evacuatingLRPLogData := lager.Data{
		"process-guid": request.ActualLrpKey.ProcessGuid,
		"index":        request.ActualLrpKey.Index,
		"instance-key": request.ActualLrpInstanceKey,
	}

	evacuatingLRP := evacuatingLRP(beforeActualLRPs)
	if evacuatingLRP != nil {
		evacuatingLRPLogData["replacement-lrp-instance-key"] = evacuatingLRP.ActualLRPInstanceKey
		evacuatingLRPLogData["replacement-state"] = evacuatingLRP.State
		evacuatingLRPLogData["replacement-lrp-placement-error"] = evacuatingLRP.PlacementError
	}

	logger.Info("removing-stranded-evacuating-actual-lrp", evacuatingLRPLogData)

	err = h.db.RemoveEvacuatingActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	if evacuatingLRP == nil {
		logger.Info("evacuating-lrp-is-emtpy")
		response.Error = models.ConvertError(models.ErrResourceNotFound)
		return
	}

	go h.actualHub.Emit(models.NewFlattenedActualLRPRemovedEvent(evacuatingLRP))
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

	beforeActualLRPs, err := h.actualLRPDB.ActualLRPs(logger, models.ActualLRPFilter{
		ProcessGUID: &request.ActualLrpKey.ProcessGuid,
		Index:       &request.ActualLrpKey.Index,
	})
	if err == nil {
		lrp := evacuatingLRP(beforeActualLRPs)
		if lrp != nil {
			err = h.db.RemoveEvacuatingActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey)
			if err != nil {
				logger.Error("failed-removing-evacuating-actual-lrp", err)
				exitIfUnrecoverable(logger, h.exitChan, models.ConvertError(err))
			} else {
				go h.actualHub.Emit(models.NewFlattenedActualLRPRemovedEvent(lrp))
			}
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

	beforeActualLRPs, err := h.actualLRPDB.ActualLRPs(logger, models.ActualLRPFilter{
		ProcessGUID: &request.ActualLrpKey.ProcessGuid,
		Index:       &request.ActualLrpKey.Index,
	})
	if err == nil {
		lrp := evacuatingLRP(beforeActualLRPs)
		if lrp != nil {
			err = h.db.RemoveEvacuatingActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey)
			if err != nil {
				logger.Error("failed-removing-evacuating-actual-lrp", err)
				exitIfUnrecoverable(logger, h.exitChan, models.ConvertError(err))
			} else {
				go h.actualHub.Emit(models.NewFlattenedActualLRPRemovedEvent(lrp))
			}
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

	actualLRPs, err := h.actualLRPDB.ActualLRPs(logger, models.ActualLRPFilter{
		ProcessGUID: &request.ActualLrpKey.ProcessGuid,
		Index:       &request.ActualLrpKey.Index,
	})
	if err != nil {
		if err == models.ErrResourceNotFound {
			response.KeepContainer = false
			return
		}
		logger.Error("failed-fetching-lrp-group", err)
		response.Error = models.ConvertError(err)
		return
	}

	normal := normalLRP(actualLRPs)
	evacuating := evacuatingLRP(actualLRPs)
	// If the instance is not there, clean up the corresponding evacuating LRP, if one exists.
	if normal == nil {
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

		if evacuating != nil {
			go h.actualHub.Emit(models.NewFlattenedActualLRPRemovedEvent(evacuating))
		}
		response.KeepContainer = false
		return
	}

	if (normal.State == models.ActualLRPStateUnclaimed && normal.PlacementError == "") ||
		(normal.State == models.ActualLRPStateClaimed && !normal.ActualLRPInstanceKey.Equal(request.ActualLrpInstanceKey)) {
		if evacuating != nil && !evacuating.ActualLRPInstanceKey.Equal(request.ActualLrpInstanceKey) {
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

	if (normal.State == models.ActualLRPStateRunning && !normal.ActualLRPInstanceKey.Equal(request.ActualLrpInstanceKey)) ||
		normal.State == models.ActualLRPStateCrashed {
		response.KeepContainer = false

		// if there is not evacuating instance, it probably got removed when the
		// new instance transitioned to a Running state
		if evacuating == nil {
			return
		}

		err = h.db.RemoveEvacuatingActualLRP(logger, &evacuating.ActualLRPKey, &evacuating.ActualLRPInstanceKey)
		if err == nil {
			go h.actualHub.Emit(models.NewFlattenedActualLRPRemovedEvent(evacuating))
		}
		if err != nil && err != models.ErrActualLRPCannotBeRemoved {
			response.KeepContainer = true
			response.Error = models.ConvertError(err)
		}
		return
	}

	if (normal.State == models.ActualLRPStateClaimed || normal.State == models.ActualLRPStateRunning) &&
		normal.ActualLRPInstanceKey.Equal(request.ActualLrpInstanceKey) {
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
	actualLRPs, err := h.actualLRPDB.ActualLRPs(logger, models.ActualLRPFilter{
		ProcessGUID: &guid,
		Index:       &index,
	})

	if err != nil {
		logger.Error("failed-fetching-actual-lrp-group", err)
		bbsErr = models.ConvertError(err)
		response.Error = bbsErr
		return
	}

	normal := normalLRP(actualLRPs)
	evacuating := evacuatingLRP(actualLRPs)

	err = h.db.RemoveEvacuatingActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey)
	if err != nil {
		logger.Error("failed-removing-evacuating-actual-lrp", err)
		bbsErr = models.ConvertError(err)
	} else if evacuating == nil {
		go h.actualHub.Emit(models.NewFlattenedActualLRPRemovedEvent(evacuating))
	}

	if normal == nil || !normal.ActualLRPInstanceKey.Equal(request.ActualLrpInstanceKey) {
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
		if normal != nil {
			go h.actualHub.Emit(models.NewFlattenedActualLRPRemovedEvent(normal))
		}
	}
}

func (h *EvacuationHandler) unclaimAndRequestAuction(logger lager.Logger, lrpKey *models.ActualLRPKey) error {
	before, after, err := h.actualLRPDB.UnclaimActualLRP(logger, lrpKey)
	if err != nil {
		return err
	}

	go h.actualHub.Emit(models.NewFlattenedActualLRPChangedEvent(before, after))

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
