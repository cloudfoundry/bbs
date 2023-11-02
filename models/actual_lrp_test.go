package models_test

import (
	"encoding/json"
	"fmt"
	"time"

	"code.cloudfoundry.org/bbs/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func defaultCrashedActual(crashCount int32, lastCrashed int64) *models.ActualLRP {
	return &models.ActualLRP{
		ActualLRPKey: models.NewActualLRPKey("p-guid", 0, "domain"),
		State:        models.ActualLRPStateCrashed,
		CrashCount:   crashCount,
		Since:        lastCrashed,
	}
}

type crashInfoTest interface {
	Test()
}

type crashInfoTests []crashInfoTest

func (tests crashInfoTests) Test() {
	for _, test := range tests {
		test.Test()
	}
}

type crashInfoBackoffTest struct {
	*models.ActualLRP
	WaitTime time.Duration
}

func newCrashInfoBackoffTest(crashCount int32, lastCrashed int64, waitTime time.Duration) crashInfoTest {
	return crashInfoBackoffTest{
		ActualLRP: defaultCrashedActual(crashCount, lastCrashed),
		WaitTime:  waitTime,
	}
}

func (test crashInfoBackoffTest) Test() {
	Context(fmt.Sprintf("when the crashCount is %d and the wait time is %s", test.CrashCount, test.WaitTime), func() {
		It("should NOT restart before the expected wait time", func() {
			calc := models.NewDefaultRestartCalculator()
			currentTimestamp := test.GetSince() + test.WaitTime.Nanoseconds() - time.Second.Nanoseconds()
			Expect(test.ShouldRestartCrash(time.Unix(0, currentTimestamp), calc)).To(BeFalse())
		})

		It("should restart after the expected wait time", func() {
			calc := models.NewDefaultRestartCalculator()
			currentTimestamp := test.GetSince() + test.WaitTime.Nanoseconds()
			Expect(test.ShouldRestartCrash(time.Unix(0, currentTimestamp), calc)).To(BeTrue())
		})
	})
}

type crashInfoNeverStartTest struct {
	*models.ActualLRP
}

func newCrashInfoNeverStartTest(crashCount int32, lastCrashed int64) crashInfoTest {
	return crashInfoNeverStartTest{
		ActualLRP: defaultCrashedActual(crashCount, lastCrashed),
	}
}

func (test crashInfoNeverStartTest) Test() {
	Context(fmt.Sprintf("when the crashCount is %d", test.CrashCount), func() {
		It("should never restart regardless of the wait time", func() {
			calc := models.NewDefaultRestartCalculator()
			theFuture := test.GetSince() + time.Hour.Nanoseconds()
			Expect(test.ShouldRestartCrash(time.Unix(0, 0), calc)).To(BeFalse())
			Expect(test.ShouldRestartCrash(time.Unix(0, test.GetSince()), calc)).To(BeFalse())
			Expect(test.ShouldRestartCrash(time.Unix(0, theFuture), calc)).To(BeFalse())
		})
	})
}

type crashInfoAlwaysStartTest struct {
	*models.ActualLRP
}

func newCrashInfoAlwaysStartTest(crashCount int32, lastCrashed int64) crashInfoTest {
	return crashInfoAlwaysStartTest{
		ActualLRP: defaultCrashedActual(crashCount, lastCrashed),
	}
}

func (test crashInfoAlwaysStartTest) Test() {
	Context(fmt.Sprintf("when the crashCount is %d", test.CrashCount), func() {
		It("should restart regardless of the wait time", func() {
			calc := models.NewDefaultRestartCalculator()
			theFuture := test.GetSince() + time.Hour.Nanoseconds()
			Expect(test.ShouldRestartCrash(time.Unix(0, 0), calc)).To(BeTrue())
			Expect(test.ShouldRestartCrash(time.Unix(0, test.GetSince()), calc)).To(BeTrue())
			Expect(test.ShouldRestartCrash(time.Unix(0, theFuture), calc)).To(BeTrue())
		})
	})
}

func testBackoffCount(maxBackoffDuration time.Duration, expectedBackoffCount int32) {
	It(fmt.Sprintf("sets the MaxBackoffCount to %d based on the MaxBackoffDuration %s and the CrashBackoffMinDuration", expectedBackoffCount, maxBackoffDuration), func() {
		calc := models.NewRestartCalculator(models.DefaultImmediateRestarts, maxBackoffDuration, models.DefaultMaxRestarts)
		Expect(calc.MaxBackoffCount).To(Equal(expectedBackoffCount))
	})
}

