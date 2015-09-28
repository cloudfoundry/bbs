package etcd

import (
	"fmt"
	"path"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cloudfoundry-incubator/auctioneer"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry/gunk/workpool"

	"github.com/cloudfoundry-incubator/runtime-schema/metric"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
)

const (
	convergeLRPRunsCounter = metric.Counter("ConvergenceLRPRuns")
	convergeLRPDuration    = metric.Duration("ConvergenceLRPDuration")

	malformedSchedulingInfosMetric = metric.Counter("ConvergenceLRPPreProcessingMalformedSchedulingInfos")
	malformedRunInfosMetric        = metric.Counter("ConvergenceLRPPreProcessingMalformedRunInfos")
	actualLRPsDeleted              = metric.Counter("ConvergenceLRPPreProcessingActualLRPsDeleted")
	lrpsDeletedCounter             = metric.Counter("ConvergenceLRPsDeleted")

	domainMetricPrefix = "Domain."

	desiredLRPs         = metric.Metric("LRPsDesired")
	startingLRPs        = metric.Metric("LRPsStarting")
	runningLRPs         = metric.Metric("LRPsRunning")
	crashedActualLRPs   = metric.Metric("CrashedActualLRPs")
	crashingDesiredLRPs = metric.Metric("CrashingDesiredLRPs")
)

func (db *ETCDDB) ConvergeLRPs(logger lager.Logger) {
	convergeStart := db.clock.Now()
	convergeLRPRunsCounter.Increment()
	logger = logger.Session("converge-lrps")
	logger.Info("starting-convergence")
	defer logger.Info("finished-convergence")

	defer func() {
		convergeLRPDuration.Send(time.Since(convergeStart))
	}()

	logger.Debug("gathering-convergence-input")
	input, err := db.GatherAndPruneLRPs(logger)
	if err != nil {
		logger.Error("failed-gathering-convergence-input", err)
		return
	}
	logger.Debug("succeeded-gathering-convergence-input")

	changes := CalculateConvergence(logger, db.clock, models.NewDefaultRestartCalculator(), input)

	db.ResolveConvergence(logger, input.DesiredLRPs, changes)
}

func (db *ETCDDB) GatherAndPruneLRPs(logger lager.Logger) (*models.ConvergenceInput, error) {
	guids := map[string]struct{}{}

	// always fetch actualLRPs before desiredLRPs to ensure correctness
	logger.Info("gathering-and-pruning-actual-lrps")
	actuals, err := db.gatherAndPruneActualLRPs(logger, guids) // modifies guids
	if err != nil {
		logger.Error("failed-gathering-and-pruning-actual-lrps", err)

		startingLRPs.Send(-1)
		runningLRPs.Send(-1)
		crashedActualLRPs.Send(-1)
		crashingDesiredLRPs.Send(-1)
		desiredLRPs.Send(-1)

		return &models.ConvergenceInput{}, err
	}
	logger.Info("succeeded-gathering-and-pruning-actual-lrps")

	// always fetch desiredLRPs after actualLRPs to ensure correctness
	logger.Info("gathering-desired-lrps")
	desireds, err := db.GatherDesiredLRPs(logger, guids) // modifies guids
	if err != nil {
		logger.Error("failed-gathering-desired-lrps", err)

		desiredLRPs.Send(-1)

		return &models.ConvergenceInput{}, err
	}
	logger.Info("succeeded-gathering-desired-lrps")

	logger.Debug("listing-domains")
	domains, err := db.Domains(logger)
	if err != nil {
		return &models.ConvergenceInput{}, err
	}
	logger.Debug("succeeded-listing-domains")
	for _, domain := range domains {
		metric.Metric(domainMetricPrefix + domain).Send(1)
	}

	logger.Debug("listing-cells")
	cellSet, modelErr := db.serviceClient.Cells(logger)
	if modelErr != nil {
		if !models.ErrResourceNotFound.Equal(modelErr) {
			logger.Debug("failed-listing-cells")
			return &models.ConvergenceInput{}, modelErr
		}
		logger.Debug("no-cells-found")
		cellSet = models.CellSet{}
	}
	logger.Debug("succeeded-listing-cells")

	return &models.ConvergenceInput{
		AllProcessGuids: guids,
		DesiredLRPs:     desireds,
		ActualLRPs:      actuals,
		Domains:         models.NewDomainSet(domains),
		Cells:           cellSet,
	}, nil
}

