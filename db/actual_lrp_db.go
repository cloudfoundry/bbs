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

	ActualLRPs(logger lager.Logger, filter models.ActualLRPFilter) ([]*models.FlattenedActualLRP, error)
	// ActualLRPsByProcessGuid(logger lager.Logger, processGuid string) ([]*models.FlattenedActualLRP, error)
	// ActualLRPByProcessGuidAndIndex(logger lager.Logger, processGuid string, index int32) ([][][][]*models.FlattenedActualLRP, error)

	CreateUnclaimedActualLRP(logger lager.Logger, key *models.ActualLRPKey) (after *models.FlattenedActualLRP, err error)
	UnclaimActualLRP(logger lager.Logger, key *models.ActualLRPKey) (before *models.FlattenedActualLRP, after *models.FlattenedActualLRP, err error)
	ClaimActualLRP(logger lager.Logger, processGuid string, index int32, instanceKey *models.ActualLRPInstanceKey) (before *models.FlattenedActualLRP, after *models.FlattenedActualLRP, err error)
	StartActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, netInfo *models.ActualLRPNetInfo) (before *models.FlattenedActualLRP, after *models.FlattenedActualLRP, err error)
	CrashActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, crashReason string) (before *models.FlattenedActualLRP, after *models.FlattenedActualLRP, shouldRestart bool, err error)
	FailActualLRP(logger lager.Logger, key *models.ActualLRPKey, placementError string) (before *models.FlattenedActualLRP, after *models.FlattenedActualLRP, err error)
	RemoveActualLRP(logger lager.Logger, processGuid string, index int32, instanceKey *models.ActualLRPInstanceKey) error
}
