package etcd_test

import (
	"errors"
	"fmt"

	etcddb "github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/coreos/go-etcd/etcd"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Evacuation", func() {
	Describe("Tabular tests", func() {
		claimedTest := func(base evacuationTest) evacuationTest {
			request := models.EvacuateClaimedActualLRPRequest{
				ActualLrpKey:         &lrpKey,
				ActualLrpInstanceKey: &alphaInstanceKey,
			}
			return evacuationTest{
				Name: base.Name,
				Subject: func() (bool, error) {
					return etcdDB.EvacuateClaimedActualLRP(logger, &request)
				},
				InstanceLRP:   base.InstanceLRP,
				EvacuatingLRP: base.EvacuatingLRP,
				Result:        base.Result,
			}
		}

		runningTest := func(base evacuationTest) evacuationTest {
			request := models.EvacuateRunningActualLRPRequest{
				ActualLrpKey:         &lrpKey,
				ActualLrpInstanceKey: &alphaInstanceKey,
				ActualLrpNetInfo:     &alphaNetInfo,
				Ttl:                  alphaEvacuationTTL,
			}
			return evacuationTest{
				Name: base.Name,
				Subject: func() (bool, error) {
					return etcdDB.EvacuateRunningActualLRP(logger, &request)
				},
				InstanceLRP:   base.InstanceLRP,
				EvacuatingLRP: base.EvacuatingLRP,
				Result:        base.Result,
			}
		}

		stoppedTest := func(base evacuationTest) evacuationTest {
			request := models.EvacuateStoppedActualLRPRequest{
				ActualLrpKey:         &lrpKey,
				ActualLrpInstanceKey: &alphaInstanceKey,
			}
			return evacuationTest{
				Name: base.Name,
				Subject: func() (bool, error) {
					return etcdDB.EvacuateStoppedActualLRP(logger, &request)
				},
				InstanceLRP:   base.InstanceLRP,
				EvacuatingLRP: base.EvacuatingLRP,
				Result:        base.Result,
			}
		}

		crashedTest := func(base evacuationTest) evacuationTest {
			request := models.EvacuateCrashedActualLRPRequest{
				ActualLrpKey:         &lrpKey,
				ActualLrpInstanceKey: &alphaInstanceKey,
				ErrorMessage:         "crashed",
			}
			return evacuationTest{
				Name: base.Name,
				Subject: func() (bool, error) {
					return etcdDB.EvacuateCrashedActualLRP(logger, &request)
				},
				InstanceLRP:   base.InstanceLRP,
				EvacuatingLRP: base.EvacuatingLRP,
				Result:        base.Result,
			}
		}

		removalTest := func(base evacuationTest) evacuationTest {
			request := models.RemoveEvacuatingActualLRPRequest{
				ActualLrpKey:         &lrpKey,
				ActualLrpInstanceKey: &alphaInstanceKey,
			}
			return evacuationTest{
				Name: base.Name,
				Subject: func() (bool, error) {
					err := etcdDB.RemoveEvacuatingActualLRP(logger, &request)
					return false, err
				},
				InstanceLRP:   base.InstanceLRP,
				EvacuatingLRP: base.EvacuatingLRP,
				Result:        base.Result,
			}
		}

		claimedTests := []testable{
			claimedTest(evacuationTest{
				Name:   "when there is no instance or evacuating LRP",
				Result: noInstanceNoEvacuating(false, nil),
			}),
			claimedTest(evacuationTest{
				Name:        "when the instance is UNCLAIMED",
				InstanceLRP: unclaimedLRP(),
				Result:      instanceNoEvacuating(anUnchangedUnclaimedInstanceLRP(), false, nil),
			}),
			claimedTest(evacuationTest{
				Name:        "when the instance is CLAIMED on alpha",
				InstanceLRP: claimedLRP(alphaInstanceKey),
				Result:      instanceNoEvacuating(anUpdatedUnclaimedInstanceLRP(), false, nil),
			}),
			claimedTest(evacuationTest{
				Name:        "when the instance is CLAIMED on omega",
				InstanceLRP: claimedLRP(omegaInstanceKey),
				Result: instanceNoEvacuating(
					anUnchangedClaimedInstanceLRP(omegaInstanceKey),
					false,
					models.ErrActualLRPCannotBeUnclaimed,
				),
			}),
			claimedTest(evacuationTest{
				Name:        "when the instance is RUNNING on alpha",
				InstanceLRP: runningLRP(alphaInstanceKey, alphaNetInfo),
				Result:      instanceNoEvacuating(anUpdatedUnclaimedInstanceLRP(), false, nil),
			}),
			claimedTest(evacuationTest{
				Name:        "when the instance is RUNNING on omega",
				InstanceLRP: runningLRP(omegaInstanceKey, omegaNetInfo),
				Result: instanceNoEvacuating(
					anUnchangedRunningInstanceLRP(omegaInstanceKey, omegaNetInfo),
					false,
					models.ErrActualLRPCannotBeUnclaimed,
				),
			}),
			claimedTest(evacuationTest{
				Name:        "when the instance is CRASHED",
				InstanceLRP: crashedLRP(),
				Result: instanceNoEvacuating(
					anUnchangedCrashedInstanceLRP(),
					false,
					models.ErrActualLRPCannotBeUnclaimed,
				),
			}),
			claimedTest(evacuationTest{
				Name:          "when the evacuating LRP is RUNNING on alpha",
				EvacuatingLRP: runningLRP(alphaInstanceKey, alphaNetInfo),
				Result:        noInstanceNoEvacuating(false, nil),
			}),
			claimedTest(evacuationTest{
				Name:          "when the evacuating LRP is RUNNING on beta",
				EvacuatingLRP: runningLRP(betaInstanceKey, betaNetInfo),
				Result:        evacuatingNoInstance(anUnchangedBetaEvacuatingLRP(), false, nil),
			}),
		}

		runningTests := []testable{
			runningTest(evacuationTest{
				Name:        "when the instance is UNCLAIMED and there is no evacuating LRP",
				InstanceLRP: unclaimedLRP(),
				Result:      newTestResult(anUnchangedUnclaimedInstanceLRP(), anUpdatedAlphaEvacuatingLRP(), true, nil),
			}),
			runningTest(evacuationTest{
				Name:        "when the instance is UNCLAIMED with a placement error and there is no evacuating LRP",
				InstanceLRP: unclaimedLRPWithPlacementError(),
				Result: instanceNoEvacuating(
					anUnchangedUnclaimedInstanceLRP(),
					false,
					nil,
				),
			}),
			runningTest(evacuationTest{
				Name:          "when the instance is UNCLAIMED and an evacuating LRP is RUNNING on alpha",
				InstanceLRP:   unclaimedLRP(),
				EvacuatingLRP: runningLRP(alphaInstanceKey, alphaNetInfo),
				Result:        newTestResult(anUnchangedUnclaimedInstanceLRP(), anUnchangedAlphaEvacuatingLRP(), true, nil),
			}),
			runningTest(evacuationTest{
				Name:          "when the instance is UNCLAIMED with a placement error and an evacuating LRP is RUNNING on alpha",
				InstanceLRP:   unclaimedLRPWithPlacementError(),
				EvacuatingLRP: runningLRP(alphaInstanceKey, alphaNetInfo),
				Result: instanceNoEvacuating(
					anUnchangedUnclaimedInstanceLRP(),
					false,
					nil,
				),
			}),
			runningTest(evacuationTest{
				Name:          "when the instance is UNCLAIMED and an evacuating LRP is RUNNING on alpha with out-of-date net info",
				InstanceLRP:   unclaimedLRP(),
				EvacuatingLRP: runningLRP(alphaInstanceKey, betaNetInfo),
				Result:        newTestResult(anUnchangedUnclaimedInstanceLRP(), anUpdatedAlphaEvacuatingLRP(), true, nil),
			}),
			runningTest(evacuationTest{
				Name:          "when the instance is UNCLAIMED and an evacuating LRP is RUNNING on beta",
				InstanceLRP:   unclaimedLRP(),
				EvacuatingLRP: runningLRP(betaInstanceKey, betaNetInfo),
				Result: newTestResult(
					anUnchangedUnclaimedInstanceLRP(),
					anUnchangedBetaEvacuatingLRP(),
					false,
					nil,
				),
			}),
			runningTest(evacuationTest{
				Name:        "when the instance is CLAIMED on alpha and there is no evacuating LRP",
				InstanceLRP: claimedLRP(alphaInstanceKey),
				Result:      newTestResult(anUpdatedUnclaimedInstanceLRP(), anUpdatedAlphaEvacuatingLRP(), true, nil),
			}),
			runningTest(evacuationTest{
				Name:          "when the instance is CLAIMED on alpha and an evacuating LRP is RUNNING on alpha",
				InstanceLRP:   claimedLRP(alphaInstanceKey),
				EvacuatingLRP: runningLRP(alphaInstanceKey, alphaNetInfo),
				Result:        newTestResult(anUpdatedUnclaimedInstanceLRP(), anUnchangedAlphaEvacuatingLRP(), true, nil),
			}),
			runningTest(evacuationTest{
				Name:          "when the instance is CLAIMED on alpha and an evacuating LRP is RUNNING on alpha with out-of-date net info",
				InstanceLRP:   claimedLRP(alphaInstanceKey),
				EvacuatingLRP: runningLRP(alphaInstanceKey, betaNetInfo),
				Result:        newTestResult(anUpdatedUnclaimedInstanceLRP(), anUpdatedAlphaEvacuatingLRP(), true, nil),
			}),
			runningTest(evacuationTest{
				Name:          "when the instance is CLAIMED on alpha and an evacuating LRP is RUNNING on beta",
				InstanceLRP:   claimedLRP(alphaInstanceKey),
				EvacuatingLRP: runningLRP(betaInstanceKey, betaNetInfo),
				Result:        newTestResult(anUpdatedUnclaimedInstanceLRP(), anUpdatedAlphaEvacuatingLRP(), true, nil),
			}),
			runningTest(evacuationTest{
				Name:        "when the instance is CLAIMED remotely and there is no evacuating LRP",
				InstanceLRP: claimedLRP(omegaInstanceKey),
				Result:      newTestResult(anUnchangedClaimedInstanceLRP(omegaInstanceKey), anUpdatedAlphaEvacuatingLRP(), true, nil),
			}),
			runningTest(evacuationTest{
				Name:          "when the instance is CLAIMED remotely and an evacuating LRP is RUNNING on alpha",
				InstanceLRP:   claimedLRP(omegaInstanceKey),
				EvacuatingLRP: runningLRP(alphaInstanceKey, alphaNetInfo),
				Result:        newTestResult(anUnchangedClaimedInstanceLRP(omegaInstanceKey), anUnchangedAlphaEvacuatingLRP(), true, nil),
			}),
			runningTest(evacuationTest{
				Name:          "when the instance is CLAIMED remotely and an evacuating LRP is RUNNING on alpha with out-of-date net info",
				InstanceLRP:   claimedLRP(omegaInstanceKey),
				EvacuatingLRP: runningLRP(alphaInstanceKey, betaNetInfo),
				Result:        newTestResult(anUnchangedClaimedInstanceLRP(omegaInstanceKey), anUpdatedAlphaEvacuatingLRP(), true, nil),
			}),
			runningTest(evacuationTest{
				Name:          "when the instance is CLAIMED remotely and an evacuating LRP is RUNNING on beta",
				InstanceLRP:   claimedLRP(omegaInstanceKey),
				EvacuatingLRP: runningLRP(betaInstanceKey, betaNetInfo),
				Result: newTestResult(
					anUnchangedClaimedInstanceLRP(omegaInstanceKey),
					anUnchangedBetaEvacuatingLRP(),
					false,
					nil,
				),
			}),
			runningTest(evacuationTest{
				Name:        "when the instance is RUNNING on alpha and there is no evacuating LRP",
				InstanceLRP: runningLRP(alphaInstanceKey, alphaNetInfo),
				Result:      newTestResult(anUpdatedUnclaimedInstanceLRP(), anUpdatedAlphaEvacuatingLRP(), true, nil),
			}),
			runningTest(evacuationTest{
				Name:          "when the instance is RUNNING on alpha and an evacuating LRP is RUNNING on alpha",
				InstanceLRP:   runningLRP(alphaInstanceKey, alphaNetInfo),
				EvacuatingLRP: runningLRP(alphaInstanceKey, alphaNetInfo),
				Result:        newTestResult(anUpdatedUnclaimedInstanceLRP(), anUnchangedAlphaEvacuatingLRP(), true, nil),
			}),
			runningTest(evacuationTest{
				Name:          "when the instance is RUNNING on alpha and an evacuating LRP is RUNNING on alpha with out-of-date net info",
				InstanceLRP:   runningLRP(alphaInstanceKey, alphaNetInfo),
				EvacuatingLRP: runningLRP(alphaInstanceKey, betaNetInfo),
				Result:        newTestResult(anUpdatedUnclaimedInstanceLRP(), anUpdatedAlphaEvacuatingLRP(), true, nil),
			}),
			runningTest(evacuationTest{
				Name:          "when the instance is RUNNING on alpha and an evacuating LRP is RUNNING on beta",
				InstanceLRP:   runningLRP(alphaInstanceKey, alphaNetInfo),
				EvacuatingLRP: runningLRP(betaInstanceKey, betaNetInfo),
				Result:        newTestResult(anUpdatedUnclaimedInstanceLRP(), anUpdatedAlphaEvacuatingLRP(), true, nil),
			}),
			runningTest(evacuationTest{
				Name:        "when the instance is RUNNING on omega and there is no evacuating LRP",
				InstanceLRP: runningLRP(omegaInstanceKey, omegaNetInfo),
				Result: instanceNoEvacuating(
					anUnchangedRunningInstanceLRP(omegaInstanceKey, omegaNetInfo),
					false,
					nil,
				),
			}),
			runningTest(evacuationTest{
				Name:          "when the instance is RUNNING on omega and an evacuating LRP is RUNNING on alpha",
				InstanceLRP:   runningLRP(omegaInstanceKey, omegaNetInfo),
				EvacuatingLRP: runningLRP(alphaInstanceKey, alphaNetInfo),
				Result: instanceNoEvacuating(
					anUnchangedRunningInstanceLRP(omegaInstanceKey, omegaNetInfo),
					false,
					nil,
				),
			}),
			runningTest(evacuationTest{
				Name:          "when the instance is RUNNING on omega and an evacuating LRP is RUNNING on beta",
				InstanceLRP:   runningLRP(omegaInstanceKey, omegaNetInfo),
				EvacuatingLRP: runningLRP(betaInstanceKey, betaNetInfo),
				Result: newTestResult(
					anUnchangedRunningInstanceLRP(omegaInstanceKey, omegaNetInfo),
					anUnchangedBetaEvacuatingLRP(),
					false,
					nil,
				),
			}),
			runningTest(evacuationTest{
				Name:        "when the instance is CRASHED and there is no evacuating LRP",
				InstanceLRP: crashedLRP(),
				Result: instanceNoEvacuating(
					anUnchangedCrashedInstanceLRP(),
					false,
					nil,
				),
			}),
			runningTest(evacuationTest{
				Name:          "when the instance is CRASHED and an evacuating LRP is RUNNING on alpha",
				InstanceLRP:   crashedLRP(),
				EvacuatingLRP: runningLRP(alphaInstanceKey, alphaNetInfo),
				Result: instanceNoEvacuating(
					anUnchangedCrashedInstanceLRP(),
					false,
					nil,
				),
			}),
			runningTest(evacuationTest{
				Name:          "when the instance is CRASHED and an evacuating LRP is RUNNING on beta",
				InstanceLRP:   crashedLRP(),
				EvacuatingLRP: runningLRP(betaInstanceKey, betaNetInfo),
				Result: newTestResult(
					anUnchangedCrashedInstanceLRP(),
					anUnchangedBetaEvacuatingLRP(),
					false,
					nil,
				),
			}),
			runningTest(evacuationTest{
				Name: "when the instance is MISSING and there is no evacuating LRP",
				Result: noInstanceNoEvacuating(
					false,
					nil,
				),
			}),
			runningTest(evacuationTest{
				Name:          "when the instance is MISSING and an evacuating LRP is RUNNING on alpha",
				EvacuatingLRP: runningLRP(alphaInstanceKey, alphaNetInfo),
				Result: noInstanceNoEvacuating(
					false,
					nil,
				),
			}),
			runningTest(evacuationTest{
				Name:          "when the instance is MISSING and an evacuating LRP is RUNNING on beta",
				EvacuatingLRP: runningLRP(betaInstanceKey, betaNetInfo),
				Result: evacuatingNoInstance(
					anUnchangedBetaEvacuatingLRP(),
					false,
					nil,
				),
			}),
		}

		stoppedTests := []testable{
			stoppedTest(evacuationTest{
				Name:   "when there is no instance or evacuating LRP",
				Result: noInstanceNoEvacuating(false, nil),
			}),
			stoppedTest(evacuationTest{
				Name:        "when the instance is UNCLAIMED",
				InstanceLRP: unclaimedLRP(),
				Result: instanceNoEvacuating(
					anUnchangedUnclaimedInstanceLRP(),
					false,
					models.ErrActualLRPCannotBeRemoved,
				),
			}),
			stoppedTest(evacuationTest{
				Name:        "when the instance is CLAIMED on alpha",
				InstanceLRP: claimedLRP(alphaInstanceKey),
				Result:      noInstanceNoEvacuating(false, nil),
			}),
			stoppedTest(evacuationTest{
				Name:        "when the instance is CLAIMED on omega",
				InstanceLRP: claimedLRP(omegaInstanceKey),
				Result: instanceNoEvacuating(
					anUnchangedClaimedInstanceLRP(omegaInstanceKey),
					false,
					models.ErrActualLRPCannotBeRemoved,
				),
			}),
			stoppedTest(evacuationTest{
				Name:        "when the instance is RUNNING on alpha",
				InstanceLRP: runningLRP(alphaInstanceKey, alphaNetInfo),
				Result:      noInstanceNoEvacuating(false, nil),
			}),
			stoppedTest(evacuationTest{
				Name:        "when the instance is RUNNING on omega",
				InstanceLRP: runningLRP(omegaInstanceKey, omegaNetInfo),
				Result: instanceNoEvacuating(
					anUnchangedRunningInstanceLRP(omegaInstanceKey, omegaNetInfo),
					false,
					models.ErrActualLRPCannotBeRemoved,
				),
			}),
			stoppedTest(evacuationTest{
				Name:        "when the instance is CRASHED",
				InstanceLRP: crashedLRP(),
				Result: instanceNoEvacuating(
					anUnchangedCrashedInstanceLRP(),
					false,
					models.ErrActualLRPCannotBeRemoved,
				),
			}),
			stoppedTest(evacuationTest{
				Name:          "when the evacuating LRP is RUNNING on alpha",
				EvacuatingLRP: runningLRP(alphaInstanceKey, alphaNetInfo),
				Result:        noInstanceNoEvacuating(false, nil),
			}),
			stoppedTest(evacuationTest{
				Name:          "when the evacuating LRP is RUNNING on beta",
				EvacuatingLRP: runningLRP(betaInstanceKey, betaNetInfo),
				Result:        evacuatingNoInstance(anUnchangedBetaEvacuatingLRP(), false, nil),
			}),
		}

		crashedTests := []testable{
			crashedTest(evacuationTest{
				Name:   "when there is no instance or evacuating LRP",
				Result: noInstanceNoEvacuating(false, nil),
			}),
			crashedTest(evacuationTest{
				Name:        "when the instance is UNCLAIMED",
				InstanceLRP: unclaimedLRP(),
				Result: instanceNoEvacuating(
					anUnchangedUnclaimedInstanceLRP(),
					false,
					models.ErrActualLRPCannotBeCrashed,
				),
			}),
			crashedTest(evacuationTest{
				Name:        "when the instance is CLAIMED on alpha",
				InstanceLRP: claimedLRP(alphaInstanceKey),
				Result:      instanceNoEvacuating(anUpdatedUnclaimedInstanceLRPWithCrashCount(1), false, nil),
			}),
			crashedTest(evacuationTest{
				Name:        "when the instance is CLAIMED on omega",
				InstanceLRP: claimedLRP(omegaInstanceKey),
				Result: instanceNoEvacuating(
					anUnchangedClaimedInstanceLRP(omegaInstanceKey),
					false,
					models.ErrActualLRPCannotBeCrashed,
				),
			}),
			crashedTest(evacuationTest{
				Name:        "when the instance is RUNNING on alpha",
				InstanceLRP: runningLRP(alphaInstanceKey, alphaNetInfo),
				Result:      instanceNoEvacuating(anUpdatedUnclaimedInstanceLRPWithCrashCount(1), false, nil),
			}),
			crashedTest(evacuationTest{
				Name:        "when the instance is RUNNING on omega",
				InstanceLRP: runningLRP(omegaInstanceKey, omegaNetInfo),
				Result: instanceNoEvacuating(
					anUnchangedRunningInstanceLRP(omegaInstanceKey, omegaNetInfo),
					false,
					models.ErrActualLRPCannotBeCrashed,
				),
			}),
			crashedTest(evacuationTest{
				Name:        "when the instance is CRASHED",
				InstanceLRP: crashedLRP(),
				Result: instanceNoEvacuating(
					anUnchangedCrashedInstanceLRP(),
					false,
					models.ErrActualLRPCannotBeCrashed,
				),
			}),
			crashedTest(evacuationTest{
				Name:          "when the evacuating LRP is RUNNING on alpha",
				EvacuatingLRP: runningLRP(alphaInstanceKey, alphaNetInfo),
				Result:        noInstanceNoEvacuating(false, nil),
			}),
			crashedTest(evacuationTest{
				Name:          "when the evacuating LRP is RUNNING on beta",
				EvacuatingLRP: runningLRP(betaInstanceKey, betaNetInfo),
				Result:        evacuatingNoInstance(anUnchangedBetaEvacuatingLRP(), false, nil),
			}),
		}

		removalTests := []testable{
			removalTest(evacuationTest{
				Name:   "when there is no instance or evacuating LRP",
				Result: noInstanceNoEvacuating(false, nil),
			}),
			removalTest(evacuationTest{
				Name:        "when the instance is UNCLAIMED",
				InstanceLRP: unclaimedLRP(),
				Result:      instanceNoEvacuating(anUnchangedUnclaimedInstanceLRP(), false, nil),
			}),
			removalTest(evacuationTest{
				Name:        "when the instance is CLAIMED on alpha",
				InstanceLRP: claimedLRP(alphaInstanceKey),
				Result:      instanceNoEvacuating(anUnchangedClaimedInstanceLRP(alphaInstanceKey), false, nil),
			}),
			removalTest(evacuationTest{
				Name:        "when the instance is CLAIMED on omega",
				InstanceLRP: claimedLRP(omegaInstanceKey),
				Result:      instanceNoEvacuating(anUnchangedClaimedInstanceLRP(omegaInstanceKey), false, nil),
			}),
			removalTest(evacuationTest{
				Name:        "when the instance is RUNNING on alpha",
				InstanceLRP: runningLRP(alphaInstanceKey, alphaNetInfo),
				Result:      instanceNoEvacuating(anUnchangedRunningInstanceLRP(alphaInstanceKey, alphaNetInfo), false, nil),
			}),
			removalTest(evacuationTest{
				Name:        "when the instance is RUNNING on omega",
				InstanceLRP: runningLRP(omegaInstanceKey, omegaNetInfo),
				Result:      instanceNoEvacuating(anUnchangedRunningInstanceLRP(omegaInstanceKey, omegaNetInfo), false, nil),
			}),
			removalTest(evacuationTest{
				Name:        "when the instance is CRASHED",
				InstanceLRP: crashedLRP(),
				Result:      instanceNoEvacuating(anUnchangedCrashedInstanceLRP(), false, nil),
			}),
			removalTest(evacuationTest{
				Name:          "when the evacuating LRP is RUNNING on alpha",
				EvacuatingLRP: runningLRP(alphaInstanceKey, alphaNetInfo),
				Result:        noInstanceNoEvacuating(false, nil),
			}),
			removalTest(evacuationTest{
				Name:          "when the evacuating LRP is RUNNING on beta",
				EvacuatingLRP: runningLRP(betaInstanceKey, betaNetInfo),
				Result: evacuatingNoInstance(
					anUnchangedBetaEvacuatingLRP(),
					false,
					models.ErrActualLRPCannotBeRemoved,
				),
			}),
		}

		Context("when the LRP is to be CLAIMED", func() {
			for _, test := range claimedTests {
				test.Test()
			}
		})

		Context("when the LRP is to be RUNNING", func() {
			for _, test := range runningTests {
				test.Test()
			}
		})

		Context("when the LRP is to be STOPPED", func() {
			for _, test := range stoppedTests {
				test.Test()
			}
		})

		Context("when the LRP is to be CRASHED", func() {
			for _, test := range crashedTests {
				test.Test()
			}
		})

		Context("when the evacuating LRP is to be removed", func() {
			for _, test := range removalTests {
				test.Test()
			}
		})
	})
})

