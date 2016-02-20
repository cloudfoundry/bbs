package sqldb

import (
	"fmt"
	"time"

	"github.com/pivotal-golang/lager"
)

func (db *SQLDB) UpsertDomain(logger lager.Logger, domain string, ttl uint32) error {
	expireTime := db.clock.Now().Add(time.Duration(ttl) * time.Second)

	result, err := db.db.Exec("UPDATE domains SET expireTime = ? WHERE domain = ?", expireTime, domain)
	if err != nil {
		panic("test me")
	}

	count, err := result.RowsAffected()
	if count == 0 || err != nil {
		_, err = db.db.Exec("INSERT INTO domains VALUES (?, ?)", domain, expireTime)
	}
	return err
}

func (db *SQLDB) Domains(logger lager.Logger) ([]string, error) {
	expirationTime := db.clock.Now().Round(time.Second)
	rows, err := db.db.Query("SELECT domain FROM domains WHERE expireTime > ?", expirationTime)
	if err != nil {
		panic(fmt.Sprintf("test me too! %s", err.Error()))
	}
	var domain string
	var results []string
	for rows.Next() {
		err = rows.Scan(&domain)
		if err != nil {
			panic("how does scan fail?")
		}
		results = append(results, domain)
	}
	return results, nil
}
