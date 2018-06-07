package sqldb

import (
	"strings"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/db/sqldb/internal"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

func (db *SQLDB) DesireTask(logger lager.Logger, taskDef *models.TaskDefinition, taskGuid, domain string) (*models.Task, error) {
	logger = logger.Session("desire-task", lager.Data{"task_guid": taskGuid})
	logger.Info("starting")
	defer logger.Info("complete")

	taskDefData, err := db.serializeModel(logger, taskDef)
	if err != nil {
		logger.Error("failed-serializing-task-definition", err)
		return nil, err
	}

	now := db.clock.Now().UnixNano()
	err = db.transact(logger, func(logger lager.Logger, tx helpers.Tx) error {
		_, err = db.insert(logger, tx, internal.TasksTable,
			helpers.SQLAttributes{
				"guid":               taskGuid,
				"domain":             domain,
				"created_at":         now,
				"updated_at":         now,
				"first_completed_at": 0,
				"state":              models.Task_Pending,
				"task_definition":    taskDefData,
			},
		)

		return err
	})

	if err != nil {
		logger.Error("failed-inserting-task", err)
		return nil, err
	}

	return &models.Task{
		TaskDefinition:   taskDef,
		TaskGuid:         taskGuid,
		Domain:           domain,
		CreatedAt:        now,
		UpdatedAt:        now,
		FirstCompletedAt: 0,
		State:            models.Task_Pending,
	}, nil
}

func (db *SQLDB) Tasks(logger lager.Logger, filter models.TaskFilter) ([]*models.Task, error) {
	logger = logger.Session("tasks", lager.Data{"filter": filter})
	logger.Debug("starting")
	defer logger.Debug("complete")

	wheres := []string{}
	values := []interface{}{}

	if filter.Domain != "" {
		wheres = append(wheres, "domain = ?")
		values = append(values, filter.Domain)
	}

	if filter.CellID != "" {
		wheres = append(wheres, "cell_id = ?")
		values = append(values, filter.CellID)
	}

	results := []*models.Task{}

	err := db.transact(logger, func(logger lager.Logger, tx helpers.Tx) error {
		rows, err := db.all(logger, tx, internal.TasksTable,
			internal.TaskColumns, helpers.NoLockRow,
			strings.Join(wheres, " AND "), values...,
		)
		if err != nil {
			logger.Error("failed-query", err)
			return err
		}
		defer rows.Close()

		results, _, _, err = db.taskDb.FetchTasks(logger, rows, tx, true)
		if err != nil {
			logger.Error("failed-fetch", err)
			return err
		}

		return nil
	})

	return results, err
}

func (db *SQLDB) TaskByGuid(logger lager.Logger, taskGuid string) (*models.Task, error) {
	logger = logger.Session("task-by-guid", lager.Data{"task_guid": taskGuid})
	logger.Debug("starting")
	defer logger.Debug("complete")

	var task *models.Task

	err := db.transact(logger, func(logger lager.Logger, tx helpers.Tx) error {
		var err error
		row := db.one(logger, tx, internal.TasksTable,
			internal.TaskColumns, helpers.NoLockRow,
			"guid = ?", taskGuid,
		)

		task, err = db.taskDb.FetchTask(logger, row, tx)
		return err
	})

	return task, err
}

func (db *SQLDB) StartTask(logger lager.Logger, taskGuid, cellId string) (*models.Task, *models.Task, bool, error) {
	logger = logger.Session("start-task", lager.Data{"task_guid": taskGuid, "cell_id": cellId})

	var started bool
	var beforeTask models.Task
	var afterTask *models.Task

	err := db.transact(logger, func(logger lager.Logger, tx helpers.Tx) error {
		var err error
		afterTask, err = db.taskDb.FetchTaskForUpdate(logger, taskGuid, tx)
		if err != nil {
			logger.Error("failed-locking-task", err)
			return err
		}

		beforeTask = *afterTask
		if afterTask.State == models.Task_Running && afterTask.CellId == cellId {
			logger.Debug("task-already-running-on-cell")
			return nil
		}

		if err = afterTask.ValidateTransitionTo(models.Task_Running); err != nil {
			logger.Error("failed-to-transition-task-to-running", err)
			return err
		}

		logger.Info("starting")
		defer logger.Info("complete")
		now := db.clock.Now().UnixNano()
		_, err = db.update(logger, tx, internal.TasksTable,
			helpers.SQLAttributes{
				"state":      models.Task_Running,
				"updated_at": now,
				"cell_id":    cellId,
			},
			"guid = ?", taskGuid,
		)
		if err != nil {
			return err
		}

		afterTask.State = models.Task_Running
		afterTask.UpdatedAt = now
		afterTask.CellId = cellId

		started = true
		return nil
	})

	return &beforeTask, afterTask, started, err
}

func (db *SQLDB) CancelTask(logger lager.Logger, taskGuid string) (*models.Task, *models.Task, string, error) {
	logger = logger.Session("cancel-task", lager.Data{"task_guid": taskGuid})
	logger.Info("starting")
	defer logger.Info("complete")

	var beforeTask models.Task
	var afterTask *models.Task
	var cellID string

	err := db.transact(logger, func(logger lager.Logger, tx helpers.Tx) error {
		var err error
		afterTask, err = db.taskDb.FetchTaskForUpdate(logger, taskGuid, tx)
		if err != nil {
			logger.Error("failed-locking-task", err)
			return err
		}

		beforeTask = *afterTask
		cellID = afterTask.CellId

		if err = afterTask.ValidateTransitionTo(models.Task_Completed); err != nil {
			if afterTask.State != models.Task_Pending {
				logger.Error("failed-to-transition-task-to-completed", err)
				return err
			}
		}
		err = db.taskDb.CompleteTask(logger, afterTask, true, "task was cancelled", "", tx)
		if err != nil {
			return err
		}

		return nil
	})

	return &beforeTask, afterTask, cellID, err
}

func (db *SQLDB) CompleteTask(logger lager.Logger, taskGuid, cellID string, failed bool, failureReason, taskResult string) (*models.Task, *models.Task, error) {
	logger = logger.Session("complete-task", lager.Data{"task_guid": taskGuid, "cell_id": cellID})
	logger.Info("starting")
	defer logger.Info("complete")

	var beforeTask models.Task
	var afterTask *models.Task

	err := db.transact(logger, func(logger lager.Logger, tx helpers.Tx) error {
		var err error
		afterTask, err = db.taskDb.FetchTaskForUpdate(logger, taskGuid, tx)
		if err != nil {
			logger.Error("failed-locking-task", err)
			return err
		}
		beforeTask = *afterTask

		if afterTask.CellId != cellID && afterTask.State == models.Task_Running {
			logger.Error("failed-task-already-running-on-different-cell", err)
			return models.NewRunningOnDifferentCellError(cellID, afterTask.CellId)
		}

		if err = afterTask.ValidateTransitionTo(models.Task_Completed); err != nil {
			logger.Error("failed-to-transition-task-to-completed", err)
			return err
		}

		err = db.taskDb.CompleteTask(logger, afterTask, failed, failureReason, taskResult, tx)
		if err != nil {
			return err
		}

		return nil
	})

	return &beforeTask, afterTask, err
}

func (db *SQLDB) FailTask(logger lager.Logger, taskGuid, failureReason string) (*models.Task, *models.Task, error) {
	logger = logger.Session("fail-task", lager.Data{"task_guid": taskGuid})
	logger.Info("starting")
	defer logger.Info("complete")

	var beforeTask models.Task
	var afterTask *models.Task

	err := db.transact(logger, func(logger lager.Logger, tx helpers.Tx) error {
		var err error
		afterTask, err = db.taskDb.FetchTaskForUpdate(logger, taskGuid, tx)
		if err != nil {
			logger.Error("failed-locking-task", err)
			return err
		}

		beforeTask = *afterTask

		if err = afterTask.ValidateTransitionTo(models.Task_Completed); err != nil {
			if afterTask.State != models.Task_Pending {
				logger.Error("failed-to-transition-task-to-completed", err)
				return err
			}
		}

		err = db.taskDb.CompleteTask(logger, afterTask, true, failureReason, "", tx)
		if err != nil {
			return err
		}

		return nil
	})

	return &beforeTask, afterTask, err
}

func (db *SQLDB) RejectTask(logger lager.Logger, taskGuid, rejectionReason string) (*models.Task, *models.Task, error) {
	var beforeTask models.Task
	var afterTask *models.Task

	err := db.transact(logger, func(logger lager.Logger, tx helpers.Tx) error {
		var err error
		afterTask, err = db.taskDb.FetchTaskForUpdate(logger, taskGuid, tx)
		if err != nil {
			logger.Error("failed-locking-task", err)
			return err
		}

		if afterTask.State != models.Task_Pending && afterTask.State != models.Task_Running {
			logger.Info("invalid-task-state", lager.Data{"task_state": afterTask.State})
			return models.ErrBadRequest
		}

		beforeTask = *afterTask

		afterTask.RejectionCount++
		afterTask.RejectionReason = rejectionReason
		afterTask.State = models.Task_Pending

		now := db.clock.Now().UnixNano()
		_, err = db.update(logger, tx, internal.TasksTable,
			helpers.SQLAttributes{
				"rejection_count":  afterTask.RejectionCount,
				"rejection_reason": rejectionReason,
				"updated_at":       now,
				"state":            afterTask.State,
			},
			"guid = ?", taskGuid,
		)
		if err != nil {
			logger.Error("failed-updating-tasks", err)
			return err
		}

		return nil
	})

	return &beforeTask, afterTask, err
}

// The stager calls this when it wants to claim a completed task.  This ensures that only one
// stager ever attempts to handle a completed task
func (db *SQLDB) ResolvingTask(logger lager.Logger, taskGuid string) (*models.Task, *models.Task, error) {
	logger = logger.WithData(lager.Data{"task_guid": taskGuid})
	logger.Info("starting")
	defer logger.Info("complete")

	var beforeTask models.Task
	var afterTask *models.Task

	err := db.transact(logger, func(logger lager.Logger, tx helpers.Tx) error {
		var err error
		afterTask, err = db.taskDb.FetchTaskForUpdate(logger, taskGuid, tx)
		if err != nil {
			logger.Error("failed-locking-task", err)
			return err
		}

		beforeTask = *afterTask

		if err = afterTask.ValidateTransitionTo(models.Task_Resolving); err != nil {
			logger.Error("invalid-state-transition", err)
			return err
		}

		now := db.clock.Now().UnixNano()
		_, err = db.update(logger, tx, internal.TasksTable,
			helpers.SQLAttributes{
				"state":      models.Task_Resolving,
				"updated_at": now,
			},
			"guid = ?", taskGuid,
		)
		if err != nil {
			logger.Error("failed-updating-tasks", err)
			return err
		}

		afterTask.State = models.Task_Resolving
		afterTask.UpdatedAt = now

		return nil
	})

	return &beforeTask, afterTask, err
}

func (db *SQLDB) DeleteTask(logger lager.Logger, taskGuid string) (*models.Task, error) {
	logger = logger.Session("delete-task", lager.Data{"task_guid": taskGuid})
	logger.Info("starting")
	defer logger.Info("complete")

	var task *models.Task

	err := db.transact(logger, func(logger lager.Logger, tx helpers.Tx) error {
		var err error
		task, err = db.taskDb.FetchTaskForUpdate(logger, taskGuid, tx)
		if err != nil {
			logger.Error("failed-locking-task", err)
			return err
		}

		if task.State != models.Task_Resolving {
			err = models.NewTaskTransitionError(task.State, models.Task_Resolving)
			logger.Error("invalid-state-transition", err)
			return err
		}

		_, err = db.delete(logger, tx, internal.TasksTable, "guid = ?", taskGuid)
		if err != nil {
			logger.Error("failed-deleting-task", err)
			return err
		}

		return nil
	})
	return task, err
}
