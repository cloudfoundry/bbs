package metrics

import (
	"os"

	"github.com/tedsuo/ifrit"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/runtimeschema/metric"
)

const (
	bbsMasterElected = metric.Counter("BBSMasterElected")
)

type BBSElectionMetronNotifier struct {
	Logger lager.Logger
}

func NewBBSElectionMetronNotifier(logger lager.Logger) ifrit.Runner {
	return &BBSElectionMetronNotifier{
		Logger: logger,
	}
}

func (notifier BBSElectionMetronNotifier) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	logger := notifier.Logger.Session("metrics-notifier")
	logger.Info("starting")

	close(ready)

	logger.Info("started")
	defer logger.Info("finished")

	bbsMasterElected.Increment()

	<-signals
	return nil
}
