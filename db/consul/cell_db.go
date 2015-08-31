package consul

import (
	"encoding/json"
	"path"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/consuladapter"
	"github.com/pivotal-golang/lager"
)

func CellSchemaPath(cellID string) string {
	return path.Join(CellSchemaRoot, cellID)
}

func (db *ConsulDB) Cells(logger lager.Logger) ([]*models.CellPresence, error) {
	cells, err := db.session.ListAcquiredValues(CellSchemaRoot)
	if err != nil {
		bbsErr := models.ConvertError(convertConsulError(err))
		if bbsErr.Type != models.Error_ResourceNotFound {
			return nil, err
		}
	}

	cellPresences := []*models.CellPresence{}
	for _, cell := range cells {
		cellPresence := new(models.CellPresence)
		err := json.Unmarshal(cell, cellPresence)
		if err != nil {
			logger.Error("failed-to-unmarshal-cells-json", err)
			continue
		}
		err = cellPresence.Validate()
		if err != nil {
			logger.Error("invalid-cell-presence", err)
			continue
		}

		cellPresences = append(cellPresences, cellPresence)
	}

	return cellPresences, nil
}

func (db *ConsulDB) CellById(logger lager.Logger, cellId string) (*models.CellPresence, error) {
	cellPresence := new(models.CellPresence)

	value, err := db.session.GetAcquiredValue(CellSchemaPath(cellId))
	if err != nil {
		return nil, convertConsulError(err)
	}

	err = json.Unmarshal(value, cellPresence)
	if err != nil {
		return nil, models.NewError(models.Error_InvalidJSON, err.Error())
	}

	err = cellPresence.Validate()
	if err != nil {
		return nil, models.NewError(models.Error_InvalidJSON, err.Error())
	}

	return cellPresence, nil
}

func convertConsulError(err error) error {
	switch err.(type) {
	case consuladapter.KeyNotFoundError:
		return models.ErrResourceNotFound
	case consuladapter.PrefixNotFoundError:
		return models.ErrResourceNotFound
	default:
		return models.ErrUnknownError
	}
}
