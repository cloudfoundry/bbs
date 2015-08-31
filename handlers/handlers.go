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

func New(logger lager.Logger, db db.DB, hub events.Hub) http.Handler {
	domainHandler := NewDomainHandler(logger, db)
	actualLRPHandler := NewActualLRPHandler(logger, db)
	actualLRPLifecycleHandler := NewActualLRPLifecycleHandler(logger, db)
	evacuationHandler := NewEvacuationHandler(logger, db)
	desiredLRPHandler := NewDesiredLRPHandler(logger, db)
	lrpConvergenceHandler := NewLRPConvergenceHandler(logger, db)
	taskHandler := NewTaskHandler(logger, db)
	eventsHandler := NewEventHandler(logger, hub)

	actions := rata.Handlers{
		// Domains
		bbs.DomainsRoute:      route(domainHandler.Domains),
		bbs.UpsertDomainRoute: route(domainHandler.Upsert),

		// Actual LRPs
		bbs.ActualLRPGroupsRoute:                     route(actualLRPHandler.ActualLRPGroups),
		bbs.ActualLRPGroupsByProcessGuidRoute:        route(actualLRPHandler.ActualLRPGroupsByProcessGuid),
		bbs.ActualLRPGroupByProcessGuidAndIndexRoute: route(actualLRPHandler.ActualLRPGroupByProcessGuidAndIndex),

		// Actual LRP Lifecycle
		bbs.ClaimActualLRPRoute:  route(actualLRPLifecycleHandler.ClaimActualLRP),
		bbs.StartActualLRPRoute:  route(actualLRPLifecycleHandler.StartActualLRP),
		bbs.CrashActualLRPRoute:  route(actualLRPLifecycleHandler.CrashActualLRP),
		bbs.RetireActualLRPRoute: route(actualLRPLifecycleHandler.RetireActualLRP),
		bbs.FailActualLRPRoute:   route(actualLRPLifecycleHandler.FailActualLRP),
		bbs.RemoveActualLRPRoute: route(actualLRPLifecycleHandler.RemoveActualLRP),

		// Evacuation
		bbs.RemoveEvacuatingActualLRPRoute: route(evacuationHandler.RemoveEvacuatingActualLRP),
		bbs.EvacuateClaimedActualLRPRoute:  route(evacuationHandler.EvacuateClaimedActualLRP),
		bbs.EvacuateCrashedActualLRPRoute:  route(evacuationHandler.EvacuateCrashedActualLRP),
		bbs.EvacuateStoppedActualLRPRoute:  route(evacuationHandler.EvacuateStoppedActualLRP),
		bbs.EvacuateRunningActualLRPRoute:  route(evacuationHandler.EvacuateRunningActualLRP),

		// LRP Convergence
		bbs.ConvergeLRPsRoute: route(lrpConvergenceHandler.ConvergeLRPs),

		// Desired LRPs
		bbs.DesiredLRPsRoute:             route(desiredLRPHandler.DesiredLRPs),
		bbs.DesiredLRPByProcessGuidRoute: route(desiredLRPHandler.DesiredLRPByProcessGuid),
		bbs.DesireDesiredLRPRoute:        route(desiredLRPHandler.DesireDesiredLRP),
		bbs.UpdateDesiredLRPRoute:        route(desiredLRPHandler.UpdateDesiredLRP),
		bbs.RemoveDesiredLRPRoute:        route(desiredLRPHandler.RemoveDesiredLRP),

		// Tasks
		bbs.TasksRoute:         route(taskHandler.Tasks),
		bbs.TaskByGuidRoute:    route(taskHandler.TaskByGuid),
		bbs.DesireTaskRoute:    route(taskHandler.DesireTask),
		bbs.StartTaskRoute:     route(taskHandler.StartTask),
		bbs.CancelTaskRoute:    route(taskHandler.CancelTask),
		bbs.FailTaskRoute:      route(taskHandler.FailTask),
		bbs.CompleteTaskRoute:  route(taskHandler.CompleteTask),
		bbs.ResolvingTaskRoute: route(taskHandler.ResolvingTask),
		bbs.DeleteTaskRoute:    route(taskHandler.DeleteTask),
		bbs.ConvergeTasksRoute: route(taskHandler.ConvergeTasks),

		// Events
		bbs.EventStreamRoute: route(eventsHandler.Subscribe),
	}

	handler, err := rata.NewRouter(bbs.Routes, actions)
	if err != nil {
		panic("unable to create router: " + err.Error())
	}

	return LogWrap(logger, handler)
}

func route(f func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return http.HandlerFunc(f)
}

func parseRequest(logger lager.Logger, req *http.Request, request MessageValidator) *models.Error {
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

	logger.Debug("parsed-request-body", lager.Data{"request": request})
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
