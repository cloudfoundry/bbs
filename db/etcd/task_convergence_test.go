package etcd_test

import (
	"errors"
	"time"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/models/test/model_helpers"
	"github.com/cloudfoundry/dropsonde/metric_sender/fake"
	"github.com/cloudfoundry/dropsonde/metrics"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const TTL = 0

var _ = Describe("Convergence of Tasks", func() {
	var (
		sender *fake.FakeMetricSender

		kickTasksDurationInSeconds, expirePendingTaskDurationInSeconds            uint64
		kickTasksDuration, expirePendingTaskDuration, expireCompletedTaskDuration time.Duration
	)

	BeforeEach(func() {
		sender = fake.NewFakeMetricSender()
		metrics.Initialize(sender, nil)

		kickTasksDurationInSeconds = 10
		kickTasksDuration = time.Duration(kickTasksDurationInSeconds) * time.Second
		expirePendingTaskDurationInSeconds = 30
		expirePendingTaskDuration = time.Duration(expirePendingTaskDurationInSeconds) * time.Second
		expireCompletedTaskDuration = time.Hour
	})

	Describe("ConvergeTasks", func() {
		const (
			taskGuid  = "some-guid"
			taskGuid2 = "some-other-guid"
			domain    = "some-domain"
			cellId    = "cell-id"
		)
		JustBeforeEach(func() {
			etcdDB.ConvergeTasks(logger, kickTasksDuration, expirePendingTaskDuration, expireCompletedTaskDuration)
		})

		It("bumps the convergence counter", func() {
			Expect(sender.GetCounter("ConvergenceTaskRuns")).To(Equal(uint64(1)))
		})

		It("reports the duration that it took to converge", func() {
			reportedDuration := sender.GetValue("ConvergenceTaskDuration")
			Expect(reportedDuration.Unit).To(Equal("nanos"))
			Expect(reportedDuration.Value).NotTo(BeZero())
		})

		Context("when a Task is malformed", func() {
			BeforeEach(func() {
				etcdHelper.CreateMalformedTask(taskGuid)

			})

			It("should delete it", func() {
				_, modelErr := etcdDB.TaskByGuid(logger, taskGuid)
				Expect(modelErr).To(BeEquivalentTo(models.ErrResourceNotFound))
			})

			It("bumps the pruned counter", func() {
				Expect(sender.GetCounter("ConvergenceTasksPruned")).To(Equal(uint64(1)))
			})
		})

		Context("when Tasks are pending", func() {
			BeforeEach(func() {
				expectedTasks := []*models.Task{
					model_helpers.NewValidTask(taskGuid), model_helpers.NewValidTask(taskGuid2),
				}

				for _, t := range expectedTasks {
					t.CreatedAt = clock.Now().UnixNano()
					t.UpdatedAt = clock.Now().UnixNano()
					t.FirstCompletedAt = 0
					etcdHelper.SetRawTask(t)
				}
			})

			Context("when the Task has NOT been pending for too long", func() {
				BeforeEach(func() {
					clock.IncrementBySeconds(kickTasksDurationInSeconds - 1)
				})

				It("does not request an auction for the task", func() {
					Consistently(auctioneerClient.RequestTaskAuctionsCallCount).Should(Equal(0))
				})
			})

			Context("when the Tasks have been pending for longer than the kick interval", func() {
				BeforeEach(func() {
					clock.IncrementBySeconds(kickTasksDurationInSeconds + 1)
				})

				It("bumps the compare-and-swap counter", func() {
					Expect(sender.GetCounter("ConvergenceTasksKicked")).To(Equal(uint64(2)))
				})

				It("logs that it sends an auction for the pending task", func() {
					Expect(logger.TestSink.LogMessages()).To(ContainElement("test.converge-tasks.requesting-auction-for-pending-task"))
				})

				Context("when able to fetch the auctioneer address", func() {

					It("requests an auction", func() {
						Expect(auctioneerClient.RequestTaskAuctionsCallCount()).To(Equal(1))

						requestedTasks := auctioneerClient.RequestTaskAuctionsArgsForCall(0)
						Expect(requestedTasks).To(HaveLen(2))
						Expect([]string{requestedTasks[0].TaskGuid, requestedTasks[1].TaskGuid}).To(ConsistOf(taskGuid, taskGuid2))
					})

					Context("when requesting an auction is unsuccessful", func() {
						BeforeEach(func() {
							auctioneerClient.RequestTaskAuctionsReturns(errors.New("oops"))
						})

						It("logs an error", func() {
							Expect(logger.TestSink.LogMessages()).To(ContainElement("test.converge-tasks.failed-to-request-auctions-for-pending-tasks"))
						})
					})
				})
			})

			Context("when the Task has been pending for longer than the expirePendingTasksDuration", func() {
				BeforeEach(func() {
					clock.IncrementBySeconds(expirePendingTaskDurationInSeconds + 1)
				})

				It("should mark the Task as completed & failed", func() {
					returnedTask, err := etcdDB.TaskByGuid(logger, taskGuid)
					Expect(err).NotTo(HaveOccurred())
					Expect(returnedTask.State).To(Equal(models.Task_Completed))

					Expect(returnedTask.Failed).To(Equal(true))
					Expect(returnedTask.FailureReason).To(ContainSubstring("time limit"))
				})

				It("bumps the compare-and-swap counter", func() {
					Expect(sender.GetCounter("ConvergenceTasksKicked")).To(Equal(uint64(2)))
				})

				It("logs an error", func() {
					Expect(logger.TestSink.LogMessages()).To(ContainElement("test.converge-tasks.failed-to-start-in-time"))
				})
			})
		})

		Context("when a Task is running", func() {
			BeforeEach(func() {
				err := etcdDB.DesireTask(logger, model_helpers.NewValidTaskDefinition(), taskGuid, domain)
				Expect(err).NotTo(HaveOccurred())

				_, err = etcdDB.StartTask(logger, taskGuid, "cell-id")
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when the associated cell is present", func() {
				BeforeEach(func() {
					cellPresence := models.NewCellPresence("cell-id", "1.2.3.4", "the-zone", models.NewCellCapacity(128, 1024, 3), []string{}, []string{})
					registerCell(cellPresence)
				})

				It("leaves the task running", func() {
					returnedTask, err := etcdDB.TaskByGuid(logger, taskGuid)
					Expect(err).NotTo(HaveOccurred())
					Expect(returnedTask.State).To(Equal(models.Task_Running))
				})
			})

			Context("when the associated cell is missing", func() {
				It("should mark the Task as completed & failed", func() {
					returnedTask, err := etcdDB.TaskByGuid(logger, taskGuid)
					Expect(err).NotTo(HaveOccurred())
					Expect(returnedTask.State).To(Equal(models.Task_Completed))

					Expect(returnedTask.Failed).To(Equal(true))
					Expect(returnedTask.FailureReason).To(ContainSubstring("cell"))
				})

				It("logs that the cell disappeared", func() {
					Expect(logger.TestSink.LogMessages()).To(ContainElement("test.converge-tasks.cell-disappeared"))
				})

				It("bumps the compare-and-swap counter", func() {
					Expect(sender.GetCounter("ConvergenceTasksKicked")).To(Equal(uint64(1)))
				})
			})
		})

		Describe("Completed tasks", func() {
			Context("when Tasks with a complete URL are completed", func() {
				BeforeEach(func() {
					taskDef := model_helpers.NewValidTaskDefinition()
					taskDef.CompletionCallbackUrl = "blah"

					err := etcdDB.DesireTask(logger, taskDef, taskGuid, domain)
					Expect(err).NotTo(HaveOccurred())

					_, err = etcdDB.StartTask(logger, taskGuid, cellId)
					Expect(err).NotTo(HaveOccurred())

					err = etcdDB.CompleteTask(logger, taskGuid, cellId, true, "'cause I said so", "a magical result")
					Expect(err).NotTo(HaveOccurred())

					err = etcdDB.DesireTask(logger, taskDef, taskGuid2, domain)

					_, err = etcdDB.StartTask(logger, taskGuid2, cellId)
					Expect(err).NotTo(HaveOccurred())

					err = etcdDB.CompleteTask(logger, taskGuid2, cellId, true, "'cause I said so", "a magical result")
					Expect(err).NotTo(HaveOccurred())
				})

				Context("for longer than the convergence interval", func() {
					BeforeEach(func() {
						clock.IncrementBySeconds(expirePendingTaskDurationInSeconds + 1)
					})

					It("resubmits the completed tasks to the callback workpool", func() {
						expectedCallCount := 4
						Expect(fakeTaskCBFactory.TaskCallbackWorkCallCount()).To(Equal(expectedCallCount)) // 2 initial completes + 2 times for convergence

						task1Completions := 0
						task2Completions := 0
						for i := 0; i < expectedCallCount; i++ {
							_, db, task := fakeTaskCBFactory.TaskCallbackWorkArgsForCall(i)
							Expect(db).To(Equal(etcdDB))
							if task.TaskGuid == taskGuid {
								task1Completions++
							} else if task.TaskGuid == taskGuid2 {
								task2Completions++
							}
							Expect(task.Failed).To(BeTrue())
							Expect(task.FailureReason).To(Equal("'cause I said so"))
							Expect(task.Result).To(Equal("a magical result"))
						}

						Expect(task1Completions).To(Equal(2))
						Expect(task2Completions).To(Equal(2))
					})

					It("logs that it kicks the completed task", func() {
						Expect(logger.TestSink.LogMessages()).To(ContainElement("test.converge-tasks.kicking-completed-task"))
					})

					It("bumps the convergence tasks kicked counter", func() {
						Expect(sender.GetCounter("ConvergenceTasksKicked")).To(Equal(uint64(2)))
					})
				})

				Context("when the task has been completed for longer than the time-to-resolve interval", func() {
					BeforeEach(func() {
						clock.IncrementBySeconds(uint64(expireCompletedTaskDuration.Seconds()) + 1)
					})

					It("should delete the task", func() {
						_, modelErr := etcdDB.TaskByGuid(logger, taskGuid)
						Expect(modelErr).To(Equal(models.ErrResourceNotFound))
					})

					It("logs that it failed to start resolving the task in time", func() {
						Expect(logger.TestSink.LogMessages()).To(ContainElement("test.converge-tasks.failed-to-start-resolving-in-time"))
					})
				})

				Context("when the task has been completed for less than the convergence interval", func() {
					var previousTime int64

					BeforeEach(func() {
						previousTime = clock.Now().UnixNano()
						clock.IncrementBySeconds(1)
					})

					It("should NOT kick the Task", func() {
						returnedTask, err := etcdDB.TaskByGuid(logger, taskGuid)
						Expect(err).NotTo(HaveOccurred())
						Expect(returnedTask.State).To(Equal(models.Task_Completed))
						Expect(returnedTask.UpdatedAt).To(Equal(previousTime))
					})
				})
			})
		})

		Context("when a Task is resolving", func() {
			BeforeEach(func() {
				taskDef := model_helpers.NewValidTaskDefinition()
				taskDef.CompletionCallbackUrl = "blah"

				err := etcdDB.DesireTask(logger, taskDef, taskGuid, domain)
				Expect(err).NotTo(HaveOccurred())

				_, err = etcdDB.StartTask(logger, taskGuid, cellId)
				Expect(err).NotTo(HaveOccurred())

				err = etcdDB.CompleteTask(logger, taskGuid, cellId, true, "'cause I said so", "a magical result")
				Expect(err).NotTo(HaveOccurred())

				err = etcdDB.ResolvingTask(logger, taskGuid)
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when the task is in resolving state for less than the convergence interval", func() {
				var previousTime int64

				BeforeEach(func() {
					previousTime = clock.Now().UnixNano()
					clock.IncrementBySeconds(1)
				})

				It("should do nothing", func() {
					Expect(fakeTaskCBFactory.TaskCallbackWorkCallCount()).To(Equal(1))

					returnedTask, err := etcdDB.TaskByGuid(logger, taskGuid)
					Expect(err).NotTo(HaveOccurred())
					Expect(returnedTask.State).To(Equal(models.Task_Resolving))
					Expect(returnedTask.UpdatedAt).To(Equal(previousTime))
				})
			})

			Context("when the task has been resolving for longer than a convergence interval", func() {
				BeforeEach(func() {
					clock.IncrementBySeconds(expirePendingTaskDurationInSeconds)
				})

				It("should put the Task back into the completed state", func() {
					returnedTask, err := etcdDB.TaskByGuid(logger, taskGuid)
					Expect(err).NotTo(HaveOccurred())
					Expect(returnedTask.State).To(Equal(models.Task_Completed))
					Expect(returnedTask.UpdatedAt).To(Equal(clock.Now().UnixNano()))
				})

				It("logs that it is demoting task from resolving to completed", func() {
					Expect(logger.TestSink.LogMessages()).To(ContainElement("test.converge-tasks.demoting-resolving-to-completed"))
				})

				It("submits the completed task to the workpool", func() {
					Expect(fakeTaskCBFactory.TaskCallbackWorkCallCount()).To(Equal(2))

					_, _, task := fakeTaskCBFactory.TaskCallbackWorkArgsForCall(1)
					Expect(task.TaskGuid).To(Equal(taskGuid))
				})

				It("bumps the compare-and-swap counter", func() {
					Expect(sender.GetCounter("ConvergenceTasksKicked")).To(Equal(uint64(1)))
				})
			})

			Context("when the resolving task has been completed for longer than the time-to-resolve interval", func() {
				BeforeEach(func() {
					clock.IncrementBySeconds(uint64(expireCompletedTaskDuration.Seconds()) + 1)
				})

				It("should delete the task", func() {
					_, err := etcdDB.TaskByGuid(logger, taskGuid)
					Expect(err).To(Equal(models.ErrResourceNotFound))
				})

				It("logs that has failed to resolve task in time", func() {
					Expect(logger.TestSink.LogMessages()).To(ContainElement("test.converge-tasks.failed-to-resolve-in-time"))
				})
			})
		})
	})
})
