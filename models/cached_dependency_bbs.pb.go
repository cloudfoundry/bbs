// Code generated by protoc-gen-go-bbs. DO NOT EDIT.
// versions:
// - protoc-gen-go-bbs v0.0.1
// - protoc            v5.27.0--rc1
// source: cached_dependency.proto

package models

// Prevent copylock errors when using ProtoCachedDependency directly
type CachedDependency struct {
	Name              string
	From              string
	To                string
	CacheKey          string
	LogSource         string
	ChecksumAlgorithm string
	ChecksumValue     string
}

func (this *CachedDependency) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*CachedDependency)
	if !ok {
		that2, ok := that.(CachedDependency)
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

	if this.Name != that1.Name {
		return false
	}
	if this.From != that1.From {
		return false
	}
	if this.To != that1.To {
		return false
	}
	if this.CacheKey != that1.CacheKey {
		return false
	}
	if this.LogSource != that1.LogSource {
		return false
	}
	if this.ChecksumAlgorithm != that1.ChecksumAlgorithm {
		return false
	}
	if this.ChecksumValue != that1.ChecksumValue {
		return false
	}
	return true
}
func (m *CachedDependency) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}
func (m *CachedDependency) SetName(value string) {
	if m != nil {
		m.Name = value
	}
}
func (m *CachedDependency) GetFrom() string {
	if m != nil {
		return m.From
	}
	return ""
}
func (m *CachedDependency) SetFrom(value string) {
	if m != nil {
		m.From = value
	}
}
func (m *CachedDependency) GetTo() string {
	if m != nil {
		return m.To
	}
	return ""
}
func (m *CachedDependency) SetTo(value string) {
	if m != nil {
		m.To = value
	}
}
func (m *CachedDependency) GetCacheKey() string {
	if m != nil {
		return m.CacheKey
	}
	return ""
}
func (m *CachedDependency) SetCacheKey(value string) {
	if m != nil {
		m.CacheKey = value
	}
}
func (m *CachedDependency) GetLogSource() string {
	if m != nil {
		return m.LogSource
	}
	return ""
}
func (m *CachedDependency) SetLogSource(value string) {
	if m != nil {
		m.LogSource = value
	}
}
func (m *CachedDependency) GetChecksumAlgorithm() string {
	if m != nil {
		return m.ChecksumAlgorithm
	}
	return ""
}
func (m *CachedDependency) SetChecksumAlgorithm(value string) {
	if m != nil {
		m.ChecksumAlgorithm = value
	}
}
func (m *CachedDependency) GetChecksumValue() string {
	if m != nil {
		return m.ChecksumValue
	}
	return ""
}
func (m *CachedDependency) SetChecksumValue(value string) {
	if m != nil {
		m.ChecksumValue = value
	}
}
func (x *CachedDependency) ToProto() *ProtoCachedDependency {
	proto := &ProtoCachedDependency{
		Name:              x.Name,
		From:              x.From,
		To:                x.To,
		CacheKey:          x.CacheKey,
		LogSource:         x.LogSource,
		ChecksumAlgorithm: x.ChecksumAlgorithm,
		ChecksumValue:     x.ChecksumValue,
	}
	return proto
}

func CachedDependencyProtoMap(values []*CachedDependency) []*ProtoCachedDependency {
	result := make([]*ProtoCachedDependency, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}