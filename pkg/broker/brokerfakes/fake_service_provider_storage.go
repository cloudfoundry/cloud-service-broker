// Code generated by counterfeiter. DO NOT EDIT.
package brokerfakes

import (
	"sync"

	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/pkg/broker"
)

type FakeServiceProviderStorage struct {
	ExistsTerraformDeploymentStub        func(string) (bool, error)
	existsTerraformDeploymentMutex       sync.RWMutex
	existsTerraformDeploymentArgsForCall []struct {
		arg1 string
	}
	existsTerraformDeploymentReturns struct {
		result1 bool
		result2 error
	}
	existsTerraformDeploymentReturnsOnCall map[int]struct {
		result1 bool
		result2 error
	}
	GetServiceBindingsForServiceInstanceStub        func(string) ([]string, error)
	getServiceBindingsForServiceInstanceMutex       sync.RWMutex
	getServiceBindingsForServiceInstanceArgsForCall []struct {
		arg1 string
	}
	getServiceBindingsForServiceInstanceReturns struct {
		result1 []string
		result2 error
	}
	getServiceBindingsForServiceInstanceReturnsOnCall map[int]struct {
		result1 []string
		result2 error
	}
	GetTerraformDeploymentStub        func(string) (storage.TerraformDeployment, error)
	getTerraformDeploymentMutex       sync.RWMutex
	getTerraformDeploymentArgsForCall []struct {
		arg1 string
	}
	getTerraformDeploymentReturns struct {
		result1 storage.TerraformDeployment
		result2 error
	}
	getTerraformDeploymentReturnsOnCall map[int]struct {
		result1 storage.TerraformDeployment
		result2 error
	}
	StoreTerraformDeploymentStub        func(storage.TerraformDeployment) error
	storeTerraformDeploymentMutex       sync.RWMutex
	storeTerraformDeploymentArgsForCall []struct {
		arg1 storage.TerraformDeployment
	}
	storeTerraformDeploymentReturns struct {
		result1 error
	}
	storeTerraformDeploymentReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeServiceProviderStorage) ExistsTerraformDeployment(arg1 string) (bool, error) {
	fake.existsTerraformDeploymentMutex.Lock()
	ret, specificReturn := fake.existsTerraformDeploymentReturnsOnCall[len(fake.existsTerraformDeploymentArgsForCall)]
	fake.existsTerraformDeploymentArgsForCall = append(fake.existsTerraformDeploymentArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.ExistsTerraformDeploymentStub
	fakeReturns := fake.existsTerraformDeploymentReturns
	fake.recordInvocation("ExistsTerraformDeployment", []interface{}{arg1})
	fake.existsTerraformDeploymentMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeServiceProviderStorage) ExistsTerraformDeploymentCallCount() int {
	fake.existsTerraformDeploymentMutex.RLock()
	defer fake.existsTerraformDeploymentMutex.RUnlock()
	return len(fake.existsTerraformDeploymentArgsForCall)
}

func (fake *FakeServiceProviderStorage) ExistsTerraformDeploymentCalls(stub func(string) (bool, error)) {
	fake.existsTerraformDeploymentMutex.Lock()
	defer fake.existsTerraformDeploymentMutex.Unlock()
	fake.ExistsTerraformDeploymentStub = stub
}

func (fake *FakeServiceProviderStorage) ExistsTerraformDeploymentArgsForCall(i int) string {
	fake.existsTerraformDeploymentMutex.RLock()
	defer fake.existsTerraformDeploymentMutex.RUnlock()
	argsForCall := fake.existsTerraformDeploymentArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeServiceProviderStorage) ExistsTerraformDeploymentReturns(result1 bool, result2 error) {
	fake.existsTerraformDeploymentMutex.Lock()
	defer fake.existsTerraformDeploymentMutex.Unlock()
	fake.ExistsTerraformDeploymentStub = nil
	fake.existsTerraformDeploymentReturns = struct {
		result1 bool
		result2 error
	}{result1, result2}
}

func (fake *FakeServiceProviderStorage) ExistsTerraformDeploymentReturnsOnCall(i int, result1 bool, result2 error) {
	fake.existsTerraformDeploymentMutex.Lock()
	defer fake.existsTerraformDeploymentMutex.Unlock()
	fake.ExistsTerraformDeploymentStub = nil
	if fake.existsTerraformDeploymentReturnsOnCall == nil {
		fake.existsTerraformDeploymentReturnsOnCall = make(map[int]struct {
			result1 bool
			result2 error
		})
	}
	fake.existsTerraformDeploymentReturnsOnCall[i] = struct {
		result1 bool
		result2 error
	}{result1, result2}
}

func (fake *FakeServiceProviderStorage) GetServiceBindingsForServiceInstance(arg1 string) ([]string, error) {
	fake.getServiceBindingsForServiceInstanceMutex.Lock()
	ret, specificReturn := fake.getServiceBindingsForServiceInstanceReturnsOnCall[len(fake.getServiceBindingsForServiceInstanceArgsForCall)]
	fake.getServiceBindingsForServiceInstanceArgsForCall = append(fake.getServiceBindingsForServiceInstanceArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.GetServiceBindingsForServiceInstanceStub
	fakeReturns := fake.getServiceBindingsForServiceInstanceReturns
	fake.recordInvocation("GetServiceBindingsForServiceInstance", []interface{}{arg1})
	fake.getServiceBindingsForServiceInstanceMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeServiceProviderStorage) GetServiceBindingsForServiceInstanceCallCount() int {
	fake.getServiceBindingsForServiceInstanceMutex.RLock()
	defer fake.getServiceBindingsForServiceInstanceMutex.RUnlock()
	return len(fake.getServiceBindingsForServiceInstanceArgsForCall)
}

func (fake *FakeServiceProviderStorage) GetServiceBindingsForServiceInstanceCalls(stub func(string) ([]string, error)) {
	fake.getServiceBindingsForServiceInstanceMutex.Lock()
	defer fake.getServiceBindingsForServiceInstanceMutex.Unlock()
	fake.GetServiceBindingsForServiceInstanceStub = stub
}

func (fake *FakeServiceProviderStorage) GetServiceBindingsForServiceInstanceArgsForCall(i int) string {
	fake.getServiceBindingsForServiceInstanceMutex.RLock()
	defer fake.getServiceBindingsForServiceInstanceMutex.RUnlock()
	argsForCall := fake.getServiceBindingsForServiceInstanceArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeServiceProviderStorage) GetServiceBindingsForServiceInstanceReturns(result1 []string, result2 error) {
	fake.getServiceBindingsForServiceInstanceMutex.Lock()
	defer fake.getServiceBindingsForServiceInstanceMutex.Unlock()
	fake.GetServiceBindingsForServiceInstanceStub = nil
	fake.getServiceBindingsForServiceInstanceReturns = struct {
		result1 []string
		result2 error
	}{result1, result2}
}

func (fake *FakeServiceProviderStorage) GetServiceBindingsForServiceInstanceReturnsOnCall(i int, result1 []string, result2 error) {
	fake.getServiceBindingsForServiceInstanceMutex.Lock()
	defer fake.getServiceBindingsForServiceInstanceMutex.Unlock()
	fake.GetServiceBindingsForServiceInstanceStub = nil
	if fake.getServiceBindingsForServiceInstanceReturnsOnCall == nil {
		fake.getServiceBindingsForServiceInstanceReturnsOnCall = make(map[int]struct {
			result1 []string
			result2 error
		})
	}
	fake.getServiceBindingsForServiceInstanceReturnsOnCall[i] = struct {
		result1 []string
		result2 error
	}{result1, result2}
}

func (fake *FakeServiceProviderStorage) GetTerraformDeployment(arg1 string) (storage.TerraformDeployment, error) {
	fake.getTerraformDeploymentMutex.Lock()
	ret, specificReturn := fake.getTerraformDeploymentReturnsOnCall[len(fake.getTerraformDeploymentArgsForCall)]
	fake.getTerraformDeploymentArgsForCall = append(fake.getTerraformDeploymentArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.GetTerraformDeploymentStub
	fakeReturns := fake.getTerraformDeploymentReturns
	fake.recordInvocation("GetTerraformDeployment", []interface{}{arg1})
	fake.getTerraformDeploymentMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeServiceProviderStorage) GetTerraformDeploymentCallCount() int {
	fake.getTerraformDeploymentMutex.RLock()
	defer fake.getTerraformDeploymentMutex.RUnlock()
	return len(fake.getTerraformDeploymentArgsForCall)
}

func (fake *FakeServiceProviderStorage) GetTerraformDeploymentCalls(stub func(string) (storage.TerraformDeployment, error)) {
	fake.getTerraformDeploymentMutex.Lock()
	defer fake.getTerraformDeploymentMutex.Unlock()
	fake.GetTerraformDeploymentStub = stub
}

func (fake *FakeServiceProviderStorage) GetTerraformDeploymentArgsForCall(i int) string {
	fake.getTerraformDeploymentMutex.RLock()
	defer fake.getTerraformDeploymentMutex.RUnlock()
	argsForCall := fake.getTerraformDeploymentArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeServiceProviderStorage) GetTerraformDeploymentReturns(result1 storage.TerraformDeployment, result2 error) {
	fake.getTerraformDeploymentMutex.Lock()
	defer fake.getTerraformDeploymentMutex.Unlock()
	fake.GetTerraformDeploymentStub = nil
	fake.getTerraformDeploymentReturns = struct {
		result1 storage.TerraformDeployment
		result2 error
	}{result1, result2}
}

func (fake *FakeServiceProviderStorage) GetTerraformDeploymentReturnsOnCall(i int, result1 storage.TerraformDeployment, result2 error) {
	fake.getTerraformDeploymentMutex.Lock()
	defer fake.getTerraformDeploymentMutex.Unlock()
	fake.GetTerraformDeploymentStub = nil
	if fake.getTerraformDeploymentReturnsOnCall == nil {
		fake.getTerraformDeploymentReturnsOnCall = make(map[int]struct {
			result1 storage.TerraformDeployment
			result2 error
		})
	}
	fake.getTerraformDeploymentReturnsOnCall[i] = struct {
		result1 storage.TerraformDeployment
		result2 error
	}{result1, result2}
}

func (fake *FakeServiceProviderStorage) StoreTerraformDeployment(arg1 storage.TerraformDeployment) error {
	fake.storeTerraformDeploymentMutex.Lock()
	ret, specificReturn := fake.storeTerraformDeploymentReturnsOnCall[len(fake.storeTerraformDeploymentArgsForCall)]
	fake.storeTerraformDeploymentArgsForCall = append(fake.storeTerraformDeploymentArgsForCall, struct {
		arg1 storage.TerraformDeployment
	}{arg1})
	stub := fake.StoreTerraformDeploymentStub
	fakeReturns := fake.storeTerraformDeploymentReturns
	fake.recordInvocation("StoreTerraformDeployment", []interface{}{arg1})
	fake.storeTerraformDeploymentMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeServiceProviderStorage) StoreTerraformDeploymentCallCount() int {
	fake.storeTerraformDeploymentMutex.RLock()
	defer fake.storeTerraformDeploymentMutex.RUnlock()
	return len(fake.storeTerraformDeploymentArgsForCall)
}

func (fake *FakeServiceProviderStorage) StoreTerraformDeploymentCalls(stub func(storage.TerraformDeployment) error) {
	fake.storeTerraformDeploymentMutex.Lock()
	defer fake.storeTerraformDeploymentMutex.Unlock()
	fake.StoreTerraformDeploymentStub = stub
}

func (fake *FakeServiceProviderStorage) StoreTerraformDeploymentArgsForCall(i int) storage.TerraformDeployment {
	fake.storeTerraformDeploymentMutex.RLock()
	defer fake.storeTerraformDeploymentMutex.RUnlock()
	argsForCall := fake.storeTerraformDeploymentArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeServiceProviderStorage) StoreTerraformDeploymentReturns(result1 error) {
	fake.storeTerraformDeploymentMutex.Lock()
	defer fake.storeTerraformDeploymentMutex.Unlock()
	fake.StoreTerraformDeploymentStub = nil
	fake.storeTerraformDeploymentReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeServiceProviderStorage) StoreTerraformDeploymentReturnsOnCall(i int, result1 error) {
	fake.storeTerraformDeploymentMutex.Lock()
	defer fake.storeTerraformDeploymentMutex.Unlock()
	fake.StoreTerraformDeploymentStub = nil
	if fake.storeTerraformDeploymentReturnsOnCall == nil {
		fake.storeTerraformDeploymentReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.storeTerraformDeploymentReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeServiceProviderStorage) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.existsTerraformDeploymentMutex.RLock()
	defer fake.existsTerraformDeploymentMutex.RUnlock()
	fake.getServiceBindingsForServiceInstanceMutex.RLock()
	defer fake.getServiceBindingsForServiceInstanceMutex.RUnlock()
	fake.getTerraformDeploymentMutex.RLock()
	defer fake.getTerraformDeploymentMutex.RUnlock()
	fake.storeTerraformDeploymentMutex.RLock()
	defer fake.storeTerraformDeploymentMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeServiceProviderStorage) recordInvocation(key string, args []interface{}) {
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

var _ broker.ServiceProviderStorage = new(FakeServiceProviderStorage)
