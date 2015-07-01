package test_helpers

import etcdclient "github.com/coreos/go-etcd/etcd"

func NewTestHelper(etcdClient *etcdclient.Client) *TestHelper {
	return &TestHelper{etcdClient: etcdClient}
}

type TestHelper struct {
	etcdClient *etcdclient.Client
}
