package db

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter . DomainDB
type DomainDB interface {
	GetAllDomains(logger lager.Logger) (*models.Domains, *models.Error)
	UpsertDomain(lgger lager.Logger, domain string, ttl int) *models.Error
}
