package etcd

import (
	"encoding/json"
	"fmt"
	"path"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry/gunk/workpool"
	"github.com/coreos/go-etcd/etcd"
	"github.com/nu7hatch/gouuid"
	"github.com/pivotal-golang/lager"
)

const retireActualThrottlerSize = 20

func (db *ETCDDB) ActualLRPGroups(logger lager.Logger, filter models.ActualLRPFilter) ([]*models.ActualLRPGroup, error) {
	node, err := db.fetchRecursiveRaw(logger, ActualLRPSchemaRoot)
	bbsErr := models.ConvertError(err)
	if bbsErr != nil {
		if bbsErr.Type == models.Error_ResourceNotFound {
			return []*models.ActualLRPGroup{}, nil
		}
		return nil, err
	}
	if node.Nodes.Len() == 0 {
		return []*models.ActualLRPGroup{}, nil
	}

	groups := []*models.ActualLRPGroup{}

	groupsLock := sync.Mutex{}
	var workErr atomic.Value
	works := []func(){}

	for _, node := range node.Nodes {
		node := node

		works = append(works, func() {
			g, err := parseActualLRPGroups(logger, node, filter)
			if err != nil {
				workErr.Store(err)
				return
			}
			groupsLock.Lock()
			groups = append(groups, g...)
			groupsLock.Unlock()
		})
	}

	throttler, err := workpool.NewThrottler(maxActualGroupGetterWorkPoolSize, works)
	if err != nil {
		logger.Error("failed-constructing-throttler", err, lager.Data{"max-workers": maxActualGroupGetterWorkPoolSize, "num-works": len(works)})
		return []*models.ActualLRPGroup{}, models.ErrUnknownError
	}

	logger.Debug("performing-deserialization-work")
	throttler.Work()
	if err, ok := workErr.Load().(error); ok {
		logger.Error("failed-performing-deserialization-work", err)
		return []*models.ActualLRPGroup{}, models.ErrUnknownError
	}
	logger.Debug("succeeded-performing-deserialization-work", lager.Data{"num-actual-lrp-groups": len(groups)})

	return groups, nil
}

func (db *ETCDDB) ActualLRPGroupsByProcessGuid(logger lager.Logger, processGuid string) ([]*models.ActualLRPGroup, error) {
	node, err := db.fetchRecursiveRaw(logger, ActualLRPProcessDir(processGuid))
	bbsErr := models.ConvertError(err)
	if bbsErr != nil {
		if bbsErr.Type == models.Error_ResourceNotFound {
			return []*models.ActualLRPGroup{}, nil
		}
		return nil, err
	}
	if node.Nodes.Len() == 0 {
		return []*models.ActualLRPGroup{}, nil
	}

	return parseActualLRPGroups(logger, node, models.ActualLRPFilter{})
}

func (db *ETCDDB) instanceActualLRPsByProcessGuid(logger lager.Logger, processGuid string) (map[int32]*models.ActualLRP, error) {
	node, err := db.fetchRecursiveRaw(logger, ActualLRPProcessDir(processGuid))
	bbsErr := models.ConvertError(err)
	if bbsErr != nil {
		if bbsErr.Type == models.Error_ResourceNotFound {
			return nil, nil
		}
		return nil, err
	}
	if node.Nodes.Len() == 0 {
		return nil, nil
	}

	var instances = map[int32]*models.ActualLRP{}

	logger.Debug("performing-parsing-actual-lrps")
	for _, indexNode := range node.Nodes {
		for _, instanceNode := range indexNode.Nodes {
			if !isInstanceActualLRPNode(instanceNode) {
				continue
			}

			instance := &models.ActualLRP{}
			deserializeErr := models.FromJSON([]byte(instanceNode.Value), instance)
			if deserializeErr != nil {
				logger.Error("failed-parsing-actual-lrs", deserializeErr, lager.Data{"key": instanceNode.Key})
				return nil, models.ErrDeserializeJSON
			}

			instances[instance.Index] = instance
		}

	}
	logger.Debug("succeeded-performing-parsing-actual-lrps", lager.Data{"num-actual-lrps": len(instances)})

	return instances, nil
}

