package sqldb

import (
	"encoding/json"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

const VersionID = "version"

func (db *SQLDB) SetVersion(logger lager.Logger, version *models.Version) error {
	versionJSON, err := json.Marshal(version)
	if err != nil {
		return err
	}

	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.Exec("UPDATE configurations SET value = ? WHERE id = ?", versionJSON, VersionID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected < 1 {
		_, err = tx.Exec("INSERT INTO configurations (id, value) VALUES (?, ?)", VersionID, versionJSON)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (db *SQLDB) Version(logger lager.Logger) (*models.Version, error) {
	var versionJSON string
	err := db.db.QueryRow(
		"SELECT value FROM configurations WHERE id = ?",
		VersionID,
	).Scan(&versionJSON)
	if err != nil {
		return nil, models.ErrResourceNotFound
	}

	var version models.Version
	err = json.Unmarshal([]byte(versionJSON), &version)
	if err != nil {
		return nil, models.ErrDeserializeJSON
	}

	return &version, nil
}
