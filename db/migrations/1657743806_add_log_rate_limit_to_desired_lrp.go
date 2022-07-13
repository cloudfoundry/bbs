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
	appendMigration(NewAddLogRateLimitToDesiredLrp())
}

type AddLogRateLimitToDesiredLrp struct {
	serializer format.Serializer
	clock      clock.Clock
	rawSQLDB   *sql.DB
	dbFlavor   string
}

func NewAddLogRateLimitToDesiredLrp() migration.Migration {
	return new(AddLogRateLimitToDesiredLrp)
}

func (e *AddLogRateLimitToDesiredLrp) String() string {
	return migrationString(e)
}

func (e *AddLogRateLimitToDesiredLrp) Version() int64 {
	return 1657743806
}

func (e *AddLogRateLimitToDesiredLrp) SetCryptor(cryptor encryption.Cryptor) {
	e.serializer = format.NewSerializer(cryptor)
}

func (e *AddLogRateLimitToDesiredLrp) SetRawSQLDB(db *sql.DB)    { e.rawSQLDB = db }
func (e *AddLogRateLimitToDesiredLrp) SetClock(c clock.Clock)    { e.clock = c }
func (e *AddLogRateLimitToDesiredLrp) SetDBFlavor(flavor string) { e.dbFlavor = flavor }

func (e *AddLogRateLimitToDesiredLrp) Up(logger lager.Logger) error {
	logger = logger.Session("add-log-rate-limit")
	logger.Info("starting")
	defer logger.Info("completed")

	alterTableSQL := "ALTER TABLE desired_lrps ADD COLUMN log_rate_limit BIGINT;"
	logger.Info("altering the table", lager.Data{"query": alterTableSQL})
	_, err := e.rawSQLDB.Exec(helpers.RebindForFlavor(alterTableSQL, e.dbFlavor))
	if err != nil {
		logger.Error("failed-altering-table", err)
		return err
	}
	logger.Info("altered the table", lager.Data{"query": alterTableSQL})

	return nil
}
