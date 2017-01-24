package lock

import (
	"os"
	"time"

	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"
)

type lockRunner struct {
	logger lager.Logger

	locker        Locker
	lock          models.Lock
	clock         clock.Clock
	retryInterval time.Duration
}

//go:generate counterfeiter . Locker
type Locker interface {
	Lock(logger lager.Logger, lock models.Lock) error
	Release(logger lager.Logger, lock models.Lock) error
}

func NewLockRunner(
	logger lager.Logger,
	locker Locker,
	lock models.Lock,
	clock clock.Clock,
	retryInterval time.Duration,
) *lockRunner {
	return &lockRunner{
		logger:        logger,
		locker:        locker,
		lock:          lock,
		clock:         clock,
		retryInterval: retryInterval,
	}
}

func (l *lockRunner) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	logger := l.logger.Session("lock")

	logger.Info("started")
	defer logger.Info("completed")

	close(ready)

	retry := l.clock.NewTimer(l.retryInterval)

	err := l.locker.Lock(logger, l.lock)
	if err != nil {
		logger.Error("failed-to-acquire-lock", err)
		retry.Reset(l.retryInterval)
	} else {
		logger.Info("acquired-lock")
		retry.Stop()
	}

	for {
		select {
		case sig := <-signals:
			logger.Info("signalled", lager.Data{"signal": sig})

			err := l.locker.Release(logger, l.lock)
			if err != nil {
				logger.Error("failed-to-release-lock", err)
			} else {
				logger.Info("released-lock")
			}

			return nil

		case <-retry.C():
			err = l.locker.Lock(logger, l.lock)
			if err != nil {
				logger.Error("failed-to-acquire-lock", err)
				retry.Reset(l.retryInterval)
			} else {
				logger.Info("acquired-lock")
				retry.Stop()
			}
		}
	}

	return nil
}
