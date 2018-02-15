// Code generated by counterfeiter. DO NOT EDIT.
package fake_controllers

import (
	"sync"

	"code.cloudfoundry.org/bbs/converger"
	"code.cloudfoundry.org/lager"
)

type FakeLrpConvergenceController struct {
	ConvergeLRPsStub        func(logger lager.Logger) error
	convergeLRPsMutex       sync.RWMutex
	convergeLRPsArgsForCall []struct {
		logger lager.Logger
	}
	convergeLRPsReturns struct {
		result1 error
	}
	convergeLRPsReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeLrpConvergenceController) ConvergeLRPs(logger lager.Logger) error {
	fake.convergeLRPsMutex.Lock()
	ret, specificReturn := fake.convergeLRPsReturnsOnCall[len(fake.convergeLRPsArgsForCall)]
	fake.convergeLRPsArgsForCall = append(fake.convergeLRPsArgsForCall, struct {
		logger lager.Logger
	}{logger})
	fake.recordInvocation("ConvergeLRPs", []interface{}{logger})
	fake.convergeLRPsMutex.Unlock()
	if fake.ConvergeLRPsStub != nil {
		return fake.ConvergeLRPsStub(logger)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.convergeLRPsReturns.result1
}

func (fake *FakeLrpConvergenceController) ConvergeLRPsCallCount() int {
	fake.convergeLRPsMutex.RLock()
	defer fake.convergeLRPsMutex.RUnlock()
	return len(fake.convergeLRPsArgsForCall)
}

func (fake *FakeLrpConvergenceController) ConvergeLRPsArgsForCall(i int) lager.Logger {
	fake.convergeLRPsMutex.RLock()
	defer fake.convergeLRPsMutex.RUnlock()
	return fake.convergeLRPsArgsForCall[i].logger
}

func (fake *FakeLrpConvergenceController) ConvergeLRPsReturns(result1 error) {
	fake.ConvergeLRPsStub = nil
	fake.convergeLRPsReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeLrpConvergenceController) ConvergeLRPsReturnsOnCall(i int, result1 error) {
	fake.ConvergeLRPsStub = nil
	if fake.convergeLRPsReturnsOnCall == nil {
		fake.convergeLRPsReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.convergeLRPsReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeLrpConvergenceController) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.convergeLRPsMutex.RLock()
	defer fake.convergeLRPsMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeLrpConvergenceController) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ converger.LrpConvergenceController = new(FakeLrpConvergenceController)
