package controllers

import (
	"sync"

	"code.cloudfoundry.org/auctioneer"
	"code.cloudfoundry.org/bbs/db"
	"code.cloudfoundry.org/bbs/events"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/serviceclient"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/workpool"
)

//go:generate counterfeiter -o fakes/fake_retirer.go . Retirer
type Retirer interface {
	RetireActualLRP(logger lager.Logger, key *models.ActualLRPKey) error
}

type LRPConvergenceController struct {
	logger                    lager.Logger
	lrpDB                     db.LRPDB
	suspectDB                 db.SuspectDB
	actualHub                 events.Hub
	auctioneerClient          auctioneer.Client
	serviceClient             serviceclient.ServiceClient
	retirer                   Retirer
	convergenceWorkersSize    int
	generateSuspectAcutalLRPs bool
}

func NewLRPConvergenceController(
	logger lager.Logger,
	db db.LRPDB,
	suspectDB db.SuspectDB,
	actualHub events.Hub,
	auctioneerClient auctioneer.Client,
	serviceClient serviceclient.ServiceClient,
	retirer Retirer,
	convergenceWorkersSize int,
	generateSuspectAcutalLRPs bool,
) *LRPConvergenceController {
	return &LRPConvergenceController{
		logger:                    logger,
		lrpDB:                     db,
		suspectDB:                 suspectDB,
		actualHub:                 actualHub,
		auctioneerClient:          auctioneerClient,
		serviceClient:             serviceClient,
		retirer:                   retirer,
		convergenceWorkersSize:    convergenceWorkersSize,
		generateSuspectAcutalLRPs: generateSuspectAcutalLRPs,
	}
}

