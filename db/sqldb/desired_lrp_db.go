package sqldb

import (
	"database/sql"
	"encoding/json"
	"strings"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/format"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

func (db *SQLDB) DesireLRP(logger lager.Logger, desiredLRP *models.DesiredLRP) (string, err) {
	// logger = logger.WithData(lager.Data{"process_guid": desiredLRP.ProcessGuid})
	// logger.Info("starting")
	// defer logger.Info("complete")

	// return db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
	// 	routesData, err := db.encodeRouteData(logger, desiredLRP.Routes)
	// 	if err != nil {
	// 		logger.Error("failed-encoding-route-data", err)
	// 		return err
	// 	}

	// 	runInfo := desiredLRP.DesiredLRPRunInfo(db.clock.Now())

	// 	runInfoData, err := db.serializeModel(logger, &runInfo)
	// 	if err != nil {
	// 		logger.Error("failed-to-serialize-model", err)
	// 		return err
	// 	}

	// 	volumePlacement := &models.VolumePlacement{}
	// 	volumePlacement.DriverNames = []string{}
	// 	for _, mount := range desiredLRP.VolumeMounts {
	// 		volumePlacement.DriverNames = append(volumePlacement.DriverNames, mount.Driver)
	// 	}

	// 	volumePlacementData, err := db.serializeModel(logger, volumePlacement)
	// 	if err != nil {
	// 		logger.Error("failed-to-serialize-model", err)
	// 		return err
	// 	}

	guid, err := db.guidProvider.NextGUID()
	if err != nil {
		logger.Error("failed-to-generate-guid", err)
		return "", models.ErrGUIDGeneration
	}

	// 	placementTagData, err := json.Marshal(desiredLRP.PlacementTags)
	// 	if err != nil {
	// 		logger.Error("failed-to-serialize-model", err)
	// 		return err
	// 	}

	// 	desiredLRP.ModificationTag = &models.ModificationTag{Epoch: guid, Index: 0}

	// 	_, err = db.insert(logger, tx, desiredLRPsTable,
	// 		helpers.SQLAttributes{
	// 			"process_guid":           desiredLRP.ProcessGuid,
	// 			"domain":                 desiredLRP.Domain,
	// 			"log_guid":               desiredLRP.LogGuid,
	// 			"annotation":             desiredLRP.Annotation,
	// 			"instances":              desiredLRP.Instances,
	// 			"memory_mb":              desiredLRP.MemoryMb,
	// 			"disk_mb":                desiredLRP.DiskMb,
	// 			"max_pids":               desiredLRP.MaxPids,
	// 			"rootfs":                 desiredLRP.RootFs,
	// 			"volume_placement":       volumePlacementData,
	// 			"modification_tag_epoch": desiredLRP.ModificationTag.Epoch,
	// 			"modification_tag_index": desiredLRP.ModificationTag.Index,
	// 			"routes":                 routesData,
	// 			"run_info":               runInfoData,
	// 			"placement_tags":         placementTagData,
	// 		},
	// 	)
	// 	if err != nil {
	// 		logger.Error("failed-inserting-desired", err)
	// 		return err
	// 	}
	// 	return nil
	// })

	lrp := &models.LRPDeploymentCreation{
		ProcessGuid:  guid,
		Domain:       desiredLRP.Domain,
		Instances:    desiredLRP.Instances,
		Annotation:   desiredLRP.Annotation,
		Routes:       desiredLRP.Routes,
		DefinitionId: desiredLRP.ProcessGuid,
		Definition: &models.LRPDefinition{
			DefinitionId:                  desiredLRP.ProcessGuid,
			RootFs:                        desiredLRP.RootFs,
			EnvironmentVariables:          desiredLRP.EnvironmentVariables,
			Setup:                         desiredLRP.Setup,
			Action:                        desiredLRP.Action,
			StartTimeoutMs:                desiredLRP.StartTimeoutMs,
			DeprecatedStartTimeoutS:       desiredLRP.DeprecatedStartTimeoutS,
			Monitor:                       desiredLRP.Monitor,
			DiskMb:                        desiredLRP.DiskMb,
			MemoryMb:                      desiredLRP.MemoryMb,
			CpuWeight:                     desiredLRP.CpuWeight,
			Privileged:                    desiredLRP.Privileged,
			Ports:                         desiredLRP.Ports,
			LogSource:                     desiredLRP.LogSource,
			LogGuid:                       desiredLRP.LogGuid,
			MetricsGuid:                   desiredLRP.MetricsGuid,
			EgressRules:                   desiredLRP.EgressRules,
			CachedDependencies:            desiredLRP.CachedDependencies,
			LegacyDownloadUser:            desiredLRP.LegacyDownloadUser,
			TrustedSystemCertificatesPath: desiredLRP.TrustedSystemCertificatesPath,
			VolumeMounts:                  desiredLRP.VolumeMounts,
			Network:                       desiredLRP.Network,
			PlacementTags:                 desiredLRP.PlacementTags,
			MaxPids:                       desiredLRP.MaxPids,
			CertificateProperties:         desiredLRP.CertificateProperties,
			ImageUsername:                 desiredLRP.ImageUsername,
			ImagePassword:                 desiredLRP.ImagePassword,
			CheckDefinition:               desiredLRP.CheckDefinition,
		},
	}
	lrpDeployment, err := db.CreateLRPDeployment(logger, lrp)
	return lrpDeployment.ProcessGuid, err
}

func (db *SQLDB) DesiredLRPByProcessGuid(logger lager.Logger, processGuid string) (*models.DesiredLRP, error) {
	logger = logger.WithData(lager.Data{"process_guid": processGuid})
	logger.Debug("starting")
	defer logger.Debug("complete")

	var desiredLRP *models.DesiredLRP

	err := db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		var err error
		wheresClause := " WHERE lrp_definitions.definition_guid = ?"
		values := []interface{}{processGuid}
		//TODO: now using QueryRow which doesn't return an error. How do we check for errors?
		row := db.oneLRPDeploymentWithDefinitions(logger, tx, desiredLRPColumns, wheresClause, values)

		if row != nil {
			desiredLRP, err = db.fetchDesiredLRP(logger, row, tx)
		} else {
			return helpers.ErrResourceNotFound
		}

		return err
	})

	return desiredLRP, err
}

