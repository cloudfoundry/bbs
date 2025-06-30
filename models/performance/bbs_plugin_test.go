package performance

import (
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gmeasure "github.com/onsi/gomega/gmeasure"
	"google.golang.org/protobuf/proto"

	"code.cloudfoundry.org/bbs/models"
)

var skipPerformanceTests = true

var _ = Describe("BBS Plugin Conversion: Desired LRP", func() {
	var nanosecondPrecision = gmeasure.Precision(1 * time.Nanosecond)
	var desiredLRP models.DesiredLRP

	desiredLRPjson := `{
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
			  "volume_mounted_files": null
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
			  "volume_mounted_files": null
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
			"volume_mounted_files": null
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
								"volume_mounted_files": null
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
	"volume_mounted_files": null
  }`

	Context("Conversion Performance Testing", func() {
		Context("Starting with a BBS Plugin Struct", func() {
			BeforeEach(func() {
				if skipPerformanceTests {
					Skip("Skipping Performnace Tests")
				}
				desiredLRP = models.DesiredLRP{}
				err := json.Unmarshal([]byte(desiredLRPjson), &desiredLRP)
				Expect(err).NotTo(HaveOccurred())
			})

			DescribeTable("Measures conversion time",
				func(experimentName string, number int, callback func(models.DesiredLRP)) {
					experiment := gmeasure.NewExperiment(experimentName)
					AddReportEntry(experiment.Name, experiment)
					experiment.Sample(func(i int) {
						stopwatch := experiment.NewStopwatch()
						callback(desiredLRP)
						stopwatch.Record("convert", nanosecondPrecision)
					}, gmeasure.SamplingConfig{N: number})
				},
				// bbs struct to protobuf struct
				Entry("1 conversion", "BBS Struct => Protobuf Struct", 1, bbsStructToProtobufStruct),
				Entry("1,000 conversions", "BBS Struct => Protobuf Struct", 1000, bbsStructToProtobufStruct),
				Entry("1,000,000 conversions", "BBS Struct => Protobuf Struct", 1000000, bbsStructToProtobufStruct),
				// bbs struct to protobuf binary
				Entry("1 conversion", "BBS Struct => Protobuf Binary", 1, bbsStructToProtobufBinary),
				Entry("1,000 conversions", "BBS Struct => Protobuf Binary", 1000, bbsStructToProtobufBinary),
				Entry("1,000,000 conversions", "BBS Struct => Protobuf Binary", 1000000, bbsStructToProtobufBinary),
				// bbs struct to protobuf binary to bbs struct
				Entry("1 conversion", "BBS Struct => Protobuf Binary => BBS Struct", 1, bbsStructRoundtrip),
				Entry("1,000 conversions", "BBS Struct => Protobuf Binary => BBS Struct", 1000, bbsStructRoundtrip),
				Entry("1,000,000 conversions", "BBS Struct => Protobuf Binary => BBS Struct", 1000000, bbsStructRoundtrip),
			)
		})

		Context("Starting with a Protobuf Struct", func() {
			var protoDesiredLRP *models.ProtoDesiredLRP
			BeforeEach(func() {
				if skipPerformanceTests {
					Skip("Skipping Performnace Tests")
				}
				desiredLRP = models.DesiredLRP{}
				err := json.Unmarshal([]byte(desiredLRPjson), &desiredLRP)
				protoDesiredLRP = desiredLRP.ToProto()
				Expect(err).NotTo(HaveOccurred())
			})

			DescribeTable("Measures conversion time",
				func(experimentName string, number int, callback func(*models.ProtoDesiredLRP)) {
					experiment := gmeasure.NewExperiment(experimentName)
					AddReportEntry(experiment.Name, experiment)
					experiment.Sample(func(i int) {
						stopwatch := experiment.NewStopwatch()
						callback(protoDesiredLRP)
						stopwatch.Record("convert", nanosecondPrecision)
					}, gmeasure.SamplingConfig{N: number})
				},
				// protobuf struct to bbs struct
				Entry("1 conversion", "Protobuf Struct => BBS Struct", 1, protobufStructToBbsStruct),
				Entry("1,000 conversions", "Protobuf Struct => BBS Struct", 1000, protobufStructToBbsStruct),
				Entry("1,000,000 conversions", "Protobuf Struct => BBS Struct", 1000000, protobufStructToBbsStruct),
				// protobuf struct to protobuf binary
				Entry("1 conversion", "Protobuf Struct => Protobuf Binary", 1, protobufStructToProtobufBinary),
				Entry("1,000 conversions", "Protobuf Struct => Protobuf Binary", 1000, protobufStructToProtobufBinary),
				Entry("1,000,000 conversions", "Protobuf Struct => Protobuf Binary", 1000000, protobufStructToProtobufBinary),
				// protobuf struct to protobuf binary to protobuf struct
				Entry("1 conversion", "Protobuf Struct => Protobuf Binary", 1, protobufStructRoundtrip),
				Entry("1,000 conversions", "Protobuf Struct => Protobuf Binary", 1000, protobufStructRoundtrip),
				Entry("1,000,000 conversions", "Protobuf Struct => Protobuf Binary", 1000000, protobufStructRoundtrip),
			)
		})

		Context("Starting with a Protobuf Binary", func() {
			var binaryDesiredLRP []byte
			var protoDesiredLRP *models.ProtoDesiredLRP
			BeforeEach(func() {
				if skipPerformanceTests {
					Skip("Skipping Performnace Tests")
				}
				desiredLRP = models.DesiredLRP{}
				err := json.Unmarshal([]byte(desiredLRPjson), &desiredLRP)
				protoDesiredLRP = desiredLRP.ToProto()
				Expect(err).NotTo(HaveOccurred())
				binaryDesiredLRP, _ = proto.Marshal(protoDesiredLRP)
			})

			DescribeTable("Measures conversion time",
				func(experimentName string, number int, callback func([]byte)) {
					experiment := gmeasure.NewExperiment(experimentName)
					AddReportEntry(experiment.Name, experiment)
					experiment.Sample(func(i int) {
						stopwatch := experiment.NewStopwatch()
						callback(binaryDesiredLRP)
						stopwatch.Record("convert", nanosecondPrecision)
					}, gmeasure.SamplingConfig{N: number})
				},
				// protobuf binary to bbs struct
				Entry("1 conversion", "Protobuf Binary => BBS Struct", 1, protobufBinaryToBbsStruct),
				Entry("1,000 conversions", "Protobuf Binary => BBS Struct", 1000, protobufBinaryToBbsStruct),
				Entry("1,000,000 conversions", "Protobuf Binary => BBS Struct", 1000000, protobufBinaryToBbsStruct),
				// protobuf binary to protobuf struct
				Entry("1 conversion", "Protobuf Binary => Protobuf Struct", 1, protobufBinaryToProtobufStruct),
				Entry("1,000 conversions", "Protobuf Binary => Protobuf Struct", 1000, protobufBinaryToProtobufStruct),
				Entry("1,000,000 conversions", "Protobuf Binary => Protobuf Struct", 1000000, protobufBinaryToProtobufStruct),
			)
		})
	})
})

