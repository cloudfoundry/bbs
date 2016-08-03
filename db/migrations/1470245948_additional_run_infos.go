package migrations

import (
	"database/sql"
	"errors"

	"code.cloudfoundry.org/bbs/db/etcd"
	"code.cloudfoundry.org/bbs/encryption"
	"code.cloudfoundry.org/bbs/format"
	"code.cloudfoundry.org/bbs/migration"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"
)

func init() {
	AppendMigration(NewAdditionalRunInfos())
}

type AdditionalRunInfos struct {
	serializer  format.Serializer
	storeClient etcd.StoreClient
	clock       clock.Clock
	rawSQLDB    *sql.DB
	dbFlavor    string
}

func NewAdditionalRunInfos() migration.Migration {
	return &AdditionalRunInfos{}
}

func (e *AdditionalRunInfos) String() string {
	return "1470245948"
}

func (e *AdditionalRunInfos) Version() int64 {
	return 1470245948
}

func (e *AdditionalRunInfos) SetStoreClient(storeClient etcd.StoreClient) {
	e.storeClient = storeClient
}

func (e *AdditionalRunInfos) SetCryptor(cryptor encryption.Cryptor) {
	e.serializer = format.NewSerializer(cryptor)
}

func (e *AdditionalRunInfos) SetRawSQLDB(db *sql.DB) {
	e.rawSQLDB = db
}

func (e *AdditionalRunInfos) RequiresSQL() bool         { return true }
func (e *AdditionalRunInfos) SetClock(c clock.Clock)    { e.clock = c }
func (e *AdditionalRunInfos) SetDBFlavor(flavor string) { e.dbFlavor = flavor }

func (e *AdditionalRunInfos) Up(logger lager.Logger) error {
	logger = logger.Session("additional-run-infos")
	logger.Info("updating tables")

	err := createRunInfoTable(logger, e.rawSQLDB)
	if err != nil {
		return err
	}

	err = alterTables(logger, e.rawSQLDB)
	if err != nil {
		return err
	}

	err = migrateDesiredLRPRunInfos(logger, e.rawSQLDB)
	if err != nil {
		return err
	}

	err = updateRunInfoReferences(logger, e.rawSQLDB)
	if err != nil {
		return err
	}

	return nil
}

func (e *AdditionalRunInfos) Down(logger lager.Logger) error {
	return errors.New("not implemented")
}

func createRunInfoTable(logger lager.Logger, db *sql.DB) error {
	var createTablesSQL = []string{
		createRunInfoSQL,
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

func alterTables(logger lager.Logger, db *sql.DB) error {
	var alterTablesSQL = []string{
		alterDesiredLRPsSQL,
		alterActualLRPsSQL,
	}

	logger.Info("altering-tables")
	for _, query := range alterTablesSQL {
		logger.Info("altering the table", lager.Data{"query": query})
		_, err := db.Exec(query)
		if err != nil {
			logger.Error("failed-altering-tables", err)
			return err
		}
		logger.Info("altered the table", lager.Data{"query": query})
	}

	return nil
}

func migrateDesiredLRPRunInfos(logger lager.Logger, db *sql.DB) error {
	logger = logger.Session("migrating-desired-lrp")
	logger.Debug("starting")
	defer logger.Debug("finished")

	_, err := db.Exec(`
			INSERT INTO run_infos
				(guid, tag, data)
			VALUES
				(select process_guid, 'current', run_info from desired_lrps)
		`)
	if err != nil {
		logger.Error("failed-inserting-run_info", err)
	}
	return nil
}

func updateRunInfoReferences(logger lager.Logger, db *sql.DB) error {
	logger = logger.Session("migrating-actual-lrps")
	logger.Debug("starting")
	defer logger.Debug("finished")
	_, err := db.Exec(`
			UPDATE actual_lrps
				SET run_info_guid = process_guid
			`)
	if err != nil {
		logger.Error("failed-updating-run_info", err)
	}

	return nil
}

const createRunInfoSQL = `CREATE TABLE run_infos(
	guid VARCHAR(255) PRIMARY KEY,
	tag VARCHAR(255) NOT NULL,
	data TEXT NOT NULL
);`

const alterDesiredLRPsSQL = `ALTER TABLE desired_lrps
	ADD COLUMN run_info_guid VARCHAR(255),
	ADD COLUMN run_info_guid_1 VARCHAR(255),
	ADD COLUMN run_info_guid_2 VARCHAR(255)
;`

const alterActualLRPsSQL = `ALTER TABLE actual_lrps
	ADD COLUMN run_info_guid VARCHAR(255)
;`
