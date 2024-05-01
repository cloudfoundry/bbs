// Code generated by protoc-gen-go-bbs. DO NOT EDIT.
// versions:
// - protoc-gen-go-bbs v0.0.1
// - protoc            v5.27.0--rc1
// source: certificate_properties.proto

package models

// Prevent copylock errors when using ProtoCertificateProperties directly
type CertificateProperties struct {
	OrganizationalUnit []string
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
	proto := &ProtoCertificateProperties{
		OrganizationalUnit: x.OrganizationalUnit,
	}
	return proto
}

func CertificatePropertiesProtoMap(values []*CertificateProperties) []*ProtoCertificateProperties {
	result := make([]*ProtoCertificateProperties, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}
