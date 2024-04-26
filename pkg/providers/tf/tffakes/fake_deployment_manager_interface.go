// Code generated by counterfeiter. DO NOT EDIT.
package tffakes

import (
	"sync"

	"github.com/cloudfoundry/cloud-service-broker/v3/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/providers/tf"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/providers/tf/workspace"
)

type FakeDeploymentManagerInterface struct {
	CreateAndSaveDeploymentStub        func(string, *workspace.TerraformWorkspace) (storage.TerraformDeployment, error)
	createAndSaveDeploymentMutex       sync.RWMutex
	createAndSaveDeploymentArgsForCall []struct {
		arg1 string
		arg2 *workspace.TerraformWorkspace
	}
	createAndSaveDeploymentReturns struct {
		result1 storage.TerraformDeployment
		result2 error
	}
	createAndSaveDeploymentReturnsOnCall map[int]struct {
		result1 storage.TerraformDeployment
		result2 error
	}
	DeleteTerraformDeploymentStub        func(string) error
	deleteTerraformDeploymentMutex       sync.RWMutex
	deleteTerraformDeploymentArgsForCall []struct {
		arg1 string
	}
	deleteTerraformDeploymentReturns struct {
		result1 error
	}
	deleteTerraformDeploymentReturnsOnCall map[int]struct {
		result1 error
	}
	GetBindingDeploymentsStub        func(string) ([]storage.TerraformDeployment, error)
	getBindingDeploymentsMutex       sync.RWMutex
	getBindingDeploymentsArgsForCall []struct {
		arg1 string
	}
	getBindingDeploymentsReturns struct {
		result1 []storage.TerraformDeployment
		result2 error
	}
	getBindingDeploymentsReturnsOnCall map[int]struct {
		result1 []storage.TerraformDeployment
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
	MarkOperationFinishedStub        func(*storage.TerraformDeployment, error) error
	markOperationFinishedMutex       sync.RWMutex
	markOperationFinishedArgsForCall []struct {
		arg1 *storage.TerraformDeployment
		arg2 error
	}
	markOperationFinishedReturns struct {
		result1 error
	}
	markOperationFinishedReturnsOnCall map[int]struct {
		result1 error
	}
	MarkOperationStartedStub        func(*storage.TerraformDeployment, string) error
	markOperationStartedMutex       sync.RWMutex
	markOperationStartedArgsForCall []struct {
		arg1 *storage.TerraformDeployment
		arg2 string
	}
	markOperationStartedReturns struct {
		result1 error
	}
	markOperationStartedReturnsOnCall map[int]struct {
		result1 error
	}
	OperationStatusStub        func(string) (bool, string, error)
	operationStatusMutex       sync.RWMutex
	operationStatusArgsForCall []struct {
		arg1 string
	}
	operationStatusReturns struct {
		result1 bool
		result2 string
		result3 error
	}
	operationStatusReturnsOnCall map[int]struct {
		result1 bool
		result2 string
		result3 error
	}
	UpdateWorkspaceHCLStub        func(string, tf.TfServiceDefinitionV1Action, map[string]any) error
	updateWorkspaceHCLMutex       sync.RWMutex
	updateWorkspaceHCLArgsForCall []struct {
		arg1 string
		arg2 tf.TfServiceDefinitionV1Action
		arg3 map[string]any
	}
	updateWorkspaceHCLReturns struct {
		result1 error
	}
	updateWorkspaceHCLReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeDeploymentManagerInterface) CreateAndSaveDeployment(arg1 string, arg2 *workspace.TerraformWorkspace) (storage.TerraformDeployment, error) {
	fake.createAndSaveDeploymentMutex.Lock()
	ret, specificReturn := fake.createAndSaveDeploymentReturnsOnCall[len(fake.createAndSaveDeploymentArgsForCall)]
	fake.createAndSaveDeploymentArgsForCall = append(fake.createAndSaveDeploymentArgsForCall, struct {
		arg1 string
		arg2 *workspace.TerraformWorkspace
	}{arg1, arg2})
	stub := fake.CreateAndSaveDeploymentStub
	fakeReturns := fake.createAndSaveDeploymentReturns
	fake.recordInvocation("CreateAndSaveDeployment", []interface{}{arg1, arg2})
	fake.createAndSaveDeploymentMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeDeploymentManagerInterface) CreateAndSaveDeploymentCallCount() int {
	fake.createAndSaveDeploymentMutex.RLock()
	defer fake.createAndSaveDeploymentMutex.RUnlock()
	return len(fake.createAndSaveDeploymentArgsForCall)
}

func (fake *FakeDeploymentManagerInterface) CreateAndSaveDeploymentCalls(stub func(string, *workspace.TerraformWorkspace) (storage.TerraformDeployment, error)) {
	fake.createAndSaveDeploymentMutex.Lock()
	defer fake.createAndSaveDeploymentMutex.Unlock()
	fake.CreateAndSaveDeploymentStub = stub
}

func (fake *FakeDeploymentManagerInterface) CreateAndSaveDeploymentArgsForCall(i int) (string, *workspace.TerraformWorkspace) {
	fake.createAndSaveDeploymentMutex.RLock()
	defer fake.createAndSaveDeploymentMutex.RUnlock()
	argsForCall := fake.createAndSaveDeploymentArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeDeploymentManagerInterface) CreateAndSaveDeploymentReturns(result1 storage.TerraformDeployment, result2 error) {
	fake.createAndSaveDeploymentMutex.Lock()
	defer fake.createAndSaveDeploymentMutex.Unlock()
	fake.CreateAndSaveDeploymentStub = nil
	fake.createAndSaveDeploymentReturns = struct {
		result1 storage.TerraformDeployment
		result2 error
	}{result1, result2}
}

func (fake *FakeDeploymentManagerInterface) CreateAndSaveDeploymentReturnsOnCall(i int, result1 storage.TerraformDeployment, result2 error) {
	fake.createAndSaveDeploymentMutex.Lock()
	defer fake.createAndSaveDeploymentMutex.Unlock()
	fake.CreateAndSaveDeploymentStub = nil
	if fake.createAndSaveDeploymentReturnsOnCall == nil {
		fake.createAndSaveDeploymentReturnsOnCall = make(map[int]struct {
			result1 storage.TerraformDeployment
			result2 error
		})
	}
	fake.createAndSaveDeploymentReturnsOnCall[i] = struct {
		result1 storage.TerraformDeployment
		result2 error
	}{result1, result2}
}

func (fake *FakeDeploymentManagerInterface) DeleteTerraformDeployment(arg1 string) error {
	fake.deleteTerraformDeploymentMutex.Lock()
	ret, specificReturn := fake.deleteTerraformDeploymentReturnsOnCall[len(fake.deleteTerraformDeploymentArgsForCall)]
	fake.deleteTerraformDeploymentArgsForCall = append(fake.deleteTerraformDeploymentArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.DeleteTerraformDeploymentStub
	fakeReturns := fake.deleteTerraformDeploymentReturns
	fake.recordInvocation("DeleteTerraformDeployment", []interface{}{arg1})
	fake.deleteTerraformDeploymentMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeDeploymentManagerInterface) DeleteTerraformDeploymentCallCount() int {
	fake.deleteTerraformDeploymentMutex.RLock()
	defer fake.deleteTerraformDeploymentMutex.RUnlock()
	return len(fake.deleteTerraformDeploymentArgsForCall)
}

func (fake *FakeDeploymentManagerInterface) DeleteTerraformDeploymentCalls(stub func(string) error) {
	fake.deleteTerraformDeploymentMutex.Lock()
	defer fake.deleteTerraformDeploymentMutex.Unlock()
	fake.DeleteTerraformDeploymentStub = stub
}

func (fake *FakeDeploymentManagerInterface) DeleteTerraformDeploymentArgsForCall(i int) string {
	fake.deleteTerraformDeploymentMutex.RLock()
	defer fake.deleteTerraformDeploymentMutex.RUnlock()
	argsForCall := fake.deleteTerraformDeploymentArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeDeploymentManagerInterface) DeleteTerraformDeploymentReturns(result1 error) {
	fake.deleteTerraformDeploymentMutex.Lock()
	defer fake.deleteTerraformDeploymentMutex.Unlock()
	fake.DeleteTerraformDeploymentStub = nil
	fake.deleteTerraformDeploymentReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeDeploymentManagerInterface) DeleteTerraformDeploymentReturnsOnCall(i int, result1 error) {
	fake.deleteTerraformDeploymentMutex.Lock()
	defer fake.deleteTerraformDeploymentMutex.Unlock()
	fake.DeleteTerraformDeploymentStub = nil
	if fake.deleteTerraformDeploymentReturnsOnCall == nil {
		fake.deleteTerraformDeploymentReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.deleteTerraformDeploymentReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeDeploymentManagerInterface) GetBindingDeployments(arg1 string) ([]storage.TerraformDeployment, error) {
	fake.getBindingDeploymentsMutex.Lock()
	ret, specificReturn := fake.getBindingDeploymentsReturnsOnCall[len(fake.getBindingDeploymentsArgsForCall)]
	fake.getBindingDeploymentsArgsForCall = append(fake.getBindingDeploymentsArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.GetBindingDeploymentsStub
	fakeReturns := fake.getBindingDeploymentsReturns
	fake.recordInvocation("GetBindingDeployments", []interface{}{arg1})
	fake.getBindingDeploymentsMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeDeploymentManagerInterface) GetBindingDeploymentsCallCount() int {
	fake.getBindingDeploymentsMutex.RLock()
	defer fake.getBindingDeploymentsMutex.RUnlock()
	return len(fake.getBindingDeploymentsArgsForCall)
}

func (fake *FakeDeploymentManagerInterface) GetBindingDeploymentsCalls(stub func(string) ([]storage.TerraformDeployment, error)) {
	fake.getBindingDeploymentsMutex.Lock()
	defer fake.getBindingDeploymentsMutex.Unlock()
	fake.GetBindingDeploymentsStub = stub
}

func (fake *FakeDeploymentManagerInterface) GetBindingDeploymentsArgsForCall(i int) string {
	fake.getBindingDeploymentsMutex.RLock()
	defer fake.getBindingDeploymentsMutex.RUnlock()
	argsForCall := fake.getBindingDeploymentsArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeDeploymentManagerInterface) GetBindingDeploymentsReturns(result1 []storage.TerraformDeployment, result2 error) {
	fake.getBindingDeploymentsMutex.Lock()
	defer fake.getBindingDeploymentsMutex.Unlock()
	fake.GetBindingDeploymentsStub = nil
	fake.getBindingDeploymentsReturns = struct {
		result1 []storage.TerraformDeployment
		result2 error
	}{result1, result2}
}

func (fake *FakeDeploymentManagerInterface) GetBindingDeploymentsReturnsOnCall(i int, result1 []storage.TerraformDeployment, result2 error) {
	fake.getBindingDeploymentsMutex.Lock()
	defer fake.getBindingDeploymentsMutex.Unlock()
	fake.GetBindingDeploymentsStub = nil
	if fake.getBindingDeploymentsReturnsOnCall == nil {
		fake.getBindingDeploymentsReturnsOnCall = make(map[int]struct {
			result1 []storage.TerraformDeployment
			result2 error
		})
	}
	fake.getBindingDeploymentsReturnsOnCall[i] = struct {
		result1 []storage.TerraformDeployment
		result2 error
	}{result1, result2}
}

func (fake *FakeDeploymentManagerInterface) GetTerraformDeployment(arg1 string) (storage.TerraformDeployment, error) {
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

func (fake *FakeDeploymentManagerInterface) GetTerraformDeploymentCallCount() int {
	fake.getTerraformDeploymentMutex.RLock()
	defer fake.getTerraformDeploymentMutex.RUnlock()
	return len(fake.getTerraformDeploymentArgsForCall)
}

func (fake *FakeDeploymentManagerInterface) GetTerraformDeploymentCalls(stub func(string) (storage.TerraformDeployment, error)) {
	fake.getTerraformDeploymentMutex.Lock()
	defer fake.getTerraformDeploymentMutex.Unlock()
	fake.GetTerraformDeploymentStub = stub
}

func (fake *FakeDeploymentManagerInterface) GetTerraformDeploymentArgsForCall(i int) string {
	fake.getTerraformDeploymentMutex.RLock()
	defer fake.getTerraformDeploymentMutex.RUnlock()
	argsForCall := fake.getTerraformDeploymentArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeDeploymentManagerInterface) GetTerraformDeploymentReturns(result1 storage.TerraformDeployment, result2 error) {
	fake.getTerraformDeploymentMutex.Lock()
	defer fake.getTerraformDeploymentMutex.Unlock()
	fake.GetTerraformDeploymentStub = nil
	fake.getTerraformDeploymentReturns = struct {
		result1 storage.TerraformDeployment
		result2 error
	}{result1, result2}
}

func (fake *FakeDeploymentManagerInterface) GetTerraformDeploymentReturnsOnCall(i int, result1 storage.TerraformDeployment, result2 error) {
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

func (fake *FakeDeploymentManagerInterface) MarkOperationFinished(arg1 *storage.TerraformDeployment, arg2 error) error {
	fake.markOperationFinishedMutex.Lock()
	ret, specificReturn := fake.markOperationFinishedReturnsOnCall[len(fake.markOperationFinishedArgsForCall)]
	fake.markOperationFinishedArgsForCall = append(fake.markOperationFinishedArgsForCall, struct {
		arg1 *storage.TerraformDeployment
		arg2 error
	}{arg1, arg2})
	stub := fake.MarkOperationFinishedStub
	fakeReturns := fake.markOperationFinishedReturns
	fake.recordInvocation("MarkOperationFinished", []interface{}{arg1, arg2})
	fake.markOperationFinishedMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeDeploymentManagerInterface) MarkOperationFinishedCallCount() int {
	fake.markOperationFinishedMutex.RLock()
	defer fake.markOperationFinishedMutex.RUnlock()
	return len(fake.markOperationFinishedArgsForCall)
}

func (fake *FakeDeploymentManagerInterface) MarkOperationFinishedCalls(stub func(*storage.TerraformDeployment, error) error) {
	fake.markOperationFinishedMutex.Lock()
	defer fake.markOperationFinishedMutex.Unlock()
	fake.MarkOperationFinishedStub = stub
}

func (fake *FakeDeploymentManagerInterface) MarkOperationFinishedArgsForCall(i int) (*storage.TerraformDeployment, error) {
	fake.markOperationFinishedMutex.RLock()
	defer fake.markOperationFinishedMutex.RUnlock()
	argsForCall := fake.markOperationFinishedArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeDeploymentManagerInterface) MarkOperationFinishedReturns(result1 error) {
	fake.markOperationFinishedMutex.Lock()
	defer fake.markOperationFinishedMutex.Unlock()
	fake.MarkOperationFinishedStub = nil
	fake.markOperationFinishedReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeDeploymentManagerInterface) MarkOperationFinishedReturnsOnCall(i int, result1 error) {
	fake.markOperationFinishedMutex.Lock()
	defer fake.markOperationFinishedMutex.Unlock()
	fake.MarkOperationFinishedStub = nil
	if fake.markOperationFinishedReturnsOnCall == nil {
		fake.markOperationFinishedReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.markOperationFinishedReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeDeploymentManagerInterface) MarkOperationStarted(arg1 *storage.TerraformDeployment, arg2 string) error {
	fake.markOperationStartedMutex.Lock()
	ret, specificReturn := fake.markOperationStartedReturnsOnCall[len(fake.markOperationStartedArgsForCall)]
	fake.markOperationStartedArgsForCall = append(fake.markOperationStartedArgsForCall, struct {
		arg1 *storage.TerraformDeployment
		arg2 string
	}{arg1, arg2})
	stub := fake.MarkOperationStartedStub
	fakeReturns := fake.markOperationStartedReturns
	fake.recordInvocation("MarkOperationStarted", []interface{}{arg1, arg2})
	fake.markOperationStartedMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeDeploymentManagerInterface) MarkOperationStartedCallCount() int {
	fake.markOperationStartedMutex.RLock()
	defer fake.markOperationStartedMutex.RUnlock()
	return len(fake.markOperationStartedArgsForCall)
}

func (fake *FakeDeploymentManagerInterface) MarkOperationStartedCalls(stub func(*storage.TerraformDeployment, string) error) {
	fake.markOperationStartedMutex.Lock()
	defer fake.markOperationStartedMutex.Unlock()
	fake.MarkOperationStartedStub = stub
}

func (fake *FakeDeploymentManagerInterface) MarkOperationStartedArgsForCall(i int) (*storage.TerraformDeployment, string) {
	fake.markOperationStartedMutex.RLock()
	defer fake.markOperationStartedMutex.RUnlock()
	argsForCall := fake.markOperationStartedArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeDeploymentManagerInterface) MarkOperationStartedReturns(result1 error) {
	fake.markOperationStartedMutex.Lock()
	defer fake.markOperationStartedMutex.Unlock()
	fake.MarkOperationStartedStub = nil
	fake.markOperationStartedReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeDeploymentManagerInterface) MarkOperationStartedReturnsOnCall(i int, result1 error) {
	fake.markOperationStartedMutex.Lock()
	defer fake.markOperationStartedMutex.Unlock()
	fake.MarkOperationStartedStub = nil
	if fake.markOperationStartedReturnsOnCall == nil {
		fake.markOperationStartedReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.markOperationStartedReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeDeploymentManagerInterface) OperationStatus(arg1 string) (bool, string, error) {
	fake.operationStatusMutex.Lock()
	ret, specificReturn := fake.operationStatusReturnsOnCall[len(fake.operationStatusArgsForCall)]
	fake.operationStatusArgsForCall = append(fake.operationStatusArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.OperationStatusStub
	fakeReturns := fake.operationStatusReturns
	fake.recordInvocation("OperationStatus", []interface{}{arg1})
	fake.operationStatusMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2, ret.result3
	}
	return fakeReturns.result1, fakeReturns.result2, fakeReturns.result3
}

func (fake *FakeDeploymentManagerInterface) OperationStatusCallCount() int {
	fake.operationStatusMutex.RLock()
	defer fake.operationStatusMutex.RUnlock()
	return len(fake.operationStatusArgsForCall)
}

func (fake *FakeDeploymentManagerInterface) OperationStatusCalls(stub func(string) (bool, string, error)) {
	fake.operationStatusMutex.Lock()
	defer fake.operationStatusMutex.Unlock()
	fake.OperationStatusStub = stub
}

func (fake *FakeDeploymentManagerInterface) OperationStatusArgsForCall(i int) string {
	fake.operationStatusMutex.RLock()
	defer fake.operationStatusMutex.RUnlock()
	argsForCall := fake.operationStatusArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeDeploymentManagerInterface) OperationStatusReturns(result1 bool, result2 string, result3 error) {
	fake.operationStatusMutex.Lock()
	defer fake.operationStatusMutex.Unlock()
	fake.OperationStatusStub = nil
	fake.operationStatusReturns = struct {
		result1 bool
		result2 string
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeDeploymentManagerInterface) OperationStatusReturnsOnCall(i int, result1 bool, result2 string, result3 error) {
	fake.operationStatusMutex.Lock()
	defer fake.operationStatusMutex.Unlock()
	fake.OperationStatusStub = nil
	if fake.operationStatusReturnsOnCall == nil {
		fake.operationStatusReturnsOnCall = make(map[int]struct {
			result1 bool
			result2 string
			result3 error
		})
	}
	fake.operationStatusReturnsOnCall[i] = struct {
		result1 bool
		result2 string
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeDeploymentManagerInterface) UpdateWorkspaceHCL(arg1 string, arg2 tf.TfServiceDefinitionV1Action, arg3 map[string]any) error {
	fake.updateWorkspaceHCLMutex.Lock()
	ret, specificReturn := fake.updateWorkspaceHCLReturnsOnCall[len(fake.updateWorkspaceHCLArgsForCall)]
	fake.updateWorkspaceHCLArgsForCall = append(fake.updateWorkspaceHCLArgsForCall, struct {
		arg1 string
		arg2 tf.TfServiceDefinitionV1Action
		arg3 map[string]any
	}{arg1, arg2, arg3})
	stub := fake.UpdateWorkspaceHCLStub
	fakeReturns := fake.updateWorkspaceHCLReturns
	fake.recordInvocation("UpdateWorkspaceHCL", []interface{}{arg1, arg2, arg3})
	fake.updateWorkspaceHCLMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeDeploymentManagerInterface) UpdateWorkspaceHCLCallCount() int {
	fake.updateWorkspaceHCLMutex.RLock()
	defer fake.updateWorkspaceHCLMutex.RUnlock()
	return len(fake.updateWorkspaceHCLArgsForCall)
}

func (fake *FakeDeploymentManagerInterface) UpdateWorkspaceHCLCalls(stub func(string, tf.TfServiceDefinitionV1Action, map[string]any) error) {
	fake.updateWorkspaceHCLMutex.Lock()
	defer fake.updateWorkspaceHCLMutex.Unlock()
	fake.UpdateWorkspaceHCLStub = stub
}

func (fake *FakeDeploymentManagerInterface) UpdateWorkspaceHCLArgsForCall(i int) (string, tf.TfServiceDefinitionV1Action, map[string]any) {
	fake.updateWorkspaceHCLMutex.RLock()
	defer fake.updateWorkspaceHCLMutex.RUnlock()
	argsForCall := fake.updateWorkspaceHCLArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeDeploymentManagerInterface) UpdateWorkspaceHCLReturns(result1 error) {
	fake.updateWorkspaceHCLMutex.Lock()
	defer fake.updateWorkspaceHCLMutex.Unlock()
	fake.UpdateWorkspaceHCLStub = nil
	fake.updateWorkspaceHCLReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeDeploymentManagerInterface) UpdateWorkspaceHCLReturnsOnCall(i int, result1 error) {
	fake.updateWorkspaceHCLMutex.Lock()
	defer fake.updateWorkspaceHCLMutex.Unlock()
	fake.UpdateWorkspaceHCLStub = nil
	if fake.updateWorkspaceHCLReturnsOnCall == nil {
		fake.updateWorkspaceHCLReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.updateWorkspaceHCLReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeDeploymentManagerInterface) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.createAndSaveDeploymentMutex.RLock()
	defer fake.createAndSaveDeploymentMutex.RUnlock()
	fake.deleteTerraformDeploymentMutex.RLock()
	defer fake.deleteTerraformDeploymentMutex.RUnlock()
	fake.getBindingDeploymentsMutex.RLock()
	defer fake.getBindingDeploymentsMutex.RUnlock()
	fake.getTerraformDeploymentMutex.RLock()
	defer fake.getTerraformDeploymentMutex.RUnlock()
	fake.markOperationFinishedMutex.RLock()
	defer fake.markOperationFinishedMutex.RUnlock()
	fake.markOperationStartedMutex.RLock()
	defer fake.markOperationStartedMutex.RUnlock()
	fake.operationStatusMutex.RLock()
	defer fake.operationStatusMutex.RUnlock()
	fake.updateWorkspaceHCLMutex.RLock()
	defer fake.updateWorkspaceHCLMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeDeploymentManagerInterface) recordInvocation(key string, args []interface{}) {
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

var _ tf.DeploymentManagerInterface = new(FakeDeploymentManagerInterface)
