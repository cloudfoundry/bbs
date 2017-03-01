package bbs

import (
	"os"
	"path"

	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/consuladapter"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/locket"
	"code.cloudfoundry.org/rep/maintain"
	"github.com/tedsuo/ifrit"
)

const BBSLockSchemaKey = "bbs_lock"

func CellSchemaRoot() string {
	return locket.LockSchemaPath(maintain.CellSchemaKey)
}

func BBSLockSchemaPath() string {
	return locket.LockSchemaPath(BBSLockSchemaKey)
}

//go:generate counterfeiter -o fake_bbs/fake_service_client.go . ServiceClient

type ServiceClient interface {
	CellById(logger lager.Logger, cellId string) (*models.CellPresence, error)
	Cells(logger lager.Logger) (models.CellSet, error)
	CellEvents(logger lager.Logger) <-chan models.CellEvent
}

type serviceClient struct {
	consulClient consuladapter.Client
	clock        clock.Clock
}

func NewServiceClient(client consuladapter.Client, clock clock.Clock) ServiceClient {
	return &serviceClient{
		consulClient: client,
		clock:        clock,
	}
}

func (db *serviceClient) Cells(logger lager.Logger) (models.CellSet, error) {
	kvPairs, _, err := db.consulClient.KV().List(CellSchemaRoot(), nil)
	if err != nil {
		bbsErr := models.ConvertError(convertConsulError(err))
		if bbsErr.Type != models.Error_ResourceNotFound {
			return nil, bbsErr
		}
	}

	if kvPairs == nil {
		err = consuladapter.NewPrefixNotFoundError(CellSchemaRoot())
		bbsErr := models.ConvertError(convertConsulError(err))
		if bbsErr.Type != models.Error_ResourceNotFound {
			return nil, bbsErr
		}
	}

	cellPresences := models.NewCellSet()
	for _, kvPair := range kvPairs {
		if kvPair.Session == "" {
			continue
		}

		cell := kvPair.Value
		presence := new(models.CellPresence)
		err := models.FromJSON(cell, presence)
		if err != nil {
			logger.Error("failed-to-unmarshal-cells-json", err)
			continue
		}
		cellPresences.Add(presence)
	}

	return cellPresences, nil
}

func (db *serviceClient) CellById(logger lager.Logger, cellId string) (*models.CellPresence, error) {
	value, err := db.getAcquiredValue(maintain.CellSchemaPath(cellId))
	if err != nil {
		return nil, convertConsulError(err)
	}

	presence := new(models.CellPresence)
	err = models.FromJSON(value, presence)
	if err != nil {
		return nil, models.NewError(models.Error_InvalidJSON, err.Error())
	}

	return presence, nil
}

func (db *serviceClient) CellEvents(logger lager.Logger) <-chan models.CellEvent {
	logger = logger.Session("cell-events")

	disappearanceWatcher, disappeared := locket.NewDisappearanceWatcher(logger, db.consulClient, CellSchemaRoot(), db.clock)
	process := ifrit.Invoke(disappearanceWatcher)

	events := make(chan models.CellEvent)
	go func() {
		for {
			select {
			case keys, ok := <-disappeared:
				if !ok {
					process.Signal(os.Interrupt)
					return
				}

				cellIDs := make([]string, len(keys))
				for i, key := range keys {
					cellIDs[i] = path.Base(key)
				}
				logger.Info("cell-disappeared", lager.Data{"cell_ids": cellIDs})
				events <- models.NewCellDisappearedEvent(cellIDs)
			}
		}
	}()

	return events
}

func convertConsulError(err error) error {
	switch err.(type) {
	case consuladapter.KeyNotFoundError:
		return models.NewError(models.Error_ResourceNotFound, err.Error())
	case consuladapter.PrefixNotFoundError:
		return models.NewError(models.Error_ResourceNotFound, err.Error())
	default:
		return models.NewError(models.Error_UnknownError, err.Error())
	}
}

func (db *serviceClient) getAcquiredValue(key string) ([]byte, error) {
	kvPair, _, err := db.consulClient.KV().Get(key, nil)
	if err != nil {
		return nil, err
	}

	if kvPair == nil || kvPair.Session == "" {
		return nil, consuladapter.NewKeyNotFoundError(key)
	}

	return kvPair.Value, nil
}
