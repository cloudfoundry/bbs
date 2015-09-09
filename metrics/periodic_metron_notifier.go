package metrics

import (
	"os"
	"time"

	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/runtime-schema/metric"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
)

const metricsReportingDuration = metric.Duration("MetricsReportingDuration")

type PeriodicMetronNotifier struct {
	Interval    time.Duration
	ETCDOptions *etcd.ETCDOptions
	Logger      lager.Logger
	Clock       clock.Clock
}

func NewPeriodicMetronNotifier(logger lager.Logger,
	interval time.Duration,
	etcdOptions *etcd.ETCDOptions,
	clock clock.Clock,
) *PeriodicMetronNotifier {
	return &PeriodicMetronNotifier{
		Interval:    interval,
		ETCDOptions: etcdOptions,
		Logger:      logger,
		Clock:       clock,
	}
}

func (notifier PeriodicMetronNotifier) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	logger := notifier.Logger.Session("metrics-notifier", lager.Data{"interval": notifier.Interval.String()})
	logger.Info("starting")

	etcdMetrics, err := etcd.NewETCDMetrics(notifier.Logger, notifier.ETCDOptions)
	if err != nil {
		return err
	}

	ticker := notifier.Clock.NewTicker(notifier.Interval)
	defer ticker.Stop()

	close(ready)

	logger.Info("started")
	defer logger.Info("finished")

	for {
		select {
		case <-ticker.C():
			logger.Info("emitting")
			startedAt := notifier.Clock.Now()

			etcdMetrics.Send()

			finishedAt := notifier.Clock.Now()

			metricsReportingDuration.Send(finishedAt.Sub(startedAt))

		case <-signals:
			return nil
		}
	}

	return nil
}
