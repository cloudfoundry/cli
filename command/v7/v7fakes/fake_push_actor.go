// Code generated by counterfeiter. DO NOT EDIT.
package v7fakes

import (
	"sync"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7pushaction"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/util/manifestparser"
)

type FakePushActor struct {
	ActualizeStub        func(v7pushaction.PushPlan, v7pushaction.ProgressBar) <-chan *v7pushaction.PushEvent
	actualizeMutex       sync.RWMutex
	actualizeArgsForCall []struct {
		arg1 v7pushaction.PushPlan
		arg2 v7pushaction.ProgressBar
	}
	actualizeReturns struct {
		result1 <-chan *v7pushaction.PushEvent
	}
	actualizeReturnsOnCall map[int]struct {
		result1 <-chan *v7pushaction.PushEvent
	}
	CreatePushPlansStub        func(string, string, manifestparser.Manifest, v7pushaction.FlagOverrides) ([]v7pushaction.PushPlan, v7action.Warnings, error)
	createPushPlansMutex       sync.RWMutex
	createPushPlansArgsForCall []struct {
		arg1 string
		arg2 string
		arg3 manifestparser.Manifest
		arg4 v7pushaction.FlagOverrides
	}
	createPushPlansReturns struct {
		result1 []v7pushaction.PushPlan
		result2 v7action.Warnings
		result3 error
	}
	createPushPlansReturnsOnCall map[int]struct {
		result1 []v7pushaction.PushPlan
		result2 v7action.Warnings
		result3 error
	}
	HandleDeploymentScaleFlagOverridesStub        func(manifestparser.Manifest, v7pushaction.FlagOverrides) (manifestparser.Manifest, error)
	handleDeploymentScaleFlagOverridesMutex       sync.RWMutex
	handleDeploymentScaleFlagOverridesArgsForCall []struct {
		arg1 manifestparser.Manifest
		arg2 v7pushaction.FlagOverrides
	}
	handleDeploymentScaleFlagOverridesReturns struct {
		result1 manifestparser.Manifest
		result2 error
	}
	handleDeploymentScaleFlagOverridesReturnsOnCall map[int]struct {
		result1 manifestparser.Manifest
		result2 error
	}
	HandleFlagOverridesStub        func(manifestparser.Manifest, v7pushaction.FlagOverrides) (manifestparser.Manifest, error)
	handleFlagOverridesMutex       sync.RWMutex
	handleFlagOverridesArgsForCall []struct {
		arg1 manifestparser.Manifest
		arg2 v7pushaction.FlagOverrides
	}
	handleFlagOverridesReturns struct {
		result1 manifestparser.Manifest
		result2 error
	}
	handleFlagOverridesReturnsOnCall map[int]struct {
		result1 manifestparser.Manifest
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakePushActor) Actualize(arg1 v7pushaction.PushPlan, arg2 v7pushaction.ProgressBar) <-chan *v7pushaction.PushEvent {
	fake.actualizeMutex.Lock()
	ret, specificReturn := fake.actualizeReturnsOnCall[len(fake.actualizeArgsForCall)]
	fake.actualizeArgsForCall = append(fake.actualizeArgsForCall, struct {
		arg1 v7pushaction.PushPlan
		arg2 v7pushaction.ProgressBar
	}{arg1, arg2})
	stub := fake.ActualizeStub
	fakeReturns := fake.actualizeReturns
	fake.recordInvocation("Actualize", []interface{}{arg1, arg2})
	fake.actualizeMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakePushActor) ActualizeCallCount() int {
	fake.actualizeMutex.RLock()
	defer fake.actualizeMutex.RUnlock()
	return len(fake.actualizeArgsForCall)
}

func (fake *FakePushActor) ActualizeCalls(stub func(v7pushaction.PushPlan, v7pushaction.ProgressBar) <-chan *v7pushaction.PushEvent) {
	fake.actualizeMutex.Lock()
	defer fake.actualizeMutex.Unlock()
	fake.ActualizeStub = stub
}

func (fake *FakePushActor) ActualizeArgsForCall(i int) (v7pushaction.PushPlan, v7pushaction.ProgressBar) {
	fake.actualizeMutex.RLock()
	defer fake.actualizeMutex.RUnlock()
	argsForCall := fake.actualizeArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakePushActor) ActualizeReturns(result1 <-chan *v7pushaction.PushEvent) {
	fake.actualizeMutex.Lock()
	defer fake.actualizeMutex.Unlock()
	fake.ActualizeStub = nil
	fake.actualizeReturns = struct {
		result1 <-chan *v7pushaction.PushEvent
	}{result1}
}

func (fake *FakePushActor) ActualizeReturnsOnCall(i int, result1 <-chan *v7pushaction.PushEvent) {
	fake.actualizeMutex.Lock()
	defer fake.actualizeMutex.Unlock()
	fake.ActualizeStub = nil
	if fake.actualizeReturnsOnCall == nil {
		fake.actualizeReturnsOnCall = make(map[int]struct {
			result1 <-chan *v7pushaction.PushEvent
		})
	}
	fake.actualizeReturnsOnCall[i] = struct {
		result1 <-chan *v7pushaction.PushEvent
	}{result1}
}

func (fake *FakePushActor) CreatePushPlans(arg1 string, arg2 string, arg3 manifestparser.Manifest, arg4 v7pushaction.FlagOverrides) ([]v7pushaction.PushPlan, v7action.Warnings, error) {
	fake.createPushPlansMutex.Lock()
	ret, specificReturn := fake.createPushPlansReturnsOnCall[len(fake.createPushPlansArgsForCall)]
	fake.createPushPlansArgsForCall = append(fake.createPushPlansArgsForCall, struct {
		arg1 string
		arg2 string
		arg3 manifestparser.Manifest
		arg4 v7pushaction.FlagOverrides
	}{arg1, arg2, arg3, arg4})
	stub := fake.CreatePushPlansStub
	fakeReturns := fake.createPushPlansReturns
	fake.recordInvocation("CreatePushPlans", []interface{}{arg1, arg2, arg3, arg4})
	fake.createPushPlansMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4)
	}
	if specificReturn {
		return ret.result1, ret.result2, ret.result3
	}
	return fakeReturns.result1, fakeReturns.result2, fakeReturns.result3
}

