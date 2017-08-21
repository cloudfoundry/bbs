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
	AppendMigration(NewActualLRP())
}

type ActualLRP struct {
	serializer  format.Serializer
	storeClient etcd.StoreClient
	clock       clock.Clock
	rawSQLDB    *sql.DB
	dbFlavor    string
}

func NewActualLRP() migration.Migration {
	return &ActualLRP{}
}

func (e *ActualLRP) Up(logger lager.Logger) error {
	logger.Info("altering the table", lager.Data{"query": alterActualLRPAddDeploymentGUID})
	_, err := e.rawSQLDB.Exec(alterActualLRPAddDeploymentGUID)
	if err != nil {
		logger.Error("failed-altering-tables", err)
		return err
	}
	logger.Info("altered the table", lager.Data{"query": alterActualLRPAddDeploymentGUID})

	return nil
}

const alterActualLRPAddDeploymentGUID = `ALTER TABLE actual_lrps
	ADD COLUMN lrp_deployment_guid VARCHAR(255);`

func (e *ActualLRP) createIndices(logger lager.Logger, db *sql.DB) error {
	return nil
}

func (e *ActualLRP) Down(logger lager.Logger) error {
	return errors.New("not implemented")
}

func (e *ActualLRP) RequiresSQL() bool         { return true }
func (e *ActualLRP) SetClock(c clock.Clock)    { e.clock = c }
func (e *ActualLRP) SetDBFlavor(flavor string) { e.dbFlavor = flavor }

func (e *ActualLRP) SetStoreClient(storeClient etcd.StoreClient) {
	e.storeClient = storeClient
}

func (e *ActualLRP) SetCryptor(cryptor encryption.Cryptor) {
	e.serializer = format.NewSerializer(cryptor)
}

func (e *ActualLRP) SetRawSQLDB(db *sql.DB) {
	e.rawSQLDB = db
}

func (e *ActualLRP) Version() int64 {
	return 1503338576
}

// TODO: add primary key on lrp deployment definition = (defintiion_guid + process_guid)
// process guid should have a foreign key constraint against LRP deployment
// Should we just rely on the client's providing a globally unqiue definition_guid?
