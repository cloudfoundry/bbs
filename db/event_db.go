package db

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter . EventDB
type EventDB interface {
	WatchForActualLRPChanges(lager.Logger,
		func(created *models.ActualLRPGroup),
		func(changed *models.ActualLRPChange),
		func(deleted *models.ActualLRPGroup)) (chan<- bool, <-chan error)
	WatchForDesiredLRPChanges(lager.Logger,
		func(created *models.DesiredLRP),
		func(changed *models.DesiredLRPChange),
		func(deleted *models.DesiredLRP)) (chan<- bool, <-chan error)
}
