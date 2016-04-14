package sqldb

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

// Using a whereClause instead of a string, to make it more awkward for
// someone to use selectActualLRPs in a sql-injectable way. It's not perfect,
// but it will hopefully make someone stop and consider when using that method.
type whereClause struct {
	string
}

var (
	whereProcessGuidEquals     = whereClause{"process_guid = ?"}
	whereCellIdEquals          = whereClause{"cell_id = ?"}
	whereDomainEquals          = whereClause{"domain = ?"}
	whereInstanceIndexEquals   = whereClause{"instance_index = ?"}
	whereEvacuatingEquals      = whereClause{"evacuating = ?"}
	whereExpireTimeGreaterThan = whereClause{"expire_time > ?"}
)

func (db *SQLDB) ActualLRPGroups(logger lager.Logger, filter models.ActualLRPFilter) ([]*models.ActualLRPGroup, error) {
	logger = logger.Session("actual-lrp-groups-sqldb", lager.Data{"filter": filter})
	logger.Debug("starting")
	defer logger.Debug("complete")

	return db.selectActualLRPs(logger, db.db, map[whereClause]interface{}{
		whereDomainEquals: filter.Domain,
		whereCellIdEquals: filter.CellID,
	}, NoLock)
}

func (db *SQLDB) ActualLRPGroupsByProcessGuid(logger lager.Logger, processGuid string) ([]*models.ActualLRPGroup, error) {
	logger = logger.Session("actual-lrp-groups-by-process-guid-sqldb", lager.Data{"process_guid": processGuid})
	logger.Debug("starting")
	defer logger.Debug("complete")

	return db.selectActualLRPs(logger, db.db, map[whereClause]interface{}{
		whereProcessGuidEquals: processGuid,
	}, NoLock)
}

func (db *SQLDB) ActualLRPGroupByProcessGuidAndIndex(logger lager.Logger, processGuid string, index int32) (*models.ActualLRPGroup, error) {
	logger = logger.Session("actual-lrp-groups-by-process-guid-and-index-sqldb", lager.Data{"process_guid": processGuid, "index": index})
	logger.Debug("starting")
	defer logger.Debug("complete")

	groups, err := db.selectActualLRPs(logger, db.db, map[whereClause]interface{}{
		whereProcessGuidEquals:   processGuid,
		whereInstanceIndexEquals: index,
	}, NoLock)
	if err != nil {
		logger.Error("failed-select-query", err)
		return nil, err
	}

	if len(groups) == 0 {
		logger.Error("failed-to-find-actual-lrp-group", models.ErrResourceNotFound)
		return nil, models.ErrResourceNotFound
	}

	return groups[0], nil
}

