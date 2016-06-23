package handlers

import (
	"net/http"
	"sync"

	"github.com/cloudfoundry-incubator/auctioneer"
	"github.com/cloudfoundry-incubator/bbs"
	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/events"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry/gunk/workpool"
	"github.com/pivotal-golang/lager"
)

type LRPConvergenceHandler struct {
	logger                 lager.Logger
	db                     db.LRPDB
	actualHub              events.Hub
	auctioneerClient       auctioneer.Client
	serviceClient          bbs.ServiceClient
	retirer                ActualLRPRetirer
	convergenceWorkersSize int
	exitChan               chan<- struct{}
}

func NewLRPConvergenceHandler(
	logger lager.Logger,
	db db.LRPDB,
	actualHub events.Hub,
	auctioneerClient auctioneer.Client,
	serviceClient bbs.ServiceClient,
	retirer ActualLRPRetirer,
	convergenceWorkersSize int,
	exitChan chan<- struct{},
) *LRPConvergenceHandler {
	return &LRPConvergenceHandler{
		logger:                 logger,
		db:                     db,
		actualHub:              actualHub,
		auctioneerClient:       auctioneerClient,
		serviceClient:          serviceClient,
		retirer:                retirer,
		convergenceWorkersSize: convergenceWorkersSize,
		exitChan:               exitChan,
	}
}

func (h *LRPConvergenceHandler) ConvergeLRPs(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("converge-lrps")

	logger.Debug("listing-cells")
	cellSet, err := h.serviceClient.Cells(logger)
	if err == models.ErrResourceNotFound {
		logger.Debug("no-cells-found")
		cellSet = models.CellSet{}
	} else if err != nil {
		logger.Debug("failed-listing-cells")
		return
	}
	logger.Debug("succeeded-listing-cells")

	startRequests, keysWithMissingCells, keysToRetire := h.db.ConvergeLRPs(logger, cellSet)

	retireLogger := logger.WithData(lager.Data{"retiring-lrp-count": len(keysToRetire)})
	works := []func(){}
	for _, key := range keysToRetire {
		key := key
		works = append(works, func() { h.retirer.RetireActualLRP(retireLogger, key.ProcessGuid, key.Index) })
	}

	startRequestLock := &sync.Mutex{}
	for _, key := range keysWithMissingCells {
		key := key
		works = append(works, func() {
			before, after, err := h.db.UnclaimActualLRP(logger, key.Key)
			if err == nil {
				h.actualHub.Emit(models.NewActualLRPChangedEvent(before, after))
				startRequest := auctioneer.NewLRPStartRequestFromSchedulingInfo(key.SchedulingInfo, int(key.Key.Index))
				startRequestLock.Lock()
				startRequests = append(startRequests, &startRequest)
				startRequestLock.Unlock()
			}
		})
	}

	throttler, err := workpool.NewThrottler(h.convergenceWorkersSize, works)
	if err != nil {
		logger.Error("failed-constructing-throttler", err, lager.Data{"max-workers": h.convergenceWorkersSize, "num-works": len(works)})
		return
	}

	retireLogger.Debug("retiring-actual-lrps")
	throttler.Work()
	retireLogger.Debug("done-retiring-actual-lrps")

	startLogger := logger.WithData(lager.Data{"start-requests-count": len(startRequests)})
	startLogger.Debug("requesting-start-auctions")
	err = h.auctioneerClient.RequestLRPAuctions(startRequests)
	if err != nil {
		startLogger.Error("failed-to-request-starts", err, lager.Data{"lrp-start-auctions": startRequests})
	}
	startLogger.Debug("done-requesting-start-auctions")

	response := &models.ConvergeLRPsResponse{}
	response.Error = models.ConvertError(err)

	writeResponse(w, response)
	exitIfUnrecoverable(logger, h.exitChan, response.Error)
}
