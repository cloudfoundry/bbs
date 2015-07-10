package db

import (
	"github.com/cloudfoundry-incubator/bbs"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter . DesiredLRPDB
type DesiredLRPDB interface {
	DesiredLRPs(filter models.DesiredLRPFilter, logger lager.Logger) (*models.DesiredLRPs, *bbs.Error)
}
