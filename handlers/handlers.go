package handlers

import (
	"net/http"

	"github.com/cloudfoundry-incubator/bbs"
	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/events"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/rata"
)

func New(logger lager.Logger, db db.DB, hub events.Hub) http.Handler {
	domainHandler := NewDomainHandler(logger, db)
	actualLRPHandler := NewActualLRPHandler(logger, db)
	actualLRPLifecycleHandler := NewActualLRPLifecycleHandler(logger, db)
	desiredLRPHandler := NewDesiredLRPHandler(logger, db)
	taskHandler := NewTaskHandler(logger, db)
	eventsHandler := NewEventHandler(logger, hub)

	actions := rata.Handlers{
		// Domains
		bbs.DomainsRoute:      route(domainHandler.GetAll),
		bbs.UpsertDomainRoute: route(domainHandler.Upsert),

		// Actual LRPs
		bbs.ActualLRPGroupsRoute:                     route(actualLRPHandler.ActualLRPGroups),
		bbs.ActualLRPGroupsByProcessGuidRoute:        route(actualLRPHandler.ActualLRPGroupsByProcessGuid),
		bbs.ActualLRPGroupByProcessGuidAndIndexRoute: route(actualLRPHandler.ActualLRPGroupByProcessGuidAndIndex),
		bbs.ClaimActualLRPRoute:                      route(actualLRPLifecycleHandler.ClaimActualLRP),
		bbs.StartActualLRPRoute:                      route(actualLRPLifecycleHandler.StartActualLRP),
		bbs.CrashActualLRPRoute:                      route(actualLRPLifecycleHandler.CrashActualLRP),
		bbs.RetireActualLRPRoute:                     route(actualLRPLifecycleHandler.RetireActualLRP),
		bbs.FailActualLRPRoute:                       route(actualLRPLifecycleHandler.FailActualLRP),
		bbs.RemoveActualLRPRoute:                     route(actualLRPLifecycleHandler.RemoveActualLRP),

		// Desired LRPs
		bbs.DesiredLRPsRoute:             route(desiredLRPHandler.DesiredLRPs),
		bbs.DesiredLRPByProcessGuidRoute: route(desiredLRPHandler.DesiredLRPByProcessGuid),

		// Tasks
		bbs.TasksRoute:      route(taskHandler.Tasks),
		bbs.TaskByGuidRoute: route(taskHandler.TaskByGuid),

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
