package metrics

import (
	"os"
	"time"

	"code.cloudfoundry.org/bbs/db/etcd"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/runtimeschema/metric"
)

const (
	metricsReportingDuration = metric.Duration("MetricsReportingDuration")

	bbsMasterElected = metric.Counter("BBSMasterElected")
)

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
	var etcdMetrics *etcd.ETCDMetrics
	var err error

	if notifier.ETCDOptions.IsConfigured {
		etcdMetrics, err = etcd.NewETCDMetrics(notifier.Logger, notifier.ETCDOptions)
		if err != nil {
			return err
		}
	}

	ticker := notifier.Clock.NewTicker(notifier.Interval)
	defer ticker.Stop()

	close(ready)

	logger.Info("started")
	defer logger.Info("finished")

	bbsMasterElected.Increment()

	for {
		select {
		case <-ticker.C():
			startedAt := notifier.Clock.Now()

			if etcdMetrics != nil {
				etcdMetrics.Send()
			}

			finishedAt := notifier.Clock.Now()

			err = metricsReportingDuration.Send(finishedAt.Sub(startedAt))
			if err != nil {
				logger.Error("failed-to-send-metrics-reporting-duration-metric", err)
			}

		case <-signals:
			return nil
		}
	}

	return nil
}
