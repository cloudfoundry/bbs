package handlers

import (
	"net/http"

	"github.com/cloudfoundry-incubator/bbs/models"
)

func (h *TaskHandler) Tasks_V2(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("tasks")

	request := &models.TasksRequest{}
	response := &models.TasksResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		filter := models.TaskFilter{Domain: request.Domain, CellID: request.CellId}
		response.Tasks, err = h.db.Tasks(logger, filter)
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}

func (h *TaskHandler) TaskByGuid_V2(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("task-by-guyid")

	request := &models.TaskByGuidRequest{}
	response := &models.TaskResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		response.Task, err = h.db.TaskByGuid(logger, request.TaskGuid)
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}
