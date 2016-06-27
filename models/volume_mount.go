package models

import "code.cloudfoundry.org/bbs/format"

func (*VolumePlacement) Version() format.Version {
	return format.V1
}

func (*VolumePlacement) Validate() error {
	return nil
}
