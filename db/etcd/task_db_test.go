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
		var errDesire error
		var taskDef *models.TaskDefinition
		const taskGuid = "some-guid"
		const domain = "some-domain"

		JustBeforeEach(func() {
			errDesire = etcdDB.DesireTask(logger, taskGuid, domain, taskDef)
		})

		Context("when given a valid task", func() {
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
					err := etcdDB.DesireTask(logger, taskGuid, otherDomain, taskDef)
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

		Context("when given an invalid task", func() {
			BeforeEach(func() {
				taskDef = &models.TaskDefinition{}
			})

			It("does not persist a task", func() {
				Consistently(func() ([]*models.Task, error) {
					tasks, err := etcdDB.Tasks(logger, nil)
					return tasks.GetTasks(), err
				}).Should(BeEmpty())
			})

			It("does not request an auction", func() {
				Consistently(auctioneerClient.RequestTaskAuctionsCallCount).Should(BeZero())
			})

			It("returns an error", func() {
				Expect(errDesire).To(ContainElement(models.ErrInvalidField{"rootfs"}))
			})
		})
	})
})
