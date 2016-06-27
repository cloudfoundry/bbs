package models_test

import (
	"code.cloudfoundry.org/bbs/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CellPresence", func() {
	var (
		cellPresence models.CellPresence
		capacity     models.CellCapacity
	)

	BeforeEach(func() {
		capacity = models.NewCellCapacity(128, 1024, 3)
		rootfsProviders := []string{"provider-1"}
		preloadedRootFSes := []string{"provider-2"}
		cellPresence = models.NewCellPresence("some-id", "some-address", "some-zone", capacity, rootfsProviders, preloadedRootFSes)
	})

	Describe("Validate", func() {
		Context("when cell presence is valid", func() {
			It("does not return an error", func() {
				Expect(cellPresence.Validate()).NotTo(HaveOccurred())
			})
		})

		Context("when cell presence is invalid", func() {
			Context("when cell id is invalid", func() {
				BeforeEach(func() {
					cellPresence.CellId = ""
				})

				It("returns an error", func() {
					err := cellPresence.Validate()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("cell_id"))
				})
			})

			Context("when rep address is invalid", func() {
				BeforeEach(func() {
					cellPresence.RepAddress = ""
				})

				It("returns an error", func() {
					err := cellPresence.Validate()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("rep_address"))
				})
			})

			Context("when cell capacity is invalid", func() {
				Context("when memory is zero", func() {
					BeforeEach(func() {
						cellPresence.Capacity.MemoryMb = 0
					})
					It("returns an error", func() {
						err := cellPresence.Validate()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("memory_mb"))
					})
				})

				Context("when memory is negative", func() {
					BeforeEach(func() {
						cellPresence.Capacity.MemoryMb = -1
					})
					It("returns an error", func() {
						err := cellPresence.Validate()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("memory_mb"))
					})
				})

				Context("when containers are zero", func() {
					BeforeEach(func() {
						cellPresence.Capacity.Containers = 0
					})
					It("returns an error", func() {
						err := cellPresence.Validate()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("containers"))
					})
				})

				Context("when containers are negative", func() {
					BeforeEach(func() {
						cellPresence.Capacity.Containers = -1
					})
					It("returns an error", func() {
						err := cellPresence.Validate()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("containers"))
					})
				})

				Context("when disk is negative", func() {
					BeforeEach(func() {
						cellPresence.Capacity.DiskMb = -1
					})
					It("returns an error", func() {
						err := cellPresence.Validate()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("disk_mb"))
					})
				})
			})
		})
	})
})
