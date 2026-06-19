package format_test

import (
	"code.cloudfoundry.org/lager/v3"
	"code.cloudfoundry.org/lager/v3/lagertest"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/bbs/encryption"
	"code.cloudfoundry.org/bbs/encryption/encryptionfakes"
	"code.cloudfoundry.org/bbs/format"
)

var _ = Describe("Format", func() {
	var (
		serializer format.Serializer
		cryptor    *encryptionfakes.FakeCryptor
		encoder    format.Encoder
		logger     lager.Logger
		msg        *types.StringValue
	)

	BeforeEach(func() {
		msg = &types.StringValue{Value: "test-message"}
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
				encoded, err := serializer.Marshal(logger, msg)
				Expect(err).NotTo(HaveOccurred())

				unencoded, err := encoder.Decode(encoded)
				Expect(err).NotTo(HaveOccurred())

				Expect(unencoded[0]).To(BeEquivalentTo(format.PROTO))
				var actualMsg types.StringValue
				err = proto.Unmarshal(unencoded[2:], &actualMsg)
				Expect(err).NotTo(HaveOccurred())
				Expect(actualMsg.Value).To(Equal(msg.Value))
			})
		})
	})

	Describe("Unmarshal", func() {
		Describe("ENCRYPTED_PROTO", func() {
			It("unmarshals the protobuf data from a base64 encoded ciphertext envelope", func() {
				payload, err := serializer.Marshal(logger, msg)
				Expect(err).NotTo(HaveOccurred())

				var decodedMsg types.StringValue
				err = serializer.Unmarshal(logger, payload, &decodedMsg)
				Expect(err).NotTo(HaveOccurred())
				Expect(decodedMsg.Value).To(Equal(msg.Value))
			})
		})
	})
})
