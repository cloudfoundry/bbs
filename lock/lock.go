package lock

import (
	"os"
	"time"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"
)

type lockRunner struct {
	logger lager.Logger

	lock          Lock
	key           string
	clock         clock.Clock
	retryInterval time.Duration
}

//go:generate counterfeiter . Lock
type Lock interface {
	Lock(logger lager.Logger, key string) error
	Release(logger lager.Logger, key string) error
}

func NewLockRunner(
	logger lager.Logger,
	lock Lock,
	key string,
	clock clock.Clock,
	retryInterval time.Duration,
) *lockRunner {
	return &lockRunner{
		logger:        logger,
		lock:          lock,
		key:           key,
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

	err := l.lock.Lock(logger, l.key)
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

			err := l.lock.Release(logger, l.key)
			if err != nil {
				logger.Error("failed-to-release-lock", err)
			} else {
				logger.Info("released-lock")
			}

			return nil

		case <-retry.C():
			err = l.lock.Lock(logger, l.key)
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
