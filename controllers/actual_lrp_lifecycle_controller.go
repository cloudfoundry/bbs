package controllers

import (
	"code.cloudfoundry.org/auctioneer"
	"code.cloudfoundry.org/bbs/db"
	"code.cloudfoundry.org/bbs/events"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/serviceclient"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/rep"
)

type ActualLRPLifecycleController struct {
	db                   db.ActualLRPDB
	suspectDB            db.SuspectDB
	evacuationDB         db.EvacuationDB
	desiredLRPDB         db.DesiredLRPDB
	auctioneerClient     auctioneer.Client
	serviceClient        serviceclient.ServiceClient
	repClientFactory     rep.ClientFactory
	actualHub            events.Hub
	actualLRPInstanceHub events.Hub
}

func NewActualLRPLifecycleController(
	db db.ActualLRPDB,
	suspectDB db.SuspectDB,
	evacuationDB db.EvacuationDB,
	desiredLRPDB db.DesiredLRPDB,
	auctioneerClient auctioneer.Client,
	serviceClient serviceclient.ServiceClient,
	repClientFactory rep.ClientFactory,
	actualHub events.Hub,
	actualLRPInstanceHub events.Hub,
) *ActualLRPLifecycleController {
	return &ActualLRPLifecycleController{
		db:                   db,
		suspectDB:            suspectDB,
		evacuationDB:         evacuationDB,
		desiredLRPDB:         desiredLRPDB,
		auctioneerClient:     auctioneerClient,
		serviceClient:        serviceClient,
		repClientFactory:     repClientFactory,
		actualHub:            actualHub,
		actualLRPInstanceHub: actualLRPInstanceHub,
	}
}

func findWithPresence(lrps []*models.ActualLRP, presence models.ActualLRP_Presence) *models.ActualLRP {
	for _, lrp := range lrps {
		if lrp.Presence == presence {
			return lrp
		}
	}
	return nil
}

func lookupLRPInSlice(lrps []*models.ActualLRP, key *models.ActualLRPInstanceKey) *models.ActualLRP {
	for _, lrp := range lrps {
		if lrp.ActualLRPInstanceKey == *key {
			return lrp
		}
	}
	return nil
}

func (h *ActualLRPLifecycleController) ClaimActualLRP(logger lager.Logger, processGuid string, index int32, actualLRPInstanceKey *models.ActualLRPInstanceKey) error {
	before, after, err := h.db.ClaimActualLRP(logger, processGuid, index, actualLRPInstanceKey)
	if err != nil {
		return err
	}

	lrps, err := h.db.ActualLRPs(logger, models.ActualLRPFilter{ProcessGuid: processGuid, Index: &index})
	if err != nil {
		return err
	}

	suspectLRP := findWithPresence(lrps, models.ActualLRP_Suspect)
	if !after.Equal(before) {
		// emit lrp instance event
		go h.actualLRPInstanceHub.Emit(models.NewActualLRPInstanceChangedEvent(before, after))

		// emit lrp group instance event
		if suspectLRP == nil {
			go h.actualHub.Emit(models.NewActualLRPChangedEvent(before.ToActualLRPGroup(), after.ToActualLRPGroup()))
		}
	}
	return nil
}

func (h *ActualLRPLifecycleController) emitV0StartEvent(before, after, suspect *models.ActualLRP) {
	if suspect != nil || before == nil {
		// This behavior (emitting a create event if there was a suspect LRP) is
		// designed to be backward compatible with old clients.  The API will
		// project the Suspect LRP as the ActualLRPGroup's Instance until the new
		// Instance is in the Running state (i.e. is started).  Once the new LRP is
		// in the running state the following two lines will emit
		// ActualLRPRemovedEvent for the Suspect LRP and a ActualLRPCreatedEvent
		// for the Ordinary LRP.  At any point in time calls to ActualLRPs should
		// get a consistent result with the events being received.
		//
		// see https://www.pivotaltracker.com/story/show/158123373 for more
		// information
		h.actualHub.Emit(models.NewActualLRPCreatedEvent(after.ToActualLRPGroup()))
		return
	}

	if !before.Equal(after) {
		h.actualHub.Emit(models.NewActualLRPChangedEvent(before.ToActualLRPGroup(), after.ToActualLRPGroup()))
	}
}

func (h *ActualLRPLifecycleController) emitV1StartEvent(before, after, suspect *models.ActualLRP) {
	if before == nil {
		h.actualLRPInstanceHub.Emit(models.NewActualLRPInstanceCreatedEvent(after))
		return
	}

	if !before.Equal(after) {
		h.actualLRPInstanceHub.Emit(models.NewActualLRPInstanceChangedEvent(before, after))
	}
}

