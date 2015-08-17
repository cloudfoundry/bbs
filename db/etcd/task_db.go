package etcd

import (
	"fmt"
	"path"

	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

const (
	TaskSchemaRoot = DataSchemaRoot + "task"
	NO_TTL         = uint64(0)
)

func TaskSchemaPath(task *models.Task) string {
	return TaskSchemaPathByGuid(task.GetTaskGuid())
}

func TaskSchemaPathByGuid(taskGuid string) string {
	return path.Join(TaskSchemaRoot, taskGuid)
}

func (db *ETCDDB) Tasks(logger lager.Logger, taskFilter db.TaskFilter) (*models.Tasks, *models.Error) {
	root, bbsErr := db.fetchRecursiveRaw(logger, TaskSchemaRoot)
	if bbsErr.Equal(models.ErrResourceNotFound) {
		return &models.Tasks{}, nil
	}
	if bbsErr != nil {
		return nil, bbsErr
	}
	if root.Nodes.Len() == 0 {
		return &models.Tasks{}, nil
	}

	tasks := models.Tasks{}

	for _, node := range root.Nodes {
		node := node

		var task models.Task
		deserializeErr := models.FromJSON([]byte(node.Value), &task)
		if deserializeErr != nil {
			logger.Error("failed-parsing-task", deserializeErr, lager.Data{"key": node.Key})
			return nil, models.ErrUnknownError
		}

		if taskFilter == nil || taskFilter(&task) {
			tasks.Tasks = append(tasks.Tasks, &task)
		}
	}

	logger.Debug("succeeded-performing-deserialization", lager.Data{"num-tasks": len(tasks.GetTasks())})

	return &tasks, nil
}

func (db *ETCDDB) TaskByGuid(logger lager.Logger, taskGuid string) (*models.Task, *models.Error) {
	task, _, err := db.taskByGuidWithIndex(logger, taskGuid)
	return task, err
}

func (db *ETCDDB) taskByGuidWithIndex(logger lager.Logger, taskGuid string) (*models.Task, uint64, *models.Error) {
	node, bbsErr := db.fetchRaw(logger, TaskSchemaPathByGuid(taskGuid))
	if bbsErr != nil {
		return nil, 0, bbsErr
	}

	var task models.Task
	deserializeErr := models.FromJSON([]byte(node.Value), &task)
	if deserializeErr != nil {
		logger.Error("failed-parsing-desired-task", deserializeErr)
		return nil, 0, models.ErrDeserializeJSON
	}

	return &task, node.ModifiedIndex, nil
}

func (db *ETCDDB) DesireTask(logger lager.Logger, request *models.DesireTaskRequest) *models.Error {
	taskGuid := request.TaskGuid
	domain := request.Domain
	taskDef := request.TaskDefinition
	taskLogger := logger.Session("desire-task", lager.Data{"task-guid": taskGuid})

	taskLogger.Info("starting")
	defer taskLogger.Info("finished")

	task := &models.Task{
		TaskDefinition: taskDef,
		TaskGuid:       taskGuid,
		Domain:         domain,
		State:          models.Task_Pending,
		CreatedAt:      db.clock.Now().UnixNano(),
		UpdatedAt:      db.clock.Now().UnixNano(),
	}

	value, modelErr := models.ToJSON(task)
	if modelErr != nil {
		logger.Error("failed-to-json", modelErr)
		return modelErr
	}

	taskLogger.Debug("persisting-task")
	_, err := db.client.Create(TaskSchemaPathByGuid(task.TaskGuid), string(value), NO_TTL)
	if err != nil {
		return ErrorFromEtcdError(logger, err)
	}
	taskLogger.Debug("succeeded-persisting-task")

	taskLogger.Debug("requesting-task-auction")
	err = db.auctioneerClient.RequestTaskAuctions([]*models.Task{task})
	if err != nil {
		taskLogger.Error("failed-requesting-task-auction", err)
		// The creation succeeded, the auction request error can be dropped
	} else {
		taskLogger.Debug("succeeded-requesting-task-auction")
	}

	return nil
}

func (db *ETCDDB) StartTask(logger lager.Logger, request *models.StartTaskRequest) (bool, *models.Error) {
	taskGuid := request.TaskGuid
	cellID := request.CellId
	if taskGuid == "" {
		return false, &models.Error{Type: models.InvalidRequest, Message: "missing task guid"}
	}
	if cellID == "" {
		return false, &models.Error{Type: models.InvalidRequest, Message: "missing cellId"}
	}

	logger = logger.Session("start-task", lager.Data{"task-guid": taskGuid, "cell-id": cellID})

	logger.Info("starting")
	defer logger.Info("finished")

	task, index, modelErr := db.taskByGuidWithIndex(logger, taskGuid)
	if modelErr != nil {
		logger.Error("failed-to-fetch-task", modelErr)
		return false, modelErr
	}

	logger = logger.WithData(lager.Data{"task": task.LagerData()})

	if task.State == models.Task_Running && task.CellId == cellID {
		logger.Info("task-already-running")
		return false, nil
	}

	modelErr = validateStateTransition(task.State, models.Task_Running)
	if modelErr != nil {
		logger.Error("invalid-state-transition", modelErr)
		return false, modelErr
	}

	task.UpdatedAt = db.clock.Now().UnixNano()
	task.State = models.Task_Running
	task.CellId = cellID

	value, modelErr := models.ToJSON(task)
	if modelErr != nil {
		logger.Error("failed-to-json", modelErr)
		return false, modelErr
	}

	_, err := db.client.CompareAndSwap(TaskSchemaPathByGuid(taskGuid), string(value), NO_TTL, "", index)
	if err != nil {
		logger.Error("failed-persisting-task", err)
		return false, ErrorFromEtcdError(logger, err)
	}

	return true, nil
}

// The cell calls this when the user requested to cancel the task
// stagerTaskBBS will retry this repeatedly if it gets a StoreTimeout error (up to N seconds?)
// Will fail if the task has already been cancelled or completed normally
func (db *ETCDDB) CancelTask(logger lager.Logger, taskGuid string) *models.Error {
	logger = logger.Session("cancel-task", lager.Data{"task-guid": taskGuid})

	logger.Info("starting")
	defer logger.Info("finished")

	task, index, modelErr := db.taskByGuidWithIndex(logger, taskGuid)
	if modelErr != nil {
		logger.Error("failed-to-fetch-task", modelErr)
		return modelErr
	}
	logger.Info("succeeded-getting-task")

	if task.State == models.Task_Resolving || task.State == models.Task_Completed {
		modelErr = models.NewTaskTransitionError(task.State, models.Task_Completed)
		logger.Error("invalid-state-transition", modelErr)
		return modelErr
	}

	logger.Info("completing-task")
	cellId := task.CellId
	modelErr = db.completeTask(logger, task, index, &models.CompleteTaskRequest{
		Failed:        true,
		FailureReason: "task was cancelled",
		Result:        "",
	})
	if modelErr != nil {
		logger.Error("failed-completing-task", modelErr)
		return modelErr
	}
	logger.Info("succeeded-completing-task")

	if cellId == "" {
		return nil
	}

	logger.Info("getting-cell-info")
	cellPresence, modelErr := db.cellDB.CellById(logger, cellId)
	if modelErr != nil {
		logger.Error("failed-getting-cell-info", modelErr)
		return nil
	}
	logger.Info("succeeded-getting-cell-info")

	logger.Info("cell-client-cancelling-task")
	err := db.cellClient.CancelTask(cellPresence.RepAddress, task.TaskGuid)
	if err != nil {
		logger.Error("cell-client-failed-cancelling-task", err)
		return nil
	}
	logger.Info("cell-client-succeeded-cancelling-task")

	return nil
}

func (db *ETCDDB) FailTask(logger lager.Logger, request *models.FailTaskRequest) *models.Error {
	logger = logger.Session("fail-task", lager.Data{"task-guid": request.TaskGuid})

	logger.Info("starting")
	defer logger.Info("finished")

	logger.Info("getting-task")
	task, index, modelErr := db.taskByGuidWithIndex(logger, request.TaskGuid)
	if modelErr != nil {
		logger.Error("failed-getting-task", modelErr)
		return modelErr
	}
	logger.Info("succeeded-getting-task")

	if task.State == models.Task_Resolving || task.State == models.Task_Completed {
		modelErr = models.NewTaskTransitionError(task.State, models.Task_Completed)
		logger.Error("invalid-state-transition", modelErr)
		return modelErr
	}

	return db.completeTask(logger, task, index, &models.CompleteTaskRequest{
		Failed:        true,
		FailureReason: request.FailureReason,
		Result:        "",
	})
}

// The cell calls this when it has finished running the task (be it success or failure)
// stagerTaskBBS will retry this repeatedly if it gets a StoreTimeout error (up to N seconds?)
// This really really shouldn't fail.  If it does, blog about it and walk away. If it failed in a
// consistent way (i.e. key already exists), there's probably a flaw in our design.
func (db *ETCDDB) CompleteTask(logger lager.Logger, request *models.CompleteTaskRequest) *models.Error {
	logger = logger.Session("complete-task", lager.Data{"task-guid": request.TaskGuid, "cell-id": request.CellId})

	logger.Info("starting")
	defer logger.Info("finished")

	logger.Info("getting-task")
	task, index, modelErr := db.taskByGuidWithIndex(logger, request.TaskGuid)
	if modelErr != nil {
		logger.Error("failed-getting-task", modelErr)
		return modelErr
	}
	logger.Info("succeeded-getting-task")

	if task.State == models.Task_Running && task.CellId != request.CellId {
		modelErr = models.NewRunningOnDifferentCellError(request.CellId, task.CellId)
		logger.Error("invalid-cell-id", modelErr)
		return modelErr
	}

	modelErr = validateStateTransition(task.State, models.Task_Completed)
	if modelErr != nil {
		logger.Error("invalid-state-transition", modelErr)
		return modelErr
	}

	return db.completeTask(logger, task, index, request)
}

func (db *ETCDDB) completeTask(logger lager.Logger, task *models.Task, index uint64, request *models.CompleteTaskRequest) *models.Error {
	db.markTaskCompleted(task, request.Failed, request.FailureReason, request.Result)

	value, modelErr := models.ToJSON(task)
	if modelErr != nil {
		logger.Error("failed-to-json", modelErr)
		return modelErr
	}

	logger.Info("persisting-task")
	_, err := db.client.CompareAndSwap(TaskSchemaPathByGuid(task.TaskGuid), string(value), NO_TTL, "", index)
	if err != nil {
		return ErrorFromEtcdError(logger, err)
	}
	logger.Info("succeded-persisting-task")

	if task.CompletionCallbackUrl == "" {
		return nil
	}

	logger.Info("task-client-completing-task")
	db.callbackWorkPool.Submit(db.taskCallbackFactory(logger, db, task))

	return nil
}

func (db *ETCDDB) markTaskCompleted(task *models.Task, failed bool, failureReason, result string) {
	now := db.clock.Now().UnixNano()
	task.CellId = ""
	task.UpdatedAt = now
	task.FirstCompletedAt = now
	task.State = models.Task_Completed
	task.Failed = failed
	task.FailureReason = failureReason
	task.Result = result
}

// The stager calls this when it wants to claim a completed task.  This ensures that only one
// stager ever attempts to handle a completed task
func (db *ETCDDB) ResolvingTask(logger lager.Logger, taskGuid string) *models.Error {
	logger = logger.Session("resolving-task", lager.Data{"task-guid": taskGuid})

	logger.Info("starting")
	defer logger.Info("finished")

	logger.Info("getting-task")
	task, index, modelErr := db.taskByGuidWithIndex(logger, taskGuid)
	if modelErr != nil {
		logger.Error("failed-getting-task", modelErr)
		return modelErr
	}
	logger.Info("succeeded-getting-task")

	modelErr = validateStateTransition(task.State, models.Task_Resolving)
	if modelErr != nil {
		logger.Error("invalid-state-transition", modelErr)
		return modelErr
	}

	task.UpdatedAt = db.clock.Now().UnixNano()
	task.State = models.Task_Resolving

	value, modelErr := models.ToJSON(task)
	if modelErr != nil {
		logger.Error("failed-to-json", modelErr)
		return modelErr
	}

	_, err := db.client.CompareAndSwap(TaskSchemaPathByGuid(taskGuid), string(value), NO_TTL, "", index)
	if err != nil {
		return ErrorFromEtcdError(logger, err)
	}
	return nil
}

// The stager calls this when it wants to signal that it has received a completion and is handling it
// stagerTaskBBS will retry this repeatedly if it gets a StoreTimeout error (up to N seconds?)
// If this fails, the stager should assume that someone else is handling the completion and should bail
func (db *ETCDDB) ResolveTask(logger lager.Logger, taskGuid string) *models.Error {
	logger = logger.Session("resolve-task", lager.Data{"task-guid": taskGuid})

	logger.Info("starting")
	defer logger.Info("finished")

	logger.Info("getting-task")
	task, _, modelErr := db.taskByGuidWithIndex(logger, taskGuid)
	if modelErr != nil {
		logger.Error("failed-getting-task", modelErr)
		return modelErr
	}
	logger.Info("succeeded-getting-task")

	if task.State != models.Task_Resolving {
		modelErr = models.NewTaskTransitionError(task.State, models.Task_Resolving)
		logger.Error("invalid-state-transition", modelErr)
		return modelErr
	}

	_, err := db.client.Delete(TaskSchemaPathByGuid(taskGuid), false)
	return ErrorFromEtcdError(logger, err)
}

func validateStateTransition(from, to models.Task_State) *models.Error {
	if (from == models.Task_Pending && to == models.Task_Running) ||
		(from == models.Task_Running && to == models.Task_Completed) ||
		(from == models.Task_Completed && to == models.Task_Resolving) {
		return nil
	} else {
		return &models.Error{
			Type:    models.InvalidStateTransition,
			Message: fmt.Sprintf("Cannot transition from %s to %s", from.String(), to.String()),
		}
	}
}
