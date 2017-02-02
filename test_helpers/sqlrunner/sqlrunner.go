package sqlrunner

import (
	"database/sql"

	"github.com/tedsuo/ifrit"
)

type SQLRunner interface {
	ifrit.Runner
	ConnectionString() string
	Reset()
	ResetTables(tables []string)
	DriverName() string
	DB() *sql.DB
}
