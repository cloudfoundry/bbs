package metrics

import (
	"os"
	"sync"
	"sync/atomic"
	"time"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/runtimeschema/metric"
)

const (
	requestCounter = metric.Counter("RequestCount")
	requestLatency = metric.Duration("RequestLatency")
)

type RequestStatMetronNotifier struct {
	logger            lager.Logger
	ticker            clock.Ticker
	requestCount      uint64
	maxRequestLatency time.Duration
	lock              sync.Mutex
}

func NewRequestStatMetronNotifier(logger lager.Logger, ticker clock.Ticker) *RequestStatMetronNotifier {
	return &RequestStatMetronNotifier{
		logger: logger,
		ticker: ticker,
	}
}

func (notifier *RequestStatMetronNotifier) IncrementCounter(delta int) {
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

	for {
		select {
		case <-notifier.ticker.C():
			add := atomic.SwapUint64(&notifier.requestCount, 0)
			logger.Info("adding-counter", lager.Data{"add": add})
			requestCounter.Add(add)

			latency := notifier.ReadAndResetLatency()
			if latency != 0 {
				logger.Info("sending-latency", lager.Data{"latency": latency})
				requestLatency.Send(latency)
			}
		case <-signals:
			return nil
		}
	}

	defer logger.Info("finished")

	<-signals
	return nil
}
