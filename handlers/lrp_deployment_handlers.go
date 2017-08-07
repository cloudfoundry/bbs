package handlers

import (
	"net/http"

	"code.cloudfoundry.org/auctioneer"
	"code.cloudfoundry.org/bbs/db"
	"code.cloudfoundry.org/bbs/events"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/serviceclient"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/rep"
)

type LRPDeploymentHandler struct {
	lrpDeploymentDB    db.LRPDeploymentDB
	desiredLRPDB       db.DesiredLRPDB
	desiredLRPHandler  *DesiredLRPHandler
	desiredHub         events.Hub
	auctioneerClient   auctioneer.Client
	repClientFactory   rep.ClientFactory
	serviceClient      serviceclient.ServiceClient
	updateWorkersCount int
	exitChan           chan<- struct{}
}

func NewLRPDeploymentHandler(
	updateWorkersCount int,
	lrpDeploymentDB db.LRPDeploymentDB,
	desiredLRPDB db.DesiredLRPDB,
	desiredLRPHandler *DesiredLRPHandler,
	desiredHub events.Hub,
	actualHub events.Hub,
	auctioneerClient auctioneer.Client,
	repClientFactory rep.ClientFactory,
	serviceClient serviceclient.ServiceClient,
	exitChan chan<- struct{},
) *LRPDeploymentHandler {
	return &LRPDeploymentHandler{
		lrpDeploymentDB:    lrpDeploymentDB,
		desiredLRPDB:       desiredLRPDB,
		desiredLRPHandler:  desiredLRPHandler,
		desiredHub:         desiredHub,
		auctioneerClient:   auctioneerClient,
		repClientFactory:   repClientFactory,
		serviceClient:      serviceClient,
		updateWorkersCount: updateWorkersCount,
		exitChan:           exitChan,
	}
}

func (h *LRPDeploymentHandler) LRPDeploymentSchedulingInfo(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	logger = logger.Session("lrp-deployments-schedulingInfo")

	request := &models.LRPDeploymentsRequest{}
	response := &models.LRPDeploymentsSchedulingInfoResponse{}

	defer func() { exitIfUnrecoverable(logger, h.exitChan, response.Error) }()
	defer writeResponse(w, response)

	err := parseRequest(logger, req, request)

	if err != nil {
		logger.Error("failed-parsing-request", err)
		response.Error = models.ConvertError(err)
		return
	}

	filter := models.LRPDeploymentFilter{
		Ids: request.Ids,
	}
	schedulingInfo, err := h.lrpDeploymentDB.LRPDeploymentSchedulingInfo(logger, filter)

	if err != nil {
		logger.Error("failed-retrieving-lrp-deployment-scheduling-info", err)
		response.Error = models.ConvertError(err)
		return
	}

	response.LrpDeploymentSchedulingInfo = schedulingInfo
}

func (h *LRPDeploymentHandler) CreateLRPDeployment(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	logger = logger.Session("create-lrp-deployment")

	request := &models.CreateLRPDeploymentRequest{}
	response := &models.LRPDeploymentLifecycleResponse{}
	defer func() { exitIfUnrecoverable(logger, h.exitChan, response.Error) }()
	defer writeResponse(w, response)

	err := parseRequest(logger, req, request)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	lrpDeployment, err := h.lrpDeploymentDB.CreateLRPDeployment(logger, request.Creation)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	lrp, err := h.desiredLRPDB.DesiredLRPByProcessGuid(logger, request.Creation.DefinitionId)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	go h.desiredHub.Emit(models.NewDesiredLRPCreatedEvent(lrp))
	go h.desiredHub.Emit(models.NewLRPDeploymentCreatedEvent(lrpDeployment))

	schedulingInfo := lrp.DesiredLRPSchedulingInfo()
	h.desiredLRPHandler.startInstanceRange(logger, 0, lrp.Instances, &schedulingInfo)
}

func (h *LRPDeploymentHandler) LRPDeployments(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	logger = logger.Session("lrp-deployments")

	request := &models.LRPDeploymentsRequest{}
	response := &models.LRPDeploymentsResponse{}

	defer func() { exitIfUnrecoverable(logger, h.exitChan, response.Error) }()
	defer writeResponse(w, response)

	err := parseRequest(logger, req, request)

	if err != nil {
		logger.Error("failed-parsing-request", err)
		response.Error = models.ConvertError(err)
		return
	}

	lrpDerps, err := h.lrpDeploymentDB.LRPDeployments(logger, request.Ids)

	if err != nil {
		logger.Error("failed-to-retrieve-deployments", err)
		response.Error = models.ConvertError(err)
		return
	}

	response.Deployments = lrpDerps
}

