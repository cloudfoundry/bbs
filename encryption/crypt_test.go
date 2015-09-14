package encryption_test

import (
	"bytes"
	"crypto/des"
	"crypto/rand"
	"io"

	"github.com/cloudfoundry-incubator/bbs/encryption"
	"github.com/cloudfoundry-incubator/bbs/encryption/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Crypt", func() {
	var cryptor encryption.Cryptor
	var keyManager encryption.KeyManager
	var prng io.Reader

	BeforeEach(func() {
		key, err := encryption.NewKey("label", "pass phrase")
		Expect(err).NotTo(HaveOccurred())
		keyManager, err = encryption.NewKeyManager(key, nil)
		Expect(err).NotTo(HaveOccurred())
		prng = rand.Reader
	})

	JustBeforeEach(func() {
		cryptor = encryption.NewCryptor(keyManager, prng)
	})

	It("successfully encrypts and decrypts with a key", func() {
		input := []byte("some plaintext data")

		encrypted, err := cryptor.Encrypt(input)
		Expect(err).NotTo(HaveOccurred())
		Expect(encrypted.CipherText).NotTo(HaveLen(0))
		Expect(encrypted.CipherText).NotTo(Equal(input))

		plaintext, err := cryptor.Decrypt(encrypted)
		Expect(err).NotTo(HaveOccurred())
		Expect(plaintext).NotTo(HaveLen(0))
		Expect(plaintext).To(Equal(input))
	})

	Context("when the nonce is incorrect", func() {
		It("fails to decrypt", func() {
			input := []byte("some plaintext data")

			encrypted, err := cryptor.Encrypt(input)
			Expect(err).NotTo(HaveOccurred())
			Expect(encrypted.CipherText).NotTo(HaveLen(0))
			Expect(encrypted.CipherText).NotTo(Equal(input))

			encrypted.Nonce = []byte("123456789012")

			_, err = cryptor.Decrypt(encrypted)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("cipher: message authentication failed"))
		})
	})

	Context("when the key is not found", func() {
		It("fails to decrypt", func() {
			encrypted := encryption.Encrypted{
				KeyLabel: "doesnt-exist",
				Nonce:    []byte("123456789012"),
			}

			_, err := cryptor.Decrypt(encrypted)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(`Key with label "doesnt-exist" was not found`))
		})
	})

	Context("when the ciphertext is modified", func() {
		It("fails to decrypt", func() {
			input := []byte("some plaintext data")

			encrypted, err := cryptor.Encrypt(input)
			Expect(err).NotTo(HaveOccurred())
			Expect(encrypted.CipherText).NotTo(HaveLen(0))
			Expect(encrypted.CipherText).NotTo(Equal(input))

			encrypted.CipherText[0] ^= 1

			_, err = cryptor.Decrypt(encrypted)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("cipher: message authentication failed"))
		})
	})

	Context("when the random number generator fails", func() {
		BeforeEach(func() {
			prng = bytes.NewBuffer([]byte{})
		})

		It("fails to encrypt", func() {
			input := []byte("some plaintext data")

			_, err := cryptor.Encrypt(input)
			Expect(err).To(MatchError(`Unable to generate random nonce: "EOF"`))
		})
	})

	Context("when the random number generator fails to generate enough data", func() {
		BeforeEach(func() {
			prng = bytes.NewBufferString("goo")
		})

		It("fails to encrypt", func() {
			input := []byte("some plaintext data")

			_, err := cryptor.Encrypt(input)
			Expect(err).To(MatchError("Unable to generate random nonce"))
		})
	})

	Context("when the encryption key is invalid", func() {
		var key *fakes.FakeKey

		BeforeEach(func() {
			desCipher, err := des.NewCipher([]byte("12345678"))
			Expect(err).NotTo(HaveOccurred())

			key = &fakes.FakeKey{}
			key.BlockReturns(desCipher)
			keyManager, err = encryption.NewKeyManager(key, nil)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns an error", func() {
			input := []byte("some plaintext data")

			_, err := cryptor.Encrypt(input)
			Expect(err).To(MatchError(HavePrefix("Unable to create GCM-wrapped cipher:")))
		})
	})
})
