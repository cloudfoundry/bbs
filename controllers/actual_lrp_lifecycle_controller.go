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
	db               db.ActualLRPDB
	suspectDB        db.SuspectDB
	evacuationDB     db.EvacuationDB
	desiredLRPDB     db.DesiredLRPDB
	auctioneerClient auctioneer.Client
	serviceClient    serviceclient.ServiceClient
	repClientFactory rep.ClientFactory
	actualHub        events.Hub
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
) *ActualLRPLifecycleController {
	return &ActualLRPLifecycleController{
		db:               db,
		suspectDB:        suspectDB,
		evacuationDB:     evacuationDB,
		desiredLRPDB:     desiredLRPDB,
		auctioneerClient: auctioneerClient,
		serviceClient:    serviceClient,
		repClientFactory: repClientFactory,
		actualHub:        actualHub,
	}
}

func (h *ActualLRPLifecycleController) ClaimActualLRP(logger lager.Logger, processGuid string, index int32, actualLRPInstanceKey *models.ActualLRPInstanceKey) error {
	before, after, err := h.db.ClaimActualLRP(logger, processGuid, index, actualLRPInstanceKey)
	if err != nil {
		return err
	}

	lrpGroup, err := h.db.ActualLRPGroupByProcessGuidAndIndex(logger, processGuid, index)
	if err != nil {
		return err
	}

	// Only emit ActualLRPChangedEvent if there was no Suspect instance.
	// Otherwise, we shouldn't emit any events until the replacement instance is
	// up.  This combined with the API internal resolve logic (i.e. to return the
	// Suspect LRP in the Instance field while the replacement is starting) will
	// give consistent view to the clients.
	if !after.Equal(before) && lrpGroup.Instance.ActualLRPInstanceKey == *actualLRPInstanceKey {
		go h.actualHub.Emit(models.NewActualLRPChangedEvent(before, after))
	}
	return nil
}

func (h *ActualLRPLifecycleController) StartActualLRP(logger lager.Logger, actualLRPKey *models.ActualLRPKey, actualLRPInstanceKey *models.ActualLRPInstanceKey, actualLRPNetInfo *models.ActualLRPNetInfo) error {
	lrpGroup, err := h.db.ActualLRPGroupByProcessGuidAndIndex(logger, actualLRPKey.ProcessGuid, actualLRPKey.Index)
	if err != nil && err != models.ErrResourceNotFound {
		return err
	}

	if lrpGroup != nil && lrpGroup.Instance.Presence == models.ActualLRP_Suspect && lrpGroup.Instance.ActualLRPInstanceKey == *actualLRPInstanceKey {
		// nothing to do
		return nil
	}

	before, after, err := h.db.StartActualLRP(logger, actualLRPKey, actualLRPInstanceKey, actualLRPNetInfo)
	if err != nil {
		return err
	}

	if lrpGroup != nil && lrpGroup.Evacuating != nil {
		h.evacuationDB.RemoveEvacuatingActualLRP(logger, &lrpGroup.Evacuating.ActualLRPKey, &lrpGroup.Evacuating.ActualLRPInstanceKey)
	}

	// prior to starting this ActualLRP there was a suspect LRP that we need to remove
	var suspectLRPGroup *models.ActualLRPGroup
	if lrpGroup != nil && lrpGroup.Instance.Presence == models.ActualLRP_Suspect {
		suspectLRPGroup, err = h.suspectDB.RemoveSuspectActualLRP(logger, actualLRPKey)
		if err != nil {
			logger.Error("failed-to-remove-suspect-lrp", err)
		}
	}

	go func() {
		if suspectLRPGroup == nil {
			// there is no suspect LRP proceed like normal
			if before == nil {
				h.actualHub.Emit(models.NewActualLRPCreatedEvent(after))
			} else if !before.Equal(after) {
				h.actualHub.Emit(models.NewActualLRPChangedEvent(before, after))
			}
		} else {
			// Otherwise, emit an ActualLRPCreatedEvent.  This behavior is designed
			// to be backward compatible with old clients.  The API will project the
			// Suspect LRP as the ActualLRPGroup's Instance until the new Instance is
			// in the Running state (i.e. is started).  Once the new LRP is in the
			// running state the following two lines will emit ActualLRPRemovedEvent
			// for the Suspect LRP and a ActualLRPCreatedEvent for the Ordinary LRP.
			// At any point in time calls to ActualLRPGroups or
			// ActualLRPGroupByProcessGuidAndIndex should get a consistent result
			// with the events being received.
			//
			// see https://www.pivotaltracker.com/story/show/158123373 for more
			// information
			h.actualHub.Emit(models.NewActualLRPCreatedEvent(after))
			h.actualHub.Emit(models.NewActualLRPRemovedEvent(suspectLRPGroup))
		}
		if lrpGroup != nil && lrpGroup.Evacuating != nil {
			h.actualHub.Emit(models.NewActualLRPRemovedEvent(&models.ActualLRPGroup{Evacuating: lrpGroup.Evacuating}))
		}
	}()
	return nil
}

