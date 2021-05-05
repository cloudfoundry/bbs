package rep_test

import (
	"fmt"

	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/clients/rep"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Resources", func() {
	var (
		cellState      rep.CellState
		linuxRootFSURL string
	)

	BeforeEach(func() {
		linuxOnlyRootFSProviders := rep.RootFSProviders{models.PreloadedRootFSScheme: rep.NewFixedSetRootFSProvider("linux")}
		total := rep.NewResources(1000, 2000, 10)
		avail := rep.NewResources(950, 1900, 3)
		linuxRootFSURL = models.PreloadedRootFS("linux")

		lrps := []rep.LRP{
			*buildLRP("ig-1", "pg-1", "domain", 0, linuxRootFSURL, 10, 20, 30, []string{}, []string{}, models.ActualLRPStateClaimed),
			*buildLRP("ig-2", "pg-1", "domain", 1, linuxRootFSURL, 10, 20, 30, []string{}, []string{}, models.ActualLRPStateClaimed),
			*buildLRP("ig-3", "pg-2", "domain", 0, linuxRootFSURL, 10, 20, 30, []string{}, []string{}, models.ActualLRPStateClaimed),
			*buildLRP("ig-4", "pg-3", "domain", 0, linuxRootFSURL, 10, 20, 30, []string{}, []string{}, models.ActualLRPStateClaimed),
			*buildLRP("ig-5", "pg-4", "domain", 0, linuxRootFSURL, 10, 20, 30, []string{}, []string{}, models.ActualLRPStateClaimed),
		}

		tasks := []rep.Task{
			*buildTask("tg-big", "domain", linuxRootFSURL, 20, 10, 10, []string{}, []string{}, models.Task_Running, false),
			*buildTask("tg-small", "domain", linuxRootFSURL, 10, 10, 10, []string{}, []string{}, models.Task_Running, false),
		}

		cellState = rep.NewCellState(
			"cell-id",
			0,
			"https://foo.cell.service.cf.internal",
			linuxOnlyRootFSProviders,
			avail,
			total,
			lrps,
			tasks,
			"my-zone",
			7,
			false,
			nil,
			nil,
			nil,
			0,
		)
	})

	Describe("MatchPlacementTags", func() {
		Context("when cell state does not have placement tags", func() {
			It("does not allow lrps with placement tags", func() {
				state := rep.CellState{
					PlacementTags:         []string{},
					OptionalPlacementTags: []string{},
				}
				Expect(state.MatchPlacementTags([]string{"foo"})).To(BeFalse())
				Expect(state.MatchPlacementTags([]string{})).To(BeTrue())
			})
		})

		Context("when it has require placement tags", func() {
			It("requires the placement tags to be present in the lrp", func() {
				state := rep.CellState{
					PlacementTags:         []string{"foo", "bar"},
					OptionalPlacementTags: []string{},
				}
				Expect(state.MatchPlacementTags([]string{})).To(BeFalse())
				Expect(state.MatchPlacementTags([]string{"foo"})).To(BeFalse())
				Expect(state.MatchPlacementTags([]string{"foo", "bar"})).To(BeTrue())
			})
		})

		Context("when it has optional placement tags", func() {
			It("does not require placement tags to be present on the desired lrp", func() {
				state := rep.CellState{
					PlacementTags:         []string{},
					OptionalPlacementTags: []string{"foo"},
				}
				Expect(state.MatchPlacementTags([]string{})).To(BeTrue())
				Expect(state.MatchPlacementTags([]string{"foo"})).To(BeTrue())
			})

			It("does not allow extra placement tags to be defined in the lrp", func() {
				state := rep.CellState{
					PlacementTags:         []string{},
					OptionalPlacementTags: []string{"foo"},
				}
				Expect(state.MatchPlacementTags([]string{"bar"})).To(BeFalse())
			})
		})

		Context("when both placement tags and optional placement tags are present", func() {
			It("requires all required placement tags to be on the lrp", func() {
				state := rep.CellState{
					PlacementTags:         []string{"foo"},
					OptionalPlacementTags: []string{"bar"},
				}
				Expect(state.MatchPlacementTags([]string{})).To(BeFalse())
				Expect(state.MatchPlacementTags([]string{"bar"})).To(BeFalse())
				Expect(state.MatchPlacementTags([]string{"foo"})).To(BeTrue())
				Expect(state.MatchPlacementTags([]string{"foo", "bar"})).To(BeTrue())
				Expect(state.MatchPlacementTags([]string{"foo", "bar", "baz"})).To(BeFalse())
			})
		})
	})

	Describe("Resource Matching", func() {
		var requiredResource rep.Resource
		var err error
		BeforeEach(func() {
			requiredResource = rep.NewResource(10, 10, 10)
		})

		JustBeforeEach(func() {
			err = cellState.ResourceMatch(&requiredResource)
		})

		Context("when insufficient memory", func() {
			BeforeEach(func() {
				requiredResource.MemoryMB = 5000
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("insufficient resources: memory"))
			})
		})

		Context("when insufficient disk", func() {
			BeforeEach(func() {
				requiredResource.DiskMB = 5000
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("insufficient resources: disk"))
			})
		})

		Context("when insufficient disk and memory", func() {
			BeforeEach(func() {
				requiredResource.MemoryMB = 5000
				requiredResource.DiskMB = 5000
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("insufficient resources: disk, memory"))
			})
		})

		Context("when insufficient disk, memory and containers", func() {
			BeforeEach(func() {
				requiredResource.MemoryMB = 5000
				requiredResource.DiskMB = 5000
				cellState.AvailableResources.Containers = 0
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("insufficient resources: containers, disk, memory"))
			})
		})

		Context("when there are no available containers", func() {
			BeforeEach(func() {
				cellState.AvailableResources.Containers = 0
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("insufficient resources: containers"))
			})
		})

		Context("when there is sufficient room", func() {
			It("does not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("StackPathMap", func() {
		Describe("PathForRootFS", func() {
			var stackPathMap rep.StackPathMap
			BeforeEach(func() {
				stackPathMap = rep.StackPathMap{
					"cflinuxfs3": "cflinuxfs3:/var/vcap/packages/cflinuxfs3/rootfs.tar",
				}
			})
			It("returns the resolved path if the RootFS URL scheme is preloaded", func() {
				p, err := stackPathMap.PathForRootFS("preloaded:cflinuxfs3")
				Expect(err).NotTo(HaveOccurred())
				Expect(p).To(Equal("cflinuxfs3:/var/vcap/packages/cflinuxfs3/rootfs.tar"))
			})
			It("returns the correct URL if the RootFS URL scheme is preloaded+layer", func() {
				queryString := "?layer=https://blobstore.internal/layer1.tgz?layer_path=/tmp/asset1&layer_digest=alkjsdflkj"
				p, err := stackPathMap.PathForRootFS(fmt.Sprintf("preloaded+layer:cflinuxfs3%s", queryString))
				Expect(err).NotTo(HaveOccurred())
				Expect(p).To(Equal(fmt.Sprintf("preloaded+layer:cflinuxfs3:/var/vcap/packages/cflinuxfs3/rootfs.tar%s", queryString)))
			})
			It("returns a blank string and no error if the RootFS URL is blank", func() {
				p, err := stackPathMap.PathForRootFS("")
				Expect(err).NotTo(HaveOccurred())
				Expect(p).To(Equal(""))
			})
			It("returns the same URL and no error if the RootFS scheme is docker", func() {
				p, err := stackPathMap.PathForRootFS("docker:///cfdiegodocker/grace")
				Expect(err).NotTo(HaveOccurred())
				Expect(p).To(Equal("docker:///cfdiegodocker/grace"))
			})
			It("returns an error if the RootFS URL is invalid", func() {
				_, err := stackPathMap.PathForRootFS("%x")
				Expect(err).To(HaveOccurred())
			})
			It("returns an error if the Preloaded RootFS path could not be found in the map", func() {
				_, err := stackPathMap.PathForRootFS("preloaded:not-on-cell")
				Expect(err).To(MatchError(rep.ErrPreloadedRootFSNotFound))
			})
		})
	})
})

func buildLRP(instanceGuid,
	guid,
	domain string,
	index int,
	rootFS string,
	memoryMB,
	diskMB,
	maxPids int32,
	placementTags,
	volumeDrivers []string,
	state string,
) *rep.LRP {
	lrpKey := models.NewActualLRPKey(guid, int32(index), domain)
	lrp := rep.NewLRP(instanceGuid, lrpKey, rep.NewResource(memoryMB, diskMB, maxPids), rep.PlacementConstraint{RootFs: rootFS,
		PlacementTags: placementTags,
		VolumeDrivers: volumeDrivers,
	})
	lrp.State = state
	return &lrp
}

func buildTask(taskGuid, domain, rootFS string, memoryMB, diskMB, maxPids int32, placementTags, volumeDrivers []string, state models.Task_State, failed bool) *rep.Task {
	task := rep.NewTask(taskGuid, domain, rep.NewResource(memoryMB, diskMB, maxPids), rep.PlacementConstraint{RootFs: rootFS, VolumeDrivers: volumeDrivers})
	return &task
}
