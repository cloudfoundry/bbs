package sqldb

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cloudfoundry-incubator/auctioneer"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/runtime-schema/metric"
	"github.com/cloudfoundry/gunk/workpool"
	"github.com/pivotal-golang/lager"
)

const (
	convergeLRPRunsCounter = metric.Counter("ConvergenceLRPRuns")
	convergeLRPDuration    = metric.Duration("ConvergenceLRPDuration")

	domainMetricPrefix = "Domain."

	instanceLRPs  = metric.Metric("LRPsDesired") // this is the number of desired instances
	claimedLRPs   = metric.Metric("LRPsClaimed")
	unclaimedLRPs = metric.Metric("LRPsUnclaimed")
	runningLRPs   = metric.Metric("LRPsRunning")

	missingLRPs = metric.Metric("LRPsMissing")
	extraLRPs   = metric.Metric("LRPsExtra")

	crashedActualLRPs   = metric.Metric("CrashedActualLRPs")
	crashingDesiredLRPs = metric.Metric("CrashingDesiredLRPs")
)

func (db *SQLDB) ConvergeLRPs(logger lager.Logger, cellSet models.CellSet) ([]*auctioneer.LRPStartRequest, []*models.ActualLRPKeyWithSchedulingInfo, []*models.ActualLRPKey) {
	convergeStart := db.clock.Now()
	convergeLRPRunsCounter.Increment()
	logger.Info("starting")
	defer logger.Info("completed")

	defer func() {
		err := convergeLRPDuration.Send(time.Since(convergeStart))
		if err != nil {
			logger.Error("failed-sending-converge-lrp-duration-metric", err)
		}
	}()

	now := db.clock.Now()

	db.pruneDomains(logger, now)
	db.pruneEvacuatingActualLRPs(logger, now)

	domainSet, err := db.domainSet(logger)
	if err != nil {
		return nil, nil, nil
	}

	db.emitDomainMetrics(logger, domainSet)

	converge := newConvergence(db)
	converge.staleUnclaimedActualLRPs(logger, now)
	converge.actualLRPsWithMissingCells(logger, cellSet)
	converge.lrpInstanceCounts(logger, domainSet)
	converge.orphanedActualLRPs(logger)
	converge.crashedActualLRPs(logger, now)

	return converge.result(logger)
}

type convergence struct {
	*SQLDB

	guidsToStartRequests map[string]*auctioneer.LRPStartRequest
	startRequestsMutex   sync.Mutex

	keysWithMissingCells []*models.ActualLRPKeyWithSchedulingInfo

	keysToRetire []*models.ActualLRPKey
	keysMutex    sync.Mutex

	pool   *workpool.WorkPool
	poolWg sync.WaitGroup
}

func newConvergence(db *SQLDB) *convergence {
	pool, err := workpool.NewWorkPool(db.convergenceWorkersSize)
	if err != nil {
		panic(fmt.Sprintf("failing to create workpool is irrecoverable %v", err))
	}

	return &convergence{
		SQLDB:                db,
		guidsToStartRequests: map[string]*auctioneer.LRPStartRequest{},
		keysToRetire:         []*models.ActualLRPKey{},
		pool:                 pool,
	}
}

// Adds stale UNCLAIMED Actual LRPs to the list of start requests.
func (c *convergence) staleUnclaimedActualLRPs(logger lager.Logger, now time.Time) {
	logger = logger.Session("stale-unclaimed-actual-lrps")

	rows, err := c.db.Query(c.getQuery(SelectStaleLRPsQuery), models.ActualLRPStateUnclaimed,
		now.Add(-models.StaleUnclaimedActualLRPDuration).UnixNano(),
		false)
	if err != nil {
		logger.Error("failed-query", err)
		return
	}

	for rows.Next() {
		var index int
		schedulingInfo, err := c.fetchDesiredLRPSchedulingInfoAndMore(logger, rows, &index)
		if err == nil {
			c.addStartRequestFromSchedulingInfo(logger, schedulingInfo, index)
		}
	}

	if rows.Err() != nil {
		logger.Error("failed-getting-next-row", rows.Err())
	}

	return
}

