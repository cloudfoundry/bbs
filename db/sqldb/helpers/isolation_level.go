package helpers

import (
	"database/sql"
	"fmt"

	"code.cloudfoundry.org/lager"
)

const (
	IsolationLevelReadUncommitted = "READ UNCOMMITTED"
	IsolationLevelReadCommitted   = "READ COMMITTED"
	IsolationLevelSerializable    = "SERIALIZABLE"
	IsolationLevelRepeatableRead  = "REPEATABLE READ"
)

// mysql: SET SESSION TRANSACTION ISOLATION LEVEL level;
// postgres: SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL level;
func (h *sqlHelper) SetIsolationLevel(logger lager.Logger, db *sql.DB, level string) error {
	logger = logger.Session("set-isolation-level", lager.Data{"level": level})
	logger.Info("starting")
	defer logger.Info("done")

	var query string
	if h.flavor == MySQL {
		query = fmt.Sprintf("SET SESSION TRANSACTION ISOLATION LEVEL %s", level)
	} else if h.flavor == Postgres {
		query = fmt.Sprintf("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL %s", level)
	} else if h.flavor == MSSQL {
		query = fmt.Sprintf("SET TRANSACTION ISOLATION LEVEL %s", level)
	} else {
		panic("database flavor not implemented: " + h.flavor)
	}

	_, err := db.Exec(query)
	return err
}
