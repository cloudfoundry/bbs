package bbs

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

const (
	CellSchemaKey    = "cell"
	BBSLockSchemaKey = "bbs_lock"
)

func CellSchemaRoot() string {
	return locket.LockSchemaPath(CellSchemaKey)
}

func CellSchemaPath(cellID string) string {
	return locket.LockSchemaPath(CellSchemaKey, cellID)
}

func BBSLockSchemaPath() string {
	return locket.LockSchemaPath(BBSLockSchemaKey)
}

//go:generate counterfeiter -o fake_bbs/fake_service_client.go . ServiceClient

type ServiceClient interface {
	CellById(logger lager.Logger, cellId string) (*models.CellPresence, error)
	Cells(logger lager.Logger) (models.CellSet, error)
	CellEvents(logger lager.Logger) <-chan models.CellEvent
	NewCellPresenceRunner(logger lager.Logger, cellPresence *models.CellPresence, retryInterval, lockTTL time.Duration) ifrit.Runner
	NewBBSLockRunner(logger lager.Logger, bbsPresence *models.BBSPresence, retryInterval, lockTTL time.Duration) (ifrit.Runner, error)
	CurrentBBS(logger lager.Logger) (*models.BBSPresence, error)
	CurrentBBSURL(logger lager.Logger) (string, error)
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

func (db *serviceClient) NewCellPresenceRunner(logger lager.Logger, cellPresence *models.CellPresence, retryInterval time.Duration, lockTTL time.Duration) ifrit.Runner {
	payload, err := models.ToJSON(cellPresence)
	if err != nil {
		panic(err)
	}

	return locket.NewPresence(logger, db.consulClient, CellSchemaPath(cellPresence.CellId), payload, db.clock, retryInterval, lockTTL)
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
	value, err := db.getAcquiredValue(CellSchemaPath(cellId))
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

func (db *serviceClient) NewBBSLockRunner(logger lager.Logger, bbsPresence *models.BBSPresence, retryInterval, lockTTL time.Duration) (ifrit.Runner, error) {
	bbsPresenceJSON, err := models.ToJSON(bbsPresence)
	if err != nil {
		return nil, err
	}
	return locket.NewLock(logger, db.consulClient, locket.LockSchemaPath("bbs_lock"), bbsPresenceJSON, db.clock, retryInterval, lockTTL), nil
}

func (db *serviceClient) CurrentBBS(logger lager.Logger) (*models.BBSPresence, error) {
	value, err := db.getAcquiredValue(BBSLockSchemaPath())
	if err != nil {
		return nil, convertConsulError(err)
	}

	presence := new(models.BBSPresence)
	err = models.FromJSON(value, presence)
	if err != nil {
		return nil, err
	}

	return presence, nil
}

func (db *serviceClient) CurrentBBSURL(logger lager.Logger) (string, error) {
	presence, err := db.CurrentBBS(logger)
	if err != nil {
		return "", err
	}

	return presence.URL, nil
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
