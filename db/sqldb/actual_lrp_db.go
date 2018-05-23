package sqldb

import (
	"database/sql"
	"strings"
	"time"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

func (db *SQLDB) getActualLRPS(logger lager.Logger, wheres string, whereBindinngs ...interface{}) ([]*models.ActualLRPGroup, error) {
	var groups []*models.ActualLRPGroup
	err := db.transact(logger, func(logger lager.Logger, tx helpers.Tx) error {
		rows, err := db.all(logger, tx, actualLRPsTable,
			actualLRPColumns, helpers.NoLockRow,
			wheres, whereBindinngs...,
		)
		if err != nil {
			logger.Error("failed-query", err)
			return err
		}
		defer rows.Close()
		groups, err = db.scanAndCleanupActualLRPs(logger, tx, rows)
		return err
	})

	return groups, err
}

func (db *SQLDB) getFlattenedActualLRPS(logger lager.Logger, wheres string, whereBindinngs ...interface{}) ([]*models.ActualLRP, error) {
	var alrps []*models.ActualLRP
	err := db.transact(logger, func(logger lager.Logger, tx helpers.Tx) error {
		rows, err := db.all(logger, tx, actualLRPsTable,
			actualLRPColumns, helpers.NoLockRow,
			wheres, whereBindinngs...,
		)
		if err != nil {
			logger.Error("failed-query", err)
			return err
		}
		defer rows.Close()
		alrps, err = db.scanAndCleanupFlattenedActualLRPs(logger, tx, rows)
		return err
	})

	return alrps, err
}

func (db *SQLDB) ActualLRPs(logger lager.Logger, filter models.ActualLRPFilter) ([]*models.ActualLRP, error) {
	logger = logger.WithData(lager.Data{"filter": filter})
	logger.Debug("starting")
	defer logger.Debug("complete")

	var wheres []string
	var values []interface{}

	if filter.Domain != "" {
		wheres = append(wheres, "domain = ?")
		values = append(values, filter.Domain)
	}

	if filter.CellID != "" {
		wheres = append(wheres, "cell_id = ?")
		values = append(values, filter.CellID)
	}

	if filter.ProcessGUID != nil {
		wheres = append(wheres, "process_guid = ?")
		values = append(values, *filter.ProcessGUID)
	}

	if filter.Index != nil {
		wheres = append(wheres, "instance_index = ?")
		values = append(values, *filter.Index)
	}
	return db.getFlattenedActualLRPS(logger, strings.Join(wheres, " AND "), values...)
}

func (db *SQLDB) ActualLRPGroups(logger lager.Logger, filter models.ActualLRPFilter) ([]*models.ActualLRPGroup, error) {
	logger = logger.WithData(lager.Data{"filter": filter})
	logger.Debug("starting")
	defer logger.Debug("complete")

	var wheres []string
	var values []interface{}

	if filter.Domain != "" {
		wheres = append(wheres, "domain = ?")
		values = append(values, filter.Domain)
	}

	if filter.CellID != "" {
		wheres = append(wheres, "cell_id = ?")
		values = append(values, filter.CellID)
	}
	return db.getActualLRPS(logger, strings.Join(wheres, " AND "), values...)
}

func (db *SQLDB) ActualLRPGroupsByProcessGuid(logger lager.Logger, processGuid string) ([]*models.ActualLRPGroup, error) {
	logger = logger.WithData(lager.Data{"process_guid": processGuid})
	logger.Debug("starting")
	defer logger.Debug("complete")

	return db.getActualLRPS(logger, "process_guid = ?", processGuid)
}

func (db *SQLDB) ActualLRPGroupByProcessGuidAndIndex(logger lager.Logger, processGuid string, index int32) (*models.ActualLRPGroup, error) {
	logger = logger.WithData(lager.Data{"process_guid": processGuid, "index": index})
	logger.Debug("starting")
	defer logger.Debug("complete")

	groups, err := db.getActualLRPS(logger, "process_guid = ? AND instance_index = ?", processGuid, index)
	if err != nil {
		return nil, err
	}

	if len(groups) == 0 {
		logger.Error("failed-to-find-actual-lrp-group", models.ErrResourceNotFound)
		return nil, models.ErrResourceNotFound
	}

	return groups[0], nil
}

func (db *SQLDB) CreateUnclaimedActualLRP(logger lager.Logger, key *models.ActualLRPKey) (*models.ActualLRP, error) {
	logger = logger.WithData(lager.Data{"key": key})
	logger.Info("starting")
	defer logger.Info("complete")

	guid, err := db.guidProvider.NextGUID()
	if err != nil {
		logger.Error("failed-to-generate-guid", err)
		return nil, models.ErrGUIDGeneration
	}

	netInfoData, err := db.serializeModel(logger, &models.ActualLRPNetInfo{})
	if err != nil {
		logger.Error("failed-to-serialize-net-info", err)
		return nil, err
	}
	now := db.clock.Now().UnixNano()
	err = db.transact(logger, func(logger lager.Logger, tx helpers.Tx) error {
		_, err := db.insert(logger, tx, actualLRPsTable,
			helpers.SQLAttributes{
				"process_guid":           key.ProcessGuid,
				"instance_index":         key.Index,
				"domain":                 key.Domain,
				"state":                  models.ActualLRPStateUnclaimed,
				"since":                  now,
				"net_info":               netInfoData,
				"modification_tag_epoch": guid,
				"modification_tag_index": 0,
			},
		)

		return err
	})

	if err != nil {
		logger.Error("failed-to-create-unclaimed-actual-lrp", err)
		return nil, err
	}
	return &models.ActualLRP{
		ActualLRPKey:    *key,
		State:           models.ActualLRPStateUnclaimed,
		PlacementState:  models.PlacementStateType_Normal,
		Since:           now,
		ModificationTag: models.ModificationTag{Epoch: guid, Index: 0},
	}, nil
}

func (db *SQLDB) UnclaimActualLRP(logger lager.Logger, key *models.ActualLRPKey) (*models.ActualLRP, *models.ActualLRP, error) {
	logger = logger.WithData(lager.Data{"key": key})

	var beforeActualLRP models.ActualLRP
	var actualLRP *models.ActualLRP
	processGuid := key.ProcessGuid
	index := key.Index

	err := db.transact(logger, func(logger lager.Logger, tx helpers.Tx) error {
		var err error
		actualLRP, err = db.fetchActualLRPForUpdate(logger, processGuid, index, false, false, tx)
		if err != nil {
			logger.Error("failed-fetching-actual-lrp-for-share", err)
			return err
		}
		beforeActualLRP = *actualLRP

		if actualLRP.State == models.ActualLRPStateUnclaimed {
			logger.Debug("already-unclaimed")
			return models.ErrActualLRPCannotBeUnclaimed
		}
		logger.Info("starting")
		defer logger.Info("complete")

		now := db.clock.Now().UnixNano()
		actualLRP.ModificationTag.Increment()
		actualLRP.State = models.ActualLRPStateUnclaimed
		actualLRP.PlacementState = models.PlacementStateType_Normal
		actualLRP.ActualLRPInstanceKey.CellId = ""
		actualLRP.ActualLRPInstanceKey.InstanceGuid = ""
		actualLRP.Since = now
		actualLRP.ActualLRPNetInfo = models.ActualLRPNetInfo{}
		netInfoData, err := db.serializeModel(logger, &models.ActualLRPNetInfo{})
		if err != nil {
			logger.Error("failed-to-serialize-net-info", err)
			return err
		}

		_, err = db.update(logger, tx, actualLRPsTable,
			helpers.SQLAttributes{
				"state":                  actualLRP.State,
				"cell_id":                actualLRP.CellId,
				"instance_guid":          actualLRP.InstanceGuid,
				"modification_tag_index": actualLRP.ModificationTag.Index,
				"since":                  actualLRP.Since,
				"net_info":               netInfoData,
			},
			"process_guid = ? AND instance_index = ? AND evacuating = ?",
			processGuid, index, false,
		)
		if err != nil {
			logger.Error("failed-to-unclaim-actual-lrp", err)
			return err
		}

		return nil
	})

	return &beforeActualLRP, actualLRP, err
}

func (db *SQLDB) ClaimActualLRP(logger lager.Logger, processGuid string, index int32, instanceKey *models.ActualLRPInstanceKey) (*models.ActualLRP, *models.ActualLRP, error) {
	logger = logger.WithData(lager.Data{"process_guid": processGuid, "index": index, "instance_key": instanceKey})
	logger.Info("starting")
	defer logger.Info("complete")

	var beforeActualLRP models.ActualLRP
	var actualLRP *models.ActualLRP
	err := db.transact(logger, func(logger lager.Logger, tx helpers.Tx) error {
		var err error
		actualLRP, err = db.fetchActualLRPForUpdate(logger, processGuid, index, false, false, tx)
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
		actualLRP.PlacementState = models.PlacementStateType_Normal
		actualLRP.ActualLRPInstanceKey = *instanceKey
		actualLRP.PlacementError = ""
		actualLRP.ActualLRPNetInfo = models.ActualLRPNetInfo{}
		actualLRP.Since = db.clock.Now().UnixNano()
		netInfoData, err := db.serializeModel(logger, &models.ActualLRPNetInfo{})
		if err != nil {
			logger.Error("failed-to-serialize-net-info", err)
			return err
		}

		_, err = db.update(logger, tx, actualLRPsTable,
			helpers.SQLAttributes{
				"state":                  actualLRP.State,
				"cell_id":                actualLRP.CellId,
				"instance_guid":          actualLRP.InstanceGuid,
				"modification_tag_index": actualLRP.ModificationTag.Index,
				"placement_error":        actualLRP.PlacementError,
				"since":                  actualLRP.Since,
				"net_info":               netInfoData,
			},
			"process_guid = ? AND instance_index = ? AND evacuating = ? AND suspect = ?",
			processGuid, index, false, false,
		)
		if err != nil {
			logger.Error("failed-claiming-actual-lrp", err)
			return err
		}

		return nil
	})

	return &beforeActualLRP, actualLRP, err
}

func (db *SQLDB) StartActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, netInfo *models.ActualLRPNetInfo) (*models.ActualLRP, *models.ActualLRP, error) {
	logger = logger.WithData(lager.Data{"actual_lrp_key": key, "actual_lrp_instance_key": instanceKey, "net_info": netInfo})

	var beforeActualLRP models.ActualLRP
	var actualLRP *models.ActualLRP

	err := db.transact(logger, func(logger lager.Logger, tx helpers.Tx) error {
		var err error
		actualLRP, err = db.fetchActualLRPForUpdate(logger, key.ProcessGuid, key.Index, false, false, tx)
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
			actualLRP.State == models.ActualLRPStateRunning &&
			actualLRP.PlacementState == models.PlacementStateType_Normal {
			logger.Debug("nothing-to-change")
			return nil
		}

		if !actualLRP.AllowsTransitionTo(key, instanceKey, models.ActualLRPStateRunning) {
			logger.Error("failed-to-transition-actual-lrp-to-started", nil)
			return models.ErrActualLRPCannotBeStarted
		}

		logger.Info("starting")
		defer logger.Info("completed")

		now := db.clock.Now().UnixNano()
		evacuating := false
		suspect := false

		actualLRP.ActualLRPInstanceKey = *instanceKey
		actualLRP.ActualLRPNetInfo = *netInfo
		actualLRP.State = models.ActualLRPStateRunning
		actualLRP.Since = now
		actualLRP.PlacementState = models.PlacementStateType_Normal
		actualLRP.ModificationTag.Increment()
		actualLRP.PlacementError = ""

		netInfoData, err := db.serializeModel(logger, &actualLRP.ActualLRPNetInfo)
		if err != nil {
			logger.Error("failed-to-serialize-net-info", err)
			return err
		}

		_, err = db.update(logger, tx, actualLRPsTable,
			helpers.SQLAttributes{
				"state":                  actualLRP.State,
				"cell_id":                actualLRP.CellId,
				"instance_guid":          actualLRP.InstanceGuid,
				"modification_tag_index": actualLRP.ModificationTag.Index,
				"placement_error":        actualLRP.PlacementError,
				"since":                  actualLRP.Since,
				"net_info":               netInfoData,
			},
			"process_guid = ? AND instance_index = ? AND evacuating = ? AND suspect = ?",
			key.ProcessGuid, key.Index, evacuating, suspect,
		)
		if err != nil {
			logger.Error("failed-starting-actual-lrp", err)
			return err
		}

		return nil
	})

	return &beforeActualLRP, actualLRP, err
}

