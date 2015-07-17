package etcd

import (
	"sync"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/coreos/go-etcd/etcd"
	"github.com/pivotal-golang/lager"
)

const DataSchemaRoot = "/v1/"

const (
	ETCDErrKeyNotFound  = 100
	ETCDErrIndexCleared = 401
)

type ETCDDB struct {
	client            *etcd.Client
	inflightWatches   map[chan bool]bool
	inflightWatchLock *sync.Mutex
}

func NewETCD(etcdClient *etcd.Client) *ETCDDB {
	return &ETCDDB{etcdClient, map[chan bool]bool{}, &sync.Mutex{}}
}

func (db *ETCDDB) fetchRecursiveRaw(key string, logger lager.Logger) (*etcd.Node, *models.Error) {
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

func (db *ETCDDB) fetchRaw(key string, logger lager.Logger) (*etcd.Node, *models.Error) {
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
