package models_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/bbs/models"
)

var _ = Describe("ImageLayer", func() {
	Describe("Validate", func() {
		var layer *models.ImageLayer

		Context("when 'url', 'destination_path', 'content_type' are specified", func() {
			It("is valid", func() {
				layer = &models.ImageLayer{
					Url:             "web_location",
					DestinationPath: "local_location",
					ContentType:     "content_type",
				}

				err := layer.Validate()
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when the action also has valid 'checksum_value' and 'checksum_algorithm'", func() {
				It("is valid", func() {
					layer = &models.ImageLayer{
						Url:               "web_location",
						DestinationPath:   "local_location",
						ChecksumValue:     "some checksum",
						ChecksumAlgorithm: "md5",
						ContentType:       "content_type",
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
					Url:               "web_location",
					DestinationPath:   "local_location",
					ChecksumAlgorithm: "md5",
					ContentType:       "some-type",
				},
			},
			{
				"checksum algorithm",
				&models.ImageLayer{
					Url:             "web_location",
					DestinationPath: "local_location",
					ChecksumValue:   "some checksum",
					ContentType:     "some-type",
				},
			},
			{
				"checksum value",
				&models.ImageLayer{
					Url:             "web_location",
					DestinationPath: "local_location",
					ContentType:     "some-type",
					LayerType:       models.ImageLayer_Exclusive,
				},
			},
			{
				"checksum algorithm",
				&models.ImageLayer{
					Url:             "web_location",
					DestinationPath: "local_location",
					ContentType:     "some-type",
					LayerType:       models.ImageLayer_Exclusive,
				},
			},
			{
				"invalid algorithm",
				&models.ImageLayer{
					Url:               "web_location",
					DestinationPath:   "local_location",
					ChecksumAlgorithm: "invalid",
					ChecksumValue:     "some checksum",
					ContentType:       "some-type",
				},
			},
			{
				"content_type",
				&models.ImageLayer{
					Url:               "web_location",
					DestinationPath:   "local_location",
					ChecksumAlgorithm: "md5",
					ChecksumValue:     "some checksum",
				},
			},
		} {
			testValidatorErrorCase(testCase)
		}
	})
})
