package sqldb

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

func (db *SQLDB) CreateLRPDeployment(logger lager.Logger, lrp *models.LRPDeploymentCreation) (*models.LRPDeployment, error) {
	logger = logger.WithData(lager.Data{"process_guid": lrp.ProcessGuid})
	logger.Info("starting")
	defer logger.Info("complete")

	guid, err := db.guidProvider.NextGUID()
	if err != nil {
		logger.Error("failed-to-generate-guid", err)
		return nil, models.ErrGUIDGeneration
	}
	modificationTag := &models.ModificationTag{Epoch: guid, Index: 0}
	lrpDeployment := lrp.LRPDeployment(modificationTag)

	definition := lrp.Definition
	err = db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		routesData, err := db.encodeRouteData(logger, lrp.Routes)
		if err != nil {
			logger.Error("failed-encoding-route-data", err)
			return err
		}

		runInfo := lrp.DesiredLRPRunInfo(db.clock.Now())

		runInfoData, err := db.serializeModel(logger, &runInfo)
		if err != nil {
			logger.Error("failed-to-serialize-model", err)
			return err
		}

		volumePlacement := &models.VolumePlacement{}
		volumePlacement.DriverNames = []string{}
		for _, mount := range definition.VolumeMounts {
			volumePlacement.DriverNames = append(volumePlacement.DriverNames, mount.Driver)
		}

		volumePlacementData, err := db.serializeModel(logger, volumePlacement)
		if err != nil {
			logger.Error("failed-to-serialize-model", err)
			return err
		}

		placementTagData, err := json.Marshal(definition.PlacementTags)
		if err != nil {
			logger.Error("failed-to-serialize-model", err)
			return err
		}

		logger.Info("================going-to-save-run-info", lager.Data{"ProcessGuid": lrp.ProcessGuid, "definition-id": lrp.DefinitionId, "rootfs": definition.RootFs, "run-info": runInfoData})

		// _, err = db.insert(logger, tx, lrpDefinitionsTable, helpers.SQLAttributes{
		// 	"process_guid":     lrp.ProcessGuid,
		// 	"definition_guid":  lrp.DefinitionId,
		// 	"log_guid":         definition.LogGuid,
		// 	"memory_mb":        definition.MemoryMb,
		// 	"disk_mb":          definition.DiskMb,
		// 	"rootfs":           definition.RootFs,
		// 	"volume_placement": volumePlacementData,
		// 	"placement_tags":   placementTagData,
		// 	"max_pids":         definition.MaxPids,
		// 	"run_info":         runInfoData,
		// })
		// if err != nil {
		// 	logger.Error("failed-inserting-lrp-definition", err)
		// 	return err
		// }

		_, err = db.insert(logger, tx, lrpDeploymentsTable,
			helpers.SQLAttributes{
				"process_guid":           lrp.ProcessGuid,
				"domain":                 lrp.Domain,
				"annotation":             lrp.Annotation,
				"instances":              lrp.Instances,
				"modification_tag_epoch": modificationTag.Epoch,
				"modification_tag_index": modificationTag.Index,
				"routes":                 routesData,
				"active":                 true,
				"definition_guid":        lrp.DefinitionId,
				"log_guid":               definition.LogGuid,
				"memory_mb":              definition.MemoryMb,
				"disk_mb":                definition.DiskMb,
				"rootfs":                 definition.RootFs,
				"volume_placement":       volumePlacementData,
				"placement_tags":         placementTagData,
				"max_pids":               definition.MaxPids,
				"run_info":               runInfoData,
			},
		)
		if err != nil {
			logger.Error("failed-inserting-lrp-deployment", err)
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return lrpDeployment, nil
}

func (db *SQLDB) UpdateLRPDeployment(logger lager.Logger, id string, update *models.LRPDeploymentUpdate) (*models.LRPDeployment, error) {

	// - update the list of definition ids to include the new one we wrote
	// - update the deployment with new active definition id
	// Currrently always operating on the assumption there is a new defintion, therefore always inserting into the definitions table

	logger = logger.WithData(lager.Data{"process_guid": id})
	logger.Info("starting")
	defer logger.Info("complete")

	var updatedLRPDeployment *models.LRPDeployment

	err := db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		//TODO: what happens if the lrp_deployment row that is active is not the healthy one
		wheresClause := "lrp_deployments.process_guid = ? AND lrp_deployments.active = ?"
		row := db.one(logger, tx, lrpDeploymentsTable, lrpDeploymentColumns, false, wheresClause, id, true)
		desiredLRPDeployment, err := db.fetchLRPDeployment(logger, row)
		if err != nil {
			logger.Error("failed-to-get-lrp-deployment", err)
			return err
		}

		desiredLRPDeployment.ModificationTag.Increment()
		updatedLRPDeployment, err = db.updateDeploymentAndInsertDefinition(logger, tx, id, desiredLRPDeployment, update)

		return err
	})

	return updatedLRPDeployment, err
}

