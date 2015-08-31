package etcd

import (
	"reflect"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

func (db *ETCDDB) EvacuateClaimedActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey) (keepContainer bool, modelErr error) {
	logger = logger.Session("evacuate-claimed", lager.Data{"actual_lrp_key": key, "actual_lrp_instance_key": instanceKey})
	logger.Info("started")
	defer func() { logger.Info("finished", lager.Data{"keepContainer": keepContainer, "err": modelErr}) }()

	_ = db.RemoveEvacuatingActualLRP(logger, key, instanceKey)

	changed, err := db.unclaimActualLRP(logger, key, instanceKey)
	bbsErr := models.ConvertError(err)
	if bbsErr != nil {
		if bbsErr.Type == models.Error_ResourceNotFound {
			return false, nil
		}
		return false, err
	}

	if !changed {
		return false, nil
	}

	logger.Info("requesting-start-lrp-auction")

	err = db.requestLRPAuctionForLRPKey(logger, key)
	if err != nil {
		logger.Error("failed-requesting-start-lrp-auction", err)
		return false, err
	}

	logger.Info("succeeded-requesting-start-lrp-auction")

	logger.Info("succeeded")
	return false, nil
}

func (db *ETCDDB) EvacuateRunningActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, netInfo *models.ActualLRPNetInfo, ttl uint64) (keepContainer bool, modelErr error) {
	logger = logger.Session("evacuate-running", lager.Data{
		"actual_lrp_key":          key,
		"actual_lrp_instance_key": instanceKey,
		"actual_lrp_net_info":     netInfo,
		"ttl": ttl,
	})
	logger.Info("started")
	defer func() { logger.Info("finished", lager.Data{"keepContainer": keepContainer, "err": modelErr}) }()

	instanceLRP, storeIndex, err := db.rawActuaLLRPByProcessGuidAndIndex(logger, key.ProcessGuid, key.Index)
	bbsErr := models.ConvertError(err)
	if bbsErr != nil {
		if bbsErr.Type == models.Error_ResourceNotFound {
			err := db.RemoveEvacuatingActualLRP(logger, key, instanceKey)
			bbsErr := models.ConvertError(err)
			if bbsErr != nil {
				if bbsErr.Type == models.Error_ActualLRPCannotBeRemoved {
					logger.Debug("remove-evacuating-actual-lrp-failed")
					return false, nil
				}
				logger.Debug("remove-evacuating-actual-lrp-errored")
				return true, err
			}
			logger.Debug("remove-evacuating-actual-lrp-success")
			return false, nil
		}
		logger.Debug("fetch-actual-lrp-errored")
		return true, err
	}

	// if the instance is unclaimed or claimed by another cell,
	// mark this cell as evacuating the lrp as long as it isn't already marked by another cell.
	if (instanceLRP.State == models.ActualLRPStateUnclaimed && instanceLRP.PlacementError == "") ||
		(instanceLRP.State == models.ActualLRPStateClaimed && !instanceLRP.ActualLRPInstanceKey.Equal(instanceKey)) {
		logger.Debug("conditionally-evacuate-actual-lrp")
		err = db.conditionallyEvacuateActualLRP(logger, key, instanceKey, netInfo, ttl)
		if err == models.ErrResourceExists || err == models.ErrActualLRPCannotBeEvacuated {
			logger.Debug("conditionally-cannot-evacuate")
			return false, nil
		}
		if err != nil {
			logger.Debug("conditionally-unknown-evacuation-error")
			return true, err
		}
		logger.Info("conditionally-succeeded")
		return true, nil
	}

	// if the instance is claimed by or running on this cell, unconditionally mark this cell as evacuating the lrp.
	if (instanceLRP.State == models.ActualLRPStateClaimed || instanceLRP.State == models.ActualLRPStateRunning) &&
		instanceLRP.ActualLRPInstanceKey.Equal(instanceKey) {
		logger.Debug("unconditionally-evacuate-actual-lrp")
		err := db.unconditionallyEvacuateActualLRP(logger, key, instanceKey, netInfo, ttl)
		if err != nil {
			logger.Debug("unconditionally-unknown-evacuation-error")
			return true, err
		}

		changed, err := db.unclaimActualLRPWithIndex(logger, instanceLRP, storeIndex, key, instanceKey)
		if err != nil {
			logger.Debug("error-unclaiming-actual-lrp", lager.Data{"err": err})
			return true, err
		}

		if !changed {
			logger.Info("unconditionally-succeeded")
			return true, nil
		}

		logger.Info("requesting-start-lrp-auction")
		err = db.requestLRPAuctionForLRPKey(logger, key)
		if err != nil {
			logger.Error("failed-requesting-start-lrp-auction", err)
		} else {
			logger.Info("succeeded-requesting-start-lrp-auction")
		}

		return true, err
	}

	if (instanceLRP.State == models.ActualLRPStateUnclaimed && instanceLRP.PlacementError != "") ||
		(instanceLRP.State == models.ActualLRPStateRunning && !instanceLRP.ActualLRPInstanceKey.Equal(instanceKey)) ||
		instanceLRP.State == models.ActualLRPStateCrashed {
		err := db.RemoveEvacuatingActualLRP(logger, key, instanceKey)
		bbsErr := models.ConvertError(err)
		if bbsErr != nil {
			if bbsErr.Type == models.Error_ActualLRPCannotBeRemoved {
				return false, nil
			}
			return true, err
		}

		return false, nil
	}

	logger.Info("succeeded")
	return true, nil
}

