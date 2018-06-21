package migrations

import (
	"database/sql"

	"code.cloudfoundry.org/bbs/encryption"
	"code.cloudfoundry.org/bbs/format"
	"code.cloudfoundry.org/bbs/migration"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"
)

func init() {
	appendMigration(NewAddPresenceToActualLrp())
}

type AddPresenceToActualLrp struct {
	serializer format.Serializer
	clock      clock.Clock
	rawSQLDB   *sql.DB
	dbFlavor   string
}

func NewAddPresenceToActualLrp() migration.Migration {
	return new(AddPresenceToActualLrp)
}

func (e *AddPresenceToActualLrp) String() string {
	return migrationString(e)
}

func (e *AddPresenceToActualLrp) Version() int64 {
	return 1529530809
}

func (e *AddPresenceToActualLrp) SetCryptor(cryptor encryption.Cryptor) {
	e.serializer = format.NewSerializer(cryptor)
}

func (e *AddPresenceToActualLrp) SetRawSQLDB(db *sql.DB)    { e.rawSQLDB = db }
func (e *AddPresenceToActualLrp) SetClock(c clock.Clock)    { e.clock = c }
func (e *AddPresenceToActualLrp) SetDBFlavor(flavor string) { e.dbFlavor = flavor }

func (e *AddPresenceToActualLrp) Up(logger lager.Logger) error {
	logger = logger.Session("add-presence")
	logger.Info("starting")
	defer logger.Info("completed")

	const query = "ALTER TABLE actual_lrps ADD COLUMN presence VARCHAR(255) NOT NULL DEFAULT '';"
	_, err := e.rawSQLDB.Exec(query)
	if err != nil {
		logger.Error("failed-altering-table", err)
		return err
	}

	return nil
}