func (db *SQLDB) updateDeploymentAndInsertDefinition(
	logger lager.Logger,
	tx *sql.Tx,
	processGuid string,
	desiredLRPDeployment *models.LRPDeployment,
	update *models.LRPDeploymentUpdate,
) (*models.LRPDeployment, error) {

	updatedLRPDeployment := *desiredLRPDeployment
	definition := update.Definition
	lrpDeploymentAttrs := helpers.SQLAttributes{
		"modification_tag_index": desiredLRPDeployment.ModificationTag.Index,
	}

	var routesToEncode *models.Routes
	if update.Routes != nil {
		routesToEncode = update.Routes
	} else {
		routesToEncode = desiredLRPDeployment.Routes
	}
	routesData, err := db.encodeRouteData(logger, routesToEncode)
	if err != nil {
		logger.Error("failed-encoding-route-data", err)
		return nil, err
	}
	lrpDeploymentAttrs["routes"] = routesData
	updatedLRPDeployment.Routes = update.Routes

	if update.Instances != nil {
		lrpDeploymentAttrs["instances"] = *update.Instances
		updatedLRPDeployment.Instances = *update.Instances
	}

	if update.Annotation != nil {
		lrpDeploymentAttrs["annotation"] = *update.Annotation
		updatedLRPDeployment.Annotation = *update.Annotation
	}

	if definition != nil {
		lrpKey := models.NewDesiredLRPKey(definition.DefinitionId, desiredLRPDeployment.Domain, definition.LogGuid)
		runInfo := definition.DesiredLRPRunInfo(lrpKey, db.clock.Now())

		runInfoData, err := db.serializeModel(logger, &runInfo)
		if err != nil {
			logger.Error("failed-to-serialize-model", err)
			return nil, err
		}

		volumePlacement := &models.VolumePlacement{}
		volumePlacement.DriverNames = []string{}
		for _, mount := range definition.VolumeMounts {
			volumePlacement.DriverNames = append(volumePlacement.DriverNames, mount.Driver)
		}

		volumePlacementData, err := db.serializeModel(logger, volumePlacement)
		if err != nil {
			logger.Error("failed-to-serialize-model", err)
			return nil, err
		}

		placementTagData, err := json.Marshal(definition.PlacementTags)
		if err != nil {
			logger.Error("failed-to-serialize-model", err)
			return nil, err
		}

		_, err = db.insert(logger, tx, lrpDeploymentsTable, helpers.SQLAttributes{
			"process_guid":           processGuid,
			"definition_guid":        update.DefinitionId,
			"log_guid":               definition.LogGuid,
			"memory_mb":              definition.MemoryMb,
			"disk_mb":                definition.DiskMb,
			"rootfs":                 definition.RootFs,
			"volume_placement":       volumePlacementData,
			"placement_tags":         placementTagData,
			"max_pids":               definition.MaxPids,
			"run_info":               runInfoData,
			"domain":                 updatedLRPDeployment.Domain,
			"annotation":             updatedLRPDeployment.Annotation,
			"instances":              updatedLRPDeployment.Instances,
			"modification_tag_epoch": updatedLRPDeployment.ModificationTag.Epoch,
			"modification_tag_index": updatedLRPDeployment.ModificationTag.Index,
			"routes":                 routesData,
			"active":                 true,
		})
		if err != nil {
			logger.Error("failed-inserting-lrp-definition", err)
			return nil, err
		}

		_, err = db.update(logger, tx, lrpDeploymentsTable, helpers.SQLAttributes{"active": false},
			"lrp_deployments.definition_guid = ?", desiredLRPDeployment.ActiveDefinitionId)

		if err != nil {
			logger.Error("failed-updating-lrp-deployment", err)
			return nil, err
		}

		// // Save new active_definition_id
		// lrpDeploymentAttrs["active_definition_id"] = *update.DefinitionId
		// updatedLRPDeployment.ActiveDefinitionId = *update.DefinitionId
	} else {

		wheresClause := " WHERE process_guid = ? AND active = ?"
		_, err = db.update(logger, tx, lrpDeploymentsTable,
			lrpDeploymentAttrs, wheresClause, updatedLRPDeployment.ProcessGuid, true,
		)
		if err != nil {
			logger.Error("failed-updating-lrp-deployments", err)
			return nil, err
		}
	}

	wheresClause := " WHERE process_guid = ?"
	values := []interface{}{desiredLRPDeployment.ProcessGuid}
	definitionRows, err := db.selectDefinitions(logger, tx, lrpDefinitionsColumns, wheresClause, values)
	if err != nil {
		logger.Error("failed-selecting-lrp-definitions", err)
		return nil, err
	}

	definitions, err := db.fetchLRPDefinitions(logger, definitionRows)
	if err != nil {
		logger.Error("failed-fetching-lrp-definitions", err)
		return nil, err
	}

	updatedLRPDeployment.Definitions = definitions

	return &updatedLRPDeployment, nil
}

