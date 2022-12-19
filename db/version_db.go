package db

import (
	"context"

	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

//counterfeiter:generate . VersionDB
type VersionDB interface {
	Version(ctx context.Context, logger lager.Logger) (*models.Version, error)
	SetVersion(ctx context.Context, logger lager.Logger, version *models.Version) error
}
