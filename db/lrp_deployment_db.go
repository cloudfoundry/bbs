package db

import (
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

//go:generate counterfeiter . LRPDeploymentDB

type LRPDeploymentDB interface {
	CreateLRPDeployment(logger lager.Logger, lrp *models.LRPDeploymentCreation) (*models.LRPDeployment, error)
	UpdateLRPDeployment(logger lager.Logger, id string, definition *models.LRPDeploymentUpdate) (*models.LRPDeployment, error)
	SaveLRPDeployment(logger lager.Logger, lrpDeployment *models.LRPDeployment, oldHealthyDefinitionId string) (*models.LRPDeployment, error)
	DeleteLRPDeployment(logger lager.Logger, id string) (*models.LRPDeployment, error)
	ActivateLRPDeploymentDefinition(logger lager.Logger, id string, definitionID string) (*models.LRPDeployment, error)
	// LRPDeploymentByDefinitionGuid(logger lager.Logger, id string) (*models.LRPDeployment, error)
	LRPDeploymentByProcessGuid(logger lager.Logger, id string) (*models.LRPDeployment, error)
	LRPDeployments(logger lager.Logger, deploymentIds []string) ([]*models.LRPDeployment, error)
	LRPDeploymentSchedulingInfo(logger lager.Logger, filter models.LRPDeploymentFilter) ([]*models.LRPDeploymentSchedulingInfo, error)
}
