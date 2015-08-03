package etcd_helpers

import etcdclient "github.com/coreos/go-etcd/etcd"

func NewETCDHelper(etcdClient *etcdclient.Client) *ETCDHelper {
	return &ETCDHelper{etcdClient: etcdClient}
}

type ETCDHelper struct {
	etcdClient *etcdclient.Client
}
