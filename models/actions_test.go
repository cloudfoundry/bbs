package models_test

import (
	"fmt"
	"log"
	"time"

	"code.cloudfoundry.org/bbs/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/encoding/protojson"
)

var _ = Describe("Actions", func() {
	itSerializes := func(actionPayload string, a *models.Action) {
		action := models.UnwrapAction(a)
		It("Action -> JSON for "+string(action.ActionType()), func() {
			marshalJson, err := protojson.Marshal(a.GetProto())
			Expect(err).NotTo(HaveOccurred())
			Expect(string(marshalJson)).To(MatchJSON(actionPayload))
		})
	}

	itDeserializes := func(actionPayload string, a *models.Action) {
		action := models.UnwrapAction(a)
		It("JSON -> Action for "+string(action.ActionType()), func() {
			wrappedJSON := fmt.Sprintf(`{"%s":%s}`, action.ActionType(), actionPayload)
			marshalledAction := new(models.ProtoAction)
			err := protojson.Unmarshal([]byte(wrappedJSON), marshalledAction)
			Expect(err).NotTo(HaveOccurred())
			Expect(marshalledAction.FromProto()).To(BeEquivalentTo(a))
		})
	}

	itSerializesAndDeserializes := func(actionPayload string, action *models.Action) {
		itSerializes(actionPayload, action)
		itDeserializes(actionPayload, action)
	}

	Describe("WrapAction", func() {
		It("wraps an action into *Action", func() {
			action := &models.DownloadAction{
				Artifact: "mouse",
				From:     "web_location",
				To:       "local_location",
				CacheKey: "elephant",
				User:     "someone",
			}
			wrapped := models.WrapAction(action)
			Expect(wrapped).NotTo(BeNil())
			Expect(wrapped.GetValue()).To(Equal(action))
		})

		It("does not wrap nil", func() {
			wrapped := models.WrapAction(nil)
			Expect(wrapped).To(BeNil())
		})
	})

	Describe("Nil Actions", func() {
		It("Action -> JSON for a Nil action", func() {
			var action *models.ProtoAction = nil
			By("marshalling to JSON", func() {
				json, err := protojson.Marshal(action)
				Expect(err).NotTo(HaveOccurred())
				log.Printf("%+v\n", json)
				Expect(json).To(MatchJSON("{}"))
			})
		})

		It("JSON -> Action for Nil action", func() {
			By("unwrapping", func() {
				unmarshalledAction := new(models.ProtoAction)
				err := protojson.Unmarshal([]byte("{}"), unmarshalledAction)
				Expect(err).NotTo(HaveOccurred())
				// TODO: protojson.Unmarshal requires a non-nil pointer in order to work
				// see if we can find an alternative way of doing this so the following line works
				// Expect(unmarshalledAction).To(BeNil())
				Expect(unmarshalledAction.DownloadAction).To(BeNil())
				Expect(unmarshalledAction.UploadAction).To(BeNil())
				Expect(unmarshalledAction.RunAction).To(BeNil())
				Expect(unmarshalledAction.TimeoutAction).To(BeNil())
				Expect(unmarshalledAction.EmitProgressAction).To(BeNil())
				Expect(unmarshalledAction.TryAction).To(BeNil())
				Expect(unmarshalledAction.ParallelAction).To(BeNil())
				Expect(unmarshalledAction.SerialAction).To(BeNil())
				Expect(unmarshalledAction.CodependentAction).To(BeNil())
			})
		})

		Describe("Validate", func() {
			var action *models.Action

			Context("when the action has no inner actions", func() {
				It("is valid", func() {
					action = nil

					err := action.Validate()
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})

	Describe("Download", func() {
		var downloadAction *models.DownloadAction

		Context("with checksum algorithm and value missing", func() {
			itSerializesAndDeserializes(
				`{
					"artifact": "mouse",
					"from": "web_location",
					"to": "local_location",
					"cache_key": "elephant",
					"user": "someone"
			}`,
				models.WrapAction(&models.DownloadAction{
					Artifact: "mouse",
					From:     "web_location",
					To:       "local_location",
					CacheKey: "elephant",
					User:     "someone",
				}),
			)

			Describe("Validate", func() {

				Context("when the action has 'from', 'to', and 'user' specified", func() {
					It("is valid", func() {
						downloadAction = &models.DownloadAction{
							From: "web_location",
							To:   "local_location",
							User: "someone",
						}

						err := downloadAction.Validate()
						Expect(err).NotTo(HaveOccurred())
					})
				})

				for _, testCase := range []ValidatorErrorCase{
					{
						"from",
						&models.DownloadAction{
							To: "local_location",
						},
					},
					{
						"to",
						&models.DownloadAction{
							From: "web_location",
						},
					},
					{
						"user",
						&models.DownloadAction{
							From: "web_location",
							To:   "local_location",
						},
					},
				} {
					testValidatorErrorCase(testCase)
				}
			})
		})

		Context("with checksum algorithm / value", func() {
			itSerializesAndDeserializes(
				`{
					"artifact": "mouse",
					"from": "web_location",
					"to": "local_location",
					"cache_key": "elephant",
					"user": "someone",
					"checksum_algorithm": "md5",
					"checksum_value": "some checksum"
			}`,
				models.WrapAction(&models.DownloadAction{
					Artifact:          "mouse",
					From:              "web_location",
					To:                "local_location",
					CacheKey:          "elephant",
					User:              "someone",
					ChecksumAlgorithm: "md5",
					ChecksumValue:     "some checksum",
				}),
			)

			Describe("Validate", func() {
				BeforeEach(func() {
					downloadAction = &models.DownloadAction{
						From:              "web_location",
						To:                "local_location",
						User:              "someone",
						ChecksumAlgorithm: "md5",
						ChecksumValue:     "some checksum",
					}
				})

				It("is valid", func() {
					err := downloadAction.Validate()
					Expect(err).NotTo(HaveOccurred())
				})

				Context("with checksum", func() {
					for _, testCase := range []ValidatorErrorCase{
						{
							"checksum value",
							&models.DownloadAction{
								From:              "web_location",
								To:                "local_location",
								User:              "someone",
								ChecksumAlgorithm: "md5",
								ChecksumValue:     "",
							},
						},
						{
							"checksum algorithm",
							&models.DownloadAction{
								From:              "web_location",
								To:                "local_location",
								User:              "someone",
								ChecksumAlgorithm: "",
								ChecksumValue:     "some checksum",
							},
						},
						{
							"invalid algorithm",
							&models.DownloadAction{
								From:              "web_location",
								To:                "local_location",
								User:              "someone",
								ChecksumAlgorithm: "invlalid-alg",
								ChecksumValue:     "some checksum",
							},
						},
					} {
						testValidatorErrorCase(testCase)
					}
				})
			})
		})
	})

	Describe("Upload", func() {
		itSerializesAndDeserializes(
			`{
					"artifact": "mouse",
					"from": "local_location",
					"to": "web_location",
					"user": "someone"
			}`,
			models.WrapAction(&models.UploadAction{
				Artifact: "mouse",
				From:     "local_location",
				To:       "web_location",
				User:     "someone",
			}),
		)

		Describe("Validate", func() {
			var uploadAction *models.UploadAction

			Context("when the action has 'from', 'to', and 'user' specified", func() {
				It("is valid", func() {
					uploadAction = &models.UploadAction{
						To:   "web_location",
						From: "local_location",
						User: "someone",
					}

					err := uploadAction.Validate()
					Expect(err).NotTo(HaveOccurred())
				})
			})

			for _, testCase := range []ValidatorErrorCase{
				{
					"from",
					&models.UploadAction{
						To: "web_location",
					},
				},
				{
					"to",
					&models.UploadAction{
						From: "local_location",
					},
				},
				{
					"user",
					&models.UploadAction{
						To:   "web_location",
						From: "local_location",
					},
				},
			} {
				testValidatorErrorCase(testCase)
			}
		})
	})

	Describe("Run", func() {
		var (
			nofile uint64 = 10
			nproc  uint64 = 20
		)
		resourceLimits := &models.ResourceLimits{}
		resourceLimits.SetNofile(&nofile)
		resourceLimits.SetNproc(&nproc)
		itSerializesAndDeserializes(
			`{
					"user": "me",
					"path": "rm",
					"args": ["-rf", "/"],
					"dir": "./some-dir",
					"env": [
						{"name":"FOO", "value":"1"},
						{"name":"BAR", "value":"2"}
					],
					"resource_limits":{"nofile": "10", "nproc": "20"}
			}`,
			//TODO: double-check these properties
			// protobuf v3 makes uint64 a string
			//"resource_limits":{"nofile": 10, "nproc": 20}
			// protojson does not emit this
			//"suppress_log_output": false
			models.WrapAction(&models.RunAction{
				User: "me",
				Path: "rm",
				Dir:  "./some-dir",
				Args: []string{"-rf", "/"},
				Env: []*models.EnvironmentVariable{
					{"FOO", "1"},
					{"BAR", "2"},
				},
				ResourceLimits: resourceLimits,
				// SuppressLogOutput: false,
			}),
		)

		Describe("Validate", func() {
			var runAction *models.RunAction

			Context("when the action has the required fields", func() {
				It("is valid", func() {
					runAction = &models.RunAction{
						Path: "ls",
						User: "foo",
					}

					err := runAction.Validate()
					Expect(err).NotTo(HaveOccurred())
				})
			})

			for _, testCase := range []ValidatorErrorCase{
				{
					"path",
					&models.RunAction{
						User: "me",
					},
				},
				{
					"user",
					&models.RunAction{
						Path: "ls",
					},
				},
			} {
				testValidatorErrorCase(testCase)
			}
		})
	})

	Describe("Timeout", func() {
		var nofile uint64 = 10

		resourceLimits := &models.ResourceLimits{}
		resourceLimits.SetNofile(&nofile)

		itSerializesAndDeserializes(
			`{
				"action": {
					"run": {
						"path": "echo",
						"resource_limits":{
							"nofile": "10"
						},
						"user": "someone"
					}
				},
				"timeout_ms": "10"
			}`,
			//TODO: double-check these properties
			// protobuf v3 makes uint64 a string
			//"resource_limits":{"nofile": 10}
			//"timeout_ms": 10
			// protojson does not emit this
			//"suppress_log_output": false
			models.WrapAction(
				models.Timeout(
					&models.RunAction{
						Path:           "echo",
						User:           "someone",
						ResourceLimits: resourceLimits,
					},
					10*time.Millisecond,
				)),
		)

		Describe("Validate", func() {
			var timeoutAction *models.TimeoutAction

			Context("when the action has 'action' specified and a positive timeout", func() {
				It("is valid", func() {
					timeoutAction = &models.TimeoutAction{
						Action: &models.Action{
							UploadAction: &models.UploadAction{
								From: "local_location",
								To:   "web_location",
								User: "someone",
							},
						},
						TimeoutMs: int64(time.Second / 1000000),
					}

					err := timeoutAction.Validate()
					Expect(err).NotTo(HaveOccurred())
				})
			})

			for _, testCase := range []ValidatorErrorCase{
				{
					"action",
					&models.TimeoutAction{
						TimeoutMs: int64(time.Second / 1000000),
					},
				},
				{
					"from",
					&models.TimeoutAction{
						Action: &models.Action{
							UploadAction: &models.UploadAction{
								To:   "web_location",
								User: "someone",
							},
						},
						TimeoutMs: int64(time.Second / 1000000),
					},
				},
				{
					"timeout_ms",
					&models.TimeoutAction{
						Action: &models.Action{
							UploadAction: &models.UploadAction{
								From: "local_location",
								To:   "web_location",
								User: "someone",
							},
						},
					},
				},
			} {
				testValidatorErrorCase(testCase)
			}
		})
	})

	Describe("Try", func() {
		itSerializesAndDeserializes(
			`{
					"action": {
						"run": {
							"path": "echo",
							"resource_limits":{},
							"user": "me"
						}
					}
			}`,
			//TODO: double-check these properties
			// protojson does not emit this
			//"suppress_log_output": false
			models.WrapAction(models.Try(&models.RunAction{
				Path:           "echo",
				User:           "me",
				ResourceLimits: &models.ResourceLimits{},
			})),
		)

		Describe("Validate", func() {
			var tryAction *models.TryAction

			Context("when the action has 'action' specified", func() {
				It("is valid", func() {
					tryAction = &models.TryAction{
						Action: &models.Action{
							UploadAction: &models.UploadAction{
								From: "local_location",
								To:   "web_location",
								User: "someone",
							},
						},
					}

					err := tryAction.Validate()
					Expect(err).NotTo(HaveOccurred())
				})
			})

			for _, testCase := range []ValidatorErrorCase{
				{
					"action",
					&models.TryAction{},
				},
				{
					"from",
					&models.TryAction{
						Action: &models.Action{
							UploadAction: &models.UploadAction{
								To: "web_location",
							},
						},
					},
				},
			} {
				testValidatorErrorCase(testCase)
			}
		})
	})

	Describe("Parallel", func() {
		itSerializesAndDeserializes(
			`{
					"actions": [
						{
							"download": {
								"cache_key": "elephant",
								"to": "local_location",
								"from": "web_location",
								"user": "someone"
							}
						},
						{
							"run": {
								"resource_limits": {},
								"path": "echo",
								"user": "me"
							}
						}
					]
			}`,
			//TODO: double-check these properties
			// protojson does not emit this
			//"suppress_log_output": false
			models.WrapAction(models.Parallel(
				&models.DownloadAction{
					From:     "web_location",
					To:       "local_location",
					CacheKey: "elephant",
					User:     "someone",
				},
				&models.RunAction{
					Path:           "echo",
					User:           "me",
					ResourceLimits: &models.ResourceLimits{},
				},
			)),
		)

		Describe("Validate", func() {
			var parallelAction *models.ParallelAction

			Context("when the action has 'actions' as a slice of valid actions", func() {
				It("is valid", func() {
					parallelAction = &models.ParallelAction{
						Actions: []*models.Action{
							{
								UploadAction: &models.UploadAction{
									From: "local_location",
									To:   "web_location",
									User: "someone",
								},
							},
						},
					}

					err := parallelAction.Validate()
					Expect(err).NotTo(HaveOccurred())
				})
			})

			for _, testCase := range []ValidatorErrorCase{
				{
					"actions",
					&models.ParallelAction{},
				},
				{
					"actions",
					&models.ParallelAction{
						Actions: []*models.Action{},
					},
				},
				{
					"action at index 0",
					&models.ParallelAction{
						Actions: []*models.Action{nil},
					},
				},
				{
					"from",
					&models.ParallelAction{
						Actions: []*models.Action{
							{
								UploadAction: &models.UploadAction{
									To: "web_location",
								},
							},
						},
					},
				},
			} {
				testValidatorErrorCase(testCase)
			}
		})
	})

	Describe("Serial", func() {
		itSerializesAndDeserializes(
			`{
					"actions": [
						{
							"download": {
								"cache_key": "elephant",
								"to": "local_location",
								"from": "web_location",
								"user": "someone"
							}
						},
						{
							"run": {
								"resource_limits": {},
								"path": "echo",
								"user": "me"
							}
						}
					]
			}`,
			//TODO: double-check these properties
			// protojson does not emit this
			//"suppress_log_output": false
			models.WrapAction(models.Serial(
				&models.DownloadAction{
					From:     "web_location",
					To:       "local_location",
					CacheKey: "elephant",
					User:     "someone",
				},
				&models.RunAction{
					Path:           "echo",
					User:           "me",
					ResourceLimits: &models.ResourceLimits{},
				},
			)),
		)

		Describe("Validate", func() {
			var serialAction *models.SerialAction

			Context("when the action has 'actions' as a slice of valid actions", func() {
				It("is valid", func() {
					serialAction = models.Serial(
						&models.UploadAction{
							From: "local_location",
							To:   "web_location",
							User: "someone",
						},
					)
					err := serialAction.Validate()
					Expect(err).NotTo(HaveOccurred())
				})
			})

			for _, testCase := range []ValidatorErrorCase{
				{
					"actions",
					&models.SerialAction{},
				},
				{
					"actions",
					&models.SerialAction{
						Actions: []*models.Action{},
					},
				},
				{
					"action at index 0",
					&models.SerialAction{
						Actions: []*models.Action{nil},
					},
				},
				{
					"from",
					models.Serial(
						&models.UploadAction{
							To: "web_location",
						},
					),
				},
			} {
				testValidatorErrorCase(testCase)
			}
		})
	})

	Describe("EmitProgressAction", func() {
		itSerializesAndDeserializes(
			`{
					"start_message": "reticulating splines",
					"success_message": "reticulated splines",
					"failure_message_prefix": "reticulation failed",
					"action": {
						"run": {
							"path": "echo",
							"resource_limits":{},
							"user": "me"
						}
					}
			}`,
			//TODO: double-check these properties
			// protojson does not emit this
			//"suppress_log_output": false
			models.WrapAction(models.EmitProgressFor(
				&models.RunAction{
					Path:           "echo",
					User:           "me",
					ResourceLimits: &models.ResourceLimits{},
				},
				"reticulating splines", "reticulated splines", "reticulation failed",
			)),
		)

		Describe("Validate", func() {
			var emitProgressAction *models.EmitProgressAction

			Context("when the action has 'action' specified", func() {
				It("is valid", func() {
					emitProgressAction = models.EmitProgressFor(
						&models.UploadAction{
							From: "local_location",
							To:   "web_location",
							User: "someone",
						},
						"", "", "",
					)

					err := emitProgressAction.Validate()
					Expect(err).NotTo(HaveOccurred())
				})
			})

			for _, testCase := range []ValidatorErrorCase{
				{
					"action",
					&models.EmitProgressAction{},
				},
				{
					"from",
					models.EmitProgressFor(
						&models.UploadAction{
							To: "web_location",
						},
						"", "", "",
					),
				},
			} {
				testValidatorErrorCase(testCase)
			}
		})
	})

	Describe("Codependent", func() {
		itSerializesAndDeserializes(
			`{
					"actions": [
						{
							"download": {
								"cache_key": "elephant",
								"to": "local_location",
								"from": "web_location",
								"user": "someone"
							}
						},
						{
							"run": {
								"resource_limits": {},
								"path": "echo",
								"user": "me"
							}
						}
					]
			}`,
			//TODO: double-check these properties
			// protojson does not emit this
			//"suppress_log_output": false
			models.WrapAction(models.Codependent(
				&models.DownloadAction{
					From:     "web_location",
					To:       "local_location",
					CacheKey: "elephant",
					User:     "someone",
				},
				&models.RunAction{Path: "echo", User: "me", ResourceLimits: &models.ResourceLimits{}},
			)),
		)

		Describe("Validate", func() {
			var codependentAction *models.CodependentAction

			Context("when the action has 'actions' as a slice of valid actions", func() {
				It("is valid", func() {
					codependentAction = models.Codependent(
						&models.UploadAction{
							From: "local_location",
							To:   "web_location",
							User: "someone",
						},
					)

					err := codependentAction.Validate()
					Expect(err).NotTo(HaveOccurred())
				})
			})

			for _, testCase := range []ValidatorErrorCase{
				{
					"actions",
					&models.CodependentAction{},
				},
				{
					"actions",
					&models.CodependentAction{
						Actions: []*models.Action{}},
				},
				{
					"action at index 0",
					&models.CodependentAction{
						Actions: []*models.Action{nil}},
				},
				{
					"from",
					models.Codependent(&models.UploadAction{
						To: "web_location",
					}),
				},
			} {
				testValidatorErrorCase(testCase)
			}
		})
	})
})
