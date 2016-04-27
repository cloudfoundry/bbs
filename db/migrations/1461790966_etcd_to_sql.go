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
	AppendMigration(NewETCDToSQL())
}

type ETCDToSQL struct {
	serializer  format.Serializer
	storeClient etcd.StoreClient
	clock       clock.Clock
	rawSQLDB    *sql.DB
}

func NewETCDToSQL() migration.Migration {
	return &ETCDToSQL{}
}

func (e *ETCDToSQL) Version() int64 {
	return 1461790966
}

func (e *ETCDToSQL) SetStoreClient(storeClient etcd.StoreClient) {
	e.storeClient = storeClient
}

func (e *ETCDToSQL) SetCryptor(cryptor encryption.Cryptor) {
	e.serializer = format.NewSerializer(cryptor)
}

func (e *ETCDToSQL) SetClock(c clock.Clock) {
	e.clock = c
}

func (e *ETCDToSQL) Up(logger lager.Logger) error {
	return nil
}

func (e *ETCDToSQL) Down(logger lager.Logger) error {
	return errors.New("not implemented")
}

func (e *ETCDToSQL) SetRawSQLDB(rawSQLDB *sql.DB) {
	e.rawSQLDB = rawSQLDB
}

func (e *ETCDToSQL) RequiresSQL() bool {
	return true
}
