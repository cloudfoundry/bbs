package models_test

import (
	"encoding/json"

	"code.cloudfoundry.org/bbs/models"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Routes", func() {
	var update models.DesiredLRPUpdate
	var aJson models.ProtoDesiredLRPUpdate
	var aProto models.DesiredLRPUpdate
	var resultProto models.ProtoDesiredLRPUpdate

	itSerializes := func(routes *models.Routes) {
		BeforeEach(func() {
			update = models.DesiredLRPUpdate{
				Routes: routes,
			}
			/*
				The point of these tests is to go from non-proto struct
				to JSON/Protobuf (binary) representation and back.
				With the new protobuf requirements we have to add a step
				to convert to the Proto struct before we can get the
				Proto binary representation.

				Old way:
				DesiredLRPUpdate -> Protobuf binary -> DesiredLRPUpdate

				New way:
				DesiredLRPUpdate -> ProtoDesiredLRPUpdate -> Protobuf binary -> ProtoDesiredLRPUpdate -> DesiredLRPUpdate

				2024-05-15: It remains to be seen if this extra layer is going to cause performance issues
			*/

			b, err := protojson.Marshal(update.ToProto())
			Expect(err).NotTo(HaveOccurred())
			err = protojson.Unmarshal(b, &aJson)
			Expect(err).NotTo(HaveOccurred())

			protoUpdate := update.ToProto()
			b, err = proto.Marshal(protoUpdate)
			Expect(err).NotTo(HaveOccurred())
			err = proto.Unmarshal(b, &resultProto)
			Expect(err).NotTo(HaveOccurred())
			aProto = *resultProto.FromProto() // make sure we convert back to non-proto
		})

		It("marshals JSON properly", func() {
			Expect(update.Equal(aJson.FromProto())).To(BeTrue())
			Expect(update).To(Equal(*aJson.FromProto()))
		})

		It("marshals Proto properly", func() {
			Expect(update.Equal(&aProto)).To(BeTrue())
			Expect(update).To(Equal(aProto))
		})
	}

	itSerializes(nil)
	itSerializes(&models.Routes{
		"abc": &(json.RawMessage{'"', 'd', '"'}),
		"def": &(json.RawMessage{'"', 'g', '"'}),
	})
})
