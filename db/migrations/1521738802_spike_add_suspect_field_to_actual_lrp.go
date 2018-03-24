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
	appendMigration(NewSpikeAddSuspectFieldToActualLrp())
}

type SpikeAddSuspectFieldToActualLrp struct {
	serializer format.Serializer
	clock      clock.Clock
	rawSQLDB   *sql.DB
	dbFlavor   string
}

func NewSpikeAddSuspectFieldToActualLrp() migration.Migration {
	return new(SpikeAddSuspectFieldToActualLrp)
}

func (e *SpikeAddSuspectFieldToActualLrp) String() string {
	return migrationString(e)
}

func (e *SpikeAddSuspectFieldToActualLrp) Version() int64 {
	return 1521738802
}

func (e *SpikeAddSuspectFieldToActualLrp) SetCryptor(cryptor encryption.Cryptor) {
	e.serializer = format.NewSerializer(cryptor)
}

func (e *SpikeAddSuspectFieldToActualLrp) SetRawSQLDB(db *sql.DB)    { e.rawSQLDB = db }
func (e *SpikeAddSuspectFieldToActualLrp) SetClock(c clock.Clock)    { e.clock = c }
func (e *SpikeAddSuspectFieldToActualLrp) SetDBFlavor(flavor string) { e.dbFlavor = flavor }

func (e *SpikeAddSuspectFieldToActualLrp) Up(logger lager.Logger) error {
	_, err := e.rawSQLDB.Exec(addSuspectToActualLRP)
	if err != nil {
		logger.Error("failed-altering-table", err)
		return err
	}

	_, err = e.rawSQLDB.Exec(dropPrimaryKey)
	if err != nil {
		logger.Error("failed-dropping-primary-key", err)
		return err
	}

	_, err = e.rawSQLDB.Exec(updatePrimaryKey)
	if err != nil {
		logger.Error("failed-altering-primary-key", err)
		return err
	}

	return nil
}

const addSuspectToActualLRP = `ALTER TABLE actual_lrps ADD COLUMN suspect BOOL DEFAULT false;`

const dropPrimaryKey = `ALTER TABLE actual_lrps DROP CONSTRAINT actual_lrps_pkey;`

const updatePrimaryKey = `ALTER TABLE actual_lrps ADD PRIMARY KEY(process_guid, instance_index, evacuating, suspect);`
