// Code generated by protoc-gen-go-bbs. DO NOT EDIT.
// versions:
// - protoc-gen-go-bbs v0.0.1
// - protoc            v5.26.1
// source: security_group.proto

package models

// Prevent copylock errors when using ProtoPortRange directly
type PortRange struct {
	Start uint32
	End   uint32
}

func (this *PortRange) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*PortRange)
	if !ok {
		that2, ok := that.(PortRange)
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

	if this.Start != that1.Start {
		return false
	}
	if this.End != that1.End {
		return false
	}
	return true
}
func (m *PortRange) GetStart() uint32 {
	if m != nil {
		return m.Start
	}
	return 0
}
func (m *PortRange) SetStart(value uint32) {
	if m != nil {
		m.Start = value
	}
}
func (m *PortRange) GetEnd() uint32 {
	if m != nil {
		return m.End
	}
	return 0
}
func (m *PortRange) SetEnd(value uint32) {
	if m != nil {
		m.End = value
	}
}
func (x *PortRange) ToProto() *ProtoPortRange {
	if x == nil {
		return nil
	}

	proto := &ProtoPortRange{
		Start: x.Start,
		End:   x.End,
	}
	return proto
}

func (x *ProtoPortRange) FromProto() *PortRange {
	if x == nil {
		return nil
	}

	copysafe := &PortRange{
		Start: x.Start,
		End:   x.End,
	}
	return copysafe
}

