package db

import (
	"context"

	"code.cloudfoundry.org/diegosqldb"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/locket/guidprovider"
	"code.cloudfoundry.org/locket/models"
)

//go:generate counterfeiter . LockDB
type LockDB interface {
	Lock(ctx context.Context, logger lager.Logger, resource *models.Resource, ttl int64) (*Lock, error)
	Release(ctx context.Context, logger lager.Logger, resource *models.Resource) error
	Fetch(ctx context.Context, logger lager.Logger, key string) (*Lock, error)
	FetchAndRelease(ctx context.Context, logger lager.Logger, lock *Lock) (bool, error)
	FetchAll(ctx context.Context, logger lager.Logger, lockType string) ([]*Lock, error)
	Count(ctx context.Context, logger lager.Logger, lockType string) (int, error)
}

type Lock struct {
	*models.Resource
	TtlInSeconds  int64
	ModifiedIndex int64
	ModifiedId    string
}

type SQLDB struct {
	diegosqldb.QueryableDB
	flavor       string
	helper       diegosqldb.SQLHelper
	guidProvider guidprovider.GUIDProvider
}

func NewSQLDB(
	db diegosqldb.QueryableDB,
	flavor string,
	guidProvider guidprovider.GUIDProvider,
) *SQLDB {
	helper := diegosqldb.NewSQLHelper(flavor)
	return &SQLDB{
		QueryableDB:  db,
		flavor:       flavor,
		helper:       helper,
		guidProvider: guidProvider,
	}
}
