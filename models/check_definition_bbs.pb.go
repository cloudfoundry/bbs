// Code generated by protoc-gen-go-bbs. DO NOT EDIT.
// versions:
// - protoc-gen-go-bbs v0.0.1
// - protoc            v5.28.1
// source: check_definition.proto

package models

// Prevent copylock errors when using ProtoCheckDefinition directly
type CheckDefinition struct {
	Checks          []*Check `json:"checks,omitempty"`
	LogSource       string   `json:"log_source"`
	ReadinessChecks []*Check `json:"readiness_checks,omitempty"`
}

func (this *CheckDefinition) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*CheckDefinition)
	if !ok {
		that2, ok := that.(CheckDefinition)
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

	if this.Checks == nil {
		if that1.Checks != nil {
			return false
		}
	} else if len(this.Checks) != len(that1.Checks) {
		return false
	}
	for i := range this.Checks {
		if !this.Checks[i].Equal(that1.Checks[i]) {
			return false
		}
	}
	if this.LogSource != that1.LogSource {
		return false
	}
	if this.ReadinessChecks == nil {
		if that1.ReadinessChecks != nil {
			return false
		}
	} else if len(this.ReadinessChecks) != len(that1.ReadinessChecks) {
		return false
	}
	for i := range this.ReadinessChecks {
		if !this.ReadinessChecks[i].Equal(that1.ReadinessChecks[i]) {
			return false
		}
	}
	return true
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
	var defaultValue string
	defaultValue = ""
	return defaultValue
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
	if x == nil {
		return nil
	}

	proto := &ProtoCheckDefinition{
		Checks:          CheckToProtoSlice(x.Checks),
		LogSource:       x.LogSource,
		ReadinessChecks: CheckToProtoSlice(x.ReadinessChecks),
	}
	return proto
}

func (x *ProtoCheckDefinition) FromProto() *CheckDefinition {
	if x == nil {
		return nil
	}

	copysafe := &CheckDefinition{
		Checks:          CheckFromProtoSlice(x.Checks),
		LogSource:       x.LogSource,
		ReadinessChecks: CheckFromProtoSlice(x.ReadinessChecks),
	}
	return copysafe
}

