package db

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter . ActualLRPDB
type ActualLRPDB interface {
	ActualLRPGroups(logger lager.Logger, filter models.ActualLRPFilter) (*models.ActualLRPGroups, *models.Error)
	ActualLRPGroupsByProcessGuid(logger lager.Logger, processGuid string) (*models.ActualLRPGroups, *models.Error)
	ActualLRPGroupByProcessGuidAndIndex(logger lager.Logger, processGuid string, index int32) (*models.ActualLRPGroup, *models.Error)
}
