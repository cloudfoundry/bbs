package sqldb

import (
	"sync"
	"sync/atomic"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry/gunk/workpool"
	"github.com/nu7hatch/gouuid"
	"github.com/pivotal-golang/lager"
)

type actualInstance struct {
	blob         string
	isEvacuating bool
}

func (db *SQLDB) ActualLRPGroups(logger lager.Logger, filter models.ActualLRPFilter) ([]*models.ActualLRPGroup, error) {
	filterString := ""
	if filter.Domain != "" {
		filterString += " where domain = '" + filter.Domain + "'"
	}
	if filter.CellID != "" {
		if len(filterString) == 0 {
			filterString += " where "
		} else {
			filterString += " and "
		}
		filterString += "cellID = '" + filter.CellID + "'"
	}

	instQuery := "select processGuid, data, isEvacuating from actuals" + filterString
	logger.Info("actuallrp-groups-query", lager.Data{"query": instQuery})
	rows, err := db.sql.Query(instQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	actualsByProcessGuid := make(map[string][]actualInstance)
	var data, pGuid string
	var isEvac bool
	for rows.Next() {
		if err := rows.Scan(&pGuid, &data, &isEvac); err != nil {
			logger.Error("actual-lrp-groups", err)
			panic(err)
		}
		actualsByProcessGuid[pGuid] = append(actualsByProcessGuid[pGuid],
			actualInstance{data, isEvac})
	}

	var workErr atomic.Value
	groupChan := make(chan []*models.ActualLRPGroup, len(actualsByProcessGuid))
	wg := sync.WaitGroup{}

	logger.Debug("performing-deserialization-work")
	for _, actualSlice := range actualsByProcessGuid {
		actualSlice := actualSlice

		wg.Add(1)
		go func() {
			defer wg.Done()
			g, err := db.parseActualLRPGroups(logger, actualSlice)
			if err != nil {
				workErr.Store(err)
				return
			}
			groupChan <- g
		}()
	}

	go func() {
		wg.Wait()
		close(groupChan)
	}()
	groups := []*models.ActualLRPGroup{}

	for g := range groupChan {
		groups = append(groups, g...)
	}

	if err, ok := workErr.Load().(error); ok {
		logger.Error("failed-performing-deserialization-work", err)
		return []*models.ActualLRPGroup{}, models.ErrUnknownError
	}
	logger.Debug("succeeded-performing-deserialization-work", lager.Data{"num-actual-lrp-groups": len(groups)})

	return groups, nil
}

func (db *SQLDB) parseActualLRPGroups(logger lager.Logger, actualSlice []actualInstance) ([]*models.ActualLRPGroup, error) {
	var groups = []*models.ActualLRPGroup{}

	logger.Debug("performing-parsing-actual-lrp-groups")
	group := &models.ActualLRPGroup{}
	for _, aInstance := range actualSlice { // instances/evacs
		var lrp models.ActualLRP
		deserializeErr := db.deserializeModel(logger, aInstance.blob, &lrp)
		if deserializeErr != nil {
			panic(deserializeErr)
			return []*models.ActualLRPGroup{}, deserializeErr
		}

		if !aInstance.isEvacuating {
			group.Instance = &lrp
		}

		if aInstance.isEvacuating {
			group.Evacuating = &lrp
		}
	}

	if group.Instance != nil || group.Evacuating != nil {
		groups = append(groups, group)
	}
	logger.Debug("succeeded-performing-parsing-actual-lrp-groups", lager.Data{"num-actual-lrp-groups": len(groups)})

	return groups, nil
}

func (db *SQLDB) ActualLRPGroupsByProcessGuid(logger lager.Logger, processGuid string) ([]*models.ActualLRPGroup, error) {
	query := "select processGuid, data, isEvacuating from actuals where processGuid = ?"
	rows, err := db.sql.Query(query, processGuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	actuals := []actualInstance{}
	var data, pGuid string
	var isEvac bool
	for rows.Next() {
		if err := rows.Scan(&pGuid, &data, &isEvac); err != nil {
			logger.Error("actuallrp-groups-by-pg", err)
			panic(err)
		}
		actuals = append(actuals, actualInstance{data, isEvac})
	}

	if len(actuals) == 0 {
		return []*models.ActualLRPGroup{}, nil
	}

	return db.parseActualLRPGroups(logger, actuals)
}

// func (db *SQLDB) instanceActualLRPsByProcessGuid(logger lager.Logger, processGuid string) (map[int32]*models.ActualLRP, error) {
// 	node, err := db.fetchRecursiveRaw(logger, ActualLRPProcessDir(processGuid))
// 	bbsErr := models.ConvertError(err)
// 	if bbsErr != nil {
// 		if bbsErr.Type == models.Error_ResourceNotFound {
// 			return nil, nil
// 		}
// 		return nil, err
// 	}
// 	if node.Nodes.Len() == 0 {
// 		return nil, nil
// 	}

// 	var instances = map[int32]*models.ActualLRP{}

// 	logger.Debug("performing-parsing-actual-lrps")
// 	for _, indexNode := range node.Nodes {
// 		for _, instanceNode := range indexNode.Nodes {
// 			if !isInstanceActualLRPNode(instanceNode) {
// 				continue
// 			}

// 			instance := &models.ActualLRP{}
// 			deserializeErr := db.deserializeModel(logger, instanceNode, instance)
// 			if deserializeErr != nil {
// 				logger.Error("failed-parsing-actual-lrs", deserializeErr, lager.Data{"key": instanceNode.Key})
// 				return nil, deserializeErr
// 			}

// 			instances[instance.Index] = instance
// 		}

// 	}
// 	logger.Debug("succeeded-performing-parsing-actual-lrps", lager.Data{"num-actual-lrps": len(instances)})

// 	return instances, nil
// }

func (db *SQLDB) ActualLRPGroupByProcessGuidAndIndex(logger lager.Logger, processGuid string, index int32) (*models.ActualLRPGroup, error) {
	group, _, err := db.rawActualLRPGroupByProcessGuidAndIndex(logger, processGuid, index)
	return group, err
}

func (db *SQLDB) rawActualLRPGroupByProcessGuidAndIndex(logger lager.Logger, processGuid string, index int32) (*models.ActualLRPGroup, uint64, error) {

	instQuery := "select processGuid, data, isEvacuating from actuals where processGuid = ? and idx = ?"
	rows, err := db.sql.Query(instQuery, processGuid, index)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	actualsByProcessGuid := []actualInstance{}
	var data, pGuid string
	var isEvac bool
	for rows.Next() {
		if err := rows.Scan(&pGuid, &data, &isEvac); err != nil {
			logger.Error("rawActualLRPGroupByProcessGuidAndIndex", err)
			panic(err)
		}
		actualsByProcessGuid = append(actualsByProcessGuid,
			actualInstance{data, isEvac})
	}

	group := models.ActualLRPGroup{}
	for _, instanceNode := range actualsByProcessGuid {
		var lrp models.ActualLRP
		deserializeErr := db.deserializeModel(logger, instanceNode.blob, &lrp)
		if deserializeErr != nil {
			return nil, 0, deserializeErr
		}

		if !instanceNode.isEvacuating {
			group.Instance = &lrp
		}

		if instanceNode.isEvacuating {
			group.Evacuating = &lrp
		}
	}

	if group.Evacuating == nil && group.Instance == nil {
		return nil, 0, models.ErrResourceNotFound
	}

	return &group, 0, nil
}

func (db *SQLDB) rawActuaLLRPByProcessGuidAndIndex(logger lager.Logger, processGuid string, index int32) (*models.ActualLRP, uint64, error) {
	logger.Debug("raw-actual-lrp-by-process-guid-and-index")
	instQuery := "select data from actuals where processGuid = ? and idx = ? and isEvacuating = false"
	row := db.sql.QueryRow(instQuery, processGuid, index)

	var data string
	lrp := new(models.ActualLRP)
	if err := row.Scan(&data); err != nil {
		logger.Error("failed-to-get-data", err, lager.Data{"processGuid": processGuid, "index": index})
		return nil, 0, err
	}

	deserializeErr := db.deserializeModel(logger, data, lrp)
	if deserializeErr != nil {
		return nil, 0, deserializeErr
	}

	return lrp, 0, nil
}

func (db *SQLDB) ClaimActualLRP(logger lager.Logger, processGuid string, index int32, instanceKey *models.ActualLRPInstanceKey) error {
	logger = logger.Session("claim-actual-lrp", lager.Data{"process_guid": processGuid, "index": index, "actual_lrp_instance-key": instanceKey})
	logger.Info("starting")

	lrp, _, err := db.rawActuaLLRPByProcessGuidAndIndex(logger, processGuid, index)
	if err != nil {
		logger.Error("failed", err)
		return err
	}

	if !lrp.AllowsTransitionTo(&lrp.ActualLRPKey, instanceKey, models.ActualLRPStateClaimed) {
		return models.ErrActualLRPCannotBeClaimed
	}

	lrp.PlacementError = ""
	lrp.State = models.ActualLRPStateClaimed
	lrp.ActualLRPInstanceKey = *instanceKey
	lrp.ActualLRPNetInfo = models.ActualLRPNetInfo{}
	lrp.ModificationTag.Increment()

	err = lrp.Validate()
	if err != nil {
		logger.Error("failed", err)
		return models.NewError(models.Error_InvalidRecord, err.Error())
	}

	lrpData, serializeErr := db.serializeModel(logger, lrp)
	if serializeErr != nil {
		return serializeErr
	}

	update := "update actuals set processGuid=?, idx=?, cellId=?, domain=?, data=?  where processGuid = ? and idx = ? and isEvacuating=false"
	_, err = db.sql.Exec(update, lrp.ProcessGuid, lrp.Index, lrp.CellId, lrp.Domain, lrpData, lrp.ProcessGuid, lrp.Index)
	if err != nil {
		logger.Error("update-failed", err)
		return models.ErrActualLRPCannotBeClaimed
	}
	logger.Info("succeeded")

	return nil
}

func (db *SQLDB) StartActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, netInfo *models.ActualLRPNetInfo) error {
	logger = logger.Session("start-actual-lrp", lager.Data{"actual_lrp_key": key, "actual_lrp_instance_key": instanceKey, "net_info": netInfo})
	logger.Info("starting")
	lrp, _, err := db.rawActuaLLRPByProcessGuidAndIndex(logger, key.ProcessGuid, key.Index)
	bbsErr := models.ConvertError(err)
	if bbsErr != nil {
		logger.Error("failed-to-get-actual-lrp", err)
		return err
	}

	if lrp.ActualLRPKey.Equal(key) &&
		lrp.ActualLRPInstanceKey.Equal(instanceKey) &&
		lrp.ActualLRPNetInfo.Equal(netInfo) &&
		lrp.State == models.ActualLRPStateRunning {
		logger.Info("succeeded")
		return nil
	}

	if !lrp.AllowsTransitionTo(&lrp.ActualLRPKey, instanceKey, models.ActualLRPStateRunning) {
		logger.Error("failed-to-transition-actual-lrp-to-started", nil)
		return models.ErrActualLRPCannotBeStarted
	}

	lrp.ModificationTag.Increment()
	lrp.State = models.ActualLRPStateRunning
	lrp.Since = db.clock.Now().UnixNano()
	lrp.ActualLRPInstanceKey = *instanceKey
	lrp.ActualLRPNetInfo = *netInfo
	lrp.PlacementError = ""

	err = lrp.Validate()
	if err != nil {
		logger.Error("failed", err)
		return models.NewError(models.Error_InvalidRecord, err.Error())
	}

	lrpData, serializeErr := db.serializeModel(logger, lrp)
	if serializeErr != nil {
		return serializeErr
	}

	update := "update actuals set processGuid=?, idx=?, cellId=?, domain=?, data=?  where processGuid = ? and idx = ? and isEvacuating=false"
	_, err = db.sql.Exec(update, lrp.ProcessGuid, lrp.Index, lrp.CellId, lrp.Domain, lrpData, lrp.ProcessGuid, lrp.Index)
	if err != nil {
		logger.Error("update-failed", err)
		return models.ErrActualLRPCannotBeStarted
	}
	logger.Info("succeeded")

	return nil
}

func (db *SQLDB) CrashActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, errorMessage string) error {
	return db.etcdDB.CrashActualLRP(logger, key, instanceKey, errorMessage)
}

