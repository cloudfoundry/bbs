package sqldb

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"code.cloudfoundry.org/bbs/db"
	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

const (
	domainMetricPrefix = "Domain."

	convergeLRPRunsCounter = "ConvergenceLRPRuns"
	convergeLRPDuration    = "ConvergenceLRPDuration"

	instanceLRPsMetric  = "LRPsDesired"
	claimedLRPsMetric   = "LRPsClaimed"
	unclaimedLRPsMetric = "LRPsUnclaimed"
	runningLRPsMetric   = "LRPsRunning"

	missingLRPsMetric = "LRPsMissing"
	extraLRPsMetric   = "LRPsExtra"

	crashedActualLRPsMetric   = "CrashedActualLRPs"
	crashingDesiredLRPsMetric = "CrashingDesiredLRPs"
)

func (sqldb *SQLDB) ConvergeLRPs(logger lager.Logger, cellSet models.CellSet) db.ConvergenceResult {
	convergeStart := sqldb.clock.Now()
	sqldb.metronClient.IncrementCounter(convergeLRPRunsCounter)
	logger.Info("starting")
	defer logger.Info("completed")

	defer func() {
		err := sqldb.metronClient.SendDuration(convergeLRPDuration, time.Since(convergeStart))
		if err != nil {
			logger.Error("failed-sending-converge-lrp-duration-metric", err)
		}
	}()

	now := sqldb.clock.Now()

	sqldb.pruneDomains(logger, now)
	events := sqldb.pruneEvacuatingActualLRPs(logger, cellSet)
	domainSet, err := sqldb.domainSet(logger)
	if err != nil {
		return db.ConvergenceResult{}
	}

	sqldb.emitDomainMetrics(logger, domainSet)

	converge := newConvergence(sqldb)
	converge.staleUnclaimedActualLRPs(logger, now)
	converge.actualLRPsWithMissingCells(logger, cellSet)
	converge.lrpInstanceCounts(logger, domainSet)
	converge.orphanedActualLRPs(logger)
	converge.extraSuspectActualLRPs(logger)
	converge.suspectActualLRPsWithExistingCells(logger, cellSet)
	converge.suspectActualLRPs(logger)
	converge.crashedActualLRPs(logger, now)

	return db.ConvergenceResult{
		MissingLRPKeys:         converge.result(logger),
		UnstartedLRPKeys:       converge.unstartedLRPKeys,
		KeysToRetire:           converge.keysToRetire,
		SuspectLRPKeysToRetire: converge.suspectKeysToRetire,
		KeysWithMissingCells:   converge.keysWithMissingCells,
		Events:                 events,
		SuspectKeysWithExistingCells: converge.suspectKeysWithExistingCells,
		SuspectKeys:                  converge.suspectKeys,
	}
}

type convergence struct {
	*SQLDB

	keysWithMissingCells         []*models.ActualLRPKeyWithSchedulingInfo
	suspectKeysWithExistingCells []*models.ActualLRPKey

	suspectKeysToRetireMutex sync.Mutex
	suspectKeysToRetire      []*models.ActualLRPKey

	suspectKeys []*models.ActualLRPKey

	keysMutex    sync.Mutex
	keysToRetire []*models.ActualLRPKey

	missingLRPKeysMutex sync.Mutex
	missingLRPKeys      []*models.ActualLRPKeyWithSchedulingInfo

	unstartedLRPKeysMutex sync.Mutex
	unstartedLRPKeys      []*models.ActualLRPKeyWithSchedulingInfo
}

func newConvergence(db *SQLDB) *convergence {
	return &convergence{
		SQLDB: db,
	}
}

