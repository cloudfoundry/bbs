package sqldb

import (
	"database/sql"
	"math"
	"time"

	"code.cloudfoundry.org/lager"
)

func (db *SQLDB) Domains(logger lager.Logger) ([]string, error) {
	logger = logger.Session("domains")
	logger.Debug("starting")
	defer logger.Debug("complete")

	expireTime := db.clock.Now().Round(time.Second).UnixNano()
	rows, err := db.all(logger, db.db, domainsTable,
		domainColumns, NoLockRow,
		"expire_time > ?", expireTime,
	)
	if err != nil {
		logger.Error("failed-query", err)
		return nil, db.convertSQLError(err)
	}

	defer rows.Close()

	var domain string
	var results []string
	for rows.Next() {
		err = rows.Scan(&domain)
		if err != nil {
			logger.Error("failed-scan-row", err)
			return nil, db.convertSQLError(err)
		}
		results = append(results, domain)
	}

	if rows.Err() != nil {
		logger.Error("failed-fetching-row", err)
		return nil, db.convertSQLError(err)
	}
	return results, nil
}

func (db *SQLDB) UpsertDomain(logger lager.Logger, domain string, ttl uint32) error {
	logger = logger.Session("upsert-domain", lager.Data{"domain": domain, "ttl": ttl})
	logger.Debug("starting")
	defer logger.Debug("complete")

	return db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		expireTime := db.clock.Now().Add(time.Duration(ttl) * time.Second).UnixNano()
		if ttl == 0 {
			expireTime = math.MaxInt64
		}

		_, err := db.upsert(logger, tx, domainsTable,
			SQLAttributes{"domain": domain},
			SQLAttributes{"expire_time": expireTime},
		)
		if err != nil {
			logger.Error("failed-upsert-domain", err)
			return db.convertSQLError(err)
		}
		return nil
	})
}