func (fake *FakePushActor) CreatePushPlansCallCount() int {
	fake.createPushPlansMutex.RLock()
	defer fake.createPushPlansMutex.RUnlock()
	return len(fake.createPushPlansArgsForCall)
}

func (fake *FakePushActor) CreatePushPlansCalls(stub func(string, string, manifestparser.Manifest, v7pushaction.FlagOverrides) ([]v7pushaction.PushPlan, v7action.Warnings, error)) {
	fake.createPushPlansMutex.Lock()
	defer fake.createPushPlansMutex.Unlock()
	fake.CreatePushPlansStub = stub
}

func (fake *FakePushActor) CreatePushPlansArgsForCall(i int) (string, string, manifestparser.Manifest, v7pushaction.FlagOverrides) {
	fake.createPushPlansMutex.RLock()
	defer fake.createPushPlansMutex.RUnlock()
	argsForCall := fake.createPushPlansArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4
}

func (fake *FakePushActor) CreatePushPlansReturns(result1 []v7pushaction.PushPlan, result2 v7action.Warnings, result3 error) {
	fake.createPushPlansMutex.Lock()
	defer fake.createPushPlansMutex.Unlock()
	fake.CreatePushPlansStub = nil
	fake.createPushPlansReturns = struct {
		result1 []v7pushaction.PushPlan
		result2 v7action.Warnings
		result3 error
	}{result1, result2, result3}
}

