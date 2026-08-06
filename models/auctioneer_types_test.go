package models_test

import (
	"encoding/json"

	"code.cloudfoundry.org/bbs/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TaskStartRequest", func() {
	Describe("JSON wire format", func() {
		var task models.SchedulingTask

		BeforeEach(func() {
			task = models.NewSchedulingTask(
				"task-guid-1",
				"cf-tasks",
				models.NewResource(256, 1024, 0),
				models.NewPlacementConstraint("cflinuxfs4", nil, nil),
			)
		})

		It("marshals as a flat object, not nested under a \"Task\" key", func() {
			req := models.NewTaskStartRequest(task)

			payload, err := json.Marshal(req)
			Expect(err).NotTo(HaveOccurred())

			var raw map[string]interface{}
			Expect(json.Unmarshal(payload, &raw)).To(Succeed())

			Expect(raw).To(HaveKeyWithValue("TaskGuid", "task-guid-1"))
			Expect(raw).NotTo(HaveKey("Task"))
		})

		It("round-trips through the flat wire format", func() {
			req := models.NewTaskStartRequest(task)
			payload, err := json.Marshal(req)
			Expect(err).NotTo(HaveOccurred())

			var decoded models.TaskStartRequest
			Expect(json.Unmarshal(payload, &decoded)).To(Succeed())
			Expect(decoded.Task.TaskGuid).To(Equal("task-guid-1"))
			Expect(decoded.Validate()).To(Succeed())
		})

		It("decodes the legacy pre-8b9ea83 flat auctioneer wire format", func() {
			legacyPayload := []byte(`{"TaskGuid":"task-guid-1","Domain":"cf-tasks","RootFs":"cflinuxfs4","MemoryMB":256,"DiskMB":1024,"MaxPids":0}`)

			var decoded models.TaskStartRequest
			Expect(json.Unmarshal(legacyPayload, &decoded)).To(Succeed())
			Expect(decoded.Task.TaskGuid).To(Equal("task-guid-1"))
			Expect(decoded.Validate()).To(Succeed())
		})
	})
})
