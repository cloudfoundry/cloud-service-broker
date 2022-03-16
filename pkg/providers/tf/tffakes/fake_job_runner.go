// Code generated by counterfeiter. DO NOT EDIT.
package tffakes

import (
	"context"
	"sync"

	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/workspace"
)

type FakeJobRunner struct {
	CreateStub        func(context.Context, string) error
	createMutex       sync.RWMutex
	createArgsForCall []struct {
		arg1 context.Context
		arg2 string
	}
	createReturns struct {
		result1 error
	}
	createReturnsOnCall map[int]struct {
		result1 error
	}
	DestroyStub        func(context.Context, string, map[string]interface{}) error
	destroyMutex       sync.RWMutex
	destroyArgsForCall []struct {
		arg1 context.Context
		arg2 string
		arg3 map[string]interface{}
	}
	destroyReturns struct {
		result1 error
	}
	destroyReturnsOnCall map[int]struct {
		result1 error
	}
	ImportStub        func(context.Context, string, []tf.ImportResource) error
	importMutex       sync.RWMutex
	importArgsForCall []struct {
		arg1 context.Context
		arg2 string
		arg3 []tf.ImportResource
	}
	importReturns struct {
		result1 error
	}
	importReturnsOnCall map[int]struct {
		result1 error
	}
	OutputsStub        func(context.Context, string, string) (map[string]interface{}, error)
	outputsMutex       sync.RWMutex
	outputsArgsForCall []struct {
		arg1 context.Context
		arg2 string
		arg3 string
	}
	outputsReturns struct {
		result1 map[string]interface{}
		result2 error
	}
	outputsReturnsOnCall map[int]struct {
		result1 map[string]interface{}
		result2 error
	}
	ShowStub        func(context.Context, string) (string, error)
	showMutex       sync.RWMutex
	showArgsForCall []struct {
		arg1 context.Context
		arg2 string
	}
	showReturns struct {
		result1 string
		result2 error
	}
	showReturnsOnCall map[int]struct {
		result1 string
		result2 error
	}
	StageJobStub        func(string, *workspace.TerraformWorkspace) error
	stageJobMutex       sync.RWMutex
	stageJobArgsForCall []struct {
		arg1 string
		arg2 *workspace.TerraformWorkspace
	}
	stageJobReturns struct {
		result1 error
	}
	stageJobReturnsOnCall map[int]struct {
		result1 error
	}
	StatusStub        func(context.Context, string) (bool, string, error)
	statusMutex       sync.RWMutex
	statusArgsForCall []struct {
		arg1 context.Context
		arg2 string
	}
	statusReturns struct {
		result1 bool
		result2 string
		result3 error
	}
	statusReturnsOnCall map[int]struct {
		result1 bool
		result2 string
		result3 error
	}
	UpdateStub        func(context.Context, string, map[string]interface{}) error
	updateMutex       sync.RWMutex
	updateArgsForCall []struct {
		arg1 context.Context
		arg2 string
		arg3 map[string]interface{}
	}
	updateReturns struct {
		result1 error
	}
	updateReturnsOnCall map[int]struct {
		result1 error
	}
	WaitStub        func(context.Context, string) error
	waitMutex       sync.RWMutex
	waitArgsForCall []struct {
		arg1 context.Context
		arg2 string
	}
	waitReturns struct {
		result1 error
	}
	waitReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeJobRunner) Create(arg1 context.Context, arg2 string) error {
	fake.createMutex.Lock()
	ret, specificReturn := fake.createReturnsOnCall[len(fake.createArgsForCall)]
	fake.createArgsForCall = append(fake.createArgsForCall, struct {
		arg1 context.Context
		arg2 string
	}{arg1, arg2})
	stub := fake.CreateStub
	fakeReturns := fake.createReturns
	fake.recordInvocation("Create", []interface{}{arg1, arg2})
	fake.createMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeJobRunner) CreateCallCount() int {
	fake.createMutex.RLock()
	defer fake.createMutex.RUnlock()
	return len(fake.createArgsForCall)
}