func (db *ETCDDB) EvacuateStoppedActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey) (bool, error) {
	logger = logger.Session("evacuating-stopped", lager.Data{"actual_lrp_key": key, "actual_lrp_instance_key": instanceKey})
	logger.Info("started")

	// ignore the error if we can't remove the LRP
	_ = db.RemoveEvacuatingActualLRP(logger, key, instanceKey)

	lrp, storeIndex, err := db.rawActuaLLRPByProcessGuidAndIndex(logger, key.ProcessGuid, key.Index)
	bbsErr := models.ConvertError(err)
	if bbsErr != nil {
		if bbsErr.Type == models.Error_ResourceNotFound {
			return false, nil
		}
	} else if !lrp.ActualLRPInstanceKey.Equal(instanceKey) {
		return false, models.ErrActualLRPCannotBeRemoved
	} else if err != nil {
		return false, err
	}

	err = db.removeActualLRP(logger, lrp, storeIndex)
	if err != nil {
		return false, err
	}

	logger.Info("succeeded")
	return false, nil
}

func (db *ETCDDB) EvacuateCrashedActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, errorMessage string) (bool, error) {
	logger = logger.Session("evacuating-crashed", lager.Data{"actual_lrp_key": key, "actual_lrp_instance_key": instanceKey, "error_message": errorMessage})

	logger.Info("started")

	err := db.RemoveEvacuatingActualLRP(logger, key, instanceKey)
	if err != nil {
		logger.Debug("failed-to-remove-evacuating-actual-lrp", lager.Data{"error": err})
	}

	err = db.CrashActualLRP(logger, key, instanceKey, errorMessage)
	bbsErr := models.ConvertError(err)
	if bbsErr != nil {
		if bbsErr.Type == models.Error_ResourceNotFound {
			return false, nil
		}
		return false, err
	}

	logger.Info("succeeded")
	return false, nil
}

func (db *ETCDDB) RemoveEvacuatingActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey) error {
	var err error
	var prevIndex uint64
	var lrp *models.ActualLRP
	processGuid := key.ProcessGuid
	index := key.Index
	logger = logger.Session("removing-evacuating", lager.Data{"process_guid": processGuid, "index": index})

	lrp, prevIndex, err = db.rawEvacuatingActuaLLRPByProcessGuidAndIndex(logger, processGuid, index)
	bbsErr := models.ConvertError(err)
	if bbsErr != nil {
		if bbsErr.Type == models.Error_ResourceNotFound {
			logger.Debug("evacuating-actual-lrp-already-removed")
			return nil
		}
		return err
	}

	if !lrp.ActualLRPKey.Equal(key) ||
		!lrp.ActualLRPInstanceKey.Equal(instanceKey) {
		return models.ErrActualLRPCannotBeRemoved
	}

	logger.Info("starting")
	_, etcdErr := db.client.CompareAndDelete(EvacuatingActualLRPSchemaPath(lrp.ProcessGuid, lrp.Index), prevIndex)
	if etcdErr != nil {
		logger.Error("failed", etcdErr)
		return models.ErrActualLRPCannotBeRemoved
	}

	logger.Info("succeeded")
	return nil
}

func (db *ETCDDB) rawEvacuatingActuaLLRPByProcessGuidAndIndex(logger lager.Logger, processGuid string, index int32) (*models.ActualLRP, uint64, error) {
	node, bbsErr := db.fetchRaw(logger, EvacuatingActualLRPSchemaPath(processGuid, index))
	if bbsErr != nil {
		return nil, 0, bbsErr
	}

	lrp := new(models.ActualLRP)
	deserializeErr := db.deserializeModel(logger, node, lrp)
	if deserializeErr != nil {
		return nil, 0, models.ErrDeserializeJSON
	}

	return lrp, node.ModifiedIndex, nil
}

type stateChange bool

