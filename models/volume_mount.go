package models

import "code.cloudfoundry.org/bbs/format"

func (*VolumePlacement) Version() format.Version {
	return format.V1
}

func (*VolumePlacement) Validate() error {
	return nil
}

// to handle old cases, can be removed as soon as the bridge speaks V1
func (v *VolumeMount) VersionUpToV1() *VolumeMount {
	mode := "rw"
	if v.DeprecatedMode == BindMountMode_RO {
		mode = "r"
	}

	return &VolumeMount{
		Driver:       v.Driver,
		ContainerDir: v.ContainerDir,
		Mode:         mode,
		Shared: &SharedDevice{
			VolumeId:    v.DeprecatedVolumeId,
			MountConfig: string(v.DeprecatedConfig),
		},
	}
}