const (
	initialTimestamp = 1138
	timeIncrement    = 2279
	finalTimestamp   = initialTimestamp + timeIncrement

	alphaEvacuationTTL = 30
	omegaEvacuationTTL = 1000
	allowedTTLDecay    = 2

	processGuid       = "process-guid"
	alphaInstanceGuid = "alpha-instance-guid"
	betaInstanceGuid  = "beta-instance-guid"
	omegaInstanceGuid = "omega-instance-guid"
	alphaCellID       = "alpha-cell-id"
	betaCellID        = "beta-cell-id"
	omegaCellID       = "omega-cell-id"
	alphaAddress      = "alpha-address"
	betaAddress       = "beta-address"
	omegaAddress      = "omega-address"
)

var (
	desiredLRP = models.DesiredLRP{
		ProcessGuid: processGuid,
		Domain:      "domain",
		Instances:   1,
		RootFs:      "some:rootfs",
		Action:      models.WrapAction(&models.RunAction{Path: "/bin/true", User: "name"}),
	}

	index  int32 = 0
	lrpKey       = models.NewActualLRPKey(desiredLRP.ProcessGuid, index, desiredLRP.Domain)

	alphaInstanceKey = models.NewActualLRPInstanceKey(alphaInstanceGuid, alphaCellID)
	betaInstanceKey  = models.NewActualLRPInstanceKey(betaInstanceGuid, betaCellID)
	omegaInstanceKey = models.NewActualLRPInstanceKey(omegaInstanceGuid, omegaCellID)
	emptyInstanceKey = models.ActualLRPInstanceKey{}

	alphaPorts   = models.NewPortMapping(9872, 2349)
	alphaNetInfo = models.NewActualLRPNetInfo(alphaAddress, alphaPorts)
	betaPorts    = models.NewPortMapping(9868, 2353)
	betaNetInfo  = models.NewActualLRPNetInfo(betaAddress, betaPorts)
	omegaPorts   = models.NewPortMapping(9876, 2345)
	omegaNetInfo = models.NewActualLRPNetInfo(omegaAddress, omegaPorts)
	emptyNetInfo = models.EmptyActualLRPNetInfo()
)

