// Code generated by counterfeiter. DO NOT EDIT.
package v6fakes

import (
	"sync"

	"code.cloudfoundry.org/cli/actor/v2action"
	v6 "code.cloudfoundry.org/cli/command/v6"
)

type FakeCreateServiceActor struct {
	CreateServiceInstanceStub        func(string, string, string, string, string, map[string]interface{}, []string) (v2action.ServiceInstance, v2action.Warnings, error)
	createServiceInstanceMutex       sync.RWMutex
	createServiceInstanceArgsForCall []struct {
		arg1 string
		arg2 string
		arg3 string
		arg4 string
		arg5 string
		arg6 map[string]interface{}
		arg7 []string
	}
	createServiceInstanceReturns struct {
		result1 v2action.ServiceInstance
		result2 v2action.Warnings
		result3 error
	}
	createServiceInstanceReturnsOnCall map[int]struct {
		result1 v2action.ServiceInstance
		result2 v2action.Warnings
		result3 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeCreateServiceActor) CreateServiceInstance(arg1 string, arg2 string, arg3 string, arg4 string, arg5 string, arg6 map[string]interface{}, arg7 []string) (v2action.ServiceInstance, v2action.Warnings, error) {
	var arg7Copy []string
	if arg7 != nil {
		arg7Copy = make([]string, len(arg7))
		copy(arg7Copy, arg7)
	}
	fake.createServiceInstanceMutex.Lock()
	ret, specificReturn := fake.createServiceInstanceReturnsOnCall[len(fake.createServiceInstanceArgsForCall)]
	fake.createServiceInstanceArgsForCall = append(fake.createServiceInstanceArgsForCall, struct {
		arg1 string
		arg2 string
		arg3 string
		arg4 string
		arg5 string
		arg6 map[string]interface{}
		arg7 []string
	}{arg1, arg2, arg3, arg4, arg5, arg6, arg7Copy})
	fake.recordInvocation("CreateServiceInstance", []interface{}{arg1, arg2, arg3, arg4, arg5, arg6, arg7Copy})
	fake.createServiceInstanceMutex.Unlock()
	if fake.CreateServiceInstanceStub != nil {
		return fake.CreateServiceInstanceStub(arg1, arg2, arg3, arg4, arg5, arg6, arg7)
	}
	if specificReturn {
		return ret.result1, ret.result2, ret.result3
	}
	fakeReturns := fake.createServiceInstanceReturns
	return fakeReturns.result1, fakeReturns.result2, fakeReturns.result3
}

func (fake *FakeCreateServiceActor) CreateServiceInstanceCallCount() int {
	fake.createServiceInstanceMutex.RLock()
	defer fake.createServiceInstanceMutex.RUnlock()
	return len(fake.createServiceInstanceArgsForCall)
}

func (fake *FakeCreateServiceActor) CreateServiceInstanceCalls(stub func(string, string, string, string, string, map[string]interface{}, []string) (v2action.ServiceInstance, v2action.Warnings, error)) {
	fake.createServiceInstanceMutex.Lock()
	defer fake.createServiceInstanceMutex.Unlock()
	fake.CreateServiceInstanceStub = stub
}

func (fake *FakeCreateServiceActor) CreateServiceInstanceArgsForCall(i int) (string, string, string, string, string, map[string]interface{}, []string) {
	fake.createServiceInstanceMutex.RLock()
	defer fake.createServiceInstanceMutex.RUnlock()
	argsForCall := fake.createServiceInstanceArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4, argsForCall.arg5, argsForCall.arg6, argsForCall.arg7
}

func (fake *FakeCreateServiceActor) CreateServiceInstanceReturns(result1 v2action.ServiceInstance, result2 v2action.Warnings, result3 error) {
	fake.createServiceInstanceMutex.Lock()
	defer fake.createServiceInstanceMutex.Unlock()
	fake.CreateServiceInstanceStub = nil
	fake.createServiceInstanceReturns = struct {
		result1 v2action.ServiceInstance
		result2 v2action.Warnings
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeCreateServiceActor) CreateServiceInstanceReturnsOnCall(i int, result1 v2action.ServiceInstance, result2 v2action.Warnings, result3 error) {
	fake.createServiceInstanceMutex.Lock()
	defer fake.createServiceInstanceMutex.Unlock()
	fake.CreateServiceInstanceStub = nil
	if fake.createServiceInstanceReturnsOnCall == nil {
		fake.createServiceInstanceReturnsOnCall = make(map[int]struct {
			result1 v2action.ServiceInstance
			result2 v2action.Warnings
			result3 error
		})
	}
	fake.createServiceInstanceReturnsOnCall[i] = struct {
		result1 v2action.ServiceInstance
		result2 v2action.Warnings
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeCreateServiceActor) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.createServiceInstanceMutex.RLock()
	defer fake.createServiceInstanceMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeCreateServiceActor) recordInvocation(key string, args []interface{}) {
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

var _ v6.CreateServiceActor = new(FakeCreateServiceActor)