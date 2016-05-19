package handlers

import (
	"net/http"

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
