package models_test

import (
	"code.cloudfoundry.org/bbs/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Events", func() {
	Describe("NewActualLRPInstanceChangeEvent", func() {
		var (
			before, after *models.ActualLRP
		)
		Context("when before is set", func() {
			BeforeEach(func() {
				before = &models.ActualLRP{
					ActualLrpKey: models.ActualLRPKey{
						ProcessGuid: "before-process-guid",
						Index:       5,
						Domain:      "before-domain",
					},
					ActualLrpInstanceKey: models.ActualLRPInstanceKey{
						InstanceGuid: "before-instance-guid",
						CellId:       "before-cell-id",
					},
				}
			})

			Context("when after is set", func() {
				BeforeEach(func() {
					after = &models.ActualLRP{
						ActualLrpKey: models.ActualLRPKey{
							ProcessGuid: "after-process-guid",
							Index:       7,
							Domain:      "after-domain",
						},
						ActualLrpInstanceKey: models.ActualLRPInstanceKey{
							InstanceGuid: "after-instance-guid",
							CellId:       "after-cell-id",
						},
					}
				})

				It("returns the after's actual lrp key and instance key", func() {
					event := models.NewActualLRPInstanceChangedEvent(before, after, "some-trace-id")
					Expect(event.ActualLrpKey).To(Equal(after.ActualLrpKey))
					Expect(event.ActualLrpInstanceKey).To(Equal(after.ActualLrpInstanceKey))
				})
			})

			Context("when after is unclaimed", func() {
				BeforeEach(func() {
					after = &models.ActualLRP{
						ActualLrpKey: models.ActualLRPKey{
							ProcessGuid: "after-process-guid",
							Index:       7,
							Domain:      "after-domain",
						},
						State: models.ActualLRPStateUnclaimed,
					}
				})

				It("returns after's actual lrp key and before's instance key", func() {
					event := models.NewActualLRPInstanceChangedEvent(before, after, "some-trace-id")
					Expect(event.ActualLrpKey).To(Equal(after.ActualLrpKey))
					Expect(event.ActualLrpInstanceKey).To(Equal(before.ActualLrpInstanceKey))
				})
			})

			Context("when after is not set", func() {
				BeforeEach(func() {
					after = &models.ActualLRP{}
				})

				It("returns the before's actual lrp key and instance key", func() {
					event := models.NewActualLRPInstanceChangedEvent(before, after, "some-trace-id")
					Expect(event.ActualLrpKey).To(Equal(before.ActualLrpKey))
					Expect(event.ActualLrpInstanceKey).To(Equal(before.ActualLrpInstanceKey))
				})
			})
		})

		Context("when before is not set", func() {
			BeforeEach(func() {
				before = &models.ActualLRP{}
			})

			Context("when after is set", func() {
				BeforeEach(func() {
					after = &models.ActualLRP{
						ActualLrpKey: models.ActualLRPKey{
							ProcessGuid: "after-process-guid",
							Index:       7,
							Domain:      "after-domain",
						},
						ActualLrpInstanceKey: models.ActualLRPInstanceKey{
							InstanceGuid: "after-instance-guid",
							CellId:       "after-cell-id",
						},
					}
				})

				It("returns the after's actual lrp key and instance key", func() {
					event := models.NewActualLRPInstanceChangedEvent(before, after, "some-trace-id")
					Expect(event.ActualLrpKey).To(Equal(after.ActualLrpKey))
					Expect(event.ActualLrpInstanceKey).To(Equal(after.ActualLrpInstanceKey))
				})

				Context("when before is unclaimed", func() {
					BeforeEach(func() {
						before = &models.ActualLRP{
							ActualLrpKey: models.ActualLRPKey{
								ProcessGuid: "before-process-guid",
								Index:       5,
								Domain:      "before-domain",
							},
							State: models.ActualLRPStateUnclaimed,
						}
					})

					It("returns after's actual lrp key and after's instance key", func() {
						event := models.NewActualLRPInstanceChangedEvent(before, after, "some-trace-id")
						Expect(event.ActualLrpKey).To(Equal(after.ActualLrpKey))
						Expect(event.ActualLrpInstanceKey).To(Equal(after.ActualLrpInstanceKey))
					})
				})
			})

			Context("when after is not set", func() {
				BeforeEach(func() {
					after = &models.ActualLRP{}
				})

				It("returns the empty actual lrp key and instance key", func() {
					event := models.NewActualLRPInstanceChangedEvent(before, after, "some-trace-id")
					Expect(event.ActualLrpKey).To(Equal(models.ActualLRPKey{}))
					Expect(event.ActualLrpInstanceKey).To(Equal(models.ActualLRPInstanceKey{}))
				})
			})
		})
	})
})
