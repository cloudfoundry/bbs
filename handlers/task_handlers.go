package handlers

import (
	"time"

	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
	"golang.org/x/net/context"
)

//go:generate counterfeiter -o fake_controllers/fake_task_controller.go . TaskController

type TaskController interface {
	Tasks(logger lager.Logger, domain, cellId string) ([]*models.Task, error)
	TaskByGuid(logger lager.Logger, taskGuid string) (*models.Task, error)
	DesireTask(logger lager.Logger, taskDefinition *models.TaskDefinition, taskGuid, domain string) error
	StartTask(logger lager.Logger, taskGuid, cellId string) (shouldStart bool, err error)
	CancelTask(logger lager.Logger, taskGuid string) error
	FailTask(logger lager.Logger, taskGuid, failureReason string) error
	CompleteTask(logger lager.Logger, taskGuid, cellId string, failed bool, failureReason, result string) error
	ResolvingTask(logger lager.Logger, taskGuid string) error
	DeleteTask(logger lager.Logger, taskGuid string) error
	ConvergeTasks(logger lager.Logger, kickTaskDuration, expirePendingTaskDuration, expireCompletedTaskDuration time.Duration) error
}

type TaskHandler struct {
	controller TaskController
	exitChan   chan<- struct{}
}

func NewTaskHandler(
	controller TaskController,
	exitChan chan<- struct{},
) *TaskHandler {
	return &TaskHandler{
		controller: controller,
		exitChan:   exitChan,
	}
}

func (h *bbsServer) Tasks(ctx context.Context, request *models.TasksRequest) (*models.TasksResponse, error) {
	var err error
	logger := h.logger.Session("tasks")
	response := &models.TasksResponse{}
	response.Tasks, err = h.taskController.Tasks(logger, request.Domain, request.CellId)
	response.Error = models.ConvertError(err)
	return response, nil
}

func (h *bbsServer) TaskByGuid(ctx context.Context, request *models.TaskByGuidRequest) (*models.TaskResponse, error) {
	var err error
	logger := h.logger.Session("task-by-guid")
	response := &models.TaskResponse{}
	response.Task, err = h.taskController.TaskByGuid(logger, request.TaskGuid)
	response.Error = models.ConvertError(err)
	return response, nil
}

func (h *bbsServer) DesireTask(ctx context.Context, request *models.DesireTaskRequest) (*models.TaskLifecycleResponse, error) {
	logger := h.logger.Session("desire-task")
	response := &models.TaskLifecycleResponse{}
	err := h.taskController.DesireTask(logger, request.TaskDefinition, request.TaskGuid, request.Domain)
	response.Error = models.ConvertError(err)
	return response, nil
}

func (h *bbsServer) StartTask(ctx context.Context, request *models.StartTaskRequest) (*models.StartTaskResponse, error) {
	var err error
	logger := h.logger.Session("start-task")
	response := &models.StartTaskResponse{}
	response.ShouldStart, err = h.taskController.StartTask(logger, request.TaskGuid, request.CellId)
	response.Error = models.ConvertError(err)
	return response, nil
}

func (h *bbsServer) CancelTask(context context.Context, req *models.TaskGuidRequest) (*models.TaskLifecycleResponse, error) {
	logger := h.logger.Session("cancel-task")

	response := &models.TaskLifecycleResponse{}
	err := h.taskController.CancelTask(logger, req.TaskGuid)
	response.Error = models.ConvertError(err)

	return response, nil
}

func (h *bbsServer) FailTask(ctx context.Context, request *models.FailTaskRequest) (*models.TaskLifecycleResponse, error) {
	logger := h.logger.Session("fail-task")
	response := &models.TaskLifecycleResponse{}
	err := h.taskController.FailTask(logger, request.TaskGuid, request.FailureReason)
	response.Error = models.ConvertError(err)
	return response, nil
}

func (h *bbsServer) CompleteTask(ctx context.Context, request *models.CompleteTaskRequest) (*models.TaskLifecycleResponse, error) {
	logger := h.logger.Session("complete-task")
	response := &models.TaskLifecycleResponse{}
	err := h.taskController.CompleteTask(logger, request.TaskGuid, request.CellId, request.Failed, request.FailureReason, request.Result)
	response.Error = models.ConvertError(err)
	return response, nil
}

func (h *bbsServer) ResolvingTask(ctx context.Context, request *models.TaskGuidRequest) (*models.TaskLifecycleResponse, error) {
	logger := h.logger.Session("resolving-task")
	response := &models.TaskLifecycleResponse{}
	err := h.taskController.ResolvingTask(logger, request.TaskGuid)
	response.Error = models.ConvertError(err)
	return response, nil
}

func (h *bbsServer) DeleteTask(ctx context.Context, request *models.TaskGuidRequest) (*models.TaskLifecycleResponse, error) {
	logger := h.logger.Session("delete-task")
	response := &models.TaskLifecycleResponse{}
	err := h.taskController.DeleteTask(logger, request.TaskGuid)
	response.Error = models.ConvertError(err)
	return response, nil
}