func (h *ActualLRPLifecycleController) StartActualLRP(logger lager.Logger, actualLRPKey *models.ActualLRPKey, actualLRPInstanceKey *models.ActualLRPInstanceKey, actualLRPNetInfo *models.ActualLRPNetInfo) error {
	lrps, err := h.db.ActualLRPs(logger, models.ActualLRPFilter{ProcessGuid: actualLRPKey.ProcessGuid, Index: &actualLRPKey.Index})
	if err != nil && err != models.ErrResourceNotFound {
		return err
	}

	lrp := lookupLRPInSlice(lrps, actualLRPInstanceKey)
	if lrp != nil && lrp.Presence == models.ActualLRP_Suspect {
		logger.Info("ignored-start-request-from-suspect", lager.Data{
			"process_guid":  actualLRPKey.ProcessGuid,
			"index":         actualLRPKey.Index,
			"instance_guid": actualLRPInstanceKey,
			"state":         lrp.State,
		})
		return nil
	}

	before, after, err := h.db.StartActualLRP(logger, actualLRPKey, actualLRPInstanceKey, actualLRPNetInfo)
	if err != nil {
		return err
	}

	evacuating := findWithPresence(lrps, models.ActualLRP_Evacuating)
	suspect := findWithPresence(lrps, models.ActualLRP_Suspect)

	var suspectLRP *models.ActualLRP
	if evacuating != nil {
		h.evacuationDB.RemoveEvacuatingActualLRP(logger, &evacuating.ActualLRPKey, &evacuating.ActualLRPInstanceKey)
	}

	// prior to starting this ActualLRP there was a suspect LRP that we need to remove
	if suspect != nil {
		suspectLRP, err = h.suspectDB.RemoveSuspectActualLRP(logger, actualLRPKey)
		if err != nil {
			logger.Error("failed-to-remove-suspect-lrp", err)
		}
	}

	go func() {
		h.emitV0StartEvent(before, after, suspect)
		h.emitV1StartEvent(before, after, suspect)

		if suspectLRP != nil {
			h.actualHub.Emit(models.NewActualLRPRemovedEvent(suspect.ToActualLRPGroup()))
			h.actualLRPInstanceHub.Emit(models.NewActualLRPInstanceRemovedEvent(suspect))
		}

		if evacuating != nil {
			h.actualHub.Emit(models.NewActualLRPRemovedEvent(evacuating.ToActualLRPGroup()))
			h.actualLRPInstanceHub.Emit(models.NewActualLRPInstanceRemovedEvent(evacuating))
		}
	}()

	return nil
}

func (h *ActualLRPLifecycleController) CrashActualLRP(logger lager.Logger, actualLRPKey *models.ActualLRPKey, actualLRPInstanceKey *models.ActualLRPInstanceKey, errorMessage string) error {
	lrps, err := h.db.ActualLRPs(logger, models.ActualLRPFilter{ProcessGuid: actualLRPKey.ProcessGuid, Index: &actualLRPKey.Index})
	if err != nil {
		return err
	}

	eventCalculator := EventCalculator{
		ActualLRPGroupHub:    h.actualHub,
		ActualLRPInstanceHub: h.actualLRPInstanceHub,
	}

	lrp := lookupLRPInSlice(lrps, actualLRPInstanceKey)
	if lrp != nil && lrp.Presence == models.ActualLRP_Suspect {
		suspectLRP, err := h.suspectDB.RemoveSuspectActualLRP(logger, actualLRPKey)
		if err != nil {
			return err
		}

		afterLRPs := eventCalculator.RecordChange(suspectLRP, nil, lrps)
		logger.Info("removing-suspect-lrp", lager.Data{"ig": suspectLRP.InstanceGuid})
		go eventCalculator.EmitEvents(lrps, afterLRPs)

		return nil
	}

	before, after, shouldRestart, err := h.db.CrashActualLRP(logger, actualLRPKey, actualLRPInstanceKey, errorMessage)
	if err != nil {
		return err
	}

	afterLRPs := eventCalculator.RecordChange(before, after, lrps)
	eventCalculator.EmitEvents(lrps, afterLRPs)

	if !shouldRestart {
		return nil
	}

	desiredLRP, err := h.desiredLRPDB.DesiredLRPByProcessGuid(logger, actualLRPKey.ProcessGuid)
	if err != nil {
		logger.Error("failed-fetching-desired-lrp", err)
		return err
	}

	schedInfo := desiredLRP.DesiredLRPSchedulingInfo()
	startRequest := auctioneer.NewLRPStartRequestFromSchedulingInfo(&schedInfo, int(actualLRPKey.Index))
	logger.Info("start-lrp-auction-request", lager.Data{"app_guid": schedInfo.ProcessGuid, "index": int(actualLRPKey.Index)})
	err = h.auctioneerClient.RequestLRPAuctions(logger, []*auctioneer.LRPStartRequest{&startRequest})
	logger.Info("finished-lrp-auction-request", lager.Data{"app_guid": schedInfo.ProcessGuid, "index": int(actualLRPKey.Index)})
	if err != nil {
		logger.Error("failed-requesting-auction", err)
	}
	return nil
}

