package migrations

import (
	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/migration"
	"github.com/pivotal-golang/lager"
)

func init() {
	appendMigration(NewTestMigration(9999999999))
}

type TestMigration struct {
	version     int64
	storeClient etcd.StoreClient
}

func NewTestMigration(version int64) migration.Migration {
	return &TestMigration{
		version: version,
	}
}

func (t *TestMigration) SetStoreClient(storeClient etcd.StoreClient) {
	t.storeClient = storeClient
}

func (t TestMigration) Up(logger lager.Logger) error {
	_, err := t.storeClient.Create("/test/key", []byte("jim is awesome"), etcd.NO_TTL)
	return err
}

func (t TestMigration) Down(logger lager.Logger) error {
	// do nothing until we get rollback
	return nil
}

func (t TestMigration) Version() int64 {
	return t.version
}
