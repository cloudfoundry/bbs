package sqldb_test

import (
	"crypto/rand"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/bbs/db/sqldb"
	"github.com/cloudfoundry-incubator/bbs/encryption"
	"github.com/cloudfoundry-incubator/bbs/format"
	"github.com/cloudfoundry-incubator/bbs/models"
)

var _ = Describe("Encryption", func() {
	Describe("SetEncryptionKeyLabel", func() {
		It("sets the encryption key label into the database", func() {
			expectedLabel := "expectedLabel"
			err := sqlDB.SetEncryptionKeyLabel(logger, expectedLabel)
			Expect(err).NotTo(HaveOccurred())

			rows, err := db.Query("SELECT value FROM configurations WHERE id = $1", sqldb.EncryptionKeyID)
			Expect(err).NotTo(HaveOccurred())
			defer rows.Close()
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
				Expect(err).To(Equal(models.ErrBadRequest))
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

				rows, err := db.Query("SELECT value FROM configurations WHERE id = $1", sqldb.EncryptionKeyID)
				Expect(err).NotTo(HaveOccurred())
				defer rows.Close()
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
					Expect(err).To(Equal(models.ErrBadRequest))
				})
			})
		})
	})

	Describe("EncryptionKeyLabel", func() {
		Context("when the encription key label key exists", func() {
			It("retrieves the encrption key label from the database", func() {
				label := "expectedLabel"
				_, err := db.Exec("INSERT INTO configurations VALUES ($1, $2)", sqldb.EncryptionKeyID, label)
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

	makeCryptor := func(activeLabel string, decryptionLabels ...string) encryption.Cryptor {
		activeKey, err := encryption.NewKey(activeLabel, fmt.Sprintf("%s-passphrase", activeLabel))
		Expect(err).NotTo(HaveOccurred())

		decryptionKeys := []encryption.Key{}
		for _, label := range decryptionLabels {
			key, err := encryption.NewKey(label, fmt.Sprintf("%s-passphrase", label))
			Expect(err).NotTo(HaveOccurred())
			decryptionKeys = append(decryptionKeys, key)
		}
		if len(decryptionKeys) == 0 {
			decryptionKeys = nil
		}

		keyManager, err := encryption.NewKeyManager(activeKey, decryptionKeys)
		Expect(err).NotTo(HaveOccurred())
		return encryption.NewCryptor(keyManager, rand.Reader)
	}

	Describe("PerformEncryption", func() {
		It("recursively re-encrypts all existing records", func() {
			var cryptor encryption.Cryptor
			var encoder format.Encoder

			value1 := []byte("some text")
			value2 := []byte("another value")
			value3 := []byte("more value")
			value4 := []byte("actual value")
			taskGuid := "uniquetaskguid"
			processGuid := "uniqueprocessguid"

			cryptor = makeCryptor("old")
			encoder = format.NewEncoder(cryptor)

			encoded1, err := encoder.Encode(format.BASE64_ENCRYPTED, value1)
			Expect(err).NotTo(HaveOccurred())

			encoded2, err := encoder.Encode(format.BASE64_ENCRYPTED, value2)
			Expect(err).NotTo(HaveOccurred())

			encoded3, err := encoder.Encode(format.BASE64_ENCRYPTED, value3)
			Expect(err).NotTo(HaveOccurred())

			encoded4, err := encoder.Encode(format.BASE64_ENCRYPTED, value4)
			Expect(err).NotTo(HaveOccurred())

			_, err = db.Exec("INSERT INTO tasks (guid, domain, task_definition) VALUES ($1, $2, $3)", taskGuid, "fake-domain", encoded1)
			Expect(err).NotTo(HaveOccurred())

			_, err = db.Exec(`
				INSERT INTO desired_lrps
					(process_guid, domain, log_guid, instances, run_info, memory_mb,
					disk_mb, rootfs, routes, volume_placement, modification_tag_epoch)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
				processGuid, "fake-domain", "some-log-guid", 1, encoded2, 10, 10,
				"some-root-fs", []byte("{}"), encoded3, 10)
			Expect(err).NotTo(HaveOccurred())

			_, err = db.Exec(`
				INSERT INTO actual_lrps
					(process_guid, domain, net_info, instance_index, modification_tag_epoch, state)
				VALUES ($1, $2, $3, $4, $5, $6)`,
				processGuid, "fake-domain", encoded4, 0, 10, "yo")
			Expect(err).NotTo(HaveOccurred())

			cryptor = makeCryptor("new", "old")

			sqlDB := sqldb.NewSQLDB(db, 5, 5, format.ENCRYPTED_PROTO, cryptor, fakeGUIDProvider, fakeClock)
			err = sqlDB.PerformEncryption(logger)
			Expect(err).NotTo(HaveOccurred())

			cryptor = makeCryptor("new")
			encoder = format.NewEncoder(cryptor)

			var result []byte
			row := db.QueryRow("SELECT task_definition FROM tasks WHERE guid = $1", taskGuid)
			err = row.Scan(&result)
			Expect(err).NotTo(HaveOccurred())
			decrypted1, err := encoder.Decode(result)
			Expect(err).NotTo(HaveOccurred())
			Expect(decrypted1).To(Equal(value1))

			var runInfo, volumePlacement []byte
			row = db.QueryRow("SELECT run_info, volume_placement FROM desired_lrps WHERE process_guid = $1", processGuid)
			err = row.Scan(&runInfo, &volumePlacement)
			Expect(err).NotTo(HaveOccurred())
			decrypted2, err := encoder.Decode(runInfo)
			Expect(err).NotTo(HaveOccurred())
			decrypted3, err := encoder.Decode(volumePlacement)
			Expect(err).NotTo(HaveOccurred())
			Expect(decrypted2).To(Equal(value2))
			Expect(decrypted3).To(Equal(value3))

			var netInfo []byte
			row = db.QueryRow("SELECT net_info FROM actual_lrps WHERE process_guid = $1", processGuid)
			err = row.Scan(&netInfo)
			Expect(err).NotTo(HaveOccurred())
			decrypted4, err := encoder.Decode(netInfo)
			Expect(err).NotTo(HaveOccurred())
			Expect(decrypted4).To(Equal(value4))
		})

		It("does not fail encryption if it can't read a record", func() {
			var cryptor encryption.Cryptor
			var encoder format.Encoder

			value1 := []byte("some text")
			taskGuid := "uniquetaskguid"

			cryptor = makeCryptor("unknown")
			encoder = format.NewEncoder(cryptor)

			encoded1, err := encoder.Encode(format.BASE64_ENCRYPTED, value1)
			Expect(err).NotTo(HaveOccurred())

			_, err = db.Exec("INSERT INTO tasks (guid, domain, task_definition) VALUES ($1, $2, $3)", taskGuid, "fake-domain", encoded1)
			Expect(err).NotTo(HaveOccurred())

			cryptor = makeCryptor("new", "old")

			sqlDB := sqldb.NewSQLDB(db, 5, 5, format.ENCRYPTED_PROTO, cryptor, fakeGUIDProvider, fakeClock)
			err = sqlDB.PerformEncryption(logger)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
