package etcd

import (
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
	node, bbsErr := db.fetchRaw(logger, TaskSchemaPathByGuid(taskGuid))
	if bbsErr != nil {
		return nil, bbsErr
	}

	var task models.Task
	deserializeErr := models.FromJSON([]byte(node.Value), &task)
	if deserializeErr != nil {
		logger.Error("failed-parsing-desired-task", deserializeErr)
		return nil, models.ErrDeserializeJSON
	}

	return &task, nil
}

func (db *ETCDDB) DesireTask(logger lager.Logger, taskGuid, domain string, taskDef *models.TaskDefinition) error {
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

	err := task.Validate()
	if err != nil {
		return err
	}

	value, err := models.ToJSON(task)
	if err != nil {
		return err
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
