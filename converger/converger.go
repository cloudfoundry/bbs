package converger

import (
	"os"
	"sync"
	"time"

	"code.cloudfoundry.org/lager"
	"github.com/nu7hatch/gouuid"

	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/clock"
)

type Converger struct {
	id                          string
	serviceClient               bbs.ServiceClient
	lrpConvergenceController    LrpConvergenceController
	taskController              TaskController
	logger                      lager.Logger
	clock                       clock.Clock
	convergeRepeatInterval      time.Duration
	kickTaskDuration            time.Duration
	expirePendingTaskDuration   time.Duration
	expireCompletedTaskDuration time.Duration
	closeOnce                   *sync.Once
}

func New(
	logger lager.Logger,
	clock clock.Clock,
	lrpConvergenceController LrpConvergenceController,
	taskController TaskController,
	serviceClient bbs.ServiceClient,
	convergeRepeatInterval,
	kickTaskDuration,
	expirePendingTaskDuration,
	expireCompletedTaskDuration time.Duration,
) *Converger {

	uuid, err := uuid.NewV4()
	if err != nil {
		panic("Failed to generate a random guid....:" + err.Error())
	}

	return &Converger{
		id:                          uuid.String(),
		logger:                      logger,
		clock:                       clock,
		serviceClient:               serviceClient,
		lrpConvergenceController:    lrpConvergenceController,
		taskController:              taskController,
		convergeRepeatInterval:      convergeRepeatInterval,
		kickTaskDuration:            kickTaskDuration,
		expirePendingTaskDuration:   expirePendingTaskDuration,
		expireCompletedTaskDuration: expireCompletedTaskDuration,
		closeOnce:                   &sync.Once{},
	}
}

func (c *Converger) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	logger := c.logger.Session("converger-process")
	logger.Info("started")

	convergeTimer := c.clock.NewTimer(c.convergeRepeatInterval)
	defer func() {
		logger.Info("done")
		convergeTimer.Stop()
	}()

	cellEvents := c.serviceClient.CellEvents(logger)

	close(ready)

	for {
		select {
		case <-signals:
			return nil

		case event := <-cellEvents:
			switch event.EventType() {
			case models.EventTypeCellDisappeared:
				logger.Info("received-cell-disappeared-event", lager.Data{"cell-id": event.CellIDs()})
				c.converge()
			}

		case <-convergeTimer.C():
			c.converge()
		}

		convergeTimer.Reset(c.convergeRepeatInterval)
	}
}

func (c *Converger) converge() {
	logger := c.logger.Session("executing-convergence")
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		logger.Info("converge-tasks-started")

		defer func() {
			logger.Info("converge-tasks-done")
			wg.Done()
		}()

		err := c.taskController.ConvergeTasks(
			c.logger,
			c.kickTaskDuration,
			c.expirePendingTaskDuration,
			c.expireCompletedTaskDuration,
		)
		if err != nil {
			logger.Error("failed-to-converge-tasks", err)
		}
	}()

	wg.Add(1)
	go func() {
		logger.Info("converge-lrps-started")

		defer func() {
			logger.Info("converge-lrps-done")
			wg.Done()
		}()

		err := c.lrpConvergenceController.ConvergeLRPs(c.logger)
		if err != nil {
			logger.Error("failed-to-converge-lrps", err)
		}
	}()

	wg.Wait()
}
