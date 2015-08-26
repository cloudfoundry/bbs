package codec_test

import (
	"github.com/cloudfoundry-incubator/bbs/db/codec"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Codecs", func() {
	var encoderKind codec.Kind
	var codecs *codec.Codecs

	BeforeEach(func() {
		encoderKind = codec.NONE
	})

	JustBeforeEach(func() {
		codecs = codec.NewCodecs(encoderKind)
	})

	Describe("Decode", func() {
		var payload []byte

		Context("when the payload is missing the kind header", func() {
			BeforeEach(func() {
				payload = []byte(`{"foo": "bar"}`)
			})

			It("returns the same data from decode", func() {
				Expect(codecs.Decode(payload)).To(Equal(payload))
			})
		})

		Context("when the payload is an UNENCODED kind", func() {
			BeforeEach(func() {
				payload = []byte(`00{"foo": "bar"}`)
			})

			It("returns the data without the header", func() {
				Expect(codecs.Decode(payload)).To(Equal(payload[2:]))
			})
		})

		Context("when the payload is a BASE64 kind", func() {
			BeforeEach(func() {
				payload = []byte(`01eyJlbmNvZGVkIjoianNvbiJ9`)
			})

			It("returns the data without the header", func() {
				Expect(codecs.Decode(payload)).To(Equal([]byte(`{"encoded":"json"}`)))
			})
		})

		Context("when the payload is an unknown kind", func() {
			BeforeEach(func() {
				payload = []byte(`99{"foo": "bar"}`)
			})

			It("returns an error", func() {
				_, err := codecs.Decode(payload)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Encode", func() {
		var data []byte

		BeforeEach(func() {
			data = []byte(`{"bob":"uncle"}`)
		})

		Context("when no encoder has been set", func() {
			It("doesn't change the data", func() {
				payload, err := codecs.Encode(data)
				Expect(err).NotTo(HaveOccurred())
				Expect(payload).Should(Equal(data))
			})
		})

		Context("when UNENCODED is the encoding kind", func() {
			BeforeEach(func() {
				encoderKind = codec.UNENCODED
			})

			It("adds a header without encoding the data", func() {
				payload, err := codecs.Encode(data)
				Expect(err).NotTo(HaveOccurred())
				Expect(payload).Should(Equal([]byte(`00{"bob":"uncle"}`)))
			})
		})

		Context("when a base 64 encoder has been set", func() {
			BeforeEach(func() {
				encoderKind = codec.BASE64
			})

			It("generates a header followed by the encoded data", func() {
				payload, err := codecs.Encode(data)
				Expect(err).NotTo(HaveOccurred())
				Expect(payload).Should(Equal([]byte(`01eyJib2IiOiJ1bmNsZSJ9`)))
			})
		})
	})

	Describe("SupportsBinary", func() {
		It("indicates whether the encoding supports binary", func() {
			Expect(codec.NONE.SupportsBinary()).To(BeFalse())
			Expect(codec.UNENCODED.SupportsBinary()).To(BeFalse())
			Expect(codec.BASE64.SupportsBinary()).To(BeTrue())
		})
	})
})