func (db *ETCDDB) ActualLRPGroupByProcessGuidAndIndex(logger lager.Logger, processGuid string, index int32) (*models.ActualLRPGroup, error) {
	group, _, err := db.rawActualLRPGroupByProcessGuidAndIndex(logger, processGuid, index)
	return group, err
}

func (db *ETCDDB) rawActualLRPGroupByProcessGuidAndIndex(logger lager.Logger, processGuid string, index int32) (*models.ActualLRPGroup, uint64, error) {
	node, err := db.fetchRecursiveRaw(logger, ActualLRPIndexDir(processGuid, index))
	if err != nil {
		return nil, 0, err
	}

	group := models.ActualLRPGroup{}
	for _, instanceNode := range node.Nodes {
		var lrp models.ActualLRP
		deserializeErr := models.FromJSON([]byte(instanceNode.Value), &lrp)
		if deserializeErr != nil {
			logger.Error("failed-parsing-actual-lrp", deserializeErr, lager.Data{"key": instanceNode.Key})
			return nil, 0, models.ErrDeserializeJSON
		}

		if isInstanceActualLRPNode(instanceNode) {
			group.Instance = &lrp
		}

		if isEvacuatingActualLRPNode(instanceNode) {
			group.Evacuating = &lrp
		}
	}

	if group.Evacuating == nil && group.Instance == nil {
		return nil, 0, models.ErrResourceNotFound
	}

	return &group, node.ModifiedIndex, nil
}

func (db *ETCDDB) rawActuaLLRPByProcessGuidAndIndex(logger lager.Logger, processGuid string, index int32) (*models.ActualLRP, uint64, error) {
	logger.Debug("raw-actual-lrp-by-process-guid-and-index")
	node, err := db.fetchRaw(logger, ActualLRPSchemaPath(processGuid, index))
	if err != nil {
		return nil, 0, err
	}

	var lrp models.ActualLRP
	deserializeErr := json.Unmarshal([]byte(node.Value), &lrp)
	if deserializeErr != nil {
		return nil, 0, models.ErrDeserializeJSON
	}

	return &lrp, node.ModifiedIndex, nil
}

func (db *ETCDDB) ClaimActualLRP(logger lager.Logger, processGuid string, index int32, instanceKey *models.ActualLRPInstanceKey) error {
	logger = logger.Session("claim-actual-lrp", lager.Data{"process_guid": processGuid, "index": index, "actual_lrp_instance-key": instanceKey})
	logger.Info("starting")

	lrp, prevIndex, err := db.rawActuaLLRPByProcessGuidAndIndex(logger, processGuid, index)
	if err != nil {
		logger.Error("failed", err)
		return err
	}

	if !lrp.AllowsTransitionTo(&lrp.ActualLRPKey, instanceKey, models.ActualLRPStateClaimed) {
		return models.ErrActualLRPCannotBeClaimed
	}

	lrp.PlacementError = ""
	lrp.State = models.ActualLRPStateClaimed
	lrp.ActualLRPInstanceKey = *instanceKey
	lrp.ActualLRPNetInfo = models.ActualLRPNetInfo{}
	lrp.ModificationTag.Increment()

	err = lrp.Validate()
	if err != nil {
		logger.Error("failed", err)
		return models.NewError(models.Error_InvalidRecord, err.Error())
	}

	lrpRawJSON, err := json.Marshal(lrp)
	if err != nil {
		logger.Error("failed", err)
		return models.ErrSerializeJSON
	}

	_, err = db.client.CompareAndSwap(ActualLRPSchemaPath(processGuid, index), lrpRawJSON, 0, prevIndex)
	if err != nil {
		logger.Error("failed", err)
		return models.ErrActualLRPCannotBeClaimed
	}
	logger.Info("succeeded")

	return nil
}

