package sqldb_test

import (
	"time"

	"code.cloudfoundry.org/auctioneer"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/models/test/model_helpers"
	"github.com/cloudfoundry/dropsonde/metric_sender/fake"
	"github.com/cloudfoundry/dropsonde/metrics"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Convergence of Tasks", func() {
	var (
		sender *fake.FakeMetricSender

		kickTasksDurationInSeconds, expirePendingTaskDurationInSeconds            uint64
		kickTasksDuration, expirePendingTaskDuration, expireCompletedTaskDuration time.Duration
	)

	BeforeEach(func() {
		sender = fake.NewFakeMetricSender()
		metrics.Initialize(sender, nil)

		kickTasksDurationInSeconds = 10
		kickTasksDuration = time.Duration(kickTasksDurationInSeconds) * time.Second

		expirePendingTaskDurationInSeconds = 30
		expirePendingTaskDuration = time.Duration(expirePendingTaskDurationInSeconds) * time.Second

		expireCompletedTaskDuration = time.Hour
	})

	Describe("ConvergeTasks", func() {
		var (
			domain          string
			tasksToAuction  []*auctioneer.TaskStartRequest
			tasksToComplete []*models.Task
			cellSet         models.CellSet

			taskDef *models.TaskDefinition
		)

		BeforeEach(func() {
			var err error
			domain = "my-domain"
			cellSet = models.NewCellSetFromList([]*models.CellPresence{
				{CellId: "existing-cell"},
			})
			taskDef = model_helpers.NewValidTaskDefinition()

			fakeClock.IncrementBySeconds(-expirePendingTaskDurationInSeconds)
			err = sqlDB.DesireTask(logger, taskDef, "pending-expired-task", domain)
			Expect(err).NotTo(HaveOccurred())
			fakeClock.IncrementBySeconds(expirePendingTaskDurationInSeconds)

			fakeClock.IncrementBySeconds(-kickTasksDurationInSeconds)
			err = sqlDB.DesireTask(logger, taskDef, "pending-kickable-task", domain)
			Expect(err).NotTo(HaveOccurred())
			fakeClock.IncrementBySeconds(kickTasksDurationInSeconds)

			fakeClock.IncrementBySeconds(-kickTasksDurationInSeconds)
			err = sqlDB.DesireTask(logger, taskDef, "pending-kickable-invalid-task", domain)
			Expect(err).NotTo(HaveOccurred())
			_, err = db.Exec("UPDATE tasks SET task_definition = 'garbage' WHERE guid = 'pending-kickable-invalid-task'")
			Expect(err).NotTo(HaveOccurred())
			fakeClock.IncrementBySeconds(kickTasksDurationInSeconds)

			err = sqlDB.DesireTask(logger, taskDef, "pending-task", domain)
			Expect(err).NotTo(HaveOccurred())

			err = sqlDB.DesireTask(logger, taskDef, "running-task-no-cell", domain)
			Expect(err).NotTo(HaveOccurred())
			_, err = sqlDB.StartTask(logger, "running-task-no-cell", "non-existant-cell")
			Expect(err).NotTo(HaveOccurred())

			err = sqlDB.DesireTask(logger, taskDef, "running-task", domain)
			Expect(err).NotTo(HaveOccurred())
			_, err = sqlDB.StartTask(logger, "running-task", "existing-cell")
			Expect(err).NotTo(HaveOccurred())

			fakeClock.Increment(-expireCompletedTaskDuration)
			err = sqlDB.DesireTask(logger, taskDef, "completed-expired-task", domain)
			Expect(err).NotTo(HaveOccurred())
			_, err = sqlDB.StartTask(logger, "completed-expired-task", "existing-cell")
			Expect(err).NotTo(HaveOccurred())
			_, err = sqlDB.CompleteTask(logger, "completed-expired-task", "existing-cell", false, "", "")
			Expect(err).NotTo(HaveOccurred())
			fakeClock.Increment(expireCompletedTaskDuration)

			fakeClock.IncrementBySeconds(-kickTasksDurationInSeconds)
			err = sqlDB.DesireTask(logger, taskDef, "completed-kickable-task", domain)
			Expect(err).NotTo(HaveOccurred())
			_, err = sqlDB.StartTask(logger, "completed-kickable-task", "existing-cell")
			Expect(err).NotTo(HaveOccurred())
			_, err = sqlDB.CompleteTask(logger, "completed-kickable-task", "existing-cell", false, "", "")
			Expect(err).NotTo(HaveOccurred())
			fakeClock.IncrementBySeconds(kickTasksDurationInSeconds)

			fakeClock.IncrementBySeconds(-kickTasksDurationInSeconds)
			err = sqlDB.DesireTask(logger, taskDef, "completed-kickable-invalid-task", domain)
			Expect(err).NotTo(HaveOccurred())
			_, err = sqlDB.StartTask(logger, "completed-kickable-invalid-task", "existing-cell")
			Expect(err).NotTo(HaveOccurred())
			_, err = sqlDB.CompleteTask(logger, "completed-kickable-invalid-task", "existing-cell", false, "", "")
			Expect(err).NotTo(HaveOccurred())
			_, err = db.Exec("UPDATE tasks SET task_definition = 'garbage' WHERE guid = 'completed-kickable-invalid-task'")
			Expect(err).NotTo(HaveOccurred())
			fakeClock.IncrementBySeconds(kickTasksDurationInSeconds)

			err = sqlDB.DesireTask(logger, taskDef, "completed-task", domain)
			Expect(err).NotTo(HaveOccurred())
			_, err = sqlDB.StartTask(logger, "completed-task", "existing-cell")
			Expect(err).NotTo(HaveOccurred())
			_, err = sqlDB.CompleteTask(logger, "completed-task", "existing-cell", false, "", "")
			Expect(err).NotTo(HaveOccurred())

			fakeClock.Increment(-expireCompletedTaskDuration)
			err = sqlDB.DesireTask(logger, taskDef, "resolving-expired-task", domain)
			Expect(err).NotTo(HaveOccurred())
			_, err = sqlDB.StartTask(logger, "resolving-expired-task", "existing-cell")
			Expect(err).NotTo(HaveOccurred())
			_, err = sqlDB.CompleteTask(logger, "resolving-expired-task", "existing-cell", false, "", "")
			Expect(err).NotTo(HaveOccurred())
			err = sqlDB.ResolvingTask(logger, "resolving-expired-task")
			Expect(err).NotTo(HaveOccurred())
			fakeClock.Increment(expireCompletedTaskDuration)

			fakeClock.IncrementBySeconds(-kickTasksDurationInSeconds)
			err = sqlDB.DesireTask(logger, taskDef, "resolving-kickable-task", domain)
			Expect(err).NotTo(HaveOccurred())
			_, err = sqlDB.StartTask(logger, "resolving-kickable-task", "existing-cell")
			Expect(err).NotTo(HaveOccurred())
			_, err = sqlDB.CompleteTask(logger, "resolving-kickable-task", "existing-cell", false, "", "")
			Expect(err).NotTo(HaveOccurred())
			err = sqlDB.ResolvingTask(logger, "resolving-kickable-task")
			Expect(err).NotTo(HaveOccurred())
			fakeClock.IncrementBySeconds(kickTasksDurationInSeconds)

			err = sqlDB.DesireTask(logger, taskDef, "resolving-task", domain)
			Expect(err).NotTo(HaveOccurred())
			_, err = sqlDB.StartTask(logger, "resolving-task", "existing-cell")
			Expect(err).NotTo(HaveOccurred())
			_, err = sqlDB.CompleteTask(logger, "resolving-task", "existing-cell", false, "", "")
			Expect(err).NotTo(HaveOccurred())
			err = sqlDB.ResolvingTask(logger, "resolving-task")
			Expect(err).NotTo(HaveOccurred())

			fakeClock.IncrementBySeconds(1)
		})

		JustBeforeEach(func() {
			tasksToAuction, tasksToComplete = sqlDB.ConvergeTasks(logger, cellSet, kickTasksDuration, expirePendingTaskDuration, expireCompletedTaskDuration)
		})

		It("bumps the convergence counter", func() {
			Expect(sender.GetCounter("ConvergenceTaskRuns")).To(Equal(uint64(1)))
		})

		It("reports the duration that it took to converge", func() {
			reportedDuration := sender.GetValue("ConvergenceTaskDuration")
			Expect(reportedDuration.Unit).To(Equal("nanos"))
			Expect(reportedDuration.Value).NotTo(BeZero())
		})

		It("emits task status count metrics", func() {
			Expect(sender.GetValue("TasksPending").Value).To(Equal(float64(2)))
			Expect(sender.GetValue("TasksRunning").Value).To(Equal(float64(1)))
			Expect(sender.GetValue("TasksCompleted").Value).To(Equal(float64(5)))
			Expect(sender.GetValue("TasksResolving").Value).To(Equal(float64(1)))

			Expect(sender.GetCounter("ConvergenceTasksPruned")).To(Equal(uint64(4)))
			Expect(sender.GetCounter("ConvergenceTasksKicked")).To(Equal(uint64(5)))
		})

		Context("pending tasks", func() {
			It("fails expired tasks", func() {
				task, err := sqlDB.TaskByGuid(logger, "pending-expired-task")
				Expect(err).NotTo(HaveOccurred())
				Expect(task.FailureReason).To(Equal("not started within time limit"))
				Expect(task.Failed).To(BeTrue())
				Expect(task.Result).To(Equal(""))
				Expect(task.State).To(Equal(models.Task_Completed))
				Expect(task.UpdatedAt).To(Equal(fakeClock.Now().UnixNano()))
				Expect(task.FirstCompletedAt).To(Equal(fakeClock.Now().UnixNano()))

				taskRequest := auctioneer.NewTaskStartRequestFromModel("pending-expired-task", domain, taskDef)
				Expect(tasksToAuction).NotTo(ContainElement(&taskRequest))
			})

			It("returns tasks that should be kicked for auctioning", func() {
				task, err := sqlDB.TaskByGuid(logger, "pending-kickable-task")
				Expect(err).NotTo(HaveOccurred())
				Expect(task.FailureReason).NotTo(Equal("not started within time limit"))
				Expect(task.Failed).NotTo(BeTrue())

				taskRequest := auctioneer.NewTaskStartRequestFromModel("pending-kickable-task", domain, taskDef)
				Expect(tasksToAuction).To(ContainElement(&taskRequest))
			})

			It("delete tasks that should be kicked if they're invalid", func() {
				_, err := sqlDB.TaskByGuid(logger, "pending-kickable-invalid-task")
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})

			It("doesn't do anything with unexpired tasks that should not be kicked", func() {
				taskRequest := auctioneer.NewTaskStartRequestFromModel("pending-task", domain, taskDef)
				Expect(tasksToAuction).NotTo(ContainElement(&taskRequest))

				task, err := sqlDB.TaskByGuid(logger, "pending-task")
				Expect(err).NotTo(HaveOccurred())
				Expect(task.FailureReason).NotTo(Equal("not started within time limit"))
				Expect(task.Failed).NotTo(BeTrue())
			})
		})

		Context("running tasks", func() {
			It("fails them when their cells are not present", func() {
				task, err := sqlDB.TaskByGuid(logger, "running-task-no-cell")
				Expect(err).NotTo(HaveOccurred())
				Expect(task.FailureReason).To(Equal("cell disappeared before completion"))
				Expect(task.Failed).To(BeTrue())
				Expect(task.Result).To(Equal(""))
				Expect(task.State).To(Equal(models.Task_Completed))
				Expect(task.UpdatedAt).To(Equal(fakeClock.Now().UnixNano()))
				Expect(task.FirstCompletedAt).To(Equal(fakeClock.Now().UnixNano()))
			})

			It("doesn't do anything when their cells are present", func() {
				taskRequest := auctioneer.NewTaskStartRequestFromModel("running-task", domain, taskDef)
				Expect(tasksToAuction).NotTo(ContainElement(taskRequest))

				task, err := sqlDB.TaskByGuid(logger, "running-task")
				Expect(err).NotTo(HaveOccurred())
				Expect(task.FailureReason).NotTo(Equal("cell disappeared before completion"))
				Expect(task.Failed).NotTo(BeTrue())
				Expect(task.State).To(Equal(models.Task_Running))
			})
		})

		Context("completed tasks", func() {
			It("deletes expired tasks", func() {
				_, err := sqlDB.TaskByGuid(logger, "completed-expired-task")
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})

			It("returns tasks that should be kicked for completion", func() {
				task, err := sqlDB.TaskByGuid(logger, "completed-kickable-task")
				Expect(err).NotTo(HaveOccurred())
				Expect(tasksToComplete).To(ContainElement(task))
			})

			It("doesn't do anything with unexpired tasks that should not be kicked", func() {
				task, err := sqlDB.TaskByGuid(logger, "completed-task")
				Expect(err).NotTo(HaveOccurred())
				Expect(tasksToComplete).NotTo(ContainElement(task))
			})

			It("delete tasks that should be kicked if they're invalid", func() {
				_, err := sqlDB.TaskByGuid(logger, "completed-kickable-invalid-task")
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})
		})

		Context("resolving tasks", func() {
			It("deletes expired tasks", func() {
				_, err := sqlDB.TaskByGuid(logger, "resolving-expired-task")
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})

			It("transitions the task back to the completed state if it should be kicked", func() {
				task, err := sqlDB.TaskByGuid(logger, "resolving-kickable-task")
				Expect(err).NotTo(HaveOccurred())
				Expect(task.State).To(Equal(models.Task_Completed))
			})

			It("returns tasks that should be kicked for completion", func() {
				task, err := sqlDB.TaskByGuid(logger, "resolving-kickable-task")
				Expect(err).NotTo(HaveOccurred())
				Expect(tasksToComplete).To(ContainElement(task))
			})

			It("doesn't do anything with unexpired tasks that should not be kicked", func() {
				task, err := sqlDB.TaskByGuid(logger, "resolving-task")
				Expect(err).NotTo(HaveOccurred())
				Expect(task.State).To(Equal(models.Task_Resolving))
				Expect(tasksToComplete).NotTo(ContainElement(task))
			})
		})

		Context("when the cell state list is empty", func() {
			BeforeEach(func() {
				cellSet = models.NewCellSetFromList([]*models.CellPresence{})
			})

			It("fails the running task", func() {
				task, err := sqlDB.TaskByGuid(logger, "running-task")
				Expect(err).NotTo(HaveOccurred())
				Expect(task.Failed).To(BeTrue())
				Expect(task.FailureReason).To(Equal("cell disappeared before completion"))
				Expect(task.Result).To(Equal(""))
			})
		})
	})
})
