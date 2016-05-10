package handlers

import (
	"io/ioutil"
	"net/http"

	"github.com/cloudfoundry-incubator/auctioneer"
	"github.com/cloudfoundry-incubator/bbs/format"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

func (h *TaskHandler) Tasks_r0(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("tasks", lager.Data{"revision": 0})

	request := &models.TasksRequest{}
	response := &models.TasksResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		filter := models.TaskFilter{Domain: request.Domain, CellID: request.CellId}
		response.Tasks, err = h.db.Tasks(logger, filter)
		if err == nil {
			for i := range response.Tasks {
				task := response.Tasks[i]
				if task.TaskDefinition == nil {
					continue
				}
				response.Tasks[i] = task.VersionDownTo(format.V0)
			}
		}
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}

func (h *TaskHandler) Tasks_r1(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("tasks", lager.Data{"revision": 0})

	request := &models.TasksRequest{}
	response := &models.TasksResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		filter := models.TaskFilter{Domain: request.Domain, CellID: request.CellId}
		response.Tasks, err = h.db.Tasks(logger, filter)
		if err == nil {
			for i := range response.Tasks {
				task := response.Tasks[i]
				if task.TaskDefinition == nil {
					continue
				}
				response.Tasks[i] = task.VersionDownTo(format.V1)
			}
		}
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}

func (h *TaskHandler) TaskByGuid_r0(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("task-by-guid", lager.Data{"revision": 0})

	request := &models.TaskByGuidRequest{}
	response := &models.TaskResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		response.Task, err = h.db.TaskByGuid(logger, request.TaskGuid)
		if err == nil && response.Task.TaskDefinition != nil {
			response.Task = response.Task.VersionDownTo(format.V0)
		}
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}

func (h *TaskHandler) TaskByGuid_r1(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("task-by-guid", lager.Data{"revision": 0})

	request := &models.TaskByGuidRequest{}
	response := &models.TaskResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		response.Task, err = h.db.TaskByGuid(logger, request.TaskGuid)
		if err == nil && response.Task.TaskDefinition != nil {
			response.Task = response.Task.VersionDownTo(format.V1)
		}
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}

func (h *TaskHandler) DesireTask_r0(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("desire-task")

	request := &models.DesireTaskRequest{}
	response := &models.TaskLifecycleResponse{}

	defer writeResponse(w, response)

	err = parseRequestForDesireTask_r0(logger, req, request)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	err = h.db.DesireTask(logger, request.TaskDefinition, request.TaskGuid, request.Domain)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	taskStartRequest := auctioneer.NewTaskStartRequestFromModel(request.TaskGuid, request.Domain, request.TaskDefinition)
	err = h.auctioneerClient.RequestTaskAuctions([]*auctioneer.TaskStartRequest{&taskStartRequest})
	if err != nil {
		logger.Error("failed-requesting-task-auction", err)
		// The creation succeeded, the auction request error can be dropped
	} else {
		logger.Debug("succeeded-requesting-task-auction")
	}
}

func parseRequestForDesireTask_r0(logger lager.Logger, req *http.Request, request *models.DesireTaskRequest) error {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logger.Error("failed-to-read-body", err)
		return models.ErrUnknownError
	}

	err = request.Unmarshal(data)
	if err != nil {
		logger.Error("failed-to-parse-request-body", err)
		return models.ErrBadRequest
	}

	request.TaskDefinition.Action.SetTimeoutMsFromDeprecatedTimeoutNs()

	if err := request.Validate(); err != nil {
		logger.Error("invalid-request", err)
		return models.NewError(models.Error_InvalidRequest, err.Error())
	}

	return nil
}
