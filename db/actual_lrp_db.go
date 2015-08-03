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

	ClaimActualLRP(logger lager.Logger, processGuid string, index int32, instanceKey *models.ActualLRPInstanceKey) (*models.ActualLRP, *models.Error)
	StartActualLRP(logger lager.Logger, request *models.StartActualLRPRequest) (*models.ActualLRP, *models.Error)
	CrashActualLRP(logger lager.Logger, request *models.CrashActualLRPRequest) *models.Error
	FailActualLRP(logger lager.Logger, request *models.FailActualLRPRequest) *models.Error
	RemoveActualLRP(logger lager.Logger, processGuid string, index int32) *models.Error
	RetireActualLRP(logger lager.Logger, request *models.RetireActualLRPRequest) *models.Error
}