func (db *SQLDB) DesiredLRPs(logger lager.Logger, filter models.DesiredLRPFilter) ([]*models.DesiredLRP, error) {
	logger = logger.WithData(lager.Data{"filter": filter})
	logger.Debug("start")
	defer logger.Debug("complete")

	var wheres []string
	var values []interface{}
	var wheresClause string
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

	if len(wheres) != 0 {
		wheresClause = " WHERE "
		wheresClause = wheresClause + strings.Join(wheres, " AND ")
	}
	err := db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		rows, err := db.selectLRPDeploymentsWithDefinitions(logger, tx, desiredLRPColumns, wheresClause, values)
		if err != nil {
			logger.Error("failed-query", err)
			return err
		}
		defer rows.Close()

		results, err = db.fetchDesiredLRPs(logger, rows, tx)
		if err != nil {
			logger.Error("failed-fetching-row", rows.Err())
			return db.convertSQLError(rows.Err())
		}

		return nil
	})

	return results, err
}

func (db *SQLDB) DesiredLRPSchedulingInfos(logger lager.Logger, filter models.DesiredLRPFilter) ([]*models.DesiredLRPSchedulingInfo, error) {
	logger = logger.WithData(lager.Data{"filter": filter})
	logger.Debug("start")
	defer logger.Debug("complete")

	var wheres []string
	var values []interface{}
	var wheresClause string

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

	if len(wheres) != 0 {
		wheresClause = " WHERE "
		wheresClause = wheresClause + strings.Join(wheres, " AND ")
	}
	results := []*models.DesiredLRPSchedulingInfo{}
	err := db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		rows, err := db.selectLRPDeploymentsWithDefinitions(logger, tx, schedulingInfoColumns, wheresClause, values)
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

func (db *SQLDB) UpdateDesiredLRP(logger lager.Logger, processGuid string, update *models.DesiredLRPUpdate) (*models.DesiredLRP, error) {
	logger = logger.WithData(lager.Data{"process_guid": processGuid})
	logger.Info("starting")
	defer logger.Info("complete")

	var beforeDesiredLRP *models.DesiredLRP
	err := db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		var err error
		definitionRow := db.one(logger, tx, lrpDefinitionsTable,
			[]string{"process_guid"}, helpers.NoLockRow,
			"definition_guid = ?", processGuid,
		)

		var desiredLRPGuid string
		err = definitionRow.Scan(&desiredLRPGuid)
		if err != nil {
			logger.Error("failed-lock-desired1", err)
			return err
		}
		// row = db.one(logger, tx, lrpDeploymentsTable,
		// 	desiredLRPColumns, helpers.LockRow,
		// 	"process_guid = ?", desiredLRPGuid,
		// )
		// beforeDesiredLRP, err = db.fetchDesiredLRP(logger, row, tx)

		wheresClause := " WHERE lrp_definitions.definition_guid = ?"
		values := []interface{}{processGuid}
		row := db.oneLRPDeploymentWithDefinitions(logger, tx, desiredLRPColumns, wheresClause, values)
		if row != nil {
			beforeDesiredLRP, err = db.fetchDesiredLRP(logger, row, tx)
			if err != nil {
				logger.Error("failed-to-fetch-desired-lrp", err)
				return err
			}
		} else {
			return helpers.ErrResourceNotFound
		}

		updateAttributes := helpers.SQLAttributes{"modification_tag_index": beforeDesiredLRP.ModificationTag.Index + 1}

		if update.Annotation != nil {
			updateAttributes["annotation"] = *update.Annotation
		}

		if update.Instances != nil {
			updateAttributes["instances"] = *update.Instances
		}

		if update.Routes != nil {
			encodedData, err := db.encodeRouteData(logger, update.Routes)
			if err != nil {
				return err
			}
			updateAttributes["routes"] = encodedData
		}

		_, err = db.update(logger, tx, lrpDeploymentsTable, updateAttributes, `process_guid = ?`, desiredLRPGuid)
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
	encodedData, err := db.encoder.Encode(format.BASE64_ENCRYPTED, routeData)
	if err != nil {
		logger.Error("failed-encrypting-routes", err)
		return nil, models.ErrBadRequest
	}
	return encodedData, nil
}

func (db *SQLDB) encodeDefinitionData(logger lager.Logger, definitions map[string]*models.LRPDefinition) ([]byte, error) {
	definitionsData, err := json.Marshal(definitions)
	if err != nil {
		logger.Error("failed-marshalling-definitions", err)
		return nil, models.ErrBadRequest
	}
	encodedData, err := db.encoder.Encode(format.BASE64_ENCRYPTED, definitionsData)
	if err != nil {
		logger.Error("failed-encrypting-definitions", err)
		return nil, models.ErrBadRequest
	}
	return encodedData, nil
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

		_, err = db.delete(logger, tx, lrpDefinitionsTable, "definition_guid = ?", processGuid)
		if err != nil {
			logger.Error("failed-deleting-from-db", err)
			return err
		}

		return nil
	})
}