func truncateString(s string, maxLen int) string {
	l := len(s)
	if l < maxLen {
		return s
	}
	return s[:maxLen]
}

func (db *SQLDB) CrashActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, crashReason string) (*models.ActualLRP, *models.ActualLRP, bool, error) {
	logger = logger.WithData(lager.Data{"key": key, "instance_key": instanceKey, "crash_reason": crashReason})
	logger.Info("starting")
	defer logger.Info("complete")

	var immediateRestart = false
	var beforeActualLRP models.ActualLRP
	var actualLRP *models.ActualLRP

	err := db.transact(logger, func(logger lager.Logger, tx helpers.Tx) error {
		var err error
		actualLRP, err = db.fetchActualLRPForUpdate(logger, key.ProcessGuid, key.Index, false, false, tx)
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
		actualLRP.PlacementState = models.PlacementStateType_Normal

		actualLRP.ActualLRPInstanceKey.InstanceGuid = ""
		actualLRP.ActualLRPInstanceKey.CellId = ""
		actualLRP.ActualLRPNetInfo = models.ActualLRPNetInfo{}
		actualLRP.CrashCount = newCrashCount
		actualLRP.CrashReason = crashReason
		netInfoData, err := db.serializeModel(logger, &actualLRP.ActualLRPNetInfo)
		if err != nil {
			logger.Error("failed-to-serialize-net-info", err)
			return err
		}
		evacuating := false
		suspect := false

		if actualLRP.ShouldRestartImmediately(models.NewDefaultRestartCalculator()) {
			actualLRP.State = models.ActualLRPStateUnclaimed
			immediateRestart = true
		}

		now := db.clock.Now().UnixNano()
		actualLRP.Since = now

		_, err = db.update(logger, tx, actualLRPsTable,
			helpers.SQLAttributes{
				"state":                  actualLRP.State,
				"cell_id":                actualLRP.CellId,
				"instance_guid":          actualLRP.InstanceGuid,
				"modification_tag_index": actualLRP.ModificationTag.Index,
				"crash_count":            actualLRP.CrashCount,
				"crash_reason":           truncateString(actualLRP.CrashReason, 1024),
				"since":                  actualLRP.Since,
				"net_info":               netInfoData,
			},
			"process_guid = ? AND instance_index = ? AND evacuating = ? AND suspect = ?",
			key.ProcessGuid, key.Index, evacuating, suspect,
		)
		if err != nil {
			logger.Error("failed-to-crash-actual-lrp", err)
			return err
		}

		return nil
	})

	return &beforeActualLRP, actualLRP, immediateRestart, err
}

