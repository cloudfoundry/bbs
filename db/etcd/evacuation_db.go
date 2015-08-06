package etcd

import (
	"encoding/json"
	"reflect"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

func (db *ETCDDB) EvacuateClaimedActualLRP(logger lager.Logger, request *models.EvacuateClaimedActualLRPRequest) (keepContainer bool, modelErr *models.Error) {
	logger = logger.Session("evacuate-claimed", lager.Data{"request": request})
	logger.Info("started")
	defer func() { logger.Info("finished", lager.Data{"keepContainer": keepContainer, "err": modelErr}) }()

	_ = db.removeEvacuatingActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey)

	changed, err := db.unclaimActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey)
	if err == models.ErrResourceNotFound {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	if !changed {
		return false, nil
	}

	logger.Info("requesting-start-lrp-auction")

	err = db.requestLRPAuctionForLRPKey(logger, request.ActualLrpKey)
	if err != nil {
		logger.Error("failed-requesting-start-lrp-auction", err)
		return false, err
	}

	logger.Info("succeeded-requesting-start-lrp-auction")

	logger.Info("succeeded")
	return false, nil
}

func (db *ETCDDB) EvacuateRunningActualLRP(logger lager.Logger, request *models.EvacuateRunningActualLRPRequest) (keepContainer bool, modelErr *models.Error) {
	logger = logger.Session("evacuate-running", lager.Data{"request": request})
	logger.Info("started")
	defer func() { logger.Info("finished", lager.Data{"keepContainer": keepContainer, "err": modelErr}) }()

	instanceLRP, storeIndex, err := db.rawActuaLLRPByProcessGuidAndIndex(logger, request.ActualLrpKey.ProcessGuid, request.ActualLrpKey.Index)
	if err == models.ErrResourceNotFound {
		err := db.removeEvacuatingActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey)
		if err == models.ErrActualLRPCannotBeRemoved {
			logger.Debug("remove-evacuating-actual-lrp-failed")
			return false, nil
		} else if err != nil {
			logger.Debug("remove-evacuating-actual-lrp-errored")
			return true, err
		}
		logger.Debug("remove-evacuating-actual-lrp-success")
		return false, nil
	} else if err != nil {
		logger.Debug("fetch-actual-lrp-errored")
		return true, err
	}

	// if the instance is unclaimed or claimed by another cell,
	// mark this cell as evacuating the lrp as long as it isn't already marked by another cell.
	if (instanceLRP.State == models.ActualLRPStateUnclaimed && instanceLRP.PlacementError == "") ||
		(instanceLRP.State == models.ActualLRPStateClaimed && !instanceLRP.ActualLRPInstanceKey.Equal(request.ActualLrpInstanceKey)) {
		logger.Debug("conditionally-evacuate-actual-lrp")
		err = db.conditionallyEvacuateActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey, request.ActualLrpNetInfo, request.Ttl)
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
		instanceLRP.ActualLRPInstanceKey.Equal(request.ActualLrpInstanceKey) {
		logger.Debug("unconditionally-evacuate-actual-lrp")
		err := db.unconditionallyEvacuateActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey, request.ActualLrpNetInfo, request.Ttl)
		if err != nil {
			logger.Debug("unconditionally-unknown-evacuation-error")
			return true, err
		}

		changed, err := db.unclaimActualLRPWithIndex(logger, instanceLRP, storeIndex, request.ActualLrpKey, request.ActualLrpInstanceKey)
		if err != nil {
			logger.Debug("error-unclaiming-actual-lrp", lager.Data{"err": err})
			return true, err
		}

		if !changed {
			logger.Info("unconditionally-succeeded")
			return true, nil
		}

		logger.Info("requesting-start-lrp-auction")
		err = db.requestLRPAuctionForLRPKey(logger, request.ActualLrpKey)
		if err != nil {
			logger.Error("failed-requesting-start-lrp-auction", err)
		} else {
			logger.Info("succeeded-requesting-start-lrp-auction")
		}

		return true, err
	}

	if (instanceLRP.State == models.ActualLRPStateUnclaimed && instanceLRP.PlacementError != "") ||
		(instanceLRP.State == models.ActualLRPStateRunning && !instanceLRP.ActualLRPInstanceKey.Equal(request.ActualLrpInstanceKey)) ||
		instanceLRP.State == models.ActualLRPStateCrashed {
		err := db.removeEvacuatingActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey)
		if err == models.ErrActualLRPCannotBeRemoved {
			return false, nil
		}
		if err != nil {
			return true, err
		}

		return false, nil
	}

	logger.Info("succeeded")
	return true, nil
}

