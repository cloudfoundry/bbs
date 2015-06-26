package db

import "github.com/coreos/go-etcd/etcd"

const DataSchemaRoot = "/v1/"

type ETCDDB struct {
	client *etcd.Client
}

func NewETCD(etcdClient *etcd.Client) *ETCDDB {
	return &ETCDDB{etcdClient}
}