func (db *ETCDDB) gatherAndPruneActualLRPs(logger lager.Logger, guids map[string]struct{}) (map[string]map[int32]*models.ActualLRP, error) {
	return db.gatherAndOptionallyPruneActualLRPs(logger, guids, true)
}

func (db *ETCDDB) GatherActualLRPs(logger lager.Logger, guids map[string]struct{}) (map[string]map[int32]*models.ActualLRP, error) {
	return db.gatherAndOptionallyPruneActualLRPs(logger, guids, false)
}

func (db *ETCDDB) gatherAndOptionallyPruneActualLRPs(logger lager.Logger, guids map[string]struct{}, doPrune bool) (map[string]map[int32]*models.ActualLRP, error) {
	response, modelErr := db.fetchRecursiveRaw(logger, ActualLRPSchemaRoot)

	if modelErr == models.ErrResourceNotFound {
		logger.Info("actual-lrp-schema-root-not-found")
		return map[string]map[int32]*models.ActualLRP{}, nil
	}

	if modelErr != nil {
		return nil, modelErr
	}

	actuals := map[string]map[int32]*models.ActualLRP{}
	var guidKeysToDelete, indexKeysToDelete []string
	var actualsToDelete []string
	var guidsLock, actualsLock, guidKeysToDeleteLock, indexKeysToDeleteLock,
		crashingDesiredsLock, actualsToDeleteLock sync.Mutex

	logger.Info("walking-actual-lrp-tree")
	works := []func(){}
	var startingCount int32 = 0
	var runningCount int32 = 0
	var crashedCount int32 = 0
	crashingDesireds := map[string]struct{}{}

	for _, guidGroup := range response.Nodes {
		guidGroup := guidGroup
		works = append(works, func() {
			guidGroupWillBeEmpty := true

			for _, indexGroup := range guidGroup.Nodes {
				indexGroupWillBeEmpty := true

				for _, actualNode := range indexGroup.Nodes {
					actual := new(models.ActualLRP)
					err := db.deserializeModel(logger, actualNode, actual)
					if err != nil {
						actualsToDeleteLock.Lock()
						actualsToDelete = append(actualsToDelete, actualNode.Key)
						actualsToDeleteLock.Unlock()

						continue
					}

					indexGroupWillBeEmpty = false
					guidGroupWillBeEmpty = false

					switch actual.State {
					case models.ActualLRPStateClaimed:
						atomic.AddInt32(&startingCount, 1)
					case models.ActualLRPStateRunning:
						atomic.AddInt32(&runningCount, 1)
					case models.ActualLRPStateCrashed:
						crashingDesiredsLock.Lock()
						crashingDesireds[actual.ProcessGuid] = struct{}{}
						crashingDesiredsLock.Unlock()
						atomic.AddInt32(&crashedCount, 1)
					}

					guidsLock.Lock()
					guids[actual.ProcessGuid] = struct{}{}
					guidsLock.Unlock()

					if path.Base(actualNode.Key) == ActualLRPInstanceKey {
						actualsLock.Lock()
						if actuals[actual.ProcessGuid] == nil {
							actuals[actual.ProcessGuid] = map[int32]*models.ActualLRP{}
						}
						actuals[actual.ProcessGuid][actual.Index] = actual
						actualsLock.Unlock()
					}
				}

				if indexGroupWillBeEmpty {
					indexKeysToDeleteLock.Lock()
					indexKeysToDelete = append(indexKeysToDelete, indexGroup.Key)
					indexKeysToDeleteLock.Unlock()
				}
			}

			if guidGroupWillBeEmpty {
				guidKeysToDeleteLock.Lock()
				guidKeysToDelete = append(guidKeysToDelete, guidGroup.Key)
				guidKeysToDeleteLock.Unlock()
			}
		})
	}
	logger.Info("done-walking-actual-lrp-tree")

	throttler, err := workpool.NewThrottler(throttlerSize, works)
	if err != nil {
		logger.Error("failed-to-create-throttler", err)
	}

	throttler.Work()

	if doPrune {
		logger.Info("deleting-invalid-actual-lrps", lager.Data{"num-lrps": len(actualsToDelete)})
		db.batchDeleteNodes(actualsToDelete, logger)
		actualLRPsDeleted.Add(uint64(len(actualsToDelete)))

		logger.Info("deleting-empty-actual-indices", lager.Data{"num-indices": len(indexKeysToDelete)})
		err = db.deleteLeaves(logger, indexKeysToDelete)
		if err != nil {
			logger.Error("failed-deleting-empty-actual-indices", err, lager.Data{"num-indices": len(indexKeysToDelete)})
		} else {
			logger.Info("succeeded-deleting-empty-actual-indices", lager.Data{"num-indices": len(indexKeysToDelete)})
		}

		logger.Info("deleting-empty-actual-guids", lager.Data{"num-guids": len(guidKeysToDelete)})
		err = db.deleteLeaves(logger, guidKeysToDelete)
		if err != nil {
			logger.Error("failed-deleting-empty-actual-guids", err, lager.Data{"num-guids": len(guidKeysToDelete)})
		} else {
			logger.Info("succeeded-deleting-empty-actual-guids", lager.Data{"num-guids": len(guidKeysToDelete)})
		}

		startingLRPs.Send(int(startingCount))
		runningLRPs.Send(int(runningCount))
		crashedActualLRPs.Send(int(crashedCount))
		crashingDesiredLRPs.Send(len(crashingDesireds))
	}

	return actuals, nil
}

