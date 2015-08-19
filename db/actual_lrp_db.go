package db

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter . ActualLRPDB
type ActualLRPDB interface {
	ActualLRPGroups(logger lager.Logger, filter models.ActualLRPFilter) ([]*models.ActualLRPGroup, *models.Error)
	ActualLRPGroupsByProcessGuid(logger lager.Logger, processGuid string) ([]*models.ActualLRPGroup, *models.Error)
	ActualLRPGroupByProcessGuidAndIndex(logger lager.Logger, processGuid string, index int32) (*models.ActualLRPGroup, *models.Error)

	ClaimActualLRP(logger lager.Logger, processGuid string, index int32, instanceKey *models.ActualLRPInstanceKey) *models.Error
	StartActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, netInfo *models.ActualLRPNetInfo) *models.Error
	CrashActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, errorMessage string) *models.Error
	FailActualLRP(logger lager.Logger, key *models.ActualLRPKey, errorMessage string) *models.Error
	RemoveActualLRP(logger lager.Logger, processGuid string, index int32) *models.Error
	RetireActualLRP(logger lager.Logger, key *models.ActualLRPKey) *models.Error
}