func (db *ETCDDB) EvacuateStoppedActualLRP(logger lager.Logger, request *models.EvacuateStoppedActualLRPRequest) (bool, *models.Error) {
	logger = logger.Session("evacuating-stopped", lager.Data{"request": request})
	logger.Info("started")

	_ = db.removeEvacuatingActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey)

	lrp, storeIndex, err := db.rawActuaLLRPByProcessGuidAndIndex(logger, request.ActualLrpKey.ProcessGuid, request.ActualLrpKey.Index)
	if err == models.ErrResourceNotFound {
		return false, nil
	} else if !lrp.ActualLRPInstanceKey.Equal(request.ActualLrpInstanceKey) {
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

func (db *ETCDDB) EvacuateCrashedActualLRP(logger lager.Logger, request *models.EvacuateCrashedActualLRPRequest) (bool, *models.Error) {
	logger = logger.Session("evacuating-crashed", lager.Data{"request": request})
	logger.Info("started")

	err := db.removeEvacuatingActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey)
	if err != nil {
		logger.Debug("failed-to-remove-evacuating-actual-lrp", lager.Data{"error": err})
	}

	err = db.CrashActualLRP(logger, &models.CrashActualLRPRequest{
		ActualLrpKey:         request.ActualLrpKey,
		ActualLrpInstanceKey: request.ActualLrpInstanceKey,
		ErrorMessage:         request.ErrorMessage,
	})
	if err == models.ErrResourceNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}

	logger.Info("succeeded")
	return false, nil
}

func (db *ETCDDB) RemoveEvacuatingActualLRP(logger lager.Logger, request *models.RemoveEvacuatingActualLRPRequest) *models.Error {
	return db.removeEvacuatingActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey)
}

func (db *ETCDDB) removeEvacuatingActualLRP(logger lager.Logger, lrpKey *models.ActualLRPKey, lrpInstanceKey *models.ActualLRPInstanceKey) *models.Error {
	var err *models.Error
	var prevIndex uint64
	var lrp *models.ActualLRP
	processGuid := lrpKey.ProcessGuid
	index := lrpKey.Index
	logger = logger.Session("removing-evacuating", lager.Data{"process_guid": processGuid, "index": index})

	lrp, prevIndex, err = db.rawEvacuatingActuaLLRPByProcessGuidAndIndex(logger, processGuid, index)
	if err == models.ErrResourceNotFound {
		logger.Debug("evacuating-actual-lrp-already-removed")
		return nil
	}

	if err != nil {
		return err
	}

	if !lrp.ActualLRPKey.Equal(lrpKey) ||
		!lrp.ActualLRPInstanceKey.Equal(lrpInstanceKey) {
		return models.ErrActualLRPCannotBeRemoved
	}

	logger.Info("starting")
	_, etcdErr := db.client.CompareAndDelete(EvacuatingActualLRPSchemaPath(lrp.ProcessGuid, lrp.Index), "", prevIndex)
	if etcdErr != nil {
		logger.Error("failed", etcdErr)
		return models.ErrActualLRPCannotBeRemoved
	}

	logger.Info("succeeded")
	return nil
}

// func (db *ETCDDB) ActualLRPGroupByProcessGuidAndIndex(logger lager.Logger, processGuid string, index int32) (*models.ActualLRPGroup, *models.Error) {
// 	group, _, err := db.rawActualLRPGroupByProcessGuidAndIndex(logger, processGuid, index)
// 	return group, err
// }

func (db *ETCDDB) rawEvacuatingActuaLLRPByProcessGuidAndIndex(logger lager.Logger, processGuid string, index int32) (*models.ActualLRP, uint64, *models.Error) {
	node, bbsErr := db.fetchRaw(logger, EvacuatingActualLRPSchemaPath(processGuid, index))
	if bbsErr != nil {
		return nil, 0, bbsErr
	}

	var lrp models.ActualLRP
	deserializeErr := json.Unmarshal([]byte(node.Value), &lrp)
	if deserializeErr != nil {
		return nil, 0, models.ErrDeserializeJSON
	}

	return &lrp, node.ModifiedIndex, nil
}

type stateChange bool

const (
	stateDidChange    stateChange = true
	stateDidNotChange stateChange = false
)

func (db *ETCDDB) unclaimActualLRP(
	logger lager.Logger,
	actualLRPKey *models.ActualLRPKey,
	actualLRPInstanceKey *models.ActualLRPInstanceKey,
) (stateChange, *models.Error) {
	logger = logger.Session("unclaim-actual-lrp")
	logger.Info("starting")

	lrp, storeIndex, err := db.rawActuaLLRPByProcessGuidAndIndex(logger, actualLRPKey.ProcessGuid, actualLRPKey.Index)
	if err != nil {
		return stateDidNotChange, err
	}

	changed, err := db.unclaimActualLRPWithIndex(logger, lrp, storeIndex, actualLRPKey, actualLRPInstanceKey)
	if err != nil {
		return changed, err
	}

	logger.Info("succeeded")
	return changed, nil
}

