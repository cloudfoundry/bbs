package handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"

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
		request          *http.Request

		handler *handlers.TaskHandler

		task1 models.Task
		task2 models.Task
	)

	BeforeEach(func() {
		fakeTaskDB = new(fakes.FakeTaskDB)
		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		responseRecorder = httptest.NewRecorder()
		request = nil
		handler = handlers.NewTaskHandler(logger, fakeTaskDB)
	})

	Describe("Tasks", func() {
		BeforeEach(func() {
			request = newTestRequest("")

			task1 = models.Task{Domain: "domain-1"}
			task2 = models.Task{CellId: "cell-id"}
		})

		JustBeforeEach(func() {
			handler.Tasks(responseRecorder, request)
		})

		Context("when reading tasks from DB succeeds", func() {
			var tasks *models.Tasks

			BeforeEach(func() {
				tasks = &models.Tasks{
					[]*models.Task{&task1, &task2},
				}
				fakeTaskDB.TasksReturns(tasks, nil)
			})

			It("responds with 200 Status OK", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})

			It("returns a list of task", func() {
				response := &models.Tasks{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response).To(Equal(tasks))
			})

			It("calls the DB with no filter", func() {
				Expect(fakeTaskDB.TasksCallCount()).To(Equal(1))
				_, filter := fakeTaskDB.TasksArgsForCall(0)
				Expect(filter).To(BeNil())
			})

			Context("and filtering by domain", func() {
				BeforeEach(func() {
					var err error
					request, err = http.NewRequest("", "http://example.com?domain=domain-1", nil)
					Expect(err).NotTo(HaveOccurred())
				})

				It("calls the DB with a domain filter", func() {
					Expect(fakeTaskDB.TasksCallCount()).To(Equal(1))
					_, filter := fakeTaskDB.TasksArgsForCall(0)
					Expect(filter(&task1)).To(BeTrue())
					Expect(filter(&task2)).To(BeFalse())
				})
			})

			Context("and filtering by cell id", func() {
				BeforeEach(func() {
					var err error
					request, err = http.NewRequest("", "http://example.com?cell_id=cell-id", nil)
					Expect(err).NotTo(HaveOccurred())
				})

				It("calls the DB with a cell filter", func() {
					Expect(fakeTaskDB.TasksCallCount()).To(Equal(1))
					_, filter := fakeTaskDB.TasksArgsForCall(0)
					Expect(filter(&task1)).To(BeFalse())
					Expect(filter(&task2)).To(BeTrue())
				})
			})

			Context("filtering by domain and cell", func() {
				BeforeEach(func() {
					var err error
					request, err = http.NewRequest("", "http://example.com?domain=d&cell_id=cell-id", nil)
					Expect(err).NotTo(HaveOccurred())
				})

				It("responds with 400 Bad Request", func() {
					Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
				})
			})
		})

		Context("when the DB errors out", func() {
			BeforeEach(func() {
				fakeTaskDB.TasksReturns(&models.Tasks{}, models.ErrUnknownError)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})

			It("provides relevant error information", func() {
				var bbsError models.Error
				err := bbsError.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(bbsError.Equal(models.ErrUnknownError)).To(BeTrue())
			})
		})
	})

	Describe("TaskByGuid", func() {
		var taskGuid = "task-guid"

		BeforeEach(func() {
			request = newTestRequest("")
			request.URL.RawQuery = url.Values{":task_guid": []string{taskGuid}}.Encode()
		})

		JustBeforeEach(func() {
			handler.TaskByGuid(responseRecorder, request)
		})

		Context("when reading a task from the DB succeeds", func() {
			var task *models.Task

			BeforeEach(func() {
				task = &models.Task{TaskGuid: taskGuid}
				fakeTaskDB.TaskByGuidReturns(task, nil)
			})

			It("responds with 200 Status OK", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})

			It("fetches task by guid", func() {
				Expect(fakeTaskDB.TaskByGuidCallCount()).To(Equal(1))
				_, actualGuid := fakeTaskDB.TaskByGuidArgsForCall(0)
				Expect(actualGuid).To(Equal(taskGuid))
			})

			It("returns the task", func() {
				response := &models.Task{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response).To(Equal(task))
			})
		})

		Context("when the DB returns no task", func() {
			BeforeEach(func() {
				fakeTaskDB.TaskByGuidReturns(nil, models.ErrResourceNotFound)
			})

			It("responds with 404", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusNotFound))
			})

			It("returns a resource not found error", func() {
				var bbsError models.Error
				err := bbsError.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(bbsError.Equal(models.ErrResourceNotFound)).To(BeTrue())
			})
		})

		Context("when the DB errors out", func() {
			BeforeEach(func() {
				fakeTaskDB.TaskByGuidReturns(nil, models.ErrUnknownError)
			})

			It("responds with a 500", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})

			It("provides relevant error information", func() {
				var bbsError models.Error
				err := bbsError.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(bbsError.Equal(models.ErrUnknownError)).To(BeTrue())
			})
		})
	})

	Describe("DesireTask", func() {
		var (
			request  *http.Request
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
			request = newTestRequest(requestBody)
			handler.DesireTask(responseRecorder, request)
		})

		Context("when the desire is successful", func() {
			It("responds with 201 Status Created", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusCreated))
			})

			It("desires the task with the requested definitions", func() {
				Expect(fakeTaskDB.DesireTaskCallCount()).To(Equal(1))
				_, actualRequest := fakeTaskDB.DesireTaskArgsForCall(0)
				Expect(actualRequest).To(Equal(requestBody))
			})
		})

		Context("when request is invalid", func() {
			BeforeEach(func() {
				requestBody = &models.DesireTaskRequest{}
			})

			It("responds with 400 Bad Request", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("when parsing the body fails", func() {
			BeforeEach(func() {
				requestBody = "beep boop beep boop -- i am a robot"
			})

			It("responds with 400 Bad Request", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("when desiring the task fails", func() {
			BeforeEach(func() {
				fakeTaskDB.DesireTaskReturns(models.ErrUnknownError)
			})

			It("responds with 500 Internal Server Error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
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
				taskLogger, startTaskRequest := fakeTaskDB.StartTaskArgsForCall(0)
				Expect(taskLogger.SessionName()).To(ContainSubstring("start-task"))
				Expect(startTaskRequest.TaskGuid).To(Equal("task-guid"))
				Expect(startTaskRequest.CellId).To(Equal("cell-id"))
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
		Context("when the fail request is normal", func() {
			var expected *models.FailTaskRequest
			BeforeEach(func() {
				expected = &models.FailTaskRequest{
					TaskGuid:      "task-guid",
					FailureReason: "just cuz ;)",
				}
				request = newTestRequest(expected)
			})

			It("responds with 200 OK", func() {
				handler.FailTask(responseRecorder, request)
				Expect(responseRecorder.Code).To(Equal(http.StatusNoContent))
			})

			It("calls FailTask", func() {
				handler.FailTask(responseRecorder, request)
				Expect(fakeTaskDB.FailTaskCallCount()).To(Equal(1))
				taskLogger, failReq := fakeTaskDB.FailTaskArgsForCall(0)
				Expect(taskLogger.SessionName()).To(ContainSubstring("fail-task"))
				Expect(failReq).To(BeEquivalentTo(expected))
			})

			It("bubbles up the underlying model error", func() {
				fakeTaskDB.FailTaskReturns(models.ErrResourceExists)
				handler.FailTask(responseRecorder, request)
				res := &models.Error{}
				err := res.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal(models.ErrResourceExists))
			})
		})
		Context("when the request body is not a FailTaskRequest", func() {
			BeforeEach(func() {
				request = newTestRequest("foo")
			})

			It("responds with 400 BAD REQUEST", func() {
				handler.FailTask(responseRecorder, request)
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("returns an Invalid Request error", func() {
				handler.FailTask(responseRecorder, request)
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
				handler.FailTask(responseRecorder, request)
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("returns an Invalid Request error", func() {
				handler.FailTask(responseRecorder, request)
				var bbsError models.Error
				err := bbsError.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(bbsError.Equal(models.ErrBadRequest)).To(BeTrue())
				Expect(bbsError.Message).To(ContainSubstring("foobar"))
			})
		})
	})

	Describe("CompleteTask", func() {
		Context("when the request is normal", func() {
			var expected *models.CompleteTaskRequest
			BeforeEach(func() {
				expected = &models.CompleteTaskRequest{
					TaskGuid:      "task-guid",
					CellId:        "cell_id",
					Failed:        false,
					FailureReason: "just cuz ;)",
					Result:        "a result",
				}
				request = newTestRequest(expected)
			})

			It("responds with 200 OK", func() {
				handler.CompleteTask(responseRecorder, request)
				Expect(responseRecorder.Code).To(Equal(http.StatusNoContent))
			})

			It("calls CompleteTask", func() {
				handler.CompleteTask(responseRecorder, request)
				Expect(fakeTaskDB.CompleteTaskCallCount()).To(Equal(1))
				taskLogger, completeReq := fakeTaskDB.CompleteTaskArgsForCall(0)
				Expect(taskLogger.SessionName()).To(ContainSubstring("complete-task"))
				Expect(completeReq).To(BeEquivalentTo(expected))
			})

			It("bubbles up the underlying model error", func() {
				fakeTaskDB.CompleteTaskReturns(models.ErrResourceExists)
				handler.CompleteTask(responseRecorder, request)
				res := &models.Error{}
				err := res.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal(models.ErrResourceExists))
			})
		})

		Context("when the request body is not a CompleteTaskRequest", func() {
			BeforeEach(func() {
				request = newTestRequest("foo")
			})

			It("responds with 400 BAD REQUEST", func() {
				handler.CompleteTask(responseRecorder, request)
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("returns an Invalid Request error", func() {
				handler.CompleteTask(responseRecorder, request)
				var bbsError models.Error
				err := bbsError.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(bbsError.Equal(models.ErrBadRequest)).To(BeTrue())
				Expect(bbsError.Message).To(ContainSubstring("unmarshal"))
			})
		})

		Context("when the request body completes to stream", func() {
			BeforeEach(func() {
				request = newTestRequest(newExplodingReader(errors.New("foobar")))
			})

			It("responds with 400 BAD REQUEST", func() {
				handler.CompleteTask(responseRecorder, request)
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("returns an Invalid Request error", func() {
				handler.CompleteTask(responseRecorder, request)
				var bbsError models.Error
				err := bbsError.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(bbsError.Equal(models.ErrBadRequest)).To(BeTrue())
				Expect(bbsError.Message).To(ContainSubstring("foobar"))
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
})
