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

func New(logger lager.Logger, db db.DB, hub events.Hub, migrationsDone <-chan struct{}) http.Handler {
	pingHandler := NewPingHandler(logger)
	domainHandler := NewDomainHandler(logger, db)
	actualLRPHandler := NewActualLRPHandler(logger, db)
	actualLRPLifecycleHandler := NewActualLRPLifecycleHandler(logger, db)
	evacuationHandler := NewEvacuationHandler(logger, db)
	desiredLRPHandler := NewDesiredLRPHandler(logger, db)
	lrpConvergenceHandler := NewLRPConvergenceHandler(logger, db)
	taskHandler := NewTaskHandler(logger, db)
	eventsHandler := NewEventHandler(logger, hub)

	actions := rata.Handlers{
		// Ping
		bbs.PingRoute: route(middleware.EmitLatency(pingHandler.Ping)),

		// Domains
		bbs.DomainsRoute:      route(middleware.EmitLatency(domainHandler.Domains)),
		bbs.UpsertDomainRoute: route(middleware.EmitLatency(domainHandler.Upsert)),

		// Actual LRPs
		bbs.ActualLRPGroupsRoute:                     route(middleware.EmitLatency(actualLRPHandler.ActualLRPGroups)),
		bbs.ActualLRPGroupsByProcessGuidRoute:        route(middleware.EmitLatency(actualLRPHandler.ActualLRPGroupsByProcessGuid)),
		bbs.ActualLRPGroupByProcessGuidAndIndexRoute: route(middleware.EmitLatency(actualLRPHandler.ActualLRPGroupByProcessGuidAndIndex)),

		// Actual LRP Lifecycle
		bbs.ClaimActualLRPRoute:  route(middleware.EmitLatency(actualLRPLifecycleHandler.ClaimActualLRP)),
		bbs.StartActualLRPRoute:  route(middleware.EmitLatency(actualLRPLifecycleHandler.StartActualLRP)),
		bbs.CrashActualLRPRoute:  route(middleware.EmitLatency(actualLRPLifecycleHandler.CrashActualLRP)),
		bbs.RetireActualLRPRoute: route(middleware.EmitLatency(actualLRPLifecycleHandler.RetireActualLRP)),
		bbs.FailActualLRPRoute:   route(middleware.EmitLatency(actualLRPLifecycleHandler.FailActualLRP)),
		bbs.RemoveActualLRPRoute: route(middleware.EmitLatency(actualLRPLifecycleHandler.RemoveActualLRP)),

		// Evacuation
		bbs.RemoveEvacuatingActualLRPRoute: route(middleware.EmitLatency(evacuationHandler.RemoveEvacuatingActualLRP)),
		bbs.EvacuateClaimedActualLRPRoute:  route(middleware.EmitLatency(evacuationHandler.EvacuateClaimedActualLRP)),
		bbs.EvacuateCrashedActualLRPRoute:  route(middleware.EmitLatency(evacuationHandler.EvacuateCrashedActualLRP)),
		bbs.EvacuateStoppedActualLRPRoute:  route(middleware.EmitLatency(evacuationHandler.EvacuateStoppedActualLRP)),
		bbs.EvacuateRunningActualLRPRoute:  route(middleware.EmitLatency(evacuationHandler.EvacuateRunningActualLRP)),

		// LRP Convergence
		bbs.ConvergeLRPsRoute: route(middleware.EmitLatency(lrpConvergenceHandler.ConvergeLRPs)),

		// Desired LRPs
		bbs.DesiredLRPsRoute:               route(middleware.EmitLatency(desiredLRPHandler.DesiredLRPs)),
		bbs.DesiredLRPByProcessGuidRoute:   route(middleware.EmitLatency(desiredLRPHandler.DesiredLRPByProcessGuid)),
		bbs.DesiredLRPSchedulingInfosRoute: route(middleware.EmitLatency(desiredLRPHandler.DesiredLRPSchedulingInfos)),
		bbs.DesireDesiredLRPRoute:          route(middleware.EmitLatency(desiredLRPHandler.DesireDesiredLRP)),
		bbs.UpdateDesiredLRPRoute:          route(middleware.EmitLatency(desiredLRPHandler.UpdateDesiredLRP)),
		bbs.RemoveDesiredLRPRoute:          route(middleware.EmitLatency(desiredLRPHandler.RemoveDesiredLRP)),

		bbs.DesiredLRPsRoute_r0:             route(middleware.EmitLatency(desiredLRPHandler.DesiredLRPs_r0)),
		bbs.DesiredLRPByProcessGuidRoute_r0: route(middleware.EmitLatency(desiredLRPHandler.DesiredLRPByProcessGuid_r0)),

		// Tasks
		bbs.TasksRoute:         route(middleware.EmitLatency(taskHandler.Tasks)),
		bbs.TaskByGuidRoute:    route(middleware.EmitLatency(taskHandler.TaskByGuid)),
		bbs.DesireTaskRoute:    route(middleware.EmitLatency(taskHandler.DesireTask)),
		bbs.StartTaskRoute:     route(middleware.EmitLatency(taskHandler.StartTask)),
		bbs.CancelTaskRoute:    route(middleware.EmitLatency(taskHandler.CancelTask)),
		bbs.FailTaskRoute:      route(middleware.EmitLatency(taskHandler.FailTask)),
		bbs.CompleteTaskRoute:  route(middleware.EmitLatency(taskHandler.CompleteTask)),
		bbs.ResolvingTaskRoute: route(middleware.EmitLatency(taskHandler.ResolvingTask)),
		bbs.DeleteTaskRoute:    route(middleware.EmitLatency(taskHandler.DeleteTask)),
		bbs.ConvergeTasksRoute: route(middleware.EmitLatency(taskHandler.ConvergeTasks)),

		bbs.TasksRoute_r0:      route(middleware.EmitLatency(taskHandler.Tasks_r0)),
		bbs.TaskByGuidRoute_r0: route(middleware.EmitLatency(taskHandler.TaskByGuid_r0)),

		// Events
		bbs.EventStreamRoute: route(eventsHandler.Subscribe),
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
