package etcd_test

import (
	"github.com/cloudfoundry-incubator/bbs/db"
	. "github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TaskDB", func() {
	var (
		etcdDB db.TaskDB
	)

	BeforeEach(func() {
		etcdDB = NewETCD(etcdClient, auctioneerClient, clock)
	})

	Describe("Tasks", func() {
		Context("when there are tasks", func() {
			var expectedTasks []*models.Task

			BeforeEach(func() {
				expectedTasks = []*models.Task{
					testHelper.NewValidTask("a-guid"), testHelper.NewValidTask("b-guid"),
				}

				for _, t := range expectedTasks {
					testHelper.SetRawTask(t)
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
				testHelper.CreateValidTask("some-guid")
				testHelper.CreateMalformedTask("some-other-guid")
				testHelper.CreateValidTask("some-third-guid")
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
				expectedTask = testHelper.NewValidTask("task-guid")
				testHelper.SetRawTask(expectedTask)
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
				testHelper.CreateMalformedTask("some-other-guid")
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
