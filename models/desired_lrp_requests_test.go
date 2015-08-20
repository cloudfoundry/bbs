package models_test

import (
	"github.com/cloudfoundry-incubator/bbs/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DesiredLRP Requests", func() {
	Describe("DesiredLRPsByProcessGuidRequest", func() {
		Describe("Validate", func() {
			var request models.DesiredLRPByProcessGuidRequest

			BeforeEach(func() {
				request = models.DesiredLRPByProcessGuidRequest{
					ProcessGuid: "something",
				}
			})

			Context("when valid", func() {
				It("returns nil", func() {
					Expect(request.Validate()).To(BeNil())
				})
			})

			Context("when the ProcessGuid is blank", func() {
				BeforeEach(func() {
					request.ProcessGuid = ""
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"process_guid"}))
				})
			})
		})
	})
})
