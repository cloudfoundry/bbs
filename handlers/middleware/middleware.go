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

func LogWrap(logger lager.Logger, handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestLog := logger.Session("request", lager.Data{
			"method":  r.Method,
			"request": r.URL.String(),
		})

		requestLog.Debug("serving")
		handler.ServeHTTP(w, r)
		requestLog.Debug("done")
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
