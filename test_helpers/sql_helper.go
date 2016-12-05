package test_helpers

import (
	"fmt"
	"os"
	"strings"

	"code.cloudfoundry.org/bbs/test_helpers/sqlrunner"
)

func UseSQL() bool {
	return UseMySQL() || UsePostgres()
}

func driver() string {
	return os.Getenv("SQL_FLAVOR")
}

func UseMySQL() bool {
	return driver() == "mysql"
}

func UsePostgres() bool {
	return driver() == "postgres"
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
