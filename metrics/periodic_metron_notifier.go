package metrics

import (
	"os"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/runtimeschema/metric"
)

const (
	bbsMasterElected = metric.Counter("BBSMasterElected")
)

type PeriodicMetronNotifier struct {
	Logger lager.Logger
}

func NewPeriodicMetronNotifier(logger lager.Logger) *PeriodicMetronNotifier {
	return &PeriodicMetronNotifier{
		Logger: logger,
	}
}

func (notifier PeriodicMetronNotifier) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	logger := notifier.Logger.Session("metrics-notifier")
	logger.Info("starting")

	close(ready)

	logger.Info("started")
	defer logger.Info("finished")

	bbsMasterElected.Increment()

	<-signals
	return nil
}
