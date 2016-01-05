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

type V0CellCapacity struct {
	MemoryMB   int `json:"memory_mb"`
	DiskMB     int `json:"disk_mb"`
	Containers int `json:"containers"`
}

func (cap V0CellCapacity) Validate() error {
	var validationError models.ValidationError

	if cap.MemoryMB <= 0 {
		validationError = validationError.Append(models.ErrInvalidField{"memory_mb"})
	}

	if cap.DiskMB < 0 {
		validationError = validationError.Append(models.ErrInvalidField{"disk_mb"})
	}

	if cap.Containers <= 0 {
		validationError = validationError.Append(models.ErrInvalidField{"containers"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

type V0CellPresence struct {
	CellID          string              `json:"cell_id"`
	RepAddress      string              `json:"rep_address"`
	Zone            string              `json:"zone"`
	Capacity        V0CellCapacity      `json:"capacity"`
	RootFSProviders map[string][]string `json:"rootfs_providers"`
}

func (c V0CellPresence) Validate() error {
	var validationError models.ValidationError

	if c.CellID == "" {
		validationError = validationError.Append(models.ErrInvalidField{"cell_id"})
	}

	if c.RepAddress == "" {
		validationError = validationError.Append(models.ErrInvalidField{"rep_address"})
	}

	if err := c.Capacity.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}
