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
	suspectLRPDB     db.SuspectDB
	actualHub        events.Hub
	auctioneerClient auctioneer.Client
	exitChan         chan<- struct{}
}

func NewEvacuationHandler(
	db db.EvacuationDB,
	actualLRPDB db.ActualLRPDB,
	desiredLRPDB db.DesiredLRPDB,
	suspectLRPDB db.SuspectDB,
	actualHub events.Hub,
	auctioneerClient auctioneer.Client,
	exitChan chan<- struct{},
) *EvacuationHandler {
	return &EvacuationHandler{
		db:               db,
		actualLRPDB:      actualLRPDB,
		desiredLRPDB:     desiredLRPDB,
		suspectLRPDB:     suspectLRPDB,
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

	beforeActualLRPs, err := h.actualLRPDB.ActualLRPs(logger, models.ActualLRPFilter{ProcessGuid: request.ActualLrpKey.ProcessGuid, Index: &request.ActualLrpKey.Index})
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	evacuatingLRPLogData := lager.Data{
		"process-guid": request.ActualLrpKey.ProcessGuid,
		"index":        request.ActualLrpKey.Index,
		"instance-key": request.ActualLrpInstanceKey,
	}

	beforeActualLRPGroup := resolveToActualLRPGroup(beforeActualLRPs)
	if beforeActualLRPGroup.Instance != nil {
		evacuatingLRPLogData["replacement-lrp-instance-key"] = beforeActualLRPGroup.Instance.ActualLRPInstanceKey
		evacuatingLRPLogData["replacement-state"] = beforeActualLRPGroup.Instance.State
		evacuatingLRPLogData["replacement-lrp-placement-error"] = beforeActualLRPGroup.Instance.PlacementError
	}

	logger.Info("removing-stranded-evacuating-actual-lrp", evacuatingLRPLogData)

	err = h.db.RemoveEvacuatingActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	if beforeActualLRPGroup.Evacuating == nil {
		logger.Info("evacuating-lrp-is-emtpy")
		response.Error = models.ConvertError(models.ErrResourceNotFound)
		return
	}

	actualLRPGroup := &models.ActualLRPGroup{
		Evacuating: beforeActualLRPGroup.Evacuating,
	}

	go h.actualHub.Emit(models.NewActualLRPRemovedEvent(actualLRPGroup))
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

	events := []models.Event{}
	defer func() {
		go func() {
			for _, event := range events {
				h.actualHub.Emit(event)
			}
		}()
	}()

	beforeActualLRPs, err := h.actualLRPDB.ActualLRPs(logger, models.ActualLRPFilter{ProcessGuid: request.ActualLrpKey.ProcessGuid, Index: &request.ActualLrpKey.Index})
	beforeActualLRPGroup := resolveToActualLRPGroup(beforeActualLRPs)

	// remove any existing evacuating LRP
	if beforeActualLRPGroup.Evacuating != nil {
		lrp := beforeActualLRPGroup.Evacuating
		err = h.db.RemoveEvacuatingActualLRP(logger, &lrp.ActualLRPKey, &lrp.ActualLRPInstanceKey)
		if err != nil {
			logger.Error("failed-removing-evacuating-actual-lrp", err)
			exitIfUnrecoverable(logger, h.exitChan, models.ConvertError(err))
		} else {
			events = append(events, models.NewActualLRPRemovedEvent(beforeActualLRPGroup.Evacuating.ToActualLRPGroup()))
		}
	}

	before, after, err := h.actualLRPDB.UnclaimActualLRP(logger, request.ActualLrpKey)
	if err == nil && beforeActualLRPGroup.Instance.Presence != models.ActualLRP_Suspect {
		// only emit the event if there is no suspect LRP running and the LRP was indeed transitioned to UNCLAIMED
		events = append(events, models.NewActualLRPChangedEvent(before.ToActualLRPGroup(), after.ToActualLRPGroup()))
	}
	bbsErr := models.ConvertError(err)
	if bbsErr != nil && bbsErr.Type != models.Error_ResourceNotFound {
		response.Error = bbsErr
		response.KeepContainer = true
		return
	}

	err = h.requestAuction(logger, request.ActualLrpKey)
	bbsErr = models.ConvertError(err)
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

	beforeActualLRPs, err := h.actualLRPDB.ActualLRPs(logger, models.ActualLRPFilter{ProcessGuid: request.ActualLrpKey.ProcessGuid, Index: &request.ActualLrpKey.Index})
	beforeActualLRPGroup := resolveToActualLRPGroup(beforeActualLRPs)

	// check if this is the suspect LRP
	instance := beforeActualLRPGroup.Instance
	if err == nil &&
		instance != nil &&
		instance.Presence == models.ActualLRP_Suspect &&
		instance.ActualLRPInstanceKey == *request.ActualLrpInstanceKey {
		suspect, err := h.suspectLRPDB.RemoveSuspectActualLRP(logger, request.ActualLrpKey)
		if err != nil {
			logger.Error("failed-removing-suspect-actual-lrp", err)
			response.Error = models.ConvertError(err)
		} else {
			go h.actualHub.Emit(models.NewActualLRPRemovedEvent(suspect.ToActualLRPGroup()))
		}
		return
	}

	if err == nil {
		err = h.db.RemoveEvacuatingActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey)
		if err != nil {
			logger.Error("failed-removing-evacuating-actual-lrp", err)
			exitIfUnrecoverable(logger, h.exitChan, models.ConvertError(err))
		} else {
			go h.actualHub.Emit(models.NewActualLRPRemovedEvent(&models.ActualLRPGroup{Evacuating: beforeActualLRPGroup.Evacuating}))
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
	actualLRPs, err := h.actualLRPDB.ActualLRPs(logger, models.ActualLRPFilter{ProcessGuid: guid, Index: &index})
	if len(actualLRPs) == 0 {
		response.KeepContainer = false
		return
	}
	if err != nil {
		logger.Error("failed-fetching-lrp-group", err)
		response.Error = models.ConvertError(err)
		return
	}

	actualLRPGroup := resolveToActualLRPGroup(actualLRPs)
	instance := actualLRPGroup.Instance
	evacuating := actualLRPGroup.Evacuating

	// If the instance is not there, clean up the corresponding evacuating LRP, if one exists.
	if instance == nil {
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

		go h.actualHub.Emit(models.NewActualLRPRemovedEvent(&models.ActualLRPGroup{Evacuating: evacuating}))
		response.KeepContainer = false
		return
	}

	if (instance.State == models.ActualLRPStateUnclaimed && instance.PlacementError == "") ||
		(instance.State == models.ActualLRPStateClaimed && !instance.ActualLRPInstanceKey.Equal(request.ActualLrpInstanceKey)) {
		if evacuating != nil && !evacuating.ActualLRPInstanceKey.Equal(request.ActualLrpInstanceKey) {
			logger.Error("already-evacuated-by-different-cell", err)
			response.KeepContainer = false
			return
		}

		evacuating, err := h.db.EvacuateActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey, request.ActualLrpNetInfo)
		if err == models.ErrActualLRPCannotBeEvacuated {
			logger.Error("cannot-evacuate-actual-lrp", err)
			response.KeepContainer = false
			return
		}

		response.KeepContainer = true
		if err == models.ErrResourceExists {
			return
		}

		if err != nil {
			logger.Error("failed-evacuating-actual-lrp", err)
			response.Error = models.ConvertError(err)
		} else {
			go h.actualHub.Emit(models.NewActualLRPCreatedEvent(evacuating.ToActualLRPGroup()))
		}

		return
	}

	if (instance.State == models.ActualLRPStateRunning && !instance.ActualLRPInstanceKey.Equal(request.ActualLrpInstanceKey)) ||
		instance.State == models.ActualLRPStateCrashed {
		response.KeepContainer = false

		// if there is not evacuating instance, it probably got removed when the
		// new instance transitioned to a Running state
		if evacuating == nil {
			return
		}

		err = h.db.RemoveEvacuatingActualLRP(logger, &evacuating.ActualLRPKey, &evacuating.ActualLRPInstanceKey)
		if err == nil {
			go h.actualHub.Emit(models.NewActualLRPRemovedEvent(&models.ActualLRPGroup{Evacuating: evacuating}))
		}
		if err != nil && err != models.ErrActualLRPCannotBeRemoved {
			response.KeepContainer = true
			response.Error = models.ConvertError(err)
		}
		return
	}

	if (instance.State == models.ActualLRPStateClaimed || instance.State == models.ActualLRPStateRunning) &&
		instance.ActualLRPInstanceKey.Equal(request.ActualLrpInstanceKey) {
		evacuating, err := h.db.EvacuateActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey, request.ActualLrpNetInfo)
		if err != nil {
			response.Error = models.ConvertError(err)
			return
		}

		events := []models.Event{}
		defer func() {
			go func() {
				for _, event := range events {
					h.actualHub.Emit(event)
				}
			}()
		}()

		events = append(events, models.NewActualLRPCreatedEvent(evacuating.ToActualLRPGroup()))

		if instance.Presence == models.ActualLRP_Suspect {
			suspect, err := h.suspectLRPDB.RemoveSuspectActualLRP(logger, request.ActualLrpKey)
			if err != nil {
				logger.Error("failed-removing-suspect-actual-lrp", err)
				response.Error = models.ConvertError(err)
				return
			}

			// after removing the running suspect instance, if the replacement instance is claimed we can now
			// emit a created event since this instance is taking over from the evacuating one
			for _, lrp := range actualLRPs {
				if lrp.State == models.ActualLRPStateClaimed {
					events = append(events, models.NewActualLRPCreatedEvent(lrp.ToActualLRPGroup()))
				}
			}

			events = append(events, models.NewActualLRPRemovedEvent(suspect.ToActualLRPGroup()))
			return
		}

		before, after, err := h.actualLRPDB.UnclaimActualLRP(logger, request.ActualLrpKey)
		if err != nil {
			response.Error = models.ConvertError(err)
			return
		}

		events = append(events, models.NewActualLRPChangedEvent(before.ToActualLRPGroup(), after.ToActualLRPGroup()))

		err = h.requestAuction(logger, request.ActualLrpKey)
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

	actualLRPs, err := h.actualLRPDB.ActualLRPs(logger, models.ActualLRPFilter{ProcessGuid: guid, Index: &index})
	if err != nil {
		logger.Error("failed-fetching-actual-lrp-group", err)
		bbsErr = models.ConvertError(err)
		response.Error = bbsErr
		return
	}

	actualLRPGroup := resolveToActualLRPGroup(actualLRPs)

	// check if this is the suspect LRP
	if err == nil &&
		actualLRPGroup.Instance != nil &&
		actualLRPGroup.Instance.Presence == models.ActualLRP_Suspect &&
		actualLRPGroup.Instance.ActualLRPInstanceKey == *request.ActualLrpInstanceKey {
		suspect, err := h.suspectLRPDB.RemoveSuspectActualLRP(logger, request.ActualLrpKey)
		if err != nil {
			logger.Error("failed-removing-suspect-actual-lrp", err)
			bbsErr = models.ConvertError(err)
			response.Error = bbsErr
			return
		}
		go h.actualHub.Emit(models.NewActualLRPRemovedEvent(suspect.ToActualLRPGroup()))
		return
	}

	err = h.db.RemoveEvacuatingActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey)
	if err != nil {
		logger.Error("failed-removing-evacuating-actual-lrp", err)
		bbsErr = models.ConvertError(err)
	} else if actualLRPGroup.Evacuating != nil {
		go h.actualHub.Emit(models.NewActualLRPRemovedEvent(&models.ActualLRPGroup{Evacuating: actualLRPGroup.Evacuating}))
	}

	if actualLRPGroup.Instance == nil || !actualLRPGroup.Instance.ActualLRPInstanceKey.Equal(request.ActualLrpInstanceKey) {
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
		go h.actualHub.Emit(models.NewActualLRPRemovedEvent(&models.ActualLRPGroup{Instance: actualLRPGroup.Instance}))
	}
}

func (h *EvacuationHandler) requestAuction(logger lager.Logger, lrpKey *models.ActualLRPKey) error {
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
