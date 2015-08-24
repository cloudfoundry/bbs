package codec

import (
	"encoding/base64"
	"fmt"
)

type Encoder interface {
	Encode(data []byte) ([]byte, error)
}

type Decoder interface {
	Decode(data []byte) ([]byte, error)
}

type Codec interface {
	Encoder
	Decoder
}

type Kind [2]byte

type Codecs struct {
	codecs      map[Kind]Codec
	encoder     Encoder
	encoderKind Kind
}

var (
	NONE      Kind = [2]byte{}
	UNENCODED Kind = [2]byte{'0', '0'}
	BASE64    Kind = [2]byte{'0', '1'}
)

type IdentityCodec struct{}

func (id *IdentityCodec) Encode(data []byte) ([]byte, error) { return data, nil }
func (id *IdentityCodec) Decode(data []byte) ([]byte, error) { return data, nil }

type Base64Codec struct{}

func (c *Base64Codec) Encode(payload []byte) ([]byte, error) {
	encodedLen := base64.StdEncoding.EncodedLen(len(payload))
	data := make([]byte, encodedLen)
	base64.StdEncoding.Encode(data, payload)
	return data, nil
}

func (c *Base64Codec) Decode(data []byte) ([]byte, error) {
	decodedLen := base64.StdEncoding.DecodedLen(len(data))
	payload := make([]byte, decodedLen)
	n, err := base64.StdEncoding.Decode(payload, data)
	return payload[:n], err
}

func NewCodecs(encoderKind Kind) *Codecs {
	codecs := map[Kind]Codec{
		NONE:      nil,
		UNENCODED: &IdentityCodec{},
		BASE64:    &Base64Codec{},
	}

	return &Codecs{
		codecs:      codecs,
		encoder:     codecs[encoderKind],
		encoderKind: encoderKind,
	}
}

func (c *Codecs) Encode(payload []byte) ([]byte, error) {
	if c.encoder == nil {
		return payload, nil
	}

	encoded, err := c.encoder.Encode(payload)
	if err != nil {
		return nil, err
	}

	return append(c.encoderKind[:], encoded...), nil
}

func (c *Codecs) Decode(data []byte) ([]byte, error) {
	if !isEncoded(data) {
		return data, nil
	}

	kind := Kind{data[0], data[1]}

	codec := c.codecs[kind]
	if codec == nil {
		return nil, fmt.Errorf("Unknown kind: %v", kind)
	}

	return codec.Decode(data[2:])
}

func isEncoded(data []byte) bool {
	if len(data) < len(UNENCODED) {
		return false
	}

	if data[0] < '0' || data[0] > '9' {
		return false
	}

	return true
}
