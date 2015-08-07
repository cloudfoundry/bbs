package etcd

import (
	"fmt"
	"path"

	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

const TaskSchemaRoot = DataSchemaRoot + "task"

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

	value, err := models.ToJSON(task)
	if err != nil {
		return models.NewError(models.InvalidRecord, err.Error())
	}

	taskLogger.Debug("persisting-task")
	_, err = db.client.Create(TaskSchemaPathByGuid(task.TaskGuid), string(value), 0)
	if err != nil {
		taskLogger.Error("failed-persisting-task", err)
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

func (db *ETCDDB) StartTask(logger lager.Logger, taskGuid string, cellID string) (bool, *models.Error) {
	if taskGuid == "" {
		return false, &models.Error{Type: models.InvalidRequest, Message: "missing task guid"}
	}
	if cellID == "" {
		return false, &models.Error{Type: models.InvalidRequest, Message: "missing cellId"}
	}

	logger = logger.WithData(lager.Data{"requested-task-guid": taskGuid, "requested-cell-id": cellID})

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

	value, err := models.ToJSON(task)
	if err != nil {
		logger.Error("failed-converting-to-json", err)
		return false, models.NewError(models.InvalidRecord, err.Error())
	}

	logger.Info("persisting-task")
	_, err = db.client.CompareAndSwap(TaskSchemaPathByGuid(taskGuid), string(value), 0, "", index)
	if err != nil {
		logger.Error("failed-persisting-task", err)
		return false, ErrorFromEtcdError(logger, err)
	}
	logger.Info("succeeded-persisting-task")

	return true, nil
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