func (db *SQLDB) SaveLRPDeployment(logger lager.Logger, lrpDeployment *models.LRPDeployment) (*models.LRPDeployment, error) {
	wheresClause := "lrp_deployments.process_guid = ?"
	_, err := db.update(logger, db.db, lrpDeploymentsTable,
		helpers.SQLAttributes{
			"healthy_definition_id": lrpDeployment.HealthyDefinitionId,
		}, wheresClause, lrpDeployment.ProcessGuid,
	)
	return lrpDeployment, err
}

func (db *SQLDB) DeleteLRPDeployment(logger lager.Logger, id string) (*models.LRPDeployment, error) {
	logger = logger.WithData(lager.Data{"process_guid": id})
	logger.Info("starting")
	defer logger.Info("complete")

	var lrpDeployment *models.LRPDeployment

	return lrpDeployment, db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		var err error
		lrpDeployment, err = db.findLRPDeployment(logger, tx, "process_guid = ?", id)

		_, err = db.delete(logger, tx, lrpDefinitionsTable, "process_guid = ?", id)
		if err != nil {
			logger.Error("failed-deleting-lrp-definitions-from-db", err)
			return err
		}
		_, err = db.delete(logger, tx, lrpDeploymentsTable, "process_guid = ?", id)
		if err != nil {
			logger.Error("failed-deleting-lrp-deployment-from-db", err)
			return err
		}

		return nil
	})
}

func (db *SQLDB) ActivateLRPDeploymentDefinition(logger lager.Logger, id, definitionId string) (*models.LRPDeployment, error) {
	var lrpDeployment *models.LRPDeployment
	return lrpDeployment, db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		var err error
		lrpDeployment, err = db.findLRPDeployment(logger, tx, "process_guid = ?", id)
		_, err = db.update(logger, db.db, lrpDeploymentsTable,
			helpers.SQLAttributes{
				"active_definition_id": definitionId,
			}, "lrp_deployments.process_guid = ?", id,
		)
		if err != nil {
			logger.Error("failed-activating-definition-id", err)
			return err
		}
		lrpDeployment.ActiveDefinitionId = definitionId
		return nil
	})
}

func (db *SQLDB) fetchLRPDeployment(logger lager.Logger, row RowScanner) (*models.LRPDeployment, error) {
	lrpDeployment := &models.LRPDeployment{
		ModificationTag: &models.ModificationTag{},
	}
	var routeData []byte
	var active, healthy bool
	var defID string
	values := []interface{}{
		&lrpDeployment.ProcessGuid,
		&lrpDeployment.Domain,
		&lrpDeployment.Instances,
		&lrpDeployment.Annotation,
		&routeData,
		&defID,
		&active,
		&healthy,
		&lrpDeployment.ModificationTag.Epoch,
		&lrpDeployment.ModificationTag.Index,
	}

	err := row.Scan(values...)
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
	lrpDeployment.Routes = &routes

	if active {
		lrpDeployment.ActiveDefinitionId = defID
	}

	if healthy {
		lrpDeployment.HealthyDefinitionId = defID
	}
	return lrpDeployment, nil
}

