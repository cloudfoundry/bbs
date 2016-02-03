package sqldb

import (
	"log"
	"sync"
	"sync/atomic"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

type actualInstance struct {
	blob         string
	isEvacuating bool
}

func (db *SQLDB) ActualLRPGroups(logger lager.Logger, filter models.ActualLRPFilter) ([]*models.ActualLRPGroup, error) {
	filterString := " "
	if filter.Domain != "" {
		filterString += "domain = " + filter.Domain + " "
	}
	if filter.CellID != "" {
		filterString += "cellId = " + filter.CellID
	}

	instQuery := "select processGuid, data, isEvacuating from actuals" + filterString
	rows, err := db.sql.Query(instQuery)
	if err != nil {
		return nil, err
	}

	actualsByProcessGuid := make(map[string][]actualInstance)
	var data, pGuid string
	var isEvac bool
	for rows.Next() {
		if err := rows.Scan(&pGuid, &data, &isEvac); err != nil {
			log.Fatal(err)
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
	query := "select data, isEvacuating from actuals where processGuid = $1"
	rows, err := db.sql.Query(query, processGuid)
	if err != nil {
		return nil, err
	}

	actuals := []actualInstance{}
	var data, pGuid string
	var isEvac bool
	for rows.Next() {
		if err := rows.Scan(&pGuid, &data, &isEvac); err != nil {
			log.Fatal(err)
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

	instQuery := "select processGuid, data, isEvacuating, modifiedIndex from actuals where processGuid = $1 and inx = $2"
	rows, err := db.sql.Query(instQuery, processGuid, index)
	if err != nil {
		return nil, 0, err
	}

	actualsByProcessGuid := []actualInstance{}
	var data, pGuid string
	var modifiedIndex int
	var isEvac bool
	for rows.Next() {
		if err := rows.Scan(&pGuid, &data, &isEvac); err != nil {
			log.Fatal(err)
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

	return &group, uint64(modifiedIndex), nil
}

func (db *SQLDB) rawActuaLLRPByProcessGuidAndIndex(logger lager.Logger, processGuid string, index int32) (*models.ActualLRP, uint64, error) {
	logger.Debug("raw-actual-lrp-by-process-guid-and-index")
	instQuery := "select data, modifiedIndex from actuals where processGuid = $1 and idx = $2 and isEvacuating = false"
	row := db.sql.QueryRow(instQuery, processGuid, index)

	var data string
	var modifiedIndex int
	lrp := new(models.ActualLRP)
	if err := row.Scan(&data, &modifiedIndex); err != nil {
		log.Fatal(err)
	}

	deserializeErr := db.deserializeModel(logger, data, lrp)
	if deserializeErr != nil {
		return nil, 0, deserializeErr
	}

	return lrp, uint64(modifiedIndex), nil
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

	insert := "insert into actuals (processGuid, idx, cellId, data, isEvacuating) values ($1, $2, $3, $4, false)"
	_, err = db.sql.Exec(insert, lrp.ProcessGuid, lrp.Index, lrp.CellId, lrpData)
	if err != nil {
		logger.Error("insert-failed", err)
		return models.ErrActualLRPCannotBeClaimed
	}
	logger.Info("succeeded")

	return nil
}

// stubbed out to etcd
func (db *SQLDB) StartActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, netInfo *models.ActualLRPNetInfo) error {
	return db.etcdDB.StartActualLRP(logger, key, instanceKey, netInfo)
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