func CheckDefinitionToProtoSlice(values []*CheckDefinition) []*ProtoCheckDefinition {
	if values == nil {
		return nil
	}
	result := make([]*ProtoCheckDefinition, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func CheckDefinitionFromProtoSlice(values []*ProtoCheckDefinition) []*CheckDefinition {
	if values == nil {
		return nil
	}
	result := make([]*CheckDefinition, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoCheck directly
type Check struct {
	TcpCheck  *TCPCheck  `json:"tcp_check,omitempty"`
	HttpCheck *HTTPCheck `json:"http_check,omitempty"`
}

func (this *Check) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*Check)
	if !ok {
		that2, ok := that.(Check)
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

	if this.TcpCheck == nil {
		if that1.TcpCheck != nil {
			return false
		}
	} else if !this.TcpCheck.Equal(*that1.TcpCheck) {
		return false
	}
	if this.HttpCheck == nil {
		if that1.HttpCheck != nil {
			return false
		}
	} else if !this.HttpCheck.Equal(*that1.HttpCheck) {
		return false
	}
	return true
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
	if x == nil {
		return nil
	}

	proto := &ProtoCheck{
		TcpCheck:  x.TcpCheck.ToProto(),
		HttpCheck: x.HttpCheck.ToProto(),
	}
	return proto
}

func (x *ProtoCheck) FromProto() *Check {
	if x == nil {
		return nil
	}

	copysafe := &Check{
		TcpCheck:  x.TcpCheck.FromProto(),
		HttpCheck: x.HttpCheck.FromProto(),
	}
	return copysafe
}

func CheckToProtoSlice(values []*Check) []*ProtoCheck {
	if values == nil {
		return nil
	}
	result := make([]*ProtoCheck, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func CheckFromProtoSlice(values []*ProtoCheck) []*Check {
	if values == nil {
		return nil
	}
	result := make([]*Check, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoTCPCheck directly
type TCPCheck struct {
	Port             uint32 `json:"port"`
	ConnectTimeoutMs uint64 `json:"connect_timeout_ms,omitempty"`
	IntervalMs       uint64 `json:"interval_ms,omitempty"`
}

func (this *TCPCheck) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*TCPCheck)
	if !ok {
		that2, ok := that.(TCPCheck)
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

	if this.Port != that1.Port {
		return false
	}
	if this.ConnectTimeoutMs != that1.ConnectTimeoutMs {
		return false
	}
	if this.IntervalMs != that1.IntervalMs {
		return false
	}
	return true
}
func (m *TCPCheck) GetPort() uint32 {
	if m != nil {
		return m.Port
	}
	var defaultValue uint32
	defaultValue = 0
	return defaultValue
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
	var defaultValue uint64
	defaultValue = 0
	return defaultValue
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
	var defaultValue uint64
	defaultValue = 0
	return defaultValue
}
func (m *TCPCheck) SetIntervalMs(value uint64) {
	if m != nil {
		m.IntervalMs = value
	}
}
func (x *TCPCheck) ToProto() *ProtoTCPCheck {
	if x == nil {
		return nil
	}

	proto := &ProtoTCPCheck{
		Port:             x.Port,
		ConnectTimeoutMs: x.ConnectTimeoutMs,
		IntervalMs:       x.IntervalMs,
	}
	return proto
}

func (x *ProtoTCPCheck) FromProto() *TCPCheck {
	if x == nil {
		return nil
	}

	copysafe := &TCPCheck{
		Port:             x.Port,
		ConnectTimeoutMs: x.ConnectTimeoutMs,
		IntervalMs:       x.IntervalMs,
	}
	return copysafe
}

func TCPCheckToProtoSlice(values []*TCPCheck) []*ProtoTCPCheck {
	if values == nil {
		return nil
	}
	result := make([]*ProtoTCPCheck, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func TCPCheckFromProtoSlice(values []*ProtoTCPCheck) []*TCPCheck {
	if values == nil {
		return nil
	}
	result := make([]*TCPCheck, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoHTTPCheck directly
type HTTPCheck struct {
	Port             uint32 `json:"port"`
	RequestTimeoutMs uint64 `json:"request_timeout_ms,omitempty"`
	Path             string `json:"path"`
	IntervalMs       uint64 `json:"interval_ms,omitempty"`
}

func (this *HTTPCheck) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*HTTPCheck)
	if !ok {
		that2, ok := that.(HTTPCheck)
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

	if this.Port != that1.Port {
		return false
	}
	if this.RequestTimeoutMs != that1.RequestTimeoutMs {
		return false
	}
	if this.Path != that1.Path {
		return false
	}
	if this.IntervalMs != that1.IntervalMs {
		return false
	}
	return true
}
func (m *HTTPCheck) GetPort() uint32 {
	if m != nil {
		return m.Port
	}
	var defaultValue uint32
	defaultValue = 0
	return defaultValue
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
	var defaultValue uint64
	defaultValue = 0
	return defaultValue
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
	var defaultValue string
	defaultValue = ""
	return defaultValue
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
	var defaultValue uint64
	defaultValue = 0
	return defaultValue
}
func (m *HTTPCheck) SetIntervalMs(value uint64) {
	if m != nil {
		m.IntervalMs = value
	}
}
func (x *HTTPCheck) ToProto() *ProtoHTTPCheck {
	if x == nil {
		return nil
	}

	proto := &ProtoHTTPCheck{
		Port:             x.Port,
		RequestTimeoutMs: x.RequestTimeoutMs,
		Path:             x.Path,
		IntervalMs:       x.IntervalMs,
	}
	return proto
}

func (x *ProtoHTTPCheck) FromProto() *HTTPCheck {
	if x == nil {
		return nil
	}

	copysafe := &HTTPCheck{
		Port:             x.Port,
		RequestTimeoutMs: x.RequestTimeoutMs,
		Path:             x.Path,
		IntervalMs:       x.IntervalMs,
	}
	return copysafe
}

func HTTPCheckToProtoSlice(values []*HTTPCheck) []*ProtoHTTPCheck {
	if values == nil {
		return nil
	}
	result := make([]*ProtoHTTPCheck, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func HTTPCheckFromProtoSlice(values []*ProtoHTTPCheck) []*HTTPCheck {
	if values == nil {
		return nil
	}
	result := make([]*HTTPCheck, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}
