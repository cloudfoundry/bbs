package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"code.cloudfoundry.org/bbs/db"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager/v3"
)

type DBHealthCheckRunner struct {
	logger                      lager.Logger
	sqlDB                       db.BBSHealthCheckDB
	clock                       clock.Clock
	migrationsDone              chan struct{}
	HealthCheckFailureThreshold int
	HealthCheckTimeout          time.Duration
	HealthCheckInterval         time.Duration
}

func NewDBHealthCheckRunner(logger lager.Logger, sqlDB db.BBSHealthCheckDB, clock clock.Clock, failureCount int, timeout, interval time.Duration, migrationsDone chan struct{}) *DBHealthCheckRunner {
	if failureCount == 0 {
		failureCount = 3
	}
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	if interval == 0 {
		interval = 10 * time.Second
	}
	return &DBHealthCheckRunner{
		logger:                      logger.Session("db-health-check-runner"),
		sqlDB:                       sqlDB,
		clock:                       clock,
		HealthCheckFailureThreshold: failureCount,
		HealthCheckTimeout:          timeout,
		HealthCheckInterval:         interval,
		migrationsDone:              migrationsDone,
	}
}

func (runner *DBHealthCheckRunner) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	close(ready)
	runner.logger.Info("waiting-for-db-migrations")
	select {
	case <-runner.migrationsDone:
	case <-signals:
		runner.logger.Debug("exiting-due-to-signal")
		return nil
	}
	runner.logger.Info("starting")
	defer runner.logger.Info("exiting")
	ticker := runner.clock.NewTicker(runner.HealthCheckInterval)
	for {
		runner.logger.Debug("reentering-run-loop")
		select {
		case <-ticker.C():
			runner.logger.Debug("executing-health-check")
			err := runner.ExecuteTimedHealthCheckWithRetries()
			if err != nil {
				runner.logger.Error("catastrophic-database-failure-detected", err)
				return err
			}
			runner.logger.Debug("health-check-succeeded")

		case <-signals:
			runner.logger.Debug("exiting-due-to-signal")
			return nil
		}
	}
}

func (runner *DBHealthCheckRunner) ExecuteTimedHealthCheckWithRetries() error {
	var errs []error
	var i int
	for i = range runner.HealthCheckFailureThreshold {
		logger := runner.logger.WithData(lager.Data{"attempt": i})
		err := runner.ExecuteTimedHealthCheck()
		if err != nil {
			logger.Error("failed-healthcheck", err)
			errs = append(errs, err)
		} else {
			logger.Debug("succeeded-healthcheck")
			return nil
		}
	}
	finalErr := errors.Join(errs...)
	runner.logger.Error("failed-healthcheck-attempts-exceeded", finalErr, lager.Data{"attempts": i, "max-attempts": runner.HealthCheckFailureThreshold})
	return finalErr
}

func (runner *DBHealthCheckRunner) ExecuteTimedHealthCheck() error {
	timer := runner.clock.NewTimer(runner.HealthCheckTimeout)
	errChan := make(chan error)
	go runner.runDBHealthCheck(errChan)

	select {
	case err := <-errChan:
		if err == nil {
			runner.logger.Debug("health-check-succeeded")
			return nil
		} else {
			return err
		}
	case <-timer.C():
		err := fmt.Errorf("timed out after %s while executing DB health check", runner.HealthCheckTimeout)
		runner.logger.Error("health-check-timed-out", err)
		return err
	}
}

func (runner *DBHealthCheckRunner) runDBHealthCheck(errChan chan error) {
	err := runner.sqlDB.PerformBBSHealthCheck(context.Background(), runner.logger, time.Now())
	errChan <- err
	runner.logger.Debug("runDBHealthCheck-complete")
}
