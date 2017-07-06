package middleware

import (
	"net/http"
	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/runtimeschema/metric"
)

const (
	RequestLatency = metric.Duration("RequestLatency")
	RequestCount   = metric.Counter("RequestCount")
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

func NewLatencyEmitterWrapper(emitter Emitter) LatencyEmitterWrapper {
	return LatencyEmitterWrapper{emitter: emitter}
}

type LatencyEmitterWrapper struct {
	emitter Emitter
}

func (l LatencyEmitterWrapper) RecordLatency(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		f(w, r)
		l.emitter.UpdateLatency(time.Since(startTime))
	}
}

//go:generate counterfeiter -o fakes/fake_emitter.go . Emitter
type Emitter interface {
	IncrementCounter(delta int)
	UpdateLatency(latency time.Duration)
}

type defaultEmitter struct {
}

func (e *defaultEmitter) IncrementCounter(delta int) {
	RequestCount.Add(uint64(delta))
}

func (e *defaultEmitter) UpdateLatency(latency time.Duration) {
}

func RequestCountWrap(handler http.Handler) http.HandlerFunc {
	return RequestCountWrapWithCustomEmitter(handler, &defaultEmitter{})
}

func RequestCountWrapWithCustomEmitter(handler http.Handler, emitter Emitter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		emitter.IncrementCounter(1)
		handler.ServeHTTP(w, r)
	}
}
