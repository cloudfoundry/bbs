package sqldb

import (
	"database/sql"
	"reflect"
	"time"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

func (db *SQLDB) EvacuateActualLRP(
	logger lager.Logger,
	lrpKey *models.ActualLRPKey,
	instanceKey *models.ActualLRPInstanceKey,
	netInfo *models.ActualLRPNetInfo,
	ttl uint64,
) (*models.ActualLRPGroup, error) {
	logger = logger.Session("evacuate-lrp-sqldb", lager.Data{"lrp_key": lrpKey, "instance_key": instanceKey, "net_info": netInfo})
	logger.Debug("starting")
	defer logger.Debug("complete")

	var actualLRP *models.ActualLRP

	err := db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		var err error
		processGuid := lrpKey.ProcessGuid
		index := lrpKey.Index

		actualLRP, err = db.fetchActualLRPForUpdate(logger, processGuid, index, true, tx)
		if err == models.ErrResourceNotFound {
			logger.Debug("creating-evacuating-lrp")
			actualLRP, err = db.createEvacuatingActualLRP(logger, lrpKey, instanceKey, netInfo, ttl, tx)
			return err
		}

		if err != nil {
			logger.Error("failed-locking-lrp", err)
			return err
		}

		if actualLRP.ActualLRPKey.Equal(lrpKey) &&
			actualLRP.ActualLRPInstanceKey.Equal(instanceKey) &&
			reflect.DeepEqual(actualLRP.ActualLRPNetInfo, *netInfo) {
			logger.Debug("evacuating-lrp-already-exists")
			return nil
		}

		now := db.clock.Now().UnixNano()
		actualLRP.ModificationTag.Increment()
		actualLRP.ActualLRPKey = *lrpKey
		actualLRP.ActualLRPInstanceKey = *instanceKey
		actualLRP.Since = now
		actualLRP.ActualLRPNetInfo = *netInfo

		netInfoData, err := db.serializeModel(logger, netInfo)
		if err != nil {
			logger.Error("failed-serializing-net-info", err)
			return err
		}

		_, err = tx.Exec(`
					UPDATE actual_lrps SET domain = ?, instance_guid = ?, cell_id = ?, net_info = ?,
					  state = ?, since = ?, modification_tag_index = ?
					  WHERE process_guid = ? AND instance_index = ? AND evacuating = ?
				`,
			actualLRP.Domain,
			actualLRP.InstanceGuid,
			actualLRP.CellId,
			netInfoData,
			actualLRP.State,
			actualLRP.Since,
			actualLRP.ModificationTag.Index,
			actualLRP.ProcessGuid,
			actualLRP.Index,
			true,
		)

		if err != nil {
			logger.Error("failed-update-evacuating-lrp", err)
			return db.convertSQLError(err)
		}

		return nil
	})

	return &models.ActualLRPGroup{Evacuating: actualLRP}, err
}

func (db *SQLDB) RemoveEvacuatingActualLRP(logger lager.Logger, lrpKey *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey) error {
	logger = logger.Session("remove-evacuating-lrp-sqldb", lager.Data{"lrp_key": lrpKey, "instance_key": instanceKey})
	logger.Debug("starting")
	defer logger.Debug("complete")

	return db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		processGuid := lrpKey.ProcessGuid
		index := lrpKey.Index

		lrp, err := db.fetchActualLRPForUpdate(logger, processGuid, index, true, tx)
		if err == models.ErrResourceNotFound {
			logger.Debug("evacuating-lrp-does-not-exist")
			return nil
		}

		if err != nil {
			logger.Error("failed-fetching-actual-lrp", err)
			return err
		}

		if !lrp.ActualLRPInstanceKey.Equal(instanceKey) {
			logger.Debug("actual-lrp-instance-key-mismatch", lager.Data{"instance-key-param": instanceKey, "instance-key-from-db": lrp.ActualLRPInstanceKey})
			return models.ErrActualLRPCannotBeRemoved
		}

		_, err = tx.Exec(`
				DELETE FROM actual_lrps
					WHERE process_guid = ? AND instance_index = ? AND evacuating = ?
			`,
			processGuid, index, true,
		)

		if err != nil {
			logger.Error("failed-delete", err)
			return models.ErrActualLRPCannotBeRemoved
		}

		return nil
	})
}

func (db *SQLDB) createEvacuatingActualLRP(logger lager.Logger,
	lrpKey *models.ActualLRPKey,
	instanceKey *models.ActualLRPInstanceKey,
	netInfo *models.ActualLRPNetInfo,
	ttl uint64,
	tx *sql.Tx) (*models.ActualLRP, error) {
	netInfoData, err := db.serializeModel(logger, netInfo)
	if err != nil {
		logger.Error("failed-serializing-net-info", err)
		return nil, err
	}

	now := db.clock.Now()
	guid, err := db.guidProvider.NextGUID()
	if err != nil {
		return nil, models.ErrGUIDGeneration
	}

	expireTime := now.Add(time.Duration(ttl) * time.Second)
	actualLRP := &models.ActualLRP{
		ActualLRPKey:         *lrpKey,
		ActualLRPInstanceKey: *instanceKey,
		ActualLRPNetInfo:     *netInfo,
		State:                models.ActualLRPStateRunning,
		Since:                now.UnixNano(),
		ModificationTag:      models.ModificationTag{Epoch: guid, Index: 0},
	}

	_, err = tx.Exec(`
					INSERT INTO actual_lrps
						(process_guid, instance_index, domain, instance_guid, cell_id, state, net_info, since,
						  modification_tag_epoch, modification_tag_index, evacuating)
						VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
						ON DUPLICATE KEY UPDATE expire_time = ?, domain = VALUES(domain),
						instance_guid = VALUES(instance_guid), cell_id = VALUES(cell_id),
						state = VALUES(state), net_info = VALUES(net_info), since = VALUES(since),
						modification_tag_epoch = VALUES(modification_tag_epoch),
						modification_tag_index = VALUES(modification_tag_index)
						`,
		actualLRP.ProcessGuid,
		actualLRP.Index,
		actualLRP.Domain,
		actualLRP.InstanceGuid,
		actualLRP.CellId,
		actualLRP.State,
		netInfoData,
		actualLRP.Since,
		actualLRP.ModificationTag.Epoch,
		actualLRP.ModificationTag.Index,
		true,
		expireTime.UnixNano(),
	)

	if err != nil {
		logger.Error("failed-insert-evacuating-lrp", err)
		return nil, db.convertSQLError(err)
	}

	return actualLRP, nil
}
