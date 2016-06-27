package encryption_test

import (
	"flag"

	"code.cloudfoundry.org/bbs/encryption"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Encryption Flags", func() {
	var (
		flagSet         *flag.FlagSet
		encryptionFlags *encryption.EncryptionFlags
	)

	JustBeforeEach(func() {
		flagSet = flag.NewFlagSet("test", 2)
		encryptionFlags = encryption.AddEncryptionFlags(flagSet)
	})

	Describe("Validate", func() {
		var args []string

		BeforeEach(func() {
			args = []string{}
		})

		It("ensures there's at least one encryption key", func() {
			args = append(args, "-activeKeyLabel="+"label")
			flagSet.Parse(args)

			_, _, err := encryptionFlags.Parse()
			Expect(err).To(HaveOccurred())
		})

		It("splits keys on the first colon", func() {
			args = append(args, "-encryptionKey="+"label:key:with:colon")
			args = append(args, "-activeKeyLabel="+"label")
			flagSet.Parse(args)

			_, _, err := encryptionFlags.Parse()
			Expect(err).ToNot(HaveOccurred())

			flagSet.Parse(args)
			key, keys, err := encryptionFlags.Parse()
			Expect(err).ToNot(HaveOccurred())

			km, err := encryption.NewKeyManager(key, keys)
			Expect(km.EncryptionKey().Label()).To(Equal("label"))
		})

		It("ensures there's a selected active key", func() {
			args = append(args, "-encryptionKey="+"label:key")
			flagSet.Parse(args)

			_, _, err := encryptionFlags.Parse()
			Expect(err).To(HaveOccurred())
		})

		It("fails if the active key is not on the list", func() {
			args = append(args, "-encryptionKey="+"label:key")
			args = append(args, "-activeKeyLabel="+"other-label")
			flagSet.Parse(args)

			_, _, err := encryptionFlags.Parse()
			Expect(err).To(HaveOccurred())
		})

		It("fails if creating a key fails to parse", func() {
			args = append(args, "-encryptionKey="+"label:key")
			args = append(args, "-encryptionKey="+"invalid")
			args = append(args, "-activeKeyLabel="+"label")
			flagSet.Parse(args)

			_, _, err := encryptionFlags.Parse()
			Expect(err).To(HaveOccurred())
		})

		It("returns an active key and all the keys", func() {
			args = append(args, "-encryptionKey="+"label:key:with:colon")
			args = append(args, "-encryptionKey="+"old-label:old-key")
			args = append(args, "-activeKeyLabel="+"label")
			flagSet.Parse(args)

			activeKey, keys, err := encryptionFlags.Parse()
			keyLabels := make([]string, len(keys))
			for _, key := range keys {
				keyLabels = append(keyLabels, key.Label())
			}

			Expect(err).NotTo(HaveOccurred())
			Expect(activeKey.Label()).To(Equal("label"))
			Expect(keyLabels).To(ContainElement("label"))
			Expect(keyLabels).To(ContainElement("old-label"))
		})
	})
})
