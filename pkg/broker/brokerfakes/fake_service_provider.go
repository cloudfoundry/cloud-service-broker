// Code generated by counterfeiter. DO NOT EDIT.
package brokerfakes

import (
	"context"
	"sync"

	"github.com/cloudfoundry/cloud-service-broker/v3/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/v3/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/broker"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/varcontext"
)

type FakeServiceProvider struct {
	BindStub        func(context.Context, *varcontext.VarContext) (map[string]any, error)
	bindMutex       sync.RWMutex
	bindArgsForCall []struct {
		arg1 context.Context
		arg2 *varcontext.VarContext
	}
	bindReturns struct {
		result1 map[string]any
		result2 error
	}
	bindReturnsOnCall map[int]struct {
		result1 map[string]any
		result2 error
	}
	CheckOperationConstraintsStub        func(string, string) error
	checkOperationConstraintsMutex       sync.RWMutex
	checkOperationConstraintsArgsForCall []struct {
		arg1 string
		arg2 string
	}
	checkOperationConstraintsReturns struct {
		result1 error
	}
	checkOperationConstraintsReturnsOnCall map[int]struct {
		result1 error
	}
	CheckUpgradeAvailableStub        func(string) error
	checkUpgradeAvailableMutex       sync.RWMutex
	checkUpgradeAvailableArgsForCall []struct {
		arg1 string
	}
	checkUpgradeAvailableReturns struct {
		result1 error
	}
	checkUpgradeAvailableReturnsOnCall map[int]struct {
		result1 error
	}
	DeleteBindingDataStub        func(context.Context, string, string) error
	deleteBindingDataMutex       sync.RWMutex
	deleteBindingDataArgsForCall []struct {
		arg1 context.Context
		arg2 string
		arg3 string
	}
	deleteBindingDataReturns struct {
		result1 error
	}
	deleteBindingDataReturnsOnCall map[int]struct {
		result1 error
	}
	DeleteInstanceDataStub        func(context.Context, string) error
	deleteInstanceDataMutex       sync.RWMutex
	deleteInstanceDataArgsForCall []struct {
		arg1 context.Context
		arg2 string
	}
	deleteInstanceDataReturns struct {
		result1 error
	}
	deleteInstanceDataReturnsOnCall map[int]struct {
		result1 error
	}
	DeprovisionStub        func(context.Context, string, *varcontext.VarContext) (*string, error)
	deprovisionMutex       sync.RWMutex
	deprovisionArgsForCall []struct {
		arg1 context.Context
		arg2 string
		arg3 *varcontext.VarContext
	}
	deprovisionReturns struct {
		result1 *string
		result2 error
	}
	deprovisionReturnsOnCall map[int]struct {
		result1 *string
		result2 error
	}
	GetImportedPropertiesStub        func(context.Context, string, []broker.BrokerVariable, map[string]any) (map[string]any, error)
	getImportedPropertiesMutex       sync.RWMutex
	getImportedPropertiesArgsForCall []struct {
		arg1 context.Context
		arg2 string
		arg3 []broker.BrokerVariable
		arg4 map[string]any
	}
	getImportedPropertiesReturns struct {
		result1 map[string]any
		result2 error
	}
	getImportedPropertiesReturnsOnCall map[int]struct {
		result1 map[string]any
		result2 error
	}
	GetTerraformOutputsStub        func(context.Context, string) (storage.JSONObject, error)
	getTerraformOutputsMutex       sync.RWMutex
	getTerraformOutputsArgsForCall []struct {
		arg1 context.Context
		arg2 string
	}
	getTerraformOutputsReturns struct {
		result1 storage.JSONObject
		result2 error
	}
	getTerraformOutputsReturnsOnCall map[int]struct {
		result1 storage.JSONObject
		result2 error
	}
	PollInstanceStub        func(context.Context, string) (bool, string, error)
	pollInstanceMutex       sync.RWMutex
	pollInstanceArgsForCall []struct {
		arg1 context.Context
		arg2 string
	}
	pollInstanceReturns struct {
		result1 bool
		result2 string
		result3 error
	}
	pollInstanceReturnsOnCall map[int]struct {
		result1 bool
		result2 string
		result3 error
	}
	ProvisionStub        func(context.Context, *varcontext.VarContext) (storage.ServiceInstanceDetails, error)
	provisionMutex       sync.RWMutex
	provisionArgsForCall []struct {
		arg1 context.Context
		arg2 *varcontext.VarContext
	}
	provisionReturns struct {
		result1 storage.ServiceInstanceDetails
		result2 error
	}
	provisionReturnsOnCall map[int]struct {
		result1 storage.ServiceInstanceDetails
		result2 error
	}
	UnbindStub        func(context.Context, string, string, *varcontext.VarContext) error
	unbindMutex       sync.RWMutex
	unbindArgsForCall []struct {
		arg1 context.Context
		arg2 string
		arg3 string
		arg4 *varcontext.VarContext
	}
	unbindReturns struct {
		result1 error
	}
	unbindReturnsOnCall map[int]struct {
		result1 error
	}
	UpdateStub        func(context.Context, *varcontext.VarContext) (models.ServiceInstanceDetails, error)
	updateMutex       sync.RWMutex
	updateArgsForCall []struct {
		arg1 context.Context
		arg2 *varcontext.VarContext
	}
	updateReturns struct {
		result1 models.ServiceInstanceDetails
		result2 error
	}
	updateReturnsOnCall map[int]struct {
		result1 models.ServiceInstanceDetails
		result2 error
	}
	UpgradeBindingsStub        func(context.Context, *varcontext.VarContext, []*varcontext.VarContext) error
	upgradeBindingsMutex       sync.RWMutex
	upgradeBindingsArgsForCall []struct {
		arg1 context.Context
		arg2 *varcontext.VarContext
		arg3 []*varcontext.VarContext
	}
	upgradeBindingsReturns struct {
		result1 error
	}
	upgradeBindingsReturnsOnCall map[int]struct {
		result1 error
	}
	UpgradeInstanceStub        func(context.Context, *varcontext.VarContext) (*sync.WaitGroup, error)
	upgradeInstanceMutex       sync.RWMutex
	upgradeInstanceArgsForCall []struct {
		arg1 context.Context
		arg2 *varcontext.VarContext
	}
	upgradeInstanceReturns struct {
		result1 *sync.WaitGroup
		result2 error
	}
	upgradeInstanceReturnsOnCall map[int]struct {
		result1 *sync.WaitGroup
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeServiceProvider) Bind(arg1 context.Context, arg2 *varcontext.VarContext) (map[string]any, error) {
	fake.bindMutex.Lock()
	ret, specificReturn := fake.bindReturnsOnCall[len(fake.bindArgsForCall)]
	fake.bindArgsForCall = append(fake.bindArgsForCall, struct {
		arg1 context.Context
		arg2 *varcontext.VarContext
	}{arg1, arg2})
	stub := fake.BindStub
	fakeReturns := fake.bindReturns
	fake.recordInvocation("Bind", []interface{}{arg1, arg2})
	fake.bindMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeServiceProvider) BindCallCount() int {
	fake.bindMutex.RLock()
	defer fake.bindMutex.RUnlock()
	return len(fake.bindArgsForCall)
}

func (fake *FakeServiceProvider) BindCalls(stub func(context.Context, *varcontext.VarContext) (map[string]any, error)) {
	fake.bindMutex.Lock()
	defer fake.bindMutex.Unlock()
	fake.BindStub = stub
}

func (fake *FakeServiceProvider) BindArgsForCall(i int) (context.Context, *varcontext.VarContext) {
	fake.bindMutex.RLock()
	defer fake.bindMutex.RUnlock()
	argsForCall := fake.bindArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeServiceProvider) BindReturns(result1 map[string]any, result2 error) {
	fake.bindMutex.Lock()
	defer fake.bindMutex.Unlock()
	fake.BindStub = nil
	fake.bindReturns = struct {
		result1 map[string]any
		result2 error
	}{result1, result2}
}

func (fake *FakeServiceProvider) BindReturnsOnCall(i int, result1 map[string]any, result2 error) {
	fake.bindMutex.Lock()
	defer fake.bindMutex.Unlock()
	fake.BindStub = nil
	if fake.bindReturnsOnCall == nil {
		fake.bindReturnsOnCall = make(map[int]struct {
			result1 map[string]any
			result2 error
		})
	}
	fake.bindReturnsOnCall[i] = struct {
		result1 map[string]any
		result2 error
	}{result1, result2}
}

func (fake *FakeServiceProvider) CheckOperationConstraints(arg1 string, arg2 string) error {
	fake.checkOperationConstraintsMutex.Lock()
	ret, specificReturn := fake.checkOperationConstraintsReturnsOnCall[len(fake.checkOperationConstraintsArgsForCall)]
	fake.checkOperationConstraintsArgsForCall = append(fake.checkOperationConstraintsArgsForCall, struct {
		arg1 string
		arg2 string
	}{arg1, arg2})
	stub := fake.CheckOperationConstraintsStub
	fakeReturns := fake.checkOperationConstraintsReturns
	fake.recordInvocation("CheckOperationConstraints", []interface{}{arg1, arg2})
	fake.checkOperationConstraintsMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeServiceProvider) CheckOperationConstraintsCallCount() int {
	fake.checkOperationConstraintsMutex.RLock()
	defer fake.checkOperationConstraintsMutex.RUnlock()
	return len(fake.checkOperationConstraintsArgsForCall)
}

func (fake *FakeServiceProvider) CheckOperationConstraintsCalls(stub func(string, string) error) {
	fake.checkOperationConstraintsMutex.Lock()
	defer fake.checkOperationConstraintsMutex.Unlock()
	fake.CheckOperationConstraintsStub = stub
}

func (fake *FakeServiceProvider) CheckOperationConstraintsArgsForCall(i int) (string, string) {
	fake.checkOperationConstraintsMutex.RLock()
	defer fake.checkOperationConstraintsMutex.RUnlock()
	argsForCall := fake.checkOperationConstraintsArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeServiceProvider) CheckOperationConstraintsReturns(result1 error) {
	fake.checkOperationConstraintsMutex.Lock()
	defer fake.checkOperationConstraintsMutex.Unlock()
	fake.CheckOperationConstraintsStub = nil
	fake.checkOperationConstraintsReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeServiceProvider) CheckOperationConstraintsReturnsOnCall(i int, result1 error) {
	fake.checkOperationConstraintsMutex.Lock()
	defer fake.checkOperationConstraintsMutex.Unlock()
	fake.CheckOperationConstraintsStub = nil
	if fake.checkOperationConstraintsReturnsOnCall == nil {
		fake.checkOperationConstraintsReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.checkOperationConstraintsReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeServiceProvider) CheckUpgradeAvailable(arg1 string) error {
	fake.checkUpgradeAvailableMutex.Lock()
	ret, specificReturn := fake.checkUpgradeAvailableReturnsOnCall[len(fake.checkUpgradeAvailableArgsForCall)]
	fake.checkUpgradeAvailableArgsForCall = append(fake.checkUpgradeAvailableArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.CheckUpgradeAvailableStub
	fakeReturns := fake.checkUpgradeAvailableReturns
	fake.recordInvocation("CheckUpgradeAvailable", []interface{}{arg1})
	fake.checkUpgradeAvailableMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeServiceProvider) CheckUpgradeAvailableCallCount() int {
	fake.checkUpgradeAvailableMutex.RLock()
	defer fake.checkUpgradeAvailableMutex.RUnlock()
	return len(fake.checkUpgradeAvailableArgsForCall)
}

func (fake *FakeServiceProvider) CheckUpgradeAvailableCalls(stub func(string) error) {
	fake.checkUpgradeAvailableMutex.Lock()
	defer fake.checkUpgradeAvailableMutex.Unlock()
	fake.CheckUpgradeAvailableStub = stub
}

func (fake *FakeServiceProvider) CheckUpgradeAvailableArgsForCall(i int) string {
	fake.checkUpgradeAvailableMutex.RLock()
	defer fake.checkUpgradeAvailableMutex.RUnlock()
	argsForCall := fake.checkUpgradeAvailableArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeServiceProvider) CheckUpgradeAvailableReturns(result1 error) {
	fake.checkUpgradeAvailableMutex.Lock()
	defer fake.checkUpgradeAvailableMutex.Unlock()
	fake.CheckUpgradeAvailableStub = nil
	fake.checkUpgradeAvailableReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeServiceProvider) CheckUpgradeAvailableReturnsOnCall(i int, result1 error) {
	fake.checkUpgradeAvailableMutex.Lock()
	defer fake.checkUpgradeAvailableMutex.Unlock()
	fake.CheckUpgradeAvailableStub = nil
	if fake.checkUpgradeAvailableReturnsOnCall == nil {
		fake.checkUpgradeAvailableReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.checkUpgradeAvailableReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeServiceProvider) DeleteBindingData(arg1 context.Context, arg2 string, arg3 string) error {
	fake.deleteBindingDataMutex.Lock()
	ret, specificReturn := fake.deleteBindingDataReturnsOnCall[len(fake.deleteBindingDataArgsForCall)]
	fake.deleteBindingDataArgsForCall = append(fake.deleteBindingDataArgsForCall, struct {
		arg1 context.Context
		arg2 string
		arg3 string
	}{arg1, arg2, arg3})
	stub := fake.DeleteBindingDataStub
	fakeReturns := fake.deleteBindingDataReturns
	fake.recordInvocation("DeleteBindingData", []interface{}{arg1, arg2, arg3})
	fake.deleteBindingDataMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeServiceProvider) DeleteBindingDataCallCount() int {
	fake.deleteBindingDataMutex.RLock()
	defer fake.deleteBindingDataMutex.RUnlock()
	return len(fake.deleteBindingDataArgsForCall)
}

func (fake *FakeServiceProvider) DeleteBindingDataCalls(stub func(context.Context, string, string) error) {
	fake.deleteBindingDataMutex.Lock()
	defer fake.deleteBindingDataMutex.Unlock()
	fake.DeleteBindingDataStub = stub
}

func (fake *FakeServiceProvider) DeleteBindingDataArgsForCall(i int) (context.Context, string, string) {
	fake.deleteBindingDataMutex.RLock()
	defer fake.deleteBindingDataMutex.RUnlock()
	argsForCall := fake.deleteBindingDataArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeServiceProvider) DeleteBindingDataReturns(result1 error) {
	fake.deleteBindingDataMutex.Lock()
	defer fake.deleteBindingDataMutex.Unlock()
	fake.DeleteBindingDataStub = nil
	fake.deleteBindingDataReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeServiceProvider) DeleteBindingDataReturnsOnCall(i int, result1 error) {
	fake.deleteBindingDataMutex.Lock()
	defer fake.deleteBindingDataMutex.Unlock()
	fake.DeleteBindingDataStub = nil
	if fake.deleteBindingDataReturnsOnCall == nil {
		fake.deleteBindingDataReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.deleteBindingDataReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeServiceProvider) DeleteInstanceData(arg1 context.Context, arg2 string) error {
	fake.deleteInstanceDataMutex.Lock()
	ret, specificReturn := fake.deleteInstanceDataReturnsOnCall[len(fake.deleteInstanceDataArgsForCall)]
	fake.deleteInstanceDataArgsForCall = append(fake.deleteInstanceDataArgsForCall, struct {
		arg1 context.Context
		arg2 string
	}{arg1, arg2})
	stub := fake.DeleteInstanceDataStub
	fakeReturns := fake.deleteInstanceDataReturns
	fake.recordInvocation("DeleteInstanceData", []interface{}{arg1, arg2})
	fake.deleteInstanceDataMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeServiceProvider) DeleteInstanceDataCallCount() int {
	fake.deleteInstanceDataMutex.RLock()
	defer fake.deleteInstanceDataMutex.RUnlock()
	return len(fake.deleteInstanceDataArgsForCall)
}

func (fake *FakeServiceProvider) DeleteInstanceDataCalls(stub func(context.Context, string) error) {
	fake.deleteInstanceDataMutex.Lock()
	defer fake.deleteInstanceDataMutex.Unlock()
	fake.DeleteInstanceDataStub = stub
}

func (fake *FakeServiceProvider) DeleteInstanceDataArgsForCall(i int) (context.Context, string) {
	fake.deleteInstanceDataMutex.RLock()
	defer fake.deleteInstanceDataMutex.RUnlock()
	argsForCall := fake.deleteInstanceDataArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeServiceProvider) DeleteInstanceDataReturns(result1 error) {
	fake.deleteInstanceDataMutex.Lock()
	defer fake.deleteInstanceDataMutex.Unlock()
	fake.DeleteInstanceDataStub = nil
	fake.deleteInstanceDataReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeServiceProvider) DeleteInstanceDataReturnsOnCall(i int, result1 error) {
	fake.deleteInstanceDataMutex.Lock()
	defer fake.deleteInstanceDataMutex.Unlock()
	fake.DeleteInstanceDataStub = nil
	if fake.deleteInstanceDataReturnsOnCall == nil {
		fake.deleteInstanceDataReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.deleteInstanceDataReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeServiceProvider) Deprovision(arg1 context.Context, arg2 string, arg3 *varcontext.VarContext) (*string, error) {
	fake.deprovisionMutex.Lock()
	ret, specificReturn := fake.deprovisionReturnsOnCall[len(fake.deprovisionArgsForCall)]
	fake.deprovisionArgsForCall = append(fake.deprovisionArgsForCall, struct {
		arg1 context.Context
		arg2 string
		arg3 *varcontext.VarContext
	}{arg1, arg2, arg3})
	stub := fake.DeprovisionStub
	fakeReturns := fake.deprovisionReturns
	fake.recordInvocation("Deprovision", []interface{}{arg1, arg2, arg3})
	fake.deprovisionMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeServiceProvider) DeprovisionCallCount() int {
	fake.deprovisionMutex.RLock()
	defer fake.deprovisionMutex.RUnlock()
	return len(fake.deprovisionArgsForCall)
}

func (fake *FakeServiceProvider) DeprovisionCalls(stub func(context.Context, string, *varcontext.VarContext) (*string, error)) {
	fake.deprovisionMutex.Lock()
	defer fake.deprovisionMutex.Unlock()
	fake.DeprovisionStub = stub
}

func (fake *FakeServiceProvider) DeprovisionArgsForCall(i int) (context.Context, string, *varcontext.VarContext) {
	fake.deprovisionMutex.RLock()
	defer fake.deprovisionMutex.RUnlock()
	argsForCall := fake.deprovisionArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeServiceProvider) DeprovisionReturns(result1 *string, result2 error) {
	fake.deprovisionMutex.Lock()
	defer fake.deprovisionMutex.Unlock()
	fake.DeprovisionStub = nil
	fake.deprovisionReturns = struct {
		result1 *string
		result2 error
	}{result1, result2}
}

func (fake *FakeServiceProvider) DeprovisionReturnsOnCall(i int, result1 *string, result2 error) {
	fake.deprovisionMutex.Lock()
	defer fake.deprovisionMutex.Unlock()
	fake.DeprovisionStub = nil
	if fake.deprovisionReturnsOnCall == nil {
		fake.deprovisionReturnsOnCall = make(map[int]struct {
			result1 *string
			result2 error
		})
	}
	fake.deprovisionReturnsOnCall[i] = struct {
		result1 *string
		result2 error
	}{result1, result2}
}

func (fake *FakeServiceProvider) GetImportedProperties(arg1 context.Context, arg2 string, arg3 []broker.BrokerVariable, arg4 map[string]any) (map[string]any, error) {
	var arg3Copy []broker.BrokerVariable
	if arg3 != nil {
		arg3Copy = make([]broker.BrokerVariable, len(arg3))
		copy(arg3Copy, arg3)
	}
	fake.getImportedPropertiesMutex.Lock()
	ret, specificReturn := fake.getImportedPropertiesReturnsOnCall[len(fake.getImportedPropertiesArgsForCall)]
	fake.getImportedPropertiesArgsForCall = append(fake.getImportedPropertiesArgsForCall, struct {
		arg1 context.Context
		arg2 string
		arg3 []broker.BrokerVariable
		arg4 map[string]any
	}{arg1, arg2, arg3Copy, arg4})
	stub := fake.GetImportedPropertiesStub
	fakeReturns := fake.getImportedPropertiesReturns
	fake.recordInvocation("GetImportedProperties", []interface{}{arg1, arg2, arg3Copy, arg4})
	fake.getImportedPropertiesMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeServiceProvider) GetImportedPropertiesCallCount() int {
	fake.getImportedPropertiesMutex.RLock()
	defer fake.getImportedPropertiesMutex.RUnlock()
	return len(fake.getImportedPropertiesArgsForCall)
}

func (fake *FakeServiceProvider) GetImportedPropertiesCalls(stub func(context.Context, string, []broker.BrokerVariable, map[string]any) (map[string]any, error)) {
	fake.getImportedPropertiesMutex.Lock()
	defer fake.getImportedPropertiesMutex.Unlock()
	fake.GetImportedPropertiesStub = stub
}

func (fake *FakeServiceProvider) GetImportedPropertiesArgsForCall(i int) (context.Context, string, []broker.BrokerVariable, map[string]any) {
	fake.getImportedPropertiesMutex.RLock()
	defer fake.getImportedPropertiesMutex.RUnlock()
	argsForCall := fake.getImportedPropertiesArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4
}

func (fake *FakeServiceProvider) GetImportedPropertiesReturns(result1 map[string]any, result2 error) {
	fake.getImportedPropertiesMutex.Lock()
	defer fake.getImportedPropertiesMutex.Unlock()
	fake.GetImportedPropertiesStub = nil
	fake.getImportedPropertiesReturns = struct {
		result1 map[string]any
		result2 error
	}{result1, result2}
}

func (fake *FakeServiceProvider) GetImportedPropertiesReturnsOnCall(i int, result1 map[string]any, result2 error) {
	fake.getImportedPropertiesMutex.Lock()
	defer fake.getImportedPropertiesMutex.Unlock()
	fake.GetImportedPropertiesStub = nil
	if fake.getImportedPropertiesReturnsOnCall == nil {
		fake.getImportedPropertiesReturnsOnCall = make(map[int]struct {
			result1 map[string]any
			result2 error
		})
	}
	fake.getImportedPropertiesReturnsOnCall[i] = struct {
		result1 map[string]any
		result2 error
	}{result1, result2}
}

func (fake *FakeServiceProvider) GetTerraformOutputs(arg1 context.Context, arg2 string) (storage.JSONObject, error) {
	fake.getTerraformOutputsMutex.Lock()
	ret, specificReturn := fake.getTerraformOutputsReturnsOnCall[len(fake.getTerraformOutputsArgsForCall)]
	fake.getTerraformOutputsArgsForCall = append(fake.getTerraformOutputsArgsForCall, struct {
		arg1 context.Context
		arg2 string
	}{arg1, arg2})
	stub := fake.GetTerraformOutputsStub
	fakeReturns := fake.getTerraformOutputsReturns
	fake.recordInvocation("GetTerraformOutputs", []interface{}{arg1, arg2})
	fake.getTerraformOutputsMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeServiceProvider) GetTerraformOutputsCallCount() int {
	fake.getTerraformOutputsMutex.RLock()
	defer fake.getTerraformOutputsMutex.RUnlock()
	return len(fake.getTerraformOutputsArgsForCall)
}

func (fake *FakeServiceProvider) GetTerraformOutputsCalls(stub func(context.Context, string) (storage.JSONObject, error)) {
	fake.getTerraformOutputsMutex.Lock()
	defer fake.getTerraformOutputsMutex.Unlock()
	fake.GetTerraformOutputsStub = stub
}

func (fake *FakeServiceProvider) GetTerraformOutputsArgsForCall(i int) (context.Context, string) {
	fake.getTerraformOutputsMutex.RLock()
	defer fake.getTerraformOutputsMutex.RUnlock()
	argsForCall := fake.getTerraformOutputsArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeServiceProvider) GetTerraformOutputsReturns(result1 storage.JSONObject, result2 error) {
	fake.getTerraformOutputsMutex.Lock()
	defer fake.getTerraformOutputsMutex.Unlock()
	fake.GetTerraformOutputsStub = nil
	fake.getTerraformOutputsReturns = struct {
		result1 storage.JSONObject
		result2 error
	}{result1, result2}
}

func (fake *FakeServiceProvider) GetTerraformOutputsReturnsOnCall(i int, result1 storage.JSONObject, result2 error) {
	fake.getTerraformOutputsMutex.Lock()
	defer fake.getTerraformOutputsMutex.Unlock()
	fake.GetTerraformOutputsStub = nil
	if fake.getTerraformOutputsReturnsOnCall == nil {
		fake.getTerraformOutputsReturnsOnCall = make(map[int]struct {
			result1 storage.JSONObject
			result2 error
		})
	}
	fake.getTerraformOutputsReturnsOnCall[i] = struct {
		result1 storage.JSONObject
		result2 error
	}{result1, result2}
}

func (fake *FakeServiceProvider) PollInstance(arg1 context.Context, arg2 string) (bool, string, error) {
	fake.pollInstanceMutex.Lock()
	ret, specificReturn := fake.pollInstanceReturnsOnCall[len(fake.pollInstanceArgsForCall)]
	fake.pollInstanceArgsForCall = append(fake.pollInstanceArgsForCall, struct {
		arg1 context.Context
		arg2 string
	}{arg1, arg2})
	stub := fake.PollInstanceStub
	fakeReturns := fake.pollInstanceReturns
	fake.recordInvocation("PollInstance", []interface{}{arg1, arg2})
	fake.pollInstanceMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2, ret.result3
	}
	return fakeReturns.result1, fakeReturns.result2, fakeReturns.result3
}

func (fake *FakeServiceProvider) PollInstanceCallCount() int {
	fake.pollInstanceMutex.RLock()
	defer fake.pollInstanceMutex.RUnlock()
	return len(fake.pollInstanceArgsForCall)
}

func (fake *FakeServiceProvider) PollInstanceCalls(stub func(context.Context, string) (bool, string, error)) {
	fake.pollInstanceMutex.Lock()
	defer fake.pollInstanceMutex.Unlock()
	fake.PollInstanceStub = stub
}

func (fake *FakeServiceProvider) PollInstanceArgsForCall(i int) (context.Context, string) {
	fake.pollInstanceMutex.RLock()
	defer fake.pollInstanceMutex.RUnlock()
	argsForCall := fake.pollInstanceArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeServiceProvider) PollInstanceReturns(result1 bool, result2 string, result3 error) {
	fake.pollInstanceMutex.Lock()
	defer fake.pollInstanceMutex.Unlock()
	fake.PollInstanceStub = nil
	fake.pollInstanceReturns = struct {
		result1 bool
		result2 string
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeServiceProvider) PollInstanceReturnsOnCall(i int, result1 bool, result2 string, result3 error) {
	fake.pollInstanceMutex.Lock()
	defer fake.pollInstanceMutex.Unlock()
	fake.PollInstanceStub = nil
	if fake.pollInstanceReturnsOnCall == nil {
		fake.pollInstanceReturnsOnCall = make(map[int]struct {
			result1 bool
			result2 string
			result3 error
		})
	}
	fake.pollInstanceReturnsOnCall[i] = struct {
		result1 bool
		result2 string
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeServiceProvider) Provision(arg1 context.Context, arg2 *varcontext.VarContext) (storage.ServiceInstanceDetails, error) {
	fake.provisionMutex.Lock()
	ret, specificReturn := fake.provisionReturnsOnCall[len(fake.provisionArgsForCall)]
	fake.provisionArgsForCall = append(fake.provisionArgsForCall, struct {
		arg1 context.Context
		arg2 *varcontext.VarContext
	}{arg1, arg2})
	stub := fake.ProvisionStub
	fakeReturns := fake.provisionReturns
	fake.recordInvocation("Provision", []interface{}{arg1, arg2})
	fake.provisionMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeServiceProvider) ProvisionCallCount() int {
	fake.provisionMutex.RLock()
	defer fake.provisionMutex.RUnlock()
	return len(fake.provisionArgsForCall)
}

func (fake *FakeServiceProvider) ProvisionCalls(stub func(context.Context, *varcontext.VarContext) (storage.ServiceInstanceDetails, error)) {
	fake.provisionMutex.Lock()
	defer fake.provisionMutex.Unlock()
	fake.ProvisionStub = stub
}

func (fake *FakeServiceProvider) ProvisionArgsForCall(i int) (context.Context, *varcontext.VarContext) {
	fake.provisionMutex.RLock()
	defer fake.provisionMutex.RUnlock()
	argsForCall := fake.provisionArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeServiceProvider) ProvisionReturns(result1 storage.ServiceInstanceDetails, result2 error) {
	fake.provisionMutex.Lock()
	defer fake.provisionMutex.Unlock()
	fake.ProvisionStub = nil
	fake.provisionReturns = struct {
		result1 storage.ServiceInstanceDetails
		result2 error
	}{result1, result2}
}

func (fake *FakeServiceProvider) ProvisionReturnsOnCall(i int, result1 storage.ServiceInstanceDetails, result2 error) {
	fake.provisionMutex.Lock()
	defer fake.provisionMutex.Unlock()
	fake.ProvisionStub = nil
	if fake.provisionReturnsOnCall == nil {
		fake.provisionReturnsOnCall = make(map[int]struct {
			result1 storage.ServiceInstanceDetails
			result2 error
		})
	}
	fake.provisionReturnsOnCall[i] = struct {
		result1 storage.ServiceInstanceDetails
		result2 error
	}{result1, result2}
}

func (fake *FakeServiceProvider) Unbind(arg1 context.Context, arg2 string, arg3 string, arg4 *varcontext.VarContext) error {
	fake.unbindMutex.Lock()
	ret, specificReturn := fake.unbindReturnsOnCall[len(fake.unbindArgsForCall)]
	fake.unbindArgsForCall = append(fake.unbindArgsForCall, struct {
		arg1 context.Context
		arg2 string
		arg3 string
		arg4 *varcontext.VarContext
	}{arg1, arg2, arg3, arg4})
	stub := fake.UnbindStub
	fakeReturns := fake.unbindReturns
	fake.recordInvocation("Unbind", []interface{}{arg1, arg2, arg3, arg4})
	fake.unbindMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeServiceProvider) UnbindCallCount() int {
	fake.unbindMutex.RLock()
	defer fake.unbindMutex.RUnlock()
	return len(fake.unbindArgsForCall)
}

func (fake *FakeServiceProvider) UnbindCalls(stub func(context.Context, string, string, *varcontext.VarContext) error) {
	fake.unbindMutex.Lock()
	defer fake.unbindMutex.Unlock()
	fake.UnbindStub = stub
}

func (fake *FakeServiceProvider) UnbindArgsForCall(i int) (context.Context, string, string, *varcontext.VarContext) {
	fake.unbindMutex.RLock()
	defer fake.unbindMutex.RUnlock()
	argsForCall := fake.unbindArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4
}

func (fake *FakeServiceProvider) UnbindReturns(result1 error) {
	fake.unbindMutex.Lock()
	defer fake.unbindMutex.Unlock()
	fake.UnbindStub = nil
	fake.unbindReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeServiceProvider) UnbindReturnsOnCall(i int, result1 error) {
	fake.unbindMutex.Lock()
	defer fake.unbindMutex.Unlock()
	fake.UnbindStub = nil
	if fake.unbindReturnsOnCall == nil {
		fake.unbindReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.unbindReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeServiceProvider) Update(arg1 context.Context, arg2 *varcontext.VarContext) (models.ServiceInstanceDetails, error) {
	fake.updateMutex.Lock()
	ret, specificReturn := fake.updateReturnsOnCall[len(fake.updateArgsForCall)]
	fake.updateArgsForCall = append(fake.updateArgsForCall, struct {
		arg1 context.Context
		arg2 *varcontext.VarContext
	}{arg1, arg2})
	stub := fake.UpdateStub
	fakeReturns := fake.updateReturns
	fake.recordInvocation("Update", []interface{}{arg1, arg2})
	fake.updateMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeServiceProvider) UpdateCallCount() int {
	fake.updateMutex.RLock()
	defer fake.updateMutex.RUnlock()
	return len(fake.updateArgsForCall)
}

func (fake *FakeServiceProvider) UpdateCalls(stub func(context.Context, *varcontext.VarContext) (models.ServiceInstanceDetails, error)) {
	fake.updateMutex.Lock()
	defer fake.updateMutex.Unlock()
	fake.UpdateStub = stub
}

func (fake *FakeServiceProvider) UpdateArgsForCall(i int) (context.Context, *varcontext.VarContext) {
	fake.updateMutex.RLock()
	defer fake.updateMutex.RUnlock()
	argsForCall := fake.updateArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeServiceProvider) UpdateReturns(result1 models.ServiceInstanceDetails, result2 error) {
	fake.updateMutex.Lock()
	defer fake.updateMutex.Unlock()
	fake.UpdateStub = nil
	fake.updateReturns = struct {
		result1 models.ServiceInstanceDetails
		result2 error
	}{result1, result2}
}

func (fake *FakeServiceProvider) UpdateReturnsOnCall(i int, result1 models.ServiceInstanceDetails, result2 error) {
	fake.updateMutex.Lock()
	defer fake.updateMutex.Unlock()
	fake.UpdateStub = nil
	if fake.updateReturnsOnCall == nil {
		fake.updateReturnsOnCall = make(map[int]struct {
			result1 models.ServiceInstanceDetails
			result2 error
		})
	}
	fake.updateReturnsOnCall[i] = struct {
		result1 models.ServiceInstanceDetails
		result2 error
	}{result1, result2}
}

func (fake *FakeServiceProvider) UpgradeBindings(arg1 context.Context, arg2 *varcontext.VarContext, arg3 []*varcontext.VarContext) error {
	var arg3Copy []*varcontext.VarContext
	if arg3 != nil {
		arg3Copy = make([]*varcontext.VarContext, len(arg3))
		copy(arg3Copy, arg3)
	}
	fake.upgradeBindingsMutex.Lock()
	ret, specificReturn := fake.upgradeBindingsReturnsOnCall[len(fake.upgradeBindingsArgsForCall)]
	fake.upgradeBindingsArgsForCall = append(fake.upgradeBindingsArgsForCall, struct {
		arg1 context.Context
		arg2 *varcontext.VarContext
		arg3 []*varcontext.VarContext
	}{arg1, arg2, arg3Copy})
	stub := fake.UpgradeBindingsStub
	fakeReturns := fake.upgradeBindingsReturns
	fake.recordInvocation("UpgradeBindings", []interface{}{arg1, arg2, arg3Copy})
	fake.upgradeBindingsMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeServiceProvider) UpgradeBindingsCallCount() int {
	fake.upgradeBindingsMutex.RLock()
	defer fake.upgradeBindingsMutex.RUnlock()
	return len(fake.upgradeBindingsArgsForCall)
}

func (fake *FakeServiceProvider) UpgradeBindingsCalls(stub func(context.Context, *varcontext.VarContext, []*varcontext.VarContext) error) {
	fake.upgradeBindingsMutex.Lock()
	defer fake.upgradeBindingsMutex.Unlock()
	fake.UpgradeBindingsStub = stub
}

func (fake *FakeServiceProvider) UpgradeBindingsArgsForCall(i int) (context.Context, *varcontext.VarContext, []*varcontext.VarContext) {
	fake.upgradeBindingsMutex.RLock()
	defer fake.upgradeBindingsMutex.RUnlock()
	argsForCall := fake.upgradeBindingsArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeServiceProvider) UpgradeBindingsReturns(result1 error) {
	fake.upgradeBindingsMutex.Lock()
	defer fake.upgradeBindingsMutex.Unlock()
	fake.UpgradeBindingsStub = nil
	fake.upgradeBindingsReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeServiceProvider) UpgradeBindingsReturnsOnCall(i int, result1 error) {
	fake.upgradeBindingsMutex.Lock()
	defer fake.upgradeBindingsMutex.Unlock()
	fake.UpgradeBindingsStub = nil
	if fake.upgradeBindingsReturnsOnCall == nil {
		fake.upgradeBindingsReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.upgradeBindingsReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeServiceProvider) UpgradeInstance(arg1 context.Context, arg2 *varcontext.VarContext) (*sync.WaitGroup, error) {
	fake.upgradeInstanceMutex.Lock()
	ret, specificReturn := fake.upgradeInstanceReturnsOnCall[len(fake.upgradeInstanceArgsForCall)]
	fake.upgradeInstanceArgsForCall = append(fake.upgradeInstanceArgsForCall, struct {
		arg1 context.Context
		arg2 *varcontext.VarContext
	}{arg1, arg2})
	stub := fake.UpgradeInstanceStub
	fakeReturns := fake.upgradeInstanceReturns
	fake.recordInvocation("UpgradeInstance", []interface{}{arg1, arg2})
	fake.upgradeInstanceMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeServiceProvider) UpgradeInstanceCallCount() int {
	fake.upgradeInstanceMutex.RLock()
	defer fake.upgradeInstanceMutex.RUnlock()
	return len(fake.upgradeInstanceArgsForCall)
}

func (fake *FakeServiceProvider) UpgradeInstanceCalls(stub func(context.Context, *varcontext.VarContext) (*sync.WaitGroup, error)) {
	fake.upgradeInstanceMutex.Lock()
	defer fake.upgradeInstanceMutex.Unlock()
	fake.UpgradeInstanceStub = stub
}

func (fake *FakeServiceProvider) UpgradeInstanceArgsForCall(i int) (context.Context, *varcontext.VarContext) {
	fake.upgradeInstanceMutex.RLock()
	defer fake.upgradeInstanceMutex.RUnlock()
	argsForCall := fake.upgradeInstanceArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeServiceProvider) UpgradeInstanceReturns(result1 *sync.WaitGroup, result2 error) {
	fake.upgradeInstanceMutex.Lock()
	defer fake.upgradeInstanceMutex.Unlock()
	fake.UpgradeInstanceStub = nil
	fake.upgradeInstanceReturns = struct {
		result1 *sync.WaitGroup
		result2 error
	}{result1, result2}
}

func (fake *FakeServiceProvider) UpgradeInstanceReturnsOnCall(i int, result1 *sync.WaitGroup, result2 error) {
	fake.upgradeInstanceMutex.Lock()
	defer fake.upgradeInstanceMutex.Unlock()
	fake.UpgradeInstanceStub = nil
	if fake.upgradeInstanceReturnsOnCall == nil {
		fake.upgradeInstanceReturnsOnCall = make(map[int]struct {
			result1 *sync.WaitGroup
			result2 error
		})
	}
	fake.upgradeInstanceReturnsOnCall[i] = struct {
		result1 *sync.WaitGroup
		result2 error
	}{result1, result2}
}

func (fake *FakeServiceProvider) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.bindMutex.RLock()
	defer fake.bindMutex.RUnlock()
	fake.checkOperationConstraintsMutex.RLock()
	defer fake.checkOperationConstraintsMutex.RUnlock()
	fake.checkUpgradeAvailableMutex.RLock()
	defer fake.checkUpgradeAvailableMutex.RUnlock()
	fake.deleteBindingDataMutex.RLock()
	defer fake.deleteBindingDataMutex.RUnlock()
	fake.deleteInstanceDataMutex.RLock()
	defer fake.deleteInstanceDataMutex.RUnlock()
	fake.deprovisionMutex.RLock()
	defer fake.deprovisionMutex.RUnlock()
	fake.getImportedPropertiesMutex.RLock()
	defer fake.getImportedPropertiesMutex.RUnlock()
	fake.getTerraformOutputsMutex.RLock()
	defer fake.getTerraformOutputsMutex.RUnlock()
	fake.pollInstanceMutex.RLock()
	defer fake.pollInstanceMutex.RUnlock()
	fake.provisionMutex.RLock()
	defer fake.provisionMutex.RUnlock()
	fake.unbindMutex.RLock()
	defer fake.unbindMutex.RUnlock()
	fake.updateMutex.RLock()
	defer fake.updateMutex.RUnlock()
	fake.upgradeBindingsMutex.RLock()
	defer fake.upgradeBindingsMutex.RUnlock()
	fake.upgradeInstanceMutex.RLock()
	defer fake.upgradeInstanceMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeServiceProvider) recordInvocation(key string, args []interface{}) {
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

var _ broker.ServiceProvider = new(FakeServiceProvider)