type testable interface {
	Test()
}

type evacuationTest struct {
	Name          string
	Subject       func() (bool, error)
	InstanceLRP   lrpSetupFunc
	EvacuatingLRP lrpSetupFunc
	Result        testResult
}

func lrp(state string, instanceKey models.ActualLRPInstanceKey, netInfo models.ActualLRPNetInfo, placementError string) lrpSetupFunc {
	return func() models.ActualLRP {
		return models.ActualLRP{
			ActualLRPKey:         lrpKey,
			ActualLRPInstanceKey: instanceKey,
			ActualLRPNetInfo:     netInfo,
			State:                state,
			Since:                clock.Now().UnixNano(),
			PlacementError:       placementError,
		}
	}
}

func unclaimedLRP() lrpSetupFunc {
	return lrp(models.ActualLRPStateUnclaimed, emptyInstanceKey, emptyNetInfo, "")
}

func unclaimedLRPWithPlacementError() lrpSetupFunc {
	return lrp(models.ActualLRPStateUnclaimed, emptyInstanceKey, emptyNetInfo, "some-placement-error")
}

func claimedLRP(instanceKey models.ActualLRPInstanceKey) lrpSetupFunc {
	return lrp(models.ActualLRPStateClaimed, instanceKey, emptyNetInfo, "")
}

