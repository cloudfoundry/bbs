package handlers

import (
	"os"

	"github.com/tedsuo/ifrit"
)

type lockReadyNotifier struct {
	lockReady chan<- struct{}
}

func NewLockReadyNotifier(lockReady chan<- struct{}) ifrit.Runner {
	return &lockReadyNotifier{
		lockReady: lockReady,
	}
}

func (n *lockReadyNotifier) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	close(n.lockReady)
	close(ready)
	<-signals
	return nil
}
