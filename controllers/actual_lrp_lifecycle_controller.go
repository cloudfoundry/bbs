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
	evacuationDB     db.EvacuationDB
	desiredLRPDB     db.DesiredLRPDB
	auctioneerClient auctioneer.Client
	serviceClient    serviceclient.ServiceClient
	repClientFactory rep.ClientFactory
	actualHub        events.Hub
}

func NewActualLRPLifecycleController(
	db db.ActualLRPDB,
	evacuationDB db.EvacuationDB,
	desiredLRPDB db.DesiredLRPDB,
	auctioneerClient auctioneer.Client,
	serviceClient serviceclient.ServiceClient,
	repClientFactory rep.ClientFactory,
	actualHub events.Hub,
) *ActualLRPLifecycleController {
	return &ActualLRPLifecycleController{
		db:               db,
		evacuationDB:     evacuationDB,
		desiredLRPDB:     desiredLRPDB,
		auctioneerClient: auctioneerClient,
		serviceClient:    serviceClient,
		repClientFactory: repClientFactory,
		actualHub:        actualHub,
	}
}

func (h *ActualLRPLifecycleController) ClaimActualLRP(logger lager.Logger, processGuid string, index int32, actualLRPInstanceKey *models.ActualLRPInstanceKey) error {
	_, after, err := h.db.ClaimActualLRP(logger, processGuid, index, actualLRPInstanceKey)
	if err != nil {
		return err
	}

	go h.actualHub.Emit(models.NewFlattenedActualLRPCreatedEvent(after))
	return nil
}
func (h *ActualLRPLifecycleController) StartActualLRP(logger lager.Logger, actualLRPKey *models.ActualLRPKey, actualLRPInstanceKey *models.ActualLRPInstanceKey, actualLRPNetInfo *models.ActualLRPNetInfo) error {
	before, after, err := h.db.StartActualLRP(logger, actualLRPKey, actualLRPInstanceKey, actualLRPNetInfo)
	if err != nil {
		return err
	}

	lrp, err := h.db.ActualLRPByProcessGuidAndIndex(logger, actualLRPKey.ProcessGuid, actualLRPKey.Index)
	if err != nil {
		return err
	}

	if lrp.ActualLRPInfo.PlacementState == models.PlacementStateType_Evacuating {
		h.evacuationDB.RemoveEvacuatingActualLRP(logger, &lrp.ActualLRPKey, &lrp.ActualLRPInstanceKey)
	}

	go func() {
		if before == nil {
			h.actualHub.Emit(models.NewFlattenedActualLRPCreatedEvent(after))
		} else if !before.Equal(after) {
			h.actualHub.Emit(models.NewFlattenedActualLRPCreatedEvent(after))
			h.actualHub.Emit(models.NewFlattenedActualLRPRemovedEvent(before))
		}
		if lrp.ActualLRPInfo.PlacementState == models.PlacementStateType_Evacuating {
			h.actualHub.Emit(models.NewFlattenedActualLRPRemovedEvent(lrp))
		}
	}()
	return nil
}

func (h *ActualLRPLifecycleController) CrashActualLRP(logger lager.Logger, actualLRPKey *models.ActualLRPKey, actualLRPInstanceKey *models.ActualLRPInstanceKey, errorMessage string) error {
	before, after, shouldRestart, err := h.db.CrashActualLRP(logger, actualLRPKey, actualLRPInstanceKey, errorMessage)
	if err != nil {
		return err
	}

	// TODO need to figure out the events
	go h.actualHub.Emit(models.NewFlattenedActualLRPCrashedEvent(before, after))
	event := models.NewFlattenedActualLRPChangedEvent(before, after)
	if event == nil {
		h.actualHub.Emit(models.NewFlattenedActualLRPCreatedEvent(after))
		h.actualHub.Emit(models.NewFlattenedActualLRPRemovedEvent(before))
	} else {
		go h.actualHub.Emit(event)
	}

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
	if err != nil {
		return err
	}

	event := models.NewFlattenedActualLRPChangedEvent(before, after)
	if event == nil {
		h.actualHub.Emit(models.NewFlattenedActualLRPCreatedEvent(after))
		h.actualHub.Emit(models.NewFlattenedActualLRPRemovedEvent(before))
	} else {
		go h.actualHub.Emit(event)
	}
	return nil
}

func (h *ActualLRPLifecycleController) RemoveActualLRP(logger lager.Logger, processGuid string, index int32, instanceKey *models.ActualLRPInstanceKey) error {
	beforeActualLRP, err := h.db.ActualLRPByProcessGuidAndIndex(logger, processGuid, index)
	if err != nil {
		return err
	}

	err = h.db.RemoveActualLRP(logger, processGuid, index, instanceKey)
	if err != nil {
		return err

	}
	go h.actualHub.Emit(models.NewFlattenedActualLRPRemovedEvent(beforeActualLRP))
	return nil
}

func (h *ActualLRPLifecycleController) RetireActualLRP(logger lager.Logger, key *models.ActualLRPKey) error {
	var err error
	var cell *models.CellPresence

	logger = logger.Session("retire-actual-lrp", lager.Data{"process_guid": key.ProcessGuid, "index": key.Index})

	for retryCount := 0; retryCount < models.RetireActualLRPRetryAttempts; retryCount++ {
		var lrp *models.FlattenedActualLRP
		lrp, err = h.db.ActualLRPByProcessGuidAndIndex(logger, key.ProcessGuid, key.Index)
		if err != nil {
			return err
		}

		if lrp == nil {
			return models.ErrResourceNotFound
		}

		switch lrp.State {
		case models.ActualLRPStateUnclaimed, models.ActualLRPStateCrashed:
			err = h.db.RemoveActualLRP(logger, lrp.ProcessGuid, lrp.Index, &lrp.ActualLRPInstanceKey)
			if err == nil {
				go h.actualHub.Emit(models.NewFlattenedActualLRPRemovedEvent(lrp))
			}
		case models.ActualLRPStateClaimed, models.ActualLRPStateRunning:
			cell, err = h.serviceClient.CellById(logger, lrp.CellId)
			if err != nil {
				bbsErr := models.ConvertError(err)
				if bbsErr.Type == models.Error_ResourceNotFound {
					err = h.db.RemoveActualLRP(logger, lrp.ProcessGuid, lrp.Index, &lrp.ActualLRPInstanceKey)
					if err == nil {
						go h.actualHub.Emit(models.NewFlattenedActualLRPRemovedEvent(lrp))
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
