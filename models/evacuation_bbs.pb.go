// Code generated by protoc-gen-go-bbs. DO NOT EDIT.
// versions:
// - protoc-gen-go-bbs v0.0.1
// - protoc            v5.29.0
// source: evacuation.proto

package models

// Prevent copylock errors when using ProtoEvacuationResponse directly
type EvacuationResponse struct {
	Error         *Error `json:"error,omitempty"`
	KeepContainer bool   `json:"keep_container"`
}

func (this *EvacuationResponse) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*EvacuationResponse)
	if !ok {
		that2, ok := that.(EvacuationResponse)
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
	if this.KeepContainer != that1.KeepContainer {
		return false
	}
	return true
}
func (m *EvacuationResponse) GetError() *Error {
	if m != nil {
		return m.Error
	}
	return nil
}
func (m *EvacuationResponse) SetError(value *Error) {
	if m != nil {
		m.Error = value
	}
}
func (m *EvacuationResponse) GetKeepContainer() bool {
	if m != nil {
		return m.KeepContainer
	}
	var defaultValue bool
	defaultValue = false
	return defaultValue
}
func (m *EvacuationResponse) SetKeepContainer(value bool) {
	if m != nil {
		m.KeepContainer = value
	}
}
func (x *EvacuationResponse) ToProto() *ProtoEvacuationResponse {
	if x == nil {
		return nil
	}

	proto := &ProtoEvacuationResponse{
		Error:         x.Error.ToProto(),
		KeepContainer: x.KeepContainer,
	}
	return proto
}

func (x *ProtoEvacuationResponse) FromProto() *EvacuationResponse {
	if x == nil {
		return nil
	}

	copysafe := &EvacuationResponse{
		Error:         x.Error.FromProto(),
		KeepContainer: x.KeepContainer,
	}
	return copysafe
}

