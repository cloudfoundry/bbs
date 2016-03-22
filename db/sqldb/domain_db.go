package sqldb

import (
	"time"

	"github.com/pivotal-golang/lager"
)

func (db *SQLDB) Domains(logger lager.Logger) ([]string, error) {
	expireTime := db.clock.Now().Round(time.Second)
	rows, err := db.db.Query("SELECT domain FROM domains WHERE expire_time > ?", expireTime)
	if err != nil {
		return nil, db.convertSQLError(err)
	}

	var domain string
	var results []string
	for rows.Next() {
		err = rows.Scan(&domain)
		if err != nil {
			return nil, err
		}
		results = append(results, domain)
	}

	return results, nil
}

func (db *SQLDB) UpsertDomain(logger lager.Logger, domain string, ttl uint32) error {
	expireTime := db.clock.Now().Add(time.Duration(ttl) * time.Second)
	_, err := db.db.Exec(
		`INSERT INTO domains (domain, expire_time) VALUES (?, ?)
										ON DUPLICATE KEY UPDATE expire_time = ?`,
		domain,
		expireTime,
		expireTime,
	)
	if err != nil {
		return db.convertSQLError(err)
	}
	return nil
}
