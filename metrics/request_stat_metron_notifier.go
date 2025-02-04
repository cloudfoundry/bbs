package metrics

import (
	"code.cloudfoundry.org/bbs/cmd/bbs/config"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"code.cloudfoundry.org/clock"
	loggingclient "code.cloudfoundry.org/diego-logging-client"
	"code.cloudfoundry.org/lager/v3"
)

const (
	requestCounter         = "RequestCount"
	requestLatencyDuration = "RequestLatency"
)

type requestMetrics struct {
	requestCount      uint64
	maxRequestLatency time.Duration
}

type RequestStatMetronNotifier struct {
	logger                 lager.Logger
	ticker                 clock.Ticker
	requestMetricsAll      requestMetrics
	requestMetricsPerRoute map[string]*requestMetrics
	lock                   sync.Mutex
	metronClient           loggingclient.IngressClient
	advancedMetricsConfig  config.AdvancedMetrics
}

func NewRequestStatMetronNotifier(
	logger lager.Logger,
	ticker clock.Ticker,
	metronClient loggingclient.IngressClient,
	advancedMetricsConfig config.AdvancedMetrics) *RequestStatMetronNotifier {

	requestMetricsPerRoute := make(map[string]*requestMetrics)
	initRouteMaps := func(routes []string) {
		for _, route := range routes {
			requestMetricsPerRoute[route] = &requestMetrics{}
		}
	}

	initRouteMaps(advancedMetricsConfig.RouteConfig.RequestCountRoutes)
	initRouteMaps(advancedMetricsConfig.RouteConfig.RequestLatencyRoutes)

	return &RequestStatMetronNotifier{
		logger:                 logger,
		ticker:                 ticker,
		metronClient:           metronClient,
		requestMetricsPerRoute: requestMetricsPerRoute,
		advancedMetricsConfig:  advancedMetricsConfig,
	}
}

func (notifier *RequestStatMetronNotifier) GetAdvancedMetricsConfig() config.AdvancedMetrics {
	return notifier.advancedMetricsConfig
}

func (notifier *RequestStatMetronNotifier) IncrementRequestCounter(delta int, route string) {
	if route != "" {
		atomic.AddUint64(&notifier.requestMetricsPerRoute[route].requestCount, uint64(delta))

		return
	}

	atomic.AddUint64(&notifier.requestMetricsAll.requestCount, uint64(delta))
}

func (notifier *RequestStatMetronNotifier) UpdateLatency(latency time.Duration, route string) {
	updateLatency := func(metrics *requestMetrics) {
		notifier.lock.Lock()
		defer notifier.lock.Unlock()

		if latency > metrics.maxRequestLatency {
			metrics.maxRequestLatency = latency
		}
	}

	if route != "" {
		updateLatency(notifier.requestMetricsPerRoute[route])

		return
	}

	updateLatency(&notifier.requestMetricsAll)
}

func readAndResetMetric[MetricType uint64 | time.Duration](metric *MetricType, lock *sync.Mutex) MetricType {
	lock.Lock()
	defer lock.Unlock()

	currentMetricValue := *metric
	*metric = 0

	return currentMetricValue
}

func (notifier *RequestStatMetronNotifier) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	logger := notifier.logger.Session("periodic-count-metrics-notifier")
	close(ready)

	logger.Info("started")
	defer logger.Info("finished")

	for {
		select {
		case <-notifier.ticker.C():
			// Emit Default Metrics
			requestCountMetricValue := readAndResetMetric(&notifier.requestMetricsAll.requestCount, &notifier.lock)
			notifier.emitRequestCount(requestCounter, requestCountMetricValue, logger)

			requestLatencyMetricValue := readAndResetMetric(&notifier.requestMetricsAll.maxRequestLatency, &notifier.lock)
			notifier.emitRequestLatency(requestLatencyDuration, requestLatencyMetricValue, logger)

			// Emit Route Specific/Advanced Metrics
			if !notifier.advancedMetricsConfig.Enabled {
				break
			}

			for _, route := range notifier.advancedMetricsConfig.RouteConfig.RequestCountRoutes {
				requestCountMetricValue := readAndResetMetric(&notifier.requestMetricsPerRoute[route].requestCount, &notifier.lock)
				notifier.emitRequestCount(requestCounter + "." + route, requestCountMetricValue, logger)
			}

			for _, route := range notifier.advancedMetricsConfig.RouteConfig.RequestLatencyRoutes {
				requestLatencyMetricValue := readAndResetMetric(&notifier.requestMetricsPerRoute[route].maxRequestLatency, &notifier.lock)
				notifier.emitRequestLatency(requestLatencyDuration + "." + route, requestLatencyMetricValue, logger)
			}
		case <-signals:
			return nil
		}
	}
}

func (notifier *RequestStatMetronNotifier) emitRequestLatency(
	requestLatencyMetricName string,
	requestLatencyMetricValue time.Duration,
	logger lager.Logger) {

	//TODO: Check what happens if the latency is 0 and emitted/not emitted
	logger.Info("sending-latency", lager.Data{"latency": requestLatencyMetricValue})
	metricErr := notifier.metronClient.SendDuration(requestLatencyMetricName, requestLatencyMetricValue)
	if metricErr != nil {
		logger.Debug("failed-to-emit-request-latency-metric", lager.Data{"error": metricErr})
	}
}

func (notifier *RequestStatMetronNotifier) emitRequestCount(
	requestCountMetricName string,
	requestCountMetricValue uint64,
	logger lager.Logger) {

	logger.Info("adding-counter", lager.Data{"add": requestCountMetricValue})
	metricErr := notifier.metronClient.IncrementCounterWithDelta(requestCountMetricName, requestCountMetricValue)
	if metricErr != nil {
		logger.Debug("failed-to-emit-request-counter", lager.Data{"error": metricErr})
	}
}
