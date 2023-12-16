package handlers

import (
	"net/http"
	"sync"

	"code.cloudfoundry.org/lager/v3"
)

type UnavailableHandler struct {
	handler http.Handler
	logger  lager.Logger
	waitCh  <-chan struct{}
}

func NewUnavailableHandler(handler http.Handler, logger lager.Logger, serviceReadyChan ...<-chan struct{}) *UnavailableHandler {
	wg := sync.WaitGroup{}
	logger.Info("setting-up")
	for _, ch := range serviceReadyChan {
		wg.Add(1)
		go func(ch <-chan struct{}) {
			defer wg.Done()
			<-ch
			logger.Info("closed-hello")
		}(ch)
	}

	waitCh := make(chan struct{})
	go func() {
		wg.Wait()
		logger.Info("wait-closed")
		close(waitCh)
	}()

	u := &UnavailableHandler{
		handler: handler,
		logger:  logger,
		waitCh:  waitCh,
	}

	return u
}

func (u *UnavailableHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	u.logger.Info("in-serve")
	select {
	case <-u.waitCh:
		u.logger.Info("in-serve-1")
		u.handler.ServeHTTP(w, r)
	default:
		u.logger.Info("in-serve-2")
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}

func UnavailableWrap(handler http.Handler, logger lager.Logger, serviceReady ...<-chan struct{}) http.HandlerFunc {
	handler = NewUnavailableHandler(handler, logger, serviceReady...)

	return func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
	}
}
