package controllers

import (
	"code.cloudfoundry.org/auctioneer"
	"code.cloudfoundry.org/bbs/db"
	"code.cloudfoundry.org/bbs/events"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

type EvacuationController struct {
	db                   db.EvacuationDB
	actualLRPDB          db.ActualLRPDB
	suspectLRPDB         db.SuspectDB
	desiredLRPDB         db.DesiredLRPDB
	auctioneerClient     auctioneer.Client
	actualHub            events.Hub
	actualLRPInstanceHub events.Hub
}

func NewEvacuationController(
	db db.EvacuationDB,
	actualLRPDB db.ActualLRPDB,
	suspectLRPDB db.SuspectDB,
	desiredLRPDB db.DesiredLRPDB,
	auctioneerClient auctioneer.Client,
	actualHub events.Hub,
	actualLRPInstanceHub events.Hub,
) *EvacuationController {
	return &EvacuationController{
		db:                   db,
		actualLRPDB:          actualLRPDB,
		suspectLRPDB:         suspectLRPDB,
		desiredLRPDB:         desiredLRPDB,
		auctioneerClient:     auctioneerClient,
		actualHub:            actualHub,
		actualLRPInstanceHub: actualLRPInstanceHub,
	}
}

func (h *EvacuationController) RemoveEvacuatingActualLRP(logger lager.Logger, actualLRPKey *models.ActualLRPKey, actualLRPInstanceKey *models.ActualLRPInstanceKey) error {
	actualLRPs, err := h.actualLRPDB.ActualLRPs(logger, models.ActualLRPFilter{ProcessGuid: actualLRPKey.ProcessGuid, Index: &actualLRPKey.Index})
	if err != nil {
		return err
	}

	evacuatingLRPLogData := lager.Data{
		"process-guid": actualLRPKey.ProcessGuid,
		"index":        actualLRPKey.Index,
		"instance-key": actualLRPInstanceKey,
	}

	instance := findWithPresence(actualLRPs, models.ActualLRP_Ordinary)
	lrpGroup := models.ResolveActualLRPGroup(actualLRPs)
	instance = lrpGroup.Instance

	if instance != nil {
		evacuatingLRPLogData["replacement-lrp-instance-key"] = instance.ActualLRPInstanceKey
		evacuatingLRPLogData["replacement-state"] = instance.State
		evacuatingLRPLogData["replacement-lrp-placement-error"] = instance.PlacementError
	}

	logger.Info("removing-stranded-evacuating-actual-lrp", evacuatingLRPLogData)

	err = h.db.RemoveEvacuatingActualLRP(logger, actualLRPKey, actualLRPInstanceKey)
	if err != nil {
		return err
	}

	evacuating := findWithPresence(actualLRPs, models.ActualLRP_Evacuating)
	if evacuating == nil {
		logger.Info("evacuating-lrp-is-emtpy")
		return models.ErrResourceNotFound
	}

	go h.actualHub.Emit(models.NewActualLRPRemovedEvent(evacuating.ToActualLRPGroup()))
	go h.actualLRPInstanceHub.Emit(models.NewActualLRPInstanceRemovedEvent(evacuating))
	return nil
}

func (h *EvacuationController) EvacuateClaimedActualLRP(logger lager.Logger, actualLRPKey *models.ActualLRPKey, actualLRPInstanceKey *models.ActualLRPInstanceKey) (error, bool) {

	eventCalculator := EventCalculator{
		ActualLRPGroupHub:    h.actualHub,
		ActualLRPInstanceHub: h.actualLRPInstanceHub,
	}

	guid := actualLRPKey.ProcessGuid
	index := actualLRPKey.Index
	actualLRPs, err := h.actualLRPDB.ActualLRPs(logger, models.ActualLRPFilter{ProcessGuid: guid, Index: &index})
	if err != nil {
		logger.Error("failed-querying-actualLRPs", err, lager.Data{"guid": guid, "index": index})
		return err, false
	}

	newLRPs := make([]*models.ActualLRP, len(actualLRPs))
	copy(newLRPs, actualLRPs)

	defer func() {
		go eventCalculator.EmitEvents(actualLRPs, newLRPs)
	}()

	targetActualLRP := lookupLRPInSlice(actualLRPs, actualLRPInstanceKey)
	if targetActualLRP == nil {
		logger.Debug("actual-lrp-not-found", lager.Data{"guid": guid, "index": index})
		return models.ErrResourceNotFound, false
	}

	switch targetActualLRP.Presence {
	case models.ActualLRP_Evacuating:
		err = h.db.RemoveEvacuatingActualLRP(logger, actualLRPKey, actualLRPInstanceKey)
		if err != nil {
			return err, false
		}

		newLRPs = eventCalculator.RecordChange(targetActualLRP, nil, newLRPs)
	case models.ActualLRP_Suspect:
		_, err = h.suspectLRPDB.RemoveSuspectActualLRP(logger, actualLRPKey)
		if err != nil {
			return err, false
		}

		newLRPs = eventCalculator.RecordChange(targetActualLRP, nil, newLRPs)
	case models.ActualLRP_Ordinary:
		before, after, err := h.actualLRPDB.UnclaimActualLRP(logger, actualLRPKey)
		bbsErr := models.ConvertError(err)
		if bbsErr != nil {
			if bbsErr.Type == models.Error_ResourceNotFound {
				return nil, false
			}
			return bbsErr, true
		}

		newLRPs = eventCalculator.RecordChange(before, after, newLRPs)

		h.requestAuction(logger, actualLRPKey)
	}

	return nil, false
}

func (h *EvacuationController) EvacuateCrashedActualLRP(logger lager.Logger, actualLRPKey *models.ActualLRPKey, actualLRPInstanceKey *models.ActualLRPInstanceKey, errorMessage string) error {
	eventCalculator := EventCalculator{
		ActualLRPGroupHub:    h.actualHub,
		ActualLRPInstanceHub: h.actualLRPInstanceHub,
	}

	guid := actualLRPKey.ProcessGuid
	index := actualLRPKey.Index

	actualLRPs, err := h.actualLRPDB.ActualLRPs(logger, models.ActualLRPFilter{ProcessGuid: guid, Index: &index})
	if err != nil {
		logger.Error("failed-fetching-actual-lrps", err)
		return err
	}

	newLRPs := make([]*models.ActualLRP, len(actualLRPs))
	copy(newLRPs, actualLRPs)

	defer func() {
		go eventCalculator.EmitEvents(actualLRPs, newLRPs)
	}()

	targetActualLRP := lookupLRPInSlice(actualLRPs, actualLRPInstanceKey)
	if targetActualLRP == nil {
		logger.Debug("actual-lrp-not-found", lager.Data{"guid": guid, "index": index})
		return models.ErrResourceNotFound
	}

	switch targetActualLRP.Presence {
	case models.ActualLRP_Evacuating:
		err = h.db.RemoveEvacuatingActualLRP(logger, actualLRPKey, actualLRPInstanceKey)
		if err != nil {
			logger.Error("failed-removing-evacuating-actual-lrp", err)
			return err
		}
		newLRPs = eventCalculator.RecordChange(targetActualLRP, nil, newLRPs)
	case models.ActualLRP_Suspect:
		_, err := h.suspectLRPDB.RemoveSuspectActualLRP(logger, actualLRPKey)
		if err != nil {
			logger.Error("failed-removing-suspect-actual-lrp", err)
			return err
		}
		newLRPs = eventCalculator.RecordChange(targetActualLRP, nil, newLRPs)
	case models.ActualLRP_Ordinary:
		_, _, _, err = h.actualLRPDB.CrashActualLRP(logger, actualLRPKey, actualLRPInstanceKey, errorMessage)
		if err != nil {
			logger.Error("failed-to-crash-actual-lrp", err)
			return err
		}
	}

	return nil
}

func (h *EvacuationController) EvacuateRunningActualLRP(logger lager.Logger, actualLRPKey *models.ActualLRPKey, actualLRPInstanceKey *models.ActualLRPInstanceKey, netInfo *models.ActualLRPNetInfo) (error, bool) {
	guid := actualLRPKey.ProcessGuid
	index := actualLRPKey.Index
	actualLRPs, err := h.actualLRPDB.ActualLRPs(logger, models.ActualLRPFilter{ProcessGuid: guid, Index: &index})
	if err != nil {
		logger.Error("failed-fetching-lrp-group", err)
		return err, true
	}

	if len(actualLRPs) == 0 {
		return nil, false
	}

	targetActualLRP := lookupLRPInSlice(actualLRPs, actualLRPInstanceKey)
	evacuating := findWithPresence(actualLRPs, models.ActualLRP_Evacuating)
	instance := findWithPresence(actualLRPs, models.ActualLRP_Ordinary)

	lrpGroup := models.ResolveActualLRPGroup(actualLRPs)
	instance = lrpGroup.Instance

	if instance == nil {
		if targetActualLRP != nil && targetActualLRP.Equal(evacuating) {
			err = h.db.RemoveEvacuatingActualLRP(logger, actualLRPKey, actualLRPInstanceKey)
			if err != nil {
				if err == models.ErrActualLRPCannotBeRemoved {
					logger.Debug("remove-evacuating-actual-lrp-failed")
					return nil, false
				}
				logger.Error("failed-removing-evacuating-actual-lrp", err)
				return err, true
			}

			go h.actualHub.Emit(models.NewActualLRPRemovedEvent(&models.ActualLRPGroup{Evacuating: evacuating}))
			go h.actualLRPInstanceHub.Emit(models.NewActualLRPInstanceRemovedEvent(evacuating))
			return nil, false
		}
	}

	switch instance.State {
	case models.ActualLRPStateUnclaimed:
		if instance.PlacementError == "" {
			if evacuating != nil && !evacuating.Equal(targetActualLRP) {
				logger.Info("already-evacuated-by-different-cell")
				return nil, false
			}
			err := h.evacuateRequesting(logger, actualLRPKey, actualLRPInstanceKey, netInfo)
			switch err {
			case models.ErrActualLRPCannotBeEvacuated:
				return nil, false
			default:
				return err, true
			}
		}
		return nil, true
	case models.ActualLRPStateClaimed:
		if !instance.Equal(targetActualLRP) {
			if evacuating != nil && !evacuating.Equal(targetActualLRP) {
				logger.Info("already-evacuated-by-different-cell")
				return nil, false
			}
			err := h.evacuateRequesting(logger, actualLRPKey, actualLRPInstanceKey, netInfo)
			switch err {
			case models.ErrActualLRPCannotBeEvacuated:
				return nil, false
			case models.ErrResourceExists:
				return nil, true
			default:
				return err, true
			}
		}
		err = h.evacuateInstance(logger, actualLRPs, instance)
		return err, true
	case models.ActualLRPStateRunning:
		var err error
		if !instance.Equal(targetActualLRP) {
			err = h.removeEvacuating(logger, evacuating)
			keepContainer := err != nil
			return err, keepContainer
		}
		err = h.evacuateInstance(logger, actualLRPs, instance)
		return err, true
	case models.ActualLRPStateCrashed:
		err := h.removeEvacuating(logger, evacuating)
		keepContainer := err != nil
		return err, keepContainer
	}
	return nil, false
}

func (h *EvacuationController) EvacuateStoppedActualLRP(logger lager.Logger, actualLRPKey *models.ActualLRPKey, actualLRPInstanceKey *models.ActualLRPInstanceKey) error {
	guid := actualLRPKey.ProcessGuid
	index := actualLRPKey.Index

	actualLRPs, err := h.actualLRPDB.ActualLRPs(logger, models.ActualLRPFilter{ProcessGuid: guid, Index: &index})
	if err != nil {
		logger.Error("failed-fetching-actual-lrp-group", err)
		return err
	}

	targetActualLRP := lookupLRPInSlice(actualLRPs, actualLRPInstanceKey)
	if targetActualLRP == nil {
		logger.Debug("actual-lrp-not-found", lager.Data{"guid": guid, "index": index})
		return models.ErrResourceNotFound
	}

	switch targetActualLRP.Presence {
	case models.ActualLRP_Evacuating:
		err = h.db.RemoveEvacuatingActualLRP(logger, actualLRPKey, actualLRPInstanceKey)
		if err != nil {
			logger.Error("failed-removing-evacuating-actual-lrp", err)
			return err
		}
	case models.ActualLRP_Suspect:
		_, err := h.suspectLRPDB.RemoveSuspectActualLRP(logger, actualLRPKey)
		if err != nil {
			logger.Error("failed-removing-suspect-actual-lrp", err)
			return err
		}
	case models.ActualLRP_Ordinary:
		err = h.actualLRPDB.RemoveActualLRP(logger, guid, index, actualLRPInstanceKey)
		if err != nil {
			logger.Error("failed-to-remove-actual-lrp", err)
			return err
		}
	}

	go h.actualHub.Emit(models.NewActualLRPRemovedEvent(targetActualLRP.ToActualLRPGroup()))
	go h.actualLRPInstanceHub.Emit(models.NewActualLRPInstanceRemovedEvent(targetActualLRP))
	return nil
}

func (h *EvacuationController) requestAuction(logger lager.Logger, lrpKey *models.ActualLRPKey) error {
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

func (h *EvacuationController) evacuateRequesting(logger lager.Logger, actualLRPKey *models.ActualLRPKey, actualLRPInstanceKey *models.ActualLRPInstanceKey, netInfo *models.ActualLRPNetInfo) error {
	evacuating, err := h.db.EvacuateActualLRP(logger, actualLRPKey, actualLRPInstanceKey, netInfo)
	if err == models.ErrActualLRPCannotBeEvacuated || err == models.ErrResourceExists {
		return err
	}

	if err != nil {
		logger.Error("failed-evacuating-actual-lrp", err)
	}
	go h.actualHub.Emit(models.NewActualLRPCreatedEvent(evacuating.ToActualLRPGroup()))
	go h.actualLRPInstanceHub.Emit(models.NewActualLRPInstanceCreatedEvent(evacuating))
	return err
}

func (h *EvacuationController) evacuateInstance(logger lager.Logger, allLRPs []*models.ActualLRP, actualLRP *models.ActualLRP) error {
	eventCalculator := EventCalculator{
		ActualLRPGroupHub:    h.actualHub,
		ActualLRPInstanceHub: h.actualLRPInstanceHub,
	}

	evacuating, err := h.db.EvacuateActualLRP(logger, &actualLRP.ActualLRPKey, &actualLRP.ActualLRPInstanceKey, &actualLRP.ActualLRPNetInfo)
	if err != nil {
		return err
	}

	// although EvacuateActualLRP above creates a new database record.  We
	// would like to record that as a change event instead, since the instance
	// guid hasn't changed.  This will produce a simpler instance event stream
	// with a single changed event and keep the group events backward
	// compatible.
	newLRPs := eventCalculator.RecordChange(actualLRP, evacuating, allLRPs)

	defer func() {
		go eventCalculator.EmitEvents(allLRPs, newLRPs)
	}()

	// events = append(events, models.NewActualLRPCreatedEvent(evacuating.ToActualLRPGroup()))
	// instanceEvents = append(instanceEvents, models.NewActualLRPInstanceCreatedEvent(evacuating))

	if actualLRP.Presence == models.ActualLRP_Suspect {
		_, err := h.suspectLRPDB.RemoveSuspectActualLRP(logger, &actualLRP.ActualLRPKey)
		if err != nil {
			logger.Error("failed-removing-suspect-actual-lrp", err)
			return err
		}

		return nil
	}

	_, after, err := h.actualLRPDB.UnclaimActualLRP(logger, &actualLRP.ActualLRPKey)
	if err != nil {
		return err
	}

	// although UnclaimActualLRP above updates a database record.  We would
	// like to record that as a create event instead.  This will produce a
	// simpler instance event stream and keep the group events backward
	// compatible.
	newLRPs = eventCalculator.RecordChange(nil, after, newLRPs)

	return h.requestAuction(logger, &actualLRP.ActualLRPKey)
}

func (h *EvacuationController) removeEvacuating(logger lager.Logger, evacuating *models.ActualLRP) error {
	if evacuating == nil {
		return nil
	}

	err := h.db.RemoveEvacuatingActualLRP(logger, &evacuating.ActualLRPKey, &evacuating.ActualLRPInstanceKey)
	if err == nil {
		go h.actualHub.Emit(models.NewActualLRPRemovedEvent(&models.ActualLRPGroup{Evacuating: evacuating}))
		go h.actualLRPInstanceHub.Emit(models.NewActualLRPInstanceRemovedEvent(evacuating))
	}
	if err != nil && err != models.ErrActualLRPCannotBeRemoved {
		return err
	}
	return nil
}
