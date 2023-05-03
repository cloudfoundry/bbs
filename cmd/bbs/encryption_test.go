package main_test

import (
	"code.cloudfoundry.org/bbs/cmd/bbs/testrunner"
	"code.cloudfoundry.org/bbs/models"
	. "code.cloudfoundry.org/bbs/models/test/matchers"
	"code.cloudfoundry.org/bbs/models/test/model_helpers"
	ginkgomon "github.com/tedsuo/ifrit/ginkgomon_v2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Encryption", func() {
	var (
		task         *models.Task
		expectedTask *models.Task
	)

	BeforeEach(func() {
		task = model_helpers.NewValidTask("task-1")
		expectedTask = task.Copy()
	})

	JustBeforeEach(func() {
		bbsRunner = testrunner.New(bbsBinPath, bbsConfig)
		bbsProcess = ginkgomon.Invoke(bbsRunner)
	})

	Describe("read-write encrypted data", func() {
		JustBeforeEach(func() {
			err := client.DesireTask(logger, "some-trace-id", task.TaskGuid, task.Domain, task.TaskDefinition)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when provided a single encryption key", func() {
			BeforeEach(func() {
				bbsConfig.ActiveKeyLabel = "label"
				bbsConfig.EncryptionKeys = map[string]string{"label": "some phrase"}
			})

			It("can write/read to the database", func() {
				tasks, err := client.Tasks(logger, "some-trace-id")
				Expect(err).NotTo(HaveOccurred())
				Expect(tasks).To(ContainElement(MatchTask(expectedTask)))
			})
		})

		Context("when provided a multiple encryption keys", func() {
			var expectedOldTask *models.Task

			BeforeEach(func() {
				oldTask := model_helpers.NewValidTask("old-task")
				expectedOldTask = oldTask.Copy()
				expectedOldTask = oldTask

				bbsConfig.ActiveKeyLabel = "oldkey"
				bbsConfig.EncryptionKeys = map[string]string{"oldkey": "old phrase"}
				bbsRunner = testrunner.New(bbsBinPath, bbsConfig)
				bbsProcess = ginkgomon.Invoke(bbsRunner)

				err := client.DesireTask(logger, "some-trace-id", oldTask.TaskGuid, oldTask.Domain, oldTask.TaskDefinition)
				Expect(err).NotTo(HaveOccurred())

				ginkgomon.Interrupt(bbsProcess)

				bbsConfig.ActiveKeyLabel = "newkey"
				bbsConfig.EncryptionKeys = map[string]string{
					"newkey": "new phrase",
					"oldkey": "old phrase",
				}
			})

			It("can read data that was written with old/new keys", func() {
				tasks, err := client.Tasks(logger, "some-trace-id")
				Expect(err).NotTo(HaveOccurred())
				Expect(tasks).To(ContainElement(MatchTask(expectedOldTask)))
				Expect(tasks).To(ContainElement(MatchTask(expectedTask)))
			})

			It("doesn't need the oldkey after migrating", func() {
				ginkgomon.Interrupt(bbsProcess)

				bbsConfig.ActiveKeyLabel = "newkey"
				bbsConfig.EncryptionKeys = map[string]string{"newkey": "new phrase"}

				bbsRunner = testrunner.New(bbsBinPath, bbsConfig)
				bbsProcess = ginkgomon.Invoke(bbsRunner)

				tasks, err := client.Tasks(logger, "some-trace-id")
				Expect(err).NotTo(HaveOccurred())
				Expect(tasks).To(ContainElement(MatchTask(expectedOldTask)))
				Expect(tasks).To(ContainElement(MatchTask(expectedTask)))
			})
		})
	})
})
