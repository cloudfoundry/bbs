// Code generated by protoc-gen-go-bbs. DO NOT EDIT.
// versions:
// - protoc-gen-go-bbs v0.0.1
// - protoc            v5.27.0
// source: modification_tag.proto

package models

// Prevent copylock errors when using ProtoModificationTag directly
type ModificationTag struct {
	Epoch string `json:"epoch"`
	Index uint32 `json:"index"`
}

func (this *ModificationTag) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ModificationTag)
	if !ok {
		that2, ok := that.(ModificationTag)
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

	if this.Epoch != that1.Epoch {
		return false
	}
	if this.Index != that1.Index {
		return false
	}
	return true
}
func (m *ModificationTag) GetEpoch() string {
	if m != nil {
		return m.Epoch
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *ModificationTag) SetEpoch(value string) {
	if m != nil {
		m.Epoch = value
	}
}
func (m *ModificationTag) GetIndex() uint32 {
	if m != nil {
		return m.Index
	}
	var defaultValue uint32
	defaultValue = 0
	return defaultValue
}
func (m *ModificationTag) SetIndex(value uint32) {
	if m != nil {
		m.Index = value
	}
}
func (x *ModificationTag) ToProto() *ProtoModificationTag {
	if x == nil {
		return nil
	}

	proto := &ProtoModificationTag{
		Epoch: x.Epoch,
		Index: x.Index,
	}
	return proto
}

func (x *ProtoModificationTag) FromProto() *ModificationTag {
	if x == nil {
		return nil
	}

	copysafe := &ModificationTag{
		Epoch: x.Epoch,
		Index: x.Index,
	}
	return copysafe
}

func ModificationTagToProtoSlice(values []*ModificationTag) []*ProtoModificationTag {
	if values == nil {
		return nil
	}
	result := make([]*ProtoModificationTag, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func ModificationTagFromProtoSlice(values []*ProtoModificationTag) []*ModificationTag {
	if values == nil {
		return nil
	}
	result := make([]*ModificationTag, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}
