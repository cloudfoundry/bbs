package etcd_test

import (
	"errors"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/models/internal/model_helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TaskDB", func() {
	Describe("Tasks", func() {
		Context("when there are tasks", func() {
			var expectedTasks []*models.Task

			BeforeEach(func() {
				expectedTasks = []*models.Task{
					model_helpers.NewValidTask("a-guid"), model_helpers.NewValidTask("b-guid"),
				}

				for _, t := range expectedTasks {
					etcdHelper.SetRawTask(t)
				}
			})

			It("returns all the tasks", func() {
				tasks, err := etcdDB.Tasks(logger, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(tasks.GetTasks()).To(ConsistOf(expectedTasks))
			})

			It("can filter", func() {
				tasks, err := etcdDB.Tasks(logger, func(t *models.Task) bool { return t.TaskGuid == "b-guid" })
				Expect(err).NotTo(HaveOccurred())
				Expect(tasks.Tasks).To(HaveLen(1))
				Expect(tasks.Tasks[0]).To(Equal(expectedTasks[1]))
			})
		})

		Context("when there are no tasks", func() {
			It("returns an empty list", func() {
				tasks, err := etcdDB.Tasks(logger, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(tasks).NotTo(BeNil())
				Expect(tasks.GetTasks()).To(BeEmpty())
			})
		})

		Context("when there is invalid data", func() {
			BeforeEach(func() {
				etcdHelper.CreateValidTask("some-guid")
				etcdHelper.CreateMalformedTask("some-other-guid")
				etcdHelper.CreateValidTask("some-third-guid")
			})

			It("errors", func() {
				_, err := etcdDB.Tasks(logger, nil)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when etcd is not there", func() {
			BeforeEach(func() {
				etcdRunner.Stop()
			})

			AfterEach(func() {
				etcdRunner.Start()
			})

			It("errors", func() {
				_, err := etcdDB.Tasks(logger, nil)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("TaskByGuid", func() {
		Context("when there is a task", func() {
			var expectedTask *models.Task

			BeforeEach(func() {
				expectedTask = model_helpers.NewValidTask("task-guid")
				etcdHelper.SetRawTask(expectedTask)
			})

			It("returns the task", func() {
				task, err := etcdDB.TaskByGuid(logger, "task-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(task).To(Equal(expectedTask))
			})
		})

		Context("when there is no task", func() {
			It("returns a ResourceNotFound", func() {
				_, err := etcdDB.TaskByGuid(logger, "nota-guid")
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})
		})

		Context("when there is invalid data", func() {
			BeforeEach(func() {
				etcdHelper.CreateMalformedTask("some-other-guid")
			})

			It("errors", func() {
				_, err := etcdDB.TaskByGuid(logger, "some-other-guid")
				Expect(err).To(Equal(models.ErrDeserializeJSON))
			})
		})

		Context("when etcd is not there", func() {
			BeforeEach(func() {
				etcdRunner.Stop()
			})

			AfterEach(func() {
				etcdRunner.Start()
			})

			It("errors", func() {
				_, err := etcdDB.TaskByGuid(logger, "some-other-guid")
				Expect(err).To(Equal(models.ErrUnknownError))
			})
		})
	})

	Describe("DesireTask", func() {
		const (
			taskGuid = "some-guid"
			domain   = "some-domain"
		)

		var (
			taskDef   *models.TaskDefinition
			errDesire *models.Error
		)

		JustBeforeEach(func() {
			request := models.DesireTaskRequest{
				Domain:         domain,
				TaskGuid:       taskGuid,
				TaskDefinition: taskDef,
			}
			errDesire = etcdDB.DesireTask(logger, &request)
		})

		BeforeEach(func() {
			taskDef = model_helpers.NewValidTaskDefinition()
		})

		Context("when a task is not already present at the desired key", func() {
			It("does not error", func() {
				Expect(errDesire).NotTo(HaveOccurred())
			})

			It("persists the task", func() {
				persistedTask, err := etcdDB.TaskByGuid(logger, taskGuid)
				Expect(err).NotTo(HaveOccurred())

				Expect(persistedTask.Domain).To(Equal(domain))
				Expect(*persistedTask.TaskDefinition).To(Equal(*taskDef))
			})

			It("provides a CreatedAt time", func() {
				persistedTask, err := etcdDB.TaskByGuid(logger, taskGuid)
				Expect(err).NotTo(HaveOccurred())
				Expect(persistedTask.CreatedAt).To(Equal(clock.Now().UnixNano()))
			})

			It("sets the UpdatedAt time", func() {
				persistedTask, err := etcdDB.TaskByGuid(logger, taskGuid)
				Expect(err).NotTo(HaveOccurred())
				Expect(persistedTask.UpdatedAt).To(Equal(clock.Now().UnixNano()))
			})

			Context("when able to fetch the Auctioneer address", func() {
				It("requests an auction", func() {
					Expect(auctioneerClient.RequestTaskAuctionsCallCount()).To(Equal(1))

					requestedTasks := auctioneerClient.RequestTaskAuctionsArgsForCall(0)
					Expect(requestedTasks).To(HaveLen(1))
					Expect(*requestedTasks[0].TaskDefinition).To(Equal(*taskDef))
				})

				Context("when requesting a task auction succeeds", func() {
					BeforeEach(func() {
						auctioneerClient.RequestTaskAuctionsReturns(nil)
					})

					It("does not return an error", func() {
						Expect(errDesire).NotTo(HaveOccurred())
					})
				})

				Context("when requesting a task auction fails", func() {
					BeforeEach(func() {
						auctioneerClient.RequestTaskAuctionsReturns(errors.New("oops"))
					})

					It("does not return an error", func() {
						// The creation succeeded, we can ignore the auction request error (converger will eventually do it)
						Expect(errDesire).NotTo(HaveOccurred())
					})
				})
			})
		})

		Context("when a task is already present at the desired key", func() {
			const otherDomain = "other-domain"

			BeforeEach(func() {
				request := models.DesireTaskRequest{
					Domain:         otherDomain,
					TaskGuid:       taskGuid,
					TaskDefinition: taskDef,
				}
				err := etcdDB.DesireTask(logger, &request)
				Expect(err).NotTo(HaveOccurred())
			})

			It("does not persist a second task", func() {
				tasks, err := etcdDB.Tasks(logger, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(tasks.GetTasks()).To(HaveLen(1))
				Expect(tasks.GetTasks()[0].Domain).To(Equal(otherDomain))
			})

			It("does not request a second auction", func() {
				Consistently(auctioneerClient.RequestTaskAuctionsCallCount).Should(Equal(1))
			})

			It("returns an error", func() {
				Expect(errDesire).To(Equal(models.ErrResourceExists))
			})
		})
	})

	Describe("StartTask", func() {
		var startRequest *models.StartTaskRequest
		var taskDef *models.TaskDefinition
		const taskGuid = "some-guid"
		const cellId = "cell-id"

		BeforeEach(func() {
			taskDef = model_helpers.NewValidTaskDefinition()
			startRequest = &models.StartTaskRequest{
				TaskGuid: taskGuid,
				CellId:   "cell-id",
			}
		})

		Context("when starting a pending Task", func() {
			BeforeEach(func() {
				desireRequest := models.DesireTaskRequest{
					Domain:         "domain",
					TaskGuid:       taskGuid,
					TaskDefinition: taskDef,
				}
				err := etcdDB.DesireTask(logger, &desireRequest)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns shouldStart as true", func() {
				started, err := etcdDB.StartTask(logger, startRequest)
				Expect(err).NotTo(HaveOccurred())
				Expect(started).To(BeTrue())
			})

			It("correctly updates the task record", func() {
				clock.IncrementBySeconds(1)

				_, err := etcdDB.StartTask(logger, startRequest)
				Expect(err).NotTo(HaveOccurred())

				task, err := etcdDB.TaskByGuid(logger, taskGuid)
				Expect(err).NotTo(HaveOccurred())

				Expect(task.TaskGuid).To(Equal(taskGuid))
				Expect(task.State).To(Equal(models.Task_Running))
				Expect(*task.TaskDefinition).To(Equal(*taskDef))
				Expect(task.UpdatedAt).To(Equal(clock.Now().UnixNano()))
			})
		})

		Context("When starting a Task that is already started", func() {
			BeforeEach(func() {
				request := models.DesireTaskRequest{
					Domain:         "domain",
					TaskGuid:       taskGuid,
					TaskDefinition: taskDef,
				}
				err := etcdDB.DesireTask(logger, &request)
				Expect(err).NotTo(HaveOccurred())

				_, err = etcdDB.StartTask(logger, startRequest)
				Expect(err).NotTo(HaveOccurred())
			})

			Context("on the same cell", func() {
				It("returns shouldStart as false", func() {
					changed, err := etcdDB.StartTask(logger, startRequest)
					Expect(err).NotTo(HaveOccurred())
					Expect(changed).To(BeFalse())
				})

				It("does not change the Task in the store", func() {
					previousTime := clock.Now().UnixNano()
					clock.IncrementBySeconds(1)

					_, err := etcdDB.StartTask(logger, startRequest)
					Expect(err).NotTo(HaveOccurred())

					task, err := etcdDB.TaskByGuid(logger, taskGuid)
					Expect(err).NotTo(HaveOccurred())

					Expect(task.UpdatedAt).To(Equal(previousTime))
				})
			})

			Context("on another cell", func() {
				It("returns an error", func() {
					startRequest.CellId = "some-other-cell"
					_, err := etcdDB.StartTask(logger, startRequest)
					Expect(err).NotTo(BeNil())
					Expect(err.Type).To(Equal(models.InvalidStateTransition))
				})

				It("does not change the Task in the store", func() {
					previousTime := clock.Now().UnixNano()
					clock.IncrementBySeconds(1)

					_, err := etcdDB.StartTask(logger, startRequest)
					Expect(err).NotTo(HaveOccurred())

					task, err := etcdDB.TaskByGuid(logger, taskGuid)
					Expect(err).NotTo(HaveOccurred())

					Expect(task.UpdatedAt).To(Equal(previousTime))
				})
			})
		})
	})
})
