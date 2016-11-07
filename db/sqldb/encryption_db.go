package sqldb

import (
	"database/sql"
	"fmt"

	"code.cloudfoundry.org/bbs/format"
	"code.cloudfoundry.org/lager"
)

const EncryptionKeyID = "encryption_key_label"

func (db *SQLDB) SetEncryptionKeyLabel(logger lager.Logger, label string) error {
	logger = logger.Session("set-encrption-key-label", lager.Data{"label": label})
	logger.Debug("starting")
	defer logger.Debug("complete")

	return db.setConfigurationValue(logger, EncryptionKeyID, label)
}

func (db *SQLDB) EncryptionKeyLabel(logger lager.Logger) (string, error) {
	logger = logger.Session("encrption-key-label")
	logger.Debug("starting")
	defer logger.Debug("complete")

	return db.getConfigurationValue(logger, EncryptionKeyID)
}

func (db *SQLDB) PerformEncryption(logger lager.Logger) error {
	errCh := make(chan error)
	go func() {
		errCh <- db.reEncrypt(logger, tasksTable, "guid", "task_definition", true)
	}()
	go func() {
		errCh <- db.reEncrypt(logger, desiredLRPsTable, "process_guid", "run_info", true)
	}()
	go func() {
		errCh <- db.reEncrypt(logger, desiredLRPsTable, "process_guid", "volume_placement", true)
	}()
	go func() {
		errCh <- db.reEncrypt(logger, actualLRPsTable, "process_guid", "net_info", false)
	}()

	for i := 0; i < 4; i++ {
		err := <-errCh
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *SQLDB) reEncrypt(logger lager.Logger, tableName, primaryKey, blobColumn string, encryptIfEmpty bool) error {
	logger = logger.WithData(
		lager.Data{"table_name": tableName, "primary_key": primaryKey, "blob_column": blobColumn},
	)
	rows, err := db.db.Query(fmt.Sprintf("SELECT %s FROM %s", primaryKey, tableName))
	if err != nil {
		return db.convertSQLError(err)
	}
	defer rows.Close()

	where := fmt.Sprintf("%s = ?", primaryKey)
	for rows.Next() {
		var guid string
		err := rows.Scan(&guid)
		if err != nil {
			logger.Error("failed-to-scan-primary-key", err)
			continue
		}

		err = db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
			var blob []byte
			row := db.one(logger, tx, tableName, ColumnList{blobColumn}, LockRow, where, guid)
			err := row.Scan(&blob)
			if err != nil {
				logger.Error("failed-to-scan-blob", err)
				return nil
			}

			// don't encrypt column if it doesn't contain any data, see #132626553 for more info
			if !encryptIfEmpty && len(blob) == 0 {
				return nil
			}

			encoder := format.NewEncoder(db.cryptor)
			payload, err := encoder.Decode(blob)
			if err != nil {
				logger.Error("failed-to-decode-blob", err)
				return nil
			}
			encryptedPayload, err := encoder.Encode(format.BASE64_ENCRYPTED, payload)
			if err != nil {
				logger.Error("failed-to-encode-blob", err)
				return err
			}
			_, err = db.update(logger, tx, tableName,
				SQLAttributes{blobColumn: encryptedPayload},
				where, guid,
			)
			if err != nil {
				logger.Error("failed-to-update-blob", err)
				return db.convertSQLError(err)
			}
			return nil
		})

		if err != nil {
			return err
		}
	}
	return nil
}
