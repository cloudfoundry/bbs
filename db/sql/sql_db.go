package sqldb

import (
	"database/sql"

	"os"

	"github.com/cloudfoundry-incubator/auctioneer"
	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/encryption"
	"github.com/cloudfoundry-incubator/bbs/format"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"

	_ "github.com/go-sql-driver/mysql"
)

type SQLDB struct {
	sql               *sql.DB
	etcdDB            db.DB
	serializer        format.Serializer
	clock             clock.Clock
	format            *format.Format
	updateWorkersSize int
	auctioneerClient  auctioneer.Client
}

func NewSQLDB(cryptor encryption.Cryptor, etcdDB db.DB, auctioneerClient auctioneer.Client) *SQLDB {
	// db, err := sql.Open("postgres", "host=10.244.0.30 port=5524 user=ccadmin password=admin dbname=diego sslmode=disable")

	// "diego:buttercup@tcp(benchmark-diego.cwqji29dpdny.us-east-1.rds.amazonaws.com)/diego"
	dsn := os.Getenv("SQL_DSN")
	if dsn == "" {
		dsn = "root:password@tcp(10.244.7.6)/diego"
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	return &SQLDB{
		sql:               db,
		etcdDB:            etcdDB,
		serializer:        format.NewSerializer(cryptor),
		clock:             clock.NewClock(),
		format:            format.ENCRYPTED_PROTO,
		updateWorkersSize: 50,
		auctioneerClient:  auctioneerClient,
	}
}

func (db *SQLDB) serializeModel(logger lager.Logger, model format.Versioner) ([]byte, error) {
	encodedPayload, err := db.serializer.Marshal(logger, db.format, model)
	if err != nil {
		logger.Error("failed-to-serialize-model", err)
		return nil, models.NewError(models.Error_InvalidRecord, err.Error())
	}
	return encodedPayload, nil
}

func (db *SQLDB) deserializeModel(logger lager.Logger, rawModel string, model format.Versioner) error {
	err := db.serializer.Unmarshal(logger, []byte(rawModel), model)
	if err != nil {
		logger.Error("failed-to-deserialize-model", err)
		return models.NewError(models.Error_InvalidRecord, err.Error())
	}
	return nil
}

// func (db *SQLDB) EncryptionKeyLabel(logger lager.Logger) (string, error) {
// 	return db.etcdDB.EncryptionKeyLabel(logger)
// }

// func (db *SQLDB) SetEncryptionKeyLabel(logger lager.Logger, encryptionKeyLabel string) error {
// 	return db.etcdDB.SetEncryptionKeyLabel(logger, encryptionKeyLabel)
// }

// func (db *SQLDB) EvacuateClaimedActualLRP(logger lager.Logger, actual *models.ActualLRPKey, instance *models.ActualLRPInstanceKey) (bool, error) {
// 	return db.etcdDB.EvacuateClaimedActualLRP(logger, actual, instance)
// }
// func (db *SQLDB) EvacuateRunningActualLRP(logger lager.Logger, actual *models.ActualLRPKey, instance *models.ActualLRPInstanceKey, netInfo *models.ActualLRPNetInfo, uint64) (bool, error) {
// 	return db.etcdDB.EvacuateRunningActualLRP(logger, actual, instance)
// }
// func (db *SQLDB) EvacuateStoppedActualLRP(lager.Logger, *models.ActualLRPKey, *models.ActualLRPInstanceKey) (bool, error)
// func (db *SQLDB) EvacuateCrashedActualLRP(lager.Logger, *models.ActualLRPKey, *models.ActualLRPInstanceKey, string) (bool, error)
// func (db *SQLDB) RemoveEvacuatingActualLRP(lager.Logger, *models.ActualLRPKey, *models.ActualLRPInstanceKey) error
