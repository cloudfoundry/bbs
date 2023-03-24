package format_test

import (
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/bbs/encryption"
	"code.cloudfoundry.org/bbs/encryption/encryptionfakes"
	"code.cloudfoundry.org/bbs/format"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/models/test/model_helpers"
)

var _ = Describe("Format", func() {
	var (
		serializer format.Serializer
		cryptor    *encryptionfakes.FakeCryptor
		encoder    format.Encoder
		logger     lager.Logger
		task       *models.Task
	)

	BeforeEach(func() {
		task = model_helpers.NewValidTask("a-guid")
		logger = lagertest.NewTestLogger("test")
		cryptor = &encryptionfakes.FakeCryptor{}
		cryptor.EncryptStub = func(plaintext []byte) (encryption.Encrypted, error) {
			nonce := [12]byte{}
			return encryption.Encrypted{
				KeyLabel:   "label",
				Nonce:      nonce[:],
				CipherText: plaintext,
			}, nil
		}
		cryptor.DecryptStub = func(ciphered encryption.Encrypted) ([]byte, error) {
			return ciphered.CipherText, nil
		}
		encoder = format.NewEncoder(cryptor)
		serializer = format.NewSerializer(cryptor)
	})

	Describe("Marshal", func() {
		Describe("ENCRYPTED_PROTO", func() {
			It("marshals the data as protobuf with an base64 encoded ciphertext envelope", func() {
				encoded, err := serializer.Marshal(logger, task)
				Expect(err).NotTo(HaveOccurred())

				unencoded, err := encoder.Decode(encoded)
				Expect(err).NotTo(HaveOccurred())

				Expect(unencoded[0]).To(BeEquivalentTo(format.PROTO))
				var actualTask models.Task
				err = proto.Unmarshal(unencoded[2:], &actualTask)
				Expect(err).NotTo(HaveOccurred())
				Expect(actualTask).To(Equal(*task))
			})
		})
	})

	Describe("Unmarshal", func() {
		Describe("ENCRYPTED_PROTO", func() {
			It("unmarshals the protobuf data from a base64 encoded ciphertext envelope", func() {
				payload, err := serializer.Marshal(logger, task)
				Expect(err).NotTo(HaveOccurred())

				var decodedTask models.Task
				err = serializer.Unmarshal(logger, payload, &decodedTask)
				Expect(err).NotTo(HaveOccurred())
				Expect(*task).To(Equal(decodedTask))
			})
		})
	})
})