// Adds CRASHED Actual LRPs that can be restarted to the list of start requests
// and transitions them to UNCLAIMED.
func (c *convergence) crashedActualLRPs(logger lager.Logger, now time.Time) {

	logger = logger.Session("crashed-actual-lrps")
	restartCalculator := models.NewDefaultRestartCalculator()

	rows, err := c.db.Query(c.getQuery(SelectLRPsByStateQuery), false, models.ActualLRPStateCrashed)
	if err != nil {
		logger.Error("failed-query", err)
		return
	}

	for rows.Next() {
		var index int
		actual := &models.ActualLRP{}

		schedulingInfo, err := c.fetchDesiredLRPSchedulingInfoAndMore(logger, rows, &index, &actual.Since, &actual.CrashCount)
		if err != nil {
			continue
		}

		actual.ProcessGuid = schedulingInfo.ProcessGuid
		actual.Domain = schedulingInfo.Domain
		actual.State = models.ActualLRPStateCrashed

		if actual.ShouldRestartCrash(now, restartCalculator) {
			c.submit(func() {
				_, _, err = c.UnclaimActualLRP(logger, &actual.ActualLRPKey)
				if err != nil {
					logger.Error("failed-unclaiming-actual-lrp", err)
					return
				}

				c.addStartRequestFromSchedulingInfo(logger, schedulingInfo, index)
			})
		}
	}

	if rows.Err() != nil {
		logger.Error("failed-getting-next-row", rows.Err())
	}

	return
}

// Adds orphaned Actual LRPs (ones with no corresponding Desired LRP) to the
// list of keys to retire.
func (c *convergence) orphanedActualLRPs(logger lager.Logger) {
	logger = logger.Session("orphaned-actual-lrps")

	rows, err := c.db.Query(c.getQuery(SelectOrphanedActualLRPsQuery), false)
	if err != nil {
		logger.Error("failed-query", err)
		return
	}

	for rows.Next() {
		actualLRPKey := &models.ActualLRPKey{}

		err := rows.Scan(
			&actualLRPKey.ProcessGuid,
			&actualLRPKey.Index,
			&actualLRPKey.Domain,
		)
		if err != nil {
			logger.Error("failed-scanning", err)
			continue
		}

		c.addKeyToRetire(logger, actualLRPKey)
	}

	if rows.Err() != nil {
		logger.Error("failed-getting-next-row", rows.Err())
	}
}

// Creates and adds missing Actual LRPs to the list of start requests.
// Adds extra Actual LRPs  to the list of keys to retire.
func (c *convergence) lrpInstanceCounts(logger lager.Logger, domainSet map[string]struct{}) {
	logger = logger.Session("lrp-instance-counts")

	rows, err := c.db.Query(c.getQuery(SelectLRPInstanceCountsQuery), false)
	if err != nil {
		logger.Error("failed-query", err)
		return
	}

	missingLRPCount := 0
	for rows.Next() {
		var existingIndicesStr sql.NullString
		var actualInstances int

		schedulingInfo, err := c.fetchDesiredLRPSchedulingInfoAndMore(logger, rows, &actualInstances, &existingIndicesStr)
		if err != nil {
			continue
		}

		indices := []int{}
		existingIndices := strings.Split(existingIndicesStr.String, ",")

		for i := 0; i < int(schedulingInfo.Instances); i++ {
			found := false
			for _, indexStr := range existingIndices {
				if indexStr == strconv.Itoa(i) {
					found = true
					break
				}
			}
			if !found {
				missingLRPCount++
				indices = append(indices, i)
				index := int32(i)

				c.submit(func() {
					_, err := c.CreateUnclaimedActualLRP(logger, &models.ActualLRPKey{ProcessGuid: schedulingInfo.ProcessGuid, Domain: schedulingInfo.Domain, Index: index})
					if err != nil {
						logger.Error("failed-creating-missing-actual-lrp", err)
					}
				})
			}
		}

		c.addStartRequestFromSchedulingInfo(logger, schedulingInfo, indices...)

		if actualInstances > int(schedulingInfo.Instances) {
			for i := int(schedulingInfo.Instances); i < actualInstances; i++ {
				if _, ok := domainSet[schedulingInfo.Domain]; ok {
					c.addKeyToRetire(logger, &models.ActualLRPKey{
						ProcessGuid: schedulingInfo.ProcessGuid,
						Index:       int32(i),
						Domain:      schedulingInfo.Domain,
					})
				}
			}
		}
	}

	if rows.Err() != nil {
		logger.Error("failed-getting-next-row", rows.Err())
	}

	missingLRPs.Send(missingLRPCount)
}

