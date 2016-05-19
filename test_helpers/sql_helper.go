package test_helpers

import (
	"fmt"
	"os"
	"strings"
)

func UseSQL() bool {
	return os.Getenv("USE_SQL") == "true"
}

func UsePostgres() bool {
	return os.Getenv("USE_POSTGRES") == "true"
}

func ReplaceQuestionMarks(queryString string) string {
	strParts := strings.Split(queryString, "?")
	for i := 1; i < len(strParts); i++ {
		strParts[i-1] = fmt.Sprintf("%s$%d", strParts[i-1], i)
	}
	return strings.Join(strParts, "")
}
