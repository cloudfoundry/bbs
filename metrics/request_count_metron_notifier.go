package metrics

import (
	"os"
	"sync/atomic"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/runtimeschema/metric"
)

const (
	requestCounter = metric.Counter("RequestCount")
)

type RequestCountMetronNotifier struct {
	logger       lager.Logger
	ticker       clock.Ticker
	requestCount uint64
}

func NewRequestCountMetronNotifier(logger lager.Logger, ticker clock.Ticker) *RequestCountMetronNotifier {
	return &RequestCountMetronNotifier{
		logger: logger,
		ticker: ticker,
	}
}

func (notifier *RequestCountMetronNotifier) IncrementCounter(delta int) {
	atomic.AddUint64(&notifier.requestCount, uint64(delta))
}

func (notifier *RequestCountMetronNotifier) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	logger := notifier.logger.Session("periodic-metrics-notifier")
	close(ready)

	logger.Info("started")

	for {
		select {
		case <-notifier.ticker.C():
			add := atomic.SwapUint64(&notifier.requestCount, 0)
			logger.Info("adding-counter", lager.Data{"add": add})
			requestCounter.Add(add)
		case <-signals:
			return nil
		}
	}

	defer logger.Info("finished")

	<-signals
	return nil
}