func (db *ETCDDB) deleteLeaves(logger lager.Logger, keys []string) error {
	works := []func(){}

	for _, key := range keys {
		key := key
		works = append(works, func() {
			_, err := db.client.DeleteDir(key)
			if err != nil {
				logger.Error("failed-deleting-leaf-node", err, lager.Data{"key": key})
			}
		})
	}

	throttler, err := workpool.NewThrottler(throttlerSize, works)
	if err != nil {
		return err
	}

	throttler.Work()

	return nil
}

func (db *ETCDDB) GatherDesiredLRPs(logger lager.Logger, guids map[string]struct{}) (map[string]*models.DesiredLRP, error) {
	desiredLRPsRoot, modelErr := db.fetchRecursiveRaw(logger, DesiredLRPComponentsSchemaRoot)

	if modelErr == models.ErrResourceNotFound {
		logger.Info("actual-lrp-schema-root-not-found")
		return map[string]*models.DesiredLRP{}, nil
	}

	if modelErr != nil {
		return nil, modelErr
	}

	schedulingInfos := map[string]*models.DesiredLRPSchedulingInfo{}
	runInfos := map[string]*models.DesiredLRPRunInfo{}

	var malformedSchedulingInfos int32
	var malformedRunInfos int32
	var guidsLock, schedulingInfosLock, runInfosLock sync.Mutex
	var desiredInstanceCount int32

	works := []func(){}
	logger.Info("walking-desired-lrp-components-tree")

	for _, componentRoot := range desiredLRPsRoot.Nodes {
		switch componentRoot.Key {
		case DesiredLRPSchedulingInfoSchemaRoot:
			for _, node := range componentRoot.Nodes {
				node := node
				works = append(works, func() {
					var schedulingInfo models.DesiredLRPSchedulingInfo
					err := db.deserializeModel(logger, node, &schedulingInfo)
					if err != nil {
						logger.Error("failed-to-deserialize-scheduling-info", err)
						atomic.AddInt32(&malformedSchedulingInfos, 1)
					} else {
						schedulingInfosLock.Lock()
						schedulingInfos[schedulingInfo.ProcessGuid] = &schedulingInfo
						schedulingInfosLock.Unlock()
						atomic.AddInt32(&desiredInstanceCount, schedulingInfo.Instances)

						guidsLock.Lock()
						guids[schedulingInfo.ProcessGuid] = struct{}{}
						guidsLock.Unlock()
					}
				})
			}
		case DesiredLRPRunInfoSchemaRoot:
			for _, node := range componentRoot.Nodes {
				node := node
				works = append(works, func() {
					var runInfo models.DesiredLRPRunInfo
					err := db.deserializeModel(logger, node, &runInfo)
					if err != nil {
						logger.Error("failed-to-deserialize-run-info", err)
						atomic.AddInt32(&malformedRunInfos, 1)
					} else {
						runInfosLock.Lock()
						runInfos[runInfo.ProcessGuid] = &runInfo
						runInfosLock.Unlock()
					}
				})
			}
		default:
			err := fmt.Errorf("unrecognized node under desired LRPs root node: %s", componentRoot.Key)
			logger.Error("unrecognized-node", err)
			return nil, err
		}
	}

	throttler, err := workpool.NewThrottler(throttlerSize, works)
	if err != nil {
		logger.Error("failed-to-create-throttler", err)
	}

	throttler.Work()

	malformedSchedulingInfosMetric.Add(uint64(malformedSchedulingInfos))
	malformedRunInfosMetric.Add(uint64(malformedRunInfos))
	desiredLRPs.Send(int(desiredInstanceCount))

	logger.Info("done-walking-desired-lrp-tree")

	desireds := make(map[string]*models.DesiredLRP)
	for guid, schedulingInfo := range schedulingInfos {
		runInfo := runInfos[guid]
		desiredLRP := models.NewDesiredLRP(*schedulingInfo, *runInfo)
		desireds[guid] = &desiredLRP
	}

	return desireds, nil
}

