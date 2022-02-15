package migrations

import (
	"database/sql"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/encryption"
	"code.cloudfoundry.org/bbs/format"
	"code.cloudfoundry.org/bbs/migration"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"
)

func init() {
	appendMigration(NewAddInternalRoutesToActualLrp())
}

type AddInternalRoutesToActualLrp struct {
	serializer format.Serializer
	clock      clock.Clock
	rawSQLDB   *sql.DB
	dbFlavor   string
}

func NewAddInternalRoutesToActualLrp() migration.Migration {
	return new(AddInternalRoutesToActualLrp)
}

func (e *AddInternalRoutesToActualLrp) String() string {
	return migrationString(e)
}

func (e *AddInternalRoutesToActualLrp) Version() int64 {
	return 1643660541
}

func (e *AddInternalRoutesToActualLrp) SetCryptor(cryptor encryption.Cryptor) {
	e.serializer = format.NewSerializer(cryptor)
}

func (e *AddInternalRoutesToActualLrp) SetRawSQLDB(db *sql.DB)    { e.rawSQLDB = db }
func (e *AddInternalRoutesToActualLrp) SetClock(c clock.Clock)    { e.clock = c }
func (e *AddInternalRoutesToActualLrp) SetDBFlavor(flavor string) { e.dbFlavor = flavor }

func (e *AddInternalRoutesToActualLrp) Up(logger lager.Logger) error {
	logger = logger.Session("add-internal-routes")
	logger.Info("starting")
	defer logger.Info("completed")

	alterTableSQL := "ALTER TABLE actual_lrps ADD COLUMN internal_routes MEDIUMTEXT;"
	logger.Info("altering the table", lager.Data{"query": alterTableSQL})
	_, err := e.rawSQLDB.Exec(helpers.RebindForFlavor(alterTableSQL, e.dbFlavor))
	if err != nil {
		logger.Error("failed-altering-table", err)
		return err
	}
	logger.Info("altered the table", lager.Data{"query": alterTableSQL})

	return nil
}
