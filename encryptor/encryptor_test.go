package encryptor_test

import (
	"crypto/rand"
	"os"

	etcddb "github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/encryption"
	"github.com/cloudfoundry-incubator/bbs/encryptor"
	"github.com/cloudfoundry-incubator/bbs/format"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/models/test/model_helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("Encryptor", func() {
	var (
		runner           ifrit.Runner
		encryptorProcess ifrit.Process

		logger     *lagertest.TestLogger
		cryptor    encryption.Cryptor
		keyManager encryption.KeyManager

		ready      chan struct{}
		signals    chan os.Signal
		runErrChan chan error

		task *models.Task
	)

	BeforeEach(func() {
		runErrChan = make(chan error, 1)
		ready = make(chan struct{})
		signals = make(chan os.Signal)

		logger = lagertest.NewTestLogger("test")

		oldKey, err := encryption.NewKey("old-key", "old-passphrase")
		encryptionKey, err := encryption.NewKey("label", "passphrase")
		Expect(err).NotTo(HaveOccurred())
		keyManager, err = encryption.NewKeyManager(encryptionKey, []encryption.Key{oldKey})
		Expect(err).NotTo(HaveOccurred())
		cryptor = encryption.NewCryptor(keyManager, rand.Reader)

		task = model_helpers.NewValidTask("task-1")
		serializer := format.NewSerializer(nil)

		encodedPayload, err := serializer.Marshal(logger, format.ENCODED_PROTO, task)
		Expect(err).NotTo(HaveOccurred())

		_, err = storeClient.Set(etcddb.TaskSchemaPathByGuid(task.TaskGuid), encodedPayload, etcddb.NO_TTL)
		Expect(err).NotTo(HaveOccurred())
	})

	JustBeforeEach(func() {
		runner = encryptor.New(logger, etcdDB, keyManager, cryptor, storeClient)
		encryptorProcess = ifrit.Background(runner)
	})

	AfterEach(func() {
		ginkgomon.Kill(encryptorProcess)
	})

	Context("when there is no current encryption key", func() {
		BeforeEach(func() {
			// intentianally left blank
		})

		It("encrypts all the existing records", func() {
			Eventually(encryptorProcess.Ready()).Should(BeClosed())
			Eventually(logger.LogMessages).Should(ContainElement("test.encryptor.encryption-finished"))

			res, err := storeClient.Get(etcddb.TaskSchemaPathByGuid(task.TaskGuid), false, false)
			Expect(err).NotTo(HaveOccurred())

			var decodedTask models.Task

			encryptionKey, err := encryption.NewKey("label", "passphrase")
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

		It("writes the current encryption key", func() {
			Eventually(func() (string, error) {
				return etcdDB.EncryptionKeyLabel(logger)
			}).Should(Equal("label"))
		})
	})

	Context("when there's a broken value in the database", func() {
		BeforeEach(func() {
			_, err := storeClient.Set(etcddb.TaskSchemaPathByGuid("invalid-task"), []byte("01borked-data"), etcddb.NO_TTL)
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not fail and logs the error", func() {
			Eventually(encryptorProcess.Ready()).Should(BeClosed())
			Eventually(logger.LogMessages).Should(ContainElement("test.encryptor.encryption-finished"))

			Expect(logger.LogMessages()).To(ContainElement("test.encryptor.failed-to-read-node"))
		})

		It("writes the current encryption key", func() {
			Eventually(func() (string, error) {
				return etcdDB.EncryptionKeyLabel(logger)
			}).Should(Equal("label"))
		})
	})

	Context("when fetching the current encryption key fails", func() {
		BeforeEach(func() {
			etcdRunner.Stop()
		})
		AfterEach(func() {
			etcdRunner.Start()
		})

		It("fails early", func() {
			var err error
			Eventually(encryptorProcess.Wait()).Should(Receive(&err))
			Expect(err).To(HaveOccurred())
			Expect(encryptorProcess.Ready()).ToNot(BeClosed())
		})
	})

	Context("when the current encryption key is not known to the encryptor", func() {
		BeforeEach(func() {
			etcdDB.SetEncryptionKeyLabel(logger, "some-unknown-key")
		})

		It("shuts down wihtout signalling ready", func() {
			var err error
			Eventually(encryptorProcess.Wait()).Should(Receive(&err))
			Expect(err).To(MatchError("Existing encryption key version (some-unknown-key) is not among the known keys"))
			Expect(encryptorProcess.Ready()).ToNot(BeClosed())
		})

		It("does not change the version", func() {
			Consistently(func() (string, error) {
				return etcdDB.EncryptionKeyLabel(logger)
			}).Should(Equal("some-unknown-key"))
		})
	})

	Context("when the current encryption key is the same as the encryptor's encryption key", func() {
		BeforeEach(func() {
			etcdDB.SetEncryptionKeyLabel(logger, "label")
		})

		It("signals ready and does not change the version", func() {
			Eventually(encryptorProcess.Ready()).Should(BeClosed())
			Consistently(func() (string, error) {
				return etcdDB.EncryptionKeyLabel(logger)
			}).Should(Equal("label"))
		})
	})

	Context("when the current encryption key is one of the encryptor's decryption keys", func() {
		BeforeEach(func() {
			etcdDB.SetEncryptionKeyLabel(logger, "old-key")
		})

		It("encrypts all the existing records", func() {
			Eventually(encryptorProcess.Ready()).Should(BeClosed())
			Eventually(logger.LogMessages).Should(ContainElement("test.encryptor.encryption-finished"))

			res, err := storeClient.Get(etcddb.TaskSchemaPathByGuid(task.TaskGuid), false, false)
			Expect(err).NotTo(HaveOccurred())

			var decodedTask models.Task

			encryptionKey, err := encryption.NewKey("label", "passphrase")
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

		It("writes the current encryption key", func() {
			Eventually(func() (string, error) {
				return etcdDB.EncryptionKeyLabel(logger)
			}).Should(Equal("label"))
		})
	})
})
