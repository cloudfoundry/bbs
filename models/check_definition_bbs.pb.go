// Code generated by protoc-gen-go-bbs. DO NOT EDIT.
// versions:
// - protoc-gen-go-bbs v0.0.1
// - protoc            v5.27.0--rc1
// source: check_definition.proto

package models

// Prevent copylock errors when using ProtoCheckDefinition directly
type CheckDefinition struct {
	Checks          []*Check
	LogSource       string
	ReadinessChecks []*Check
}

func (x *CheckDefinition) ToProto() *ProtoCheckDefinition {
	proto := &ProtoCheckDefinition{
		Checks:          CheckProtoMap(x.Checks),
		LogSource:       x.LogSource,
		ReadinessChecks: CheckProtoMap(x.ReadinessChecks),
	}
	return proto
}

func CheckDefinitionProtoMap(values []*CheckDefinition) []*ProtoCheckDefinition {
	result := make([]*ProtoCheckDefinition, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

// Prevent copylock errors when using ProtoCheck directly
type Check struct {
	TcpCheck  *TCPCheck
	HttpCheck *HTTPCheck
}

func (x *Check) ToProto() *ProtoCheck {
	proto := &ProtoCheck{
		TcpCheck:  x.TcpCheck.ToProto(),
		HttpCheck: x.HttpCheck.ToProto(),
	}
	return proto
}

func CheckProtoMap(values []*Check) []*ProtoCheck {
	result := make([]*ProtoCheck, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

// Prevent copylock errors when using ProtoTCPCheck directly
type TCPCheck struct {
	Port             uint32
	ConnectTimeoutMs uint64
	IntervalMs       uint64
}

func (x *TCPCheck) ToProto() *ProtoTCPCheck {
	proto := &ProtoTCPCheck{
		Port:             x.Port,
		ConnectTimeoutMs: x.ConnectTimeoutMs,
		IntervalMs:       x.IntervalMs,
	}
	return proto
}

func TCPCheckProtoMap(values []*TCPCheck) []*ProtoTCPCheck {
	result := make([]*ProtoTCPCheck, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

// Prevent copylock errors when using ProtoHTTPCheck directly
type HTTPCheck struct {
	Port             uint32
	RequestTimeoutMs uint64
	Path             string
	IntervalMs       uint64
}

func (x *HTTPCheck) ToProto() *ProtoHTTPCheck {
	proto := &ProtoHTTPCheck{
		Port:             x.Port,
		RequestTimeoutMs: x.RequestTimeoutMs,
		Path:             x.Path,
		IntervalMs:       x.IntervalMs,
	}
	return proto
}

func HTTPCheckProtoMap(values []*HTTPCheck) []*ProtoHTTPCheck {
	result := make([]*ProtoHTTPCheck, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}
