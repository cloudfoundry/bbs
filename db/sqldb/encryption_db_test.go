package sqldb_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/bbs/db/sqldb"
	"github.com/cloudfoundry-incubator/bbs/models"
)

var _ = Describe("Encryption", func() {
	Describe("SetEncryptionKeyLabel", func() {
		It("sets the encryption key label into the database", func() {
			expectedLabel := "expectedLabel"
			err := sqlDB.SetEncryptionKeyLabel(logger, expectedLabel)
			Expect(err).NotTo(HaveOccurred())

			rows, err := db.Query("SELECT value FROM configurations WHERE id = ?", sqldb.EncryptionKeyId)
			Expect(err).NotTo(HaveOccurred())
			Expect(rows.Next()).To(BeTrue())
			var label string
			err = rows.Scan(&label)
			Expect(err).NotTo(HaveOccurred())
			Expect(label).To(Equal(expectedLabel))
		})

		Context("when the label is too long", func() {
			It("returns an error trying to insert", func() {
				expectedLabel := randStr(256)
				err := sqlDB.SetEncryptionKeyLabel(logger, expectedLabel)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Error creating encryption key label"))
			})
		})

		Context("When the encryption key is already set", func() {
			BeforeEach(func() {
				previouslySetLabel := "jim-likes-kittens-meow"
				err := sqlDB.SetEncryptionKeyLabel(logger, previouslySetLabel)
				Expect(err).NotTo(HaveOccurred())
			})

			It("replaces the encryption key label in the database", func() {
				expectedLabel := "expectedLabel"
				err := sqlDB.SetEncryptionKeyLabel(logger, expectedLabel)
				Expect(err).NotTo(HaveOccurred())

				rows, err := db.Query("SELECT value FROM configurations WHERE id = ?", sqldb.EncryptionKeyId)
				Expect(err).NotTo(HaveOccurred())
				Expect(rows.Next()).To(BeTrue())
				var label string
				err = rows.Scan(&label)
				Expect(err).NotTo(HaveOccurred())
				Expect(label).To(Equal(expectedLabel))
			})

			Context("when the label is too long", func() {
				It("returns an error trying to insert", func() {
					expectedLabel := randStr(256)
					err := sqlDB.SetEncryptionKeyLabel(logger, expectedLabel)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Error updating encryption key label"))
				})
			})
		})
	})

	Describe("EncryptionKeyLabel", func() {
		Context("when the encription key label key exists", func() {
			It("retrieves the encrption key label from the database", func() {
				label := "expectedLabel"
				_, err := db.Exec("INSERT INTO configurations VALUES (?, ?)", sqldb.EncryptionKeyId, label)
				Expect(err).NotTo(HaveOccurred())

				keyLabel, err := sqlDB.EncryptionKeyLabel(logger)
				Expect(err).NotTo(HaveOccurred())

				Expect(keyLabel).To(Equal(label))
			})
		})

		Context("when the encryption key label key does not exist", func() {
			It("returns a ErrResourceNotFound", func() {
				keyLabel, err := sqlDB.EncryptionKeyLabel(logger)
				Expect(err).To(MatchError(models.ErrResourceNotFound))
				Expect(keyLabel).To(Equal(""))
			})
		})
	})
})
