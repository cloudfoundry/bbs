package sqldb

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

func (db *SQLDB) DesireLRP(ctx context.Context, logger lager.Logger, desiredLRP *models.DesiredLRP) error {
	logger = logger.Session("db-desire-lrp", lager.Data{"process_guid": desiredLRP.ProcessGuid})
	logger.Info("starting")
	defer logger.Info("complete")

	return db.transact(ctx, logger, func(logger lager.Logger, tx helpers.Tx) error {
		routesData, err := db.encodeRouteData(logger, desiredLRP.Routes)
		if err != nil {
			logger.Error("failed-encoding-route-data", err)
			return err
		}

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

		guid, err := db.guidProvider.NextGUID()
		if err != nil {
			logger.Error("failed-to-generate-guid", err)
			return models.ErrGUIDGeneration
		}

		placementTagData, err := json.Marshal(desiredLRP.PlacementTags)
		if err != nil {
			logger.Error("failed-to-serialize-model", err)
			return err
		}

		desiredLRP.ModificationTag = &models.ModificationTag{Epoch: guid, Index: 0}

		_, err = db.insert(ctx, logger, tx, desiredLRPsTable,
			helpers.SQLAttributes{
				"process_guid":           desiredLRP.ProcessGuid,
				"domain":                 desiredLRP.Domain,
				"log_guid":               desiredLRP.LogGuid,
				"annotation":             desiredLRP.Annotation,
				"instances":              desiredLRP.Instances,
				"memory_mb":              desiredLRP.MemoryMb,
				"disk_mb":                desiredLRP.DiskMb,
				"max_pids":               desiredLRP.MaxPids,
				"rootfs":                 desiredLRP.RootFs,
				"volume_placement":       volumePlacementData,
				"modification_tag_epoch": desiredLRP.ModificationTag.Epoch,
				"modification_tag_index": desiredLRP.ModificationTag.Index,
				"routes":                 routesData,
				"run_info":               runInfoData,
				"placement_tags":         placementTagData,
			},
		)
		if err != nil {
			logger.Error("failed-inserting-desired", err)
			return err
		}
		return nil
	})
}

func (db *SQLDB) DesiredLRPByProcessGuid(ctx context.Context, logger lager.Logger, processGuid string) (*models.DesiredLRP, error) {
	logger = logger.Session("db-desired-lrp-by-process-guid", lager.Data{"process_guid": processGuid})
	logger.Debug("starting")
	defer logger.Debug("complete")

	var desiredLRP *models.DesiredLRP

	err := db.transact(ctx, logger, func(logger lager.Logger, tx helpers.Tx) error {
		var err error
		row := db.one(ctx, logger, tx, desiredLRPsTable,
			desiredLRPColumns, helpers.NoLockRow,
			"process_guid = ?", processGuid,
		)

		desiredLRP, err = db.fetchDesiredLRP(ctx, logger, row, tx)
		return err
	})

	return desiredLRP, err
}

func (db *SQLDB) DesiredLRPs(ctx context.Context, logger lager.Logger, filter models.DesiredLRPFilter) ([]*models.DesiredLRP, error) {
	logger = logger.Session("db-desired-lrps", lager.Data{"filter": filter})
	logger.Debug("start")
	defer logger.Debug("complete")

	var wheres []string
	var values []interface{}

	if filter.Domain != "" {
		wheres = append(wheres, "domain = ?")
		values = append(values, filter.Domain)
	}

	if len(filter.ProcessGuids) > 0 {
		wheres = append(wheres, whereClauseForProcessGuids(filter.ProcessGuids))

		for _, guid := range filter.ProcessGuids {
			values = append(values, guid)
		}
	}

	results := []*models.DesiredLRP{}

	err := db.transact(ctx, logger, func(logger lager.Logger, tx helpers.Tx) error {
		rows, err := db.all(ctx, logger, tx, desiredLRPsTable,
			desiredLRPColumns, helpers.NoLockRow,
			strings.Join(wheres, " AND "), values...,
		)
		if err != nil {
			logger.Error("failed-query", err)
			return err
		}
		defer rows.Close()

		results, err = db.fetchDesiredLRPs(ctx, logger, rows, tx)
		if err != nil {
			logger.Error("failed-fetching-row", rows.Err())
			return db.convertSQLError(rows.Err())
		}

		return nil
	})

	return results, err
}

