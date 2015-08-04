package auctionhandlers

import (
	"net/http"

	"github.com/cloudfoundry-incubator/auction/auctiontypes"
	"github.com/cloudfoundry/dropsonde"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/rata"
)

type loggableHandler interface {
	WithLogger(lager.Logger) http.Handler
}

type loggableHandlerFunc func(http.ResponseWriter, *http.Request, lager.Logger)

func (f loggableHandlerFunc) WithLogger(logger lager.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f(w, r, logger)
	})
}

func New(runner auctiontypes.AuctionRunner, logger lager.Logger) http.Handler {
	taskAuctionHandler := logWrap(route(NewTaskAuctionHandler(runner).Create), logger)
	lrpAuctionHandler := logWrap(route(NewLRPAuctionHandler(runner).Create), logger)

	actions := rata.Handlers{
		CreateTaskAuctionsRoute: taskAuctionHandler,
		CreateLRPAuctionsRoute:  lrpAuctionHandler,
	}

	handler, err := rata.NewRouter(Routes, actions)
	if err != nil {
		panic("unable to create router: " + err.Error())
	}

	return handler
}

func route(f func(_ http.ResponseWriter, _ *http.Request, _ lager.Logger)) loggableHandler {
	return loggableHandlerFunc(f)
}

func logWrap(loggable loggableHandler, logger lager.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestLog := logger.Session("request", lager.Data{
			"method":  r.Method,
			"request": r.URL.String(),
		})

		handler := dropsonde.InstrumentedHandler(loggable.WithLogger(requestLog))

		requestLog.Info("serving")
		handler.ServeHTTP(w, r)
		requestLog.Info("done")
	}
}
