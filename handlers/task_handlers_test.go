package handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/cloudfoundry-incubator/auctioneer"
	"github.com/cloudfoundry-incubator/auctioneer/auctioneerfakes"
	"github.com/cloudfoundry-incubator/bbs/db/fakes"
	"github.com/cloudfoundry-incubator/bbs/handlers"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/models/test/model_helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
)

var _ = Describe("Task Handlers", func() {
	var (
		logger               lager.Logger
		fakeTaskDB           *fakes.FakeTaskDB
		fakeAuctioneerClient *auctioneerfakes.FakeClient

		responseRecorder *httptest.ResponseRecorder

		handler *handlers.TaskHandler

		task1 models.Task
		task2 models.Task

		requestBody interface{}

		request *http.Request
	)

	BeforeEach(func() {
		fakeTaskDB = new(fakes.FakeTaskDB)
		fakeAuctioneerClient = new(auctioneerfakes.FakeClient)

		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		responseRecorder = httptest.NewRecorder()
		handler = handlers.NewTaskHandler(logger, fakeTaskDB, fakeAuctioneerClient, fakeServiceClient, fakeRepClientFactory)
	})

	Describe("Tasks", func() {
		BeforeEach(func() {
			task1 = models.Task{Domain: "domain-1"}
			task2 = models.Task{CellId: "cell-id"}
			requestBody = &models.TasksRequest{}
		})

		JustBeforeEach(func() {
			request := newTestRequest(requestBody)
			handler.Tasks(responseRecorder, request)
		})

		Context("when reading tasks from DB succeeds", func() {
			var tasks []*models.Task

			BeforeEach(func() {
				tasks = []*models.Task{&task1, &task2}
				fakeTaskDB.TasksReturns(tasks, nil)
			})

			It("returns a list of task", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := models.TasksResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(BeNil())
				Expect(response.Tasks).To(Equal(tasks))
			})

			It("calls the DB with no filter", func() {
				Expect(fakeTaskDB.TasksCallCount()).To(Equal(1))
				_, filter := fakeTaskDB.TasksArgsForCall(0)
				Expect(filter).To(Equal(models.TaskFilter{}))
			})

			Context("and filtering by domain", func() {
				BeforeEach(func() {
					requestBody = &models.TasksRequest{
						Domain: "domain-1",
					}
				})

				It("calls the DB with a domain filter", func() {
					Expect(fakeTaskDB.TasksCallCount()).To(Equal(1))
					_, filter := fakeTaskDB.TasksArgsForCall(0)
					Expect(filter.Domain).To(Equal("domain-1"))
				})
			})

			Context("and filtering by cell id", func() {
				BeforeEach(func() {
					requestBody = &models.TasksRequest{
						CellId: "cell-id",
					}
				})

				It("calls the DB with a cell filter", func() {
					Expect(fakeTaskDB.TasksCallCount()).To(Equal(1))
					_, filter := fakeTaskDB.TasksArgsForCall(0)
					Expect(filter.CellID).To(Equal("cell-id"))
				})
			})
		})

		Context("when the DB errors out", func() {
			BeforeEach(func() {
				fakeTaskDB.TasksReturns(nil, models.ErrUnknownError)
			})

			It("provides relevant error information", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := models.TasksResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrUnknownError))
			})
		})
	})

	Describe("TaskByGuid", func() {
		var taskGuid = "task-guid"

		BeforeEach(func() {
			requestBody = &models.TaskByGuidRequest{
				TaskGuid: taskGuid,
			}
		})

		JustBeforeEach(func() {
			request := newTestRequest(requestBody)
			handler.TaskByGuid(responseRecorder, request)
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
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := models.TaskResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(BeNil())
				Expect(response.Task).To(Equal(task))
			})
		})

		Context("when the DB returns no task", func() {
			BeforeEach(func() {
				fakeTaskDB.TaskByGuidReturns(nil, models.ErrResourceNotFound)
			})

			It("returns a resource not found error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := models.TaskResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrResourceNotFound))
			})
		})

		Context("when the DB errors out", func() {
			BeforeEach(func() {
				fakeTaskDB.TaskByGuidReturns(nil, models.ErrUnknownError)
			})

			It("provides relevant error information", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := models.TaskResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrUnknownError))
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
			requestBody = &models.DesireTaskRequest{
				TaskGuid:       taskGuid,
				Domain:         domain,
				TaskDefinition: taskDef,
			}
		})

		JustBeforeEach(func() {
			request := newTestRequest(requestBody)
			handler.DesireTask(responseRecorder, request)
		})

		Context("when the desire is successful", func() {
			It("desires the task with the requested definitions", func() {
				Expect(fakeTaskDB.DesireTaskCallCount()).To(Equal(1))
				_, actualTaskDef, actualTaskGuid, actualDomain := fakeTaskDB.DesireTaskArgsForCall(0)
				Expect(actualTaskDef).To(Equal(taskDef))
				Expect(actualTaskGuid).To(Equal(taskGuid))
				Expect(actualDomain).To(Equal(domain))

				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := &models.TaskLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(BeNil())
			})

			It("requests an auction", func() {
				Expect(fakeAuctioneerClient.RequestTaskAuctionsCallCount()).To(Equal(1))

				expectedStartRequest := auctioneer.NewTaskStartRequestFromModel(taskGuid, domain, taskDef)

				requestedTasks := fakeAuctioneerClient.RequestTaskAuctionsArgsForCall(0)
				Expect(requestedTasks).To(HaveLen(1))
				Expect(*requestedTasks[0]).To(Equal(expectedStartRequest))
			})

			Context("when requesting a task auction succeeds", func() {
				BeforeEach(func() {
					fakeAuctioneerClient.RequestTaskAuctionsReturns(nil)
				})

				It("does not return an error", func() {
					Expect(responseRecorder.Code).To(Equal(http.StatusOK))
					response := &models.TaskLifecycleResponse{}
					err := response.Unmarshal(responseRecorder.Body.Bytes())
					Expect(err).NotTo(HaveOccurred())

					Expect(response.Error).To(BeNil())
				})
			})

			Context("when requesting a task auction fails", func() {
				BeforeEach(func() {
					fakeAuctioneerClient.RequestTaskAuctionsReturns(errors.New("oops"))
				})

				It("does not return an error", func() {
					Expect(responseRecorder.Code).To(Equal(http.StatusOK))
					response := &models.TaskLifecycleResponse{}
					err := response.Unmarshal(responseRecorder.Body.Bytes())
					Expect(err).NotTo(HaveOccurred())

					Expect(response.Error).To(BeNil())
				})

				It("does not request a second auction", func() {
					Consistently(fakeAuctioneerClient.RequestTaskAuctionsCallCount).Should(Equal(1))
				})
			})
		})

		Context("when desiring the task fails", func() {
			BeforeEach(func() {
				fakeTaskDB.DesireTaskReturns(models.ErrUnknownError)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := &models.TaskLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrUnknownError))
			})
		})
	})

	Describe("StartTask", func() {
		Context("when the start is successful", func() {
			BeforeEach(func() {
				requestBody = &models.StartTaskRequest{
					TaskGuid: "task-guid",
					CellId:   "cell-id",
				}
			})

			JustBeforeEach(func() {
				request := newTestRequest(requestBody)
				handler.StartTask(responseRecorder, request)
			})

			It("calls StartTask", func() {
				Expect(fakeTaskDB.StartTaskCallCount()).To(Equal(1))
				taskLogger, taskGuid, cellId := fakeTaskDB.StartTaskArgsForCall(0)
				Expect(taskLogger.SessionName()).To(ContainSubstring("start-task"))
				Expect(taskGuid).To(Equal("task-guid"))
				Expect(cellId).To(Equal("cell-id"))
			})

			Context("when the task should start", func() {
				BeforeEach(func() {
					fakeTaskDB.StartTaskReturns(true, nil)
				})

				It("responds with true", func() {
					Expect(responseRecorder.Code).To(Equal(http.StatusOK))
					response := &models.StartTaskResponse{}
					err := response.Unmarshal(responseRecorder.Body.Bytes())
					Expect(err).NotTo(HaveOccurred())

					Expect(response.Error).To(BeNil())
					Expect(response.ShouldStart).To(BeTrue())
				})
			})

			Context("when the task should not start", func() {
				BeforeEach(func() {
					fakeTaskDB.StartTaskReturns(false, nil)
				})

				It("responds with false", func() {
					Expect(responseRecorder.Code).To(Equal(http.StatusOK))
					response := &models.StartTaskResponse{}
					err := response.Unmarshal(responseRecorder.Body.Bytes())
					Expect(err).NotTo(HaveOccurred())

					Expect(response.Error).To(BeNil())
					Expect(response.ShouldStart).To(BeFalse())
				})
			})

			Context("when the DB fails", func() {
				BeforeEach(func() {
					fakeTaskDB.StartTaskReturns(false, models.ErrResourceExists)
				})

				It("bubbles up the underlying model error", func() {
					Expect(responseRecorder.Code).To(Equal(http.StatusOK))
					response := &models.StartTaskResponse{}
					err := response.Unmarshal(responseRecorder.Body.Bytes())
					Expect(err).NotTo(HaveOccurred())

					Expect(response.Error).To(Equal(models.ErrResourceExists))
				})
			})
		})
	})

	Describe("CancelTask", func() {
		var request *http.Request

		BeforeEach(func() {
			requestBody = &models.TaskGuidRequest{
				TaskGuid: "task-guid",
			}

			request = newTestRequest(requestBody)
		})

		JustBeforeEach(func() {
			handler.CancelTask(responseRecorder, request)
			Expect(responseRecorder.Code).To(Equal(http.StatusOK))
		})

		Context("when the cancel request is normal", func() {
			Context("when canceling the task in the db succeeds", func() {
				BeforeEach(func() {
					task1 = *model_helpers.NewValidTask("guid")
					cellPresence := models.CellPresence{CellId: "cell-id"}
					fakeTaskDB.TaskByGuidReturns(&task1, nil)
					fakeServiceClient.CellByIdReturns(&cellPresence, nil)
				})

				It("returns no error", func() {
					Expect(fakeTaskDB.CancelTaskCallCount()).To(Equal(1))
					taskLogger, taskGuid := fakeTaskDB.CancelTaskArgsForCall(0)
					Expect(taskLogger.SessionName()).To(ContainSubstring("cancel-task"))
					Expect(taskGuid).To(Equal("task-guid"))

					response := &models.TaskLifecycleResponse{}
					err := response.Unmarshal(responseRecorder.Body.Bytes())
					Expect(err).NotTo(HaveOccurred())

					Expect(response.Error).To(BeNil())
				})

				It("stops the task on the rep", func() {
					Expect(fakeRepClient.CancelTaskCallCount()).To(Equal(1))
					guid := fakeRepClient.CancelTaskArgsForCall(0)
					Expect(guid).To(Equal("task-guid"))
				})

				// after persisting the task in the DB, all additional functionality is best-effort
				Context("when fetching the task fails", func() {
					BeforeEach(func() {
						fakeTaskDB.TaskByGuidReturns(nil, errors.New("nope"))
					})

					It("does not return an error", func() {
						response := &models.TaskLifecycleResponse{}
						err := response.Unmarshal(responseRecorder.Body.Bytes())
						Expect(err).NotTo(HaveOccurred())
						Expect(response.Error).To(BeNil())

						Expect(fakeServiceClient.CellByIdCallCount()).To(Equal(0))
						Expect(fakeRepClient.CancelTaskCallCount()).To(Equal(0))
					})
				})

				Context("when the task has no cell id", func() {
					BeforeEach(func() {
						task1.CellId = ""
					})

					It("does not return an error", func() {
						response := &models.TaskLifecycleResponse{}
						err := response.Unmarshal(responseRecorder.Body.Bytes())
						Expect(err).NotTo(HaveOccurred())
						Expect(response.Error).To(BeNil())

						Expect(fakeServiceClient.CellByIdCallCount()).To(Equal(0))
						Expect(fakeRepClient.CancelTaskCallCount()).To(Equal(0))
					})
				})

				Context("when fetching the cell presence fails", func() {
					BeforeEach(func() {
						fakeServiceClient.CellByIdReturns(nil, errors.New("lol"))
					})

					It("does not return an error", func() {
						response := &models.TaskLifecycleResponse{}
						err := response.Unmarshal(responseRecorder.Body.Bytes())
						Expect(err).NotTo(HaveOccurred())
						Expect(response.Error).To(BeNil())

						Expect(fakeRepClient.CancelTaskCallCount()).To(Equal(0))
					})
				})

				Context("when we fail to cancel the task on the rep", func() {
					BeforeEach(func() {
						fakeRepClient.CancelTaskReturns(errors.New("lol"))
					})

					It("does not return an error", func() {
						response := &models.TaskLifecycleResponse{}
						err := response.Unmarshal(responseRecorder.Body.Bytes())
						Expect(err).NotTo(HaveOccurred())
						Expect(response.Error).To(BeNil())
					})
				})
			})

			Context("when cancelling the task fails", func() {
				BeforeEach(func() {
					fakeTaskDB.CancelTaskReturns(models.ErrUnknownError)
				})

				It("responds with an error", func() {
					response := &models.TaskLifecycleResponse{}
					err := response.Unmarshal(responseRecorder.Body.Bytes())
					Expect(err).NotTo(HaveOccurred())

					Expect(response.Error).To(Equal(models.ErrUnknownError))
				})
			})
		})

		Context("when the cancel task request is not valid", func() {
			BeforeEach(func() {
				request = newTestRequest("{{")
			})

			It("returns an BadRequest error", func() {
				response := &models.TaskLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrBadRequest))
			})
		})
	})

	Describe("FailTask", func() {
		var (
			taskGuid      string
			failureReason string
		)

		BeforeEach(func() {
			taskGuid = "task-guid"
			failureReason = "just cuz ;)"

			requestBody = &models.FailTaskRequest{
				TaskGuid:      taskGuid,
				FailureReason: failureReason,
			}
		})

		JustBeforeEach(func() {
			request = newTestRequest(requestBody)
			handler.FailTask(responseRecorder, request)
		})
		Context("when failing the task succeeds", func() {
			It("returns no error", func() {
				_, actualTaskGuid, actualFailureReason := fakeTaskDB.FailTaskArgsForCall(0)
				Expect(actualTaskGuid).To(Equal(taskGuid))
				Expect(actualFailureReason).To(Equal(failureReason))

				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := &models.TaskLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(BeNil())
			})
		})

		Context("when failing the task fails", func() {
			BeforeEach(func() {
				fakeTaskDB.FailTaskReturns(models.ErrUnknownError)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := &models.TaskLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrUnknownError))
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
		)

		BeforeEach(func() {
			taskGuid = "t-guid"
			cellId = "c-id"
			failed = true
			failureReason = "some-error"
			result = "yeah"

			requestBody = &models.CompleteTaskRequest{
				TaskGuid:      taskGuid,
				CellId:        cellId,
				Failed:        failed,
				FailureReason: failureReason,
				Result:        result,
			}
		})

		JustBeforeEach(func() {
			request := newTestRequest(requestBody)
			handler.CompleteTask(responseRecorder, request)
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

				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := &models.TaskLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(BeNil())
			})
		})

		Context("when completing the task fails", func() {
			BeforeEach(func() {
				fakeTaskDB.CompleteTaskReturns(models.ErrUnknownError)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := &models.TaskLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrUnknownError))
			})
		})
	})

	Describe("ResolvingTask", func() {
		Context("when the resolving request is normal", func() {
			BeforeEach(func() {
				requestBody = &models.TaskGuidRequest{
					TaskGuid: "task-guid",
				}
			})
			JustBeforeEach(func() {
				request := newTestRequest(requestBody)
				handler.ResolvingTask(responseRecorder, request)
			})

			Context("when resolvinging the task succeeds", func() {
				It("returns no error", func() {
					Expect(fakeTaskDB.ResolvingTaskCallCount()).To(Equal(1))
					_, taskGuid := fakeTaskDB.ResolvingTaskArgsForCall(0)
					Expect(taskGuid).To(Equal("task-guid"))

					Expect(responseRecorder.Code).To(Equal(http.StatusOK))
					response := &models.TaskLifecycleResponse{}
					err := response.Unmarshal(responseRecorder.Body.Bytes())
					Expect(err).NotTo(HaveOccurred())

					Expect(response.Error).To(BeNil())
				})
			})

			Context("when desiring the task fails", func() {
				BeforeEach(func() {
					fakeTaskDB.ResolvingTaskReturns(models.ErrUnknownError)
				})

				It("responds with an error", func() {
					Expect(responseRecorder.Code).To(Equal(http.StatusOK))
					response := &models.TaskLifecycleResponse{}
					err := response.Unmarshal(responseRecorder.Body.Bytes())
					Expect(err).NotTo(HaveOccurred())

					Expect(response.Error).To(Equal(models.ErrUnknownError))
				})
			})
		})
	})

	Describe("DeleteTask", func() {
		Context("when the delete request is normal", func() {
			BeforeEach(func() {
				requestBody = &models.TaskGuidRequest{
					TaskGuid: "task-guid",
				}
			})
			JustBeforeEach(func() {
				request := newTestRequest(requestBody)
				handler.DeleteTask(responseRecorder, request)
			})

			Context("when deleting the task succeeds", func() {
				It("returns no error", func() {
					Expect(fakeTaskDB.DeleteTaskCallCount()).To(Equal(1))
					_, taskGuid := fakeTaskDB.DeleteTaskArgsForCall(0)
					Expect(taskGuid).To(Equal("task-guid"))

					Expect(responseRecorder.Code).To(Equal(http.StatusOK))
					response := &models.TaskLifecycleResponse{}
					err := response.Unmarshal(responseRecorder.Body.Bytes())
					Expect(err).NotTo(HaveOccurred())

					Expect(response.Error).To(BeNil())
				})
			})

			Context("when desiring the task fails", func() {
				BeforeEach(func() {
					fakeTaskDB.DeleteTaskReturns(models.ErrUnknownError)
				})

				It("responds with an error", func() {
					Expect(responseRecorder.Code).To(Equal(http.StatusOK))
					response := &models.TaskLifecycleResponse{}
					err := response.Unmarshal(responseRecorder.Body.Bytes())
					Expect(err).NotTo(HaveOccurred())

					Expect(response.Error).To(Equal(models.ErrUnknownError))
				})
			})
		})
	})

	Describe("ConvergeTasks", func() {
		Context("when the request is normal", func() {
			var (
				kickTaskDuration            = int64(10 * time.Second)
				expirePendingTaskDuration   = int64(10 * time.Second)
				expireCompletedTaskDuration = int64(10 * time.Second)
			)
			BeforeEach(func() {
				requestBody = &models.ConvergeTasksRequest{
					KickTaskDuration:            kickTaskDuration,
					ExpirePendingTaskDuration:   expirePendingTaskDuration,
					ExpireCompletedTaskDuration: expireCompletedTaskDuration,
				}
			})

			JustBeforeEach(func() {
				request := newTestRequest(requestBody)
				handler.ConvergeTasks(responseRecorder, request)
			})

			It("calls ConvergeTasks", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				Expect(fakeTaskDB.ConvergeTasksCallCount()).To(Equal(1))
				taskLogger, actualKickDuration, actualPendingDuration, actualCompletedDuration := fakeTaskDB.ConvergeTasksArgsForCall(0)
				Expect(taskLogger.SessionName()).To(ContainSubstring("converge-tasks"))
				Expect(actualKickDuration).To(BeEquivalentTo(kickTaskDuration))
				Expect(actualPendingDuration).To(BeEquivalentTo(expirePendingTaskDuration))
				Expect(actualCompletedDuration).To(BeEquivalentTo(expireCompletedTaskDuration))

				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := &models.TaskLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(BeNil())
			})
		})
	})
})
