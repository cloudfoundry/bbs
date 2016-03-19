package handlers

import (
	"net/http"

	"github.com/cloudfoundry-incubator/auctioneer"
	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry/gunk/workpool"
	"github.com/pivotal-golang/lager"
)

type LRPConvergenceHandler struct {
	logger                 lager.Logger
	db                     db.LRPDB
	auctioneerClient       auctioneer.Client
	retirer                ActualLRPRetirer
	convergenceWorkersSize int
}

func NewLRPConvergenceHandler(logger lager.Logger, db db.LRPDB, auctioneerClient auctioneer.Client, retirer ActualLRPRetirer, convergenceWorkersSize int) *LRPConvergenceHandler {
	return &LRPConvergenceHandler{logger, db, auctioneerClient, retirer, convergenceWorkersSize}
}

func (h *LRPConvergenceHandler) ConvergeLRPs(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("converge-lrps")
	startRequests, keysToRetire := h.db.ConvergeLRPs(logger)

	startLogger := logger.WithData(lager.Data{"start-requests-count": len(startRequests)})
	startLogger.Debug("requesting-start-auctions")
	err := h.auctioneerClient.RequestLRPAuctions(startRequests)
	if err != nil {
		startLogger.Error("failed-to-request-starts", err, lager.Data{"lrp-start-auctions": startRequests})
	}
	startLogger.Debug("done-requesting-start-auctions")

	retireLogger := logger.WithData(lager.Data{"retiring-lrp-count": len(keysToRetire)})
	works := []func(){}
	for _, key := range keysToRetire {
		key := key
		works = append(works, func() { h.retirer.RetireActualLRP(retireLogger, key.ProcessGuid, key.Index) })
	}

	throttler, err := workpool.NewThrottler(h.convergenceWorkersSize, works)
	if err != nil {
		logger.Error("failed-constructing-throttler", err, lager.Data{"max-workers": h.convergenceWorkersSize, "num-works": len(works)})
		return
	}

	startLogger.Debug("retiring-actual-lrps")
	throttler.Work()
	retireLogger.Debug("done-retiring-actual-lrps")

	response := &models.ConvergeLRPsResponse{}
	writeResponse(w, response)
}
