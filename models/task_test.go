package models_test

import (
	"encoding/json"
	"strings"
	"time"

	"code.cloudfoundry.org/bbs/format"
	"code.cloudfoundry.org/bbs/models"
	. "code.cloudfoundry.org/bbs/test_helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"
)

var _ = Describe("Task", func() {
	var taskPayload string
	var task models.Task

	BeforeEach(func() {
		taskPayload = `{
		"task_guid":"some-guid",
		"domain":"some-domain",
		"cell_id":"cell",
		"result": "turboencabulated",
		"failed":true,
		"failure_reason":"because i said so",
		"created_at": 1393371971000000000,
		"updated_at": 1393371971000000010,
		"first_completed_at": 1393371971000000030,
		"state": "Pending",
		"rejection_count": 0,
		"rejection_reason": "",
		"task_definition" : {
			"rootfs": "docker:///docker.com/docker",
			"env":[
				{
					"name":"ENV_VAR_NAME",
					"value":"an environmment value"
				}
			],
			"action": {
				"download":{
					"from":"old_location",
					"to":"new_location",
					"cache_key":"the-cache-key",
					"user":"someone",
					"checksum_algorithm": "md5",
					"checksum_value": "some value"
				}
			},
			"result_file":"some-file.txt",
			"memory_mb":256,
			"disk_mb":1024,
			"log_rate_limit": {
				"bytes_per_second": 2048
			},
			"cpu_weight": 42,
			"privileged": true,
			"log_guid": "123",
			"log_source": "APP",
			"metrics_guid": "456",
			"annotation": "[{\"anything\": \"you want!\"}]... dude",
			"network": {
				"properties": {
					"some-key": "some-value",
					"some-other-key": "some-other-value"
				}
			},
			"egress_rules": [
				{
					"protocol": "tcp",
					"destinations": ["0.0.0.0/0"],
					"port_range": {
						"start": 1,
						"end": 1024
					},
					"log": true
				},
				{
					"protocol": "udp",
					"destinations": ["8.8.0.0/16"],
					"ports": [53],
					"log": false
				}
			],
			"completion_callback_url":"http://user:password@a.b.c/d/e/f",
			"max_pids": 256,
			"certificate_properties": {
				"organizational_unit": ["stuff"]
			},
			"image_username": "jake",
			"image_password": "thedog",
			"image_layers": [
				{
					"url": "some-url",
					"destination_path": "/tmp",
					"media_type": "TGZ",
					"layer_type": "SHARED"
				}
			],
			"legacy_download_user": "some-user",
			"metric_tags": {
			  "source_id": {
				  "static": "some-guid"
			  },
			  "foo": {
				  "static": "some-value"
			  },
			  "bar": {
				  "dynamic": "INDEX"
			  }
			},
			"volume_mounted_files": [
				{"path": "/redis/username", "content": "username"}
			]
		  }
		}`

		err := json.Unmarshal([]byte(taskPayload), &task)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("serialization", func() {
		It("successfully round trips through json and protobuf", func() {
			jsonSerialization, err := json.Marshal(task)
			Expect(err).NotTo(HaveOccurred())
			Expect(jsonSerialization).To(MatchJSON(taskPayload))

			protoSerialization, err := proto.Marshal(task.ToProto())
			Expect(err).NotTo(HaveOccurred())

			var protoDeserialization models.ProtoTask
			err = proto.Unmarshal(protoSerialization, &protoDeserialization)
			Expect(err).NotTo(HaveOccurred())
			Expect(*protoDeserialization.FromProto()).To(Equal(task))
		})
	})

	Describe("Validate", func() {
		Context("when the task has a domain, valid guid, stack, and valid action", func() {
			It("is valid", func() {
				task = models.Task{
					Domain:   "some-domain",
					TaskGuid: "some-task-guid",
					TaskDefinition: &models.TaskDefinition{
						RootFs: "some:rootfs",
						Action: models.WrapAction(&models.RunAction{
							Path: "ls",
							User: "me",
						}),
					},
				}

				err := task.Validate()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the task GUID is present but invalid", func() {
			It("returns an error indicating so", func() {
				task = models.Task{
					Domain:   "some-domain",
					TaskGuid: "invalid/guid",
					TaskDefinition: &models.TaskDefinition{
						RootFs: "some:rootfs",
						Action: models.WrapAction(&models.RunAction{
							Path: "ls",
							User: "me",
						}),
					},
				}

				err := task.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("task_guid"))
			})
		})

		for _, testCase := range []ValidatorErrorCase{
			{
				"task_guid",
				&models.Task{
					Domain: "some-domain",
					TaskDefinition: &models.TaskDefinition{
						RootFs: "some:rootfs",
						Action: models.WrapAction(&models.RunAction{
							Path: "ls",
							User: "me",
						}),
					},
				},
			},
			{
				"rootfs",
				&models.Task{
					Domain:   "some-domain",
					TaskGuid: "task-guid",
					TaskDefinition: &models.TaskDefinition{
						Action: models.WrapAction(&models.RunAction{
							Path: "ls",
							User: "me",
						}),
					},
				},
			},
			{
				"rootfs",
				&models.Task{
					Domain:   "some-domain",
					TaskGuid: "task-guid",
					TaskDefinition: &models.TaskDefinition{
						RootFs: ":invalid-url",
						Action: models.WrapAction(&models.RunAction{
							Path: "ls",
							User: "me",
						}),
					},
				},
			},
			{
				"rootfs",
				&models.Task{
					Domain:   "some-domain",
					TaskGuid: "task-guid",
					TaskDefinition: &models.TaskDefinition{
						RootFs: "invalid-absolute-url",
						Action: models.WrapAction(&models.RunAction{
							Path: "ls",
							User: "me",
						}),
					},
				},
			},
			{
				"domain",
				&models.Task{
					TaskGuid: "task-guid",
					TaskDefinition: &models.TaskDefinition{
						RootFs: "some:rootfs",
						Action: models.WrapAction(&models.RunAction{
							Path: "ls",
							User: "me",
						}),
					},
				},
			},
			{
				"action",
				&models.Task{
					Domain:   "some-domain",
					TaskGuid: "task-guid",
					TaskDefinition: &models.TaskDefinition{
						RootFs: "some:rootfs",
						Action: nil,
					},
				}},
			{
				"path",
				&models.Task{
					Domain:   "some-domain",
					TaskGuid: "task-guid",
					TaskDefinition: &models.TaskDefinition{
						RootFs: "some:rootfs",
						Action: models.WrapAction(&models.RunAction{User: "me"}),
					},
				},
			},
			{
				"annotation",
				&models.Task{
					Domain:   "some-domain",
					TaskGuid: "task-guid",
					TaskDefinition: &models.TaskDefinition{
						RootFs: "some:rootfs",
						Action: models.WrapAction(&models.RunAction{
							Path: "ls",
							User: "me",
						}),
						Annotation: strings.Repeat("a", 10*1024+1),
					},
				},
			},
			{
				"memory_mb",
				&models.Task{
					Domain:   "some-domain",
					TaskGuid: "task-guid",
					TaskDefinition: &models.TaskDefinition{
						RootFs: "some:rootfs",
						Action: models.WrapAction(&models.RunAction{
							Path: "ls",
							User: "me",
						}),
						MemoryMb: -1,
					},
				},
			},
			{
				"disk_mb",
				&models.Task{
					Domain:   "some-domain",
					TaskGuid: "task-guid",
					TaskDefinition: &models.TaskDefinition{
						RootFs: "some:rootfs",
						Action: models.WrapAction(&models.RunAction{
							Path: "ls",
							User: "me",
						}),
						DiskMb: -1,
					},
				},
			},
			{
				"log_rate_limit",
				&models.Task{
					Domain:   "some-domain",
					TaskGuid: "task-guid",
					TaskDefinition: &models.TaskDefinition{
						RootFs: "some:rootfs",
						Action: models.WrapAction(&models.RunAction{
							Path: "ls",
							User: "me",
						}),
						LogRateLimit: &models.LogRateLimit{BytesPerSecond: -2},
					},
				},
			},

			{
				"max_pids",
				&models.Task{
					Domain:   "some-domain",
					TaskGuid: "task-guid",
					TaskDefinition: &models.TaskDefinition{
						RootFs: "some:rootfs",
						Action: models.WrapAction(&models.RunAction{
							Path: "ls",
							User: "me",
						}),
						MaxPids: -1,
					},
				},
			},
			{
				"egress_rules",
				&models.Task{
					Domain:   "some-domain",
					TaskGuid: "task-guid",
					TaskDefinition: &models.TaskDefinition{
						RootFs: "some:rootfs",
						Action: models.WrapAction(&models.RunAction{
							Path: "ls",
							User: "me",
						}),
						EgressRules: []*models.SecurityGroupRule{
							{Protocol: "invalid"},
						},
					},
				},
			},
			{
				"cached_dependency",
				&models.Task{
					TaskGuid: "guid-1",
					Domain:   "some-domain",
					TaskDefinition: &models.TaskDefinition{
						RootFs: "some-rootfs",
						CachedDependencies: []*models.CachedDependency{
							{
								To: "here",
							},
						},
					},
				},
			},
			{
				"invalid algorithm",
				&models.Task{
					TaskGuid: "guid-1",
					Domain:   "some-domain",
					TaskDefinition: &models.TaskDefinition{
						RootFs: "some-rootfs",
						CachedDependencies: []*models.CachedDependency{
							{
								To:                "here",
								From:              "there",
								ChecksumAlgorithm: "wrong algorithm",
								ChecksumValue:     "some value",
							},
						},
					},
				},
			},
			{
				"image_username",
				&models.Task{
					Domain:   "some-domain",
					TaskGuid: "task-guid",
					TaskDefinition: &models.TaskDefinition{
						RootFs: "some:rootfs",
						Action: models.WrapAction(&models.RunAction{
							Path: "ls",
							User: "me",
						}),
						ImageUsername: "",
						ImagePassword: "thedog",
					},
				},
			},
			{
				"image_password",
				&models.Task{
					Domain:   "some-domain",
					TaskGuid: "task-guid",
					TaskDefinition: &models.TaskDefinition{
						RootFs: "some:rootfs",
						Action: models.WrapAction(&models.RunAction{
							Path: "ls",
							User: "me",
						}),
						ImageUsername: "jake",
						ImagePassword: "",
					},
				},
			},
			{
				"image_layer",
				&models.Task{
					Domain:   "some-domain",
					TaskGuid: "task-guid",
					TaskDefinition: &models.TaskDefinition{
						RootFs: "some:rootfs",
						Action: models.WrapAction(&models.RunAction{
							Path: "ls",
							User: "me",
						}),
						ImageUsername: "jake",
						ImagePassword: "pass",
						ImageLayers: []*models.ImageLayer{
							{Url: "some-url", DestinationPath: "", MediaType: models.ImageLayer_MediaTypeTgz}, // invalid destination path
						},
					},
				},
			},
			{
				"legacy_download_user",
				&models.Task{
					Domain:   "some-domain",
					TaskGuid: "task-guid",
					TaskDefinition: &models.TaskDefinition{
						RootFs: "some:rootfs",
						Action: models.WrapAction(&models.RunAction{
							Path: "ls",
							User: "me",
						}),
						ImageUsername: "jake",
						ImagePassword: "pass",
						ImageLayers: []*models.ImageLayer{
							{Url: "some-url", DestinationPath: "/tmp", MediaType: models.ImageLayer_MediaTypeTgz, LayerType: models.ImageLayer_LayerTypeExclusive}, // exclusive layers require legacy_download_user to be set
						},
					},
				},
			},
		} {
			testValidatorErrorCase(testCase)
		}
	})

	Describe("VersionDownTo", func() {
		var task *models.Task

		BeforeEach(func() {
			task = &models.Task{
				TaskDefinition: &models.TaskDefinition{},
			}
		})

		Context("V3->V2", func() {
			Context("when there are no image layers", func() {
				BeforeEach(func() {
					task.TaskDefinition.ImageLayers = nil
				})

				It("does not add any cached dependencies to the TaskDefinition", func() {
					convertedTask := task.VersionDownTo(format.V2)
					Expect(convertedTask.TaskDefinition.CachedDependencies).To(BeEmpty())
				})

				It("does not add any Download Actions", func() {
					convertedTask := task.VersionDownTo(format.V2)
					Expect(convertedTask.TaskDefinition.Action).To(Equal(task.TaskDefinition.Action))
				})
			})

			Context("when there are shared image layers", func() {
				BeforeEach(func() {
					task.TaskDefinition.ImageLayers = []*models.ImageLayer{
						{
							Name:            "dep0",
							Url:             "u0",
							DestinationPath: "/tmp/0",
							LayerType:       models.ImageLayer_LayerTypeShared,
							MediaType:       models.ImageLayer_MediaTypeTgz,
							DigestAlgorithm: models.ImageLayer_DigestAlgorithmSha256,
							DigestValue:     "some-sha",
						},
						{
							Name:            "dep1",
							Url:             "u1",
							DestinationPath: "/tmp/1",
							LayerType:       models.ImageLayer_LayerTypeShared,
							MediaType:       models.ImageLayer_MediaTypeTgz,
						},
					}

					task.TaskDefinition.CachedDependencies = []*models.CachedDependency{
						{
							Name:      "dep2",
							From:      "u2",
							To:        "/tmp/2",
							CacheKey:  "key2",
							LogSource: "download",
						},
					}
				})

				It("converts them to cached dependencies and prepends them to the list", func() {
					convertedTask := task.VersionDownTo(format.V2)
					Expect(convertedTask.TaskDefinition.CachedDependencies).To(DeepEqual([]*models.CachedDependency{
						{
							Name:              "dep0",
							From:              "u0",
							To:                "/tmp/0",
							CacheKey:          "sha256:some-sha",
							LogSource:         "",
							ChecksumAlgorithm: "sha256",
							ChecksumValue:     "some-sha",
						},
						{
							Name:      "dep1",
							From:      "u1",
							To:        "/tmp/1",
							CacheKey:  "u1",
							LogSource: "",
						},
						{
							Name:      "dep2",
							From:      "u2",
							To:        "/tmp/2",
							CacheKey:  "key2",
							LogSource: "download",
						},
					}))
				})

				It("sets removes the existing image layers", func() {
					convertedTask := task.VersionDownTo(format.V2)
					Expect(convertedTask.TaskDefinition.ImageLayers).To(BeNil())
				})
			})

			Context("when there are exclusive image layers", func() {
				var (
					downloadAction1, downloadAction2 models.DownloadAction
				)

				BeforeEach(func() {
					task.TaskDefinition.ImageLayers = []*models.ImageLayer{
						{
							Name:            "dep0",
							Url:             "u0",
							DestinationPath: "/tmp/0",
							LayerType:       models.ImageLayer_LayerTypeExclusive,
							MediaType:       models.ImageLayer_MediaTypeTgz,
							DigestAlgorithm: models.ImageLayer_DigestAlgorithmSha256,
							DigestValue:     "some-sha",
						},
						{
							Name:            "dep1",
							Url:             "u1",
							DestinationPath: "/tmp/1",
							LayerType:       models.ImageLayer_LayerTypeExclusive,
							MediaType:       models.ImageLayer_MediaTypeTgz,
							DigestAlgorithm: models.ImageLayer_DigestAlgorithmSha256,
							DigestValue:     "some-other-sha",
						},
					}
					task.TaskDefinition.LegacyDownloadUser = "the user"
					task.TaskDefinition.Action = models.WrapAction(models.Timeout(
						&models.RunAction{
							Path: "/the/path",
							User: "the user",
						},
						20*time.Millisecond,
					))

					downloadAction1 = models.DownloadAction{
						Artifact:          "dep0",
						From:              "u0",
						To:                "/tmp/0",
						CacheKey:          "sha256:some-sha",
						LogSource:         "",
						User:              "the user",
						ChecksumAlgorithm: "sha256",
						ChecksumValue:     "some-sha",
					}
					downloadAction2 = models.DownloadAction{
						Artifact:          "dep1",
						From:              "u1",
						To:                "/tmp/1",
						CacheKey:          "sha256:some-other-sha",
						LogSource:         "",
						User:              "the user",
						ChecksumAlgorithm: "sha256",
						ChecksumValue:     "some-other-sha",
					}
				})

				It("converts them to download actions with the correct user and prepends them to the action", func() {
					convertedTask := task.VersionDownTo(format.V2)

					Expect(convertedTask.TaskDefinition.Action.GetValue()).To(DeepEqual(
						models.Serial(
							models.Parallel(&downloadAction1, &downloadAction2),
							task.TaskDefinition.Action.GetValue().(models.ActionInterface),
						)))
				})

				It("sets removes the existing image layers", func() {
					convertedTask := task.VersionDownTo(format.V2)
					Expect(convertedTask.TaskDefinition.ImageLayers).To(BeNil())
				})

				Context("when there is no existing action", func() {
					BeforeEach(func() {
						task.TaskDefinition.Action = nil
					})

					It("creates an action with exclusive layers converted to download actions", func() {
						convertedLRP := task.VersionDownTo(format.V2)
						Expect(convertedLRP.TaskDefinition.Action.GetValue()).To(DeepEqual(
							models.Parallel(&downloadAction1, &downloadAction2),
						))
					})
				})
			})
		})
	})

	Describe("State", func() {
		Describe("MarshalJSON", func() {
			DescribeTable("marshals and unmarshals between the value and the expected JSON output",
				func(v models.Task_State, expectedJSON string) {
					Expect(json.Marshal(v)).To(MatchJSON(expectedJSON))
					var testV models.Task_State
					Expect(json.Unmarshal([]byte(expectedJSON), &testV)).To(Succeed())
					Expect(testV).To(Equal(v))
				},
				Entry("invalid", models.Task_Invalid, `"Invalid"`),
				Entry("pending", models.Task_Pending, `"Pending"`),
				Entry("running", models.Task_Running, `"Running"`),
				Entry("completed", models.Task_Completed, `"Completed"`),
				Entry("resolving", models.Task_Resolving, `"Resolving"`),
			)
		})
	})
})
