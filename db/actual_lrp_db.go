package db

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter . ActualLRPDB
type ActualLRPDB interface {
	ActualLRPGroups(filter models.ActualLRPFilter, logger lager.Logger) (*models.ActualLRPGroups, *models.Error)
	ActualLRPGroupsByProcessGuid(processGuid string, logger lager.Logger) (*models.ActualLRPGroups, *models.Error)
	ActualLRPGroupByProcessGuidAndIndex(processGuid string, index int32, logger lager.Logger) (*models.ActualLRPGroup, *models.Error)
}
