package models_test

import (
	"encoding/json"
	"log"

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
				Routes: *routes.ToProto(),
			}
			log.Printf("INPUT routes: %+v\n", routes)
			log.Printf("INPUT update: %+v\n", update)
			log.Printf("INPUT update.routes: %+v\n", update.Routes)
			// for k, v := range *update.Routes {
			// 	log.Printf("key: %+v, value: %+v", k, string(*v))
			// }
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

			log.Printf("update.ToProto(): %+v\n", update.ToProto())
			b, err := protojson.Marshal(update.ToProto())
			Expect(err).NotTo(HaveOccurred())
			err = protojson.Unmarshal(b, &aJson)
			log.Printf("JSON-aJson: %+v\n", aJson.FromProto())
			log.Printf("JSON-aJson routes: %+v\n", aJson.FromProto().Routes)
			Expect(err).NotTo(HaveOccurred())

			protoUpdate := update.ToProto()
			log.Printf("protoUpdate: %+v\n", protoUpdate)
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
