package controllers_test

import (
	"context"
	"errors"
	"time"

	"code.cloudfoundry.org/auctioneer"
	"code.cloudfoundry.org/auctioneer/auctioneerfakes"
	"code.cloudfoundry.org/bbs/controllers"
	"code.cloudfoundry.org/bbs/db"
	"code.cloudfoundry.org/bbs/db/dbfakes"
	"code.cloudfoundry.org/bbs/events/eventfakes"
	"code.cloudfoundry.org/bbs/metrics/fakes"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/models/test/model_helpers"
	"code.cloudfoundry.org/bbs/taskworkpool/taskworkpoolfakes"
	"code.cloudfoundry.org/bbs/trace"
	"code.cloudfoundry.org/lager/v3/lagertest"
	"code.cloudfoundry.org/rep"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Task Controller", func() {
	var (
		logger                   *lagertest.TestLogger
		fakeTaskDB               *dbfakes.FakeTaskDB
		fakeAuctioneerClient     *auctioneerfakes.FakeClient
		fakeTaskCompletionClient *taskworkpoolfakes.FakeTaskCompletionClient
		taskHub                  *eventfakes.FakeHub
		maxPlacementRetries      int

		controller           *controllers.TaskController
		fakeTaskStatNotifier *fakes.FakeTaskStatMetronNotifier
		err                  error
	)

	BeforeEach(func() {
		fakeTaskDB = new(dbfakes.FakeTaskDB)
		fakeAuctioneerClient = new(auctioneerfakes.FakeClient)
		fakeTaskCompletionClient = new(taskworkpoolfakes.FakeTaskCompletionClient)
		fakeTaskStatNotifier = &fakes.FakeTaskStatMetronNotifier{}

		logger = lagertest.NewTestLogger("test")
		err = nil

		taskHub = &eventfakes.FakeHub{}
		maxPlacementRetries = 0
	})

	JustBeforeEach(func() {
		controller = controllers.NewTaskController(
			fakeTaskDB,
			fakeTaskCompletionClient,
			fakeAuctioneerClient,
			fakeServiceClient,
			fakeRepClientFactory,
			taskHub,
			fakeTaskStatNotifier,
			maxPlacementRetries,
		)
	})

	Describe("Tasks", func() {
		var (
			domain, cellId string
			task1          models.Task
			task2          models.Task
			actualTasks    []*models.Task
		)

		BeforeEach(func() {
			task1 = models.Task{Domain: "domain-1"}
			task2 = models.Task{CellId: "cell-id"}
			domain = ""
			cellId = ""
		})

		JustBeforeEach(func() {
			actualTasks, err = controller.Tasks(ctx, logger, domain, cellId)
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
				_, _, filter := fakeTaskDB.TasksArgsForCall(0)
				Expect(filter).To(Equal(models.TaskFilter{}))
			})

			Context("and filtering by domain", func() {
				BeforeEach(func() {
					domain = "domain-1"
				})

				It("calls the DB with a domain filter", func() {
					Expect(fakeTaskDB.TasksCallCount()).To(Equal(1))
					_, _, filter := fakeTaskDB.TasksArgsForCall(0)
					Expect(filter.Domain).To(Equal(domain))
				})
			})

			Context("and filtering by cell id", func() {
				BeforeEach(func() {
					cellId = "cell-id"
				})

				It("calls the DB with a cell filter", func() {
					Expect(fakeTaskDB.TasksCallCount()).To(Equal(1))
					_, _, filter := fakeTaskDB.TasksArgsForCall(0)
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
		)

		JustBeforeEach(func() {
			actualTask, err = controller.TaskByGuid(ctx, logger, taskGuid)
		})

		Context("when reading a task from the DB succeeds", func() {
			var task *models.Task

			BeforeEach(func() {
				task = &models.Task{TaskGuid: taskGuid}
				fakeTaskDB.TaskByGuidReturns(task, nil)
			})

			It("fetches task by guid", func() {
				Expect(fakeTaskDB.TaskByGuidCallCount()).To(Equal(1))
				_, _, actualGuid := fakeTaskDB.TaskByGuidArgsForCall(0)
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
		)

		BeforeEach(func() {
			taskDef = model_helpers.NewValidTaskDefinition()
		})

		JustBeforeEach(func() {
			err = controller.DesireTask(ctx, logger, taskDef, taskGuid, domain)
		})

		Context("when the desire is successful", func() {
			BeforeEach(func() {
				fakeTaskDB.DesireTaskReturns(&models.Task{TaskGuid: taskGuid}, err)
			})

			It("desires the task with the requested definitions", func() {
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeTaskDB.DesireTaskCallCount()).To(Equal(1))
				_, _, actualTaskDef, actualTaskGuid, actualDomain := fakeTaskDB.DesireTaskArgsForCall(0)
				Expect(actualTaskDef).To(Equal(taskDef))
				Expect(actualTaskGuid).To(Equal(taskGuid))
				Expect(actualDomain).To(Equal(domain))
			})

			It("requests an auction", func() {
				Eventually(fakeAuctioneerClient.RequestTaskAuctionsCallCount).Should(Equal(1))

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
							MaxPids:  1024,
						},
						PlacementConstraint: rep.PlacementConstraint{
							RootFs:        "docker:///docker.com/docker",
							VolumeDrivers: volumeMounts,
							PlacementTags: taskDef.PlacementTags,
						},
					},
				}

				_, traceId, requestedTasks := fakeAuctioneerClient.RequestTaskAuctionsArgsForCall(0)
				Expect(traceId).NotTo(BeNil())
				Expect(requestedTasks).To(HaveLen(1))
				Expect(*requestedTasks[0]).To(Equal(expectedStartRequest))
			})

			It("emits a TaskCreateEvent to the hub", func() {
				Eventually(taskHub.EmitCallCount).Should(Equal(1))
				event := taskHub.EmitArgsForCall(0)
				create, ok := event.(*models.TaskCreatedEvent)
				Expect(ok).To(BeTrue())
				Expect(create.Key()).To(Equal(taskGuid))
			})

			Context("when requesting a task auction succeeds", func() {
				BeforeEach(func() {
					fakeAuctioneerClient.RequestTaskAuctionsReturns(nil)
				})

				It("does not return an error", func() {
					Expect(err).NotTo(HaveOccurred())
				})

				It("does not emit a TaskChangedEvent", func() {
					Eventually(taskHub.EmitCallCount).Should(Equal(1))
					Consistently(taskHub.EmitCallCount).Should(Equal(1))
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

				It("does not emit a TaskChangedEvent", func() {
					Eventually(taskHub.EmitCallCount).Should(Equal(1))
					Consistently(taskHub.EmitCallCount).Should(Equal(1))
				})
			})
		})

		Context("when desiring the task fails", func() {
			BeforeEach(func() {
				fakeTaskDB.DesireTaskReturns(nil, errors.New("kaboom"))
			})

			It("responds with an error", func() {
				Expect(err).To(MatchError("kaboom"))
			})

			It("does not emit a TaskChangedEvent", func() {
				Consistently(taskHub.EmitCallCount).Should(Equal(0))
			})
		})
	})

	Describe("StartTask", func() {
		Context("when the start is successful", func() {
			var (
				taskGuid, cellId string
				shouldStart      bool
			)

			BeforeEach(func() {
				taskGuid = "task-guid"
				cellId = "cell-id"
			})

			JustBeforeEach(func() {
				shouldStart, err = controller.StartTask(ctx, logger, taskGuid, cellId)
			})

			It("calls StartTask", func() {
				Expect(fakeTaskDB.StartTaskCallCount()).To(Equal(1))
				taskContext, taskLogger, taskGuid, cellId := fakeTaskDB.StartTaskArgsForCall(0)
				Expect(taskContext).To(Equal(ctx))
				Expect(taskLogger.SessionName()).To(ContainSubstring("start-task"))
				Expect(taskGuid).To(Equal(taskGuid))
				Expect(cellId).To(Equal(cellId))
			})

			Context("when the task should start", func() {
				var before, after *models.Task
				BeforeEach(func() {

					before = &models.Task{State: models.Task_Pending}
					after = &models.Task{State: models.Task_Running}
					fakeTaskDB.StartTaskReturns(before, after, true, nil)
				})

				It("updates the task stats", func() {
					Expect(fakeTaskStatNotifier.RecordTaskStartedCallCount()).To(Equal(1))
					actualCellId := fakeTaskStatNotifier.RecordTaskStartedArgsForCall(0)
					Expect(actualCellId).To(Equal(cellId))
				})

				It("responds with true", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(shouldStart).To(BeTrue())
				})

				It("emits a change to the hub", func() {
					Eventually(taskHub.EmitCallCount).Should(Equal(1))
					event := taskHub.EmitArgsForCall(0)
					changedEvent, ok := event.(*models.TaskChangedEvent)
					Expect(ok).To(BeTrue())
					Expect(changedEvent.Before).To(Equal(before))
					Expect(changedEvent.After).To(Equal(after))
				})
			})

			Context("when the task should not start", func() {
				BeforeEach(func() {
					fakeTaskDB.StartTaskReturns(nil, nil, false, nil)
				})

				It("responds with false", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(shouldStart).To(BeFalse())
				})

				It("does not update the task stats", func() {
					Expect(fakeTaskStatNotifier.RecordTaskStartedCallCount()).To(BeZero())
				})

				It("does not emit a change to the hub", func() {
					Consistently(taskHub.EmitCallCount).Should(Equal(0))
				})
			})

			Context("when the DB fails", func() {
				BeforeEach(func() {
					fakeTaskDB.StartTaskReturns(nil, nil, false, errors.New("kaboom"))
				})

				It("does not update the task stats", func() {
					Expect(fakeTaskStatNotifier.RecordTaskStartedCallCount()).To(BeZero())
				})

				It("bubbles up the underlying model error", func() {
					Expect(err).To(MatchError("kaboom"))
				})

				It("does not emit a change to the hub", func() {
					Consistently(taskHub.EmitCallCount).Should(Equal(0))
				})
			})
		})
	})

	Describe("CancelTask", func() {
		var (
			taskGuid, cellID string
			before, after    *models.Task
			ctxWithTrace     context.Context
		)

		BeforeEach(func() {
			taskGuid = "task-guid"
			cellID = "the-cell"
			after = model_helpers.NewValidTask("hi-bob")
			fakeTaskDB.CancelTaskReturns(before, after, cellID, nil)
			ctxWithTrace = context.WithValue(ctx, trace.RequestIdHeader, "some-trace-id")
		})

		JustBeforeEach(func() {
			err = controller.CancelTask(ctxWithTrace, logger, taskGuid)
		})

		Context("when the cancel request is normal", func() {
			Context("when canceling the task in the db succeeds", func() {
				BeforeEach(func() {
					cellPresence := models.CellPresence{CellId: "cell-id"}
					fakeServiceClient.CellByIdReturns(&cellPresence, nil)
				})

				It("returns no error", func() {
					Expect(fakeTaskDB.CancelTaskCallCount()).To(Equal(1))
					taskContext, taskLogger, taskGuid := fakeTaskDB.CancelTaskArgsForCall(0)
					Expect(taskContext).To(Equal(ctxWithTrace))
					Expect(taskLogger.SessionName()).To(ContainSubstring("cancel-task"))
					Expect(taskGuid).To(Equal("task-guid"))
					Expect(err).NotTo(HaveOccurred())
				})

				It("emits a change to the hub", func() {
					Eventually(taskHub.EmitCallCount).Should(Equal(1))
					event := taskHub.EmitArgsForCall(0)
					changedEvent, ok := event.(*models.TaskChangedEvent)
					Expect(ok).To(BeTrue())
					Expect(changedEvent.Before).To(Equal(before))
					Expect(changedEvent.After).To(Equal(after))
				})

				Context("and the task has a complete URL", func() {
					BeforeEach(func() {
						task := model_helpers.NewValidTask("hi-bob")
						task.CompletionCallbackUrl = "bogus"
						fakeTaskDB.CancelTaskReturns(nil, task, cellID, nil)
					})

					It("causes the workpool to complete its callback work", func() {
						Eventually(fakeTaskCompletionClient.SubmitCallCount).Should(Equal(1))
					})

				})

				Context("but the task has no complete URL", func() {
					BeforeEach(func() {
						after = model_helpers.NewValidTask("hi-bob")
						fakeTaskDB.CancelTaskReturns(nil, after, cellID, nil)
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
					_, guid := fakeRepClient.CancelTaskArgsForCall(0)
					Expect(guid).To(Equal("task-guid"))
				})

				Context("when the rep announces a url", func() {
					BeforeEach(func() {
						cellPresence := models.CellPresence{CellId: "cell-id", RepAddress: "some-address", RepUrl: "http://some-address"}
						fakeServiceClient.CellByIdReturns(&cellPresence, nil)
					})

					It("creates a rep client using the rep url", func() {
						repAddr, repURL, traceID := fakeRepClientFactory.CreateClientArgsForCall(0)
						Expect(repAddr).To(Equal("some-address"))
						Expect(repURL).To(Equal("http://some-address"))
						Expect(traceID).To(Equal("some-trace-id"))
					})

					Context("when creating a rep client fails", func() {
						BeforeEach(func() {
							err := errors.New("BOOM!!!")
							fakeRepClientFactory.CreateClientReturns(nil, err)
						})

						It("should log the error", func() {
							Expect(logger.Buffer()).To(gbytes.Say("BOOM!!!"))
						})

						It("should return the error", func() {
							Expect(err).To(MatchError("BOOM!!!"))
						})
					})
				})

				Context("when the task has no cell id", func() {
					BeforeEach(func() {
						task := model_helpers.NewValidTask("hi-bob")
						fakeTaskDB.CancelTaskReturns(nil, task, "", nil)
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
					fakeTaskDB.CancelTaskReturns(nil, nil, "", errors.New("kaboom"))
				})

				It("responds with an error", func() {
					Expect(err).To(MatchError("kaboom"))
				})

				It("does not emit a change to the hub", func() {
					Consistently(taskHub.EmitCallCount).Should(Equal(0))
				})
			})
		})
	})

	AssertTaskFailing := func(taskGuid, failureReason string) {
		Context("when failing the task succeeds", func() {
			It("returns no error", func() {
				_, _, actualTaskGuid, actualFailureReason := fakeTaskDB.FailTaskArgsForCall(0)
				Expect(actualTaskGuid).To(Equal(taskGuid))
				Expect(actualFailureReason).To(Equal(failureReason))
				Expect(err).NotTo(HaveOccurred())
			})

			It("emits a change to the hub", func() {
				Eventually(taskHub.EmitCallCount).Should(Equal(1))
				event := taskHub.EmitArgsForCall(0)
				changedEvent, ok := event.(*models.TaskChangedEvent)
				Expect(ok).To(BeTrue())
				Expect(changedEvent.Before).To(Equal(&models.Task{}))
				Expect(changedEvent.After).To(Equal(model_helpers.NewValidTask("hi-bob")))
			})

			Context("and the task has a complete URL", func() {
				BeforeEach(func() {
					task := model_helpers.NewValidTask("hi-bob")
					task.CompletionCallbackUrl = "bogus"
					fakeTaskDB.FailTaskReturns(nil, task, nil)
				})

				It("causes the workpool to complete its callback work", func() {
					Eventually(fakeTaskCompletionClient.SubmitCallCount).Should(Equal(1))
				})
			})

			Context("but the task has no complete URL", func() {
				BeforeEach(func() {
					task := model_helpers.NewValidTask("hi-bob")
					fakeTaskDB.FailTaskReturns(nil, task, nil)
				})

				It("does not complete the task callback", func() {
					Consistently(fakeTaskCompletionClient.SubmitCallCount).Should(Equal(0))
				})
			})
		})

		Context("when failing the task fails", func() {
			BeforeEach(func() {
				fakeTaskDB.FailTaskReturns(nil, nil, errors.New("kaboom"))
			})

			It("responds with an error", func() {
				Expect(err).To(MatchError("kaboom"))
			})

			It("does not emit a change to the hub", func() {
				Consistently(taskHub.EmitCallCount).Should(Equal(0))
			})
		})
	}

	Describe("FailTask", func() {
		var (
			taskGuid      = "task-guid"
			failureReason = "just cuz ;)"
		)

		BeforeEach(func() {
			before := &models.Task{}
			after := model_helpers.NewValidTask("hi-bob")
			fakeTaskDB.FailTaskReturns(before, after, nil)
		})

		JustBeforeEach(func() {
			err = controller.FailTask(ctx, logger, taskGuid, failureReason)
		})

		AssertTaskFailing(taskGuid, failureReason)
	})

	Describe("RejectTask", func() {
		var (
			taskGuid        = "task-guid"
			rejectionReason = "rejection"
			before, after   *models.Task
		)

		BeforeEach(func() {
			before = &models.Task{}
			after = model_helpers.NewValidTask("hi-bob")
			fakeTaskDB.FailTaskReturns(before, after, nil)
			fakeTaskDB.RejectTaskReturns(before, after, nil)
			fakeTaskDB.TaskByGuidReturns(before, nil)
		})

		JustBeforeEach(func() {
			err = controller.RejectTask(ctx, logger, taskGuid, rejectionReason)
		})

		Context("when fetching the task returns an error", func() {
			BeforeEach(func() {
				fakeTaskDB.TaskByGuidReturns(nil, errors.New("some db error"))
			})

			It("forwards the error", func() {
				Expect(err).To(MatchError(errors.New("some db error")))
			})
		})

		Context("when max_task_retries is 0", func() {
			BeforeEach(func() {
				maxPlacementRetries = 0
			})
			AssertTaskFailing(taskGuid, rejectionReason)

			It("rejects the task", func() {
				Expect(fakeTaskDB.RejectTaskCallCount()).To(Equal(1))
				_, _, actualTaskGuid, actualRejectionReason := fakeTaskDB.RejectTaskArgsForCall(0)
				Expect(actualTaskGuid).To(Equal(taskGuid))
				Expect(actualRejectionReason).To(Equal(rejectionReason))
			})

			It("logs the rejection reason", func() {
				Eventually(logger.Buffer()).Should(gbytes.Say(rejectionReason))
			})
		})

		Context("when max_task_retries is 1", func() {
			BeforeEach(func() {
				maxPlacementRetries = 1
			})

			Context("when the task has a rejection count of 0", func() {
				It("rejects the task", func() {
					Expect(fakeTaskDB.RejectTaskCallCount()).To(Equal(1))
					_, _, actualTaskGuid, actualRejectionReason := fakeTaskDB.RejectTaskArgsForCall(0)
					Expect(actualTaskGuid).To(Equal(taskGuid))
					Expect(actualRejectionReason).To(Equal(rejectionReason))
				})

				It("emits a change to the hub", func() {
					Eventually(taskHub.EmitCallCount).Should(Equal(1))
					event := taskHub.EmitArgsForCall(0)
					changedEvent, ok := event.(*models.TaskChangedEvent)
					Expect(ok).To(BeTrue())
					Expect(changedEvent.Before).To(Equal(before))
					Expect(changedEvent.After).To(Equal(after))
				})

				It("logs the rejection reason", func() {
					Eventually(logger.Buffer()).Should(gbytes.Say(rejectionReason))
				})
			})

			Context("when the task has a rejection count of 1", func() {
				BeforeEach(func() {
					fakeTaskDB.TaskByGuidReturns(&models.Task{RejectionCount: 1}, nil)
				})
				AssertTaskFailing(taskGuid, rejectionReason)

				It("rejects the task", func() {
					Expect(fakeTaskDB.RejectTaskCallCount()).To(Equal(1))
					_, _, actualTaskGuid, actualRejectionReason := fakeTaskDB.RejectTaskArgsForCall(0)
					Expect(actualTaskGuid).To(Equal(taskGuid))
					Expect(actualRejectionReason).To(Equal(rejectionReason))
				})

				It("logs the rejection reason", func() {
					Eventually(logger.Buffer()).Should(gbytes.Say(rejectionReason))
				})

				Context("when RejectTask returns an error", func() {
					BeforeEach(func() {
						fakeTaskDB.RejectTaskReturns(nil, nil, errors.New("o noes!"))
					})
					AssertTaskFailing(taskGuid, rejectionReason)

				})
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
			before, after *models.Task
		)

		BeforeEach(func() {
			taskGuid = "t-guid"
			cellId = "c-id"
			failed = true
			failureReason = "some-error"
			result = "yeah"

			before = &models.Task{}
			after = model_helpers.NewValidTask("hi-bob")
			fakeTaskDB.CompleteTaskReturns(before, after, nil)
		})

		JustBeforeEach(func() {
			err = controller.CompleteTask(ctx, logger, taskGuid, cellId, failed, failureReason, result)
		})

		Context("when the task is not marked failed", func() {
			BeforeEach(func() {
				failed = false
			})

			It("updates the task stats", func() {
				Expect(fakeTaskStatNotifier.RecordTaskSucceededCallCount()).To(Equal(1))
				actualCellId := fakeTaskStatNotifier.RecordTaskSucceededArgsForCall(0)
				Expect(actualCellId).To(Equal(cellId))
			})
		})

		Context("when completing the task succeeds", func() {
			It("returns no error", func() {
				Expect(fakeTaskDB.CompleteTaskCallCount()).To(Equal(1))
				_, _, actualTaskGuid, actualCellId, actualFailed, actualFailureReason, actualResult := fakeTaskDB.CompleteTaskArgsForCall(0)
				Expect(actualTaskGuid).To(Equal(taskGuid))
				Expect(actualCellId).To(Equal(cellId))
				Expect(actualFailed).To(Equal(failed))
				Expect(actualFailureReason).To(Equal(failureReason))
				Expect(actualResult).To(Equal(result))
				Expect(err).NotTo(HaveOccurred())
			})

			It("emits a change to the hub", func() {
				Eventually(taskHub.EmitCallCount).Should(Equal(1))
				event := taskHub.EmitArgsForCall(0)
				changedEvent, ok := event.(*models.TaskChangedEvent)
				Expect(ok).To(BeTrue())
				Expect(changedEvent.Before).To(Equal(before))
				Expect(changedEvent.After).To(Equal(after))
			})

			It("updates the task stats", func() {
				Expect(fakeTaskStatNotifier.RecordTaskFailedCallCount()).To(Equal(1))
				actualCellId := fakeTaskStatNotifier.RecordTaskFailedArgsForCall(0)
				Expect(actualCellId).To(Equal(cellId))
			})

			Context("and completing succeeds", func() {
				Context("and the task has a complete URL", func() {
					BeforeEach(func() {
						task := model_helpers.NewValidTask("hi-bob")
						task.CompletionCallbackUrl = "bogus"
						fakeTaskDB.CompleteTaskReturns(nil, task, nil)
					})

					It("causes the workpool to complete its callback work", func() {
						Eventually(fakeTaskCompletionClient.SubmitCallCount).Should(Equal(1))
					})
				})

				Context("but the task has no complete URL", func() {
					BeforeEach(func() {
						task := model_helpers.NewValidTask("hi-bob")
						fakeTaskDB.CompleteTaskReturns(nil, task, nil)
					})

					It("does not complete the task callback", func() {
						Consistently(fakeTaskCompletionClient.SubmitCallCount).Should(Equal(0))
					})
				})
			})
		})

		Context("when completing the task fails", func() {
			BeforeEach(func() {
				fakeTaskDB.CompleteTaskReturns(nil, nil, errors.New("kaboom"))
			})

			It("responds with an error", func() {
				Expect(err).To(MatchError("kaboom"))
			})

			It("does not update the task stats", func() {
				Expect(fakeTaskStatNotifier.RecordTaskSucceededCallCount()).To(BeZero())
			})

			It("does not emit a change to the hub", func() {
				Consistently(taskHub.EmitCallCount).Should(Equal(0))
			})
		})
	})

	Describe("ResolvingTask", func() {
		Context("when the resolving request is normal", func() {
			var (
				taskGuid      string
				before, after *models.Task
			)

			BeforeEach(func() {
				taskGuid = "task-guid"
				before = &models.Task{}
				after = &models.Task{State: models.Task_Resolving}
				fakeTaskDB.ResolvingTaskReturns(before, after, nil)
			})

			JustBeforeEach(func() {
				err = controller.ResolvingTask(ctx, logger, taskGuid)
			})

			Context("when resolvinging the task succeeds", func() {
				It("returns no error", func() {
					Expect(fakeTaskDB.ResolvingTaskCallCount()).To(Equal(1))
					_, _, taskGuid := fakeTaskDB.ResolvingTaskArgsForCall(0)
					Expect(taskGuid).To(Equal("task-guid"))
					Expect(err).NotTo(HaveOccurred())
				})

				It("emits a change to the hub", func() {
					Eventually(taskHub.EmitCallCount).Should(Equal(1))
					event := taskHub.EmitArgsForCall(0)
					changedEvent, ok := event.(*models.TaskChangedEvent)
					Expect(ok).To(BeTrue())
					Expect(changedEvent.Before).To(Equal(before))
					Expect(changedEvent.After).To(Equal(after))
				})
			})

			Context("when desiring the task fails", func() {
				BeforeEach(func() {
					fakeTaskDB.ResolvingTaskReturns(nil, nil, errors.New("kaboom"))
				})

				It("responds with an error", func() {
					Expect(err).To(MatchError("kaboom"))
				})

				It("does not emit a change to the hub", func() {
					Consistently(taskHub.EmitCallCount).Should(Equal(0))
				})
			})
		})
	})

	Describe("DeleteTask", func() {
		Context("when the delete request is normal", func() {
			var (
				taskGuid string
				task     *models.Task
			)

			BeforeEach(func() {
				taskGuid = "task-guid"
				task = &models.Task{TaskGuid: "guid"}
				fakeTaskDB.DeleteTaskReturns(task, nil)
			})

			JustBeforeEach(func() {
				err = controller.DeleteTask(ctx, logger, taskGuid)
			})

			Context("when deleting the task succeeds", func() {
				It("returns no error", func() {
					Expect(fakeTaskDB.DeleteTaskCallCount()).To(Equal(1))
					_, _, taskGuid := fakeTaskDB.DeleteTaskArgsForCall(0)
					Expect(taskGuid).To(Equal("task-guid"))
					Expect(err).NotTo(HaveOccurred())
				})

				It("emits a change to the hub", func() {
					Eventually(taskHub.EmitCallCount).Should(Equal(1))
					event := taskHub.EmitArgsForCall(0)
					removedEvent, ok := event.(*models.TaskRemovedEvent)
					Expect(ok).To(BeTrue())
					Expect(removedEvent.Task).To(Equal(task))
				})
			})

			Context("when desiring the task fails", func() {
				BeforeEach(func() {
					fakeTaskDB.DeleteTaskReturns(nil, errors.New("kaboom"))
				})

				It("responds with an error", func() {
					Expect(err).To(MatchError("kaboom"))
				})

				It("does not emit a change to the hub", func() {
					Consistently(taskHub.EmitCallCount).Should(Equal(0))
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
			)

			BeforeEach(func() {
				cellPresence := models.NewCellPresence("cell-id", "1.1.1.1", "", "z1", models.CellCapacity{}, nil, nil, nil, nil)
				cellSet = models.CellSet{"cell-id": &cellPresence}
				fakeServiceClient.CellsReturns(cellSet, nil)

				fakeTaskDB.ConvergeTasksReturns(db.TaskConvergenceResult{
					TasksToAuction:  nil,
					TasksToComplete: nil,
					Events:          nil,
					Metrics: db.TaskMetrics{
						TasksPending:   1,
						TasksRunning:   2,
						TasksCompleted: 3,
						TasksResolving: 4,
						TasksPruned:    5,
						TasksKicked:    6,
					},
				})
			})

			JustBeforeEach(func() {
				err = controller.ConvergeTasks(ctx, logger, kickTaskDuration, expirePendingTaskDuration, expireCompletedTaskDuration)
			})

			It("calls ConvergeTasks", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeTaskDB.ConvergeTasksCallCount()).To(Equal(1))
				taskContext, taskLogger, actualCellSet, actualKickDuration, actualPendingDuration, actualCompletedDuration := fakeTaskDB.ConvergeTasksArgsForCall(0)
				Expect(taskContext).To(Equal(ctx))
				Expect(taskLogger.SessionName()).To(ContainSubstring("converge-tasks"))
				Expect(actualCellSet).To(BeEquivalentTo(cellSet))
				Expect(actualKickDuration).To(BeEquivalentTo(kickTaskDuration))
				Expect(actualPendingDuration).To(BeEquivalentTo(expirePendingTaskDuration))
				Expect(actualCompletedDuration).To(BeEquivalentTo(expireCompletedTaskDuration))
			})

			It("records task count metrics", func() {
				Expect(fakeTaskStatNotifier.RecordTaskCountsCallCount()).To(Equal(1))

				pending, running, completed, resolving, pruned, kicked := fakeTaskStatNotifier.RecordTaskCountsArgsForCall(0)
				Expect(pending).To(Equal(1))
				Expect(running).To(Equal(2))
				Expect(completed).To(Equal(3))
				Expect(resolving).To(Equal(4))
				Expect(pruned).To(Equal(uint64(5)))
				Expect(kicked).To(Equal(uint64(6)))
			})

			It("records the convergence runs counter and duration", func() {
				Expect(fakeTaskStatNotifier.RecordConvergenceDurationCallCount()).To(Equal(1))
				Expect(fakeTaskStatNotifier.RecordConvergenceDurationArgsForCall(0)).To(BeNumerically(">", 0))
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
					_, _, actualCellSet, _, _, _ := fakeTaskDB.ConvergeTasksArgsForCall(0)
					Expect(actualCellSet).To(BeEquivalentTo(models.CellSet{}))
				})
			})

			Context("when there are tasks to complete", func() {
				const taskGuid1 = "to-complete-1"
				const taskGuid2 = "to-complete-2"

				BeforeEach(func() {
					task1 := model_helpers.NewValidTask(taskGuid1)
					task2 := model_helpers.NewValidTask(taskGuid2)
					convergenceResult := db.TaskConvergenceResult{
						TasksToAuction:  nil,
						TasksToComplete: []*models.Task{task1, task2},
						Events:          []models.Event{},
					}
					fakeTaskDB.ConvergeTasksReturns(convergenceResult)
				})

				It("submits the tasks to the workpool", func() {
					expectedCallCount := 2
					Expect(fakeTaskCompletionClient.SubmitCallCount()).To(Equal(expectedCallCount))

					_, _, submittedTask1 := fakeTaskCompletionClient.SubmitArgsForCall(0)
					_, _, submittedTask2 := fakeTaskCompletionClient.SubmitArgsForCall(1)
					Expect([]string{submittedTask1.TaskGuid, submittedTask2.TaskGuid}).To(ConsistOf(taskGuid1, taskGuid2))

					task1Completions := 0
					task2Completions := 0
					for i := 0; i < expectedCallCount; i++ {
						db, _, task := fakeTaskCompletionClient.SubmitArgsForCall(i)
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
					convergenceResult := db.TaskConvergenceResult{
						TasksToAuction:  []*auctioneer.TaskStartRequest{&taskStartRequest1, &taskStartRequest2},
						TasksToComplete: nil,
						Events:          nil,
					}
					fakeTaskDB.ConvergeTasksReturns(convergenceResult)
				})

				It("requests an auction", func() {
					Expect(fakeAuctioneerClient.RequestTaskAuctionsCallCount()).To(Equal(1))

					_, traceId, requestedTasks := fakeAuctioneerClient.RequestTaskAuctionsArgsForCall(0)
					Expect(traceId).NotTo(BeNil())
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

			Context("when there are events to emit", func() {
				var event1, event2 models.Event

				BeforeEach(func() {
					event1 = models.NewTaskRemovedEvent(&models.Task{TaskGuid: "removed-task-1"})
					event2 = models.NewTaskRemovedEvent(&models.Task{TaskGuid: "removed-task-2"})
					convergenceResult := db.TaskConvergenceResult{
						TasksToAuction:  nil,
						TasksToComplete: nil,
						Events:          []models.Event{event1, event2},
					}
					fakeTaskDB.ConvergeTasksReturns(convergenceResult)
				})

				It("emits a Task event to the hub", func() {
					Eventually(taskHub.EmitCallCount).Should(Equal(2))

					e1 := taskHub.EmitArgsForCall(0)
					e2 := taskHub.EmitArgsForCall(1)

					events := []*models.TaskRemovedEvent{
						e1.(*models.TaskRemovedEvent),
						e2.(*models.TaskRemovedEvent),
					}

					Expect(events).To(ContainElement(event1))
					Expect(events).To(ContainElement(event2))
				})
			})
		})
	})
})
