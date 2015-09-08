package migrations_test

import (
	"encoding/base64"
	"encoding/json"

	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/db/migrations"
	"github.com/cloudfoundry-incubator/bbs/format"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/models/test/model_helpers"
	goetcd "github.com/coreos/go-etcd/etcd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("Base 64 Protobuf Encode Migration", func() {
	var (
		migration migrations.Base64ProtobufEncode

		logger *lagertest.TestLogger
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")

		migration = migrations.NewBase64ProtobufEncode()
	})

	Describe("Version", func() {
		It("returns the timestamp from which it was created", func() {
			Expect(migration.Version()).To(BeEquivalentTo(1441411196))
		})
	})

	Describe("Up", func() {
		var (
			expectedDesiredLRP                             *models.DesiredLRP
			expectedActualLRP, expectedEvacuatingActualLRP *models.ActualLRP
			expectedTask                                   *models.Task
			migrationErr                                   error
		)

		BeforeEach(func() {
			// DesiredLRP
			expectedDesiredLRP = model_helpers.NewValidDesiredLRP("process-guid")
			jsonValue, err := json.Marshal(expectedDesiredLRP)
			Expect(err).NotTo(HaveOccurred())
			_, err = storeClient.Set(etcd.DesiredLRPSchemaPath(expectedDesiredLRP), jsonValue, 0)
			Expect(err).NotTo(HaveOccurred())

			// ActualLRP
			expectedActualLRP = model_helpers.NewValidActualLRP("process-guid", 1)
			jsonValue, err = json.Marshal(expectedActualLRP)
			Expect(err).NotTo(HaveOccurred())
			_, err = storeClient.Set(etcd.ActualLRPSchemaPath(expectedActualLRP.ProcessGuid, 1), jsonValue, 0)
			Expect(err).NotTo(HaveOccurred())

			// Evacuating ActualLRP
			expectedEvacuatingActualLRP = model_helpers.NewValidActualLRP("process-guid", 4)
			jsonValue, err = json.Marshal(expectedEvacuatingActualLRP)
			Expect(err).NotTo(HaveOccurred())
			_, err = storeClient.Set(
				etcd.EvacuatingActualLRPSchemaPath(expectedEvacuatingActualLRP.ProcessGuid, 1),
				jsonValue,
				0,
			)
			Expect(err).NotTo(HaveOccurred())

			// Tasks
			expectedTask = model_helpers.NewValidTask("task-guid")
			jsonValue, err = json.Marshal(expectedTask)
			Expect(err).NotTo(HaveOccurred())
			_, err = storeClient.Set(etcd.TaskSchemaPath(expectedTask), jsonValue, 0)
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			migrationErr = migration.Up(logger, storeClient)
		})

		var validateConversionToProto = func(node *goetcd.Node, actual, expected format.Versioner) {
			value := node.Value

			Expect(value[:2]).To(BeEquivalentTo(format.BASE64[:]))
			payload, err := base64.StdEncoding.DecodeString(string(value[2:]))
			Expect(err).NotTo(HaveOccurred())
			Expect(payload[0]).To(BeEquivalentTo(format.PROTO))
			Expect(payload[1]).To(BeEquivalentTo(format.V0))
			format.Unmarshal(logger, []byte(value), actual)
			Expect(actual).To(Equal(expected))
		}

		It("converts all data stored in the etcd store to base 64 protobuf", func() {
			Expect(migrationErr).NotTo(HaveOccurred())

			By("Converting DesiredLRPs to Encoded Proto")
			response, err := storeClient.Get(etcd.DesiredLRPSchemaRoot, false, true)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Node.Nodes).To(HaveLen(1))
			for _, node := range response.Node.Nodes {
				var desiredLRP models.DesiredLRP
				value := node.Value
				format.Unmarshal(logger, []byte(value), &desiredLRP)
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
						format.Unmarshal(logger, []byte(lrpNode.Value), &actualLRP)
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
				format.Unmarshal(logger, []byte(taskNode.Value), &task)
				validateConversionToProto(taskNode, &task, expectedTask)
			}
		})

		Context("when fetching desired lrps fails", func() {
			Context("because the root node does not exist", func() {
				BeforeEach(func() {
					_, err := storeClient.Delete(etcd.DesiredLRPSchemaRoot, true)
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
			Expect(migration.Down(logger, storeClient)).To(HaveOccurred())
		})
	})
})