func (db *SQLDB) FailActualLRP(logger lager.Logger, key *models.ActualLRPKey, placementError string) (*models.ActualLRP, *models.ActualLRP, error) {
	logger = logger.WithData(lager.Data{"actual_lrp_key": key, "placement_error": placementError})
	logger.Info("starting")
	defer logger.Info("complete")

	var beforeActualLRP models.ActualLRP
	var actualLRP *models.ActualLRP

	err := db.transact(logger, func(logger lager.Logger, tx helpers.Tx) error {
		var err error
		actualLRP, err = db.fetchActualLRPForUpdate(logger, key.ProcessGuid, key.Index, false, false, tx)
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

		_, err = db.update(logger, tx, actualLRPsTable,
			helpers.SQLAttributes{
				"modification_tag_index": actualLRP.ModificationTag.Index,
				"placement_error":        truncateString(actualLRP.PlacementError, 1024),
				"since":                  actualLRP.Since,
			},
			"process_guid = ? AND instance_index = ? AND evacuating = ?",
			key.ProcessGuid, key.Index, evacuating,
		)
		if err != nil {
			logger.Error("failed-failing-actual-lrp", err)
			return err
		}

		return nil
	})

	return &beforeActualLRP, actualLRP, err
}

func (db *SQLDB) RemoveActualLRP(logger lager.Logger, processGuid string, index int32, instanceKey *models.ActualLRPInstanceKey) error {
	logger = logger.WithData(lager.Data{"process_guid": processGuid, "index": index})
	logger.Info("starting")
	defer logger.Info("complete")

	return db.transact(logger, func(logger lager.Logger, tx helpers.Tx) error {
		var err error
		var result sql.Result
		if instanceKey == nil {
			result, err = db.delete(logger, tx, actualLRPsTable,
				"process_guid = ? AND instance_index = ? AND evacuating = ?",
				processGuid, index, false,
			)
		} else {
			result, err = db.delete(logger, tx, actualLRPsTable,
				"process_guid = ? AND instance_index = ? AND evacuating = ? AND instance_guid = ? AND cell_id = ?",
				processGuid, index, false, instanceKey.InstanceGuid, instanceKey.CellId,
			)
		}
		if err != nil {
			logger.Error("failed-removing-actual-lrp", err)
			return err
		}

		numRows, err := result.RowsAffected()
		if err != nil {
			logger.Error("failed-getting-rows-affected", err)
			return err
		}
		if numRows == 0 {
			logger.Debug("not-found", lager.Data{"instance_key": instanceKey})
			return models.ErrResourceNotFound
		}

		return nil
	})
}

