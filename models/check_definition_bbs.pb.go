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

func (m *CheckDefinition) GetChecks() []*Check {
	if m != nil {
		return m.Checks
	}
	return nil
}
func (m *CheckDefinition) SetChecks(value []*Check) {
	if m != nil {
		m.Checks = value
	}
}
func (m *CheckDefinition) GetLogSource() string {
	if m != nil {
		return m.LogSource
	}
	return ""
}
func (m *CheckDefinition) SetLogSource(value string) {
	if m != nil {
		m.LogSource = value
	}
}
func (m *CheckDefinition) GetReadinessChecks() []*Check {
	if m != nil {
		return m.ReadinessChecks
	}
	return nil
}
func (m *CheckDefinition) SetReadinessChecks(value []*Check) {
	if m != nil {
		m.ReadinessChecks = value
	}
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

func (m *Check) GetTcpCheck() *TCPCheck {
	if m != nil {
		return m.TcpCheck
	}
	return nil
}
func (m *Check) SetTcpCheck(value *TCPCheck) {
	if m != nil {
		m.TcpCheck = value
	}
}
func (m *Check) GetHttpCheck() *HTTPCheck {
	if m != nil {
		return m.HttpCheck
	}
	return nil
}
func (m *Check) SetHttpCheck(value *HTTPCheck) {
	if m != nil {
		m.HttpCheck = value
	}
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

func (m *TCPCheck) GetPort() uint32 {
	if m != nil {
		return m.Port
	}
	return 0
}
func (m *TCPCheck) SetPort(value uint32) {
	if m != nil {
		m.Port = value
	}
}
func (m *TCPCheck) GetConnectTimeoutMs() uint64 {
	if m != nil {
		return m.ConnectTimeoutMs
	}
	return 0
}
func (m *TCPCheck) SetConnectTimeoutMs(value uint64) {
	if m != nil {
		m.ConnectTimeoutMs = value
	}
}
func (m *TCPCheck) GetIntervalMs() uint64 {
	if m != nil {
		return m.IntervalMs
	}
	return 0
}
func (m *TCPCheck) SetIntervalMs(value uint64) {
	if m != nil {
		m.IntervalMs = value
	}
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

func (m *HTTPCheck) GetPort() uint32 {
	if m != nil {
		return m.Port
	}
	return 0
}
func (m *HTTPCheck) SetPort(value uint32) {
	if m != nil {
		m.Port = value
	}
}
func (m *HTTPCheck) GetRequestTimeoutMs() uint64 {
	if m != nil {
		return m.RequestTimeoutMs
	}
	return 0
}
func (m *HTTPCheck) SetRequestTimeoutMs(value uint64) {
	if m != nil {
		m.RequestTimeoutMs = value
	}
}
func (m *HTTPCheck) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}
func (m *HTTPCheck) SetPath(value string) {
	if m != nil {
		m.Path = value
	}
}
func (m *HTTPCheck) GetIntervalMs() uint64 {
	if m != nil {
		return m.IntervalMs
	}
	return 0
}
func (m *HTTPCheck) SetIntervalMs(value uint64) {
	if m != nil {
		m.IntervalMs = value
	}
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