func (db *ETCDDB) unclaimActualLRPWithIndex(
	logger lager.Logger,
	lrp *models.ActualLRP,
	storeIndex uint64,
	actualLRPKey *models.ActualLRPKey,
	actualLRPInstanceKey *models.ActualLRPInstanceKey,
) (change stateChange, modelErr *models.Error) {
	logger = logger.Session("unclaim-actual-lrp-with-index")
	defer func() {
		logger.Debug("complete", lager.Data{"stateChange": change, "error": modelErr})
	}()
	if !lrp.ActualLRPKey.Equal(actualLRPKey) {
		logger.Error("failed-actual-lrp-key-differs", models.ErrActualLRPCannotBeUnclaimed)
		return stateDidNotChange, models.ErrActualLRPCannotBeUnclaimed
	}

	if lrp.State == models.ActualLRPStateUnclaimed {
		logger.Info("already-unclaimed")
		return stateDidNotChange, nil
	}

	if !lrp.ActualLRPInstanceKey.Equal(actualLRPInstanceKey) {
		logger.Error("failed-actual-lrp-instance-key-differs", models.ErrActualLRPCannotBeUnclaimed)
		return stateDidNotChange, models.ErrActualLRPCannotBeUnclaimed
	}

	lrp.Since = db.clock.Now().UnixNano()
	lrp.State = models.ActualLRPStateUnclaimed
	lrp.ActualLRPInstanceKey = models.ActualLRPInstanceKey{}
	lrp.ActualLRPNetInfo = models.EmptyActualLRPNetInfo()
	lrp.ModificationTag.Increment()

	err := lrp.Validate()
	if err != nil {
		logger.Error("failed-to-validate-unclaimed-lrp", err)
		return stateDidNotChange, &models.Error{Type: models.InvalidRecord, Message: err.Error()}
	}

	lrpRawJSON, err := json.Marshal(lrp)
	if err != nil {
		logger.Error("failed-to-marshal-unclaimed-lrp", err)
		return stateDidNotChange, models.ErrSerializeJSON
	}

	_, err = db.client.CompareAndSwap(ActualLRPSchemaPath(actualLRPKey.ProcessGuid, actualLRPKey.Index), string(lrpRawJSON), 0, "", storeIndex)
	if err != nil {
		logger.Error("failed-to-compare-and-swap", err)
		return stateDidNotChange, models.ErrActualLRPCannotBeUnclaimed
	}

	logger.Debug("changed-to-unclaimed")
	return stateDidChange, nil
}

func (db *ETCDDB) unconditionallyEvacuateActualLRP(
	logger lager.Logger,
	actualLRPKey *models.ActualLRPKey,
	actualLRPInstanceKey *models.ActualLRPInstanceKey,
	actualLRPNetInfo *models.ActualLRPNetInfo,
	evacuationTTLInSeconds uint64,
) *models.Error {
	existingLRP, storeIndex, err := db.rawEvacuatingActuaLLRPByProcessGuidAndIndex(logger, actualLRPKey.ProcessGuid, actualLRPKey.Index)
	logger = logger.Session("unconditionally-evacuate")
	if err == models.ErrResourceNotFound {
		return db.createEvacuatingActualLRP(logger, actualLRPKey, actualLRPInstanceKey, actualLRPNetInfo, evacuationTTLInSeconds)
	} else if err != nil {
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
) *models.Error {
	existingLRP, storeIndex, err := db.rawEvacuatingActuaLLRPByProcessGuidAndIndex(logger, actualLRPKey.ProcessGuid, actualLRPKey.Index)
	logger = logger.Session("conditionally-evacuate")
	if err == models.ErrResourceNotFound {
		return db.createEvacuatingActualLRP(logger, actualLRPKey, actualLRPInstanceKey, actualLRPNetInfo, evacuationTTLInSeconds)
	} else if err != nil {
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
) *models.Error {
	lrpRawJSON, err := json.Marshal(lrp)
	if err != nil {
		logger.Error("failed-to-marshal-actual-lrp", err, lager.Data{"actual-lrp": lrp})
		return models.ErrSerializeJSON
	}

	_, err = db.client.CompareAndSwap(
		EvacuatingActualLRPSchemaPath(lrp.ActualLRPKey.ProcessGuid, lrp.ActualLRPKey.Index),
		string(lrpRawJSON),
		evacuationTTLInSeconds,
		"",
		storeIndex,
	)
	if err != nil {
		logger.Error("failed-to-compare-and-swap-evacuating-actual-lrp", err, lager.Data{"actual-lrp": lrp})
		return models.ErrResourceConflict
	}

	return nil
}
