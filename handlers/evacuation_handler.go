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

func findWithPresence(lrps []*models.ActualLRP, presence models.ActualLRP_Presence) *models.ActualLRP {
	for _, lrp := range lrps {
		if lrp.Presence == presence {
			return lrp
		}
	}
	return nil
}

func findLRP(key *models.ActualLRPInstanceKey, lrps []*models.ActualLRP) (*models.ActualLRP, bool) {
	for _, lrp := range lrps {
		if lrp.ActualLRPInstanceKey == *key {
			return lrp, lrp.Presence == models.ActualLRP_Suspect
		}
	}
	return nil, false
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

	actualLRPs, err := h.actualLRPDB.ActualLRPs(logger, models.ActualLRPFilter{ProcessGuid: request.ActualLrpKey.ProcessGuid, Index: &request.ActualLrpKey.Index})
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	evacuatingLRPLogData := lager.Data{
		"process-guid": request.ActualLrpKey.ProcessGuid,
		"index":        request.ActualLrpKey.Index,
		"instance-key": request.ActualLrpInstanceKey,
	}

	instance := findWithPresence(actualLRPs, models.ActualLRP_Ordinary)
	suspect := findWithPresence(actualLRPs, models.ActualLRP_Suspect)
	instance = getHigherPriorityActualLRP(instance, suspect)

	if instance != nil {
		evacuatingLRPLogData["replacement-lrp-instance-key"] = instance.ActualLRPInstanceKey
		evacuatingLRPLogData["replacement-state"] = instance.State
		evacuatingLRPLogData["replacement-lrp-placement-error"] = instance.PlacementError
	}

	logger.Info("removing-stranded-evacuating-actual-lrp", evacuatingLRPLogData)

	err = h.db.RemoveEvacuatingActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	evacuating := findWithPresence(actualLRPs, models.ActualLRP_Evacuating)
	if evacuating == nil {
		response.Error = models.ConvertError(models.ErrResourceNotFound)
		return
	}

	go h.actualHub.Emit(models.NewActualLRPRemovedEvent(evacuating.ToActualLRPGroup()))
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

	guid := request.ActualLrpKey.ProcessGuid
	index := request.ActualLrpKey.Index

	actualLRPs, err := h.actualLRPDB.ActualLRPs(logger, models.ActualLRPFilter{ProcessGuid: guid, Index: &index})
	if err != nil {
		logger.Error("failed-querying-actualLRPs", err, lager.Data{"guid": guid, "indec": index})
		response.Error = models.ConvertError(err)
		return
	}

	// TODO: check if it is ok to return errors here
	targetActualLRP, _ := findLRP(request.ActualLrpInstanceKey, actualLRPs)
	if targetActualLRP == nil {
		logger.Debug("actual-lrp-not-found", lager.Data{"guid": guid, "index": index})
		response.Error = models.ErrResourceNotFound
		return
	}

	evacuating := findWithPresence(actualLRPs, models.ActualLRP_Evacuating)
	suspect := findWithPresence(actualLRPs, models.ActualLRP_Suspect)
	ordinary := findWithPresence(actualLRPs, models.ActualLRP_Ordinary)

	if evacuating != nil {
		err = h.db.RemoveEvacuatingActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey)
		if err != nil {
			logger.Error("failed-removing-evacuating-actual-lrp", err)
			exitIfUnrecoverable(logger, h.exitChan, models.ConvertError(err))
		}
		events = append(events, models.NewActualLRPRemovedEvent(evacuating.ToActualLRPGroup()))
	}

	if ordinary != nil && targetActualLRP.Equal(suspect) {
		h.actualLRPDB.RemoveActualLRP(logger, guid, index, request.ActualLrpInstanceKey)
		response.KeepContainer = false
		return
	}

	before, after, err := h.actualLRPDB.UnclaimActualLRP(logger, request.ActualLrpKey)
	if err != nil {
		bbsErr := models.ConvertError(err)
		if bbsErr != nil && bbsErr.Type != models.Error_ResourceNotFound {
			response.Error = bbsErr
			response.KeepContainer = true
		}
		return
	}
	err = h.requestAuction(logger, request.ActualLrpKey)
	bbsErr := models.ConvertError(err)
	if bbsErr != nil && bbsErr.Type != models.Error_ResourceNotFound {
		response.Error = bbsErr
		response.KeepContainer = true
		return
	}

	go func() {
		if suspect == nil || targetActualLRP.Equal(suspect) {
			events = append(events, models.NewActualLRPChangedEvent(before.ToActualLRPGroup(), after.ToActualLRPGroup()))
		}
	}()
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

	guid := request.ActualLrpKey.ProcessGuid
	index := request.ActualLrpKey.Index
	actualLRPs, err := h.actualLRPDB.ActualLRPs(logger, models.ActualLRPFilter{ProcessGuid: guid, Index: &index})

	if err != nil {
		logger.Error("failed-querying-actualLRPs", err, lager.Data{"guid": guid, "indec": index})
		response.Error = models.ConvertError(err)
		return
	}

	targetActualLRP, _ := findLRP(request.ActualLrpInstanceKey, actualLRPs)
	if targetActualLRP != nil && targetActualLRP.Presence == models.ActualLRP_Suspect {
		suspect, err := h.suspectLRPDB.RemoveSuspectActualLRP(logger, request.ActualLrpKey)
		if err != nil {
			logger.Error("failed-removing-suspect-actual-lrp", err)
			response.Error = models.ConvertError(err)
		} else {
			go h.actualHub.Emit(models.NewActualLRPRemovedEvent(suspect.ToActualLRPGroup()))
		}
		return
	}

	// try removing the evacuating instance if present
	err = h.db.RemoveEvacuatingActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey)
	if err != nil {
		logger.Error("failed-removing-evacuating-actual-lrp", err)
		exitIfUnrecoverable(logger, h.exitChan, models.ConvertError(err))
	} else {
		evacuating := findWithPresence(actualLRPs, models.ActualLRP_Evacuating)
		go h.actualHub.Emit(models.NewActualLRPRemovedEvent(evacuating.ToActualLRPGroup()))
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
	if err != nil {
		logger.Error("failed-fetching-lrp-group", err)
		response.Error = models.ConvertError(err)
		return
	}

	if len(actualLRPs) == 0 {
		response.KeepContainer = false
		return
	}

	evacuateRequesting := func(request *models.EvacuateRunningActualLRPRequest) error {
		evacuating, err := h.db.EvacuateActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey, request.ActualLrpNetInfo)
		if err == models.ErrActualLRPCannotBeEvacuated {
			logger.Error("cannot-evacuate-actual-lrp", err)
			return err
		}

		if err == models.ErrResourceExists {
			return err
		}

		if err != nil {
			logger.Error("failed-evacuating-actual-lrp", err)
			response.Error = models.ConvertError(err)
		}
		go h.actualHub.Emit(models.NewActualLRPCreatedEvent(evacuating.ToActualLRPGroup()))
		return nil
	}

	evacuateInstance := func(actualLRP *models.ActualLRP) {
		evacuating, err := h.db.EvacuateActualLRP(logger, &actualLRP.ActualLRPKey, &actualLRP.ActualLRPInstanceKey, &actualLRP.ActualLRPNetInfo)
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

		if actualLRP.Presence == models.ActualLRP_Suspect {
			suspect, err := h.suspectLRPDB.RemoveSuspectActualLRP(logger, &actualLRP.ActualLRPKey)
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

		before, after, err := h.actualLRPDB.UnclaimActualLRP(logger, &actualLRP.ActualLRPKey)
		if err != nil {
			response.Error = models.ConvertError(err)
			return
		}

		events = append(events, models.NewActualLRPChangedEvent(before.ToActualLRPGroup(), after.ToActualLRPGroup()))

		err = h.requestAuction(logger, &actualLRP.ActualLRPKey)
		if err != nil {
			response.Error = models.ConvertError(err)
			return
		}
	}

	removeEvacuating := func(evacuating *models.ActualLRP) error {
		if evacuating == nil {
			return nil
		}
		err = h.db.RemoveEvacuatingActualLRP(logger, &evacuating.ActualLRPKey, &evacuating.ActualLRPInstanceKey)
		if err == nil {
			go h.actualHub.Emit(models.NewActualLRPRemovedEvent(&models.ActualLRPGroup{Evacuating: evacuating}))
		}
		if err != nil && err != models.ErrActualLRPCannotBeRemoved {
			return err
		}
		return nil
	}

	targetActualLRP, _ := findLRP(request.ActualLrpInstanceKey, actualLRPs)
	evacuating := findWithPresence(actualLRPs, models.ActualLRP_Evacuating)
	instance := findWithPresence(actualLRPs, models.ActualLRP_Ordinary)
	suspect := findWithPresence(actualLRPs, models.ActualLRP_Suspect)
	instance = getHigherPriorityActualLRP(instance, suspect)

	if instance == nil {
		if targetActualLRP != nil && targetActualLRP.Equal(evacuating) {
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
	}

	switch instance.State {
	case models.ActualLRPStateUnclaimed:
		response.KeepContainer = true
		if instance.PlacementError == "" {
			if evacuating != nil && !evacuating.Equal(targetActualLRP) {
				logger.Error("already-evacuated-by-different-cell", err)
				response.KeepContainer = false
				return
			}
			err := evacuateRequesting(request)
			if err == models.ErrActualLRPCannotBeEvacuated {
				response.KeepContainer = false
			}
		}
	case models.ActualLRPStateClaimed:
		response.KeepContainer = true
		if !instance.Equal(targetActualLRP) {
			if evacuating != nil && !evacuating.Equal(targetActualLRP) {
				response.KeepContainer = false
				logger.Error("already-evacuated-by-different-cell", err)
				return
			}
			err := evacuateRequesting(request)
			if err == models.ErrActualLRPCannotBeEvacuated {
				response.KeepContainer = false
			}
			return
		}
		evacuateInstance(instance)
	case models.ActualLRPStateRunning:
		if !instance.Equal(targetActualLRP) {
			response.KeepContainer = false
			err := removeEvacuating(evacuating)
			if err != nil {
				response.KeepContainer = true
				response.Error = models.ConvertError(err)
			}
			return
		}
		evacuateInstance(instance)
	case models.ActualLRPStateCrashed:
		response.KeepContainer = false
		err := removeEvacuating(evacuating)
		if err != nil {
			response.KeepContainer = true
			response.Error = models.ConvertError(err)
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

	targetActualLRP, _ := findLRP(request.ActualLrpInstanceKey, actualLRPs)
	if targetActualLRP == nil {
		logger.Debug("actual-lrp-not-found", lager.Data{"guid": guid, "index": index})
		response.Error = models.ErrActualLRPCannotBeRemoved
		return
	}

	switch targetActualLRP.Presence {
	case models.ActualLRP_Evacuating:
		err = h.db.RemoveEvacuatingActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey)
		if err != nil {
			logger.Error("failed-removing-evacuating-actual-lrp", err)
			bbsErr = models.ConvertError(err)
			return
		}
		go h.actualHub.Emit(models.NewActualLRPRemovedEvent(targetActualLRP.ToActualLRPGroup()))
	case models.ActualLRP_Suspect:
		suspect, err := h.suspectLRPDB.RemoveSuspectActualLRP(logger, request.ActualLrpKey)
		if err != nil {
			logger.Error("failed-removing-suspect-actual-lrp", err)
			bbsErr = models.ConvertError(err)
			response.Error = bbsErr
			return
		}
		go h.actualHub.Emit(models.NewActualLRPRemovedEvent(suspect.ToActualLRPGroup()))
	case models.ActualLRP_Ordinary:
		err = h.actualLRPDB.RemoveActualLRP(logger, guid, index, request.ActualLrpInstanceKey)
		if err != nil {
			logger.Error("failed-to-remove-actual-lrp", err)
			bbsErr = models.ConvertError(err)
			response.Error = bbsErr
			return
		}
		go h.actualHub.Emit(models.NewActualLRPRemovedEvent(targetActualLRP.ToActualLRPGroup()))
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