func (db *SQLDB) fetchLRPDefinitionsInternal(logger lager.Logger, scanner RowScanner) (*models.LRPDefinition, string, error) {
	definition := &models.LRPDefinition{}
	var placementData, runInfoData []byte
	values := []interface{}{
		&definition.DefinitionId,
		&definition.LogGuid,
		&definition.MemoryMb,
		&definition.DiskMb,
		&definition.MaxPids,
		&definition.RootFs,
		&placementData,
		&runInfoData,
	}

	err := scanner.Scan(values...)
	if err == sql.ErrNoRows {
		return nil, "", err
	}
	if err != nil {
		logger.Error("failed-scanning", err)
		return nil, "", err
	}

	if placementData != nil {
		err = json.Unmarshal(placementData, &definition.PlacementTags)
		if err != nil {
			logger.Error("failed-parsing-placement-tags", err)
			return nil, "", err
		}
	}

	var runInfo models.DesiredLRPRunInfo
	err = db.deserializeModel(logger, runInfoData, &runInfo)
	if err != nil {
		return nil, "", models.ErrDeserialize
	}

	environmentVariables := make([]*models.EnvironmentVariable, len(runInfo.EnvironmentVariables))
	for i := range runInfo.EnvironmentVariables {
		environmentVariables[i] = &runInfo.EnvironmentVariables[i]
	}

	egressRules := make([]*models.SecurityGroupRule, len(runInfo.EgressRules))
	for i := range runInfo.EgressRules {
		egressRules[i] = &runInfo.EgressRules[i]
	}

	definition.EnvironmentVariables = environmentVariables
	definition.CachedDependencies = runInfo.CachedDependencies
	definition.Setup = runInfo.Setup
	definition.Action = runInfo.Action
	definition.Monitor = runInfo.Monitor
	definition.StartTimeoutMs = runInfo.StartTimeoutMs
	definition.Privileged = runInfo.Privileged
	definition.CpuWeight = runInfo.CpuWeight
	definition.Ports = runInfo.Ports
	definition.EgressRules = egressRules
	definition.LogSource = runInfo.LogSource
	definition.MetricsGuid = runInfo.MetricsGuid
	definition.LegacyDownloadUser = runInfo.LegacyDownloadUser
	definition.TrustedSystemCertificatesPath = runInfo.TrustedSystemCertificatesPath
	definition.VolumeMounts = runInfo.VolumeMounts
	definition.Network = runInfo.Network
	definition.CertificateProperties = runInfo.CertificateProperties
	definition.ImageUsername = runInfo.ImageUsername
	definition.ImagePassword = runInfo.ImagePassword
	definition.CheckDefinition = runInfo.CheckDefinition

	return definition, definition.DefinitionId, nil
}

func (db *SQLDB) fetchLRPDefinitions(logger lager.Logger, rows *sql.Rows) (map[string]*models.LRPDefinition, error) {
	definitions := map[string]*models.LRPDefinition{}
	for rows.Next() {
		def, guid, err := db.fetchLRPDefinitionsInternal(logger, rows)
		if err != nil {
			logger.Error("failed-reading-row", err)
			continue
		}
		definitions[guid] = def
	}
	// if len(guids) > 0 {
	// 	db.deleteInvalidLRPs(logger, queryable, guids...)
	// }

	if err := rows.Err(); err != nil {
		return definitions, err
	}

	return definitions, nil
}

func (db *SQLDB) LRPDeployments(logger lager.Logger, deploymentIds []string) ([]*models.LRPDeployment, error) {
	logger.Debug("starting")
	defer logger.Debug("complete")

	var deployments []*models.LRPDeployment

	err := db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		var err error
		var wheresClause string
		values := []interface{}{}

		whereClauseForIds := func(filter []string) string {
			var questionMarks []string

			where := "process_guid IN ("
			for range filter {
				questionMarks = append(questionMarks, "?")

			}

			where += strings.Join(questionMarks, ", ")
			return where + ")"
		}

		if len(deploymentIds) > 0 {
			wheresClause = whereClauseForIds(deploymentIds)
			for _, guid := range deploymentIds {
				values = append(values, guid)
			}
		}

		rows, err := db.all(logger, tx, lrpDeploymentsTable, lrpDeploymentColumns, helpers.NoLockRow, wheresClause, values...)
		if err != nil {
			logger.Error("failed-to-fetch-deployments", err)
			return err
		}

		if rows != nil {
			for rows.Next() {
				lrpDeployment, err := db.fetchLRPDeployment(logger, rows)
				if err != nil {
					logger.Error("failed-to-fetch-deployment", err)
					return err
				}

				if lrpDeployment != nil {
					deployments = append(deployments, lrpDeployment)
				}
			}
		} else {
			return nil
		}

		for _, dep := range deployments {
			wheresClause := " WHERE process_guid = ?"
			values = []interface{}{dep.ProcessGuid}
			definitionRows, err := db.selectDefinitions(logger, tx, lrpDefinitionsColumns, wheresClause, values)
			if err != nil {
				logger.Error("failed-selecting-lrp-definitions", err)
				return err
			}

			definitions, err := db.fetchLRPDefinitions(logger, definitionRows)
			if err != nil {
				logger.Error("failed-fetching-lrp-definitions", err)
				return err
			}
			dep.Definitions = definitions
		}

		return err
	})

	return deployments, err
}

