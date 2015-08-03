package etcd

import (
	"sync"

	"github.com/cloudfoundry-incubator/auctioneer"
	"github.com/cloudfoundry-incubator/rep"

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

type ETCDDB struct {
	client            *etcd.Client
	clock             clock.Clock
	inflightWatches   map[chan bool]bool
	inflightWatchLock *sync.Mutex
	auctioneerClient  auctioneer.Client
	cellClient        rep.CellClient

	cellDB db.CellDB
}

func NewETCD(etcdClient *etcd.Client, auctioneerClient auctioneer.Client, cellClient rep.CellClient, cellDB db.CellDB, clock clock.Clock) *ETCDDB {
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
