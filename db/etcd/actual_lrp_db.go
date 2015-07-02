package db

import (
	"fmt"
	"path"
	"strconv"
	"sync"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry/gunk/workpool"
	"github.com/coreos/go-etcd/etcd"
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

func (db *ETCDDB) ActualLRPGroups(filter models.ActualLRPFilter, logger lager.Logger) (*models.ActualLRPGroups, error) {
	logger.Debug("fetching-actual-lrps-from-bbs")
	response, err := db.client.Get(ActualLRPSchemaRoot, false, true)
	if etcdErr, ok := err.(*etcd.EtcdError); ok && etcdErr.ErrorCode == 100 {
		logger.Debug("no-actual-lrps-to-fetch")
		return &models.ActualLRPGroups{}, nil
	} else if err != nil {
		logger.Error("failed-fetching-actual-lrps-from-bbs", err)
		return &models.ActualLRPGroups{}, err
	}
	logger.Debug("succeeded-fetching-actual-lrps-from-bbs", lager.Data{"num-lrps": response.Node.Nodes.Len()})

	if response.Node.Nodes.Len() == 0 {
		return &models.ActualLRPGroups{}, nil
	}

	var groups = &models.ActualLRPGroups{}
	groupsLock := sync.Mutex{}
	var workErr error
	workErrLock := sync.Mutex{}

	works := []func(){}

	for _, node := range response.Node.Nodes {
		node := node

		works = append(works, func() {
			for _, indexNode := range node.Nodes {
				group := &models.ActualLRPGroup{}
				for _, instanceNode := range indexNode.Nodes {
					var lrp models.ActualLRP
					deserializeErr := models.FromJSON([]byte(instanceNode.Value), &lrp)
					if deserializeErr != nil {
						logger.Error("invalid-instance-node", deserializeErr)
						workErrLock.Lock()
						workErr = fmt.Errorf("cannot parse lrp JSON for key %s: %s", instanceNode.Key, deserializeErr.Error())
						workErrLock.Unlock()
						continue
					}
					if filter.Domain != "" && lrp.GetDomain() != filter.Domain {
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
					groupsLock.Lock()
					groups.ActualLrpGroups = append(groups.ActualLrpGroups, group)
					groupsLock.Unlock()
				}
			}
		})
	}

	throttler, err := workpool.NewThrottler(maxActualGroupGetterWorkPoolSize, works)
	if err != nil {
		logger.Error("failed-constructing-throttler", err, lager.Data{"max-workers": maxActualGroupGetterWorkPoolSize, "num-works": len(works)})
		return &models.ActualLRPGroups{}, err
	}

	logger.Debug("performing-deserialization-work")
	throttler.Work()
	if workErr != nil {
		logger.Error("failed-performing-deserialization-work", workErr)
		return &models.ActualLRPGroups{}, workErr
	}
	logger.Debug("succeeded-performing-deserialization-work", lager.Data{"num-actual-lrp-groups": len(groups.GetActualLrpGroups())})

	return groups, nil
}

func isInstanceActualLRPNode(node *etcd.Node) bool {
	return path.Base(node.Key) == ActualLRPInstanceKey
}

func isEvacuatingActualLRPNode(node *etcd.Node) bool {
	return path.Base(node.Key) == ActualLRPEvacuatingKey
}
