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
	appendMigration(NewAddTaskFailureCount())
}

type AddTaskFailureCount struct {
	serializer format.Serializer
	clock      clock.Clock
	rawSQLDB   *sql.DB
	dbFlavor   string
}

func NewAddTaskFailureCount() migration.Migration {
	return new(AddTaskFailureCount)
}

func (e *AddTaskFailureCount) String() string {
	return migrationString(e)
}

func (e *AddTaskFailureCount) Version() int64 {
	return 1521745876
}

func (e *AddTaskFailureCount) SetCryptor(cryptor encryption.Cryptor) {
	e.serializer = format.NewSerializer(cryptor)
}

func (e *AddTaskFailureCount) SetRawSQLDB(db *sql.DB)    { e.rawSQLDB = db }
func (e *AddTaskFailureCount) SetClock(c clock.Clock)    { e.clock = c }
func (e *AddTaskFailureCount) SetDBFlavor(flavor string) { e.dbFlavor = flavor }

func (e *AddTaskFailureCount) Up(logger lager.Logger) error {
	query := helpers.RebindForFlavor(addTaskFailureReasonAndCountQuery, e.dbFlavor)
	_, err := e.rawSQLDB.Exec(query)
	return err
}

const (
	addTaskFailureReasonAndCountQuery = `
ALTER TABLE tasks ADD failure_count INT NOT NULL DEFAULT 0;
`
)
