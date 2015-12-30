package models_test

import (
	"encoding/json"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/gogo/protobuf/proto"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RootFSProviders", func() {
	var update models.CellPresence
	var aJson models.CellPresence
	var aProto models.CellPresence

	itSerializes := func(rootFSProviders *models.RootFSProviders) {
		BeforeEach(func() {
			update = models.CellPresence{
				RootfsProviders: *rootFSProviders,
			}

			b, err := json.Marshal(update)
			Expect(err).NotTo(HaveOccurred())
			err = json.Unmarshal(b, &aJson)
			Expect(err).NotTo(HaveOccurred())

			b, err = proto.Marshal(&update)
			Expect(err).NotTo(HaveOccurred())
			err = proto.Unmarshal(b, &aProto)
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

	itSerializes(&models.RootFSProviders{
		"provider1": &models.Providers{ProvidersList: []string{"test1", "test2"}},
		"provider2": &models.Providers{ProvidersList: []string{"test3", "test4"}},
	})
})
