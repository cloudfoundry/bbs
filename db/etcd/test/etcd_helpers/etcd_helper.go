package etcd_helpers

import (
	"crypto/rand"

	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/encryption"
	"github.com/cloudfoundry-incubator/bbs/format"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
)

type ETCDHelper struct {
	client     etcd.StoreClient
	format     *format.Format
	serializer format.Serializer
	logger     lager.Logger
}

func NewETCDHelper(serializationFormat *format.Format, client etcd.StoreClient) *ETCDHelper {
	key, err := encryption.NewKey("keylabel", "passphrase")
	Expect(err).NotTo(HaveOccurred())
	keyManager, err := encryption.NewKeyManager(key, nil)
	Expect(err).NotTo(HaveOccurred())
	cryptor := encryption.NewCryptor(keyManager, rand.Reader)

	logger := lagertest.NewTestLogger("etcd-helper")

	return &ETCDHelper{
		client:     client,
		format:     serializationFormat,
		serializer: format.NewSerializer(cryptor),
		logger:     logger,
	}
}
