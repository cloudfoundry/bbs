package migrations

import (
	"database/sql"
	"errors"

	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/encryption"
	"github.com/cloudfoundry-incubator/bbs/format"
	"github.com/cloudfoundry-incubator/bbs/migration"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
)

func init() {
	AppendMigration(NewTimeoutMilliseconds())
}

type TimeoutToMilliseconds struct {
	serializer  format.Serializer
	storeClient etcd.StoreClient
}

func NewTimeoutMilliseconds() migration.Migration {
	return &TimeoutToMilliseconds{}
}

func (b *TimeoutToMilliseconds) Version() int64 {
	return 1451635200
}

func (b *TimeoutToMilliseconds) SetStoreClient(storeClient etcd.StoreClient) {
	b.storeClient = storeClient
}

func (b *TimeoutToMilliseconds) SetCryptor(cryptor encryption.Cryptor) {
	b.serializer = format.NewSerializer(cryptor)
}

func (b *TimeoutToMilliseconds) Up(logger lager.Logger) error {
	return nil
}

func (b *TimeoutToMilliseconds) Down(logger lager.Logger) error {
	return errors.New("not implemented")
}

func (b *TimeoutToMilliseconds) SetRawSQLDB(*sql.DB) {}

func (b *TimeoutToMilliseconds) SetClock(clock.Clock) {}

func (b *TimeoutToMilliseconds) RequiresSQL() bool {
	return false
}
