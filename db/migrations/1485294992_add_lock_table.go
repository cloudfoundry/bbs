package migrations

import (
	"database/sql"
	"errors"

	"code.cloudfoundry.org/bbs/db/etcd"
	"code.cloudfoundry.org/bbs/encryption"
	"code.cloudfoundry.org/bbs/format"
	"code.cloudfoundry.org/bbs/migration"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"
)

func init() {
	AppendMigration(NewAddLockTable())
}

type AddLockTable struct {
	clock       clock.Clock
	dbFlavor    string
	serializer  format.Serializer
	rawSQLDB    *sql.DB
	storeClient etcd.StoreClient
}

func NewAddLockTable() migration.Migration {
	return &AddLockTable{}
}

func (*AddLockTable) String() string {
	return "1485294992"
}

func (*AddLockTable) Version() int64 {
	return 1485294992
}

func (a *AddLockTable) SetStoreClient(storeClient etcd.StoreClient) {
	a.storeClient = storeClient
}

func (a *AddLockTable) SetCryptor(cryptor encryption.Cryptor) {
	a.serializer = format.NewSerializer(cryptor)
}

func (a *AddLockTable) SetRawSQLDB(db *sql.DB)    { a.rawSQLDB = db }
func (*AddLockTable) RequiresSQL() bool           { return true }
func (a *AddLockTable) SetClock(c clock.Clock)    { a.clock = c }
func (a *AddLockTable) SetDBFlavor(flavor string) { a.dbFlavor = flavor }

func (a *AddLockTable) Up(logger lager.Logger) error {
	_, err := a.rawSQLDB.Exec(dropLockTable)
	if err != nil {
		logger.Debug("failed-to-drop-table")
	}

	_, err = a.rawSQLDB.Exec(createLockTable)
	if err != nil {
		logger.Error("failed-to-create-lock-table", err)
	}

	return err
}

func (*AddLockTable) Down(logger lager.Logger) error {
	return errors.New("not implemented")
}

const createLockTable = `CREATE TABLE locks (
	key VARCHAR(255) PRIMARY KEY,
	owner VARCHAR(255),
	value VARCHAR(255)
);`

const dropLockTable = `DROP TABLE locks;`
