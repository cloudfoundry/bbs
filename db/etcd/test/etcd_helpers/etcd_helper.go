package etcd_helpers

import (
	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/format"
)

type ETCDHelper struct {
	client etcd.StoreClient
	format *format.Format
}

func NewETCDHelper(format *format.Format, client etcd.StoreClient) *ETCDHelper {
	return &ETCDHelper{
		client: client,
		format: format,
	}
}