func (db *SQLDB) createRunningActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, netInfo *models.ActualLRPNetInfo, tx helpers.Tx) (*models.ActualLRP, error) {
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

	_, err = db.insert(logger, tx, actualLRPsTable,
		helpers.SQLAttributes{
			"process_guid":           actualLRP.ActualLRPKey.ProcessGuid,
			"instance_index":         actualLRP.ActualLRPKey.Index,
			"domain":                 actualLRP.ActualLRPKey.Domain,
			"instance_guid":          actualLRP.ActualLRPInstanceKey.InstanceGuid,
			"cell_id":                actualLRP.ActualLRPInstanceKey.CellId,
			"state":                  actualLRP.State,
			"net_info":               netInfoData,
			"since":                  actualLRP.Since,
			"modification_tag_epoch": actualLRP.ModificationTag.Epoch,
			"modification_tag_index": actualLRP.ModificationTag.Index,
		},
	)
	if err != nil {
		logger.Error("failed-creating-running-actual-lrp", err)
		return nil, err
	}
	return actualLRP, nil
}

func (db *SQLDB) scanToFlattenedActualLRP(logger lager.Logger, row helpers.RowScanner) (*models.ActualLRP, error) {
	var netInfoData []byte
	var actualLRP models.ActualLRP
	var evacuating, suspect bool

	err := row.Scan(
		&actualLRP.ProcessGuid,
		&actualLRP.Index,
		&evacuating,
		&suspect,
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
		return nil, err
	}

	if len(netInfoData) > 0 {
		logger.Debug("unmarshalling-net-info-data", lager.Data{"net_info": string(netInfoData)})
		err = db.deserializeModel(logger, netInfoData, &actualLRP.ActualLRPNetInfo)
		if err != nil {
			logger.Error("failed-unmarshaling-net-info-data", err)
			return &actualLRP, models.ErrDeserialize
		}
	}

	if evacuating {
		actualLRP.PlacementState = models.PlacementStateType_Evacuating
	} else if suspect {
		actualLRP.PlacementState = models.PlacementStateType_Suspect
	} else {
		actualLRP.PlacementState = models.PlacementStateType_Normal
	}

	return &actualLRP, nil
}