func (db *SQLDB) LRPDeploymentByDefinitionGuid(logger lager.Logger, id string) (*models.LRPDeployment, error) {
	logger = logger.WithData(lager.Data{"definition_guid": id})
	logger.Info("starting")
	defer logger.Info("complete")

	var lrpDeployment *models.LRPDeployment

	err := db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		var err error

		lrpDeployment, err = db.findLRPDeployment(logger, tx, "definition_guid = ?", id)
		return err
	})

	return lrpDeployment, err
	// err := db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
	// 	var err error
	// 	wheresClause := " WHERE lrp_deployments.definition_guid = ?"
	// 	values := []interface{}{id}
	// 	//TODO: now using QueryRow which doesn't return an error. How do we check for errors?
	// 	row := db.oneLRPDeploymentWithDefinitions(logger, tx, lrpDeploymentColumns, wheresClause, values)
	// 	if row != nil {
	// 		lrpDeployment, err = db.fetchLRPDeployment(logger, row)
	// 		if err != nil {
	// 			return err
	// 		}
	// 	} else {
	// 		return helpers.ErrResourceNotFound
	// 	}

	// 	wheresClause = " WHERE process_guid = ?"
	// 	values = []interface{}{lrpDeployment.ProcessGuid}
	// 	definitionRows, err := db.selectDefinitions(logger, tx, lrpDefinitionsColumns, wheresClause, values)
	// 	if err != nil {
	// 		logger.Error("failed-selecting-lrp-definitions", err)
	// 		return err
	// 	}

	// 	definitions, err := db.fetchLRPDefinitions(logger, definitionRows)
	// 	if err != nil {
	// 		logger.Error("failed-fetching-lrp-definitions", err)
	// 		return err
	// 	}

	// 	lrpDeployment.Definitions = definitions

	// 	return err
	// })

	// return lrpDeployment, err
}

func (db *SQLDB) LRPDeploymentByProcessGuid(logger lager.Logger, id string) (*models.LRPDeployment, error) {
	logger = logger.WithData(lager.Data{"process_guid": id})
	logger.Debug("starting")
	defer logger.Debug("complete")

	var lrpDeployment *models.LRPDeployment

	err := db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		var err error

		lrpDeployment, err = db.findLRPDeployment(logger, tx, "process_guid = ?", id)
		return err
	})

	return lrpDeployment, err
}

func (db *SQLDB) LRPDeploymentSchedulingInfo(logger lager.Logger, filter models.LRPDeploymentFilter) ([]*models.LRPDeploymentSchedulingInfo, error) {
	logger = logger.WithData(lager.Data{"filter": filter})
	logger.Info("start")
	defer logger.Info("complete")

	var wheres []string
	var values []interface{}
	var wheresClause string

	if len(filter.DefinitionIds) > 0 {
		wheres = append(wheres, whereClauseForDefinitionGuids(filter.DefinitionIds))

		for _, guid := range filter.DefinitionIds {
			values = append(values, guid)
		}
	}

	if len(wheres) != 0 {
		wheresClause = " WHERE "
		wheresClause = wheresClause + strings.Join(wheres, " AND ")
	}
	results := map[string]*models.LRPDeploymentSchedulingInfo{}
	err := db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		rows, err := db.selectDefinitions(logger, tx, schedulingInfoColumns, wheresClause, values)
		if err != nil {
			logger.Error("failed-query", err)
			return err
		}
		defer rows.Close()

		for rows.Next() {
			col, _ := rows.Columns()
			fmt.Printf(">>>>>>>>>>%#v", col)
			deployment, definition, err := db.fetchSchedulingInfo(logger, rows)
			if err != nil {
				logger.Error("failed-reading-row", err)
				continue
			}
			if foundDeployment, ok := results[deployment.ProcessGuid]; ok {
				foundDeployment.Definitions[definition.DefinitionId] = definition
			} else {
				deployment.Definitions = map[string]*models.LRPDefinitionSchedulingInfo{
					definition.DefinitionId: definition,
				}
				results[deployment.ProcessGuid] = deployment
			}
		}

		if rows.Err() != nil {
			logger.Error("failed-fetching-row", rows.Err())
			return db.convertSQLError(rows.Err())
		}

		return nil
	})

	deploymentsSchedulingInfo := []*models.LRPDeploymentSchedulingInfo{}
	for _, deploymentSchedulingInfo := range results {
		deploymentsSchedulingInfo = append(deploymentsSchedulingInfo, deploymentSchedulingInfo)
	}

	return deploymentsSchedulingInfo, err
}

