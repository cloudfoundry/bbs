package helpers

import (
	"database/sql"
	"time"

	"code.cloudfoundry.org/lager"
)

// BEGIN TRANSACTION; f ... ; COMMIT; or
// BEGIN TRANSACTION; f ... ; ROLLBACK; if f returns an error.
func (h *sqlHelper) Transact(logger lager.Logger, db *sql.DB, f func(logger lager.Logger, tx *sql.Tx) error) error {
	var err error

	for attempts := 0; attempts < 3; attempts++ {
		err = func() error {
			tx, err := db.Begin()
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

		if attempts >= 2 || h.ConvertSQLError(err) != ErrDeadlock {
			break
		} else {
			logger.Error("deadlock-transaction", err, lager.Data{"attempts": attempts})
			time.Sleep(500 * time.Millisecond)
		}
	}

	return err
}
