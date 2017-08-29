package sqldb_test

import (
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/models/test/model_helpers"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LRPDeployment", func() {
	var (
		logger lager.Logger
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("lrp-deployment")
	})

	Describe("LRPDeploymentByDefinitionGuid", func() {
		var (
			processGuid string
			lrpCreate   *models.LRPDeploymentCreation
			lrpUpdate   *models.LRPDeploymentUpdate
		)
		BeforeEach(func() {
			var err error
			lrpCreate = model_helpers.NewValidLRPDeploymentCreation("process-guid", "definition_guid")
			lerpDerp, err := sqlDB.CreateLRPDeployment(logger, lrpCreate)
			Expect(err).ToNot(HaveOccurred())
			processGuid = lerpDerp.ProcessGuid

			lrpUpdate = model_helpers.NewValidLRPDeploymentUpdate("update")
			_, err = sqlDB.UpdateLRPDeployment(logger, lerpDerp.ProcessGuid, lrpUpdate)
			Expect(err).ToNot(HaveOccurred())
		})

		It("fetches the LRP deployment associated with a LRPDefinition's definition guid", func() {
			lrpDeployment, err := sqlDB.LRPDeploymentByProcessGuid(logger, lrpCreate.ProcessGuid)
			Expect(err).ToNot(HaveOccurred())

			Expect(lrpDeployment.Definitions).To(HaveLen(2))
			Expect(lrpDeployment.Definitions[lrpCreate.DefinitionId]).To(Equal(lrpCreate.Definition))
			Expect(lrpDeployment.Definitions[*lrpUpdate.DefinitionId]).To(Equal(lrpUpdate.Definition))
			Expect(processGuid).To(Equal(lrpDeployment.ProcessGuid))
		})

		Context("when the deployment has been deleted", func() {
			It("does not leave any record of the deployment behind", func() {
				_, err := sqlDB.DeleteLRPDeployment(logger, lrpCreate.ProcessGuid)
				Expect(err).ToNot(HaveOccurred())

				lrpDeployment, err := sqlDB.LRPDeploymentByProcessGuid(logger, lrpCreate.ProcessGuid)
				Expect(err).To(Equal(models.ErrResourceNotFound))
				Expect(lrpDeployment).To(BeNil())
			})
		})
	})
})
