package db

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter . ActualLRPDB
type ActualLRPDB interface {
	ActualLRPGroups(models.ActualLRPFilter, lager.Logger) (*models.ActualLRPGroups, error)
	ActualLRPGroupsByProcessGuid(string, lager.Logger) (*models.ActualLRPGroups, error)
}
