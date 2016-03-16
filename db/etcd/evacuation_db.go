package etcd

import (
	"reflect"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

func (db *ETCDDB) RemoveEvacuatingActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey) error {
	processGuid := key.ProcessGuid
	index := key.Index

	logger = logger.Session("remove-evacuating", lager.Data{"process_guid": processGuid, "index": index})

	logger.Debug("starting")
	defer logger.Debug("complete")

	node, err := db.fetchRaw(logger, EvacuatingActualLRPSchemaPath(processGuid, index))
	bbsErr := models.ConvertError(err)
	if bbsErr != nil {
		if bbsErr.Type == models.Error_ResourceNotFound {
			logger.Debug("evacuating-actual-lrp-not-found")
			return nil
		}
		return bbsErr
	}

	lrp := models.ActualLRP{}
	err = db.deserializeModel(logger, node, &lrp)
	if err != nil {
		return err
	}

	if !lrp.ActualLRPKey.Equal(key) || !lrp.ActualLRPInstanceKey.Equal(instanceKey) {
		return models.ErrActualLRPCannotBeRemoved
	}

	_, err = db.client.CompareAndDelete(EvacuatingActualLRPSchemaPath(lrp.ProcessGuid, lrp.Index), node.ModifiedIndex)
	if err != nil {
		logger.Error("failed-compare-and-delete", err)
		return models.ErrActualLRPCannotBeRemoved
	}

	return nil
}

func (db *ETCDDB) EvacuateActualLRP(
	logger lager.Logger,
	lrpKey *models.ActualLRPKey,
	instanceKey *models.ActualLRPInstanceKey,
	netInfo *models.ActualLRPNetInfo,
	ttl uint64,
) error {
	logger = logger.Session("evacuate-actual-lrp", lager.Data{"process_guid": lrpKey.ProcessGuid, "index": lrpKey.Index})

	logger.Debug("starting")
	defer logger.Debug("complete")

	node, err := db.fetchRaw(logger, EvacuatingActualLRPSchemaPath(lrpKey.ProcessGuid, lrpKey.Index))
	bbsErr := models.ConvertError(err)
	if bbsErr != nil {
		if bbsErr.Type == models.Error_ResourceNotFound {
			return db.createEvacuatingActualLRP(logger, lrpKey, instanceKey, netInfo, ttl)
		}
		return bbsErr
	}

	lrp := models.ActualLRP{}
	err = db.deserializeModel(logger, node, &lrp)
	if err != nil {
		return err
	}

	if lrp.ActualLRPKey.Equal(lrpKey) && lrp.ActualLRPInstanceKey.Equal(instanceKey) &&
		reflect.DeepEqual(lrp.ActualLRPNetInfo, *netInfo) {
		return nil
	}

	lrp.ActualLRPNetInfo = *netInfo
	lrp.ActualLRPKey = *lrpKey
	lrp.ActualLRPInstanceKey = *instanceKey
	lrp.Since = db.clock.Now().UnixNano()

	data, err := db.serializeModel(logger, &lrp)
	if err != nil {
		logger.Error("failed-serializing", err)
		return err
	}

	_, err = db.client.CompareAndSwap(EvacuatingActualLRPSchemaPath(lrp.ProcessGuid, lrp.Index), data, ttl, node.ModifiedIndex)
	if err != nil {
		return ErrorFromEtcdError(logger, err)
	}

	return nil
}

func (db *ETCDDB) createEvacuatingActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, netInfo *models.ActualLRPNetInfo, evacuatingTTLInSeconds uint64) (err error) {
	logger.Debug("create-evacuating-actual-lrp")
	defer logger.Debug("create-evacuating-actual-lrp-complete", lager.Data{"error": err})

	lrp, err := db.newRunningActualLRP(key, instanceKey, netInfo)
	if err != nil {
		return models.ErrActualLRPCannotBeEvacuated
	}

	lrp.ModificationTag.Increment()

	lrpData, err := db.serializeModel(logger, lrp)
	if err != nil {
		return err
	}

	_, err = db.client.Create(EvacuatingActualLRPSchemaPath(key.ProcessGuid, key.Index), lrpData, evacuatingTTLInSeconds)
	if err != nil {
		logger.Error("failed", err)
		return models.ErrActualLRPCannotBeEvacuated
	}

	return nil
}
