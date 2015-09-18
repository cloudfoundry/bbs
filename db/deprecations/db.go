package deprecations

import (
	"path"

	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/models"
)

var (
	DesiredLRPSchemaRoot = etcd.V1SchemaRoot + "desired"
)

func DesiredLRPSchemaPath(lrp *models.DesiredLRP) string {
	return DesiredLRPSchemaPathByProcessGuid(lrp.ProcessGuid)
}

func DesiredLRPSchemaPathByProcessGuid(processGuid string) string {
	return path.Join(DesiredLRPSchemaRoot, processGuid)
}
