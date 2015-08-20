package handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/cloudfoundry-incubator/bbs/db/fakes"
	"github.com/cloudfoundry-incubator/bbs/handlers"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/models/internal/model_helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
)

var _ = Describe("Task Handlers", func() {
	var (
		logger           lager.Logger
		fakeTaskDB       *fakes.FakeTaskDB
		responseRecorder *httptest.ResponseRecorder

		handler *handlers.TaskHandler

		task1 models.Task
		task2 models.Task

		requestBody interface{}

		request *http.Request
	)

	BeforeEach(func() {
		fakeTaskDB = new(fakes.FakeTaskDB)
		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		responseRecorder = httptest.NewRecorder()
		handler = handlers.NewTaskHandler(logger, fakeTaskDB)
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

			requestBody interface{}
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
				request = newTestRequest(&models.StartTaskRequest{
					TaskGuid: "task-guid",
					CellId:   "cell-id",
				})
			})

			It("responds with 200 OK", func() {
				handler.StartTask(responseRecorder, request)
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})

			It("calls StartTask", func() {
				handler.StartTask(responseRecorder, request)
				Expect(fakeTaskDB.StartTaskCallCount()).To(Equal(1))
				taskLogger, taskGuid, cellId := fakeTaskDB.StartTaskArgsForCall(0)
				Expect(taskLogger.SessionName()).To(ContainSubstring("start-task"))
				Expect(taskGuid).To(Equal("task-guid"))
				Expect(cellId).To(Equal("cell-id"))
			})

			It("responds with true when the task should start", func() {
				fakeTaskDB.StartTaskReturns(true, nil)
				handler.StartTask(responseRecorder, request)
				res := &models.StartTaskResponse{}
				err := res.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())
				Expect(res.GetShouldStart()).To(BeTrue())
			})

			It("responds with false when the task should not start", func() {
				fakeTaskDB.StartTaskReturns(false, nil)
				handler.StartTask(responseRecorder, request)
				res := &models.StartTaskResponse{}
				err := res.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())
				Expect(res.GetShouldStart()).To(BeFalse())
			})

			It("bubbles up the underlying model error", func() {
				fakeTaskDB.StartTaskReturns(false, models.ErrResourceExists)
				handler.StartTask(responseRecorder, request)
				res := &models.Error{}
				err := res.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal(models.ErrResourceExists))
			})
		})

		Context("when the request body is not a StartRequest", func() {
			BeforeEach(func() {
				request = newTestRequest("foo")
			})

			It("responds with 400 BAD REQUEST", func() {
				handler.StartTask(responseRecorder, request)
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("returns an Invalid Request error", func() {
				handler.StartTask(responseRecorder, request)
				var bbsError models.Error
				err := bbsError.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(bbsError.Equal(models.ErrBadRequest)).To(BeTrue())
				Expect(bbsError.Message).To(ContainSubstring("unmarshal"))
			})
		})

		Context("when the request body fails to stream", func() {
			BeforeEach(func() {
				request = newTestRequest(newExplodingReader(errors.New("foobar")))
			})

			It("responds with 400 BAD REQUEST", func() {
				handler.StartTask(responseRecorder, request)
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("returns an Invalid Request error", func() {
				handler.StartTask(responseRecorder, request)
				var bbsError models.Error
				err := bbsError.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(bbsError.Equal(models.ErrBadRequest)).To(BeTrue())
				Expect(bbsError.Message).To(ContainSubstring("foobar"))
			})
		})
	})
	Describe("CancelTask", func() {
		Context("when the cancel request is normal", func() {
			BeforeEach(func() {
				request = newTestRequest(&models.TaskGuidRequest{
					TaskGuid: "task-guid",
				})
			})

			It("responds with 200 OK", func() {
				handler.CancelTask(responseRecorder, request)
				Expect(responseRecorder.Code).To(Equal(http.StatusNoContent))
			})

			It("calls CancelTask", func() {
				handler.CancelTask(responseRecorder, request)
				Expect(fakeTaskDB.CancelTaskCallCount()).To(Equal(1))
				taskLogger, taskGuid := fakeTaskDB.CancelTaskArgsForCall(0)
				Expect(taskLogger.SessionName()).To(ContainSubstring("cancel-task"))
				Expect(taskGuid).To(Equal("task-guid"))
			})

			It("bubbles up the underlying model error", func() {
				fakeTaskDB.CancelTaskReturns(models.ErrResourceExists)
				handler.CancelTask(responseRecorder, request)
				res := &models.Error{}
				err := res.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal(models.ErrResourceExists))
			})
		})
		Context("when the request body is not a CancelRequest", func() {
			BeforeEach(func() {
				request = newTestRequest("foo")
			})

			It("responds with 400 BAD REQUEST", func() {
				handler.CancelTask(responseRecorder, request)
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("returns an Invalid Request error", func() {
				handler.CancelTask(responseRecorder, request)
				var bbsError models.Error
				err := bbsError.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(bbsError.Equal(models.ErrBadRequest)).To(BeTrue())
				Expect(bbsError.Message).To(ContainSubstring("unmarshal"))
			})
		})

		Context("when the request body fails to stream", func() {
			BeforeEach(func() {
				request = newTestRequest(newExplodingReader(errors.New("foobar")))
			})

			It("responds with 400 BAD REQUEST", func() {
				handler.CancelTask(responseRecorder, request)
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("returns an Invalid Request error", func() {
				handler.CancelTask(responseRecorder, request)
				var bbsError models.Error
				err := bbsError.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(bbsError.Equal(models.ErrBadRequest)).To(BeTrue())
				Expect(bbsError.Message).To(ContainSubstring("foobar"))
			})
		})
	})
	Describe("FailTask", func() {
		var (
			taskGuid      string
			failureReason string

			requestBody interface{}
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
			It("calls FailTask", func() {
				taskLogger, actualTaskGuid, actualFailureReason := fakeTaskDB.FailTaskArgsForCall(0)
				Expect(taskLogger.SessionName()).To(ContainSubstring("fail-task"))
				Expect(actualTaskGuid).To(Equal(taskGuid))
				Expect(actualFailureReason).To(Equal(failureReason))
			})
		})

		Context("when failing the task fails ", func() {
			BeforeEach(func() {
				fakeTaskDB.FailTaskReturns(models.ErrResourceExists)
			})

			It("bubbles up the underlying model error", func() {
				res := &models.Error{}
				err := res.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal(models.ErrResourceExists))
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

			requestBody interface{}
		)

		BeforeEach(func() {
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
			It("responds with 200 OK", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusNoContent))
			})

			It("calls CompleteTask", func() {
				Expect(fakeTaskDB.CompleteTaskCallCount()).To(Equal(1))
				taskLogger, actualTaskGuid, actualCellId, actualFailed, actualFailureReason, actualResult := fakeTaskDB.CompleteTaskArgsForCall(0)
				Expect(taskLogger.SessionName()).To(ContainSubstring("complete-task"))
				Expect(actualTaskGuid).To(Equal(taskGuid))
				Expect(actualCellId).To(Equal(cellId))
				Expect(actualFailed).To(Equal(failed))
				Expect(actualFailureReason).To(Equal(failureReason))
				Expect(actualResult).To(Equal(result))
			})

		})
		Context("when completing the task fails", func() {
			BeforeEach(func() {
				fakeTaskDB.CompleteTaskReturns(models.ErrResourceExists)
			})

			It("bubbles up the underlying model error", func() {
				res := &models.Error{}
				err := res.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal(models.ErrResourceExists))
			})
		})
	})

	Describe("ResolvingTask", func() {
		Context("when the request is normal", func() {
			var expected *models.TaskGuidRequest
			BeforeEach(func() {
				expected = &models.TaskGuidRequest{
					TaskGuid: "task-guid",
				}
				request = newTestRequest(expected)
			})

			It("responds with 200 OK", func() {
				handler.ResolvingTask(responseRecorder, request)
				Expect(responseRecorder.Code).To(Equal(http.StatusNoContent))
			})

			It("calls ResolvingTask", func() {
				handler.ResolvingTask(responseRecorder, request)
				Expect(fakeTaskDB.ResolvingTaskCallCount()).To(Equal(1))
				taskLogger, taskGuid := fakeTaskDB.ResolvingTaskArgsForCall(0)
				Expect(taskLogger.SessionName()).To(ContainSubstring("resolving-task"))
				Expect(taskGuid).To(BeEquivalentTo("task-guid"))
			})

			It("bubbles up the underlying model error", func() {
				fakeTaskDB.ResolvingTaskReturns(models.ErrResourceExists)
				handler.ResolvingTask(responseRecorder, request)
				res := &models.Error{}
				err := res.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal(models.ErrResourceExists))
			})
		})

		Context("when the request body is not a TaskGuidRequest", func() {
			BeforeEach(func() {
				request = newTestRequest("foo")
			})

			It("responds with 400 BAD REQUEST", func() {
				handler.ResolvingTask(responseRecorder, request)
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("returns an Invalid Request error", func() {
				handler.ResolvingTask(responseRecorder, request)
				var bbsError models.Error
				err := bbsError.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(bbsError.Equal(models.ErrBadRequest)).To(BeTrue())
				Expect(bbsError.Message).To(ContainSubstring("unmarshal"))
			})
		})

		Context("when the request body resolvings to stream", func() {
			BeforeEach(func() {
				request = newTestRequest(newExplodingReader(errors.New("foobar")))
			})

			It("responds with 400 BAD REQUEST", func() {
				handler.ResolvingTask(responseRecorder, request)
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("returns an Invalid Request error", func() {
				handler.ResolvingTask(responseRecorder, request)
				var bbsError models.Error
				err := bbsError.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(bbsError.Equal(models.ErrBadRequest)).To(BeTrue())
				Expect(bbsError.Message).To(ContainSubstring("foobar"))
			})
		})
	})

	Describe("ResolveTask", func() {
		Context("when the request is normal", func() {
			var expected *models.TaskGuidRequest
			BeforeEach(func() {
				expected = &models.TaskGuidRequest{
					TaskGuid: "task-guid",
				}
				request = newTestRequest(expected)
			})

			It("responds with 200 OK", func() {
				handler.ResolveTask(responseRecorder, request)
				Expect(responseRecorder.Code).To(Equal(http.StatusNoContent))
			})

			It("calls ResolveTask", func() {
				handler.ResolveTask(responseRecorder, request)
				Expect(fakeTaskDB.ResolveTaskCallCount()).To(Equal(1))
				taskLogger, taskGuid := fakeTaskDB.ResolveTaskArgsForCall(0)
				Expect(taskLogger.SessionName()).To(ContainSubstring("resolve-task"))
				Expect(taskGuid).To(BeEquivalentTo("task-guid"))
			})

			It("bubbles up the underlying model error", func() {
				fakeTaskDB.ResolveTaskReturns(models.ErrResourceExists)
				handler.ResolveTask(responseRecorder, request)
				res := &models.Error{}
				err := res.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal(models.ErrResourceExists))
			})
		})

		Context("when the request body is not a TaskGuidRequest", func() {
			BeforeEach(func() {
				request = newTestRequest("foo")
			})

			It("responds with 400 BAD REQUEST", func() {
				handler.ResolveTask(responseRecorder, request)
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("returns an Invalid Request error", func() {
				handler.ResolveTask(responseRecorder, request)
				var bbsError models.Error
				err := bbsError.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(bbsError.Equal(models.ErrBadRequest)).To(BeTrue())
				Expect(bbsError.Message).To(ContainSubstring("unmarshal"))
			})
		})

		Context("when the request body resolves to stream", func() {
			BeforeEach(func() {
				request = newTestRequest(newExplodingReader(errors.New("foobar")))
			})

			It("responds with 400 BAD REQUEST", func() {
				handler.ResolveTask(responseRecorder, request)
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("returns an Invalid Request error", func() {
				handler.ResolveTask(responseRecorder, request)
				var bbsError models.Error
				err := bbsError.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(bbsError.Equal(models.ErrBadRequest)).To(BeTrue())
				Expect(bbsError.Message).To(ContainSubstring("foobar"))
			})
		})
	})

	Describe("ConvergeTasks", func() {
		Context("when the request is normal", func() {
			var expected *models.ConvergeTasksRequest
			BeforeEach(func() {
				expected = &models.ConvergeTasksRequest{
					KickTaskDuration:            int64(10 * time.Second),
					ExpirePendingTaskDuration:   int64(10 * time.Second),
					ExpireCompletedTaskDuration: int64(10 * time.Second),
				}
				request = newTestRequest(expected)
			})

			It("responds with 200 OK", func() {
				handler.ConvergeTasks(responseRecorder, request)
				Expect(responseRecorder.Code).To(Equal(http.StatusNoContent))
			})

			It("calls ConvergeTasks", func() {
				handler.ConvergeTasks(responseRecorder, request)
				Expect(fakeTaskDB.ConvergeTasksCallCount()).To(Equal(1))
				taskLogger, kickDuration, pendingDuration, completedDuration := fakeTaskDB.ConvergeTasksArgsForCall(0)
				Expect(taskLogger.SessionName()).To(ContainSubstring("converge-tasks"))
				Expect(kickDuration).To(BeEquivalentTo(expected.KickTaskDuration))
				Expect(pendingDuration).To(BeEquivalentTo(expected.ExpirePendingTaskDuration))
				Expect(completedDuration).To(BeEquivalentTo(expected.ExpireCompletedTaskDuration))
			})
		})

		Context("when the request body is not a TaskGuidRequest", func() {
			BeforeEach(func() {
				request = newTestRequest("foo")
			})

			It("responds with 400 BAD REQUEST", func() {
				handler.ConvergeTasks(responseRecorder, request)
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("returns an Invalid Request error", func() {
				handler.ConvergeTasks(responseRecorder, request)
				var bbsError models.Error
				err := bbsError.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(bbsError.Equal(models.ErrBadRequest)).To(BeTrue())
				Expect(bbsError.Message).To(ContainSubstring("unmarshal"))
			})
		})

		Context("when the request body resolves to stream", func() {
			BeforeEach(func() {
				request = newTestRequest(newExplodingReader(errors.New("foobar")))
			})

			It("responds with 400 BAD REQUEST", func() {
				handler.ConvergeTasks(responseRecorder, request)
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("returns an Invalid Request error", func() {
				handler.ConvergeTasks(responseRecorder, request)
				var bbsError models.Error
				err := bbsError.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(bbsError.Equal(models.ErrBadRequest)).To(BeTrue())
				Expect(bbsError.Message).To(ContainSubstring("foobar"))
			})
		})
	})
})
