package sqldb

import (
	"database/sql"

	"github.com/pivotal-golang/clock"
)

type SQLDB struct {
	db    *sql.DB
	clock clock.Clock
}

func NewSQLDB(db *sql.DB, clock clock.Clock) *SQLDB {
	return &SQLDB{
		db:    db,
		clock: clock,
	}
}
