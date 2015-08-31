package taskworkpool_test

import (
	"net/http"
	"time"

	dbFakes "github.com/cloudfoundry-incubator/bbs/db/fakes"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/models/test/model_helpers"
	"github.com/cloudfoundry-incubator/bbs/taskworkpool"
	"github.com/cloudfoundry-incubator/cf_http"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("TaskWorker", func() {
	var (
		fakeServer *ghttp.Server
		logger     *lagertest.TestLogger
		timeout    time.Duration
	)

	BeforeEach(func() {
		timeout = 1 * time.Second
		cf_http.Initialize(timeout)
		fakeServer = ghttp.NewServer()

		logger = lagertest.NewTestLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.INFO))
	})

	AfterEach(func() {
		fakeServer.Close()
	})

	Describe("HandleCompletedTask", func() {
		var (
			callbackURL string
			taskDB      *dbFakes.FakeTaskDB
			statusCodes chan int
			reqCount    chan struct{}
		)

		BeforeEach(func() {
			statusCodes = make(chan int)
			reqCount = make(chan struct{})

			fakeServer.RouteToHandler("POST", "/the-callback/url", func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(<-statusCodes)
			})

			callbackURL = fakeServer.URL() + "/the-callback/url"
			taskDB = new(dbFakes.FakeTaskDB)
			taskDB.ResolvingTaskReturns(nil)
			taskDB.DeleteTaskReturns(nil)
		})

		simulateTaskCompleting := func() {
			task := model_helpers.NewValidTask("the-task-guid")
			task.CompletionCallbackUrl = callbackURL
			taskworkpool.HandleCompletedTask(logger, taskDB, task)
		}

		Context("when the task has a completion callback URL", func() {
			It("marks the task as resolving", func() {
				Expect(taskDB.ResolvingTaskCallCount()).To(Equal(0))

				go simulateTaskCompleting()
				statusCodes <- 200

				Eventually(taskDB.ResolvingTaskCallCount).Should(Equal(1))
				_, actualGuid := taskDB.ResolvingTaskArgsForCall(0)
				Expect(actualGuid).To(Equal("the-task-guid"))
			})

			Context("when marking the task as resolving fails", func() {
				BeforeEach(func() {
					taskDB.ResolvingTaskReturns(models.NewError(models.Error_UnknownError, "failed to resolve task"))
				})

				It("does not make a request to the task's callback URL", func() {
					go simulateTaskCompleting()

					Consistently(fakeServer.ReceivedRequests, 0.25).Should(BeEmpty())
				})
			})

			Context("when marking the task as resolving succeeds", func() {
				It("POSTs to the task's callback URL", func() {
					go simulateTaskCompleting()

					statusCodes <- 200

					Eventually(fakeServer.ReceivedRequests).Should(HaveLen(1))
				})

				Context("when the request succeeds", func() {
					It("resolves the task", func() {
						go simulateTaskCompleting()

						statusCodes <- 200

						Eventually(taskDB.DeleteTaskCallCount).Should(Equal(1))
						_, actualGuid := taskDB.DeleteTaskArgsForCall(0)
						Expect(actualGuid).To(Equal("the-task-guid"))
					})
				})

				Context("when the request fails with a 4xx response code", func() {
					It("resolves the task", func() {
						go simulateTaskCompleting()

						statusCodes <- 403

						Eventually(taskDB.DeleteTaskCallCount).Should(Equal(1))
						_, actualGuid := taskDB.DeleteTaskArgsForCall(0)
						Expect(actualGuid).To(Equal("the-task-guid"))
					})
				})

				Context("when the request fails with a 500 response code", func() {
					It("resolves the task", func() {
						go simulateTaskCompleting()

						statusCodes <- 500

						Eventually(taskDB.DeleteTaskCallCount).Should(Equal(1))
						_, actualGuid := taskDB.DeleteTaskArgsForCall(0)
						Expect(actualGuid).To(Equal("the-task-guid"))
					})
				})

				Context("when the request fails with a 503 or 504 response code", func() {
					It("retries the request 2 more times", func() {
						go simulateTaskCompleting()
						Eventually(fakeServer.ReceivedRequests).Should(HaveLen(1))

						statusCodes <- 503

						Consistently(taskDB.DeleteTaskCallCount, 0.25).Should(Equal(0))
						Eventually(fakeServer.ReceivedRequests).Should(HaveLen(2))

						statusCodes <- 504

						Consistently(taskDB.DeleteTaskCallCount, 0.25).Should(Equal(0))
						Eventually(fakeServer.ReceivedRequests).Should(HaveLen(3))

						statusCodes <- 200

						Eventually(taskDB.DeleteTaskCallCount, 0.25).Should(Equal(1))
						_, actualGuid := taskDB.DeleteTaskArgsForCall(0)
						Expect(actualGuid).To(Equal("the-task-guid"))
					})

					Context("when the request fails every time", func() {
						It("does not resolve the task", func() {
							go simulateTaskCompleting()
							Eventually(fakeServer.ReceivedRequests).Should(HaveLen(1))

							statusCodes <- 503

							Consistently(taskDB.DeleteTaskCallCount, 0.25).Should(Equal(0))
							Eventually(fakeServer.ReceivedRequests).Should(HaveLen(2))

							statusCodes <- 504

							Consistently(taskDB.DeleteTaskCallCount, 0.25).Should(Equal(0))
							Eventually(fakeServer.ReceivedRequests).Should(HaveLen(3))

							statusCodes <- 503

							Consistently(taskDB.DeleteTaskCallCount, 0.25).Should(Equal(0))
							Consistently(fakeServer.ReceivedRequests, 0.25).Should(HaveLen(3))
						})
					})
				})

				Context("when DeleteTask fails", func() {
					It("logs an error and returns", func() {
						taskDB.DeleteTaskReturns(&models.Error{})
						go simulateTaskCompleting()
						Eventually(fakeServer.ReceivedRequests).Should(HaveLen(1))
						statusCodes <- 200

						Eventually(taskDB.DeleteTaskCallCount).Should(Equal(1))
						Expect(logger.TestSink.LogMessages()).To(ContainElement("test.delete-task-failed"))
					})
				})

				Context("when the request fails with a timeout", func() {
					var sleepCh chan time.Duration

					BeforeEach(func() {
						sleepCh = make(chan time.Duration)
						fakeServer.RouteToHandler("POST", "/the-callback/url", func(w http.ResponseWriter, req *http.Request) {
							time.Sleep(<-sleepCh)
							w.WriteHeader(200)
						})
					})

					It("retries the request 2 more times", func() {
						go simulateTaskCompleting()

						sleepCh <- timeout + 100*time.Millisecond
						Eventually(fakeServer.ReceivedRequests).Should(HaveLen(1))

						sleepCh <- timeout + 100*time.Millisecond
						Consistently(taskDB.DeleteTaskCallCount, 0.25).Should(Equal(0))
						Eventually(fakeServer.ReceivedRequests).Should(HaveLen(2))

						sleepCh <- timeout + 100*time.Millisecond
						Consistently(taskDB.DeleteTaskCallCount, 0.25).Should(Equal(0))
						Eventually(fakeServer.ReceivedRequests).Should(HaveLen(3))

						Eventually(taskDB.DeleteTaskCallCount, 0.25).Should(Equal(0))
					})

					Context("when the request fails with timeout once and then succeeds", func() {
						It("deletes the task", func() {
							go simulateTaskCompleting()
							sleepCh <- (timeout + 100*time.Millisecond)

							Eventually(fakeServer.ReceivedRequests).Should(HaveLen(1))
							Consistently(taskDB.DeleteTaskCallCount, 0.25).Should(Equal(0))

							sleepCh <- 0
							Eventually(fakeServer.ReceivedRequests).Should(HaveLen(2))
							Eventually(taskDB.DeleteTaskCallCount, 0.25).Should(Equal(1))

							_, resolvedTaskGuid := taskDB.DeleteTaskArgsForCall(0)
							Expect(resolvedTaskGuid).To(Equal("the-task-guid"))
						})
					})
				})
			})
		})

		Context("when the task doesn't have a completion callback URL", func() {
			BeforeEach(func() {
				callbackURL = ""
				go simulateTaskCompleting()
			})

			It("does not mark the task as resolving", func() {
				Consistently(taskDB.ResolvingTaskCallCount).Should(Equal(0))
			})
		})
	})
})
