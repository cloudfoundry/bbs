// Code generated by protoc-gen-go-bbs. DO NOT EDIT.
// versions:
// - protoc-gen-go-bbs v0.0.1
// - protoc            v5.29.4
// source: cached_dependency.proto

package models

// Prevent copylock errors when using ProtoCachedDependency directly
type CachedDependency struct {
	Name              string `json:"name"`
	From              string `json:"from"`
	To                string `json:"to"`
	CacheKey          string `json:"cache_key"`
	LogSource         string `json:"log_source"`
	ChecksumAlgorithm string `json:"checksum_algorithm,omitempty"`
	ChecksumValue     string `json:"checksum_value,omitempty"`
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
	var defaultValue string
	defaultValue = ""
	return defaultValue
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
	var defaultValue string
	defaultValue = ""
	return defaultValue
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
	var defaultValue string
	defaultValue = ""
	return defaultValue
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
	var defaultValue string
	defaultValue = ""
	return defaultValue
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
	var defaultValue string
	defaultValue = ""
	return defaultValue
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
	var defaultValue string
	defaultValue = ""
	return defaultValue
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
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *CachedDependency) SetChecksumValue(value string) {
	if m != nil {
		m.ChecksumValue = value
	}
}
func (x *CachedDependency) ToProto() *ProtoCachedDependency {
	if x == nil {
		return nil
	}

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

func (x *ProtoCachedDependency) FromProto() *CachedDependency {
	if x == nil {
		return nil
	}

	copysafe := &CachedDependency{
		Name:              x.Name,
		From:              x.From,
		To:                x.To,
		CacheKey:          x.CacheKey,
		LogSource:         x.LogSource,
		ChecksumAlgorithm: x.ChecksumAlgorithm,
		ChecksumValue:     x.ChecksumValue,
	}
	return copysafe
}

func CachedDependencyToProtoSlice(values []*CachedDependency) []*ProtoCachedDependency {
	if values == nil {
		return nil
	}
	result := make([]*ProtoCachedDependency, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func CachedDependencyFromProtoSlice(values []*ProtoCachedDependency) []*CachedDependency {
	if values == nil {
		return nil
	}
	result := make([]*CachedDependency, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}
