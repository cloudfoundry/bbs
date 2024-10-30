// Code generated by protoc-gen-go-bbs. DO NOT EDIT.
// versions:
// - protoc-gen-go-bbs v0.0.1
// - protoc            v5.28.3
// source: certificate_properties.proto

package models

// Prevent copylock errors when using ProtoCertificateProperties directly
type CertificateProperties struct {
	OrganizationalUnit []string `json:"organizational_unit,omitempty"`
}

func (this *CertificateProperties) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*CertificateProperties)
	if !ok {
		that2, ok := that.(CertificateProperties)
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

	if this.OrganizationalUnit == nil {
		if that1.OrganizationalUnit != nil {
			return false
		}
	} else if len(this.OrganizationalUnit) != len(that1.OrganizationalUnit) {
		return false
	}
	for i := range this.OrganizationalUnit {
		if this.OrganizationalUnit[i] != that1.OrganizationalUnit[i] {
			return false
		}
	}
	return true
}
func (m *CertificateProperties) GetOrganizationalUnit() []string {
	if m != nil {
		return m.OrganizationalUnit
	}
	return nil
}
func (m *CertificateProperties) SetOrganizationalUnit(value []string) {
	if m != nil {
		m.OrganizationalUnit = value
	}
}
func (x *CertificateProperties) ToProto() *ProtoCertificateProperties {
	if x == nil {
		return nil
	}

	proto := &ProtoCertificateProperties{
		OrganizationalUnit: x.OrganizationalUnit,
	}
	return proto
}

func (x *ProtoCertificateProperties) FromProto() *CertificateProperties {
	if x == nil {
		return nil
	}

	copysafe := &CertificateProperties{
		OrganizationalUnit: x.OrganizationalUnit,
	}
	return copysafe
}

func CertificatePropertiesToProtoSlice(values []*CertificateProperties) []*ProtoCertificateProperties {
	if values == nil {
		return nil
	}
	result := make([]*ProtoCertificateProperties, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func CertificatePropertiesFromProtoSlice(values []*ProtoCertificateProperties) []*CertificateProperties {
	if values == nil {
		return nil
	}
	result := make([]*CertificateProperties, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}
