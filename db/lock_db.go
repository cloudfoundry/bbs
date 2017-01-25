package db

import (
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

//go:generate counterfeiter . DB
type LockDB interface {
	Lock(logger lager.Logger, lock models.Lock) error
	ReleaseLock(logger lager.Logger, lock models.Lock) error
}
