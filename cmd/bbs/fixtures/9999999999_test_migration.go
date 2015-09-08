package migrations

import (
	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/pivotal-golang/lager"
)

func init() {
	appendMigration(NewTestMigration(9999999999))
}

type TestMigration struct {
	version int64
}

func NewTestMigration(version int64) TestMigration {
	return TestMigration{
		version: version,
	}
}

func (t TestMigration) Up(logger lager.Logger, storeClient etcd.StoreClient) error {
	_, err := storeClient.Create("/test/key", []byte("jim is awesome"), etcd.NO_TTL)
	return err
}

func (t TestMigration) Down(logger lager.Logger, storeClient etcd.StoreClient) error {
	// do nothing until we get rollback
	return nil
}

func (t TestMigration) Version() int64 {
	return t.version
}
