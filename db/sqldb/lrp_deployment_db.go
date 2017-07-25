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
			logger.Error("failed-inserting-desired", err)
			return err
		}

		_, err = db.insert(logger, tx, lrpDefinitionsTable, helpers.SQLAttributes{
			"process_guid":     lrp.ProcessGuid,
			"definition_name":  lrp.DefinitionId,
			"definition_guid":  lrp.ProcessGuid,
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
			logger.Error("failed-inserting-desired", err)
			return err
		}
		return nil
	})

	return "", err
}

func (db *SQLDB) UpdateLRPDeployment(logger lager.Logger, id string, update *models.LRPDeploymentUpdate) (string, error) {
	return "", nil
}

func (db *SQLDB) DeleteLRPDeployment(logger lager.Logger, id string) error {
	return nil
}

func (db *SQLDB) ActivateLRPDeploymentDefinition(logger lager.Logger, id, definitionId string) error {
	return nil
}
