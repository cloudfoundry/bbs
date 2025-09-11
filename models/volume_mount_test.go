package models_test

import (
	"code.cloudfoundry.org/bbs/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("VolumeMount", func() {
	Context("Validate", func() {
		var (
			mount models.VolumeMount
			err   error
		)

		BeforeEach(func() {
			mount = models.VolumeMount{
				Driver:       "my-driver",
				ContainerDir: "/mnt/mypath",
				Mode:         "r",
				Shared: &models.SharedDevice{
					VolumeId:    "my-volume",
					MountConfig: `{"foo":"bar"}`,
				},
			}
		})

		JustBeforeEach(func() {
			err = mount.Validate()
		})

		It("doesnt error with a good mount", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		Context("given an invalid driver", func() {
			BeforeEach(func() {
				mount.Driver = ""
			})

			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		Context("given an invalid volumeId", func() {
			BeforeEach(func() {
				mount.Shared.VolumeId = ""
			})

			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		Context("given an unset mode", func() {
			BeforeEach(func() {
				mount.Mode = ""
			})

			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		Context("given a DedicatedDevice", func() {
			BeforeEach(func() {
				mount.Dedicated = &models.DedicatedDevice{
					MounterId:    "my-mounter",
					MountConfig:  `{"foo":"bar"}`,
					DeviceConfig: `{"baz":"qux"}`,
				}
			})

			It("doesnt error with a good mount", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			Context("given an invalid mounterId", func() {
				BeforeEach(func() {
					mount.Dedicated.MounterId = ""
				})

				It("should return an error", func() {
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})
})
