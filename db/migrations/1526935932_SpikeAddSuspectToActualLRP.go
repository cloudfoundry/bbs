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
	appendMigration(NewSpikeAddSuspectToActualLRP())
}

type SpikeAddSuspectToActualLRP struct {
	serializer format.Serializer
	clock      clock.Clock
	rawSQLDB   *sql.DB
	dbFlavor   string
}

func NewSpikeAddSuspectToActualLRP() migration.Migration {
	return new(SpikeAddSuspectToActualLRP)
}

func (e *SpikeAddSuspectToActualLRP) String() string {
	return migrationString(e)
}

func (e *SpikeAddSuspectToActualLRP) Version() int64 {
	return 1526935932
}

func (e *SpikeAddSuspectToActualLRP) SetCryptor(cryptor encryption.Cryptor) {
	e.serializer = format.NewSerializer(cryptor)
}

func (e *SpikeAddSuspectToActualLRP) SetRawSQLDB(db *sql.DB)    { e.rawSQLDB = db }
func (e *SpikeAddSuspectToActualLRP) SetClock(c clock.Clock)    { e.clock = c }
func (e *SpikeAddSuspectToActualLRP) SetDBFlavor(flavor string) { e.dbFlavor = flavor }

func (e *SpikeAddSuspectToActualLRP) Up(logger lager.Logger) error {
	// TODO: add migration code here
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
