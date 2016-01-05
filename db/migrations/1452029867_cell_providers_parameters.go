package migrations

import (
	"errors"

	"github.com/cloudfoundry-incubator/bbs"
	"github.com/cloudfoundry-incubator/bbs/db/deprecations"
	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/encryption"
	"github.com/cloudfoundry-incubator/bbs/format"
	"github.com/cloudfoundry-incubator/bbs/migration"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/consuladapter"
	"github.com/pivotal-golang/lager"
)

func init() {
	AppendMigration(NewCellProvidersParameters())
}

type CellProvidersParameters struct {
	serializer    format.Serializer
	consulSession consuladapter.Session
}

func NewCellProvidersParameters() migration.Migration {
	return &CellProvidersParameters{}
}

func (m *CellProvidersParameters) Version() int64 {
	return 1452029867
}

func (m *CellProvidersParameters) SetStoreClient(storeClient etcd.StoreClient) {
}

func (m *CellProvidersParameters) SetConsulSession(consulSession consuladapter.Session) {
	m.consulSession = consulSession
}

func (m *CellProvidersParameters) SetCryptor(cryptor encryption.Cryptor) {
	m.serializer = format.NewSerializer(cryptor)
}

func (m *CellProvidersParameters) Up(logger lager.Logger) error {
	cells, err := m.consulSession.ListAcquiredValues(bbs.CellSchemaRoot())
	if err != nil {
		return err
	}

	for _, cell := range cells {
		presence := new(deprecations.V0CellPresence)
		err := models.FromJSON(cell, presence)
		if err != nil {
			logger.Error("failed-to-unmarshal-cells-json", err)
			continue
		}

		updatedProviders := make(models.RootFSProviders)
		for k, v := range presence.RootFSProviders {
			updatedProviders[k] = &models.Providers{Parameters: v}
		}
		updatedPresence := &models.CellPresence{
			Capacity: &models.CellCapacity{
				DiskMb:     int32(presence.Capacity.DiskMB),
				MemoryMb:   int32(presence.Capacity.MemoryMB),
				Containers: int32(presence.Capacity.Containers),
			},
			CellId:          presence.CellID,
			RepAddress:      presence.RepAddress,
			Zone:            presence.Zone,
			RootfsProviders: updatedProviders,
		}
		value, err := models.ToJSON(updatedPresence)
		if err != nil {
			logger.Error("failed-to-marshal-cells-json", err)
			continue
		}
		m.consulSession.SetPresence(bbs.CellSchemaPath(updatedPresence.CellId),
			value)
	}

	return nil
}

func (m *CellProvidersParameters) Down(logger lager.Logger) error {
	return errors.New("not implemented")
}
