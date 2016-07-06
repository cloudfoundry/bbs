package handlers

import (
	"net/http"
	"time"

	"code.cloudfoundry.org/auctioneer"
	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/bbs/db"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/taskworkpool"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/rep"
)

type TaskHandler struct {
	db                   db.TaskDB
	logger               lager.Logger
	taskCompletionClient taskworkpool.TaskCompletionClient
	auctioneerClient     auctioneer.Client
	serviceClient        bbs.ServiceClient
	repClientFactory     rep.ClientFactory
	exitChan             chan<- struct{}
}

func NewTaskHandler(
	logger lager.Logger,
	db db.TaskDB,
	taskCompletionClient taskworkpool.TaskCompletionClient,
	auctioneerClient auctioneer.Client,
	serviceClient bbs.ServiceClient,
	repClientFactory rep.ClientFactory,
	exitChan chan<- struct{},
) *TaskHandler {
	return &TaskHandler{
		db:                   db,
		logger:               logger.Session("task-handler"),
		taskCompletionClient: taskCompletionClient,
		auctioneerClient:     auctioneerClient,
		serviceClient:        serviceClient,
		repClientFactory:     repClientFactory,
		exitChan:             exitChan,
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
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
	exitIfUnrecoverable(logger, h.exitChan, response.Error)
}

func (h *TaskHandler) TaskByGuid(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("task-by-guid")

	request := &models.TaskByGuidRequest{}
	response := &models.TaskResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		response.Task, err = h.db.TaskByGuid(logger, request.TaskGuid)
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
	exitIfUnrecoverable(logger, h.exitChan, response.Error)
}

func (h *TaskHandler) DesireTask(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("desire-task")

	request := &models.DesireTaskRequest{}
	response := &models.TaskLifecycleResponse{}

	defer func() { exitIfUnrecoverable(logger, h.exitChan, response.Error) }()
	defer writeResponse(w, response)

	err = parseRequest(logger, req, request)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	logger = logger.WithData(lager.Data{"task_guid": request.TaskGuid})
	err = h.db.DesireTask(logger, request.TaskDefinition, request.TaskGuid, request.Domain)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	logger.Debug("start-task-auction-request")
	taskStartRequest := auctioneer.NewTaskStartRequestFromModel(request.TaskGuid, request.Domain, request.TaskDefinition)
	err = h.auctioneerClient.RequestTaskAuctions([]*auctioneer.TaskStartRequest{&taskStartRequest})
	if err != nil {
		logger.Error("failed-requesting-task-auction", err)
		// The creation succeeded, the auction request error can be dropped
	} else {
		logger.Debug("succeeded-requesting-task-auction")
	}
}

func (h *TaskHandler) StartTask(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("start-task")

	request := &models.StartTaskRequest{}
	response := &models.StartTaskResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		logger = logger.WithData(lager.Data{"task_guid": request.TaskGuid, "cell_id": request.CellId})
		response.ShouldStart, err = h.db.StartTask(logger, request.TaskGuid, request.CellId)
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
	exitIfUnrecoverable(logger, h.exitChan, response.Error)
}

func (h *TaskHandler) CancelTask(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("cancel-task")

	request := &models.TaskGuidRequest{}
	response := &models.TaskLifecycleResponse{}
	defer func() { exitIfUnrecoverable(logger, h.exitChan, response.Error) }()
	defer writeResponse(w, response)

	err := parseRequest(logger, req, request)
	if err != nil {
		logger.Error("failed-parsing-request", err)
		response.Error = models.ConvertError(err)
		return
	}

	task, cellID, err := h.db.CancelTask(logger, request.TaskGuid)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	if task.CompletionCallbackUrl != "" {
		logger.Info("task-client-completing-task")
		go h.taskCompletionClient.Submit(h.db, task)
	}

	if cellID == "" {
		return
	}

	logger.Info("start-check-cell-presence", lager.Data{"cell_id": cellID})
	cellPresence, err := h.serviceClient.CellById(logger, cellID)
	if err != nil {
		logger.Error("failed-fetching-cell-presence", err)
		return
	}
	logger.Info("finished-check-cell-presence", lager.Data{"cell_id": cellID})

	repClient := h.repClientFactory.CreateClient(cellPresence.RepAddress)
	logger.Info("start-rep-cancel-task", lager.Data{"task_guid": request.TaskGuid})
	repClient.CancelTask(request.TaskGuid)
	if err != nil {
		logger.Error("failed-rep-cancel-task", err)
		return
	}
	logger.Info("finished-rep-cancel-task", lager.Data{"task_guid": request.TaskGuid})
}

func (h *TaskHandler) FailTask(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("fail-task")

	request := &models.FailTaskRequest{}
	response := &models.TaskLifecycleResponse{}

	defer func() { exitIfUnrecoverable(logger, h.exitChan, response.Error) }()
	defer writeResponse(w, response)

	err = parseRequest(logger, req, request)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	task, err := h.db.FailTask(logger, request.TaskGuid, request.FailureReason)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	if task.CompletionCallbackUrl != "" {
		logger.Info("task-client-completing-task")
		go h.taskCompletionClient.Submit(h.db, task)
	}
}

func (h *TaskHandler) CompleteTask(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("complete-task")

	request := &models.CompleteTaskRequest{}
	response := &models.TaskLifecycleResponse{}

	defer func() { exitIfUnrecoverable(logger, h.exitChan, response.Error) }()
	defer writeResponse(w, response)

	err = parseRequest(logger, req, request)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	task, err := h.db.CompleteTask(logger, request.TaskGuid, request.CellId, request.Failed, request.FailureReason, request.Result)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	if task.CompletionCallbackUrl != "" {
		logger.Info("task-client-completing-task")
		go h.taskCompletionClient.Submit(h.db, task)
	}
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
	exitIfUnrecoverable(logger, h.exitChan, response.Error)
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
	exitIfUnrecoverable(logger, h.exitChan, response.Error)
}

func (h *TaskHandler) DeprecatedConvergeTasks(w http.ResponseWriter, req *http.Request) {
}

func (h *TaskHandler) ConvergeTasks(kickTaskDuration, expirePendingTaskDuration, expireCompletedTaskDuration time.Duration) error {
	var err error
	logger := h.logger.Session("converge-tasks")

	defer func() { exitIfUnrecoverable(logger, h.exitChan, models.ConvertError(err)) }()

	logger.Debug("listing-cells")
	cellSet, err := h.serviceClient.Cells(logger)
	if err == models.ErrResourceNotFound {
		logger.Debug("no-cells-found")
		cellSet = models.CellSet{}
	} else if err != nil {
		logger.Debug("failed-listing-cells")
		return err
	}
	logger.Debug("succeeded-listing-cells")

	tasksToAuction, tasksToComplete := h.db.ConvergeTasks(
		logger,
		cellSet,
		kickTaskDuration,
		expirePendingTaskDuration,
		expireCompletedTaskDuration,
	)

	if len(tasksToAuction) > 0 {
		logger.Debug("requesting-task-auctions", lager.Data{"num_tasks_to_auction": len(tasksToAuction)})
		err = h.auctioneerClient.RequestTaskAuctions(tasksToAuction)
		if err != nil {
			taskGuids := make([]string, len(tasksToAuction))
			for i, task := range tasksToAuction {
				taskGuids[i] = task.TaskGuid
			}
			logger.Error("failed-to-request-auctions-for-pending-tasks", err,
				lager.Data{"task_guids": taskGuids})
		}
		logger.Debug("done-requesting-task-auctions", lager.Data{"num_tasks_to_auction": len(tasksToAuction)})
	}

	logger.Debug("submitting-tasks-to-be-completed", lager.Data{"num_tasks_to_complete": len(tasksToComplete)})
	for _, task := range tasksToComplete {
		h.taskCompletionClient.Submit(h.db, task)
	}
	logger.Debug("done-submitting-tasks-to-be-completed", lager.Data{"num_tasks_to_complete": len(tasksToComplete)})
	return nil
}