func (h *LRPDeploymentHandler) UpdateLRPDeployment(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	logger = logger.Session("update-lrp-deployment")

	request := &models.UpdateLRPDeploymentRequest{}
	response := &models.LRPDeploymentLifecycleResponse{}
	defer func() { exitIfUnrecoverable(logger, h.exitChan, response.Error) }()
	defer writeResponse(w, response)

	err := parseRequest(logger, req, request)
	if err != nil {
		logger.Error("failed-parsing-request", err)
		response.Error = models.ConvertError(err)
		return
	}

	logger = logger.WithData(lager.Data{"guid": request.Id})

	beforeLrpDeployment, err := h.lrpDeploymentDB.LRPDeploymentByProcessGuid(logger, request.Id)
	if err != nil {
		logger.Error("failed-retrieving-lrp-deployment", err)
		response.Error = models.ConvertError(err)
		return
	}

	afterLrpDeployment, err := h.lrpDeploymentDB.UpdateLRPDeployment(logger, request.Id, request.Update)
	if err != nil {
		logger.Debug("failed-updating-desired-lrp")
		response.Error = models.ConvertError(err)
		return
	}
	logger.Debug("completed-updating-desired-lrp")

	if request.Update.Definition != nil && request.Update.DefinitionId != nil {
		lrp, err := h.desiredLRPDB.DesiredLRPByProcessGuid(logger, *request.Update.DefinitionId)
		if err != nil {
			response.Error = models.ConvertError(err)
			return
		}

		go h.desiredHub.Emit(models.NewDesiredLRPCreatedEvent(lrp))
		go h.desiredHub.Emit(models.NewLRPDeploymentCreatedEvent(afterLrpDeployment))

		schedulingInfo := lrp.DesiredLRPSchedulingInfo()
		h.desiredLRPHandler.startInstanceRange(logger, 0, lrp.Instances, &schedulingInfo)
	} else {
		before, err := beforeLrpDeployment.DesiredLRP(beforeLrpDeployment.ActiveDefinitionId)
		if err != nil {
			response.Error = models.ConvertError(err)
			return
		}
		after, err := afterLrpDeployment.DesiredLRP(afterLrpDeployment.ActiveDefinitionId)
		if err != nil {
			response.Error = models.ConvertError(err)
			return
		}
		go h.desiredHub.Emit(models.NewDesiredLRPChangedEvent(&before, &after))
		go h.desiredHub.Emit(models.NewLRPDeploymentChangedEvent(beforeLrpDeployment, afterLrpDeployment))
	}

	if request.Update.Instances != nil {
		logger.Debug("updating-lrp-instances")
		lrp, err := h.desiredLRPDB.DesiredLRPByProcessGuid(logger, afterLrpDeployment.ActiveDefinitionId)
		if err != nil {
			response.Error = models.ConvertError(err)
			return
		}

		previousInstanceCount := beforeLrpDeployment.Instances

		requestedInstances := *request.Update.Instances - previousInstanceCount

		logger = logger.WithData(lager.Data{"instances_delta": requestedInstances})
		if requestedInstances > 0 {
			logger.Debug("increasing-the-instances")
			schedulingInfo := lrp.DesiredLRPSchedulingInfo()
			h.desiredLRPHandler.startInstanceRange(logger, previousInstanceCount, *request.Update.Instances, &schedulingInfo)
		}

		if requestedInstances < 0 {
			logger.Debug("decreasing-the-instances")
			numExtraActualLRP := previousInstanceCount + requestedInstances
			h.desiredLRPHandler.stopInstancesFrom(logger, lrp.ProcessGuid, int(numExtraActualLRP))
		}
	}
}

func (h *LRPDeploymentHandler) DeleteLRPDeployment(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	logger = logger.Session("delete-lrp-deployment")

	request := &models.RemoveLRPDeploymentRequest{}
	response := &models.LRPDeploymentLifecycleResponse{}
	defer func() { exitIfUnrecoverable(logger, h.exitChan, response.Error) }()
	defer writeResponse(w, response)

	err := parseRequest(logger, req, request)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}
	logger = logger.WithData(lager.Data{"process_guid": request.Id})

	lrpDeployment, err := h.lrpDeploymentDB.LRPDeploymentByProcessGuid(logger, request.Id)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	_, err = h.lrpDeploymentDB.DeleteLRPDeployment(logger.Session("remove-desired"), request.Id)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	for defID, _ := range lrpDeployment.Definitions {
		lrp, err := lrpDeployment.DesiredLRP(defID)
		if err != nil {
			logger.Error("failed-to-convert-to-desired-lrp", err)
			continue
		}
		go h.desiredHub.Emit(models.NewDesiredLRPRemovedEvent(&lrp))
		go h.desiredHub.Emit(models.NewLRPDeploymentRemovedEvent(lrpDeployment))
		h.desiredLRPHandler.stopInstancesFrom(logger, defID, 0)
	}
}

func (h *LRPDeploymentHandler) ActivateLRPDeploymentDefinition(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	logger = logger.Session("activate-lrp-deployment-definition")

	request := &models.ActivateLRPDeploymentDefinitionRequest{}
	response := &models.LRPDeploymentLifecycleResponse{}
	defer func() { exitIfUnrecoverable(logger, h.exitChan, response.Error) }()
	defer writeResponse(w, response)

	err := parseRequest(logger, req, request)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}
	logger = logger.WithData(lager.Data{"process_guid": request.Id})

	_, err = h.lrpDeploymentDB.ActivateLRPDeploymentDefinition(logger, request.Id, request.DefinitionId)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	lrp, err := h.desiredLRPDB.DesiredLRPByProcessGuid(logger, request.DefinitionId)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}
	schedulingInfo := lrp.DesiredLRPSchedulingInfo()
	h.desiredLRPHandler.startInstanceRange(logger, 0, lrp.Instances, &schedulingInfo)

	// TODO: what should we do here ?
}
