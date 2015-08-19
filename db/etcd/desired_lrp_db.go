package etcd

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry/gunk/workpool"
	"github.com/pivotal-golang/lager"
)

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

func (db *ETCDDB) DesiredLRPByProcessGuid(logger lager.Logger, processGuid string) (*models.DesiredLRP, *models.Error) {
	node, bbsErr := db.fetchRaw(logger, DesiredLRPSchemaPathByProcessGuid(processGuid))
	if bbsErr != nil {
		return nil, bbsErr
	}

	var lrp models.DesiredLRP
	deserializeErr := models.FromJSON([]byte(node.Value), &lrp)
	if deserializeErr != nil {
		logger.Error("failed-parsing-desired-lrp", deserializeErr)
		return nil, models.ErrDeserializeJSON
	}

	return &lrp, nil
}
