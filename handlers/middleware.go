package handlers

import (
	"net/http"
	"time"

	"github.com/cloudfoundry-incubator/runtime-schema/metric"
	"github.com/pivotal-golang/lager"
)

const (
	requestLatency = metric.Duration("RequestLatency")
	requestCount   = metric.Counter("RequestCount")
)

func LogWrap(logger lager.Logger, handler http.Handler) http.HandlerFunc {
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

func MeasureWrap(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestCount.Increment()
		startTime := time.Now()
		handler.ServeHTTP(w, r)
		requestLatency.Send(time.Since(startTime))
	}
}
