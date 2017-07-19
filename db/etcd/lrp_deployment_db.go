package etcd

import (
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

func (db *ETCDDB) CreateLRPDeployment(logger lager.Logger, lrp *models.LRPDeploymentDefinition) error {
	return nil
}

func (db *ETCDDB) UpdateLRPDeployment(logger lager.Logger, id string, update *models.LRPDeploymentUpdate) (*models.LRPDeploymentDefinition, error) {
	return nil, nil
}

func (db *ETCDDB) DeleteLRPDeployment(logger lager.Logger, id string) error {
	return nil
}

func (db *ETCDDB) ActivateLRPDeploymentDefinition(logger lager.Logger, id, definitionId string) error {
	return nil
}
