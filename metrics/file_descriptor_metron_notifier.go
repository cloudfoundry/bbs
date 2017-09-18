package metrics

import (
	"io/ioutil"
	"os"

	"github.com/tedsuo/ifrit"

	"code.cloudfoundry.org/clock"
	loggingclient "code.cloudfoundry.org/diego-logging-client"
	"code.cloudfoundry.org/lager"
)

const (
	OpenFileDescriptors = "OpenFileDescriptors"
	FileDescriptorUnits = "descriptors"
)

type FileDescriptorMetronNotifier struct {
	Logger       lager.Logger
	metronClient loggingclient.IngressClient
	ticker       clock.Ticker
	procFSPath   string
}

func NewFileDescriptorMetronNotifier(logger lager.Logger, metronClient loggingclient.IngressClient, newTicker clock.Ticker, procPath string) ifrit.Runner {
	return &FileDescriptorMetronNotifier{
		Logger:       logger,
		metronClient: metronClient,
		ticker:       newTicker,
		procFSPath:   procPath,
	}
}

func (notifier FileDescriptorMetronNotifier) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	logger := notifier.Logger.Session("metrics-notifier")
	logger.Info("starting")

	close(ready)

	logger.Info("started")
	defer logger.Info("finished")

	for {
		select {
		case <-notifier.ticker.C():
			nDescriptors, err := notifier.descriptorCount()

			if err == nil {
				err := notifier.metronClient.SendMetric(OpenFileDescriptors, nDescriptors)

				if err != nil {
					logger.Error("error-sending-metric", err)
				}

			} else {
				logger.Error("failed-to-read-proc-filesystem", err)
			}

		case <-signals:
			return nil
		}
	}
}

func (notifier FileDescriptorMetronNotifier) descriptorCount() (int, error) {
	descriptorInfos, err := ioutil.ReadDir(notifier.procFSPath)

	if err != nil {
		return 0, err
	}

	count := 0

	for _, descriptorInfo := range descriptorInfos {
		notifier.Logger.Info("file-info", lager.Data{"name": descriptorInfo.Name(), "mode": descriptorInfo.Mode()})
		if descriptorInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
			count++
		}
	}

	return count, nil
}
