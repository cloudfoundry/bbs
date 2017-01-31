package sqldb

import (
	"database/sql"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

func (db *SQLDB) Lock(logger lager.Logger, lock models.Lock) error {
	logger = logger.Session("lock")

	err := db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		_, err := db.insert(logger, tx, "locks", helpers.SQLAttributes{
			"key":   lock.Key,
			"owner": lock.Owner,
			"value": lock.Value,
		})

		modelErr := db.convertSQLError(err)
		if modelErr == models.ErrResourceExists {
			logger.Error("lock-already-exists", err)
			return models.ErrLockCollision
		}

		return err
	})

	return err
}

func (db *SQLDB) ReleaseLock(logger lager.Logger, lock models.Lock) error {
	logger = logger.Session("release-lock")

	return db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		var owner string
		row := db.one(logger, tx, "locks", []string{"owner"}, true, "key = ?", lock.Key)
		err := row.Scan(&owner)
		if err != nil {
			return err
		}

		if owner != lock.Owner {
			logger.Error("cannot-release-lock", models.ErrLockCollision, lager.Data{
				"key":   lock.Key,
				"owner": owner,
				"actor": lock.Owner,
			})

			return models.ErrLockCollision
		}

		_, err = db.delete(logger, tx, "locks", "key = ?", lock.Key)
		return err
	})
}
