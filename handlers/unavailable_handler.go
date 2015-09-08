package handlers

import "net/http"

type UnavailableHandler struct {
	handler          http.Handler
	serviceReadyChan <-chan struct{}
	serviceReady     bool
}

func NewUnavailableHandler(handler http.Handler, serviceReadyChan <-chan struct{}) *UnavailableHandler {
	u := &UnavailableHandler{
		handler:          handler,
		serviceReadyChan: serviceReadyChan,
		serviceReady:     false,
	}

	go u.waitForMigrations()

	return u
}

func (u *UnavailableHandler) waitForMigrations() {
	select {
	case <-u.serviceReadyChan:
		u.serviceReady = true
	}
}

func (u *UnavailableHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if u.serviceReady {
		u.handler.ServeHTTP(w, r)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}
