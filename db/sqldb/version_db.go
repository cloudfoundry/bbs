package sqldb

import (
	"context"
	"encoding/json"

	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager/v3"
)

const VersionID = "version"

func (db *SQLDB) SetVersion(ctx context.Context, logger lager.Logger, version *models.Version) error {
	logger = logger.Session("db-set-version", lager.Data{"version": version})
	logger.Debug("starting")
	defer logger.Debug("complete")

	versionJSON, err := json.Marshal(version)
	if err != nil {
		logger.Error("failed-marshalling-version", err)
		return err
	}

	return db.setConfigurationValue(ctx, logger, VersionID, string(versionJSON))
}

func (db *SQLDB) Version(ctx context.Context, logger lager.Logger) (*models.Version, error) {
	logger = logger.Session("db-version")
	logger.Debug("starting")
	defer logger.Debug("complete")

	versionJSON, err := db.getConfigurationValue(ctx, logger, VersionID)
	if err != nil {
		return nil, err
	}

	var version models.Version
	err = json.Unmarshal([]byte(versionJSON), &version)
	if err != nil {
		logger.Error("failed-to-deserialize-version", err)
		return nil, models.ErrDeserialize
	}

	return &version, nil
}
