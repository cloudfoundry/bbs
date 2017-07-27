package sqldb

import (
	"database/sql"
	"encoding/json"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

func (db *SQLDB) CreateLRPDeployment(logger lager.Logger, lrp *models.LRPDeploymentDefinition) (string, error) {
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
		columns := helpers.ColumnList{
			"lrp_deployments.process_guid",
			"lrp_deployments.domain",
		}

		row := db.one(logger, tx, lrpDeploymentsTable, columns, false, wheresClause, id)
		desiredLRPDeployment, err := db.fetchLRPDeployment(logger, row)
		if err != nil {
			logger.Error("failed-to-get-desired-lrp", err)
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

func (db *SQLDB) DeleteLRPDeployment(logger lager.Logger, id string) error {
	return nil
}

func (db *SQLDB) ActivateLRPDeploymentDefinition(logger lager.Logger, id, definitionId string) error {
	return nil
}

func (db *SQLDB) fetchLRPDeployment(logger lager.Logger, row *sql.Row) (*models.LRPDeployment, error) {
	lrpDeployment := &models.LRPDeployment{}
	values := []interface{}{
		&lrpDeployment.ProcessGuid,
		&lrpDeployment.Domain,
	}

	err := row.Scan(values...)
	if err == sql.ErrNoRows {
		return nil, err
	}
	if err != nil {
		logger.Error("failed-scanning", err)
		return nil, err
	}
	return lrpDeployment, nil
}