var _ = Describe("RestartCalculator", func() {

	Describe("NewRestartCalculator", func() {
		testBackoffCount(20*time.Minute, 5)
		testBackoffCount(16*time.Minute, 5)
		testBackoffCount(8*time.Minute, 4)
		testBackoffCount(119*time.Second, 2)
		testBackoffCount(120*time.Second, 2)
		testBackoffCount(models.CrashBackoffMinDuration, 0)

		It("should work...", func() {
			nanoseconds := func(seconds int) int64 {
				return int64(seconds * 1000000000)
			}

			calc := models.NewRestartCalculator(3, 119*time.Second, 200)
			Expect(calc.ShouldRestart(0, 0, 0)).To(BeTrue())
			Expect(calc.ShouldRestart(0, 0, 1)).To(BeTrue())
			Expect(calc.ShouldRestart(0, 0, 2)).To(BeTrue())

			Expect(calc.ShouldRestart(0, 0, 3)).To(BeFalse())
			Expect(calc.ShouldRestart(nanoseconds(30), 0, 3)).To(BeTrue())

			Expect(calc.ShouldRestart(nanoseconds(30), 0, 4)).To(BeFalse())
			Expect(calc.ShouldRestart(nanoseconds(59), 0, 4)).To(BeFalse())
			Expect(calc.ShouldRestart(nanoseconds(60), 0, 4)).To(BeTrue())
			Expect(calc.ShouldRestart(nanoseconds(60), 0, 5)).To(BeFalse())
			Expect(calc.ShouldRestart(nanoseconds(118), 0, 5)).To(BeFalse())
			Expect(calc.ShouldRestart(nanoseconds(119), 0, 5)).To(BeTrue())
		})
	})

	Describe("Validate", func() {
		It("the default values are valid", func() {
			calc := models.NewDefaultRestartCalculator()
			Expect(calc.Validate()).NotTo(HaveOccurred())
		})

		It("invalid when MaxBackoffDuration is lower than the CrashBackoffMinDuration", func() {
			calc := models.NewRestartCalculator(models.DefaultImmediateRestarts, models.CrashBackoffMinDuration-time.Second, models.DefaultMaxRestarts)
			Expect(calc.Validate()).To(HaveOccurred())
		})
	})
})

