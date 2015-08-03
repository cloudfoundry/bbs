package handlers

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

type TaskHandler struct {
	db     db.TaskDB
	logger lager.Logger
}

func NewTaskHandler(logger lager.Logger, db db.TaskDB) *TaskHandler {
	return &TaskHandler{
		db:     db,
		logger: logger.Session("task-handler"),
	}
}

func (h *TaskHandler) Tasks(w http.ResponseWriter, req *http.Request) {
	domain := req.FormValue("domain")
	cellID := req.FormValue("cell_id")
	logger := h.logger.Session("tasks", lager.Data{
		"domain":  domain,
		"cell_id": cellID,
	})

	if domain != "" && cellID != "" {
		writeBadRequestResponse(w, models.InvalidRequest, errors.New("too many filters"))
		return
	}

	tasks, err := h.db.Tasks(h.logger, taskFilter(domain, cellID))
	if err != nil {
		logger.Error("failed-to-fetch-tasks", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	writeProtoResponse(w, http.StatusOK, tasks)
}

func taskFilter(domain string, cellID string) db.TaskFilter {
	if domain != "" {
		return func(t *models.Task) bool {
			return domain == t.Domain
		}
	}
	if cellID != "" {
		return func(t *models.Task) bool {
			return cellID == t.CellId
		}
	}

	return nil
}

func (h *TaskHandler) TaskByGuid(w http.ResponseWriter, req *http.Request) {
	taskGuid := req.FormValue(":task_guid")
	logger := h.logger.Session("task-by-guid", lager.Data{
		"task_guid": taskGuid,
	})

	task, err := h.db.TaskByGuid(h.logger, taskGuid)
	if err == models.ErrResourceNotFound {
		writeNotFoundResponse(w, err)
		return
	}
	if err != nil {
		logger.Error("failed-to-fetch-task", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	writeProtoResponse(w, http.StatusOK, task)
}

func (h *TaskHandler) DesireTask(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("desire-task")

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logger.Error("failed-to-read-body", err)
		writeBadRequestResponse(w, models.InvalidRequest, fmt.Errorf("failed to read request body: %s", err))
		return
	}

	logger.Info("data", lager.Data{"data": data})
	taskReq := &models.DesireTaskRequest{}
	err = taskReq.Unmarshal(data)
	if err != nil {
		logger.Error("failed-to-unmarshal-task", err)
		writeBadRequestResponse(w, models.InvalidRequest, fmt.Errorf("failed to unmarshal TaskDescription: %s", err))
		return
	}

	err = h.db.DesireTask(logger, taskReq.TaskGuid, taskReq.Domain, taskReq.TaskDefinition)
	if err != nil {
		logger.Error("failed-to-desire-task", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	writeEmptyResponse(w, http.StatusCreated)
}
