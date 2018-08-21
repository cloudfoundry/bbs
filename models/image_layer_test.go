package models_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/bbs/models"
)

var _ = Describe("ImageLayer", func() {
	Describe("Validate", func() {
		var layer *models.ImageLayer

		Context("when 'url', 'destination_path', 'media_type' are specified", func() {
			It("is valid", func() {
				layer = &models.ImageLayer{
					Url:             "web_location",
					DestinationPath: "local_location",
					MediaType:       "media_type",
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
						DigestAlgorithm: "md5",
						MediaType:       "media_type",
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
				"checksum value",
				&models.ImageLayer{
					Url:             "web_location",
					DestinationPath: "local_location",
					DigestAlgorithm: "md5",
					MediaType:       "some-type",
				},
			},
			{
				"checksum algorithm",
				&models.ImageLayer{
					Url:             "web_location",
					DestinationPath: "local_location",
					DigestValue:     "some digest",
					MediaType:       "some-type",
				},
			},
			{
				"checksum value",
				&models.ImageLayer{
					Url:             "web_location",
					DestinationPath: "local_location",
					MediaType:       "some-type",
					LayerType:       models.ImageLayer_Exclusive,
				},
			},
			{
				"checksum algorithm",
				&models.ImageLayer{
					Url:             "web_location",
					DestinationPath: "local_location",
					MediaType:       "some-type",
					LayerType:       models.ImageLayer_Exclusive,
				},
			},
			{
				"invalid algorithm",
				&models.ImageLayer{
					Url:             "web_location",
					DestinationPath: "local_location",
					DigestAlgorithm: "invalid",
					DigestValue:     "some digest",
					MediaType:       "some-type",
				},
			},
			{
				"media_type",
				&models.ImageLayer{
					Url:             "web_location",
					DestinationPath: "local_location",
					DigestAlgorithm: "md5",
					DigestValue:     "some digest",
				},
			},
		} {
			testValidatorErrorCase(testCase)
		}
	})
})
