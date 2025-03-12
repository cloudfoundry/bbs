// Code generated by protoc-gen-go-bbs. DO NOT EDIT.
// versions:
// - protoc-gen-go-bbs v0.0.1
// - protoc            v6.30.0
// source: domain.proto

package models

// Prevent copylock errors when using ProtoDomainsResponse directly
type DomainsResponse struct {
	Error   *Error   `json:"error,omitempty"`
	Domains []string `json:"domains,omitempty"`
}

func (this *DomainsResponse) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*DomainsResponse)
	if !ok {
		that2, ok := that.(DomainsResponse)
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
	if this.Domains == nil {
		if that1.Domains != nil {
			return false
		}
	} else if len(this.Domains) != len(that1.Domains) {
		return false
	}
	for i := range this.Domains {
		if this.Domains[i] != that1.Domains[i] {
			return false
		}
	}
	return true
}
func (m *DomainsResponse) GetError() *Error {
	if m != nil {
		return m.Error
	}
	return nil
}
func (m *DomainsResponse) SetError(value *Error) {
	if m != nil {
		m.Error = value
	}
}
func (m *DomainsResponse) GetDomains() []string {
	if m != nil {
		return m.Domains
	}
	return nil
}
func (m *DomainsResponse) SetDomains(value []string) {
	if m != nil {
		m.Domains = value
	}
}
func (x *DomainsResponse) ToProto() *ProtoDomainsResponse {
	if x == nil {
		return nil
	}

	proto := &ProtoDomainsResponse{
		Error:   x.Error.ToProto(),
		Domains: x.Domains,
	}
	return proto
}

func (x *ProtoDomainsResponse) FromProto() *DomainsResponse {
	if x == nil {
		return nil
	}

	copysafe := &DomainsResponse{
		Error:   x.Error.FromProto(),
		Domains: x.Domains,
	}
	return copysafe
}

func DomainsResponseToProtoSlice(values []*DomainsResponse) []*ProtoDomainsResponse {
	if values == nil {
		return nil
	}
	result := make([]*ProtoDomainsResponse, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func DomainsResponseFromProtoSlice(values []*ProtoDomainsResponse) []*DomainsResponse {
	if values == nil {
		return nil
	}
	result := make([]*DomainsResponse, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoUpsertDomainResponse directly
type UpsertDomainResponse struct {
	Error *Error `json:"error,omitempty"`
}

func (this *UpsertDomainResponse) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*UpsertDomainResponse)
	if !ok {
		that2, ok := that.(UpsertDomainResponse)
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
func (m *UpsertDomainResponse) GetError() *Error {
	if m != nil {
		return m.Error
	}
	return nil
}
func (m *UpsertDomainResponse) SetError(value *Error) {
	if m != nil {
		m.Error = value
	}
}
func (x *UpsertDomainResponse) ToProto() *ProtoUpsertDomainResponse {
	if x == nil {
		return nil
	}

	proto := &ProtoUpsertDomainResponse{
		Error: x.Error.ToProto(),
	}
	return proto
}

func (x *ProtoUpsertDomainResponse) FromProto() *UpsertDomainResponse {
	if x == nil {
		return nil
	}

	copysafe := &UpsertDomainResponse{
		Error: x.Error.FromProto(),
	}
	return copysafe
}

func UpsertDomainResponseToProtoSlice(values []*UpsertDomainResponse) []*ProtoUpsertDomainResponse {
	if values == nil {
		return nil
	}
	result := make([]*ProtoUpsertDomainResponse, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func UpsertDomainResponseFromProtoSlice(values []*ProtoUpsertDomainResponse) []*UpsertDomainResponse {
	if values == nil {
		return nil
	}
	result := make([]*UpsertDomainResponse, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoUpsertDomainRequest directly
type UpsertDomainRequest struct {
	Domain string `json:"domain"`
	Ttl    uint32 `json:"ttl"`
}

func (this *UpsertDomainRequest) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*UpsertDomainRequest)
	if !ok {
		that2, ok := that.(UpsertDomainRequest)
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
	if this.Ttl != that1.Ttl {
		return false
	}
	return true
}
func (m *UpsertDomainRequest) GetDomain() string {
	if m != nil {
		return m.Domain
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *UpsertDomainRequest) SetDomain(value string) {
	if m != nil {
		m.Domain = value
	}
}
func (m *UpsertDomainRequest) GetTtl() uint32 {
	if m != nil {
		return m.Ttl
	}
	var defaultValue uint32
	defaultValue = 0
	return defaultValue
}
func (m *UpsertDomainRequest) SetTtl(value uint32) {
	if m != nil {
		m.Ttl = value
	}
}
func (x *UpsertDomainRequest) ToProto() *ProtoUpsertDomainRequest {
	if x == nil {
		return nil
	}

	proto := &ProtoUpsertDomainRequest{
		Domain: x.Domain,
		Ttl:    x.Ttl,
	}
	return proto
}

func (x *ProtoUpsertDomainRequest) FromProto() *UpsertDomainRequest {
	if x == nil {
		return nil
	}

	copysafe := &UpsertDomainRequest{
		Domain: x.Domain,
		Ttl:    x.Ttl,
	}
	return copysafe
}

func UpsertDomainRequestToProtoSlice(values []*UpsertDomainRequest) []*ProtoUpsertDomainRequest {
	if values == nil {
		return nil
	}
	result := make([]*ProtoUpsertDomainRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func UpsertDomainRequestFromProtoSlice(values []*ProtoUpsertDomainRequest) []*UpsertDomainRequest {
	if values == nil {
		return nil
	}
	result := make([]*UpsertDomainRequest, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}
