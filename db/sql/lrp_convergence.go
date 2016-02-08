package sqldb

import (
	"github.com/cloudfoundry-incubator/bbs/models"

	"github.com/pivotal-golang/lager"
)

type actualLRPCompositeKey struct {
	processGuid  string
	index        int32
	isEvacuating bool
}

type LRPMetricCounter struct {
}

func (db *SQLDB) GatherActualLRPs(logger lager.Logger, guids map[string]struct{}, lmc *LRPMetricCounter) (map[string]map[int32]*models.ActualLRP, error) {
	logger = logger.Session("gather-actual-lrps")
	logger.Info("starting")
	defer logger.Info("complete")

	return db.gatherAndOptionallyPruneActualLRPs(logger, guids, false, lmc)
}

func (db *SQLDB) gatherAndOptionallyPruneActualLRPs(logger lager.Logger, guids map[string]struct{}, doPrune bool, lmc *LRPMetricCounter) (map[string]map[int32]*models.ActualLRP, error) {
	fetchQuery := "select processGuid, idx, isEvacuating, data from actuals"
	logger.Info("built-fetch-query", lager.Data{"query": fetchQuery})
	rows, err := db.sql.Query(fetchQuery)
	if err != nil {
		logger.Error("failed fetching actual lrps", err)
		return nil, err
	}
	defer rows.Close()

	actualLRPs := map[string]map[int32]*models.ActualLRP{}
	actualsToDelete := []actualLRPCompositeKey{}

	var processGuid, data string
	var isEvacuating bool
	var index int32

	logger.Info("scanning-actual-lrps")
	for rows.Next() {
		err := rows.Scan(&processGuid, &index, &isEvacuating, &data)
		if err != nil {
			logger.Error("failed to scan row", err)
			return nil, err
		}

		var actualLRP models.ActualLRP
		deserializeErr := db.deserializeModel(logger, data, &actualLRP)
		if deserializeErr != nil {
			key := actualLRPCompositeKey{
				processGuid:  processGuid,
				index:        index,
				isEvacuating: isEvacuating,
			}
			actualsToDelete = append(actualsToDelete, key)
		}

		indexMap, ok := actualLRPs[processGuid]
		if !ok {
			indexMap = map[int32]*models.ActualLRP{}
		}

		indexMap[index] = &actualLRP
		actualLRPs[processGuid] = indexMap
	}

	logger.Info("gathered-actual-lrps", lager.Data{"num": len(actualLRPs)})
	return actualLRPs, nil
}

func (db *SQLDB) GatherAndPruneDesiredLRPs(logger lager.Logger, guids map[string]struct{}, lmc *LRPMetricCounter) (map[string]*models.DesiredLRP, error) {
	logger = logger.Session("gathering-desired-lrps")

	logger.Info("starting")
	defer logger.Info("complete")

	fetchQuery := "select processGuid, scheduleInfo, runInfo from desired"
	logger.Info("built-fetch-query", lager.Data{"query": fetchQuery})
	rows, err := db.sql.Query(fetchQuery)
	if err != nil {
		logger.Error("failed fetching desired lrps", err)
		return nil, err
	}
	defer rows.Close()

	desiredLRPs := map[string]*models.DesiredLRP{}
	var processGuid, schedulingInfo, runInfo string

	logger.Info("scanning-desired-lrps")
	for rows.Next() {
		err := rows.Scan(&processGuid, &schedulingInfo, &runInfo)
		if err != nil {
			logger.Error("failed to scan row", err)
			return nil, err
		}

		var desiredSchedulingInfo models.DesiredLRPSchedulingInfo
		deserializeErr := db.deserializeModel(logger, schedulingInfo, &desiredSchedulingInfo)
		if deserializeErr != nil {
			logger.Error("failed to deserialize scheduling info", err)
			continue
		}

		var desiredRunInfo models.DesiredLRPRunInfo
		deserializeErr = db.deserializeModel(logger, runInfo, &desiredRunInfo)
		if deserializeErr != nil {
			logger.Error("failed to deserialize run info", err)
			continue
		}

		desiredLRP := models.NewDesiredLRP(desiredSchedulingInfo, desiredRunInfo)
		desiredLRPs[processGuid] = &desiredLRP
	}

	logger.Info("gathered-desired-lrps", lager.Data{"num": len(desiredLRPs)})
	return desiredLRPs, nil
}