func (db *ETCDDB) batchDeleteNodes(keys []string, logger lager.Logger) {
	if len(keys) == 0 {
		return
	}

	works := []func(){}

	for _, key := range keys {
		key := key
		works = append(works, func() {
			logger.Info("deleting", lager.Data{"key": key})
			_, err := db.client.Delete(key, true)
			if err != nil {
				logger.Error("failed-to-delete", err, lager.Data{
					"key": key,
				})
			}
		})
	}

	throttler, err := workpool.NewThrottler(throttlerSize, works)
	if err != nil {
		logger.Error("failed-to-create-throttler", err)
	}

	throttler.Work()
	return
}

func CalculateConvergence(
	logger lager.Logger,
	clock clock.Clock,
	restartCalculator models.RestartCalculator,
	input *models.ConvergenceInput,
) *models.ConvergenceChanges {
	sess := logger.Session("calculate-convergence")

	sess.Info("start")
	defer sess.Info("done")

	changes := &models.ConvergenceChanges{}

	now := clock.Now()

	for processGuid, _ := range input.AllProcessGuids {
		pLog := sess.WithData(lager.Data{
			"process-guid": processGuid,
		})

		desired, hasDesired := input.DesiredLRPs[processGuid]

		actualsByIndex := input.ActualLRPs[processGuid]

		if hasDesired {
			for i := int32(0); i < desired.Instances; i++ {
				if _, hasIndex := actualsByIndex[i]; !hasIndex {
					pLog.Info("missing", lager.Data{"index": i})
					lrpKey := models.NewActualLRPKey(desired.ProcessGuid, i, desired.Domain)
					changes.ActualLRPKeysForMissingIndices = append(
						changes.ActualLRPKeysForMissingIndices,
						&lrpKey,
					)
				}
			}

			for i, actual := range actualsByIndex {
				if actual.CellIsMissing(input.Cells) {
					pLog.Info("missing-cell", lager.Data{"index": i})
					changes.ActualLRPsWithMissingCells = append(changes.ActualLRPsWithMissingCells, actual)
					continue
				}

				if actual.Index >= desired.Instances && input.Domains.Contains(desired.Domain) {
					pLog.Info("extra", lager.Data{"index": i})
					changes.ActualLRPsForExtraIndices = append(changes.ActualLRPsForExtraIndices, actual)
					continue
				}

				if actual.ShouldRestartCrash(now, restartCalculator) {
					pLog.Info("restart-crash", lager.Data{"index": i})
					changes.RestartableCrashedActualLRPs = append(changes.RestartableCrashedActualLRPs, actual)
					continue
				}

				if actual.ShouldStartUnclaimed(now) {
					pLog.Info("stale-unclaimed", lager.Data{"index": i})
					changes.StaleUnclaimedActualLRPs = append(changes.StaleUnclaimedActualLRPs, actual)
					continue
				}
			}
		} else {
			for i, actual := range actualsByIndex {
				if actual.CellIsMissing(input.Cells) {
					pLog.Info("missing-cell", lager.Data{"index": i})
					changes.ActualLRPsWithMissingCells = append(changes.ActualLRPsWithMissingCells, actual)
					continue
				}

				if !input.Domains.Contains(actual.Domain) {
					pLog.Info("skipping-unfresh-domain")
					continue
				}

				pLog.Info("no-longer-desired", lager.Data{"index": i})
				changes.ActualLRPsForExtraIndices = append(changes.ActualLRPsForExtraIndices, actual)
			}
		}
	}

	return changes
}

