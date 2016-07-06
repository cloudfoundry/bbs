package main_test

import (
	"encoding/json"
	"fmt"
	"syscall"
	"time"

	"code.cloudfoundry.org/clock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"

	"code.cloudfoundry.org/bbs"
	bbsrunner "code.cloudfoundry.org/bbs/cmd/bbs/testrunner"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/models/test/model_helpers"
	"code.cloudfoundry.org/consuladapter"
	convergerrunner "code.cloudfoundry.org/converger/cmd/converger/testrunner"
	"code.cloudfoundry.org/locket"
)

type BinPaths struct {
	Converger string
	Bbs       string
}

var _ = Describe("Converger", func() {
	const exitDuration = 4 * time.Second

	var (
		bbsProcess       ifrit.Process
		bbsClient        bbs.InternalClient
		convergerProcess ifrit.Process
		runner           *ginkgomon.Runner

		consulClient consuladapter.Client
	)

	BeforeEach(func() {
		etcdRunner.Start()
		consulRunner.Start()
		consulRunner.WaitUntilReady()

		bbsProcess = ginkgomon.Invoke(bbsrunner.New(binPaths.Bbs, bbsArgs))
		bbsClient = bbs.NewClient(fmt.Sprint("http://", bbsArgs.Address))

		consulClient = consulRunner.NewClient()

		capacity := models.NewCellCapacity(512, 1024, 124)
		cellPresence := models.NewCellPresence("the-cell-id", "1.2.3.4", "the-zone", capacity, []string{}, []string{})

		value, err := json.Marshal(cellPresence)
		Expect(err).NotTo(HaveOccurred())

		presenceRunner := locket.NewPresence(logger, consulClient, bbs.CellSchemaPath(cellPresence.CellId), value, clock.NewClock(), locket.RetryInterval, locket.LockTTL)
		ifrit.Invoke(presenceRunner)
	})

	AfterEach(func() {
		ginkgomon.Kill(bbsProcess)
		ginkgomon.Kill(convergerProcess)
		consulRunner.Stop()
		etcdRunner.Stop()
	})

	startConverger := func() {
		runner = convergerrunner.New(convergerConfig)
		convergerProcess = ginkgomon.Invoke(runner)
		time.Sleep(convergeRepeatInterval)
	}

	createRunningTaskWithDeadCell := func() {
		task := model_helpers.NewValidTask("task-guid")

		err := bbsClient.DesireTask(logger, task.TaskGuid, task.Domain, task.TaskDefinition)
		Expect(err).NotTo(HaveOccurred())

		_, err = bbsClient.StartTask(logger, task.TaskGuid, "dead-cell")
		Expect(err).NotTo(HaveOccurred())
	}

	Context("when the converger has the lock", func() {
		Describe("when a task is desired but its cell is dead", func() {
			JustBeforeEach(createRunningTaskWithDeadCell)

			It("marks the task as completed and failed", func() {
				Consistently(func() []*models.Task {
					return getTasksByState(bbsClient, models.Task_Completed)
				}, 0.5).Should(BeEmpty())

				startConverger()

				Eventually(func() []*models.Task {
					return getTasksByState(bbsClient, models.Task_Completed)
				}, 10*convergeRepeatInterval).Should(HaveLen(1))
			})
		})
	})

	Context("when the converger loses the lock", func() {
		BeforeEach(func() {
			startConverger()
			Eventually(runner, 5*time.Second).Should(gbytes.Say("acquire-lock-succeeded"))

			_, err := consulClient.KV().DeleteTree(locket.LockSchemaPath("converge_lock"), nil)
			Expect(err).NotTo(HaveOccurred())
		})

		It("exits with an error", func() {
			Eventually(runner, exitDuration).Should(Exit(1))
		})
	})

	Context("when the converger initially does not have the lock", func() {
		var competingConvergerProcess ifrit.Process

		BeforeEach(func() {
			competingConvergerLock := locket.NewLock(logger, consulClient, locket.LockSchemaPath("converge_lock"), []byte{}, clock.NewClock(), 500*time.Millisecond, 10*time.Second)
			competingConvergerProcess = ifrit.Invoke(competingConvergerLock)

			startConverger()
		})

		Describe("when a task is desired but its cell is dead", func() {
			JustBeforeEach(createRunningTaskWithDeadCell)

			It("does not converge the task", func() {
				Consistently(func() []*models.Task {
					return getTasksByState(bbsClient, models.Task_Completed)
				}, 10*convergeRepeatInterval).Should(BeEmpty())
			})
		})

		Describe("when the lock becomes available", func() {
			BeforeEach(func() {
				ginkgomon.Kill(competingConvergerProcess)
				time.Sleep(convergeRepeatInterval + 10*time.Millisecond)
			})

			Describe("when a running task with a dead cell is present", func() {
				JustBeforeEach(createRunningTaskWithDeadCell)

				It("eventually marks the task as failed", func() {
					Eventually(func() []*models.Task {
						completedTasks := getTasksByState(bbsClient, models.Task_Completed)
						return failedTasks(completedTasks)
					}, 20*convergeRepeatInterval).Should(HaveLen(1))
				})
			})
		})
	})

	Describe("signal handling", func() {
		BeforeEach(func() {
			startConverger()
		})

		Describe("when it receives SIGINT", func() {
			It("exits successfully", func() {
				convergerProcess.Signal(syscall.SIGINT)
				Eventually(runner, exitDuration).Should(Exit(0))
			})
		})

		Describe("when it receives SIGTERM", func() {
			It("exits successfully", func() {
				convergerProcess.Signal(syscall.SIGTERM)
				Eventually(runner, exitDuration).Should(Exit(0))
			})
		})
	})

	Context("when etcd is down", func() {
		BeforeEach(func() {
			etcdRunner.Stop()
			startConverger()
		})

		It("starts", func() {
			Consistently(runner).ShouldNot(Exit())
		})
	})
})

func getTasksByState(client bbs.InternalClient, state models.Task_State) []*models.Task {
	tasks, err := client.Tasks(logger)
	Expect(err).NotTo(HaveOccurred())

	filteredTasks := make([]*models.Task, 0)
	for _, task := range tasks {
		if task.State == state {
			filteredTasks = append(filteredTasks, task)
		}
	}
	return filteredTasks
}

func failedTasks(tasks []*models.Task) []*models.Task {
	failedTasks := make([]*models.Task, 0)

	for _, task := range tasks {
		if task.Failed {
			failedTasks = append(failedTasks, task)
		}
	}

	return failedTasks
}
