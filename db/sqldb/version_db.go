package sqldb

import (
	"encoding/json"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

const VERSION_ID = "version"

func (db *SQLDB) SetVersion(logger lager.Logger, version *models.Version) error {
	versionJSON, err := json.Marshal(version)
	if err != nil {
		return err
	}

	result, err := db.db.Exec("UPDATE configurations SET value = ? WHERE id = ?", versionJSON, VERSION_ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected < 1 {
		_, err = db.db.Exec("INSERT INTO configurations (id, value) VALUES (?, ?)", VERSION_ID, versionJSON)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *SQLDB) Version(logger lager.Logger) (*models.Version, error) {
	rows, err := db.db.Query("SELECT value FROM configurations WHERE id = ?", VERSION_ID)
	if err != nil {
		return nil, err
	}

	if rows.Next() {
		var versionJSON string
		err = rows.Scan(&versionJSON)
		if err != nil {
			return nil, err
		}

		var version models.Version
		err = json.Unmarshal([]byte(versionJSON), &version)
		if err != nil {
			return nil, models.ErrDeserializeJSON
		}

		return &version, nil
	}

	return nil, models.ErrResourceNotFound
}
