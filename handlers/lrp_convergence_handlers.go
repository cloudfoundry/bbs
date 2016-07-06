package handlers

import (
	"net/http"
	"sync"

	"code.cloudfoundry.org/auctioneer"
	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/bbs/db"
	"code.cloudfoundry.org/bbs/events"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/gunk/workpool"
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

func (h *LRPConvergenceHandler) DeprecatedConvergeLRPs(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (h *LRPConvergenceHandler) ConvergeLRPs() error {
	logger := h.logger.Session("converge-lrps")
	var err error

	defer func() { exitIfUnrecoverable(logger, h.exitChan, models.ConvertError(err)) }()

	logger.Debug("listing-cells")
	var cellSet models.CellSet
	cellSet, err = h.serviceClient.Cells(logger)
	if err == models.ErrResourceNotFound {
		logger.Debug("no-cells-found")
		cellSet = models.CellSet{}
	} else if err != nil {
		logger.Debug("failed-listing-cells")
		return err
	}
	logger.Debug("succeeded-listing-cells")

	startRequests, keysWithMissingCells, keysToRetire := h.db.ConvergeLRPs(logger, cellSet)

	retireLogger := logger.WithData(lager.Data{"retiring_lrp_count": len(keysToRetire)})
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
			} else {
				bbsErr := models.ConvertError(err)
				exitIfUnrecoverable(logger, h.exitChan, bbsErr)
			}
		})
	}

	var throttler *workpool.Throttler
	throttler, err = workpool.NewThrottler(h.convergenceWorkersSize, works)
	if err != nil {
		logger.Error("failed-constructing-throttler", err, lager.Data{"max_workers": h.convergenceWorkersSize, "num_works": len(works)})
		return err
	}

	retireLogger.Debug("retiring-actual-lrps")
	throttler.Work()
	retireLogger.Debug("done-retiring-actual-lrps")

	startLogger := logger.WithData(lager.Data{"start_requests_count": len(startRequests)})
	if len(startRequests) > 0 {
		startLogger.Debug("requesting-start-auctions")
		err = h.auctioneerClient.RequestLRPAuctions(startRequests)
		if err != nil {
			startLogger.Error("failed-to-request-starts", err, lager.Data{"lrp_start_auctions": startRequests})
		}
		startLogger.Debug("done-requesting-start-auctions")
	}

	return err
}
