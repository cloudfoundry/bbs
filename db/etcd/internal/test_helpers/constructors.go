package test_helpers

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/gomega"
)

func (t *TestHelper) NewValidActualLRP(guid string, index int32) models.ActualLRP {
	actualLRP := models.ActualLRP{
		ActualLRPKey:         models.NewActualLRPKey(guid, index, "some-domain"),
		ActualLRPInstanceKey: models.NewActualLRPInstanceKey("some-guid", "some-cell"),
		ActualLRPNetInfo: models.NewActualLRPNetInfo("some-address", []*models.PortMapping{
			{HostPort: proto.Uint32(2222), ContainerPort: proto.Uint32(4444)},
		}),
		CrashCount:  proto.Int32(33),
		CrashReason: proto.String("badness"),
		State:       proto.String(models.ActualLRPStateRunning),
		Since:       proto.Int64(1138),
		ModificationTag: &models.ModificationTag{
			Epoch: proto.String("some-epoch"),
			Index: proto.Uint32(999),
		},
	}
	err := actualLRP.Validate()
	Expect(err).NotTo(HaveOccurred())

	return actualLRP
}
