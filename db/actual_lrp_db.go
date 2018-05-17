package db

import (
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

//go:generate counterfeiter . ActualLRPDB

type ActualLRPDB interface {
	ActualLRPGroups(logger lager.Logger, filter models.ActualLRPFilter) ([]*models.ActualLRPGroup, error)
	ActualLRPGroupsByProcessGuid(logger lager.Logger, processGuid string) ([]*models.ActualLRPGroup, error)
	ActualLRPGroupByProcessGuidAndIndex(logger lager.Logger, processGuid string, index int32) (*models.ActualLRPGroup, error)

	ActualLRPs(logger lager.Logger, filter models.ActualLRPFilter) ([]*models.ActualLRP, error)
	// ActualLRPsByProcessGuid(logger lager.Logger, processGuid string) ([]*models.ActualLRP, error)
	// ActualLRPByProcessGuidAndIndex(logger lager.Logger, processGuid string, index int32) ([][][][]*models.ActualLRP, error)

	CreateUnclaimedActualLRP(logger lager.Logger, key *models.ActualLRPKey) (after *models.ActualLRP, err error)
	UnclaimActualLRP(logger lager.Logger, key *models.ActualLRPKey) (before *models.ActualLRP, after *models.ActualLRP, err error)
	ClaimActualLRP(logger lager.Logger, processGuid string, index int32, instanceKey *models.ActualLRPInstanceKey) (before *models.ActualLRP, after *models.ActualLRP, err error)
	StartActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, netInfo *models.ActualLRPNetInfo) (before *models.ActualLRP, after *models.ActualLRP, err error)
	CrashActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, crashReason string) (before *models.ActualLRP, after *models.ActualLRP, shouldRestart bool, err error)
	FailActualLRP(logger lager.Logger, key *models.ActualLRPKey, placementError string) (before *models.ActualLRP, after *models.ActualLRP, err error)
	RemoveActualLRP(logger lager.Logger, processGuid string, index int32, instanceKey *models.ActualLRPInstanceKey) error
}
