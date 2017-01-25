package etcd

import (
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

/*************** NO-OP IMPLEMENTATION FOR LEGACY DATABASE ******************/

func (db *ETCDDB) Lock(logger lager.Logger, lock models.Lock) error {
	return nil
}

func (db *ETCDDB) ReleaseLock(logger lager.Logger, lock models.Lock) error {
	return nil
}
