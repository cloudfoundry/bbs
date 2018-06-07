package internal

import (
	"database/sql"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/format"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"
)

const (
	TasksTable = "tasks"
)

var (
	TaskColumns = helpers.ColumnList{
		TasksTable + ".guid",
		TasksTable + ".domain",
		TasksTable + ".updated_at",
		TasksTable + ".created_at",
		TasksTable + ".first_completed_at",
		TasksTable + ".state",
		TasksTable + ".cell_id",
		TasksTable + ".result",
		TasksTable + ".failed",
		TasksTable + ".failure_reason",
		TasksTable + ".task_definition",
		TasksTable + ".rejection_count",
		TasksTable + ".rejection_reason",
	}
)

type TaskDbInternal interface {
	CompleteTask(logger lager.Logger, task *models.Task, failed bool, failureReason, result string, tx helpers.Tx) error
	FetchTaskForUpdate(logger lager.Logger, taskGuid string, queryable helpers.Queryable) (*models.Task, error)
	FetchTasks(logger lager.Logger, rows *sql.Rows, queryable helpers.Queryable, abortOnError bool) ([]*models.Task, []string, int, error)
	FetchTask(logger lager.Logger, scanner helpers.RowScanner, queryable helpers.Queryable) (*models.Task, error)
}

type taskDbInternal struct {
	helper     helpers.SQLHelper
	serializer format.Serializer
	clock      clock.Clock
}

func NewTaskDbInternal(helper helpers.SQLHelper, serializer format.Serializer, clock clock.Clock) TaskDbInternal {
	return &taskDbInternal{
		helper:     helper,
		serializer: serializer,
		clock:      clock,
	}
}

func (db *taskDbInternal) CompleteTask(logger lager.Logger, task *models.Task, failed bool, failureReason, result string, tx helpers.Tx) error {
	now := db.clock.Now().UnixNano()
	_, err := db.helper.Update(logger, tx, TasksTable,
		helpers.SQLAttributes{
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
		return err
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

func (db *taskDbInternal) FetchTaskForUpdate(logger lager.Logger, taskGuid string, queryable helpers.Queryable) (*models.Task, error) {
	row := db.helper.One(logger, queryable, TasksTable,
		TaskColumns, helpers.LockRow,
		"guid = ?", taskGuid,
	)
	return db.FetchTask(logger, row, queryable)
}

func (db *taskDbInternal) FetchTasks(logger lager.Logger, rows *sql.Rows, queryable helpers.Queryable, abortOnError bool) ([]*models.Task, []string, int, error) {
	tasks := []*models.Task{}
	invalidGuids := []string{}
	validGuids := []string{}
	var err error
	for rows.Next() {
		var task *models.Task
		var guid string

		task, guid, err = db.fetchTaskInternal(logger, rows)
		if err == models.ErrDeserialize {
			invalidGuids = append(invalidGuids, guid)
			if abortOnError {
				break
			}
			continue
		}
		tasks = append(tasks, task)
		validGuids = append(validGuids, task.TaskGuid)
	}

	if err == nil {
		err = rows.Err()
	}

	rows.Close()

	if len(invalidGuids) > 0 {
		db.deleteInvalidTasks(logger, queryable, invalidGuids...)
	}

	return tasks, validGuids, len(invalidGuids), err
}

func (db *taskDbInternal) FetchTask(logger lager.Logger, scanner helpers.RowScanner, queryable helpers.Queryable) (*models.Task, error) {
	task, guid, err := db.fetchTaskInternal(logger, scanner)
	if err == models.ErrDeserialize {
		db.deleteInvalidTasks(logger, queryable, guid)
	}
	return task, err
}

func (db *taskDbInternal) fetchTaskInternal(logger lager.Logger, scanner helpers.RowScanner) (*models.Task, string, error) {
	var guid, domain, cellID, failureReason, rejectionReason string
	var result sql.NullString
	var createdAt, updatedAt, firstCompletedAt int64
	var state, rejectionCount int32
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
		&rejectionCount,
		&rejectionReason,
	)

	if err == sql.ErrNoRows {
		return nil, "", err
	}

	if err != nil {
		logger.Error("failed-scanning-row", err)
		return nil, "", err
	}

	var taskDef models.TaskDefinition
	err = deserializeModel(logger, db.serializer, taskDefData, &taskDef)
	if err != nil {
		return nil, guid, models.ErrDeserialize
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
		RejectionCount:   rejectionCount,
		RejectionReason:  rejectionReason,
	}
	return task, guid, nil
}

func (db *taskDbInternal) deleteInvalidTasks(logger lager.Logger, queryable helpers.Queryable, guids ...string) error {
	for _, guid := range guids {
		logger.Info("deleting-invalid-task-from-db", lager.Data{"guid": guid})
		_, err := db.helper.Delete(logger, queryable, TasksTable, "guid = ?", guid)
		if err != nil {
			logger.Error("failed-deleting-task", err)
			return err
		}
	}
	return nil
}
