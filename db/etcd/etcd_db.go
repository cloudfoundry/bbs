package etcd

import (
	"path"
	"strconv"
	"sync"

	"github.com/cloudfoundry-incubator/bbs/auctionhandlers"
	"github.com/cloudfoundry-incubator/bbs/cellhandlers"
	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/format"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/taskworkpool"
	"github.com/coreos/go-etcd/etcd"
	etcdclient "github.com/coreos/go-etcd/etcd"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
)

const DataSchemaRoot = "/v1/"
const VersionKey = "/version"

const maxActualGroupGetterWorkPoolSize = 50
const ActualLRPSchemaRoot = DataSchemaRoot + "actual"
const ActualLRPInstanceKey = "instance"
const ActualLRPEvacuatingKey = "evacuating"

func ActualLRPProcessDir(processGuid string) string {
	return path.Join(ActualLRPSchemaRoot, processGuid)
}

func ActualLRPIndexDir(processGuid string, index int32) string {
	return path.Join(ActualLRPProcessDir(processGuid), strconv.Itoa(int(index)))
}

func ActualLRPSchemaPath(processGuid string, index int32) string {
	return path.Join(ActualLRPIndexDir(processGuid, index), ActualLRPInstanceKey)
}

func EvacuatingActualLRPSchemaPath(processGuid string, index int32) string {
	return path.Join(ActualLRPIndexDir(processGuid, index), ActualLRPEvacuatingKey)
}

const maxDesiredLRPGetterWorkPoolSize = 50
const DesiredLRPSchemaRoot = DataSchemaRoot + "desired"

func DesiredLRPSchemaPath(lrp *models.DesiredLRP) string {
	return DesiredLRPSchemaPathByProcessGuid(lrp.ProcessGuid)
}

func DesiredLRPSchemaPathByProcessGuid(processGuid string) string {
	return path.Join(DesiredLRPSchemaRoot, processGuid)
}

type ETCDDB struct {
	format            *format.Format
	client            StoreClient
	clock             clock.Clock
	inflightWatches   map[chan bool]bool
	inflightWatchLock *sync.Mutex
	auctioneerClient  auctionhandlers.Client
	cellClient        cellhandlers.Client

	taskCompletionClient taskworkpool.TaskCompletionClient

	cellDB db.CellDB
}

func NewETCD(
	format *format.Format,
	storeClient StoreClient,
	auctioneerClient auctionhandlers.Client,
	cellClient cellhandlers.Client,
	cellDB db.CellDB,
	clock clock.Clock,
	taskCC taskworkpool.TaskCompletionClient,
) *ETCDDB {
	return &ETCDDB{
		format,
		storeClient,
		clock,
		map[chan bool]bool{},
		&sync.Mutex{},
		auctioneerClient,
		cellClient,
		taskCC,
		cellDB,
	}
}

func (db *ETCDDB) serializeModel(logger lager.Logger, model format.Versioner) ([]byte, error) {
	encodedPayload, err := format.Marshal(db.format, model)
	if err != nil {
		logger.Error("failed-ro-serialize-model", err)
		return nil, models.NewError(models.Error_InvalidRecord, err.Error())
	}
	return encodedPayload, nil
}

func (db *ETCDDB) deserializeModel(logger lager.Logger, node *etcdclient.Node, model format.Versioner) error {
	err := format.Unmarshal(logger, []byte(node.Value), model)
	if err != nil {
		logger.Error("failed-ro-serialize-model", err)
		return models.NewError(models.Error_InvalidRecord, err.Error())
	}
	return nil
}

func (db *ETCDDB) fetchRecursiveRaw(logger lager.Logger, key string) (*etcd.Node, error) {
	logger.Debug("fetching-recursive-from-etcd")
	response, err := db.client.Get(key, false, true)
	if err != nil {
		return nil, ErrorFromEtcdError(logger, err)
	}
	logger.Debug("succeeded-fetching-recursive-from-etcd", lager.Data{"num-nodes": response.Node.Nodes.Len()})
	return response.Node, nil
}

func (db *ETCDDB) fetchRaw(logger lager.Logger, key string) (*etcd.Node, error) {
	logger.Debug("fetching-from-etcd")
	response, err := db.client.Get(key, false, false)
	if err != nil {
		return nil, ErrorFromEtcdError(logger, err)
	}
	logger.Debug("succeeded-fetching-from-etcd")
	return response.Node, nil
}

const (
	ETCDErrKeyNotFound  = 100
	ETCDErrKeyExists    = 105
	ETCDErrIndexCleared = 401
)

func ErrorFromEtcdError(logger lager.Logger, err error) error {
	if err == nil {
		return nil
	}

	logger = logger.Session("etcd-error", lager.Data{"error": err})
	switch etcdErrCode(err) {
	case ETCDErrKeyNotFound:
		logger.Debug("resource-not-found")
		return models.ErrResourceNotFound
	case ETCDErrKeyExists:
		logger.Debug("resource-exits")
		return models.ErrResourceExists
	default:
		logger.Error("unknown-error", err)
		return models.ErrUnknownError
	}
}

func etcdErrCode(err error) int {
	if err != nil {
		switch err.(type) {
		case etcd.EtcdError:
			return err.(etcd.EtcdError).ErrorCode
		case *etcd.EtcdError:
			return err.(*etcd.EtcdError).ErrorCode
		}
	}
	return 0
}