var _ = Describe("ActualLRP", func() {
	Describe("ToActualLRP", func() {
		var actualLRPInfo models.ActualLRPInfo

		BeforeEach(func() {
			actualLRPInfo = models.ActualLRPInfo{}
		})

		It("updates availability zone", func() {
			actualLRPInfo.AvailabilityZone = "some-zone-1"
			actualLRP := actualLRPInfo.ToActualLRP(models.NewActualLRPKey("p-guid", 0, "domain"), models.NewActualLRPInstanceKey("i-1", "cell-1"))
			Expect(actualLRP.AvailabilityZone).To(Equal("some-zone-1"))
		})

		Context("when Routable is not provided", func() {
			It("does not set routable", func() {
				actualLRP := actualLRPInfo.ToActualLRP(models.NewActualLRPKey("p-guid", 0, "domain"), models.NewActualLRPInstanceKey("i-1", "cell-1"))
				Expect(actualLRP.RoutableExists()).To(Equal(false))
			})
		})

		Context("when Routable is false", func() {
			BeforeEach(func() {
				actualLRPInfo.SetRoutable(false)
			})

			It("sets routable to provided value", func() {
				actualLRP := actualLRPInfo.ToActualLRP(models.NewActualLRPKey("p-guid", 0, "domain"), models.NewActualLRPInstanceKey("i-1", "cell-1"))
				Expect(actualLRP.RoutableExists()).To(Equal(true))
				Expect(actualLRP.GetRoutable()).To(Equal(false))
			})
		})

		Context("when Routable is true", func() {
			BeforeEach(func() {
				actualLRPInfo.SetRoutable(true)
			})

			It("sets routable to provided value", func() {
				actualLRP := actualLRPInfo.ToActualLRP(models.NewActualLRPKey("p-guid", 0, "domain"), models.NewActualLRPInstanceKey("i-1", "cell-1"))
				Expect(actualLRP.RoutableExists()).To(Equal(true))
				Expect(actualLRP.GetRoutable()).To(Equal(true))
			})
		})
	})

	Describe("ShouldRestartCrash", func() {
		Context("when the lpr is CRASHED", func() {
			const maxWaitTime = 16 * time.Minute
			var now = time.Now().UnixNano()
			var crashTests = crashInfoTests{
				newCrashInfoAlwaysStartTest(0, now),
				newCrashInfoAlwaysStartTest(1, now),
				newCrashInfoAlwaysStartTest(2, now),
				newCrashInfoBackoffTest(3, now, 30*time.Second),
				newCrashInfoBackoffTest(7, now, 8*time.Minute),
				newCrashInfoBackoffTest(8, now, maxWaitTime),
				newCrashInfoBackoffTest(199, now, maxWaitTime),
				newCrashInfoNeverStartTest(200, now),
				newCrashInfoNeverStartTest(201, now),
			}

			crashTests.Test()
		})

		Context("when the lrp is not CRASHED", func() {
			It("returns false", func() {
				now := time.Now()
				actual := defaultCrashedActual(0, now.UnixNano())
				calc := models.NewDefaultRestartCalculator()
				for _, state := range models.ActualLRPStates {
					actual.State = state
					if state == models.ActualLRPStateCrashed {
						Expect(actual.ShouldRestartCrash(now, calc)).To(BeTrue(), "should restart CRASHED lrp")
					} else {
						Expect(actual.ShouldRestartCrash(now, calc)).To(BeFalse(), fmt.Sprintf("should not restart %s lrp", state))
					}
				}
			})
		})
	})

	Describe("ActualLRPKey", func() {
		Describe("Validate", func() {
			var actualLRPKey models.ActualLRPKey

			BeforeEach(func() {
				actualLRPKey = models.NewActualLRPKey("process-guid", 1, "domain")
			})

			Context("when valid", func() {
				It("returns nil", func() {
					Expect(actualLRPKey.Validate()).To(BeNil())
				})
			})

			Context("when the ProcessGuid is blank", func() {
				BeforeEach(func() {
					actualLRPKey.ProcessGuid = ""
				})

				It("returns a validation error", func() {
					Expect(actualLRPKey.Validate()).To(ConsistOf(models.ErrInvalidField{"process_guid"}))
				})
			})

			Context("when the Domain is blank", func() {
				BeforeEach(func() {
					actualLRPKey.Domain = ""
				})

				It("returns a validation error", func() {
					Expect(actualLRPKey.Validate()).To(ConsistOf(models.ErrInvalidField{"domain"}))
				})
			})

			Context("when the Index is negative", func() {
				BeforeEach(func() {
					actualLRPKey.Index = -1
				})

				It("returns a validation error", func() {
					Expect(actualLRPKey.Validate()).To(ConsistOf(models.ErrInvalidField{"index"}))
				})
			})
		})
	})

	Describe("ActualLRPInstanceKey", func() {
		Describe("Validate", func() {
			var actualLRPInstanceKey models.ActualLRPInstanceKey

			Context("when both instance guid and cell id are specified", func() {
				It("returns nil", func() {
					actualLRPInstanceKey = models.NewActualLRPInstanceKey("instance-guid", "cell-id")
					Expect(actualLRPInstanceKey.Validate()).To(BeNil())
				})
			})

			Context("when both instance guid and cell id are empty", func() {
				It("returns a validation error", func() {
					actualLRPInstanceKey = models.NewActualLRPInstanceKey("", "")
					Expect(actualLRPInstanceKey.Validate()).To(ConsistOf(
						models.ErrInvalidField{"cell_id"},
						models.ErrInvalidField{"instance_guid"},
					))

				})
			})

			Context("when only the instance guid is specified", func() {
				It("returns a validation error", func() {
					actualLRPInstanceKey = models.NewActualLRPInstanceKey("instance-guid", "")
					Expect(actualLRPInstanceKey.Validate()).To(ConsistOf(models.ErrInvalidField{"cell_id"}))
				})
			})

			Context("when only the cell id is specified", func() {
				It("returns a validation error", func() {
					actualLRPInstanceKey = models.NewActualLRPInstanceKey("", "cell-id")
					Expect(actualLRPInstanceKey.Validate()).To(ConsistOf(models.ErrInvalidField{"instance_guid"}))
				})
			})
		})

		Describe("ActualLRPNetInfo", func() {
			Describe("EmptyActualLRPNetInfo", func() {
				It("returns a net info with an empty address, non-nil empty PortMapping slice, and unknown address preference", func() {
					netInfo := models.EmptyActualLRPNetInfo()

					Expect(netInfo.GetAddress()).To(BeEmpty())
					Expect(netInfo.GetPorts()).To(BeEmpty())
					Expect(netInfo.PreferredAddress).To(Equal(models.ActualLRPNetInfo_PreferredAddressUnknown))
				})
			})

			Describe("ActualLRPNetInfo_PreferredAddress", func() {
				Describe("serialization", func() {
					DescribeTable("marshals and unmarshals between the value and the expected JSON output",
						func(v models.ActualLRPNetInfo_PreferredAddress, expectedJSON string) {
							Expect(json.Marshal(v)).To(MatchJSON(expectedJSON))
							var testV models.ActualLRPNetInfo_PreferredAddress
							Expect(json.Unmarshal([]byte(expectedJSON), &testV)).To(Succeed())
							Expect(testV).To(Equal(v))
						},
						Entry("UNKNOWN", models.ActualLRPNetInfo_PreferredAddressUnknown, `"UNKNOWN"`),
						Entry("INSTANCE", models.ActualLRPNetInfo_PreferredAddressInstance, `"INSTANCE"`),
						Entry("HOST", models.ActualLRPNetInfo_PreferredAddressHost, `"HOST"`),
					)
				})
			})
		})
	})

	Describe("ActualLRPGroup", func() {
		Describe("Resolve", func() {
			var (
				instanceLRP   *models.ActualLRP
				evacuatingLRP *models.ActualLRP

				group models.ActualLRPGroup

				resolvedLRP *models.ActualLRP
				evacuating  bool

				resolveError error
			)

			BeforeEach(func() {
				lrpKey := models.NewActualLRPKey("process-guid", 1, "domain")
				instanceLRP = &models.ActualLRP{
					ActualLRPKey: lrpKey,
					Since:        1138,
				}
				evacuatingLRP = &models.ActualLRP{
					ActualLRPKey: lrpKey,
					Since:        3417,
				}
			})

			JustBeforeEach(func() {
				resolvedLRP, evacuating, resolveError = group.Resolve()
			})

			Context("When only the Instance LRP is set", func() {
				BeforeEach(func() {
					group = models.ActualLRPGroup{
						Instance: instanceLRP,
					}
				})

				JustBeforeEach(func() {
					Expect(resolveError).NotTo(HaveOccurred())
				})

				It("returns the Instance LRP", func() {
					Expect(resolvedLRP).To(Equal(instanceLRP))
					Expect(evacuating).To(BeFalse())
				})
			})

			Context("When only the Evacuating LRP is set", func() {
				BeforeEach(func() {
					group = models.ActualLRPGroup{
						Evacuating: evacuatingLRP,
					}
				})

				JustBeforeEach(func() {
					Expect(resolveError).NotTo(HaveOccurred())
				})

				It("returns the Evacuating LRP", func() {
					Expect(resolvedLRP).To(Equal(evacuatingLRP))
					Expect(evacuating).To(BeTrue())
				})
			})

			Context("When both the Instance and the Evacuating LRP are set", func() {
				BeforeEach(func() {
					group = models.ActualLRPGroup{
						Evacuating: evacuatingLRP,
						Instance:   instanceLRP,
					}
				})

				JustBeforeEach(func() {
					Expect(resolveError).NotTo(HaveOccurred())
				})

				Context("When the Instance is UNCLAIMED", func() {
					BeforeEach(func() {
						instanceLRP.State = models.ActualLRPStateUnclaimed
					})

					It("returns the Evacuating LRP", func() {
						Expect(resolvedLRP).To(Equal(evacuatingLRP))
						Expect(evacuating).To(BeTrue())
					})
				})

				Context("When the Instance is CLAIMED", func() {
					BeforeEach(func() {
						instanceLRP.State = models.ActualLRPStateClaimed
					})

					It("returns the Evacuating LRP", func() {
						Expect(resolvedLRP).To(Equal(evacuatingLRP))
						Expect(evacuating).To(BeTrue())
					})
				})

				Context("When the Instance is RUNNING", func() {
					BeforeEach(func() {
						instanceLRP.State = models.ActualLRPStateRunning
					})

					It("returns the Instance LRP", func() {
						Expect(resolvedLRP).To(Equal(instanceLRP))
						Expect(evacuating).To(BeFalse())
					})
				})

				Context("When the Instance is CRASHED", func() {
					BeforeEach(func() {
						instanceLRP.State = models.ActualLRPStateCrashed
					})

					It("returns the Instance LRP", func() {
						Expect(resolvedLRP).To(Equal(instanceLRP))
						Expect(evacuating).To(BeFalse())
					})
				})
			})

			Context("When both the Instance and the Evacuating are nil", func() {
				BeforeEach(func() {
					group = models.ActualLRPGroup{
						Evacuating: nil,
						Instance:   nil,
					}
				})

				It("returns an error", func() {
					Expect(resolveError).To(MatchError("ActualLRPGroup invalid"))
				})
			})
		})
	})

	Describe("ActualLRP", func() {
		var lrp models.ActualLRP
		var lrpKey models.ActualLRPKey
		var instanceKey models.ActualLRPInstanceKey
		var netInfo models.ActualLRPNetInfo

		BeforeEach(func() {
			lrpKey = models.NewActualLRPKey("some-guid", 2, "some-domain")
			instanceKey = models.NewActualLRPInstanceKey("some-instance-guid", "some-cell-id")
			netInfo = models.NewActualLRPNetInfo("1.2.3.4", "2.2.2.2", models.ActualLRPNetInfo_PreferredAddressUnknown, models.NewPortMapping(5678, 8080), models.NewPortMapping(1234, 8081))

			lrp = models.ActualLRP{
				ActualLRPKey:         lrpKey,
				ActualLRPInstanceKey: instanceKey,
				ActualLRPNetInfo:     netInfo,
				CrashCount:           1,
				State:                models.ActualLRPStateRunning,
				Since:                1138,
				ModificationTag: models.ModificationTag{
					Epoch: "some-guid",
					Index: 50,
				},
			}
		})

		Describe("AllowsTransitionTo", func() {
			var (
				before   *models.ActualLRP
				afterKey models.ActualLRPKey
			)

			BeforeEach(func() {
				before = &models.ActualLRP{
					ActualLRPKey: models.NewActualLRPKey("fake-process-guid", 1, "fake-domain"),
				}
				afterKey = models.ActualLRPKey{}
				afterKey = before.ActualLRPKey
			})

			Context("when the ProcessGuid fields differ", func() {
				BeforeEach(func() {
					before.ProcessGuid = "some-process-guid"
					afterKey.ProcessGuid = "another-process-guid"
				})

				It("is not allowed", func() {
					Expect(before.AllowsTransitionTo(&afterKey, &before.ActualLRPInstanceKey, before.GetState())).To(BeFalse())
				})
			})

			Context("when the Index fields differ", func() {
				BeforeEach(func() {
					before.Index = 1138
					afterKey.Index = 3417
				})

				It("is not allowed", func() {
					Expect(before.AllowsTransitionTo(&afterKey, &before.ActualLRPInstanceKey, before.GetState())).To(BeFalse())
				})
			})

			Context("when the Domain fields differ", func() {
				BeforeEach(func() {
					before.Domain = "some-domain"
					afterKey.Domain = "another-domain"
				})

				It("is not allowed", func() {
					Expect(before.AllowsTransitionTo(&afterKey, &before.ActualLRPInstanceKey, before.GetState())).To(BeFalse())
				})
			})

			Context("when the ProcessGuid, Index, and Domain are equivalent", func() {
				var (
					emptyKey                 = models.NewActualLRPInstanceKey("", "")
					claimedKey               = models.NewActualLRPInstanceKey("some-instance-guid", "some-cell-id")
					differentInstanceGuidKey = models.NewActualLRPInstanceKey("some-other-instance-guid", "some-cell-id")
					differentCellIDKey       = models.NewActualLRPInstanceKey("some-instance-guid", "some-other-cell-id")

					equivalentEmptyKey   = models.NewActualLRPInstanceKey("", "")
					equivalentClaimedKey = models.NewActualLRPInstanceKey("some-instance-guid", "some-cell-id")
				)

				type stateTableEntry struct {
					BeforeState       string
					AfterState        string
					BeforeInstanceKey models.ActualLRPInstanceKey
					AfterInstanceKey  models.ActualLRPInstanceKey
					Allowed           bool
				}

				var EntryToString = func(entry stateTableEntry) string {
					return fmt.Sprintf("is %t when the before has state %s and instance guid '%s' and cell id '%s' and the after has state %s and instance guid '%s' and cell id '%s'",
						entry.Allowed,
						entry.BeforeState,
						entry.BeforeInstanceKey.GetInstanceGuid(),
						entry.BeforeInstanceKey.GetCellId(),
						entry.AfterState,
						entry.AfterInstanceKey.GetInstanceGuid(),
						entry.AfterInstanceKey.GetCellId(),
					)
				}

				stateTable := []stateTableEntry{
					{models.ActualLRPStateUnclaimed, models.ActualLRPStateUnclaimed, emptyKey, equivalentEmptyKey, true},
					{models.ActualLRPStateUnclaimed, models.ActualLRPStateClaimed, emptyKey, claimedKey, true},
					{models.ActualLRPStateUnclaimed, models.ActualLRPStateRunning, emptyKey, claimedKey, true},
					{models.ActualLRPStateUnclaimed, models.ActualLRPStateCrashed, emptyKey, claimedKey, false},
					{models.ActualLRPStateClaimed, models.ActualLRPStateUnclaimed, claimedKey, emptyKey, true},
					{models.ActualLRPStateClaimed, models.ActualLRPStateClaimed, claimedKey, equivalentClaimedKey, true},
					{models.ActualLRPStateClaimed, models.ActualLRPStateClaimed, claimedKey, differentInstanceGuidKey, false},
					{models.ActualLRPStateClaimed, models.ActualLRPStateClaimed, claimedKey, differentCellIDKey, false},
					{models.ActualLRPStateClaimed, models.ActualLRPStateRunning, claimedKey, equivalentClaimedKey, true},
					{models.ActualLRPStateClaimed, models.ActualLRPStateRunning, claimedKey, differentInstanceGuidKey, true},
					{models.ActualLRPStateClaimed, models.ActualLRPStateRunning, claimedKey, differentCellIDKey, true},
					{models.ActualLRPStateClaimed, models.ActualLRPStateCrashed, claimedKey, equivalentClaimedKey, true},
					{models.ActualLRPStateClaimed, models.ActualLRPStateCrashed, claimedKey, differentInstanceGuidKey, false},
					{models.ActualLRPStateClaimed, models.ActualLRPStateCrashed, claimedKey, differentCellIDKey, false},
					{models.ActualLRPStateRunning, models.ActualLRPStateUnclaimed, claimedKey, emptyKey, true},
					{models.ActualLRPStateRunning, models.ActualLRPStateClaimed, claimedKey, equivalentClaimedKey, true},
					{models.ActualLRPStateRunning, models.ActualLRPStateClaimed, claimedKey, differentInstanceGuidKey, false},
					{models.ActualLRPStateRunning, models.ActualLRPStateClaimed, claimedKey, differentCellIDKey, false},
					{models.ActualLRPStateRunning, models.ActualLRPStateRunning, claimedKey, equivalentClaimedKey, true},
					{models.ActualLRPStateRunning, models.ActualLRPStateRunning, claimedKey, differentInstanceGuidKey, false},
					{models.ActualLRPStateRunning, models.ActualLRPStateRunning, claimedKey, differentCellIDKey, false},
					{models.ActualLRPStateRunning, models.ActualLRPStateCrashed, claimedKey, equivalentClaimedKey, true},
					{models.ActualLRPStateRunning, models.ActualLRPStateCrashed, claimedKey, differentInstanceGuidKey, false},
					{models.ActualLRPStateRunning, models.ActualLRPStateCrashed, claimedKey, differentCellIDKey, false},
					{models.ActualLRPStateCrashed, models.ActualLRPStateUnclaimed, claimedKey, emptyKey, true},
					{models.ActualLRPStateCrashed, models.ActualLRPStateClaimed, claimedKey, equivalentClaimedKey, true},
					{models.ActualLRPStateCrashed, models.ActualLRPStateClaimed, claimedKey, differentInstanceGuidKey, false},
					{models.ActualLRPStateCrashed, models.ActualLRPStateClaimed, claimedKey, differentCellIDKey, false},
					{models.ActualLRPStateCrashed, models.ActualLRPStateRunning, claimedKey, equivalentClaimedKey, true},
					{models.ActualLRPStateCrashed, models.ActualLRPStateRunning, claimedKey, differentInstanceGuidKey, false},
					{models.ActualLRPStateCrashed, models.ActualLRPStateRunning, claimedKey, differentCellIDKey, false},
					{models.ActualLRPStateCrashed, models.ActualLRPStateCrashed, claimedKey, equivalentClaimedKey, false},
					{models.ActualLRPStateCrashed, models.ActualLRPStateCrashed, claimedKey, differentInstanceGuidKey, false},
					{models.ActualLRPStateCrashed, models.ActualLRPStateCrashed, claimedKey, differentCellIDKey, false},
				}

				for _, entry := range stateTable {
					entry := entry
					It(EntryToString(entry), func() {
						before.State = entry.BeforeState
						before.ActualLRPInstanceKey = entry.BeforeInstanceKey
						Expect(before.AllowsTransitionTo(&before.ActualLRPKey, &entry.AfterInstanceKey, entry.AfterState)).To(Equal(entry.Allowed))
					})
				}
			})
		})

		Describe("Validate", func() {
			Context("when state is unclaimed", func() {
				BeforeEach(func() {
					lrp = models.ActualLRP{
						ActualLRPKey: lrpKey,
						State:        models.ActualLRPStateUnclaimed,
						Since:        1138,
					}
				})

				itValidatesPresenceOfTheLRPKey(&lrp)
				itValidatesAbsenceOfTheInstanceKey(&lrp)
				itValidatesAbsenceOfNetInfo(&lrp)
				itValidatesPresenceOfPlacementError(&lrp)
				itValidatesOrdinaryPresence(&lrp)
			})

			Context("when state is claimed", func() {
				BeforeEach(func() {
					lrp = models.ActualLRP{
						ActualLRPKey:         lrpKey,
						ActualLRPInstanceKey: instanceKey,
						State:                models.ActualLRPStateClaimed,
						Since:                1138,
					}
				})

				itValidatesPresenceOfTheLRPKey(&lrp)
				itValidatesPresenceOfTheInstanceKey(&lrp)
				itValidatesAbsenceOfNetInfo(&lrp)
				itValidatesAbsenceOfPlacementError(&lrp)
			})

			Context("when state is running", func() {
				BeforeEach(func() {
					lrp = models.ActualLRP{
						ActualLRPKey:         lrpKey,
						ActualLRPInstanceKey: instanceKey,
						ActualLRPNetInfo:     netInfo,
						State:                models.ActualLRPStateRunning,
						Since:                1138,
					}
				})

				itValidatesPresenceOfTheLRPKey(&lrp)
				itValidatesPresenceOfTheInstanceKey(&lrp)
				itValidatesPresenceOfNetInfo(&lrp)
				itValidatesAbsenceOfPlacementError(&lrp)
			})

			Context("when state is not set", func() {
				BeforeEach(func() {
					lrp = models.ActualLRP{
						ActualLRPKey: lrpKey,
						State:        "",
						Since:        1138,
					}
				})

				It("validate returns an error", func() {
					err := lrp.Validate()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("state"))
				})

			})

			Context("when since is not set", func() {
				BeforeEach(func() {
					lrp = models.ActualLRP{
						ActualLRPKey: lrpKey,
						State:        models.ActualLRPStateUnclaimed,
						Since:        0,
					}
				})

				It("validate returns an error", func() {
					err := lrp.Validate()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("since"))
				})
			})

			Context("when state is crashed", func() {
				BeforeEach(func() {
					lrp = models.ActualLRP{
						ActualLRPKey: lrpKey,
						State:        models.ActualLRPStateCrashed,
						Since:        1138,
					}
				})

				itValidatesPresenceOfTheLRPKey(&lrp)
				itValidatesAbsenceOfTheInstanceKey(&lrp)
				itValidatesAbsenceOfNetInfo(&lrp)
				itValidatesAbsenceOfPlacementError(&lrp)
			})
		})
	})

	Describe("ResolveActualLRPGroups", func() {
		It("returns ordinary ActualLRPs in the instance slot of ActualLRPGroups", func() {
			lrp1 := &models.ActualLRP{
				ActualLRPKey:         models.NewActualLRPKey("process-guid-0", 0, "domain-0"),
				ActualLRPInstanceKey: models.NewActualLRPInstanceKey("instance-guid-0", "cell-id-0"),
				Presence:             models.ActualLRP_Ordinary,
				State:                models.ActualLRPStateRunning,
			}
			lrp2 := &models.ActualLRP{
				ActualLRPKey:         models.NewActualLRPKey("process-guid-1", 1, "domain-1"),
				ActualLRPInstanceKey: models.NewActualLRPInstanceKey("instance-guid-1", "cell-id-0"),
				Presence:             models.ActualLRP_Ordinary,
				State:                models.ActualLRPStateRunning,
			}
			groups := models.ResolveActualLRPGroups([]*models.ActualLRP{lrp1, lrp2})
			Expect(groups).To(ConsistOf(
				&models.ActualLRPGroup{Instance: lrp1},
				&models.ActualLRPGroup{Instance: lrp2},
			))
		})

		It("returns evacuating ActualLRPs in the evacuating slot of ActualLRPGroups", func() {
			lrp1 := &models.ActualLRP{
				ActualLRPKey:         models.NewActualLRPKey("process-guid-0", 0, "domain-0"),
				ActualLRPInstanceKey: models.NewActualLRPInstanceKey("instance-guid-0", "cell-id-0"),
				Presence:             models.ActualLRP_Evacuating,
				State:                models.ActualLRPStateRunning,
			}
			lrp2 := &models.ActualLRP{
				ActualLRPKey:         models.NewActualLRPKey("process-guid-0", 0, "domain-0"),
				ActualLRPInstanceKey: models.NewActualLRPInstanceKey("instance-guid-1", "cell-id-1"),
				Presence:             models.ActualLRP_Ordinary,
				State:                models.ActualLRPStateRunning,
			}
			groups := models.ResolveActualLRPGroups([]*models.ActualLRP{lrp1, lrp2})
			Expect(groups).To(ConsistOf(
				&models.ActualLRPGroup{Instance: lrp2, Evacuating: lrp1},
			))

		})

		DescribeTable("resolution priority of the Instance slot",
			func(
				supLRPState string, supLRPPresence models.ActualLRP_Presence,
				infLRPState string, infLRPPresence models.ActualLRP_Presence,
			) {
				supLRP := &models.ActualLRP{
					ActualLRPKey:         models.NewActualLRPKey("process-guid-0", 0, "domain-0"),
					ActualLRPInstanceKey: models.NewActualLRPInstanceKey("instance-guid-0", "cell-id-0"),
					Presence:             supLRPPresence,
					State:                supLRPState,
				}
				infLRP := &models.ActualLRP{
					ActualLRPKey:         models.NewActualLRPKey("process-guid-0", 0, "domain-0"),
					ActualLRPInstanceKey: models.NewActualLRPInstanceKey("instance-guid-1", "cell-id-1"),
					Presence:             infLRPPresence,
					State:                infLRPState,
				}
				groups := models.ResolveActualLRPGroups([]*models.ActualLRP{supLRP, infLRP})
				Expect(groups).To(ConsistOf(
					&models.ActualLRPGroup{Instance: supLRP},
				))
			},
			Entry("chooses RUNNING/Ordinary over RUNNING/Suspect",
				models.ActualLRPStateRunning, models.ActualLRP_Ordinary,
				models.ActualLRPStateRunning, models.ActualLRP_Suspect,
			),
			Entry("chooses RUNNING/Ordinary over CLAIMED/Suspect",
				models.ActualLRPStateRunning, models.ActualLRP_Ordinary,
				models.ActualLRPStateClaimed, models.ActualLRP_Suspect,
			),
			Entry("chooses RUNNING/Suspect over CLAIMED/Ordinary",
				models.ActualLRPStateRunning, models.ActualLRP_Suspect,
				models.ActualLRPStateClaimed, models.ActualLRP_Ordinary,
			),
			Entry("chooses RUNNING/Suspect over UNCLAIMED/Ordinary",
				models.ActualLRPStateRunning, models.ActualLRP_Suspect,
				models.ActualLRPStateUnclaimed, models.ActualLRP_Ordinary,
			),
			Entry("chooses RUNNING/Suspect over CRASHED/Ordinary",
				models.ActualLRPStateRunning, models.ActualLRP_Suspect,
				models.ActualLRPStateCrashed, models.ActualLRP_Ordinary,
			),
			Entry("chooses CLAIMED/Suspect over CLAIMED/Ordinary",
				models.ActualLRPStateClaimed, models.ActualLRP_Suspect,
				models.ActualLRPStateClaimed, models.ActualLRP_Ordinary,
			),
			Entry("chooses CLAIMED/Suspect over UNCLAIMED/Ordinary",
				models.ActualLRPStateClaimed, models.ActualLRP_Suspect,
				models.ActualLRPStateUnclaimed, models.ActualLRP_Ordinary,
			),
			Entry("chooses CLAIMED/Suspect over CRASHED/Ordinary",
				models.ActualLRPStateClaimed, models.ActualLRP_Suspect,
				models.ActualLRPStateCrashed, models.ActualLRP_Ordinary,
			),
		)

		Describe("ActualLRP_Presence", func() {
			Describe("serialization", func() {
				DescribeTable("marshals and unmarshals between the value and the expected JSON output",
					func(v models.ActualLRP_Presence, expectedJSON string) {
						Expect(json.Marshal(v)).To(MatchJSON(expectedJSON))
						var testV models.ActualLRP_Presence
						Expect(json.Unmarshal([]byte(expectedJSON), &testV)).To(Succeed())
						Expect(testV).To(Equal(v))
					},
					Entry("Ordinary", models.ActualLRP_Ordinary, `"ORDINARY"`),
					Entry("EVACUATING", models.ActualLRP_Evacuating, `"EVACUATING"`),
					Entry("SUSPECT", models.ActualLRP_Suspect, `"SUSPECT"`),
				)
			})
		})
	})
})

