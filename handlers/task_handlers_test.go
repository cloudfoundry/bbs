package handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/cloudfoundry-incubator/bbs/db/fakes"
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
		JustBeforeEach(func() {
			handler.DesireTask(responseRecorder, request)
		})

		Context("when the desire is successful", func() {
			BeforeEach(func() {
				request = newTestRequest(&models.TaskDefinition{})
			})

			It("responds with 201 CREATED", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusCreated))
			})
		})

		Context("when the request body is not a TaskDefinition", func() {
			BeforeEach(func() {
				request = newTestRequest("foo")
			})

			It("responds with 400 BAD REQUEST", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("returns an Invalid Request error", func() {
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
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("returns an Invalid Request error", func() {
				var bbsError models.Error
				err := bbsError.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(bbsError.Equal(models.ErrBadRequest)).To(BeTrue())
				Expect(bbsError.Message).To(ContainSubstring("foobar"))
			})
		})
	})
})
