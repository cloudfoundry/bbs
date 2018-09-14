package controllers

import (
	"time"

	"code.cloudfoundry.org/auctioneer"
	"code.cloudfoundry.org/bbs/db"
	"code.cloudfoundry.org/bbs/events"
	"code.cloudfoundry.org/bbs/metrics"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/serviceclient"
	"code.cloudfoundry.org/bbs/taskworkpool"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/rep"
)

type TaskController struct {
	db                     db.TaskDB
	taskCompletionClient   taskworkpool.TaskCompletionClient
	auctioneerClient       auctioneer.Client
	serviceClient          serviceclient.ServiceClient
	repClientFactory       rep.ClientFactory
	taskHub                events.Hub
	taskStatMetronNotifier metrics.TaskStatMetronNotifier
	maxRetries             int
}

func NewTaskController(
	db db.TaskDB,
	taskCompletionClient taskworkpool.TaskCompletionClient,
	auctioneerClient auctioneer.Client,
	serviceClient serviceclient.ServiceClient,
	repClientFactory rep.ClientFactory,
	taskHub events.Hub,
	taskStatMetronNotifier metrics.TaskStatMetronNotifier,
	maxRetries int,
) *TaskController {
	return &TaskController{
		db:                     db,
		taskCompletionClient:   taskCompletionClient,
		auctioneerClient:       auctioneerClient,
		serviceClient:          serviceClient,
		repClientFactory:       repClientFactory,
		taskHub:                taskHub,
		taskStatMetronNotifier: taskStatMetronNotifier,
		maxRetries:             maxRetries,
	}
}

func (c *TaskController) Tasks(logger lager.Logger, domain, cellId string) ([]*models.Task, error) {
	logger = logger.Session("tasks")

	filter := models.TaskFilter{Domain: domain, CellID: cellId}
	return c.db.Tasks(logger, filter)
}

func (c *TaskController) TaskByGuid(logger lager.Logger, taskGuid string) (*models.Task, error) {
	logger = logger.Session("task-by-guid")

	return c.db.TaskByGuid(logger, taskGuid)
}

func (c *TaskController) DesireTask(logger lager.Logger, taskDefinition *models.TaskDefinition, taskGuid, domain string) error {
	var err error
	var task *models.Task
	logger = logger.Session("desire-task")

	logger = logger.WithData(lager.Data{"task_guid": taskGuid})

	task, err = c.db.DesireTask(logger, taskDefinition, taskGuid, domain)
	if err != nil {
		return err
	}
	go c.taskHub.Emit(models.NewTaskCreatedEvent(task))

	logger.Debug("start-task-auction-request")
	taskStartRequest := auctioneer.NewTaskStartRequestFromModel(taskGuid, domain, taskDefinition)
	err = c.auctioneerClient.RequestTaskAuctions(logger, []*auctioneer.TaskStartRequest{&taskStartRequest})
	if err != nil {
		logger.Error("failed-requesting-task-auction", err)
		// The creation succeeded, the auction request error can be dropped
	} else {
		logger.Debug("succeeded-requesting-task-auction")
	}

	return nil
}

func (c *TaskController) StartTask(logger lager.Logger, taskGuid, cellId string) (shouldStart bool, err error) {
	logger = logger.Session("start-task", lager.Data{"task_guid": taskGuid, "cell_id": cellId})
	before, after, shouldStart, err := c.db.StartTask(logger, taskGuid, cellId)
	if err == nil && shouldStart {
		go c.taskHub.Emit(models.NewTaskChangedEvent(before, after))
		c.taskStatMetronNotifier.TaskStarted(cellId)
	}
	return shouldStart, err
}

func (c *TaskController) CancelTask(logger lager.Logger, taskGuid string) error {
	logger = logger.Session("cancel-task")

	before, after, cellID, err := c.db.CancelTask(logger, taskGuid)
	if err != nil {
		return err
	}
	go c.taskHub.Emit(models.NewTaskChangedEvent(before, after))

	if after.CompletionCallbackUrl != "" {
		logger.Info("task-client-completing-task")
		go c.taskCompletionClient.Submit(c.db, c.taskHub, after)
	}

	if cellID == "" {
		return nil
	}

	logger.Info("start-check-cell-presence", lager.Data{"cell_id": cellID})
	cellPresence, err := c.serviceClient.CellById(logger, cellID)
	if err != nil {
		logger.Error("failed-fetching-cell-presence", err)
		// don't return an error, the rep will converge later
		return nil
	}
	logger.Info("finished-check-cell-presence", lager.Data{"cell_id": cellID})

	repClient, err := c.repClientFactory.CreateClient(cellPresence.RepAddress, cellPresence.RepUrl)
	if err != nil {
		logger.Error("create-rep-client-failed", err)
		return err
	}
	logger.Info("start-rep-cancel-task", lager.Data{"task_guid": taskGuid})
	repClient.CancelTask(logger, taskGuid)
	if err != nil {
		logger.Error("failed-rep-cancel-task", err)
		// don't return an error, the rep will converge later
		return nil
	}
	logger.Info("finished-rep-cancel-task", lager.Data{"task_guid": taskGuid})
	return nil
}