func itValidatesPresenceOfTheLRPKey(lrp *models.ActualLRP) {
	Context("when the lrp key is set", func() {
		BeforeEach(func() {
			lrp.ActualLRPKey = models.NewActualLRPKey("some-guid", 1, "domain")
		})

		It("validate does not return an error", func() {
			Expect(lrp.Validate()).NotTo(HaveOccurred())
		})
	})

	Context("when the lrp key is not set", func() {
		BeforeEach(func() {
			lrp.ActualLRPKey = models.ActualLRPKey{}
		})

		It("validate returns an error", func() {
			err := lrp.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("process_guid"))
		})
	})
}

func itValidatesPresenceOfTheInstanceKey(lrp *models.ActualLRP) {
	Context("when the instance key is set", func() {
		BeforeEach(func() {
			lrp.ActualLRPInstanceKey = models.NewActualLRPInstanceKey("some-instance", "some-cell")
		})

		It("validate does not return an error", func() {
			Expect(lrp.Validate()).NotTo(HaveOccurred())
		})
	})

	Context("when the instance key is not set", func() {
		BeforeEach(func() {
			lrp.ActualLRPInstanceKey = models.ActualLRPInstanceKey{}
		})

		It("validate returns an error", func() {
			err := lrp.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("instance_guid"))
		})
	})
}

