package sqldb

import (
	"github.com/cloudfoundry-incubator/auctioneer"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

type guidSet struct {
	set map[string]struct{}
}

func newGuidSet() guidSet {
	return guidSet{
		set: map[string]struct{}{},
	}
}

func (g guidSet) Add(guid string) {
	g.set[guid] = struct{}{}
}

func (g guidSet) Merge(other guidSet) {
	for guid := range other.set {
		g.set[guid] = struct{}{}
	}
}

func (g guidSet) ToMap() map[string]struct{} {
	return g.set
}

func (db *SQLDB) DesiredLRPs(logger lager.Logger, filter models.DesiredLRPFilter) ([]*models.DesiredLRP, error) {
	logger = logger.Session("desired-lrps", lager.Data{"filter": filter})
	logger.Info("start")
	defer logger.Info("complete")
	return nil, nil

	// desireds, _, err := db.desiredLRPs(logger, filter)
	// if err != nil {
	// 	logger.Error("failed", err)
	// }
	// return desireds, err
}

func (db *SQLDB) DesiredLRPSchedulingInfos(logger lager.Logger, filter models.DesiredLRPFilter) ([]*models.DesiredLRPSchedulingInfo, error) {
	logger = logger.Session("desired-lrp-scheduling-infos", lager.Data{"filter": filter})
	logger.Info("start")
	defer logger.Info("complete")

	filterString := " "
	if filter.Domain != "" {
		filterString += "where domain = " + filter.Domain + " "
	}

	lrpQuery := "select processGuid, scheduleInfo from desired" + filterString
	rows, err := db.sql.Query(lrpQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var pGuid, sInfo string

	schedInfoByProcessGuid := make(map[string]string)
	for rows.Next() {
		if err := rows.Scan(&pGuid, &sInfo); err != nil {
			logger.Error("DesireLRPSchedGroup", err)
			panic(err)
		}
		schedInfoByProcessGuid[pGuid] = sInfo
	}

	schedulingInfoMap, _ := db.deserializeScheduleInfos(logger, schedInfoByProcessGuid)

	schedulingInfos := make([]*models.DesiredLRPSchedulingInfo, 0, len(schedulingInfoMap))
	for _, schedulingInfo := range schedulingInfoMap {
		schedulingInfos = append(schedulingInfos, schedulingInfo)
	}
	return schedulingInfos, nil
}

// func (db *SQLDB) desiredLRPs(logger lager.Logger, filter models.DesiredLRPFilter) ([]*models.DesiredLRP, guidSet, error) {
// 	root, err := db.fetchRecursiveRaw(logger, DesiredLRPComponentsSchemaRoot)
// 	bbsErr := models.ConvertError(err)
// 	if bbsErr != nil {
// 		if bbsErr.Type == models.Error_ResourceNotFound {
// 			return []*models.DesiredLRP{}, newGuidSet(), nil
// 		}
// 		return nil, newGuidSet(), err
// 	}
// 	if root.Nodes.Len() == 0 {
// 		return []*models.DesiredLRP{}, newGuidSet(), nil
// 	}

// 	var schedules map[string]*models.DesiredLRPSchedulingInfo
// 	var runs map[string]*models.DesiredLRPRunInfo
// 	var malformedInfos guidSet
// 	var malformedRunInfos guidSet
// 	var wg sync.WaitGroup
// 	for i := range root.Nodes {
// 		node := root.Nodes[i]
// 		switch node.Key {
// 		case DesiredLRPSchedulingInfoSchemaRoot:
// 			wg.Add(1)
// 			go func() {
// 				defer wg.Done()
// 				schedules, malformedInfos = db.deserializeScheduleInfos(logger, node.Nodes, filter)
// 			}()
// 		case DesiredLRPRunInfoSchemaRoot:
// 			wg.Add(1)
// 			go func() {
// 				defer wg.Done()
// 				runs, malformedRunInfos = db.deserializeRunInfos(logger, node.Nodes, filter)
// 			}()
// 		default:
// 			logger.Error("unexpected-etcd-key", nil, lager.Data{"key": node.Key})
// 		}
// 	}

// 	wg.Wait()

// 	desiredLRPs := []*models.DesiredLRP{}
// 	for processGuid, schedule := range schedules {
// 		desired := models.NewDesiredLRP(*schedule, *runs[processGuid])
// 		desiredLRPs = append(desiredLRPs, &desired)
// 	}

// 	malformedInfos.Merge(malformedRunInfos)
// 	return desiredLRPs, malformedInfos, nil
// }

func (db *SQLDB) deserializeScheduleInfos(logger lager.Logger, sInfoByPGuid map[string]string) (map[string]*models.DesiredLRPSchedulingInfo, guidSet) {
	logger.Info("deserializing-scheduling-infos", lager.Data{"count": len(sInfoByPGuid)})

	components := make(map[string]*models.DesiredLRPSchedulingInfo)
	malformedModels := newGuidSet()

	for guid, sInfo := range sInfoByPGuid {
		model := new(models.DesiredLRPSchedulingInfo)
		err := db.deserializeModel(logger, sInfo, model)
		if err != nil {
			logger.Error("failed-parsing-desired-lrp-scheduling-info", err)
			malformedModels.Add(guid)
			continue
		}
		components[guid] = model
	}

	return components, malformedModels
}

// func (db *SQLDB) deserializeRunInfos(logger lager.Logger, nodes etcd.Nodes, filter models.DesiredLRPFilter) (map[string]*models.DesiredLRPRunInfo, guidSet) {
// 	logger.Info("deserializing-run-infos", lager.Data{"count": len(nodes)})

// 	components := make(map[string]*models.DesiredLRPRunInfo, len(nodes))
// 	malformedModels := newGuidSet()

// 	for i := range nodes {
// 		node := nodes[i]
// 		model := new(models.DesiredLRPRunInfo)
// 		err := db.deserializeModel(logger, node, model)
// 		if err != nil {
// 			logger.Error("failed-parsing-desired-lrp-run-info", err)
// 			malformedModels.Add(model.ProcessGuid)
// 			continue
// 		}
// 		if filter.Domain == "" || model.Domain == filter.Domain {
// 			components[model.ProcessGuid] = model
// 		}
// 	}

// 	return components, malformedModels
// }

// func (db *SQLDB) rawDesiredLRPSchedulingInfo(logger lager.Logger, processGuid string) (*models.DesiredLRPSchedulingInfo, uint64, error) {
// 	node, err := db.fetchRaw(logger, DesiredLRPSchedulingInfoSchemaPath(processGuid))
// 	if err != nil {
// 		logger.Error("failed-to-fetch-existing-scheduling-info", err)
// 		return nil, 0, err
// 	}

// 	model := new(models.DesiredLRPSchedulingInfo)
// 	err = db.deserializeModel(logger, node, model)
// 	if err != nil {
// 		logger.Error("failed-parsing-desired-lrp-scheduling-info", err)
// 		return nil, 0, err
// 	}

// 	return model, node.ModifiedIndex, nil
// }

// func (db *SQLDB) rawDesiredLRPRunInfo(logger lager.Logger, processGuid string) (*models.DesiredLRPRunInfo, error) {
// 	node, err := db.fetchRaw(logger, DesiredLRPRunInfoSchemaPath(processGuid))
// 	if err != nil {
// 		return nil, err
// 	}

// 	model := new(models.DesiredLRPRunInfo)
// 	err = db.deserializeModel(logger, node, model)
// 	if err != nil {
// 		logger.Error("failed-parsing-desired-lrp-run-info", err)
// 		return nil, err
// 	}

// 	return model, nil
// }

// func (db *SQLDB) rawDesiredLRPByProcessGuid(logger lager.Logger, processGuid string) (*models.DesiredLRP, error) {
// 	var wg sync.WaitGroup

// 	var schedulingInfo *models.DesiredLRPSchedulingInfo
// 	var runInfo *models.DesiredLRPRunInfo
// 	var schedulingErr, runErr error

// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		schedulingInfo, _, schedulingErr = db.rawDesiredLRPSchedulingInfo(logger, processGuid)
// 	}()

// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		runInfo, runErr = db.rawDesiredLRPRunInfo(logger, processGuid)
// 	}()

// 	wg.Wait()

// 	if schedulingErr != nil {
// 		return nil, schedulingErr
// 	}

// 	if runErr != nil {
// 		return nil, runErr
// 	}

// 	desiredLRP := models.NewDesiredLRP(*schedulingInfo, *runInfo)
// 	return &desiredLRP, nil
// }

func (db *SQLDB) DesiredLRPByProcessGuid(logger lager.Logger, processGuid string) (*models.DesiredLRP, error) {
	return nil, nil
	// lrp, err := db.rawDesiredLRPByProcessGuid(logger, processGuid)
	// return lrp, err
}

func (db *SQLDB) startInstanceRange(logger lager.Logger, lower, upper int32, schedulingInfo *models.DesiredLRPSchedulingInfo) {
	logger = logger.Session("start-instance-range", lager.Data{"lower": lower, "upper": upper})
	logger.Info("starting")
	defer logger.Info("complete")

	keys := make([]*models.ActualLRPKey, upper-lower)
	i := 0
	for actualIndex := lower; actualIndex < upper; actualIndex++ {
		key := models.NewActualLRPKey(schedulingInfo.ProcessGuid, int32(actualIndex), schedulingInfo.Domain)
		keys[i] = &key
		i++
	}

	createdIndices := db.createUnclaimedActualLRPs(logger, keys)
	start := auctioneer.NewLRPStartRequestFromSchedulingInfo(schedulingInfo, createdIndices...)

	err := db.auctioneerClient.RequestLRPAuctions([]*auctioneer.LRPStartRequest{&start})
	if err != nil {
		logger.Error("failed-to-request-auction", err)
	}
}

// func (db *SQLDB) stopInstancesForProcessGuid(logger lager.Logger, processGuid string) {
// 	logger = logger.Session("stop-instance-for-process-guid", lager.Data{"process-guid": processGuid})
// 	logger.Info("starting")
// 	defer logger.Info("complete")

// 	actualsMap, err := db.instanceActualLRPsByProcessGuid(logger, processGuid)
// 	if err != nil {
// 		logger.Error("failed-to-get-actual-lrps", err)
// 		return
// 	}

// 	actualKeys := make([]*models.ActualLRPKey, len(actualsMap))
// 	for i, actual := range actualsMap {
// 		actualKeys[i] = &actual.ActualLRPKey
// 	}

// 	db.retireActualLRPs(logger, actualKeys)
// }

// func (db *SQLDB) stopInstanceRange(logger lager.Logger, lower, upper int32, schedInfo *models.DesiredLRPSchedulingInfo) {
// 	logger = logger.Session("stop-instance-range", lager.Data{"lower": lower, "upper": upper})
// 	logger.Info("starting")
// 	defer logger.Info("complete")

// 	actualsMap, err := db.instanceActualLRPsByProcessGuid(logger, schedInfo.ProcessGuid)
// 	if err != nil {
// 		logger.Error("failed-to-get-actual-lrps", err)
// 		return
// 	}

// 	actualKeys := make([]*models.ActualLRPKey, 0)
// 	for i := lower; i < upper; i++ {
// 		actual, ok := actualsMap[i]
// 		if ok {
// 			actualKeys = append(actualKeys, &actual.ActualLRPKey)
// 		}
// 	}

// 	db.retireActualLRPs(logger, actualKeys)
// }

// DesireLRP creates a DesiredLRPSchedulingInfo and a DesiredLRPRunInfo. In order
// to ensure that the complete model is available and there are no races in
// Desired Watches, DesiredLRPRunInfo is created before DesiredLRPSchedulingInfo.
func (db *SQLDB) DesireLRP(logger lager.Logger, desiredLRP *models.DesiredLRP) error {
	logger = logger.Session("create-desired-lrp", lager.Data{"process-guid": desiredLRP.ProcessGuid})
	logger.Info("starting")
	defer logger.Info("complete")

	schedulingInfo, runInfo := desiredLRP.CreateComponents(db.clock.Now())
	serializedSchedInfo, err := db.serializeModel(logger, &schedulingInfo)
	if err != nil {
		logger.Error("failed-to-serialize", err)
		return err
	}
	serializedRunInfo, err := db.serializeModel(logger, &runInfo)
	if err != nil {
		logger.Error("failed-to-serialize", err)
		return err
	}

	insert := "insert into desired (processGuid, domain, scheduleInfo, runInfo) values ($1, $2, $3, $4)"
	_, err = db.sql.Exec(insert, desiredLRP.ProcessGuid, desiredLRP.Domain, serializedSchedInfo, serializedRunInfo)
	if err != nil {
		logger.Error("insert-failed", err)
		return err
	}

	db.startInstanceRange(logger, 0, schedulingInfo.Instances, &schedulingInfo)
	return nil
}

// func (db *SQLDB) createDesiredLRPSchedulingInfo(logger lager.Logger, schedulingInfo *models.DesiredLRPSchedulingInfo) error {
// 	epochGuid, err := uuid.NewV4()
// 	if err != nil {
// 		logger.Error("failed-to-generate-epoch", err)
// 		return models.ErrUnknownError
// 	}

// 	schedulingInfo.ModificationTag = models.NewModificationTag(epochGuid.String(), 0)

// 	serializedSchedInfo, err := db.serializeModel(logger, schedulingInfo)
// 	if err != nil {
// 		logger.Error("failed-to-serialize", err)
// 		return err
// 	}

// 	logger.Info("persisting-scheduling-info")
// 	_, err = db.client.Create(DesiredLRPSchedulingInfoSchemaPath(schedulingInfo.ProcessGuid), serializedSchedInfo, NO_TTL)
// 	err = ErrorFromEtcdError(logger, err)
// 	if err != nil {
// 		logger.Error("failed-persisting-scheduling-info", err)
// 		return err
// 	}

// 	logger.Info("succeeded-persisting-scheduling-info")
// 	return nil
// }

// func (db *SQLDB) updateDesiredLRPSchedulingInfo(logger lager.Logger, schedulingInfo *models.DesiredLRPSchedulingInfo, index uint64) error {
// 	value, err := db.serializeModel(logger, schedulingInfo)
// 	if err != nil {
// 		logger.Error("failed-to-serialize-scheduling-info", err)
// 		return err
// 	}

// 	_, err = db.client.CompareAndSwap(DesiredLRPSchedulingInfoSchemaPath(schedulingInfo.ProcessGuid), value, NO_TTL, index)
// 	if err != nil {
// 		logger.Error("failed-to-CAS-scheduling-info", err)
// 		return ErrorFromEtcdError(logger, err)
// 	}

// 	return nil
// }

// func (db *SQLDB) createDesiredLRPRunInfo(logger lager.Logger, runInfo *models.DesiredLRPRunInfo) error {
// 	serializedRunInfo, err := db.serializeModel(logger, runInfo)
// 	if err != nil {
// 		logger.Error("failed-to-serialize", err)
// 		return err
// 	}

// 	logger.Debug("persisting-run-info")
// 	_, err = db.client.Create(DesiredLRPRunInfoSchemaPath(runInfo.ProcessGuid), serializedRunInfo, NO_TTL)
// 	if err != nil {
// 		return ErrorFromEtcdError(logger, err)
// 	}
// 	logger.Debug("succeeded-persisting-run-info")

// 	return nil
// }

func (db *SQLDB) UpdateDesiredLRP(logger lager.Logger, processGuid string, update *models.DesiredLRPUpdate) error {
	logger = logger.Session("update-desired-lrp", lager.Data{"process-guid": processGuid})
	logger.Info("starting")
	defer logger.Info("complete")

	// var schedulingInfo *models.DesiredLRPSchedulingInfo
	// var existingInstances int32
	// var err error

	// for i := 0; i < 2; i++ {
	// 	var index uint64

	// 	schedulingInfo, index, err = db.rawDesiredLRPSchedulingInfo(logger, processGuid)
	// 	if err != nil {
	// 		logger.Error("failed-to-fetch-scheduling-info", err)
	// 		break
	// 	}

	// 	existingInstances = schedulingInfo.Instances

	// 	schedulingInfo.ApplyUpdate(update)

	// 	err = db.updateDesiredLRPSchedulingInfo(logger, schedulingInfo, index)
	// 	if err != nil {
	// 		logger.Error("update-scheduling-info-failed", err)
	// 		modelErr := models.ConvertError(err)
	// 		if modelErr != models.ErrResourceConflict {
	// 			break
	// 		}
	// 		// Retry on CAS fail
	// 		continue
	// 	}

	// 	break
	// }

	// if err != nil {
	// 	return err
	// }

	// switch diff := schedulingInfo.Instances - existingInstances; {
	// case diff > 0:
	// 	db.startInstanceRange(logger, existingInstances, schedulingInfo.Instances, schedulingInfo)

	// case diff < 0:
	// 	db.stopInstanceRange(logger, schedulingInfo.Instances, existingInstances, schedulingInfo)

	// case diff == 0:
	// 	// this space intentionally left blank
	// }

	return nil
}

// RemoveDesiredLRP deletes the DesiredLRPSchedulingInfo and the DesiredLRPRunInfo
// from the database. We delete DesiredLRPSchedulingInfo first because the system
// uses it to determine wheter the lrp is present. In the event that only the
// RunInfo fails to delete, the orphaned DesiredLRPRunInfo will be garbage
// collected later by convergence.
func (db *SQLDB) RemoveDesiredLRP(logger lager.Logger, processGuid string) error {
	logger = logger.Session("remove-desired-lrp", lager.Data{"process-guid": processGuid})
	logger.Info("starting")
	defer logger.Info("complete")
	return nil

	// logger.Info("deleting-scheduling-info")
	// _, schedulingInfoErr := db.client.Delete(DesiredLRPSchedulingInfoSchemaPath(processGuid), true)
	// schedulingInfoErr = ErrorFromEtcdError(logger, schedulingInfoErr)
	// if schedulingInfoErr != nil && schedulingInfoErr != models.ErrResourceNotFound {
	// 	logger.Error("failed-deleting-scheduling-info", schedulingInfoErr)
	// 	return schedulingInfoErr
	// }

	// logger.Info("deleting-run-info")
	// _, runInfoErr := db.client.Delete(DesiredLRPRunInfoSchemaPath(processGuid), true)
	// runInfoErr = ErrorFromEtcdError(logger, runInfoErr)
	// if runInfoErr != nil && runInfoErr != models.ErrResourceNotFound {
	// 	logger.Error("failed-deleting-run-info", runInfoErr)
	// 	return runInfoErr
	// }

	// if schedulingInfoErr == models.ErrResourceNotFound && runInfoErr == models.ErrResourceNotFound {
	// 	// If neither component of the desired LRP exists, don't bother trying to delete running instances
	// 	return models.ErrResourceNotFound
	// }

	// db.stopInstancesForProcessGuid(logger, processGuid)
	// return nil
}
