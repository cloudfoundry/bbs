package sqldb

import (
	"database/sql"
	"strings"

	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

func (db *SQLDB) DesireTask(logger lager.Logger, taskDef *models.TaskDefinition, taskGuid, domain string) error {
	logger = logger.Session("desire-task-sql", lager.Data{"task_guid": taskGuid})
	logger.Info("starting")
	defer logger.Info("complete")

	taskDefData, err := db.serializeModel(logger, taskDef)
	if err != nil {
		logger.Error("failed-serializing-task-definition", err)
		return err
	}

	now := db.clock.Now().UnixNano()

	_, err = db.insert(logger, db.db, tasksTable,
		SQLAttributes{
			"guid":               taskGuid,
			"domain":             domain,
			"created_at":         now,
			"updated_at":         now,
			"first_completed_at": 0,
			"state":              models.Task_Pending,
			"task_definition":    taskDefData,
		},
	)
	if err != nil {
		logger.Error("failed-inserting-task", err)
		return db.convertSQLError(err)
	}

	return nil
}

func (db *SQLDB) Tasks(logger lager.Logger, filter models.TaskFilter) ([]*models.Task, error) {
	logger = logger.Session("tasks-sql", lager.Data{"filter": filter})
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

	rows, err := db.all(logger, db.db, tasksTable,
		taskColumns, NoLockRow,
		strings.Join(wheres, " AND "), values...,
	)
	if err != nil {
		logger.Error("failed-query", err)
		return nil, db.convertSQLError(err)
	}
	defer rows.Close()

	results := []*models.Task{}
	for rows.Next() {
		task, err := db.fetchTask(logger, rows, db.db)
		if err != nil {
			logger.Error("failed-fetch", err)
			return nil, err
		}
		results = append(results, task)
	}

	if rows.Err() != nil {
		logger.Error("failed-getting-next-row", rows.Err())
		return nil, db.convertSQLError(rows.Err())
	}

	return results, nil
}

func (db *SQLDB) TaskByGuid(logger lager.Logger, taskGuid string) (*models.Task, error) {
	logger = logger.Session("task-by-guid-sql", lager.Data{"task_guid": taskGuid})
	logger.Debug("starting")
	defer logger.Debug("complete")

	row := db.one(logger, db.db, tasksTable,
		taskColumns, NoLockRow,
		"guid = ?", taskGuid,
	)
	return db.fetchTask(logger, row, db.db)
}

func (db *SQLDB) StartTask(logger lager.Logger, taskGuid, cellId string) (bool, error) {
	logger = logger.Session("start-task-sql", lager.Data{"task_guid": taskGuid, "cell_id": cellId})

	var started bool

	err := db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		task, err := db.fetchTaskForUpdate(logger, taskGuid, tx)
		if err != nil {
			logger.Error("failed-locking-task", err)
			return err
		}

		if task.State == models.Task_Running && task.CellId == cellId {
			logger.Debug("task-already-running-on-cell")
			return nil
		}

		if err = task.ValidateTransitionTo(models.Task_Running); err != nil {
			logger.Error("failed-to-transition-task-to-running", err)
			return err
		}

		logger.Info("starting")
		defer logger.Info("complete")
		now := db.clock.Now().UnixNano()
		_, err = db.update(logger, tx, tasksTable,
			SQLAttributes{
				"state":      models.Task_Running,
				"updated_at": now,
				"cell_id":    cellId,
			},
			"guid = ?", taskGuid,
		)
		if err != nil {
			return db.convertSQLError(err)
		}

		started = true
		return nil
	})

	return started, err
}

func (db *SQLDB) CancelTask(logger lager.Logger, taskGuid string) (*models.Task, string, error) {
	logger = logger.Session("cancel-task-sql", lager.Data{"task_guid": taskGuid})
	logger.Info("starting")
	defer logger.Info("complete")

	var task *models.Task
	var cellID string

	err := db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		var err error
		task, err = db.fetchTaskForUpdate(logger, taskGuid, tx)
		if err != nil {
			logger.Error("failed-locking-task", err)
			return err
		}

		cellID = task.CellId

		if err = task.ValidateTransitionTo(models.Task_Completed); err != nil {
			if task.State != models.Task_Pending {
				logger.Error("failed-to-transition-task-to-completed", err)
				return err
			}
		}
		return db.completeTask(logger, task, true, "task was cancelled", "", tx)
	})

	return task, cellID, err
}

func (db *SQLDB) CompleteTask(logger lager.Logger, taskGuid, cellID string, failed bool, failureReason, taskResult string) (*models.Task, error) {
	logger = logger.Session("complete-task-sql", lager.Data{"task_guid": taskGuid, "cell_id": cellID})
	logger.Info("starting")
	defer logger.Info("complete")

	var task *models.Task

	err := db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		var err error
		task, err = db.fetchTaskForUpdate(logger, taskGuid, tx)
		if err != nil {
			logger.Error("failed-locking-task", err)
			return err
		}

		if task.CellId != cellID && task.State == models.Task_Running {
			logger.Error("failed-task-already-running-on-different-cell", err)
			return models.NewRunningOnDifferentCellError(cellID, task.CellId)
		}

		if err = task.ValidateTransitionTo(models.Task_Completed); err != nil {
			logger.Error("failed-to-transition-task-to-completed", err)
			return err
		}

		return db.completeTask(logger, task, failed, failureReason, taskResult, tx)
	})

	return task, err
}

