package migrations_test

import (
	"crypto/rand"

	"github.com/cloudfoundry-incubator/bbs/db/migrations"
	"github.com/cloudfoundry-incubator/bbs/encryption"
	"github.com/cloudfoundry-incubator/bbs/format"
	"github.com/cloudfoundry-incubator/bbs/migration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = FDescribe("Change Timeouts to Milliseconds Migration", func() {
	var (
		migration  migration.Migration
		serializer format.Serializer
		cryptor    encryption.Cryptor

		logger *lagertest.TestLogger
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")

		encryptionKey, err := encryption.NewKey("label", "passphrase")
		Expect(err).NotTo(HaveOccurred())
		keyManager, err := encryption.NewKeyManager(encryptionKey, nil)
		Expect(err).NotTo(HaveOccurred())
		cryptor = encryption.NewCryptor(keyManager, rand.Reader)
		serializer = format.NewSerializer(cryptor)
		migration = migrations.NewTimeoutMilliseconds()
	})

	It("appends itself to the migration list", func() {
		Expect(migrations.Migrations).To(ContainElement(migration))
	})

	Describe("Version", func() {
		It("returns the timestamp from which it was created", func() {
			Expect(migration.Version()).To(BeEquivalentTo(1451635200))
		})
	})

	Describe("Down", func() {
		It("returns a not implemented error", func() {
			Expect(migration.Down(logger)).To(HaveOccurred())
		})
	})

})