func itValidatesAbsenceOfTheInstanceKey(lrp *models.ActualLRP) {
	Context("when the instance key is set", func() {
		BeforeEach(func() {
			lrp.ActualLRPInstanceKey = models.NewActualLRPInstanceKey("some-instance", "some-cell")
		})

		It("validate returns an error", func() {
			err := lrp.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("instance key"))
		})
	})

	Context("when the instance key is not set", func() {
		BeforeEach(func() {
			lrp.ActualLRPInstanceKey = models.ActualLRPInstanceKey{}
		})

		It("validate does not return an error", func() {
			Expect(lrp.Validate()).NotTo(HaveOccurred())
		})
	})
}

func itValidatesPresenceOfNetInfo(lrp *models.ActualLRP) {
	Context("when net info is set", func() {
		BeforeEach(func() {
			lrp.ActualLRPNetInfo = models.NewActualLRPNetInfo("1.2.3.4", "2.2.2.2", models.ActualLRPNetInfo_PreferredAddressUnknown)
		})

		It("validate does not return an error", func() {
			Expect(lrp.Validate()).NotTo(HaveOccurred())
		})
	})

	Context("when net info is not set", func() {
		BeforeEach(func() {
			lrp.ActualLRPNetInfo = models.ActualLRPNetInfo{}
		})

		It("validate returns an error", func() {
			err := lrp.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("address"))
		})
	})
}

