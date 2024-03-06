package sqldb_test

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/format"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/models/test/model_helpers"
	"code.cloudfoundry.org/bbs/test_helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TaskDB", func() {
	Describe("DesireTask", func() {
		var (
			errDesire            error
			task                 *models.Task
			desiredTask          *models.Task
			taskDef              *models.TaskDefinition
			taskGuid, taskDomain string
		)

		JustBeforeEach(func() {
			desiredTask, errDesire = sqlDB.DesireTask(ctx, logger, taskDef, taskGuid, taskDomain)
		})

		BeforeEach(func() {
			taskGuid = "the-task-guid"
			task = model_helpers.NewValidTask(taskGuid)
			taskDomain = task.Domain
			taskDef = task.TaskDefinition
		})

		Context("when a task is not already present at the desired key", func() {
			It("persists the task", func() {
				Expect(errDesire).NotTo(HaveOccurred())

				queryStr := "SELECT * FROM tasks WHERE guid = ?"
				if test_helpers.UsePostgres() {
					queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
				}
				rows, err := db.QueryContext(ctx, queryStr, taskGuid)
				Expect(err).NotTo(HaveOccurred())
				defer rows.Close()
				Expect(rows.Next()).To(BeTrue())

				var guid, domain, cellID, failureReason, rejectionReason string
				var result sql.NullString
				var createdAt, updatedAt, firstCompletedAt int64
				var state, rejectionCount int32
				var failed bool
				var taskDefData []byte

				err = rows.Scan(
					&guid,
					&domain,
					&createdAt,
					&updatedAt,
					&firstCompletedAt,
					&state,
					&cellID,
					&result,
					&failed,
					&failureReason,
					&taskDefData,
					&rejectionCount,
					&rejectionReason,
				)
				Expect(err).NotTo(HaveOccurred())

				Expect(domain).To(Equal(taskDomain))
				Expect(guid).To(Equal(taskGuid))
				Expect(createdAt).To(Equal(fakeClock.Now().UTC().UnixNano()))
				Expect(updatedAt).To(Equal(fakeClock.Now().UTC().UnixNano()))
				Expect(firstCompletedAt).To(BeEquivalentTo(0))
				Expect(state).To(BeEquivalentTo(models.Task_Pending))
				Expect(result.String).To(Equal(""))
				Expect(failureReason).To(Equal(""))
				Expect(cellID).To(Equal(""))
				Expect(failed).To(BeFalse())
				Expect(rejectionCount).To(BeEquivalentTo(0))
				Expect(rejectionReason).To(Equal(""))

				var actualTaskDef models.TaskDefinition
				err = serializer.Unmarshal(logger, taskDefData, &actualTaskDef)
				Expect(err).NotTo(HaveOccurred())
				Expect(actualTaskDef).To(Equal(*taskDef))

				Expect(desiredTask.Domain).To(Equal(taskDomain))
				Expect(desiredTask.TaskGuid).To(Equal(taskGuid))
				Expect(desiredTask.CreatedAt).To(Equal(fakeClock.Now().UTC().UnixNano()))
				Expect(desiredTask.UpdatedAt).To(Equal(fakeClock.Now().UTC().UnixNano()))
				Expect(desiredTask.FirstCompletedAt).To(BeEquivalentTo(0))
				Expect(desiredTask.State).To(BeEquivalentTo(models.Task_Pending))
				Expect(desiredTask.Result).To(Equal(""))
				Expect(desiredTask.FailureReason).To(Equal(""))
				Expect(desiredTask.CellId).To(Equal(""))
				Expect(desiredTask.Failed).To(BeFalse())
				Expect(desiredTask.RejectionCount).To(BeEquivalentTo(0))
			})
		})

		Context("when a task is already present with the desired task guid", func() {
			BeforeEach(func() {
				otherDomain := "my-other-domain"
				_, err := sqlDB.DesireTask(ctx, logger, taskDef, taskGuid, otherDomain)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error and does not persist the task", func() {
				Expect(errDesire).To(Equal(models.ErrResourceExists))

				rows, err := db.QueryContext(ctx, "SELECT count(*) FROM tasks;")
				Expect(err).NotTo(HaveOccurred())
				defer rows.Close()
				Expect(rows.Next()).To(BeTrue())

				var count int
				err = rows.Scan(&count)
				Expect(err).NotTo(HaveOccurred())
				Expect(count).To(Equal(1))
			})
		})
	})

	Describe("Tasks", func() {
		Context("when there are tasks", func() {
			var expectedTasks []*models.Task

			BeforeEach(func() {
				task1 := model_helpers.NewValidTask("a-guid")
				task1.Domain = "domain-1"
				task1.CellId = "cell-1"
				task2 := model_helpers.NewValidTask("b-guid")
				task2.Domain = "domain-2"
				task2.CellId = "cell-2"
				task3 := model_helpers.NewValidTask("c-guid")
				task3.Domain = "domain-2"
				task3.CellId = "cell-1"
				expectedTasks = []*models.Task{task1, task2, task3}

				for _, t := range expectedTasks {
					insertTask(ctx, db, serializer, t, false)
				}
			})

			It("returns all the tasks", func() {
				tasks, err := sqlDB.Tasks(ctx, logger, models.TaskFilter{})
				Expect(err).NotTo(HaveOccurred())
				Expect(tasks).To(ConsistOf(expectedTasks))
			})

			It("can filter by domain", func() {
				tasks, err := sqlDB.Tasks(ctx, logger, models.TaskFilter{Domain: "domain-1"})
				Expect(err).NotTo(HaveOccurred())
				Expect(tasks).To(HaveLen(1))
				Expect(tasks[0]).To(Equal(expectedTasks[0]))
			})

			It("can filter by cell id", func() {
				tasks, err := sqlDB.Tasks(ctx, logger, models.TaskFilter{CellID: "cell-2"})
				Expect(err).NotTo(HaveOccurred())
				Expect(tasks).To(HaveLen(1))
				Expect(tasks[0]).To(Equal(expectedTasks[1]))
			})

			It("can filter by domain and cell id", func() {
				tasks, err := sqlDB.Tasks(ctx, logger, models.TaskFilter{CellID: "cell-1", Domain: "domain-2"})
				Expect(err).NotTo(HaveOccurred())
				Expect(tasks).To(HaveLen(1))
				Expect(tasks[0]).To(Equal(expectedTasks[2]))
			})
		})

		Context("when there are no tasks", func() {
			It("returns an empty list", func() {
				tasks, err := sqlDB.Tasks(ctx, logger, models.TaskFilter{})
				Expect(err).NotTo(HaveOccurred())
				Expect(tasks).NotTo(BeNil())
				Expect(tasks).To(BeEmpty())
			})
		})

		Context("when there is invalid task definition data", func() {
			BeforeEach(func() {
				task1 := model_helpers.NewValidTask("a-guid")
				insertTask(ctx, db, serializer, task1, true)
			})

			It("errors", func() {
				_, err := sqlDB.Tasks(ctx, logger, models.TaskFilter{})
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("TaskByGuid", func() {
		Context("when there is a task", func() {
			var expectedTask *models.Task

			BeforeEach(func() {
				expectedTask = model_helpers.NewValidTask("task-guid")
				insertTask(ctx, db, serializer, expectedTask, false)
			})

			It("returns the task", func() {
				task, err := sqlDB.TaskByGuid(ctx, logger, "task-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(task).To(Equal(expectedTask))
			})
		})

		Context("when there is no task", func() {
			It("returns a ResourceNotFound", func() {
				_, err := sqlDB.TaskByGuid(ctx, logger, "nota-guid")
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})
		})

		Context("when there is invalid data", func() {
			BeforeEach(func() {
				task1 := model_helpers.NewValidTask("a-guid")
				insertTask(ctx, db, serializer, task1, true)
			})

			It("errors", func() {
				_, err := sqlDB.TaskByGuid(ctx, logger, "a-guid")
				Expect(err).To(Equal(models.ErrDeserialize))
			})
		})
	})

	Describe("StartTask", func() {
		var (
			expectedTask, beforeTask *models.Task
		)

		BeforeEach(func() {
			var err error
			expectedTask = model_helpers.NewValidTask("task-guid")
			beforeTask, err = sqlDB.DesireTask(ctx, logger, expectedTask.TaskDefinition, expectedTask.TaskGuid, expectedTask.Domain)
			Expect(err).NotTo(HaveOccurred())
		})

		It("starts the task", func() {
			fakeClock.IncrementBySeconds(1)

			expectedTask.CellId = "expectedCellId"

			before, after, started, err := sqlDB.StartTask(ctx, logger, expectedTask.TaskGuid, expectedTask.CellId)
			Expect(err).NotTo(HaveOccurred())
			Expect(started).To(BeTrue())
			Expect(before).To(Equal(beforeTask))

			Expect(after.TaskGuid).To(Equal(expectedTask.TaskGuid))
			Expect(after.State).To(Equal(models.Task_Running))
			Expect(after.CellId).To(Equal(expectedTask.CellId))
			Expect(after.TaskDefinition).To(BeEquivalentTo(expectedTask.TaskDefinition))
			Expect(after.UpdatedAt).To(Equal(fakeClock.Now().UnixNano()))

			task, err := sqlDB.TaskByGuid(ctx, logger, expectedTask.TaskGuid)
			Expect(err).NotTo(HaveOccurred())

			Expect(task.TaskGuid).To(Equal(expectedTask.TaskGuid))
			Expect(task.State).To(Equal(models.Task_Running))
			Expect(task.CellId).To(Equal(expectedTask.CellId))
			Expect(task.TaskDefinition).To(BeEquivalentTo(expectedTask.TaskDefinition))
			Expect(task.UpdatedAt).To(Equal(fakeClock.Now().UnixNano()))
		})

		Context("when the cell id is toooooo long", func() {
			It("returns a BadRequest error", func() {
				_, _, started, err := sqlDB.StartTask(ctx, logger, expectedTask.TaskGuid, randStr(256))
				Expect(err).To(Equal(models.ErrBadRequest))
				Expect(started).To(BeFalse())
			})
		})

		Context("When starting a Task that is already started", func() {
			var beforeTask *models.Task
			BeforeEach(func() {
				var err error
				var started bool
				_, beforeTask, started, err = sqlDB.StartTask(ctx, logger, expectedTask.TaskGuid, "cell-id")
				Expect(err).NotTo(HaveOccurred())
				Expect(started).To(BeTrue())
			})

			Context("on the same cell", func() {
				It("returns shouldStart as false", func() {
					fakeClock.IncrementBySeconds(1)

					before, after, changed, err := sqlDB.StartTask(ctx, logger, expectedTask.TaskGuid, "cell-id")
					Expect(err).NotTo(HaveOccurred())
					Expect(changed).To(BeFalse())

					Expect(before).To(BeEquivalentTo(after))

					task, err := sqlDB.TaskByGuid(ctx, logger, expectedTask.TaskGuid)
					Expect(err).NotTo(HaveOccurred())
					Expect(task).To(BeEquivalentTo(beforeTask))
				})
			})

			Context("on another cell", func() {
				It("returns an error", func() {
					fakeClock.IncrementBySeconds(1)

					_, _, _, err := sqlDB.StartTask(ctx, logger, expectedTask.TaskGuid, "some-other-cell")
					modelErr := models.ConvertError(err)
					Expect(modelErr).NotTo(BeNil())
					Expect(modelErr.Type).To(Equal(models.Error_InvalidStateTransition))

					task, err := sqlDB.TaskByGuid(ctx, logger, expectedTask.TaskGuid)
					Expect(err).NotTo(HaveOccurred())
					Expect(task).To(BeEquivalentTo(beforeTask))
				})
			})
		})

		Context("when the task does not exist", func() {
			It("returns an error", func() {
				_, _, started, err := sqlDB.StartTask(ctx, logger, "invalid-guid", "cell-id")
				Expect(err).To(Equal(models.ErrResourceNotFound))
				Expect(started).To(BeFalse())
			})
		})

		Context("when the task is already completed", func() {
			BeforeEach(func() {
				expectedTask = model_helpers.NewValidTask("task-other-guid")
				expectedTask.State = models.Task_Completed
				expectedTask.CellId = "completed-guid"
				insertTask(ctx, db, serializer, expectedTask, false)
			})

			It("returns an invalid state transition", func() {
				_, _, started, err := sqlDB.StartTask(ctx, logger, "task-other-guid", "completed-guid")
				modelErr := models.ConvertError(err)
				Expect(modelErr).NotTo(BeNil())
				Expect(modelErr.Type).To(Equal(models.Error_InvalidStateTransition))
				Expect(started).To(BeFalse())

				task, err := sqlDB.TaskByGuid(ctx, logger, expectedTask.TaskGuid)
				Expect(err).NotTo(HaveOccurred())
				Expect(task).To(BeEquivalentTo(expectedTask))
			})
		})
	})

	Describe("CancelTask", func() {
		var (
			taskGuid, taskDomain string
			taskDefinition       *models.TaskDefinition
		)

		BeforeEach(func() {
			taskGuid = "the-task-guid"
			taskDomain = "the-task-domain"
			taskDefinition = model_helpers.NewValidTaskDefinition()
		})

		Context("when the task is pending", func() {
			var beforeTask *models.Task
			BeforeEach(func() {
				var err error
				beforeTask, err = sqlDB.DesireTask(ctx, logger, taskDefinition, taskGuid, taskDomain)
				Expect(err).NotTo(HaveOccurred())
			})

			It("cancels the task", func() {
				fakeClock.Increment(time.Second)
				now := fakeClock.Now().UnixNano()

				before, after, cellID, err := sqlDB.CancelTask(ctx, logger, taskGuid)
				Expect(err).NotTo(HaveOccurred())

				Expect(before).To(Equal(beforeTask))
				Expect(after.State).To(Equal(models.Task_Completed))
				Expect(after.UpdatedAt).To(Equal(now))
				Expect(after.FirstCompletedAt).To(Equal(now))
				Expect(after.Failed).To(BeTrue())
				Expect(after.FailureReason).To(Equal("task was cancelled"))
				Expect(after.Result).To(Equal(""))
				Expect(after.CellId).To(Equal(""))
				Expect(cellID).To(Equal(""))

				task, err := sqlDB.TaskByGuid(ctx, logger, taskGuid)
				Expect(err).NotTo(HaveOccurred())

				Expect(task.State).To(Equal(models.Task_Completed))
				Expect(task.UpdatedAt).To(Equal(now))
				Expect(task.FirstCompletedAt).To(Equal(now))
				Expect(task.Failed).To(BeTrue())
				Expect(task.FailureReason).To(Equal("task was cancelled"))
				Expect(task.Result).To(Equal(""))
				Expect(task.CellId).To(Equal(""))
			})

			Context("when there are multiple tasks", func() {
				var anotherTask *models.Task

				BeforeEach(func() {
					var err error
					anotherTaskGuid := "the-other-task-guid"
					anotherTask, err = sqlDB.DesireTask(ctx, logger, taskDefinition, anotherTaskGuid, taskDomain)
					Expect(err).NotTo(HaveOccurred())
				})

				It("does not update the other task", func() {
					fakeClock.Increment(time.Second)
					now := fakeClock.Now().UnixNano()

					_, after, _, err := sqlDB.CancelTask(ctx, logger, taskGuid)
					Expect(err).NotTo(HaveOccurred())

					Expect(after.State).To(Equal(models.Task_Completed))
					Expect(after.UpdatedAt).To(Equal(now))

					task, err := sqlDB.TaskByGuid(ctx, logger, anotherTask.TaskGuid)
					Expect(err).NotTo(HaveOccurred())
					Expect(task).To(BeEquivalentTo(anotherTask))
				})
			})
		})

		Context("when the task is running", func() {
			var beforeTask *models.Task

			BeforeEach(func() {
				_, err := sqlDB.DesireTask(ctx, logger, taskDefinition, taskGuid, taskDomain)
				Expect(err).NotTo(HaveOccurred())

				var started bool
				_, beforeTask, started, err = sqlDB.StartTask(ctx, logger, taskGuid, "the-cell")
				Expect(err).NotTo(HaveOccurred())
				Expect(started).To(BeTrue())
			})

			It("cancels the task", func() {
				fakeClock.Increment(time.Second)
				now := fakeClock.Now().UnixNano()

				before, after, cellID, err := sqlDB.CancelTask(ctx, logger, taskGuid)
				Expect(err).NotTo(HaveOccurred())
				Expect(before).To(Equal(beforeTask))

				Expect(after.State).To(Equal(models.Task_Completed))
				Expect(after.UpdatedAt).To(Equal(now))
				Expect(after.FirstCompletedAt).To(Equal(now))
				Expect(after.Failed).To(BeTrue())
				Expect(after.FailureReason).To(Equal("task was cancelled"))
				Expect(after.Result).To(Equal(""))
				Expect(after.CellId).To(Equal(""))
				Expect(cellID).To(Equal("the-cell"))

				task, err := sqlDB.TaskByGuid(ctx, logger, taskGuid)
				Expect(err).NotTo(HaveOccurred())

				Expect(task.State).To(Equal(models.Task_Completed))
				Expect(task.UpdatedAt).To(Equal(now))
				Expect(task.FirstCompletedAt).To(Equal(now))
				Expect(task.Failed).To(BeTrue())
				Expect(task.FailureReason).To(Equal("task was cancelled"))
				Expect(task.Result).To(Equal(""))
				Expect(task.CellId).To(Equal(""))
			})
		})

		Context("when the task is already completed", func() {
			var beforeTask *models.Task

			BeforeEach(func() {
				_, err := sqlDB.DesireTask(ctx, logger, taskDefinition, taskGuid, taskDomain)
				Expect(err).NotTo(HaveOccurred())

				_, beforeTask, _, err = sqlDB.CancelTask(ctx, logger, taskGuid)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an InvalidStateTransition error", func() {
				_, _, _, err := sqlDB.CancelTask(ctx, logger, taskGuid)
				modelErr := models.ConvertError(err)
				Expect(modelErr).NotTo(BeNil())
				Expect(modelErr.Type).To(Equal(models.Error_InvalidStateTransition))

				task, err := sqlDB.TaskByGuid(ctx, logger, taskGuid)
				Expect(err).NotTo(HaveOccurred())
				Expect(task).To(BeEquivalentTo(beforeTask))
			})
		})

		Context("when the task is already resolving", func() {
			var beforeTask *models.Task

			BeforeEach(func() {
				beforeTask = model_helpers.NewValidTask(taskGuid)
				beforeTask.State = models.Task_Resolving
				insertTask(ctx, db, serializer, beforeTask, false)
			})

			It("returns an InvalidStateTransition error", func() {
				_, _, _, err := sqlDB.CancelTask(ctx, logger, taskGuid)
				modelErr := models.ConvertError(err)
				Expect(modelErr).NotTo(BeNil())
				Expect(modelErr.Type).To(Equal(models.Error_InvalidStateTransition))

				task, err := sqlDB.TaskByGuid(ctx, logger, taskGuid)
				Expect(err).NotTo(HaveOccurred())
				Expect(task).To(BeEquivalentTo(beforeTask))
			})
		})

		Context("when the task does not exist", func() {
			It("returns an InvalidStateTransition error", func() {
				_, _, _, err := sqlDB.CancelTask(ctx, logger, taskGuid)
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})
		})
	})

	Describe("CompleteTask", func() {
		var (
			taskGuid, taskDomain, cellID string
			taskDefinition               *models.TaskDefinition
			taskBefore                   *models.Task
		)

		BeforeEach(func() {
			taskGuid = "the-task-guid"
			taskDomain = "the-task-domain"
			taskDefinition = model_helpers.NewValidTaskDefinition()
			cellID = "the-cell"
		})

		Context("when the task exists", func() {
			JustBeforeEach(func() {
				var err error
				taskBefore, err = sqlDB.TaskByGuid(ctx, logger, taskGuid)
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when the task is running", func() {
				var beforeTask *models.Task
				BeforeEach(func() {
					_, err := sqlDB.DesireTask(ctx, logger, taskDefinition, taskGuid, taskDomain)
					Expect(err).NotTo(HaveOccurred())

					var started bool
					_, beforeTask, started, err = sqlDB.StartTask(ctx, logger, taskGuid, cellID)
					Expect(err).NotTo(HaveOccurred())
					Expect(started).To(BeTrue())
				})

				Context("on the same cell", func() {
					It("completes the task with the specified values", func() {
						fakeClock.Increment(time.Second)
						nowTruncateMicroseconds := fakeClock.Now()
						now := fakeClock.Now()

						before, after, err := sqlDB.CompleteTask(ctx, logger, taskGuid, cellID, true, "it blew up", "i am the result")
						Expect(err).NotTo(HaveOccurred())

						Expect(before).To(Equal(beforeTask))

						Expect(after.State).To(Equal(models.Task_Completed))
						Expect(after.UpdatedAt).To(Equal(now.UnixNano()))
						Expect(after.FirstCompletedAt).To(Equal(now.UnixNano()))
						Expect(after.Failed).To(BeTrue())
						Expect(after.FailureReason).To(Equal("it blew up"))
						Expect(after.Result).To(Equal("i am the result"))
						Expect(after.CellId).To(Equal(""))

						task, err := sqlDB.TaskByGuid(ctx, logger, taskGuid)
						Expect(err).NotTo(HaveOccurred())

						Expect(task.State).To(Equal(models.Task_Completed))
						Expect(task.UpdatedAt).To(Equal(nowTruncateMicroseconds.UnixNano()))
						Expect(task.FirstCompletedAt).To(Equal(nowTruncateMicroseconds.UnixNano()))
						Expect(task.Failed).To(BeTrue())
						Expect(task.FailureReason).To(Equal("it blew up"))
						Expect(task.Result).To(Equal("i am the result"))
						Expect(task.CellId).To(Equal(""))
					})

					Context("when the rejection reason is longer than 1K", func() {
						var (
							failureReason string
							after         *models.Task
						)

						JustBeforeEach(func() {
							failureReason = strings.Repeat("x", 2*1024)
							var err error
							_, after, err = sqlDB.CompleteTask(ctx, logger, taskGuid, cellID, true, failureReason, "i am the result")
							Expect(err).NotTo(HaveOccurred())
						})

						It("truncates the crash reason in the returned task", func() {
							expectedFailureReason := strings.Repeat("x", 1013) + "(truncated)"
							Expect(after.FailureReason).To(Equal(expectedFailureReason))
						})

						It("truncates the crash reason", func() {
							expectedFailureReason := strings.Repeat("x", 1013) + "(truncated)"
							task, err := sqlDB.TaskByGuid(ctx, logger, taskGuid)
							Expect(err).NotTo(HaveOccurred())
							Expect(task.FailureReason).To(Equal(expectedFailureReason))
						})
					})

					Context("with multiple tasks", func() {
						var anotherTask *models.Task

						BeforeEach(func() {
							anotherTaskGuid := "another-task-guid"
							_, err := sqlDB.DesireTask(ctx, logger, taskDefinition, anotherTaskGuid, taskDomain)
							Expect(err).NotTo(HaveOccurred())

							_, _, started, err := sqlDB.StartTask(ctx, logger, anotherTaskGuid, cellID)
							Expect(err).NotTo(HaveOccurred())
							Expect(started).To(BeTrue())

							anotherTask, err = sqlDB.TaskByGuid(ctx, logger, anotherTaskGuid)
							Expect(err).NotTo(HaveOccurred())
						})

						It("only updates the task with the corresponding guid", func() {
							_, _, err := sqlDB.CompleteTask(ctx, logger, taskGuid, cellID, true, "it blew up", "i am the result")
							Expect(err).NotTo(HaveOccurred())

							task, err := sqlDB.TaskByGuid(ctx, logger, anotherTask.TaskGuid)
							Expect(err).NotTo(HaveOccurred())
							Expect(task).To(BeEquivalentTo(anotherTask))
						})
					})
				})

				Context("on a different cell", func() {
					It("errors and does not change the task", func() {
						_, _, err := sqlDB.CompleteTask(ctx, logger, taskGuid, "a-different-cell", true, "it blue up", "i am the result")
						modelErr := models.ConvertError(err)
						Expect(modelErr).NotTo(BeNil())
						Expect(modelErr.Type).To(Equal(models.Error_RunningOnDifferentCell))
						Expect(modelErr.Message).To(Equal("Running on cell the-cell not a-different-cell"))

						task, err := sqlDB.TaskByGuid(ctx, logger, taskGuid)
						Expect(err).NotTo(HaveOccurred())
						Expect(task).To(BeEquivalentTo(taskBefore))
					})
				})
			})

			Context("when the task is not running", func() {
				BeforeEach(func() {
					task := model_helpers.NewValidTask(taskGuid)
					task.State = models.Task_Pending
					task.CellId = cellID
					insertTask(ctx, db, serializer, task, false)
				})

				It("errors and does not change the task", func() {
					before, after, err := sqlDB.CompleteTask(ctx, logger, taskGuid, cellID, true, "it blue up", "i am the result")
					modelErr := models.ConvertError(err)
					Expect(modelErr).NotTo(BeNil())
					Expect(modelErr.Type).To(Equal(models.Error_InvalidStateTransition))

					Expect(before).To(BeEquivalentTo(taskBefore))
					Expect(before).To(BeEquivalentTo(after))

					task, err := sqlDB.TaskByGuid(ctx, logger, taskGuid)
					Expect(err).NotTo(HaveOccurred())
					Expect(task).To(BeEquivalentTo(taskBefore))
				})
			})
		})

		Context("when the task does not exist", func() {
			It("errors", func() {
				_, _, err := sqlDB.CompleteTask(ctx, logger, "task-not-here", "a-different-cell", true, "it blue up", "i am the result")
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})
		})
	})

	Describe("FailTask", func() {
		Context("when the task exists", func() {
			var (
				taskGuid, taskDomain, failureReason, cellID string
				taskDefinition                              *models.TaskDefinition
				beforeTask                                  *models.Task
			)

			BeforeEach(func() {
				var err error

				taskGuid = "the-task-guid"
				taskDomain = "the-task-domain"
				taskDefinition = model_helpers.NewValidTaskDefinition()
				failureReason = "I failed."

				beforeTask, err = sqlDB.DesireTask(ctx, logger, taskDefinition, taskGuid, taskDomain)
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when the task is pending", func() {
				It("fails the task", func() {
					fakeClock.Increment(time.Second)
					nowTruncateMicroseconds := fakeClock.Now()
					now := fakeClock.Now()

					before, after, err := sqlDB.FailTask(ctx, logger, taskGuid, failureReason)
					Expect(err).NotTo(HaveOccurred())
					Expect(before).To(Equal(beforeTask))

					Expect(after.State).To(Equal(models.Task_Completed))
					Expect(after.UpdatedAt).To(Equal(now.UnixNano()))
					Expect(after.FirstCompletedAt).To(Equal(now.UnixNano()))
					Expect(after.Failed).To(BeTrue())
					Expect(after.FailureReason).To(Equal("I failed."))
					Expect(after.Result).To(Equal(""))
					Expect(after.CellId).To(Equal(""))

					task, err := sqlDB.TaskByGuid(ctx, logger, taskGuid)
					Expect(err).NotTo(HaveOccurred())

					Expect(task.State).To(Equal(models.Task_Completed))
					Expect(task.UpdatedAt).To(Equal(nowTruncateMicroseconds.UnixNano()))
					Expect(task.FirstCompletedAt).To(Equal(nowTruncateMicroseconds.UnixNano()))
					Expect(task.Failed).To(BeTrue())
					Expect(task.FailureReason).To(Equal("I failed."))
					Expect(task.Result).To(Equal(""))
					Expect(task.CellId).To(Equal(""))
				})

				Context("with multiple tasks pending", func() {
					var anotherTask *models.Task
					BeforeEach(func() {
						anotherTaskGuid := "another-task-guid"
						_, err := sqlDB.DesireTask(ctx, logger, taskDefinition, anotherTaskGuid, taskDomain)
						Expect(err).NotTo(HaveOccurred())

						anotherTask, err = sqlDB.TaskByGuid(ctx, logger, anotherTaskGuid)
						Expect(err).NotTo(HaveOccurred())
					})

					It("updates only the task with the corresponding guid", func() {
						_, _, err := sqlDB.FailTask(ctx, logger, taskGuid, failureReason)
						Expect(err).NotTo(HaveOccurred())

						task, err := sqlDB.TaskByGuid(ctx, logger, anotherTask.TaskGuid)
						Expect(err).NotTo(HaveOccurred())
						Expect(task).To(BeEquivalentTo(anotherTask))
					})
				})

				Context("when the failure reason is longer than 1K", func() {
					It("truncates the crash reason", func() {
						failureReason := strings.Repeat("x", 2*1024)
						expectedFailureReason := strings.Repeat("x", 1013) + "(truncated)"
						_, _, err := sqlDB.FailTask(ctx, logger, taskGuid, failureReason)
						Expect(err).NotTo(HaveOccurred())

						task, err := sqlDB.TaskByGuid(ctx, logger, taskGuid)
						Expect(err).NotTo(HaveOccurred())
						Expect(task.FailureReason).To(Equal(expectedFailureReason))
					})
				})
			})

			Context("when the task is running", func() {
				var beforeTask *models.Task
				BeforeEach(func() {
					var err error
					var started bool
					cellID = "the-cell-id"
					_, beforeTask, started, err = sqlDB.StartTask(ctx, logger, taskGuid, cellID)
					Expect(err).NotTo(HaveOccurred())
					Expect(started).To(BeTrue())
				})

				It("fails the task", func() {
					fakeClock.Increment(time.Second)
					nowTruncateMicroseconds := fakeClock.Now()
					now := fakeClock.Now()

					failureReason := "I failed."

					before, after, err := sqlDB.FailTask(ctx, logger, taskGuid, failureReason)
					Expect(err).NotTo(HaveOccurred())

					Expect(before).To(Equal(beforeTask))

					Expect(after.State).To(Equal(models.Task_Completed))
					Expect(after.UpdatedAt).To(Equal(now.UnixNano()))
					Expect(after.FirstCompletedAt).To(Equal(now.UnixNano()))
					Expect(after.Failed).To(BeTrue())
					Expect(after.FailureReason).To(Equal("I failed."))
					Expect(after.Result).To(Equal(""))
					Expect(after.CellId).To(Equal(""))

					task, err := sqlDB.TaskByGuid(ctx, logger, taskGuid)
					Expect(err).NotTo(HaveOccurred())

					Expect(task.State).To(Equal(models.Task_Completed))
					Expect(task.UpdatedAt).To(Equal(nowTruncateMicroseconds.UnixNano()))
					Expect(task.FirstCompletedAt).To(Equal(nowTruncateMicroseconds.UnixNano()))
					Expect(task.Failed).To(BeTrue())
					Expect(task.FailureReason).To(Equal("I failed."))
					Expect(task.Result).To(Equal(""))
					Expect(task.CellId).To(Equal(""))
				})
			})

			Context("when the task is completed", func() {
				var beforeTask *models.Task

				BeforeEach(func() {
					cellID = "the-cell-id"
					_, _, started, err := sqlDB.StartTask(ctx, logger, taskGuid, cellID)
					Expect(err).NotTo(HaveOccurred())
					Expect(started).To(BeTrue())

					_, _, err = sqlDB.CompleteTask(ctx, logger, taskGuid, cellID, false, "", "I am the result.")
					Expect(err).NotTo(HaveOccurred())

					beforeTask, err = sqlDB.TaskByGuid(ctx, logger, taskGuid)
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns an InvalidStateTransition error", func() {
					_, _, err := sqlDB.FailTask(ctx, logger, taskGuid, failureReason)
					modelErr := models.ConvertError(err)
					Expect(modelErr).NotTo(BeNil())
					Expect(modelErr.Type).To(Equal(models.Error_InvalidStateTransition))

					task, err := sqlDB.TaskByGuid(ctx, logger, taskGuid)
					Expect(err).NotTo(HaveOccurred())
					Expect(task).To(BeEquivalentTo(beforeTask))
				})
			})

			Context("when the task is resolving", func() {
				var beforeTask *models.Task

				BeforeEach(func() {
					var err error
					taskGuid = "new-task-guid"

					beforeTask = model_helpers.NewValidTask(taskGuid)
					beforeTask.State = models.Task_Resolving
					insertTask(ctx, db, serializer, beforeTask, false)

					beforeTask, err = sqlDB.TaskByGuid(ctx, logger, taskGuid)
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns an InvalidStateTransition error", func() {
					_, _, err := sqlDB.FailTask(ctx, logger, taskGuid, failureReason)
					modelErr := models.ConvertError(err)
					Expect(modelErr).NotTo(BeNil())
					Expect(modelErr.Type).To(Equal(models.Error_InvalidStateTransition))

					task, err := sqlDB.TaskByGuid(ctx, logger, taskGuid)
					Expect(err).NotTo(HaveOccurred())
					Expect(task).To(BeEquivalentTo(beforeTask))
				})
			})
		})

		Context("when the task does not exist", func() {
			It("returns an ResourceNotFound error", func() {
				_, _, err := sqlDB.FailTask(ctx, logger, "", "nota-guid")
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})
		})
	})

	Describe("ResolvingTask", func() {
		var taskGuid string

		BeforeEach(func() {
			taskGuid = "the-task-guid"
		})

		Context("when the task exists", func() {
			var (
				taskDomain, cellID string
				taskDefinition     *models.TaskDefinition
			)

			BeforeEach(func() {
				taskDomain = "the-task-domain"
				cellID = "the-cell-id"
				taskDefinition = model_helpers.NewValidTaskDefinition()

				_, err := sqlDB.DesireTask(ctx, logger, taskDefinition, taskGuid, taskDomain)
				Expect(err).NotTo(HaveOccurred())

				_, _, started, err := sqlDB.StartTask(ctx, logger, taskGuid, cellID)
				Expect(err).NotTo(HaveOccurred())
				Expect(started).To(BeTrue())
			})

			Context("when the task is completed", func() {
				var beforeTask *models.Task
				BeforeEach(func() {
					var err error
					_, beforeTask, err = sqlDB.CompleteTask(ctx, logger, taskGuid, cellID, false, "", "some-result")
					Expect(err).NotTo(HaveOccurred())
				})

				It("resolves the task", func() {
					fakeClock.Increment(time.Second)
					nowTruncateMicroseconds := fakeClock.Now()

					before, after, err := sqlDB.ResolvingTask(ctx, logger, taskGuid)
					Expect(err).NotTo(HaveOccurred())

					Expect(before).To(Equal(beforeTask))
					Expect(after.State).To(Equal(models.Task_Resolving))
					Expect(after.UpdatedAt).To(Equal(nowTruncateMicroseconds.UnixNano()))

					task, err := sqlDB.TaskByGuid(ctx, logger, taskGuid)
					Expect(err).NotTo(HaveOccurred())

					Expect(task.State).To(Equal(models.Task_Resolving))
					Expect(task.UpdatedAt).To(Equal(nowTruncateMicroseconds.UnixNano()))
				})

				Context("with multiple completed tasks", func() {
					var anotherTask *models.Task

					BeforeEach(func() {
						anotherTaskGuid := "another-guid"
						_, err := sqlDB.DesireTask(ctx, logger, taskDefinition, anotherTaskGuid, taskDomain)
						Expect(err).NotTo(HaveOccurred())

						_, _, started, err := sqlDB.StartTask(ctx, logger, anotherTaskGuid, cellID)
						Expect(err).NotTo(HaveOccurred())
						Expect(started).To(BeTrue())

						_, _, err = sqlDB.CompleteTask(ctx, logger, anotherTaskGuid, cellID, false, "", "some-result")
						Expect(err).NotTo(HaveOccurred())

						anotherTask, err = sqlDB.TaskByGuid(ctx, logger, anotherTaskGuid)
						Expect(err).NotTo(HaveOccurred())
					})

					It("should only update the task with the corresponding guid", func() {
						_, _, err := sqlDB.ResolvingTask(ctx, logger, taskGuid)
						Expect(err).NotTo(HaveOccurred())

						task, err := sqlDB.TaskByGuid(ctx, logger, anotherTask.TaskGuid)
						Expect(err).NotTo(HaveOccurred())

						Expect(task).To(BeEquivalentTo(anotherTask))
					})
				})
			})

			Context("when the task is not completed", func() {
				var taskBefore *models.Task

				BeforeEach(func() {
					var err error
					taskBefore, err = sqlDB.TaskByGuid(ctx, logger, taskGuid)
					Expect(err).NotTo(HaveOccurred())
				})

				It("errors and does not change the task", func() {
					_, _, err := sqlDB.ResolvingTask(ctx, logger, taskGuid)
					modelErr := models.ConvertError(err)
					Expect(modelErr).NotTo(BeNil())
					Expect(modelErr.Type).To(Equal(models.Error_InvalidStateTransition))

					task, err := sqlDB.TaskByGuid(ctx, logger, taskGuid)
					Expect(err).NotTo(HaveOccurred())
					Expect(task).To(BeEquivalentTo(taskBefore))
				})
			})

			Context("when the task is already resolving", func() {
				var taskBefore *models.Task

				BeforeEach(func() {
					_, _, err := sqlDB.CompleteTask(ctx, logger, taskGuid, cellID, false, "", "some-result")
					Expect(err).NotTo(HaveOccurred())

					_, taskBefore, err = sqlDB.ResolvingTask(ctx, logger, taskGuid)
					Expect(err).NotTo(HaveOccurred())
				})

				It("errors and does not change the task", func() {
					_, _, err := sqlDB.ResolvingTask(ctx, logger, taskGuid)
					modelErr := models.ConvertError(err)
					Expect(modelErr).NotTo(BeNil())
					Expect(modelErr.Type).To(Equal(models.Error_InvalidStateTransition))

					task, err := sqlDB.TaskByGuid(ctx, logger, taskGuid)
					Expect(err).NotTo(HaveOccurred())
					Expect(task).To(BeEquivalentTo(taskBefore))
				})
			})
		})

		Context("when the task does not exist", func() {
			It("returns a ResourceNotFound error", func() {
				_, _, err := sqlDB.ResolvingTask(ctx, logger, taskGuid)
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})
		})
	})

	Describe("DeleteTask", func() {
		var taskGuid string

		BeforeEach(func() {
			taskGuid = "the-task-guid"
		})

		Context("when the task exists", func() {
			var (
				taskDomain, cellID string
				taskDefinition     *models.TaskDefinition
			)

			BeforeEach(func() {
				taskDomain = "the-task-domain"
				cellID = "the-cell-id"
				taskDefinition = model_helpers.NewValidTaskDefinition()

				_, err := sqlDB.DesireTask(ctx, logger, taskDefinition, taskGuid, taskDomain)
				Expect(err).NotTo(HaveOccurred())

				_, _, started, err := sqlDB.StartTask(ctx, logger, taskGuid, cellID)
				Expect(err).NotTo(HaveOccurred())
				Expect(started).To(BeTrue())

				_, _, err = sqlDB.CompleteTask(ctx, logger, taskGuid, cellID, false, "", "some-result")
				Expect(err).NotTo(HaveOccurred())
			})

			Context("and the task is resolving", func() {
				var beforeTask *models.Task
				BeforeEach(func() {
					var err error
					_, beforeTask, err = sqlDB.ResolvingTask(ctx, logger, taskGuid)
					Expect(err).NotTo(HaveOccurred())
				})

				It("removes the task from the database", func() {
					task, err := sqlDB.DeleteTask(ctx, logger, taskGuid)
					Expect(err).NotTo(HaveOccurred())

					Expect(task).To(Equal(beforeTask))

					_, err = sqlDB.TaskByGuid(ctx, logger, taskGuid)
					Expect(err).To(Equal(models.ErrResourceNotFound))
				})

				Context("with multiple resolving tasks", func() {
					var anotherTask *models.Task

					BeforeEach(func() {
						anotherTaskGuid := "another-guid"

						_, err := sqlDB.DesireTask(ctx, logger, taskDefinition, anotherTaskGuid, taskDomain)
						Expect(err).NotTo(HaveOccurred())

						_, _, started, err := sqlDB.StartTask(ctx, logger, anotherTaskGuid, cellID)
						Expect(err).NotTo(HaveOccurred())
						Expect(started).To(BeTrue())

						_, _, err = sqlDB.CompleteTask(ctx, logger, anotherTaskGuid, cellID, false, "", "some-result")
						Expect(err).NotTo(HaveOccurred())

						_, _, err = sqlDB.ResolvingTask(ctx, logger, anotherTaskGuid)
						Expect(err).NotTo(HaveOccurred())

						anotherTask, err = sqlDB.TaskByGuid(ctx, logger, anotherTaskGuid)
						Expect(err).NotTo(HaveOccurred())
					})

					It("only removes the task with the corresponding guid", func() {
						_, err := sqlDB.DeleteTask(ctx, logger, taskGuid)
						Expect(err).NotTo(HaveOccurred())

						task, err := sqlDB.TaskByGuid(ctx, logger, anotherTask.TaskGuid)
						Expect(err).NotTo(HaveOccurred())
						Expect(task).To(BeEquivalentTo(anotherTask))
					})
				})
			})

			Context("and the task is not resolving", func() {
				It("returns an error", func() {
					_, err := sqlDB.DeleteTask(ctx, logger, taskGuid)
					expectedErr := models.NewTaskTransitionError(models.Task_Completed, models.Task_Resolving)
					Expect(err).To(Equal(expectedErr))
				})
			})
		})

		Context("when the task does not exist", func() {
			It("returns a ResourceNotFound error", func() {
				_, err := sqlDB.DeleteTask(ctx, logger, taskGuid)
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})
		})
	})

	Describe("RejectTask", func() {
		Context("when the task exists", func() {
			var (
				taskGuid, taskDomain string
				taskDefinition       *models.TaskDefinition
				beforeTask           *models.Task
			)

			BeforeEach(func() {
				var err error

				taskGuid = "the-task-guid"
				taskDomain = "the-task-domain"
				taskDefinition = model_helpers.NewValidTaskDefinition()

				beforeTask, err = sqlDB.DesireTask(ctx, logger, taskDefinition, taskGuid, taskDomain)
				Expect(err).NotTo(HaveOccurred())
			})

			Context("and the task is pending", func() {
				It("increments the rejection count, stores the rejection reason, and does not mutate the task state", func() {
					before, after, err := sqlDB.RejectTask(ctx, logger, taskGuid, "some failure")
					Expect(err).NotTo(HaveOccurred())
					Expect(before).To(Equal(beforeTask))

					Expect(after.RejectionCount).To(Equal(beforeTask.RejectionCount + 1))
					Expect(after.RejectionReason).To(Equal("some failure"))
					Expect(after.State).To(Equal(models.Task_Pending))

					task, err := sqlDB.TaskByGuid(ctx, logger, taskGuid)
					Expect(err).NotTo(HaveOccurred())

					Expect(task).To(Equal(after))
				})
			})

			Context("and the task is running", func() {
				BeforeEach(func() {
					_, afterStartTask, _, err := sqlDB.StartTask(ctx, logger, taskGuid, "cell-id")
					Expect(err).NotTo(HaveOccurred())
					beforeTask = afterStartTask
				})

				It("increments the rejection count, stores the rejection reason, and resets state to pending", func() {
					before, after, err := sqlDB.RejectTask(ctx, logger, taskGuid, "some failure")
					Expect(err).NotTo(HaveOccurred())
					Expect(before).To(Equal(beforeTask))

					Expect(after.RejectionCount).To(Equal(beforeTask.RejectionCount + 1))
					Expect(after.RejectionReason).To(Equal("some failure"))
					Expect(after.State).To(Equal(models.Task_Pending))

					task, err := sqlDB.TaskByGuid(ctx, logger, taskGuid)
					Expect(err).NotTo(HaveOccurred())

					Expect(task).To(Equal(after))
				})
			})

			Context("and the task is completed", func() {
				BeforeEach(func() {
					_, _, _, err := sqlDB.CancelTask(ctx, logger, taskGuid)
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns a BadRequest error and does not mutate the task", func() {
					_, _, err := sqlDB.RejectTask(ctx, logger, taskGuid, "rejected")
					Expect(err).To(Equal(models.ErrBadRequest))

					task, err := sqlDB.TaskByGuid(ctx, logger, taskGuid)
					Expect(err).NotTo(HaveOccurred())

					Expect(task.RejectionCount).To(BeEquivalentTo(0))
					Expect(task.RejectionReason).To(Equal(""))
				})
			})

			Context("when the rejection reason is longer than 1K", func() {
				var (
					failureReason string
					after         *models.Task
				)

				JustBeforeEach(func() {
					failureReason = strings.Repeat("x", 2*1024)
					var err error
					_, after, err = sqlDB.RejectTask(ctx, logger, taskGuid, failureReason)
					Expect(err).NotTo(HaveOccurred())

				})

				It("truncates the crash reason in the returned task", func() {
					expectedFailureReason := strings.Repeat("x", 1013) + "(truncated)"
					Expect(after.RejectionReason).To(Equal(expectedFailureReason))
				})

				It("truncates the crash reason", func() {
					expectedFailureReason := strings.Repeat("x", 1013) + "(truncated)"
					task, err := sqlDB.TaskByGuid(ctx, logger, taskGuid)
					Expect(err).NotTo(HaveOccurred())
					Expect(task.RejectionReason).To(Equal(expectedFailureReason))
				})
			})

			Context("and the task is resolving", func() {
				BeforeEach(func() {
					_, _, _, err := sqlDB.CancelTask(ctx, logger, taskGuid)
					Expect(err).NotTo(HaveOccurred())
					_, _, err = sqlDB.ResolvingTask(ctx, logger, taskGuid)
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns a BadRequest error and does not mutate the task", func() {
					_, _, err := sqlDB.RejectTask(ctx, logger, taskGuid, "rejected")
					Expect(err).To(Equal(models.ErrBadRequest))

					task, err := sqlDB.TaskByGuid(ctx, logger, taskGuid)
					Expect(err).NotTo(HaveOccurred())

					Expect(task.RejectionCount).To(BeEquivalentTo(0))
					Expect(task.RejectionReason).To(Equal(""))
				})
			})
		})

		Context("when the task does not exist", func() {
			It("returns an ResourceNotFound error", func() {
				_, _, err := sqlDB.RejectTask(ctx, logger, "nota-guid", "")
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})
		})
	})
})

func insertTask(ctx context.Context, db helpers.QueryableDB, serializer format.Serializer, task *models.Task, malformedTaskDefinition bool) {
	taskDefData, err := serializer.Marshal(logger, task.TaskDefinition)
	Expect(err).NotTo(HaveOccurred())

	if malformedTaskDefinition {
		taskDefData = []byte("{{{{{{{{{{")
	}

	queryStr := `INSERT INTO tasks
						  (guid, domain, created_at, updated_at, first_completed_at, state,
							cell_id, result, failed, failure_reason, task_definition)
					    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	if test_helpers.UsePostgres() {
		queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
	}
	result, err := db.ExecContext(
		ctx,
		queryStr,
		task.TaskGuid,
		task.Domain,
		task.CreatedAt,
		task.UpdatedAt,
		task.FirstCompletedAt,
		task.State,
		task.CellId,
		task.Result,
		task.Failed,
		task.FailureReason,
		taskDefData,
	)
	Expect(err).NotTo(HaveOccurred())
	Expect(result.RowsAffected()).NotTo(Equal(1))
}

func updateTaskToInvalid(ctx context.Context, db helpers.QueryableDB, serializer format.Serializer, task *models.Task) {
	_, err := serializer.Marshal(logger, task.TaskDefinition)
	Expect(err).NotTo(HaveOccurred())

	taskDefData := []byte("{{{{{{{{{{")
	queryStr := `UPDATE tasks
						  SET task_definition=?
							WHERE guid=?`
	if test_helpers.UsePostgres() {
		queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
	}
	result, err := db.ExecContext(
		ctx,
		queryStr,
		taskDefData,
		task.TaskGuid,
	)
	Expect(err).NotTo(HaveOccurred())
	Expect(result.RowsAffected()).NotTo(Equal(1))
}