func (db *ETCDDB) createActualLRP(logger lager.Logger, desiredLRP *models.DesiredLRP, index int32) error {
	logger = logger.Session("create-actual-lrp")
	var err error
	if index >= desiredLRP.Instances {
		err = models.NewError(models.Error_InvalidRecord, "Index too large")
		logger.Error("actual-lrp-index-too-large", err, lager.Data{"actual-index": index, "desired-instances": desiredLRP.Instances})
		return err
	}

	guid, err := uuid.NewV4()
	if err != nil {
		return err
	}

	actualLRP := &models.ActualLRP{
		ActualLRPKey: models.NewActualLRPKey(
			desiredLRP.ProcessGuid,
			index,
			desiredLRP.Domain,
		),
		State: models.ActualLRPStateUnclaimed,
		Since: db.clock.Now().UnixNano(),
		ModificationTag: models.ModificationTag{
			Epoch: guid.String(),
			Index: 0,
		},
	}

	err = db.createRawActualLRP(logger, actualLRP)
	if err != nil {
		return err
	}

	return nil
}

func (db *ETCDDB) newRunningActualLRP(key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, netInfo *models.ActualLRPNetInfo) (*models.ActualLRP, error) {
	guid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	return &models.ActualLRP{
		ActualLRPKey:         *key,
		ActualLRPInstanceKey: *instanceKey,
		ActualLRPNetInfo:     *netInfo,
		Since:                db.clock.Now().UnixNano(),
		State:                models.ActualLRPStateRunning,
		ModificationTag: models.ModificationTag{
			Epoch: guid.String(),
			Index: 0,
		},
	}, nil
}

func (db *ETCDDB) newUnclaimedActualLRP(key *models.ActualLRPKey) (*models.ActualLRP, error) {
	guid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	return &models.ActualLRP{
		ActualLRPKey: *key,
		Since:        db.clock.Now().UnixNano(),
		State:        models.ActualLRPStateUnclaimed,
		ModificationTag: models.ModificationTag{
			Epoch: guid.String(),
			Index: 0,
		},
	}, nil
}

func (db *ETCDDB) createRawActualLRP(logger lager.Logger, lrp *models.ActualLRP) error {
	lrpRawJSON, err := json.Marshal(lrp)
	if err != nil {
		logger.Error("failed-to-marshal-actual-lrp", err, lager.Data{"actual-lrp": lrp})
		return models.ErrSerializeJSON
	}

	_, err = db.client.Create(ActualLRPSchemaPath(lrp.ProcessGuid, lrp.Index), lrpRawJSON, 0)
	if err != nil {
		logger.Error("failed-to-create-actual-lrp", err, lager.Data{"actual-lrp": lrp})
		return models.ErrActualLRPCannotBeStarted
	}

	return nil
}

func (db *ETCDDB) createRunningActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, netInfo *models.ActualLRPNetInfo) error {
	lrp, err := db.newRunningActualLRP(key, instanceKey, netInfo)
	if err != nil {
		return models.ErrActualLRPCannotBeStarted
	}

	return db.createRawActualLRP(logger, lrp)
}

func (db *ETCDDB) createUnclaimedActualLRP(logger lager.Logger, key *models.ActualLRPKey) error {
	lrp, err := db.newUnclaimedActualLRP(key)
	if err != nil {
		return models.ErrActualLRPCannotBeUnclaimed
	}

	return db.createRawActualLRP(logger, lrp)
}

