package migrations

import (
	"database/sql"

	"code.cloudfoundry.org/bbs/encryption"
	"code.cloudfoundry.org/bbs/format"
	"code.cloudfoundry.org/bbs/migration"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager/v3"
)

func init() {
	appendMigration(NewAddRoutableToActualLrps())
}

type AddRoutableToActualLrps struct {
	serializer format.Serializer
	clock      clock.Clock
	rawSQLDB   *sql.DB
	dbFlavor   string
}

func NewAddRoutableToActualLrps() migration.Migration {
	return new(AddRoutableToActualLrps)
}

func (e *AddRoutableToActualLrps) String() string {
	return migrationString(e)
}

func (e *AddRoutableToActualLrps) Version() int64 {
	return 1686692176
}

func (e *AddRoutableToActualLrps) SetCryptor(cryptor encryption.Cryptor) {
	e.serializer = format.NewSerializer(cryptor)
}

func (e *AddRoutableToActualLrps) SetRawSQLDB(db *sql.DB)    { e.rawSQLDB = db }
func (e *AddRoutableToActualLrps) SetClock(c clock.Clock)    { e.clock = c }
func (e *AddRoutableToActualLrps) SetDBFlavor(flavor string) { e.dbFlavor = flavor }

func (e *AddRoutableToActualLrps) Up(logger lager.Logger) error {
	// TODO: add migration code here
	logger.Info("altering the table", lager.Data{"query": alterActualLRPAddRoutableSQL})
	_, err := e.rawSQLDB.Exec(alterActualLRPAddRoutableSQL)
	if err != nil {
		logger.Error("failed-altering-tables", err)
		return err
	}
	logger.Info("altered the table", lager.Data{"query": alterActualLRPAddRoutableSQL})

	return nil
}

const alterActualLRPAddRoutableSQL = `ALTER TABLE actual_lrps
ADD COLUMN routable BOOL DEFAULT false;`
