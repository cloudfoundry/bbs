package db

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter . ActualLRPDB

type ActualLRPDB interface {
	ActualLRPGroups(logger lager.Logger, filter models.ActualLRPFilter) ([]*models.ActualLRPGroup, error)
	ActualLRPGroupsByProcessGuid(logger lager.Logger, processGuid string) ([]*models.ActualLRPGroup, error)
	ActualLRPGroupByProcessGuidAndIndex(logger lager.Logger, processGuid string, index int32) (*models.ActualLRPGroup, error)

	CreateUnclaimedActualLRP(logger lager.Logger, key *models.ActualLRPKey) (actualLRPGroup *models.ActualLRPGroup, err error)
	UnclaimActualLRP(logger lager.Logger, key *models.ActualLRPKey) (beforeActualLRPGroup *models.ActualLRPGroup, err error)
	ClaimActualLRP(logger lager.Logger, processGuid string, index int32, instanceKey *models.ActualLRPInstanceKey) (beforeActualLRPGroup *models.ActualLRPGroup, err error)
	StartActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, netInfo *models.ActualLRPNetInfo) (beforeActualLRPGroup *models.ActualLRPGroup, updated bool, err error)
	CrashActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, crashReason string) (beforeActualLRPGroup *models.ActualLRPGroup, shouldRestart bool, err error)
	FailActualLRP(logger lager.Logger, key *models.ActualLRPKey, placementError string) (beforeActualLRPGroup *models.ActualLRPGroup, err error)
	RemoveActualLRP(logger lager.Logger, processGuid string, index int32) error
}