func (h *ActualLRPLifecycleController) FailActualLRP(logger lager.Logger, key *models.ActualLRPKey, errorMessage string) error {
	before, after, err := h.db.FailActualLRP(logger, key, errorMessage)
	if err != nil && err != models.ErrResourceNotFound {
		return err
	}

	lrps, err := h.db.ActualLRPs(logger, models.ActualLRPFilter{ProcessGuid: key.ProcessGuid, Index: &key.Index})
	if err != nil {
		return err
	}

	go h.actualLRPInstanceHub.Emit(models.NewActualLRPInstanceChangedEvent(before, after))
	go func() {
		suspectExists := findWithPresence(lrps, models.ActualLRP_Suspect)
		if suspectExists == nil {
			h.actualHub.Emit(models.NewActualLRPChangedEvent(before.ToActualLRPGroup(), after.ToActualLRPGroup()))
		}
	}()

	return nil
}

func (h *ActualLRPLifecycleController) RemoveActualLRP(logger lager.Logger, processGuid string, index int32, instanceKey *models.ActualLRPInstanceKey) error {
	beforeLRPs, err := h.db.ActualLRPs(logger, models.ActualLRPFilter{ProcessGuid: processGuid, Index: &index})
	if err != nil {
		return err
	}

	err = h.db.RemoveActualLRP(logger, processGuid, index, instanceKey)
	if err != nil {
		return err
	}
	go h.actualHub.Emit(models.NewActualLRPRemovedEvent(beforeLRPs[0].ToActualLRPGroup()))
	go h.actualLRPInstanceHub.Emit(models.NewActualLRPInstanceRemovedEvent(beforeLRPs[0]))
	return nil
}

func (h *ActualLRPLifecycleController) RetireActualLRP(logger lager.Logger, key *models.ActualLRPKey) error {
	var err error
	var cell *models.CellPresence

	logger = logger.Session("retire-actual-lrp", lager.Data{"process_guid": key.ProcessGuid, "index": key.Index})

	for retryCount := 0; retryCount < models.RetireActualLRPRetryAttempts; retryCount++ {
		var lrps []*models.ActualLRP
		lrps, err = h.db.ActualLRPs(logger, models.ActualLRPFilter{ProcessGuid: key.ProcessGuid, Index: &key.Index})
		if err != nil {
			return err
		}

		lrp := findWithPresence(lrps, models.ActualLRP_Ordinary)
		if lrp == nil {
			return models.ErrResourceNotFound
		}

		switch lrp.State {
		case models.ActualLRPStateUnclaimed, models.ActualLRPStateCrashed:
			err = h.db.RemoveActualLRP(logger, lrp.ProcessGuid, lrp.Index, &lrp.ActualLRPInstanceKey)
			if err == nil {
				go h.actualHub.Emit(models.NewActualLRPRemovedEvent(lrp.ToActualLRPGroup()))
				go h.actualLRPInstanceHub.Emit(models.NewActualLRPInstanceRemovedEvent(lrp))
			}
		case models.ActualLRPStateClaimed, models.ActualLRPStateRunning:
			cell, err = h.serviceClient.CellById(logger, lrp.CellId)
			if err != nil {
				bbsErr := models.ConvertError(err)
				if bbsErr.Type == models.Error_ResourceNotFound {
					err = h.db.RemoveActualLRP(logger, lrp.ProcessGuid, lrp.Index, &lrp.ActualLRPInstanceKey)
					if err == nil {
						go h.actualHub.Emit(models.NewActualLRPRemovedEvent(lrp.ToActualLRPGroup()))
						go h.actualLRPInstanceHub.Emit(models.NewActualLRPInstanceRemovedEvent(lrp))
					}
				}
				return err
			}

			var client rep.Client
			client, err = h.repClientFactory.CreateClient(cell.RepAddress, cell.RepUrl)
			if err != nil {
				return err
			}
			err = client.StopLRPInstance(logger, lrp.ActualLRPKey, lrp.ActualLRPInstanceKey)
		}

		if err == nil {
			return nil
		}

		logger.Error("retrying-failed-retire-of-actual-lrp", err, lager.Data{"attempt": retryCount + 1})
	}

	return err
}
