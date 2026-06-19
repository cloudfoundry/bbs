package format_test

import (
	"code.cloudfoundry.org/bbs/format"
	"code.cloudfoundry.org/lager/v3/lagertest"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Envelope", func() {
	var logger *lagertest.TestLogger

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")
	})

	Describe("Marshal", func() {
		It("can successfully marshal a protobuf message envelope", func() {
			msg := &types.StringValue{Value: "test-message"}
			encoded, err := format.MarshalEnvelope(msg)
			Expect(err).NotTo(HaveOccurred())

			Expect(format.EnvelopeFormat(encoded[0])).To(Equal(format.PROTO))

			var newMsg types.StringValue
			modelErr := proto.Unmarshal(encoded[2:], &newMsg)
			Expect(modelErr).To(BeNil())

			Expect(newMsg.Value).To(Equal(msg.Value))
		})
	})

	Describe("Unmarshal", func() {
		It("can marshal and unmarshal a protobuf message without losing data", func() {
			msg := &types.StringValue{Value: "test-message"}
			payload, err := format.MarshalEnvelope(msg)
			Expect(err).NotTo(HaveOccurred())

			resultingMsg := new(types.StringValue)
			err = format.UnmarshalEnvelope(logger, payload, resultingMsg)
			Expect(err).NotTo(HaveOccurred())

			Expect(resultingMsg.Value).To(Equal(msg.Value))
		})

		It("returns an error when the protobuf payload is invalid", func() {
			msg := &types.StringValue{Value: "test"}
			payload := []byte{byte(format.PROTO), byte(format.V0), 'f', 'o', 'o'}
			err := format.UnmarshalEnvelope(logger, payload, msg)
			Expect(err).To(HaveOccurred())
		})
	})
})
