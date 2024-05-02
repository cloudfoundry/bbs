// Code generated by protoc-gen-go-bbs. DO NOT EDIT.
// versions:
// - protoc-gen-go-bbs v0.0.1
// - protoc            v5.27.0--rc1
// source: events.proto

package models

// Prevent copylock errors when using ProtoActualLRPCreatedEvent directly
type ActualLRPCreatedEvent struct {
	ActualLRPGroup *ActualLRPGroup
}

func (this *ActualLRPCreatedEvent) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ActualLRPCreatedEvent)
	if !ok {
		that2, ok := that.(ActualLRPCreatedEvent)
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

	if !this.ActualLRPGroup.Equal(that1.ActualLRPGroup) {
		return false
	}
	return true
}
func (m *ActualLRPCreatedEvent) GetActualLRPGroup() *ActualLRPGroup {
	if m != nil {
		return m.ActualLRPGroup
	}
	return nil
}
func (m *ActualLRPCreatedEvent) SetActualLRPGroup(value *ActualLRPGroup) {
	if m != nil {
		m.ActualLRPGroup = value
	}
}
func (x *ActualLRPCreatedEvent) ToProto() *ProtoActualLRPCreatedEvent {
	proto := &ProtoActualLRPCreatedEvent{
		ActualLrpGroup: x.ActualLRPGroup.ToProto(),
	}
	return proto
}

