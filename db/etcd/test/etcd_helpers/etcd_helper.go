package etcd_helpers

import "github.com/cloudfoundry-incubator/bbs/db/etcd"

func NewETCDHelper(client etcd.StoreClient) *ETCDHelper {
	return &ETCDHelper{client: client}
}

type ETCDHelper struct {
	client etcd.StoreClient
}
