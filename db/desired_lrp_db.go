package db

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter . DesiredLRPDB

type DesiredLRPDB interface {
	DesiredLRPs(logger lager.Logger, filter models.DesiredLRPFilter) ([]*models.DesiredLRP, *models.Error)
	DesiredLRPByProcessGuid(logger lager.Logger, processGuid string) (*models.DesiredLRP, *models.Error)

	DesireLRP(logger lager.Logger, desiredLRP *models.DesiredLRP) *models.Error
	UpdateDesiredLRP(logger lager.Logger, processGuid string, update *models.DesiredLRPUpdate) *models.Error
	RemoveDesiredLRP(logger lager.Logger, processGuid string) *models.Error
}