func runningLRP(instanceKey models.ActualLRPInstanceKey, netInfo models.ActualLRPNetInfo) lrpSetupFunc {
	return lrp(models.ActualLRPStateRunning, instanceKey, netInfo, "")
}

func crashedLRP() lrpSetupFunc {
	actualFunc := lrp(models.ActualLRPStateCrashed, emptyInstanceKey, emptyNetInfo, "")
	return func() models.ActualLRP {
		actual := actualFunc()
		actual.CrashReason = "crashed"
		return actual
	}
}

type lrpStatus struct {
	State string
	models.ActualLRPInstanceKey
	models.ActualLRPNetInfo
	ShouldUpdate bool
}

type instanceLRPStatus struct {
	lrpStatus
	CrashCount  int32
	CrashReason string
}

type evacuatingLRPStatus struct {
	lrpStatus
	TTL uint64
}

type testResult struct {
	Instance         *instanceLRPStatus
	Evacuating       *evacuatingLRPStatus
	AuctionRequested bool
	ReturnedError    error
	RetainContainer  bool
}

func anUpdatedAlphaEvacuatingLRP() *evacuatingLRPStatus {
	return newEvacuatingLRPStatus(alphaInstanceKey, alphaNetInfo, true)
}

func anUnchangedAlphaEvacuatingLRP() *evacuatingLRPStatus {
	return newEvacuatingLRPStatus(alphaInstanceKey, alphaNetInfo, false)
}

