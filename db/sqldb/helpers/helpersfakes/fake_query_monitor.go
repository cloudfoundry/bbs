// Code generated by counterfeiter. DO NOT EDIT.
package helpersfakes

import (
	"sync"
	"time"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
)

type FakeQueryMonitor struct {
	MonitorQueryStub        func(func() error) error
	monitorQueryMutex       sync.RWMutex
	monitorQueryArgsForCall []struct {
		arg1 func() error
	}
	monitorQueryReturns struct {
		result1 error
	}
	monitorQueryReturnsOnCall map[int]struct {
		result1 error
	}
	QueriesTotalStub        func() int64
	queriesTotalMutex       sync.RWMutex
	queriesTotalArgsForCall []struct{}
	queriesTotalReturns     struct {
		result1 int64
	}
	queriesTotalReturnsOnCall map[int]struct {
		result1 int64
	}
	QueriesSucceededStub        func() int64
	queriesSucceededMutex       sync.RWMutex
	queriesSucceededArgsForCall []struct{}
	queriesSucceededReturns     struct {
		result1 int64
	}
	queriesSucceededReturnsOnCall map[int]struct {
		result1 int64
	}
	QueriesFailedStub        func() int64
	queriesFailedMutex       sync.RWMutex
	queriesFailedArgsForCall []struct{}
	queriesFailedReturns     struct {
		result1 int64
	}
	queriesFailedReturnsOnCall map[int]struct {
		result1 int64
	}
	ReadAndResetQueryDurationMaxStub        func() time.Duration
	readAndResetQueryDurationMaxMutex       sync.RWMutex
	readAndResetQueryDurationMaxArgsForCall []struct{}
	readAndResetQueryDurationMaxReturns     struct {
		result1 time.Duration
	}
	readAndResetQueryDurationMaxReturnsOnCall map[int]struct {
		result1 time.Duration
	}
	QueriesInFlightStub        func() int64
	queriesInFlightMutex       sync.RWMutex
	queriesInFlightArgsForCall []struct{}
	queriesInFlightReturns     struct {
		result1 int64
	}
	queriesInFlightReturnsOnCall map[int]struct {
		result1 int64
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeQueryMonitor) MonitorQuery(arg1 func() error) error {
	fake.monitorQueryMutex.Lock()
	ret, specificReturn := fake.monitorQueryReturnsOnCall[len(fake.monitorQueryArgsForCall)]
	fake.monitorQueryArgsForCall = append(fake.monitorQueryArgsForCall, struct {
		arg1 func() error
	}{arg1})
	fake.recordInvocation("MonitorQuery", []interface{}{arg1})
	fake.monitorQueryMutex.Unlock()
	if fake.MonitorQueryStub != nil {
		return fake.MonitorQueryStub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.monitorQueryReturns.result1
}

func (fake *FakeQueryMonitor) MonitorQueryCallCount() int {
	fake.monitorQueryMutex.RLock()
	defer fake.monitorQueryMutex.RUnlock()
	return len(fake.monitorQueryArgsForCall)
}

func (fake *FakeQueryMonitor) MonitorQueryArgsForCall(i int) func() error {
	fake.monitorQueryMutex.RLock()
	defer fake.monitorQueryMutex.RUnlock()
	return fake.monitorQueryArgsForCall[i].arg1
}

func (fake *FakeQueryMonitor) MonitorQueryReturns(result1 error) {
	fake.MonitorQueryStub = nil
	fake.monitorQueryReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeQueryMonitor) MonitorQueryReturnsOnCall(i int, result1 error) {
	fake.MonitorQueryStub = nil
	if fake.monitorQueryReturnsOnCall == nil {
		fake.monitorQueryReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.monitorQueryReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeQueryMonitor) QueriesTotal() int64 {
	fake.queriesTotalMutex.Lock()
	ret, specificReturn := fake.queriesTotalReturnsOnCall[len(fake.queriesTotalArgsForCall)]
	fake.queriesTotalArgsForCall = append(fake.queriesTotalArgsForCall, struct{}{})
	fake.recordInvocation("QueriesTotal", []interface{}{})
	fake.queriesTotalMutex.Unlock()
	if fake.QueriesTotalStub != nil {
		return fake.QueriesTotalStub()
	}
	if specificReturn {
		return ret.result1
	}
	return fake.queriesTotalReturns.result1
}

func (fake *FakeQueryMonitor) QueriesTotalCallCount() int {
	fake.queriesTotalMutex.RLock()
	defer fake.queriesTotalMutex.RUnlock()
	return len(fake.queriesTotalArgsForCall)
}

func (fake *FakeQueryMonitor) QueriesTotalReturns(result1 int64) {
	fake.QueriesTotalStub = nil
	fake.queriesTotalReturns = struct {
		result1 int64
	}{result1}
}

func (fake *FakeQueryMonitor) QueriesTotalReturnsOnCall(i int, result1 int64) {
	fake.QueriesTotalStub = nil
	if fake.queriesTotalReturnsOnCall == nil {
		fake.queriesTotalReturnsOnCall = make(map[int]struct {
			result1 int64
		})
	}
	fake.queriesTotalReturnsOnCall[i] = struct {
		result1 int64
	}{result1}
}

func (fake *FakeQueryMonitor) QueriesSucceeded() int64 {
	fake.queriesSucceededMutex.Lock()
	ret, specificReturn := fake.queriesSucceededReturnsOnCall[len(fake.queriesSucceededArgsForCall)]
	fake.queriesSucceededArgsForCall = append(fake.queriesSucceededArgsForCall, struct{}{})
	fake.recordInvocation("QueriesSucceeded", []interface{}{})
	fake.queriesSucceededMutex.Unlock()
	if fake.QueriesSucceededStub != nil {
		return fake.QueriesSucceededStub()
	}
	if specificReturn {
		return ret.result1
	}
	return fake.queriesSucceededReturns.result1
}

func (fake *FakeQueryMonitor) QueriesSucceededCallCount() int {
	fake.queriesSucceededMutex.RLock()
	defer fake.queriesSucceededMutex.RUnlock()
	return len(fake.queriesSucceededArgsForCall)
}

func (fake *FakeQueryMonitor) QueriesSucceededReturns(result1 int64) {
	fake.QueriesSucceededStub = nil
	fake.queriesSucceededReturns = struct {
		result1 int64
	}{result1}
}

func (fake *FakeQueryMonitor) QueriesSucceededReturnsOnCall(i int, result1 int64) {
	fake.QueriesSucceededStub = nil
	if fake.queriesSucceededReturnsOnCall == nil {
		fake.queriesSucceededReturnsOnCall = make(map[int]struct {
			result1 int64
		})
	}
	fake.queriesSucceededReturnsOnCall[i] = struct {
		result1 int64
	}{result1}
}

func (fake *FakeQueryMonitor) QueriesFailed() int64 {
	fake.queriesFailedMutex.Lock()
	ret, specificReturn := fake.queriesFailedReturnsOnCall[len(fake.queriesFailedArgsForCall)]
	fake.queriesFailedArgsForCall = append(fake.queriesFailedArgsForCall, struct{}{})
	fake.recordInvocation("QueriesFailed", []interface{}{})
	fake.queriesFailedMutex.Unlock()
	if fake.QueriesFailedStub != nil {
		return fake.QueriesFailedStub()
	}
	if specificReturn {
		return ret.result1
	}
	return fake.queriesFailedReturns.result1
}

func (fake *FakeQueryMonitor) QueriesFailedCallCount() int {
	fake.queriesFailedMutex.RLock()
	defer fake.queriesFailedMutex.RUnlock()
	return len(fake.queriesFailedArgsForCall)
}

func (fake *FakeQueryMonitor) QueriesFailedReturns(result1 int64) {
	fake.QueriesFailedStub = nil
	fake.queriesFailedReturns = struct {
		result1 int64
	}{result1}
}

func (fake *FakeQueryMonitor) QueriesFailedReturnsOnCall(i int, result1 int64) {
	fake.QueriesFailedStub = nil
	if fake.queriesFailedReturnsOnCall == nil {
		fake.queriesFailedReturnsOnCall = make(map[int]struct {
			result1 int64
		})
	}
	fake.queriesFailedReturnsOnCall[i] = struct {
		result1 int64
	}{result1}
}

func (fake *FakeQueryMonitor) ReadAndResetQueryDurationMax() time.Duration {
	fake.readAndResetQueryDurationMaxMutex.Lock()
	ret, specificReturn := fake.readAndResetQueryDurationMaxReturnsOnCall[len(fake.readAndResetQueryDurationMaxArgsForCall)]
	fake.readAndResetQueryDurationMaxArgsForCall = append(fake.readAndResetQueryDurationMaxArgsForCall, struct{}{})
	fake.recordInvocation("ReadAndResetQueryDurationMax", []interface{}{})
	fake.readAndResetQueryDurationMaxMutex.Unlock()
	if fake.ReadAndResetQueryDurationMaxStub != nil {
		return fake.ReadAndResetQueryDurationMaxStub()
	}
	if specificReturn {
		return ret.result1
	}
	return fake.readAndResetQueryDurationMaxReturns.result1
}

func (fake *FakeQueryMonitor) ReadAndResetQueryDurationMaxCallCount() int {
	fake.readAndResetQueryDurationMaxMutex.RLock()
	defer fake.readAndResetQueryDurationMaxMutex.RUnlock()
	return len(fake.readAndResetQueryDurationMaxArgsForCall)
}

func (fake *FakeQueryMonitor) ReadAndResetQueryDurationMaxReturns(result1 time.Duration) {
	fake.ReadAndResetQueryDurationMaxStub = nil
	fake.readAndResetQueryDurationMaxReturns = struct {
		result1 time.Duration
	}{result1}
}

func (fake *FakeQueryMonitor) ReadAndResetQueryDurationMaxReturnsOnCall(i int, result1 time.Duration) {
	fake.ReadAndResetQueryDurationMaxStub = nil
	if fake.readAndResetQueryDurationMaxReturnsOnCall == nil {
		fake.readAndResetQueryDurationMaxReturnsOnCall = make(map[int]struct {
			result1 time.Duration
		})
	}
	fake.readAndResetQueryDurationMaxReturnsOnCall[i] = struct {
		result1 time.Duration
	}{result1}
}

func (fake *FakeQueryMonitor) QueriesInFlight() int64 {
	fake.queriesInFlightMutex.Lock()
	ret, specificReturn := fake.queriesInFlightReturnsOnCall[len(fake.queriesInFlightArgsForCall)]
	fake.queriesInFlightArgsForCall = append(fake.queriesInFlightArgsForCall, struct{}{})
	fake.recordInvocation("QueriesInFlight", []interface{}{})
	fake.queriesInFlightMutex.Unlock()
	if fake.QueriesInFlightStub != nil {
		return fake.QueriesInFlightStub()
	}
	if specificReturn {
		return ret.result1
	}
	return fake.queriesInFlightReturns.result1
}

func (fake *FakeQueryMonitor) QueriesInFlightCallCount() int {
	fake.queriesInFlightMutex.RLock()
	defer fake.queriesInFlightMutex.RUnlock()
	return len(fake.queriesInFlightArgsForCall)
}

func (fake *FakeQueryMonitor) QueriesInFlightReturns(result1 int64) {
	fake.QueriesInFlightStub = nil
	fake.queriesInFlightReturns = struct {
		result1 int64
	}{result1}
}

func (fake *FakeQueryMonitor) QueriesInFlightReturnsOnCall(i int, result1 int64) {
	fake.QueriesInFlightStub = nil
	if fake.queriesInFlightReturnsOnCall == nil {
		fake.queriesInFlightReturnsOnCall = make(map[int]struct {
			result1 int64
		})
	}
	fake.queriesInFlightReturnsOnCall[i] = struct {
		result1 int64
	}{result1}
}

func (fake *FakeQueryMonitor) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.monitorQueryMutex.RLock()
	defer fake.monitorQueryMutex.RUnlock()
	fake.queriesTotalMutex.RLock()
	defer fake.queriesTotalMutex.RUnlock()
	fake.queriesSucceededMutex.RLock()
	defer fake.queriesSucceededMutex.RUnlock()
	fake.queriesFailedMutex.RLock()
	defer fake.queriesFailedMutex.RUnlock()
	fake.readAndResetQueryDurationMaxMutex.RLock()
	defer fake.readAndResetQueryDurationMaxMutex.RUnlock()
	fake.queriesInFlightMutex.RLock()
	defer fake.queriesInFlightMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeQueryMonitor) recordInvocation(key string, args []interface{}) {
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

var _ helpers.QueryMonitor = new(FakeQueryMonitor)
