// Code generated by protoc-gen-gogo.
// source: network.proto
// DO NOT EDIT!

package models

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/protobuf/gogoproto"

import strings "strings"
import reflect "reflect"
import github_com_gogo_protobuf_sortkeys "github.com/gogo/protobuf/sortkeys"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type Network struct {
	Properties map[string]string `protobuf:"bytes,1,rep,name=properties" json:"properties,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *Network) Reset()                    { *m = Network{} }
func (*Network) ProtoMessage()               {}
func (*Network) Descriptor() ([]byte, []int) { return fileDescriptorNetwork, []int{0} }

func (m *Network) GetProperties() map[string]string {
	if m != nil {
		return m.Properties
	}
	return nil
}

func init() {
	proto.RegisterType((*Network)(nil), "models.Network")
}
func (this *Network) Equal(that interface{}) bool {
	if that == nil {
		if this == nil {
			return true
		}
		return false
	}

	that1, ok := that.(*Network)
	if !ok {
		that2, ok := that.(Network)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		if this == nil {
			return true
		}
		return false
	} else if this == nil {
		return false
	}
	if len(this.Properties) != len(that1.Properties) {
		return false
	}
	for i := range this.Properties {
		if this.Properties[i] != that1.Properties[i] {
			return false
		}
	}
	return true
}
func (this *Network) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 5)
	s = append(s, "&models.Network{")
	keysForProperties := make([]string, 0, len(this.Properties))
	for k, _ := range this.Properties {
		keysForProperties = append(keysForProperties, k)
	}
	github_com_gogo_protobuf_sortkeys.Strings(keysForProperties)
	mapStringForProperties := "map[string]string{"
	for _, k := range keysForProperties {
		mapStringForProperties += fmt.Sprintf("%#v: %#v,", k, this.Properties[k])
	}
	mapStringForProperties += "}"
	if this.Properties != nil {
		s = append(s, "Properties: "+mapStringForProperties+",\n")
	}
	s = append(s, "}")
	return strings.Join(s, "")
}
func valueToGoStringNetwork(v interface{}, typ string) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("func(v %v) *%v { return &v } ( %#v )", typ, typ, pv)
}
func (m *Network) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Network) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Properties) > 0 {
		for k, _ := range m.Properties {
			dAtA[i] = 0xa
			i++
			v := m.Properties[k]
			mapSize := 1 + len(k) + sovNetwork(uint64(len(k))) + 1 + len(v) + sovNetwork(uint64(len(v)))
			i = encodeVarintNetwork(dAtA, i, uint64(mapSize))
			dAtA[i] = 0xa
			i++
			i = encodeVarintNetwork(dAtA, i, uint64(len(k)))
			i += copy(dAtA[i:], k)
			dAtA[i] = 0x12
			i++
			i = encodeVarintNetwork(dAtA, i, uint64(len(v)))
			i += copy(dAtA[i:], v)
		}
	}
	return i, nil
}

func encodeFixed64Network(dAtA []byte, offset int, v uint64) int {
	dAtA[offset] = uint8(v)
	dAtA[offset+1] = uint8(v >> 8)
	dAtA[offset+2] = uint8(v >> 16)
	dAtA[offset+3] = uint8(v >> 24)
	dAtA[offset+4] = uint8(v >> 32)
	dAtA[offset+5] = uint8(v >> 40)
	dAtA[offset+6] = uint8(v >> 48)
	dAtA[offset+7] = uint8(v >> 56)
	return offset + 8
}
func encodeFixed32Network(dAtA []byte, offset int, v uint32) int {
	dAtA[offset] = uint8(v)
	dAtA[offset+1] = uint8(v >> 8)
	dAtA[offset+2] = uint8(v >> 16)
	dAtA[offset+3] = uint8(v >> 24)
	return offset + 4
}
func encodeVarintNetwork(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *Network) Size() (n int) {
	var l int
	_ = l
	if len(m.Properties) > 0 {
		for k, v := range m.Properties {
			_ = k
			_ = v
			mapEntrySize := 1 + len(k) + sovNetwork(uint64(len(k))) + 1 + len(v) + sovNetwork(uint64(len(v)))
			n += mapEntrySize + 1 + sovNetwork(uint64(mapEntrySize))
		}
	}
	return n
}

func sovNetwork(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozNetwork(x uint64) (n int) {
	return sovNetwork(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (this *Network) String() string {
	if this == nil {
		return "nil"
	}
	keysForProperties := make([]string, 0, len(this.Properties))
	for k, _ := range this.Properties {
		keysForProperties = append(keysForProperties, k)
	}
	github_com_gogo_protobuf_sortkeys.Strings(keysForProperties)
	mapStringForProperties := "map[string]string{"
	for _, k := range keysForProperties {
		mapStringForProperties += fmt.Sprintf("%v: %v,", k, this.Properties[k])
	}
	mapStringForProperties += "}"
	s := strings.Join([]string{`&Network{`,
		`Properties:` + mapStringForProperties + `,`,
		`}`,
	}, "")
	return s
}
func valueToStringNetwork(v interface{}) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("*%v", pv)
}
func (m *Network) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowNetwork
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Network: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Network: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Properties", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNetwork
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthNetwork
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			var keykey uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNetwork
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				keykey |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			var stringLenmapkey uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNetwork
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLenmapkey |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLenmapkey := int(stringLenmapkey)
			if intStringLenmapkey < 0 {
				return ErrInvalidLengthNetwork
			}
			postStringIndexmapkey := iNdEx + intStringLenmapkey
			if postStringIndexmapkey > l {
				return io.ErrUnexpectedEOF
			}
			mapkey := string(dAtA[iNdEx:postStringIndexmapkey])
			iNdEx = postStringIndexmapkey
			if m.Properties == nil {
				m.Properties = make(map[string]string)
			}
			if iNdEx < postIndex {
				var valuekey uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowNetwork
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					valuekey |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				var stringLenmapvalue uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowNetwork
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					stringLenmapvalue |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				intStringLenmapvalue := int(stringLenmapvalue)
				if intStringLenmapvalue < 0 {
					return ErrInvalidLengthNetwork
				}
				postStringIndexmapvalue := iNdEx + intStringLenmapvalue
				if postStringIndexmapvalue > l {
					return io.ErrUnexpectedEOF
				}
				mapvalue := string(dAtA[iNdEx:postStringIndexmapvalue])
				iNdEx = postStringIndexmapvalue
				m.Properties[mapkey] = mapvalue
			} else {
				var mapvalue string
				m.Properties[mapkey] = mapvalue
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipNetwork(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthNetwork
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipNetwork(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowNetwork
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowNetwork
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowNetwork
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthNetwork
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowNetwork
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipNetwork(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthNetwork = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowNetwork   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("network.proto", fileDescriptorNetwork) }

var fileDescriptorNetwork = []byte{
	// 240 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0xcd, 0x4b, 0x2d, 0x29,
	0xcf, 0x2f, 0xca, 0xd6, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0xcb, 0xcd, 0x4f, 0x49, 0xcd,
	0x29, 0x96, 0xd2, 0x4d, 0xcf, 0x2c, 0xc9, 0x28, 0x4d, 0xd2, 0x4b, 0xce, 0xcf, 0xd5, 0x4f, 0xcf,
	0x4f, 0xcf, 0xd7, 0x07, 0x4b, 0x27, 0x95, 0xa6, 0x81, 0x79, 0x60, 0x0e, 0x98, 0x05, 0xd1, 0xa6,
	0xb4, 0x9e, 0x91, 0x8b, 0xdd, 0x0f, 0x62, 0x90, 0x50, 0x24, 0x17, 0x57, 0x41, 0x51, 0x7e, 0x41,
	0x6a, 0x51, 0x49, 0x66, 0x6a, 0xb1, 0x04, 0xa3, 0x02, 0xb3, 0x06, 0xb7, 0x91, 0xbc, 0x1e, 0xc4,
	0x5c, 0x3d, 0xa8, 0x22, 0xbd, 0x00, 0xb8, 0x0a, 0xd7, 0xbc, 0x92, 0xa2, 0x4a, 0x27, 0x89, 0x57,
	0xf7, 0xe4, 0x45, 0x10, 0xda, 0x74, 0xf2, 0x73, 0x33, 0x4b, 0x52, 0x73, 0x0b, 0x4a, 0x2a, 0x83,
	0x90, 0x0c, 0x93, 0xf2, 0xe4, 0xe2, 0x47, 0xd3, 0x28, 0x24, 0xc6, 0xc5, 0x9c, 0x9d, 0x5a, 0x29,
	0xc1, 0xa8, 0xc0, 0xa8, 0xc1, 0xe9, 0xc4, 0x72, 0xe2, 0x9e, 0x3c, 0x43, 0x10, 0x48, 0x40, 0x48,
	0x8a, 0x8b, 0xb5, 0x2c, 0x31, 0xa7, 0x34, 0x55, 0x82, 0x09, 0x49, 0x06, 0x22, 0x64, 0xc5, 0x64,
	0xc1, 0xe8, 0xa4, 0x73, 0xe1, 0xa1, 0x1c, 0xe3, 0x8d, 0x87, 0x72, 0x0c, 0x1f, 0x1e, 0xca, 0x31,
	0x36, 0x3c, 0x92, 0x63, 0x5c, 0xf1, 0x48, 0x8e, 0xf1, 0xc4, 0x23, 0x39, 0xc6, 0x0b, 0x8f, 0xe4,
	0x18, 0x1f, 0x3c, 0x92, 0x63, 0x7c, 0xf1, 0x48, 0x8e, 0xe1, 0xc3, 0x23, 0x39, 0xc6, 0x09, 0x8f,
	0xe5, 0x18, 0x00, 0x01, 0x00, 0x00, 0xff, 0xff, 0x9d, 0x60, 0xf7, 0x51, 0x26, 0x01, 0x00, 0x00,
}
