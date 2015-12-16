package migrations

import (
	"errors"

	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/encryption"
	"github.com/cloudfoundry-incubator/bbs/format"
	"github.com/cloudfoundry-incubator/bbs/migration"
	"github.com/pivotal-golang/lager"
)

// null migration to bump the database version
func init() {
	AppendMigration(NewAddCacheDependencies())
}

type AddCacheDependencies struct {
	serializer  format.Serializer
	storeClient etcd.StoreClient
}

func NewAddCacheDependencies() migration.Migration {
	return &AddCacheDependencies{}
}

func (a AddCacheDependencies) Version() int64 {
	return 1450292094
}

func (a *AddCacheDependencies) SetStoreClient(storeClient etcd.StoreClient) {
	a.storeClient = storeClient
}

func (a *AddCacheDependencies) SetCryptor(cryptor encryption.Cryptor) {
	a.serializer = format.NewSerializer(cryptor)
}

func (a AddCacheDependencies) Up(logger lager.Logger) error {
	return nil
}

func (a AddCacheDependencies) Down(logger lager.Logger) error {
	return errors.New("not implemented")
}
