## BBS Migrations

### How it works

#### Schema versions

TODO: how do we keep track of the version in the version configuration record

### Writing a migration

#### Migration Requirements

Migrations are expected to be idempotent, meaning that when a migration of version N is applied against schema version N - no change occurs. There is no expectation of or requirement for migrations to be interchangeable, meaning migrations are not expected to run against any schema versions other than N or N - 1 (where N is the schema version defined by the current migration).  Additionally, rollbacks and "down" migrations are not supported.

#### Creating a migration

use `scripts/make_migration.sh <name>` to create a new migration; where `<name>` is a descriptive name of what the migration with no spaces.

TODO: explain what this does and improve the scripts to generate a migration skeleton. also make sure that this is higher than any migration

#### Testing your migration

TODO: add a test suite for all migrations to ensure they are idempotent

**Note** `scripts/make_migration.sh` generate a test skeleton for the migration