func (db *SQLDB) scanToActualLRP(logger lager.Logger, row helpers.RowScanner) (*models.ActualLRP, bool, error) {
	var netInfoData []byte
	var actualLRP models.ActualLRP
	var evacuating, suspect bool

	err := row.Scan(
		&actualLRP.ProcessGuid,
		&actualLRP.Index,
		&evacuating,
		&suspect,
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
		return nil, false, err
	}

	if len(netInfoData) > 0 {
		logger.Debug("unmarshalling-net-info-data", lager.Data{"net_info": string(netInfoData)})
		err = db.deserializeModel(logger, netInfoData, &actualLRP.ActualLRPNetInfo)
		if err != nil {
			logger.Error("failed-unmarshaling-net-info-data", err)
			return &actualLRP, evacuating, models.ErrDeserialize
		}
	}

	return &actualLRP, evacuating, nil
}

type actualToDelete struct {
	*models.ActualLRP
	evacuating bool
}

type flattenedActualToDelete struct {
	*models.ActualLRP
	evacuating bool
}

func (db *SQLDB) SuspectActualLRP(logger lager.Logger, key *models.ActualLRPKey) (*models.ActualLRP, *models.ActualLRP, error) {
	logger = logger.Session("marking-as-suspect", lager.Data{"key": key})

	var beforeActualLRP models.ActualLRP
	var actualLRP *models.ActualLRP
	processGuid := key.ProcessGuid
	index := key.Index

	err := db.transact(logger, func(logger lager.Logger, tx helpers.Tx) error {
		var err error
		actualLRP, err = db.fetchActualLRPForUpdate(logger, processGuid, index, false, false, tx)
		if err != nil {
			logger.Error("failed-fetching-actual-lrp-for-suspect-update", err)
			return err
		}
		beforeActualLRP = *actualLRP

		logger.Info("starting")
		defer logger.Info("complete")

		now := db.clock.Now().UnixNano()
		actualLRP.ModificationTag.Increment()
		actualLRP.Since = now
		actualLRP.PlacementState = models.PlacementStateType_Suspect

		_, err = db.update(logger, tx, actualLRPsTable,
			helpers.SQLAttributes{
				"modification_tag_index": actualLRP.ModificationTag.Index,
				"since":                  actualLRP.Since,
				"suspect":                true,
			},
			"process_guid = ? AND instance_index = ? AND evacuating = ? AND suspect = ?",
			processGuid, index, false, false,
		)
		if err != nil {
			logger.Error("failed-to-suspect-actual-lrp", err)
			return err
		}

		return nil
	})

	_, err = db.CreateUnclaimedActualLRP(logger, key)
	if err != nil {
		logger.Error("create-replacement-for-suspect-failed", err)
		return nil, nil, err
	}

	return &beforeActualLRP, actualLRP, err
}