// Adds stale UNCLAIMED Actual LRPs to the list of start requests.
func (c *convergence) staleUnclaimedActualLRPs(logger lager.Logger, now time.Time) {
	logger = logger.Session("stale-unclaimed-actual-lrps")

	rows, err := c.selectStaleUnclaimedLRPs(logger, c.db, now)
	if err != nil {
		logger.Error("failed-query", err)
		return
	}

	for rows.Next() {
		var index int
		schedulingInfo, err := c.fetchDesiredLRPSchedulingInfoAndMore(logger, rows, &index)
		if err == nil {
			c.addUnstartedLRPKey(logger, &models.ActualLRPKeyWithSchedulingInfo{
				Key:            &models.ActualLRPKey{schedulingInfo.ProcessGuid, int32(index), schedulingInfo.Domain},
				SchedulingInfo: schedulingInfo,
			})
			logger.Info("creating-start-request",
				lager.Data{"reason": "stale-unclaimed-lrp", "process_guid": schedulingInfo.ProcessGuid, "index": index})
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

	rows, err := c.selectCrashedLRPs(logger, c.db)
	if err != nil {
		logger.Error("failed-query", err)
		return
	}

	type crashedActualLRP struct {
		lrpKey         models.ActualLRPKey
		schedulingInfo *models.DesiredLRPSchedulingInfo
		index          int
	}
	lrps := []crashedActualLRP{}

	for rows.Next() {
		var index int
		actual := &models.ActualLRP{}

		schedulingInfo, err := c.fetchDesiredLRPSchedulingInfoAndMore(logger, rows, &index, &actual.Since, &actual.CrashCount)
		if err != nil {
			continue
		}

		actual.ActualLRPKey = models.NewActualLRPKey(schedulingInfo.ProcessGuid, int32(index), schedulingInfo.Domain)
		actual.State = models.ActualLRPStateCrashed

		if actual.ShouldRestartCrash(now, restartCalculator) {
			lrps = append(lrps, crashedActualLRP{
				lrpKey:         actual.ActualLRPKey,
				schedulingInfo: schedulingInfo,
				index:          index,
			})
			logger.Info("creating-start-request",
				lager.Data{"reason": "crashed-instance", "process_guid": actual.ProcessGuid, "index": index})
		}
	}

	for _, lrp := range lrps {
		key := lrp.lrpKey
		schedulingInfo := lrp.schedulingInfo

		c.addUnstartedLRPKey(logger, &models.ActualLRPKeyWithSchedulingInfo{
			Key:            &key,
			SchedulingInfo: schedulingInfo,
		})
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

	rows, err := c.selectOrphanedActualLRPs(logger, c.db)
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

func (c *convergence) extraSuspectActualLRPs(logger lager.Logger) {
	logger = logger.Session("extra-suspect-lrps")

	rows, err := c.selectExtraSuspectActualLRPs(logger, c.db)
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

		c.addExtraSuspectLRPKeyToRetire(logger, actualLRPKey)
	}

	if rows.Err() != nil {
		logger.Error("failed-getting-next-row", rows.Err())
	}
}

// Creates and adds missing Actual LRPs to the list of start requests.
// Adds extra Actual LRPs  to the list of keys to retire.
func (c *convergence) lrpInstanceCounts(logger lager.Logger, domainSet map[string]struct{}) {
	logger = logger.Session("lrp-instance-counts")

	rows, err := c.selectLRPInstanceCounts(logger, c.db)
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
		existingIndices := make(map[int]struct{})
		if existingIndicesStr.String != "" {
			for _, indexStr := range strings.Split(existingIndicesStr.String, ",") {
				index, err := strconv.Atoi(indexStr)
				if err != nil {
					logger.Error("cannot-parse-index", err, lager.Data{
						"index":                indexStr,
						"existing-indeces-str": existingIndicesStr,
					})
					return
				}
				existingIndices[index] = struct{}{}
			}
		}

		for i := 0; i < int(schedulingInfo.Instances); i++ {
			_, found := existingIndices[i]
			if found {
				continue
			}

			missingLRPCount++
			indices = append(indices, i)
			index := int32(i)
			c.addMissingLRPKey(logger, &models.ActualLRPKeyWithSchedulingInfo{
				Key: &models.ActualLRPKey{
					ProcessGuid: schedulingInfo.ProcessGuid,
					Domain:      schedulingInfo.Domain,
					Index:       index,
				},
				SchedulingInfo: schedulingInfo,
			})
			logger.Info("creating-start-request",
				lager.Data{"reason": "missing-instance", "process_guid": schedulingInfo.ProcessGuid, "index": index})
		}

		for index := range existingIndices {
			if index < int(schedulingInfo.Instances) {
				continue
			}

			// only take destructive actions for fresh domains
			if _, ok := domainSet[schedulingInfo.Domain]; ok {
				c.addKeyToRetire(logger, &models.ActualLRPKey{
					ProcessGuid: schedulingInfo.ProcessGuid,
					Index:       int32(index),
					Domain:      schedulingInfo.Domain,
				})
			}
		}
	}

	if rows.Err() != nil {
		logger.Error("failed-getting-next-row", rows.Err())
	}

	c.metronClient.SendMetric(missingLRPsMetric, missingLRPCount)
}

func (c *convergence) suspectActualLRPs(logger lager.Logger) {
	logger = logger.Session("suspect-lrps")

	rows, err := c.selectSuspectActualLRPs(logger, c.db)
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

		c.suspectKeys = append(c.suspectKeys, actualLRPKey)
	}

	if rows.Err() != nil {
		logger.Error("failed-getting-next-row", rows.Err())
	}
}

// Unclaim Actual LRPs that have missing cells (not in the cell set passed to
// convergence) and add them to the list of start requests.
func (c *convergence) suspectActualLRPsWithExistingCells(logger lager.Logger, cellSet models.CellSet) {
	logger = logger.Session("suspect-lrps-with-existing-cells")

	if len(cellSet) == 0 {
		return
	}

	rows, err := c.selectSuspectLRPsWithExistingCells(logger, c.db, cellSet)
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

		c.suspectKeysWithExistingCells = append(c.suspectKeysWithExistingCells, actualLRPKey)
	}

	if rows.Err() != nil {
		logger.Error("failed-getting-next-row", rows.Err())
	}
}

