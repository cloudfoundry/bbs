package etcd

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry/gunk/workpool"
	"github.com/nu7hatch/gouuid"
	"github.com/pivotal-golang/lager"
)

const createActualMaxWorkers = 100

func (db *ETCDDB) DesiredLRPs(logger lager.Logger, filter models.DesiredLRPFilter) ([]*models.DesiredLRP, *models.Error) {
	root, bbsErr := db.fetchRecursiveRaw(logger, DesiredLRPSchemaRoot)
	if bbsErr.Equal(models.ErrResourceNotFound) {
		return []*models.DesiredLRP{}, nil
	}
	if bbsErr != nil {
		return nil, bbsErr
	}
	if root.Nodes.Len() == 0 {
		return []*models.DesiredLRP{}, nil
	}

	desiredLRPs := []*models.DesiredLRP{}

	lrpsLock := sync.Mutex{}
	var workErr atomic.Value
	works := []func(){}

	for _, node := range root.Nodes {
		node := node

		works = append(works, func() {
			var lrp models.DesiredLRP
			deserializeErr := models.FromJSON([]byte(node.Value), &lrp)
			if deserializeErr != nil {
				logger.Error("failed-parsing-desired-lrp", deserializeErr)
				workErr.Store(fmt.Errorf("cannot parse lrp JSON for key %s: %s", node.Key, deserializeErr.Error()))
				return
			}

			if filter.Domain == "" || lrp.GetDomain() == filter.Domain {
				lrpsLock.Lock()
				desiredLRPs = append(desiredLRPs, &lrp)
				lrpsLock.Unlock()
			}
		})
	}

	throttler, err := workpool.NewThrottler(maxDesiredLRPGetterWorkPoolSize, works)
	if err != nil {
		logger.Error("failed-constructing-throttler", err, lager.Data{"max-workers": maxDesiredLRPGetterWorkPoolSize, "num-works": len(works)})
		return []*models.DesiredLRP{}, models.ErrUnknownError
	}

	logger.Debug("performing-deserialization-work")
	throttler.Work()
	if err, ok := workErr.Load().(error); ok {
		logger.Error("failed-performing-deserialization-work", err)
		return []*models.DesiredLRP{}, models.ErrUnknownError
	}
	logger.Debug("succeeded-performing-deserialization-work", lager.Data{"num-desired-lrps": len(desiredLRPs)})

	return desiredLRPs, nil
}

func (db *ETCDDB) rawDesiredLRPByProcessGuid(logger lager.Logger, processGuid string) (*models.DesiredLRP, uint64, *models.Error) {
	node, bbsErr := db.fetchRaw(logger, DesiredLRPSchemaPathByProcessGuid(processGuid))
	if bbsErr != nil {
		return nil, 0, bbsErr
	}

	var lrp models.DesiredLRP
	deserializeErr := models.FromJSON([]byte(node.Value), &lrp)
	if deserializeErr != nil {
		logger.Error("failed-parsing-desired-lrp", deserializeErr)
		return nil, 0, models.ErrDeserializeJSON
	}

	return &lrp, node.ModifiedIndex, nil
}

func (db *ETCDDB) DesiredLRPByProcessGuid(logger lager.Logger, processGuid string) (*models.DesiredLRP, *models.Error) {
	lrp, _, err := db.rawDesiredLRPByProcessGuid(logger, processGuid)
	return lrp, err
}

func (db *ETCDDB) startInstanceRange(logger lager.Logger, lower, upper int32, desiredLRP *models.DesiredLRP) {
	logger = logger.Session("start-instance-range", lager.Data{"lower": lower, "upper": upper})
	logger.Info("starting")
	defer logger.Info("complete")

	keys := make([]*models.ActualLRPKey, upper-lower)
	i := 0
	for actualIndex := lower; actualIndex < upper; actualIndex++ {
		key := models.NewActualLRPKey(desiredLRP.ProcessGuid, int32(actualIndex), desiredLRP.Domain)
		keys[i] = &key
		i++
	}

	createdIndices := db.createUnclaimedActualLRPs(logger, keys)
	start := models.NewLRPStartRequest(desiredLRP, createdIndices...)

	err := db.auctioneerClient.RequestLRPAuctions([]*models.LRPStartRequest{&start})
	if err != nil {
		logger.Error("failed-to-request-auction", err)
	}
}

