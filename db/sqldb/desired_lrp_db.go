package sqldb

import (
	"database/sql"
	"encoding/json"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

func (db *SQLDB) DesireLRP(logger lager.Logger, desiredLRP *models.DesiredLRP) error {
	return db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
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