// Unclaim Actual LRPs that have missing cells (not in the cell set passed to
// convergence) and add them to the list of start requests.
func (c *convergence) actualLRPsWithMissingCells(logger lager.Logger, cellSet models.CellSet) {
	// time.Sleep(1000 * time.Second)
	logger = logger.Session("actual-lrps-with-missing-cells")

	values := make([]interface{}, 0, 1+len(cellSet))
	values = append(values, false)
	keysWithMissingCells := make([]*models.ActualLRPKeyWithSchedulingInfo, 0)

	for k := range cellSet {
		values = append(values, k)
	}

	query := c.getQuery(SelectLRPsQuery)
	if len(cellSet) != 0 {
		cellSetArgs := strings.Join(strings.Split(strings.Repeat("?", len(cellSet)), ""), ",")

		if c.flavor == Postgres {
			strParts := strings.Split(cellSetArgs, "?")
			for i := 1; i < len(strParts); i++ {
				strParts[i-1] += fmt.Sprintf("$%d", i+1)
			}
			cellSetArgs = strings.Join(strParts, "")
		}
		query = fmt.Sprintf(`%s AND actual_lrps.cell_id NOT IN (%s) AND actual_lrps.cell_id <> ''`,
			query, cellSetArgs)
	}

	stmt, err := c.db.Prepare(query)
	if err != nil {
		logger.Error("failed-preparing-query", err)
		return
	}

	rows, err := stmt.Query(values...)
	if err != nil {
		logger.Error("failed-query", err)
		return
	}

	for rows.Next() {
		var index int32
		schedulingInfo, err := c.fetchDesiredLRPSchedulingInfoAndMore(logger, rows, &index)
		if err == nil {
			keysWithMissingCells = append(keysWithMissingCells, &models.ActualLRPKeyWithSchedulingInfo{
				Key: &models.ActualLRPKey{
					ProcessGuid: schedulingInfo.ProcessGuid,
					Domain:      schedulingInfo.Domain,
					Index:       index,
				},
				SchedulingInfo: schedulingInfo,
			})
		}
	}

	if rows.Err() != nil {
		logger.Error("failed-getting-next-row", rows.Err())
	}

	c.keysWithMissingCells = keysWithMissingCells
}

func (c *convergence) addStartRequestFromSchedulingInfo(logger lager.Logger, schedulingInfo *models.DesiredLRPSchedulingInfo, indices ...int) {
	if len(indices) == 0 {
		return
	}

	c.startRequestsMutex.Lock()
	defer c.startRequestsMutex.Unlock()

	if startRequest, ok := c.guidsToStartRequests[schedulingInfo.ProcessGuid]; ok {
		startRequest.Indices = append(startRequest.Indices, indices...)
		return
	}

	startRequest := auctioneer.NewLRPStartRequestFromSchedulingInfo(schedulingInfo, indices...)
	c.guidsToStartRequests[schedulingInfo.ProcessGuid] = &startRequest
}

func (c *convergence) addKeyToRetire(logger lager.Logger, key *models.ActualLRPKey) {
	c.keysMutex.Lock()
	defer c.keysMutex.Unlock()

	c.keysToRetire = append(c.keysToRetire, key)
}

func (c *convergence) submit(work func()) {
	c.poolWg.Add(1)
	c.pool.Submit(func() {
		defer c.poolWg.Done()
		work()
	})
}

