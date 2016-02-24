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

	return db.setConfigurationValue(logger, VersionID, string(versionJSON))
}

func (db *SQLDB) Version(logger lager.Logger) (*models.Version, error) {
	versionJSON, err := db.getConfigurationValue(logger, VersionID)
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
