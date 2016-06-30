package sqldb

import (
	"database/sql"
	"time"

	"code.cloudfoundry.org/bbs/encryption"
	"code.cloudfoundry.org/bbs/format"
	"code.cloudfoundry.org/bbs/guidprovider"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"
	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
)

type SQLDB struct {
	db                     *sql.DB
	convergenceWorkersSize int
	updateWorkersSize      int
	clock                  clock.Clock
	format                 *format.Format
	guidProvider           guidprovider.GUIDProvider
	serializer             format.Serializer
	cryptor                encryption.Cryptor
	flavor                 string
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
	LockForUpdate
)

func NewSQLDB(
	db *sql.DB,
	convergenceWorkersSize int,
	updateWorkersSize int,
	serializationFormat *format.Format,
	cryptor encryption.Cryptor,
	guidProvider guidprovider.GUIDProvider,
	clock clock.Clock,
	flavor string,
) *SQLDB {
	return &SQLDB{
		db: db,
		convergenceWorkersSize: convergenceWorkersSize,
		updateWorkersSize:      updateWorkersSize,
		clock:                  clock,
		format:                 serializationFormat,
		guidProvider:           guidProvider,
		serializer:             format.NewSerializer(cryptor),
		cryptor:                cryptor,
		flavor:                 flavor,
	}
}

func (db *SQLDB) transact(logger lager.Logger, f func(logger lager.Logger, tx *sql.Tx) error) error {
	var err error

	for attempts := 0; attempts < 3; attempts++ {
		err = func() error {
			tx, err := db.db.Begin()
			if err != nil {
				return err
			}
			defer tx.Rollback()

			err = f(logger, tx)
			if err != nil {
				return err
			}

			return tx.Commit()
		}()

		if attempts >= 2 || db.convertSQLError(err) != models.ErrDeadlock {
			break
		} else {
			logger.Error("deadlock-transaction", err, lager.Data{"attempts": attempts})
			time.Sleep(500 * time.Millisecond)
		}
	}

	return err
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
		case *pq.Error:
			return db.convertPostgresError(err.(*pq.Error))
		}
	}

	return models.ErrUnknownError
}

func (db *SQLDB) convertMySQLError(err *mysql.MySQLError) *models.Error {
	switch err.Number {
	case 1062:
		return models.ErrResourceExists
	case 1213:
		return models.ErrDeadlock
	case 1406:
		return models.ErrBadRequest
	case 1146:
		return models.NewUnrecoverableError(err)
	default:
		return models.ErrUnknownError
	}

	return nil
}

func (db *SQLDB) convertPostgresError(err *pq.Error) *models.Error {
	switch err.Code {
	case "22001":
		return models.ErrBadRequest
	case "23505":
		return models.ErrResourceExists
	case "42P01":
		return models.NewUnrecoverableError(err)
	default:
		return models.ErrUnknownError
	}
}
