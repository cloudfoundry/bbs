package db

import (
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

func (db *ETCDDB) fetchRecursiveRaw(key string, logger lager.Logger) (*etcd.Node, error) {
	logger.Debug("fetching-recursive-from-etcd")
	response, err := db.client.Get(key, false, true)
	if etcdErr, ok := err.(*etcd.EtcdError); ok && etcdErr.ErrorCode == ETCDErrKeyNotFound {
		logger.Debug("no-nodes-to-fetch")
		return nil, etcdErr
	} else if err != nil {
		logger.Error("failed-fetching-recd", err)
		return nil, err
	}
	logger.Debug("succeeded-fetching-recursive-from-etcd", lager.Data{"num-lrps": response.Node.Nodes.Len()})
	return response.Node, nil
}
