package handlers

import (
	"net/http"

	"code.cloudfoundry.org/bbs/controllers"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

type TaskHandler struct {
	controller *controllers.TaskController
	exitChan   chan<- struct{}
	logger     lager.Logger
}

func NewTaskHandler(
	logger lager.Logger,
	controller *controllers.TaskController,
	exitChan chan<- struct{},
) *TaskHandler {
	return &TaskHandler{
		logger:     logger.Session("task-http-handler"),
		controller: controller,
		exitChan:   exitChan,
	}
}

func (h *TaskHandler) Tasks(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("tasks")

	request := &models.TasksRequest{}
	response := &models.TasksResponse{}

	defer exitIfUnrecoverable(logger, h.exitChan, response.Error)
	defer writeResponse(w, response)

	err = parseRequest(logger, req, request)
	if err != nil {
		logger.Error("failed-parsing-request", err)
		response.Error = models.ConvertError(err)
		return
	}

	response.Tasks, err = h.controller.Tasks(logger, request.Domain, request.CellId)
	response.Error = models.ConvertError(err)
}

func (h *TaskHandler) TaskByGuid(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("task-by-guid")

	request := &models.TaskByGuidRequest{}
	response := &models.TaskResponse{}

	defer exitIfUnrecoverable(logger, h.exitChan, response.Error)
	defer writeResponse(w, response)

	err = parseRequest(logger, req, request)
	if err != nil {
		logger.Error("failed-parsing-request", err)
		response.Error = models.ConvertError(err)
		return
	}

	response.Task, err = h.controller.TaskByGuid(logger, request.TaskGuid)
	response.Error = models.ConvertError(err)
}

func (h *TaskHandler) DesireTask(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("desire-task")

	request := &models.DesireTaskRequest{}
	response := &models.TaskLifecycleResponse{}

	defer exitIfUnrecoverable(logger, h.exitChan, response.Error)
	defer writeResponse(w, response)

	err = parseRequest(logger, req, request)
	if err != nil {
		logger.Error("failed-parsing-request", err)
		response.Error = models.ConvertError(err)
		return
	}

	err = h.controller.DesireTask(h.logger, request.TaskDefinition, request.TaskGuid, request.Domain)
	response.Error = models.ConvertError(err)
}

func (h *TaskHandler) StartTask(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("start-task")

	request := &models.StartTaskRequest{}
	response := &models.StartTaskResponse{}

	defer exitIfUnrecoverable(logger, h.exitChan, response.Error)
	defer writeResponse(w, response)

	err = parseRequest(logger, req, request)
	if err != nil {
		logger.Error("failed-parsing-request", err)
		response.Error = models.ConvertError(err)
		return
	}

	response.ShouldStart, err = h.controller.StartTask(logger, request.TaskGuid, request.CellId)
	response.Error = models.ConvertError(err)
}

func (h *TaskHandler) CancelTask(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("cancel-task")

	request := &models.TaskGuidRequest{}
	response := &models.TaskLifecycleResponse{}

	defer exitIfUnrecoverable(logger, h.exitChan, response.Error)
	defer writeResponse(w, response)

	err := parseRequest(logger, req, request)
	if err != nil {
		logger.Error("failed-parsing-request", err)
		response.Error = models.ConvertError(err)
		return
	}

	err = h.controller.CancelTask(logger, request.TaskGuid)
	response.Error = models.ConvertError(err)
}

func (h *TaskHandler) FailTask(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("fail-task")

	request := &models.FailTaskRequest{}
	response := &models.TaskLifecycleResponse{}

	defer exitIfUnrecoverable(logger, h.exitChan, response.Error)
	defer writeResponse(w, response)

	err = parseRequest(logger, req, request)
	if err != nil {
		logger.Error("failed-parsing-request", err)
		response.Error = models.ConvertError(err)
		return
	}

	err = h.controller.FailTask(logger, request.TaskGuid, request.FailureReason)
	response.Error = models.ConvertError(err)
}

func (h *TaskHandler) CompleteTask(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("complete-task")

	request := &models.CompleteTaskRequest{}
	response := &models.TaskLifecycleResponse{}

	defer exitIfUnrecoverable(logger, h.exitChan, response.Error)
	defer writeResponse(w, response)

	err = parseRequest(logger, req, request)
	if err != nil {
		response.Error = models.ConvertError(err)
		logger.Error("failed-parsing-request", err)
		return
	}

	err = h.controller.CompleteTask(logger, request.TaskGuid, request.CellId, request.Failed, request.FailureReason, request.Result)
	response.Error = models.ConvertError(err)
}

func (h *TaskHandler) ResolvingTask(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("resolving-task")

	request := &models.TaskGuidRequest{}
	response := &models.TaskLifecycleResponse{}

	defer exitIfUnrecoverable(logger, h.exitChan, response.Error)
	defer writeResponse(w, response)

	err = parseRequest(logger, req, request)
	if err != nil {
		logger.Error("failed-parsing-request", err)
		response.Error = models.ConvertError(err)
		return
	}

	err = h.controller.ResolvingTask(logger, request.TaskGuid)
	response.Error = models.ConvertError(err)
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("delete-task")

	request := &models.TaskGuidRequest{}
	response := &models.TaskLifecycleResponse{}

	defer writeResponse(w, response)
	defer exitIfUnrecoverable(logger, h.exitChan, response.Error)

	err = parseRequest(logger, req, request)
	if err != nil {
		logger.Error("failed-parsing-request", err)
		response.Error = models.ConvertError(err)
		return
	}

	err = h.controller.DeleteTask(logger, request.TaskGuid)
	response.Error = models.ConvertError(err)
}