func (db *ETCDDB) ResolveConvergence(logger lager.Logger, desiredLRPs map[string]*models.DesiredLRP, changes *models.ConvergenceChanges) {
	startRequests := newStartRequests(desiredLRPs)
	for _, actual := range changes.StaleUnclaimedActualLRPs {
		startRequests.Add(logger, &actual.ActualLRPKey)
	}

	works := []func(){}

	for _, actual := range changes.ActualLRPsForExtraIndices {
		works = append(works, db.resolveActualsToBeRetired(logger, actual))
	}

	for _, actual := range changes.ActualLRPsWithMissingCells {
		works = append(works, db.resolveActualsWithMissingCells(logger, desiredLRPs[actual.ProcessGuid], actual, startRequests))
	}

	for _, actualKey := range changes.ActualLRPKeysForMissingIndices {
		works = append(works, db.resolveActualsWithMissingIndices(logger, desiredLRPs[actualKey.ProcessGuid], actualKey, startRequests))
	}

	for _, actual := range changes.RestartableCrashedActualLRPs {
		works = append(works, db.resolveRestartableCrashedActualLRPS(logger, actual, startRequests))
	}

	throttler, err := workpool.NewThrottler(throttlerSize, works)
	if err != nil {
		logger.Error("failed-constructing-throttler", err, lager.Data{"max-workers": throttlerSize, "num-works": len(works)})
		return
	}

	logger.Debug("waiting-for-lrp-convergence-work")
	throttler.Work()
	logger.Debug("done-waiting-for-lrp-convergence-work")

	logger.Debug("requesting-start-auctions", lager.Data{"start-requests-instance-count": startRequests.InstanceCount()})
	db.startActualLRPs(logger, startRequests)
	logger.Debug("done-requesting-start-auctions", lager.Data{"start-requests-instance-count": startRequests.InstanceCount()})
}

func (db *ETCDDB) resolveActualsToBeRetired(logger lager.Logger, actual *models.ActualLRP) func() {
	return func() {
		logger = logger.Session("start-retire-actual", lager.Data{
			"process-guid": actual.ProcessGuid,
			"index":        actual.Index,
		})
		logger.Debug("retiring-actual-lrp")
		retireErr := db.RetireActualLRP(logger, &actual.ActualLRPKey)
		if retireErr != nil {
			logger.Error("failed-retiring-actual-lrp", retireErr)
			return
		}
		logger.Debug("succeeded-retiring-actual-lrp")
	}
}

func (db *ETCDDB) resolveActualsWithMissingCells(logger lager.Logger, desired *models.DesiredLRP, actual *models.ActualLRP, starts *startRequests) func() {
	return func() {
		logger = logger.Session("start-missing-actual", lager.Data{
			"process-guid": actual.ProcessGuid,
			"index":        actual.Index,
		})

		logger.Debug("removing-actual-lrp")
		removeErr := db.RemoveActualLRP(logger, actual.ActualLRPKey.ProcessGuid, actual.ActualLRPKey.Index)
		if removeErr != nil {
			logger.Error("failed-removing-actual-lrp", removeErr)
			return
		}
		logger.Debug("succeeded-removing-actual-lrp")

		if actual.Index >= desired.Instances {
			return
		}

		logger.Debug("creating-actual-lrp")
		err := db.createActualLRP(logger, desired, actual.Index)
		if err != nil {
			logger.Error("failed-creating-actual-lrp", err)
			return
		}
		logger.Debug("succeeded-creating-actual-lrp")

		starts.Add(logger, &actual.ActualLRPKey)
	}
}

