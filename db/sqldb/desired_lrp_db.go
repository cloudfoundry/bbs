package sqldb

import (
	"database/sql"
	"encoding/json"
	"strings"

	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

func (db *SQLDB) DesireLRP(logger lager.Logger, desiredLRP *models.DesiredLRP) error {
	logger = logger.WithData(lager.Data{"process_guid": desiredLRP.ProcessGuid})
	logger.Info("starting")
	defer logger.Info("complete")

	return db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		routesData, err := json.Marshal(desiredLRP.Routes)
		runInfo := desiredLRP.DesiredLRPRunInfo(db.clock.Now())

		runInfoData, err := db.serializeModel(logger, &runInfo)
		if err != nil {
			logger.Error("failed-to-serialize-model", err)
			return err
		}

		volumePlacement := &models.VolumePlacement{}
		volumePlacement.DriverNames = []string{}
		for _, mount := range desiredLRP.VolumeMounts {
			volumePlacement.DriverNames = append(volumePlacement.DriverNames, mount.Driver)
		}

		volumePlacementData, err := db.serializeModel(logger, volumePlacement)
		if err != nil {
			logger.Error("failed-to-serialize-model", err)
			return err
		}

		if desiredLRP.RunInfoTag == nil {
			runInfoTag := desiredLRP.ProcessGuid + "-initial"
			desiredLRP.RunInfoTag = &runInfoTag
		}

		_, err = db.insert(logger, tx, runInfosTable,
			SQLAttributes{
				"tag":  desiredLRP.RunInfoTag,
				"data": runInfoData,
			},
		)
		if err != nil {
			logger.Error("failed-to-insert-run-info", err)
			return db.convertSQLError(err)
		}

		guid, err := db.guidProvider.NextGUID()
		if err != nil {
			logger.Error("failed-to-generate-guid", err)
			return models.ErrGUIDGeneration
		}

		desiredLRP.ModificationTag = &models.ModificationTag{Epoch: guid, Index: 0}

		_, err = db.insert(logger, tx, desiredLRPsTable,
			SQLAttributes{
				"process_guid":           desiredLRP.ProcessGuid,
				"domain":                 desiredLRP.Domain,
				"log_guid":               desiredLRP.LogGuid,
				"annotation":             desiredLRP.Annotation,
				"instances":              desiredLRP.Instances,
				"memory_mb":              desiredLRP.MemoryMb,
				"disk_mb":                desiredLRP.DiskMb,
				"rootfs":                 desiredLRP.RootFs,
				"volume_placement":       volumePlacementData,
				"modification_tag_epoch": desiredLRP.ModificationTag.Epoch,
				"modification_tag_index": desiredLRP.ModificationTag.Index,
				"routes":                 routesData,
				"run_info":               runInfoData,
				"run_info_tag":           desiredLRP.RunInfoTag,
			},
		)
		if err != nil {
			logger.Error("failed-inserting-desired", err)
			return db.convertSQLError(err)
		}
		return nil
	})
}

func (db *SQLDB) DesiredLRPByProcessGuid(logger lager.Logger, processGuid string) (*models.DesiredLRP, error) {
	logger = logger.WithData(lager.Data{"process_guid": processGuid})
	logger.Debug("starting")
	defer logger.Debug("complete")

	tables := []string{desiredLRPsTable, runInfosTable1, runInfosTable2}

	row := db.oneMultiTable(logger, db.db, tables,
		desiredLRPAndRunInfoColumns, NoLockRow,
		"process_guid = ?", processGuid,
	)

	return db.fetchDesiredLRP(logger, row)
}

func (db *SQLDB) DesiredLRPs(logger lager.Logger, filter models.DesiredLRPFilter) ([]*models.DesiredLRP, error) {
	logger = logger.WithData(lager.Data{"filter": filter})
	logger.Debug("start")
	defer logger.Debug("complete")

	var wheres []string
	var values []interface{}

	if filter.Domain != "" {
		wheres = append(wheres, "domain = ?")
		values = append(values, filter.Domain)
	}

	tables := []string{desiredLRPsTable, runInfosTable1, runInfosTable2}
	rows, err := db.allMultiTable(logger, db.db, tables,
		desiredLRPAndRunInfoColumns, NoLockRow,
		strings.Join(wheres, " AND "), values...,
	)
	if err != nil {
		logger.Error("failed-query", err)
		return nil, db.convertSQLError(err)
	}
	defer rows.Close()

	results := []*models.DesiredLRP{}
	for rows.Next() {
		desiredLRP, err := db.fetchDesiredLRP(logger, rows)
		if err != nil {
			logger.Error("failed-reading-row", err)
			continue
		}
		results = append(results, desiredLRP)
	}

	if rows.Err() != nil {
		logger.Error("failed-fetching-row", rows.Err())
		return nil, db.convertSQLError(rows.Err())
	}

	return results, nil
}

