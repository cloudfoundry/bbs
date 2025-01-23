// Code generated by protoc-gen-go-bbs. DO NOT EDIT.
// versions:
// - protoc-gen-go-bbs v0.0.1
// - protoc            v4.25.6
// source: actual_lrp_requests.proto

package models

// Prevent copylock errors when using ProtoActualLRPLifecycleResponse directly
type ActualLRPLifecycleResponse struct {
	Error *Error `json:"error,omitempty"`
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

	if this.Error == nil {
		if that1.Error != nil {
			return false
		}
	} else if !this.Error.Equal(*that1.Error) {
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
	if x == nil {
		return nil
	}

	proto := &ProtoActualLRPLifecycleResponse{
		Error: x.Error.ToProto(),
	}
	return proto
}

func (x *ProtoActualLRPLifecycleResponse) FromProto() *ActualLRPLifecycleResponse {
	if x == nil {
		return nil
	}

	copysafe := &ActualLRPLifecycleResponse{
		Error: x.Error.FromProto(),
	}
	return copysafe
}

func ActualLRPLifecycleResponseToProtoSlice(values []*ActualLRPLifecycleResponse) []*ProtoActualLRPLifecycleResponse {
	if values == nil {
		return nil
	}
	result := make([]*ProtoActualLRPLifecycleResponse, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func ActualLRPLifecycleResponseFromProtoSlice(values []*ProtoActualLRPLifecycleResponse) []*ActualLRPLifecycleResponse {
	if values == nil {
		return nil
	}
	result := make([]*ActualLRPLifecycleResponse, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Deprecated: marked deprecated in actual_lrp_requests.proto
// Prevent copylock errors when using ProtoActualLRPGroupsResponse directly
type ActualLRPGroupsResponse struct {
	Error           *Error            `json:"error,omitempty"`
	ActualLrpGroups []*ActualLRPGroup `json:"actual_lrp_groups,omitempty"`
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

	if this.Error == nil {
		if that1.Error != nil {
			return false
		}
	} else if !this.Error.Equal(*that1.Error) {
		return false
	}
	if this.ActualLrpGroups == nil {
		if that1.ActualLrpGroups != nil {
			return false
		}
	} else if len(this.ActualLrpGroups) != len(that1.ActualLrpGroups) {
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
	if x == nil {
		return nil
	}

	proto := &ProtoActualLRPGroupsResponse{
		Error:           x.Error.ToProto(),
		ActualLrpGroups: ActualLRPGroupToProtoSlice(x.ActualLrpGroups),
	}
	return proto
}

func (x *ProtoActualLRPGroupsResponse) FromProto() *ActualLRPGroupsResponse {
	if x == nil {
		return nil
	}

	copysafe := &ActualLRPGroupsResponse{
		Error:           x.Error.FromProto(),
		ActualLrpGroups: ActualLRPGroupFromProtoSlice(x.ActualLrpGroups),
	}
	return copysafe
}

func ActualLRPGroupsResponseToProtoSlice(values []*ActualLRPGroupsResponse) []*ProtoActualLRPGroupsResponse {
	if values == nil {
		return nil
	}
	result := make([]*ProtoActualLRPGroupsResponse, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func ActualLRPGroupsResponseFromProtoSlice(values []*ProtoActualLRPGroupsResponse) []*ActualLRPGroupsResponse {
	if values == nil {
		return nil
	}
	result := make([]*ActualLRPGroupsResponse, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Deprecated: marked deprecated in actual_lrp_requests.proto
// Prevent copylock errors when using ProtoActualLRPGroupResponse directly
type ActualLRPGroupResponse struct {
	Error          *Error          `json:"error,omitempty"`
	ActualLrpGroup *ActualLRPGroup `json:"actual_lrp_group,omitempty"`
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

	if this.Error == nil {
		if that1.Error != nil {
			return false
		}
	} else if !this.Error.Equal(*that1.Error) {
		return false
	}
	if this.ActualLrpGroup == nil {
		if that1.ActualLrpGroup != nil {
			return false
		}
	} else if !this.ActualLrpGroup.Equal(*that1.ActualLrpGroup) {
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
	if x == nil {
		return nil
	}

	proto := &ProtoActualLRPGroupResponse{
		Error:          x.Error.ToProto(),
		ActualLrpGroup: x.ActualLrpGroup.ToProto(),
	}
	return proto
}

func (x *ProtoActualLRPGroupResponse) FromProto() *ActualLRPGroupResponse {
	if x == nil {
		return nil
	}

	copysafe := &ActualLRPGroupResponse{
		Error:          x.Error.FromProto(),
		ActualLrpGroup: x.ActualLrpGroup.FromProto(),
	}
	return copysafe
}

func ActualLRPGroupResponseToProtoSlice(values []*ActualLRPGroupResponse) []*ProtoActualLRPGroupResponse {
	if values == nil {
		return nil
	}
	result := make([]*ProtoActualLRPGroupResponse, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func ActualLRPGroupResponseFromProtoSlice(values []*ProtoActualLRPGroupResponse) []*ActualLRPGroupResponse {
	if values == nil {
		return nil
	}
	result := make([]*ActualLRPGroupResponse, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Deprecated: marked deprecated in actual_lrp_requests.proto
// Prevent copylock errors when using ProtoActualLRPGroupsRequest directly
type ActualLRPGroupsRequest struct {
	Domain string `json:"domain"`
	CellId string `json:"cell_id"`
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
	var defaultValue string
	defaultValue = ""
	return defaultValue
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
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *ActualLRPGroupsRequest) SetCellId(value string) {
	if m != nil {
		m.CellId = value
	}
}
func (x *ActualLRPGroupsRequest) ToProto() *ProtoActualLRPGroupsRequest {
	if x == nil {
		return nil
	}

	proto := &ProtoActualLRPGroupsRequest{
		Domain: x.Domain,
		CellId: x.CellId,
	}
	return proto
}

func (x *ProtoActualLRPGroupsRequest) FromProto() *ActualLRPGroupsRequest {
	if x == nil {
		return nil
	}

	copysafe := &ActualLRPGroupsRequest{
		Domain: x.Domain,
		CellId: x.CellId,
	}
	return copysafe
}

func ActualLRPGroupsRequestToProtoSlice(values []*ActualLRPGroupsRequest) []*ProtoActualLRPGroupsRequest {
	if values == nil {
		return nil
	}
	result := make([]*ProtoActualLRPGroupsRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func ActualLRPGroupsRequestFromProtoSlice(values []*ProtoActualLRPGroupsRequest) []*ActualLRPGroupsRequest {
	if values == nil {
		return nil
	}
	result := make([]*ActualLRPGroupsRequest, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Deprecated: marked deprecated in actual_lrp_requests.proto
// Prevent copylock errors when using ProtoActualLRPGroupsByProcessGuidRequest directly
type ActualLRPGroupsByProcessGuidRequest struct {
	ProcessGuid string `json:"process_guid"`
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
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *ActualLRPGroupsByProcessGuidRequest) SetProcessGuid(value string) {
	if m != nil {
		m.ProcessGuid = value
	}
}
func (x *ActualLRPGroupsByProcessGuidRequest) ToProto() *ProtoActualLRPGroupsByProcessGuidRequest {
	if x == nil {
		return nil
	}

	proto := &ProtoActualLRPGroupsByProcessGuidRequest{
		ProcessGuid: x.ProcessGuid,
	}
	return proto
}

func (x *ProtoActualLRPGroupsByProcessGuidRequest) FromProto() *ActualLRPGroupsByProcessGuidRequest {
	if x == nil {
		return nil
	}

	copysafe := &ActualLRPGroupsByProcessGuidRequest{
		ProcessGuid: x.ProcessGuid,
	}
	return copysafe
}

func ActualLRPGroupsByProcessGuidRequestToProtoSlice(values []*ActualLRPGroupsByProcessGuidRequest) []*ProtoActualLRPGroupsByProcessGuidRequest {
	if values == nil {
		return nil
	}
	result := make([]*ProtoActualLRPGroupsByProcessGuidRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func ActualLRPGroupsByProcessGuidRequestFromProtoSlice(values []*ProtoActualLRPGroupsByProcessGuidRequest) []*ActualLRPGroupsByProcessGuidRequest {
	if values == nil {
		return nil
	}
	result := make([]*ActualLRPGroupsByProcessGuidRequest, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Deprecated: marked deprecated in actual_lrp_requests.proto
// Prevent copylock errors when using ProtoActualLRPGroupByProcessGuidAndIndexRequest directly
type ActualLRPGroupByProcessGuidAndIndexRequest struct {
	ProcessGuid string `json:"process_guid"`
	Index       int32  `json:"index"`
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
	var defaultValue string
	defaultValue = ""
	return defaultValue
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
	var defaultValue int32
	defaultValue = 0
	return defaultValue
}
func (m *ActualLRPGroupByProcessGuidAndIndexRequest) SetIndex(value int32) {
	if m != nil {
		m.Index = value
	}
}
func (x *ActualLRPGroupByProcessGuidAndIndexRequest) ToProto() *ProtoActualLRPGroupByProcessGuidAndIndexRequest {
	if x == nil {
		return nil
	}

	proto := &ProtoActualLRPGroupByProcessGuidAndIndexRequest{
		ProcessGuid: x.ProcessGuid,
		Index:       x.Index,
	}
	return proto
}

func (x *ProtoActualLRPGroupByProcessGuidAndIndexRequest) FromProto() *ActualLRPGroupByProcessGuidAndIndexRequest {
	if x == nil {
		return nil
	}

	copysafe := &ActualLRPGroupByProcessGuidAndIndexRequest{
		ProcessGuid: x.ProcessGuid,
		Index:       x.Index,
	}
	return copysafe
}

func ActualLRPGroupByProcessGuidAndIndexRequestToProtoSlice(values []*ActualLRPGroupByProcessGuidAndIndexRequest) []*ProtoActualLRPGroupByProcessGuidAndIndexRequest {
	if values == nil {
		return nil
	}
	result := make([]*ProtoActualLRPGroupByProcessGuidAndIndexRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func ActualLRPGroupByProcessGuidAndIndexRequestFromProtoSlice(values []*ProtoActualLRPGroupByProcessGuidAndIndexRequest) []*ActualLRPGroupByProcessGuidAndIndexRequest {
	if values == nil {
		return nil
	}
	result := make([]*ActualLRPGroupByProcessGuidAndIndexRequest, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoClaimActualLRPRequest directly
type ClaimActualLRPRequest struct {
	ProcessGuid          string                `json:"process_guid"`
	Index                int32                 `json:"index"`
	ActualLrpInstanceKey *ActualLRPInstanceKey `json:"actual_lrp_instance_key,omitempty"`
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
	if this.ActualLrpInstanceKey == nil {
		if that1.ActualLrpInstanceKey != nil {
			return false
		}
	} else if !this.ActualLrpInstanceKey.Equal(*that1.ActualLrpInstanceKey) {
		return false
	}
	return true
}
func (m *ClaimActualLRPRequest) GetProcessGuid() string {
	if m != nil {
		return m.ProcessGuid
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
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
	var defaultValue int32
	defaultValue = 0
	return defaultValue
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
	if x == nil {
		return nil
	}

	proto := &ProtoClaimActualLRPRequest{
		ProcessGuid:          x.ProcessGuid,
		Index:                x.Index,
		ActualLrpInstanceKey: x.ActualLrpInstanceKey.ToProto(),
	}
	return proto
}

func (x *ProtoClaimActualLRPRequest) FromProto() *ClaimActualLRPRequest {
	if x == nil {
		return nil
	}

	copysafe := &ClaimActualLRPRequest{
		ProcessGuid:          x.ProcessGuid,
		Index:                x.Index,
		ActualLrpInstanceKey: x.ActualLrpInstanceKey.FromProto(),
	}
	return copysafe
}

func ClaimActualLRPRequestToProtoSlice(values []*ClaimActualLRPRequest) []*ProtoClaimActualLRPRequest {
	if values == nil {
		return nil
	}
	result := make([]*ProtoClaimActualLRPRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func ClaimActualLRPRequestFromProtoSlice(values []*ProtoClaimActualLRPRequest) []*ClaimActualLRPRequest {
	if values == nil {
		return nil
	}
	result := make([]*ClaimActualLRPRequest, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoStartActualLRPRequest directly
type StartActualLRPRequest struct {
	ActualLrpKey            *ActualLRPKey             `json:"actual_lrp_key,omitempty"`
	ActualLrpInstanceKey    *ActualLRPInstanceKey     `json:"actual_lrp_instance_key,omitempty"`
	ActualLrpNetInfo        *ActualLRPNetInfo         `json:"actual_lrp_net_info,omitempty"`
	ActualLrpInternalRoutes []*ActualLRPInternalRoute `json:"actual_lrp_internal_routes,omitempty"`
	MetricTags              map[string]string         `json:"metric_tags,omitempty"`
	Routable                *bool                     `json:"Routable,omitempty"`
	AvailabilityZone        string                    `json:"availability_zone"`
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

	if this.ActualLrpKey == nil {
		if that1.ActualLrpKey != nil {
			return false
		}
	} else if !this.ActualLrpKey.Equal(*that1.ActualLrpKey) {
		return false
	}
	if this.ActualLrpInstanceKey == nil {
		if that1.ActualLrpInstanceKey != nil {
			return false
		}
	} else if !this.ActualLrpInstanceKey.Equal(*that1.ActualLrpInstanceKey) {
		return false
	}
	if this.ActualLrpNetInfo == nil {
		if that1.ActualLrpNetInfo != nil {
			return false
		}
	} else if !this.ActualLrpNetInfo.Equal(*that1.ActualLrpNetInfo) {
		return false
	}
	if this.ActualLrpInternalRoutes == nil {
		if that1.ActualLrpInternalRoutes != nil {
			return false
		}
	} else if len(this.ActualLrpInternalRoutes) != len(that1.ActualLrpInternalRoutes) {
		return false
	}
	for i := range this.ActualLrpInternalRoutes {
		if !this.ActualLrpInternalRoutes[i].Equal(that1.ActualLrpInternalRoutes[i]) {
			return false
		}
	}
	if this.MetricTags == nil {
		if that1.MetricTags != nil {
			return false
		}
	} else if len(this.MetricTags) != len(that1.MetricTags) {
		return false
	}
	for i := range this.MetricTags {
		if this.MetricTags[i] != that1.MetricTags[i] {
			return false
		}
	}
	if this.Routable == nil {
		if that1.Routable != nil {
			return false
		}
	} else if *this.Routable != *that1.Routable {
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
	var defaultValue bool
	defaultValue = false
	return &defaultValue
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
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *StartActualLRPRequest) SetAvailabilityZone(value string) {
	if m != nil {
		m.AvailabilityZone = value
	}
}
func (x *StartActualLRPRequest) ToProto() *ProtoStartActualLRPRequest {
	if x == nil {
		return nil
	}

	proto := &ProtoStartActualLRPRequest{
		ActualLrpKey:            x.ActualLrpKey.ToProto(),
		ActualLrpInstanceKey:    x.ActualLrpInstanceKey.ToProto(),
		ActualLrpNetInfo:        x.ActualLrpNetInfo.ToProto(),
		ActualLrpInternalRoutes: ActualLRPInternalRouteToProtoSlice(x.ActualLrpInternalRoutes),
		MetricTags:              x.MetricTags,
		Routable:                x.Routable,
		AvailabilityZone:        x.AvailabilityZone,
	}
	return proto
}

func (x *ProtoStartActualLRPRequest) FromProto() *StartActualLRPRequest {
	if x == nil {
		return nil
	}

	copysafe := &StartActualLRPRequest{
		ActualLrpKey:            x.ActualLrpKey.FromProto(),
		ActualLrpInstanceKey:    x.ActualLrpInstanceKey.FromProto(),
		ActualLrpNetInfo:        x.ActualLrpNetInfo.FromProto(),
		ActualLrpInternalRoutes: ActualLRPInternalRouteFromProtoSlice(x.ActualLrpInternalRoutes),
		MetricTags:              x.MetricTags,
		Routable:                x.Routable,
		AvailabilityZone:        x.AvailabilityZone,
	}
	return copysafe
}

func StartActualLRPRequestToProtoSlice(values []*StartActualLRPRequest) []*ProtoStartActualLRPRequest {
	if values == nil {
		return nil
	}
	result := make([]*ProtoStartActualLRPRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func StartActualLRPRequestFromProtoSlice(values []*ProtoStartActualLRPRequest) []*StartActualLRPRequest {
	if values == nil {
		return nil
	}
	result := make([]*StartActualLRPRequest, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoCrashActualLRPRequest directly
type CrashActualLRPRequest struct {
	ActualLrpKey         *ActualLRPKey         `json:"actual_lrp_key,omitempty"`
	ActualLrpInstanceKey *ActualLRPInstanceKey `json:"actual_lrp_instance_key,omitempty"`
	ErrorMessage         string                `json:"error_message"`
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

	if this.ActualLrpKey == nil {
		if that1.ActualLrpKey != nil {
			return false
		}
	} else if !this.ActualLrpKey.Equal(*that1.ActualLrpKey) {
		return false
	}
	if this.ActualLrpInstanceKey == nil {
		if that1.ActualLrpInstanceKey != nil {
			return false
		}
	} else if !this.ActualLrpInstanceKey.Equal(*that1.ActualLrpInstanceKey) {
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
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *CrashActualLRPRequest) SetErrorMessage(value string) {
	if m != nil {
		m.ErrorMessage = value
	}
}
func (x *CrashActualLRPRequest) ToProto() *ProtoCrashActualLRPRequest {
	if x == nil {
		return nil
	}

	proto := &ProtoCrashActualLRPRequest{
		ActualLrpKey:         x.ActualLrpKey.ToProto(),
		ActualLrpInstanceKey: x.ActualLrpInstanceKey.ToProto(),
		ErrorMessage:         x.ErrorMessage,
	}
	return proto
}

func (x *ProtoCrashActualLRPRequest) FromProto() *CrashActualLRPRequest {
	if x == nil {
		return nil
	}

	copysafe := &CrashActualLRPRequest{
		ActualLrpKey:         x.ActualLrpKey.FromProto(),
		ActualLrpInstanceKey: x.ActualLrpInstanceKey.FromProto(),
		ErrorMessage:         x.ErrorMessage,
	}
	return copysafe
}

func CrashActualLRPRequestToProtoSlice(values []*CrashActualLRPRequest) []*ProtoCrashActualLRPRequest {
	if values == nil {
		return nil
	}
	result := make([]*ProtoCrashActualLRPRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func CrashActualLRPRequestFromProtoSlice(values []*ProtoCrashActualLRPRequest) []*CrashActualLRPRequest {
	if values == nil {
		return nil
	}
	result := make([]*CrashActualLRPRequest, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoFailActualLRPRequest directly
type FailActualLRPRequest struct {
	ActualLrpKey *ActualLRPKey `json:"actual_lrp_key,omitempty"`
	ErrorMessage string        `json:"error_message"`
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

	if this.ActualLrpKey == nil {
		if that1.ActualLrpKey != nil {
			return false
		}
	} else if !this.ActualLrpKey.Equal(*that1.ActualLrpKey) {
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
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *FailActualLRPRequest) SetErrorMessage(value string) {
	if m != nil {
		m.ErrorMessage = value
	}
}
func (x *FailActualLRPRequest) ToProto() *ProtoFailActualLRPRequest {
	if x == nil {
		return nil
	}

	proto := &ProtoFailActualLRPRequest{
		ActualLrpKey: x.ActualLrpKey.ToProto(),
		ErrorMessage: x.ErrorMessage,
	}
	return proto
}

func (x *ProtoFailActualLRPRequest) FromProto() *FailActualLRPRequest {
	if x == nil {
		return nil
	}

	copysafe := &FailActualLRPRequest{
		ActualLrpKey: x.ActualLrpKey.FromProto(),
		ErrorMessage: x.ErrorMessage,
	}
	return copysafe
}

func FailActualLRPRequestToProtoSlice(values []*FailActualLRPRequest) []*ProtoFailActualLRPRequest {
	if values == nil {
		return nil
	}
	result := make([]*ProtoFailActualLRPRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func FailActualLRPRequestFromProtoSlice(values []*ProtoFailActualLRPRequest) []*FailActualLRPRequest {
	if values == nil {
		return nil
	}
	result := make([]*FailActualLRPRequest, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoRetireActualLRPRequest directly
type RetireActualLRPRequest struct {
	ActualLrpKey *ActualLRPKey `json:"actual_lrp_key,omitempty"`
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

	if this.ActualLrpKey == nil {
		if that1.ActualLrpKey != nil {
			return false
		}
	} else if !this.ActualLrpKey.Equal(*that1.ActualLrpKey) {
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
	if x == nil {
		return nil
	}

	proto := &ProtoRetireActualLRPRequest{
		ActualLrpKey: x.ActualLrpKey.ToProto(),
	}
	return proto
}

func (x *ProtoRetireActualLRPRequest) FromProto() *RetireActualLRPRequest {
	if x == nil {
		return nil
	}

	copysafe := &RetireActualLRPRequest{
		ActualLrpKey: x.ActualLrpKey.FromProto(),
	}
	return copysafe
}

func RetireActualLRPRequestToProtoSlice(values []*RetireActualLRPRequest) []*ProtoRetireActualLRPRequest {
	if values == nil {
		return nil
	}
	result := make([]*ProtoRetireActualLRPRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func RetireActualLRPRequestFromProtoSlice(values []*ProtoRetireActualLRPRequest) []*RetireActualLRPRequest {
	if values == nil {
		return nil
	}
	result := make([]*RetireActualLRPRequest, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoRemoveActualLRPRequest directly
type RemoveActualLRPRequest struct {
	ProcessGuid          string                `json:"process_guid"`
	Index                int32                 `json:"index"`
	ActualLrpInstanceKey *ActualLRPInstanceKey `json:"actual_lrp_instance_key,omitempty"`
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
	if this.ActualLrpInstanceKey == nil {
		if that1.ActualLrpInstanceKey != nil {
			return false
		}
	} else if !this.ActualLrpInstanceKey.Equal(*that1.ActualLrpInstanceKey) {
		return false
	}
	return true
}
func (m *RemoveActualLRPRequest) GetProcessGuid() string {
	if m != nil {
		return m.ProcessGuid
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
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
	var defaultValue int32
	defaultValue = 0
	return defaultValue
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
	if x == nil {
		return nil
	}

	proto := &ProtoRemoveActualLRPRequest{
		ProcessGuid:          x.ProcessGuid,
		Index:                x.Index,
		ActualLrpInstanceKey: x.ActualLrpInstanceKey.ToProto(),
	}
	return proto
}

func (x *ProtoRemoveActualLRPRequest) FromProto() *RemoveActualLRPRequest {
	if x == nil {
		return nil
	}

	copysafe := &RemoveActualLRPRequest{
		ProcessGuid:          x.ProcessGuid,
		Index:                x.Index,
		ActualLrpInstanceKey: x.ActualLrpInstanceKey.FromProto(),
	}
	return copysafe
}

func RemoveActualLRPRequestToProtoSlice(values []*RemoveActualLRPRequest) []*ProtoRemoveActualLRPRequest {
	if values == nil {
		return nil
	}
	result := make([]*ProtoRemoveActualLRPRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func RemoveActualLRPRequestFromProtoSlice(values []*ProtoRemoveActualLRPRequest) []*RemoveActualLRPRequest {
	if values == nil {
		return nil
	}
	result := make([]*RemoveActualLRPRequest, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoActualLRPsResponse directly
type ActualLRPsResponse struct {
	Error      *Error       `json:"error,omitempty"`
	ActualLrps []*ActualLRP `json:"actual_lrps,omitempty"`
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

	if this.Error == nil {
		if that1.Error != nil {
			return false
		}
	} else if !this.Error.Equal(*that1.Error) {
		return false
	}
	if this.ActualLrps == nil {
		if that1.ActualLrps != nil {
			return false
		}
	} else if len(this.ActualLrps) != len(that1.ActualLrps) {
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
	if x == nil {
		return nil
	}

	proto := &ProtoActualLRPsResponse{
		Error:      x.Error.ToProto(),
		ActualLrps: ActualLRPToProtoSlice(x.ActualLrps),
	}
	return proto
}

func (x *ProtoActualLRPsResponse) FromProto() *ActualLRPsResponse {
	if x == nil {
		return nil
	}

	copysafe := &ActualLRPsResponse{
		Error:      x.Error.FromProto(),
		ActualLrps: ActualLRPFromProtoSlice(x.ActualLrps),
	}
	return copysafe
}

func ActualLRPsResponseToProtoSlice(values []*ActualLRPsResponse) []*ProtoActualLRPsResponse {
	if values == nil {
		return nil
	}
	result := make([]*ProtoActualLRPsResponse, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func ActualLRPsResponseFromProtoSlice(values []*ProtoActualLRPsResponse) []*ActualLRPsResponse {
	if values == nil {
		return nil
	}
	result := make([]*ActualLRPsResponse, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoActualLRPsRequest directly
type ActualLRPsRequest struct {
	Domain      string `json:"domain"`
	CellId      string `json:"cell_id"`
	ProcessGuid string `json:"process_guid"`
	Index       *int32 `json:"index"`
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
	if this.Index == nil {
		if that1.Index != nil {
			return false
		}
	} else if *this.Index != *that1.Index {
		return false
	}
	return true
}
func (m *ActualLRPsRequest) GetDomain() string {
	if m != nil {
		return m.Domain
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
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
	var defaultValue string
	defaultValue = ""
	return defaultValue
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
	var defaultValue string
	defaultValue = ""
	return defaultValue
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
	var defaultValue int32
	defaultValue = 0
	return &defaultValue
}
func (m *ActualLRPsRequest) SetIndex(value *int32) {
	if m != nil {
		m.Index = value
	}
}
func (x *ActualLRPsRequest) ToProto() *ProtoActualLRPsRequest {
	if x == nil {
		return nil
	}

	proto := &ProtoActualLRPsRequest{
		Domain:      x.Domain,
		CellId:      x.CellId,
		ProcessGuid: x.ProcessGuid,
		Index:       x.Index,
	}
	return proto
}

func (x *ProtoActualLRPsRequest) FromProto() *ActualLRPsRequest {
	if x == nil {
		return nil
	}

	copysafe := &ActualLRPsRequest{
		Domain:      x.Domain,
		CellId:      x.CellId,
		ProcessGuid: x.ProcessGuid,
		Index:       x.Index,
	}
	return copysafe
}

func ActualLRPsRequestToProtoSlice(values []*ActualLRPsRequest) []*ProtoActualLRPsRequest {
	if values == nil {
		return nil
	}
	result := make([]*ProtoActualLRPsRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func ActualLRPsRequestFromProtoSlice(values []*ProtoActualLRPsRequest) []*ActualLRPsRequest {
	if values == nil {
		return nil
	}
	result := make([]*ActualLRPsRequest, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}
