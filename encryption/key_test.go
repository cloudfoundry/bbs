package encryption_test

import (
	"crypto/aes"
	"crypto/sha256"
	"fmt"

	"github.com/cloudfoundry-incubator/bbs/encryption"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Key", func() {
	Describe("NewKey", func() {
		It("generates a 256 bit key from a string that can be used as aes keys", func() {
			phrases := []string{
				"",
				"a",
				"a short phrase",
				"12345678901234567890123456789012",
				"1234567890123456789012345678901234567890123456789012345678901234567890",
			}

			for i, phrase := range phrases {
				label := fmt.Sprintf("%d", i)
				key, err := encryption.NewKey(label, phrase)
				Expect(err).NotTo(HaveOccurred())
				Expect(key.Label()).To(Equal(label))
				Expect(key.Block().BlockSize()).To(Equal(aes.BlockSize))

				phraseHash := sha256.Sum256([]byte(phrase))
				block, err := aes.NewCipher(phraseHash[:])
				Expect(err).NotTo(HaveOccurred())
				Expect(key.Block()).To(Equal(block))
			}
		})

		Context("when a key label is not specified", func() {
			It("returns a meaningful error", func() {
				_, err := encryption.NewKey("", "phrase")
				Expect(err).To(MatchError("A key label is required"))
			})
		})
	})
})
