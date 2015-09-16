package etcd_test

import (
	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Encryption", func() {
	Describe("SetEncryptionKeyLabel", func() {
		It("sets the encryption key label into the database", func() {
			err := etcdDB.SetEncryptionKeyLabel(logger, "expected-key")
			Expect(err).NotTo(HaveOccurred())

			response, err := storeClient.Get(etcd.EncryptionKeyLabelKey, false, false)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.Node.Value).To(Equal("expected-key"))
		})
	})

	Describe("EncryptionKeyLabel", func() {
		Context("when the encription key label key exists", func() {
			It("retrieves the encrption key label from the database", func() {
				_, err := storeClient.Set(etcd.EncryptionKeyLabelKey, []byte("expected-key"), etcd.NO_TTL)
				Expect(err).NotTo(HaveOccurred())

				keyLabel, err := etcdDB.EncryptionKeyLabel(logger)
				Expect(err).NotTo(HaveOccurred())

				Expect(keyLabel).To(Equal("expected-key"))
			})
		})

		Context("when the encryption key label key does not exist", func() {
			It("returns a ErrResourceNotFound", func() {
				keyLabel, err := etcdDB.EncryptionKeyLabel(logger)
				Expect(err).To(MatchError(models.ErrResourceNotFound))
				Expect(keyLabel).To(Equal(""))
			})
		})
	})
})