func (db *SQLDB) DesiredLRPSchedulingInfos(logger lager.Logger, filter models.DesiredLRPFilter) ([]*models.DesiredLRPSchedulingInfo, error) {
	logger = logger.WithData(lager.Data{"filter": filter})
	logger.Debug("start")
	defer logger.Debug("complete")

	var wheres []string
	var values []interface{}

	if filter.Domain != "" {
		wheres = append(wheres, "domain = ?")
		values = append(values, filter.Domain)
	}

	rows, err := db.all(logger, db.db, desiredLRPsTable,
		schedulingInfoColumns, NoLockRow,
		strings.Join(wheres, " AND "), values...,
	)
	if err != nil {
		logger.Error("failed-query", err)
		return nil, db.convertSQLError(err)
	}
	defer rows.Close()

	results := []*models.DesiredLRPSchedulingInfo{}
	for rows.Next() {
		desiredLRPSchedulingInfo, err := db.fetchDesiredLRPSchedulingInfo(logger, rows)
		if err != nil {
			logger.Error("failed-reading-row", err)
			continue
		}
		results = append(results, desiredLRPSchedulingInfo)
	}

	if rows.Err() != nil {
		logger.Error("failed-fetching-row", rows.Err())
		return nil, db.convertSQLError(rows.Err())
	}

	return results, nil
}

func (db *SQLDB) UpdateDesiredLRP(logger lager.Logger, processGuid string, update *models.DesiredLRPUpdate) (*models.DesiredLRP, error) {
	logger = logger.WithData(lager.Data{"process_guid": processGuid})
	logger.Info("starting")
	defer logger.Info("complete")

	var beforeDesiredLRP *models.DesiredLRP
	err := db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		var err error

		tables := []string{desiredLRPsTable, runInfosTable1, runInfosTable2}
		row := db.oneMultiTable(logger, tx, tables,
			desiredLRPAndRunInfoColumns, LockRow,
			"process_guid = ?", processGuid,
		)
		beforeDesiredLRP, err = db.fetchDesiredLRP(logger, row)
		if err != nil {
			logger.Error("failed-lock-desired", err)
			return err
		}

		updateAttributes := SQLAttributes{"modification_tag_index": beforeDesiredLRP.ModificationTag.Index + 1}

		if update.Annotation != nil {
			updateAttributes["annotation"] = *update.Annotation
		}

		if update.Instances != nil {
			updateAttributes["instances"] = *update.Instances
		}

		if update.Routes != nil {
			routeData, err := json.Marshal(update.Routes)
			if err != nil {
				logger.Error("failed-marshalling-routes", err)
				return models.ErrBadRequest
			}
			updateAttributes["routes"] = routeData
		}

		if update.NewDesired != nil {
			runInfo := update.NewDesired.DesiredLRPRunInfo(db.clock.Now())

			runInfoData, err := db.serializeModel(logger, &runInfo)
			if err != nil {
				logger.Error("failed-to-serialize-model", err)
				return err
			}

			newRunInfoTag := update.NewDesired.ProcessGuid + "-" + *update.NewDesired.RunInfoTag

			_, err = db.insert(logger, tx, runInfosTable,
				SQLAttributes{
					"tag":  &newRunInfoTag,
					"data": runInfoData,
				},
			)
			if err != nil {
				logger.Error("failed-to-insert-run-info", err)
				return err
			}

			if beforeDesiredLRP.RunInfo_1 != nil {
				updateAttributes["run_info_tag_2"] = beforeDesiredLRP.RunInfo_1.RunInfoTag
			}
			updateAttributes["run_info_tag_1"] = beforeDesiredLRP.RunInfoTag
			updateAttributes["run_info_tag"] = newRunInfoTag
			updateAttributes["run_info"] = runInfoData
		}

		_, err = db.update(logger, tx, desiredLRPsTable, updateAttributes, `process_guid = ?`, processGuid)
		if err != nil {
			logger.Error("failed-executing-query", err)
			return db.convertSQLError(err)
		}

		return nil
	})

	return beforeDesiredLRP, err
}

func (db *SQLDB) RemoveDesiredLRP(logger lager.Logger, processGuid string) error {
	logger = logger.WithData(lager.Data{"process_guid": processGuid})
	logger.Info("starting")
	defer logger.Info("complete")

	return db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		err := db.lockDesiredLRPByGuidForUpdate(logger, processGuid, tx)
		if err != nil {
			logger.Error("failed-lock-desired", err)
			return err
		}

		_, err = db.delete(logger, tx, desiredLRPsTable, "process_guid = ?", processGuid)
		if err != nil {
			logger.Error("failed-deleting-from-db", err)
			return db.convertSQLError(err)
		}

		return nil
	})
}

