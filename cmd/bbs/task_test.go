package main_test

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/models/test/model_helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Task API", func() {
	Context("Getters", func() {
		var expectedTasks []*models.Task

		BeforeEach(func() {
			expectedTasks = []*models.Task{model_helpers.NewValidTask("a-guid"), model_helpers.NewValidTask("b-guid")}
			expectedTasks[1].Domain = "b-domain"
			expectedTasks[1].CellId = "b-cell"
			for _, t := range expectedTasks {
				etcdHelper.SetRawTask(t)
			}
		})

		Describe("GET /v1/tasks", func() {
			Context("all tasks", func() {
				It("has the correct number of responses", func() {
					actualTasks, err := client.Tasks()
					Expect(err).NotTo(HaveOccurred())
					Expect(actualTasks).To(ConsistOf(expectedTasks))
				})
			})

			Context("when filtering by domain", func() {
				It("has the correct number of responses", func() {
					domain := expectedTasks[0].Domain
					actualTasks, err := client.TasksByDomain(domain)
					Expect(err).NotTo(HaveOccurred())
					Expect(actualTasks).To(ConsistOf(expectedTasks[0]))
				})
			})

			Context("when filtering by cell", func() {
				It("has the correct number of responses", func() {
					actualTasks, err := client.TasksByCellID("b-cell")
					Expect(err).NotTo(HaveOccurred())
					Expect(actualTasks).To(ConsistOf(expectedTasks[1]))
				})
			})
		})

		Describe("GET /v1/tasks/:task_guid", func() {
			It("returns the task", func() {
				task, err := client.TaskByGuid(expectedTasks[0].TaskGuid)
				Expect(err).NotTo(HaveOccurred())
				Expect(task).To(Equal(expectedTasks[0]))
			})
		})
	})

	Context("Setters", func() {
		Describe("POST /v1/tasks", func() {
			It("adds the desired task", func() {
				expectedTask := model_helpers.NewValidTask("task-1")
				err := client.DesireTask(expectedTask.TaskGuid, expectedTask.Domain, expectedTask.TaskDefinition)
				Expect(err).NotTo(HaveOccurred())

				task, err := client.TaskByGuid(expectedTask.TaskGuid)
				Expect(err).NotTo(HaveOccurred())
				Expect(task.TaskDefinition).To(Equal(expectedTask.TaskDefinition))
			})
		})

		Describe("POST /v1/tasks/start", func() {
			var taskDef = model_helpers.NewValidTaskDefinition()
			const taskGuid = "task-1"
			const cellId = "cell-1"

			BeforeEach(func() {
				err := client.DesireTask(taskGuid, "test", taskDef)
				Expect(err).NotTo(HaveOccurred())
			})

			It("changes the task state from pending to running", func() {
				task, err := client.TaskByGuid(taskGuid)
				Expect(err).NotTo(HaveOccurred())
				Expect(task.State).To(Equal(models.Task_Pending))

				_, err = client.StartTask(taskGuid, cellId)
				Expect(err).NotTo(HaveOccurred())

				task, err = client.TaskByGuid(taskGuid)
				Expect(err).NotTo(HaveOccurred())
				Expect(task.State).To(Equal(models.Task_Running))
			})

			It("shouldStart is true", func() {
				shouldStart, err := client.StartTask(taskGuid, cellId)
				Expect(err).NotTo(HaveOccurred())
				Expect(shouldStart).To(BeTrue())
			})
		})

		Describe("POST /v1/tasks/cancel", func() {
			It("cancel the desired task", func() {
				expectedTask := model_helpers.NewValidTask("task-1")
				err := client.DesireTask(expectedTask.TaskGuid, expectedTask.Domain, expectedTask.TaskDefinition)
				Expect(err).NotTo(HaveOccurred())

				err = client.CancelTask(expectedTask.TaskGuid)
				Expect(err).NotTo(HaveOccurred())

				task, err := client.TaskByGuid(expectedTask.TaskGuid)
				Expect(err).NotTo(HaveOccurred())
				Expect(task.FailureReason).To(Equal("task was cancelled"))
			})
		})
	})
})
