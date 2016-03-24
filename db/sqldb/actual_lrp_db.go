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
	return db.selectActualLRPs(logger, db.db, map[whereClause]interface{}{
		whereDomainEquals: filter.Domain,
		whereCellIdEquals: filter.CellID,
	}, NoLock)
}

func (db *SQLDB) ActualLRPGroupsByProcessGuid(logger lager.Logger, processGuid string) ([]*models.ActualLRPGroup, error) {
	return db.selectActualLRPs(logger, db.db, map[whereClause]interface{}{
		whereProcessGuidEquals: processGuid,
	}, NoLock)
}

func (db *SQLDB) ActualLRPGroupByProcessGuidAndIndex(logger lager.Logger, processGuid string, index int32) (*models.ActualLRPGroup, error) {
	groups, err := db.selectActualLRPs(logger, db.db, map[whereClause]interface{}{
		whereProcessGuidEquals:   processGuid,
		whereInstanceIndexEquals: index,
	}, NoLock)
	if err != nil {
		return nil, err
	}

	if len(groups) == 0 {
		logger.Error("failed-to-find-actual-lrp-group", models.ErrResourceNotFound)
		return nil, models.ErrResourceNotFound
	}

	return groups[0], nil
}

func (db *SQLDB) CreateUnclaimedActualLRP(logger lager.Logger, key *models.ActualLRPKey) error {
	guid, err := db.guidProvider.NextGUID()
	if err != nil {
		return models.ErrGUIDGeneration
	}

	now := db.clock.Now()
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
		return db.convertSQLError(err)
	}
	return nil
}

func (db *SQLDB) UnclaimActualLRP(logger lager.Logger, actualLRPKey *models.ActualLRPKey) error {
	processGuid := actualLRPKey.ProcessGuid
	index := actualLRPKey.Index
	err := db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		actualLRP, err := db.fetchActualLRPForShare(logger, processGuid, index, false, tx)
		if err != nil {
			return err
		}

		if actualLRP.State == models.ActualLRPStateUnclaimed {
			logger.Debug("already-unclaimed")
			return models.ErrActualLRPCannotBeUnclaimed
		}
		actualLRP.ModificationTag.Increment()

		_, err = tx.Exec(`
				UPDATE actual_lrps
				SET state = ?, instance_guid = ?, cell_id = ?,
					modification_tag_index = ?, since = ?, net_info = ?
				WHERE process_guid = ? AND instance_index = ? AND evacuating = ?`,
			models.ActualLRPStateUnclaimed,
			"",
			"",
			actualLRP.ModificationTag.Index,
			db.clock.Now(),
			[]byte{},
			processGuid, index, false,
		)
		if err != nil {
			logger.Error("failed-to-unclaim-actual-lrp", err)
			return db.convertSQLError(err)
		}

		return nil
	})

	return err
}

func (db *SQLDB) ClaimActualLRP(logger lager.Logger, processGuid string, index int32, instanceKey *models.ActualLRPInstanceKey) error {
	return db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		actualLRP, err := db.fetchActualLRPForShare(logger, processGuid, index, false, tx)
		if err != nil {
			logger.Error("failed-fetching-actual-lrp-for-share", err)
			return err
		}

		if !actualLRP.AllowsTransitionTo(&actualLRP.ActualLRPKey, instanceKey, models.ActualLRPStateClaimed) {
			return models.ErrActualLRPCannotBeClaimed
		}
		actualLRP.ModificationTag.Increment()

		_, err = tx.Exec(`
				UPDATE actual_lrps
				SET state = ?, instance_guid = ?, cell_id = ?, placement_error = ?,
					modification_tag_index = ?, net_info = ?
				WHERE process_guid = ? AND instance_index = ? AND evacuating = ?`,
			models.ActualLRPStateClaimed,
			instanceKey.InstanceGuid,
			instanceKey.CellId,
			"",
			actualLRP.ModificationTag.Index,
			[]byte{},
			processGuid, index, false,
		)
		if err != nil {
			logger.Error("failed-claiming-actual-lrp", err)
			return db.convertSQLError(err)
		}

		return nil
	})
}

func (db *SQLDB) StartActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, netInfo *models.ActualLRPNetInfo) error {
	return db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		actualLRP, err := db.fetchActualLRPForShare(logger, key.ProcessGuid, key.Index, false, tx)
		logger = logger.Session("start-actual-lrp", lager.Data{"actual_lrp_key": key, "actual_lrp_instance_key": instanceKey, "net_info": netInfo})
		logger.Info("starting")

		if err == models.ErrResourceNotFound {
			return db.createRunningActualLRP(logger, key, instanceKey, netInfo, tx)
		}

		if err != nil {
			logger.Error("failed-to-get-actual-lrp", err)
			return err
		}

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

		netInfoData, err := db.serializer.Marshal(logger, db.format, netInfo)
		if err != nil {
			return err
		}

		now := db.clock.Now()
		actualLRP.ModificationTag.Increment()
		placementError := ""
		evacuating := false

		_, err = tx.Exec(`
					UPDATE actual_lrps SET instance_guid = ?, cell_id = ?, net_info = ?,
					state = ?, since = ?, modification_tag_index = ?, placement_error = ?
					WHERE process_guid = ? AND instance_index = ? AND evacuating = ?
				`,
			instanceKey.InstanceGuid, instanceKey.CellId, netInfoData,
			models.ActualLRPStateRunning, now, actualLRP.ModificationTag.Index,
			placementError, key.ProcessGuid, key.Index, evacuating,
		)
		if err != nil {
			logger.Error("failed-starting-actual-lrp", err)
			return db.convertSQLError(err)
		}

		logger.Info("succeeded")
		return nil
	})
}

