package db

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter . EvacuationDB

type EvacuationDB interface {
	EvacuateClaimedActualLRP(lager.Logger, *models.EvacuateClaimedActualLRPRequest) (bool, *models.Error)
	EvacuateRunningActualLRP(lager.Logger, *models.EvacuateRunningActualLRPRequest) (bool, *models.Error)
	EvacuateStoppedActualLRP(lager.Logger, *models.EvacuateStoppedActualLRPRequest) (bool, *models.Error)
	EvacuateCrashedActualLRP(lager.Logger, *models.EvacuateCrashedActualLRPRequest) (bool, *models.Error)
	RemoveEvacuatingActualLRP(lager.Logger, *models.RemoveEvacuatingActualLRPRequest) *models.Error
}