func (db *SQLDB) DesiredLRPSchedulingInfos(ctx context.Context, logger lager.Logger, filter models.DesiredLRPFilter) ([]*models.DesiredLRPSchedulingInfo, error) {
	logger = logger.Session("db-desired-lrps-scheduling-infos", lager.Data{"filter": filter})
	logger.Debug("starting")
	defer logger.Debug("complete")

	var wheres []string
	var values []interface{}

	if filter.Domain != "" {
		wheres = append(wheres, "domain = ?")
		values = append(values, filter.Domain)
	}

	if len(filter.ProcessGuids) > 0 {
		wheres = append(wheres, whereClauseForProcessGuids(filter.ProcessGuids))

		for _, guid := range filter.ProcessGuids {
			values = append(values, guid)
		}
	}

	results := []*models.DesiredLRPSchedulingInfo{}

	err := db.transact(ctx, logger, func(logger lager.Logger, tx helpers.Tx) error {
		rows, err := db.all(ctx, logger, tx, desiredLRPsTable,
			schedulingInfoColumns, helpers.NoLockRow,
			strings.Join(wheres, " AND "), values...,
		)
		if err != nil {
			logger.Error("failed-query", err)
			return err
		}
		defer rows.Close()

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
			return db.convertSQLError(rows.Err())
		}

		return nil
	})

	return results, err
}

func (db *SQLDB) UpdateDesiredLRP(ctx context.Context, logger lager.Logger, processGuid string, update *models.DesiredLRPUpdate) (*models.DesiredLRP, error) {
	logger = logger.Session("db-update-desired-lrp", lager.Data{"process_guid": processGuid})
	logger.Info("starting")
	defer logger.Info("complete")

	var beforeDesiredLRP *models.DesiredLRP
	err := db.transact(ctx, logger, func(logger lager.Logger, tx helpers.Tx) error {
		var err error
		row := db.one(ctx, logger, tx, desiredLRPsTable,
			desiredLRPColumns, helpers.LockRow,
			"process_guid = ?", processGuid,
		)
		beforeDesiredLRP, err = db.fetchDesiredLRP(ctx, logger, row, tx)

		if err != nil {
			logger.Error("failed-lock-desired", err)
			return err
		}

		updateAttributes := helpers.SQLAttributes{"modification_tag_index": beforeDesiredLRP.ModificationTag.Index + 1}

		if update.AnnotationExists() {
			updateAttributes["annotation"] = update.GetAnnotation()
		}

		if update.InstancesExists() {
			updateAttributes["instances"] = update.GetInstances()
		}

		if update.Routes != nil {
			encodedData, err := db.encodeRouteData(logger, update.Routes)
			if err != nil {
				return err
			}
			updateAttributes["routes"] = encodedData
		}

		_, err = db.update(ctx, logger, tx, desiredLRPsTable, updateAttributes, `process_guid = ?`, processGuid)
		if err != nil {
			logger.Error("failed-executing-query", err)
			return err
		}

		return nil
	})

	return beforeDesiredLRP, err
}

func (db *SQLDB) encodeRouteData(logger lager.Logger, routes *models.Routes) ([]byte, error) {
	routeData, err := json.Marshal(routes)
	if err != nil {
		logger.Error("failed-marshalling-routes", err)
		return nil, models.ErrBadRequest
	}
	encodedData, err := db.encoder.Encode(routeData)
	if err != nil {
		logger.Error("failed-encrypting-routes", err)
		return nil, models.ErrBadRequest
	}
	return encodedData, nil
}

func (db *SQLDB) RemoveDesiredLRP(ctx context.Context, logger lager.Logger, processGuid string) error {
	logger = logger.Session("db-remove-desired-lrp", lager.Data{"process_guid": processGuid})
	logger.Info("starting")
	defer logger.Info("complete")

	return db.transact(ctx, logger, func(logger lager.Logger, tx helpers.Tx) error {
		err := db.lockDesiredLRPByGuidForUpdate(ctx, logger, processGuid, tx)
		if err != nil {
			logger.Error("failed-lock-desired", err)
			return err
		}

		_, err = db.delete(ctx, logger, tx, desiredLRPsTable, "process_guid = ?", processGuid)
		if err != nil {
			logger.Error("failed-deleting-from-db", err)
			return err
		}

		return nil
	})
}

// "rows" needs to have the columns defined in the schedulingInfoColumns constant
func (db *SQLDB) fetchDesiredLRPSchedulingInfoAndMore(logger lager.Logger, scanner helpers.RowScanner, dest ...interface{}) (*models.DesiredLRPSchedulingInfo, error) {
	schedulingInfo := &models.DesiredLRPSchedulingInfo{}
	var routeData, volumePlacementData, placementTagData []byte
	values := []interface{}{
		&schedulingInfo.ProcessGuid,
		&schedulingInfo.Domain,
		&schedulingInfo.LogGuid,
		&schedulingInfo.Annotation,
		&schedulingInfo.Instances,
		&schedulingInfo.MemoryMb,
		&schedulingInfo.DiskMb,
		&schedulingInfo.MaxPids,
		&schedulingInfo.RootFs,
		&routeData,
		&volumePlacementData,
		&schedulingInfo.ModificationTag.Epoch,
		&schedulingInfo.ModificationTag.Index,
		&placementTagData,
	}
	values = append(values, dest...)

	err := scanner.Scan(values...)
	if err == sql.ErrNoRows {
		return nil, err
	}

	if err != nil {
		logger.Error("failed-scanning", err)
		return nil, err
	}

	var routes models.Routes
	encodedData, err := db.encoder.Decode(routeData)
	if err != nil {
		logger.Error("failed-decrypting-routes", err)
		return nil, err
	}
	err = json.Unmarshal(encodedData, &routes)
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
	if placementTagData != nil {
		err = json.Unmarshal(placementTagData, &schedulingInfo.PlacementTags)
		if err != nil {
			logger.Error("failed-parsing-placement-tags", err)
			return nil, err
		}
	}

	return schedulingInfo, nil
}

