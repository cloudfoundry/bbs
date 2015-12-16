package handlers

import (
	"net/http"
	"time"

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
	var err error
	logger := h.logger.Session("tasks")

	request := &models.TasksRequest{}
	response := &models.TasksResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		filter := models.TaskFilter{Domain: request.Domain, CellID: request.CellId}
		response.Tasks, err = h.db.Tasks(logger, filter)
		if err == nil {
			for _, task := range response.Tasks {
				if task.TaskDefinition == nil {
					continue
				}
				transformedTaskDef := task.TaskDefinition.WithCacheDependenciesAsActions()
				task.TaskDefinition = &transformedTaskDef
			}
		}
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}

func (h *TaskHandler) TaskByGuid(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("task-by-guyid")

	request := &models.TaskByGuidRequest{}
	response := &models.TaskResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		response.Task, err = h.db.TaskByGuid(logger, request.TaskGuid)
		if err == nil && response.Task.TaskDefinition != nil {
			transformedTaskDef := response.Task.TaskDefinition.WithCacheDependenciesAsActions()
			response.Task.TaskDefinition = &transformedTaskDef
		}
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}

func (h *TaskHandler) DesireTask(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("desire-task")

	request := &models.DesireTaskRequest{}
	response := &models.TaskLifecycleResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		err = h.db.DesireTask(logger, request.TaskDefinition, request.TaskGuid, request.Domain)
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}

func (h *TaskHandler) StartTask(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("start-task")

	request := &models.StartTaskRequest{}
	response := &models.StartTaskResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		response.ShouldStart, err = h.db.StartTask(logger, request.TaskGuid, request.CellId)
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}

func (h *TaskHandler) CancelTask(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("cancel-task")

	request := &models.TaskGuidRequest{}
	response := &models.TaskLifecycleResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		err = h.db.CancelTask(logger, request.TaskGuid)
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}

func (h *TaskHandler) FailTask(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("fail-task")

	request := &models.FailTaskRequest{}
	response := &models.TaskLifecycleResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		err = h.db.FailTask(logger, request.TaskGuid, request.FailureReason)
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}

func (h *TaskHandler) CompleteTask(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("complete-task")

	request := &models.CompleteTaskRequest{}
	response := &models.TaskLifecycleResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		err = h.db.CompleteTask(logger, request.TaskGuid, request.CellId, request.Failed, request.FailureReason, request.Result)
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}

func (h *TaskHandler) ResolvingTask(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("resolving-task")

	request := &models.TaskGuidRequest{}
	response := &models.TaskLifecycleResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		err = h.db.ResolvingTask(logger, request.TaskGuid)
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("delete-task")

	request := &models.TaskGuidRequest{}
	response := &models.TaskLifecycleResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		err = h.db.DeleteTask(logger, request.TaskGuid)
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}

func (h *TaskHandler) ConvergeTasks(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("converge-tasks")

	request := &models.ConvergeTasksRequest{}
	response := &models.ConvergeTasksResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		h.db.ConvergeTasks(
			logger,
			time.Duration(request.KickTaskDuration),
			time.Duration(request.ExpirePendingTaskDuration),
			time.Duration(request.ExpireCompletedTaskDuration),
		)
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}
