package db

import (
	"context"

	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/routing-info/internalroutes"
)

type ActualLRPKeyWithInternalRoutes struct {
	Key                   *models.ActualLRPKey
	InstanceKey           *models.ActualLRPInstanceKey
	DesiredInternalRoutes internalroutes.InternalRoutes
}

//counterfeiter:generate . LRPDB

type ConvergenceResult struct {
	MissingLRPKeys               []*models.ActualLRPKeyWithSchedulingInfo
	UnstartedLRPKeys             []*models.ActualLRPKeyWithSchedulingInfo
	SuspectKeysWithExistingCells []*models.ActualLRPKey
	SuspectLRPKeysToRetire       []*models.ActualLRPKey
	SuspectRunningKeys           []*models.ActualLRPKey
	SuspectClaimedKeys           []*models.ActualLRPKey
	KeysToRetire                 []*models.ActualLRPKey
	KeysWithMissingCells         []*models.ActualLRPKeyWithSchedulingInfo
	KeysWithInternalRouteChanges []*ActualLRPKeyWithInternalRoutes
	MissingCellIds               []string
	Events                       []models.Event
	InstanceEvents               []models.Event
}

type LRPDB interface {
	ActualLRPDB
	DesiredLRPDB

	ConvergeLRPs(ctx context.Context, logger lager.Logger, cellSet models.CellSet) ConvergenceResult
}
