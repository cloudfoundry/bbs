// Code generated by protoc-gen-go-bbs. DO NOT EDIT.
// versions:
// - protoc-gen-go-bbs v0.0.1
// - protoc            v5.29.0
// source: ping.proto

package models

// Prevent copylock errors when using ProtoPingResponse directly
type PingResponse struct {
	Available bool `json:"available"`
}

func (this *PingResponse) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*PingResponse)
	if !ok {
		that2, ok := that.(PingResponse)
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

	if this.Available != that1.Available {
		return false
	}
	return true
}
func (m *PingResponse) GetAvailable() bool {
	if m != nil {
		return m.Available
	}
	var defaultValue bool
	defaultValue = false
	return defaultValue
}
func (m *PingResponse) SetAvailable(value bool) {
	if m != nil {
		m.Available = value
	}
}
func (x *PingResponse) ToProto() *ProtoPingResponse {
	if x == nil {
		return nil
	}

	proto := &ProtoPingResponse{
		Available: x.Available,
	}
	return proto
}

func (x *ProtoPingResponse) FromProto() *PingResponse {
	if x == nil {
		return nil
	}

	copysafe := &PingResponse{
		Available: x.Available,
	}
	return copysafe
}

func PingResponseToProtoSlice(values []*PingResponse) []*ProtoPingResponse {
	if values == nil {
		return nil
	}
	result := make([]*ProtoPingResponse, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func PingResponseFromProtoSlice(values []*ProtoPingResponse) []*PingResponse {
	if values == nil {
		return nil
	}
	result := make([]*PingResponse, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}
