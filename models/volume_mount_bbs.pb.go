// Code generated by protoc-gen-go-bbs. DO NOT EDIT.
// versions:
// - protoc-gen-go-bbs v0.0.1
// - protoc            v5.26.1
// source: volume_mount.proto

package models

// Prevent copylock errors when using ProtoSharedDevice directly
type SharedDevice struct {
	VolumeId    string
	MountConfig string
}

func (this *SharedDevice) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*SharedDevice)
	if !ok {
		that2, ok := that.(SharedDevice)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}

	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}

	if this.VolumeId != that1.VolumeId {
		return false
	}
	if this.MountConfig != that1.MountConfig {
		return false
	}
	return true
}
func (m *SharedDevice) GetVolumeId() string {
	if m != nil {
		return m.VolumeId
	}
	return ""
}
func (m *SharedDevice) SetVolumeId(value string) {
	if m != nil {
		m.VolumeId = value
	}
}
func (m *SharedDevice) GetMountConfig() string {
	if m != nil {
		return m.MountConfig
	}
	return ""
}
func (m *SharedDevice) SetMountConfig(value string) {
	if m != nil {
		m.MountConfig = value
	}
}
func (x *SharedDevice) ToProto() *ProtoSharedDevice {
	if x == nil {
		return nil
	}

	proto := &ProtoSharedDevice{
		VolumeId:    x.VolumeId,
		MountConfig: x.MountConfig,
	}
	return proto
}

func (x *ProtoSharedDevice) FromProto() *SharedDevice {
	if x == nil {
		return nil
	}

	copysafe := &SharedDevice{
		VolumeId:    x.VolumeId,
		MountConfig: x.MountConfig,
	}
	return copysafe
}

func SharedDeviceToProtoSlice(values []*SharedDevice) []*ProtoSharedDevice {
	result := make([]*ProtoSharedDevice, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func SharedDeviceFromProtoSlice(values []*ProtoSharedDevice) []*SharedDevice {
	result := make([]*SharedDevice, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoVolumeMount directly
type VolumeMount struct {
	Driver       string
	ContainerDir string
	Mode         string
	Shared       *SharedDevice
}

func (this *VolumeMount) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*VolumeMount)
	if !ok {
		that2, ok := that.(VolumeMount)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}

	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}

	if this.Driver != that1.Driver {
		return false
	}
	if this.ContainerDir != that1.ContainerDir {
		return false
	}
	if this.Mode != that1.Mode {
		return false
	}
	if !this.Shared.Equal(that1.Shared) {
		return false
	}
	return true
}
func (m *VolumeMount) GetDriver() string {
	if m != nil {
		return m.Driver
	}
	return ""
}
func (m *VolumeMount) SetDriver(value string) {
	if m != nil {
		m.Driver = value
	}
}
func (m *VolumeMount) GetContainerDir() string {
	if m != nil {
		return m.ContainerDir
	}
	return ""
}
func (m *VolumeMount) SetContainerDir(value string) {
	if m != nil {
		m.ContainerDir = value
	}
}
func (m *VolumeMount) GetMode() string {
	if m != nil {
		return m.Mode
	}
	return ""
}
func (m *VolumeMount) SetMode(value string) {
	if m != nil {
		m.Mode = value
	}
}
func (m *VolumeMount) GetShared() *SharedDevice {
	if m != nil {
		return m.Shared
	}
	return nil
}
func (m *VolumeMount) SetShared(value *SharedDevice) {
	if m != nil {
		m.Shared = value
	}
}
func (x *VolumeMount) ToProto() *ProtoVolumeMount {
	if x == nil {
		return nil
	}

	proto := &ProtoVolumeMount{
		Driver:       x.Driver,
		ContainerDir: x.ContainerDir,
		Mode:         x.Mode,
		Shared:       x.Shared.ToProto(),
	}
	return proto
}

func (x *ProtoVolumeMount) FromProto() *VolumeMount {
	if x == nil {
		return nil
	}

	copysafe := &VolumeMount{
		Driver:       x.Driver,
		ContainerDir: x.ContainerDir,
		Mode:         x.Mode,
		Shared:       x.Shared.FromProto(),
	}
	return copysafe
}

func VolumeMountToProtoSlice(values []*VolumeMount) []*ProtoVolumeMount {
	result := make([]*ProtoVolumeMount, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func VolumeMountFromProtoSlice(values []*ProtoVolumeMount) []*VolumeMount {
	result := make([]*VolumeMount, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoVolumePlacement directly
type VolumePlacement struct {
	DriverNames []string
}

func (this *VolumePlacement) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*VolumePlacement)
	if !ok {
		that2, ok := that.(VolumePlacement)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}

	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}

	if len(this.DriverNames) != len(that1.DriverNames) {
		return false
	}
	for i := range this.DriverNames {
		if this.DriverNames[i] != that1.DriverNames[i] {
			return false
		}
	}
	return true
}
func (m *VolumePlacement) GetDriverNames() []string {
	if m != nil {
		return m.DriverNames
	}
	return nil
}
func (m *VolumePlacement) SetDriverNames(value []string) {
	if m != nil {
		m.DriverNames = value
	}
}
func (x *VolumePlacement) ToProto() *ProtoVolumePlacement {
	if x == nil {
		return nil
	}

	proto := &ProtoVolumePlacement{
		DriverNames: x.DriverNames,
	}
	return proto
}

func (x *ProtoVolumePlacement) FromProto() *VolumePlacement {
	if x == nil {
		return nil
	}

	copysafe := &VolumePlacement{
		DriverNames: x.DriverNames,
	}
	return copysafe
}

func VolumePlacementToProtoSlice(values []*VolumePlacement) []*ProtoVolumePlacement {
	result := make([]*ProtoVolumePlacement, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func VolumePlacementFromProtoSlice(values []*ProtoVolumePlacement) []*VolumePlacement {
	result := make([]*VolumePlacement, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}
