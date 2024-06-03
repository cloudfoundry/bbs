package format_test

import (
	"code.cloudfoundry.org/bbs/format"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/models/test/model_helpers"
	"code.cloudfoundry.org/lager/v3/lagertest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"
)

var _ = Describe("Envelope", func() {
	var logger *lagertest.TestLogger

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")
	})

	Describe("Marshal", func() {
		It("can successfully marshal a model object envelope", func() {
			task := model_helpers.NewValidTask("some-guid")
			encoded, err := format.MarshalEnvelope(task.ToProto())
			Expect(err).NotTo(HaveOccurred())

			Expect(format.EnvelopeFormat(encoded[0])).To(Equal(format.PROTO))

			var newTask models.ProtoTask
			modelErr := proto.Unmarshal(encoded[2:], &newTask)
			Expect(modelErr).To(BeNil())

			Expect(task).To(Equal(newTask.FromProto()))
		})
	})

	Describe("Unmarshal", func() {
		It("can marshal and unmarshal a task without losing data", func() {
			task := model_helpers.NewValidTask("some-guid")
			payload, err := format.MarshalEnvelope(task.ToProto())
			Expect(err).NotTo(HaveOccurred())

			resultingTask := new(models.ProtoTask)
			err = format.UnmarshalEnvelope(logger, payload, resultingTask)
			Expect(err).NotTo(HaveOccurred())

			Expect(resultingTask.FromProto()).To(BeEquivalentTo(task))
		})

		It("returns an error when the protobuf payload is invalid", func() {
			model := model_helpers.NewValidTask("foo")
			payload := []byte{byte(format.PROTO), byte(format.V0), 'f', 'o', 'o'}
			err := format.UnmarshalEnvelope(logger, payload, model.ToProto())
			Expect(err).To(HaveOccurred())
		})
	})
})
