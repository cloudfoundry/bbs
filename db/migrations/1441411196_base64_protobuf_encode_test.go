package migrations_test

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"time"

	"code.cloudfoundry.org/bbs/db/deprecations"
	"code.cloudfoundry.org/bbs/db/etcd"
	"code.cloudfoundry.org/bbs/db/migrations"
	"code.cloudfoundry.org/bbs/encryption"
	"code.cloudfoundry.org/bbs/format"
	"code.cloudfoundry.org/bbs/migration"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager/lagertest"
	goetcd "github.com/coreos/go-etcd/etcd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Base 64 Protobuf Encode Migration", func() {
	var (
		migration  migration.Migration
		serializer format.Serializer
		cryptor    encryption.Cryptor

		logger *lagertest.TestLogger
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")

		encryptionKey, err := encryption.NewKey("label", "passphrase")
		Expect(err).NotTo(HaveOccurred())
		keyManager, err := encryption.NewKeyManager(encryptionKey, nil)
		Expect(err).NotTo(HaveOccurred())
		cryptor = encryption.NewCryptor(keyManager, rand.Reader)
		serializer = format.NewSerializer(cryptor)
		migration = migrations.NewBase64ProtobufEncode()
	})

	It("appends itself to the migration list", func() {
		Expect(migrations.Migrations).To(ContainElement(migration))
	})

	Describe("Version", func() {
		It("returns the timestamp from which it was created", func() {
			Expect(migration.Version()).To(BeEquivalentTo(1441411196))
		})
	})

	var newValidDesiredLRP = func(guid string) *models.DesiredLRP {
		myRouterJSON := json.RawMessage(`{"foo":"bar"}`)
		desiredLRP := &models.DesiredLRP{
			ProcessGuid:             guid,
			Domain:                  "some-domain",
			RootFs:                  "some:rootfs",
			Instances:               1,
			EnvironmentVariables:    []*models.EnvironmentVariable{{Name: "FOO", Value: "bar"}},
			Setup:                   models.WrapAction(&models.RunAction{Path: "ls", User: "name"}),
			Action:                  models.WrapAction(&models.RunAction{Path: "ls", User: "name"}),
			DeprecatedStartTimeoutS: 15,
			Monitor: models.WrapAction(models.EmitProgressFor(
				models.Timeout(models.Try(models.Parallel(models.Serial(&models.RunAction{Path: "ls", User: "name"}))),
					10*time.Second,
				),
				"start-message",
				"success-message",
				"failure-message",
			)),
			DiskMb:      512,
			MemoryMb:    1024,
			CpuWeight:   42,
			Routes:      &models.Routes{"my-router": &myRouterJSON},
			LogSource:   "some-log-source",
			LogGuid:     "some-log-guid",
			MetricsGuid: "some-metrics-guid",
			Annotation:  "some-annotation",
			EgressRules: []*models.SecurityGroupRule{{
				Protocol:     models.TCPProtocol,
				Destinations: []string{"1.1.1.1/32", "2.2.2.2/32"},
				PortRange:    &models.PortRange{Start: 10, End: 16000},
			}},
		}
		err := desiredLRP.Validate()
		Expect(err).NotTo(HaveOccurred())

		return desiredLRP
	}

	var newValidActualLRP = func(guid string, index int32) *models.ActualLRP {
		ports := []*models.PortMapping{
			&models.PortMapping{2222, 4444},
		}

		actualLRP := &models.ActualLRP{
			ActualLRPKey:         models.ActualLRPKey{guid, index, "some-domain"},
			ActualLRPInstanceKey: models.ActualLRPInstanceKey{"some-guid", "some-cell"},
			ActualLRPNetInfo: models.ActualLRPNetInfo{
				Address: "some-address",
				Ports:   ports,
			},
			CrashCount:  33,
			CrashReason: "badness",
			State:       models.ActualLRPStateRunning,
			Since:       1138,
			ModificationTag: models.ModificationTag{
				Epoch: "some-epoch",
				Index: 999,
			},
		}
		err := actualLRP.Validate()
		Expect(err).NotTo(HaveOccurred())

		return actualLRP
	}

	var newTaskDefinition = func() *models.TaskDefinition {
		return &models.TaskDefinition{
			RootFs: "docker:///docker.com/docker",
			EnvironmentVariables: []*models.EnvironmentVariable{
				{
					Name:  "FOO",
					Value: "BAR",
				},
			},
			Action: models.WrapAction(&models.RunAction{
				User:           "user",
				Path:           "echo",
				Args:           []string{"hello world"},
				ResourceLimits: &models.ResourceLimits{},
			}),
			MemoryMb:    256,
			DiskMb:      1024,
			CpuWeight:   42,
			Privileged:  true,
			LogGuid:     "123",
			LogSource:   "APP",
			MetricsGuid: "456",
			ResultFile:  "some-file.txt",
			EgressRules: []*models.SecurityGroupRule{
				{
					Protocol:     "tcp",
					Destinations: []string{"0.0.0.0/0"},
					PortRange: &models.PortRange{
						Start: 1,
						End:   1024,
					},
					Log: true,
				},
				{
					Protocol:     "udp",
					Destinations: []string{"8.8.0.0/16"},
					Ports:        []uint32{53},
				},
			},

			Annotation: `[{"anything": "you want!"}]... dude`,
		}
	}

	var newValidTask = func(guid string) *models.Task {

		task := &models.Task{
			TaskGuid:       guid,
			Domain:         "some-domain",
			TaskDefinition: newTaskDefinition(),

			CreatedAt:        time.Date(2014, time.February, 25, 23, 46, 11, 00, time.UTC).UnixNano(),
			UpdatedAt:        time.Date(2014, time.February, 25, 23, 46, 11, 10, time.UTC).UnixNano(),
			FirstCompletedAt: time.Date(2014, time.February, 25, 23, 46, 11, 30, time.UTC).UnixNano(),

			CellId:        "cell",
			State:         models.Task_Pending,
			Result:        "turboencabulated",
			Failed:        true,
			FailureReason: "because i said so",
		}

		err := task.Validate()
		if err != nil {
			panic(err)
		}
		return task
	}

	Describe("Up", func() {
		var (
			expectedDesiredLRP                             *models.DesiredLRP
			expectedActualLRP, expectedEvacuatingActualLRP *models.ActualLRP
			expectedTask                                   *models.Task
			migrationErr                                   error
		)

		BeforeEach(func() {
			// DesiredLRP
			expectedDesiredLRP = newValidDesiredLRP("process-guid")
			jsonValue, err := json.Marshal(expectedDesiredLRP)
			Expect(err).NotTo(HaveOccurred())
			_, err = storeClient.Set(deprecations.DesiredLRPSchemaPath(expectedDesiredLRP), jsonValue, 0)
			Expect(err).NotTo(HaveOccurred())

			// ActualLRP
			expectedActualLRP = newValidActualLRP("process-guid", 1)
			jsonValue, err = json.Marshal(expectedActualLRP)
			Expect(err).NotTo(HaveOccurred())
			_, err = storeClient.Set(etcd.ActualLRPSchemaPath(expectedActualLRP.ProcessGuid, 1), jsonValue, 0)
			Expect(err).NotTo(HaveOccurred())

			// Evacuating ActualLRP
			expectedEvacuatingActualLRP = newValidActualLRP("process-guid", 4)
			jsonValue, err = json.Marshal(expectedEvacuatingActualLRP)
			Expect(err).NotTo(HaveOccurred())
			_, err = storeClient.Set(
				etcd.EvacuatingActualLRPSchemaPath(expectedEvacuatingActualLRP.ProcessGuid, 1),
				jsonValue,
				0,
			)
			Expect(err).NotTo(HaveOccurred())

			// Tasks
			expectedTask = newValidTask("task-guid")
			jsonValue, err = json.Marshal(expectedTask)
			Expect(err).NotTo(HaveOccurred())
			_, err = storeClient.Set(etcd.TaskSchemaPath(expectedTask), jsonValue, 0)
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			migration.SetStoreClient(storeClient)
			migration.SetCryptor(cryptor)
			migrationErr = migration.Up(logger)
		})

		var validateConversionToProto = func(node *goetcd.Node, actual, expected format.Versioner) {
			value := node.Value

			Expect(value[:2]).To(BeEquivalentTo(format.BASE64[:]))
			payload, err := base64.StdEncoding.DecodeString(string(value[2:]))
			Expect(err).NotTo(HaveOccurred())
			Expect(payload[0]).To(BeEquivalentTo(format.PROTO))
			serializer.Unmarshal(logger, []byte(value), actual)
			Expect(actual).To(Equal(expected))
		}

		It("converts all data stored in the etcd store to base 64 protobuf", func() {
			Expect(migrationErr).NotTo(HaveOccurred())

			By("Converting DesiredLRPs to Encoded Proto")
			response, err := storeClient.Get(deprecations.DesiredLRPSchemaRoot, false, true)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Node.Nodes).To(HaveLen(1))
			for _, node := range response.Node.Nodes {
				var desiredLRP models.DesiredLRP
				value := node.Value
				err := serializer.Unmarshal(logger, []byte(value), &desiredLRP)
				Expect(err).NotTo(HaveOccurred())
				validateConversionToProto(node, &desiredLRP, expectedDesiredLRP)
			}

			By("Converting ActualLRPs to Encoded Proto")
			response, err = storeClient.Get(etcd.ActualLRPSchemaRoot, false, true)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Node.Nodes).To(HaveLen(1))
			for _, processNode := range response.Node.Nodes {
				for _, groupNode := range processNode.Nodes {
					for _, lrpNode := range groupNode.Nodes {
						var expected *models.ActualLRP
						if lrpNode.Key == etcd.ActualLRPSchemaPath("process-guid", 1) {
							expected = expectedActualLRP
						} else {
							expected = expectedEvacuatingActualLRP
						}
						var actualLRP models.ActualLRP
						serializer.Unmarshal(logger, []byte(lrpNode.Value), &actualLRP)
						validateConversionToProto(lrpNode, &actualLRP, expected)
					}
				}
			}

			By("Converting Tasks to Encoded Proto")
			response, err = storeClient.Get(etcd.TaskSchemaRoot, false, true)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Node.Nodes).To(HaveLen(1))
			for _, taskNode := range response.Node.Nodes {
				var task models.Task
				serializer.Unmarshal(logger, []byte(taskNode.Value), &task)
				validateConversionToProto(taskNode, &task, expectedTask)
			}
		})

		Context("when fetching desired lrps fails", func() {
			Context("because the root node does not exist", func() {
				BeforeEach(func() {
					_, err := storeClient.Delete(deprecations.DesiredLRPSchemaRoot, true)
					Expect(err).NotTo(HaveOccurred())
				})

				It("continues the migration", func() {
					Expect(migrationErr).NotTo(HaveOccurred())
				})
			})
		})

		Context("when fetching actual lrps fails", func() {
			Context("because the root node does not exist", func() {
				BeforeEach(func() {
					_, err := storeClient.Delete(etcd.ActualLRPSchemaRoot, true)
					Expect(err).NotTo(HaveOccurred())
				})

				It("continues the migration", func() {
					Expect(migrationErr).NotTo(HaveOccurred())
				})
			})
		})

		Context("when fetching tasks fails", func() {
			Context("because the root node does not exist", func() {
				BeforeEach(func() {
					_, err := storeClient.Delete(etcd.TaskSchemaRoot, true)
					Expect(err).NotTo(HaveOccurred())
				})

				It("continues the migration", func() {
					Expect(migrationErr).NotTo(HaveOccurred())
				})
			})
		})
	})

	Describe("Down", func() {
		It("returns a not implemented error", func() {
			Expect(migration.Down(logger)).To(HaveOccurred())
		})
	})
})