func anUnchangedBetaEvacuatingLRP() *evacuatingLRPStatus {
	return newEvacuatingLRPStatus(betaInstanceKey, betaNetInfo, false)
}

func newEvacuatingLRPStatus(instanceKey models.ActualLRPInstanceKey, netInfo models.ActualLRPNetInfo, shouldUpdate bool) *evacuatingLRPStatus {
	status := &evacuatingLRPStatus{
		lrpStatus: lrpStatus{
			State:                models.ActualLRPStateRunning,
			ActualLRPInstanceKey: instanceKey,
			ActualLRPNetInfo:     netInfo,
			ShouldUpdate:         shouldUpdate,
		},
	}

	if shouldUpdate {
		status.TTL = alphaEvacuationTTL
	}

	return status
}

func anUpdatedUnclaimedInstanceLRP() *instanceLRPStatus {
	return anUpdatedUnclaimedInstanceLRPWithCrashCount(0)
}

func anUpdatedUnclaimedInstanceLRPWithCrashCount(crashCount int32) *instanceLRPStatus {
	reason := ""
	if crashCount > 0 {
		reason = "crashed"
	}
	return &instanceLRPStatus{
		lrpStatus: lrpStatus{
			State:                models.ActualLRPStateUnclaimed,
			ActualLRPInstanceKey: emptyInstanceKey,
			ActualLRPNetInfo:     emptyNetInfo,
			ShouldUpdate:         true,
		},
		CrashCount:  crashCount,
		CrashReason: reason,
	}
}

