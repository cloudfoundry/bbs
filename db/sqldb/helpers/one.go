package helpers

import (
	"database/sql"
	"fmt"
	"strings"

	"code.cloudfoundry.org/lager"
)

// SELECT <columns> FROM <table> WHERE ... LIMIT 1 [FOR UPDATE]
func (h *sqlHelper) One(
	logger lager.Logger,
	q Queryable,
	table string,
	columns ColumnList,
	lockRow RowLock,
	wheres string,
	whereBindings ...interface{},
) *sql.Row {
	var query string

	if h.flavor == MSSQL {
		lockClause := ""
		if lockRow {
			lockClause = "WITH (UPDLOCK)"
		}
		query = fmt.Sprintf("SELECT TOP 1 %s FROM %s %s\n", strings.Join(columns, ", "), table, lockClause)

		if len(wheres) > 0 {
			query += "WHERE " + wheres
		}
	} else {
		query = fmt.Sprintf("SELECT %s FROM %s\n", strings.Join(columns, ", "), table)

		if len(wheres) > 0 {
			query += "WHERE " + wheres
		}

		query += "\nLIMIT 1"

		if lockRow {
			query += "\nFOR UPDATE"
		}
	}
	return q.QueryRow(h.Rebind(query), whereBindings...)
}
