package etcd

import (
	"path"
	"strconv"
	"sync"

	"github.com/cloudfoundry-incubator/bbs/auctionhandlers"
	"github.com/cloudfoundry-incubator/bbs/cellhandlers"
	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/coreos/go-etcd/etcd"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
)

const DataSchemaRoot = "/v1/"

const (
	ETCDErrKeyNotFound  = 100
	ETCDErrIndexCleared = 401
)

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
	return DesiredLRPSchemaPathByProcessGuid(lrp.GetProcessGuid())
}

func DesiredLRPSchemaPathByProcessGuid(processGuid string) string {
	return path.Join(DesiredLRPSchemaRoot, processGuid)
}

type ETCDDB struct {
	client            *etcd.Client
	clock             clock.Clock
	inflightWatches   map[chan bool]bool
	inflightWatchLock *sync.Mutex
	auctioneerClient  auctionhandlers.Client
	cellClient        cellhandlers.Client

	cellDB db.CellDB
}

func NewETCD(etcdClient *etcd.Client, auctioneerClient auctionhandlers.Client, cellClient cellhandlers.Client, cellDB db.CellDB, clock clock.Clock) *ETCDDB {
	return &ETCDDB{etcdClient,
		clock,
		map[chan bool]bool{},
		&sync.Mutex{},
		auctioneerClient,
		cellClient,
		cellDB,
	}
}

func (db *ETCDDB) fetchRecursiveRaw(logger lager.Logger, key string) (*etcd.Node, *models.Error) {
	logger.Debug("fetching-recursive-from-etcd")
	response, err := db.client.Get(key, false, true)
	if etcdErrCode(err) == ETCDErrKeyNotFound {
		logger.Debug("no-nodes-to-fetch")
		return nil, models.ErrResourceNotFound
	} else if err != nil {
		logger.Error("failed-fetching-recursive-from-etcd", err)
		return nil, models.ErrUnknownError
	}
	logger.Debug("succeeded-fetching-recursive-from-etcd", lager.Data{"num-lrps": response.Node.Nodes.Len()})
	return response.Node, nil
}

func (db *ETCDDB) fetchRaw(logger lager.Logger, key string) (*etcd.Node, *models.Error) {
	logger.Debug("fetching-from-etcd")
	response, err := db.client.Get(key, false, false)
	if etcdErrCode(err) == ETCDErrKeyNotFound {
		logger.Debug("no-node-to-fetch")
		return nil, models.ErrResourceNotFound
	} else if err != nil {
		logger.Error("failed-fetching-from-etcd", err)
		return nil, models.ErrUnknownError
	}
	logger.Debug("succeeded-fetching-from-etcd")
	return response.Node, nil
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
