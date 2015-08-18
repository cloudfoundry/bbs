package db

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter . DomainDB
type DomainDB interface {
	Domains(logger lager.Logger) ([]string, *models.Error)
	UpsertDomain(lgger lager.Logger, domain string, ttl uint32) *models.Error
}
