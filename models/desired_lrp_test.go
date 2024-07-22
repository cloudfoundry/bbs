package models_test

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"code.cloudfoundry.org/bbs/format"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/models/test/model_helpers"
	"github.com/gogo/protobuf/proto"

	. "code.cloudfoundry.org/bbs/test_helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("DesiredLRP", func() {
	var desiredLRP models.DesiredLRP

	jsonDesiredLRP := `{
    "setup": {
      "serial": {
        "actions": [
          {
            "download": {
              "from": "http://file-server.service.cf.internal:8080/v1/static/buildpack_app_lifecycle/buildpack_app_lifecycle.tgz",
              "to": "/tmp/lifecycle",
              "cache_key": "buildpack-cflinuxfs3-lifecycle",
							"user": "someone",
							"checksum_algorithm": "md5",
							"checksum_value": "some random value"
            }
          },
          {
            "download": {
              "from": "http://cloud-controller-ng.service.cf.internal:9022/internal/v2/droplets/some-guid/some-guid/download",
              "to": ".",
              "cache_key": "droplets-some-guid",
							"user": "someone"
            }
          }
        ]
      }
    },
    "action": {
      "codependent": {
        "actions": [
          {
            "run": {
              "path": "/tmp/lifecycle/launcher",
              "args": [
                "app",
                "",
                "{\"start_command\":\"bundle exec rackup config.ru -p $PORT\"}"
              ],
              "env": [
                {
                  "name": "VCAP_APPLICATION",
                  "value": "{\"limits\":{\"mem\":1024,\"disk\":1024,\"fds\":16384},\"application_id\":\"some-guid\",\"application_version\":\"some-guid\",\"application_name\":\"some-guid\",\"version\":\"some-guid\",\"name\":\"some-guid\",\"space_name\":\"CATS-SPACE-3-2015_07_01-11h28m01.515s\",\"space_id\":\"bc640806-ea03-40c6-8371-1c2b23fa4662\"}"
                },
                {
                  "name": "VCAP_SERVICES",
                  "value": "{}"
                },
                {
                  "name": "MEMORY_LIMIT",
                  "value": "1024m"
                },
                {
                  "name": "CF_STACK",
                  "value": "cflinuxfs3"
                },
                {
                  "name": "PORT",
                  "value": "8080"
                }
              ],
              "resource_limits": {
                "nofile": 16384
              },
              "user": "vcap",
              "log_source": "APP",
			  "suppress_log_output": false,
              "service_binding_files": null
            }
          },
          {
            "run": {
              "path": "/tmp/lifecycle/diego-sshd",
              "args": [
                "-address=0.0.0.0:2222",
                "-hostKey=-----BEGIN RSA PRIVATE KEY-----\nMIICWwIBAAKBgQCp72ylz6ow8P4km1Nzd2yyN9aiXAI8MHl6Crl6vjpBNQIhy+YH\nEf5fgAI/wHydaajSsk28Byf/hAm/Q/3EmT1bUmdCsVzzndzJvPNf5t11LGmPFcNV\nZ9vsfnFjMlsFM/ZHU60PT8POSoE8VnrplTLRhEtQFopdMcDN8nRl6imhUQIDAQAB\nAoGAWz8aQbZOFlVwwUs99gQsM03US/3HnXYR5DwZ+BRox1alPGx1qVo6EiF0E7NR\ntlxjsC7ZmprlGUhWy4LAom3+CUj712fI7Qnud9AH4GUHN4JrxytiDDLJJh/hRADB\niD/MKo9ih7c2bQvBU+FwLYlXyI/GViBMqIYzZ+6r7yVkp/kCQQDZIcMKzNwVV+LL\nnDXZg4nIyFgR3CGZb+cVrXnDaIEwmC5ABHlnhJJzI7FdsGuhwOJnKdMHQgI6+o+Z\nvmizsdyDAkEAyFrXDX+wRMPrEjmNga2TYaCIt6AWR3b4aLJskZQnf0iMI2DzL74e\na7Ibkxp+OxtSL2YIR7NCfDz/DiUtqvQKmwJAVRxX0K72geM+QiOMNCPMaYimhPGt\ntfBYO3YRaZhYM40ja/KVCA++PCW8i4Xw2qm51UhesNSd/TJkAZbSgcVxMwJAQSKX\nK4JJkfGHqKMhR/lgIqsIB3p6A72/wHnRJfreZFj3hkDsjqbmSOjcYhSI2Tpmm5Y2\nNukmQjGqUbTwhdVU5QJALpewrw7eiWAjnYxus6Fi0XiEduE91OEtuc3yHRrR0ubI\nCt2HP6jQ43siwcx+FAA8kBfvtQElIC2TF2qwjezEcA==\n-----END RSA PRIVATE KEY-----\n",
                "-authorizedKey=ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAgQDuOfcUnfiXE6g6Cvgur3Om6t8cEx27FAoVrDrxMzy+q2NTJaQFNYqG2DDDHZCLG2mJasryKZfDyK30c48ITpecBkCux429aZN2gEJCEsyYgsZheI+5eNYs1vzl68KQ1LdxlgNOqFZijyVjTOD60GMPCVlDICqGNUFH4aPTHA0fVw==\n",
                "-inheritDaemonEnv",
                "-logLevel=fatal"
              ],
              "env": [
                {
                  "name": "VCAP_APPLICATION",
                  "value": "{\"limits\":{\"mem\":1024,\"disk\":1024,\"fds\":16384},\"application_id\":\"some-guid\",\"application_version\":\"some-guid\",\"application_name\":\"some-guid\",\"version\":\"some-guid\",\"name\":\"some-guid\",\"space_name\":\"CATS-SPACE-3-2015_07_01-11h28m01.515s\",\"space_id\":\"some-guid\"}"
                },
                {
                  "name": "VCAP_SERVICES",
                  "value": "{}"
                },
                {
                  "name": "MEMORY_LIMIT",
                  "value": "1024m"
                },
                {
                  "name": "CF_STACK",
									"value": "cflinuxfs3"
                },
                {
                  "name": "PORT",
                  "value": "8080"
                }
              ],
              "resource_limits": {
                "nofile": 16384
              },
              "user": "vcap",
			  "suppress_log_output": false,
              "service_binding_files": null
            }
          }
        ]
      }
    },
    "monitor": {
      "timeout": {
        "action": {
          "run": {
            "path": "/tmp/lifecycle/healthcheck",
            "args": [
              "-port=8080"
            ],
            "resource_limits": {
              "nofile": 1024
            },
            "user": "vcap",
            "log_source": "HEALTH",
			"suppress_log_output": true,
		    "service_binding_files": null
          }
        },
        "timeout_ms": 30000000
      }
    },
    "process_guid": "some-guid",
    "domain": "cf-apps",
    "rootfs": "preloaded:cflinuxfs3",
    "instances": 2,
    "env": [
      {
        "name": "LANG",
        "value": "en_US.UTF-8"
      }
    ],
    "start_timeout_ms": 60000,
    "disk_mb": 1024,
    "memory_mb": 1024,
    "cpu_weight": 10,
    "privileged": true,
    "ports": [
      8080,
      2222
    ],
    "routes": {
      "cf-router": [
        {
          "hostnames": [
            "some-route.example.com"
          ],
          "port": 8080
        }
      ],
      "diego-ssh": {
        "container_port": 2222,
        "host_fingerprint": "ac:99:67:20:7e:c2:7c:2c:d2:22:37:bc:9f:14:01:ec",
        "private_key": "-----BEGIN RSA PRIVATE KEY-----\nMIICXQIBAAKBgQDuOfcUnfiXE6g6Cvgur3Om6t8cEx27FAoVrDrxMzy+q2NTJaQF\nNYqG2DDDHZCLG2mJasryKZfDyK30c48ITpecBkCux429aZN2gEJCEsyYgsZheI+5\neNYs1vzl68KQ1LdxlgNOqFZijyVjTOD60GMPCVlDICqGNUFH4aPTHA0fVwIDAQAB\nAoGBAO1Ak19YGHy1mgP8asFsAT1KitrV+vUW9xgwiB8xjRzDac8kHJ8HfKfg5Wdc\nqViw+0FdNzNH0xqsYPqkn92BECDqdWOzhlEYNj/AFSHTdRPrs9w82b7h/LhrX0H/\nRUrU2QrcI2uSV/SQfQvFwC6YaYugCo35noljJEcD8EYQTcRxAkEA+jfjumM6da8O\n8u8Rc58Tih1C5mumeIfJMPKRz3FBLQEylyMWtGlr1XT6ppqiHkAAkQRUBgKi+Ffi\nYedQOvE0/wJBAPO7I+brmrknzOGtSK2tvVKnMqBY6F8cqmG4ZUm0W9tMLKiR7JWO\nAsjSlQfEEnpOr/AmuONwTsNg+g93IILv3akCQQDnrKfmA8o0/IlS1ZfK/hcRYlZ3\nEmVoZBEciPwInkxCZ0F4Prze/l0hntYVPEeuyoO7wc4qYnaSiozJKWtXp83xAkBo\nk+ubsYv51jH6wzdkDiAlzsfSNVO/O7V/qHcNYO3o8o5W5gX1RbG8KV74rhCfmhOz\nn2nFbPLeskWZTSwOAo3BAkBWHBjvCj1sBgsIG4v6Tn2ig21akbmssJezmZRjiqeh\nqt0sAzMVixAwIFM0GsW3vQ8Hr/eBTb5EBQVZ/doRqUzf\n-----END RSA PRIVATE KEY-----\n"
      }
    },
    "log_guid": "some-guid",
    "log_source": "CELL",
    "metrics_guid": "some-guid",
    "annotation": "1435775395.194748",
    "egress_rules": [
      {
        "protocol": "all",
        "destinations": [
          "0.0.0.0-9.255.255.255"
        ],
        "log": false
      },
      {
        "protocol": "all",
        "destinations": [
          "11.0.0.0-169.253.255.255"
        ],
        "log": false
      },
      {
        "protocol": "all",
        "destinations": [
          "169.255.0.0-172.15.255.255"
        ],
        "log": false
      },
      {
        "protocol": "all",
        "destinations": [
          "172.32.0.0-192.167.255.255"
        ],
        "log": false
      },
      {
        "protocol": "all",
        "destinations": [
          "192.169.0.0-255.255.255.255"
        ],
        "log": false
      },
      {
        "protocol": "tcp",
        "destinations": [
          "0.0.0.0/0"
        ],
        "ports": [
          53
        ],
        "log": false
      },
      {
        "protocol": "udp",
        "destinations": [
          "0.0.0.0/0"
        ],
        "ports": [
          53
        ],
        "log": false
      }
    ],
    "modification_tag": {
      "epoch": "some-guid",
      "index": 0
    },
		"placement_tags": ["red-tag", "blue-tag"],
    "trusted_system_certificates_path": "/etc/cf-system-certificates",
    "network": {
			"properties": {
				"key": "value",
				"another_key": "another_value"
			}
		},
	"max_pids": 256,
	"certificate_properties": {
		"organizational_unit": ["stuff"]
	},
	"check_definition": {
		"checks": [
			{
				"tcp_check": {
					"port": 12345,
					"connect_timeout_ms": 100
				}
			}
		],
		"readiness_checks": [
			{
				"tcp_check": {
					"port": 12345
				}
			}
		],
		"log_source": "healthcheck_log_source"
	},
	"image_layers": [
	  {
			"url": "some-url",
			"destination_path": "/tmp",
			"digest_algorithm": "SHA512",
			"digest_value": "abc123",
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
	"sidecars": [
	  {
			"action": {
				"codependent": {
					"actions": [
						{
							"run": {
								"path": "/tmp/lifecycle/launcher",
								"args": [
									"app",
									"",
									"{\"start_command\":\"/usr/local/bin/nginx\"}"
								],
								"resource_limits": {
									"nofile": 16384
								},
								"user": "vcap",
								"log_source": "SIDECAR",
								"suppress_log_output": false,
								"service_binding_files": null
							}
						}
					]
				}
			},
			"disk_mb": 512,
			"memory_mb": 512
		}
	],
	"log_rate_limit": {
	  "bytes_per_second": 2048
	},
	"service_binding_files": null
  }`

	BeforeEach(func() {
		desiredLRP = models.DesiredLRP{}
		err := json.Unmarshal([]byte(jsonDesiredLRP), &desiredLRP)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("serialization", func() {
		It("successfully round trips through json and protobuf", func() {
			jsonSerialization, err := json.Marshal(desiredLRP)
			Expect(err).NotTo(HaveOccurred())
			Expect(jsonSerialization).To(MatchJSON(jsonDesiredLRP))

			protoSerialization, err := proto.Marshal(&desiredLRP)
			Expect(err).NotTo(HaveOccurred())

			var protoDeserialization models.DesiredLRP
			err = proto.Unmarshal(protoSerialization, &protoDeserialization)
			Expect(err).NotTo(HaveOccurred())

			desiredRoutes := *desiredLRP.Routes
			deserializedRoutes := *protoDeserialization.Routes

			Expect(deserializedRoutes).To(HaveLen(len(desiredRoutes)))
			for k := range desiredRoutes {
				Expect(string(*deserializedRoutes[k])).To(MatchJSON(string(*desiredRoutes[k])))
			}

			desiredLRP.Routes = nil
			protoDeserialization.Routes = nil
			Expect(protoDeserialization).To(Equal(desiredLRP))
		})
	})

	Describe("ApplyUpdate", func() {
		It("updates instances", func() {
			instances := int32(100)
			update := &models.DesiredLRPUpdate{}
			update.SetInstances(instances)
			schedulingInfo := desiredLRP.DesiredLRPSchedulingInfo()

			expectedSchedulingInfo := schedulingInfo
			expectedSchedulingInfo.Instances = instances
			expectedSchedulingInfo.ModificationTag.Increment()

			schedulingInfo.ApplyUpdate(update)
			Expect(schedulingInfo).To(Equal(expectedSchedulingInfo))
		})

		It("allows empty routes to be set", func() {
			update := &models.DesiredLRPUpdate{
				Routes: &models.Routes{},
			}

			schedulingInfo := desiredLRP.DesiredLRPSchedulingInfo()

			expectedSchedulingInfo := schedulingInfo
			expectedSchedulingInfo.Routes = models.Routes{}
			expectedSchedulingInfo.ModificationTag.Increment()

			schedulingInfo.ApplyUpdate(update)
			Expect(schedulingInfo).To(Equal(expectedSchedulingInfo))
		})

		It("allows annotation to be set", func() {
			annotation := "new-annotation"
			update := &models.DesiredLRPUpdate{}
			update.SetAnnotation(annotation)

			schedulingInfo := desiredLRP.DesiredLRPSchedulingInfo()

			expectedSchedulingInfo := schedulingInfo
			expectedSchedulingInfo.Annotation = annotation
			expectedSchedulingInfo.ModificationTag.Increment()

			schedulingInfo.ApplyUpdate(update)
			Expect(schedulingInfo).To(Equal(expectedSchedulingInfo))
		})

		It("allows empty annotation to be set", func() {
			emptyAnnotation := ""
			update := &models.DesiredLRPUpdate{}
			update.SetAnnotation(emptyAnnotation)

			schedulingInfo := desiredLRP.DesiredLRPSchedulingInfo()

			expectedSchedulingInfo := schedulingInfo
			expectedSchedulingInfo.Annotation = emptyAnnotation
			expectedSchedulingInfo.ModificationTag.Increment()

			schedulingInfo.ApplyUpdate(update)
			Expect(schedulingInfo).To(Equal(expectedSchedulingInfo))
		})

		It("updates routes", func() {
			rawMessage := json.RawMessage([]byte(`{"port": 8080,"hosts":["new-route-1","new-route-2"]}`))
			update := &models.DesiredLRPUpdate{
				Routes: &models.Routes{
					"router": &rawMessage,
				},
			}

			schedulingInfo := desiredLRP.DesiredLRPSchedulingInfo()

			expectedSchedulingInfo := schedulingInfo
			expectedSchedulingInfo.Routes = models.Routes{
				"router": &rawMessage,
			}
			expectedSchedulingInfo.ModificationTag.Increment()

			schedulingInfo.ApplyUpdate(update)
			Expect(schedulingInfo).To(Equal(expectedSchedulingInfo))
		})

		Describe("IsRoutesGroupUpdated", func() {
			var (
				routes *models.Routes
				update *models.DesiredLRPUpdate
			)

			BeforeEach(func() {
				rawMessage := json.RawMessage([]byte(`{"port": 8080,"hosts":["new-route-1","new-route-2"]}`))
				routes = &models.Routes{
					"router-group-1": &rawMessage,
				}
				update = &models.DesiredLRPUpdate{}
			})

			Context("when update does not contain routes", func() {
				BeforeEach(func() {
					update.SetInstances(2)
				})

				It("returns false", func() {
					Expect(update.IsRoutesGroupUpdated(routes, "router-group-1")).To(BeFalse())
				})
			})

			Context("when update contains routes", func() {
				BeforeEach(func() {
					update.Routes = routes
				})

				It("returns true when provided routes are not set", func() {
					Expect(update.IsRoutesGroupUpdated(nil, "router-group-1")).To(BeTrue())
				})

				Context("when the requested group does not exist in any of the routes", func() {
					It("returns false", func() {
						rawMessage := json.RawMessage([]byte(`{"port": 8080,"hosts":["new-route-1","new-route-2"]}`))
						providedRoutes := &models.Routes{
							"router-group-2": &rawMessage,
						}
						Expect(update.IsRoutesGroupUpdated(providedRoutes, "router-group-1")).To(BeTrue())
						Expect(update.IsRoutesGroupUpdated(providedRoutes, "router-group-2")).To(BeTrue())
						Expect(update.IsRoutesGroupUpdated(providedRoutes, "router-group-3")).To(BeTrue())
					})
				})

				Context("when the requested group exists in both routes", func() {
					It("returns true if contents are different", func() {
						rawMessage := json.RawMessage([]byte(`{"port": 8081,"hosts":["new-route-3","new-route-4"]}`))
						providedRoutes := &models.Routes{
							"router-group-1": &rawMessage,
						}
						Expect(update.IsRoutesGroupUpdated(providedRoutes, "router-group-1")).To(BeTrue())
					})

					It("returns false if contents are the same", func() {
						rawMessage := json.RawMessage([]byte(`{"port": 8080,"hosts":["new-route-1","new-route-2"]}`))
						providedRoutes := &models.Routes{
							"router-group-1": &rawMessage,
						}
						Expect(update.IsRoutesGroupUpdated(providedRoutes, "router-group-1")).To(BeFalse())
					})
				})
			})
		})
	})

	Describe("VersionDownTo", func() {
		Context("V2->V0", func() {
			var (
				downloadAction1, downloadAction2 models.DownloadAction
				setupAction                      *models.TimeoutAction
			)

			BeforeEach(func() {
				desiredLRP.ImageLayers = nil // V2 does not include ImageLayers
				desiredLRP.CachedDependencies = []*models.CachedDependency{
					{Name: "name-1", From: "from-1", To: "to-1", CacheKey: "cache-key-1", LogSource: "log-source-1"},
					{Name: "name-2", From: "from-2", To: "to-2", CacheKey: "cache-key-2", LogSource: "log-source-2"},
				}

				downloadAction1 = models.DownloadAction{
					Artifact:  "name-1",
					From:      "from-1",
					To:        "to-1",
					CacheKey:  "cache-key-1",
					User:      "some-user",
					LogSource: "log-source-1",
				}

				downloadAction2 = models.DownloadAction{
					Artifact:  "name-2",
					From:      "from-2",
					To:        "to-2",
					CacheKey:  "cache-key-2",
					User:      "some-user",
					LogSource: "log-source-2",
				}

				setupAction = models.Timeout(
					&models.DownloadAction{
						Artifact:  "name-3",
						From:      "from-3",
						To:        "to-3",
						CacheKey:  "cache-key-3",
						User:      "some-user",
						LogSource: "log-source-3",
					},
					20*time.Millisecond,
				)

				desiredLRP.Action = models.WrapAction(models.Timeout(
					&models.RunAction{
						Path: "/the/path",
						User: "the user",
					},
					20*time.Millisecond,
				))

				desiredLRP.Monitor = models.WrapAction(models.Timeout(
					&models.RunAction{
						Path: "/the/path",
						User: "the user",
					},
					30*time.Millisecond,
				))
				desiredLRP.StartTimeoutMs = 10000
			})

			Context("when there is no existing setup action", func() {
				BeforeEach(func() {
					desiredLRP.Setup = nil
				})

				It("converts a cache dependency into download step action", func() {
					convertedLRP := desiredLRP.VersionDownTo(format.V0)
					Expect(convertedLRP.Setup.SerialAction.Actions).To(HaveLen(1))
					Expect(convertedLRP.Setup.SerialAction.Actions[0].ParallelAction.Actions).To(HaveLen(2))

					Expect(*convertedLRP.Setup.SerialAction.Actions[0].ParallelAction.Actions[0].DownloadAction).To(Equal(downloadAction1))
					Expect(*convertedLRP.Setup.SerialAction.Actions[0].ParallelAction.Actions[1].DownloadAction).To(Equal(downloadAction2))

					Expect(*convertedLRP.Setup).To(Equal(models.Action{
						SerialAction: &models.SerialAction{
							Actions: []*models.Action{
								{
									ParallelAction: &models.ParallelAction{
										Actions: []*models.Action{
											&models.Action{DownloadAction: &downloadAction1},
											&models.Action{DownloadAction: &downloadAction2},
										},
									},
								},
							},
						},
					}))
				})
			})

			It("original LRP isn't changed", func() {
				desiredLRP.VersionDownTo(format.V0)
				Expect(desiredLRP.GetAction().GetTimeoutAction().DeprecatedTimeoutNs).To(BeZero())
				Expect(desiredLRP.GetMonitor().GetTimeoutAction().DeprecatedTimeoutNs).To(BeZero())
			})

			It("converts TimeoutMs to Timeout in Nanoseconds", func() {
				convertedLRP := desiredLRP.VersionDownTo(format.V0)
				Expect(convertedLRP.GetAction().GetTimeoutAction().DeprecatedTimeoutNs).To(BeEquivalentTo(20 * time.Millisecond))
				Expect(convertedLRP.GetMonitor().GetTimeoutAction().DeprecatedTimeoutNs).To(BeEquivalentTo(30 * time.Millisecond))
			})

			It("converts StartTimeoutMs to StartTimeout in seconds", func() {
				convertedLRP := desiredLRP.VersionDownTo(format.V0)
				Expect(convertedLRP.GetDeprecatedStartTimeoutS()).To(BeEquivalentTo(10))
			})

			Context("when there is an existing setup action", func() {
				BeforeEach(func() {
					desiredLRP.Setup = models.WrapAction(setupAction)
				})

				It("leaves original LRP unchanged", func() {
					desiredLRP.CachedDependencies = nil // avoid messing up the Setup Action

					desiredLRP.VersionDownTo(format.V0)
					Expect(desiredLRP.GetSetup().GetTimeoutAction().DeprecatedTimeoutNs).To(BeZero())
				})

				It("converts TimeoutMs to Timeout in Nanoseconds", func() {
					desiredLRP.CachedDependencies = nil // avoid messing up the Setup Action

					convertedLRP := desiredLRP.VersionDownTo(format.V0)
					Expect(convertedLRP.GetSetup().GetTimeoutAction().DeprecatedTimeoutNs).To(BeEquivalentTo(20 * time.Millisecond))
				})

				It("appends the new converted step action to the front", func() {
					convertedLRP := desiredLRP.VersionDownTo(format.V0)
					Expect(convertedLRP.Setup.SerialAction.Actions).To(HaveLen(2))
					Expect(convertedLRP.Setup.SerialAction.Actions[0].ParallelAction.Actions).To(HaveLen(2))

					Expect(*convertedLRP.Setup).To(DeepEqual(models.Action{
						SerialAction: &models.SerialAction{
							Actions: []*models.Action{
								{
									ParallelAction: &models.ParallelAction{
										Actions: []*models.Action{
											models.WrapAction(&downloadAction1),
											models.WrapAction(&downloadAction2),
										},
									},
								},
								desiredLRP.Setup.SetDeprecatedTimeoutNs(),
							},
						},
					}))
				})
			})

			Context("when there are no cache dependencies", func() {
				BeforeEach(func() {
					desiredLRP.CachedDependencies = nil
				})

				It("keeps the current setup", func() {
					convertedLRP := desiredLRP.VersionDownTo(format.V0)
					Expect(convertedLRP.Setup.SerialAction.Actions).To(HaveLen(2))

					Expect(*convertedLRP.Setup).To(Equal(*desiredLRP.Setup))
				})
			})
		})

		Context("V3->V0", func() {
			Context("when there are image layers and cached dependencies", func() {
				BeforeEach(func() {
					desiredLRP.ImageLayers = []*models.ImageLayer{
						{
							Name:            "dep0",
							Url:             "u0",
							DestinationPath: "/tmp/0",
							LayerType:       models.LayerTypeExclusive,
							MediaType:       models.MediaTypeTgz,
							DigestAlgorithm: models.DigestAlgorithmSha256,
							DigestValue:     "some-sha",
						},
						{
							Name:            "dep1",
							Url:             "u1",
							DestinationPath: "/tmp/1",
							LayerType:       models.LayerTypeShared,
							MediaType:       models.MediaTypeTgz,
						},
					}
					desiredLRP.CachedDependencies = []*models.CachedDependency{
						{
							Name:      "dep2",
							From:      "u2",
							To:        "/tmp/2",
							CacheKey:  "key2",
							LogSource: "download",
						},
					}
				})

				It("converts image layers and cached dependencies to download actions", func() {
					desiredLRP.LegacyDownloadUser = "the user"
					convertedLRP := desiredLRP.VersionDownTo(format.V0)
					Expect(models.UnwrapAction(convertedLRP.Setup)).To(DeepEqual(
						models.Serial(
							models.Parallel(
								&models.DownloadAction{
									Artifact: "dep1",
									From:     "u1",
									To:       "/tmp/1",
									CacheKey: "u1",
									User:     "the user",
								},
								&models.DownloadAction{
									Artifact:  "dep2",
									From:      "u2",
									To:        "/tmp/2",
									CacheKey:  "key2",
									LogSource: "download",
									User:      "the user",
								},
							),
							models.Serial(
								models.Parallel(
									&models.DownloadAction{
										Artifact:          "dep0",
										From:              "u0",
										To:                "/tmp/0",
										CacheKey:          "sha256:some-sha",
										User:              "the user",
										ChecksumAlgorithm: "sha256",
										ChecksumValue:     "some-sha",
									},
								),
								models.UnwrapAction(desiredLRP.Setup),
							),
						),
					))
				})

				It("removes the existing image layers", func() {
					convertedLRP := desiredLRP.VersionDownTo(format.V0)
					Expect(convertedLRP.ImageLayers).To(BeNil())
				})
			})
		})

		Context("V3->V2", func() {
			Context("when there are no image layers", func() {
				BeforeEach(func() {
					desiredLRP.ImageLayers = nil
				})

				It("does not add any cached dependencies", func() {
					convertedLRP := desiredLRP.VersionDownTo(format.V2)
					Expect(convertedLRP.CachedDependencies).To(BeEmpty())
				})

				It("does not add any download actions to the Setup", func() {
					convertedLRP := desiredLRP.VersionDownTo(format.V2)
					Expect(convertedLRP.Setup).To(Equal(desiredLRP.Setup))
				})
			})

			Context("when there are shared image layers", func() {
				BeforeEach(func() {
					desiredLRP.ImageLayers = []*models.ImageLayer{
						{
							Name:            "dep0",
							Url:             "u0",
							DestinationPath: "/tmp/0",
							LayerType:       models.LayerTypeShared,
							MediaType:       models.MediaTypeTgz,
							DigestAlgorithm: models.DigestAlgorithmSha256,
							DigestValue:     "some-sha",
						},
						{
							Name:            "dep1",
							Url:             "u1",
							DestinationPath: "/tmp/1",
							LayerType:       models.LayerTypeShared,
							MediaType:       models.MediaTypeTgz,
						},
					}
					desiredLRP.CachedDependencies = []*models.CachedDependency{
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
					convertedLRP := desiredLRP.VersionDownTo(format.V2)
					Expect(convertedLRP.CachedDependencies).To(DeepEqual([]*models.CachedDependency{
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

				It("removes the existing image layers", func() {
					convertedLRP := desiredLRP.VersionDownTo(format.V0)
					Expect(convertedLRP.ImageLayers).To(BeNil())
				})
			})

			Context("when there are exclusive image layers", func() {
				var (
					downloadAction1, downloadAction2 models.DownloadAction
				)

				BeforeEach(func() {
					desiredLRP.ImageLayers = []*models.ImageLayer{
						{
							Name:            "dep0",
							Url:             "u0",
							DestinationPath: "/tmp/0",
							LayerType:       models.LayerTypeExclusive,
							MediaType:       models.MediaTypeTgz,
							DigestAlgorithm: models.DigestAlgorithmSha256,
							DigestValue:     "some-sha",
						},
						{
							Name:            "dep1",
							Url:             "u1",
							DestinationPath: "/tmp/1",
							LayerType:       models.LayerTypeExclusive,
							MediaType:       models.MediaTypeTgz,
							DigestAlgorithm: models.DigestAlgorithmSha256,
							DigestValue:     "some-other-sha",
						},
					}
					desiredLRP.LegacyDownloadUser = "the user"
					desiredLRP.Action = models.WrapAction(models.Timeout(
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

				It("converts them to download actions with the correct user and prepends them to the setup action", func() {
					convertedLRP := desiredLRP.VersionDownTo(format.V2)
					Expect(models.UnwrapAction(convertedLRP.Setup)).To(DeepEqual(models.Serial(
						models.Parallel(&downloadAction1, &downloadAction2),
						models.UnwrapAction(desiredLRP.Setup),
					)))

				})

				It("sets removes the existing image layers", func() {
					convertedLRP := desiredLRP.VersionDownTo(format.V0)
					Expect(convertedLRP.ImageLayers).To(BeNil())
				})

				Context("when there is no existing setup action", func() {
					BeforeEach(func() {
						desiredLRP.Setup = nil
					})

					It("creates a setup action with exclusive layers converted to download actions", func() {
						convertedLRP := desiredLRP.VersionDownTo(format.V2)
						Expect(models.UnwrapAction(convertedLRP.Setup)).To(Equal(
							models.Parallel(&downloadAction1, &downloadAction2),
						))
					})
				})
			})
		})
	})

	Describe("PopulateMetricsGuid", func() {
		Context(`when both metric_tags["source_id"] and metrics_guid are provided`, func() {
			It("returns both of them unmodified", func() {
				// in practice they would always be equal if both of them are set
				// different values for test purposes only
				desiredLRP.MetricsGuid = "some-guid-1"
				desiredLRP.MetricTags = map[string]*models.MetricTagValue{
					"source_id": {Static: "some-guid-2"},
				}
				updatedLRP := desiredLRP.PopulateMetricsGuid()
				Expect(updatedLRP.MetricsGuid).To(Equal("some-guid-1"))
				Expect(updatedLRP.MetricTags["source_id"].Static).To(Equal("some-guid-2"))
			})
		})

		Context(`when both metric_tags["source_id"] and metrics_guid are missing`, func() {
			It("returns both of them as empty", func() {
				desiredLRP.MetricsGuid = ""
				desiredLRP.MetricTags = nil
				updatedLRP := desiredLRP.PopulateMetricsGuid()
				Expect(updatedLRP.MetricsGuid).To(Equal(""))
				Expect(updatedLRP.MetricTags).To(BeNil())
			})
		})

		Context(`when metric_tags["source_id"] is provided and metrics_guid is missing`, func() {
			It(`populates metrics_guid from metric_tags["source_id"]`, func() {
				desiredLRP.MetricsGuid = ""
				desiredLRP.MetricTags = map[string]*models.MetricTagValue{
					"source_id": {Static: "some-guid"},
				}
				updatedLRP := desiredLRP.PopulateMetricsGuid()
				Expect(updatedLRP.MetricsGuid).To(Equal("some-guid"))
				Expect(updatedLRP.MetricTags["source_id"].Static).To(Equal("some-guid"))
			})
		})

		Context(`when metric_tags["source_id"] is missing and metrics_guid is provided`, func() {
			It(`populates metric_tags["source_id"] from metrics_guid`, func() {
				desiredLRP.MetricsGuid = "some-guid"
				desiredLRP.MetricTags = nil
				updatedLRP := desiredLRP.PopulateMetricsGuid()
				Expect(updatedLRP.MetricsGuid).To(Equal("some-guid"))
				Expect(updatedLRP.MetricTags["source_id"].Static).To(Equal("some-guid"))
			})
		})
	})

	Describe("Validate", func() {
		var assertDesiredLRPValidationFailsWithMessage = func(lrp models.DesiredLRP, substring string) {
			validationErr := lrp.Validate()
			ExpectWithOffset(1, validationErr).To(HaveOccurred())
			ExpectWithOffset(1, validationErr.Error()).To(ContainSubstring(substring))
		}

		Context("process_guid only contains `A-Z`, `a-z`, `0-9`, `-`, and `_`", func() {
			validGuids := []string{"a", "A", "0", "-", "_", "-aaaa", "_-aaa", "09a87aaa-_aASKDn"}
			for _, validGuid := range validGuids {
				func(validGuid string) {
					It(fmt.Sprintf("'%s' is a valid process_guid", validGuid), func() {
						desiredLRP.ProcessGuid = validGuid
						err := desiredLRP.Validate()
						Expect(err).NotTo(HaveOccurred())
					})
				}(validGuid)
			}

			invalidGuids := []string{"", "bang!", "!!!", "\\slash", "star*", "params()", "invalid/key", "with.dots"}
			for _, invalidGuid := range invalidGuids {
				func(invalidGuid string) {
					It(fmt.Sprintf("'%s' is an invalid process_guid", invalidGuid), func() {
						desiredLRP.ProcessGuid = invalidGuid
						assertDesiredLRPValidationFailsWithMessage(desiredLRP, "process_guid")
					})
				}(invalidGuid)
			}
		})

		It("fails when service binding files exceed 1MB", func() {
			var exceedSize string
			for _ = range (1024 * 1024) + 100 {
				exceedSize += fmt.Sprintf("%s", "a")
			}

			var InvalidServiceBindingFiles = []*models.Files{{Name: "/redis/username", Value: exceedSize}}
			desiredLRP.ServiceBindingFiles = InvalidServiceBindingFiles

			assertDesiredLRPValidationFailsWithMessage(desiredLRP, "serviceBindingFiles")
		})

		It("requires a positive nonzero number of instances", func() {
			desiredLRP.Instances = -1
			assertDesiredLRPValidationFailsWithMessage(desiredLRP, "instances")

			desiredLRP.Instances = 0
			validationErr := desiredLRP.Validate()
			Expect(validationErr).NotTo(HaveOccurred())

			desiredLRP.Instances = 1
			validationErr = desiredLRP.Validate()
			Expect(validationErr).NotTo(HaveOccurred())
		})

		It("requires a domain", func() {
			desiredLRP.Domain = ""
			assertDesiredLRPValidationFailsWithMessage(desiredLRP, "domain")
		})

		It("requires a rootfs", func() {
			desiredLRP.RootFs = ""
			assertDesiredLRPValidationFailsWithMessage(desiredLRP, "rootfs")
		})

		It("requires a valid URL with a non-empty scheme for the rootfs", func() {
			desiredLRP.RootFs = ":not-a-url"
			assertDesiredLRPValidationFailsWithMessage(desiredLRP, "rootfs")
		})

		It("requires a valid absolute URL for the rootfs", func() {
			desiredLRP.RootFs = "not-an-absolute-url"
			assertDesiredLRPValidationFailsWithMessage(desiredLRP, "rootfs")
		})

		It("requires an action", func() {
			desiredLRP.Action = nil
			assertDesiredLRPValidationFailsWithMessage(desiredLRP, "action")
		})

		It("requires an action with an inner action", func() {
			desiredLRP.Action = &models.Action{}
			assertDesiredLRPValidationFailsWithMessage(desiredLRP, "action")
		})

		It("requires a valid action", func() {
			desiredLRP.Action = &models.Action{
				UploadAction: &models.UploadAction{
					From: "web_location",
				},
			}
			assertDesiredLRPValidationFailsWithMessage(desiredLRP, "to")
		})

		It("requires a valid setup action if specified", func() {
			desiredLRP.Setup = &models.Action{
				UploadAction: &models.UploadAction{
					From: "web_location",
				},
			}
			assertDesiredLRPValidationFailsWithMessage(desiredLRP, "to")
		})

		It("requires a setup action with an inner action", func() {
			desiredLRP.Setup = &models.Action{}
			assertDesiredLRPValidationFailsWithMessage(desiredLRP, "setup")
		})

		It("requires a valid monitor action if specified", func() {
			desiredLRP.Monitor = &models.Action{
				UploadAction: &models.UploadAction{
					From: "web_location",
				},
			}
			assertDesiredLRPValidationFailsWithMessage(desiredLRP, "to")
		})

		It("requires a monitor action with an inner action", func() {
			desiredLRP.Monitor = &models.Action{}
			assertDesiredLRPValidationFailsWithMessage(desiredLRP, "monitor")
		})

		It("requires a valid MemoryMb", func() {
			desiredLRP.MemoryMb = -1
			assertDesiredLRPValidationFailsWithMessage(desiredLRP, "memory_mb")
		})

		It("requires a valid DiskMb", func() {
			desiredLRP.DiskMb = -1
			assertDesiredLRPValidationFailsWithMessage(desiredLRP, "disk_mb")
		})

		It("requires a valid MaxPids", func() {
			desiredLRP.MaxPids = -1
			assertDesiredLRPValidationFailsWithMessage(desiredLRP, "max_pids")
		})

		It("limits the annotation length", func() {
			desiredLRP.Annotation = randStringBytes(50000)
			assertDesiredLRPValidationFailsWithMessage(desiredLRP, "annotation")
		})

		It("requires metric tags", func() {
			desiredLRP.MetricTags = nil
			assertDesiredLRPValidationFailsWithMessage(desiredLRP, "metric_tags")
		})

		Context("when a log rate limit is present", func() {
			It("cannot be less than -1", func() {
				desiredLRP.LogRateLimit = &models.LogRateLimit{BytesPerSecond: -2}
				assertDesiredLRPValidationFailsWithMessage(desiredLRP, "log_rate_limit")
			})

			It("allows -1 to indicate no log rate limit", func() {
				desiredLRP.LogRateLimit = &models.LogRateLimit{BytesPerSecond: -1}
				Expect(desiredLRP.Validate()).ToNot(HaveOccurred())
			})

			It("allows zero as a log rate limit", func() {
				desiredLRP.LogRateLimit = &models.LogRateLimit{BytesPerSecond: 0}
				Expect(desiredLRP.Validate()).ToNot(HaveOccurred())
			})
		})

		Context("when a log rate limit is not present", func() {
			It("is valid", func() {
				Expect(desiredLRP.Validate()).ToNot(HaveOccurred())
			})
		})

		Context("when security group is present", func() {
			It("must be valid", func() {
				desiredLRP.EgressRules = []*models.SecurityGroupRule{{
					Protocol: "foo",
				}}
				assertDesiredLRPValidationFailsWithMessage(desiredLRP, "egress_rules")
			})
		})

		Context("when security group is not present", func() {
			It("does not error", func() {
				desiredLRP.EgressRules = []*models.SecurityGroupRule{}

				validationErr := desiredLRP.Validate()
				Expect(validationErr).NotTo(HaveOccurred())
			})
		})

		Context("when sidecars are specified", func() {
			It("requires them to be valid", func() {
				desiredLRP.Sidecars = []*models.Sidecar{
					{
						Action: nil,
					},
				}
				assertDesiredLRPValidationFailsWithMessage(desiredLRP, "sidecars")
			})
		})

		Context("when cached dependencies are specified", func() {
			It("requires requires them to be valid", func() {
				desiredLRP.CachedDependencies = []*models.CachedDependency{
					{
						To:   "",
						From: "",
					},
				}
				assertDesiredLRPValidationFailsWithMessage(desiredLRP, "cached_dependency")
			})

			It("requires a valid checksum algorithm", func() {
				desiredLRP.CachedDependencies = []*models.CachedDependency{
					{
						To:                "here",
						From:              "there",
						ChecksumAlgorithm: "wrong algorithm",
						ChecksumValue:     "sum value",
					},
				}
				assertDesiredLRPValidationFailsWithMessage(desiredLRP, "invalid algorithm")
			})

			It("requires a valid checksum value", func() {
				desiredLRP.CachedDependencies = []*models.CachedDependency{
					{
						To:                "here",
						From:              "there",
						ChecksumAlgorithm: "md5",
					},
				}
				assertDesiredLRPValidationFailsWithMessage(desiredLRP, "value")
			})
		})

		Context("when image layers are specified", func() {
			It("requires requires them to be valid", func() {
				desiredLRP.ImageLayers = []*models.ImageLayer{
					{
						Url:             "",
						DestinationPath: "",
					},
				}
				assertDesiredLRPValidationFailsWithMessage(desiredLRP, "image_layer")
			})

			It("requires a valid digest value", func() {
				desiredLRP.ImageLayers = []*models.ImageLayer{
					{
						Url:             "here",
						DestinationPath: "there",
						DigestAlgorithm: models.DigestAlgorithmSha256,
					},
				}
				assertDesiredLRPValidationFailsWithMessage(desiredLRP, "value")
			})

			Context("when there are exclusive layers specified", func() {
				It("requires a legacy download user", func() {
					desiredLRP.LegacyDownloadUser = ""
					desiredLRP.ImageLayers = []*models.ImageLayer{
						{
							Url:             "here",
							DestinationPath: "there",
							DigestAlgorithm: models.DigestAlgorithmSha256,
							DigestValue:     "sum value",
							LayerType:       models.LayerTypeExclusive,
						},
					}
					assertDesiredLRPValidationFailsWithMessage(desiredLRP, "legacy_download_user")
				})
			})
		})

		Context("metric tags", func() {
			It("is invalid when both static and dynamic values are provided", func() {
				desiredLRP.MetricTags = map[string]*models.MetricTagValue{
					"some_metric": {Static: "some-value", Dynamic: models.MetricTagDynamicValueIndex},
				}
				assertDesiredLRPValidationFailsWithMessage(desiredLRP, "metric_tags")
				assertDesiredLRPValidationFailsWithMessage(desiredLRP, "static")
				assertDesiredLRPValidationFailsWithMessage(desiredLRP, "dynamic")
			})

			It("is valid when metric tags source_id matches metrics_guid", func() {
				desiredLRP.MetricsGuid = "some-guid"
				desiredLRP.MetricTags = map[string]*models.MetricTagValue{
					"source_id": {Static: "some-guid"},
				}
				Expect(desiredLRP.Validate()).To(Succeed())
			})

			It("is invalid when metric tags source_id does not match metrics_guid", func() {
				desiredLRP.MetricsGuid = "some-guid"
				desiredLRP.MetricTags = map[string]*models.MetricTagValue{
					"source_id": {Static: "some-another-guid"},
				}
				assertDesiredLRPValidationFailsWithMessage(desiredLRP, "metric_tags")
				assertDesiredLRPValidationFailsWithMessage(desiredLRP, "source_id should match metrics_guid")
			})

			It("is valid when metric tags source_id is provided and metrics_guid is not provided", func() {
				desiredLRP.MetricsGuid = ""
				desiredLRP.MetricTags = map[string]*models.MetricTagValue{
					"source_id": {Static: "some-other-guid"},
				}
				Expect(desiredLRP.Validate()).To(Succeed())
			})
		})

		Context("when image credentials are specified", func() {
			It("is valid when both credentials are supplied", func() {
				desiredLRP.ImageUsername = "something"
				desiredLRP.ImagePassword = "something"
				Expect(desiredLRP.Validate()).To(Succeed())
			})

			It("is valid when no credentials are supplied", func() {
				desiredLRP.ImageUsername = ""
				desiredLRP.ImagePassword = ""
				Expect(desiredLRP.Validate()).To(Succeed())
			})

			It("is invalid when providing just a username", func() {
				desiredLRP.ImageUsername = "something"
				desiredLRP.ImagePassword = ""
				assertDesiredLRPValidationFailsWithMessage(desiredLRP, "image_password")
			})

			It("is invalid when providing just a password", func() {
				desiredLRP.ImageUsername = ""
				desiredLRP.ImagePassword = "something"
				assertDesiredLRPValidationFailsWithMessage(desiredLRP, "image_username")
			})
		})
	})
})

var _ = Describe("DesiredLRPUpdate", func() {
	var desiredLRPUpdate models.DesiredLRPUpdate

	BeforeEach(func() {
		desiredLRPUpdate.SetInstances(2)
		desiredLRPUpdate.Routes = &models.Routes{
			"foo": &json.RawMessage{'"', 'b', 'a', 'r', '"'},
		}
		desiredLRPUpdate.SetAnnotation("some-text")
		desiredLRPUpdate.MetricTags = map[string]*models.MetricTagValue{
			"some-tag": {Static: "some-value"},
		}
	})

	Describe("Validate", func() {
		var assertDesiredLRPValidationFailsWithMessage = func(lrp models.DesiredLRPUpdate, substring string) {
			validationErr := lrp.Validate()
			Expect(validationErr).To(HaveOccurred())
			Expect(validationErr.Error()).To(ContainSubstring(substring))
		}

		It("requires a positive nonzero number of instances", func() {
			desiredLRPUpdate.SetInstances(-1)
			assertDesiredLRPValidationFailsWithMessage(desiredLRPUpdate, "instances")

			desiredLRPUpdate.SetInstances(0)
			validationErr := desiredLRPUpdate.Validate()
			Expect(validationErr).NotTo(HaveOccurred())

			desiredLRPUpdate.SetInstances(1)
			validationErr = desiredLRPUpdate.Validate()
			Expect(validationErr).NotTo(HaveOccurred())
		})

		It("limits the annotation length", func() {
			largeString := randStringBytes(50000)
			desiredLRPUpdate.SetAnnotation(largeString)
			assertDesiredLRPValidationFailsWithMessage(desiredLRPUpdate, "annotation")
		})

		Context("metric tags", func() {
			It("is invalid when both static and dynamic values are provided for the same key", func() {
				desiredLRPUpdate.MetricTags = map[string]*models.MetricTagValue{
					"some_metric": {Static: "some-value", Dynamic: models.MetricTagDynamicValueIndex},
				}
				assertDesiredLRPValidationFailsWithMessage(desiredLRPUpdate, "metric_tags")
				assertDesiredLRPValidationFailsWithMessage(desiredLRPUpdate, "static")
				assertDesiredLRPValidationFailsWithMessage(desiredLRPUpdate, "dynamic")
			})

			It("is valid when metric tags is empty", func() {
				desiredLRPUpdate.MetricTags = map[string]*models.MetricTagValue{}
				Expect(desiredLRPUpdate.Validate()).To(Succeed())
			})

			It("is valid when metric tags is nil", func() {
				desiredLRPUpdate.MetricTags = nil
				Expect(desiredLRPUpdate.Validate()).To(Succeed())
			})
		})
	})

	Describe("serialization", func() {
		var expectedJSON string
		BeforeEach(func() {
			expectedJSON = `{
				"instances": 2,
				"routes": {
					"foo": "bar"
				},
				"annotation": "some-text",
				"metric_tags": {
				  "some-tag": {
				    "static": "some-value"
				  }
				}
			}`
		})

		It("can marshal to JSON and back", func() {
			Expect(json.Marshal(desiredLRPUpdate)).To(MatchJSON(expectedJSON))

			var testV models.DesiredLRPUpdate
			Expect(json.Unmarshal([]byte(expectedJSON), &testV)).To(Succeed())
			Expect(testV).To(Equal(desiredLRPUpdate))
		})
	})

	Describe("IsMetricTagsUpdated", func() {
		Context("when the update does not contain metric tags", func() {
			It("returns false", func() {
				existingTags := map[string]*models.MetricTagValue{"some-tag": {Static: "some-tag-value"}}
				update := &models.DesiredLRPUpdate{}
				Expect(update.IsMetricTagsUpdated(existingTags)).To(BeFalse())
			})
		})
		Context("when the metric tags differ in a single static value", func() {
			It("returns true", func() {
				existingTags := map[string]*models.MetricTagValue{"some-tag": {Static: "some-tag-value"}}
				update := &models.DesiredLRPUpdate{MetricTags: map[string]*models.MetricTagValue{"some-tag": {Static: "some-other-tag-value"}}}
				Expect(update.IsMetricTagsUpdated(existingTags)).To(BeTrue())
			})
		})
		Context("when the metric tags differ in a single dynamic value", func() {
			It("returns true", func() {
				existingTags := map[string]*models.MetricTagValue{"some-tag": {Dynamic: models.MetricTagDynamicValueIndex}}
				update := &models.DesiredLRPUpdate{MetricTags: map[string]*models.MetricTagValue{"some-tag": {Dynamic: models.MetricTagDynamicValueInstanceGuid}}}
				Expect(update.IsMetricTagsUpdated(existingTags)).To(BeTrue())
			})
		})
		Context("when the metric tags have different keys", func() {
			It("returns true", func() {
				existingTags := map[string]*models.MetricTagValue{"some-tag": {Static: "some-tag-value"}}
				update := &models.DesiredLRPUpdate{MetricTags: map[string]*models.MetricTagValue{"some-other-tag": {Static: "some-tag-value"}}}
				Expect(update.IsMetricTagsUpdated(existingTags)).To(BeTrue())
			})
		})
		Context("when the metric tags are different lengths", func() {
			It("returns true", func() {
				existingTags := map[string]*models.MetricTagValue{}
				update := &models.DesiredLRPUpdate{MetricTags: map[string]*models.MetricTagValue{"some-tag": {Static: "some-tag-value"}}}
				Expect(update.IsMetricTagsUpdated(existingTags)).To(BeTrue())
			})
		})
		Context("when the metric tags are equal", func() {
			It("returns false", func() {
				existingTags := map[string]*models.MetricTagValue{"some-tag": {Static: "some-tag-value"}}
				update := &models.DesiredLRPUpdate{MetricTags: map[string]*models.MetricTagValue{"some-tag": {Static: "some-tag-value"}}}
				Expect(update.IsMetricTagsUpdated(existingTags)).To(BeFalse())
			})
		})
	})
})

func randStringBytes(n int) string {
	rb := make([]byte, n)
	rand.Read(rb)
	rs := base64.URLEncoding.EncodeToString(rb)
	return rs
}

var _ = Describe("DesiredLRPKey", func() {
	const guid = "valid-guid"
	const domain = "valid-domain"
	const log = "valid-log-guid"

	DescribeTable("Validation",
		func(key models.DesiredLRPKey, expectedErr string) {
			err := key.Validate()
			if expectedErr == "" {
				Expect(err).NotTo(HaveOccurred())
			} else {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(expectedErr))
			}
		},
		Entry("valid key", models.NewDesiredLRPKey(guid, domain, log), ""),
		Entry("blank process guid", models.NewDesiredLRPKey("", domain, log), "process_guid"),
		Entry("blank domain", models.NewDesiredLRPKey(guid, "", log), "domain"),
		Entry("blank log guid is valid", models.NewDesiredLRPKey(guid, domain, ""), ""),
	)
	Context("process_guid only contains `A-Z`, `a-z`, `0-9`, `-`, and `_`", func() {
		validGuids := []string{"a", "A", "0", "-", "_", "-aaaa", "_-aaa", "09a87aaa-_aASKDn"}
		for _, validGuid := range validGuids {
			func(validGuid string) {
				It(fmt.Sprintf("'%s' is a valid process_guid", validGuid), func() {
					key := models.NewDesiredLRPKey(validGuid, domain, log)
					err := key.Validate()
					Expect(err).NotTo(HaveOccurred())
				})
			}(validGuid)
		}

		invalidGuids := []string{"", "bang!", "!!!", "\\slash", "star*", "params()", "invalid/key", "with.dots"}
		for _, invalidGuid := range invalidGuids {
			func(invalidGuid string) {
				It(fmt.Sprintf("'%s' is an invalid process_guid", invalidGuid), func() {
					key := models.NewDesiredLRPKey(invalidGuid, domain, log)
					err := key.Validate()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("process_guid"))
				})
			}(invalidGuid)
		}
	})
})

var _ = Describe("DesiredLRPResource", func() {
	const rootFs = "preloaded://linux64"
	const memoryMb = 256
	const diskMb = 256
	const maxPids = 256

	DescribeTable("Validation",
		func(key models.DesiredLRPResource, expectedErr string) {
			err := key.Validate()
			if expectedErr == "" {
				Expect(err).NotTo(HaveOccurred())
			} else {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(expectedErr))
			}
		},
		Entry("valid resource", models.NewDesiredLRPResource(memoryMb, diskMb, maxPids, rootFs), ""),
		Entry("invalid rootFs", models.NewDesiredLRPResource(memoryMb, diskMb, maxPids, "BAD URL"), "rootfs"),
		Entry("invalid memoryMb", models.NewDesiredLRPResource(-1, diskMb, maxPids, rootFs), "memory_mb"),
		Entry("invalid diskMb", models.NewDesiredLRPResource(memoryMb, -1, maxPids, rootFs), "disk_mb"),
		Entry("invalid maxPids", models.NewDesiredLRPResource(memoryMb, diskMb, -1, rootFs), "max_pids"),
	)
})

var _ = Describe("DesiredLRPSchedulingInfo", func() {
	const annotation = "the annotation"
	const instances = 2
	var (
		largeString = randStringBytes(50000)
		rawMessage  = json.RawMessage([]byte(`{"port": 8080,"hosts":["new-route-1","new-route-2"]}`))
		routes      = models.Routes{
			"router": &rawMessage,
		}
		largeRoutingString = randStringBytes(129 * 1024)
		largeRoute         = json.RawMessage([]byte(largeRoutingString))
		largeRoutes        = models.Routes{
			"router": &largeRoute,
		}
		tag = models.ModificationTag{}
	)

	DescribeTable("Validation",
		func(key models.DesiredLRPSchedulingInfo, expectedErr string) {
			err := key.Validate()
			if expectedErr == "" {
				Expect(err).NotTo(HaveOccurred())
			} else {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(expectedErr))
			}
		},
		Entry("valid scheduling info", models.NewDesiredLRPSchedulingInfo(newValidLRPKey(), annotation, instances, newValidResource(), routes, tag, nil, nil), ""),
		Entry("invalid annotation", models.NewDesiredLRPSchedulingInfo(newValidLRPKey(), largeString, instances, newValidResource(), routes, tag, nil, nil), "annotation"),
		Entry("invalid instances", models.NewDesiredLRPSchedulingInfo(newValidLRPKey(), annotation, -2, newValidResource(), routes, tag, nil, nil), "instances"),
		Entry("invalid key", models.NewDesiredLRPSchedulingInfo(models.DesiredLRPKey{}, annotation, instances, newValidResource(), routes, tag, nil, nil), "process_guid"),
		Entry("invalid resource", models.NewDesiredLRPSchedulingInfo(newValidLRPKey(), annotation, instances, models.DesiredLRPResource{}, routes, tag, nil, nil), "rootfs"),
		Entry("invalid routes", models.NewDesiredLRPSchedulingInfo(newValidLRPKey(), annotation, instances, newValidResource(), largeRoutes, tag, nil, nil), "routes"),
	)
})

var _ = Describe("DesiredLRPRoutingInfo", func() {
	const instances = 2
	var (
		rawMessage = json.RawMessage([]byte(`{"port": 8080,"hosts":["new-route-1","new-route-2"]}`))
		routes     = models.Routes{
			"router": &rawMessage,
		}
		largeRoutingString = randStringBytes(129 * 1024)
		largeRoute         = json.RawMessage([]byte(largeRoutingString))
		largeRoutes        = models.Routes{
			"router": &largeRoute,
		}
		tag = models.ModificationTag{}
	)

	DescribeTable("Validation",
		func(key models.DesiredLRP, expectedErr string) {
			err := key.Validate()
			if expectedErr == "" {
				Expect(err).NotTo(HaveOccurred())
			} else {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(expectedErr))
			}
		},
		Entry("invalid instances", models.NewDesiredLRPRoutingInfo(newValidLRPKey(), -2, &routes, &tag, map[string]*models.MetricTagValue{}), "instances"),
		Entry("invalid key", models.NewDesiredLRPRoutingInfo(models.DesiredLRPKey{}, instances, &routes, &tag, map[string]*models.MetricTagValue{}), "process_guid"),
		Entry("invalid routes", models.NewDesiredLRPRoutingInfo(newValidLRPKey(), instances, &largeRoutes, &tag, map[string]*models.MetricTagValue{}), "routes"),
		Entry("invalid metricTags", models.NewDesiredLRPRoutingInfo(newValidLRPKey(), instances, &routes, &tag, map[string]*models.MetricTagValue{"foo": {Dynamic: models.DynamicValueInvalid}}), "metric_tags"),
	)
})

var _ = Describe("DesiredLRPRunInfo", func() {
	var envVars = []models.EnvironmentVariable{{"FOO", "bar"}}
	var action = model_helpers.NewValidAction()
	const startTimeoutMs int64 = 12
	const privileged = true
	var ports = []uint32{80, 443}
	var egressRules = model_helpers.NewValidEgressRules()
	const logSource = "log-source"
	const metricsGuid = "metrics-guid"
	const cpuWeight = 50
	var createdAt = time.Unix(123, 456)
	var trustedSystemCertificatesPath = "/etc/cf-system-certificates"
	var httpCheckDef = model_helpers.NewValidHTTPCheckDefinition()
	var logRateLimit = &models.LogRateLimit{BytesPerSecond: 1024}

	var serviceBindingFiles = []*models.Files{{Name: "/redis/username", Value: "username"}}

	var exceedSize string
	for _ = range (1024 * 1024) + 100 {
		exceedSize += fmt.Sprintf("%s", "a")
	}

	var InvalidServiceBindingFiles = []*models.Files{{Name: "/redis/username", Value: exceedSize}}

	DescribeTable("Validation",
		func(key models.DesiredLRPRunInfo, expectedErr string) {
			err := key.Validate()
			if expectedErr == "" {
				Expect(err).NotTo(HaveOccurred())
			} else {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(expectedErr))
			}
		},
		Entry("valid run info", models.NewDesiredLRPRunInfo(newValidLRPKey(), createdAt, envVars, nil, action, action, action, startTimeoutMs, privileged, cpuWeight, ports, egressRules, logSource, metricsGuid, "legacy-jim", trustedSystemCertificatesPath, []*models.VolumeMount{}, nil, nil, "", "", httpCheckDef, nil, map[string]*models.MetricTagValue{}, []*models.Sidecar{{Action: action}}, logRateLimit, serviceBindingFiles), ""),
		Entry("invalid key", models.NewDesiredLRPRunInfo(models.DesiredLRPKey{}, createdAt, envVars, nil, action, action, action, startTimeoutMs, privileged, cpuWeight, ports, egressRules, logSource, metricsGuid, "legacy-jim", trustedSystemCertificatesPath, []*models.VolumeMount{}, nil, nil, "", "", httpCheckDef, nil, map[string]*models.MetricTagValue{}, []*models.Sidecar{}, logRateLimit, serviceBindingFiles), "process_guid"),
		Entry("invalid env vars", models.NewDesiredLRPRunInfo(newValidLRPKey(), createdAt, append(envVars, models.EnvironmentVariable{}), nil, action, action, action, startTimeoutMs, privileged, cpuWeight, ports, egressRules, logSource, metricsGuid, "legacy-jim", trustedSystemCertificatesPath, []*models.VolumeMount{}, nil, nil, "", "", httpCheckDef, nil, map[string]*models.MetricTagValue{}, []*models.Sidecar{}, logRateLimit, serviceBindingFiles), "name"),
		Entry("invalid setup action", models.NewDesiredLRPRunInfo(newValidLRPKey(), createdAt, envVars, nil, &models.Action{}, action, action, startTimeoutMs, privileged, cpuWeight, ports, egressRules, logSource, metricsGuid, "legacy-jim", trustedSystemCertificatesPath, []*models.VolumeMount{}, nil, nil, "", "", httpCheckDef, nil, map[string]*models.MetricTagValue{}, []*models.Sidecar{}, logRateLimit, serviceBindingFiles), "inner-action"),
		Entry("invalid run action", models.NewDesiredLRPRunInfo(newValidLRPKey(), createdAt, envVars, nil, action, &models.Action{}, action, startTimeoutMs, privileged, cpuWeight, ports, egressRules, logSource, metricsGuid, "legacy-jim", trustedSystemCertificatesPath, []*models.VolumeMount{}, nil, nil, "", "", httpCheckDef, nil, map[string]*models.MetricTagValue{}, []*models.Sidecar{}, logRateLimit, serviceBindingFiles), "inner-action"),
		Entry("invalid monitor action", models.NewDesiredLRPRunInfo(newValidLRPKey(), createdAt, envVars, nil, action, action, &models.Action{}, startTimeoutMs, privileged, cpuWeight, ports, egressRules, logSource, metricsGuid, "legacy-jim", trustedSystemCertificatesPath, []*models.VolumeMount{}, nil, nil, "", "", httpCheckDef, nil, map[string]*models.MetricTagValue{}, []*models.Sidecar{}, logRateLimit, serviceBindingFiles), "inner-action"),
		Entry("invalid http check definition", models.NewDesiredLRPRunInfo(newValidLRPKey(), createdAt, envVars, nil, action, action, action, startTimeoutMs, privileged, cpuWeight, ports, egressRules, logSource, metricsGuid, "legacy-jim", trustedSystemCertificatesPath, []*models.VolumeMount{}, nil, nil, "", "", &models.CheckDefinition{[]*models.Check{&models.Check{HttpCheck: &models.HTTPCheck{Port: 65536}}}, "healthcheck_log_source", []*models.Check{&models.Check{HttpCheck: &models.HTTPCheck{Port: 77777}}}}, nil, map[string]*models.MetricTagValue{}, []*models.Sidecar{}, logRateLimit, serviceBindingFiles), "port"),
		Entry("invalid tcp check definition", models.NewDesiredLRPRunInfo(newValidLRPKey(), createdAt, envVars, nil, action, action, action, startTimeoutMs, privileged, cpuWeight, ports, egressRules, logSource, metricsGuid, "legacy-jim", trustedSystemCertificatesPath, []*models.VolumeMount{}, nil, nil, "", "", &models.CheckDefinition{[]*models.Check{&models.Check{TcpCheck: &models.TCPCheck{}}}, "healthcheck_log_source", []*models.Check{&models.Check{TcpCheck: &models.TCPCheck{}}}}, nil, map[string]*models.MetricTagValue{}, []*models.Sidecar{}, logRateLimit, serviceBindingFiles), "port"),
		Entry("invalid check in check definition", models.NewDesiredLRPRunInfo(newValidLRPKey(), createdAt, envVars, nil, action, action, action, startTimeoutMs, privileged, cpuWeight, ports, egressRules, logSource, metricsGuid, "legacy-jim", trustedSystemCertificatesPath, []*models.VolumeMount{}, nil, nil, "", "", &models.CheckDefinition{[]*models.Check{&models.Check{HttpCheck: &models.HTTPCheck{}, TcpCheck: &models.TCPCheck{}}}, "healthcheck_log_source", []*models.Check{&models.Check{HttpCheck: &models.HTTPCheck{}, TcpCheck: &models.TCPCheck{}}}}, nil, map[string]*models.MetricTagValue{}, []*models.Sidecar{}, logRateLimit, serviceBindingFiles), "check"),
		Entry("invalid legacy download user", models.NewDesiredLRPRunInfo(newValidLRPKey(), createdAt, envVars, nil, action, action, action, startTimeoutMs, privileged, cpuWeight, ports, egressRules, logSource, metricsGuid, "", trustedSystemCertificatesPath, []*models.VolumeMount{}, nil, nil, "", "", httpCheckDef, []*models.ImageLayer{{Url: "url", DestinationPath: "path", MediaType: models.MediaTypeTgz, LayerType: models.LayerTypeExclusive}}, map[string]*models.MetricTagValue{}, []*models.Sidecar{}, logRateLimit, serviceBindingFiles), "legacy_download_user"),
		Entry("invalid cached dependency", models.NewDesiredLRPRunInfo(newValidLRPKey(), createdAt, envVars, []*models.CachedDependency{{To: "here"}}, action, action, action, startTimeoutMs, privileged, cpuWeight, ports, egressRules, logSource, metricsGuid, "user", trustedSystemCertificatesPath, []*models.VolumeMount{}, nil, nil, "", "", httpCheckDef, nil, map[string]*models.MetricTagValue{}, []*models.Sidecar{}, logRateLimit, serviceBindingFiles), "cached_dependency"),
		Entry("invalid volume mount", models.NewDesiredLRPRunInfo(newValidLRPKey(), createdAt, envVars, nil, action, action, action, startTimeoutMs, privileged, cpuWeight, ports, egressRules, logSource, metricsGuid, "user", trustedSystemCertificatesPath, []*models.VolumeMount{{Mode: "lol"}}, nil, nil, "", "", httpCheckDef, nil, map[string]*models.MetricTagValue{}, []*models.Sidecar{}, logRateLimit, serviceBindingFiles), "volume_mount"),
		Entry("invalid image username", models.NewDesiredLRPRunInfo(newValidLRPKey(), createdAt, envVars, nil, action, action, action, startTimeoutMs, privileged, cpuWeight, ports, egressRules, logSource, metricsGuid, "user", trustedSystemCertificatesPath, []*models.VolumeMount{}, nil, nil, "", "password", httpCheckDef, nil, map[string]*models.MetricTagValue{}, []*models.Sidecar{}, logRateLimit, serviceBindingFiles), "image_username"),
		Entry("invalid image password", models.NewDesiredLRPRunInfo(newValidLRPKey(), createdAt, envVars, nil, action, action, action, startTimeoutMs, privileged, cpuWeight, ports, egressRules, logSource, metricsGuid, "user", trustedSystemCertificatesPath, []*models.VolumeMount{}, nil, nil, "username", "", httpCheckDef, nil, map[string]*models.MetricTagValue{}, []*models.Sidecar{}, logRateLimit, serviceBindingFiles), "image_password"),
		Entry("invalid layers", models.NewDesiredLRPRunInfo(newValidLRPKey(), createdAt, envVars, nil, action, action, action, startTimeoutMs, privileged, cpuWeight, ports, egressRules, logSource, metricsGuid, "user", trustedSystemCertificatesPath, []*models.VolumeMount{}, nil, nil, "", "", httpCheckDef, []*models.ImageLayer{{Url: "some-url"}}, map[string]*models.MetricTagValue{}, []*models.Sidecar{}, logRateLimit, serviceBindingFiles), "image_layer"),
		Entry("invalid metric tags", models.NewDesiredLRPRunInfo(newValidLRPKey(), createdAt, envVars, nil, action, action, action, startTimeoutMs, privileged, cpuWeight, ports, egressRules, logSource, metricsGuid, "user", trustedSystemCertificatesPath, []*models.VolumeMount{}, nil, nil, "", "", httpCheckDef, nil, map[string]*models.MetricTagValue{"foo": {Dynamic: models.DynamicValueInvalid}}, []*models.Sidecar{}, logRateLimit, serviceBindingFiles), "metric_tags"),
		Entry("invalid sidecars", models.NewDesiredLRPRunInfo(newValidLRPKey(), createdAt, envVars, nil, action, action, action, startTimeoutMs, privileged, cpuWeight, ports, egressRules, logSource, metricsGuid, "user", trustedSystemCertificatesPath, []*models.VolumeMount{}, nil, nil, "", "", httpCheckDef, nil, map[string]*models.MetricTagValue{}, []*models.Sidecar{{DiskMb: -1}}, logRateLimit, serviceBindingFiles), "sidecars"),
		Entry("invalid log rate limit", models.NewDesiredLRPRunInfo(newValidLRPKey(), createdAt, envVars, nil, action, action, action, startTimeoutMs, privileged, cpuWeight, ports, egressRules, logSource, metricsGuid, "user", trustedSystemCertificatesPath, []*models.VolumeMount{}, nil, nil, "", "", httpCheckDef, nil, map[string]*models.MetricTagValue{}, []*models.Sidecar{{DiskMb: -1}}, &models.LogRateLimit{BytesPerSecond: -2}, serviceBindingFiles), "log_rate_limit"),
		Entry("invalid service binding file size", models.NewDesiredLRPRunInfo(newValidLRPKey(), createdAt, envVars, nil, action, action, action, startTimeoutMs, privileged, cpuWeight, ports, egressRules, logSource, metricsGuid, "user", trustedSystemCertificatesPath, []*models.VolumeMount{}, nil, nil, "", "", httpCheckDef, nil, map[string]*models.MetricTagValue{}, []*models.Sidecar{}, logRateLimit, InvalidServiceBindingFiles), "serviceBindingFiles"),
	)
})

func newValidLRPKey() models.DesiredLRPKey {
	return models.NewDesiredLRPKey("some-guid", "domain", "log-guid")
}

func newValidResource() models.DesiredLRPResource {
	return models.NewDesiredLRPResource(256, 256, 256, "preloaded://linux64")
}
