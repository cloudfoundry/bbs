package handlers

import (
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/cloudfoundry-incubator/bbs"
	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/events"
	"github.com/cloudfoundry-incubator/bbs/handlers/middleware"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/gogo/protobuf/proto"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/rata"
)

func New(logger lager.Logger, db db.DB, desiredHub, actualHub, taskHub events.Hub, serviceClient bbs.ServiceClient, migrationsDone <-chan struct{}) http.Handler {
	pingHandler := NewPingHandler(logger)
	domainHandler := NewDomainHandler(logger, db)
	actualLRPHandler := NewActualLRPHandler(logger, db)
	actualLRPLifecycleHandler := NewActualLRPLifecycleHandler(logger, db)
	evacuationHandler := NewEvacuationHandler(logger, db)
	desiredLRPHandler := NewDesiredLRPHandler(logger, db)
	lrpConvergenceHandler := NewLRPConvergenceHandler(logger, db)
	taskHandler := NewTaskHandler(logger, db)
	eventsHandler := NewEventHandler(logger, desiredHub, actualHub, taskHub)
	cellsHandler := NewCellHandler(logger, serviceClient)

	emitter := middleware.NewLatencyEmitter(logger)

	actions := rata.Handlers{
		// Ping
		bbs.PingRoute: route(emitter.EmitLatency(pingHandler.Ping)),

		// Domains
		bbs.DomainsRoute:      route(emitter.EmitLatency(domainHandler.Domains)),
		bbs.UpsertDomainRoute: route(emitter.EmitLatency(domainHandler.Upsert)),

		// Actual LRPs
		bbs.ActualLRPGroupsRoute:                     route(emitter.EmitLatency(actualLRPHandler.ActualLRPGroups)),
		bbs.ActualLRPGroupsByProcessGuidRoute:        route(emitter.EmitLatency(actualLRPHandler.ActualLRPGroupsByProcessGuid)),
		bbs.ActualLRPGroupByProcessGuidAndIndexRoute: route(emitter.EmitLatency(actualLRPHandler.ActualLRPGroupByProcessGuidAndIndex)),

		// Actual LRP Lifecycle
		bbs.ClaimActualLRPRoute:  route(emitter.EmitLatency(actualLRPLifecycleHandler.ClaimActualLRP)),
		bbs.StartActualLRPRoute:  route(emitter.EmitLatency(actualLRPLifecycleHandler.StartActualLRP)),
		bbs.CrashActualLRPRoute:  route(emitter.EmitLatency(actualLRPLifecycleHandler.CrashActualLRP)),
		bbs.RetireActualLRPRoute: route(emitter.EmitLatency(actualLRPLifecycleHandler.RetireActualLRP)),
		bbs.FailActualLRPRoute:   route(emitter.EmitLatency(actualLRPLifecycleHandler.FailActualLRP)),
		bbs.RemoveActualLRPRoute: route(emitter.EmitLatency(actualLRPLifecycleHandler.RemoveActualLRP)),

		// Evacuation
		bbs.RemoveEvacuatingActualLRPRoute: route(emitter.EmitLatency(evacuationHandler.RemoveEvacuatingActualLRP)),
		bbs.EvacuateClaimedActualLRPRoute:  route(emitter.EmitLatency(evacuationHandler.EvacuateClaimedActualLRP)),
		bbs.EvacuateCrashedActualLRPRoute:  route(emitter.EmitLatency(evacuationHandler.EvacuateCrashedActualLRP)),
		bbs.EvacuateStoppedActualLRPRoute:  route(emitter.EmitLatency(evacuationHandler.EvacuateStoppedActualLRP)),
		bbs.EvacuateRunningActualLRPRoute:  route(emitter.EmitLatency(evacuationHandler.EvacuateRunningActualLRP)),

		// LRP Convergence
		bbs.ConvergeLRPsRoute: route(emitter.EmitLatency(lrpConvergenceHandler.ConvergeLRPs)),

		// Desired LRPs
		bbs.DesiredLRPsRoute:               route(emitter.EmitLatency(desiredLRPHandler.DesiredLRPs)),
		bbs.DesiredLRPByProcessGuidRoute:   route(emitter.EmitLatency(desiredLRPHandler.DesiredLRPByProcessGuid)),
		bbs.DesiredLRPSchedulingInfosRoute: route(emitter.EmitLatency(desiredLRPHandler.DesiredLRPSchedulingInfos)),
		bbs.DesireDesiredLRPRoute:          route(emitter.EmitLatency(desiredLRPHandler.DesireDesiredLRP)),
		bbs.UpdateDesiredLRPRoute:          route(emitter.EmitLatency(desiredLRPHandler.UpdateDesiredLRP)),
		bbs.RemoveDesiredLRPRoute:          route(emitter.EmitLatency(desiredLRPHandler.RemoveDesiredLRP)),

		bbs.DesiredLRPsRoute_r0:             route(emitter.EmitLatency(desiredLRPHandler.DesiredLRPs_r0)),
		bbs.DesiredLRPByProcessGuidRoute_r0: route(emitter.EmitLatency(desiredLRPHandler.DesiredLRPByProcessGuid_r0)),

		// Tasks
		bbs.TasksRoute:         route(emitter.EmitLatency(taskHandler.Tasks)),
		bbs.TaskByGuidRoute:    route(emitter.EmitLatency(taskHandler.TaskByGuid)),
		bbs.DesireTaskRoute:    route(emitter.EmitLatency(taskHandler.DesireTask)),
		bbs.StartTaskRoute:     route(emitter.EmitLatency(taskHandler.StartTask)),
		bbs.CancelTaskRoute:    route(emitter.EmitLatency(taskHandler.CancelTask)),
		bbs.FailTaskRoute:      route(emitter.EmitLatency(taskHandler.FailTask)),
		bbs.CompleteTaskRoute:  route(emitter.EmitLatency(taskHandler.CompleteTask)),
		bbs.ResolvingTaskRoute: route(emitter.EmitLatency(taskHandler.ResolvingTask)),
		bbs.DeleteTaskRoute:    route(emitter.EmitLatency(taskHandler.DeleteTask)),
		bbs.ConvergeTasksRoute: route(emitter.EmitLatency(taskHandler.ConvergeTasks)),

		bbs.TasksRoute_r0:      route(emitter.EmitLatency(taskHandler.Tasks_r0)),
		bbs.TaskByGuidRoute_r0: route(emitter.EmitLatency(taskHandler.TaskByGuid_r0)),

		// Events
		bbs.EventStreamRoute_r0:        route(eventsHandler.Subscribe_r0),
		bbs.DesiredLRPEventStreamRoute: route(eventsHandler.SubscribeToDesiredLRPEvents),
		bbs.ActualLRPEventStreamRoute:  route(eventsHandler.SubscribeToActualLRPEvents),
		bbs.TaskEventStreamRoute:       route(eventsHandler.SubscribeToTaskEvents),

		// Cells
		bbs.CellsRoute: route(emitter.EmitLatency(cellsHandler.Cells)),
	}

	handler, err := rata.NewRouter(bbs.Routes, actions)
	if err != nil {
		panic("unable to create router: " + err.Error())
	}

	return middleware.RequestCountWrap(middleware.LogWrap(logger, UnavailableWrap(handler, migrationsDone)))
}

func route(f http.HandlerFunc) http.Handler {
	return http.HandlerFunc(f)
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
