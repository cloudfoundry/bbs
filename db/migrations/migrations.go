package migrations

import (
	"fmt"

	"code.cloudfoundry.org/bbs/migration"
)

var Migrations = []migration.Migration{}

func AppendMigration(migration migration.Migration) {
	for _, m := range Migrations {
		if m.Version() == migration.Version() {
			panic(fmt.Sprintf("cannot have two migrations with the same version. %T & %T", m, migration))
		}
	}

	Migrations = append(Migrations, migration)
}
