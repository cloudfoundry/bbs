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
	logger, accessLogger lager.Logger,
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
		bbs.PingRoute: emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, pingHandler.Ping)),

		// Domains
		bbs.DomainsRoute:      route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, domainHandler.Domains))),
		bbs.UpsertDomainRoute: route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, domainHandler.Upsert))),

		// Actual LRPs
		bbs.ActualLRPGroupsRoute:                     route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, actualLRPHandler.ActualLRPGroups))),
		bbs.ActualLRPGroupsByProcessGuidRoute:        route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, actualLRPHandler.ActualLRPGroupsByProcessGuid))),
		bbs.ActualLRPGroupByProcessGuidAndIndexRoute: route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, actualLRPHandler.ActualLRPGroupByProcessGuidAndIndex))),

		// Actual LRP Lifecycle
		bbs.ClaimActualLRPRoute:  route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, actualLRPLifecycleHandler.ClaimActualLRP))),
		bbs.StartActualLRPRoute:  route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, actualLRPLifecycleHandler.StartActualLRP))),
		bbs.CrashActualLRPRoute:  route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, actualLRPLifecycleHandler.CrashActualLRP))),
		bbs.RetireActualLRPRoute: route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, actualLRPLifecycleHandler.RetireActualLRP))),
		bbs.FailActualLRPRoute:   route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, actualLRPLifecycleHandler.FailActualLRP))),
		bbs.RemoveActualLRPRoute: route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, actualLRPLifecycleHandler.RemoveActualLRP))),

		// Evacuation
		bbs.RemoveEvacuatingActualLRPRoute: route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, evacuationHandler.RemoveEvacuatingActualLRP))),
		bbs.EvacuateClaimedActualLRPRoute:  route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, evacuationHandler.EvacuateClaimedActualLRP))),
		bbs.EvacuateCrashedActualLRPRoute:  route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, evacuationHandler.EvacuateCrashedActualLRP))),
		bbs.EvacuateStoppedActualLRPRoute:  route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, evacuationHandler.EvacuateStoppedActualLRP))),
		bbs.EvacuateRunningActualLRPRoute:  route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, evacuationHandler.EvacuateRunningActualLRP))),

		// Desired LRPs
		bbs.DesiredLRPsRoute:               route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, desiredLRPHandler.DesiredLRPs))),
		bbs.DesiredLRPByProcessGuidRoute:   route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, desiredLRPHandler.DesiredLRPByProcessGuid))),
		bbs.DesiredLRPSchedulingInfosRoute: route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, desiredLRPHandler.DesiredLRPSchedulingInfos))),
		bbs.DesireDesiredLRPRoute:          route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, desiredLRPHandler.DesireDesiredLRP))),
		bbs.UpdateDesiredLRPRoute:          route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, desiredLRPHandler.UpdateDesiredLRP))),
		bbs.RemoveDesiredLRPRoute:          route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, desiredLRPHandler.RemoveDesiredLRP))),

		bbs.DesiredLRPsRoute_r0:             route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, desiredLRPHandler.DesiredLRPs_r0))),
		bbs.DesiredLRPsRoute_r1:             route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, desiredLRPHandler.DesiredLRPs_r1))),
		bbs.DesiredLRPByProcessGuidRoute_r0: route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, desiredLRPHandler.DesiredLRPByProcessGuid_r0))),
		bbs.DesiredLRPByProcessGuidRoute_r1: route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, desiredLRPHandler.DesiredLRPByProcessGuid_r1))),
		bbs.DesireDesiredLRPRoute_r0:        route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, desiredLRPHandler.DesireDesiredLRP_r0))),
		bbs.DesireDesiredLRPRoute_r1:        route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, desiredLRPHandler.DesireDesiredLRP_r1))),

		// Tasks
		bbs.TasksRoute:         route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, taskHandler.Tasks))),
		bbs.TaskByGuidRoute:    route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, taskHandler.TaskByGuid))),
		bbs.DesireTaskRoute:    route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, taskHandler.DesireTask))),
		bbs.StartTaskRoute:     route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, taskHandler.StartTask))),
		bbs.CancelTaskRoute:    route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, taskHandler.CancelTask))),
		bbs.FailTaskRoute:      route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, taskHandler.FailTask))),
		bbs.CompleteTaskRoute:  route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, taskHandler.CompleteTask))),
		bbs.ResolvingTaskRoute: route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, taskHandler.ResolvingTask))),
		bbs.DeleteTaskRoute:    route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, taskHandler.DeleteTask))),

		bbs.TasksRoute_r1:      route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, taskHandler.Tasks_r1))),
		bbs.TasksRoute_r0:      route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, taskHandler.Tasks_r0))),
		bbs.TaskByGuidRoute_r1: route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, taskHandler.TaskByGuid_r1))),
		bbs.TaskByGuidRoute_r0: route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, taskHandler.TaskByGuid_r0))),
		bbs.DesireTaskRoute_r1: route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, taskHandler.DesireTask_r1))),
		bbs.DesireTaskRoute_r0: route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, taskHandler.DesireTask_r0))),

		// Events
		bbs.EventStreamRoute_r0: route(middleware.LogWrap(logger, accessLogger, eventsHandler.Subscribe_r0)),

		// Cells
		bbs.CellsRoute:    route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, cellsHandler.Cells))),
		bbs.CellsRoute_r1: route(emitter.EmitLatency(middleware.LogWrap(logger, accessLogger, cellsHandler.Cells))),
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
