package db

import "github.com/pivotal-golang/lager"

//go:generate counterfeiter . DomainDB
type DomainDB interface {
	Domains(logger lager.Logger) ([]string, error)
	UpsertDomain(lgger lager.Logger, domain string, ttl uint32) error
}