func (db *SQLDB) FailTask(logger lager.Logger, taskGuid, failureReason string) (*models.Task, error) {
	logger = logger.Session("fail-task-sql", lager.Data{"task_guid": taskGuid})
	logger.Info("starting")
	defer logger.Info("complete")

	var task *models.Task

	err := db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		var err error
		task, err = db.fetchTaskForUpdate(logger, taskGuid, tx)
		if err != nil {
			logger.Error("failed-locking-task", err)
			return err
		}

		if err = task.ValidateTransitionTo(models.Task_Completed); err != nil {
			if task.State != models.Task_Pending {
				logger.Error("failed-to-transition-task-to-completed", err)
				return err
			}
		}

		return db.completeTask(logger, task, true, failureReason, "", tx)
	})

	return task, err
}

// The stager calls this when it wants to claim a completed task.  This ensures that only one
// stager ever attempts to handle a completed task
func (db *SQLDB) ResolvingTask(logger lager.Logger, taskGuid string) error {
	logger = logger.WithData(lager.Data{"task_guid": taskGuid})
	logger.Info("starting")
	defer logger.Info("complete")

	return db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		task, err := db.fetchTaskForUpdate(logger, taskGuid, tx)
		if err != nil {
			logger.Error("failed-locking-task", err)
			return err
		}

		if err = task.ValidateTransitionTo(models.Task_Resolving); err != nil {
			logger.Error("invalid-state-transition", err)
			return err
		}

		now := db.clock.Now().UnixNano()
		_, err = db.update(logger, tx, tasksTable,
			SQLAttributes{
				"state":      models.Task_Resolving,
				"updated_at": now,
			},
			"guid = ?", taskGuid,
		)
		if err != nil {
			logger.Error("failed-updating-tasks", err)
			return db.convertSQLError(err)
		}

		return nil
	})
}

func (db *SQLDB) DeleteTask(logger lager.Logger, taskGuid string) error {
	logger = logger.Session("delete-task-sql", lager.Data{"task_guid": taskGuid})
	logger.Info("starting")
	defer logger.Info("complete")

	return db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		task, err := db.fetchTaskForUpdate(logger, taskGuid, tx)
		if err != nil {
			logger.Error("failed-locking-task", err)
			return err
		}

		if task.State != models.Task_Resolving {
			err = models.NewTaskTransitionError(task.State, models.Task_Resolving)
			logger.Error("invalid-state-transition", err)
			return err
		}

		_, err = db.delete(logger, tx, tasksTable, "guid = ?", taskGuid)
		if err != nil {
			logger.Error("failed-deleting-task", err)
			return db.convertSQLError(err)
		}

		return nil
	})
}

func (db *SQLDB) completeTask(logger lager.Logger, task *models.Task, failed bool, failureReason, result string, tx *sql.Tx) error {
	now := db.clock.Now().UnixNano()
	_, err := db.update(logger, tx, tasksTable,
		SQLAttributes{
			"failed":             failed,
			"failure_reason":     failureReason,
			"result":             result,
			"state":              models.Task_Completed,
			"first_completed_at": now,
			"updated_at":         now,
			"cell_id":            "",
		},
		"guid = ?", task.TaskGuid,
	)
	if err != nil {
		logger.Error("failed-updating-tasks", err)
		return db.convertSQLError(err)
	}

	task.State = models.Task_Completed
	task.UpdatedAt = now
	task.FirstCompletedAt = now
	task.Failed = failed
	task.FailureReason = failureReason
	task.Result = result
	task.CellId = ""

	return nil
}

func (db *SQLDB) fetchTaskForUpdate(logger lager.Logger, taskGuid string, tx *sql.Tx) (*models.Task, error) {
	row := db.one(logger, tx, tasksTable,
		taskColumns, LockRow,
		"guid = ?", taskGuid,
	)
	return db.fetchTask(logger, row, tx)
}

func (db *SQLDB) fetchTask(logger lager.Logger, scanner RowScanner, tx Queryable) (*models.Task, error) {
	var guid, domain, cellID, failureReason string
	var result sql.NullString
	var createdAt, updatedAt, firstCompletedAt int64
	var state int32
	var failed bool
	var taskDefData []byte

	err := scanner.Scan(
		&guid,
		&domain,
		&updatedAt,
		&createdAt,
		&firstCompletedAt,
		&state,
		&cellID,
		&result,
		&failed,
		&failureReason,
		&taskDefData,
	)
	if err != nil {
		logger.Error("failed-scanning-row", err)
		return nil, models.ErrResourceNotFound
	}

	var taskDef models.TaskDefinition
	err = db.deserializeModel(logger, taskDefData, &taskDef)
	if err != nil {
		logger.Info("deleting-malformed-task-from-db", lager.Data{"guid": guid})
		_, err = db.delete(logger, tx, tasksTable, "guid = ?", guid)
		if err != nil {
			logger.Error("failed-deleting-task", err)
			return nil, db.convertSQLError(err)
		}
		return nil, models.ErrDeserialize
	}

	task := &models.Task{
		TaskGuid:         guid,
		Domain:           domain,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
		FirstCompletedAt: firstCompletedAt,
		State:            models.Task_State(state),
		CellId:           cellID,
		Result:           result.String,
		Failed:           failed,
		FailureReason:    failureReason,
		TaskDefinition:   &taskDef,
	}
	return task, nil
}
