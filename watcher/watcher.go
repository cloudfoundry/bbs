package watcher

import (
	"os"
	"time"

	"github.com/cloudfoundry-incubator/bbs/events"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
)

type Watcher ifrit.Runner

type watcher struct {
	logger            lager.Logger
	name              string
	retryWaitDuration time.Duration
	streamer          EventStreamer
	hub               events.Hub
	clock             clock.Clock
}

func NewWatcher(
	logger lager.Logger,
	name string,
	retryWaitDuration time.Duration,
	streamer EventStreamer,
	hub events.Hub,
	clock clock.Clock,
) Watcher {
	return &watcher{
		logger:            logger,
		name:              name,
		retryWaitDuration: retryWaitDuration,
		streamer:          streamer,
		hub:               hub,
		clock:             clock,
	}
}

func (w *watcher) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	logger := w.logger.Session("watcher", lager.Data{"watch-name": w.name})
	logger.Info("starting")

	hubNotification := make(chan int, 1)
	hubSize := 0

	logger.Info("registering-callback-with-hub")
	w.hub.RegisterCallback(func(size int) {
		hubNotification <- size
	})

	close(ready)
	logger.Info("started")
	defer logger.Info("finished")

	var stopChan chan<- bool
	var errorChan <-chan error
	eventChan := make(chan models.Event)

	reWatchTimer := w.clock.NewTimer(w.retryWaitDuration)
	defer reWatchTimer.Stop()
	reWatchTimer.Stop()

	var reWatchChan <-chan time.Time

	for {
		select {
		case hubSize = <-hubNotification:
			if hubSize == 0 {
				if stopChan != nil {
					logger.Info("stopping-watch-from-hub-notification")
					stopChan <- true
					stopChan = nil
					errorChan = nil
				}
			} else {
				if stopChan == nil {
					logger.Info("rewatching-from-hub-notification")
					stopChan, errorChan = w.streamer.Stream(logger, eventChan)
					logger.Debug("finished-rewatching-rom-hub-notification")
				}
			}

		case event := <-eventChan:
			w.hub.Emit(event)

		case err, ok := <-errorChan:
			if ok {
				reWatchTimer.Reset(w.retryWaitDuration)
				reWatchChan = reWatchTimer.C()
			}
			if err != nil {
				logger.Error("watch-failed", err)
			}
			errorChan = nil
			stopChan = nil

		case <-reWatchChan:
			reWatchChan = nil

			if stopChan == nil && hubSize > 0 {
				logger.Info("rewatching")
				stopChan, errorChan = w.streamer.Stream(logger, eventChan)
			}

		case <-signals:
			logger.Info("stopping")
			if stopChan != nil {
				stopChan <- true
				stopChan = nil
			}
			return nil
		}
	}

	return nil
}
