package main_test

import (
	"flag"

	"github.com/cloudfoundry-incubator/bbs/cmd/bbs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Encryption Flags", func() {
	var (
		flagSet         *flag.FlagSet
		encryptionFlags *main.EncryptionFlags
	)

	JustBeforeEach(func() {
		flagSet = flag.NewFlagSet("test", 2)
		encryptionFlags = main.AddEncryptionFlags(flagSet)
	})

	Describe("Validate", func() {
		var args []string

		BeforeEach(func() {
			args = []string{}
		})

		It("ensures there's at least one encryption key", func() {
			args = append(args, "-activeKeyLabel="+"label")
			flagSet.Parse(args)

			_, err := encryptionFlags.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("ensures there's a selected active key", func() {
			args = append(args, "-encryptionKey="+"label:key")
			flagSet.Parse(args)

			_, err := encryptionFlags.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("fails if the active key is not on the list", func() {
			args = append(args, "-encryptionKey="+"label:key")
			args = append(args, "-activeKeyLabel="+"other-label")
			flagSet.Parse(args)

			_, err := encryptionFlags.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("fails if creating a key fails to parse", func() {
			args = append(args, "-encryptionKey="+"label:key")
			args = append(args, "-encryptionKey="+"invalid")
			args = append(args, "-activeKeyLabel="+"label")
			flagSet.Parse(args)

			_, err := encryptionFlags.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("returns a key manager with all the keys", func() {
			args = append(args, "-encryptionKey="+"label:key:with:colon")
			args = append(args, "-encryptionKey="+"old-label:old-key")
			args = append(args, "-activeKeyLabel="+"label")
			flagSet.Parse(args)

			km, err := encryptionFlags.Validate()
			Expect(err).NotTo(HaveOccurred())
			Expect(km.EncryptionKey().Label()).To(Equal("label"))
			Expect(km.DecryptionKey("old-label").Label()).To(Equal("old-label"))
		})
	})
})
