package etcd_helpers

import (
	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

func NewETCDHelper(client etcd.StoreClient) *ETCDHelper {
	return &ETCDHelper{client: client}
}

type ETCDHelper struct {
	client etcd.StoreClient
}

//go:generate counterfeiter . TaskCallbackFactory

type TaskCallbackFactory interface {
	TaskCallbackWork(logger lager.Logger, taskDB db.TaskDB, task *models.Task) func()
}