func (db *SQLDB) lockDesiredLRPByGuidForUpdate(ctx context.Context, logger lager.Logger, processGuid string, tx helpers.Tx) error {
	row := db.one(ctx, logger, tx, desiredLRPsTable,
		helpers.ColumnList{"1"}, helpers.LockRow,
		"process_guid = ?", processGuid,
	)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return err
	}
	return nil
}

func (db *SQLDB) fetchDesiredLRPs(ctx context.Context, logger lager.Logger, rows *sql.Rows, queryable helpers.Queryable) ([]*models.DesiredLRP, error) {
	guids := []string{}
	lrps := []*models.DesiredLRP{}
	for rows.Next() {
		lrp, guid, err := db.fetchDesiredLRPInternal(logger, rows)
		if err == models.ErrDeserialize {
			guids = append(guids, guid)
		}
		if err != nil {
			logger.Error("failed-reading-row", err)
			continue
		}
		lrps = append(lrps, lrp)
	}

	if len(guids) > 0 {
		db.deleteInvalidLRPs(ctx, logger, queryable, guids...)
	}

	if err := rows.Err(); err != nil {
		return lrps, err
	}

	return lrps, nil
}

func (db *SQLDB) fetchDesiredLRP(ctx context.Context, logger lager.Logger, scanner helpers.RowScanner, queryable helpers.Queryable) (*models.DesiredLRP, error) {
	lrp, guid, err := db.fetchDesiredLRPInternal(logger, scanner)
	if err == models.ErrDeserialize {
		db.deleteInvalidLRPs(ctx, logger, queryable, guid)
	}
	return lrp, err
}

func (db *SQLDB) fetchDesiredLRPInternal(logger lager.Logger, scanner helpers.RowScanner) (*models.DesiredLRP, string, error) {
	var runInfoData []byte
	schedulingInfo, err := db.fetchDesiredLRPSchedulingInfoAndMore(logger, scanner, &runInfoData)
	if err != nil {
		return nil, "", err
	}

	var runInfo models.DesiredLRPRunInfo
	err = db.deserializeModel(logger, runInfoData, &runInfo)
	if err != nil {
		return nil, schedulingInfo.ProcessGuid, models.ErrDeserialize
	}
	// dedup the ports
	runInfo.Ports = dedupSlice(runInfo.Ports)
	desiredLRP := models.NewDesiredLRP(*schedulingInfo, runInfo)
	return &desiredLRP, "", nil
}

func (db *SQLDB) deleteInvalidLRPs(ctx context.Context, logger lager.Logger, queryable helpers.Queryable, guids ...string) error {
	for _, guid := range guids {
		logger.Info("deleting-invalid-desired-lrp-from-db", lager.Data{"guid": guid})
		_, err := db.delete(ctx, logger, queryable, desiredLRPsTable, "process_guid = ?", guid)
		if err != nil {
			logger.Error("failed-deleting-invalid-row", err)
			return err
		}
	}
	return nil
}

func (db *SQLDB) fetchDesiredLRPSchedulingInfo(logger lager.Logger, scanner helpers.RowScanner) (*models.DesiredLRPSchedulingInfo, error) {
	return db.fetchDesiredLRPSchedulingInfoAndMore(logger, scanner)
}

func whereClauseForProcessGuids(filter []string) string {
	var questionMarks []string

	where := "process_guid IN ("
	for range filter {
		questionMarks = append(questionMarks, "?")
	}

	where += strings.Join(questionMarks, ", ")
	return where + ")"
}

func dedupSlice(ints []uint32) []uint32 {
	if ints == nil {
		// this is really here to make some tests happy, otherwise we replace the
		// nil with an empty slice and they barf
		return nil
	}

	set := make(map[uint32]struct{})
	for _, i := range ints {
		set[i] = struct{}{}
	}
	if len(ints) == len(set) {
		// short circuit the copying if the set has the same number of elements as
		// the slice
		return ints
	}

	newIs := make([]uint32, 0, len(ints))
	for i := range set {
		newIs = append(newIs, i)
	}
	return newIs
}