// Unclaim Actual LRPs that have missing cells (not in the cell set passed to
// convergence) and add them to the list of start requests.
func (c *convergence) actualLRPsWithMissingCells(logger lager.Logger, cellSet models.CellSet) {
	logger = logger.Session("actual-lrps-with-missing-cells")

	var keysWithMissingCells []*models.ActualLRPKeyWithSchedulingInfo

	rows, err := c.selectLRPsWithMissingCells(logger, c.db, cellSet)
	if err != nil {
		logger.Error("failed-query", err)
		return
	}

	missingCellSet := make(map[string]struct{})
	for rows.Next() {
		var index int32
		var cellID string
		schedulingInfo, err := c.fetchDesiredLRPSchedulingInfoAndMore(logger, rows, &index, &cellID)
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
		missingCellSet[cellID] = struct{}{}
	}

	if rows.Err() != nil {
		logger.Error("failed-getting-next-row", rows.Err())
	}

	cellIDs := []string{}
	for key, _ := range missingCellSet {
		cellIDs = append(cellIDs, key)
	}

	if len(cellIDs) > 0 {
		logger.Info("detected-missing-cells", lager.Data{"cell_ids": cellIDs})
	}

	c.keysWithMissingCells = keysWithMissingCells
}

func (c *convergence) addUnstartedLRPKey(logger lager.Logger, key *models.ActualLRPKeyWithSchedulingInfo) {
	c.unstartedLRPKeysMutex.Lock()
	defer c.unstartedLRPKeysMutex.Unlock()

	c.unstartedLRPKeys = append(c.unstartedLRPKeys, key)
}

func (c *convergence) addMissingLRPKey(logger lager.Logger, key *models.ActualLRPKeyWithSchedulingInfo) {
	c.missingLRPKeysMutex.Lock()
	defer c.missingLRPKeysMutex.Unlock()

	c.missingLRPKeys = append(c.missingLRPKeys, key)
}

