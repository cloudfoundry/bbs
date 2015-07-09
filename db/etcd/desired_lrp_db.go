package db

import (
	"fmt"
	"path"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

const DesiredLRPSchemaRoot = DataSchemaRoot + "desired"

func DesiredLRPSchemaPath(lrp models.DesiredLRP) string {
	return DesiredLRPSchemaPathByProcessGuid(lrp.GetProcessGuid())
}

func DesiredLRPSchemaPathByProcessGuid(processGuid string) string {
	return path.Join(DesiredLRPSchemaRoot, processGuid)
}

func (db *ETCDDB) DesiredLRPs(logger lager.Logger) (*models.DesiredLRPs, error) {
	node, err := db.fetchRecursiveRaw(DesiredLRPSchemaRoot, logger)
	if err != nil {
		return &models.DesiredLRPs{}, nil
	}
	if node.Nodes.Len() == 0 {
		return &models.DesiredLRPs{}, nil
	}

	desiredLRPs := models.DesiredLRPs{}
	for _, instanceNode := range node.Nodes {
		var lrp models.DesiredLRP
		deserializeErr := models.FromJSON([]byte(instanceNode.Value), &lrp)
		if deserializeErr != nil {
			logger.Error("failed-parsing-desired-lrp", deserializeErr)
			return nil, fmt.Errorf("cannot parse lrp JSON for key %s: %s", instanceNode.Key, deserializeErr.Error())
		}
		desiredLRPs.DesiredLrps = append(desiredLRPs.DesiredLrps, &lrp)
	}
	return &desiredLRPs, nil
}
