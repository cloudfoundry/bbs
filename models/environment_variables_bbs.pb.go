// Code generated by protoc-gen-go-bbs. DO NOT EDIT.
// versions:
// - protoc-gen-go-bbs v0.0.1
// - protoc            v5.28.1
// source: environment_variables.proto

package models

// Prevent copylock errors when using ProtoEnvironmentVariable directly
type EnvironmentVariable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (this *EnvironmentVariable) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*EnvironmentVariable)
	if !ok {
		that2, ok := that.(EnvironmentVariable)
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
	if this.Value != that1.Value {
		return false
	}
	return true
}
func (m *EnvironmentVariable) GetName() string {
	if m != nil {
		return m.Name
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *EnvironmentVariable) SetName(value string) {
	if m != nil {
		m.Name = value
	}
}
func (m *EnvironmentVariable) GetValue() string {
	if m != nil {
		return m.Value
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *EnvironmentVariable) SetValue(value string) {
	if m != nil {
		m.Value = value
	}
}
func (x *EnvironmentVariable) ToProto() *ProtoEnvironmentVariable {
	if x == nil {
		return nil
	}

	proto := &ProtoEnvironmentVariable{
		Name:  x.Name,
		Value: x.Value,
	}
	return proto
}

func (x *ProtoEnvironmentVariable) FromProto() *EnvironmentVariable {
	if x == nil {
		return nil
	}

	copysafe := &EnvironmentVariable{
		Name:  x.Name,
		Value: x.Value,
	}
	return copysafe
}

func EnvironmentVariableToProtoSlice(values []*EnvironmentVariable) []*ProtoEnvironmentVariable {
	if values == nil {
		return nil
	}
	result := make([]*ProtoEnvironmentVariable, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func EnvironmentVariableFromProtoSlice(values []*ProtoEnvironmentVariable) []*EnvironmentVariable {
	if values == nil {
		return nil
	}
	result := make([]*EnvironmentVariable, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}
