package handlers

import (
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/cloudfoundry-incubator/bbs"
	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/events"
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
		bbs.PingRoute: route(EmitLatency(pingHandler.Ping)),

		// Domains
		bbs.DomainsRoute:      route(EmitLatency(domainHandler.Domains)),
		bbs.UpsertDomainRoute: route(EmitLatency(domainHandler.Upsert)),

		// Actual LRPs
		bbs.ActualLRPGroupsRoute:                     route(EmitLatency(actualLRPHandler.ActualLRPGroups)),
		bbs.ActualLRPGroupsByProcessGuidRoute:        route(EmitLatency(actualLRPHandler.ActualLRPGroupsByProcessGuid)),
		bbs.ActualLRPGroupByProcessGuidAndIndexRoute: route(EmitLatency(actualLRPHandler.ActualLRPGroupByProcessGuidAndIndex)),

		// Actual LRP Lifecycle
		bbs.ClaimActualLRPRoute:  route(EmitLatency(actualLRPLifecycleHandler.ClaimActualLRP)),
		bbs.StartActualLRPRoute:  route(EmitLatency(actualLRPLifecycleHandler.StartActualLRP)),
		bbs.CrashActualLRPRoute:  route(EmitLatency(actualLRPLifecycleHandler.CrashActualLRP)),
		bbs.RetireActualLRPRoute: route(EmitLatency(actualLRPLifecycleHandler.RetireActualLRP)),
		bbs.FailActualLRPRoute:   route(EmitLatency(actualLRPLifecycleHandler.FailActualLRP)),
		bbs.RemoveActualLRPRoute: route(EmitLatency(actualLRPLifecycleHandler.RemoveActualLRP)),

		// Evacuation
		bbs.RemoveEvacuatingActualLRPRoute: route(EmitLatency(evacuationHandler.RemoveEvacuatingActualLRP)),
		bbs.EvacuateClaimedActualLRPRoute:  route(EmitLatency(evacuationHandler.EvacuateClaimedActualLRP)),
		bbs.EvacuateCrashedActualLRPRoute:  route(EmitLatency(evacuationHandler.EvacuateCrashedActualLRP)),
		bbs.EvacuateStoppedActualLRPRoute:  route(EmitLatency(evacuationHandler.EvacuateStoppedActualLRP)),
		bbs.EvacuateRunningActualLRPRoute:  route(EmitLatency(evacuationHandler.EvacuateRunningActualLRP)),

		// LRP Convergence
		bbs.ConvergeLRPsRoute: route(EmitLatency(lrpConvergenceHandler.ConvergeLRPs)),

		// Desired LRPs
		bbs.DesiredLRPsRoute:               route(EmitLatency(desiredLRPHandler.DesiredLRPs)),
		bbs.DesiredLRPByProcessGuidRoute:   route(EmitLatency(desiredLRPHandler.DesiredLRPByProcessGuid)),
		bbs.DesiredLRPSchedulingInfosRoute: route(EmitLatency(desiredLRPHandler.DesiredLRPSchedulingInfos)),
		bbs.DesireDesiredLRPRoute:          route(EmitLatency(desiredLRPHandler.DesireDesiredLRP)),
		bbs.UpdateDesiredLRPRoute:          route(EmitLatency(desiredLRPHandler.UpdateDesiredLRP)),
		bbs.RemoveDesiredLRPRoute:          route(EmitLatency(desiredLRPHandler.RemoveDesiredLRP)),

		// Tasks
		bbs.TasksRoute:         route(EmitLatency(taskHandler.Tasks)),
		bbs.TaskByGuidRoute:    route(EmitLatency(taskHandler.TaskByGuid)),
		bbs.DesireTaskRoute:    route(EmitLatency(taskHandler.DesireTask)),
		bbs.StartTaskRoute:     route(EmitLatency(taskHandler.StartTask)),
		bbs.CancelTaskRoute:    route(EmitLatency(taskHandler.CancelTask)),
		bbs.FailTaskRoute:      route(EmitLatency(taskHandler.FailTask)),
		bbs.CompleteTaskRoute:  route(EmitLatency(taskHandler.CompleteTask)),
		bbs.ResolvingTaskRoute: route(EmitLatency(taskHandler.ResolvingTask)),
		bbs.DeleteTaskRoute:    route(EmitLatency(taskHandler.DeleteTask)),
		bbs.ConvergeTasksRoute: route(EmitLatency(taskHandler.ConvergeTasks)),

		// Events
		bbs.EventStreamRoute: route(eventsHandler.Subscribe),
	}

	handler, err := rata.NewRouter(bbs.Routes, actions)
	if err != nil {
		panic("unable to create router: " + err.Error())
	}

	return RequestCountWrap(LogWrap(logger, UnavailableWrap(handler, migrationsDone)))
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
