package etcd_helpers

import (
	"code.cloudfoundry.org/bbs/db/etcd"
	"code.cloudfoundry.org/bbs/encryption"
	"code.cloudfoundry.org/bbs/format"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
)

type ETCDHelper struct {
	client     etcd.StoreClient
	format     *format.Format
	serializer format.Serializer
	logger     lager.Logger
	clock      clock.Clock
}

func NewETCDHelper(serializationFormat *format.Format, cryptor encryption.Cryptor, client etcd.StoreClient, clock clock.Clock) *ETCDHelper {
	logger := lagertest.NewTestLogger("etcd-helper")

	return &ETCDHelper{
		client:     client,
		format:     serializationFormat,
		serializer: format.NewSerializer(cryptor),
		logger:     logger,
		clock:      clock,
	}
}
