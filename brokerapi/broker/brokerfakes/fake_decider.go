// Code generated by counterfeiter. DO NOT EDIT.
package brokerfakes

import (
	"sync"

	"github.com/cloudfoundry/cloud-service-broker/brokerapi/broker"
	"github.com/cloudfoundry/cloud-service-broker/brokerapi/broker/decider"
	brokera "github.com/cloudfoundry/cloud-service-broker/pkg/broker"
	"github.com/pivotal-cf/brokerapi/v8/domain"
)

type FakeDecider struct {
	DecideOperationStub        func(*brokera.ServiceDefinition, domain.UpdateDetails) (decider.Operation, error)
	decideOperationMutex       sync.RWMutex
	decideOperationArgsForCall []struct {
		arg1 *brokera.ServiceDefinition
		arg2 domain.UpdateDetails
	}
	decideOperationReturns struct {
		result1 decider.Operation
		result2 error
	}
	decideOperationReturnsOnCall map[int]struct {
		result1 decider.Operation
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeDecider) DecideOperation(arg1 *brokera.ServiceDefinition, arg2 domain.UpdateDetails) (decider.Operation, error) {
	fake.decideOperationMutex.Lock()
	ret, specificReturn := fake.decideOperationReturnsOnCall[len(fake.decideOperationArgsForCall)]
	fake.decideOperationArgsForCall = append(fake.decideOperationArgsForCall, struct {
		arg1 *brokera.ServiceDefinition
		arg2 domain.UpdateDetails
	}{arg1, arg2})
	stub := fake.DecideOperationStub
	fakeReturns := fake.decideOperationReturns
	fake.recordInvocation("DecideOperation", []interface{}{arg1, arg2})
	fake.decideOperationMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeDecider) DecideOperationCallCount() int {
	fake.decideOperationMutex.RLock()
	defer fake.decideOperationMutex.RUnlock()
	return len(fake.decideOperationArgsForCall)
}

func (fake *FakeDecider) DecideOperationCalls(stub func(*brokera.ServiceDefinition, domain.UpdateDetails) (decider.Operation, error)) {
	fake.decideOperationMutex.Lock()
	defer fake.decideOperationMutex.Unlock()
	fake.DecideOperationStub = stub
}

func (fake *FakeDecider) DecideOperationArgsForCall(i int) (*brokera.ServiceDefinition, domain.UpdateDetails) {
	fake.decideOperationMutex.RLock()
	defer fake.decideOperationMutex.RUnlock()
	argsForCall := fake.decideOperationArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeDecider) DecideOperationReturns(result1 decider.Operation, result2 error) {
	fake.decideOperationMutex.Lock()
	defer fake.decideOperationMutex.Unlock()
	fake.DecideOperationStub = nil
	fake.decideOperationReturns = struct {
		result1 decider.Operation
		result2 error
	}{result1, result2}
}

func (fake *FakeDecider) DecideOperationReturnsOnCall(i int, result1 decider.Operation, result2 error) {
	fake.decideOperationMutex.Lock()
	defer fake.decideOperationMutex.Unlock()
	fake.DecideOperationStub = nil
	if fake.decideOperationReturnsOnCall == nil {
		fake.decideOperationReturnsOnCall = make(map[int]struct {
			result1 decider.Operation
			result2 error
		})
	}
	fake.decideOperationReturnsOnCall[i] = struct {
		result1 decider.Operation
		result2 error
	}{result1, result2}
}

func (fake *FakeDecider) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.decideOperationMutex.RLock()
	defer fake.decideOperationMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeDecider) recordInvocation(key string, args []interface{}) {
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

var _ broker.Decider = new(FakeDecider)