package etcd

import (
	"fmt"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

const NO_TTL = 0

func (db *ETCDDB) Tasks(logger lager.Logger, filter models.TaskFilter) ([]*models.Task, error) {
	root, err := db.fetchRecursiveRaw(logger, TaskSchemaRoot)
	bbsErr := models.ConvertError(err)
	if bbsErr != nil {
		if bbsErr.Type == models.Error_ResourceNotFound {
			return []*models.Task{}, nil
		}
		return nil, err
	}
	if root.Nodes.Len() == 0 {
		return []*models.Task{}, nil
	}

	tasks := []*models.Task{}

	for _, node := range root.Nodes {
		task := new(models.Task)
		err := db.deserializeModel(logger, node, task)
		if err != nil {
			return nil, err
		}

		if filter.Domain != "" && task.Domain != filter.Domain {
			continue
		}
		if filter.CellID != "" && task.CellId != filter.CellID {
			continue
		}

		tasks = append(tasks, task)
	}

	logger.Debug("succeeded-performing-deserialization", lager.Data{"num-tasks": len(tasks)})

	return tasks, nil
}

func (db *ETCDDB) TaskByGuid(logger lager.Logger, taskGuid string) (*models.Task, error) {
	task, _, err := db.taskByGuidWithIndex(logger, taskGuid)
	return task, err
}

func (db *ETCDDB) taskByGuidWithIndex(logger lager.Logger, taskGuid string) (*models.Task, uint64, error) {
	node, err := db.fetchRaw(logger, TaskSchemaPathByGuid(taskGuid))
	if err != nil {
		return nil, 0, err
	}

	task := new(models.Task)
	deserializeErr := db.deserializeModel(logger, node, task)
	if deserializeErr != nil {
		logger.Error("failed-parsing-desired-task", deserializeErr)
		return nil, 0, models.ErrDeserializeJSON
	}

	return task, node.ModifiedIndex, nil
}

func (db *ETCDDB) DesireTask(logger lager.Logger, taskDef *models.TaskDefinition, taskGuid, domain string) error {
	logger = logger.Session("desire-task", lager.Data{"task-guid": taskGuid})
	logger.Info("starting")
	defer logger.Info("finished")

	task := &models.Task{
		TaskDefinition: taskDef,
		TaskGuid:       taskGuid,
		Domain:         domain,
		State:          models.Task_Pending,
		CreatedAt:      db.clock.Now().UnixNano(),
		UpdatedAt:      db.clock.Now().UnixNano(),
	}

	value, err := db.serializeModel(logger, task)
	if err != nil {
		return err
	}

	logger.Debug("persisting-task")
	_, err = db.client.Create(TaskSchemaPathByGuid(task.TaskGuid), value, NO_TTL)
	if err != nil {
		return ErrorFromEtcdError(logger, err)
	}
	logger.Debug("succeeded-persisting-task")

	return nil
}

func (db *ETCDDB) StartTask(logger lager.Logger, taskGuid, cellID string) (bool, error) {
	logger = logger.Session("start-task", lager.Data{"task-guid": taskGuid, "cell-id": cellID})

	logger.Info("starting")
	defer logger.Info("finished")

	task, index, err := db.taskByGuidWithIndex(logger, taskGuid)
	if err != nil {
		logger.Error("failed-to-fetch-task", err)
		return false, err
	}

	logger = logger.WithData(lager.Data{"task": task.LagerData()})

	if task.State == models.Task_Running && task.CellId == cellID {
		logger.Info("task-already-running")
		return false, nil
	}

	err = validateStateTransition(task.State, models.Task_Running)
	if err != nil {
		logger.Error("invalid-state-transition", err)
		return false, err
	}

	task.UpdatedAt = db.clock.Now().UnixNano()
	task.State = models.Task_Running
	task.CellId = cellID

	value, err := db.serializeModel(logger, task)
	if err != nil {
		return false, err
	}

	_, err = db.client.CompareAndSwap(TaskSchemaPathByGuid(taskGuid), value, NO_TTL, index)
	if err != nil {
		logger.Error("failed-persisting-task", err)
		return false, ErrorFromEtcdError(logger, err)
	}

	return true, nil
}

// The cell calls this when the user requested to cancel the task
// stagerTaskBBS will retry this repeatedly if it gets a StoreTimeout error (up to N seconds?)
// Will fail if the task has already been cancelled or completed normally
func (db *ETCDDB) CancelTask(logger lager.Logger, taskGuid string) (*models.Task, error) {
	logger = logger.Session("cancel-task", lager.Data{"task-guid": taskGuid})

	logger.Info("starting")
	defer logger.Info("finished")

	task, index, err := db.taskByGuidWithIndex(logger, taskGuid)
	if err != nil {
		logger.Error("failed-to-fetch-task", err)
		return nil, err
	}
	logger.Info("succeeded-getting-task")

	if task.State == models.Task_Resolving || task.State == models.Task_Completed {
		err = models.NewTaskTransitionError(task.State, models.Task_Completed)
		logger.Error("invalid-state-transition", err)
		return nil, err
	}

	logger.Info("completing-task")
	err = db.completeTask(logger, task, index, true, "task was cancelled", "")
	if err != nil {
		logger.Error("failed-completing-task", err)
		return nil, err
	}

	logger.Info("succeeded-completing-task")
	return task, nil
}

func (db *ETCDDB) FailTask(logger lager.Logger, taskGuid, failureReason string) (*models.Task, error) {
	logger = logger.Session("fail-task", lager.Data{"task-guid": taskGuid})

	logger.Info("starting")
	defer logger.Info("finished")

	logger.Info("getting-task")
	task, index, err := db.taskByGuidWithIndex(logger, taskGuid)
	if err != nil {
		logger.Error("failed-getting-task", err)
		return nil, err
	}
	logger.Info("succeeded-getting-task")

	if task.State == models.Task_Resolving || task.State == models.Task_Completed {
		err = models.NewTaskTransitionError(task.State, models.Task_Completed)
		logger.Error("invalid-state-transition", err)
		return nil, err
	}

	return task, db.completeTask(logger, task, index, true, failureReason, "")
}

// The cell calls this when it has finished running the task (be it success or failure)
// stagerTaskBBS will retry this repeatedly if it gets a StoreTimeout error (up to N seconds?)
// This really really shouldn't fail.  If it does, blog about it and walk away. If it failed in a
// consistent way (i.e. key already exists), there's probably a flaw in our design.
func (db *ETCDDB) CompleteTask(logger lager.Logger, taskGuid, cellId string, failed bool, failureReason, result string) (*models.Task, error) {
	logger = logger.Session("complete-task", lager.Data{"task-guid": taskGuid, "cell-id": cellId})

	logger.Info("starting")
	defer logger.Info("finished")

	logger.Info("getting-task")
	task, index, err := db.taskByGuidWithIndex(logger, taskGuid)
	if err != nil {
		logger.Error("failed-getting-task", err)
		return nil, err
	}
	logger.Info("succeeded-getting-task")

	if task.State == models.Task_Running && task.CellId != cellId {
		err = models.NewRunningOnDifferentCellError(cellId, task.CellId)
		logger.Error("invalid-cell-id", err)
		return nil, err
	}

	err = validateStateTransition(task.State, models.Task_Completed)
	if err != nil {
		logger.Error("invalid-state-transition", err)
		return nil, err
	}

	return task, db.completeTask(logger, task, index, failed, failureReason, result)
}

func (db *ETCDDB) completeTask(logger lager.Logger, task *models.Task, index uint64, failed bool, failureReason, result string) error {
	db.markTaskCompleted(task, failed, failureReason, result)

	value, err := db.serializeModel(logger, task)
	if err != nil {
		return err
	}

	logger.Info("persisting-task")
	_, err = db.client.CompareAndSwap(TaskSchemaPathByGuid(task.TaskGuid), value, NO_TTL, index)
	if err != nil {
		return ErrorFromEtcdError(logger, err)
	}
	logger.Info("succeded-persisting-task")

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
func (db *ETCDDB) ResolvingTask(logger lager.Logger, taskGuid string) error {
	logger = logger.Session("resolving-task", lager.Data{"task-guid": taskGuid})

	logger.Info("starting")
	defer logger.Info("finished")

	logger.Info("getting-task")
	task, index, err := db.taskByGuidWithIndex(logger, taskGuid)
	if err != nil {
		logger.Error("failed-getting-task", err)
		return err
	}
	logger.Info("succeeded-getting-task")

	err = validateStateTransition(task.State, models.Task_Resolving)
	if err != nil {
		logger.Error("invalid-state-transition", err)
		return err
	}

	task.UpdatedAt = db.clock.Now().UnixNano()
	task.State = models.Task_Resolving

	value, err := db.serializeModel(logger, task)
	if err != nil {
		return err
	}

	_, err = db.client.CompareAndSwap(TaskSchemaPathByGuid(taskGuid), value, NO_TTL, index)
	if err != nil {
		return ErrorFromEtcdError(logger, err)
	}
	return nil
}

// The stager calls this when it wants to signal that it has received a completion and is handling it
// stagerTaskBBS will retry this repeatedly if it gets a StoreTimeout error (up to N seconds?)
// If this fails, the stager should assume that someone else is handling the completion and should bail
func (db *ETCDDB) DeleteTask(logger lager.Logger, taskGuid string) error {
	logger = logger.Session("delete-task", lager.Data{"task-guid": taskGuid})

	logger.Info("starting")
	defer logger.Info("finished")

	logger.Info("getting-task")
	task, _, err := db.taskByGuidWithIndex(logger, taskGuid)
	if err != nil {
		logger.Error("failed-getting-task", err)
		return err
	}
	logger.Info("succeeded-getting-task")

	if task.State != models.Task_Resolving {
		err = models.NewTaskTransitionError(task.State, models.Task_Resolving)
		logger.Error("invalid-state-transition", err)
		return err
	}

	_, err = db.client.Delete(TaskSchemaPathByGuid(taskGuid), false)
	return ErrorFromEtcdError(logger, err)
}

func validateStateTransition(from, to models.Task_State) error {
	if (from == models.Task_Pending && to == models.Task_Running) ||
		(from == models.Task_Running && to == models.Task_Completed) ||
		(from == models.Task_Completed && to == models.Task_Resolving) {
		return nil
	} else {
		return models.NewError(models.Error_InvalidStateTransition, fmt.Sprintf("Cannot transition from %s to %s", from.String(), to.String()))
	}
}
