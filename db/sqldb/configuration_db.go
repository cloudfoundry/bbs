package sqldb

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

func (db *SQLDB) setConfigurationValue(logger lager.Logger, key, value string) error {
	_, err := db.db.Exec(db.getQuery(SetConfigurationValueQuery),
		key,
		value,
		value,
	)
	if err != nil {
		logger.Error("failed-setting-config-value", err, lager.Data{"key": key})
		return db.convertSQLError(err)
	}

	return nil
}

func (db *SQLDB) getConfigurationValue(logger lager.Logger, key string) (string, error) {
	var value string
	err := db.db.QueryRow(db.getQuery(GetConfigurationValueQuery),
		key,
	).Scan(&value)
	if err != nil {
		logger.Error("failed-fetching-config-value", err, lager.Data{"key": key})
		return "", models.ErrResourceNotFound
	}

	return value, nil
}
