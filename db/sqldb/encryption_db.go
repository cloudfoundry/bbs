package sqldb

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

const EncryptionKeyID = "encryption_key_label"

func (db *SQLDB) SetEncryptionKeyLabel(logger lager.Logger, label string) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.Exec("UPDATE configurations SET value = ? WHERE id = ? ", label, EncryptionKeyID)
	if err != nil {
		return errors.New(fmt.Sprintf("Error updating encryption key label: %s", err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected < 1 {
		_, err = tx.Exec("INSERT INTO configurations VALUES (?, ?)", EncryptionKeyID, label)
		if err != nil {
			return errors.New(fmt.Sprintf("Error creating encryption key label: %s", err))
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (db *SQLDB) EncryptionKeyLabel(logger lager.Logger) (string, error) {
	var label string
	err := db.db.QueryRow(
		"SELECT value FROM configurations WHERE id = ?",
		EncryptionKeyID,
	).Scan(&label)
	if err != nil {
		return "", models.ErrResourceNotFound
	}
	return label, nil
}
