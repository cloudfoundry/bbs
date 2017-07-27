package sqldb

import (
	"database/sql"
	"encoding/json"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

func (db *SQLDB) CreateLRPDeployment(logger lager.Logger, lrp *models.LRPDeploymentCreation) (string, error) {
	logger = logger.WithData(lager.Data{"process_guid": lrp.ProcessGuid})
	logger.Info("starting")
	defer logger.Info("complete")

	definition := lrp.Definition

	err := db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
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

		guid, err := db.guidProvider.NextGUID()
		if err != nil {
			logger.Error("failed-to-generate-guid", err)
			return models.ErrGUIDGeneration
		}

		placementTagData, err := json.Marshal(definition.PlacementTags)
		if err != nil {
			logger.Error("failed-to-serialize-model", err)
			return err
		}

		logger.Info("================going-to-save-run-info", lager.Data{"ProcessGuid": lrp.ProcessGuid, "definition-id": lrp.DefinitionId, "rootfs": definition.RootFs, "run-info": runInfoData})

		_, err = db.insert(logger, tx, lrpDefinitionsTable, helpers.SQLAttributes{
			"process_guid":     lrp.ProcessGuid,
			"definition_name":  lrp.ProcessGuid,
			"definition_guid":  lrp.DefinitionId,
			"log_guid":         definition.LogGuid,
			"memory_mb":        definition.MemoryMb,
			"disk_mb":          definition.DiskMb,
			"rootfs":           definition.RootFs,
			"volume_placement": volumePlacementData,
			"placement_tags":   placementTagData,
			"max_pids":         definition.MaxPids,
			"run_info":         runInfoData,
		})
		if err != nil {
			logger.Error("failed-inserting-lrp-definition", err)
			return err
		}

		modificationTag := &models.ModificationTag{Epoch: guid, Index: 0}

		_, err = db.insert(logger, tx, lrpDeploymentsTable,
			helpers.SQLAttributes{
				"process_guid":           lrp.ProcessGuid,
				"domain":                 lrp.Domain,
				"annotation":             lrp.Annotation,
				"instances":              lrp.Instances,
				"modification_tag_epoch": modificationTag.Epoch,
				"modification_tag_index": modificationTag.Index,
				"routes":                 routesData,
				"active_definition_id":   lrp.DefinitionId,
			},
		)
		if err != nil {
			logger.Error("failed-inserting-lrp-deployment", err)
			return err
		}

		return nil
	})

	return lrp.ProcessGuid, err
}

func (db *SQLDB) UpdateLRPDeployment(logger lager.Logger, id string, update *models.LRPDeploymentUpdate) (string, error) {

	// Write a new lrp definition to the DB based on the definition in the updateLRPDeploymentUpdate
	//Update lrp deployment with the new instances, routes, and annotation. Update the modification epoch tag
	// - update the list of definition ids to include the new one we wrote
	// - update the deployment with new active definition id

	logger = logger.WithData(lager.Data{"process_guid": id})
	logger.Info("starting")
	defer logger.Info("complete")

	definition := update.Definition

	err := db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		routesData, err := db.encodeRouteData(logger, update.Routes)
		if err != nil {
			logger.Error("failed-encoding-route-data", err)
			return err
		}

		wheresClause := "lrp_deployments.process_guid = ?"
		row := db.one(logger, tx, lrpDeploymentsTable, lrpDeploymentColumns, false, wheresClause, id)
		desiredLRPDeployment, err := db.fetchLRPDeployment(logger, row)
		if err != nil {
			logger.Error("failed-to-get-desired-lrp", err)
			return err
		}
		lrpKey := models.NewDesiredLRPKey(id, desiredLRPDeployment.Domain, definition.LogGuid)
		runInfo := definition.DesiredLRPRunInfo(lrpKey, db.clock.Now())

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

		guid, err := db.guidProvider.NextGUID()
		if err != nil {
			logger.Error("failed-to-generate-guid", err)
			return models.ErrGUIDGeneration
		}

		placementTagData, err := json.Marshal(definition.PlacementTags)
		if err != nil {
			logger.Error("failed-to-serialize-model", err)
			return err
		}
		modificationTag := &models.ModificationTag{Epoch: guid, Index: 0}

		_, err = db.insert(logger, tx, lrpDefinitionsTable, helpers.SQLAttributes{
			"process_guid":     id,
			"definition_name":  update.DefinitionId,
			"definition_guid":  update.DefinitionId,
			"log_guid":         definition.LogGuid,
			"memory_mb":        definition.MemoryMb,
			"disk_mb":          definition.DiskMb,
			"rootfs":           definition.RootFs,
			"volume_placement": volumePlacementData,
			"placement_tags":   placementTagData,
			"max_pids":         definition.MaxPids,
			"run_info":         runInfoData,
		})
		if err != nil {
			logger.Error("failed-inserting-lrp-definition", err)
			return err
		}

		_, err = db.update(logger, tx, lrpDeploymentsTable,
			helpers.SQLAttributes{
				"annotation":             update.Annotation,
				"instances":              update.Instances,
				"modification_tag_epoch": modificationTag.Epoch,
				"modification_tag_index": modificationTag.Index,
				"routes":                 routesData,
				"active_definition_id":   update.DefinitionId,
			}, wheresClause, id,
		)
		if err != nil {
			logger.Error("failed-updating-lrp-deployment", err)
			return err
		}

		return nil
	})

	return id, err
}

