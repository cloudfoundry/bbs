package controllers

import (
	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/bbs/db"
	"code.cloudfoundry.org/bbs/events"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/rep"
)

type ActualLRPRetirer interface {
	RetireActualLRP(logger lager.Logger, processGuid string, index int32) error
}

type actualLRPRetirer struct {
	db               db.ActualLRPDB
	actualHub        events.Hub
	repClientFactory rep.ClientFactory
	serviceClient    bbs.ServiceClient
}

func NewActualLRPRetirer(db db.ActualLRPDB,
	actualHub events.Hub,
	repClientFactory rep.ClientFactory,
	serviceClient bbs.ServiceClient,
) *actualLRPRetirer {
	return &actualLRPRetirer{
		db:               db,
		actualHub:        actualHub,
		repClientFactory: repClientFactory,
		serviceClient:    serviceClient,
	}
}

func (r *actualLRPRetirer) RetireActualLRP(logger lager.Logger, processGuid string, index int32) error {
	var err error
	var cell *models.CellPresence

	logger = logger.Session("retire-actual-lrp", lager.Data{"process_guid": processGuid, "index": index})

	for retryCount := 0; retryCount < models.RetireActualLRPRetryAttempts; retryCount++ {
		var lrpGroup *models.ActualLRPGroup
		lrpGroup, err = r.db.ActualLRPGroupByProcessGuidAndIndex(logger, processGuid, index)
		if err != nil {
			return err
		}

		lrp := lrpGroup.Instance
		if lrp == nil {
			return models.ErrResourceNotFound
		}

		switch lrp.State {
		case models.ActualLRPStateUnclaimed, models.ActualLRPStateCrashed:
			err = r.db.RemoveActualLRP(logger, lrp.ProcessGuid, lrp.Index, &lrp.ActualLRPInstanceKey)
			if err == nil {
				go r.actualHub.Emit(models.NewActualLRPRemovedEvent(lrpGroup))
			}
		case models.ActualLRPStateClaimed, models.ActualLRPStateRunning:
			cell, err = r.serviceClient.CellById(logger, lrp.CellId)
			if err != nil {
				bbsErr := models.ConvertError(err)
				if bbsErr.Type == models.Error_ResourceNotFound {
					err = r.db.RemoveActualLRP(logger, lrp.ProcessGuid, lrp.Index, &lrp.ActualLRPInstanceKey)
					if err == nil {
						go r.actualHub.Emit(models.NewActualLRPRemovedEvent(lrpGroup))
					}
				}
				return err
			}

			client := r.repClientFactory.CreateClient(cell.RepAddress)
			err = client.StopLRPInstance(lrp.ActualLRPKey, lrp.ActualLRPInstanceKey)
		}

		if err == nil {
			return nil
		}

		logger.Error("retrying-failed-retire-of-actual-lrp", err, lager.Data{"attempt": retryCount + 1})
	}

	return err
}
