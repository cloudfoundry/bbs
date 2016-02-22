package sqldb

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

const EncryptionKeyId = "encryption_key_label"

func (db *SQLDB) SetEncryptionKeyLabel(logger lager.Logger, label string) error {
	var err error
	result, err := db.db.Exec("UPDATE configurations SET value = ? WHERE id = ? ", label, EncryptionKeyId)
	if err != nil {
		return errors.New(fmt.Sprintf("Error updating encryption key label: %s", err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected < 1 {
		_, err = db.db.Exec("INSERT INTO configurations VALUES (?, ?)", EncryptionKeyId, label)
		if err != nil {
			return errors.New(fmt.Sprintf("Error creating encryption key label: %s", err))
		}
	}

	return nil
}

func (db *SQLDB) EncryptionKeyLabel(logger lager.Logger) (string, error) {
	rows, err := db.db.Query("SELECT value FROM configurations WHERE id = ?", EncryptionKeyId)
	if err != nil {
		return "", err
	}

	if rows.Next() {
		var label string
		err = rows.Scan(&label)
		if err != nil {
			return "", err
		}
		return label, nil
	}

	return "", models.ErrResourceNotFound
}
