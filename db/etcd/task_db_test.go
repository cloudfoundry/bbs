package etcd_test

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/models/internal/model_helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TaskDB", func() {
	Describe("Tasks", func() {
		Context("when there are tasks", func() {
			var expectedTasks []*models.Task

			BeforeEach(func() {
				expectedTasks = []*models.Task{
					model_helpers.NewValidTask("a-guid"), model_helpers.NewValidTask("b-guid"),
				}

				for _, t := range expectedTasks {
					etcdHelper.SetRawTask(t)
				}
			})

			It("returns all the tasks", func() {
				tasks, err := etcdDB.Tasks(logger, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(tasks.GetTasks()).To(ConsistOf(expectedTasks))
			})

			It("can filter", func() {
				tasks, err := etcdDB.Tasks(logger, func(t *models.Task) bool { return t.TaskGuid == "b-guid" })
				Expect(err).NotTo(HaveOccurred())
				Expect(tasks.Tasks).To(HaveLen(1))
				Expect(tasks.Tasks[0]).To(Equal(expectedTasks[1]))
			})
		})

		Context("when there are no tasks", func() {
			It("returns an empty list", func() {
				tasks, err := etcdDB.Tasks(logger, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(tasks).NotTo(BeNil())
				Expect(tasks.GetTasks()).To(BeEmpty())
			})
		})

		Context("when there is invalid data", func() {
			BeforeEach(func() {
				etcdHelper.CreateValidTask("some-guid")
				etcdHelper.CreateMalformedTask("some-other-guid")
				etcdHelper.CreateValidTask("some-third-guid")
			})

			It("errors", func() {
				_, err := etcdDB.Tasks(logger, nil)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when etcd is not there", func() {
			BeforeEach(func() {
				etcdRunner.Stop()
			})

			AfterEach(func() {
				etcdRunner.Start()
			})

			It("errors", func() {
				_, err := etcdDB.Tasks(logger, nil)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("TaskByGuid", func() {
		Context("when there is a task", func() {
			var expectedTask *models.Task

			BeforeEach(func() {
				expectedTask = model_helpers.NewValidTask("task-guid")
				etcdHelper.SetRawTask(expectedTask)
			})

			It("returns the task", func() {
				task, err := etcdDB.TaskByGuid(logger, "task-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(task).To(Equal(expectedTask))
			})
		})

		Context("when there is no task", func() {
			It("returns a ResourceNotFound", func() {
				_, err := etcdDB.TaskByGuid(logger, "nota-guid")
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})
		})

		Context("when there is invalid data", func() {
			BeforeEach(func() {
				etcdHelper.CreateMalformedTask("some-other-guid")
			})

			It("errors", func() {
				_, err := etcdDB.TaskByGuid(logger, "some-other-guid")
				Expect(err).To(Equal(models.ErrDeserializeJSON))
			})
		})

		Context("when etcd is not there", func() {
			BeforeEach(func() {
				etcdRunner.Stop()
			})

			AfterEach(func() {
				etcdRunner.Start()
			})

			It("errors", func() {
				_, err := etcdDB.TaskByGuid(logger, "some-other-guid")
				Expect(err).To(Equal(models.ErrUnknownError))
			})
		})
	})
})
