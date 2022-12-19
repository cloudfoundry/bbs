package db

import (
	"context"

	"code.cloudfoundry.org/lager"
)

//counterfeiter:generate . DomainDB
type DomainDB interface {
	FreshDomains(ctx context.Context, logger lager.Logger) ([]string, error)
	UpsertDomain(ctx context.Context, lgger lager.Logger, domain string, ttl uint32) error
}