func (db *ETCDDB) resolveActualsWithMissingIndices(logger lager.Logger, desired *models.DesiredLRP, actualKey *models.ActualLRPKey, starts *startRequests) func() {
	return func() {
		logger = logger.Session("start-missing-actual", lager.Data{
			"process-guid": actualKey.ProcessGuid,
			"index":        actualKey.Index,
		})

		logger.Debug("creating-actual-lrp")
		err := db.createActualLRP(logger, desired, actualKey.Index)
		if err != nil {
			logger.Error("failed-creating-actual-lrp", err)
			return
		}
		logger.Debug("succeeded-creating-actual-lrp")

		starts.Add(logger, actualKey)
	}
}

func (db *ETCDDB) resolveRestartableCrashedActualLRPS(logger lager.Logger, actualLRP *models.ActualLRP, starts *startRequests) func() {
	return func() {
		actualKey := actualLRP.ActualLRPKey

		logger = logger.Session("restart-crash", lager.Data{
			"process-guid": actualKey.ProcessGuid,
			"index":        actualKey.Index,
		})

		if actualLRP.State != models.ActualLRPStateCrashed {
			logger.Error("failed-actual-lrp-state-is-not-crashed", nil)
			return
		}

		logger.Debug("unclaiming-actual-lrp", lager.Data{"ProcessGuid": actualLRP.ActualLRPKey.ProcessGuid, "Index": actualLRP.ActualLRPKey.Index})
		_, err := db.unclaimActualLRP(logger, &actualLRP.ActualLRPKey, &actualLRP.ActualLRPInstanceKey)
		if err != nil {
			logger.Error("failed-unclaiming-crash", err)
			return
		}
		logger.Debug("succeeded-unclaiming-actual-lrp")

		starts.Add(logger, &actualKey)
	}
}

func (db *ETCDDB) startActualLRPs(logger lager.Logger, starts *startRequests) {
	count := starts.InstanceCount()
	if count == 0 {
		return
	}

	err := db.auctioneerClient.RequestLRPAuctions(starts.Slice())
	if err != nil {
		logger.Error("failed-to-request-starts", err, lager.Data{
			"lrp-start-auctions": starts,
		})
	}
}

type startRequests struct {
	desiredMap    map[string]*models.DesiredLRP
	startMap      map[string]*auctioneer.LRPStartRequest
	instanceCount uint64
	*sync.Mutex
}

func newStartRequests(desiredMap map[string]*models.DesiredLRP) *startRequests {
	return &startRequests{
		desiredMap: desiredMap,
		startMap:   make(map[string]*auctioneer.LRPStartRequest),
		Mutex:      new(sync.Mutex),
	}
}

func (s *startRequests) Add(logger lager.Logger, actual *models.ActualLRPKey) {
	s.Lock()
	defer s.Unlock()

	desiredLRP, found := s.desiredMap[actual.ProcessGuid]
	if !found {
		logger.Info("failed-to-find-desired-lrp-for-stale-unclaimed-actual-lrp", lager.Data{"actual-lrp": actual})
		return
	}

	start, found := s.startMap[desiredLRP.ProcessGuid]
	if !found {
		startRequest := auctioneer.NewLRPStartRequestFromModel(desiredLRP, int(actual.Index))
		start = &startRequest
	} else {
		start.Indices = append(start.Indices, int(actual.Index))
	}

	logger.Info("adding-start-auction", lager.Data{"process-guid": desiredLRP.ProcessGuid, "index": actual.Index})
	s.startMap[desiredLRP.ProcessGuid] = start
	s.instanceCount++
}

func (s *startRequests) Slice() []*auctioneer.LRPStartRequest {
	s.Lock()
	defer s.Unlock()

	starts := make([]*auctioneer.LRPStartRequest, 0, len(s.startMap))
	for _, start := range s.startMap {
		starts = append(starts, start)
	}
	return starts
}

func (s *startRequests) InstanceCount() uint64 {
	s.Lock()
	defer s.Unlock()

	return s.instanceCount
}
