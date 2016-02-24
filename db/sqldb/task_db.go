package sqldb

import (
	"database/sql"
	"fmt"
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
		`INSERT INTO tasks (guid, domain, created_at, updated_at, state, task_definition)
			VALUES (?, ?, ?, ?, ?, ?)`,
		taskGuid,
		domain,
		now,
		now,
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
		now := db.clock.Now()
		result, err := tx.Exec(
			`UPDATE tasks SET
		  state = ?, updated_at = ?, cell_id = ?
			WHERE guid = ? AND state = ? AND cell_id = ?`,
			models.Task_Running,
			now,
			cellId,
			taskGuid,
			models.Task_Pending,
			"",
		)
		if err != nil {
			return db.convertSQLError(err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return db.convertSQLError(err)
		}

		if rowsAffected < 1 {
			row := tx.QueryRow("SELECT * FROM tasks WHERE guid = ?", taskGuid)
			task, err := db.fetchTask(logger, row)
			if err != nil {
				return err
			}

			if task.State == models.Task_Running && task.CellId == cellId {
				return nil
			}

			return models.NewError(
				models.Error_InvalidStateTransition,
				fmt.Sprintf("Cannot transition from %s to %s", task.State.String(), models.Task_Running.String()),
			)
		}

		started = true
		return nil
	})

	return started, err
}

func (db *SQLDB) CancelTask(logger lager.Logger, taskGuid string) error {
	return db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		now := db.clock.Now()

		result, err := tx.Exec(
			`UPDATE tasks SET
		  state = ?, updated_at = ?, cell_id = ?, first_completed_at = ?,
			failed = ?, failure_reason = ?, result = ?
			WHERE guid = ? AND state IN (?, ?)
			`,
			models.Task_Completed,
			now,
			"",
			now,
			true,
			"task was cancelled",
			"",
			taskGuid,
			models.Task_Pending,
			models.Task_Running,
		)
		if err != nil {
			return db.convertSQLError(err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return db.convertSQLError(err)
		}

		if rowsAffected < 1 {
			row := tx.QueryRow("SELECT * FROM tasks WHERE guid = ?", taskGuid)
			task, err := db.fetchTask(logger, row)
			if err != nil {
				return err
			}

			return models.NewTaskTransitionError(task.State, models.Task_Completed)
		}

		return nil
	})
}

func (db *SQLDB) CompleteTask(logger lager.Logger, taskGuid, cellID string, failed bool, failureReason, taskResult string) error {
	return db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		now := db.clock.Now()
		result, err := tx.Exec(
			`UPDATE tasks SET
		  state = ?, updated_at = ?, first_completed_at = ?,
			failed = ?, failure_reason = ?, result = ?, cell_id = ?
			WHERE cell_id = ? AND guid = ? AND state = ?
			`,
			models.Task_Completed,
			now,
			now,
			failed,
			failureReason,
			taskResult,
			"",
			cellID,
			taskGuid,
			models.Task_Running,
		)
		if err != nil {
			return db.convertSQLError(err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return db.convertSQLError(err)
		}

		if rowsAffected < 1 {
			row := tx.QueryRow("SELECT * FROM tasks WHERE guid = ?", taskGuid)
			task, err := db.fetchTask(logger, row)
			if err != nil {
				return err
			}

			if task.State != models.Task_Running {
				return models.NewTaskTransitionError(task.State, models.Task_Completed)
			}

			return models.NewRunningOnDifferentCellError(cellID, task.CellId)
		}

		return nil
	})
}

func (db *SQLDB) FailTask(logger lager.Logger, taskGuid, failureReason string) error {
	return db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		now := db.clock.Now()

		result, err := tx.Exec(
			`UPDATE tasks SET
		  state = ?, updated_at = ?, first_completed_at = ?,
			failed = ?, failure_reason = ?, result = ?, cell_id = ?
			WHERE guid = ? AND state IN (?, ?)
			`,
			models.Task_Completed,
			now,
			now,
			true,
			failureReason,
			"",
			"",
			taskGuid,
			models.Task_Running,
			models.Task_Pending,
		)
		if err != nil {
			return db.convertSQLError(err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return db.convertSQLError(err)
		}

		if rowsAffected < 1 {
			row := tx.QueryRow("SELECT * FROM tasks WHERE guid = ?", taskGuid)
			task, err := db.fetchTask(logger, row)
			if err != nil {
				return err
			}

			return models.NewTaskTransitionError(task.State, models.Task_Completed)
		}

		return nil
	})
}

func (db *SQLDB) ResolvingTask(logger lager.Logger, taskGuid string) error {
	return db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		now := db.clock.Now()
		result, _ := tx.Exec(
			`UPDATE tasks SET
		  state = ?, updated_at = ?
			WHERE state = ? AND guid = ?
			`,
			models.Task_Resolving,
			now,
			models.Task_Completed,
			taskGuid,
		)

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return db.convertSQLError(err)
		}

		if rowsAffected < 1 {
			row := tx.QueryRow("SELECT * FROM tasks WHERE guid = ?", taskGuid)
			task, err := db.fetchTask(logger, row)
			if err != nil {
				return err
			}

			return models.NewTaskTransitionError(task.State, models.Task_Resolving)
		}

		return nil
	})
}

func (db *SQLDB) DeleteTask(logger lager.Logger, taskGuid string) error {
	return db.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		result, _ := tx.Exec(
			`DELETE FROM tasks WHERE guid = ? AND state = ?`,
			taskGuid,
			models.Task_Resolving,
		)

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return db.convertSQLError(err)
		}

		if rowsAffected < 1 {
			row := tx.QueryRow("SELECT * FROM tasks WHERE guid = ?", taskGuid)
			_, err := db.fetchTask(logger, row)
			if err != nil {
				return err
			}

			return models.ErrBadRequest
		}

		return nil
	})
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
		&createdAt,
		&updatedAt,
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