const (
	stateDidChange    stateChange = true
	stateDidNotChange stateChange = false
)

func (db *ETCDDB) unconditionallyEvacuateActualLRP(
	logger lager.Logger,
	actualLRPKey *models.ActualLRPKey,
	actualLRPInstanceKey *models.ActualLRPInstanceKey,
	actualLRPNetInfo *models.ActualLRPNetInfo,
	evacuationTTLInSeconds uint64,
) error {
	existingLRP, storeIndex, err := db.rawEvacuatingActuaLLRPByProcessGuidAndIndex(logger, actualLRPKey.ProcessGuid, actualLRPKey.Index)
	logger = logger.Session("unconditionally-evacuate")
	bbsErr := models.ConvertError(err)
	if bbsErr != nil {
		if bbsErr.Type == models.Error_ResourceNotFound {
			return db.createEvacuatingActualLRP(logger, actualLRPKey, actualLRPInstanceKey, actualLRPNetInfo, evacuationTTLInSeconds)
		}
		return err
	}

	if existingLRP.ActualLRPKey.Equal(actualLRPKey) &&
		existingLRP.ActualLRPInstanceKey.Equal(actualLRPInstanceKey) &&
		existingLRP.Address == actualLRPNetInfo.Address &&
		reflect.DeepEqual(existingLRP.Ports, actualLRPNetInfo.Ports) {
		return nil
	}

	lrp := *existingLRP

	lrp.Since = db.clock.Now().UnixNano()
	lrp.ActualLRPInstanceKey = *actualLRPInstanceKey
	lrp.ActualLRPNetInfo = *actualLRPNetInfo
	lrp.PlacementError = ""
	lrp.ModificationTag.Increment()

	return db.compareAndSwapRawEvacuatingActualLRP(logger, &lrp, storeIndex, evacuationTTLInSeconds)
}

func (db *ETCDDB) conditionallyEvacuateActualLRP(
	logger lager.Logger,
	actualLRPKey *models.ActualLRPKey,
	actualLRPInstanceKey *models.ActualLRPInstanceKey,
	actualLRPNetInfo *models.ActualLRPNetInfo,
	evacuationTTLInSeconds uint64,
) error {
	existingLRP, storeIndex, err := db.rawEvacuatingActuaLLRPByProcessGuidAndIndex(logger, actualLRPKey.ProcessGuid, actualLRPKey.Index)
	logger = logger.Session("conditionally-evacuate")

	bbsErr := models.ConvertError(err)
	if bbsErr != nil {
		if bbsErr.Type == models.Error_ResourceNotFound {
			return db.createEvacuatingActualLRP(logger, actualLRPKey, actualLRPInstanceKey, actualLRPNetInfo, evacuationTTLInSeconds)
		}
		return err
	}

	if existingLRP.ActualLRPKey.Equal(actualLRPKey) &&
		existingLRP.ActualLRPInstanceKey.Equal(actualLRPInstanceKey) &&
		existingLRP.Address == actualLRPNetInfo.Address &&
		reflect.DeepEqual(existingLRP.Ports, actualLRPNetInfo.Ports) {
		return nil
	}

	if !existingLRP.ActualLRPKey.Equal(actualLRPKey) ||
		!existingLRP.ActualLRPInstanceKey.Equal(actualLRPInstanceKey) {
		return models.ErrActualLRPCannotBeEvacuated
	}

	lrp := *existingLRP

	lrp.Since = db.clock.Now().UnixNano()
	lrp.ActualLRPInstanceKey = *actualLRPInstanceKey
	lrp.ActualLRPNetInfo = *actualLRPNetInfo
	lrp.PlacementError = ""
	lrp.ModificationTag.Increment()

	return db.compareAndSwapRawEvacuatingActualLRP(logger, &lrp, storeIndex, evacuationTTLInSeconds)
}

func (db *ETCDDB) compareAndSwapRawEvacuatingActualLRP(
	logger lager.Logger,
	lrp *models.ActualLRP,
	storeIndex uint64,
	evacuationTTLInSeconds uint64,
) error {
	payload, err := db.serializeModel(logger, lrp)
	if err != nil {
		logger.Error("failed-to-marshal-actual-lrp", err, lager.Data{"actual-lrp": lrp})
		return err
	}

	_, err = db.client.CompareAndSwap(
		EvacuatingActualLRPSchemaPath(lrp.ActualLRPKey.ProcessGuid, lrp.ActualLRPKey.Index),
		payload,
		evacuationTTLInSeconds,
		storeIndex,
	)
	if err != nil {
		logger.Error("failed-to-compare-and-swap-evacuating-actual-lrp", err, lager.Data{"actual-lrp": lrp})
		return models.ErrResourceConflict
	}

	return nil
}
