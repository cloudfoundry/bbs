// Code generated by protoc-gen-go-bbs. DO NOT EDIT.
// versions:
// - protoc-gen-go-bbs v0.0.1
// - protoc            v5.27.0--rc1
// source: actual_lrp_requests.proto

package models

// Prevent copylock errors when using ProtoActualLRPLifecycleResponse directly
type ActualLRPLifecycleResponse struct {
	Error *Error
}

func (this *ActualLRPLifecycleResponse) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ActualLRPLifecycleResponse)
	if !ok {
		that2, ok := that.(ActualLRPLifecycleResponse)
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

	if !this.Error.Equal(that1.Error) {
		return false
	}
	return true
}
func (m *ActualLRPLifecycleResponse) GetError() *Error {
	if m != nil {
		return m.Error
	}
	return nil
}
func (m *ActualLRPLifecycleResponse) SetError(value *Error) {
	if m != nil {
		m.Error = value
	}
}
func (x *ActualLRPLifecycleResponse) ToProto() *ProtoActualLRPLifecycleResponse {
	proto := &ProtoActualLRPLifecycleResponse{
		Error: x.Error.ToProto(),
	}
	return proto
}

func ActualLRPLifecycleResponseProtoMap(values []*ActualLRPLifecycleResponse) []*ProtoActualLRPLifecycleResponse {
	result := make([]*ProtoActualLRPLifecycleResponse, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

// Prevent copylock errors when using ProtoActualLRPGroupsResponse directly
type ActualLRPGroupsResponse struct {
	Error           *Error
	ActualLrpGroups []*ActualLRPGroup
}

func (this *ActualLRPGroupsResponse) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ActualLRPGroupsResponse)
	if !ok {
		that2, ok := that.(ActualLRPGroupsResponse)
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

	if !this.Error.Equal(that1.Error) {
		return false
	}
	if len(this.ActualLrpGroups) != len(that1.ActualLrpGroups) {
		return false
	}
	for i := range this.ActualLrpGroups {
		if !this.ActualLrpGroups[i].Equal(that1.ActualLrpGroups[i]) {
			return false
		}
	}
	return true
}
func (m *ActualLRPGroupsResponse) GetError() *Error {
	if m != nil {
		return m.Error
	}
	return nil
}
func (m *ActualLRPGroupsResponse) SetError(value *Error) {
	if m != nil {
		m.Error = value
	}
}
func (m *ActualLRPGroupsResponse) GetActualLrpGroups() []*ActualLRPGroup {
	if m != nil {
		return m.ActualLrpGroups
	}
	return nil
}
func (m *ActualLRPGroupsResponse) SetActualLrpGroups(value []*ActualLRPGroup) {
	if m != nil {
		m.ActualLrpGroups = value
	}
}
func (x *ActualLRPGroupsResponse) ToProto() *ProtoActualLRPGroupsResponse {
	proto := &ProtoActualLRPGroupsResponse{
		Error:           x.Error.ToProto(),
		ActualLrpGroups: ActualLRPGroupProtoMap(x.ActualLrpGroups),
	}
	return proto
}

func ActualLRPGroupsResponseProtoMap(values []*ActualLRPGroupsResponse) []*ProtoActualLRPGroupsResponse {
	result := make([]*ProtoActualLRPGroupsResponse, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

// Prevent copylock errors when using ProtoActualLRPGroupResponse directly
type ActualLRPGroupResponse struct {
	Error          *Error
	ActualLrpGroup *ActualLRPGroup
}

func (this *ActualLRPGroupResponse) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ActualLRPGroupResponse)
	if !ok {
		that2, ok := that.(ActualLRPGroupResponse)
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

	if !this.Error.Equal(that1.Error) {
		return false
	}
	if !this.ActualLrpGroup.Equal(that1.ActualLrpGroup) {
		return false
	}
	return true
}
func (m *ActualLRPGroupResponse) GetError() *Error {
	if m != nil {
		return m.Error
	}
	return nil
}
func (m *ActualLRPGroupResponse) SetError(value *Error) {
	if m != nil {
		m.Error = value
	}
}
func (m *ActualLRPGroupResponse) GetActualLrpGroup() *ActualLRPGroup {
	if m != nil {
		return m.ActualLrpGroup
	}
	return nil
}
func (m *ActualLRPGroupResponse) SetActualLrpGroup(value *ActualLRPGroup) {
	if m != nil {
		m.ActualLrpGroup = value
	}
}
func (x *ActualLRPGroupResponse) ToProto() *ProtoActualLRPGroupResponse {
	proto := &ProtoActualLRPGroupResponse{
		Error:          x.Error.ToProto(),
		ActualLrpGroup: x.ActualLrpGroup.ToProto(),
	}
	return proto
}

func ActualLRPGroupResponseProtoMap(values []*ActualLRPGroupResponse) []*ProtoActualLRPGroupResponse {
	result := make([]*ProtoActualLRPGroupResponse, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

// Prevent copylock errors when using ProtoActualLRPGroupsRequest directly
type ActualLRPGroupsRequest struct {
	Domain string
	CellId string
}

func (this *ActualLRPGroupsRequest) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ActualLRPGroupsRequest)
	if !ok {
		that2, ok := that.(ActualLRPGroupsRequest)
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

	if this.Domain != that1.Domain {
		return false
	}
	if this.CellId != that1.CellId {
		return false
	}
	return true
}
func (m *ActualLRPGroupsRequest) GetDomain() string {
	if m != nil {
		return m.Domain
	}
	return ""
}
func (m *ActualLRPGroupsRequest) SetDomain(value string) {
	if m != nil {
		m.Domain = value
	}
}
func (m *ActualLRPGroupsRequest) GetCellId() string {
	if m != nil {
		return m.CellId
	}
	return ""
}
func (m *ActualLRPGroupsRequest) SetCellId(value string) {
	if m != nil {
		m.CellId = value
	}
}
func (x *ActualLRPGroupsRequest) ToProto() *ProtoActualLRPGroupsRequest {
	proto := &ProtoActualLRPGroupsRequest{
		Domain: x.Domain,
		CellId: x.CellId,
	}
	return proto
}

func ActualLRPGroupsRequestProtoMap(values []*ActualLRPGroupsRequest) []*ProtoActualLRPGroupsRequest {
	result := make([]*ProtoActualLRPGroupsRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

// Prevent copylock errors when using ProtoActualLRPGroupsByProcessGuidRequest directly
type ActualLRPGroupsByProcessGuidRequest struct {
	ProcessGuid string
}

func (this *ActualLRPGroupsByProcessGuidRequest) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ActualLRPGroupsByProcessGuidRequest)
	if !ok {
		that2, ok := that.(ActualLRPGroupsByProcessGuidRequest)
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

	if this.ProcessGuid != that1.ProcessGuid {
		return false
	}
	return true
}
func (m *ActualLRPGroupsByProcessGuidRequest) GetProcessGuid() string {
	if m != nil {
		return m.ProcessGuid
	}
	return ""
}
func (m *ActualLRPGroupsByProcessGuidRequest) SetProcessGuid(value string) {
	if m != nil {
		m.ProcessGuid = value
	}
}
func (x *ActualLRPGroupsByProcessGuidRequest) ToProto() *ProtoActualLRPGroupsByProcessGuidRequest {
	proto := &ProtoActualLRPGroupsByProcessGuidRequest{
		ProcessGuid: x.ProcessGuid,
	}
	return proto
}

func ActualLRPGroupsByProcessGuidRequestProtoMap(values []*ActualLRPGroupsByProcessGuidRequest) []*ProtoActualLRPGroupsByProcessGuidRequest {
	result := make([]*ProtoActualLRPGroupsByProcessGuidRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

// Prevent copylock errors when using ProtoActualLRPGroupByProcessGuidAndIndexRequest directly
type ActualLRPGroupByProcessGuidAndIndexRequest struct {
	ProcessGuid string
	Index       int32
}

func (this *ActualLRPGroupByProcessGuidAndIndexRequest) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ActualLRPGroupByProcessGuidAndIndexRequest)
	if !ok {
		that2, ok := that.(ActualLRPGroupByProcessGuidAndIndexRequest)
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

	if this.ProcessGuid != that1.ProcessGuid {
		return false
	}
	if this.Index != that1.Index {
		return false
	}
	return true
}
func (m *ActualLRPGroupByProcessGuidAndIndexRequest) GetProcessGuid() string {
	if m != nil {
		return m.ProcessGuid
	}
	return ""
}
func (m *ActualLRPGroupByProcessGuidAndIndexRequest) SetProcessGuid(value string) {
	if m != nil {
		m.ProcessGuid = value
	}
}
func (m *ActualLRPGroupByProcessGuidAndIndexRequest) GetIndex() int32 {
	if m != nil {
		return m.Index
	}
	return 0
}
func (m *ActualLRPGroupByProcessGuidAndIndexRequest) SetIndex(value int32) {
	if m != nil {
		m.Index = value
	}
}
func (x *ActualLRPGroupByProcessGuidAndIndexRequest) ToProto() *ProtoActualLRPGroupByProcessGuidAndIndexRequest {
	proto := &ProtoActualLRPGroupByProcessGuidAndIndexRequest{
		ProcessGuid: x.ProcessGuid,
		Index:       x.Index,
	}
	return proto
}

func ActualLRPGroupByProcessGuidAndIndexRequestProtoMap(values []*ActualLRPGroupByProcessGuidAndIndexRequest) []*ProtoActualLRPGroupByProcessGuidAndIndexRequest {
	result := make([]*ProtoActualLRPGroupByProcessGuidAndIndexRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

// Prevent copylock errors when using ProtoClaimActualLRPRequest directly
type ClaimActualLRPRequest struct {
	ProcessGuid          string
	Index                int32
	ActualLrpInstanceKey *ActualLRPInstanceKey
}

func (this *ClaimActualLRPRequest) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ClaimActualLRPRequest)
	if !ok {
		that2, ok := that.(ClaimActualLRPRequest)
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

	if this.ProcessGuid != that1.ProcessGuid {
		return false
	}
	if this.Index != that1.Index {
		return false
	}
	if !this.ActualLrpInstanceKey.Equal(that1.ActualLrpInstanceKey) {
		return false
	}
	return true
}
func (m *ClaimActualLRPRequest) GetProcessGuid() string {
	if m != nil {
		return m.ProcessGuid
	}
	return ""
}
func (m *ClaimActualLRPRequest) SetProcessGuid(value string) {
	if m != nil {
		m.ProcessGuid = value
	}
}
func (m *ClaimActualLRPRequest) GetIndex() int32 {
	if m != nil {
		return m.Index
	}
	return 0
}
func (m *ClaimActualLRPRequest) SetIndex(value int32) {
	if m != nil {
		m.Index = value
	}
}
func (m *ClaimActualLRPRequest) GetActualLrpInstanceKey() *ActualLRPInstanceKey {
	if m != nil {
		return m.ActualLrpInstanceKey
	}
	return nil
}
func (m *ClaimActualLRPRequest) SetActualLrpInstanceKey(value *ActualLRPInstanceKey) {
	if m != nil {
		m.ActualLrpInstanceKey = value
	}
}
func (x *ClaimActualLRPRequest) ToProto() *ProtoClaimActualLRPRequest {
	proto := &ProtoClaimActualLRPRequest{
		ProcessGuid:          x.ProcessGuid,
		Index:                x.Index,
		ActualLrpInstanceKey: x.ActualLrpInstanceKey.ToProto(),
	}
	return proto
}

func ClaimActualLRPRequestProtoMap(values []*ClaimActualLRPRequest) []*ProtoClaimActualLRPRequest {
	result := make([]*ProtoClaimActualLRPRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

// Prevent copylock errors when using ProtoStartActualLRPRequest directly
type StartActualLRPRequest struct {
	ActualLrpKey            *ActualLRPKey
	ActualLrpInstanceKey    *ActualLRPInstanceKey
	ActualLrpNetInfo        *ActualLRPNetInfo
	ActualLrpInternalRoutes []*ActualLRPInternalRoute
	MetricTags              map[string]string
	Routable                *bool
	AvailabilityZone        string
}

func (this *StartActualLRPRequest) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*StartActualLRPRequest)
	if !ok {
		that2, ok := that.(StartActualLRPRequest)
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

	if !this.ActualLrpKey.Equal(that1.ActualLrpKey) {
		return false
	}
	if !this.ActualLrpInstanceKey.Equal(that1.ActualLrpInstanceKey) {
		return false
	}
	if !this.ActualLrpNetInfo.Equal(that1.ActualLrpNetInfo) {
		return false
	}
	if len(this.ActualLrpInternalRoutes) != len(that1.ActualLrpInternalRoutes) {
		return false
	}
	for i := range this.ActualLrpInternalRoutes {
		if !this.ActualLrpInternalRoutes[i].Equal(that1.ActualLrpInternalRoutes[i]) {
			return false
		}
	}
	if len(this.MetricTags) != len(that1.MetricTags) {
		return false
	}
	for i := range this.MetricTags {
		if this.MetricTags[i] != that1.MetricTags[i] {
			return false
		}
	}
	if this.Routable != that1.Routable {
		return false
	}
	if this.AvailabilityZone != that1.AvailabilityZone {
		return false
	}
	return true
}
func (m *StartActualLRPRequest) GetActualLrpKey() *ActualLRPKey {
	if m != nil {
		return m.ActualLrpKey
	}
	return nil
}
func (m *StartActualLRPRequest) SetActualLrpKey(value *ActualLRPKey) {
	if m != nil {
		m.ActualLrpKey = value
	}
}
func (m *StartActualLRPRequest) GetActualLrpInstanceKey() *ActualLRPInstanceKey {
	if m != nil {
		return m.ActualLrpInstanceKey
	}
	return nil
}
func (m *StartActualLRPRequest) SetActualLrpInstanceKey(value *ActualLRPInstanceKey) {
	if m != nil {
		m.ActualLrpInstanceKey = value
	}
}
func (m *StartActualLRPRequest) GetActualLrpNetInfo() *ActualLRPNetInfo {
	if m != nil {
		return m.ActualLrpNetInfo
	}
	return nil
}
func (m *StartActualLRPRequest) SetActualLrpNetInfo(value *ActualLRPNetInfo) {
	if m != nil {
		m.ActualLrpNetInfo = value
	}
}
func (m *StartActualLRPRequest) GetActualLrpInternalRoutes() []*ActualLRPInternalRoute {
	if m != nil {
		return m.ActualLrpInternalRoutes
	}
	return nil
}
func (m *StartActualLRPRequest) SetActualLrpInternalRoutes(value []*ActualLRPInternalRoute) {
	if m != nil {
		m.ActualLrpInternalRoutes = value
	}
}
func (m *StartActualLRPRequest) GetMetricTags() map[string]string {
	if m != nil {
		return m.MetricTags
	}
	return nil
}
func (m *StartActualLRPRequest) SetMetricTags(value map[string]string) {
	if m != nil {
		m.MetricTags = value
	}
}
func (m *StartActualLRPRequest) RoutableExists() bool {
	return m != nil && m.Routable != nil
}
func (m *StartActualLRPRequest) GetRoutable() *bool {
	if m != nil && m.Routable != nil {
		return m.Routable
	}
	return nil
}
func (m *StartActualLRPRequest) SetRoutable(value *bool) {
	if m != nil {
		m.Routable = value
	}
}
func (m *StartActualLRPRequest) GetAvailabilityZone() string {
	if m != nil {
		return m.AvailabilityZone
	}
	return ""
}
func (m *StartActualLRPRequest) SetAvailabilityZone(value string) {
	if m != nil {
		m.AvailabilityZone = value
	}
}
func (x *StartActualLRPRequest) ToProto() *ProtoStartActualLRPRequest {
	proto := &ProtoStartActualLRPRequest{
		ActualLrpKey:            x.ActualLrpKey.ToProto(),
		ActualLrpInstanceKey:    x.ActualLrpInstanceKey.ToProto(),
		ActualLrpNetInfo:        x.ActualLrpNetInfo.ToProto(),
		ActualLrpInternalRoutes: ActualLRPInternalRouteProtoMap(x.ActualLrpInternalRoutes),
		MetricTags:              x.MetricTags,
		Routable:                x.Routable,
		AvailabilityZone:        x.AvailabilityZone,
	}
	return proto
}

func StartActualLRPRequestProtoMap(values []*StartActualLRPRequest) []*ProtoStartActualLRPRequest {
	result := make([]*ProtoStartActualLRPRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

// Prevent copylock errors when using ProtoCrashActualLRPRequest directly
type CrashActualLRPRequest struct {
	ActualLrpKey         *ActualLRPKey
	ActualLrpInstanceKey *ActualLRPInstanceKey
	ErrorMessage         string
}

func (this *CrashActualLRPRequest) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*CrashActualLRPRequest)
	if !ok {
		that2, ok := that.(CrashActualLRPRequest)
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

	if !this.ActualLrpKey.Equal(that1.ActualLrpKey) {
		return false
	}
	if !this.ActualLrpInstanceKey.Equal(that1.ActualLrpInstanceKey) {
		return false
	}
	if this.ErrorMessage != that1.ErrorMessage {
		return false
	}
	return true
}
func (m *CrashActualLRPRequest) GetActualLrpKey() *ActualLRPKey {
	if m != nil {
		return m.ActualLrpKey
	}
	return nil
}
func (m *CrashActualLRPRequest) SetActualLrpKey(value *ActualLRPKey) {
	if m != nil {
		m.ActualLrpKey = value
	}
}
func (m *CrashActualLRPRequest) GetActualLrpInstanceKey() *ActualLRPInstanceKey {
	if m != nil {
		return m.ActualLrpInstanceKey
	}
	return nil
}
func (m *CrashActualLRPRequest) SetActualLrpInstanceKey(value *ActualLRPInstanceKey) {
	if m != nil {
		m.ActualLrpInstanceKey = value
	}
}
func (m *CrashActualLRPRequest) GetErrorMessage() string {
	if m != nil {
		return m.ErrorMessage
	}
	return ""
}
func (m *CrashActualLRPRequest) SetErrorMessage(value string) {
	if m != nil {
		m.ErrorMessage = value
	}
}
func (x *CrashActualLRPRequest) ToProto() *ProtoCrashActualLRPRequest {
	proto := &ProtoCrashActualLRPRequest{
		ActualLrpKey:         x.ActualLrpKey.ToProto(),
		ActualLrpInstanceKey: x.ActualLrpInstanceKey.ToProto(),
		ErrorMessage:         x.ErrorMessage,
	}
	return proto
}

func CrashActualLRPRequestProtoMap(values []*CrashActualLRPRequest) []*ProtoCrashActualLRPRequest {
	result := make([]*ProtoCrashActualLRPRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

// Prevent copylock errors when using ProtoFailActualLRPRequest directly
type FailActualLRPRequest struct {
	ActualLrpKey *ActualLRPKey
	ErrorMessage string
}

func (this *FailActualLRPRequest) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*FailActualLRPRequest)
	if !ok {
		that2, ok := that.(FailActualLRPRequest)
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

	if !this.ActualLrpKey.Equal(that1.ActualLrpKey) {
		return false
	}
	if this.ErrorMessage != that1.ErrorMessage {
		return false
	}
	return true
}
func (m *FailActualLRPRequest) GetActualLrpKey() *ActualLRPKey {
	if m != nil {
		return m.ActualLrpKey
	}
	return nil
}
func (m *FailActualLRPRequest) SetActualLrpKey(value *ActualLRPKey) {
	if m != nil {
		m.ActualLrpKey = value
	}
}
func (m *FailActualLRPRequest) GetErrorMessage() string {
	if m != nil {
		return m.ErrorMessage
	}
	return ""
}
func (m *FailActualLRPRequest) SetErrorMessage(value string) {
	if m != nil {
		m.ErrorMessage = value
	}
}
func (x *FailActualLRPRequest) ToProto() *ProtoFailActualLRPRequest {
	proto := &ProtoFailActualLRPRequest{
		ActualLrpKey: x.ActualLrpKey.ToProto(),
		ErrorMessage: x.ErrorMessage,
	}
	return proto
}

func FailActualLRPRequestProtoMap(values []*FailActualLRPRequest) []*ProtoFailActualLRPRequest {
	result := make([]*ProtoFailActualLRPRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

// Prevent copylock errors when using ProtoRetireActualLRPRequest directly
type RetireActualLRPRequest struct {
	ActualLrpKey *ActualLRPKey
}

func (this *RetireActualLRPRequest) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*RetireActualLRPRequest)
	if !ok {
		that2, ok := that.(RetireActualLRPRequest)
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

	if !this.ActualLrpKey.Equal(that1.ActualLrpKey) {
		return false
	}
	return true
}
func (m *RetireActualLRPRequest) GetActualLrpKey() *ActualLRPKey {
	if m != nil {
		return m.ActualLrpKey
	}
	return nil
}
func (m *RetireActualLRPRequest) SetActualLrpKey(value *ActualLRPKey) {
	if m != nil {
		m.ActualLrpKey = value
	}
}
func (x *RetireActualLRPRequest) ToProto() *ProtoRetireActualLRPRequest {
	proto := &ProtoRetireActualLRPRequest{
		ActualLrpKey: x.ActualLrpKey.ToProto(),
	}
	return proto
}

func RetireActualLRPRequestProtoMap(values []*RetireActualLRPRequest) []*ProtoRetireActualLRPRequest {
	result := make([]*ProtoRetireActualLRPRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

// Prevent copylock errors when using ProtoRemoveActualLRPRequest directly
type RemoveActualLRPRequest struct {
	ProcessGuid          string
	Index                int32
	ActualLrpInstanceKey *ActualLRPInstanceKey
}

func (this *RemoveActualLRPRequest) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*RemoveActualLRPRequest)
	if !ok {
		that2, ok := that.(RemoveActualLRPRequest)
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

	if this.ProcessGuid != that1.ProcessGuid {
		return false
	}
	if this.Index != that1.Index {
		return false
	}
	if !this.ActualLrpInstanceKey.Equal(that1.ActualLrpInstanceKey) {
		return false
	}
	return true
}
func (m *RemoveActualLRPRequest) GetProcessGuid() string {
	if m != nil {
		return m.ProcessGuid
	}
	return ""
}
func (m *RemoveActualLRPRequest) SetProcessGuid(value string) {
	if m != nil {
		m.ProcessGuid = value
	}
}
func (m *RemoveActualLRPRequest) GetIndex() int32 {
	if m != nil {
		return m.Index
	}
	return 0
}
func (m *RemoveActualLRPRequest) SetIndex(value int32) {
	if m != nil {
		m.Index = value
	}
}
func (m *RemoveActualLRPRequest) GetActualLrpInstanceKey() *ActualLRPInstanceKey {
	if m != nil {
		return m.ActualLrpInstanceKey
	}
	return nil
}
func (m *RemoveActualLRPRequest) SetActualLrpInstanceKey(value *ActualLRPInstanceKey) {
	if m != nil {
		m.ActualLrpInstanceKey = value
	}
}
func (x *RemoveActualLRPRequest) ToProto() *ProtoRemoveActualLRPRequest {
	proto := &ProtoRemoveActualLRPRequest{
		ProcessGuid:          x.ProcessGuid,
		Index:                x.Index,
		ActualLrpInstanceKey: x.ActualLrpInstanceKey.ToProto(),
	}
	return proto
}

func RemoveActualLRPRequestProtoMap(values []*RemoveActualLRPRequest) []*ProtoRemoveActualLRPRequest {
	result := make([]*ProtoRemoveActualLRPRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

// Prevent copylock errors when using ProtoActualLRPsResponse directly
type ActualLRPsResponse struct {
	Error      *Error
	ActualLrps []*ActualLRP
}

func (this *ActualLRPsResponse) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ActualLRPsResponse)
	if !ok {
		that2, ok := that.(ActualLRPsResponse)
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

	if !this.Error.Equal(that1.Error) {
		return false
	}
	if len(this.ActualLrps) != len(that1.ActualLrps) {
		return false
	}
	for i := range this.ActualLrps {
		if !this.ActualLrps[i].Equal(that1.ActualLrps[i]) {
			return false
		}
	}
	return true
}
func (m *ActualLRPsResponse) GetError() *Error {
	if m != nil {
		return m.Error
	}
	return nil
}
func (m *ActualLRPsResponse) SetError(value *Error) {
	if m != nil {
		m.Error = value
	}
}
func (m *ActualLRPsResponse) GetActualLrps() []*ActualLRP {
	if m != nil {
		return m.ActualLrps
	}
	return nil
}
func (m *ActualLRPsResponse) SetActualLrps(value []*ActualLRP) {
	if m != nil {
		m.ActualLrps = value
	}
}
func (x *ActualLRPsResponse) ToProto() *ProtoActualLRPsResponse {
	proto := &ProtoActualLRPsResponse{
		Error:      x.Error.ToProto(),
		ActualLrps: ActualLRPProtoMap(x.ActualLrps),
	}
	return proto
}

func ActualLRPsResponseProtoMap(values []*ActualLRPsResponse) []*ProtoActualLRPsResponse {
	result := make([]*ProtoActualLRPsResponse, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

// Prevent copylock errors when using ProtoActualLRPsRequest directly
type ActualLRPsRequest struct {
	Domain      string
	CellId      string
	ProcessGuid string
	Index       *int32
}

func (this *ActualLRPsRequest) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ActualLRPsRequest)
	if !ok {
		that2, ok := that.(ActualLRPsRequest)
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

	if this.Domain != that1.Domain {
		return false
	}
	if this.CellId != that1.CellId {
		return false
	}
	if this.ProcessGuid != that1.ProcessGuid {
		return false
	}
	if this.Index != that1.Index {
		return false
	}
	return true
}
func (m *ActualLRPsRequest) GetDomain() string {
	if m != nil {
		return m.Domain
	}
	return ""
}
func (m *ActualLRPsRequest) SetDomain(value string) {
	if m != nil {
		m.Domain = value
	}
}
func (m *ActualLRPsRequest) GetCellId() string {
	if m != nil {
		return m.CellId
	}
	return ""
}
func (m *ActualLRPsRequest) SetCellId(value string) {
	if m != nil {
		m.CellId = value
	}
}
func (m *ActualLRPsRequest) GetProcessGuid() string {
	if m != nil {
		return m.ProcessGuid
	}
	return ""
}
func (m *ActualLRPsRequest) SetProcessGuid(value string) {
	if m != nil {
		m.ProcessGuid = value
	}
}
func (m *ActualLRPsRequest) IndexExists() bool {
	return m != nil && m.Index != nil
}
func (m *ActualLRPsRequest) GetIndex() *int32 {
	if m != nil && m.Index != nil {
		return m.Index
	}
	return nil
}
func (m *ActualLRPsRequest) SetIndex(value *int32) {
	if m != nil {
		m.Index = value
	}
}
func (x *ActualLRPsRequest) ToProto() *ProtoActualLRPsRequest {
	proto := &ProtoActualLRPsRequest{
		Domain:      x.Domain,
		CellId:      x.CellId,
		ProcessGuid: x.ProcessGuid,
		Index:       x.Index,
	}
	return proto
}

func ActualLRPsRequestProtoMap(values []*ActualLRPsRequest) []*ProtoActualLRPsRequest {
	result := make([]*ProtoActualLRPsRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}
