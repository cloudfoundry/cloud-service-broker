// Code generated by counterfeiter. DO NOT EDIT.
package invokerfakes

import (
	"context"
	"sync"

	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/providers/tf/executor"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/providers/tf/invoker"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/providers/tf/workspace"
)

type FakeTerraformInvoker struct {
	ApplyStub        func(context.Context, workspace.Workspace) error
	applyMutex       sync.RWMutex
	applyArgsForCall []struct {
		arg1 context.Context
		arg2 workspace.Workspace
	}
	applyReturns struct {
		result1 error
	}
	applyReturnsOnCall map[int]struct {
		result1 error
	}
	DestroyStub        func(context.Context, workspace.Workspace) error
	destroyMutex       sync.RWMutex
	destroyArgsForCall []struct {
		arg1 context.Context
		arg2 workspace.Workspace
	}
	destroyReturns struct {
		result1 error
	}
	destroyReturnsOnCall map[int]struct {
		result1 error
	}
	ImportStub        func(context.Context, workspace.Workspace, map[string]string) error
	importMutex       sync.RWMutex
	importArgsForCall []struct {
		arg1 context.Context
		arg2 workspace.Workspace
		arg3 map[string]string
	}
	importReturns struct {
		result1 error
	}
	importReturnsOnCall map[int]struct {
		result1 error
	}
	PlanStub        func(context.Context, workspace.Workspace) (executor.ExecutionOutput, error)
	planMutex       sync.RWMutex
	planArgsForCall []struct {
		arg1 context.Context
		arg2 workspace.Workspace
	}
	planReturns struct {
		result1 executor.ExecutionOutput
		result2 error
	}
	planReturnsOnCall map[int]struct {
		result1 executor.ExecutionOutput
		result2 error
	}
	ShowStub        func(context.Context, workspace.Workspace) (string, error)
	showMutex       sync.RWMutex
	showArgsForCall []struct {
		arg1 context.Context
		arg2 workspace.Workspace
	}
	showReturns struct {
		result1 string
		result2 error
	}
	showReturnsOnCall map[int]struct {
		result1 string
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeTerraformInvoker) Apply(arg1 context.Context, arg2 workspace.Workspace) error {
	fake.applyMutex.Lock()
	ret, specificReturn := fake.applyReturnsOnCall[len(fake.applyArgsForCall)]
	fake.applyArgsForCall = append(fake.applyArgsForCall, struct {
		arg1 context.Context
		arg2 workspace.Workspace
	}{arg1, arg2})
	stub := fake.ApplyStub
	fakeReturns := fake.applyReturns
	fake.recordInvocation("Apply", []interface{}{arg1, arg2})
	fake.applyMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeTerraformInvoker) ApplyCallCount() int {
	fake.applyMutex.RLock()
	defer fake.applyMutex.RUnlock()
	return len(fake.applyArgsForCall)
}

func (fake *FakeTerraformInvoker) ApplyCalls(stub func(context.Context, workspace.Workspace) error) {
	fake.applyMutex.Lock()
	defer fake.applyMutex.Unlock()
	fake.ApplyStub = stub
}

func (fake *FakeTerraformInvoker) ApplyArgsForCall(i int) (context.Context, workspace.Workspace) {
	fake.applyMutex.RLock()
	defer fake.applyMutex.RUnlock()
	argsForCall := fake.applyArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeTerraformInvoker) ApplyReturns(result1 error) {
	fake.applyMutex.Lock()
	defer fake.applyMutex.Unlock()
	fake.ApplyStub = nil
	fake.applyReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeTerraformInvoker) ApplyReturnsOnCall(i int, result1 error) {
	fake.applyMutex.Lock()
	defer fake.applyMutex.Unlock()
	fake.ApplyStub = nil
	if fake.applyReturnsOnCall == nil {
		fake.applyReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.applyReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeTerraformInvoker) Destroy(arg1 context.Context, arg2 workspace.Workspace) error {
	fake.destroyMutex.Lock()
	ret, specificReturn := fake.destroyReturnsOnCall[len(fake.destroyArgsForCall)]
	fake.destroyArgsForCall = append(fake.destroyArgsForCall, struct {
		arg1 context.Context
		arg2 workspace.Workspace
	}{arg1, arg2})
	stub := fake.DestroyStub
	fakeReturns := fake.destroyReturns
	fake.recordInvocation("Destroy", []interface{}{arg1, arg2})
	fake.destroyMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeTerraformInvoker) DestroyCallCount() int {
	fake.destroyMutex.RLock()
	defer fake.destroyMutex.RUnlock()
	return len(fake.destroyArgsForCall)
}

func (fake *FakeTerraformInvoker) DestroyCalls(stub func(context.Context, workspace.Workspace) error) {
	fake.destroyMutex.Lock()
	defer fake.destroyMutex.Unlock()
	fake.DestroyStub = stub
}

func (fake *FakeTerraformInvoker) DestroyArgsForCall(i int) (context.Context, workspace.Workspace) {
	fake.destroyMutex.RLock()
	defer fake.destroyMutex.RUnlock()
	argsForCall := fake.destroyArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeTerraformInvoker) DestroyReturns(result1 error) {
	fake.destroyMutex.Lock()
	defer fake.destroyMutex.Unlock()
	fake.DestroyStub = nil
	fake.destroyReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeTerraformInvoker) DestroyReturnsOnCall(i int, result1 error) {
	fake.destroyMutex.Lock()
	defer fake.destroyMutex.Unlock()
	fake.DestroyStub = nil
	if fake.destroyReturnsOnCall == nil {
		fake.destroyReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.destroyReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeTerraformInvoker) Import(arg1 context.Context, arg2 workspace.Workspace, arg3 map[string]string) error {
	fake.importMutex.Lock()
	ret, specificReturn := fake.importReturnsOnCall[len(fake.importArgsForCall)]
	fake.importArgsForCall = append(fake.importArgsForCall, struct {
		arg1 context.Context
		arg2 workspace.Workspace
		arg3 map[string]string
	}{arg1, arg2, arg3})
	stub := fake.ImportStub
	fakeReturns := fake.importReturns
	fake.recordInvocation("Import", []interface{}{arg1, arg2, arg3})
	fake.importMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeTerraformInvoker) ImportCallCount() int {
	fake.importMutex.RLock()
	defer fake.importMutex.RUnlock()
	return len(fake.importArgsForCall)
}

func (fake *FakeTerraformInvoker) ImportCalls(stub func(context.Context, workspace.Workspace, map[string]string) error) {
	fake.importMutex.Lock()
	defer fake.importMutex.Unlock()
	fake.ImportStub = stub
}

func (fake *FakeTerraformInvoker) ImportArgsForCall(i int) (context.Context, workspace.Workspace, map[string]string) {
	fake.importMutex.RLock()
	defer fake.importMutex.RUnlock()
	argsForCall := fake.importArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeTerraformInvoker) ImportReturns(result1 error) {
	fake.importMutex.Lock()
	defer fake.importMutex.Unlock()
	fake.ImportStub = nil
	fake.importReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeTerraformInvoker) ImportReturnsOnCall(i int, result1 error) {
	fake.importMutex.Lock()
	defer fake.importMutex.Unlock()
	fake.ImportStub = nil
	if fake.importReturnsOnCall == nil {
		fake.importReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.importReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeTerraformInvoker) Plan(arg1 context.Context, arg2 workspace.Workspace) (executor.ExecutionOutput, error) {
	fake.planMutex.Lock()
	ret, specificReturn := fake.planReturnsOnCall[len(fake.planArgsForCall)]
	fake.planArgsForCall = append(fake.planArgsForCall, struct {
		arg1 context.Context
		arg2 workspace.Workspace
	}{arg1, arg2})
	stub := fake.PlanStub
	fakeReturns := fake.planReturns
	fake.recordInvocation("Plan", []interface{}{arg1, arg2})
	fake.planMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeTerraformInvoker) PlanCallCount() int {
	fake.planMutex.RLock()
	defer fake.planMutex.RUnlock()
	return len(fake.planArgsForCall)
}

func (fake *FakeTerraformInvoker) PlanCalls(stub func(context.Context, workspace.Workspace) (executor.ExecutionOutput, error)) {
	fake.planMutex.Lock()
	defer fake.planMutex.Unlock()
	fake.PlanStub = stub
}

func (fake *FakeTerraformInvoker) PlanArgsForCall(i int) (context.Context, workspace.Workspace) {
	fake.planMutex.RLock()
	defer fake.planMutex.RUnlock()
	argsForCall := fake.planArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeTerraformInvoker) PlanReturns(result1 executor.ExecutionOutput, result2 error) {
	fake.planMutex.Lock()
	defer fake.planMutex.Unlock()
	fake.PlanStub = nil
	fake.planReturns = struct {
		result1 executor.ExecutionOutput
		result2 error
	}{result1, result2}
}

func (fake *FakeTerraformInvoker) PlanReturnsOnCall(i int, result1 executor.ExecutionOutput, result2 error) {
	fake.planMutex.Lock()
	defer fake.planMutex.Unlock()
	fake.PlanStub = nil
	if fake.planReturnsOnCall == nil {
		fake.planReturnsOnCall = make(map[int]struct {
			result1 executor.ExecutionOutput
			result2 error
		})
	}
	fake.planReturnsOnCall[i] = struct {
		result1 executor.ExecutionOutput
		result2 error
	}{result1, result2}
}

func (fake *FakeTerraformInvoker) Show(arg1 context.Context, arg2 workspace.Workspace) (string, error) {
	fake.showMutex.Lock()
	ret, specificReturn := fake.showReturnsOnCall[len(fake.showArgsForCall)]
	fake.showArgsForCall = append(fake.showArgsForCall, struct {
		arg1 context.Context
		arg2 workspace.Workspace
	}{arg1, arg2})
	stub := fake.ShowStub
	fakeReturns := fake.showReturns
	fake.recordInvocation("Show", []interface{}{arg1, arg2})
	fake.showMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeTerraformInvoker) ShowCallCount() int {
	fake.showMutex.RLock()
	defer fake.showMutex.RUnlock()
	return len(fake.showArgsForCall)
}

func (fake *FakeTerraformInvoker) ShowCalls(stub func(context.Context, workspace.Workspace) (string, error)) {
	fake.showMutex.Lock()
	defer fake.showMutex.Unlock()
	fake.ShowStub = stub
}

func (fake *FakeTerraformInvoker) ShowArgsForCall(i int) (context.Context, workspace.Workspace) {
	fake.showMutex.RLock()
	defer fake.showMutex.RUnlock()
	argsForCall := fake.showArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeTerraformInvoker) ShowReturns(result1 string, result2 error) {
	fake.showMutex.Lock()
	defer fake.showMutex.Unlock()
	fake.ShowStub = nil
	fake.showReturns = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *FakeTerraformInvoker) ShowReturnsOnCall(i int, result1 string, result2 error) {
	fake.showMutex.Lock()
	defer fake.showMutex.Unlock()
	fake.ShowStub = nil
	if fake.showReturnsOnCall == nil {
		fake.showReturnsOnCall = make(map[int]struct {
			result1 string
			result2 error
		})
	}
	fake.showReturnsOnCall[i] = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *FakeTerraformInvoker) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.applyMutex.RLock()
	defer fake.applyMutex.RUnlock()
	fake.destroyMutex.RLock()
	defer fake.destroyMutex.RUnlock()
	fake.importMutex.RLock()
	defer fake.importMutex.RUnlock()
	fake.planMutex.RLock()
	defer fake.planMutex.RUnlock()
	fake.showMutex.RLock()
	defer fake.showMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeTerraformInvoker) recordInvocation(key string, args []interface{}) {
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

var _ invoker.TerraformInvoker = new(FakeTerraformInvoker)
