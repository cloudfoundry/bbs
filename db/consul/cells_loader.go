package consul

import (
	"sync"

	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

type cellsLoader struct {
	lock    sync.Mutex
	db      *ConsulDB
	cellSet models.CellSet
	err     *models.Error
	logger  lager.Logger
}

func (loader *cellsLoader) Cells() (models.CellSet, *models.Error) {
	loader.lock.Lock()
	if loader.cellSet == nil {
		cells, err := loader.db.Cells(loader.logger)
		if err != nil {
			loader.err = err
		} else {
			loader.cellSet = models.CellSet{}
			for _, cell := range cells {
				loader.cellSet.Add(cell)
			}
		}
	}
	loader.lock.Unlock()

	return loader.cellSet, loader.err
}

func (db *ConsulDB) NewCellsLoader(logger lager.Logger) db.CellsLoader {
	return &cellsLoader{
		logger: logger,
		db:     db,
	}
}
