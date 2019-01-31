package models_test

import (
	"code.cloudfoundry.org/bbs/models"
	proto "github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Proto Version Upgrade", func() {
	// The only potential breaking change we have when upgrading from proto2 to
	// proto3 as far as wire encoding is concerned is the
	// "oneof" type in proto3. We can fully exercise testing with just
	// the ActualLRPsRequest since that message uses a "oneof".
	var (
		oldMessage models.ActualLRPsRequestProto2
		newMessage models.ActualLRPsRequest
	)

	BeforeEach(func() {
		var index int32 = 2
		oldMessage = models.ActualLRPsRequestProto2{
			Domain:      "foobar",
			CellId:      "some-cell",
			ProcessGuid: "some-guid",
			Index:       &index,
		}

		newMessage = models.ActualLRPsRequest{
			Domain:        "foobar",
			CellId:        "some-cell",
			ProcessGuid:   "some-guid",
			OptionalIndex: &models.ActualLRPsRequest_Index{Index: 2},
		}
	})

	It("can convert proto2 messages to proto3 messages over the wire", func() {
		// This is plausible if the client is making requests with proto2 messages
		// and the server is decoding requests as proto3 messages.
		protoBytes, err := proto.Marshal(&oldMessage)
		Expect(err).NotTo(HaveOccurred())
		var target models.ActualLRPsRequest
		Expect(proto.Unmarshal(protoBytes, &target)).To(Succeed())
		Expect(target).To(Equal(newMessage))
	})

	It("can downconvert proto3 messages to proto2 messages over the wire", func() {
		// This is plausible if the server is encoding responses as proto3 messages
		// and the client is decoding responses as proto2 messages.
		protoBytes, err := proto.Marshal(&newMessage)
		Expect(err).NotTo(HaveOccurred())
		var target models.ActualLRPsRequestProto2
		Expect(proto.Unmarshal(protoBytes, &target)).To(Succeed())
		Expect(target).To(Equal(oldMessage))
	})
})
