package db

import "github.com/cloudfoundry-incubator/bbs/models"

//go:generate counterfeiter . DomainDB
type DomainDB interface {
	GetAllDomains() (*models.Domains, error)
	UpsertDomain(domain string, ttl int) error
}
