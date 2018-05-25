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

func findLRP(key *models.ActualLRPInstanceKey, lrps []*models.ActualLRP) (*models.ActualLRP, bool) {
	for _, lrp := range lrps {
		if lrp.ActualLRPInstanceKey == *key {
			return lrp, lrp.PlacementState == models.PlacementStateType_Suspect
		}
	}
	return nil, false
}

func findNormalInstance(lrps []*models.ActualLRP) *models.ActualLRP {
	for _, lrp := range lrps {
		if lrp.PlacementState == models.PlacementStateType_Normal {
			return lrp
		}
	}
	return nil
}

func (h *ActualLRPLifecycleController) StartActualLRP(logger lager.Logger, actualLRPKey *models.ActualLRPKey, actualLRPInstanceKey *models.ActualLRPInstanceKey, actualLRPNetInfo *models.ActualLRPNetInfo) error {
	lrps, err := h.db.ActualLRPs(logger, models.ActualLRPFilter{ProcessGUID: &actualLRPKey.ProcessGuid, Index: &actualLRPKey.Index})
	if err != nil {
		if err != models.ErrResourceNotFound {
			logger.Error("err-when-finding-suspect", err)
			return err
		}
	}

	if _, suspect := findLRP(actualLRPInstanceKey, lrps); suspect {
		// this is a suspect starting
		// if there is a normal instance return an error to destroy the suspect lrps
		if lrp := findNormalInstance(lrps); lrp != nil && lrp.State == models.ActualLRPStateRunning {
			return models.ErrActualLRPCannotBeStarted
		}
		return nil
	}

	events := []models.Event{}

	before, after, err := h.db.StartActualLRP(logger, actualLRPKey, actualLRPInstanceKey, actualLRPNetInfo)
	if err != nil {
		return err
	}

	if before == nil {
		events = append(events, models.NewFlattenedActualLRPCreatedEvent(after))
	} else if !before.Equal(after) {
		events = append(events, models.NewFlattenedActualLRPChangedEvent(before, after))
	}

	for _, lrp := range lrps {
		if lrp.ActualLRPInstanceKey == *actualLRPInstanceKey {
			// do not touch the lrp that just started
			continue
		}

		// otherwise remove all evacuating/suspect lrps that have the same guid+index
		h.db.RemoveActualLRP(logger, lrp.ProcessGuid, lrp.Index, &lrp.ActualLRPInstanceKey)
		events = append(events, models.NewFlattenedActualLRPRemovedEvent(lrp))
	}

	go func() {
		for _, ev := range events {
			h.actualHub.Emit(ev)
		}
	}()
	return nil
}

func (h *ActualLRPLifecycleController) CrashActualLRP(logger lager.Logger, actualLRPKey *models.ActualLRPKey, actualLRPInstanceKey *models.ActualLRPInstanceKey, errorMessage string) error {
	lrps, err := h.db.ActualLRPs(logger, models.ActualLRPFilter{ProcessGUID: &actualLRPKey.ProcessGuid, Index: &actualLRPKey.Index, CellID: actualLRPInstanceKey.CellId})
	if err != nil {
		logger.Error("err-when-finding-actual-lrp-group", err)
		return err
	}

	if lrp, suspect := findLRP(actualLRPInstanceKey, lrps); suspect {
		logger = logger.Session("found-crashed-suspect", lager.Data{"guid": actualLRPKey.ProcessGuid, "index": actualLRPKey.Index, "instance-guid": actualLRPInstanceKey.InstanceGuid})
		err = h.db.RemoveActualLRP(logger, lrp.ProcessGuid, lrp.Index, &lrp.ActualLRPInstanceKey)
		h.actualHub.Emit(models.NewFlattenedActualLRPRemovedEvent(lrp))
		return err
	}

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
		var lrps []*models.ActualLRP

		lrps, err := h.db.ActualLRPs(logger, models.ActualLRPFilter{
			ProcessGUID: &key.ProcessGuid,
			Index:       &key.Index,
		})

		var normalLRP *models.ActualLRP
		for _, lrp := range lrps {
			if lrp.PlacementState == models.PlacementStateType_Normal {
				normalLRP = lrp
				break
			}
		}

		if normalLRP == nil {
			return models.ErrResourceNotFound
		}

		removeLRP := func(lrp *models.ActualLRP) error {
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
