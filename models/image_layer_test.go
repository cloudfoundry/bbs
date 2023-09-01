package models_test

import (
	"encoding/json"

	"code.cloudfoundry.org/bbs/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ImageLayer", func() {
	Describe("Validate", func() {
		var layer *models.ImageLayer

		Context("when 'url', 'destination_path', 'media_type' are specified", func() {
			It("is valid", func() {
				layer = &models.ImageLayer{
					Url:             "web_location",
					DestinationPath: "local_location",
					MediaType:       models.ImageLayer_TGZ,
					LayerType:       models.ImageLayer_SHARED,
				}

				err := layer.Validate()
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when the action also has valid 'digest_value' and 'digest_algorithm'", func() {
				It("is valid", func() {
					layer = &models.ImageLayer{
						Url:             "web_location",
						DestinationPath: "local_location",
						DigestValue:     "some digest",
						DigestAlgorithm: models.ImageLayer_SHA256,
						MediaType:       models.ImageLayer_TGZ,
						LayerType:       models.ImageLayer_EXCLUSIVE,
					}

					err := layer.Validate()
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		for _, testCase := range []ValidatorErrorCase{
			{
				"url",
				&models.ImageLayer{
					DestinationPath: "local_location",
				},
			},
			{
				"destination_path",
				&models.ImageLayer{
					Url: "web_location",
				},
			},
			{
				"layer_type",
				&models.ImageLayer{},
			},
			{
				"layer_type",
				&models.ImageLayer{
					LayerType: models.ImageLayer_Type(10),
				},
			},
			{
				"digest_value",
				&models.ImageLayer{
					Url:             "web_location",
					DestinationPath: "local_location",
					DigestAlgorithm: models.ImageLayer_SHA256,
					MediaType:       models.ImageLayer_TGZ,
				},
			},
			{
				"digest_algorithm",
				&models.ImageLayer{
					Url:             "web_location",
					DestinationPath: "local_location",
					DigestValue:     "some digest",
					MediaType:       models.ImageLayer_TGZ,
				},
			},
			{
				"digest_value",
				&models.ImageLayer{
					Url:             "web_location",
					DestinationPath: "local_location",
					MediaType:       models.ImageLayer_TGZ,
					LayerType:       models.ImageLayer_EXCLUSIVE,
				},
			},
			{
				"digest_algorithm",
				&models.ImageLayer{
					Url:             "web_location",
					DestinationPath: "local_location",
					MediaType:       models.ImageLayer_TGZ,
					LayerType:       models.ImageLayer_EXCLUSIVE,
				},
			},
			{
				"digest_algorithm",
				&models.ImageLayer{
					Url:             "web_location",
					DestinationPath: "local_location",
					DigestAlgorithm: models.ImageLayer_DigestAlgorithm(5),
					DigestValue:     "some digest",
					MediaType:       models.ImageLayer_TGZ,
				},
			},
			{
				"media_type",
				&models.ImageLayer{
					Url:             "web_location",
					DestinationPath: "local_location",
					DigestAlgorithm: models.ImageLayer_SHA256,
					DigestValue:     "some digest",
				},
			},
			{
				"media_type",
				&models.ImageLayer{
					Url:             "web_location",
					DestinationPath: "local_location",
					DigestAlgorithm: models.ImageLayer_SHA256,
					DigestValue:     "some digest",
					MediaType:       models.ImageLayer_MediaType(9),
				},
			},
		} {
			testValidatorErrorCase(testCase)
		}
	})

	Describe("DigestAlgorithm", func() {
		Describe("serialization", func() {
			DescribeTable("marshals and unmarshals between the value and the expected JSON output",
				func(v models.ImageLayer_DigestAlgorithm, expectedJSON string) {
					Expect(json.Marshal(v)).To(MatchJSON(expectedJSON))
					var testV models.ImageLayer_DigestAlgorithm
					Expect(json.Unmarshal([]byte(expectedJSON), &testV)).To(Succeed())
					Expect(testV).To(Equal(v))
				},
				Entry("invalid", models.ImageLayer_DigestAlgorithmInvalid, `"ImageLayer_DigestAlgorithmInvalid"`),
				Entry("sha256", models.ImageLayer_SHA256, `"SHA256"`),
				Entry("sha512", models.ImageLayer_SHA512, `"SHA512"`),
			)
		})
	})

	Describe("MediaType", func() {
		Describe("serialization", func() {
			DescribeTable("marshals and unmarshals between the value and the expected JSON output",
				func(v models.ImageLayer_MediaType, expectedJSON string) {
					Expect(json.Marshal(v)).To(MatchJSON(expectedJSON))
					var testV models.ImageLayer_MediaType
					Expect(json.Unmarshal([]byte(expectedJSON), &testV)).To(Succeed())
					Expect(testV).To(Equal(v))
				},
				Entry("invalid", models.ImageLayer_MediaTypeInvalid, `"ImageLayer_MediaTypeInvalid"`),
				Entry("tgz", models.ImageLayer_TGZ, `"TGZ"`),
				Entry("tar", models.ImageLayer_TAR, `"TAR"`),
				Entry("zip", models.ImageLayer_ZIP, `"ZIP"`),
			)
		})
	})

	Describe("Type", func() {
		Describe("serialization", func() {
			DescribeTable("marshals and unmarshals between the value and the expected JSON output",
				func(v models.ImageLayer_Type, expectedJSON string) {
					Expect(json.Marshal(v)).To(MatchJSON(expectedJSON))
					var testV models.ImageLayer_Type
					Expect(json.Unmarshal([]byte(expectedJSON), &testV)).To(Succeed())
					Expect(testV).To(Equal(v))
				},
				Entry("invalid", models.ImageLayer_LayerTypeInvalid, `"ImageLayer_LayerTypeInvalid"`),
				Entry("shared", models.ImageLayer_SHARED, `"SHARED"`),
				Entry("exclusive", models.ImageLayer_EXCLUSIVE, `"EXCLUSIVE"`),
			)
		})
	})
})
