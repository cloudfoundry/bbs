package sqldb

import (
	"database/sql"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

func (db *SQLDB) getConfigurationValue(logger lager.Logger, key string) (string, error) {
	var value string
	err := db.db.QueryRow(
		"SELECT value FROM configurations WHERE id = ?",
		key,
	).Scan(&value)
	if err != nil {
		return "", models.ErrResourceNotFound
	}

	return value, nil
}

func (db *SQLDB) setConfigurationValue(logger lager.Logger, key, value string) error {
	return db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		_, err := tx.Exec(
			`INSERT INTO configurations (id, value) VALUES (?, ?)
										ON DUPLICATE KEY UPDATE value = ?`,
			key,
			value,
			value,
		)
		if err != nil {
			return db.convertSQLError(err)
		}

		return nil
	})
}
