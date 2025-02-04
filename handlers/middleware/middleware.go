package middleware

import (
	"code.cloudfoundry.org/bbs/cmd/bbs/config"
	"net/http"
	"slices"
	"time"

	"code.cloudfoundry.org/lager/v3"
)

type LoggableHandlerFunc func(logger lager.Logger, w http.ResponseWriter, r *http.Request)

//go:generate counterfeiter -generate

//counterfeiter:generate -o fakes/fake_emitter.go . Emitter
type Emitter interface {
	IncrementRequestCounter(delta int, route string)
	UpdateLatency(latency time.Duration, route string)
	GetAdvancedMetricsConfig() config.AdvancedMetrics
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
		emitter.UpdateLatency(time.Since(startTime), "")
	}
}


//TODO: remove this function
//func RecordRequestCount(handler http.Handler, emitter Emitter) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		emitter.IncrementRequestCounter(1)
//		handler.ServeHTTP(w, r)
//	}
//}

func RecordRequestCount(f http.HandlerFunc, emitter Emitter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		emitter.IncrementRequestCounter(1, "")
		f(w, r)
	}
}

func RecordDefaultMetrics(f http.HandlerFunc, emitter Emitter) http.HandlerFunc {
	return RecordLatency(RecordRequestCount(f, emitter), emitter)

	return func(w http.ResponseWriter, r *http.Request) {
		emitter.IncrementRequestCounter(1, "")
		f(w, r)
	}

	RecordAdvancedMetrics(RecordRequestCount(RecordLatency(f, emitter), emitter), emitter, route)
	RecordAdvancedMetrics(RecordRequestCount(RecordLatency(f, emitter), emitter), emitter, route)
}

func RecordMetrics(f http.HandlerFunc, emitter Emitter, calledRoute string) http.HandlerFunc {
	//TODO: Adpat this to return a function
	// Record Default Metrics
	RecordLatency(f, emitter)
	RecordRequestCount(f, emitter)

	advancedMetricsConfig := emitter.GetAdvancedMetricsConfig()
	if advancedMetricsConfig.Enabled {
		RecordAdvancedMetrics(f, emitter, calledRoute)
	}
}

func RecordAdvancedMetrics(f http.HandlerFunc, emitter Emitter,  calledRoute string) http.HandlerFunc {
	isRouteFound := slices.Contains(emitter.GetAdvancedMetricsConfig().RouteConfig.RequestCountRoutes, calledRoute)
	if isRouteFound {
		emitter.IncrementRequestCounter(1)
	}

	isRouteFound = slices.Contains(emitter.GetAdvancedMetricsConfig().RouteConfig.RequestLatencyRoutes, calledRoute)


	//
	//if indexFound == -1 {
	//	return




	return func(w http.ResponseWriter, r *http.Request) {
		emitter.IncrementRequestCounter(1)
		f(w, r)
	}


	// 1. Execute default metrics
	// 2. read config
	// 3. If enabled record metrics for endpoints
	//3. Retrieve metadata (What's the current endpoint?)
	//3. Create a structure that keeps a mapping between each endpoint and it's notifier
}