func (db *SQLDB) CreateUnclaimedActualLRP(logger lager.Logger, key *models.ActualLRPKey) (*models.ActualLRPGroup, error) {
	logger = logger.Session("create-unclaimed-actual-lrp-sqldb", lager.Data{"key": key})
	logger.Debug("starting")
	defer logger.Debug("complete")

	guid, err := db.guidProvider.NextGUID()
	if err != nil {
		logger.Error("failed-to-generate-guid", err)
		return nil, models.ErrGUIDGeneration
	}

	now := db.clock.Now().UnixNano()
	_, err = db.db.Exec(`
		INSERT INTO actual_lrps
			(process_guid, instance_index, domain, state, since, net_info, modification_tag_epoch, modification_tag_index)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		key.ProcessGuid,
		key.Index,
		key.Domain,
		models.ActualLRPStateUnclaimed,
		now,
		[]byte{},
		guid,
		0,
	)
	if err != nil {
		logger.Error("failed-to-create-unclaimed-actual-lrp", err)
		return nil, db.convertSQLError(err)
	}
	return &models.ActualLRPGroup{
		Instance: &models.ActualLRP{
			ActualLRPKey:    *key,
			State:           models.ActualLRPStateUnclaimed,
			Since:           now,
			ModificationTag: models.ModificationTag{Epoch: guid, Index: 0},
		},
	}, nil
}

func (db *SQLDB) UnclaimActualLRP(logger lager.Logger, key *models.ActualLRPKey) (*models.ActualLRPGroup, *models.ActualLRPGroup, error) {
	logger = logger.Session("unclaim-actual-lrp-sqldb", lager.Data{"key": key})
	logger.Debug("starting")
	defer logger.Debug("complete")

	var beforeActualLRP models.ActualLRP
	var actualLRP *models.ActualLRP
	processGuid := key.ProcessGuid
	index := key.Index

	err := db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		var err error
		actualLRP, err = db.fetchActualLRPForShare(logger, processGuid, index, false, tx)
		if err != nil {
			logger.Error("failed-fetching-actual-lrp-for-share", err)
			return err
		}
		beforeActualLRP = *actualLRP

		if actualLRP.State == models.ActualLRPStateUnclaimed {
			logger.Debug("already-unclaimed")
			return models.ErrActualLRPCannotBeUnclaimed
		}

		now := db.clock.Now().UnixNano()
		actualLRP.ModificationTag.Increment()
		actualLRP.State = models.ActualLRPStateUnclaimed
		actualLRP.ActualLRPInstanceKey.CellId = ""
		actualLRP.ActualLRPInstanceKey.InstanceGuid = ""
		actualLRP.Since = now
		actualLRP.ActualLRPNetInfo = models.ActualLRPNetInfo{}

		_, err = tx.Exec(`
				UPDATE actual_lrps
				SET state = ?, instance_guid = ?, cell_id = ?,
					modification_tag_index = ?, since = ?, net_info = ?
				WHERE process_guid = ? AND instance_index = ? AND evacuating = ?`,
			actualLRP.State,
			actualLRP.ActualLRPInstanceKey.CellId,
			actualLRP.ActualLRPInstanceKey.InstanceGuid,
			actualLRP.ModificationTag.Index,
			db.clock.Now().UnixNano(),
			[]byte{},
			processGuid,
			index,
			false,
		)
		if err != nil {
			logger.Error("failed-to-unclaim-actual-lrp", err)
			return db.convertSQLError(err)
		}

		return nil
	})

	return &models.ActualLRPGroup{Instance: &beforeActualLRP}, &models.ActualLRPGroup{Instance: actualLRP}, err
}

func (db *SQLDB) ClaimActualLRP(logger lager.Logger, processGuid string, index int32, instanceKey *models.ActualLRPInstanceKey) (*models.ActualLRPGroup, *models.ActualLRPGroup, error) {
	logger = logger.Session("claim-actual-lrp-sqldb", lager.Data{"process_guid": processGuid, "index": index, "instance_key": instanceKey})
	logger.Debug("starting")
	defer logger.Debug("complete")

	var beforeActualLRP models.ActualLRP
	var actualLRP *models.ActualLRP
	err := db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		var err error
		actualLRP, err = db.fetchActualLRPForShare(logger, processGuid, index, false, tx)
		if err != nil {
			logger.Error("failed-fetching-actual-lrp-for-share", err)
			return err
		}
		beforeActualLRP = *actualLRP

		if !actualLRP.AllowsTransitionTo(&actualLRP.ActualLRPKey, instanceKey, models.ActualLRPStateClaimed) {
			logger.Error("cannot-transition-to-claimed", nil, lager.Data{"from_state": actualLRP.State, "same_instance_key": actualLRP.ActualLRPInstanceKey.Equal(instanceKey)})
			return models.ErrActualLRPCannotBeClaimed
		}

		if actualLRP.State == models.ActualLRPStateClaimed && actualLRP.ActualLRPInstanceKey.Equal(instanceKey) {
			return nil
		}

		actualLRP.ModificationTag.Increment()
		actualLRP.State = models.ActualLRPStateClaimed
		actualLRP.ActualLRPInstanceKey = *instanceKey
		actualLRP.PlacementError = ""
		actualLRP.ActualLRPNetInfo = models.ActualLRPNetInfo{}
		actualLRP.Since = db.clock.Now().UnixNano()

		_, err = tx.Exec(`
				UPDATE actual_lrps
				SET state = ?, instance_guid = ?, cell_id = ?, placement_error = ?,
					modification_tag_index = ?, net_info = ?, since = ?
				WHERE process_guid = ? AND instance_index = ? AND evacuating = ?`,
			actualLRP.State,
			actualLRP.ActualLRPInstanceKey.InstanceGuid,
			actualLRP.ActualLRPInstanceKey.CellId,
			actualLRP.PlacementError,
			actualLRP.ModificationTag.Index,
			[]byte{},
			actualLRP.Since,
			processGuid, index, false,
		)
		if err != nil {
			logger.Error("failed-claiming-actual-lrp", err)
			return db.convertSQLError(err)
		}

		return nil
	})

	return &models.ActualLRPGroup{Instance: &beforeActualLRP}, &models.ActualLRPGroup{Instance: actualLRP}, err
}

func (db *SQLDB) StartActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, netInfo *models.ActualLRPNetInfo) (*models.ActualLRPGroup, *models.ActualLRPGroup, error) {
	logger = logger.Session("start-actual-lrp", lager.Data{"actual_lrp_key": key, "actual_lrp_instance_key": instanceKey, "net_info": netInfo})
	logger.Debug("starting")
	defer logger.Debug("completed")

	var beforeActualLRP models.ActualLRP
	var actualLRP *models.ActualLRP

	err := db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		var err error
		actualLRP, err = db.fetchActualLRPForShare(logger, key.ProcessGuid, key.Index, false, tx)
		if err == models.ErrResourceNotFound {
			actualLRP, err = db.createRunningActualLRP(logger, key, instanceKey, netInfo, tx)
			return err
		}

		if err != nil {
			logger.Error("failed-to-get-actual-lrp", err)
			return err
		}

		beforeActualLRP = *actualLRP

		if actualLRP.ActualLRPKey.Equal(key) &&
			actualLRP.ActualLRPInstanceKey.Equal(instanceKey) &&
			actualLRP.ActualLRPNetInfo.Equal(netInfo) &&
			actualLRP.State == models.ActualLRPStateRunning {
			logger.Info("nothing-to-change")
			return nil
		}

		if !actualLRP.AllowsTransitionTo(key, instanceKey, models.ActualLRPStateRunning) {
			logger.Error("failed-to-transition-actual-lrp-to-started", nil)
			return models.ErrActualLRPCannotBeStarted
		}

		now := db.clock.Now().UnixNano()
		evacuating := false

		actualLRP.ActualLRPInstanceKey = *instanceKey
		actualLRP.ActualLRPNetInfo = *netInfo
		actualLRP.State = models.ActualLRPStateRunning
		actualLRP.Since = now
		actualLRP.ModificationTag.Increment()
		actualLRP.PlacementError = ""

		netInfoData, err := db.serializeModel(logger, &actualLRP.ActualLRPNetInfo)
		if err != nil {
			logger.Error("failed-to-serialize-net-info", err)
			return err
		}

		_, err = tx.Exec(`
					UPDATE actual_lrps SET instance_guid = ?, cell_id = ?, net_info = ?,
					state = ?, since = ?, modification_tag_index = ?, placement_error = ?
					WHERE process_guid = ? AND instance_index = ? AND evacuating = ?
				`,
			actualLRP.InstanceGuid,
			actualLRP.CellId,
			netInfoData,
			actualLRP.State,
			now,
			actualLRP.ModificationTag.Index,
			actualLRP.PlacementError,
			key.ProcessGuid,
			key.Index,
			evacuating,
		)
		if err != nil {
			logger.Error("failed-starting-actual-lrp", err)
			return db.convertSQLError(err)
		}

		return nil
	})

	return &models.ActualLRPGroup{Instance: &beforeActualLRP}, &models.ActualLRPGroup{Instance: actualLRP}, err
}

func (db *SQLDB) CrashActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, crashReason string) (*models.ActualLRPGroup, *models.ActualLRPGroup, bool, error) {
	logger = logger.Session("crash-actual-lrp-sqldb", lager.Data{"key": key, "instanceKey": instanceKey, "crash_reason": crashReason})
	logger.Debug("starting")
	defer logger.Debug("complete")

	var immediateRestart = false
	var beforeActualLRP models.ActualLRP
	var actualLRP *models.ActualLRP

	err := db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		var err error
		actualLRP, err = db.fetchActualLRPForShare(logger, key.ProcessGuid, key.Index, false, tx)
		if err != nil {
			logger.Error("failed-to-get-actual-lrp", err)
			return err
		}
		beforeActualLRP = *actualLRP

		latestChangeTime := time.Duration(db.clock.Now().UnixNano() - actualLRP.Since)

		var newCrashCount int32
		if latestChangeTime > models.CrashResetTimeout && actualLRP.State == models.ActualLRPStateRunning {
			newCrashCount = 1
		} else {
			newCrashCount = actualLRP.CrashCount + 1
		}

		if !actualLRP.AllowsTransitionTo(&actualLRP.ActualLRPKey, instanceKey, models.ActualLRPStateCrashed) {
			logger.Error("failed-to-transition-to-crashed", nil, lager.Data{"from_state": actualLRP.State, "same_instance_key": actualLRP.ActualLRPInstanceKey.Equal(instanceKey)})
			return models.ErrActualLRPCannotBeCrashed
		}

		actualLRP.ModificationTag.Increment()
		actualLRP.State = models.ActualLRPStateCrashed

		actualLRP.ActualLRPInstanceKey.InstanceGuid = ""
		actualLRP.ActualLRPInstanceKey.CellId = ""
		actualLRP.ActualLRPNetInfo = models.ActualLRPNetInfo{}
		actualLRP.CrashCount = newCrashCount
		actualLRP.CrashReason = crashReason
		evacuating := false

		if actualLRP.ShouldRestartImmediately(models.NewDefaultRestartCalculator()) {
			actualLRP.State = models.ActualLRPStateUnclaimed
			immediateRestart = true
		}

		now := db.clock.Now().UnixNano()
		actualLRP.Since = now

		_, err = tx.Exec(`
				UPDATE actual_lrps
				SET state = ?, instance_guid = ?, cell_id = ?,
					modification_tag_index = ?, since = ?, net_info = ?,
					crash_count = ?, crash_reason = ?
				WHERE process_guid = ? AND instance_index = ? AND evacuating = ?`,
			actualLRP.State,
			actualLRP.ActualLRPInstanceKey.InstanceGuid,
			actualLRP.ActualLRPInstanceKey.CellId,
			actualLRP.ModificationTag.Index,
			actualLRP.Since,
			[]byte{},
			actualLRP.CrashCount,
			actualLRP.CrashReason,
			key.ProcessGuid, key.Index, evacuating,
		)
		if err != nil {
			logger.Error("failed-to-crash-actual-lrp", err)
			return db.convertSQLError(err)
		}

		return nil
	})

	return &models.ActualLRPGroup{Instance: &beforeActualLRP}, &models.ActualLRPGroup{Instance: actualLRP}, immediateRestart, err
}

func (db *SQLDB) FailActualLRP(logger lager.Logger, key *models.ActualLRPKey, placementError string) (*models.ActualLRPGroup, *models.ActualLRPGroup, error) {
	logger = logger.Session("fail-actual-lrp", lager.Data{"actual_lrp_key": key, "placement_error": placementError})
	logger.Debug("starting")
	defer logger.Debug("complete")

	var beforeActualLRP models.ActualLRP
	var actualLRP *models.ActualLRP

	err := db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		var err error
		actualLRP, err = db.fetchActualLRPForShare(logger, key.ProcessGuid, key.Index, false, tx)
		if err != nil {
			logger.Error("failed-to-get-actual-lrp", err)
			return err
		}
		beforeActualLRP = *actualLRP

		if actualLRP.State != models.ActualLRPStateUnclaimed {
			logger.Error("cannot-fail-actual-lrp", nil, lager.Data{"from_state": actualLRP.State})
			return models.ErrActualLRPCannotBeFailed
		}

		now := db.clock.Now().UnixNano()
		actualLRP.ModificationTag.Increment()
		actualLRP.PlacementError = placementError
		actualLRP.Since = now
		evacuating := false

		_, err = tx.Exec(`
					UPDATE actual_lrps SET since = ?, modification_tag_index = ?, placement_error = ?
					WHERE process_guid = ? AND instance_index = ? AND evacuating = ?
				`,
			now,
			actualLRP.ModificationTag.Index,
			actualLRP.PlacementError,
			actualLRP.ActualLRPKey.ProcessGuid,
			actualLRP.ActualLRPKey.Index,
			evacuating,
		)
		if err != nil {
			logger.Error("failed-failing-actual-lrp", err)
			return db.convertSQLError(err)
		}

		return nil
	})

	return &models.ActualLRPGroup{Instance: &beforeActualLRP}, &models.ActualLRPGroup{Instance: actualLRP}, err
}

func (db *SQLDB) RemoveActualLRP(logger lager.Logger, processGuid string, index int32) error {
	logger = logger.Session("remove-actual-lrp", lager.Data{"process_guid": processGuid, "index": index})
	logger.Debug("starting")
	defer logger.Debug("complete")

	return db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		result, err := tx.Exec(`
					DELETE FROM actual_lrps
					WHERE process_guid = ? AND instance_index = ? AND evacuating = ?
				`,
			processGuid, index, false,
		)
		if err != nil {
			logger.Error("failed-removing-actual-lrp", err)
			return db.convertSQLError(err)
		}

		numRows, err := result.RowsAffected()
		if err != nil {
			logger.Error("failed-getting-rows-affected", err)
			return err
		}
		if numRows == 0 {
			logger.Debug("not-found")
			return models.ErrResourceNotFound
		}

		return nil
	})
}

func (db *SQLDB) createRunningActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, netInfo *models.ActualLRPNetInfo, tx *sql.Tx) (*models.ActualLRP, error) {
	now := db.clock.Now().UnixNano()
	guid, err := db.guidProvider.NextGUID()
	if err != nil {
		return nil, models.ErrGUIDGeneration
	}

	actualLRP := &models.ActualLRP{}
	actualLRP.ModificationTag = models.NewModificationTag(guid, 0)
	actualLRP.ActualLRPKey = *key
	actualLRP.ActualLRPInstanceKey = *instanceKey
	actualLRP.ActualLRPNetInfo = *netInfo
	actualLRP.State = models.ActualLRPStateRunning
	actualLRP.Since = now

	netInfoData, err := db.serializeModel(logger, &actualLRP.ActualLRPNetInfo)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(`
				INSERT INTO actual_lrps
					(process_guid, instance_index, domain, instance_guid, cell_id, state, net_info, since, modification_tag_epoch, modification_tag_index)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		actualLRP.ActualLRPKey.ProcessGuid,
		actualLRP.ActualLRPKey.Index,
		actualLRP.ActualLRPKey.Domain,
		actualLRP.ActualLRPInstanceKey.InstanceGuid,
		actualLRP.ActualLRPInstanceKey.CellId,
		actualLRP.State,
		netInfoData,
		actualLRP.Since,
		actualLRP.ModificationTag.Epoch,
		actualLRP.ModificationTag.Index,
	)
	if err != nil {
		logger.Error("failed-creating-running-actual-lrp", err)
		return nil, db.convertSQLError(err)
	}
	return actualLRP, nil
}

