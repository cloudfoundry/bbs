package format

import (
	"errors"

	"code.cloudfoundry.org/lager"
	"github.com/gogo/protobuf/proto"
)

type EnvelopeFormat byte

const (
	PROTO EnvelopeFormat = 2
)

const EnvelopeOffset int = 2

func UnmarshalEnvelope(logger lager.Logger, unencodedPayload []byte, model Versioner) error {
	protoModel, ok := model.(ProtoVersioner)
	if !ok {
		return errors.New("Model object incompatible with envelope format")
	}
	return UnmarshalProto(logger, unencodedPayload[EnvelopeOffset:], protoModel)
}

// DEPRECATED
// dummy version for backward compatability. old BBS used to serialize proto
// messages with a 2-byte header that has the envelope format (i.e. PROTO) and
// the version of the model (e.g. 0, 1 or 2). Adding the version was a
// pre-mature optimization that we decided to get rid of in #133215113. That
// said, we have the ensure the header is a 2-byte to avoid breaking older BBS
const version = 0

func MarshalEnvelope(model Versioner) ([]byte, error) {
	var payload []byte
	var err error

	protoModel, ok := model.(ProtoVersioner)
	if !ok {
		return nil, errors.New("Model object incompatible with envelope format")
	}
	payload, err = MarshalProto(protoModel)

	if err != nil {
		return nil, err
	}

	data := make([]byte, 0, len(payload)+EnvelopeOffset)
	data = append(data, byte(PROTO), byte(version))
	data = append(data, payload...)

	return data, nil
}

func UnmarshalProto(logger lager.Logger, marshaledPayload []byte, model ProtoVersioner) error {
	err := proto.Unmarshal(marshaledPayload, model)
	if err != nil {
		logger.Error("failed-to-proto-unmarshal-payload", err)
		return err
	}
	return nil
}

func MarshalProto(v ProtoVersioner) ([]byte, error) {
	bytes, err := proto.Marshal(v)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}
