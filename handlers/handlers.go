package handlers

import (
	"net/http"

	"github.com/cloudfoundry-incubator/bbs"
	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/rata"
)

func New(domainDB db.DomainDB, logger lager.Logger) http.Handler {
	domainHandler := NewDomainHandler(domainDB, logger)

	actions := rata.Handlers{
		// Domains
		bbs.DomainsRoute: route(domainHandler.GetAll),
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
