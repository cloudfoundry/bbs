package db

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter . CellDB
type CellDB interface {
	CellById(logger lager.Logger, cellId string) (*models.CellPresence, *models.Error)
}
