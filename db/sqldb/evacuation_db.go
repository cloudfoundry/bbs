package sqldb

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

func (db *SQLDB) EvacuateClaimedActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey) (keepContainer bool, modelErr error) {
	return false, nil
}

func (db *SQLDB) EvacuateRunningActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, netInfo *models.ActualLRPNetInfo, ttl uint64) (keepContainer bool, modelErr error) {
	return true, nil
}

func (db *SQLDB) EvacuateStoppedActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey) (bool, error) {
	return false, nil
}

func (db *SQLDB) EvacuateCrashedActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, errorMessage string) (bool, error) {
	return false, nil
}

func (db *SQLDB) RemoveEvacuatingActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey) error {
	return nil
}