func (h *ActualLRPLifecycleController) CrashActualLRP(logger lager.Logger, actualLRPKey *models.ActualLRPKey, actualLRPInstanceKey *models.ActualLRPInstanceKey, errorMessage string) error {
	lrpGroup, err := h.db.ActualLRPGroupByProcessGuidAndIndex(logger, actualLRPKey.ProcessGuid, actualLRPKey.Index)
	if err != nil {
		return err
	}

	if lrpGroup.Instance != nil &&
		lrpGroup.Instance.Presence == models.ActualLRP_Suspect &&
		lrpGroup.Instance.ActualLRPInstanceKey == *actualLRPInstanceKey {
		suspectLRPGroup, err := h.suspectDB.RemoveSuspectActualLRP(logger, actualLRPKey)
		if err == nil {
			go h.actualHub.Emit(models.NewActualLRPRemovedEvent(suspectLRPGroup))
		}
		return err
	}

	before, after, shouldRestart, err := h.db.CrashActualLRP(logger, actualLRPKey, actualLRPInstanceKey, errorMessage)
	if err != nil {
		return err
	}

	beforeActualLRP, _, beforeResolveError := before.Resolve()
	if beforeResolveError != nil {
		return beforeResolveError
	}
	afterActualLRP, _, afterResolveError := after.Resolve()
	if afterResolveError != nil {
		return afterResolveError
	}
	go h.actualHub.Emit(models.NewActualLRPCrashedEvent(beforeActualLRP, afterActualLRP))
	go h.actualHub.Emit(models.NewActualLRPChangedEvent(before, after))

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
	lrpGroup, err := h.db.ActualLRPGroupByProcessGuidAndIndex(logger, key.ProcessGuid, key.Index)
	if err != nil {
		return err
	}

	if lrpGroup.Instance != nil && lrpGroup.Instance.Presence == models.ActualLRP_Suspect {
		// nothing to do
		return nil
	}

	before, after, err := h.db.FailActualLRP(logger, key, errorMessage)
	if err != nil {
		return err
	}

	if lrpGroup.Instance == nil {
		go h.actualHub.Emit(models.NewActualLRPChangedEvent(before, after))
	}

	return nil
}

func (h *ActualLRPLifecycleController) RemoveActualLRP(logger lager.Logger, processGuid string, index int32, instanceKey *models.ActualLRPInstanceKey) error {
	beforeActualLRPGroup, err := h.db.ActualLRPGroupByProcessGuidAndIndex(logger, processGuid, index)
	if err != nil {
		return err
	}

	err = h.db.RemoveActualLRP(logger, processGuid, index, instanceKey)
	if err != nil {
		return err

	}
	go h.actualHub.Emit(models.NewActualLRPRemovedEvent(beforeActualLRPGroup))
	return nil
}

func (h *ActualLRPLifecycleController) RetireActualLRP(logger lager.Logger, key *models.ActualLRPKey) error {
	var err error
	var cell *models.CellPresence

	logger = logger.Session("retire-actual-lrp", lager.Data{"process_guid": key.ProcessGuid, "index": key.Index})

	for retryCount := 0; retryCount < models.RetireActualLRPRetryAttempts; retryCount++ {
		var lrpGroup *models.ActualLRPGroup
		lrpGroup, err = h.db.ActualLRPGroupByProcessGuidAndIndex(logger, key.ProcessGuid, key.Index)
		if err != nil {
			return err
		}

		lrp := lrpGroup.Instance
		if lrp == nil {
			return models.ErrResourceNotFound
		}

		switch lrp.State {
		case models.ActualLRPStateUnclaimed, models.ActualLRPStateCrashed:
			err = h.db.RemoveActualLRP(logger, lrp.ProcessGuid, lrp.Index, &lrp.ActualLRPInstanceKey)
			if err == nil {
				go h.actualHub.Emit(models.NewActualLRPRemovedEvent(lrpGroup))
			}
		case models.ActualLRPStateClaimed, models.ActualLRPStateRunning:
			cell, err = h.serviceClient.CellById(logger, lrp.CellId)
			if err != nil {
				bbsErr := models.ConvertError(err)
				if bbsErr.Type == models.Error_ResourceNotFound {
					err = h.db.RemoveActualLRP(logger, lrp.ProcessGuid, lrp.Index, &lrp.ActualLRPInstanceKey)
					if err == nil {
						go h.actualHub.Emit(models.NewActualLRPRemovedEvent(lrpGroup))
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