// "rows" needs to have the columns defined in the schedulingInfoColumns constant
func (db *SQLDB) fetchDesiredLRPSchedulingInfoAndMore(logger lager.Logger, scanner RowScanner, dest ...interface{}) (*models.DesiredLRPSchedulingInfo, error) {
	schedulingInfo := &models.DesiredLRPSchedulingInfo{}
	var routeData, volumePlacementData []byte

	values := []interface{}{
		&schedulingInfo.ProcessGuid,
		&schedulingInfo.Domain,
		&schedulingInfo.LogGuid,
		&schedulingInfo.Annotation,
		&schedulingInfo.Instances,
		&schedulingInfo.MemoryMb,
		&schedulingInfo.DiskMb,
		&schedulingInfo.RootFs,
		&routeData,
		&volumePlacementData,
		&schedulingInfo.ModificationTag.Epoch,
		&schedulingInfo.ModificationTag.Index,
	}
	values = append(values, dest...)

	err := scanner.Scan(values...)
	if err != nil {
		logger.Error("failed-scanning", err)
		return nil, err
	}

	var routes models.Routes
	err = json.Unmarshal(routeData, &routes)
	if err != nil {
		logger.Error("failed-parsing-routes", err)
		return nil, err
	}
	schedulingInfo.Routes = routes

	var volumePlacement models.VolumePlacement
	err = db.deserializeModel(logger, volumePlacementData, &volumePlacement)
	if err != nil {
		logger.Error("failed-parsing-volume-placement", err)
		return nil, err
	}
	schedulingInfo.VolumePlacement = &volumePlacement

	return schedulingInfo, nil
}

func (db *SQLDB) lockDesiredLRPByGuidForUpdate(logger lager.Logger, processGuid string, tx *sql.Tx) error {
	row := db.one(logger, tx, desiredLRPsTable,
		ColumnList{"1"}, LockRow,
		"process_guid = ?", processGuid,
	)
	var count int
	err := row.Scan(&count)
	if err == sql.ErrNoRows {
		return models.ErrResourceNotFound
	} else if err != nil {
		return db.convertSQLError(err)
	}
	return nil
}

func (db *SQLDB) fetchDesiredLRP(logger lager.Logger, scanner RowScanner) (*models.DesiredLRP, error) {
	var runInfoData []byte
	var runInfo1Data []byte
	var runInfo2Data []byte
	var runInfoTag, runInfo1Tag, runInfo2Tag, runInfoTag1, runInfoTag2 sql.NullString
	schedulingInfo, err := db.fetchDesiredLRPSchedulingInfoAndMore(logger, scanner,
		&runInfo1Tag,
		&runInfo1Data,
		&runInfo2Tag,
		&runInfo2Data,
		&runInfoTag,
		&runInfoTag1,
		&runInfoTag2,
		&runInfoData)
	if err != nil {
		logger.Error("failed-fetching-run-info", err)
		return nil, models.ErrResourceNotFound
	}

	var runInfo, runInfo1, runInfo2 models.DesiredLRPRunInfo
	err = db.deserializeModel(logger, runInfoData, &runInfo)
	if err != nil {
		_, err := db.delete(logger, db.db, desiredLRPsTable, "process_guid = ?", schedulingInfo.ProcessGuid)
		if err != nil {
			logger.Error("failed-deleting-invalid-row", err)
		}
		return nil, models.ErrDeserialize
	}
	desiredLRP := models.NewDesiredLRP(*schedulingInfo, runInfo)
	if runInfoTag.Valid {
		desiredLRP.RunInfoTag = &runInfoTag.String
	}

	if runInfoTag1.Valid {
		err = db.deserializeModel(logger, runInfo1Data, &runInfo1)
		runInfo1.RunInfoTag = runInfoTag1.String
		desiredLRP.RunInfo_1 = &runInfo1
	}
	if runInfoTag2.Valid {
		err = db.deserializeModel(logger, runInfo2Data, &runInfo2)
		runInfo2.RunInfoTag = runInfoTag2.String
		desiredLRP.RunInfo_2 = &runInfo2
	}

	return &desiredLRP, nil
}

func (db *SQLDB) fetchDesiredLRPSchedulingInfo(logger lager.Logger, scanner RowScanner) (*models.DesiredLRPSchedulingInfo, error) {
	return db.fetchDesiredLRPSchedulingInfoAndMore(logger, scanner)
}