func (c *convergence) addExtraSuspectLRPKeyToRetire(logger lager.Logger, key *models.ActualLRPKey) {
	c.suspectKeysToRetireMutex.Lock()
	defer c.suspectKeysToRetireMutex.Unlock()

	c.suspectKeysToRetire = append(c.suspectKeysToRetire, key)
}

func (c *convergence) addKeyToRetire(logger lager.Logger, key *models.ActualLRPKey) {
	c.keysMutex.Lock()
	defer c.keysMutex.Unlock()

	c.keysToRetire = append(c.keysToRetire, key)
}

func (c *convergence) result(logger lager.Logger) []*models.ActualLRPKeyWithSchedulingInfo {
	c.metronClient.SendMetric(extraLRPsMetric, len(c.keysToRetire))
	c.emitLRPMetrics(logger)

	return c.missingLRPKeys
}

func (db *SQLDB) pruneDomains(logger lager.Logger, now time.Time) {
	logger = logger.Session("prune-domains")

	_, err := db.delete(logger, db.db, domainsTable, "expire_time <= ?", now.UnixNano())
	if err != nil {
		logger.Error("failed-query", err)
	}
}

func (db *SQLDB) pruneEvacuatingActualLRPs(logger lager.Logger, cellSet models.CellSet) []models.Event {
	logger = logger.Session("prune-evacuating-actual-lrps")

	wheres := []string{"presence = ?"}
	bindings := []interface{}{models.ActualLRP_Evacuating}

	if len(cellSet) > 0 {
		wheres = append(wheres, fmt.Sprintf("actual_lrps.cell_id NOT IN (%s)", helpers.QuestionMarks(len(cellSet))))

		for cellID := range cellSet {
			bindings = append(bindings, cellID)
		}
	}

	lrpsToDelete, err := db.getActualLRPs(logger, strings.Join(wheres, " AND "), bindings...)
	if err != nil {
		logger.Error("failed-fetching-evacuating-lrps-with-missing-cells", err)
	}

	_, err = db.delete(logger, db.db, actualLRPsTable, strings.Join(wheres, " AND "), bindings...)
	if err != nil {
		logger.Error("failed-query", err)
	}

	var events []models.Event
	for _, lrp := range lrpsToDelete {
		events = append(events, models.NewActualLRPRemovedEvent(lrp.ToActualLRPGroup()))
	}
	return events
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
		db.metronClient.SendMetric(domainMetricPrefix+domain, 1)
	}
}

func (db *SQLDB) emitLRPMetrics(logger lager.Logger) {
	var err error
	logger = logger.Session("emit-lrp-metrics")
	claimedInstances, unclaimedInstances, runningInstances, crashedInstances, crashingDesireds := db.countActualLRPsByState(logger, db.db)

	desiredInstances := db.countDesiredInstances(logger, db.db)

	err = db.metronClient.SendMetric(unclaimedLRPsMetric, unclaimedInstances)
	if err != nil {
		logger.Error("failed-sending-unclaimed-lrps-metric", err)
	}

	err = db.metronClient.SendMetric(claimedLRPsMetric, claimedInstances)
	if err != nil {
		logger.Error("failed-sending-claimed-lrps-metric", err)
	}

	err = db.metronClient.SendMetric(runningLRPsMetric, runningInstances)
	if err != nil {
		logger.Error("failed-sending-running-lrps-metric", err)
	}

	err = db.metronClient.SendMetric(crashedActualLRPsMetric, crashedInstances)
	if err != nil {
		logger.Error("failed-sending-crashed-actual-lrps-metric", err)
	}

	err = db.metronClient.SendMetric(crashingDesiredLRPsMetric, crashingDesireds)
	if err != nil {
		logger.Error("failed-sending-crashing-desired-lrps-metric", err)
	}

	err = db.metronClient.SendMetric(instanceLRPsMetric, desiredInstances)
	if err != nil {
		logger.Error("failed-sending-desired-lrps-metric", err)
	}
}