func (db *SQLDB) createRunningActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, netInfo *models.ActualLRPNetInfo, tx *sql.Tx) error {
	netInfoData, err := db.serializer.Marshal(logger, db.format, netInfo)
	if err != nil {
		return err
	}

	now := db.clock.Now()
	guid, err := db.guidProvider.NextGUID()
	if err != nil {
		return models.ErrGUIDGeneration
	}

	_, err = tx.Exec(`
				INSERT INTO actual_lrps
					(process_guid, instance_index, domain, instance_guid, cell_id, state, net_info, since, modification_tag_epoch, modification_tag_index)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		key.ProcessGuid,
		key.Index,
		key.Domain,
		instanceKey.InstanceGuid,
		instanceKey.CellId,
		models.ActualLRPStateRunning,
		netInfoData,
		now,
		guid,
		0,
	)
	if err != nil {
		logger.Error("failed-creating-running-actual-lrp", err)
		return db.convertSQLError(err)
	}
	return nil
}

func (db *SQLDB) CrashActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, crashReason string) (bool, error) {
	var immediateRestart = false

	err := db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		logger = logger.Session("crash-actual-lrp", lager.Data{"actual_lrp_key": key, "actual_lrp_instance_key": instanceKey})
		logger.Info("starting")

		actualLRP, err := db.fetchActualLRPForShare(logger, key.ProcessGuid, key.Index, false, tx)
		if err != nil {
			logger.Error("failed-to-get-actual-lrp", err)
			return err
		}

		latestChangeTime := time.Duration(db.clock.Now().UnixNano() - actualLRP.Since)

		var newCrashCount int32
		if latestChangeTime > models.CrashResetTimeout && actualLRP.State == models.ActualLRPStateRunning {
			newCrashCount = 1
		} else {
			newCrashCount = actualLRP.CrashCount + 1
		}

		if !actualLRP.AllowsTransitionTo(&actualLRP.ActualLRPKey, instanceKey, models.ActualLRPStateCrashed) {
			logger.Error("failed-to-transition-to-crashed", nil, lager.Data{"from-state": actualLRP.State, "same-instance-key": actualLRP.ActualLRPInstanceKey.Equal(instanceKey)})
			return models.ErrActualLRPCannotBeCrashed
		}

		actualLRP.ModificationTag.Increment()
		actualLRP.State = models.ActualLRPStateCrashed

		if actualLRP.ShouldRestartImmediately(models.NewDefaultRestartCalculator()) {
			actualLRP.State = models.ActualLRPStateUnclaimed
			immediateRestart = true
		}

		instanceGuid := ""
		cellID := ""
		evacuating := false

		_, err = tx.Exec(`
				UPDATE actual_lrps
				SET state = ?, instance_guid = ?, cell_id = ?,
					modification_tag_index = ?, since = ?, net_info = ?,
					crash_count = ?, crash_reason = ?
				WHERE process_guid = ? AND instance_index = ? AND evacuating = ?`,
			actualLRP.State,
			instanceGuid,
			cellID,
			actualLRP.ModificationTag.Index,
			db.clock.Now(),
			[]byte{},
			newCrashCount, crashReason,
			key.ProcessGuid, key.Index, evacuating,
		)
		if err != nil {
			logger.Error("failed-to-crash-actual-lrp", err)
			return db.convertSQLError(err)
		}

		logger.Info("succeeded")
		return nil
	})

	return immediateRestart, err
}

func (db *SQLDB) FailActualLRP(logger lager.Logger, key *models.ActualLRPKey, placementError string) error {
	return db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		logger = logger.Session("fail-actual-lrp", lager.Data{"actual_lrp_key": key})
		logger.Info("starting")
		actualLRP, err := db.fetchActualLRPForShare(logger, key.ProcessGuid, key.Index, false, tx)
		if err != nil {
			logger.Error("failed-to-get-actual-lrp", err)
			return err
		}

		if actualLRP.State != models.ActualLRPStateUnclaimed {
			return models.ErrActualLRPCannotBeFailed
		}

		now := db.clock.Now()
		actualLRP.ModificationTag.Increment()
		evacuating := false

		_, err = tx.Exec(`
					UPDATE actual_lrps SET since = ?, modification_tag_index = ?, placement_error = ?
					WHERE process_guid = ? AND instance_index = ? AND evacuating = ?
				`,
			now, actualLRP.ModificationTag.Index, placementError,
			key.ProcessGuid, key.Index, evacuating,
		)
		if err != nil {
			logger.Error("failed-failing-actual-lrp", err)
			return db.convertSQLError(err)
		}

		logger.Info("succeeded")
		return nil
	})
}

func (db *SQLDB) RemoveActualLRP(logger lager.Logger, processGuid string, index int32) error {
	return db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		logger = logger.Session("remove-actual-lrp", lager.Data{"process_guid": processGuid, "index": index})
		logger.Info("starting")

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
			return models.ErrResourceNotFound
		}

		logger.Info("succeeded")
		return nil
	})
}

func (db *SQLDB) scanToActualLRP(logger lager.Logger, row RowScanner) (*models.ActualLRP, bool, error) {
	var netInfoData []byte
	var since time.Time
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
		&since,
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
		err = db.serializer.Unmarshal(logger, netInfoData, &actualLRP.ActualLRPNetInfo)
		if err != nil {
			logger.Error("failed-unmarshaling-net-info-data", err)
			return nil, false, err
		}
	}

	actualLRP.Since = since.UnixNano()
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
	return result, nil
}

func (db *SQLDB) fetchActualLRPForShare(logger lager.Logger, processGuid string, index int32, evacuating bool, tx *sql.Tx) (*models.ActualLRP, error) {
	expireTime := db.clock.Now().Round(time.Second)
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
