package db

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter . CellDB
type CellDB interface {
	NewCellsLoader(logger lager.Logger) CellsLoader
	CellById(logger lager.Logger, cellId string) (*models.CellPresence, *models.Error)
}

type CellsLoader interface {
	Cells() (models.CellSet, *models.Error)
}
