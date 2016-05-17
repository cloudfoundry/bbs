package test_helpers

import "os"

func UseSQL() bool {
	return os.Getenv("USE_SQL") == "true"
}
