package models

import "errors"

func (*VolumePlacement) Version() Version {
	return V1
}

func (*VolumePlacement) Validate() error {
	return nil
}

func (v *VolumeMount) Validate() error {
	var ve ValidationError
	if v.Driver == "" {
		ve = ve.Append(errors.New("invalid volume_mount driver"))
	}
	if !(v.Mode == "r" || v.Mode == "rw") {
		ve = ve.Append(errors.New("invalid volume_mount mode"))
	}
	if v.Shared != nil && v.Shared.VolumeId == "" {
		ve = ve.Append(errors.New("invalid volume_mount volume id"))
	}

	if !ve.Empty() {
		return ve
	}

	return nil
}
