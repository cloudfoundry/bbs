package etcd_helpers

import (
	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/models"
	etcdclient "github.com/coreos/go-etcd/etcd"
	"github.com/pivotal-golang/lager"
)

func NewETCDHelper(etcdClient *etcdclient.Client) *ETCDHelper {
	return &ETCDHelper{etcdClient: etcdClient}
}

type ETCDHelper struct {
	etcdClient *etcdclient.Client
}

//go:generate counterfeiter . TaskCallbackFactory

type TaskCallbackFactory interface {
	TaskCallbackWork(logger lager.Logger, taskDB db.TaskDB, task *models.Task) func()
}
