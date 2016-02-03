package sqldb

import (
	"log"
	"time"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

func (db *SQLDB) Domains(logger lager.Logger) ([]string, error) {
	rows, err := db.sql.Query("select * from domains where expireTime > $1", time.Now())
	if err != nil {
		logger.Error("failed-to-fetch-domains", err)
		return nil, models.ErrUnknownError
	}
	defer rows.Close()

	domains := []string{}
	for rows.Next() {
		var domain string
		if err := rows.Scan(&domain); err != nil {
			log.Fatal(err)
		}
		domains = append(domains, domain)
	}

	return domains, nil
}

func (db *SQLDB) UpsertDomain(logger lager.Logger, domain string, ttl uint32) error {
	expireTime := time.Now().Add(time.Duration(ttl) * time.Second)
	result, err := db.sql.Exec("update domains set domain = $1, expireTime = $2 where domain = $3", domain, expireTime, domain)
	if err != nil {
		panic(err)
	}
	count, err := result.RowsAffected()
	logger.Info("update-domain", lager.Data{"count": count, "error": err})
	if err != nil || count == 0 {
		logger.Error("failed-to-update-domain", err)
		_, err = db.sql.Exec("insert into domains (domain, expireTime) values ($1, $2)", domain, expireTime)
		logger.Info("insert-domain", lager.Data{"result": result})
		if err != nil {
			logger.Error("failed-to-insert-domain", err)
			return models.ErrUnknownError
		}
	}
	return nil
}
