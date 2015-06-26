package db

//go:generate counterfeiter . DomainDB
type DomainDB interface {
	GetAllDomains() ([]string, error)
	UpsertDomain(domain string, ttl int) error
}