func anUnchangedInstanceLRP(state string, instanceKey models.ActualLRPInstanceKey, netInfo models.ActualLRPNetInfo) *instanceLRPStatus {
	return &instanceLRPStatus{
		lrpStatus: lrpStatus{
			State:                state,
			ActualLRPInstanceKey: instanceKey,
			ActualLRPNetInfo:     netInfo,
			ShouldUpdate:         false,
		},
	}
}

func anUnchangedUnclaimedInstanceLRP() *instanceLRPStatus {
	return anUnchangedInstanceLRP(models.ActualLRPStateUnclaimed, emptyInstanceKey, emptyNetInfo)
}

func anUnchangedClaimedInstanceLRP(instanceKey models.ActualLRPInstanceKey) *instanceLRPStatus {
	return anUnchangedInstanceLRP(models.ActualLRPStateClaimed, instanceKey, emptyNetInfo)
}

func anUnchangedRunningInstanceLRP(instanceKey models.ActualLRPInstanceKey, netInfo models.ActualLRPNetInfo) *instanceLRPStatus {
	return anUnchangedInstanceLRP(models.ActualLRPStateRunning, instanceKey, netInfo)
}

func anUnchangedCrashedInstanceLRP() *instanceLRPStatus {
	instance := anUnchangedInstanceLRP(models.ActualLRPStateCrashed, emptyInstanceKey, emptyNetInfo)
	instance.CrashReason = "crashed"
	return instance
}

