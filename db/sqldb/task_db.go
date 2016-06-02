package sqldb

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

func (db *SQLDB) DesireTask(logger lager.Logger, taskDef *models.TaskDefinition, taskGuid, domain string) error {
	logger = logger.Session("desire-task-sql", lager.Data{"task_guid": taskGuid})
	logger.Debug("starting")
	defer logger.Debug("complete")

	taskDefData, err := db.serializeModel(logger, taskDef)
	if err != nil {
		logger.Error("failed-serializing-task-definition", err)
		return err
	}

	now := db.clock.Now().UnixNano()

	_, err = db.db.Exec(db.getQuery(InsertTaskQuery),
		taskGuid,
		domain,
		now,
		now,
		0,
		models.Task_Pending,
		taskDefData,
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

	var rows *sql.Rows
	var err error
	conditions := map[whereClause]interface{}{
		whereDomainEquals: filter.Domain,
		whereCellIdEquals: filter.CellID,
	}
	wheres := []string{}
	values := []interface{}{}

	index := 1
	for field, value := range conditions {
		if value == "" {
			continue
		}

		if db.flavor == Postgres {
			postgresField := strings.Replace(field.string, "?", fmt.Sprintf("$%d", index), -1)
			wheres = append(wheres, postgresField)
		} else {
			wheres = append(wheres, field.string)
		}

		values = append(values, value)

		index++
	}

	query := db.getQuery(SelectTasksBaseQuery)
	if len(wheres) > 0 {
		query += fmt.Sprintf(" WHERE %s\n", strings.Join(wheres, " AND "))
	}
	rows, err = db.db.Query(query, values...)
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

	row := db.db.QueryRow(db.getQuery(SelectTaskByGuidQuery), taskGuid)
	return db.fetchTask(logger, row, db.db)
}

func (db *SQLDB) StartTask(logger lager.Logger, taskGuid, cellId string) (bool, error) {
	logger = logger.Session("start-task-sql", lager.Data{"task_guid": taskGuid, "cell_id": cellId})
	logger.Debug("starting")
	defer logger.Debug("complete")

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

		now := db.clock.Now().UnixNano()
		_, err = tx.Exec(db.getQuery(UpdateTaskByGuidQuery),
			models.Task_Running,
			now,
			cellId,
			taskGuid,
		)
		if err != nil {
			logger.Error("failed-to-update-task", err)
			return db.convertSQLError(err)
		}

		started = true
		return nil
	})

	return started, err
}

func (db *SQLDB) CancelTask(logger lager.Logger, taskGuid string) (*models.Task, string, error) {
	logger = logger.Session("cancel-task-sql", lager.Data{"task_guid": taskGuid})
	logger.Debug("starting")
	defer logger.Debug("complete")

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
	logger.Debug("starting")
	defer logger.Debug("complete")

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
	logger.Debug("starting")
	defer logger.Debug("complete")

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
	logger.Debug("starting")
	defer logger.Debug("complete")

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
		_, err = tx.Exec(db.getQuery(UpdateTaskByGuidQuery),
			models.Task_Resolving,
			now,
			task.CellId,
			taskGuid,
		)
		if err != nil {
			logger.Error("failed-resolving-task", err)
			return db.convertSQLError(err)
		}

		return nil
	})
}

func (db *SQLDB) DeleteTask(logger lager.Logger, taskGuid string) error {
	logger = logger.Session("delete-task-sql", lager.Data{"task_guid": taskGuid})
	logger.Debug("starting")
	defer logger.Debug("complete")

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

		_, err = tx.Exec(db.getQuery(DeleteTaskQuery), taskGuid)
		if err != nil {
			logger.Error("failed-deleting-task", err)
			return db.convertSQLError(err)
		}

		return nil
	})
}

func (db *SQLDB) completeTask(logger lager.Logger, task *models.Task, failed bool, failureReason, result string, tx *sql.Tx) error {
	now := db.clock.Now().UnixNano()
	_, err := tx.Exec(db.getQuery(CompleteTaskByGuidQuery),
		models.Task_Completed,
		now,
		now,
		failed,
		failureReason,
		result,
		"",
		task.TaskGuid,
	)
	if err != nil {
		logger.Error("failed-updating-task", err)
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
	query := db.getQuery(SelectTaskByGuidQuery) + " FOR UPDATE"
	row := tx.QueryRow(query, taskGuid)
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
		_, err = tx.Exec(db.getQuery(DeleteTaskQuery), guid)
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
