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
	return db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		logger = logger.Session("create-desired-lrp", lager.Data{"process-guid": desiredLRP.ProcessGuid})
		logger.Debug("starting")
		defer logger.Debug("complete")

		routesData, err := json.Marshal(desiredLRP.Routes)
		runInfo := desiredLRP.DesiredLRPRunInfo(db.clock.Now())

		runInfoData, err := db.serializeModel(logger, &runInfo)
		if err != nil {
			return err
		}

		_, err = tx.Exec(`
			INSERT INTO desired_lrps
				(process_guid, domain, log_guid, annotation, instances, memory_mb,
				disk_mb, rootfs, modification_tag_epoch, modification_tag_index,
				routes, run_info)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			desiredLRP.ProcessGuid,
			desiredLRP.Domain,
			desiredLRP.LogGuid,
			desiredLRP.Annotation,
			desiredLRP.Instances,
			desiredLRP.MemoryMb,
			desiredLRP.DiskMb,
			desiredLRP.RootFs,
			desiredLRP.ModificationTag.Epoch,
			desiredLRP.ModificationTag.Index,
			routesData,
			runInfoData,
		)
		if err != nil {
			return db.convertSQLError(err)
		}
		return nil
	})
}

func (db *SQLDB) DesiredLRPByProcessGuid(logger lager.Logger, processGuid string) (*models.DesiredLRP, error) {
	row := db.db.QueryRow(`
		SELECT process_guid, domain, log_guid, annotation, instances, memory_mb,
			disk_mb, rootfs, routes, modification_tag_epoch, modification_tag_index,
			run_info
		FROM desired_lrps
		WHERE process_guid = ?`,
		processGuid)
	return db.fetchDesiredLRP(logger, row)
}

func (db *SQLDB) DesiredLRPs(logger lager.Logger, filter models.DesiredLRPFilter) ([]*models.DesiredLRP, error) {
	var rows *sql.Rows
	var err error
	if filter.Domain != "" {
		rows, err = db.db.Query(`
			SELECT process_guid, domain, log_guid, annotation, instances, memory_mb,
				disk_mb, rootfs, routes, modification_tag_epoch, modification_tag_index,
				run_info
			FROM desired_lrps
			WHERE domain = ?`,
			filter.Domain)
	} else {
		rows, err = db.db.Query(`
			SELECT process_guid, domain, log_guid, annotation, instances, memory_mb,
				disk_mb, rootfs, routes, modification_tag_epoch, modification_tag_index,
				run_info
			FROM desired_lrps`)
	}
	if err != nil {
		return nil, err
	}

	results := []*models.DesiredLRP{}
	for rows.Next() {
		desiredLRP, err := db.fetchDesiredLRP(logger, rows)
		if err != nil {
			continue
		}
		results = append(results, desiredLRP)
	}

	return results, nil
}

func (db *SQLDB) DesiredLRPSchedulingInfos(logger lager.Logger, filter models.DesiredLRPFilter) ([]*models.DesiredLRPSchedulingInfo, error) {
	var rows *sql.Rows
	var err error
	if filter.Domain != "" {
		rows, err = db.db.Query(`
			SELECT process_guid, domain, log_guid, annotation, instances, memory_mb,
				disk_mb, rootfs, routes, modification_tag_epoch, modification_tag_index
			FROM desired_lrps
			WHERE domain = ?`,
			filter.Domain,
		)
	} else {
		rows, err = db.db.Query(`
			SELECT process_guid, domain, log_guid, annotation, instances, memory_mb,
				disk_mb, rootfs, routes, modification_tag_epoch, modification_tag_index
			FROM desired_lrps`,
		)
	}
	if err != nil {
		return nil, err
	}

	results := []*models.DesiredLRPSchedulingInfo{}
	for rows.Next() {
		desiredLRPSchedulingInfo, err := db.fetchDesiredLRPSchedulingInfo(logger, rows)
		if err != nil {
			continue
		}
		results = append(results, desiredLRPSchedulingInfo)
	}

	return results, nil
}

func (db *SQLDB) UpdateDesiredLRP(logger lager.Logger, processGuid string, update *models.DesiredLRPUpdate) (int32, error) {
	var previousInstanceCount int32
	err := db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		err := db.lockDesiredLRPByGuidForShare(logger, processGuid, tx)
		if err != nil {
			return err
		}

		row := db.db.QueryRow("SELECT instances FROM desired_lrps WHERE process_guid = ?", processGuid)

		err = row.Scan(&previousInstanceCount)
		if err != nil {
			return models.ErrResourceNotFound
		}

		setKeys := []string{}
		setValues := []interface{}{}
		if update.Annotation != nil {
			setKeys = append(setKeys, "annotation = ?")
			setValues = append(setValues, *update.Annotation)
		}

		if update.Instances != nil {
			setKeys = append(setKeys, "instances = ?")
			setValues = append(setValues, *update.Instances)
		}

		if update.Routes != nil {
			routeData, err := json.Marshal(update.Routes)
			if err != nil {
				return models.ErrBadRequest
			}
			setKeys = append(setKeys, "routes = ?")
			setValues = append(setValues, routeData)
		}

		if len(setKeys) == 0 {
			return nil
		}

		setValues = append(setValues, processGuid)

		query := fmt.Sprintf("UPDATE desired_lrps SET %s WHERE process_guid = ?", strings.Join(setKeys, ", "))
		stmt, err := tx.Prepare(query)
		if err != nil {
			return db.convertSQLError(err)
		}

		_, err = stmt.Exec(setValues...)
		if err != nil {
			return db.convertSQLError(err)
		}

		return nil
	})

	return previousInstanceCount, err
}

func (db *SQLDB) RemoveDesiredLRP(logger lager.Logger, processGuid string) error {
	return db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		err := db.lockDesiredLRPByGuidForShare(logger, processGuid, tx)
		if err != nil {
			return err
		}

		_, err = tx.Exec("DELETE FROM desired_lrps WHERE process_guid = ?", processGuid)
		if err != nil {
			panic(err)
			return db.convertSQLError(err)
		}

		return nil
	})
}

func (db *SQLDB) lockDesiredLRPByGuidForShare(logger lager.Logger, processGuid string, tx *sql.Tx) error {
	row := tx.QueryRow("SELECT count(*) FROM desired_lrps WHERE process_guid = ? LOCK IN SHARE MODE", processGuid)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return db.convertSQLError(err)
	}
	if count == 0 {
		return models.ErrResourceNotFound
	}
	return nil
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
	err = db.serializer.Unmarshal(logger, runInformationData, &runInformation)
	if err != nil {
		return nil, err
	}
	desiredLRP.AddRunInfo(runInformation)

	var routes models.Routes
	err = json.Unmarshal(routeData, &routes)
	if err != nil {
		return nil, models.ErrDeserializeJSON
	}

	desiredLRP.Routes = &routes

	return &desiredLRP, nil
}

func (db *SQLDB) fetchDesiredLRPSchedulingInfo(logger lager.Logger, scanner RowScanner) (*models.DesiredLRPSchedulingInfo, error) {
	var desiredLRPSchedulingInfo models.DesiredLRPSchedulingInfo
	desiredLRPSchedulingInfo.ModificationTag = models.ModificationTag{}
	var routeData []byte
	err := scanner.Scan(
		&desiredLRPSchedulingInfo.ProcessGuid,
		&desiredLRPSchedulingInfo.Domain,
		&desiredLRPSchedulingInfo.LogGuid,
		&desiredLRPSchedulingInfo.Annotation,
		&desiredLRPSchedulingInfo.Instances,
		&desiredLRPSchedulingInfo.MemoryMb,
		&desiredLRPSchedulingInfo.DiskMb,
		&desiredLRPSchedulingInfo.RootFs,
		&routeData,
		&desiredLRPSchedulingInfo.ModificationTag.Epoch,
		&desiredLRPSchedulingInfo.ModificationTag.Index,
	)
	if err != nil {
		return nil, models.ErrResourceNotFound
	}

	var routes models.Routes
	err = json.Unmarshal(routeData, &routes)
	if err != nil {
		return nil, models.ErrDeserializeJSON
	}

	desiredLRPSchedulingInfo.Routes = routes

	return &desiredLRPSchedulingInfo, nil
}
