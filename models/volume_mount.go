package models

import "github.com/cloudfoundry-incubator/bbs/format"

func (*VolumePlacement) Version() format.Version {
	return format.V1
}

func (*VolumePlacement) MigrateFromVersion(v format.Version) error {
	return nil
}

func (*VolumePlacement) Validate() error {
	return nil
}