func (db *SQLDB) SaveLRPDeployment(logger lager.Logger, lrpDeployment *models.LRPDeployment) error {
	wheresClause := "lrp_deployments.process_guid = ?"
	_, err := db.update(logger, db.db, lrpDeploymentsTable,
		helpers.SQLAttributes{
			"healthy_definition_id": lrpDeployment.HealthyDefinitionId,
		}, wheresClause, lrpDeployment.ProcessGuid,
	)
	return err
}

func (db *SQLDB) DeleteLRPDeployment(logger lager.Logger, id string) error {
	return nil
}

func (db *SQLDB) ActivateLRPDeploymentDefinition(logger lager.Logger, id, definitionId string) error {
	return nil
}

func (db *SQLDB) fetchLRPDeployment(logger lager.Logger, row *sql.Row) (*models.LRPDeployment, error) {
	lrpDeployment := &models.LRPDeployment{
		ModificationTag: &models.ModificationTag{},
	}
	var routeData []byte
	values := []interface{}{
		&lrpDeployment.ProcessGuid,
		&lrpDeployment.Domain,
		&lrpDeployment.Instances,
		&lrpDeployment.Annotation,
		&routeData,
		// &lrpDeployment.HealthyDefinitionId,
		&lrpDeployment.ActiveDefinitionId,
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
	return lrpDeployment, nil
}

func (db *SQLDB) fetchLRPDefinitionsInternal(logger lager.Logger, scanner RowScanner) (*models.LRPDefinition, string, error) {
	definition := &models.LRPDefinition{}
	var volumeData, placementData, runInfoData []byte
	values := []interface{}{
		&definition.DefinitionId,
		&definition.LogGuid,
		&definition.MemoryMb,
		&definition.DiskMb,
		&definition.MaxPids,
		&definition.RootFs,
		&volumeData,
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

	// var volumePlacement models.VolumePlacement
	// err = db.deserializeModel(logger, volumeData, &volumePlacement)
	// if err != nil {
	// 	logger.Error("failed-parsing-volume-placement", err)
	// 	return nil, "", err
	// }
	// definition.VolumeMounts = &volumePlacement
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

func (db *SQLDB) LRPDeploymentByDefinitionGuid(logger lager.Logger, id string) (*models.LRPDeployment, error) {
	logger = logger.WithData(lager.Data{"definition_guid": id})
	logger.Debug("starting")
	defer logger.Debug("complete")

	var lrpDeployment *models.LRPDeployment

	err := db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		var err error
		wheresClause := " WHERE lrp_definitions.definition_guid = ?"
		values := []interface{}{id}
		//TODO: now using QueryRow which doesn't return an error. How do we check for errors?
		row := db.oneLRPDeploymentWithDefinitions(logger, tx, lrpDeploymentColumns, wheresClause, values)
		if row != nil {
			lrpDeployment, err = db.fetchLRPDeployment(logger, row)
		} else {
			return helpers.ErrResourceNotFound
		}

		wheresClause = " WHERE process_guid = ?"
		values = []interface{}{lrpDeployment.ProcessGuid}
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

		lrpDeployment.Definitions = definitions

		return err
	})

	return lrpDeployment, err
}