func (db *ETCDDB) createUnclaimedActualLRPs(logger lager.Logger, keys []*models.ActualLRPKey) []uint {
	count := len(keys)
	createdIndicesChan := make(chan uint, count)

	works := make([]func(), count)

	for i, key := range keys {
		key := key
		works[i] = func() {
			err := db.createUnclaimedActualLRP(logger, key)
			if err != nil {
				logger.Info("failed-creating-actual-lrp", lager.Data{"actual_lrp_key": key, "err-message": err.Error()})
			} else {
				createdIndicesChan <- uint(key.Index)
			}
		}
	}

	throttler, err := workpool.NewThrottler(createActualMaxWorkers, works)
	if err != nil {
		logger.Error("failed-constructing-throttler", err, lager.Data{"max-workers": createActualMaxWorkers, "num-works": len(works)})
		return []uint{}
	}

	throttler.Work()
	close(createdIndicesChan)

	createdIndices := make([]uint, 0, count)
	for createdIndex := range createdIndicesChan {
		createdIndices = append(createdIndices, createdIndex)
	}

	return createdIndices
}

func (db *ETCDDB) createEvacuatingActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, netInfo *models.ActualLRPNetInfo, evacuatingTTLInSeconds uint64) (err error) {
	logger.Debug("create-evacuating-actual-lrp")
	defer func() { logger.Debug("create-evacuating-actual-lrp-complete", lager.Data{"error": err}) }()
	lrp, err := db.newRunningActualLRP(key, instanceKey, netInfo)
	if err != nil {
		return models.ErrActualLRPCannotBeStarted
	}

	lrp.ModificationTag.Increment()

	lrpRawJSON, err := json.Marshal(lrp)
	if err != nil {
		return models.ErrSerializeJSON
	}

	_, err = db.client.Create(EvacuatingActualLRPSchemaPath(key.ProcessGuid, key.Index), lrpRawJSON, evacuatingTTLInSeconds)
	if err != nil {
		logger.Error("failed", err)
		return models.ErrActualLRPCannotBeStarted
	}

	return nil
}

func (db *ETCDDB) StartActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, netInfo *models.ActualLRPNetInfo) error {
	logger = logger.Session("start-actual-lrp", lager.Data{"actual_lrp_key": key, "actual_lrp_instance_key": instanceKey})
	logger.Info("starting")
	lrp, prevIndex, err := db.rawActuaLLRPByProcessGuidAndIndex(logger, key.ProcessGuid, key.Index)
	bbsErr := models.ConvertError(err)
	if bbsErr != nil {
		if bbsErr.Type == models.Error_ResourceNotFound {
			return db.createRunningActualLRP(logger, key, instanceKey, netInfo)
		}
		logger.Error("failed-to-get-actual-lrp", err)
		return err
	}

	if lrp.ActualLRPKey.Equal(key) &&
		lrp.ActualLRPInstanceKey.Equal(instanceKey) &&
		lrp.ActualLRPNetInfo.Equal(netInfo) &&
		lrp.State == models.ActualLRPStateRunning {
		logger.Info("succeeded")
		return nil
	}

	if !lrp.AllowsTransitionTo(key, instanceKey, models.ActualLRPStateRunning) {
		logger.Error("failed-to-transition-actual-lrp-to-started", nil)
		return models.ErrActualLRPCannotBeStarted
	}

	lrp.ModificationTag.Increment()
	lrp.State = models.ActualLRPStateRunning
	lrp.Since = db.clock.Now().UnixNano()
	lrp.ActualLRPInstanceKey = *instanceKey
	lrp.ActualLRPNetInfo = *netInfo
	lrp.PlacementError = ""

	lrpRawJSON, err := json.Marshal(lrp)
	if err != nil {
		return models.ErrSerializeJSON
	}

	_, err = db.client.CompareAndSwap(ActualLRPSchemaPath(key.ProcessGuid, key.Index), lrpRawJSON, 0, prevIndex)
	if err != nil {
		logger.Error("failed", err)
		return models.ErrActualLRPCannotBeStarted
	}

	logger.Info("succeeded")
	return nil
}

