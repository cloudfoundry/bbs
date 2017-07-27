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
		auctioneerClient:   auctioneerClient,
		repClientFactory:   repClientFactory,
		serviceClient:      serviceClient,
		updateWorkersCount: updateWorkersCount,
		exitChan:           exitChan,
	}
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

	guid, err := h.lrpDeploymentDB.CreateLRPDeployment(logger, request.Creation)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	lrp, err := h.desiredLRPDB.DesiredLRPByProcessGuid(logger, guid)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	// go h.desiredHub.Emit(models.NewDesiredLRPCreatedEvent(desiredLRP))

	schedulingInfo := lrp.DesiredLRPSchedulingInfo()
	h.desiredLRPHandler.startInstanceRange(logger, 0, lrp.Instances, &schedulingInfo)
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

	guid, err := h.lrpDeploymentDB.UpdateLRPDeployment(logger, request.Id, request.Update)
	if err != nil {
		logger.Debug("failed-updating-desired-lrp")
		response.Error = models.ConvertError(err)
		return
	}
	logger.Debug("completed-updating-desired-lrp")

	// TODO: what should we do here ?

	// TODO: scale up or down

	lrp, err := h.desiredLRPDB.DesiredLRPByProcessGuid(logger, guid)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	schedulingInfo := lrp.DesiredLRPSchedulingInfo()
	h.desiredLRPHandler.startInstanceRange(logger, 0, lrp.Instances, &schedulingInfo)
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

	err = h.lrpDeploymentDB.DeleteLRPDeployment(logger.Session("remove-desired"), request.Id)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	// h.stopInstancesFrom(logger, request.ProcessGuid, 0)
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

	err = h.lrpDeploymentDB.ActivateLRPDeploymentDefinition(logger, request.Id, request.DefinitionId)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	// go h.desiredHub.Emit(models.NewDesiredLRPRemovedEvent(desiredLRP))

	// TODO: what should we do here ?
}