func (c *convergence) result(logger lager.Logger) ([]*auctioneer.LRPStartRequest, []*models.ActualLRPKeyWithSchedulingInfo, []*models.ActualLRPKey) {
	c.poolWg.Wait()
	c.startRequestsMutex.Lock()
	defer c.startRequestsMutex.Unlock()
	c.keysMutex.Lock()
	defer c.keysMutex.Unlock()

	startRequests := make([]*auctioneer.LRPStartRequest, 0, len(c.guidsToStartRequests))
	for _, startRequest := range c.guidsToStartRequests {
		startRequests = append(startRequests, startRequest)
	}

	extraLRPs.Send(len(c.keysToRetire))
	c.emitLRPMetrics(logger)

	return startRequests, c.keysWithMissingCells, c.keysToRetire
}

func (db *SQLDB) pruneDomains(logger lager.Logger, now time.Time) {
	logger = logger.Session("prune-domains")

	_, err := db.db.Exec(db.getQuery(PruneDomainsQuery), now.UnixNano())

	if err != nil {
		logger.Error("failed-query", err)
	}
}

func (db *SQLDB) pruneEvacuatingActualLRPs(logger lager.Logger, now time.Time) {
	logger = logger.Session("prune-evacuating-actual-lrps")

	_, err := db.db.Exec(db.getQuery(PruneActualLRPsQuery), true, now.UnixNano())
	if err != nil {
		logger.Error("failed-query", err)
	}
}

func (db *SQLDB) domainSet(logger lager.Logger) (map[string]struct{}, error) {
	logger.Debug("listing-domains")
	domains, err := db.Domains(logger)
	if err != nil {
		logger.Error("failed-listing-domains", err)
		return nil, err
	}
	logger.Debug("succeeded-listing-domains")
	m := make(map[string]struct{}, len(domains))
	for _, domain := range domains {
		m[domain] = struct{}{}
	}
	return m, nil
}

func (db *SQLDB) emitDomainMetrics(logger lager.Logger, domainSet map[string]struct{}) {
	for domain := range domainSet {
		metric.Metric(domainMetricPrefix + domain).Send(1)
	}
}

func (db *SQLDB) emitLRPMetrics(logger lager.Logger) {
	logger = logger.Session("emit-lrp-metrics")
	var desiredInstances, claimedInstances, unclaimedInstances, runningInstances, crashedInstances, crashingDesireds int

	row := db.db.QueryRow(db.getQuery(SelectLRPMetricsQuery),
		models.ActualLRPStateClaimed,
		models.ActualLRPStateUnclaimed,
		models.ActualLRPStateRunning,
		models.ActualLRPStateCrashed,
		models.ActualLRPStateCrashed,
		false,
	)

	err := row.Scan(&claimedInstances, &unclaimedInstances, &runningInstances, &crashedInstances, &crashingDesireds)
	if err != nil {
		logger.Error("failed-query", err)
	}

	row = db.db.QueryRow(db.getQuery(CountDesiredInstancesQuery))

	err = row.Scan(&desiredInstances)
	if err != nil {
		logger.Error("failed-desired-instances-query", err)
	}

	err = unclaimedLRPs.Send(unclaimedInstances)
	if err != nil {
		logger.Error("failed-sending-unclaimed-lrps-metric", err)
	}

	err = claimedLRPs.Send(claimedInstances)
	if err != nil {
		logger.Error("failed-sending-claimed-lrps-metric", err)
	}

	err = runningLRPs.Send(runningInstances)
	if err != nil {
		logger.Error("failed-sending-running-lrps-metric", err)
	}

	err = crashedActualLRPs.Send(crashedInstances)
	if err != nil {
		logger.Error("failed-sending-crashed-actual-lrps-metric", err)
	}

	err = crashingDesiredLRPs.Send(crashingDesireds)
	if err != nil {
		logger.Error("failed-sending-crashing-desired-lrps-metric", err)
	}

	err = instanceLRPs.Send(desiredInstances)
	if err != nil {
		logger.Error("failed-sending-desired-lrps-metric", err)
	}
}

func (db *SQLDB) GatherAndPruneLRPs(logger lager.Logger, cellSet models.CellSet) (*models.ConvergenceInput, error) {
	panic("not implemented")
}