func (db *ETCDDB) CrashActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, errorMessage string) error {
	logger = logger.Session("crash-actual-lrp", lager.Data{"actual_lrp_key": key, "actual_lrp_instance_key": instanceKey})
	logger.Info("starting")

	lrp, prevIndex, err := db.rawActuaLLRPByProcessGuidAndIndex(logger, key.ProcessGuid, key.Index)
	if err != nil {
		logger.Error("failed-to-get-actual-lrp", err)
		return err
	}

	latestChangeTime := time.Duration(db.clock.Now().UnixNano() - lrp.Since)

	var newCrashCount int32
	if latestChangeTime > models.CrashResetTimeout && lrp.State == models.ActualLRPStateRunning {
		newCrashCount = 1
	} else {
		newCrashCount = lrp.CrashCount + 1
	}

	logger.Debug("retrieved-lrp")
	if !lrp.AllowsTransitionTo(key, instanceKey, models.ActualLRPStateCrashed) {
		err := fmt.Errorf("cannot transition crashed lrp from state %s to state %s", lrp.State, models.ActualLRPStateCrashed)
		logger.Error("failed-to-transition-actual", err)
		return models.ErrActualLRPCannotBeCrashed
	}

	if lrp.State == models.ActualLRPStateUnclaimed || lrp.State == models.ActualLRPStateCrashed ||
		((lrp.State == models.ActualLRPStateClaimed || lrp.State == models.ActualLRPStateRunning) &&
			!lrp.ActualLRPInstanceKey.Equal(instanceKey)) {
		logger.Debug("cannot-be-crashed", lager.Data{"state": lrp.State, "same-instance-key": lrp.ActualLRPInstanceKey.Equal(instanceKey)})
		return models.ErrActualLRPCannotBeCrashed
	}

	lrp.State = models.ActualLRPStateCrashed
	lrp.Since = db.clock.Now().UnixNano()
	lrp.CrashCount = newCrashCount
	lrp.ActualLRPInstanceKey = models.ActualLRPInstanceKey{}
	lrp.ActualLRPNetInfo = models.EmptyActualLRPNetInfo()
	lrp.ModificationTag.Increment()
	lrp.CrashReason = errorMessage

	var immediateRestart bool
	if lrp.ShouldRestartImmediately(models.NewDefaultRestartCalculator()) {
		lrp.State = models.ActualLRPStateUnclaimed
		immediateRestart = true
	}

	lrpRawJSON, err := json.Marshal(lrp)
	if err != nil {
		return models.ErrSerializeJSON
	}

	_, err = db.client.CompareAndSwap(ActualLRPSchemaPath(key.ProcessGuid, key.Index), lrpRawJSON, 0, prevIndex)
	if err != nil {
		logger.Error("failed", err)
		return models.ErrActualLRPCannotBeCrashed
	}

	if immediateRestart {
		auctionErr := db.requestLRPAuctionForLRPKey(logger, key)
		if auctionErr != nil {
			return auctionErr
		}
	}

	logger.Info("succeeded")
	return nil
}

func (db *ETCDDB) requestLRPAuctionForLRPKey(logger lager.Logger, key *models.ActualLRPKey) error {
	desiredLRP, err := db.DesiredLRPByProcessGuid(logger, key.ProcessGuid)
	bbsErr := models.ConvertError(err)
	if bbsErr != nil {
		if bbsErr.Type == models.Error_ResourceNotFound {
			_, err := db.client.Delete(ActualLRPSchemaPath(key.ProcessGuid, key.Index), false)
			if err != nil {
				logger.Error("failed-to-delete-actual", err)
				return models.ErrUnknownError
			}
			return err
		} else {
			return err
		}
	}

	lrpStart := models.NewLRPStartRequest(desiredLRP, uint(key.Index))
	err = db.auctioneerClient.RequestLRPAuctions([]*models.LRPStartRequest{&lrpStart})
	if err != nil {
		logger.Error("failed-to-request-auction", err)
		return models.ErrUnknownError
	}
	return nil
}

