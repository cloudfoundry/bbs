package etcd

import (
	"encoding/json"
	"path"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry/gunk/workpool"
	"github.com/coreos/go-etcd/etcd"
	"github.com/nu7hatch/gouuid"
	"github.com/pivotal-golang/lager"
)

const maxActualGroupGetterWorkPoolSize = 50
const ActualLRPSchemaRoot = DataSchemaRoot + "actual"
const ActualLRPInstanceKey = "instance"
const ActualLRPEvacuatingKey = "evacuating"

func ActualLRPProcessDir(processGuid string) string {
	return path.Join(ActualLRPSchemaRoot, processGuid)
}

func ActualLRPIndexDir(processGuid string, index int32) string {
	return path.Join(ActualLRPProcessDir(processGuid), strconv.Itoa(int(index)))
}
func ActualLRPSchemaPath(processGuid string, index int32) string {
	return path.Join(ActualLRPIndexDir(processGuid, index), ActualLRPInstanceKey)
}

func EvacuatingActualLRPSchemaPath(processGuid string, index int32) string {
	return path.Join(ActualLRPIndexDir(processGuid, index), ActualLRPEvacuatingKey)
}

func (db *ETCDDB) ActualLRPGroups(logger lager.Logger, filter models.ActualLRPFilter) (*models.ActualLRPGroups, *models.Error) {
	node, bbsErr := db.fetchRecursiveRaw(logger, ActualLRPSchemaRoot)
	if bbsErr.Equal(models.ErrResourceNotFound) {
		return &models.ActualLRPGroups{}, nil
	}
	if bbsErr != nil {
		return nil, bbsErr
	}
	if node.Nodes.Len() == 0 {
		return &models.ActualLRPGroups{}, nil
	}

	groups := &models.ActualLRPGroups{}

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
			groups.ActualLrpGroups = append(groups.ActualLrpGroups, g.ActualLrpGroups...)
			groupsLock.Unlock()
		})
	}

	throttler, err := workpool.NewThrottler(maxActualGroupGetterWorkPoolSize, works)
	if err != nil {
		logger.Error("failed-constructing-throttler", err, lager.Data{"max-workers": maxActualGroupGetterWorkPoolSize, "num-works": len(works)})
		return &models.ActualLRPGroups{}, models.ErrUnknownError
	}

	logger.Debug("performing-deserialization-work")
	throttler.Work()
	if err, ok := workErr.Load().(error); ok {
		logger.Error("failed-performing-deserialization-work", err)
		return &models.ActualLRPGroups{}, models.ErrUnknownError
	}
	logger.Debug("succeeded-performing-deserialization-work", lager.Data{"num-actual-lrp-groups": len(groups.ActualLrpGroups)})

	return groups, nil
}

func (db *ETCDDB) ActualLRPGroupsByProcessGuid(logger lager.Logger, processGuid string) (*models.ActualLRPGroups, *models.Error) {
	node, bbsErr := db.fetchRecursiveRaw(logger, ActualLRPProcessDir(processGuid))
	if bbsErr.Equal(models.ErrResourceNotFound) {
		return &models.ActualLRPGroups{}, nil
	}
	if bbsErr != nil {
		return nil, bbsErr
	}
	if node.Nodes.Len() == 0 {
		return &models.ActualLRPGroups{}, nil
	}

	return parseActualLRPGroups(logger, node, models.ActualLRPFilter{})
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

func (db *ETCDDB) ClaimActualLRP(logger lager.Logger, processGuid string, index int32, instanceKey *models.ActualLRPInstanceKey) (*models.ActualLRP, *models.Error) {
	lrp, prevIndex, bbsErr := db.rawActuaLLRPByProcessGuidAndIndex(logger, processGuid, index)
	if bbsErr != nil {
		return nil, bbsErr
	}

	prevValue, err := json.Marshal(lrp)
	if err != nil {
		return nil, models.ErrSerializeJSON
	}

	if !lrp.AllowsTransitionTo(lrp.ActualLRPKey, *instanceKey, models.ActualLRPStateClaimed) {
		return nil, models.ErrActualLRPCannotBeClaimed
	}

	lrp.PlacementError = ""
	lrp.State = models.ActualLRPStateClaimed
	lrp.ActualLRPInstanceKey = *instanceKey
	lrp.ActualLRPNetInfo = models.ActualLRPNetInfo{}
	lrp.ModificationTag.Increment()

	err = lrp.Validate()
	if err != nil {
		return nil, &models.Error{Type: models.InvalidRecord, Message: err.Error()}
	}

	lrpRawJSON, err := json.Marshal(lrp)
	if err != nil {
		return nil, models.ErrSerializeJSON
	}

	logger.Info("starting")
	_, err = db.client.CompareAndSwap(ActualLRPSchemaPath(processGuid, index), string(lrpRawJSON), 0, string(prevValue), prevIndex)
	if err != nil {
		logger.Error("failed", err)
		return nil, models.ErrActualLRPCannotBeClaimed
	}
	logger.Info("succeeded")

	return lrp, nil
}

func (db *ETCDDB) createRunningActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, netInfo *models.ActualLRPNetInfo) (*models.ActualLRP, *models.Error) {
	guid, err := uuid.NewV4()
	if err != nil {
		return nil, models.ErrActualLRPCannotBeStarted
	}
	lrp := &models.ActualLRP{
		ActualLRPKey:         *key,
		ActualLRPInstanceKey: *instanceKey,
		ActualLRPNetInfo:     *netInfo,
		Since:                db.clock.Now().UnixNano(),
		State:                models.ActualLRPStateRunning,
		ModificationTag: models.ModificationTag{
			Epoch: guid.String(),
			Index: 0,
		},
	}

	lrpRawJSON, err := json.Marshal(lrp)
	if err != nil {
		return nil, models.ErrSerializeJSON
	}

	_, err = db.client.Set(ActualLRPSchemaPath(key.ProcessGuid, key.Index), string(lrpRawJSON), 0)
	if err != nil {
		logger.Error("failed", err)
		return nil, models.ErrActualLRPCannotBeStarted
	}
	return lrp, nil
}

