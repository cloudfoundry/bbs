package sqldb

import (
	"database/sql"

	"github.com/cloudfoundry-incubator/bbs/encryption"
	"github.com/cloudfoundry-incubator/bbs/format"
	"github.com/cloudfoundry-incubator/bbs/guidprovider"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/go-sql-driver/mysql"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
)

type SQLDB struct {
	db                     *sql.DB
	convergenceWorkersSize int
	clock                  clock.Clock
	format                 *format.Format
	guidProvider           guidprovider.GUIDProvider
	serializer             format.Serializer
}

type RowScanner interface {
	Scan(dest ...interface{}) error
}

type Queryable interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

const (
	NoLock = iota
	LockForShare
	LockForUpdate
)

func NewSQLDB(
	db *sql.DB,
	convergenceWorkersSize int,
	serializationFormat *format.Format,
	cryptor encryption.Cryptor,
	guidProvider guidprovider.GUIDProvider,
	clock clock.Clock,
) *SQLDB {
	return &SQLDB{
		db: db,
		convergenceWorkersSize: convergenceWorkersSize,
		clock:        clock,
		format:       serializationFormat,
		guidProvider: guidProvider,
		serializer:   format.NewSerializer(cryptor),
	}
}

func (db *SQLDB) transact(logger lager.Logger, f func(logger lager.Logger, tx *sql.Tx) error) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = f(logger, tx)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (db *SQLDB) serializeModel(logger lager.Logger, model format.Versioner) ([]byte, error) {
	encodedPayload, err := db.serializer.Marshal(logger, db.format, model)
	if err != nil {
		logger.Error("failed-to-serialize-model", err)
		return nil, models.NewError(models.Error_InvalidRecord, err.Error())
	}
	return encodedPayload, nil
}

func (db *SQLDB) deserializeModel(logger lager.Logger, data []byte, model format.Versioner) error {
	err := db.serializer.Unmarshal(logger, data, model)
	if err != nil {
		logger.Error("failed-to-deserialize-model", err)
		return models.NewError(models.Error_InvalidRecord, err.Error())
	}
	return nil
}

func (db *SQLDB) convertSQLError(err error) *models.Error {
	if err != nil {
		switch err.(type) {
		case *mysql.MySQLError:
			return db.convertMySQLError(err.(*mysql.MySQLError))
		}
	}

	return models.ErrUnknownError
}

func (db *SQLDB) convertMySQLError(err *mysql.MySQLError) *models.Error {
	switch err.Number {
	case 1062:
		return models.ErrResourceExists
	case 1406:
		return models.ErrBadRequest
	default:
		return models.ErrUnknownError
	}
}
