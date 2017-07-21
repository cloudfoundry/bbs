package migrations

import (
	"database/sql"
	"errors"

	"code.cloudfoundry.org/bbs/db/etcd"
	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/encryption"
	"code.cloudfoundry.org/bbs/format"
	"code.cloudfoundry.org/bbs/migration"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"
)

func init() {
	AppendMigration(NewLRPDeployment())
}

type LRPDeployment struct {
	serializer  format.Serializer
	storeClient etcd.StoreClient
	clock       clock.Clock
	rawSQLDB    *sql.DB
	dbFlavor    string
}

func NewLRPDeployment() migration.Migration {
	return &LRPDeployment{}
}

func (e *LRPDeployment) Up(logger lager.Logger) error {
	logger = logger.Session("lrp-deployments")
	logger.Info("truncating-tables")

	// Ignore the error as the tables may not exist
	_ = e.dropTables(e.rawSQLDB)

	err := e.createTables(logger, e.rawSQLDB, e.dbFlavor)
	if err != nil {
		return err
	}

	err = e.createIndices(logger, e.rawSQLDB)
	if err != nil {
		return err
	}

	return nil
}

func (e *LRPDeployment) createTables(logger lager.Logger, db *sql.DB, flavor string) error {
	var createTablesSQL = []string{
		helpers.RebindForFlavor(createLRPDeploymentsSQL, flavor),
		helpers.RebindForFlavor(createLRPDeploymentDefinitionsSQL, flavor),
		helpers.RebindForFlavor(dropDesiredLRPSQL, flavor),
	}

	logger.Info("creating-tables")
	for _, query := range createTablesSQL {
		logger.Info("creating the table", lager.Data{"query": query})
		_, err := db.Exec(query)
		if err != nil {
			logger.Error("failed-creating-tables", err)
			return err
		}
		logger.Info("created the table", lager.Data{"query": query})
	}

	return nil
}

func (e *LRPDeployment) createIndices(logger lager.Logger, db *sql.DB) error {
	return nil
}

func (e *LRPDeployment) Down(logger lager.Logger) error {
	return errors.New("not implemented")
}

func (e *LRPDeployment) RequiresSQL() bool         { return true }
func (e *LRPDeployment) SetClock(c clock.Clock)    { e.clock = c }
func (e *LRPDeployment) SetDBFlavor(flavor string) { e.dbFlavor = flavor }

func (e *LRPDeployment) SetStoreClient(storeClient etcd.StoreClient) {
	e.storeClient = storeClient
}

func (e *LRPDeployment) SetCryptor(cryptor encryption.Cryptor) {
	e.serializer = format.NewSerializer(cryptor)
}

func (e *LRPDeployment) SetRawSQLDB(db *sql.DB) {
	e.rawSQLDB = db
}

func (e *LRPDeployment) Version() int64 {
	return 1500558085
}

func (e *LRPDeployment) dropTables(db *sql.DB) error {
	tableNames := []string{
		"lrp_deployments",
		"lrp_deployment_definitions",
	}
	for _, tableName := range tableNames {
		_, err := db.Exec("DROP TABLE IF EXISTS " + tableName)
		if err != nil {
			return err
		}
	}
	return nil
}

const dropDesiredLRPSQL = `DROP TABLE desired_lrps`

const createLRPDeploymentsSQL = `CREATE TABLE lrp_deployments(
	process_guid VARCHAR(255) PRIMARY KEY,
	domain VARCHAR(255) NOT NULL,
	instances INT NOT NULL,
	annotation MEDIUMTEXT,
	routes MEDIUMTEXT NOT NULL,
	active_definition_id VARCHAR(255),
	healthy_definition_id VARCHAR(255),
	modification_tag_epoch VARCHAR(255) NOT NULL,
	modification_tag_index INT
);`

const createLRPDeploymentDefinitionsSQL = `CREATE TABLE lrp_deployment_definitions(
	process_guid VARCHAR(255),
	definition_guid VARCHAR(255),
	definition_name VARCHAR(255),
	log_guid VARCHAR(255) NOT NULL,
	memory_mb INT NOT NULL,
	disk_mb INT NOT NULL,
	rootfs VARCHAR(255) NOT NULL,
	volume_placement MEDIUMTEXT NOT NULL,
	placement_tags TEXT,
	max_pids INTEGER DEFAULT 0,
	run_info MEDIUMTEXT NOT NULL
);`

// todo: add primary key on lrp deployment definition = (defintiion_guid + process_guid)
