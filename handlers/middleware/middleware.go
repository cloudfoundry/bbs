package middleware

import (
	"net/http"
	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/runtimeschema/metric"
)

const (
	requestLatency = metric.Duration("RequestLatency")
	requestCount   = metric.Counter("RequestCount")
)

type LoggableHandlerFunc func(logger lager.Logger, w http.ResponseWriter, r *http.Request)

func LogWrap(logger, accessLogger lager.Logger, loggableHandlerFunc LoggableHandlerFunc) http.HandlerFunc {
	lagerDataFromReq := func(r *http.Request) lager.Data {
		return lager.Data{
			"method":  r.Method,
			"request": r.URL.String(),
		}
	}

	if accessLogger != nil {
		return func(w http.ResponseWriter, r *http.Request) {
			requestLog := logger.Session("request", lagerDataFromReq(r))
			requestAccessLogger := accessLogger.Session("request", lagerDataFromReq(r))

			requestAccessLogger.Info("serving")
			requestLog.Debug("serving")

			start := time.Now()
			loggableHandlerFunc(requestLog, w, r)

			defer requestAccessLogger.Info("done", lager.Data{"duration": time.Since(start)})
			defer requestLog.Debug("done")
		}
	} else {
		return func(w http.ResponseWriter, r *http.Request) {
			requestLog := logger.Session("request", lagerDataFromReq(r))

			requestLog.Debug("serving")

			loggableHandlerFunc(requestLog, w, r)

			defer requestLog.Debug("done")
		}
	}
}

func NewLatencyEmitter(logger lager.Logger) LatencyEmitter {
	return LatencyEmitter{
		logger: logger,
	}
}

type LatencyEmitter struct {
	logger lager.Logger
}

func (l LatencyEmitter) EmitLatency(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		f(w, r)
		err := requestLatency.Send(time.Since(startTime))
		if err != nil {
			l.logger.Error("failed-to-send-request-latency-metric", err)
		}
	}
}

func RequestCountWrap(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestCount.Increment()
		handler.ServeHTTP(w, r)
	}
}