func bbsStructToProtobufStruct(which models.DesiredLRP) {
	// simple conversion between structs
	which.ToProto()
}

func bbsStructToProtobufBinary(which models.DesiredLRP) {
	// simple conversion between structs
	protoStruct := which.ToProto()
	// marshal into protobuf binary / wire format
	proto.Marshal(protoStruct)
}

func protobufBinaryToBbsStruct(which []byte) {
	var unmarshalProto models.ProtoDesiredLRP
	// umarshal into protobuf binary / wire format
	proto.Unmarshal(which, &unmarshalProto)
	// single conversion to bbs struct
	unmarshalProto.FromProto()
}

func bbsStructRoundtrip(which models.DesiredLRP) {
	var unmarshalProto models.ProtoDesiredLRP
	// simple conversion to proto struct
	protoStruct := which.ToProto()
	// marshal into protobuf binary / wire format
	marshalProto, _ := proto.Marshal(protoStruct)
	// unmarshal from protobuf binary / wire format
	proto.Unmarshal(marshalProto, &unmarshalProto)
	// simple conversion to bbs struct
	unmarshalProto.FromProto()
}

func protobufStructToBbsStruct(which *models.ProtoDesiredLRP) {
	// simple conversion to bbs struct
	which.FromProto()
}

func protobufStructToProtobufBinary(which *models.ProtoDesiredLRP) {
	// marshal into protobuf binary / wire format
	proto.Marshal(which)
}

func protobufBinaryToProtobufStruct(which []byte) {
	var unmarshalProto models.ProtoDesiredLRP
	// umarshal into protobuf binary / wire format
	proto.Unmarshal(which, &unmarshalProto)
}

func protobufStructRoundtrip(which *models.ProtoDesiredLRP) {
	var unmarshalProto models.ProtoDesiredLRP
	// marshal into protobuf binary / wire format
	marshalProto, _ := proto.Marshal(which)
	// unmarshal from protobuf binary / wire format
	proto.Unmarshal(marshalProto, &unmarshalProto)
}
