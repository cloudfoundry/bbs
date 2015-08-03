package consul

import (
	"path"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/consuladapter"
	"github.com/pivotal-golang/lager"
)

func CellSchemaPath(cellID string) string {
	return path.Join(CellSchemaRoot, cellID)
}

func (db *ConsulDB) CellById(logger lager.Logger, cellId string) (*models.CellPresence, *models.Error) {
	cellPresence := models.CellPresence{}

	value, err := db.session.GetAcquiredValue(CellSchemaPath(cellId))
	if err != nil {
		return nil, convertConsulError(err)
	}

	err = models.FromJSON(value, &cellPresence)
	if err != nil {
		return nil, models.ErrDeserializeJSON
	}

	return &cellPresence, nil
}

func convertConsulError(err error) *models.Error {
	switch err.(type) {
	case consuladapter.KeyNotFoundError:
		return models.ErrResourceNotFound
	case consuladapter.PrefixNotFoundError:
		return models.ErrResourceNotFound
	default:
		return models.ErrUnknownError
	}
}
