package models_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/bbs/models"
)

var _ = Describe("CachedDependency", func() {
	Describe("Validate", func() {
		var cachedDep *models.CachedDependency

		Context("when the action has 'from', 'to', and 'user' specified", func() {
			It("is valid", func() {
				cachedDep = &models.CachedDependency{
					From: "web_location",
					To:   "local_location",
				}

				err := cachedDep.Validate()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		for _, testCase := range []ValidatorErrorCase{
			{
				"from",
				&models.CachedDependency{
					To: "local_location",
				},
			},
			{
				"to",
				&models.CachedDependency{
					From: "web_location",
				},
			},
		} {
			testValidatorErrorCase(testCase)
		}
	})
})
