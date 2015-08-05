package db

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter . EvacuationDB

type EvacuationDB interface {
	EvacuateClaimedActualLRP(lager.Logger, *models.EvacuateClaimedActualLRPRequest) (models.ContainerRetainment, *models.Error)
	EvacuateRunningActualLRP(lager.Logger, *models.EvacuateRunningActualLRPRequest) (models.ContainerRetainment, *models.Error)
	EvacuateStoppedActualLRP(lager.Logger, *models.EvacuateStoppedActualLRPRequest) (models.ContainerRetainment, *models.Error)
	EvacuateCrashedActualLRP(lager.Logger, *models.EvacuateCrashedActualLRPRequest) (models.ContainerRetainment, *models.Error)
	RemoveEvacuatingActualLRP(lager.Logger, *models.RemoveEvacuatingActualLRPRequest) *models.Error
}