func (db *SQLDB) scanToActualLRP(logger lager.Logger, row RowScanner) (*models.ActualLRP, bool, error) {
	var netInfoData []byte
	var actualLRP models.ActualLRP
	var evacuating bool

	err := row.Scan(
		&actualLRP.ProcessGuid,
		&actualLRP.Index,
		&evacuating,
		&actualLRP.Domain,
		&actualLRP.State,
		&actualLRP.InstanceGuid,
		&actualLRP.CellId,
		&actualLRP.PlacementError,
		&actualLRP.Since,
		&netInfoData,
		&actualLRP.ModificationTag.Epoch,
		&actualLRP.ModificationTag.Index,
		&actualLRP.CrashCount,
		&actualLRP.CrashReason,
	)
	if err != nil {
		logger.Error("failed-scanning-actual-lrp", err)
		return nil, false, db.convertSQLError(err)
	}

	if len(netInfoData) > 0 {
		err = db.deserializeModel(logger, netInfoData, &actualLRP.ActualLRPNetInfo)
		if err != nil {
			logger.Error("failed-unmarshaling-net-info-data", err)
			return &actualLRP, evacuating, models.ErrDeserialize
		}
	}

	return &actualLRP, evacuating, nil
}

func (db *SQLDB) selectActualLRPs(logger lager.Logger, q Queryable, conditions map[whereClause]interface{}, lockMode int) ([]*models.ActualLRPGroup, error) {
	wheres := []string{}
	values := []interface{}{}
	for field, value := range conditions {
		if value == "" {
			continue
		}
		wheres = append(wheres, field.string)
		values = append(values, value)
	}

	query := `
		SELECT process_guid, instance_index, evacuating, domain, state,
			instance_guid, cell_id, placement_error, since, net_info,
			modification_tag_epoch, modification_tag_index, crash_count,
			crash_reason
		FROM actual_lrps
	`
	if len(wheres) > 0 {
		query += fmt.Sprintf("WHERE %s\n", strings.Join(wheres, " AND "))
	}
	switch lockMode {
	case LockForShare:
		query += "LOCK IN SHARE MODE\n"
	case LockForUpdate:
		query += "FOR UPDATE\n"
	}

	rows, err := q.Query(query, values...)
	if err != nil {
		logger.Error("failed-fetching-actual-lrps", err)
		return nil, db.convertSQLError(err)
	}
	defer rows.Close()

	mapOfGroups := map[models.ActualLRPKey]*models.ActualLRPGroup{}
	result := []*models.ActualLRPGroup{}
	for rows.Next() {
		actualLRP, evacuating, err := db.scanToActualLRP(logger, rows)
		if err == models.ErrDeserialize && lockMode != LockForShare {
			_, err := q.Exec(`
				DELETE FROM actual_lrps
				WHERE process_guid = ? AND instance_index = ? AND evacuating = ?
				`, actualLRP.ProcessGuid, actualLRP.Index, evacuating)
			if err != nil {
				logger.Error("failed-cleaning-up-invalid-actual-lrp", err)
			}
			continue
		}

		if err != nil {
			logger.Error("failed-scanning-actual-lrp", err)
			return nil, err
		}

		// Every actual LRP has potentially 2 rows in the database: one for the instance
		// one for the evacuating.  When building the list of actual LRP groups (where
		// a group is the instance and corresponding evacuating), make sure we don't add the same
		// actual lrp twice.
		if mapOfGroups[actualLRP.ActualLRPKey] == nil {
			mapOfGroups[actualLRP.ActualLRPKey] = &models.ActualLRPGroup{}
			result = append(result, mapOfGroups[actualLRP.ActualLRPKey])
		}
		if evacuating {
			mapOfGroups[actualLRP.ActualLRPKey].Evacuating = actualLRP
		} else {
			mapOfGroups[actualLRP.ActualLRPKey].Instance = actualLRP
		}
	}

	if rows.Err() != nil {
		logger.Error("failed-getting-next-row", rows.Err())
		return nil, db.convertSQLError(rows.Err())
	}

	return result, nil
}

func (db *SQLDB) fetchActualLRPForShare(logger lager.Logger, processGuid string, index int32, evacuating bool, tx *sql.Tx) (*models.ActualLRP, error) {
	expireTime := db.clock.Now().Round(time.Second).UnixNano()
	conditions := map[whereClause]interface{}{
		whereProcessGuidEquals:   processGuid,
		whereInstanceIndexEquals: index,
		whereEvacuatingEquals:    evacuating,
	}

	if evacuating {
		conditions[whereExpireTimeGreaterThan] = expireTime
	}

	groups, err := db.selectActualLRPs(logger, tx, conditions, LockForShare)
	if err != nil {
		return nil, err
	}

	if len(groups) == 0 {
		return nil, models.ErrResourceNotFound
	}

	actualLRP, _ := groups[0].Resolve()

	return actualLRP, nil
}
