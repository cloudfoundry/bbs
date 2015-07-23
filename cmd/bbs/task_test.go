package main_test

import (
	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Task API", func() {
	var (
		actualTasks   []*models.Task
		expectedTasks []*models.Task

		filter db.TaskFilter

		getErr error
	)

	BeforeEach(func() {
		filter = nil
		actualTasks = nil
		expectedTasks = []*models.Task{testHelper.NewValidTask("a-guid"), testHelper.NewValidTask("b-guid")}
		expectedTasks[1].Domain = "b-domain"
		expectedTasks[1].CellId = "b-cell"
		for _, t := range expectedTasks {
			testHelper.SetRawTask(t)
		}
	})

	Describe("GET /v1/tasks", func() {
		Context("all tasks", func() {
			BeforeEach(func() {
				actualTasks, getErr = client.Tasks()
			})

			It("responds without error", func() {
				Expect(getErr).NotTo(HaveOccurred())
			})

			It("has the correct number of responses", func() {
				Expect(actualTasks).To(ConsistOf(expectedTasks))
			})
		})

		Context("when filtering by domain", func() {
			var domain string
			BeforeEach(func() {
				domain = expectedTasks[0].Domain
				actualTasks, getErr = client.TasksByDomain(domain)
			})

			It("has the correct number of responses", func() {
				Expect(actualTasks).To(ConsistOf(expectedTasks[0]))
			})
		})

		Context("when filtering by cell", func() {
			BeforeEach(func() {
				actualTasks, getErr = client.TasksByCellID("b-cell")
			})

			It("has the correct number of responses", func() {
				Expect(actualTasks).To(ConsistOf(expectedTasks[1]))
			})
		})
	})

	Describe("GET /v1/tasks/:task_guid", func() {
		It("returns the task", func() {
			task, getErr := client.TaskByGuid(expectedTasks[0].TaskGuid)
			Expect(getErr).NotTo(HaveOccurred())
			Expect(task).To(Equal(expectedTasks[0]))
		})
	})
})
