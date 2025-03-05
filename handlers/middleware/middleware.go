package middleware

import (
	"net/http"
	"slices"
	"time"

	"code.cloudfoundry.org/bbs/cmd/bbs/config"

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

type metadata struct {
	latency time.Duration
}

type handlerWithMetadata func(w http.ResponseWriter, r *http.Request) metadata

func initHandlerWithMetadata(f http.HandlerFunc) handlerWithMetadata {
	return func(w http.ResponseWriter, r *http.Request) metadata {
		f(w, r)
		return metadata{}
	}
}

func stripMetadata(f handlerWithMetadata) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f(w, r)
	}
}

func recordLatency(f handlerWithMetadata) handlerWithMetadata {
	return func(w http.ResponseWriter, r *http.Request) metadata {
		startTime := time.Now()
		metadata := f(w, r)
		metadata.latency = time.Since(startTime)
		return metadata
	}
}

func updateLatency(f handlerWithMetadata, emitter Emitter, route string) handlerWithMetadata {
	return func(w http.ResponseWriter, r *http.Request) metadata {
		metadata := f(w, r)
		emitter.UpdateLatency(metadata.latency, route)
		return metadata
	}
}

func updateRequestCount(f handlerWithMetadata, emitter Emitter, route string) handlerWithMetadata {
	return func(w http.ResponseWriter, r *http.Request) metadata {
		emitter.IncrementRequestCounter(1, route)
		return f(w, r)
	}
}

func UpdateRequestCount(f http.HandlerFunc, emitter Emitter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		emitter.IncrementRequestCounter(1, "")
		f(w, r)
	}
}

func RecordMetrics(f http.HandlerFunc, emitter Emitter, calledRoute string) http.HandlerFunc {
	// Record Default Metrics
	handlerMeta := initHandlerWithMetadata(f)
	handlerMeta = recordLatency(handlerMeta)

	handlerMeta = updateLatency(handlerMeta, emitter, "")
	handlerMeta = updateRequestCount(handlerMeta, emitter, "")

	// Record Advanced Metrics
	advancedMetricsConfig := emitter.GetAdvancedMetricsConfig()
	if advancedMetricsConfig.Enabled {
		handlerMeta = recordAdvancedMetrics(handlerMeta, emitter, calledRoute)
	}

	return stripMetadata(handlerMeta)
}

func recordAdvancedMetrics(handlerMeta handlerWithMetadata, emitter Emitter, calledRoute string) handlerWithMetadata {
	isRouteFound := slices.Contains(emitter.GetAdvancedMetricsConfig().RouteConfig.RequestCountRoutes, calledRoute)
	if isRouteFound {
		handlerMeta = updateRequestCount(handlerMeta, emitter, calledRoute)
	}

	isRouteFound = slices.Contains(emitter.GetAdvancedMetricsConfig().RouteConfig.RequestLatencyRoutes, calledRoute)
	if isRouteFound {
		handlerMeta = updateLatency(handlerMeta, emitter, calledRoute)
	}

	return handlerMeta
}