func (db *SQLDB) findLRPDeployment(logger lager.Logger, q Queryable, where, id string) (*models.LRPDeployment, error) {
	var lrpDeployment *models.LRPDeployment
	var err error

	values := []interface{}{id}
	row := db.one(logger, q, lrpDeploymentsTable, lrpDeploymentColumns, false, where, id)
	if row != nil {
		lrpDeployment, err = db.fetchLRPDeployment(logger, row)
		if err != nil {
			logger.Error("failed-fetching-lrp-deployment", err)
			return nil, err
		}
	} else {
		return nil, helpers.ErrResourceNotFound
	}

	wheresClause := " WHERE process_guid = ?"
	values = []interface{}{lrpDeployment.ProcessGuid}
	definitionRows, err := db.selectDefinitions(logger, q, lrpDefinitionsColumns, wheresClause, values)
	if err != nil {
		logger.Error("failed-selecting-lrp-definitions", err)
		return nil, err
	}

	definitions, err := db.fetchLRPDefinitions(logger, definitionRows)
	if err != nil {
		logger.Error("failed-fetching-lrp-definitions", err)
		return nil, err
	}

	lrpDeployment.Definitions = definitions
	return lrpDeployment, nil
}

func whereClauseForDefinitionGuids(filter []string) string {
	var questionMarks []string

	where := "lrp_deployments.definition_guid IN ("
	for range filter {
		questionMarks = append(questionMarks, "?")

	}

	where += strings.Join(questionMarks, ", ")
	return where + ")"
}

func (db *SQLDB) fetchSchedulingInfo(logger lager.Logger, scanner RowScanner, dest ...interface{}) (*models.LRPDeploymentSchedulingInfo, *models.LRPDefinitionSchedulingInfo, error) {
	deployment := &models.LRPDeploymentSchedulingInfo{}
	definition := &models.LRPDefinitionSchedulingInfo{}
	var routeData, volumePlacementData, placementTagData []byte
	values := []interface{}{
		&deployment.ProcessGuid,
		&definition.DefinitionId,
		&deployment.Domain,
		&deployment.Instances,
		&deployment.Annotation,
		&routeData,
		&deployment.ModificationTag.Epoch,
		&deployment.ModificationTag.Index,
		&definition.LogGuid,
		&definition.MemoryMb,
		&definition.DiskMb,
		&definition.MaxPids,
		&definition.RootFs,
		&volumePlacementData,
		&placementTagData,
	}
	values = append(values, dest...)

	err := scanner.Scan(values...)
	if err == sql.ErrNoRows {
		return nil, nil, err
	}

	if err != nil {
		logger.Error("failed-scanning", err)
		return nil, nil, err
	}

	var routes models.Routes
	encodedData, err := db.encoder.Decode(routeData)
	if err != nil {
		logger.Error("failed-decrypting-routes", err)
		return nil, nil, err
	}
	err = json.Unmarshal(encodedData, &routes)
	if err != nil {
		logger.Error("failed-parsing-routes", err)
		return nil, nil, err
	}
	deployment.Routes = &routes

	var volumePlacement models.VolumePlacement
	err = db.deserializeModel(logger, volumePlacementData, &volumePlacement)
	if err != nil {
		logger.Error("failed-parsing-volume-placement", err)
		return nil, nil, err
	}
	definition.VolumePlacement = &volumePlacement
	if placementTagData != nil {
		err = json.Unmarshal(placementTagData, &definition.PlacementTags)
		if err != nil {
			logger.Error("failed-parsing-placement-tags", err)
			return nil, nil, err
		}
	}

	return deployment, definition, nil
}
