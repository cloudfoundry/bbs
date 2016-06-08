package test_helpers

import (
	"fmt"
	"os"
	"strings"

	"github.com/cloudfoundry/storeadapter/storerunner/mysqlrunner"
	"github.com/cloudfoundry/storeadapter/storerunner/postgresrunner"
	"github.com/cloudfoundry/storeadapter/storerunner/sqlrunner"
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
		sqlRunner = mysqlrunner.NewMySQLRunner(dbName)
	} else if UsePostgres() {
		sqlRunner = postgresrunner.NewPostgresRunner(dbName)
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