func (db *ETCDDB) FailActualLRP(logger lager.Logger, key *models.ActualLRPKey, errorMessage string) error {
	logger = logger.Session("fail-actual-lrp", lager.Data{"actual_lrp_key": key})
	logger.Info("starting")
	lrp, prevIndex, err := db.rawActuaLLRPByProcessGuidAndIndex(logger, key.ProcessGuid, key.Index)
	if err != nil {
		logger.Error("failed-to-get-actual-lrp", err)
		return err
	}

	if lrp.State != models.ActualLRPStateUnclaimed {
		return models.ErrActualLRPCannotBeFailed
	}

	lrp.ModificationTag.Increment()
	lrp.PlacementError = errorMessage
	lrp.Since = db.clock.Now().UnixNano()

	lrpRawJSON, err := json.Marshal(lrp)
	if err != nil {
		return models.ErrSerializeJSON
	}

	_, err = db.client.CompareAndSwap(ActualLRPSchemaPath(key.ProcessGuid, key.Index), lrpRawJSON, 0, prevIndex)
	if err != nil {
		logger.Error("failed", err)
		return models.ErrActualLRPCannotBeFailed
	}

	logger.Info("succeeded")
	return nil
}

func (db *ETCDDB) RemoveActualLRP(logger lager.Logger, processGuid string, index int32) error {
	logger = logger.Session("remove-actual-lrp", lager.Data{"process_guid": processGuid, "index": index})
	lrp, prevIndex, err := db.rawActuaLLRPByProcessGuidAndIndex(logger, processGuid, index)
	if err != nil {
		return err
	}

	return db.removeActualLRP(logger, lrp, prevIndex)
}

func (db *ETCDDB) RetireActualLRP(logger lager.Logger, key *models.ActualLRPKey) error {
	logger = logger.Session("retire-actual-lrp", lager.Data{"actual_lrp_key": key})
	var err error
	var prevIndex uint64
	var lrp *models.ActualLRP
	processGuid := key.ProcessGuid
	index := key.Index

	for i := 0; i < models.RetireActualLRPRetryAttempts; i++ {
		lrp, prevIndex, err = db.rawActuaLLRPByProcessGuidAndIndex(logger, processGuid, index)
		if err != nil {
			break
		}

		switch lrp.State {
		case models.ActualLRPStateUnclaimed, models.ActualLRPStateCrashed:
			err = db.removeActualLRP(logger, lrp, prevIndex)
		default:
			var cell *models.CellPresence
			key := lrp.ActualLRPKey
			instanceKey := lrp.ActualLRPInstanceKey
			cell, err = db.cellDB.CellById(logger, instanceKey.CellId)
			if err != nil {
				bbsErr := models.ConvertError(err)
				if bbsErr.Type == models.Error_ResourceNotFound {
					err = db.removeActualLRP(logger, lrp, prevIndex)
				}
				err = err
				break
			}

			logger.Info("stopping-lrp-instance", lager.Data{
				"actual-lrp-key": key,
			})
			cellErr := db.cellClient.StopLRPInstance(cell.RepAddress, key, instanceKey)
			if cellErr != nil {
				err = models.ErrActualLRPCannotBeStopped
			}
		}

		if err == nil {
			break
		}

		if i+1 < models.RetireActualLRPRetryAttempts {
			logger.Error("retrying-failed-retire-of-actual-lrp", err, lager.Data{"attempt": i + 1})
		}
	}

	return err
}

func (db *ETCDDB) retireActualLRPs(logger lager.Logger, keys []*models.ActualLRPKey) {
	logger = logger.Session("retire-actual-lrps")

	works := make([]func(), len(keys))

	for i, key := range keys {
		key := key

		works[i] = func() {
			err := db.RetireActualLRP(logger, key)
			if err != nil {
				logger.Error("failed-to-retire", err, lager.Data{"lrp-key": key})
			}
		}
	}

	throttler, err := workpool.NewThrottler(retireActualThrottlerSize, works)
	if err != nil {
		logger.Error("failed-constructing-throttler", err, lager.Data{"max-workers": retireActualThrottlerSize, "num-works": len(works)})
		return
	}

	throttler.Work()
}