func itValidatesAbsenceOfNetInfo(lrp *models.ActualLRP) {
	Context("when net info is set", func() {
		BeforeEach(func() {
			lrp.ActualLRPNetInfo = models.NewActualLRPNetInfo("1.2.3.4", "2.2.2.2", models.ActualLRPNetInfo_PreferredAddressUnknown)
		})

		It("validate returns an error", func() {
			err := lrp.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("net info"))
		})
	})

	Context("when net info is not set", func() {
		BeforeEach(func() {
			lrp.ActualLRPNetInfo = models.ActualLRPNetInfo{}
		})

		It("validate does not return an error", func() {
			Expect(lrp.Validate()).NotTo(HaveOccurred())
		})
	})
}

func itValidatesPresenceOfPlacementError(lrp *models.ActualLRP) {
	Context("when placement error is set", func() {
		BeforeEach(func() {
			lrp.PlacementError = "insufficient capacity"
		})

		It("validate does not return an error", func() {
			Expect(lrp.Validate()).NotTo(HaveOccurred())
		})
	})

	Context("when placement error is not set", func() {
		BeforeEach(func() {
			lrp.PlacementError = ""
		})

		It("validate does not return an error", func() {
			Expect(lrp.Validate()).NotTo(HaveOccurred())
		})
	})
}

func itValidatesAbsenceOfPlacementError(lrp *models.ActualLRP) {
	Context("when placement error is set", func() {
		BeforeEach(func() {
			lrp.PlacementError = "insufficient capacity"
		})

		It("validate returns an error", func() {
			err := lrp.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("placement error"))
		})
	})

	Context("when placement error is not set", func() {
		BeforeEach(func() {
			lrp.PlacementError = ""
		})

		It("validate does not return an error", func() {
			Expect(lrp.Validate()).NotTo(HaveOccurred())
		})
	})
}

func itValidatesOrdinaryPresence(lrp *models.ActualLRP) {
	Context("when presence is set", func() {
		BeforeEach(func() {
			lrp.Presence = models.ActualLRP_Evacuating
		})

		It("validate returns an error", func() {
			err := lrp.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("presence cannot be set"))
		})
	})

	Context("when presence is not set", func() {
		BeforeEach(func() {
			lrp.Presence = models.ActualLRP_Ordinary
		})

		It("validate does not return an error", func() {
			Expect(lrp.Validate()).NotTo(HaveOccurred())
		})
	})
}