// "rows" needs to have the columns defined in the schedulingInfoColumns constant
func (db *SQLDB) fetchDesiredLRPSchedulingInfoAndMore(logger lager.Logger, scanner RowScanner, dest ...interface{}) (*models.DesiredLRPSchedulingInfo, error) {
	schedulingInfo := &models.DesiredLRPSchedulingInfo{}
	var routeData, volumePlacementData, placementTagData []byte
	var deploymentProcessGuid string
	values := []interface{}{
		&deploymentProcessGuid,
		&schedulingInfo.ProcessGuid,
		&schedulingInfo.Domain,
		&schedulingInfo.Instances,
		&schedulingInfo.Annotation,
		&routeData,
		&schedulingInfo.ModificationTag.Epoch,
		&schedulingInfo.ModificationTag.Index,
		&schedulingInfo.LogGuid,
		&schedulingInfo.MemoryMb,
		&schedulingInfo.DiskMb,
		&schedulingInfo.MaxPids,
		&schedulingInfo.RootFs,
		&volumePlacementData,
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

func (db *SQLDB) lockDesiredLRPByGuidForUpdate(logger lager.Logger, processGuid string, tx *sql.Tx) error {
	row := db.one(logger, tx, lrpDefinitionsTable,
		helpers.ColumnList{"1"}, helpers.LockRow,
		"definition_guid = ?", processGuid,
	)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return err
	}
	return nil
}

func (db *SQLDB) fetchDesiredLRPs(logger lager.Logger, rows *sql.Rows, queryable Queryable) ([]*models.DesiredLRP, error) {
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
		db.deleteInvalidLRPs(logger, queryable, guids...)
	}

	if err := rows.Err(); err != nil {
		return lrps, err
	}

	return lrps, nil
}

func (db *SQLDB) fetchDesiredLRP(logger lager.Logger, scanner RowScanner, queryable Queryable) (*models.DesiredLRP, error) {
	lrp, guid, err := db.fetchDesiredLRPInternal(logger, scanner)
	if err == models.ErrDeserialize {
		db.deleteInvalidLRPs(logger, queryable, guid)
	}
	return lrp, err
}

func (db *SQLDB) fetchDesiredLRPInternal(logger lager.Logger, scanner RowScanner) (*models.DesiredLRP, string, error) {
	var runInfoData []byte
	schedulingInfo, err := db.fetchDesiredLRPSchedulingInfoAndMore(logger, scanner, &runInfoData)
	if err != nil {
		logger.Error("failed-fetching-run-info", err)
		return nil, "", err
	}

	var runInfo models.DesiredLRPRunInfo
	err = db.deserializeModel(logger, runInfoData, &runInfo)
	if err != nil {
		return nil, schedulingInfo.ProcessGuid, models.ErrDeserialize
	}
	desiredLRP := models.NewDesiredLRP(*schedulingInfo, runInfo)
	return &desiredLRP, "", nil
}

func (db *SQLDB) deleteInvalidLRPs(logger lager.Logger, queryable Queryable, guids ...string) error {
	for _, guid := range guids {
		logger.Info("deleting-invalid-desired-lrp-from-db", lager.Data{"guid": guid})
		_, err := db.delete(logger, queryable, lrpDefinitionsTable, "process_guid = ?", guid)
		if err != nil {
			logger.Error("failed-deleting-invalid-row", err)
			return err
		}
	}
	return nil
}

func (db *SQLDB) fetchDesiredLRPSchedulingInfo(logger lager.Logger, scanner RowScanner) (*models.DesiredLRPSchedulingInfo, error) {
	return db.fetchDesiredLRPSchedulingInfoAndMore(logger, scanner)
}

func whereClauseForProcessGuids(filter []string) string {
	var questionMarks []string

	where := "definition_guid IN ("
	for range filter {
		questionMarks = append(questionMarks, "?")

	}

	where += strings.Join(questionMarks, ", ")
	return where + ")"
}