func (db *ETCDDB) removeActualLRP(logger lager.Logger, lrp *models.ActualLRP, prevIndex uint64) error {
	logger.Info("starting")
	_, err := db.client.CompareAndDelete(ActualLRPSchemaPath(lrp.ProcessGuid, lrp.Index), prevIndex)
	if err != nil {
		logger.Error("failed", err)
		return models.ErrActualLRPCannotBeRemoved
	}
	logger.Info("succeeded")
	return nil
}

func parseActualLRPGroups(logger lager.Logger, node *etcd.Node, filter models.ActualLRPFilter) ([]*models.ActualLRPGroup, error) {
	var groups = []*models.ActualLRPGroup{}

	logger.Debug("performing-parsing-actual-lrp-groups")
	for _, indexNode := range node.Nodes {
		group := &models.ActualLRPGroup{}
		for _, instanceNode := range indexNode.Nodes {
			var lrp models.ActualLRP
			deserializeErr := models.FromJSON([]byte(instanceNode.Value), &lrp)
			if deserializeErr != nil {
				logger.Error("failed-parsing-actual-lrp-groups", deserializeErr, lager.Data{"key": instanceNode.Key})
				return []*models.ActualLRPGroup{}, models.ErrDeserializeJSON
			}
			if filter.Domain != "" && lrp.Domain != filter.Domain {
				continue
			}
			if filter.CellID != "" && lrp.CellId != filter.CellID {
				continue
			}

			if isInstanceActualLRPNode(instanceNode) {
				group.Instance = &lrp
			}

			if isEvacuatingActualLRPNode(instanceNode) {
				group.Evacuating = &lrp
			}
		}

		if group.Instance != nil || group.Evacuating != nil {
			groups = append(groups, group)
		}
	}
	logger.Debug("succeeded-performing-parsing-actual-lrp-groups", lager.Data{"num-actual-lrp-groups": len(groups)})

	return groups, nil
}

func (db *ETCDDB) unclaimActualLRP(
	logger lager.Logger,
	actualLRPKey *models.ActualLRPKey,
	actualLRPInstanceKey *models.ActualLRPInstanceKey,
) (stateChange, error) {
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
) (change stateChange, err error) {
	logger = logger.Session("unclaim-actual-lrp-with-index")
	defer func() {
		logger.Debug("complete", lager.Data{"stateChange": change, "error": err})
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

	err = lrp.Validate()
	if err != nil {
		logger.Error("failed-to-validate-unclaimed-lrp", err)
		return stateDidNotChange, models.NewError(models.Error_InvalidRecord, err.Error())
	}

	lrpRawJSON, err := json.Marshal(lrp)
	if err != nil {
		logger.Error("failed-to-marshal-unclaimed-lrp", err)
		return stateDidNotChange, models.ErrSerializeJSON
	}

	_, err = db.client.CompareAndSwap(ActualLRPSchemaPath(actualLRPKey.ProcessGuid, actualLRPKey.Index), lrpRawJSON, 0, storeIndex)
	if err != nil {
		logger.Error("failed-to-compare-and-swap", err)
		return stateDidNotChange, models.ErrActualLRPCannotBeUnclaimed
	}

	logger.Debug("changed-to-unclaimed")
	return stateDidChange, nil
}

func isInstanceActualLRPNode(node *etcd.Node) bool {
	return path.Base(node.Key) == ActualLRPInstanceKey
}

func isEvacuatingActualLRPNode(node *etcd.Node) bool {
	return path.Base(node.Key) == ActualLRPEvacuatingKey
}
