package etcd

import (
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

func (db *ETCDDB) CreateLRPDeployment(logger lager.Logger, lrp *models.LRPDeploymentCreation) (*models.LRPDeployment, error) {
	return nil, nil
}

func (db *ETCDDB) UpdateLRPDeployment(logger lager.Logger, id string, update *models.LRPDeploymentUpdate) (*models.LRPDeployment, error) {
	return nil, nil
}

func (db *ETCDDB) SaveLRPDeployment(logger lager.Logger, lrp *models.LRPDeployment) (*models.LRPDeployment, error) {
	return nil, nil
}

func (db *ETCDDB) LRPDeploymentByDefinitionGuid(logger lager.Logger, guid string) (*models.LRPDeployment, error) {
	return nil, nil
}

func (db *ETCDDB) DeleteLRPDeployment(logger lager.Logger, id string) (*models.LRPDeployment, error) {
	return nil, nil
}

func (db *ETCDDB) ActivateLRPDeploymentDefinition(logger lager.Logger, id, definitionId string) (*models.LRPDeployment, error) {
	return nil, nil
}

func (db *ETCDDB) LRPDeploymentByProcessGuid(logger lager.Logger, guid string) (*models.LRPDeployment, error) {
	return nil, nil
}

func (db *ETCDDB) LRPDeployments(logger lager.Logger, ids []string) ([]*models.LRPDeployment, error) {
	return nil, nil
}
