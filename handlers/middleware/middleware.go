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
			defer requestLog.Debug("done")
			defer func() {
				requestAccessLogger.Info("done", lager.Data{"duration": time.Since(start)})
			}()
			loggableHandlerFunc(requestLog, w, r)
		}
	} else {
		return func(w http.ResponseWriter, r *http.Request) {
			requestLog := logger.Session("request", lagerDataFromReq(r))

			requestLog.Debug("serving")
			defer requestLog.Debug("done")

			loggableHandlerFunc(requestLog, w, r)

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
