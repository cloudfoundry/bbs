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
	desiredLRPHandler := NewDesiredLRPHandler(logger, db)
	eventsHandler := NewEventHandler(logger, hub)

	actions := rata.Handlers{
		// Domains
		bbs.DomainsRoute:      route(domainHandler.GetAll),
		bbs.UpsertDomainRoute: route(domainHandler.Upsert),

		// Actual LRPs
		bbs.ActualLRPGroupsRoute:                     route(actualLRPHandler.ActualLRPGroups),
		bbs.ActualLRPGroupsByProcessGuidRoute:        route(actualLRPHandler.ActualLRPGroupsByProcessGuid),
		bbs.ActualLRPGroupByProcessGuidAndIndexRoute: route(actualLRPHandler.ActualLRPGroupByProcessGuidAndIndex),

		// Desired LRPs
		bbs.DesiredLRPsRoute:             route(desiredLRPHandler.DesiredLRPs),
		bbs.DesiredLRPByProcessGuidRoute: route(desiredLRPHandler.DesiredLRPByProcessGuid),

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