func (db *SQLDB) UnsuspectActualLRP(logger lager.Logger, key *models.ActualLRPKey) error {
	logger = logger.Session("marking-as-unsuspect", lager.Data{"key": key})

	var actualLRPs []*models.ActualLRP
	processGuid := key.ProcessGuid
	index := key.Index

	err := db.transact(logger, func(logger lager.Logger, tx helpers.Tx) error {
		var err error
		actualLRPs, err = db.ActualLRPs(logger, models.ActualLRPFilter{ProcessGUID: &processGuid, Index: &index})
		if err != nil {
			logger.Error("failed-fetching-actual-lrp-for-share", err)
			return err
		}

		unsuspectActualLRP := func(actualLRP *models.ActualLRP) error {
			logger.Info("starting")
			defer logger.Info("complete")

			now := db.clock.Now().UnixNano()
			actualLRP.ModificationTag.Increment()
			actualLRP.Since = now
			actualLRP.PlacementState = models.PlacementStateType_Normal

			_, err = db.update(logger, tx, actualLRPsTable,
				helpers.SQLAttributes{
					"modification_tag_index": actualLRP.ModificationTag.Index,
					"since":                  actualLRP.Since,
					"suspect":                false,
				},
				"process_guid = ? AND instance_index = ? AND evacuating = ? AND suspect = ?",
				processGuid, index, false, true,
			)
			if err != nil {
				logger.Error("failed-to-suspect-actual-lrp", err)
				return err
			}
			return nil
		}

		deleteActualLRP := func(actualLRP *models.ActualLRP) error {
			_, err := db.delete(logger, tx, actualLRPsTable,
				"process_guid = ? AND instance_index = ? AND evacuating = ? AND instance_guid = ? AND cell_id = ?",
				actualLRP.ProcessGuid, actualLRP.Index, false, actualLRP.ActualLRPInstanceKey.InstanceGuid, actualLRP.ActualLRPInstanceKey.CellId,
			)
			if err != nil {
				logger.Error("failed-removing-actual-lrp", err)
				return err
			}
			return nil
		}

		err = nil
		for _, actualLRP := range actualLRPs {
			switch actualLRP.PlacementState {
			case models.PlacementStateType_Suspect:
				err = unsuspectActualLRP(actualLRP)
			default:
				err = deleteActualLRP(actualLRP)
			}
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

func (db *SQLDB) RemoveSuspectActualLRP(logger lager.Logger, lrpKey *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey) error {
	logger = logger.Session("remove-suspect-lrp", lager.Data{"lrp_key": lrpKey, "instance_key": instanceKey})
	logger.Debug("starting")
	defer logger.Debug("complete")

	return db.transact(logger, func(logger lager.Logger, tx helpers.Tx) error {
		processGuid := lrpKey.ProcessGuid
		index := lrpKey.Index

		lrp, err := db.fetchActualLRPForUpdate(logger, processGuid, index, false, true, tx)
		if err == models.ErrResourceNotFound {
			logger.Debug("suspect-lrp-does-not-exist")
			return nil
		}

		if err != nil {
			logger.Error("failed-fetching-actual-lrp", err)
			return err
		}

		if !lrp.ActualLRPInstanceKey.Equal(instanceKey) {
			logger.Debug("actual-lrp-instance-key-mismatch", lager.Data{"instance_key_param": instanceKey, "instance_key_from_db": lrp.ActualLRPInstanceKey})
			return models.ErrActualLRPCannotBeRemoved
		}

		_, err = db.delete(logger, tx, "actual_lrps",
			"process_guid = ? AND instance_index = ? AND evacuating = ? AND suspect = ?",
			processGuid, index, false, true,
		)
		if err != nil {
			logger.Error("failed-delete", err)
			return models.ErrActualLRPCannotBeRemoved
		}

		return nil
	})
}

func (db *SQLDB) fetchActualLRPForUpdate(logger lager.Logger, processGuid string, index int32, evacuating, suspect bool, tx helpers.Tx) (*models.ActualLRP, error) {
	wheres := "process_guid = ? AND instance_index = ? AND evacuating = ? AND suspect = ?"
	bindings := []interface{}{processGuid, index, evacuating, suspect}

	rows, err := db.all(logger, tx, actualLRPsTable,
		actualLRPColumns, helpers.LockRow, wheres, bindings...)
	if err != nil {
		logger.Error("failed-query", err)
		return nil, err
	}
	lrps, err := db.scanAndCleanupFlattenedActualLRPs(logger, tx, rows)
	if err != nil {
		return nil, err
	}

	if len(lrps) == 0 {
		return nil, models.ErrResourceNotFound
	}

	return lrps[0], nil
}

func (db *SQLDB) scanAndCleanupFlattenedActualLRPs(logger lager.Logger, q helpers.Queryable, rows *sql.Rows) ([]*models.ActualLRP, error) {
	result := []*models.ActualLRP{}
	actualsToDelete := []*flattenedActualToDelete{}
	for rows.Next() {
		actualLRP, err := db.scanToFlattenedActualLRP(logger, rows)
		if err == models.ErrDeserialize {
			actualsToDelete = append(actualsToDelete, &flattenedActualToDelete{actualLRP, actualLRP.PlacementState == models.PlacementStateType_Evacuating})
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
		result = append(result, actualLRP)
	}

	if rows.Err() != nil {
		logger.Error("failed-getting-next-row", rows.Err())
		return nil, db.convertSQLError(rows.Err())
	}

	for _, actual := range actualsToDelete {
		_, err := db.delete(logger, q, actualLRPsTable,
			"process_guid = ? AND instance_index = ? AND evacuating = ?",
			actual.ProcessGuid, actual.Index, actual.evacuating,
		)
		if err != nil {
			logger.Error("failed-cleaning-up-invalid-actual-lrp", err)
		}
	}

	return result, nil
}

func (db *SQLDB) scanAndCleanupActualLRPs(logger lager.Logger, q helpers.Queryable, rows *sql.Rows) ([]*models.ActualLRPGroup, error) {
	mapOfGroups := map[models.ActualLRPKey]*models.ActualLRPGroup{}
	result := []*models.ActualLRPGroup{}
	actualsToDelete := []*actualToDelete{}
	for rows.Next() {
		actualLRP, evacuating, err := db.scanToActualLRP(logger, rows)
		if err == models.ErrDeserialize {
			actualsToDelete = append(actualsToDelete, &actualToDelete{actualLRP, evacuating})
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

	for _, actual := range actualsToDelete {
		_, err := db.delete(logger, q, actualLRPsTable,
			"process_guid = ? AND instance_index = ? AND evacuating = ?",
			actual.ProcessGuid, actual.Index, actual.evacuating,
		)
		if err != nil {
			logger.Error("failed-cleaning-up-invalid-actual-lrp", err)
		}
	}

	return result, nil
}