func (db *ETCDDB) StartActualLRP(logger lager.Logger, request *models.StartActualLRPRequest) (*models.ActualLRP, *models.Error) {
	key := request.ActualLrpKey
	instanceKey := request.ActualLrpInstanceKey
	netInfo := request.ActualLrpNetInfo

	logger.Info("starting")
	lrp, prevIndex, bbsErr := db.rawActuaLLRPByProcessGuidAndIndex(logger, key.ProcessGuid, key.Index)
	if bbsErr == models.ErrResourceNotFound {
		return db.createRunningActualLRP(logger, key, instanceKey, netInfo)
	} else if bbsErr != nil {
		logger.Error("failed-to-get-actual-lrp", bbsErr)
		return nil, bbsErr
	}

	prevValue, err := json.Marshal(lrp)
	if err != nil {
		return nil, models.ErrSerializeJSON
	}

	if lrp.ActualLRPKey.Equal(key) &&
		lrp.ActualLRPInstanceKey.Equal(instanceKey) &&
		lrp.ActualLRPNetInfo.Equal(netInfo) &&
		lrp.State == models.ActualLRPStateRunning {
		logger.Info("succeeded")
		return lrp, nil
	}

	if !lrp.AllowsTransitionTo(*key, *instanceKey, models.ActualLRPStateRunning) {
		logger.Error("failed-to-transition-actual-lrp-to-started", nil)
		return nil, models.ErrActualLRPCannotBeStarted
	}

	lrp.State = models.ActualLRPStateRunning
	lrp.Since = db.clock.Now().UnixNano()
	lrp.ActualLRPInstanceKey = *instanceKey
	lrp.ActualLRPNetInfo = *netInfo
	lrp.PlacementError = ""

	lrpRawJSON, err := json.Marshal(lrp)
	if err != nil {
		return nil, models.ErrSerializeJSON
	}

	_, err = db.client.CompareAndSwap(ActualLRPSchemaPath(key.ProcessGuid, key.Index), string(lrpRawJSON), 0, string(prevValue), prevIndex)
	if err != nil {
		logger.Error("failed", err)
		return nil, models.ErrActualLRPCannotBeStarted
	}

	logger.Info("succeeded")
	return lrp, nil
}

func (db *ETCDDB) RemoveActualLRP(logger lager.Logger, processGuid string, index int32) *models.Error {
	lrp, prevIndex, bbsErr := db.rawActuaLLRPByProcessGuidAndIndex(logger, processGuid, index)
	if bbsErr != nil {
		return bbsErr
	}

	prevValue, err := json.Marshal(lrp)
	if err != nil {
		return models.ErrSerializeJSON
	}

	logger.Info("starting")
	_, err = db.client.CompareAndDelete(ActualLRPSchemaPath(processGuid, index), string(prevValue), prevIndex)
	if err != nil {
		logger.Error("failed", err)
		return models.ErrActualLRPCannotBeRemoved
	}
	logger.Info("succeeded")

	return nil
}

func parseActualLRPGroups(logger lager.Logger, node *etcd.Node, filter models.ActualLRPFilter) (*models.ActualLRPGroups, *models.Error) {
	var groups = &models.ActualLRPGroups{}

	logger.Debug("performing-parsing-actual-lrp-groups")
	for _, indexNode := range node.Nodes {
		group := &models.ActualLRPGroup{}
		for _, instanceNode := range indexNode.Nodes {
			var lrp models.ActualLRP
			deserializeErr := models.FromJSON([]byte(instanceNode.Value), &lrp)
			if deserializeErr != nil {
				logger.Error("failed-parsing-actual-lrp-groups", deserializeErr, lager.Data{"key": instanceNode.Key})
				return &models.ActualLRPGroups{}, models.ErrDeserializeJSON
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
			groups.ActualLrpGroups = append(groups.ActualLrpGroups, group)
		}
	}
	logger.Debug("succeeded-performing-parsing-actual-lrp-groups", lager.Data{"num-actual-lrp-groups": len(groups.ActualLrpGroups)})

	return groups, nil
}

func isInstanceActualLRPNode(node *etcd.Node) bool {
	return path.Base(node.Key) == ActualLRPInstanceKey
}

func isEvacuatingActualLRPNode(node *etcd.Node) bool {
	return path.Base(node.Key) == ActualLRPEvacuatingKey
}
