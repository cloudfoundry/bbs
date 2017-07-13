package metrics

import (
	"os"

	"github.com/tedsuo/ifrit"

	loggregator_v2 "code.cloudfoundry.org/go-loggregator/compatibility"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/runtimeschema/metric"
)

const (
	bbsMasterElected = metric.Counter("BBSMasterElected")
)

type BBSElectionMetronNotifier struct {
	Logger       lager.Logger
	metronClient loggregator_v2.IngressClient
}

func NewBBSElectionMetronNotifier(logger lager.Logger, metronClient loggregator_v2.IngressClient) ifrit.Runner {
	return &BBSElectionMetronNotifier{
		Logger:       logger,
		metronClient: metronClient,
	}
}

func (notifier BBSElectionMetronNotifier) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	logger := notifier.Logger.Session("metrics-notifier")
	logger.Info("starting")

	close(ready)

	logger.Info("started")
	defer logger.Info("finished")

	notifier.metronClient.SendMetric("BBSMasterElected", 1)

	<-signals
	return nil
}
