package etcd_helpers

import (
	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/encryption"
	"github.com/cloudfoundry-incubator/bbs/format"
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
