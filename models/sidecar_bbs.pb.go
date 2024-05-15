// Code generated by protoc-gen-go-bbs. DO NOT EDIT.
// versions:
// - protoc-gen-go-bbs v0.0.1
// - protoc            v5.26.1
// source: sidecar.proto

package models

// Prevent copylock errors when using ProtoSidecar directly
type Sidecar struct {
	Action   *Action
	DiskMb   int32
	MemoryMb int32
}

func (this *Sidecar) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*Sidecar)
	if !ok {
		that2, ok := that.(Sidecar)
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

	if !this.Action.Equal(*that1.Action) {
		return false
	}
	if this.DiskMb != that1.DiskMb {
		return false
	}
	if this.MemoryMb != that1.MemoryMb {
		return false
	}
	return true
}
func (m *Sidecar) GetAction() *Action {
	if m != nil {
		return m.Action
	}
	return nil
}
func (m *Sidecar) SetAction(value *Action) {
	if m != nil {
		m.Action = value
	}
}
func (m *Sidecar) GetDiskMb() int32 {
	if m != nil {
		return m.DiskMb
	}
	return 0
}
func (m *Sidecar) SetDiskMb(value int32) {
	if m != nil {
		m.DiskMb = value
	}
}
func (m *Sidecar) GetMemoryMb() int32 {
	if m != nil {
		return m.MemoryMb
	}
	return 0
}
func (m *Sidecar) SetMemoryMb(value int32) {
	if m != nil {
		m.MemoryMb = value
	}
}
func (x *Sidecar) ToProto() *ProtoSidecar {
	if x == nil {
		return nil
	}

	proto := &ProtoSidecar{
		Action:   x.Action.ToProto(),
		DiskMb:   x.DiskMb,
		MemoryMb: x.MemoryMb,
	}
	return proto
}

func (x *ProtoSidecar) FromProto() *Sidecar {
	if x == nil {
		return nil
	}

	copysafe := &Sidecar{
		Action:   x.Action.FromProto(),
		DiskMb:   x.DiskMb,
		MemoryMb: x.MemoryMb,
	}
	return copysafe
}

func SidecarToProtoSlice(values []*Sidecar) []*ProtoSidecar {
	if values == nil {
		return nil
	}
	result := make([]*ProtoSidecar, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func SidecarFromProtoSlice(values []*ProtoSidecar) []*Sidecar {
	if values == nil {
		return nil
	}
	result := make([]*Sidecar, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}
