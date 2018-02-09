package migrations

import (
	"reflect"

	"code.cloudfoundry.org/bbs/migration"
)

var migrationsRegistry = migration.Migrations{}

func appendMigration(migrationTemplate migration.Migration) {
	migrationsRegistry = append(migrationsRegistry, migrationTemplate)
}

func AllMigrations() migration.Migrations {
	migs := make(migration.Migrations, len(migrationsRegistry))
	for i, mig := range migrationsRegistry {
		rt := reflect.TypeOf(mig)
		if rt.Kind() == reflect.Ptr {
			rt = rt.Elem()
		}
		migs[i] = reflect.New(rt).Interface().(migration.Migration)
	}
	return migs
}
