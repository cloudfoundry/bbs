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
	lrpDeploymentDB  db.LRPDeploymentDB
	auctioneerClient auctioneer.Client
	serviceClient    serviceclient.ServiceClient
	repClientFactory rep.ClientFactory
	actualHub        events.Hub
}

func NewActualLRPLifecycleController(
	db db.ActualLRPDB,
	evacuationDB db.EvacuationDB,
	desiredLRPDB db.DesiredLRPDB,
	lrpDeploymentDB db.LRPDeploymentDB,
	auctioneerClient auctioneer.Client,
	serviceClient serviceclient.ServiceClient,
	repClientFactory rep.ClientFactory,
	actualHub events.Hub,
) *ActualLRPLifecycleController {
	return &ActualLRPLifecycleController{
		db:               db,
		evacuationDB:     evacuationDB,
		desiredLRPDB:     desiredLRPDB,
		lrpDeploymentDB:  lrpDeploymentDB,
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
		go h.actualHub.Emit(models.NewActualLRPChangedEvent(before, after))
	}
	return nil
}
func (h *ActualLRPLifecycleController) StartActualLRP(logger lager.Logger, actualLRPKey *models.ActualLRPKey, actualLRPInstanceKey *models.ActualLRPInstanceKey, actualLRPNetInfo *models.ActualLRPNetInfo) error {
	before, after, err := h.db.StartActualLRP(logger, actualLRPKey, actualLRPInstanceKey, actualLRPNetInfo)
	if err != nil {
		return err
	}

	lrpGroup, err := h.db.ActualLRPGroupByProcessGuidAndIndex(logger, actualLRPKey.ProcessGuid, actualLRPKey.Index)
	if err != nil {
		return err
	}

	if lrpGroup.Evacuating != nil {
		h.evacuationDB.RemoveEvacuatingActualLRP(logger, &lrpGroup.Evacuating.ActualLRPKey, &lrpGroup.Evacuating.ActualLRPInstanceKey)
	}

	go func() {
		if before == nil {
			h.actualHub.Emit(models.NewActualLRPCreatedEvent(after))
		} else if !before.Equal(after) {
			h.actualHub.Emit(models.NewActualLRPChangedEvent(before, after))
		}
		if lrpGroup.Evacuating != nil {
			h.actualHub.Emit(models.NewActualLRPRemovedEvent(&models.ActualLRPGroup{Evacuating: lrpGroup.Evacuating}))
		}
	}()

	//TODO: Remove the actual lrp of the previous definition
	//Set the healthy definition id of lrp deployment if its the first instance
	// 1. Find if the healthy_definition_id for this lrpdeployment is different from the one that is running
	// 2. Update  the healthy_definition_id with the  definition id of the update
	// 3. Remove actual lrps for the old definition id in the healthy_definition_id
	// 4. Maybe we need to do this only for index==0 because that indicates the first instance coming up?

	lrpDeployment, err := h.lrpDeploymentDB.LRPDeploymentByDefinitionGuid(logger, actualLRPKey.ProcessGuid)
	if err != nil {
		logger.Error("failed-retrieving-lrpdeployment", err)
		return err
	}

	logger.Info("lrp-deployment-found", lager.Data{"deployment": lrpDeployment})

	if lrpDeployment.ActiveDefinitionId == actualLRPKey.ProcessGuid {
		for defID, _ := range lrpDeployment.Definitions {
			if defID != actualLRPKey.ProcessGuid {

				beforeActualLRPGroup, err := h.db.ActualLRPGroupByProcessGuidAndIndex(logger, defID, actualLRPKey.Index)
				if err == models.ErrResourceNotFound {
					continue
				}
				if err != nil {
					return err
				}

				oldActualLRP, _ := beforeActualLRPGroup.Resolve()
				h.RetireActualLRP(logger, &oldActualLRP.ActualLRPKey)
			}
		}
		if lrpDeployment.HealthyDefinitionId != actualLRPKey.ProcessGuid {
			lrpDeployment.HealthyDefinitionId = actualLRPKey.ProcessGuid
			err := h.lrpDeploymentDB.SaveLRPDeployment(logger, lrpDeployment)
			if err != nil {
				logger.Error("failed-to-save-lrp-deployment", err)
				return err
			}
		}
	}

	return nil
}

func (h *ActualLRPLifecycleController) CrashActualLRP(logger lager.Logger, actualLRPKey *models.ActualLRPKey, actualLRPInstanceKey *models.ActualLRPInstanceKey, errorMessage string) error {
	before, after, shouldRestart, err := h.db.CrashActualLRP(logger, actualLRPKey, actualLRPInstanceKey, errorMessage)
	if err != nil {
		return err
	}

	if shouldRestart {
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
			return err
		}
	}

	beforeActualLRP, _ := before.Resolve()
	afterActualLRP, _ := after.Resolve()
	go h.actualHub.Emit(models.NewActualLRPCrashedEvent(beforeActualLRP, afterActualLRP))
	go h.actualHub.Emit(models.NewActualLRPChangedEvent(before, after))
	return nil
}

func (h *ActualLRPLifecycleController) FailActualLRP(logger lager.Logger, key *models.ActualLRPKey, errorMessage string) error {
	before, after, err := h.db.FailActualLRP(logger, key, errorMessage)
	if err != nil {
		return err
	}

	go h.actualHub.Emit(models.NewActualLRPChangedEvent(before, after))
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
