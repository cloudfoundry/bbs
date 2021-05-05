package sqldb

import (
	"context"

	"code.cloudfoundry.org/diegosqldb"
	"code.cloudfoundry.org/lager"
)

const configurationsTable = "configurations"

func (db *SQLDB) setConfigurationValue(ctx context.Context, logger lager.Logger, key, value string) error {
	return db.transact(ctx, logger, func(logger lager.Logger, tx diegosqldb.Tx) error {
		_, err := db.upsert(
			ctx,
			logger,
			tx,
			configurationsTable,
			diegosqldb.SQLAttributes{"value": value, "id": key},
			"id = ?", key,
		)
		if err != nil {
			logger.Error("failed-setting-config-value", err, lager.Data{"key": key})
			return err
		}

		return nil
	})
}

func (db *SQLDB) getConfigurationValue(ctx context.Context, logger lager.Logger, key string) (string, error) {
	var value string
	err := db.transact(ctx, logger, func(logger lager.Logger, tx diegosqldb.Tx) error {
		return db.one(ctx, logger, tx, "configurations",
			diegosqldb.ColumnList{"value"}, diegosqldb.NoLockRow,
			"id = ?", key,
		).Scan(&value)
	})

	if err != nil {
		return "", err
	}

	return value, nil
}
