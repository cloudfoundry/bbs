package models_test

import (
	"errors"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/models/fakes"
	"github.com/cloudfoundry-incubator/bbs/models/test/model_helpers"
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("Envelope", func() {
	var logger *lagertest.TestLogger

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")
	})

	Describe("Open", func() {
		It("prserves nil data and tags it V0 JSON", func() {
			envelope := models.OpenEnvelope(nil)
			Expect(envelope.Version).To(Equal(models.V0))
			Expect(envelope.Payload).To(BeNil())
			Expect(envelope.SerializationFormat).To(BeEquivalentTo(models.JSON))
		})

		It("preserves empty data and tags it V0 JSON", func() {
			envelope := models.OpenEnvelope([]byte{})
			Expect(envelope.Version).To(Equal(models.V0))
			Expect(envelope.Payload).To(Equal([]byte{}))
			Expect(envelope.SerializationFormat).To(BeEquivalentTo(models.JSON))
		})

		It("preserves unencoded data and tags it as V0 JSON", func() {
			envelope := models.OpenEnvelope([]byte("{}"))
			Expect(envelope.SerializationFormat).To(BeEquivalentTo(models.JSON))
			Expect(envelope.Version).To(Equal(models.V0))
			Expect(envelope.Payload).To(Equal([]byte("{}")))
		})

		It("handles JSON encoded data and tags it as JSON with the correct version and payload", func() {
			envelope := models.OpenEnvelope(bytesForEnvelope(models.JSON, models.V0, "{}"))
			Expect(envelope.SerializationFormat).To(Equal(models.JSON))
			Expect(envelope.Version).To(Equal(models.V0))
			Expect(envelope.Payload).To(Equal([]byte("{}")))
		})

		It("handles protobuf encoded data and tags it as as PROTO with the correct version and payload", func() {
			task := &models.Task{}
			protoData, err := task.Marshal()
			Expect(err).NotTo(HaveOccurred())

			envelope := models.OpenEnvelope(bytesForEnvelope(models.PROTO, models.V0, string(protoData)))
			Expect(envelope.SerializationFormat).To(Equal(models.PROTO))
			Expect(envelope.Version).To(Equal(models.V0))
			Expect(envelope.Payload).To(Equal(protoData))
		})
	})

	Describe("Marshal", func() {
		It("can successfully marshal a model object envelope", func() {
			task := model_helpers.NewValidTask("some-guid")
			encoded, err := models.MarshalEnvelope(models.PROTO, task)
			Expect(err).NotTo(HaveOccurred())

			Expect(models.SerializationFormat(encoded[0])).To(Equal(models.PROTO))
			Expect(models.Version(encoded[1])).To(Equal(models.V0))

			var newTask models.Task
			modelErr := proto.Unmarshal(encoded[2:], &newTask)
			Expect(modelErr).To(BeNil())

			Expect(*task).To(Equal(newTask))
		})

		Context("when model validation fails", func() {
			It("returns an error ", func() {
				model := &fakes.FakeVersioner{}
				model.ValidateReturns(errors.New("go away"))

				_, err := models.MarshalEnvelope(models.PROTO, model)
				Expect(err).To(Equal(models.NewError(models.InvalidRecord, "go away")))
			})
		})
	})

	Describe("Unmarshal", func() {
		It("can marshal and unmarshal a task without losing data", func() {
			task := model_helpers.NewValidTask("some-guid")
			ValueInEtcd, err := models.MarshalEnvelope(models.PROTO, task)
			Expect(err).NotTo(HaveOccurred())

			envelope := models.OpenEnvelope(ValueInEtcd)

			var resultingTask models.Task
			err = envelope.Unmarshal(logger, &resultingTask)
			Expect(err).NotTo(HaveOccurred())

			Expect(resultingTask).To(BeEquivalentTo(*task))
		})

		It("calls MigrateFromVersion on on the model object with the envelope version", func() {
			envelope := &models.Envelope{
				SerializationFormat: models.JSON,
				Version:             models.Version(99),
				Payload:             []byte(`{}`),
			}
			model := &fakes.FakeVersioner{}

			err := envelope.Unmarshal(logger, model)
			Expect(err).NotTo(HaveOccurred())

			Expect(model.MigrateFromVersionCallCount()).To(Equal(1))
			Expect(model.MigrateFromVersionArgsForCall(0)).To(Equal(models.Version(99)))
		})

		It("returns an error when the serialization format is unknown", func() {
			envelope := &models.Envelope{
				SerializationFormat: models.SerializationFormat(99),
			}
			model := &fakes.FakeVersioner{}

			err := envelope.Unmarshal(logger, model)
			Expect(err).To(HaveOccurred())
			Expect(err.Type).To(Equal(models.FailedToOpenEnvelope))
		})

		It("returns an error when the json payload is invalid", func() {
			model := &fakes.FakeVersioner{}
			envelope := &models.Envelope{
				SerializationFormat: models.JSON,
				Payload:             []byte(`foobar: baz`),
			}

			err := envelope.Unmarshal(logger, model)
			Expect(err.Type).To(Equal(models.InvalidRecord))
		})

		It("returns an error when the protobuf payload is invalid", func() {
			model := &models.Task{}
			envelope := &models.Envelope{
				SerializationFormat: models.PROTO,
				Payload:             []byte(`foobar: baz`),
			}

			err := envelope.Unmarshal(logger, model)
			Expect(err.Type).To(Equal(models.InvalidRecord))
		})
	})
})

func bytesForEnvelope(f models.SerializationFormat, v models.Version, payloads ...string) []byte {
	env := []byte{byte(f), byte(v)}
	for i := range payloads {
		env = append(env, []byte(payloads[i])...)
	}
	return env
}
