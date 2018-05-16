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
	before, after, err := h.db.ClaimActualLRP(logger, processGuid, index, actualLRPInstanceKey)
	if err != nil {
		return err
	}
	if !after.Equal(before) {
		go h.actualHub.Emit(models.NewFlattenedActualLRPChangedEvent(before, after))
	}
	return nil
}

func (h *ActualLRPLifecycleController) StartActualLRP(logger lager.Logger, actualLRPKey *models.ActualLRPKey, actualLRPInstanceKey *models.ActualLRPInstanceKey, actualLRPNetInfo *models.ActualLRPNetInfo) error {
	before, after, err := h.db.StartActualLRP(logger, actualLRPKey, actualLRPInstanceKey, actualLRPNetInfo)
	if err != nil {
		return err
	}

	lrps, err := h.db.ActualLRPs(logger, models.ActualLRPFilter{
		ProcessGUID: &actualLRPKey.ProcessGuid,
		Index:       &actualLRPKey.Index,
	})
	if err != nil {
		return err
	}

	var evacuatingLRP *models.FlattenedActualLRP
	for _, lrp := range lrps {
		if lrp.ActualLRPInfo.PlacementState == models.PlacementStateType_Evacuating {
			evacuatingLRP = lrp
			h.evacuationDB.RemoveEvacuatingActualLRP(logger, &lrp.ActualLRPKey, &lrp.ActualLRPInstanceKey)
			break
		}
	}

	go func() {
		if before == nil {
			h.actualHub.Emit(models.NewFlattenedActualLRPCreatedEvent(after))
		} else if !before.Equal(after) {
			h.actualHub.Emit(models.NewFlattenedActualLRPChangedEvent(before, after))
		}
		if evacuatingLRP != nil {
			h.actualHub.Emit(models.NewFlattenedActualLRPRemovedEvent(evacuatingLRP))
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
	go h.actualHub.Emit(models.NewFlattenedActualLRPChangedEvent(before, after))

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

	go h.actualHub.Emit(models.NewFlattenedActualLRPChangedEvent(before, after))
	return nil
}

func (h *ActualLRPLifecycleController) RemoveActualLRP(logger lager.Logger, processGuid string, index int32, instanceKey *models.ActualLRPInstanceKey) error {
	lrps, err := h.db.ActualLRPs(logger, models.ActualLRPFilter{
		ProcessGUID: &processGuid,
		Index:       &index,
	})
	if err != nil {
		return err
	}

	err = h.db.RemoveActualLRP(logger, processGuid, index, instanceKey)
	if err != nil {
		return err
	}

	for _, lrp := range lrps {
		if lrp.GetPlacementState() == models.PlacementStateType_Normal {
			go h.actualHub.Emit(models.NewFlattenedActualLRPRemovedEvent(lrp))
			break
		}
	}
	return nil
}

func (h *ActualLRPLifecycleController) RetireActualLRP(logger lager.Logger, key *models.ActualLRPKey) error {
	var err error
	var cell *models.CellPresence

	logger = logger.Session("retire-actual-lrp", lager.Data{"process_guid": key.ProcessGuid, "index": key.Index})

	for retryCount := 0; retryCount < models.RetireActualLRPRetryAttempts; retryCount++ {
		var lrps []*models.FlattenedActualLRP

		lrps, err := h.db.ActualLRPs(logger, models.ActualLRPFilter{
			ProcessGUID: &key.ProcessGuid,
			Index:       &key.Index,
		})

		var normalLRP *models.FlattenedActualLRP
		for _, lrp := range lrps {
			if lrp.PlacementState == models.PlacementStateType_Normal {
				normalLRP = lrp
				break
			}
		}

		if normalLRP == nil {
			return models.ErrResourceNotFound
		}

		removeLRP := func(lrp *models.FlattenedActualLRP) error {
			err = h.db.RemoveActualLRP(logger, lrp.ProcessGuid, lrp.Index, &lrp.ActualLRPInstanceKey)
			if err == nil {
				go h.actualHub.Emit(models.NewFlattenedActualLRPRemovedEvent(lrp))
			}
			return err
		}

		switch normalLRP.State {
		case models.ActualLRPStateUnclaimed, models.ActualLRPStateCrashed:
			removeLRP(normalLRP)
		case models.ActualLRPStateClaimed, models.ActualLRPStateRunning:
			cell, err = h.serviceClient.CellById(logger, normalLRP.CellId)
			if err != nil {
				bbsErr := models.ConvertError(err)
				if bbsErr.Type == models.Error_ResourceNotFound {
					err = removeLRP(normalLRP)
				}
				return err
			}

			var client rep.Client
			client, err = h.repClientFactory.CreateClient(cell.RepAddress, cell.RepUrl)
			if err != nil {
				return err
			}
			err = client.StopLRPInstance(logger, normalLRP.ActualLRPKey, normalLRP.ActualLRPInstanceKey)
		}

		if err == nil {
			return nil
		}

		logger.Error("retrying-failed-retire-of-actual-lrp", err, lager.Data{"attempt": retryCount + 1})
	}

	return err
}
