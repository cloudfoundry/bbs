package sqldb

import (
	"log"
	"time"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

func (db *SQLDB) Domains(logger lager.Logger) ([]string, error) {
	rows, err := db.sql.Query("select * from domains where expireTime > $1", time.Now)
	if err != nil {
		logger.Error("failed-to-fetch-domains", err)
		return nil, models.ErrUnknownError
	}

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
	_, err := db.sql.Exec("update domains set name = $1, ttl = $2 where name = $3", domain, expireTime, domain)
	if err != nil {
		logger.Error("failed-to-upsert-domain", err)
		return models.ErrUnknownError
	}
	return nil
}
