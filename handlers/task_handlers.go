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
	logger := h.logger.Session("tasks")

	request := &models.TasksRequest{}
	response := &models.TasksResponse{}

	response.Error = parseRequest(logger, req, request)
	if response.Error == nil {
		filter := models.TaskFilter{Domain: request.Domain, CellID: request.CellId}
		response.Tasks, response.Error = h.db.Tasks(logger, filter)
	}

	writeResponse(w, response)
}

func (h *TaskHandler) TaskByGuid(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("task-by-guyid")

	request := &models.TaskByGuidRequest{}
	response := &models.TaskResponse{}

	response.Error = parseRequest(logger, req, request)
	if response.Error == nil {
		response.Task, response.Error = h.db.TaskByGuid(logger, request.TaskGuid)
	}

	writeResponse(w, response)
}

func (h *TaskHandler) DesireTask(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("desire-task")

	request := &models.DesireTaskRequest{}
	response := &models.TaskLifecycleResponse{}

	response.Error = parseRequest(logger, req, request)
	if response.Error == nil {
		response.Error = h.db.DesireTask(logger, request.TaskDefinition, request.TaskGuid, request.Domain)
	}

	writeResponse(w, response)
}

func (h *TaskHandler) StartTask(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("start-task")

	request := &models.StartTaskRequest{}
	response := &models.StartTaskResponse{}

	response.Error = parseRequest(logger, req, request)
	if response.Error == nil {
		response.ShouldStart, response.Error = h.db.StartTask(logger, request.TaskGuid, request.CellId)
	}

	writeResponse(w, response)
}

func (h *TaskHandler) CancelTask(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("cancel-task")

	request := &models.TaskGuidRequest{}
	response := &models.TaskLifecycleResponse{}

	response.Error = parseRequest(logger, req, request)
	if response.Error == nil {
		response.Error = h.db.CancelTask(logger, request.TaskGuid)
	}

	writeResponse(w, response)
}

func (h *TaskHandler) FailTask(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("fail-task")

	request := &models.FailTaskRequest{}
	response := &models.TaskLifecycleResponse{}

	response.Error = parseRequest(logger, req, request)
	if response.Error == nil {
		response.Error = h.db.FailTask(logger, request.TaskGuid, request.FailureReason)
	}

	writeResponse(w, response)
}

func (h *TaskHandler) CompleteTask(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("complete-task")

	request := &models.CompleteTaskRequest{}
	response := &models.TaskLifecycleResponse{}

	response.Error = parseRequest(logger, req, request)
	if response.Error == nil {
		response.Error = h.db.CompleteTask(logger, request.TaskGuid, request.CellId, request.Failed, request.FailureReason, request.Result)
	}

	writeResponse(w, response)
}

func (h *TaskHandler) ResolvingTask(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("resolving-task")

	request := &models.TaskGuidRequest{}
	response := &models.TaskLifecycleResponse{}

	response.Error = parseRequest(logger, req, request)
	if response.Error == nil {
		response.Error = h.db.ResolvingTask(logger, request.TaskGuid)
	}

	writeResponse(w, response)
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("delete-task")

	request := &models.TaskGuidRequest{}
	response := &models.TaskLifecycleResponse{}

	response.Error = parseRequest(logger, req, request)
	if response.Error == nil {
		response.Error = h.db.DeleteTask(logger, request.TaskGuid)
	}

	writeResponse(w, response)
}

func (h *TaskHandler) ConvergeTasks(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("converge-tasks")

	request := &models.ConvergeTasksRequest{}
	response := &models.ConvergeTasksResponse{}

	response.Error = parseRequest(logger, req, request)
	if response.Error == nil {
		h.db.ConvergeTasks(
			logger,
			time.Duration(request.KickTaskDuration),
			time.Duration(request.ExpirePendingTaskDuration),
			time.Duration(request.ExpireCompletedTaskDuration),
		)
	}

	writeResponse(w, response)
}