func (h *LRPConvergenceController) ConvergeLRPs(logger lager.Logger) {
	logger = h.logger.Session("converge-lrps")
	var err error

	logger.Debug("listing-cells")
	var cellSet models.CellSet
	cellSet, err = h.serviceClient.Cells(logger)
	if err == models.ErrResourceNotFound {
		logger.Info("no-cells-found")
		cellSet = models.CellSet{}
	} else if err != nil {
		logger.Error("failed-listing-cells", err)
		// convergence should run again later
		return
	}
	logger.Debug("succeeded-listing-cells")

	convergenceResult := h.lrpDB.ConvergeLRPs(logger, cellSet)
	keysToRetire := convergenceResult.KeysToRetire
	events := convergenceResult.Events

	for _, e := range events {
		go h.actualHub.Emit(e)
	}
	retireLogger := logger.WithData(lager.Data{"retiring_lrp_count": len(keysToRetire)})
	works := []func(){}
	for _, key := range keysToRetire {
		key := key
		works = append(works, func() {
			err := h.retirer.RetireActualLRP(retireLogger, key)
			if err != nil {
				logger.Error("retiring-lrp-failed", err)
			}
		})
	}

	startRequests := []*auctioneer.LRPStartRequest{}
	startRequestLock := &sync.Mutex{}

	defer func() {
		startLogger := logger.WithData(lager.Data{"start_requests_count": len(startRequests)})
		if len(startRequests) > 0 {
			startLogger.Debug("requesting-start-auctions")
			err = h.auctioneerClient.RequestLRPAuctions(logger, startRequests)
			if err != nil {
				startLogger.Error("failed-to-request-starts", err, lager.Data{"lrp_start_auctions": startRequests})
			}
			startLogger.Debug("done-requesting-start-auctions")
		}
	}()

	for _, lrpKey := range convergenceResult.MissingLRPKeys {
		key := lrpKey
		works = append(works, func() {
			lrpGroup, err := h.lrpDB.CreateUnclaimedActualLRP(logger, key.Key)
			if err != nil {
				logger.Error("failed-to-create-unclaimed-lrp", err, lager.Data{"key": key.Key})
				return
			}

			go h.actualHub.Emit(models.NewActualLRPCreatedEvent(lrpGroup))

			startRequest := auctioneer.NewLRPStartRequestFromSchedulingInfo(key.SchedulingInfo, int(key.Key.Index))
			startRequestLock.Lock()
			startRequests = append(startRequests, &startRequest)
			startRequestLock.Unlock()
		})
	}

	for _, lrpKey := range convergenceResult.UnstartedLRPKeys {
		key := lrpKey
		works = append(works, func() {
			before, after, err := h.lrpDB.UnclaimActualLRP(logger, key.Key)
			if err != nil && err != models.ErrActualLRPCannotBeUnclaimed {
				logger.Error("cannot-unclaim-lrp", err, lager.Data{"key": key})
				return
			} else if !after.Equal(before) {
				logger.Info("emitting-changed-event", lager.Data{"before": before, "after": after})
				go h.actualHub.Emit(models.NewActualLRPChangedEvent(before, after))
			}

			startRequest := auctioneer.NewLRPStartRequestFromSchedulingInfo(key.SchedulingInfo, int(key.Key.Index))
			startRequestLock.Lock()
			startRequests = append(startRequests, &startRequest)
			startRequestLock.Unlock()
		})
	}

	for _, key := range convergenceResult.KeysWithMissingCells {
		key := key
		var handleLRP func()
		if h.generateSuspectAcutalLRPs {
			handleLRP = func() {
				logger := logger.Session("keys-with-missing-cells")

				_, _, err := h.lrpDB.ChangeActualLRPPresence(logger, key.Key, models.ActualLRP_Ordinary, models.ActualLRP_Suspect)
				if err == models.ErrResourceExists {
					logger.Debug("found-suspect-lrp-unclaiming", lager.Data{"key": key.Key})
					// there is a Suspect LRP already, unclaim this one and reauction it
					_, _, err := h.lrpDB.UnclaimActualLRP(logger, key.Key)
					if err != nil {
						logger.Error("failed-unclaiming-lrp", err)
						return
					}

					return
				}

				if err != nil {
					logger.Error("failed-changing-presence", err)
					return
				}

				_, err = h.lrpDB.CreateUnclaimedActualLRP(logger.Session("create-unclaimed-actual"), key.Key)
				if err != nil {
					logger.Error("cannot-unclaim-lrp", err)
					return
				}

				startRequest := auctioneer.NewLRPStartRequestFromSchedulingInfo(key.SchedulingInfo, int(key.Key.Index))
				startRequestLock.Lock()
				startRequests = append(startRequests, &startRequest)
				startRequestLock.Unlock()
				logger.Info("creating-start-request",
					lager.Data{"reason": "missing-cell", "process_guid": key.Key.ProcessGuid, "index": key.Key.Index})
			}
		} else {
			handleLRP = func() {
				before, after, err := h.lrpDB.UnclaimActualLRP(logger, key.Key)
				if err == nil {
					h.actualHub.Emit(models.NewActualLRPChangedEvent(before, after))
					startRequest := auctioneer.NewLRPStartRequestFromSchedulingInfo(key.SchedulingInfo, int(key.Key.Index))
					startRequestLock.Lock()
					startRequests = append(startRequests, &startRequest)
					startRequestLock.Unlock()
				} else {
					bbsErr := models.ConvertError(err)
					if bbsErr.GetType() != models.Error_Unrecoverable {
						return
					}

					logger.Error("unrecoverable-error", bbsErr)
				}
			}
		}
		works = append(works, handleLRP)
	}

	for _, key := range convergenceResult.SuspectKeysWithExistingCells {
		key := key
		works = append(works, func() {
			logger := logger.Session("suspect-keys-with-existing-cells")
			err := h.lrpDB.RemoveActualLRP(logger, key.ProcessGuid, key.Index, nil)
			if err != nil {
				logger.Error("cannot-remove-lrp", err, lager.Data{"key": key})
				return
			}
			_, _, err = h.lrpDB.ChangeActualLRPPresence(logger, key, models.ActualLRP_Suspect, models.ActualLRP_Ordinary)
			if err != nil {
				logger.Error("cannot-change-lrp-presence", err, lager.Data{"key": key})
				return
			}
		})
	}

	for _, key := range convergenceResult.SuspectLRPKeysToRetire {
		key := key
		works = append(works, func() {
			logger := logger.Session("suspect-keys-to-retire")
			suspectLRPGroup, err := h.suspectDB.RemoveSuspectActualLRP(logger, key)
			if err != nil {
				logger.Error("cannot-remove-suspect-lrp", err, lager.Data{"key": key})
				return
			}

			go h.actualHub.Emit(models.NewActualLRPRemovedEvent(suspectLRPGroup))
		})
	}

	var throttler *workpool.Throttler
	throttler, err = workpool.NewThrottler(h.convergenceWorkersSize, works)
	if err != nil {
		logger.Error("failed-constructing-throttler", err, lager.Data{"max_workers": h.convergenceWorkersSize, "num_works": len(works)})
		return
	}

	retireLogger.Debug("retiring-actual-lrps")
	throttler.Work()
	retireLogger.Debug("done-retiring-actual-lrps")

	return
}
