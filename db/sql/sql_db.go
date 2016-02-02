package sqldb

import (
	"github.com/jmoiron/sqlx"

	"github.com/cloudfoundry-incubator/bbs/db/etcd"

	_ "github.com/lib/pq"
)

type SQLDB struct {
	sql    *sqlx.DB
	etcdDB *etcd.ETCDDB
}

func NewSQLDB(etcdDB *etcd.ETCDDB) *SQLDB {
	db, err := sqlx.Open("postgres", "user=pqgotest dbname=pqgotest sslmode=verify-full")
	if err != nil {
		panic(err)
	}

	return &SQLDB{sql: db, etcdDB: etcdDB}
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
