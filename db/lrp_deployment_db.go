package db

import (
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

//go:generate counterfeiter . DesiredLRPDB

type LRPDeploymentDB interface {
	CreateLRPDeployment(logger lager.Logger, lrp *models.LRPDeploymentDefinition) error
	UpdateLRPDeployment(logger lager.Logger, id string, definition *models.LRPDeploymentUpdate) (*models.LRPDeploymentDefinition, error)
	DeleteLRPDeployment(logger lager.Logger, id string) error
	ActivateLRPDeploymentDefinition(logger lager.Logger, id string, definitionID string) error
}
