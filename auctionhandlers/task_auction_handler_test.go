package auctionhandlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	fake_auction_runner "github.com/cloudfoundry-incubator/auction/auctiontypes/fakes"
	"github.com/cloudfoundry-incubator/bbs/auctionhandlers"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/models/test/model_helpers"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("TaskAuctionHandler", func() {
	var (
		logger           *lagertest.TestLogger
		runner           *fake_auction_runner.FakeAuctionRunner
		responseRecorder *httptest.ResponseRecorder
		handler          *auctionhandlers.TaskAuctionHandler
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		runner = new(fake_auction_runner.FakeAuctionRunner)
		responseRecorder = httptest.NewRecorder()
		handler = auctionhandlers.NewTaskAuctionHandler(runner)
	})

	Describe("Create", func() {
		Context("when the request body is a task", func() {
			var tasks []*models.Task

			BeforeEach(func() {
				tasks = []*models.Task{
					model_helpers.NewValidTask("the-task-guid"),
				}

				handler.Create(responseRecorder, newTestRequest(tasks), logger)
			})

			It("responds with 202", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusAccepted))
			})

			It("responds with an empty JSON body", func() {
				Expect(responseRecorder.Body.String()).To(Equal("{}"))
			})

			It("should submit the task to the auction runner", func() {
				Expect(runner.ScheduleTasksForAuctionsCallCount()).To(Equal(1))

				submittedTasks := runner.ScheduleTasksForAuctionsArgsForCall(0)
				Expect(submittedTasks).To(Equal(tasks))
			})
		})

		Context("when the request body is a not a valid task", func() {
			var tasks []*models.Task

			BeforeEach(func() {
				tasks = []*models.Task{{TaskGuid: "the-task-guid"}}

				handler.Create(responseRecorder, newTestRequest(tasks), logger)
			})

			It("responds with 202", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusAccepted))
			})

			It("logs an error", func() {
				Expect(logger).To(Say("test.task-auction-handler.create.task-validate-failed"))
			})

			It("should submit the task to the auction runner", func() {
				Expect(runner.ScheduleTasksForAuctionsCallCount()).To(Equal(1))

				submittedTasks := runner.ScheduleTasksForAuctionsArgsForCall(0)
				Expect(submittedTasks).To(BeEmpty())
			})
		})

		Context("when the request body is a not a task", func() {
			BeforeEach(func() {
				handler.Create(responseRecorder, newTestRequest(`{invalidjson}`), logger)
			})

			It("responds with 400", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("responds with a JSON body containing the error", func() {
				handlerError := auctionhandlers.HandlerError{}
				err := json.NewDecoder(responseRecorder.Body).Decode(&handlerError)
				Expect(err).NotTo(HaveOccurred())
				Expect(handlerError.Error).NotTo(BeEmpty())
			})

			It("should not submit the task to the auction runner", func() {
				Expect(runner.ScheduleTasksForAuctionsCallCount()).To(Equal(0))
			})
		})

		Context("when the request body returns a non-EOF error on read", func() {
			BeforeEach(func() {
				req := newTestRequest("")
				req.Body = badReader{}
				handler.Create(responseRecorder, req, logger)
			})

			It("responds with 500", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})

			It("responds with a JSON body containing the error", func() {
				handlerError := auctionhandlers.HandlerError{}
				err := json.NewDecoder(responseRecorder.Body).Decode(&handlerError)
				Expect(err).NotTo(HaveOccurred())
				Expect(handlerError.Error).To(Equal(ErrBadRead.Error()))
			})

			It("should not submit the task auction to the auction runner", func() {
				Expect(runner.ScheduleTasksForAuctionsCallCount()).To(Equal(0))
			})
		})
	})
})