func (db *SQLDB) FailActualLRP(logger lager.Logger, key *models.ActualLRPKey, errorMessage string) error {
	return db.etcdDB.FailActualLRP(logger, key, errorMessage)
}

func (db *SQLDB) RemoveActualLRP(logger lager.Logger, processGuid string, index int32) error {
	return db.etcdDB.RemoveActualLRP(logger, processGuid, index)
}

func (db *SQLDB) RetireActualLRP(logger lager.Logger, key *models.ActualLRPKey) error {
	return db.etcdDB.RetireActualLRP(logger, key)
}

// we need the create methods!!!
//
func (db *SQLDB) newUnclaimedActualLRP(key *models.ActualLRPKey) (*models.ActualLRP, error) {
	guid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	return &models.ActualLRP{
		ActualLRPKey: *key,
		Since:        db.clock.Now().UnixNano(),
		State:        models.ActualLRPStateUnclaimed,
		ModificationTag: models.ModificationTag{
			Epoch: guid.String(),
			Index: 0,
		},
	}, nil
}

func (db *SQLDB) createUnclaimedActualLRP(logger lager.Logger, key *models.ActualLRPKey) error {
	lrp, err := db.newUnclaimedActualLRP(key)
	if err != nil {
		return models.ErrActualLRPCannotBeUnclaimed
	}

	return db.createRawActualLRP(logger, lrp)
}

