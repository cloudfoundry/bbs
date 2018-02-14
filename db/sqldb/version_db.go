package sqldb

import (
	"encoding/json"

	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

const VersionID = "version"

func (db *SQLDB) SetVersion(logger lager.Logger, version *models.Version) error {
	logger = logger.Session("set-version", lager.Data{"version": version})
	logger.Debug("starting")
	defer logger.Debug("complete")

	versionJSON, err := json.Marshal(version)
	if err != nil {
		return F("set-version", err)
	}

	return db.setConfigurationValue(logger, VersionID, string(versionJSON))
}

func (db *SQLDB) Version(logger lager.Logger) (*models.Version, error) {
	logger = logger.Session("version")
	logger.Debug("starting")
	defer logger.Debug("complete")

	versionJSON, err := db.getConfigurationValue(logger, VersionID)
	if err != nil {
		return nil, E("version", err)
	}

	var version models.Version
	err = json.Unmarshal([]byte(versionJSON), &version)
	if err != nil {
		return nil, F("version", err)
	}

	return &version, nil
}
