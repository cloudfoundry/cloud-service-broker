// Code generated by counterfeiter. DO NOT EDIT.
package workspacefakes

import (
	"sync"

	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/workspace"
)

type FakeWorkspaceBuilder struct {
	CreateWorkspaceStub        func(storage.TerraformDeployment) (workspace.Workspace, error)
	createWorkspaceMutex       sync.RWMutex
	createWorkspaceArgsForCall []struct {
		arg1 storage.TerraformDeployment
	}
	createWorkspaceReturns struct {
		result1 workspace.Workspace
		result2 error
	}
	createWorkspaceReturnsOnCall map[int]struct {
		result1 workspace.Workspace
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeWorkspaceBuilder) CreateWorkspace(arg1 storage.TerraformDeployment) (workspace.Workspace, error) {
	fake.createWorkspaceMutex.Lock()
	ret, specificReturn := fake.createWorkspaceReturnsOnCall[len(fake.createWorkspaceArgsForCall)]
	fake.createWorkspaceArgsForCall = append(fake.createWorkspaceArgsForCall, struct {
		arg1 storage.TerraformDeployment
	}{arg1})
	stub := fake.CreateWorkspaceStub
	fakeReturns := fake.createWorkspaceReturns
	fake.recordInvocation("CreateWorkspace", []interface{}{arg1})
	fake.createWorkspaceMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeWorkspaceBuilder) CreateWorkspaceCallCount() int {
	fake.createWorkspaceMutex.RLock()
	defer fake.createWorkspaceMutex.RUnlock()
	return len(fake.createWorkspaceArgsForCall)
}

func (fake *FakeWorkspaceBuilder) CreateWorkspaceCalls(stub func(storage.TerraformDeployment) (workspace.Workspace, error)) {
	fake.createWorkspaceMutex.Lock()
	defer fake.createWorkspaceMutex.Unlock()
	fake.CreateWorkspaceStub = stub
}

func (fake *FakeWorkspaceBuilder) CreateWorkspaceArgsForCall(i int) storage.TerraformDeployment {
	fake.createWorkspaceMutex.RLock()
	defer fake.createWorkspaceMutex.RUnlock()
	argsForCall := fake.createWorkspaceArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeWorkspaceBuilder) CreateWorkspaceReturns(result1 workspace.Workspace, result2 error) {
	fake.createWorkspaceMutex.Lock()
	defer fake.createWorkspaceMutex.Unlock()
	fake.CreateWorkspaceStub = nil
	fake.createWorkspaceReturns = struct {
		result1 workspace.Workspace
		result2 error
	}{result1, result2}
}

func (fake *FakeWorkspaceBuilder) CreateWorkspaceReturnsOnCall(i int, result1 workspace.Workspace, result2 error) {
	fake.createWorkspaceMutex.Lock()
	defer fake.createWorkspaceMutex.Unlock()
	fake.CreateWorkspaceStub = nil
	if fake.createWorkspaceReturnsOnCall == nil {
		fake.createWorkspaceReturnsOnCall = make(map[int]struct {
			result1 workspace.Workspace
			result2 error
		})
	}
	fake.createWorkspaceReturnsOnCall[i] = struct {
		result1 workspace.Workspace
		result2 error
	}{result1, result2}
}

func (fake *FakeWorkspaceBuilder) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.createWorkspaceMutex.RLock()
	defer fake.createWorkspaceMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeWorkspaceBuilder) recordInvocation(key string, args []interface{}) {
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

var _ workspace.WorkspaceBuilder = new(FakeWorkspaceBuilder)
