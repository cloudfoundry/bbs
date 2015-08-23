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

func (db *ETCDDB) ActualLRPGroups(logger lager.Logger, filter models.ActualLRPFilter) ([]*models.ActualLRPGroup, *models.Error) {
	node, bbsErr := db.fetchRecursiveRaw(logger, ActualLRPSchemaRoot)
	if bbsErr.Equal(models.ErrResourceNotFound) {
		return []*models.ActualLRPGroup{}, nil
	}
	if bbsErr != nil {
		return nil, bbsErr
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

func (db *ETCDDB) ActualLRPGroupsByProcessGuid(logger lager.Logger, processGuid string) ([]*models.ActualLRPGroup, *models.Error) {
	node, bbsErr := db.fetchRecursiveRaw(logger, ActualLRPProcessDir(processGuid))
	if bbsErr.Equal(models.ErrResourceNotFound) {
		return []*models.ActualLRPGroup{}, nil
	}
	if bbsErr != nil {
		return nil, bbsErr
	}
	if node.Nodes.Len() == 0 {
		return []*models.ActualLRPGroup{}, nil
	}

	return parseActualLRPGroups(logger, node, models.ActualLRPFilter{})
}

func (db *ETCDDB) instanceActualLRPsByProcessGuid(logger lager.Logger, processGuid string) (map[int32]*models.ActualLRP, *models.Error) {
	node, bbsErr := db.fetchRecursiveRaw(logger, ActualLRPProcessDir(processGuid))
	if bbsErr.Equal(models.ErrResourceNotFound) {
		return nil, nil
	}
	if bbsErr != nil {
		return nil, bbsErr
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

func (db *ETCDDB) ActualLRPGroupByProcessGuidAndIndex(logger lager.Logger, processGuid string, index int32) (*models.ActualLRPGroup, *models.Error) {
	group, _, err := db.rawActualLRPGroupByProcessGuidAndIndex(logger, processGuid, index)
	return group, err
}

func (db *ETCDDB) rawActualLRPGroupByProcessGuidAndIndex(logger lager.Logger, processGuid string, index int32) (*models.ActualLRPGroup, uint64, *models.Error) {
	node, bbsErr := db.fetchRecursiveRaw(logger, ActualLRPIndexDir(processGuid, index))
	if bbsErr != nil {
		return nil, 0, bbsErr
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

func (db *ETCDDB) rawActuaLLRPByProcessGuidAndIndex(logger lager.Logger, processGuid string, index int32) (*models.ActualLRP, uint64, *models.Error) {
	logger.Debug("raw-actual-lrp-by-process-guid-and-index")
	node, bbsErr := db.fetchRaw(logger, ActualLRPSchemaPath(processGuid, index))
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

func (db *ETCDDB) ClaimActualLRP(logger lager.Logger, processGuid string, index int32, instanceKey *models.ActualLRPInstanceKey) *models.Error {
	logger = logger.Session("claim-actual-lrp", lager.Data{"process_guid": processGuid, "index": index, "actual_lrp_instance-key": instanceKey})
	logger.Info("starting")

	lrp, prevIndex, bbsErr := db.rawActuaLLRPByProcessGuidAndIndex(logger, processGuid, index)
	if bbsErr != nil {
		logger.Error("failed", bbsErr)
		return bbsErr
	}

	if !lrp.AllowsTransitionTo(&lrp.ActualLRPKey, instanceKey, models.ActualLRPStateClaimed) {
		return models.ErrActualLRPCannotBeClaimed
	}

	lrp.PlacementError = ""
	lrp.State = models.ActualLRPStateClaimed
	lrp.ActualLRPInstanceKey = *instanceKey
	lrp.ActualLRPNetInfo = models.ActualLRPNetInfo{}
	lrp.ModificationTag.Increment()

	err := lrp.Validate()
	if err != nil {
		logger.Error("failed", err)
		return &models.Error{Type: models.InvalidRecord, Message: err.Error()}
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

func (db *ETCDDB) createRawActualLRP(logger lager.Logger, lrp *models.ActualLRP) *models.Error {
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

func (db *ETCDDB) createRunningActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, netInfo *models.ActualLRPNetInfo) *models.Error {
	lrp, err := db.newRunningActualLRP(key, instanceKey, netInfo)
	if err != nil {
		return models.ErrActualLRPCannotBeStarted
	}

	return db.createRawActualLRP(logger, lrp)
}

func (db *ETCDDB) createUnclaimedActualLRP(logger lager.Logger, key *models.ActualLRPKey) *models.Error {
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

func (db *ETCDDB) createEvacuatingActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, netInfo *models.ActualLRPNetInfo, evacuatingTTLInSeconds uint64) (modelErr *models.Error) {
	logger.Debug("create-evacuating-actual-lrp")
	defer func() { logger.Debug("create-evacuating-actual-lrp-complete", lager.Data{"error": modelErr}) }()
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

func (db *ETCDDB) StartActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, netInfo *models.ActualLRPNetInfo) *models.Error {
	logger = logger.Session("start-actual-lrp", lager.Data{"actual_lrp_key": key, "actual_lrp_instance_key": instanceKey})
	logger.Info("starting")
	lrp, prevIndex, bbsErr := db.rawActuaLLRPByProcessGuidAndIndex(logger, key.ProcessGuid, key.Index)
	if bbsErr == models.ErrResourceNotFound {
		return db.createRunningActualLRP(logger, key, instanceKey, netInfo)
	} else if bbsErr != nil {
		logger.Error("failed-to-get-actual-lrp", bbsErr)
		return bbsErr
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

func (db *ETCDDB) CrashActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, errorMessage string) *models.Error {
	logger = logger.Session("crash-actual-lrp", lager.Data{"actual_lrp_key": key, "actual_lrp_instance_key": instanceKey})
	logger.Info("starting")

	lrp, prevIndex, bbsErr := db.rawActuaLLRPByProcessGuidAndIndex(logger, key.ProcessGuid, key.Index)
	if bbsErr != nil {
		logger.Error("failed-to-get-actual-lrp", bbsErr)
		return bbsErr
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

func (db *ETCDDB) requestLRPAuctionForLRPKey(logger lager.Logger, key *models.ActualLRPKey) *models.Error {
	desiredLRP, bbsErr := db.DesiredLRPByProcessGuid(logger, key.ProcessGuid)
	if bbsErr == models.ErrResourceNotFound {
		_, err := db.client.Delete(ActualLRPSchemaPath(key.ProcessGuid, key.Index), false)
		if err != nil {
			logger.Error("failed-to-delete-actual", err)
			return models.ErrUnknownError
		}
	} else if bbsErr != nil {
		return bbsErr
	}

	lrpStart := models.NewLRPStartRequest(desiredLRP, uint(key.Index))
	err := db.auctioneerClient.RequestLRPAuctions([]*models.LRPStartRequest{&lrpStart})
	if err != nil {
		logger.Error("failed-to-request-auction", err)
		return models.ErrUnknownError
	}
	return nil
}

func (db *ETCDDB) FailActualLRP(logger lager.Logger, key *models.ActualLRPKey, errorMessage string) *models.Error {
	logger = logger.Session("fail-actual-lrp", lager.Data{"actual_lrp_key": key})
	logger.Info("starting")
	lrp, prevIndex, bbsErr := db.rawActuaLLRPByProcessGuidAndIndex(logger, key.ProcessGuid, key.Index)
	if bbsErr != nil {
		logger.Error("failed-to-get-actual-lrp", bbsErr)
		return bbsErr
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

func (db *ETCDDB) RemoveActualLRP(logger lager.Logger, processGuid string, index int32) *models.Error {
	logger = logger.Session("remove-actual-lrp", lager.Data{"process_guid": processGuid, "index": index})
	lrp, prevIndex, bbsErr := db.rawActuaLLRPByProcessGuidAndIndex(logger, processGuid, index)
	if bbsErr != nil {
		return bbsErr
	}

	return db.removeActualLRP(logger, lrp, prevIndex)
}

func (db *ETCDDB) RetireActualLRP(logger lager.Logger, key *models.ActualLRPKey) *models.Error {
	logger = logger.Session("retire-actual-lrp", lager.Data{"actual_lrp_key": key})
	var err *models.Error
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
				if err == models.ErrResourceNotFound {
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

func (db *ETCDDB) removeActualLRP(logger lager.Logger, lrp *models.ActualLRP, prevIndex uint64) *models.Error {
	logger.Info("starting")
	_, err := db.client.CompareAndDelete(ActualLRPSchemaPath(lrp.ProcessGuid, lrp.Index), prevIndex)
	if err != nil {
		logger.Error("failed", err)
		return models.ErrActualLRPCannotBeRemoved
	}
	logger.Info("succeeded")
	return nil
}

func parseActualLRPGroups(logger lager.Logger, node *etcd.Node, filter models.ActualLRPFilter) ([]*models.ActualLRPGroup, *models.Error) {
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

func isInstanceActualLRPNode(node *etcd.Node) bool {
	return path.Base(node.Key) == ActualLRPInstanceKey
}

func isEvacuatingActualLRPNode(node *etcd.Node) bool {
	return path.Base(node.Key) == ActualLRPEvacuatingKey
}
