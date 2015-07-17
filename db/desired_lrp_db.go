package db

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter . DesiredLRPDB
type DesiredLRPDB interface {
	DesiredLRPs(logger lager.Logger, filter models.DesiredLRPFilter) (*models.DesiredLRPs, *models.Error)
	DesiredLRPByProcessGuid(logger lager.Logger, processGuid string) (*models.DesiredLRP, *models.Error)
}
