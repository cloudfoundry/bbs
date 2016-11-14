package middleware

import (
	"errors"
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
	latencyChannel := make(chan time.Duration, 10000)
	max_latency := time.Duration(0)

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		for {
			select {
			case <-ticker.C:
				for {
					select {
					case latency := <-latencyChannel:
						if latency > max_latency {
							max_latency = latency
						}
					default:
						break
					}
				}

				err := requestLatency.Send(max_latency)
				if err != nil {
					logger.Error("failed-to-send-request-latency-metric", err)
				}
			}
		}
	}()

	return LatencyEmitter{
		logger:         logger,
		latencyChannel: latencyChannel,
	}
}

type LatencyEmitter struct {
	logger         lager.Logger
	latencyChannel chan time.Duration
}

func (l LatencyEmitter) EmitLatency(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		f(w, r)
		select {
		case l.latencyChannel <- time.Since(startTime):
		default:
			l.logger.Error("dropped-latency-metric", errors.New("Channel too full"))
		}
	}
}

func RequestCountWrap(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestCount.Increment()
		handler.ServeHTTP(w, r)
	}
}
