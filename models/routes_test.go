package models_test

import (
	"encoding/json"

	"code.cloudfoundry.org/bbs/models"
	"google.golang.org/protobuf/proto"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Routes", func() {
	var update models.DesiredLRPUpdate
	var aJson models.DesiredLRPUpdate
	var aProto models.DesiredLRPUpdate

	itSerializes := func(routes *models.Routes) {
		BeforeEach(func() {
			update = models.DesiredLRPUpdate{
				Routes: routes,
			}

			b, err := json.Marshal(update)
			Expect(err).NotTo(HaveOccurred())
			err = json.Unmarshal(b, &aJson)
			Expect(err).NotTo(HaveOccurred())

			protoUpdate := update.ToProto()
			b, err = proto.Marshal(protoUpdate)
			Expect(err).NotTo(HaveOccurred())
			protoA := aProto.ToProto()
			err = proto.Unmarshal(b, protoA)
			Expect(err).NotTo(HaveOccurred())
		})

		It("marshals JSON properly", func() {
			Expect(update.Equal(&aJson)).To(BeTrue())
			Expect(update).To(Equal(aJson))
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
