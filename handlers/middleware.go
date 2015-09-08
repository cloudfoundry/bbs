package handlers

import (
	"net/http"

	"github.com/cloudfoundry/dropsonde"
	"github.com/pivotal-golang/lager"
)

func LogWrap(logger lager.Logger, handler http.Handler) http.HandlerFunc {
	handler = dropsonde.InstrumentedHandler(handler)

	return func(w http.ResponseWriter, r *http.Request) {
		requestLog := logger.Session("request", lager.Data{
			"method":  r.Method,
			"request": r.URL.String(),
		})

		requestLog.Info("serving")
		handler.ServeHTTP(w, r)
		requestLog.Info("done")
	}
}

func UnavailableWrap(handler http.Handler, serviceReady <-chan struct{}) http.HandlerFunc {
	handler = NewUnavailableHandler(handler, serviceReady)

	return func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
	}
}