func newTestResult(instanceStatus *instanceLRPStatus, evacuatingStatus *evacuatingLRPStatus, retainContainer bool, err error) testResult {
	result := testResult{
		Instance:        instanceStatus,
		Evacuating:      evacuatingStatus,
		ReturnedError:   err,
		RetainContainer: retainContainer,
	}

	if instanceStatus != nil && instanceStatus.ShouldUpdate {
		result.AuctionRequested = true
	}

	return result
}

func instanceNoEvacuating(instanceStatus *instanceLRPStatus, retainContainer bool, err error) testResult {
	return newTestResult(instanceStatus, nil, retainContainer, err)
}

func evacuatingNoInstance(evacuatingStatus *evacuatingLRPStatus, retainContainer bool, err error) testResult {
	return newTestResult(nil, evacuatingStatus, retainContainer, err)
}

func noInstanceNoEvacuating(retainContainer bool, err error) testResult {
	return newTestResult(nil, nil, retainContainer, err)
}

func (t evacuationTest) Test() {
	Context(t.Name, func() {
		var evacuateErr error
		var initialTimestamp int64
		var initialInstanceModificationIndex uint32
		var initialEvacuatingModificationIndex uint32
		var retainContainer bool

		BeforeEach(func() {
			initialTimestamp = clock.Now().UnixNano()

			etcdHelper.SetRawDesiredLRP(&desiredLRP)
			if t.InstanceLRP != nil {
				actualLRP := t.InstanceLRP()
				initialInstanceModificationIndex = actualLRP.ModificationTag.Index
				etcdHelper.SetRawActualLRP(&actualLRP)
			}
			if t.EvacuatingLRP != nil {
				evacuatingLRP := t.EvacuatingLRP()
				initialEvacuatingModificationIndex = evacuatingLRP.ModificationTag.Index
				etcdHelper.SetRawEvacuatingActualLRP(&evacuatingLRP, omegaEvacuationTTL)
			}
		})

		JustBeforeEach(func() {
			clock.Increment(timeIncrement)
			retainContainer, evacuateErr = t.Subject()
		})

		if t.Result.ReturnedError == nil {
			It("does not return an error", func() {
				Expect(evacuateErr).NotTo(HaveOccurred())
			})
		} else {
			It(fmt.Sprintf("returned error should be '%s'", t.Result.ReturnedError.Error()), func() {
				Expect(evacuateErr).To(Equal(t.Result.ReturnedError))
			})
		}

		if t.Result.RetainContainer == true {
			It("returns true", func() {
				Expect(retainContainer).To(Equal(true))
			})

		} else {
			It("returns false", func() {
				Expect(retainContainer).To(Equal(false))
			})

		}

		if t.Result.AuctionRequested {
			It("starts an auction", func() {
				Expect(auctioneerClient.RequestLRPAuctionsCallCount()).To(Equal(1))

				requestedAuctions := auctioneerClient.RequestLRPAuctionsArgsForCall(0)
				Expect(requestedAuctions).To(HaveLen(1))

				Expect(*requestedAuctions[0].DesiredLRP).To(Equal(desiredLRP))
				Expect(requestedAuctions[0].Indices).To(ConsistOf(uint(index)))
			})

			Context("when starting the auction fails", func() {
				BeforeEach(func() {
					auctioneerClient.RequestLRPAuctionsReturns(errors.New("error"))
				})

				It("returns an UnknownError", func() {
					Expect(evacuateErr).To(Equal(models.ErrUnknownError))
				})
			})

			Context("when the desired LRP no longer exists", func() {
				BeforeEach(func() {
					_, err := etcdClient.Delete(etcddb.DesiredLRPSchemaPath(&desiredLRP), true)
					Expect(err).NotTo(HaveOccurred())
				})

				It("the actual LRP is also deleted", func() {
					Expect(evacuateErr).NotTo(HaveOccurred())

					group, err := etcdDB.ActualLRPGroupByProcessGuidAndIndex(logger, t.InstanceLRP().ProcessGuid, t.InstanceLRP().Index)
					if err == nil {
						// LRP remaining in one of evacuating or ...not
						Expect(group.Instance).To(BeNil())
					} else {
						// No LRP remaining at all (no group returned)
						Expect(err).To(Equal(models.ErrResourceNotFound))
					}
				})
			})
		} else {
			It("does not start an auction", func() {
				Expect(auctioneerClient.RequestLRPAuctionsCallCount()).To(Equal(0))
			})
		}

		if t.Result.Instance == nil {
			It("removes the /instance actualLRP", func() {
				_, err := etcdHelper.GetInstanceActualLRP(&lrpKey)
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})
		} else {
			if t.Result.Instance.ShouldUpdate {
				It("updates the /instance Since", func() {
					lrpInBBS, err := etcdHelper.GetInstanceActualLRP(&lrpKey)
					Expect(err).NotTo(HaveOccurred())

					Expect(lrpInBBS.Since).To(Equal(clock.Now().UnixNano()))
				})

				It("updates the /instance ModificationTag", func() {
					lrpInBBS, err := etcdHelper.GetInstanceActualLRP(&lrpKey)
					Expect(err).NotTo(HaveOccurred())

					Expect(lrpInBBS.ModificationTag.Index).To(Equal(initialInstanceModificationIndex + 1))
				})
			} else {
				It("does not update the /instance Since", func() {
					lrpInBBS, err := etcdHelper.GetInstanceActualLRP(&lrpKey)
					Expect(err).NotTo(HaveOccurred())

					Expect(lrpInBBS.Since).To(Equal(initialTimestamp))
				})
			}

			It("has the expected /instance state", func() {
				lrpInBBS, err := etcdHelper.GetInstanceActualLRP(&lrpKey)
				Expect(err).NotTo(HaveOccurred())

				Expect(lrpInBBS.State).To(Equal(t.Result.Instance.State))
			})

			It("has the expected /instance crash count", func() {
				lrpInBBS, err := etcdHelper.GetInstanceActualLRP(&lrpKey)
				Expect(err).NotTo(HaveOccurred())

				Expect(lrpInBBS.CrashCount).To(Equal(t.Result.Instance.CrashCount))
			})

			It("has the expected /instance crash reason", func() {
				lrpInBBS, err := etcdHelper.GetInstanceActualLRP(&lrpKey)
				Expect(err).NotTo(HaveOccurred())

				Expect(lrpInBBS.CrashReason).To(Equal(t.Result.Instance.CrashReason))
			})

			It("has the expected /instance instance key", func() {
				lrpInBBS, err := etcdHelper.GetInstanceActualLRP(&lrpKey)
				Expect(err).NotTo(HaveOccurred())

				Expect(lrpInBBS.ActualLRPInstanceKey).To(Equal(t.Result.Instance.ActualLRPInstanceKey))
			})

			It("has the expected /instance net info", func() {
				lrpInBBS, err := etcdHelper.GetInstanceActualLRP(&lrpKey)
				Expect(err).NotTo(HaveOccurred())

				Expect(lrpInBBS.ActualLRPNetInfo).To(Equal(t.Result.Instance.ActualLRPNetInfo))
			})
		}

		if t.Result.Evacuating == nil {
			It("removes the /evacuating actualLRP", func() {
				_, _, err := getEvacuatingActualLRP(lrpKey)
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})
		} else {
			if t.Result.Evacuating.ShouldUpdate {
				It("updates the /evacuating Since", func() {
					lrpInBBS, _, err := getEvacuatingActualLRP(lrpKey)
					Expect(err).NotTo(HaveOccurred())

					Expect(lrpInBBS.Since).To(Equal(clock.Now().UnixNano()))
				})

				It("updates the /evacuating TTL to the desired value", func() {
					_, ttl, err := getEvacuatingActualLRP(lrpKey)
					Expect(err).NotTo(HaveOccurred())

					Expect(ttl).To(BeNumerically("~", t.Result.Evacuating.TTL, allowedTTLDecay))
				})

				It("updates the /evacuating ModificationTag", func() {
					lrpInBBS, _, err := getEvacuatingActualLRP(lrpKey)
					Expect(err).NotTo(HaveOccurred())

					Expect(lrpInBBS.ModificationTag.Index).To(Equal(initialEvacuatingModificationIndex + 1))
				})
			} else {
				It("does not update the /evacuating Since", func() {
					lrpInBBS, _, err := getEvacuatingActualLRP(lrpKey)
					Expect(err).NotTo(HaveOccurred())

					Expect(lrpInBBS.Since).To(Equal(initialTimestamp))
				})

				It("does not update the /evacuating TTL", func() {
					_, ttl, err := getEvacuatingActualLRP(lrpKey)
					Expect(err).NotTo(HaveOccurred())

					Expect(ttl).To(BeNumerically("~", omegaEvacuationTTL, allowedTTLDecay))
				})
			}

			It("has the expected /evacuating state", func() {
				lrpInBBS, _, err := getEvacuatingActualLRP(lrpKey)
				Expect(err).NotTo(HaveOccurred())

				Expect(lrpInBBS.State).To(Equal(t.Result.Evacuating.State))
			})

			It("has the expected /evacuating instance key", func() {
				lrpInBBS, _, err := getEvacuatingActualLRP(lrpKey)
				Expect(err).NotTo(HaveOccurred())

				Expect(lrpInBBS.ActualLRPInstanceKey).To(Equal(t.Result.Evacuating.ActualLRPInstanceKey))
			})

			It("has the expected /evacuating net info", func() {
				lrpInBBS, _, err := getEvacuatingActualLRP(lrpKey)
				Expect(err).NotTo(HaveOccurred())

				Expect(lrpInBBS.ActualLRPNetInfo).To(Equal(t.Result.Evacuating.ActualLRPNetInfo))
			})
		}
	})
}

func getEvacuatingActualLRP(lrpKey models.ActualLRPKey) (models.ActualLRP, int64, error) {
	node, err := etcdClient.Get(etcddb.EvacuatingActualLRPSchemaPath(lrpKey.ProcessGuid, lrpKey.Index), false, true)
	if etcdErrCode(err) == etcddb.ETCDErrKeyNotFound {
		return models.ActualLRP{}, 0, models.ErrResourceNotFound
	}
	Expect(err).NotTo(HaveOccurred())

	var lrp models.ActualLRP
	err = models.FromJSON([]byte(node.Node.Value), &lrp)
	Expect(err).NotTo(HaveOccurred())

	return lrp, node.Node.TTL, nil
}

func etcdErrCode(err error) int {
	if err != nil {
		switch err.(type) {
		case etcd.EtcdError:
			return err.(etcd.EtcdError).ErrorCode
		case *etcd.EtcdError:
			return err.(*etcd.EtcdError).ErrorCode
		}
	}
	return 0
}
