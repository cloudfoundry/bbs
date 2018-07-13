package sqldb

import (
	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

func (db *SQLDB) RemoveSuspectActualLRP(logger lager.Logger, lrpKey *models.ActualLRPKey) (*models.ActualLRPGroup, error) {
	logger = logger.Session("remove-suspect-lrp", lager.Data{"lrp_key": lrpKey})
	logger.Debug("starting")
	defer logger.Debug("complete")

	var lrpGroup *models.ActualLRPGroup

	err := db.transact(logger, func(logger lager.Logger, tx helpers.Tx) error {
		processGuid := lrpKey.ProcessGuid
		index := lrpKey.Index

		lrp, err := db.fetchActualLRPForUpdate(logger, processGuid, index, models.ActualLRP_Suspect, tx)
		if err == models.ErrResourceNotFound {
			logger.Debug("suspect-lrp-does-not-exist")
			return nil
		}

		if err != nil {
			logger.Error("failed-fetching-actual-lrp", err)
			return err
		}

		lrpGroup = &models.ActualLRPGroup{Instance: lrp}

		_, err = db.delete(logger, tx, "actual_lrps",
			"process_guid = ? AND instance_index = ? AND presence = ?",
			processGuid, index, models.ActualLRP_Suspect,
		)

		if err != nil {
			logger.Error("failed-delete", err)
			return models.ErrActualLRPCannotBeRemoved
		}

		return nil
	})

	return lrpGroup, err
}