func PortRangeToProtoSlice(values []*PortRange) []*ProtoPortRange {
	if values == nil {
		return nil
	}
	result := make([]*ProtoPortRange, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func PortRangeFromProtoSlice(values []*ProtoPortRange) []*PortRange {
	if values == nil {
		return nil
	}
	result := make([]*PortRange, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoICMPInfo directly
type ICMPInfo struct {
	Type int32
	Code int32
}

func (this *ICMPInfo) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ICMPInfo)
	if !ok {
		that2, ok := that.(ICMPInfo)
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

	if this.Type != that1.Type {
		return false
	}
	if this.Code != that1.Code {
		return false
	}
	return true
}
func (m *ICMPInfo) GetType() int32 {
	if m != nil {
		return m.Type
	}
	return 0
}
func (m *ICMPInfo) SetType(value int32) {
	if m != nil {
		m.Type = value
	}
}
func (m *ICMPInfo) GetCode() int32 {
	if m != nil {
		return m.Code
	}
	return 0
}
func (m *ICMPInfo) SetCode(value int32) {
	if m != nil {
		m.Code = value
	}
}
func (x *ICMPInfo) ToProto() *ProtoICMPInfo {
	if x == nil {
		return nil
	}

	proto := &ProtoICMPInfo{
		Type: x.Type,
		Code: x.Code,
	}
	return proto
}

func (x *ProtoICMPInfo) FromProto() *ICMPInfo {
	if x == nil {
		return nil
	}

	copysafe := &ICMPInfo{
		Type: x.Type,
		Code: x.Code,
	}
	return copysafe
}

func ICMPInfoToProtoSlice(values []*ICMPInfo) []*ProtoICMPInfo {
	if values == nil {
		return nil
	}
	result := make([]*ProtoICMPInfo, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func ICMPInfoFromProtoSlice(values []*ProtoICMPInfo) []*ICMPInfo {
	if values == nil {
		return nil
	}
	result := make([]*ICMPInfo, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoSecurityGroupRule directly
type SecurityGroupRule struct {
	Protocol     string
	Destinations []string
	Ports        []uint32
	PortRange    *PortRange
	IcmpInfo     *ICMPInfo
	Log          bool
	Annotations  []string
}

func (this *SecurityGroupRule) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*SecurityGroupRule)
	if !ok {
		that2, ok := that.(SecurityGroupRule)
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

	if this.Protocol != that1.Protocol {
		return false
	}
	if len(this.Destinations) != len(that1.Destinations) {
		return false
	}
	for i := range this.Destinations {
		if this.Destinations[i] != that1.Destinations[i] {
			return false
		}
	}
	if len(this.Ports) != len(that1.Ports) {
		return false
	}
	for i := range this.Ports {
		if this.Ports[i] != that1.Ports[i] {
			return false
		}
	}
	if !this.PortRange.Equal(*that1.PortRange) {
		return false
	}
	if !this.IcmpInfo.Equal(*that1.IcmpInfo) {
		return false
	}
	if this.Log != that1.Log {
		return false
	}
	if len(this.Annotations) != len(that1.Annotations) {
		return false
	}
	for i := range this.Annotations {
		if this.Annotations[i] != that1.Annotations[i] {
			return false
		}
	}
	return true
}
func (m *SecurityGroupRule) GetProtocol() string {
	if m != nil {
		return m.Protocol
	}
	return ""
}
func (m *SecurityGroupRule) SetProtocol(value string) {
	if m != nil {
		m.Protocol = value
	}
}
func (m *SecurityGroupRule) GetDestinations() []string {
	if m != nil {
		return m.Destinations
	}
	return nil
}
func (m *SecurityGroupRule) SetDestinations(value []string) {
	if m != nil {
		m.Destinations = value
	}
}
func (m *SecurityGroupRule) GetPorts() []uint32 {
	if m != nil {
		return m.Ports
	}
	return nil
}
func (m *SecurityGroupRule) SetPorts(value []uint32) {
	if m != nil {
		m.Ports = value
	}
}
func (m *SecurityGroupRule) GetPortRange() *PortRange {
	if m != nil {
		return m.PortRange
	}
	return nil
}
func (m *SecurityGroupRule) SetPortRange(value *PortRange) {
	if m != nil {
		m.PortRange = value
	}
}
func (m *SecurityGroupRule) GetIcmpInfo() *ICMPInfo {
	if m != nil {
		return m.IcmpInfo
	}
	return nil
}
func (m *SecurityGroupRule) SetIcmpInfo(value *ICMPInfo) {
	if m != nil {
		m.IcmpInfo = value
	}
}
func (m *SecurityGroupRule) GetLog() bool {
	if m != nil {
		return m.Log
	}
	return false
}
func (m *SecurityGroupRule) SetLog(value bool) {
	if m != nil {
		m.Log = value
	}
}
func (m *SecurityGroupRule) GetAnnotations() []string {
	if m != nil {
		return m.Annotations
	}
	return nil
}
func (m *SecurityGroupRule) SetAnnotations(value []string) {
	if m != nil {
		m.Annotations = value
	}
}
func (x *SecurityGroupRule) ToProto() *ProtoSecurityGroupRule {
	if x == nil {
		return nil
	}

	proto := &ProtoSecurityGroupRule{
		Protocol:     x.Protocol,
		Destinations: x.Destinations,
		Ports:        x.Ports,
		PortRange:    x.PortRange.ToProto(),
		IcmpInfo:     x.IcmpInfo.ToProto(),
		Log:          x.Log,
		Annotations:  x.Annotations,
	}
	return proto
}

func (x *ProtoSecurityGroupRule) FromProto() *SecurityGroupRule {
	if x == nil {
		return nil
	}

	copysafe := &SecurityGroupRule{
		Protocol:     x.Protocol,
		Destinations: x.Destinations,
		Ports:        x.Ports,
		PortRange:    x.PortRange.FromProto(),
		IcmpInfo:     x.IcmpInfo.FromProto(),
		Log:          x.Log,
		Annotations:  x.Annotations,
	}
	return copysafe
}

func SecurityGroupRuleToProtoSlice(values []*SecurityGroupRule) []*ProtoSecurityGroupRule {
	if values == nil {
		return nil
	}
	result := make([]*ProtoSecurityGroupRule, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func SecurityGroupRuleFromProtoSlice(values []*ProtoSecurityGroupRule) []*SecurityGroupRule {
	if values == nil {
		return nil
	}
	result := make([]*SecurityGroupRule, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}
