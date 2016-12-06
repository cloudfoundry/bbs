package test_helpers

import (
	"fmt"
	"os"
	"strings"

	"code.cloudfoundry.org/bbs/test_helpers/sqlrunner"
)

const (
	mysql_flavor    = "mysql"
	postgres_flavor = "postgres"
)

func UseSQL() bool {
	return true
}

func driver() string {
	flavor := os.Getenv("SQL_FLAVOR")
	if flavor == "" {
		flavor = postgres_flavor
	}
	return flavor
}

func UseMySQL() bool {
	return driver() == mysql_flavor
}

func UsePostgres() bool {
	return driver() == postgres_flavor
}

func NewSQLRunner(dbName string) sqlrunner.SQLRunner {
	var sqlRunner sqlrunner.SQLRunner

	if UseMySQL() {
		sqlRunner = sqlrunner.NewMySQLRunner(dbName)
	} else if UsePostgres() {
		sqlRunner = sqlrunner.NewPostgresRunner(dbName)
	} else {
		panic(fmt.Sprintf("driver '%s' is not supported", driver()))
	}

	return sqlRunner
}

func ReplaceQuestionMarks(queryString string) string {
	strParts := strings.Split(queryString, "?")
	for i := 1; i < len(strParts); i++ {
		strParts[i-1] = fmt.Sprintf("%s$%d", strParts[i-1], i)
	}
	return strings.Join(strParts, "")
}
