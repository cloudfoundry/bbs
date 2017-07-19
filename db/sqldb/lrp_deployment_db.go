package sqldb

import (
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

func (db *SQLDB) CreateLRPDeployment(logger lager.Logger, lrp *models.LRPDeploymentDefinition) error {
	return nil
}

func (db *SQLDB) UpdateLRPDeployment(logger lager.Logger, id string, update *models.LRPDeploymentUpdate) (*models.LRPDeploymentDefinition, error) {
	return nil, nil
}

func (db *SQLDB) DeleteLRPDeployment(logger lager.Logger, id string) error {
	return nil
}

func (db *SQLDB) ActivateLRPDeploymentDefinition(logger lager.Logger, id, definitionId string) error {
	return nil
}
