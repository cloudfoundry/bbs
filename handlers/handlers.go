package handlers

import (
	"io/ioutil"
	"net/http"
	"strconv"

	"code.cloudfoundry.org/auctioneer"
	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/bbs/controllers"
	"code.cloudfoundry.org/bbs/db"
	"code.cloudfoundry.org/bbs/events"
	"code.cloudfoundry.org/bbs/handlers/middleware"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/taskworkpool"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/rep"
	"github.com/gogo/protobuf/proto"
	"github.com/tedsuo/rata"
)

func New(
	logger lager.Logger,
	updateWorkers int,
	convergenceWorkersSize int,
	db db.DB,
	desiredHub, actualHub events.Hub,
	taskCompletionClient taskworkpool.TaskCompletionClient,
	serviceClient bbs.ServiceClient,
	auctioneerClient auctioneer.Client,
	repClientFactory rep.ClientFactory,
	migrationsDone <-chan struct{},
	exitChan chan struct{},
) http.Handler {
	retirer := controllers.NewActualLRPRetirer(db, actualHub, repClientFactory, serviceClient)
	pingHandler := NewPingHandler()
	domainHandler := NewDomainHandler(db, exitChan)
	actualLRPHandler := NewActualLRPHandler(db, exitChan)
	actualLRPLifecycleHandler := NewActualLRPLifecycleHandler(db, db, actualHub, auctioneerClient, retirer, exitChan)
	evacuationHandler := NewEvacuationHandler(db, db, db, actualHub, auctioneerClient, exitChan)
	desiredLRPHandler := NewDesiredLRPHandler(updateWorkers, db, db, desiredHub, actualHub, auctioneerClient, repClientFactory, serviceClient, exitChan)
	taskController := controllers.NewTaskController(db, taskCompletionClient, auctioneerClient, serviceClient, repClientFactory)
	taskHandler := NewTaskHandler(taskController, exitChan)
	eventsHandler := NewEventHandler(desiredHub, actualHub)
	cellsHandler := NewCellHandler(serviceClient, exitChan)

	emitter := middleware.NewLatencyEmitter(logger)

	actions := rata.Handlers{
		// Ping
		bbs.PingRoute: emitter.EmitLatency(middleware.LogWrap(logger, pingHandler.Ping)),

		// Domains
		bbs.DomainsRoute:      route(emitter.EmitLatency(middleware.LogWrap(logger, domainHandler.Domains))),
		bbs.UpsertDomainRoute: route(emitter.EmitLatency(middleware.LogWrap(logger, domainHandler.Upsert))),

		// Actual LRPs
		bbs.ActualLRPGroupsRoute:                     route(emitter.EmitLatency(middleware.LogWrap(logger, actualLRPHandler.ActualLRPGroups))),
		bbs.ActualLRPGroupsByProcessGuidRoute:        route(emitter.EmitLatency(middleware.LogWrap(logger, actualLRPHandler.ActualLRPGroupsByProcessGuid))),
		bbs.ActualLRPGroupByProcessGuidAndIndexRoute: route(emitter.EmitLatency(middleware.LogWrap(logger, actualLRPHandler.ActualLRPGroupByProcessGuidAndIndex))),

		// Actual LRP Lifecycle
		bbs.ClaimActualLRPRoute:  route(emitter.EmitLatency(middleware.LogWrap(logger, actualLRPLifecycleHandler.ClaimActualLRP))),
		bbs.StartActualLRPRoute:  route(emitter.EmitLatency(middleware.LogWrap(logger, actualLRPLifecycleHandler.StartActualLRP))),
		bbs.CrashActualLRPRoute:  route(emitter.EmitLatency(middleware.LogWrap(logger, actualLRPLifecycleHandler.CrashActualLRP))),
		bbs.RetireActualLRPRoute: route(emitter.EmitLatency(middleware.LogWrap(logger, actualLRPLifecycleHandler.RetireActualLRP))),
		bbs.FailActualLRPRoute:   route(emitter.EmitLatency(middleware.LogWrap(logger, actualLRPLifecycleHandler.FailActualLRP))),
		bbs.RemoveActualLRPRoute: route(emitter.EmitLatency(middleware.LogWrap(logger, actualLRPLifecycleHandler.RemoveActualLRP))),

		// Evacuation
		bbs.RemoveEvacuatingActualLRPRoute: route(emitter.EmitLatency(middleware.LogWrap(logger, evacuationHandler.RemoveEvacuatingActualLRP))),
		bbs.EvacuateClaimedActualLRPRoute:  route(emitter.EmitLatency(middleware.LogWrap(logger, evacuationHandler.EvacuateClaimedActualLRP))),
		bbs.EvacuateCrashedActualLRPRoute:  route(emitter.EmitLatency(middleware.LogWrap(logger, evacuationHandler.EvacuateCrashedActualLRP))),
		bbs.EvacuateStoppedActualLRPRoute:  route(emitter.EmitLatency(middleware.LogWrap(logger, evacuationHandler.EvacuateStoppedActualLRP))),
		bbs.EvacuateRunningActualLRPRoute:  route(emitter.EmitLatency(middleware.LogWrap(logger, evacuationHandler.EvacuateRunningActualLRP))),

		// Desired LRPs
		bbs.DesiredLRPsRoute:               route(emitter.EmitLatency(middleware.LogWrap(logger, desiredLRPHandler.DesiredLRPs))),
		bbs.DesiredLRPByProcessGuidRoute:   route(emitter.EmitLatency(middleware.LogWrap(logger, desiredLRPHandler.DesiredLRPByProcessGuid))),
		bbs.DesiredLRPSchedulingInfosRoute: route(emitter.EmitLatency(middleware.LogWrap(logger, desiredLRPHandler.DesiredLRPSchedulingInfos))),
		bbs.DesireDesiredLRPRoute:          route(emitter.EmitLatency(middleware.LogWrap(logger, desiredLRPHandler.DesireDesiredLRP))),
		bbs.UpdateDesiredLRPRoute:          route(emitter.EmitLatency(middleware.LogWrap(logger, desiredLRPHandler.UpdateDesiredLRP))),
		bbs.RemoveDesiredLRPRoute:          route(emitter.EmitLatency(middleware.LogWrap(logger, desiredLRPHandler.RemoveDesiredLRP))),

		bbs.DesiredLRPsRoute_r0:             route(emitter.EmitLatency(middleware.LogWrap(logger, desiredLRPHandler.DesiredLRPs_r0))),
		bbs.DesiredLRPsRoute_r1:             route(emitter.EmitLatency(middleware.LogWrap(logger, desiredLRPHandler.DesiredLRPs_r1))),
		bbs.DesiredLRPByProcessGuidRoute_r0: route(emitter.EmitLatency(middleware.LogWrap(logger, desiredLRPHandler.DesiredLRPByProcessGuid_r0))),
		bbs.DesiredLRPByProcessGuidRoute_r1: route(emitter.EmitLatency(middleware.LogWrap(logger, desiredLRPHandler.DesiredLRPByProcessGuid_r1))),
		bbs.DesireDesiredLRPRoute_r0:        route(emitter.EmitLatency(middleware.LogWrap(logger, desiredLRPHandler.DesireDesiredLRP_r0))),

		// Tasks
		bbs.TasksRoute:         route(emitter.EmitLatency(middleware.LogWrap(logger, taskHandler.Tasks))),
		bbs.TaskByGuidRoute:    route(emitter.EmitLatency(middleware.LogWrap(logger, taskHandler.TaskByGuid))),
		bbs.DesireTaskRoute:    route(emitter.EmitLatency(middleware.LogWrap(logger, taskHandler.DesireTask))),
		bbs.StartTaskRoute:     route(emitter.EmitLatency(middleware.LogWrap(logger, taskHandler.StartTask))),
		bbs.CancelTaskRoute:    route(emitter.EmitLatency(middleware.LogWrap(logger, taskHandler.CancelTask))),
		bbs.FailTaskRoute:      route(emitter.EmitLatency(middleware.LogWrap(logger, taskHandler.FailTask))),
		bbs.CompleteTaskRoute:  route(emitter.EmitLatency(middleware.LogWrap(logger, taskHandler.CompleteTask))),
		bbs.ResolvingTaskRoute: route(emitter.EmitLatency(middleware.LogWrap(logger, taskHandler.ResolvingTask))),
		bbs.DeleteTaskRoute:    route(emitter.EmitLatency(middleware.LogWrap(logger, taskHandler.DeleteTask))),

		bbs.TasksRoute_r1:      route(emitter.EmitLatency(middleware.LogWrap(logger, taskHandler.Tasks_r1))),
		bbs.TasksRoute_r0:      route(emitter.EmitLatency(middleware.LogWrap(logger, taskHandler.Tasks_r0))),
		bbs.TaskByGuidRoute_r1: route(emitter.EmitLatency(middleware.LogWrap(logger, taskHandler.TaskByGuid_r1))),
		bbs.TaskByGuidRoute_r0: route(emitter.EmitLatency(middleware.LogWrap(logger, taskHandler.TaskByGuid_r0))),
		bbs.DesireTaskRoute_r0: route(emitter.EmitLatency(middleware.LogWrap(logger, taskHandler.DesireTask_r0))),

		// Events
		bbs.EventStreamRoute_r0:        route(middleware.LogWrap(logger, eventsHandler.Subscribe_r0)),
		bbs.DesiredLRPEventStreamRoute: route(middleware.LogWrap(logger, eventsHandler.SubscribeToDesiredLRPEvents)),
		bbs.ActualLRPEventStreamRoute:  route(middleware.LogWrap(logger, eventsHandler.SubscribeToActualLRPEvents)),

		// Cells
		bbs.CellsRoute:    route(emitter.EmitLatency(middleware.LogWrap(logger, cellsHandler.Cells))),
		bbs.CellsRoute_r1: route(emitter.EmitLatency(middleware.LogWrap(logger, cellsHandler.Cells_r1))),
	}

	handler, err := rata.NewRouter(bbs.Routes, actions)
	if err != nil {
		panic("unable to create router: " + err.Error())
	}

	return middleware.RequestCountWrap(
		UnavailableWrap(handler,
			migrationsDone,
		),
	)
}

func route(f http.HandlerFunc) http.Handler {
	return f
}

func parseRequest(logger lager.Logger, req *http.Request, request MessageValidator) error {
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

	if err := request.Validate(); err != nil {
		logger.Error("invalid-request", err)
		return models.NewError(models.Error_InvalidRequest, err.Error())
	}

	return nil
}

func exitIfUnrecoverable(logger lager.Logger, exitCh chan<- struct{}, err *models.Error) {
	if err != nil && err.Type == models.Error_Unrecoverable {
		logger.Error("unrecoverable-error", err)
		select {
		case exitCh <- struct{}{}:
		default:
		}
	}
}

func writeResponse(w http.ResponseWriter, message proto.Message) {
	responseBytes, err := proto.Marshal(message)
	if err != nil {
		panic("Unable to encode Proto: " + err.Error())
	}

	w.Header().Set("Content-Length", strconv.Itoa(len(responseBytes)))
	w.Header().Set("Content-Type", "application/x-protobuf")
	w.WriteHeader(http.StatusOK)

	w.Write(responseBytes)
}
