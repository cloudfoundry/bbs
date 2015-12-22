package handlers_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry-incubator/bbs/db/fakes"
	"github.com/cloudfoundry-incubator/bbs/format"
	"github.com/cloudfoundry-incubator/bbs/handlers"
	"github.com/cloudfoundry-incubator/bbs/models"
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
	)

	BeforeEach(func() {
		fakeTaskDB = new(fakes.FakeTaskDB)
		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		responseRecorder = httptest.NewRecorder()
		handler = handlers.NewTaskHandler(logger, fakeTaskDB)
	})

	Describe("Tasks_r0", func() {
		BeforeEach(func() {
			task1 = models.Task{Domain: "domain-1"}
			task2 = models.Task{CellId: "cell-id"}
			requestBody = &models.TasksRequest{}
		})

		JustBeforeEach(func() {
			request := newTestRequest(requestBody)
			handler.Tasks_r0(responseRecorder, request)
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

			Context("and the returned tasks have cache dependencies", func() {
				BeforeEach(func() {
					task1.TaskDefinition = &models.TaskDefinition{}
					task2.TaskDefinition = &models.TaskDefinition{}

					task1.Action = &models.Action{
						UploadAction: &models.UploadAction{
							From: "web_location",
						},
					}

					task1.CachedDependencies = []*models.CachedDependency{
						{Name: "name-1", From: "from-1", To: "to-1", CacheKey: "cache-key-1", LogSource: "log-source-1"},
					}

					task2.CachedDependencies = []*models.CachedDependency{
						{Name: "name-2", From: "from-2", To: "to-2", CacheKey: "cache-key-2", LogSource: "log-source-2"},
						{Name: "name-3", From: "from-3", To: "to-3", CacheKey: "cache-key-3", LogSource: "log-source-3"},
					}
				})

				It("translates the cache dependencies into download actions", func() {
					Expect(responseRecorder.Code).To(Equal(http.StatusOK))
					response := models.TasksResponse{}
					err := response.Unmarshal(responseRecorder.Body.Bytes())
					Expect(err).NotTo(HaveOccurred())

					Expect(response.Error).To(BeNil())
					Expect(response.Tasks).To(HaveLen(2))
					Expect(response.Tasks[0]).To(Equal(task1.VersionDownTo(format.V0)))
					Expect(response.Tasks[1]).To(Equal(task2.VersionDownTo(format.V0)))
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

	Describe("TaskByGuid_r0", func() {
		var taskGuid = "task-guid"

		BeforeEach(func() {
			requestBody = &models.TaskByGuidRequest{
				TaskGuid: taskGuid,
			}
		})

		JustBeforeEach(func() {
			request := newTestRequest(requestBody)
			handler.TaskByGuid_r0(responseRecorder, request)
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

			Context("when the task has cache dependencies", func() {
				BeforeEach(func() {
					task.TaskDefinition = &models.TaskDefinition{}
					task.CachedDependencies = []*models.CachedDependency{
						{Name: "name-2", From: "from-2", To: "to-2", CacheKey: "cache-key-2", LogSource: "log-source-2"},
						{Name: "name-3", From: "from-3", To: "to-3", CacheKey: "cache-key-3", LogSource: "log-source-3"},
					}
				})

				It("moves them to the actions", func() {
					Expect(responseRecorder.Code).To(Equal(http.StatusOK))
					response := models.TaskResponse{}
					err := response.Unmarshal(responseRecorder.Body.Bytes())
					Expect(err).NotTo(HaveOccurred())

					Expect(response.Error).To(BeNil())
					Expect(response.Task).To(Equal(task.VersionDownTo(format.V0)))
				})
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

})
