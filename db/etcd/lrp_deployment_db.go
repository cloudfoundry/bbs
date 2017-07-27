package etcd

import (
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

func (db *ETCDDB) CreateLRPDeployment(logger lager.Logger, lrp *models.LRPDeploymentCreation) (string, error) {
	return "", nil
}

func (db *ETCDDB) UpdateLRPDeployment(logger lager.Logger, id string, update *models.LRPDeploymentUpdate) (string, error) {
	return "", nil
}

func (db *ETCDDB) DeleteLRPDeployment(logger lager.Logger, id string) error {
	return nil
}

func (db *ETCDDB) ActivateLRPDeploymentDefinition(logger lager.Logger, id, definitionId string) error {
	return nil
}
