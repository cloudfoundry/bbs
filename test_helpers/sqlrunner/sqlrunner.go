package sqlrunner

import (
	"database/sql"

	"github.com/tedsuo/ifrit"
)

type SQLRunner interface {
	ifrit.Runner
	ConnectionString() string
	Reset()
	DriverName() string
	DB() *sql.DB
}
