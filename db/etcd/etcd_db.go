package db

import (
	"github.com/cloudfoundry-incubator/bbs"
	"github.com/coreos/go-etcd/etcd"
	"github.com/pivotal-golang/lager"
)

const DataSchemaRoot = "/v1/"

const (
	ETCDErrKeyNotFound = 100
)

type ETCDDB struct {
	client *etcd.Client
}

func NewETCD(etcdClient *etcd.Client) *ETCDDB {
	return &ETCDDB{etcdClient}
}

func (db *ETCDDB) fetchRecursiveRaw(key string, logger lager.Logger) (*etcd.Node, *bbs.Error) {
	logger.Debug("fetching-recursive-from-etcd")
	response, err := db.client.Get(key, false, true)
	if etcdErr, ok := err.(*etcd.EtcdError); ok && etcdErr.ErrorCode == ETCDErrKeyNotFound {
		logger.Debug("no-nodes-to-fetch")
		return nil, bbs.ErrResourceNotFound
	} else if err != nil {
		logger.Error("failed-fetching-recursive-from-etcd", err)
		return nil, bbs.ErrUnknownError
	}
	logger.Debug("succeeded-fetching-recursive-from-etcd", lager.Data{"num-lrps": response.Node.Nodes.Len()})
	return response.Node, nil
}

func (db *ETCDDB) fetchRaw(key string, logger lager.Logger) (*etcd.Node, *bbs.Error) {
	logger.Debug("fetching-from-etcd")
	response, err := db.client.Get(key, false, false)
	if etcdErr, ok := err.(*etcd.EtcdError); ok && etcdErr.ErrorCode == ETCDErrKeyNotFound {
		logger.Debug("no-node-to-fetch")
		return nil, bbs.ErrResourceNotFound
	} else if err != nil {
		logger.Error("failed-fetching-from-etcd", err)
		return nil, bbs.ErrUnknownError
	}
	logger.Debug("succeeded-fetching-from-etcd")
	return response.Node, nil
}
