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
	appendMigration(NewSetRunInfoLongtext())
}

type SetRunInfoLongtext struct {
	serializer format.Serializer
	clock      clock.Clock
	rawSQLDB   *sql.DB
	dbFlavor   string
}

func NewSetRunInfoLongtext() migration.Migration {
	return &SetRunInfoLongtext{}
}

func (e *SetRunInfoLongtext) String() string {
	return migrationString(e)
}

func (e *SetRunInfoLongtext) Version() int64 {
	return 1674146125
}

func (e *SetRunInfoLongtext) SetCryptor(cryptor encryption.Cryptor) {
	e.serializer = format.NewSerializer(cryptor)
}

func (e *SetRunInfoLongtext) SetRawSQLDB(db *sql.DB) {
	e.rawSQLDB = db
}

func (e *SetRunInfoLongtext) SetClock(c clock.Clock)    { e.clock = c }
func (e *SetRunInfoLongtext) SetDBFlavor(flavor string) { e.dbFlavor = flavor }

func (e *SetRunInfoLongtext) Up(logger lager.Logger) error {
	logger = logger.Session("set-run-info-longtext")
	logger.Info("starting")
	defer logger.Info("completed")

	return e.alterTables(logger, e.rawSQLDB, e.dbFlavor)
}

func (e *SetRunInfoLongtext) alterTables(logger lager.Logger, db *sql.DB, flavor string) error {
	if flavor != "mysql" {
		return nil
	}

	alterDesiredLRPsSQL := `ALTER TABLE desired_lrps
	MODIFY annotation LONGTEXT,
	MODIFY routes LONGTEXT NOT NULL,
	MODIFY volume_placement LONGTEXT NOT NULL,
	MODIFY run_info LONGTEXT NOT NULL;`

	alterActualLRPsSQL := `ALTER TABLE actual_lrps
	MODIFY net_info LONGTEXT NOT NULL;`

	alterTasksSQL := `ALTER TABLE tasks
	MODIFY result LONGTEXT,
	MODIFY task_definition LONGTEXT NOT NULL;`

	var alterTablesSQL = []string{
		alterDesiredLRPsSQL,
		alterActualLRPsSQL,
		alterTasksSQL,
	}

	logger.Info("altering-tables")
	for _, query := range alterTablesSQL {
		logger.Info("altering the table", lager.Data{"query": query})
		_, err := db.Exec(query)
		if err != nil {
			logger.Error("failed-altering-tables", err)
			return err
		}
		logger.Info("altered the table", lager.Data{"query": query})
	}

	return nil
}
