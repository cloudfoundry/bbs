package db

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter . EvacuationDB

type EvacuationDB interface {
	EvacuateClaimedActualLRP(lager.Logger, *models.ActualLRPKey, *models.ActualLRPInstanceKey) (bool, error)
	EvacuateRunningActualLRP(lager.Logger, *models.ActualLRPKey, *models.ActualLRPInstanceKey, *models.ActualLRPNetInfo, uint64) (bool, error)
	EvacuateStoppedActualLRP(lager.Logger, *models.ActualLRPKey, *models.ActualLRPInstanceKey) (bool, error)
	EvacuateCrashedActualLRP(lager.Logger, *models.ActualLRPKey, *models.ActualLRPInstanceKey, string) (bool, error)
	RemoveEvacuatingActualLRP(lager.Logger, *models.ActualLRPKey, *models.ActualLRPInstanceKey) error
}
