package middleware

import (
	"net/http"
	"time"

	"code.cloudfoundry.org/lager/v3"
)

type LoggableHandlerFunc func(logger lager.Logger, w http.ResponseWriter, r *http.Request)

//go:generate counterfeiter -generate

//counterfeiter:generate -o fakes/fake_emitter.go . Emitter
type Emitter interface {
	IncrementRequestCounter(delta int)
	UpdateLatency(latency time.Duration)
}

func LogWrap(logger, accessLogger lager.Logger, loggableHandlerFunc LoggableHandlerFunc) http.HandlerFunc {
	lagerDataFromReq := func(r *http.Request) lager.Data {
		return lager.Data{
			"method":      r.Method,
			"remote_addr": r.RemoteAddr,
			"request":     r.URL.String(),
		}
	}

	if accessLogger != nil {
		return func(w http.ResponseWriter, r *http.Request) {
			requestLog := logger.Session("request")
			requestAccessLogger := accessLogger.Session("request")

			requestAccessLogger.Info("serving", lagerDataFromReq(r))
			requestLog.Debug("serving", lagerDataFromReq(r))

			start := time.Now()
			defer requestLog.Debug("done", lagerDataFromReq(r))
			defer func() {
				requestTime := time.Since(start)
				lagerData := lagerDataFromReq(r)
				lagerData["duration"] = requestTime
				requestAccessLogger.Info("done", lagerData)
			}()
			loggableHandlerFunc(requestLog, w, r)
		}
	} else {
		return func(w http.ResponseWriter, r *http.Request) {
			requestLog := logger.Session("request")

			requestLog.Debug("serving", lagerDataFromReq(r))
			defer requestLog.Debug("done", lagerDataFromReq(r))

			loggableHandlerFunc(requestLog, w, r)
		}
	}
}

func RecordLatency(f http.HandlerFunc, emitter Emitter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		f(w, r)
		emitter.UpdateLatency(time.Since(startTime))
	}
}

func RecordRequestCount(handler http.Handler, emitter Emitter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		emitter.IncrementRequestCounter(1)
		handler.ServeHTTP(w, r)
	}
}