func (fake *FakePushActor) CreatePushPlansReturnsOnCall(i int, result1 []v7pushaction.PushPlan, result2 v7action.Warnings, result3 error) {
	fake.createPushPlansMutex.Lock()
	defer fake.createPushPlansMutex.Unlock()
	fake.CreatePushPlansStub = nil
	if fake.createPushPlansReturnsOnCall == nil {
		fake.createPushPlansReturnsOnCall = make(map[int]struct {
			result1 []v7pushaction.PushPlan
			result2 v7action.Warnings
			result3 error
		})
	}
	fake.createPushPlansReturnsOnCall[i] = struct {
		result1 []v7pushaction.PushPlan
		result2 v7action.Warnings
		result3 error
	}{result1, result2, result3}
}

func (fake *FakePushActor) HandleDeploymentScaleFlagOverrides(arg1 manifestparser.Manifest, arg2 v7pushaction.FlagOverrides) (manifestparser.Manifest, error) {
	fake.handleDeploymentScaleFlagOverridesMutex.Lock()
	ret, specificReturn := fake.handleDeploymentScaleFlagOverridesReturnsOnCall[len(fake.handleDeploymentScaleFlagOverridesArgsForCall)]
	fake.handleDeploymentScaleFlagOverridesArgsForCall = append(fake.handleDeploymentScaleFlagOverridesArgsForCall, struct {
		arg1 manifestparser.Manifest
		arg2 v7pushaction.FlagOverrides
	}{arg1, arg2})
	stub := fake.HandleDeploymentScaleFlagOverridesStub
	fakeReturns := fake.handleDeploymentScaleFlagOverridesReturns
	fake.recordInvocation("HandleDeploymentScaleFlagOverrides", []interface{}{arg1, arg2})
	fake.handleDeploymentScaleFlagOverridesMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakePushActor) HandleDeploymentScaleFlagOverridesCallCount() int {
	fake.handleDeploymentScaleFlagOverridesMutex.RLock()
	defer fake.handleDeploymentScaleFlagOverridesMutex.RUnlock()
	return len(fake.handleDeploymentScaleFlagOverridesArgsForCall)
}

func (fake *FakePushActor) HandleDeploymentScaleFlagOverridesCalls(stub func(manifestparser.Manifest, v7pushaction.FlagOverrides) (manifestparser.Manifest, error)) {
	fake.handleDeploymentScaleFlagOverridesMutex.Lock()
	defer fake.handleDeploymentScaleFlagOverridesMutex.Unlock()
	fake.HandleDeploymentScaleFlagOverridesStub = stub
}

func (fake *FakePushActor) HandleDeploymentScaleFlagOverridesArgsForCall(i int) (manifestparser.Manifest, v7pushaction.FlagOverrides) {
	fake.handleDeploymentScaleFlagOverridesMutex.RLock()
	defer fake.handleDeploymentScaleFlagOverridesMutex.RUnlock()
	argsForCall := fake.handleDeploymentScaleFlagOverridesArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakePushActor) HandleDeploymentScaleFlagOverridesReturns(result1 manifestparser.Manifest, result2 error) {
	fake.handleDeploymentScaleFlagOverridesMutex.Lock()
	defer fake.handleDeploymentScaleFlagOverridesMutex.Unlock()
	fake.HandleDeploymentScaleFlagOverridesStub = nil
	fake.handleDeploymentScaleFlagOverridesReturns = struct {
		result1 manifestparser.Manifest
		result2 error
	}{result1, result2}
}

func (fake *FakePushActor) HandleDeploymentScaleFlagOverridesReturnsOnCall(i int, result1 manifestparser.Manifest, result2 error) {
	fake.handleDeploymentScaleFlagOverridesMutex.Lock()
	defer fake.handleDeploymentScaleFlagOverridesMutex.Unlock()
	fake.HandleDeploymentScaleFlagOverridesStub = nil
	if fake.handleDeploymentScaleFlagOverridesReturnsOnCall == nil {
		fake.handleDeploymentScaleFlagOverridesReturnsOnCall = make(map[int]struct {
			result1 manifestparser.Manifest
			result2 error
		})
	}
	fake.handleDeploymentScaleFlagOverridesReturnsOnCall[i] = struct {
		result1 manifestparser.Manifest
		result2 error
	}{result1, result2}
}