func (db *SQLDB) createRawActualLRP(logger lager.Logger, lrp *models.ActualLRP) error {
	logger = logger.Session("creating-raw-actual-lrp", lager.Data{"actual-lrp": lrp})
	logger.Debug("starting")
	defer logger.Debug("complete")

	lrpData, err := db.serializeModel(logger, lrp)
	if err != nil {
		logger.Error("failed-to-marshal-actual-lrp", err, lager.Data{"actual-lrp": lrp})
		return err
	}

	insert := "insert into actuals (processGuid, idx, cellId, domain, data, isEvacuating) values (?, ?, ?, ?, ?, ?)"
	_, err = db.sql.Exec(insert, lrp.ProcessGuid, lrp.Index, lrp.CellId, lrp.Domain, lrpData, false)
	if err != nil {
		logger.Error("failed-to-create-actual-lrp", err)
		return models.ErrActualLRPCannotBeStarted
	}
	return nil
}

func (db *SQLDB) createUnclaimedActualLRPs(logger lager.Logger, keys []*models.ActualLRPKey) []int {
	count := len(keys)
	createdIndicesChan := make(chan int, count)

	works := make([]func(), count)

	for i, key := range keys {
		key := key
		works[i] = func() {
			err := db.createUnclaimedActualLRP(logger, key)
			if err != nil {
				logger.Info("failed-creating-actual-lrp", lager.Data{"actual_lrp_key": key, "err-message": err.Error()})
			} else {
				createdIndicesChan <- int(key.Index)
			}
		}
	}

	throttler, err := workpool.NewThrottler(db.updateWorkersSize, works)
	if err != nil {
		logger.Error("failed-constructing-throttler", err, lager.Data{"max-workers": db.updateWorkersSize, "num-works": len(works)})
		return []int{}
	}

	go func() {
		throttler.Work()
		close(createdIndicesChan)
	}()

	createdIndices := make([]int, 0, count)
	for createdIndex := range createdIndicesChan {
		createdIndices = append(createdIndices, createdIndex)
	}

	return createdIndices
}
