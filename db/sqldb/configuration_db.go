package sqldb

import (
	"database/sql"

	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

func (db *SQLDB) setConfigurationValue(logger lager.Logger, key, value string) error {
	return db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		_, err := db.upsert(logger, tx, "configurations",
			SQLAttributes{"id": key},
			SQLAttributes{"value": value},
		)
		if err != nil {
			logger.Error("failed-setting-config-value", err, lager.Data{"key": key})
			return db.convertSQLError(err)
		}

		return nil
	})
}

func (db *SQLDB) getConfigurationValue(logger lager.Logger, key string) (string, error) {
	var value string
	err := db.one(logger, db.db, "configurations",
		ColumnList{"value"}, NoLockRow,
		"id = ?", key,
	).Scan(&value)
	if err != nil {
		logger.Error("failed-fetching-config-value", err, lager.Data{"key": key})
		return "", models.ErrResourceNotFound
	}

	return value, nil
}
