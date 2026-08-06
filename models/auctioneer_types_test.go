package models_test

import (
	"encoding/json"

	"code.cloudfoundry.org/bbs/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TaskStartRequest", func() {
	var request models.TaskStartRequest

	BeforeEach(func() {
		request = models.NewTaskStartRequest(models.NewSchedulingTask(
			"task-guid-123",
			"my-domain",
			models.NewResource(256, 1024, 0),
			models.NewPlacementConstraint("cflinuxfs4", []string{"tag"}, []string{"driver"}),
		))
	})

	Describe("MarshalJSON", func() {
		It("flattens the wrapped SchedulingTask to the top level, instead of nesting it under a \"Task\" key", func() {
			wire, err := json.Marshal(request)
			Expect(err).NotTo(HaveOccurred())

			var flat map[string]interface{}
			Expect(json.Unmarshal(wire, &flat)).To(Succeed())

			Expect(flat).NotTo(HaveKey("Task"))
			Expect(flat["TaskGuid"]).To(Equal("task-guid-123"))
			Expect(flat["Domain"]).To(Equal("my-domain"))
			Expect(flat["RootFs"]).To(Equal("cflinuxfs4"))
			Expect(flat["MemoryMB"]).To(Equal(float64(256)))
			Expect(flat["DiskMB"]).To(Equal(float64(1024)))
		})
	})

	Describe("UnmarshalJSON", func() {
		It("round-trips through its own MarshalJSON", func() {
			wire, err := json.Marshal(request)
			Expect(err).NotTo(HaveOccurred())

			var roundTripped models.TaskStartRequest
			Expect(json.Unmarshal(wire, &roundTripped)).To(Succeed())
			Expect(roundTripped).To(Equal(request))
		})

		It("accepts the flat, pre-move wire format produced by an old auctioneer client", func() {
			flatWire := []byte(`{
				"TaskGuid": "task-guid-123",
				"Domain": "my-domain",
				"RootFs": "cflinuxfs4",
				"PlacementTags": ["tag"],
				"VolumeDrivers": ["driver"],
				"MemoryMB": 256,
				"DiskMB": 1024,
				"MaxPids": 0,
				"state": "Invalid",
				"failed": false
			}`)

			var parsed models.TaskStartRequest
			Expect(json.Unmarshal(flatWire, &parsed)).To(Succeed())
			Expect(parsed).To(Equal(request))
			Expect(parsed.Validate()).NotTo(HaveOccurred())
		})
	})
})
