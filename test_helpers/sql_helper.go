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

func UseMySQL() bool {
	return os.Getenv("USE_SQL") == "mysql"
}

func UsePostgres() bool {
	return os.Getenv("USE_SQL") == "postgres"
}

func NewSQLRunner(dbName string) sqlrunner.SQLRunner {
	var sqlRunner sqlrunner.SQLRunner

	if UseMySQL() {
		sqlRunner = sqlrunner.NewMySQLRunner(dbName)
	} else if UsePostgres() {
		sqlRunner = sqlrunner.NewPostgresRunner(dbName)
	} else {
		panic("driver not supported")
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
