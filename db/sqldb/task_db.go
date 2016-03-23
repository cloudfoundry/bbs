package sqldb

import (
	"database/sql"
	"time"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

func (db *SQLDB) DesireTask(logger lager.Logger, taskDef *models.TaskDefinition, taskGuid, domain string) error {
	taskDefData, err := db.serializeModel(logger, taskDef)
	if err != nil {
		return err
	}

	now := db.clock.Now()

	_, err = db.db.Exec(
		`INSERT INTO tasks (guid, domain, created_at, updated_at, first_completed_at, state, task_definition)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
		taskGuid,
		domain,
		now,
		now,
		time.Time{},
		models.Task_Pending,
		taskDefData,
	)

	if err != nil {
		return db.convertSQLError(err)
	}

	return nil
}

func (db *SQLDB) Tasks(logger lager.Logger, taskFilter models.TaskFilter) ([]*models.Task, error) {
	var rows *sql.Rows
	var err error
	if taskFilter.Domain != "" && taskFilter.CellID != "" {
		rows, err = db.db.Query("SELECT * FROM tasks WHERE domain = ? AND cell_id = ?", taskFilter.Domain, taskFilter.CellID)
	} else if taskFilter.Domain != "" {
		rows, err = db.db.Query("SELECT * FROM tasks WHERE domain = ?", taskFilter.Domain)
	} else if taskFilter.CellID != "" {
		rows, err = db.db.Query("SELECT * FROM tasks WHERE cell_id = ?", taskFilter.CellID)
	} else {
		rows, err = db.db.Query("SELECT * FROM tasks")
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []*models.Task{}
	for rows.Next() {
		task, err := db.fetchTask(logger, rows)
		if err != nil {
			return nil, err
		}
		results = append(results, task)
	}

	return results, nil
}

func (db *SQLDB) TaskByGuid(logger lager.Logger, taskGuid string) (*models.Task, error) {
	row := db.db.QueryRow("SELECT * FROM tasks WHERE guid = ?", taskGuid)
	return db.fetchTask(logger, row)
}

func (db *SQLDB) StartTask(logger lager.Logger, taskGuid, cellId string) (bool, error) {
	var started bool

	err := db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		task, err := db.fetchTaskForShare(logger, taskGuid, tx)
		if err != nil {
			return err
		}

		if task.State == models.Task_Running && task.CellId == cellId {
			return nil
		}

		if err = task.ValidateTransitionTo(models.Task_Running); err != nil {
			return err
		}

		now := db.clock.Now()
		_, err = tx.Exec(`
				UPDATE tasks SET state = ?, updated_at = ?, cell_id = ?
				WHERE guid = ?`,
			models.Task_Running,
			now,
			cellId,
			taskGuid,
		)
		if err != nil {
			return db.convertSQLError(err)
		}

		started = true
		return nil
	})

	return started, err
}

func (db *SQLDB) CancelTask(logger lager.Logger, taskGuid string) error {
	return db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		task, err := db.fetchTaskForShare(logger, taskGuid, tx)
		if err != nil {
			return err
		}

		if err = task.ValidateTransitionTo(models.Task_Completed); err != nil {
			if task.State != models.Task_Pending {
				return err
			}
		}
		return db.completeTask(logger, task, true, "task was cancelled", "", tx)
	})
}

func (db *SQLDB) CompleteTask(logger lager.Logger, taskGuid, cellID string, failed bool, failureReason, taskResult string) (*models.Task, error) {
	var task *models.Task

	err := db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		var err error
		task, err = db.fetchTaskForShare(logger, taskGuid, tx)
		if err != nil {
			return err
		}

		if task.CellId != cellID && task.State == models.Task_Running {
			return models.NewRunningOnDifferentCellError(cellID, task.CellId)
		}

		if err = task.ValidateTransitionTo(models.Task_Completed); err != nil {
			return err
		}

		return db.completeTask(logger, task, failed, failureReason, taskResult, tx)
	})

	return task, err
}

func (db *SQLDB) FailTask(logger lager.Logger, taskGuid, failureReason string) (*models.Task, error) {
	var task *models.Task

	err := db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		var err error
		task, err = db.fetchTaskForShare(logger, taskGuid, tx)
		if err != nil {
			return err
		}

		if err = task.ValidateTransitionTo(models.Task_Completed); err != nil {
			if task.State != models.Task_Pending {
				return err
			}
		}

		return db.completeTask(logger, task, true, failureReason, "", tx)
	})

	return task, err
}

func (db *SQLDB) completeTask(logger lager.Logger, task *models.Task, failed bool, failureReason, result string, tx *sql.Tx) error {
	now := db.clock.Now()
	_, err := tx.Exec(
		`UPDATE tasks SET
		  state = ?, updated_at = ?, first_completed_at = ?,
			failed = ?, failure_reason = ?, result = ?, cell_id = ?
			WHERE guid = ?
			`,
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
		return db.convertSQLError(err)
	}

	task.State = models.Task_Completed
	task.UpdatedAt = now.UnixNano()
	task.FirstCompletedAt = now.UnixNano()
	task.Failed = failed
	task.FailureReason = failureReason
	task.Result = result
	task.CellId = ""

	return nil
}

// The stager calls this when it wants to claim a completed task.  This ensures that only one
// stager ever attempts to handle a completed task
func (db *SQLDB) ResolvingTask(logger lager.Logger, taskGuid string) error {
	return db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		task, err := db.fetchTaskForShare(logger, taskGuid, tx)
		if err != nil {
			logger.Error("failed-getting-task", err)
			return err
		}

		if err = task.ValidateTransitionTo(models.Task_Resolving); err != nil {
			logger.Error("invalid-state-transition", err)
			return err
		}

		now := db.clock.Now()
		_, err = tx.Exec(
			`UPDATE tasks SET
		  state = ?, updated_at = ?
			WHERE guid = ?
			`,
			models.Task_Resolving,
			now,
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
	return db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		task, err := db.fetchTaskForShare(logger, taskGuid, tx)
		if err != nil {
			logger.Error("failed-getting-task", err)
			return err
		}

		if task.State != models.Task_Resolving {
			err = models.NewTaskTransitionError(task.State, models.Task_Resolving)
			logger.Error("invalid-state-transition", err)
			return err
		}

		_, err = tx.Exec(
			`DELETE FROM tasks WHERE guid = ?`,
			taskGuid,
		)
		if err != nil {
			logger.Error("failed-deleting-task", err)
			return db.convertSQLError(err)
		}

		return nil
	})
}

func (db *SQLDB) fetchTaskForShare(logger lager.Logger, taskGuid string, tx *sql.Tx) (*models.Task, error) {
	row := tx.QueryRow("SELECT * FROM tasks WHERE guid = ? LOCK IN SHARE MODE", taskGuid)
	return db.fetchTask(logger, row)
}

func (db *SQLDB) fetchTask(logger lager.Logger, scanner RowScanner) (*models.Task, error) {
	var guid, domain, cellID, failureReason string
	var result sql.NullString
	var createdAt, updatedAt, firstCompletedAt time.Time
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
		return nil, models.ErrResourceNotFound
	}

	var taskDef models.TaskDefinition
	err = db.deserializeModel(logger, taskDefData, &taskDef)
	if err != nil {
		return nil, models.ErrDeserializeJSON
	}

	task := &models.Task{
		TaskGuid:         guid,
		Domain:           domain,
		CreatedAt:        createdAt.UnixNano(),
		UpdatedAt:        updatedAt.UnixNano(),
		FirstCompletedAt: firstCompletedAt.UnixNano(),
		State:            models.Task_State(state),
		CellId:           cellID,
		Result:           result.String,
		Failed:           failed,
		FailureReason:    failureReason,
		TaskDefinition:   &taskDef,
	}
	return task, nil
}
