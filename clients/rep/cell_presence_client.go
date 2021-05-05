package rep

import (
	"os"
	"path"
	"time"

	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/consuladapter"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/locket"
	"github.com/tedsuo/ifrit"
)

const CellSchemaKey = "cell"

//go:generate counterfeiter . CellPresenceClient

type CellPresenceClient interface {
	NewCellPresenceRunner(logger lager.Logger, cellPresence *models.CellPresence, retryInterval, lockTTL time.Duration) ifrit.Runner

	CellById(logger lager.Logger, cellId string) (*models.CellPresence, error)
	Cells(logger lager.Logger) (models.CellSet, error)
	CellEvents(logger lager.Logger) <-chan models.CellEvent
}

type cellPresenceClient struct {
	consulClient consuladapter.Client
	clock        clock.Clock
}

func CellSchemaPath(cellID string) string {
	return locket.LockSchemaPath(CellSchemaKey, cellID)
}

func CellSchemaRoot() string {
	return locket.LockSchemaPath(CellSchemaKey)
}

func NewCellPresenceClient(client consuladapter.Client, clock clock.Clock) *cellPresenceClient {
	return &cellPresenceClient{
		consulClient: client,
		clock:        clock,
	}
}

func (db *cellPresenceClient) NewCellPresenceRunner(logger lager.Logger, cellPresence *models.CellPresence, retryInterval time.Duration, lockTTL time.Duration) ifrit.Runner {
	payload, err := models.ToJSON(cellPresence)
	if err != nil {
		panic(err)
	}

	return locket.NewPresence(logger, db.consulClient, CellSchemaPath(cellPresence.CellId), payload, db.clock, retryInterval, lockTTL)
}

func (c *cellPresenceClient) Cells(logger lager.Logger) (models.CellSet, error) {
	kvPairs, _, err := c.consulClient.KV().List(CellSchemaRoot(), nil)
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

func (c *cellPresenceClient) CellById(logger lager.Logger, cellId string) (*models.CellPresence, error) {
	value, err := c.getAcquiredValue(CellSchemaPath(cellId))
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

func (c *cellPresenceClient) CellEvents(logger lager.Logger) <-chan models.CellEvent {
	logger = logger.Session("cell-events")

	disappearanceWatcher, disappeared := locket.NewDisappearanceWatcher(logger, c.consulClient, CellSchemaRoot(), c.clock)
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

func (c *cellPresenceClient) getAcquiredValue(key string) ([]byte, error) {
	kvPair, _, err := c.consulClient.KV().Get(key, nil)
	if err != nil {
		return nil, err
	}

	if kvPair == nil || kvPair.Session == "" {
		return nil, consuladapter.NewKeyNotFoundError(key)
	}

	return kvPair.Value, nil
}
