package handlers

import (
	"net/http"
	"sync"

	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/bbs/db"
	"code.cloudfoundry.org/bbs/events"
	"code.cloudfoundry.org/bbs/models"
	"github.com/cloudfoundry-incubator/auctioneer"
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
	response := &models.ConvergeLRPsResponse{}

	defer func() { exitIfUnrecoverable(logger, h.exitChan, response.Error) }()
	defer writeResponse(w, response)

	logger.Debug("listing-cells")
	cellSet, err := h.serviceClient.Cells(logger)
	if err == models.ErrResourceNotFound {
		logger.Debug("no-cells-found")
		cellSet = models.CellSet{}
	} else if err != nil {
		logger.Debug("failed-listing-cells")
		response.Error = models.ConvertError(err)
		return
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

	throttler, err := workpool.NewThrottler(h.convergenceWorkersSize, works)
	if err != nil {
		logger.Error("failed-constructing-throttler", err, lager.Data{"max_workers": h.convergenceWorkersSize, "num_works": len(works)})
		response.Error = models.ConvertError(err)
		return
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

	response.Error = models.ConvertError(err)
}
