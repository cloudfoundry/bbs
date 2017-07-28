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
			processGuid, err = sqlDB.CreateLRPDeployment(logger, lrpCreate)
			Expect(err).ToNot(HaveOccurred())

			lrpUpdate = model_helpers.NewValidLRPDeploymentUpdate("update")
			processGuid, err = sqlDB.UpdateLRPDeployment(logger, processGuid, lrpUpdate)
			Expect(err).ToNot(HaveOccurred())
		})

		It("fetches the LRP deployment associated with a LRPDefinition's definition guid", func() {
			lrpDeployment, err := sqlDB.LRPDeploymentByDefinitionGuid(logger, lrpCreate.DefinitionId)
			Expect(err).ToNot(HaveOccurred())

			Expect(lrpDeployment.Definitions).To(HaveLen(2))
			Expect(lrpDeployment.Definitions[lrpCreate.DefinitionId]).To(Equal(lrpCreate.Definition))
			Expect(lrpDeployment.Definitions[lrpUpdate.DefinitionId]).To(Equal(lrpUpdate.Definition))
			Expect(lrpDeployment.ProcessGuid).To(Equal(processGuid))
		})
	})
})
