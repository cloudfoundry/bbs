package handlers

import (
	"net/http"

	"github.com/cloudfoundry-incubator/bbs"
	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/events"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/rata"
)

func New(db db.DB, hub events.Hub, logger lager.Logger) http.Handler {
	domainHandler := NewDomainHandler(db, logger)
	actualLRPHandler := NewActualLRPHandler(db, logger)
	desiredLRPHandler := NewDesiredLRPHandler(db, logger)
	eventsHandler := NewEventHandler(hub, logger)

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

	return LogWrap(handler, logger)
}

func route(f func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return http.HandlerFunc(f)
}