func (c *TaskController) FailTask(logger lager.Logger, taskGuid, failureReason string) error {
	var err error

	before, after, err := c.db.FailTask(logger, taskGuid, failureReason)
	if err != nil {
		return err
	}

	go c.taskHub.Emit(models.NewTaskChangedEvent(before, after))

	if after.CompletionCallbackUrl != "" {
		logger.Info("task-client-completing-task")
		go c.taskCompletionClient.Submit(c.db, c.taskHub, after)
	}

	return nil
}

func (c *TaskController) RejectTask(logger lager.Logger, taskGuid, rejectionReason string) error {
	logger = logger.Session("reject-task", lager.Data{"guid": taskGuid})
	logger.Info("start")
	defer logger.Info("complete")

	task, err := c.db.TaskByGuid(logger, taskGuid)
	if err != nil {
		logger.Error("failed-to-fetch-task", err)
		return err
	}

	logger.Info("reject-task", lager.Data{"rejection-reason": rejectionReason})
	before, after, rejectTaskErr := c.db.RejectTask(logger, taskGuid, rejectionReason)
	if rejectTaskErr != nil {
		logger.Error("failed-to-reject-task", rejectTaskErr)
	}

	if int(task.RejectionCount) >= c.maxRetries {
		return c.FailTask(logger, taskGuid, rejectionReason)
	}

	go c.taskHub.Emit(models.NewTaskChangedEvent(before, after))

	return rejectTaskErr
}

func (c *TaskController) CompleteTask(
	logger lager.Logger,
	taskGuid,
	cellId string,
	failed bool,
	failureReason,
	result string,
) error {
	var err error
	logger = logger.Session("complete-task")

	before, after, err := c.db.CompleteTask(logger, taskGuid, cellId, failed, failureReason, result)
	if err != nil {
		return err
	}
	go c.taskHub.Emit(models.NewTaskChangedEvent(before, after))

	if failed {
		c.taskStatMetronNotifier.TaskFailed(cellId)
	} else {
		c.taskStatMetronNotifier.TaskSucceeded(cellId)
	}

	if after.CompletionCallbackUrl != "" {
		logger.Info("task-client-completing-task")
		go c.taskCompletionClient.Submit(c.db, c.taskHub, after)
	}

	return nil
}

func (c *TaskController) ResolvingTask(logger lager.Logger, taskGuid string) error {
	logger = logger.Session("resolving-task")

	before, after, err := c.db.ResolvingTask(logger, taskGuid)
	if err != nil {
		return err
	}
	go c.taskHub.Emit(models.NewTaskChangedEvent(before, after))

	return nil
}

func (c *TaskController) DeleteTask(logger lager.Logger, taskGuid string) error {
	logger = logger.Session("delete-task")

	task, err := c.db.DeleteTask(logger, taskGuid)
	if err != nil {
		return err
	}
	go c.taskHub.Emit(models.NewTaskRemovedEvent(task))

	return nil
}

func (c *TaskController) ConvergeTasks(
	logger lager.Logger,
	kickTaskDuration,
	expirePendingTaskDuration,
	expireCompletedTaskDuration time.Duration,
) error {
	var err error
	logger = logger.Session("converge-tasks")

	defer c.emitTaskMetrics(logger)

	logger.Debug("listing-cells")
	cellSet, err := c.serviceClient.Cells(logger)
	if err == models.ErrResourceNotFound {
		logger.Debug("no-cells-found")
		cellSet = models.CellSet{}
	} else if err != nil {
		logger.Debug("failed-listing-cells")
		return err
	}
	logger.Debug("succeeded-listing-cells")

	tasksToAuction, tasksToComplete, eventsToEmit := c.db.ConvergeTasks(
		logger,
		cellSet,
		kickTaskDuration,
		expirePendingTaskDuration,
		expireCompletedTaskDuration,
	)

	c.taskStatMetronNotifier.TaskConvergenceStarted()

	logger.Debug("emitting-events-from-convergence", lager.Data{"num_tasks_to_complete": len(tasksToComplete)})
	for _, event := range eventsToEmit {
		go c.taskHub.Emit(event)
	}

	if len(tasksToAuction) > 0 {
		logger.Debug("requesting-task-auctions", lager.Data{"num_tasks_to_auction": len(tasksToAuction)})
		err = c.auctioneerClient.RequestTaskAuctions(logger, tasksToAuction)
		if err != nil {
			taskGuids := make([]string, len(tasksToAuction))
			for i, task := range tasksToAuction {
				taskGuids[i] = task.TaskGuid
			}
			logger.Error("failed-to-request-auctions-for-pending-tasks", err,
				lager.Data{"task_guids": taskGuids})
		}
		logger.Debug("done-requesting-task-auctions", lager.Data{"num_tasks_to_auction": len(tasksToAuction)})
	}

	logger.Debug("submitting-tasks-to-be-completed", lager.Data{"num_tasks_to_complete": len(tasksToComplete)})
	for _, task := range tasksToComplete {
		c.taskCompletionClient.Submit(c.db, c.taskHub, task)
	}
	logger.Debug("done-submitting-tasks-to-be-completed", lager.Data{"num_tasks_to_complete": len(tasksToComplete)})

	return nil
}

func (c *TaskController) emitTaskMetrics(logger lager.Logger) {
	logger = logger.Session("emit-task-metrics")
	pendingCount, runningCount, completedCount, resolvingCount := c.db.GetTaskCountByState(logger)
	c.taskStatMetronNotifier.TaskConvergenceResults(pendingCount, runningCount, completedCount, resolvingCount)
}
