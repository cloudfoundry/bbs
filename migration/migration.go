package migration

import (
	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/encryption"
	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter -o migrationfakes/fake_migration.go . Migration

type Migration interface {
	Version() int64
	Up(logger lager.Logger) error
	Down(logger lager.Logger) error
	SetStoreClient(storeClient etcd.StoreClient)
	SetCryptor(cryptor encryption.Cryptor)
}
