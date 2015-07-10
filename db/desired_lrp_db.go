package db

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter . DesiredLRPDB
type DesiredLRPDB interface {
	DesiredLRPs(models.DesiredLRPFilter, lager.Logger) (*models.DesiredLRPs, error)
}
