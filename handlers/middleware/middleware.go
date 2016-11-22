package middleware

import (
	"net/http"
	"sync"
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

func NewLatencyEmitter(logger lager.Logger) *LatencyEmitter {
	l := &LatencyEmitter{
		logger:            logger,
		metricLock:        &sync.Mutex{},
		currentMaxLatency: 0 * time.Second,
	}

	go l.emitMetrics()

	return l
}

type LatencyEmitter struct {
	logger            lager.Logger
	metricLock        *sync.Mutex
	currentMaxLatency time.Duration
}

func (l *LatencyEmitter) EmitLatency(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		f(w, r)
		endTime := time.Since(startTime)

		go l.addMetric(endTime)
	}
}

func (l *LatencyEmitter) addMetric(duration time.Duration) {
	l.metricLock.Lock()
	defer l.metricLock.Unlock()

	if duration > l.currentMaxLatency {
		l.currentMaxLatency = duration
	}
}

func (l *LatencyEmitter) emitMetrics() {
	for {
		l.metricLock.Lock()
		err := requestLatency.Send(l.currentMaxLatency)
		if err != nil {
			l.logger.Error("failed-to-send-request-latency-metric", err)
		}
		l.currentMaxLatency = 0
		l.metricLock.Unlock()

		time.Sleep(30 * time.Second)
	}
}

func RequestCountWrap(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestCount.Increment()
		handler.ServeHTTP(w, r)
	}
}
