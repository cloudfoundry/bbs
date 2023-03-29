package models_test

import (
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/models/test/model_helpers"
	. "github.com/onsi/ginkgo/v2"
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

	Describe("DesireLRPRequest", func() {
		Describe("Validate", func() {
			var request models.DesireLRPRequest

			BeforeEach(func() {
				request = models.DesireLRPRequest{
					DesiredLrp: model_helpers.NewValidDesiredLRP("some-guid"),
				}
			})

			Context("when valid", func() {
				It("returns nil", func() {
					Expect(request.Validate()).To(BeNil())
				})
			})

			Context("when the DesiredLRP is blank", func() {
				BeforeEach(func() {
					request.DesiredLrp = nil
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"desired_lrp"}))
				})
			})

			Context("when the DesiredLRP is invalid", func() {
				BeforeEach(func() {
					request.DesiredLrp.ProcessGuid = ""
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"process_guid"}))
				})
			})
		})
	})

	Describe("UpdateDesiredLRPRequest", func() {
		Describe("Validate", func() {
			var request models.UpdateDesiredLRPRequest

			BeforeEach(func() {
				request = models.UpdateDesiredLRPRequest{
					ProcessGuid: "some-guid",
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

			Context("when the Update is invalid", func() {
				BeforeEach(func() {
					request.Update = &models.DesiredLRPUpdate{}
					request.Update.SetInstances(-1)
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"instances"}))
				})
			})
		})
	})

	Describe("RemoveDesiredLRPRequest", func() {
		Describe("Validate", func() {
			var request models.RemoveDesiredLRPRequest

			BeforeEach(func() {
				request = models.RemoveDesiredLRPRequest{
					ProcessGuid: "some-guid",
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