func (fake *FakePushActor) HandleFlagOverrides(arg1 manifestparser.Manifest, arg2 v7pushaction.FlagOverrides) (manifestparser.Manifest, error) {
	fake.handleFlagOverridesMutex.Lock()
	ret, specificReturn := fake.handleFlagOverridesReturnsOnCall[len(fake.handleFlagOverridesArgsForCall)]
	fake.handleFlagOverridesArgsForCall = append(fake.handleFlagOverridesArgsForCall, struct {
		arg1 manifestparser.Manifest
		arg2 v7pushaction.FlagOverrides
	}{arg1, arg2})
	stub := fake.HandleFlagOverridesStub
	fakeReturns := fake.handleFlagOverridesReturns
	fake.recordInvocation("HandleFlagOverrides", []interface{}{arg1, arg2})
	fake.handleFlagOverridesMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakePushActor) HandleFlagOverridesCallCount() int {
	fake.handleFlagOverridesMutex.RLock()
	defer fake.handleFlagOverridesMutex.RUnlock()
	return len(fake.handleFlagOverridesArgsForCall)
}

func (fake *FakePushActor) HandleFlagOverridesCalls(stub func(manifestparser.Manifest, v7pushaction.FlagOverrides) (manifestparser.Manifest, error)) {
	fake.handleFlagOverridesMutex.Lock()
	defer fake.handleFlagOverridesMutex.Unlock()
	fake.HandleFlagOverridesStub = stub
}

func (fake *FakePushActor) HandleFlagOverridesArgsForCall(i int) (manifestparser.Manifest, v7pushaction.FlagOverrides) {
	fake.handleFlagOverridesMutex.RLock()
	defer fake.handleFlagOverridesMutex.RUnlock()
	argsForCall := fake.handleFlagOverridesArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakePushActor) HandleFlagOverridesReturns(result1 manifestparser.Manifest, result2 error) {
	fake.handleFlagOverridesMutex.Lock()
	defer fake.handleFlagOverridesMutex.Unlock()
	fake.HandleFlagOverridesStub = nil
	fake.handleFlagOverridesReturns = struct {
		result1 manifestparser.Manifest
		result2 error
	}{result1, result2}
}

func (fake *FakePushActor) HandleFlagOverridesReturnsOnCall(i int, result1 manifestparser.Manifest, result2 error) {
	fake.handleFlagOverridesMutex.Lock()
	defer fake.handleFlagOverridesMutex.Unlock()
	fake.HandleFlagOverridesStub = nil
	if fake.handleFlagOverridesReturnsOnCall == nil {
		fake.handleFlagOverridesReturnsOnCall = make(map[int]struct {
			result1 manifestparser.Manifest
			result2 error
		})
	}
	fake.handleFlagOverridesReturnsOnCall[i] = struct {
		result1 manifestparser.Manifest
		result2 error
	}{result1, result2}
}

func (fake *FakePushActor) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.actualizeMutex.RLock()
	defer fake.actualizeMutex.RUnlock()
	fake.createPushPlansMutex.RLock()
	defer fake.createPushPlansMutex.RUnlock()
	fake.handleDeploymentScaleFlagOverridesMutex.RLock()
	defer fake.handleDeploymentScaleFlagOverridesMutex.RUnlock()
	fake.handleFlagOverridesMutex.RLock()
	defer fake.handleFlagOverridesMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakePushActor) recordInvocation(key string, args []interface{}) {
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

var _ v7.PushActor = new(FakePushActor)
