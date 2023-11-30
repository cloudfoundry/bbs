package migrations

import (
	"database/sql"
	"strings"

	"code.cloudfoundry.org/bbs/encryption"
	"code.cloudfoundry.org/bbs/format"
	"code.cloudfoundry.org/bbs/migration"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager/v3"
)

func init() {
	appendMigration(NewAddMaxPidsToDesiredLRPs())
}

type AddMaxPidsToDesiredLRPs struct {
	serializer format.Serializer
	clock      clock.Clock
	dbFlavor   string
}

func NewAddMaxPidsToDesiredLRPs() migration.Migration {
	return &AddMaxPidsToDesiredLRPs{}
}

func (e *AddMaxPidsToDesiredLRPs) String() string {
	return migrationString(e)
}

func (e *AddMaxPidsToDesiredLRPs) Version() int64 {
	return 1481761088
}

func (e *AddMaxPidsToDesiredLRPs) SetCryptor(cryptor encryption.Cryptor) {
	e.serializer = format.NewSerializer(cryptor)
}

func (e *AddMaxPidsToDesiredLRPs) SetClock(c clock.Clock)    { e.clock = c }
func (e *AddMaxPidsToDesiredLRPs) SetDBFlavor(flavor string) { e.dbFlavor = flavor }

func (e *AddMaxPidsToDesiredLRPs) Up(tx *sql.Tx, logger lager.Logger) error {
	_, err := tx.Exec(checkMaxPidsExistenceSQL)
	if err == nil {
		logger.Info("max-pid-already-available")
		return nil
	}

	if err != nil {
		if !strings.Contains(err.Error(), postgresColumnNotExistErr) && !strings.Contains(err.Error(), mysqlColumnNotExistErr) {
			logger.Error("failed-querying-desired-lrps", err)
			return err
		}
	}

	logger.Info("altering the table", lager.Data{"query": alterDesiredLRPAddMaxPidsSQL})
	_, err = tx.Exec(alterDesiredLRPAddMaxPidsSQL)
	if err != nil {
		logger.Error("failed-altering-tables", err)
		return err
	}
	logger.Info("altered the table", lager.Data{"query": alterDesiredLRPAddMaxPidsSQL})

	return nil
}

const postgresColumnNotExistErr = `"max_pids" does not exist`
const mysqlColumnNotExistErr = `Unknown column 'max_pids'`
const checkMaxPidsExistenceSQL = `SELECT count(max_pids) FROM desired_lrps`
const alterDesiredLRPAddMaxPidsSQL = `ALTER TABLE desired_lrps
	ADD COLUMN max_pids INTEGER DEFAULT 0;`