func ActualLRPCreatedEventProtoMap(values []*ActualLRPCreatedEvent) []*ProtoActualLRPCreatedEvent {
	result := make([]*ProtoActualLRPCreatedEvent, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

// Prevent copylock errors when using ProtoActualLRPChangedEvent directly
type ActualLRPChangedEvent struct {
	Before *ActualLRPGroup
	After  *ActualLRPGroup
}

func (this *ActualLRPChangedEvent) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ActualLRPChangedEvent)
	if !ok {
		that2, ok := that.(ActualLRPChangedEvent)
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

	if !this.Before.Equal(that1.Before) {
		return false
	}
	if !this.After.Equal(that1.After) {
		return false
	}
	return true
}
func (m *ActualLRPChangedEvent) GetBefore() *ActualLRPGroup {
	if m != nil {
		return m.Before
	}
	return nil
}
func (m *ActualLRPChangedEvent) SetBefore(value *ActualLRPGroup) {
	if m != nil {
		m.Before = value
	}
}
func (m *ActualLRPChangedEvent) GetAfter() *ActualLRPGroup {
	if m != nil {
		return m.After
	}
	return nil
}
func (m *ActualLRPChangedEvent) SetAfter(value *ActualLRPGroup) {
	if m != nil {
		m.After = value
	}
}
func (x *ActualLRPChangedEvent) ToProto() *ProtoActualLRPChangedEvent {
	proto := &ProtoActualLRPChangedEvent{
		Before: x.Before.ToProto(),
		After:  x.After.ToProto(),
	}
	return proto
}

func ActualLRPChangedEventProtoMap(values []*ActualLRPChangedEvent) []*ProtoActualLRPChangedEvent {
	result := make([]*ProtoActualLRPChangedEvent, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

// Prevent copylock errors when using ProtoActualLRPRemovedEvent directly
type ActualLRPRemovedEvent struct {
	ActualLRPGroup *ActualLRPGroup
}

func (this *ActualLRPRemovedEvent) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ActualLRPRemovedEvent)
	if !ok {
		that2, ok := that.(ActualLRPRemovedEvent)
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

	if !this.ActualLRPGroup.Equal(that1.ActualLRPGroup) {
		return false
	}
	return true
}
func (m *ActualLRPRemovedEvent) GetActualLRPGroup() *ActualLRPGroup {
	if m != nil {
		return m.ActualLRPGroup
	}
	return nil
}
func (m *ActualLRPRemovedEvent) SetActualLRPGroup(value *ActualLRPGroup) {
	if m != nil {
		m.ActualLRPGroup = value
	}
}
func (x *ActualLRPRemovedEvent) ToProto() *ProtoActualLRPRemovedEvent {
	proto := &ProtoActualLRPRemovedEvent{
		ActualLrpGroup: x.ActualLRPGroup.ToProto(),
	}
	return proto
}

func ActualLRPRemovedEventProtoMap(values []*ActualLRPRemovedEvent) []*ProtoActualLRPRemovedEvent {
	result := make([]*ProtoActualLRPRemovedEvent, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

// Prevent copylock errors when using ProtoActualLRPInstanceCreatedEvent directly
type ActualLRPInstanceCreatedEvent struct {
	ActualLRP *ActualLRP
	TraceId   string
}

func (this *ActualLRPInstanceCreatedEvent) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ActualLRPInstanceCreatedEvent)
	if !ok {
		that2, ok := that.(ActualLRPInstanceCreatedEvent)
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

	if !this.ActualLRP.Equal(that1.ActualLRP) {
		return false
	}
	if this.TraceId != that1.TraceId {
		return false
	}
	return true
}
func (m *ActualLRPInstanceCreatedEvent) GetActualLRP() *ActualLRP {
	if m != nil {
		return m.ActualLRP
	}
	return nil
}
func (m *ActualLRPInstanceCreatedEvent) SetActualLRP(value *ActualLRP) {
	if m != nil {
		m.ActualLRP = value
	}
}
func (m *ActualLRPInstanceCreatedEvent) GetTraceId() string {
	if m != nil {
		return m.TraceId
	}
	return ""
}
func (m *ActualLRPInstanceCreatedEvent) SetTraceId(value string) {
	if m != nil {
		m.TraceId = value
	}
}
func (x *ActualLRPInstanceCreatedEvent) ToProto() *ProtoActualLRPInstanceCreatedEvent {
	proto := &ProtoActualLRPInstanceCreatedEvent{
		ActualLrp: x.ActualLRP.ToProto(),
		TraceId:   x.TraceId,
	}
	return proto
}

func ActualLRPInstanceCreatedEventProtoMap(values []*ActualLRPInstanceCreatedEvent) []*ProtoActualLRPInstanceCreatedEvent {
	result := make([]*ProtoActualLRPInstanceCreatedEvent, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

// Prevent copylock errors when using ProtoActualLRPInfo directly
type ActualLRPInfo struct {
	ActualLRPNetInfo *ActualLRPNetInfo
	CrashCount       int32
	CrashReason      string
	State            string
	PlacementError   string
	Since            int64
	ModificationTag  *ModificationTag
	Presence         ActualLRP_Presence
	Routable         *bool
	AvailabilityZone string
}

func (this *ActualLRPInfo) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ActualLRPInfo)
	if !ok {
		that2, ok := that.(ActualLRPInfo)
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

	if !this.ActualLRPNetInfo.Equal(that1.ActualLRPNetInfo) {
		return false
	}
	if this.CrashCount != that1.CrashCount {
		return false
	}
	if this.CrashReason != that1.CrashReason {
		return false
	}
	if this.State != that1.State {
		return false
	}
	if this.PlacementError != that1.PlacementError {
		return false
	}
	if this.Since != that1.Since {
		return false
	}
	if !this.ModificationTag.Equal(that1.ModificationTag) {
		return false
	}
	if this.Presence != that1.Presence {
		return false
	}
	if this.Routable != that1.Routable {
		return false
	}
	if this.AvailabilityZone != that1.AvailabilityZone {
		return false
	}
	return true
}
func (m *ActualLRPInfo) GetActualLRPNetInfo() *ActualLRPNetInfo {
	if m != nil {
		return m.ActualLRPNetInfo
	}
	return nil
}
func (m *ActualLRPInfo) SetActualLRPNetInfo(value *ActualLRPNetInfo) {
	if m != nil {
		m.ActualLRPNetInfo = value
	}
}
func (m *ActualLRPInfo) GetCrashCount() int32 {
	if m != nil {
		return m.CrashCount
	}
	return 0
}
func (m *ActualLRPInfo) SetCrashCount(value int32) {
	if m != nil {
		m.CrashCount = value
	}
}
func (m *ActualLRPInfo) GetCrashReason() string {
	if m != nil {
		return m.CrashReason
	}
	return ""
}
func (m *ActualLRPInfo) SetCrashReason(value string) {
	if m != nil {
		m.CrashReason = value
	}
}
func (m *ActualLRPInfo) GetState() string {
	if m != nil {
		return m.State
	}
	return ""
}
func (m *ActualLRPInfo) SetState(value string) {
	if m != nil {
		m.State = value
	}
}
func (m *ActualLRPInfo) GetPlacementError() string {
	if m != nil {
		return m.PlacementError
	}
	return ""
}
func (m *ActualLRPInfo) SetPlacementError(value string) {
	if m != nil {
		m.PlacementError = value
	}
}
func (m *ActualLRPInfo) GetSince() int64 {
	if m != nil {
		return m.Since
	}
	return 0
}
func (m *ActualLRPInfo) SetSince(value int64) {
	if m != nil {
		m.Since = value
	}
}
func (m *ActualLRPInfo) GetModificationTag() *ModificationTag {
	if m != nil {
		return m.ModificationTag
	}
	return nil
}
func (m *ActualLRPInfo) SetModificationTag(value *ModificationTag) {
	if m != nil {
		m.ModificationTag = value
	}
}
func (m *ActualLRPInfo) GetPresence() ActualLRP_Presence {
	if m != nil {
		return m.Presence
	}
	return 0
}
func (m *ActualLRPInfo) SetPresence(value ActualLRP_Presence) {
	if m != nil {
		m.Presence = value
	}
}
func (m *ActualLRPInfo) RoutableExists() bool {
	return m != nil && m.Routable != nil
}
func (m *ActualLRPInfo) GetRoutable() *bool {
	if m != nil && m.Routable != nil {
		return m.Routable
	}
	return nil
}
func (m *ActualLRPInfo) SetRoutable(value *bool) {
	if m != nil {
		m.Routable = value
	}
}
func (m *ActualLRPInfo) GetAvailabilityZone() string {
	if m != nil {
		return m.AvailabilityZone
	}
	return ""
}
func (m *ActualLRPInfo) SetAvailabilityZone(value string) {
	if m != nil {
		m.AvailabilityZone = value
	}
}
func (x *ActualLRPInfo) ToProto() *ProtoActualLRPInfo {
	proto := &ProtoActualLRPInfo{
		ActualLrpNetInfo: x.ActualLRPNetInfo.ToProto(),
		CrashCount:       x.CrashCount,
		CrashReason:      x.CrashReason,
		State:            x.State,
		PlacementError:   x.PlacementError,
		Since:            x.Since,
		ModificationTag:  x.ModificationTag.ToProto(),
		Presence:         ProtoActualLRPInfo_Presence(x.Presence),
		Routable:         x.Routable,
		AvailabilityZone: x.AvailabilityZone,
	}
	return proto
}

func ActualLRPInfoProtoMap(values []*ActualLRPInfo) []*ProtoActualLRPInfo {
	result := make([]*ProtoActualLRPInfo, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

// Prevent copylock errors when using ProtoActualLRPInstanceChangedEvent directly
type ActualLRPInstanceChangedEvent struct {
	ActualLRPKey         *ActualLRPKey
	ActualLRPInstanceKey *ActualLRPInstanceKey
	Before               *ActualLRPInfo
	After                *ActualLRPInfo
	TraceId              string
}

func (this *ActualLRPInstanceChangedEvent) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ActualLRPInstanceChangedEvent)
	if !ok {
		that2, ok := that.(ActualLRPInstanceChangedEvent)
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

	if !this.ActualLRPKey.Equal(that1.ActualLRPKey) {
		return false
	}
	if !this.ActualLRPInstanceKey.Equal(that1.ActualLRPInstanceKey) {
		return false
	}
	if !this.Before.Equal(that1.Before) {
		return false
	}
	if !this.After.Equal(that1.After) {
		return false
	}
	if this.TraceId != that1.TraceId {
		return false
	}
	return true
}
func (m *ActualLRPInstanceChangedEvent) GetActualLRPKey() *ActualLRPKey {
	if m != nil {
		return m.ActualLRPKey
	}
	return nil
}
func (m *ActualLRPInstanceChangedEvent) SetActualLRPKey(value *ActualLRPKey) {
	if m != nil {
		m.ActualLRPKey = value
	}
}
func (m *ActualLRPInstanceChangedEvent) GetActualLRPInstanceKey() *ActualLRPInstanceKey {
	if m != nil {
		return m.ActualLRPInstanceKey
	}
	return nil
}
func (m *ActualLRPInstanceChangedEvent) SetActualLRPInstanceKey(value *ActualLRPInstanceKey) {
	if m != nil {
		m.ActualLRPInstanceKey = value
	}
}
func (m *ActualLRPInstanceChangedEvent) GetBefore() *ActualLRPInfo {
	if m != nil {
		return m.Before
	}
	return nil
}
func (m *ActualLRPInstanceChangedEvent) SetBefore(value *ActualLRPInfo) {
	if m != nil {
		m.Before = value
	}
}
func (m *ActualLRPInstanceChangedEvent) GetAfter() *ActualLRPInfo {
	if m != nil {
		return m.After
	}
	return nil
}
func (m *ActualLRPInstanceChangedEvent) SetAfter(value *ActualLRPInfo) {
	if m != nil {
		m.After = value
	}
}
func (m *ActualLRPInstanceChangedEvent) GetTraceId() string {
	if m != nil {
		return m.TraceId
	}
	return ""
}
func (m *ActualLRPInstanceChangedEvent) SetTraceId(value string) {
	if m != nil {
		m.TraceId = value
	}
}
func (x *ActualLRPInstanceChangedEvent) ToProto() *ProtoActualLRPInstanceChangedEvent {
	proto := &ProtoActualLRPInstanceChangedEvent{
		ActualLrpKey:         x.ActualLRPKey.ToProto(),
		ActualLrpInstanceKey: x.ActualLRPInstanceKey.ToProto(),
		Before:               x.Before.ToProto(),
		After:                x.After.ToProto(),
		TraceId:              x.TraceId,
	}
	return proto
}

func ActualLRPInstanceChangedEventProtoMap(values []*ActualLRPInstanceChangedEvent) []*ProtoActualLRPInstanceChangedEvent {
	result := make([]*ProtoActualLRPInstanceChangedEvent, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

// Prevent copylock errors when using ProtoActualLRPInstanceRemovedEvent directly
type ActualLRPInstanceRemovedEvent struct {
	ActualLRP *ActualLRP
	TraceId   string
}

func (this *ActualLRPInstanceRemovedEvent) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ActualLRPInstanceRemovedEvent)
	if !ok {
		that2, ok := that.(ActualLRPInstanceRemovedEvent)
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

	if !this.ActualLRP.Equal(that1.ActualLRP) {
		return false
	}
	if this.TraceId != that1.TraceId {
		return false
	}
	return true
}
func (m *ActualLRPInstanceRemovedEvent) GetActualLRP() *ActualLRP {
	if m != nil {
		return m.ActualLRP
	}
	return nil
}
func (m *ActualLRPInstanceRemovedEvent) SetActualLRP(value *ActualLRP) {
	if m != nil {
		m.ActualLRP = value
	}
}
func (m *ActualLRPInstanceRemovedEvent) GetTraceId() string {
	if m != nil {
		return m.TraceId
	}
	return ""
}
func (m *ActualLRPInstanceRemovedEvent) SetTraceId(value string) {
	if m != nil {
		m.TraceId = value
	}
}
func (x *ActualLRPInstanceRemovedEvent) ToProto() *ProtoActualLRPInstanceRemovedEvent {
	proto := &ProtoActualLRPInstanceRemovedEvent{
		ActualLrp: x.ActualLRP.ToProto(),
		TraceId:   x.TraceId,
	}
	return proto
}

func ActualLRPInstanceRemovedEventProtoMap(values []*ActualLRPInstanceRemovedEvent) []*ProtoActualLRPInstanceRemovedEvent {
	result := make([]*ProtoActualLRPInstanceRemovedEvent, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

// Prevent copylock errors when using ProtoDesiredLRPCreatedEvent directly
type DesiredLRPCreatedEvent struct {
	DesiredLRP *DesiredLRP
	TraceId    string
}

func (this *DesiredLRPCreatedEvent) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*DesiredLRPCreatedEvent)
	if !ok {
		that2, ok := that.(DesiredLRPCreatedEvent)
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

	if !this.DesiredLRP.Equal(that1.DesiredLRP) {
		return false
	}
	if this.TraceId != that1.TraceId {
		return false
	}
	return true
}
func (m *DesiredLRPCreatedEvent) GetDesiredLRP() *DesiredLRP {
	if m != nil {
		return m.DesiredLRP
	}
	return nil
}
func (m *DesiredLRPCreatedEvent) SetDesiredLRP(value *DesiredLRP) {
	if m != nil {
		m.DesiredLRP = value
	}
}
func (m *DesiredLRPCreatedEvent) GetTraceId() string {
	if m != nil {
		return m.TraceId
	}
	return ""
}
func (m *DesiredLRPCreatedEvent) SetTraceId(value string) {
	if m != nil {
		m.TraceId = value
	}
}
func (x *DesiredLRPCreatedEvent) ToProto() *ProtoDesiredLRPCreatedEvent {
	proto := &ProtoDesiredLRPCreatedEvent{
		DesiredLrp: x.DesiredLRP.ToProto(),
		TraceId:    x.TraceId,
	}
	return proto
}

func DesiredLRPCreatedEventProtoMap(values []*DesiredLRPCreatedEvent) []*ProtoDesiredLRPCreatedEvent {
	result := make([]*ProtoDesiredLRPCreatedEvent, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

// Prevent copylock errors when using ProtoDesiredLRPChangedEvent directly
type DesiredLRPChangedEvent struct {
	Before  *DesiredLRP
	After   *DesiredLRP
	TraceId string
}

func (this *DesiredLRPChangedEvent) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*DesiredLRPChangedEvent)
	if !ok {
		that2, ok := that.(DesiredLRPChangedEvent)
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

	if !this.Before.Equal(that1.Before) {
		return false
	}
	if !this.After.Equal(that1.After) {
		return false
	}
	if this.TraceId != that1.TraceId {
		return false
	}
	return true
}
func (m *DesiredLRPChangedEvent) GetBefore() *DesiredLRP {
	if m != nil {
		return m.Before
	}
	return nil
}
func (m *DesiredLRPChangedEvent) SetBefore(value *DesiredLRP) {
	if m != nil {
		m.Before = value
	}
}
func (m *DesiredLRPChangedEvent) GetAfter() *DesiredLRP {
	if m != nil {
		return m.After
	}
	return nil
}
func (m *DesiredLRPChangedEvent) SetAfter(value *DesiredLRP) {
	if m != nil {
		m.After = value
	}
}
func (m *DesiredLRPChangedEvent) GetTraceId() string {
	if m != nil {
		return m.TraceId
	}
	return ""
}
func (m *DesiredLRPChangedEvent) SetTraceId(value string) {
	if m != nil {
		m.TraceId = value
	}
}
func (x *DesiredLRPChangedEvent) ToProto() *ProtoDesiredLRPChangedEvent {
	proto := &ProtoDesiredLRPChangedEvent{
		Before:  x.Before.ToProto(),
		After:   x.After.ToProto(),
		TraceId: x.TraceId,
	}
	return proto
}

func DesiredLRPChangedEventProtoMap(values []*DesiredLRPChangedEvent) []*ProtoDesiredLRPChangedEvent {
	result := make([]*ProtoDesiredLRPChangedEvent, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

// Prevent copylock errors when using ProtoDesiredLRPRemovedEvent directly
type DesiredLRPRemovedEvent struct {
	DesiredLRP *DesiredLRP
	TraceId    string
}

func (this *DesiredLRPRemovedEvent) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*DesiredLRPRemovedEvent)
	if !ok {
		that2, ok := that.(DesiredLRPRemovedEvent)
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

	if !this.DesiredLRP.Equal(that1.DesiredLRP) {
		return false
	}
	if this.TraceId != that1.TraceId {
		return false
	}
	return true
}
func (m *DesiredLRPRemovedEvent) GetDesiredLRP() *DesiredLRP {
	if m != nil {
		return m.DesiredLRP
	}
	return nil
}
func (m *DesiredLRPRemovedEvent) SetDesiredLRP(value *DesiredLRP) {
	if m != nil {
		m.DesiredLRP = value
	}
}
func (m *DesiredLRPRemovedEvent) GetTraceId() string {
	if m != nil {
		return m.TraceId
	}
	return ""
}
func (m *DesiredLRPRemovedEvent) SetTraceId(value string) {
	if m != nil {
		m.TraceId = value
	}
}
func (x *DesiredLRPRemovedEvent) ToProto() *ProtoDesiredLRPRemovedEvent {
	proto := &ProtoDesiredLRPRemovedEvent{
		DesiredLrp: x.DesiredLRP.ToProto(),
		TraceId:    x.TraceId,
	}
	return proto
}

func DesiredLRPRemovedEventProtoMap(values []*DesiredLRPRemovedEvent) []*ProtoDesiredLRPRemovedEvent {
	result := make([]*ProtoDesiredLRPRemovedEvent, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

// Prevent copylock errors when using ProtoActualLRPCrashedEvent directly
type ActualLRPCrashedEvent struct {
	ActualLRPKey         *ActualLRPKey
	ActualLRPInstanceKey *ActualLRPInstanceKey
	CrashCount           int32
	CrashReason          string
	Since                int64
}

func (this *ActualLRPCrashedEvent) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ActualLRPCrashedEvent)
	if !ok {
		that2, ok := that.(ActualLRPCrashedEvent)
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

	if !this.ActualLRPKey.Equal(that1.ActualLRPKey) {
		return false
	}
	if !this.ActualLRPInstanceKey.Equal(that1.ActualLRPInstanceKey) {
		return false
	}
	if this.CrashCount != that1.CrashCount {
		return false
	}
	if this.CrashReason != that1.CrashReason {
		return false
	}
	if this.Since != that1.Since {
		return false
	}
	return true
}
func (m *ActualLRPCrashedEvent) GetActualLRPKey() *ActualLRPKey {
	if m != nil {
		return m.ActualLRPKey
	}
	return nil
}
func (m *ActualLRPCrashedEvent) SetActualLRPKey(value *ActualLRPKey) {
	if m != nil {
		m.ActualLRPKey = value
	}
}
func (m *ActualLRPCrashedEvent) GetActualLRPInstanceKey() *ActualLRPInstanceKey {
	if m != nil {
		return m.ActualLRPInstanceKey
	}
	return nil
}
func (m *ActualLRPCrashedEvent) SetActualLRPInstanceKey(value *ActualLRPInstanceKey) {
	if m != nil {
		m.ActualLRPInstanceKey = value
	}
}
func (m *ActualLRPCrashedEvent) GetCrashCount() int32 {
	if m != nil {
		return m.CrashCount
	}
	return 0
}
func (m *ActualLRPCrashedEvent) SetCrashCount(value int32) {
	if m != nil {
		m.CrashCount = value
	}
}
func (m *ActualLRPCrashedEvent) GetCrashReason() string {
	if m != nil {
		return m.CrashReason
	}
	return ""
}
func (m *ActualLRPCrashedEvent) SetCrashReason(value string) {
	if m != nil {
		m.CrashReason = value
	}
}
func (m *ActualLRPCrashedEvent) GetSince() int64 {
	if m != nil {
		return m.Since
	}
	return 0
}
func (m *ActualLRPCrashedEvent) SetSince(value int64) {
	if m != nil {
		m.Since = value
	}
}
func (x *ActualLRPCrashedEvent) ToProto() *ProtoActualLRPCrashedEvent {
	proto := &ProtoActualLRPCrashedEvent{
		ActualLrpKey:         x.ActualLRPKey.ToProto(),
		ActualLrpInstanceKey: x.ActualLRPInstanceKey.ToProto(),
		CrashCount:           x.CrashCount,
		CrashReason:          x.CrashReason,
		Since:                x.Since,
	}
	return proto
}

func ActualLRPCrashedEventProtoMap(values []*ActualLRPCrashedEvent) []*ProtoActualLRPCrashedEvent {
	result := make([]*ProtoActualLRPCrashedEvent, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

// Prevent copylock errors when using ProtoEventsByCellId directly
type EventsByCellId struct {
	CellId string
}

func (this *EventsByCellId) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*EventsByCellId)
	if !ok {
		that2, ok := that.(EventsByCellId)
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

	if this.CellId != that1.CellId {
		return false
	}
	return true
}
func (m *EventsByCellId) GetCellId() string {
	if m != nil {
		return m.CellId
	}
	return ""
}
func (m *EventsByCellId) SetCellId(value string) {
	if m != nil {
		m.CellId = value
	}
}
func (x *EventsByCellId) ToProto() *ProtoEventsByCellId {
	proto := &ProtoEventsByCellId{
		CellId: x.CellId,
	}
	return proto
}

func EventsByCellIdProtoMap(values []*EventsByCellId) []*ProtoEventsByCellId {
	result := make([]*ProtoEventsByCellId, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

// Prevent copylock errors when using ProtoTaskCreatedEvent directly
type TaskCreatedEvent struct {
	Task *Task
}

func (this *TaskCreatedEvent) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*TaskCreatedEvent)
	if !ok {
		that2, ok := that.(TaskCreatedEvent)
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

	if !this.Task.Equal(that1.Task) {
		return false
	}
	return true
}
func (m *TaskCreatedEvent) GetTask() *Task {
	if m != nil {
		return m.Task
	}
	return nil
}
func (m *TaskCreatedEvent) SetTask(value *Task) {
	if m != nil {
		m.Task = value
	}
}
func (x *TaskCreatedEvent) ToProto() *ProtoTaskCreatedEvent {
	proto := &ProtoTaskCreatedEvent{
		Task: x.Task.ToProto(),
	}
	return proto
}

func TaskCreatedEventProtoMap(values []*TaskCreatedEvent) []*ProtoTaskCreatedEvent {
	result := make([]*ProtoTaskCreatedEvent, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

// Prevent copylock errors when using ProtoTaskChangedEvent directly
type TaskChangedEvent struct {
	Before *Task
	After  *Task
}

func (this *TaskChangedEvent) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*TaskChangedEvent)
	if !ok {
		that2, ok := that.(TaskChangedEvent)
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

	if !this.Before.Equal(that1.Before) {
		return false
	}
	if !this.After.Equal(that1.After) {
		return false
	}
	return true
}
func (m *TaskChangedEvent) GetBefore() *Task {
	if m != nil {
		return m.Before
	}
	return nil
}
func (m *TaskChangedEvent) SetBefore(value *Task) {
	if m != nil {
		m.Before = value
	}
}
func (m *TaskChangedEvent) GetAfter() *Task {
	if m != nil {
		return m.After
	}
	return nil
}
func (m *TaskChangedEvent) SetAfter(value *Task) {
	if m != nil {
		m.After = value
	}
}
func (x *TaskChangedEvent) ToProto() *ProtoTaskChangedEvent {
	proto := &ProtoTaskChangedEvent{
		Before: x.Before.ToProto(),
		After:  x.After.ToProto(),
	}
	return proto
}

func TaskChangedEventProtoMap(values []*TaskChangedEvent) []*ProtoTaskChangedEvent {
	result := make([]*ProtoTaskChangedEvent, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

// Prevent copylock errors when using ProtoTaskRemovedEvent directly
type TaskRemovedEvent struct {
	Task *Task
}

func (this *TaskRemovedEvent) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*TaskRemovedEvent)
	if !ok {
		that2, ok := that.(TaskRemovedEvent)
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

	if !this.Task.Equal(that1.Task) {
		return false
	}
	return true
}
func (m *TaskRemovedEvent) GetTask() *Task {
	if m != nil {
		return m.Task
	}
	return nil
}
func (m *TaskRemovedEvent) SetTask(value *Task) {
	if m != nil {
		m.Task = value
	}
}
func (x *TaskRemovedEvent) ToProto() *ProtoTaskRemovedEvent {
	proto := &ProtoTaskRemovedEvent{
		Task: x.Task.ToProto(),
	}
	return proto
}

func TaskRemovedEventProtoMap(values []*TaskRemovedEvent) []*ProtoTaskRemovedEvent {
	result := make([]*ProtoTaskRemovedEvent, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}
