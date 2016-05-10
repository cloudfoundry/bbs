package migrations_test

import (
	"crypto/rand"
	"encoding/json"
	"time"

	etcddb "github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/db/migrations"
	"github.com/cloudfoundry-incubator/bbs/encryption"
	"github.com/cloudfoundry-incubator/bbs/format"
	"github.com/cloudfoundry-incubator/bbs/migration"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/models/test/model_helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("Change Timeouts to Milliseconds Migration", func() {
	var (
		migration  migration.Migration
		serializer format.Serializer
		cryptor    encryption.Cryptor
		db         *etcddb.ETCDDB

		logger *lagertest.TestLogger
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")

		encryptionKey, err := encryption.NewKey("label", "passphrase")
		Expect(err).NotTo(HaveOccurred())
		keyManager, err := encryption.NewKeyManager(encryptionKey, nil)
		Expect(err).NotTo(HaveOccurred())
		cryptor = encryption.NewCryptor(keyManager, rand.Reader)
		serializer = format.NewSerializer(cryptor)
		migration = migrations.NewTimeoutMilliseconds()
		db = etcddb.NewETCD(format.ENCRYPTED_PROTO, 1, 1, 1*time.Minute, cryptor, storeClient, fakeClock)
	})

	It("appends itself to the migration list", func() {
		Expect(migrations.Migrations).To(ContainElement(migration))
	})

	Describe("Version", func() {
		It("returns the timestamp from which it was created", func() {
			Expect(migration.Version()).To(BeEquivalentTo(1451635200))
		})
	})

	Describe("Down", func() {
		It("returns a not implemented error", func() {
			Expect(migration.Down(logger)).To(HaveOccurred())
		})
	})

	FDescribe("Up", func() {

		var (
			// expectedDesiredLRP *models.DesiredLRP
			expectedTask *models.Task
			taskGuid     string

			migrationErr error
		)

		JustBeforeEach(func() {
			migration.SetStoreClient(storeClient)
			migration.SetCryptor(cryptor)
			migration.SetClock(fakeClock)
			migrationErr = migration.Up(logger)
		})

		Describe("Task Migration", func() {
			BeforeEach(func() {
				taskGuid = "task-guid-1"
				expectedTask = model_helpers.NewValidTask(taskGuid)
				// Model changed but this is test setup and we store the timeout in nanos
				expectedTask.Action = models.WrapAction(&models.TimeoutAction{Action: model_helpers.NewValidAction(), TimeoutMs: 5 * int64(time.Second)})

				taskData, err := serializer.Marshal(logger, format.ENCRYPTED_PROTO, expectedTask)
				Expect(err).NotTo(HaveOccurred())
				_, err = storeClient.Set(etcddb.TaskSchemaPath(expectedTask), taskData, 0)
				Expect(err).NotTo(HaveOccurred())
			})

			It("changes task timeoutAction timeout to milliseconds", func() {
				Expect(migrationErr).NotTo(HaveOccurred())
				task, err := db.TaskByGuid(logger, taskGuid)
				Expect(err).NotTo(HaveOccurred())
				Expect(task.Action.GetTimeoutAction().GetTimeoutMs()).To(Equal(int64(5000)))
			})
		})

		Describe("DesiredLRP Migration", func() {
			var (
				processGuid string
				desiredLRP  *models.DesiredLRP
			)

			BeforeEach(func() {
				processGuid = "process-guid-1"
				desiredLRP = model_helpers.NewValidDesiredLRP(processGuid)
				desiredLRP.Action = models.WrapAction(models.Timeout(&models.RunAction{Path: "ls", User: "name"}, 4*time.Second))
				desiredLRP.Setup = models.WrapAction(models.Timeout(&models.RunAction{Path: "ls", User: "name"}, 7*time.Second))

				schedulingInfo, runInfo := desiredLRP.CreateComponents(fakeClock.Now())
				_, err := json.Marshal(desiredLRP.Routes)
				Expect(err).NotTo(HaveOccurred())

				schedInfoData, err := serializer.Marshal(logger, format.ENCRYPTED_PROTO, &schedulingInfo)
				Expect(err).NotTo(HaveOccurred())
				_, err = storeClient.Set(etcddb.DesiredLRPSchedulingInfoSchemaPath(processGuid), schedInfoData, 0)
				Expect(err).NotTo(HaveOccurred())
				runInfoData, err := serializer.Marshal(logger, format.ENCRYPTED_PROTO, &runInfo)
				Expect(err).NotTo(HaveOccurred())
				_, err = storeClient.Set(etcddb.DesiredLRPRunInfoSchemaPath(processGuid), runInfoData, 0)
				Expect(err).NotTo(HaveOccurred())

				encoder := format.NewEncoder(cryptor)
				encryptedVolumePlacement, err := serializer.Marshal(logger, format.ENCRYPTED_PROTO, schedulingInfo.VolumePlacement)
				Expect(err).NotTo(HaveOccurred())
				_, err = encoder.Decode(encryptedVolumePlacement)
				Expect(err).NotTo(HaveOccurred())
			})

			It("changes desiredLRP startTimeout to milliseconds", func() {
				Expect(migrationErr).NotTo(HaveOccurred())
				desiredLRP, err := db.DesiredLRPByProcessGuid(logger, processGuid)
				Expect(err).ToNot(HaveOccurred())
				// switch to uint64 when tests passed
				Expect(desiredLRP.GetStartTimeoutMs()).To(Equal(uint32(15000)))
			})

			It("changes monitor action startTimeout to milliseconds", func() {
				Expect(migrationErr).NotTo(HaveOccurred())
				desiredLRP, err := db.DesiredLRPByProcessGuid(logger, processGuid)
				Expect(err).ToNot(HaveOccurred())
				Expect(desiredLRP.GetMonitor().GetEmitProgressAction().GetAction().GetTimeoutAction().GetTimeoutMs()).To(Equal(int64(10000)))
			})

			It("changes action startTimeout to milliseconds", func() {
				Expect(migrationErr).NotTo(HaveOccurred())
				desiredLRP, err := db.DesiredLRPByProcessGuid(logger, processGuid)
				Expect(err).ToNot(HaveOccurred())
				Expect(desiredLRP.GetAction().GetTimeoutAction().GetTimeoutMs()).To(Equal(int64(4000)))
			})

			It("changes setup startTimeout to milliseconds", func() {
				Expect(migrationErr).NotTo(HaveOccurred())
				desiredLRP, err := db.DesiredLRPByProcessGuid(logger, processGuid)
				Expect(err).ToNot(HaveOccurred())
				Expect(desiredLRP.GetSetup().GetTimeoutAction().GetTimeoutMs()).To(Equal(int64(7000)))
			})
		})
	})
})
