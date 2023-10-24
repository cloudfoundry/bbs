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
	appendMigration(NewAddAvailabilityZoneToActualLrps())
}

type AddAvailabilityZoneToActualLrps struct {
	serializer format.Serializer
	clock      clock.Clock
	rawSQLDB   *sql.DB
	dbFlavor   string
}

func NewAddAvailabilityZoneToActualLrps() migration.Migration {
	return new(AddAvailabilityZoneToActualLrps)
}

func (e *AddAvailabilityZoneToActualLrps) String() string {
	return migrationString(e)
}

func (e *AddAvailabilityZoneToActualLrps) Version() int64 {
	return 1698182853
}

func (e *AddAvailabilityZoneToActualLrps) SetCryptor(cryptor encryption.Cryptor) {
	e.serializer = format.NewSerializer(cryptor)
}

func (e *AddAvailabilityZoneToActualLrps) SetRawSQLDB(db *sql.DB)    { e.rawSQLDB = db }
func (e *AddAvailabilityZoneToActualLrps) SetClock(c clock.Clock)    { e.clock = c }
func (e *AddAvailabilityZoneToActualLrps) SetDBFlavor(flavor string) { e.dbFlavor = flavor }

func (e *AddAvailabilityZoneToActualLrps) Up(logger lager.Logger) error {
	logger.Info("altering the table", lager.Data{"query": alterActualLRPAddAvailabilityZoneSQL})
	_, err := e.rawSQLDB.Exec(alterActualLRPAddAvailabilityZoneSQL)
	if err != nil {
		logger.Error("failed-altering-tables", err)
		return err
	}
	logger.Info("altered the table", lager.Data{"query": alterActualLRPAddAvailabilityZoneSQL})

	return nil
}

const alterActualLRPAddAvailabilityZoneSQL = `ALTER TABLE actual_lrps
ADD COLUMN availability_zone VARCHAR(255) NOT NULL DEFAULT '';`
