package internal

import (
	"code.cloudfoundry.org/bbs/format"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

func SerializeModel(logger lager.Logger, serializer format.Serializer, model format.Model) ([]byte, error) {
	encodedPayload, err := serializer.Marshal(logger, model)
	if err != nil {
		logger.Error("failed-to-serialize-model", err)
		return nil, models.NewError(models.Error_InvalidRecord, err.Error())
	}
	return encodedPayload, nil
}

func DeserializeModel(logger lager.Logger, serializer format.Serializer, data []byte, model format.Model) error {
	err := serializer.Unmarshal(logger, data, model)
	if err != nil {
		logger.Error("failed-to-deserialize-model", err)
		return models.NewError(models.Error_InvalidRecord, err.Error())
	}
	return nil
}
