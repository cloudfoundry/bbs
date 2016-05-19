package sqldb

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

func (db *SQLDB) DesireLRP(logger lager.Logger, desiredLRP *models.DesiredLRP) error {
	logger = logger.WithData(lager.Data{"process_guid": desiredLRP.ProcessGuid})
	logger.Debug("starting")
	defer logger.Debug("complete")

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

		guid, err := db.guidProvider.NextGUID()
		if err != nil {
			logger.Error("failed-to-generate-guid", err)
			return models.ErrGUIDGeneration
		}

		desiredLRP.ModificationTag = &models.ModificationTag{Epoch: guid, Index: 0}

		_, err = tx.Exec(`
			INSERT INTO desired_lrps
				(process_guid, domain, log_guid, annotation, instances, memory_mb,
				disk_mb, rootfs, volume_placement, modification_tag_epoch, modification_tag_index,
				routes, run_info)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
			desiredLRP.ProcessGuid,
			desiredLRP.Domain,
			desiredLRP.LogGuid,
			desiredLRP.Annotation,
			desiredLRP.Instances,
			desiredLRP.MemoryMb,
			desiredLRP.DiskMb,
			desiredLRP.RootFs,
			volumePlacementData,
			desiredLRP.ModificationTag.Epoch,
			desiredLRP.ModificationTag.Index,
			routesData,
			runInfoData,
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

	return db.selectDesiredLRPByGuid(logger, processGuid, db.db, NoLock)
}

func (db *SQLDB) DesiredLRPs(logger lager.Logger, filter models.DesiredLRPFilter) ([]*models.DesiredLRP, error) {
	logger = logger.WithData(lager.Data{"filter": filter})
	logger.Debug("start")
	defer logger.Debug("complete")

	var rows *sql.Rows
	var err error
	if filter.Domain != "" {
		rows, err = db.db.Query(`
			SELECT process_guid, domain, log_guid, annotation, instances, memory_mb,
				disk_mb, rootfs, routes, modification_tag_epoch, modification_tag_index,
				run_info
			FROM desired_lrps
			WHERE domain = $1`,
			filter.Domain)
	} else {
		rows, err = db.db.Query(`
		SELECT process_guid, domain, log_guid, annotation, instances, memory_mb,
			disk_mb, rootfs, routes, modification_tag_epoch, modification_tag_index,
			run_info
		FROM desired_lrps`)
	}
	if err != nil {
		logger.Error("failed-query", err)
		return nil, err
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

	var rows *sql.Rows
	var err error
	if filter.Domain != "" {
		rows, err = db.db.Query(`
			SELECT process_guid, domain, log_guid, annotation, instances, memory_mb,
				disk_mb, rootfs, routes, volume_placement, modification_tag_epoch,
				modification_tag_index
			FROM desired_lrps
			WHERE domain = $1`,
			filter.Domain,
		)
	} else {
		rows, err = db.db.Query(`
			SELECT process_guid, domain, log_guid, annotation, instances, memory_mb,
				disk_mb, rootfs, routes, volume_placement, modification_tag_epoch,
				modification_tag_index
			FROM desired_lrps`,
		)
	}
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
	logger.Debug("starting")
	defer logger.Debug("complete")

	var beforeDesiredLRP *models.DesiredLRP
	err := db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		var err error
		beforeDesiredLRP, err = db.selectDesiredLRPByGuid(logger, processGuid, tx, LockForUpdate)
		if err != nil {
			logger.Error("failed-lock-desired", err)
			return err
		}

		setKeys := []string{"modification_tag_index = $1"}
		setValues := []interface{}{beforeDesiredLRP.ModificationTag.Index + 1}

		bindVarIndex := 2
		if update.Annotation != nil {
			setKeys = append(setKeys, fmt.Sprintf("annotation = $%d", bindVarIndex))
			setValues = append(setValues, *update.Annotation)
			bindVarIndex++
		}

		if update.Instances != nil {
			setKeys = append(setKeys, fmt.Sprintf("instances = $%d", bindVarIndex))
			setValues = append(setValues, *update.Instances)
			bindVarIndex++
		}

		if update.Routes != nil {
			routeData, err := json.Marshal(update.Routes)
			if err != nil {
				logger.Error("failed-marshalling-routes", err)
				return models.ErrBadRequest
			}
			setKeys = append(setKeys, fmt.Sprintf("routes = $%d", bindVarIndex))
			setValues = append(setValues, routeData)
			bindVarIndex++
		}

		setValues = append(setValues, processGuid)

		query := fmt.Sprintf("UPDATE desired_lrps SET %s WHERE process_guid = $%d", strings.Join(setKeys, ", "), bindVarIndex)
		bindVarIndex++
		stmt, err := tx.Prepare(query)
		if err != nil {
			logger.Error("failed-preparing-query", err)
			return db.convertSQLError(err)
		}

		_, err = stmt.Exec(setValues...)
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
	logger.Debug("starting")
	defer logger.Debug("complete")

	return db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		err := db.lockDesiredLRPByGuidForUpdate(logger, processGuid, tx)
		if err != nil {
			logger.Error("failed-lock-desired", err)
			return err
		}

		_, err = tx.Exec("DELETE FROM desired_lrps WHERE process_guid = $1", processGuid)
		if err != nil {
			logger.Error("failed-deleting-from-db", err)
			return db.convertSQLError(err)
		}

		return nil
	})
}

var schedulingInfoColumns = `
	desired_lrps.process_guid,
	desired_lrps.domain,
	desired_lrps.log_guid,
	desired_lrps.annotation,
	desired_lrps.instances,
	desired_lrps.memory_mb,
	desired_lrps.disk_mb,
	desired_lrps.rootfs,
	desired_lrps.routes,
	desired_lrps.volume_placement,
	desired_lrps.modification_tag_epoch,
	desired_lrps.modification_tag_index`

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
	row := tx.QueryRow("SELECT 1 FROM desired_lrps WHERE process_guid = $1 FOR UPDATE", processGuid)
	var count int
	err := row.Scan(&count)
	if err == sql.ErrNoRows {
		return models.ErrResourceNotFound
	} else if err != nil {
		logger.Error("jim", err)
		return db.convertSQLError(err)
	}
	return nil
}

func (db *SQLDB) selectDesiredLRPByGuid(logger lager.Logger, processGuid string, q Queryable, lockMode int) (*models.DesiredLRP, error) {
	query := `
		SELECT process_guid, domain, log_guid, annotation, instances, memory_mb,
			disk_mb, rootfs, routes, modification_tag_epoch, modification_tag_index,
			run_info
		FROM desired_lrps
		WHERE process_guid = $1
	`
	switch lockMode {
	case LockForUpdate:
		query += "FOR UPDATE\n"
	}
	row := q.QueryRow(query, processGuid)
	return db.fetchDesiredLRP(logger, row)
}

func (db *SQLDB) fetchDesiredLRP(logger lager.Logger, scanner RowScanner) (*models.DesiredLRP, error) {
	var desiredLRP models.DesiredLRP
	desiredLRP.ModificationTag = &models.ModificationTag{}
	var routeData, runInformationData []byte
	err := scanner.Scan(
		&desiredLRP.ProcessGuid,
		&desiredLRP.Domain,
		&desiredLRP.LogGuid,
		&desiredLRP.Annotation,
		&desiredLRP.Instances,
		&desiredLRP.MemoryMb,
		&desiredLRP.DiskMb,
		&desiredLRP.RootFs,
		&routeData,
		&desiredLRP.ModificationTag.Epoch,
		&desiredLRP.ModificationTag.Index,
		&runInformationData,
	)
	if err != nil {
		return nil, models.ErrResourceNotFound
	}

	var runInformation models.DesiredLRPRunInfo
	err = db.deserializeModel(logger, runInformationData, &runInformation)
	if err != nil {
		_, err := db.db.Exec(`
				DELETE FROM desired_lrps
				WHERE process_guid = $1
				`, desiredLRP.ProcessGuid)

		if err != nil {
			logger.Error("failed-deleting-invalid-row", err)
		}
		return nil, models.ErrDeserialize
	}
	desiredLRP.AddRunInfo(runInformation)

	var routes models.Routes
	err = json.Unmarshal(routeData, &routes)
	if err != nil {
		return nil, models.ErrDeserialize
	}

	desiredLRP.Routes = &routes

	return &desiredLRP, nil
}

func (db *SQLDB) fetchDesiredLRPSchedulingInfo(logger lager.Logger, scanner RowScanner) (*models.DesiredLRPSchedulingInfo, error) {
	return db.fetchDesiredLRPSchedulingInfoAndMore(logger, scanner)
}