func (db *ETCDDB) stopInstanceRange(logger lager.Logger, lower, upper int32, desiredLRP *models.DesiredLRP) {
	logger = logger.Session("stop-instance-range", lager.Data{"lower": lower, "upper": upper})
	logger.Info("starting")
	defer logger.Info("complete")

	actualsMap, err := db.instanceActualLRPsByProcessGuid(logger, desiredLRP.ProcessGuid)
	if err != nil {
		logger.Error("failed-to-get-actual-lrps", err)
		return
	}

	actualKeys := make([]*models.ActualLRPKey, 0)
	for i := lower; i < upper; i++ {
		actual, ok := actualsMap[i]
		if ok {
			actualKeys = append(actualKeys, &actual.ActualLRPKey)
		}
	}

	db.retireActualLRPs(logger, actualKeys)
}

func (db *ETCDDB) DesireLRP(logger lager.Logger, desiredLRP *models.DesiredLRP) *models.Error {
	logger = logger.Session("create-desired-lrp", lager.Data{"process-guid": desiredLRP.ProcessGuid})
	logger.Info("starting")
	defer logger.Info("complete")

	guid, err := uuid.NewV4()
	if err != nil {
		logger.Error("failed-to-generate-epoch", err)
		return models.ErrUnknownError
	}

	desiredLRP.ModificationTag = &models.ModificationTag{
		Epoch: guid.String(),
		Index: 0,
	}

	value, modelErr := models.ToJSON(desiredLRP)
	if modelErr != nil {
		logger.Error("failed-to-json", err)
		return models.ErrSerializeJSON
	}

	logger.Debug("persisting-desired-lrp")
	_, err = db.client.Create(DesiredLRPSchemaPath(desiredLRP), value, NO_TTL)
	if err != nil {
		return ErrorFromEtcdError(logger, err)
	}
	logger.Debug("succeeded-persisting-desired-lrp")

	db.startInstanceRange(logger, 0, desiredLRP.Instances, desiredLRP)
	return nil
}

func (db *ETCDDB) UpdateDesiredLRP(logger lager.Logger, processGuid string, update *models.DesiredLRPUpdate) *models.Error {
	logger = logger.Session("update-desired-lrp", lager.Data{"process-guid": processGuid})
	logger.Info("starting")
	defer logger.Info("complete")

	desiredLRP, index, modelErr := db.rawDesiredLRPByProcessGuid(logger, processGuid)
	if modelErr != nil {
		logger.Error("failed-to-fetch-existing-desired-lrp", modelErr)
		return modelErr
	}

	existingInstances := desiredLRP.Instances
	desiredLRP = desiredLRP.ApplyUpdate(update)

	desiredLRP.ModificationTag.Increment()

	value, modelErr := models.ToJSON(desiredLRP)
	if modelErr != nil {
		logger.Error("failed-to-serialize-desired-lrp", modelErr)
		return modelErr
	}

	_, err := db.client.CompareAndSwap(DesiredLRPSchemaPath(desiredLRP), value, NO_TTL, index)
	if err != nil {
		logger.Error("failed-to-CAS-desired-lrp", err)
		return models.ErrDesiredLRPCannotBeUpdated
	}

	diff := desiredLRP.Instances - existingInstances
	switch {
	case diff > 0:
		db.startInstanceRange(logger, existingInstances, desiredLRP.Instances, desiredLRP)

	case diff < 0:
		db.stopInstanceRange(logger, desiredLRP.Instances, existingInstances, desiredLRP)

	case diff == 0:
		// this space intentionally left blank
	}

	return nil
}

func (db *ETCDDB) RemoveDesiredLRP(logger lager.Logger, processGuid string) *models.Error {
	logger = logger.Session("remove-desired-lrp", lager.Data{"process-guid": processGuid})
	logger.Info("starting")
	defer logger.Info("complete")

	desiredLRP, modelErr := db.DesiredLRPByProcessGuid(logger, processGuid)
	if modelErr != nil {
		return modelErr
	}

	logger.Info("starting")
	_, err := db.client.Delete(DesiredLRPSchemaPathByProcessGuid(processGuid), true)
	if err != nil {
		logger.Error("failed", err)
		return models.ErrUnknownError
	}
	logger.Info("succeeded")

	db.stopInstanceRange(logger, 0, desiredLRP.Instances, desiredLRP)
	return nil
}
