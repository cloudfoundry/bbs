package helpers

import (
	"database/sql"
	"fmt"
	"strings"

	"code.cloudfoundry.org/lager"
)

const (
	MySQL    = "mysql"
	Postgres = "postgres"
	MSSQL    = "mssql"

	LockRow   RowLock = true
	NoLockRow RowLock = false
)

type SQLHelper interface {
	SetIsolationLevel(logger lager.Logger, db *sql.DB, level string) error

	Transact(logger lager.Logger, db *sql.DB, f func(logger lager.Logger, tx *sql.Tx) error) error
	One(logger lager.Logger, q Queryable, table string, columns ColumnList, lockRow RowLock, wheres string, whereBindings ...interface{}) *sql.Row
	All(logger lager.Logger, q Queryable, table string, columns ColumnList, lockRow RowLock, wheres string, whereBindings ...interface{}) (*sql.Rows, error)
	Upsert(logger lager.Logger, q Queryable, table string, keyAttributes, updateAttributes SQLAttributes) (sql.Result, error)
	Insert(logger lager.Logger, q Queryable, table string, attributes SQLAttributes) (sql.Result, error)
	Update(logger lager.Logger, q Queryable, table string, updates SQLAttributes, wheres string, whereBindings ...interface{}) (sql.Result, error)
	Delete(logger lager.Logger, q Queryable, table string, wheres string, whereBindings ...interface{}) (sql.Result, error)

	ConvertSQLError(err error) error
	Rebind(query string) string
}

type sqlHelper struct {
	flavor string
}

func NewSQLHelper(flavor string) *sqlHelper {
	return &sqlHelper{flavor: flavor}
}

type Queryable interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

type RowLock bool
type SQLAttributes map[string]interface{}
type ColumnList []string

func (h *sqlHelper) Rebind(query string) string {
	return RebindForFlavor(query, h.flavor)
}

func RebindForFlavor(query, flavor string) string {
	switch flavor {
	case MySQL:
		return query
	case Postgres:
		strParts := strings.Split(query, "?")
		for i := 1; i < len(strParts); i++ {
			strParts[i-1] = fmt.Sprintf("%s$%d", strParts[i-1], i)
		}
		return strings.Replace(strings.Join(strParts, ""), "MEDIUMTEXT", "TEXT", -1)
	case MSSQL:
		query = strings.Replace(query, "MEDIUMTEXT", "NVARCHAR(MAX)", -1)
		query = strings.Replace(query, "TEXT", "NVARCHAR(MAX)", -1)
		query = strings.Replace(query, "BOOL DEFAULT false", "TINYINT DEFAULT 0", -1)
		query = strings.Replace(query, "BOOL DEFAULT true", "TINYINT DEFAULT 1", -1)
		query = strings.Replace(query, "BOOL", "TINYINT", -1)
		query = strings.Replace(query, "ADD COLUMN", "ADD", -1)
		return query
	default:
		panic("database flavor not implemented: " + flavor)
	}
}

func QuestionMarks(count int) string {
	if count == 0 {
		return ""
	}
	return strings.Repeat("?, ", count-1) + "?"
}
