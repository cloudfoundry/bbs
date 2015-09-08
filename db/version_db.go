package db

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter . VersionDB
type VersionDB interface {
	Version(logger lager.Logger) (*models.Version, error)
	SetVersion(logger lager.Logger, version *models.Version) error
}
