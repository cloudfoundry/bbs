package controllers_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/auctioneer"
	"code.cloudfoundry.org/auctioneer/auctioneerfakes"
	"code.cloudfoundry.org/bbs/controllers"
	"code.cloudfoundry.org/bbs/db/dbfakes"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/models/test/model_helpers"
	"code.cloudfoundry.org/bbs/taskworkpool/taskworkpoolfakes"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/rep"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Task Controller", func() {
	var (
		logger                   *lagertest.TestLogger
		fakeTaskDB               *dbfakes.FakeTaskDB
		fakeAuctioneerClient     *auctioneerfakes.FakeClient
		fakeTaskCompletionClient *taskworkpoolfakes.FakeTaskCompletionClient

		controller *controllers.TaskController
	)

	BeforeEach(func() {
		fakeTaskDB = new(dbfakes.FakeTaskDB)
		fakeAuctioneerClient = new(auctioneerfakes.FakeClient)
		fakeTaskCompletionClient = new(taskworkpoolfakes.FakeTaskCompletionClient)

		logger = lagertest.NewTestLogger("test")
		controller = controllers.NewTaskController(fakeTaskDB, fakeTaskCompletionClient, fakeAuctioneerClient, fakeServiceClient, fakeRepClientFactory)
	})

	Describe("Tasks", func() {
		var (
			domain, cellId string
			task1          models.Task
			task2          models.Task
			actualTasks    []*models.Task
			err            error
		)

		BeforeEach(func() {
			task1 = models.Task{Domain: "domain-1"}
			task2 = models.Task{CellId: "cell-id"}
			domain = ""
			cellId = ""
		})

		JustBeforeEach(func() {
			actualTasks, err = controller.Tasks(logger, domain, cellId)
		})

		Context("when reading tasks from DB succeeds", func() {
			var tasks []*models.Task

			BeforeEach(func() {
				tasks = []*models.Task{&task1, &task2}
				fakeTaskDB.TasksReturns(tasks, nil)
			})

			It("returns a list of task", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(actualTasks).To(Equal(tasks))
			})

			It("calls the DB with no filter", func() {
				Expect(fakeTaskDB.TasksCallCount()).To(Equal(1))
				_, filter := fakeTaskDB.TasksArgsForCall(0)
				Expect(filter).To(Equal(models.TaskFilter{}))
			})

			Context("and filtering by domain", func() {
				BeforeEach(func() {
					domain = "domain-1"
				})

				It("calls the DB with a domain filter", func() {
					Expect(fakeTaskDB.TasksCallCount()).To(Equal(1))
					_, filter := fakeTaskDB.TasksArgsForCall(0)
					Expect(filter.Domain).To(Equal(domain))
				})
			})

			Context("and filtering by cell id", func() {
				BeforeEach(func() {
					cellId = "cell-id"
				})

				It("calls the DB with a cell filter", func() {
					Expect(fakeTaskDB.TasksCallCount()).To(Equal(1))
					_, filter := fakeTaskDB.TasksArgsForCall(0)
					Expect(filter.CellID).To(Equal(cellId))
				})
			})
		})

		Context("when the DB returns an error", func() {
			BeforeEach(func() {
				fakeTaskDB.TasksReturns(nil, errors.New("kaboom"))
			})

			It("returns the error", func() {
				Expect(err).To(MatchError("kaboom"))
			})
		})
	})

	Describe("TaskByGuid", func() {
		var (
			taskGuid   = "task-guid"
			actualTask *models.Task
			err        error
		)

		JustBeforeEach(func() {
			actualTask, err = controller.TaskByGuid(logger, taskGuid)
		})

		Context("when reading a task from the DB succeeds", func() {
			var task *models.Task

			BeforeEach(func() {
				task = &models.Task{TaskGuid: taskGuid}
				fakeTaskDB.TaskByGuidReturns(task, nil)
			})

			It("fetches task by guid", func() {
				Expect(fakeTaskDB.TaskByGuidCallCount()).To(Equal(1))
				_, actualGuid := fakeTaskDB.TaskByGuidArgsForCall(0)
				Expect(actualGuid).To(Equal(taskGuid))
			})

			It("returns the task", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(actualTask).To(Equal(task))
			})
		})

		Context("when the DB errors out", func() {
			BeforeEach(func() {
				fakeTaskDB.TaskByGuidReturns(nil, errors.New("kaboom"))
			})

			It("provides relevant error information", func() {
				Expect(err).To(MatchError("kaboom"))
			})
		})
	})

	Describe("DesireTask", func() {
		var (
			taskGuid = "task-guid"
			domain   = "domain"
			taskDef  *models.TaskDefinition
			err      error
		)

		BeforeEach(func() {
			taskDef = model_helpers.NewValidTaskDefinition()
		})

		JustBeforeEach(func() {
			err = controller.DesireTask(logger, taskDef, taskGuid, domain)
		})

		Context("when the desire is successful", func() {
			It("desires the task with the requested definitions", func() {
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeTaskDB.DesireTaskCallCount()).To(Equal(1))
				_, actualTaskDef, actualTaskGuid, actualDomain := fakeTaskDB.DesireTaskArgsForCall(0)
				Expect(actualTaskDef).To(Equal(taskDef))
				Expect(actualTaskGuid).To(Equal(taskGuid))
				Expect(actualDomain).To(Equal(domain))
			})

			It("requests an auction", func() {
				Expect(fakeAuctioneerClient.RequestTaskAuctionsCallCount()).To(Equal(1))

				var volumeMounts []string
				for _, volMount := range taskDef.VolumeMounts {
					volumeMounts = append(volumeMounts, volMount.Driver)
				}

				expectedStartRequest := auctioneer.TaskStartRequest{
					Task: rep.Task{
						TaskGuid: taskGuid,
						Domain:   domain,
						Resource: rep.Resource{
							MemoryMB: 256,
							DiskMB:   1024,
						},
						PlacementConstraint: rep.PlacementConstraint{
							RootFs:        "docker:///docker.com/docker",
							VolumeDrivers: volumeMounts,
							PlacementTags: taskDef.PlacementTags,
						},
					},
				}

				requestedTasks := fakeAuctioneerClient.RequestTaskAuctionsArgsForCall(0)
				Expect(requestedTasks).To(HaveLen(1))
				Expect(*requestedTasks[0]).To(Equal(expectedStartRequest))
			})

			Context("when requesting a task auction succeeds", func() {
				BeforeEach(func() {
					fakeAuctioneerClient.RequestTaskAuctionsReturns(nil)
				})

				It("does not return an error", func() {
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when requesting a task auction fails", func() {
				BeforeEach(func() {
					fakeAuctioneerClient.RequestTaskAuctionsReturns(errors.New("oops"))
				})

				It("does not return an error", func() {
					Expect(err).NotTo(HaveOccurred())
				})

				It("does not request a second auction", func() {
					Consistently(fakeAuctioneerClient.RequestTaskAuctionsCallCount).Should(Equal(1))
				})
			})
		})

		Context("when desiring the task fails", func() {
			BeforeEach(func() {
				fakeTaskDB.DesireTaskReturns(errors.New("kaboom"))
			})

			It("responds with an error", func() {
				Expect(err).To(MatchError("kaboom"))
			})
		})
	})

	Describe("StartTask", func() {
		Context("when the start is successful", func() {
			var (
				taskGuid, cellId string
				shouldStart      bool
				err              error
			)

			BeforeEach(func() {
				taskGuid = "task-guid"
				cellId = "cell-id"
			})

			JustBeforeEach(func() {
				shouldStart, err = controller.StartTask(logger, taskGuid, cellId)
			})

			It("calls StartTask", func() {
				Expect(fakeTaskDB.StartTaskCallCount()).To(Equal(1))
				taskLogger, taskGuid, cellId := fakeTaskDB.StartTaskArgsForCall(0)
				Expect(taskLogger.SessionName()).To(ContainSubstring("start-task"))
				Expect(taskGuid).To(Equal(taskGuid))
				Expect(cellId).To(Equal(cellId))
			})

			Context("when the task should start", func() {
				BeforeEach(func() {
					fakeTaskDB.StartTaskReturns(true, nil)
				})

				It("responds with true", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(shouldStart).To(BeTrue())
				})
			})

			Context("when the task should not start", func() {
				BeforeEach(func() {
					fakeTaskDB.StartTaskReturns(false, nil)
				})

				It("responds with false", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(shouldStart).To(BeFalse())
				})
			})

			Context("when the DB fails", func() {
				BeforeEach(func() {
					fakeTaskDB.StartTaskReturns(false, errors.New("kaboom"))
				})

				It("bubbles up the underlying model error", func() {
					Expect(err).To(MatchError("kaboom"))
				})
			})
		})
	})

	Describe("CancelTask", func() {
		var (
			taskGuid, cellID string
			err              error
		)

		BeforeEach(func() {
			taskGuid = "task-guid"
			cellID = "the-cell"
			task := model_helpers.NewValidTask("hi-bob")
			fakeTaskDB.CancelTaskReturns(task, cellID, nil)
		})

		JustBeforeEach(func() {
			err = controller.CancelTask(logger, taskGuid)
		})

		Context("when the cancel request is normal", func() {
			Context("when canceling the task in the db succeeds", func() {
				BeforeEach(func() {
					cellPresence := models.CellPresence{CellId: "cell-id"}
					fakeServiceClient.CellByIdReturns(&cellPresence, nil)
				})

				It("returns no error", func() {
					Expect(fakeTaskDB.CancelTaskCallCount()).To(Equal(1))
					taskLogger, taskGuid := fakeTaskDB.CancelTaskArgsForCall(0)
					Expect(taskLogger.SessionName()).To(ContainSubstring("cancel-task"))
					Expect(taskGuid).To(Equal("task-guid"))
					Expect(err).NotTo(HaveOccurred())
				})

				Context("and the task has a complete URL", func() {
					BeforeEach(func() {
						task := model_helpers.NewValidTask("hi-bob")
						task.CompletionCallbackUrl = "bogus"
						fakeTaskDB.CancelTaskReturns(task, cellID, nil)
					})

					It("causes the workpool to complete its callback work", func() {
						Eventually(fakeTaskCompletionClient.SubmitCallCount).Should(Equal(1))
					})
				})

				Context("but the task has no complete URL", func() {
					BeforeEach(func() {
						task := model_helpers.NewValidTask("hi-bob")
						fakeTaskDB.CancelTaskReturns(task, cellID, nil)
					})

					It("does not complete the task callback", func() {
						Consistently(fakeTaskCompletionClient.SubmitCallCount).Should(Equal(0))
					})
				})

				It("stops the task on the rep", func() {
					Expect(fakeServiceClient.CellByIdCallCount()).To(Equal(1))
					_, actualCellID := fakeServiceClient.CellByIdArgsForCall(0)
					Expect(actualCellID).To(Equal(cellID))

					Expect(fakeRepClient.CancelTaskCallCount()).To(Equal(1))
					guid := fakeRepClient.CancelTaskArgsForCall(0)
					Expect(guid).To(Equal("task-guid"))
				})

				Context("when the task has no cell id", func() {
					BeforeEach(func() {
						task := model_helpers.NewValidTask("hi-bob")
						fakeTaskDB.CancelTaskReturns(task, "", nil)
					})

					It("does not return an error", func() {
						Expect(err).NotTo(HaveOccurred())
					})

					It("does not make any calls to the rep", func() {
						Expect(fakeRepClient.CancelTaskCallCount()).To(Equal(0))
					})

					It("does not make any calls to the service client", func() {
						Expect(fakeServiceClient.CellByIdCallCount()).To(Equal(0))
					})
				})

				Context("when fetching the cell presence fails", func() {
					BeforeEach(func() {
						fakeServiceClient.CellByIdReturns(nil, errors.New("lol"))
					})

					It("does not return an error", func() {
						Expect(err).NotTo(HaveOccurred())
					})

					It("does not make any calls to the rep", func() {
						Expect(fakeRepClient.CancelTaskCallCount()).To(Equal(0))
					})
				})

				Context("when we fail to cancel the task on the rep", func() {
					BeforeEach(func() {
						fakeRepClient.CancelTaskReturns(errors.New("lol"))
					})

					It("does not return an error", func() {
						Expect(err).NotTo(HaveOccurred())
					})
				})
			})

			Context("when cancelling the task fails", func() {
				BeforeEach(func() {
					fakeTaskDB.CancelTaskReturns(nil, "", errors.New("kaboom"))
				})

				It("responds with an error", func() {
					Expect(err).To(MatchError("kaboom"))
				})
			})
		})
	})

	Describe("FailTask", func() {
		var (
			taskGuid      string
			failureReason string
			err           error
		)

		BeforeEach(func() {
			taskGuid = "task-guid"
			failureReason = "just cuz ;)"
			task := model_helpers.NewValidTask("hi-bob")
			fakeTaskDB.FailTaskReturns(task, nil)
		})

		JustBeforeEach(func() {
			err = controller.FailTask(logger, taskGuid, failureReason)
		})

		Context("when failing the task succeeds", func() {
			It("returns no error", func() {
				_, actualTaskGuid, actualFailureReason := fakeTaskDB.FailTaskArgsForCall(0)
				Expect(actualTaskGuid).To(Equal(taskGuid))
				Expect(actualFailureReason).To(Equal(failureReason))
				Expect(err).NotTo(HaveOccurred())
			})

			Context("and the task has a complete URL", func() {
				BeforeEach(func() {
					task := model_helpers.NewValidTask("hi-bob")
					task.CompletionCallbackUrl = "bogus"
					fakeTaskDB.FailTaskReturns(task, nil)
				})

				It("causes the workpool to complete its callback work", func() {
					Eventually(fakeTaskCompletionClient.SubmitCallCount).Should(Equal(1))
				})
			})

			Context("but the task has no complete URL", func() {
				BeforeEach(func() {
					task := model_helpers.NewValidTask("hi-bob")
					fakeTaskDB.FailTaskReturns(task, nil)
				})

				It("does not complete the task callback", func() {
					Consistently(fakeTaskCompletionClient.SubmitCallCount).Should(Equal(0))
				})
			})
		})

		Context("when failing the task fails", func() {
			BeforeEach(func() {
				fakeTaskDB.FailTaskReturns(nil, errors.New("kaboom"))
			})

			It("responds with an error", func() {
				Expect(err).To(MatchError("kaboom"))
			})
		})
	})

	Describe("CompleteTask", func() {
		var (
			taskGuid      string
			cellId        string
			failed        bool
			failureReason string
			result        string
			err           error
		)

		BeforeEach(func() {
			taskGuid = "t-guid"
			cellId = "c-id"
			failed = true
			failureReason = "some-error"
			result = "yeah"

			task := model_helpers.NewValidTask("hi-bob")
			fakeTaskDB.CompleteTaskReturns(task, nil)
		})

		JustBeforeEach(func() {
			err = controller.CompleteTask(logger, taskGuid, cellId, failed, failureReason, result)
		})

		Context("when completing the task succeeds", func() {
			It("returns no error", func() {
				Expect(fakeTaskDB.CompleteTaskCallCount()).To(Equal(1))
				_, actualTaskGuid, actualCellId, actualFailed, actualFailureReason, actualResult := fakeTaskDB.CompleteTaskArgsForCall(0)
				Expect(actualTaskGuid).To(Equal(taskGuid))
				Expect(actualCellId).To(Equal(cellId))
				Expect(actualFailed).To(Equal(failed))
				Expect(actualFailureReason).To(Equal(failureReason))
				Expect(actualResult).To(Equal(result))
				Expect(err).NotTo(HaveOccurred())
			})

			Context("and completing succeeds", func() {
				Context("and the task has a complete URL", func() {
					BeforeEach(func() {
						task := model_helpers.NewValidTask("hi-bob")
						task.CompletionCallbackUrl = "bogus"
						fakeTaskDB.CompleteTaskReturns(task, nil)
					})

					It("causes the workpool to complete its callback work", func() {
						Eventually(fakeTaskCompletionClient.SubmitCallCount).Should(Equal(1))
					})
				})

				Context("but the task has no complete URL", func() {
					BeforeEach(func() {
						task := model_helpers.NewValidTask("hi-bob")
						fakeTaskDB.CompleteTaskReturns(task, nil)
					})

					It("does not complete the task callback", func() {
						Consistently(fakeTaskCompletionClient.SubmitCallCount).Should(Equal(0))
					})
				})
			})
		})

		Context("when completing the task fails", func() {
			BeforeEach(func() {
				fakeTaskDB.CompleteTaskReturns(nil, errors.New("kaboom"))
			})

			It("responds with an error", func() {
				Expect(err).To(MatchError("kaboom"))
			})
		})
	})

	Describe("ResolvingTask", func() {
		Context("when the resolving request is normal", func() {
			var (
				taskGuid string
				err      error
			)

			BeforeEach(func() {
				taskGuid = "task-guid"
			})

			JustBeforeEach(func() {
				err = controller.ResolvingTask(logger, taskGuid)
			})

			Context("when resolvinging the task succeeds", func() {
				It("returns no error", func() {
					Expect(fakeTaskDB.ResolvingTaskCallCount()).To(Equal(1))
					_, taskGuid := fakeTaskDB.ResolvingTaskArgsForCall(0)
					Expect(taskGuid).To(Equal("task-guid"))
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when desiring the task fails", func() {
				BeforeEach(func() {
					fakeTaskDB.ResolvingTaskReturns(errors.New("kaboom"))
				})

				It("responds with an error", func() {
					Expect(err).To(MatchError("kaboom"))
				})
			})
		})
	})

	Describe("DeleteTask", func() {
		Context("when the delete request is normal", func() {
			var (
				taskGuid string
				err      error
			)

			BeforeEach(func() {
				taskGuid = "task-guid"
			})

			JustBeforeEach(func() {
				err = controller.DeleteTask(logger, taskGuid)
			})

			Context("when deleting the task succeeds", func() {
				It("returns no error", func() {
					Expect(fakeTaskDB.DeleteTaskCallCount()).To(Equal(1))
					_, taskGuid := fakeTaskDB.DeleteTaskArgsForCall(0)
					Expect(taskGuid).To(Equal("task-guid"))
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when desiring the task fails", func() {
				BeforeEach(func() {
					fakeTaskDB.DeleteTaskReturns(errors.New("kaboom"))
				})

				It("responds with an error", func() {
					Expect(err).To(MatchError("kaboom"))
				})
			})
		})
	})

	Describe("ConvergeTasks", func() {
		Context("when the request is normal", func() {
			var (
				kickTaskDuration            = 10 * time.Second
				expirePendingTaskDuration   = 10 * time.Second
				expireCompletedTaskDuration = 10 * time.Second
				cellSet                     models.CellSet
				err                         error
			)

			BeforeEach(func() {
				cellPresence := models.NewCellPresence("cell-id", "1.1.1.1", "z1", models.CellCapacity{}, nil, nil, nil, nil)
				cellSet = models.CellSet{"cell-id": &cellPresence}
				fakeServiceClient.CellsReturns(cellSet, nil)
			})

			JustBeforeEach(func() {
				err = controller.ConvergeTasks(logger, kickTaskDuration, expirePendingTaskDuration, expireCompletedTaskDuration)
			})

			It("calls ConvergeTasks", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeTaskDB.ConvergeTasksCallCount()).To(Equal(1))
				taskLogger, actualCellSet, actualKickDuration, actualPendingDuration, actualCompletedDuration := fakeTaskDB.ConvergeTasksArgsForCall(0)
				Expect(taskLogger.SessionName()).To(ContainSubstring("converge-tasks"))
				Expect(actualCellSet).To(BeEquivalentTo(cellSet))
				Expect(actualKickDuration).To(BeEquivalentTo(kickTaskDuration))
				Expect(actualPendingDuration).To(BeEquivalentTo(expirePendingTaskDuration))
				Expect(actualCompletedDuration).To(BeEquivalentTo(expireCompletedTaskDuration))
			})

			Context("when fetching cells fails", func() {
				BeforeEach(func() {
					fakeServiceClient.CellsReturns(nil, errors.New("kaboom"))
				})

				It("does not call ConvergeTasks", func() {
					Expect(err).To(MatchError("kaboom"))
					Expect(fakeTaskDB.ConvergeTasksCallCount()).To(Equal(0))
				})
			})

			Context("when fetching cells returns ErrResourceNotFound", func() {
				BeforeEach(func() {
					fakeServiceClient.CellsReturns(nil, models.ErrResourceNotFound)
				})

				It("calls ConvergeTasks with an empty CellSet", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeTaskDB.ConvergeTasksCallCount()).To(Equal(1))
					_, actualCellSet, _, _, _ := fakeTaskDB.ConvergeTasksArgsForCall(0)
					Expect(actualCellSet).To(BeEquivalentTo(models.CellSet{}))
				})
			})

			Context("when there are tasks to complete", func() {
				const taskGuid1 = "to-complete-1"
				const taskGuid2 = "to-complete-2"

				BeforeEach(func() {
					task1 := model_helpers.NewValidTask(taskGuid1)
					task2 := model_helpers.NewValidTask(taskGuid2)
					fakeTaskDB.ConvergeTasksReturns(nil, []*models.Task{task1, task2})
				})

				It("submits the tasks to the workpool", func() {
					expectedCallCount := 2
					Expect(fakeTaskCompletionClient.SubmitCallCount()).To(Equal(expectedCallCount))

					_, submittedTask1 := fakeTaskCompletionClient.SubmitArgsForCall(0)
					_, submittedTask2 := fakeTaskCompletionClient.SubmitArgsForCall(1)
					Expect([]string{submittedTask1.TaskGuid, submittedTask2.TaskGuid}).To(ConsistOf(taskGuid1, taskGuid2))

					task1Completions := 0
					task2Completions := 0
					for i := 0; i < expectedCallCount; i++ {
						db, task := fakeTaskCompletionClient.SubmitArgsForCall(i)
						Expect(db).To(Equal(fakeTaskDB))
						if task.TaskGuid == taskGuid1 {
							task1Completions++
						} else if task.TaskGuid == taskGuid2 {
							task2Completions++
						}
					}

					Expect(task1Completions).To(Equal(1))
					Expect(task2Completions).To(Equal(1))
				})
			})

			Context("when there are tasks to auction", func() {
				const taskGuid1 = "to-auction-1"
				const taskGuid2 = "to-auction-2"

				BeforeEach(func() {
					taskStartRequest1 := auctioneer.NewTaskStartRequestFromModel(taskGuid1, "domain", model_helpers.NewValidTaskDefinition())
					taskStartRequest2 := auctioneer.NewTaskStartRequestFromModel(taskGuid2, "domain", model_helpers.NewValidTaskDefinition())
					fakeTaskDB.ConvergeTasksReturns([]*auctioneer.TaskStartRequest{&taskStartRequest1, &taskStartRequest2}, nil)
				})

				It("requests an auction", func() {
					Expect(fakeAuctioneerClient.RequestTaskAuctionsCallCount()).To(Equal(1))

					requestedTasks := fakeAuctioneerClient.RequestTaskAuctionsArgsForCall(0)
					Expect(requestedTasks).To(HaveLen(2))
					Expect([]string{requestedTasks[0].TaskGuid, requestedTasks[1].TaskGuid}).To(ConsistOf(taskGuid1, taskGuid2))
				})

				Context("when requesting an auction is unsuccessful", func() {
					BeforeEach(func() {
						fakeAuctioneerClient.RequestTaskAuctionsReturns(errors.New("oops"))
					})

					It("logs an error", func() {
						Expect(logger.TestSink.LogMessages()).To(ContainElement("test.converge-tasks.failed-to-request-auctions-for-pending-tasks"))
					})
				})
			})
		})
	})
})