func EvacuationResponseToProtoSlice(values []*EvacuationResponse) []*ProtoEvacuationResponse {
	if values == nil {
		return nil
	}
	result := make([]*ProtoEvacuationResponse, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func EvacuationResponseFromProtoSlice(values []*ProtoEvacuationResponse) []*EvacuationResponse {
	if values == nil {
		return nil
	}
	result := make([]*EvacuationResponse, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoEvacuateClaimedActualLRPRequest directly
type EvacuateClaimedActualLRPRequest struct {
	ActualLrpKey         *ActualLRPKey         `json:"actual_lrp_key,omitempty"`
	ActualLrpInstanceKey *ActualLRPInstanceKey `json:"actual_lrp_instance_key,omitempty"`
}

func (this *EvacuateClaimedActualLRPRequest) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*EvacuateClaimedActualLRPRequest)
	if !ok {
		that2, ok := that.(EvacuateClaimedActualLRPRequest)
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
	return true
}
func (m *EvacuateClaimedActualLRPRequest) GetActualLrpKey() *ActualLRPKey {
	if m != nil {
		return m.ActualLrpKey
	}
	return nil
}
func (m *EvacuateClaimedActualLRPRequest) SetActualLrpKey(value *ActualLRPKey) {
	if m != nil {
		m.ActualLrpKey = value
	}
}
func (m *EvacuateClaimedActualLRPRequest) GetActualLrpInstanceKey() *ActualLRPInstanceKey {
	if m != nil {
		return m.ActualLrpInstanceKey
	}
	return nil
}
func (m *EvacuateClaimedActualLRPRequest) SetActualLrpInstanceKey(value *ActualLRPInstanceKey) {
	if m != nil {
		m.ActualLrpInstanceKey = value
	}
}
func (x *EvacuateClaimedActualLRPRequest) ToProto() *ProtoEvacuateClaimedActualLRPRequest {
	if x == nil {
		return nil
	}

	proto := &ProtoEvacuateClaimedActualLRPRequest{
		ActualLrpKey:         x.ActualLrpKey.ToProto(),
		ActualLrpInstanceKey: x.ActualLrpInstanceKey.ToProto(),
	}
	return proto
}

func (x *ProtoEvacuateClaimedActualLRPRequest) FromProto() *EvacuateClaimedActualLRPRequest {
	if x == nil {
		return nil
	}

	copysafe := &EvacuateClaimedActualLRPRequest{
		ActualLrpKey:         x.ActualLrpKey.FromProto(),
		ActualLrpInstanceKey: x.ActualLrpInstanceKey.FromProto(),
	}
	return copysafe
}

func EvacuateClaimedActualLRPRequestToProtoSlice(values []*EvacuateClaimedActualLRPRequest) []*ProtoEvacuateClaimedActualLRPRequest {
	if values == nil {
		return nil
	}
	result := make([]*ProtoEvacuateClaimedActualLRPRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func EvacuateClaimedActualLRPRequestFromProtoSlice(values []*ProtoEvacuateClaimedActualLRPRequest) []*EvacuateClaimedActualLRPRequest {
	if values == nil {
		return nil
	}
	result := make([]*EvacuateClaimedActualLRPRequest, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoEvacuateRunningActualLRPRequest directly
type EvacuateRunningActualLRPRequest struct {
	ActualLrpKey            *ActualLRPKey             `json:"actual_lrp_key,omitempty"`
	ActualLrpInstanceKey    *ActualLRPInstanceKey     `json:"actual_lrp_instance_key,omitempty"`
	ActualLrpNetInfo        *ActualLRPNetInfo         `json:"actual_lrp_net_info,omitempty"`
	ActualLrpInternalRoutes []*ActualLRPInternalRoute `json:"actual_lrp_internal_routes,omitempty"`
	MetricTags              map[string]string         `json:"metric_tags,omitempty"`
	Routable                *bool                     `json:"Routable,omitempty"`
	AvailabilityZone        string                    `json:"availability_zone"`
}

func (this *EvacuateRunningActualLRPRequest) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*EvacuateRunningActualLRPRequest)
	if !ok {
		that2, ok := that.(EvacuateRunningActualLRPRequest)
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
func (m *EvacuateRunningActualLRPRequest) GetActualLrpKey() *ActualLRPKey {
	if m != nil {
		return m.ActualLrpKey
	}
	return nil
}
func (m *EvacuateRunningActualLRPRequest) SetActualLrpKey(value *ActualLRPKey) {
	if m != nil {
		m.ActualLrpKey = value
	}
}
func (m *EvacuateRunningActualLRPRequest) GetActualLrpInstanceKey() *ActualLRPInstanceKey {
	if m != nil {
		return m.ActualLrpInstanceKey
	}
	return nil
}
func (m *EvacuateRunningActualLRPRequest) SetActualLrpInstanceKey(value *ActualLRPInstanceKey) {
	if m != nil {
		m.ActualLrpInstanceKey = value
	}
}
func (m *EvacuateRunningActualLRPRequest) GetActualLrpNetInfo() *ActualLRPNetInfo {
	if m != nil {
		return m.ActualLrpNetInfo
	}
	return nil
}
func (m *EvacuateRunningActualLRPRequest) SetActualLrpNetInfo(value *ActualLRPNetInfo) {
	if m != nil {
		m.ActualLrpNetInfo = value
	}
}
func (m *EvacuateRunningActualLRPRequest) GetActualLrpInternalRoutes() []*ActualLRPInternalRoute {
	if m != nil {
		return m.ActualLrpInternalRoutes
	}
	return nil
}
func (m *EvacuateRunningActualLRPRequest) SetActualLrpInternalRoutes(value []*ActualLRPInternalRoute) {
	if m != nil {
		m.ActualLrpInternalRoutes = value
	}
}
func (m *EvacuateRunningActualLRPRequest) GetMetricTags() map[string]string {
	if m != nil {
		return m.MetricTags
	}
	return nil
}
func (m *EvacuateRunningActualLRPRequest) SetMetricTags(value map[string]string) {
	if m != nil {
		m.MetricTags = value
	}
}
func (m *EvacuateRunningActualLRPRequest) RoutableExists() bool {
	return m != nil && m.Routable != nil
}
func (m *EvacuateRunningActualLRPRequest) GetRoutable() *bool {
	if m != nil && m.Routable != nil {
		return m.Routable
	}
	var defaultValue bool
	defaultValue = false
	return &defaultValue
}
func (m *EvacuateRunningActualLRPRequest) SetRoutable(value *bool) {
	if m != nil {
		m.Routable = value
	}
}
func (m *EvacuateRunningActualLRPRequest) GetAvailabilityZone() string {
	if m != nil {
		return m.AvailabilityZone
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *EvacuateRunningActualLRPRequest) SetAvailabilityZone(value string) {
	if m != nil {
		m.AvailabilityZone = value
	}
}
func (x *EvacuateRunningActualLRPRequest) ToProto() *ProtoEvacuateRunningActualLRPRequest {
	if x == nil {
		return nil
	}

	proto := &ProtoEvacuateRunningActualLRPRequest{
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

func (x *ProtoEvacuateRunningActualLRPRequest) FromProto() *EvacuateRunningActualLRPRequest {
	if x == nil {
		return nil
	}

	copysafe := &EvacuateRunningActualLRPRequest{
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

func EvacuateRunningActualLRPRequestToProtoSlice(values []*EvacuateRunningActualLRPRequest) []*ProtoEvacuateRunningActualLRPRequest {
	if values == nil {
		return nil
	}
	result := make([]*ProtoEvacuateRunningActualLRPRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func EvacuateRunningActualLRPRequestFromProtoSlice(values []*ProtoEvacuateRunningActualLRPRequest) []*EvacuateRunningActualLRPRequest {
	if values == nil {
		return nil
	}
	result := make([]*EvacuateRunningActualLRPRequest, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoEvacuateStoppedActualLRPRequest directly
type EvacuateStoppedActualLRPRequest struct {
	ActualLrpKey         *ActualLRPKey         `json:"actual_lrp_key,omitempty"`
	ActualLrpInstanceKey *ActualLRPInstanceKey `json:"actual_lrp_instance_key,omitempty"`
}

func (this *EvacuateStoppedActualLRPRequest) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*EvacuateStoppedActualLRPRequest)
	if !ok {
		that2, ok := that.(EvacuateStoppedActualLRPRequest)
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
	return true
}
func (m *EvacuateStoppedActualLRPRequest) GetActualLrpKey() *ActualLRPKey {
	if m != nil {
		return m.ActualLrpKey
	}
	return nil
}
func (m *EvacuateStoppedActualLRPRequest) SetActualLrpKey(value *ActualLRPKey) {
	if m != nil {
		m.ActualLrpKey = value
	}
}
func (m *EvacuateStoppedActualLRPRequest) GetActualLrpInstanceKey() *ActualLRPInstanceKey {
	if m != nil {
		return m.ActualLrpInstanceKey
	}
	return nil
}
func (m *EvacuateStoppedActualLRPRequest) SetActualLrpInstanceKey(value *ActualLRPInstanceKey) {
	if m != nil {
		m.ActualLrpInstanceKey = value
	}
}
func (x *EvacuateStoppedActualLRPRequest) ToProto() *ProtoEvacuateStoppedActualLRPRequest {
	if x == nil {
		return nil
	}

	proto := &ProtoEvacuateStoppedActualLRPRequest{
		ActualLrpKey:         x.ActualLrpKey.ToProto(),
		ActualLrpInstanceKey: x.ActualLrpInstanceKey.ToProto(),
	}
	return proto
}

func (x *ProtoEvacuateStoppedActualLRPRequest) FromProto() *EvacuateStoppedActualLRPRequest {
	if x == nil {
		return nil
	}

	copysafe := &EvacuateStoppedActualLRPRequest{
		ActualLrpKey:         x.ActualLrpKey.FromProto(),
		ActualLrpInstanceKey: x.ActualLrpInstanceKey.FromProto(),
	}
	return copysafe
}

func EvacuateStoppedActualLRPRequestToProtoSlice(values []*EvacuateStoppedActualLRPRequest) []*ProtoEvacuateStoppedActualLRPRequest {
	if values == nil {
		return nil
	}
	result := make([]*ProtoEvacuateStoppedActualLRPRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func EvacuateStoppedActualLRPRequestFromProtoSlice(values []*ProtoEvacuateStoppedActualLRPRequest) []*EvacuateStoppedActualLRPRequest {
	if values == nil {
		return nil
	}
	result := make([]*EvacuateStoppedActualLRPRequest, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoEvacuateCrashedActualLRPRequest directly
type EvacuateCrashedActualLRPRequest struct {
	ActualLrpKey         *ActualLRPKey         `json:"actual_lrp_key,omitempty"`
	ActualLrpInstanceKey *ActualLRPInstanceKey `json:"actual_lrp_instance_key,omitempty"`
	ErrorMessage         string                `json:"error_message"`
}

func (this *EvacuateCrashedActualLRPRequest) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*EvacuateCrashedActualLRPRequest)
	if !ok {
		that2, ok := that.(EvacuateCrashedActualLRPRequest)
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
func (m *EvacuateCrashedActualLRPRequest) GetActualLrpKey() *ActualLRPKey {
	if m != nil {
		return m.ActualLrpKey
	}
	return nil
}
func (m *EvacuateCrashedActualLRPRequest) SetActualLrpKey(value *ActualLRPKey) {
	if m != nil {
		m.ActualLrpKey = value
	}
}
func (m *EvacuateCrashedActualLRPRequest) GetActualLrpInstanceKey() *ActualLRPInstanceKey {
	if m != nil {
		return m.ActualLrpInstanceKey
	}
	return nil
}
func (m *EvacuateCrashedActualLRPRequest) SetActualLrpInstanceKey(value *ActualLRPInstanceKey) {
	if m != nil {
		m.ActualLrpInstanceKey = value
	}
}
func (m *EvacuateCrashedActualLRPRequest) GetErrorMessage() string {
	if m != nil {
		return m.ErrorMessage
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *EvacuateCrashedActualLRPRequest) SetErrorMessage(value string) {
	if m != nil {
		m.ErrorMessage = value
	}
}
func (x *EvacuateCrashedActualLRPRequest) ToProto() *ProtoEvacuateCrashedActualLRPRequest {
	if x == nil {
		return nil
	}

	proto := &ProtoEvacuateCrashedActualLRPRequest{
		ActualLrpKey:         x.ActualLrpKey.ToProto(),
		ActualLrpInstanceKey: x.ActualLrpInstanceKey.ToProto(),
		ErrorMessage:         x.ErrorMessage,
	}
	return proto
}

func (x *ProtoEvacuateCrashedActualLRPRequest) FromProto() *EvacuateCrashedActualLRPRequest {
	if x == nil {
		return nil
	}

	copysafe := &EvacuateCrashedActualLRPRequest{
		ActualLrpKey:         x.ActualLrpKey.FromProto(),
		ActualLrpInstanceKey: x.ActualLrpInstanceKey.FromProto(),
		ErrorMessage:         x.ErrorMessage,
	}
	return copysafe
}

func EvacuateCrashedActualLRPRequestToProtoSlice(values []*EvacuateCrashedActualLRPRequest) []*ProtoEvacuateCrashedActualLRPRequest {
	if values == nil {
		return nil
	}
	result := make([]*ProtoEvacuateCrashedActualLRPRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func EvacuateCrashedActualLRPRequestFromProtoSlice(values []*ProtoEvacuateCrashedActualLRPRequest) []*EvacuateCrashedActualLRPRequest {
	if values == nil {
		return nil
	}
	result := make([]*EvacuateCrashedActualLRPRequest, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoRemoveEvacuatingActualLRPRequest directly
type RemoveEvacuatingActualLRPRequest struct {
	ActualLrpKey         *ActualLRPKey         `json:"actual_lrp_key,omitempty"`
	ActualLrpInstanceKey *ActualLRPInstanceKey `json:"actual_lrp_instance_key,omitempty"`
}

func (this *RemoveEvacuatingActualLRPRequest) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*RemoveEvacuatingActualLRPRequest)
	if !ok {
		that2, ok := that.(RemoveEvacuatingActualLRPRequest)
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
	return true
}
func (m *RemoveEvacuatingActualLRPRequest) GetActualLrpKey() *ActualLRPKey {
	if m != nil {
		return m.ActualLrpKey
	}
	return nil
}
func (m *RemoveEvacuatingActualLRPRequest) SetActualLrpKey(value *ActualLRPKey) {
	if m != nil {
		m.ActualLrpKey = value
	}
}
func (m *RemoveEvacuatingActualLRPRequest) GetActualLrpInstanceKey() *ActualLRPInstanceKey {
	if m != nil {
		return m.ActualLrpInstanceKey
	}
	return nil
}
func (m *RemoveEvacuatingActualLRPRequest) SetActualLrpInstanceKey(value *ActualLRPInstanceKey) {
	if m != nil {
		m.ActualLrpInstanceKey = value
	}
}
func (x *RemoveEvacuatingActualLRPRequest) ToProto() *ProtoRemoveEvacuatingActualLRPRequest {
	if x == nil {
		return nil
	}

	proto := &ProtoRemoveEvacuatingActualLRPRequest{
		ActualLrpKey:         x.ActualLrpKey.ToProto(),
		ActualLrpInstanceKey: x.ActualLrpInstanceKey.ToProto(),
	}
	return proto
}

func (x *ProtoRemoveEvacuatingActualLRPRequest) FromProto() *RemoveEvacuatingActualLRPRequest {
	if x == nil {
		return nil
	}

	copysafe := &RemoveEvacuatingActualLRPRequest{
		ActualLrpKey:         x.ActualLrpKey.FromProto(),
		ActualLrpInstanceKey: x.ActualLrpInstanceKey.FromProto(),
	}
	return copysafe
}

func RemoveEvacuatingActualLRPRequestToProtoSlice(values []*RemoveEvacuatingActualLRPRequest) []*ProtoRemoveEvacuatingActualLRPRequest {
	if values == nil {
		return nil
	}
	result := make([]*ProtoRemoveEvacuatingActualLRPRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func RemoveEvacuatingActualLRPRequestFromProtoSlice(values []*ProtoRemoveEvacuatingActualLRPRequest) []*RemoveEvacuatingActualLRPRequest {
	if values == nil {
		return nil
	}
	result := make([]*RemoveEvacuatingActualLRPRequest, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoRemoveEvacuatingActualLRPResponse directly
type RemoveEvacuatingActualLRPResponse struct {
	Error *Error `json:"error,omitempty"`
}

func (this *RemoveEvacuatingActualLRPResponse) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*RemoveEvacuatingActualLRPResponse)
	if !ok {
		that2, ok := that.(RemoveEvacuatingActualLRPResponse)
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
func (m *RemoveEvacuatingActualLRPResponse) GetError() *Error {
	if m != nil {
		return m.Error
	}
	return nil
}
func (m *RemoveEvacuatingActualLRPResponse) SetError(value *Error) {
	if m != nil {
		m.Error = value
	}
}
func (x *RemoveEvacuatingActualLRPResponse) ToProto() *ProtoRemoveEvacuatingActualLRPResponse {
	if x == nil {
		return nil
	}

	proto := &ProtoRemoveEvacuatingActualLRPResponse{
		Error: x.Error.ToProto(),
	}
	return proto
}

func (x *ProtoRemoveEvacuatingActualLRPResponse) FromProto() *RemoveEvacuatingActualLRPResponse {
	if x == nil {
		return nil
	}

	copysafe := &RemoveEvacuatingActualLRPResponse{
		Error: x.Error.FromProto(),
	}
	return copysafe
}

func RemoveEvacuatingActualLRPResponseToProtoSlice(values []*RemoveEvacuatingActualLRPResponse) []*ProtoRemoveEvacuatingActualLRPResponse {
	if values == nil {
		return nil
	}
	result := make([]*ProtoRemoveEvacuatingActualLRPResponse, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func RemoveEvacuatingActualLRPResponseFromProtoSlice(values []*ProtoRemoveEvacuatingActualLRPResponse) []*RemoveEvacuatingActualLRPResponse {
	if values == nil {
		return nil
	}
	result := make([]*RemoveEvacuatingActualLRPResponse, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}