func (fake *FakeJobRunner) CreateCalls(stub func(context.Context, string) error) {
	fake.createMutex.Lock()
	defer fake.createMutex.Unlock()
	fake.CreateStub = stub
}

func (fake *FakeJobRunner) CreateArgsForCall(i int) (context.Context, string) {
	fake.createMutex.RLock()
	defer fake.createMutex.RUnlock()
	argsForCall := fake.createArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeJobRunner) CreateReturns(result1 error) {
	fake.createMutex.Lock()
	defer fake.createMutex.Unlock()
	fake.CreateStub = nil
	fake.createReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeJobRunner) CreateReturnsOnCall(i int, result1 error) {
	fake.createMutex.Lock()
	defer fake.createMutex.Unlock()
	fake.CreateStub = nil
	if fake.createReturnsOnCall == nil {
		fake.createReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.createReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeJobRunner) Destroy(arg1 context.Context, arg2 string, arg3 map[string]interface{}) error {
	fake.destroyMutex.Lock()
	ret, specificReturn := fake.destroyReturnsOnCall[len(fake.destroyArgsForCall)]
	fake.destroyArgsForCall = append(fake.destroyArgsForCall, struct {
		arg1 context.Context
		arg2 string
		arg3 map[string]interface{}
	}{arg1, arg2, arg3})
	stub := fake.DestroyStub
	fakeReturns := fake.destroyReturns
	fake.recordInvocation("Destroy", []interface{}{arg1, arg2, arg3})
	fake.destroyMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeJobRunner) DestroyCallCount() int {
	fake.destroyMutex.RLock()
	defer fake.destroyMutex.RUnlock()
	return len(fake.destroyArgsForCall)
}

func (fake *FakeJobRunner) DestroyCalls(stub func(context.Context, string, map[string]interface{}) error) {
	fake.destroyMutex.Lock()
	defer fake.destroyMutex.Unlock()
	fake.DestroyStub = stub
}

func (fake *FakeJobRunner) DestroyArgsForCall(i int) (context.Context, string, map[string]interface{}) {
	fake.destroyMutex.RLock()
	defer fake.destroyMutex.RUnlock()
	argsForCall := fake.destroyArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeJobRunner) DestroyReturns(result1 error) {
	fake.destroyMutex.Lock()
	defer fake.destroyMutex.Unlock()
	fake.DestroyStub = nil
	fake.destroyReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeJobRunner) DestroyReturnsOnCall(i int, result1 error) {
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

func (fake *FakeJobRunner) Import(arg1 context.Context, arg2 string, arg3 []tf.ImportResource) error {
	var arg3Copy []tf.ImportResource
	if arg3 != nil {
		arg3Copy = make([]tf.ImportResource, len(arg3))
		copy(arg3Copy, arg3)
	}
	fake.importMutex.Lock()
	ret, specificReturn := fake.importReturnsOnCall[len(fake.importArgsForCall)]
	fake.importArgsForCall = append(fake.importArgsForCall, struct {
		arg1 context.Context
		arg2 string
		arg3 []tf.ImportResource
	}{arg1, arg2, arg3Copy})
	stub := fake.ImportStub
	fakeReturns := fake.importReturns
	fake.recordInvocation("Import", []interface{}{arg1, arg2, arg3Copy})
	fake.importMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeJobRunner) ImportCallCount() int {
	fake.importMutex.RLock()
	defer fake.importMutex.RUnlock()
	return len(fake.importArgsForCall)
}

func (fake *FakeJobRunner) ImportCalls(stub func(context.Context, string, []tf.ImportResource) error) {
	fake.importMutex.Lock()
	defer fake.importMutex.Unlock()
	fake.ImportStub = stub
}

func (fake *FakeJobRunner) ImportArgsForCall(i int) (context.Context, string, []tf.ImportResource) {
	fake.importMutex.RLock()
	defer fake.importMutex.RUnlock()
	argsForCall := fake.importArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeJobRunner) ImportReturns(result1 error) {
	fake.importMutex.Lock()
	defer fake.importMutex.Unlock()
	fake.ImportStub = nil
	fake.importReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeJobRunner) ImportReturnsOnCall(i int, result1 error) {
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

func (fake *FakeJobRunner) Outputs(arg1 context.Context, arg2 string, arg3 string) (map[string]interface{}, error) {
	fake.outputsMutex.Lock()
	ret, specificReturn := fake.outputsReturnsOnCall[len(fake.outputsArgsForCall)]
	fake.outputsArgsForCall = append(fake.outputsArgsForCall, struct {
		arg1 context.Context
		arg2 string
		arg3 string
	}{arg1, arg2, arg3})
	stub := fake.OutputsStub
	fakeReturns := fake.outputsReturns
	fake.recordInvocation("Outputs", []interface{}{arg1, arg2, arg3})
	fake.outputsMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeJobRunner) OutputsCallCount() int {
	fake.outputsMutex.RLock()
	defer fake.outputsMutex.RUnlock()
	return len(fake.outputsArgsForCall)
}

func (fake *FakeJobRunner) OutputsCalls(stub func(context.Context, string, string) (map[string]interface{}, error)) {
	fake.outputsMutex.Lock()
	defer fake.outputsMutex.Unlock()
	fake.OutputsStub = stub
}

func (fake *FakeJobRunner) OutputsArgsForCall(i int) (context.Context, string, string) {
	fake.outputsMutex.RLock()
	defer fake.outputsMutex.RUnlock()
	argsForCall := fake.outputsArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeJobRunner) OutputsReturns(result1 map[string]interface{}, result2 error) {
	fake.outputsMutex.Lock()
	defer fake.outputsMutex.Unlock()
	fake.OutputsStub = nil
	fake.outputsReturns = struct {
		result1 map[string]interface{}
		result2 error
	}{result1, result2}
}

func (fake *FakeJobRunner) OutputsReturnsOnCall(i int, result1 map[string]interface{}, result2 error) {
	fake.outputsMutex.Lock()
	defer fake.outputsMutex.Unlock()
	fake.OutputsStub = nil
	if fake.outputsReturnsOnCall == nil {
		fake.outputsReturnsOnCall = make(map[int]struct {
			result1 map[string]interface{}
			result2 error
		})
	}
	fake.outputsReturnsOnCall[i] = struct {
		result1 map[string]interface{}
		result2 error
	}{result1, result2}
}

func (fake *FakeJobRunner) Show(arg1 context.Context, arg2 string) (string, error) {
	fake.showMutex.Lock()
	ret, specificReturn := fake.showReturnsOnCall[len(fake.showArgsForCall)]
	fake.showArgsForCall = append(fake.showArgsForCall, struct {
		arg1 context.Context
		arg2 string
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

func (fake *FakeJobRunner) ShowCallCount() int {
	fake.showMutex.RLock()
	defer fake.showMutex.RUnlock()
	return len(fake.showArgsForCall)
}

func (fake *FakeJobRunner) ShowCalls(stub func(context.Context, string) (string, error)) {
	fake.showMutex.Lock()
	defer fake.showMutex.Unlock()
	fake.ShowStub = stub
}

func (fake *FakeJobRunner) ShowArgsForCall(i int) (context.Context, string) {
	fake.showMutex.RLock()
	defer fake.showMutex.RUnlock()
	argsForCall := fake.showArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeJobRunner) ShowReturns(result1 string, result2 error) {
	fake.showMutex.Lock()
	defer fake.showMutex.Unlock()
	fake.ShowStub = nil
	fake.showReturns = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *FakeJobRunner) ShowReturnsOnCall(i int, result1 string, result2 error) {
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

func (fake *FakeJobRunner) StageJob(arg1 string, arg2 *workspace.TerraformWorkspace) error {
	fake.stageJobMutex.Lock()
	ret, specificReturn := fake.stageJobReturnsOnCall[len(fake.stageJobArgsForCall)]
	fake.stageJobArgsForCall = append(fake.stageJobArgsForCall, struct {
		arg1 string
		arg2 *workspace.TerraformWorkspace
	}{arg1, arg2})
	stub := fake.StageJobStub
	fakeReturns := fake.stageJobReturns
	fake.recordInvocation("StageJob", []interface{}{arg1, arg2})
	fake.stageJobMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeJobRunner) StageJobCallCount() int {
	fake.stageJobMutex.RLock()
	defer fake.stageJobMutex.RUnlock()
	return len(fake.stageJobArgsForCall)
}

func (fake *FakeJobRunner) StageJobCalls(stub func(string, *workspace.TerraformWorkspace) error) {
	fake.stageJobMutex.Lock()
	defer fake.stageJobMutex.Unlock()
	fake.StageJobStub = stub
}

func (fake *FakeJobRunner) StageJobArgsForCall(i int) (string, *workspace.TerraformWorkspace) {
	fake.stageJobMutex.RLock()
	defer fake.stageJobMutex.RUnlock()
	argsForCall := fake.stageJobArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeJobRunner) StageJobReturns(result1 error) {
	fake.stageJobMutex.Lock()
	defer fake.stageJobMutex.Unlock()
	fake.StageJobStub = nil
	fake.stageJobReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeJobRunner) StageJobReturnsOnCall(i int, result1 error) {
	fake.stageJobMutex.Lock()
	defer fake.stageJobMutex.Unlock()
	fake.StageJobStub = nil
	if fake.stageJobReturnsOnCall == nil {
		fake.stageJobReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.stageJobReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeJobRunner) Status(arg1 context.Context, arg2 string) (bool, string, error) {
	fake.statusMutex.Lock()
	ret, specificReturn := fake.statusReturnsOnCall[len(fake.statusArgsForCall)]
	fake.statusArgsForCall = append(fake.statusArgsForCall, struct {
		arg1 context.Context
		arg2 string
	}{arg1, arg2})
	stub := fake.StatusStub
	fakeReturns := fake.statusReturns
	fake.recordInvocation("Status", []interface{}{arg1, arg2})
	fake.statusMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2, ret.result3
	}
	return fakeReturns.result1, fakeReturns.result2, fakeReturns.result3
}

func (fake *FakeJobRunner) StatusCallCount() int {
	fake.statusMutex.RLock()
	defer fake.statusMutex.RUnlock()
	return len(fake.statusArgsForCall)
}

func (fake *FakeJobRunner) StatusCalls(stub func(context.Context, string) (bool, string, error)) {
	fake.statusMutex.Lock()
	defer fake.statusMutex.Unlock()
	fake.StatusStub = stub
}

func (fake *FakeJobRunner) StatusArgsForCall(i int) (context.Context, string) {
	fake.statusMutex.RLock()
	defer fake.statusMutex.RUnlock()
	argsForCall := fake.statusArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeJobRunner) StatusReturns(result1 bool, result2 string, result3 error) {
	fake.statusMutex.Lock()
	defer fake.statusMutex.Unlock()
	fake.StatusStub = nil
	fake.statusReturns = struct {
		result1 bool
		result2 string
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeJobRunner) StatusReturnsOnCall(i int, result1 bool, result2 string, result3 error) {
	fake.statusMutex.Lock()
	defer fake.statusMutex.Unlock()
	fake.StatusStub = nil
	if fake.statusReturnsOnCall == nil {
		fake.statusReturnsOnCall = make(map[int]struct {
			result1 bool
			result2 string
			result3 error
		})
	}
	fake.statusReturnsOnCall[i] = struct {
		result1 bool
		result2 string
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeJobRunner) Update(arg1 context.Context, arg2 string, arg3 map[string]interface{}) error {
	fake.updateMutex.Lock()
	ret, specificReturn := fake.updateReturnsOnCall[len(fake.updateArgsForCall)]
	fake.updateArgsForCall = append(fake.updateArgsForCall, struct {
		arg1 context.Context
		arg2 string
		arg3 map[string]interface{}
	}{arg1, arg2, arg3})
	stub := fake.UpdateStub
	fakeReturns := fake.updateReturns
	fake.recordInvocation("Update", []interface{}{arg1, arg2, arg3})
	fake.updateMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeJobRunner) UpdateCallCount() int {
	fake.updateMutex.RLock()
	defer fake.updateMutex.RUnlock()
	return len(fake.updateArgsForCall)
}

func (fake *FakeJobRunner) UpdateCalls(stub func(context.Context, string, map[string]interface{}) error) {
	fake.updateMutex.Lock()
	defer fake.updateMutex.Unlock()
	fake.UpdateStub = stub
}

func (fake *FakeJobRunner) UpdateArgsForCall(i int) (context.Context, string, map[string]interface{}) {
	fake.updateMutex.RLock()
	defer fake.updateMutex.RUnlock()
	argsForCall := fake.updateArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeJobRunner) UpdateReturns(result1 error) {
	fake.updateMutex.Lock()
	defer fake.updateMutex.Unlock()
	fake.UpdateStub = nil
	fake.updateReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeJobRunner) UpdateReturnsOnCall(i int, result1 error) {
	fake.updateMutex.Lock()
	defer fake.updateMutex.Unlock()
	fake.UpdateStub = nil
	if fake.updateReturnsOnCall == nil {
		fake.updateReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.updateReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeJobRunner) Wait(arg1 context.Context, arg2 string) error {
	fake.waitMutex.Lock()
	ret, specificReturn := fake.waitReturnsOnCall[len(fake.waitArgsForCall)]
	fake.waitArgsForCall = append(fake.waitArgsForCall, struct {
		arg1 context.Context
		arg2 string
	}{arg1, arg2})
	stub := fake.WaitStub
	fakeReturns := fake.waitReturns
	fake.recordInvocation("Wait", []interface{}{arg1, arg2})
	fake.waitMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeJobRunner) WaitCallCount() int {
	fake.waitMutex.RLock()
	defer fake.waitMutex.RUnlock()
	return len(fake.waitArgsForCall)
}

func (fake *FakeJobRunner) WaitCalls(stub func(context.Context, string) error) {
	fake.waitMutex.Lock()
	defer fake.waitMutex.Unlock()
	fake.WaitStub = stub
}

func (fake *FakeJobRunner) WaitArgsForCall(i int) (context.Context, string) {
	fake.waitMutex.RLock()
	defer fake.waitMutex.RUnlock()
	argsForCall := fake.waitArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeJobRunner) WaitReturns(result1 error) {
	fake.waitMutex.Lock()
	defer fake.waitMutex.Unlock()
	fake.WaitStub = nil
	fake.waitReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeJobRunner) WaitReturnsOnCall(i int, result1 error) {
	fake.waitMutex.Lock()
	defer fake.waitMutex.Unlock()
	fake.WaitStub = nil
	if fake.waitReturnsOnCall == nil {
		fake.waitReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.waitReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeJobRunner) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.createMutex.RLock()
	defer fake.createMutex.RUnlock()
	fake.destroyMutex.RLock()
	defer fake.destroyMutex.RUnlock()
	fake.importMutex.RLock()
	defer fake.importMutex.RUnlock()
	fake.outputsMutex.RLock()
	defer fake.outputsMutex.RUnlock()
	fake.showMutex.RLock()
	defer fake.showMutex.RUnlock()
	fake.stageJobMutex.RLock()
	defer fake.stageJobMutex.RUnlock()
	fake.statusMutex.RLock()
	defer fake.statusMutex.RUnlock()
	fake.updateMutex.RLock()
	defer fake.updateMutex.RUnlock()
	fake.waitMutex.RLock()
	defer fake.waitMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeJobRunner) recordInvocation(key string, args []interface{}) {
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

var _ tf.JobRunner = new(FakeJobRunner)
