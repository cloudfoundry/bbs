package main_test

import (
	"crypto/rand"

	"github.com/cloudfoundry-incubator/bbs/cmd/bbs/testrunner"
	etcddb "github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/encryption"
	"github.com/cloudfoundry-incubator/bbs/format"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/models/test/model_helpers"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Encryption", func() {
	var task *models.Task

	BeforeEach(func() {
		task = model_helpers.NewValidTask("task-1")
	})

	JustBeforeEach(func() {
		bbsRunner = testrunner.New(bbsBinPath, bbsArgs)
		bbsProcess = ginkgomon.Invoke(bbsRunner)
	})

	AfterEach(func() {
		ginkgomon.Kill(bbsProcess)
	})

	Describe("read-write encrypted data", func() {
		JustBeforeEach(func() {
			err := client.DesireTask(task.TaskGuid, task.Domain, task.TaskDefinition)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when provided a single encryption key", func() {
			BeforeEach(func() {
				bbsArgs.ActiveKeyLabel = "label"
				bbsArgs.EncryptionKeys = []string{"label:some phrase"}
			})

			It("writes the value as base64 encoded encrypted protobufs with metadata", func() {
				res, err := etcdClient.Get(etcddb.TaskSchemaPathByGuid(task.TaskGuid), false, false)
				Expect(err).NotTo(HaveOccurred())

				var decodedTask models.Task

				encryptionKey, err := encryption.NewKey("label", "some phrase")
				Expect(err).NotTo(HaveOccurred())
				keyManager, err := encryption.NewKeyManager(encryptionKey, nil)
				Expect(err).NotTo(HaveOccurred())
				cryptor := encryption.NewCryptor(keyManager, rand.Reader)
				serializer := format.NewSerializer(cryptor)

				encoding := res.Node.Value[:format.EnvelopeOffset]
				Expect(format.Encoding{encoding[0], encoding[1]}).To(Equal(format.BASE64_ENCRYPTED))
				err = serializer.Unmarshal(logger, []byte(res.Node.Value), &decodedTask)
				Expect(err).NotTo(HaveOccurred())

				Expect(task.TaskGuid).To(Equal(decodedTask.TaskGuid))
			})

		})

		Context("when provided a multiple encryption keys", func() {
			BeforeEach(func() {
				bbsArgs.ActiveKeyLabel = "newkey"
				bbsArgs.EncryptionKeys = []string{
					"newkey:new phrase",
					"oldkey:old phrase",
				}
			})

			It("writes the value as base64 encoded encrypted protobufs with metadata using the active key", func() {
				res, err := etcdClient.Get(etcddb.TaskSchemaPathByGuid(task.TaskGuid), false, false)
				Expect(err).NotTo(HaveOccurred())

				var decodedTask models.Task

				encryptionKey, err := encryption.NewKey("newkey", "new phrase")
				Expect(err).NotTo(HaveOccurred())
				keyManager, err := encryption.NewKeyManager(encryptionKey, nil)
				Expect(err).NotTo(HaveOccurred())
				cryptor := encryption.NewCryptor(keyManager, rand.Reader)
				serializer := format.NewSerializer(cryptor)

				encoding := res.Node.Value[:format.EnvelopeOffset]
				Expect(format.Encoding{encoding[0], encoding[1]}).To(Equal(format.BASE64_ENCRYPTED))
				err = serializer.Unmarshal(logger, []byte(res.Node.Value), &decodedTask)
				Expect(err).NotTo(HaveOccurred())

				Expect(task.TaskGuid).To(Equal(decodedTask.TaskGuid))
			})

			It("can read existing data that was written used an old key", func() {
				encryptionKey, err := encryption.NewKey("oldkey", "old phrase")
				Expect(err).NotTo(HaveOccurred())
				keyManager, err := encryption.NewKeyManager(encryptionKey, nil)
				Expect(err).NotTo(HaveOccurred())
				cryptor := encryption.NewCryptor(keyManager, rand.Reader)
				serializer := format.NewSerializer(cryptor)

				encryptedPayload, err := serializer.Marshal(logger, format.ENCRYPTED_PROTO, task)
				Expect(err).NotTo(HaveOccurred())
				_, err = etcdClient.Set(etcddb.TaskSchemaPathByGuid(task.TaskGuid), string(encryptedPayload), etcddb.NO_TTL)
				Expect(err).NotTo(HaveOccurred())

				returnedTask, err := client.TaskByGuid(task.TaskGuid)
				Expect(err).NotTo(HaveOccurred())
				Expect(returnedTask).To(Equal(task))
			})
		})
	})

	Describe("encryptor", func() {
		Context("when there is data unencrypted in the database", func() {
			BeforeEach(func() {
				serializer := format.NewSerializer(nil)

				encodedPayload, err := serializer.Marshal(logger, format.ENCODED_PROTO, task)
				Expect(err).NotTo(HaveOccurred())

				_, err = etcdClient.Set(etcddb.TaskSchemaPathByGuid(task.TaskGuid), string(encodedPayload), etcddb.NO_TTL)
				Expect(err).NotTo(HaveOccurred())

				bbsArgs.EncryptionKeys = []string{"my-label:my-secure-passphrase"}
				bbsArgs.ActiveKeyLabel = "my-label"
			})

			It("rewrites existing data in encrypted format", func() {
				By("reading it with the client")
				returnedTask, err := client.TaskByGuid(task.TaskGuid)
				Expect(err).NotTo(HaveOccurred())
				Expect(returnedTask).To(Equal(task))

				By("decrypting it manually")
				res, err := etcdClient.Get(etcddb.TaskSchemaPathByGuid(task.TaskGuid), false, false)
				Expect(err).NotTo(HaveOccurred())

				var decodedTask models.Task

				encryptionKey, err := encryption.NewKey("my-label", "my-secure-passphrase")
				Expect(err).NotTo(HaveOccurred())
				keyManager, err := encryption.NewKeyManager(encryptionKey, nil)
				Expect(err).NotTo(HaveOccurred())
				cryptor := encryption.NewCryptor(keyManager, rand.Reader)
				serializer := format.NewSerializer(cryptor)

				encoding := res.Node.Value[:format.EnvelopeOffset]
				Expect(format.Encoding{encoding[0], encoding[1]}).To(Equal(format.BASE64_ENCRYPTED))
				err = serializer.Unmarshal(logger, []byte(res.Node.Value), &decodedTask)
				Expect(err).NotTo(HaveOccurred())

				Expect(task.TaskGuid).To(Equal(decodedTask.TaskGuid))
			})
		})
	})
})
