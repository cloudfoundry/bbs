package migrations

import "github.com/cloudfoundry-incubator/bbs/migration"

var Migrations = []migration.Migration{}

func appendMigration(migration migration.Migration) {
	Migrations = append(Migrations, migration)
}
