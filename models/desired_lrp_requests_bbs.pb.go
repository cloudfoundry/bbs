// Code generated by protoc-gen-go-bbs. DO NOT EDIT.
// versions:
// - protoc-gen-go-bbs v0.0.1
// - protoc            v5.28.2
// source: desired_lrp_requests.proto

package models

// Prevent copylock errors when using ProtoDesiredLRPLifecycleResponse directly
type DesiredLRPLifecycleResponse struct {
	Error *Error `json:"error,omitempty"`
}

func (this *DesiredLRPLifecycleResponse) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*DesiredLRPLifecycleResponse)
	if !ok {
		that2, ok := that.(DesiredLRPLifecycleResponse)
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
func (m *DesiredLRPLifecycleResponse) GetError() *Error {
	if m != nil {
		return m.Error
	}
	return nil
}
func (m *DesiredLRPLifecycleResponse) SetError(value *Error) {
	if m != nil {
		m.Error = value
	}
}
func (x *DesiredLRPLifecycleResponse) ToProto() *ProtoDesiredLRPLifecycleResponse {
	if x == nil {
		return nil
	}

	proto := &ProtoDesiredLRPLifecycleResponse{
		Error: x.Error.ToProto(),
	}
	return proto
}

func (x *ProtoDesiredLRPLifecycleResponse) FromProto() *DesiredLRPLifecycleResponse {
	if x == nil {
		return nil
	}

	copysafe := &DesiredLRPLifecycleResponse{
		Error: x.Error.FromProto(),
	}
	return copysafe
}

func DesiredLRPLifecycleResponseToProtoSlice(values []*DesiredLRPLifecycleResponse) []*ProtoDesiredLRPLifecycleResponse {
	if values == nil {
		return nil
	}
	result := make([]*ProtoDesiredLRPLifecycleResponse, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func DesiredLRPLifecycleResponseFromProtoSlice(values []*ProtoDesiredLRPLifecycleResponse) []*DesiredLRPLifecycleResponse {
	if values == nil {
		return nil
	}
	result := make([]*DesiredLRPLifecycleResponse, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoDesiredLRPsResponse directly
type DesiredLRPsResponse struct {
	Error       *Error        `json:"error,omitempty"`
	DesiredLrps []*DesiredLRP `json:"desired_lrps,omitempty"`
}

func (this *DesiredLRPsResponse) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*DesiredLRPsResponse)
	if !ok {
		that2, ok := that.(DesiredLRPsResponse)
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
	if this.DesiredLrps == nil {
		if that1.DesiredLrps != nil {
			return false
		}
	} else if len(this.DesiredLrps) != len(that1.DesiredLrps) {
		return false
	}
	for i := range this.DesiredLrps {
		if !this.DesiredLrps[i].Equal(that1.DesiredLrps[i]) {
			return false
		}
	}
	return true
}
func (m *DesiredLRPsResponse) GetError() *Error {
	if m != nil {
		return m.Error
	}
	return nil
}
func (m *DesiredLRPsResponse) SetError(value *Error) {
	if m != nil {
		m.Error = value
	}
}
func (m *DesiredLRPsResponse) GetDesiredLrps() []*DesiredLRP {
	if m != nil {
		return m.DesiredLrps
	}
	return nil
}
func (m *DesiredLRPsResponse) SetDesiredLrps(value []*DesiredLRP) {
	if m != nil {
		m.DesiredLrps = value
	}
}
func (x *DesiredLRPsResponse) ToProto() *ProtoDesiredLRPsResponse {
	if x == nil {
		return nil
	}

	proto := &ProtoDesiredLRPsResponse{
		Error:       x.Error.ToProto(),
		DesiredLrps: DesiredLRPToProtoSlice(x.DesiredLrps),
	}
	return proto
}

func (x *ProtoDesiredLRPsResponse) FromProto() *DesiredLRPsResponse {
	if x == nil {
		return nil
	}

	copysafe := &DesiredLRPsResponse{
		Error:       x.Error.FromProto(),
		DesiredLrps: DesiredLRPFromProtoSlice(x.DesiredLrps),
	}
	return copysafe
}

func DesiredLRPsResponseToProtoSlice(values []*DesiredLRPsResponse) []*ProtoDesiredLRPsResponse {
	if values == nil {
		return nil
	}
	result := make([]*ProtoDesiredLRPsResponse, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func DesiredLRPsResponseFromProtoSlice(values []*ProtoDesiredLRPsResponse) []*DesiredLRPsResponse {
	if values == nil {
		return nil
	}
	result := make([]*DesiredLRPsResponse, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoDesiredLRPsRequest directly
type DesiredLRPsRequest struct {
	Domain       string   `json:"domain"`
	ProcessGuids []string `json:"process_guids,omitempty"`
}

func (this *DesiredLRPsRequest) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*DesiredLRPsRequest)
	if !ok {
		that2, ok := that.(DesiredLRPsRequest)
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
	if this.ProcessGuids == nil {
		if that1.ProcessGuids != nil {
			return false
		}
	} else if len(this.ProcessGuids) != len(that1.ProcessGuids) {
		return false
	}
	for i := range this.ProcessGuids {
		if this.ProcessGuids[i] != that1.ProcessGuids[i] {
			return false
		}
	}
	return true
}
func (m *DesiredLRPsRequest) GetDomain() string {
	if m != nil {
		return m.Domain
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *DesiredLRPsRequest) SetDomain(value string) {
	if m != nil {
		m.Domain = value
	}
}
func (m *DesiredLRPsRequest) GetProcessGuids() []string {
	if m != nil {
		return m.ProcessGuids
	}
	return nil
}
func (m *DesiredLRPsRequest) SetProcessGuids(value []string) {
	if m != nil {
		m.ProcessGuids = value
	}
}
func (x *DesiredLRPsRequest) ToProto() *ProtoDesiredLRPsRequest {
	if x == nil {
		return nil
	}

	proto := &ProtoDesiredLRPsRequest{
		Domain:       x.Domain,
		ProcessGuids: x.ProcessGuids,
	}
	return proto
}

func (x *ProtoDesiredLRPsRequest) FromProto() *DesiredLRPsRequest {
	if x == nil {
		return nil
	}

	copysafe := &DesiredLRPsRequest{
		Domain:       x.Domain,
		ProcessGuids: x.ProcessGuids,
	}
	return copysafe
}

func DesiredLRPsRequestToProtoSlice(values []*DesiredLRPsRequest) []*ProtoDesiredLRPsRequest {
	if values == nil {
		return nil
	}
	result := make([]*ProtoDesiredLRPsRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func DesiredLRPsRequestFromProtoSlice(values []*ProtoDesiredLRPsRequest) []*DesiredLRPsRequest {
	if values == nil {
		return nil
	}
	result := make([]*DesiredLRPsRequest, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoDesiredLRPResponse directly
type DesiredLRPResponse struct {
	Error      *Error      `json:"error,omitempty"`
	DesiredLrp *DesiredLRP `json:"desired_lrp,omitempty"`
}

func (this *DesiredLRPResponse) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*DesiredLRPResponse)
	if !ok {
		that2, ok := that.(DesiredLRPResponse)
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
	if this.DesiredLrp == nil {
		if that1.DesiredLrp != nil {
			return false
		}
	} else if !this.DesiredLrp.Equal(*that1.DesiredLrp) {
		return false
	}
	return true
}
func (m *DesiredLRPResponse) GetError() *Error {
	if m != nil {
		return m.Error
	}
	return nil
}
func (m *DesiredLRPResponse) SetError(value *Error) {
	if m != nil {
		m.Error = value
	}
}
func (m *DesiredLRPResponse) GetDesiredLrp() *DesiredLRP {
	if m != nil {
		return m.DesiredLrp
	}
	return nil
}
func (m *DesiredLRPResponse) SetDesiredLrp(value *DesiredLRP) {
	if m != nil {
		m.DesiredLrp = value
	}
}
func (x *DesiredLRPResponse) ToProto() *ProtoDesiredLRPResponse {
	if x == nil {
		return nil
	}

	proto := &ProtoDesiredLRPResponse{
		Error:      x.Error.ToProto(),
		DesiredLrp: x.DesiredLrp.ToProto(),
	}
	return proto
}

func (x *ProtoDesiredLRPResponse) FromProto() *DesiredLRPResponse {
	if x == nil {
		return nil
	}

	copysafe := &DesiredLRPResponse{
		Error:      x.Error.FromProto(),
		DesiredLrp: x.DesiredLrp.FromProto(),
	}
	return copysafe
}

func DesiredLRPResponseToProtoSlice(values []*DesiredLRPResponse) []*ProtoDesiredLRPResponse {
	if values == nil {
		return nil
	}
	result := make([]*ProtoDesiredLRPResponse, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func DesiredLRPResponseFromProtoSlice(values []*ProtoDesiredLRPResponse) []*DesiredLRPResponse {
	if values == nil {
		return nil
	}
	result := make([]*DesiredLRPResponse, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoDesiredLRPSchedulingInfosResponse directly
type DesiredLRPSchedulingInfosResponse struct {
	Error                     *Error                      `json:"error,omitempty"`
	DesiredLrpSchedulingInfos []*DesiredLRPSchedulingInfo `json:"desired_lrp_scheduling_infos,omitempty"`
}

func (this *DesiredLRPSchedulingInfosResponse) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*DesiredLRPSchedulingInfosResponse)
	if !ok {
		that2, ok := that.(DesiredLRPSchedulingInfosResponse)
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
	if this.DesiredLrpSchedulingInfos == nil {
		if that1.DesiredLrpSchedulingInfos != nil {
			return false
		}
	} else if len(this.DesiredLrpSchedulingInfos) != len(that1.DesiredLrpSchedulingInfos) {
		return false
	}
	for i := range this.DesiredLrpSchedulingInfos {
		if !this.DesiredLrpSchedulingInfos[i].Equal(that1.DesiredLrpSchedulingInfos[i]) {
			return false
		}
	}
	return true
}
func (m *DesiredLRPSchedulingInfosResponse) GetError() *Error {
	if m != nil {
		return m.Error
	}
	return nil
}
func (m *DesiredLRPSchedulingInfosResponse) SetError(value *Error) {
	if m != nil {
		m.Error = value
	}
}
func (m *DesiredLRPSchedulingInfosResponse) GetDesiredLrpSchedulingInfos() []*DesiredLRPSchedulingInfo {
	if m != nil {
		return m.DesiredLrpSchedulingInfos
	}
	return nil
}
func (m *DesiredLRPSchedulingInfosResponse) SetDesiredLrpSchedulingInfos(value []*DesiredLRPSchedulingInfo) {
	if m != nil {
		m.DesiredLrpSchedulingInfos = value
	}
}
func (x *DesiredLRPSchedulingInfosResponse) ToProto() *ProtoDesiredLRPSchedulingInfosResponse {
	if x == nil {
		return nil
	}

	proto := &ProtoDesiredLRPSchedulingInfosResponse{
		Error:                     x.Error.ToProto(),
		DesiredLrpSchedulingInfos: DesiredLRPSchedulingInfoToProtoSlice(x.DesiredLrpSchedulingInfos),
	}
	return proto
}

func (x *ProtoDesiredLRPSchedulingInfosResponse) FromProto() *DesiredLRPSchedulingInfosResponse {
	if x == nil {
		return nil
	}

	copysafe := &DesiredLRPSchedulingInfosResponse{
		Error:                     x.Error.FromProto(),
		DesiredLrpSchedulingInfos: DesiredLRPSchedulingInfoFromProtoSlice(x.DesiredLrpSchedulingInfos),
	}
	return copysafe
}

func DesiredLRPSchedulingInfosResponseToProtoSlice(values []*DesiredLRPSchedulingInfosResponse) []*ProtoDesiredLRPSchedulingInfosResponse {
	if values == nil {
		return nil
	}
	result := make([]*ProtoDesiredLRPSchedulingInfosResponse, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func DesiredLRPSchedulingInfosResponseFromProtoSlice(values []*ProtoDesiredLRPSchedulingInfosResponse) []*DesiredLRPSchedulingInfosResponse {
	if values == nil {
		return nil
	}
	result := make([]*DesiredLRPSchedulingInfosResponse, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoDesiredLRPSchedulingInfoByProcessGuidResponse directly
type DesiredLRPSchedulingInfoByProcessGuidResponse struct {
	Error                    *Error                    `json:"error,omitempty"`
	DesiredLrpSchedulingInfo *DesiredLRPSchedulingInfo `json:"desired_lrp_scheduling_info,omitempty"`
}

func (this *DesiredLRPSchedulingInfoByProcessGuidResponse) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*DesiredLRPSchedulingInfoByProcessGuidResponse)
	if !ok {
		that2, ok := that.(DesiredLRPSchedulingInfoByProcessGuidResponse)
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
	if this.DesiredLrpSchedulingInfo == nil {
		if that1.DesiredLrpSchedulingInfo != nil {
			return false
		}
	} else if !this.DesiredLrpSchedulingInfo.Equal(*that1.DesiredLrpSchedulingInfo) {
		return false
	}
	return true
}
func (m *DesiredLRPSchedulingInfoByProcessGuidResponse) GetError() *Error {
	if m != nil {
		return m.Error
	}
	return nil
}
func (m *DesiredLRPSchedulingInfoByProcessGuidResponse) SetError(value *Error) {
	if m != nil {
		m.Error = value
	}
}
func (m *DesiredLRPSchedulingInfoByProcessGuidResponse) GetDesiredLrpSchedulingInfo() *DesiredLRPSchedulingInfo {
	if m != nil {
		return m.DesiredLrpSchedulingInfo
	}
	return nil
}
func (m *DesiredLRPSchedulingInfoByProcessGuidResponse) SetDesiredLrpSchedulingInfo(value *DesiredLRPSchedulingInfo) {
	if m != nil {
		m.DesiredLrpSchedulingInfo = value
	}
}
func (x *DesiredLRPSchedulingInfoByProcessGuidResponse) ToProto() *ProtoDesiredLRPSchedulingInfoByProcessGuidResponse {
	if x == nil {
		return nil
	}

	proto := &ProtoDesiredLRPSchedulingInfoByProcessGuidResponse{
		Error:                    x.Error.ToProto(),
		DesiredLrpSchedulingInfo: x.DesiredLrpSchedulingInfo.ToProto(),
	}
	return proto
}

func (x *ProtoDesiredLRPSchedulingInfoByProcessGuidResponse) FromProto() *DesiredLRPSchedulingInfoByProcessGuidResponse {
	if x == nil {
		return nil
	}

	copysafe := &DesiredLRPSchedulingInfoByProcessGuidResponse{
		Error:                    x.Error.FromProto(),
		DesiredLrpSchedulingInfo: x.DesiredLrpSchedulingInfo.FromProto(),
	}
	return copysafe
}

func DesiredLRPSchedulingInfoByProcessGuidResponseToProtoSlice(values []*DesiredLRPSchedulingInfoByProcessGuidResponse) []*ProtoDesiredLRPSchedulingInfoByProcessGuidResponse {
	if values == nil {
		return nil
	}
	result := make([]*ProtoDesiredLRPSchedulingInfoByProcessGuidResponse, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func DesiredLRPSchedulingInfoByProcessGuidResponseFromProtoSlice(values []*ProtoDesiredLRPSchedulingInfoByProcessGuidResponse) []*DesiredLRPSchedulingInfoByProcessGuidResponse {
	if values == nil {
		return nil
	}
	result := make([]*DesiredLRPSchedulingInfoByProcessGuidResponse, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoDesiredLRPByProcessGuidRequest directly
type DesiredLRPByProcessGuidRequest struct {
	ProcessGuid string `json:"process_guid"`
}

func (this *DesiredLRPByProcessGuidRequest) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*DesiredLRPByProcessGuidRequest)
	if !ok {
		that2, ok := that.(DesiredLRPByProcessGuidRequest)
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
func (m *DesiredLRPByProcessGuidRequest) GetProcessGuid() string {
	if m != nil {
		return m.ProcessGuid
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *DesiredLRPByProcessGuidRequest) SetProcessGuid(value string) {
	if m != nil {
		m.ProcessGuid = value
	}
}
func (x *DesiredLRPByProcessGuidRequest) ToProto() *ProtoDesiredLRPByProcessGuidRequest {
	if x == nil {
		return nil
	}

	proto := &ProtoDesiredLRPByProcessGuidRequest{
		ProcessGuid: x.ProcessGuid,
	}
	return proto
}

func (x *ProtoDesiredLRPByProcessGuidRequest) FromProto() *DesiredLRPByProcessGuidRequest {
	if x == nil {
		return nil
	}

	copysafe := &DesiredLRPByProcessGuidRequest{
		ProcessGuid: x.ProcessGuid,
	}
	return copysafe
}

func DesiredLRPByProcessGuidRequestToProtoSlice(values []*DesiredLRPByProcessGuidRequest) []*ProtoDesiredLRPByProcessGuidRequest {
	if values == nil {
		return nil
	}
	result := make([]*ProtoDesiredLRPByProcessGuidRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func DesiredLRPByProcessGuidRequestFromProtoSlice(values []*ProtoDesiredLRPByProcessGuidRequest) []*DesiredLRPByProcessGuidRequest {
	if values == nil {
		return nil
	}
	result := make([]*DesiredLRPByProcessGuidRequest, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoDesireLRPRequest directly
type DesireLRPRequest struct {
	DesiredLrp *DesiredLRP `json:"desired_lrp,omitempty"`
}

func (this *DesireLRPRequest) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*DesireLRPRequest)
	if !ok {
		that2, ok := that.(DesireLRPRequest)
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

	if this.DesiredLrp == nil {
		if that1.DesiredLrp != nil {
			return false
		}
	} else if !this.DesiredLrp.Equal(*that1.DesiredLrp) {
		return false
	}
	return true
}
func (m *DesireLRPRequest) GetDesiredLrp() *DesiredLRP {
	if m != nil {
		return m.DesiredLrp
	}
	return nil
}
func (m *DesireLRPRequest) SetDesiredLrp(value *DesiredLRP) {
	if m != nil {
		m.DesiredLrp = value
	}
}
func (x *DesireLRPRequest) ToProto() *ProtoDesireLRPRequest {
	if x == nil {
		return nil
	}

	proto := &ProtoDesireLRPRequest{
		DesiredLrp: x.DesiredLrp.ToProto(),
	}
	return proto
}

func (x *ProtoDesireLRPRequest) FromProto() *DesireLRPRequest {
	if x == nil {
		return nil
	}

	copysafe := &DesireLRPRequest{
		DesiredLrp: x.DesiredLrp.FromProto(),
	}
	return copysafe
}

func DesireLRPRequestToProtoSlice(values []*DesireLRPRequest) []*ProtoDesireLRPRequest {
	if values == nil {
		return nil
	}
	result := make([]*ProtoDesireLRPRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func DesireLRPRequestFromProtoSlice(values []*ProtoDesireLRPRequest) []*DesireLRPRequest {
	if values == nil {
		return nil
	}
	result := make([]*DesireLRPRequest, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoUpdateDesiredLRPRequest directly
type UpdateDesiredLRPRequest struct {
	ProcessGuid string            `json:"process_guid"`
	Update      *DesiredLRPUpdate `json:"update,omitempty"`
}

func (this *UpdateDesiredLRPRequest) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*UpdateDesiredLRPRequest)
	if !ok {
		that2, ok := that.(UpdateDesiredLRPRequest)
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
	if this.Update == nil {
		if that1.Update != nil {
			return false
		}
	} else if !this.Update.Equal(*that1.Update) {
		return false
	}
	return true
}
func (m *UpdateDesiredLRPRequest) GetProcessGuid() string {
	if m != nil {
		return m.ProcessGuid
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *UpdateDesiredLRPRequest) SetProcessGuid(value string) {
	if m != nil {
		m.ProcessGuid = value
	}
}
func (m *UpdateDesiredLRPRequest) GetUpdate() *DesiredLRPUpdate {
	if m != nil {
		return m.Update
	}
	return nil
}
func (m *UpdateDesiredLRPRequest) SetUpdate(value *DesiredLRPUpdate) {
	if m != nil {
		m.Update = value
	}
}
func (x *UpdateDesiredLRPRequest) ToProto() *ProtoUpdateDesiredLRPRequest {
	if x == nil {
		return nil
	}

	proto := &ProtoUpdateDesiredLRPRequest{
		ProcessGuid: x.ProcessGuid,
		Update:      x.Update.ToProto(),
	}
	return proto
}

func (x *ProtoUpdateDesiredLRPRequest) FromProto() *UpdateDesiredLRPRequest {
	if x == nil {
		return nil
	}

	copysafe := &UpdateDesiredLRPRequest{
		ProcessGuid: x.ProcessGuid,
		Update:      x.Update.FromProto(),
	}
	return copysafe
}

func UpdateDesiredLRPRequestToProtoSlice(values []*UpdateDesiredLRPRequest) []*ProtoUpdateDesiredLRPRequest {
	if values == nil {
		return nil
	}
	result := make([]*ProtoUpdateDesiredLRPRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func UpdateDesiredLRPRequestFromProtoSlice(values []*ProtoUpdateDesiredLRPRequest) []*UpdateDesiredLRPRequest {
	if values == nil {
		return nil
	}
	result := make([]*UpdateDesiredLRPRequest, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoRemoveDesiredLRPRequest directly
type RemoveDesiredLRPRequest struct {
	ProcessGuid string `json:"process_guid"`
}

func (this *RemoveDesiredLRPRequest) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*RemoveDesiredLRPRequest)
	if !ok {
		that2, ok := that.(RemoveDesiredLRPRequest)
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
func (m *RemoveDesiredLRPRequest) GetProcessGuid() string {
	if m != nil {
		return m.ProcessGuid
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *RemoveDesiredLRPRequest) SetProcessGuid(value string) {
	if m != nil {
		m.ProcessGuid = value
	}
}
func (x *RemoveDesiredLRPRequest) ToProto() *ProtoRemoveDesiredLRPRequest {
	if x == nil {
		return nil
	}

	proto := &ProtoRemoveDesiredLRPRequest{
		ProcessGuid: x.ProcessGuid,
	}
	return proto
}

func (x *ProtoRemoveDesiredLRPRequest) FromProto() *RemoveDesiredLRPRequest {
	if x == nil {
		return nil
	}

	copysafe := &RemoveDesiredLRPRequest{
		ProcessGuid: x.ProcessGuid,
	}
	return copysafe
}

func RemoveDesiredLRPRequestToProtoSlice(values []*RemoveDesiredLRPRequest) []*ProtoRemoveDesiredLRPRequest {
	if values == nil {
		return nil
	}
	result := make([]*ProtoRemoveDesiredLRPRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func RemoveDesiredLRPRequestFromProtoSlice(values []*ProtoRemoveDesiredLRPRequest) []*RemoveDesiredLRPRequest {
	if values == nil {
		return nil
	}
	result := make([]*RemoveDesiredLRPRequest, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}
