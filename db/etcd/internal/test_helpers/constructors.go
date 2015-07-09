package test_helpers

import (
	"encoding/json"
	"time"

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

func (t *TestHelper) NewValidDesiredLRP(guid string) models.DesiredLRP {
	myRouterJSON := json.RawMessage(`{"foo":"bar"}`)
	desiredLRP := models.DesiredLRP{
		ProcessGuid:          &guid,
		Domain:               proto.String("some-domain"),
		RootFs:               proto.String("some:rootfs"),
		Instances:            proto.Int32(1),
		EnvironmentVariables: []*models.EnvironmentVariable{{Name: proto.String("FOO"), Value: proto.String("bar")}},
		Setup:                &models.Action{RunAction: &models.RunAction{Path: proto.String("ls"), User: proto.String("name")}},
		Action:               &models.Action{RunAction: &models.RunAction{Path: proto.String("ls"), User: proto.String("name")}},
		StartTimeout:         proto.Uint32(15),
		Monitor: models.EmitProgressFor(
			models.Timeout(models.Try(models.Parallel(models.Serial(
				models.WrapAction(&models.RunAction{Path: proto.String("ls"), User: proto.String("name")})))),
				10*time.Second,
			),
			"start-message",
			"success-message",
			"failure-message",
		),
		DiskMb:      proto.Int32(512),
		MemoryMb:    proto.Int32(1024),
		CpuWeight:   proto.Uint32(42),
		Routes:      &models.Routes{"my-router": &myRouterJSON},
		LogSource:   proto.String("some-log-source"),
		LogGuid:     proto.String("some-log-guid"),
		MetricsGuid: proto.String("some-metrics-guid"),
		Annotation:  proto.String("some-annotation"),
		EgressRules: []*models.SecurityGroupRule{{
			Protocol:     proto.String(models.TCPProtocol),
			Destinations: []string{"1.1.1.1/32", "2.2.2.2/32"},
			PortRange:    &models.PortRange{Start: proto.Uint32(10), End: proto.Uint32(16000)},
		}},
	}
	err := desiredLRP.Validate()
	Expect(err).NotTo(HaveOccurred())

	return desiredLRP
}
