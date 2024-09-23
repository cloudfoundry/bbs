package metrics

import (
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

type RequestStatMetronNotifier struct {
	logger            lager.Logger
	ticker            clock.Ticker
	requestCount      uint64
	maxRequestLatency time.Duration
	lock              sync.Mutex
	metronClient      loggingclient.IngressClient
}

func NewRequestStatMetronNotifier(logger lager.Logger, ticker clock.Ticker, metronClient loggingclient.IngressClient) *RequestStatMetronNotifier {
	return &RequestStatMetronNotifier{
		logger:       logger,
		ticker:       ticker,
		metronClient: metronClient,
	}
}

func (notifier *RequestStatMetronNotifier) IncrementRequestCounter(delta int) {
	atomic.AddUint64(&notifier.requestCount, uint64(delta))
}

func (notifier *RequestStatMetronNotifier) UpdateLatency(latency time.Duration) {
	notifier.lock.Lock()
	defer notifier.lock.Unlock()
	if latency > notifier.maxRequestLatency {
		notifier.maxRequestLatency = latency
	}
}

func (notifier *RequestStatMetronNotifier) ReadAndResetLatency() time.Duration {
	notifier.lock.Lock()
	defer notifier.lock.Unlock()

	currentLatency := notifier.maxRequestLatency
	notifier.maxRequestLatency = 0

	return currentLatency
}

func (notifier *RequestStatMetronNotifier) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	logger := notifier.logger.Session("periodic-count-metrics-notifier")
	close(ready)

	logger.Info("started")
	defer logger.Info("finished")

	for {
		select {
		case <-notifier.ticker.C():
			add := atomic.SwapUint64(&notifier.requestCount, 0)
			logger.Info("adding-counter", lager.Data{"add": add})
			metricErr := notifier.metronClient.IncrementCounterWithDelta(requestCounter, add)
			if metricErr != nil {
				logger.Debug("failed-to-emit-request-counter", lager.Data{"error": metricErr})
			}

			latency := notifier.ReadAndResetLatency()
			if latency != 0 {
				logger.Info("sending-latency", lager.Data{"latency": latency})
				metricErr := notifier.metronClient.SendDuration(requestLatencyDuration, latency)
				if metricErr != nil {
					logger.Debug("failed-to-emit-request-latency-metric", lager.Data{"error": metricErr})
				}
			}
		case <-signals:
			return nil
		}
	}
}
