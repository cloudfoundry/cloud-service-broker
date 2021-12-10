// Code generated by counterfeiter. DO NOT EDIT.
package brokerfakes

import (
	"context"
	"sync"

	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/broker"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/varcontext"
	"github.com/pivotal-cf/brokerapi/v8/domain"
)

type FakeServiceProvider struct {
	BindStub        func(context.Context, *varcontext.VarContext) (map[string]interface{}, error)
	bindMutex       sync.RWMutex
	bindArgsForCall []struct {
		arg1 context.Context
		arg2 *varcontext.VarContext
	}
	bindReturns struct {
		result1 map[string]interface{}
		result2 error
	}
	bindReturnsOnCall map[int]struct {
		result1 map[string]interface{}
		result2 error
	}
	BuildInstanceCredentialsStub        func(context.Context, map[string]interface{}, models.ServiceInstanceDetails) (*domain.Binding, error)
	buildInstanceCredentialsMutex       sync.RWMutex
	buildInstanceCredentialsArgsForCall []struct {
		arg1 context.Context
		arg2 map[string]interface{}
		arg3 models.ServiceInstanceDetails
	}
	buildInstanceCredentialsReturns struct {
		result1 *domain.Binding
		result2 error
	}
	buildInstanceCredentialsReturnsOnCall map[int]struct {
		result1 *domain.Binding
		result2 error
	}
	DeprovisionStub        func(context.Context, models.ServiceInstanceDetails, domain.DeprovisionDetails, *varcontext.VarContext) (*string, error)
	deprovisionMutex       sync.RWMutex
	deprovisionArgsForCall []struct {
		arg1 context.Context
		arg2 models.ServiceInstanceDetails
		arg3 domain.DeprovisionDetails
		arg4 *varcontext.VarContext
	}
	deprovisionReturns struct {
		result1 *string
		result2 error
	}
	deprovisionReturnsOnCall map[int]struct {
		result1 *string
		result2 error
	}
	DeprovisionsAsyncStub        func() bool
	deprovisionsAsyncMutex       sync.RWMutex
	deprovisionsAsyncArgsForCall []struct {
	}
	deprovisionsAsyncReturns struct {
		result1 bool
	}
	deprovisionsAsyncReturnsOnCall map[int]struct {
		result1 bool
	}
	PollInstanceStub        func(context.Context, models.ServiceInstanceDetails) (bool, string, error)
	pollInstanceMutex       sync.RWMutex
	pollInstanceArgsForCall []struct {
		arg1 context.Context
		arg2 models.ServiceInstanceDetails
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
	ProvisionStub        func(context.Context, *varcontext.VarContext) (models.ServiceInstanceDetails, error)
	provisionMutex       sync.RWMutex
	provisionArgsForCall []struct {
		arg1 context.Context
		arg2 *varcontext.VarContext
	}
	provisionReturns struct {
		result1 models.ServiceInstanceDetails
		result2 error
	}
	provisionReturnsOnCall map[int]struct {
		result1 models.ServiceInstanceDetails
		result2 error
	}
	ProvisionsAsyncStub        func() bool
	provisionsAsyncMutex       sync.RWMutex
	provisionsAsyncArgsForCall []struct {
	}
	provisionsAsyncReturns struct {
		result1 bool
	}
	provisionsAsyncReturnsOnCall map[int]struct {
		result1 bool
	}
	UnbindStub        func(context.Context, models.ServiceInstanceDetails, string, *varcontext.VarContext) error
	unbindMutex       sync.RWMutex
	unbindArgsForCall []struct {
		arg1 context.Context
		arg2 models.ServiceInstanceDetails
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
	UpdateInstanceDetailsStub        func(context.Context, *models.ServiceInstanceDetails) error
	updateInstanceDetailsMutex       sync.RWMutex
	updateInstanceDetailsArgsForCall []struct {
		arg1 context.Context
		arg2 *models.ServiceInstanceDetails
	}
	updateInstanceDetailsReturns struct {
		result1 error
	}
	updateInstanceDetailsReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeServiceProvider) Bind(arg1 context.Context, arg2 *varcontext.VarContext) (map[string]interface{}, error) {
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

func (fake *FakeServiceProvider) BindCalls(stub func(context.Context, *varcontext.VarContext) (map[string]interface{}, error)) {
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

func (fake *FakeServiceProvider) BindReturns(result1 map[string]interface{}, result2 error) {
	fake.bindMutex.Lock()
	defer fake.bindMutex.Unlock()
	fake.BindStub = nil
	fake.bindReturns = struct {
		result1 map[string]interface{}
		result2 error
	}{result1, result2}
}

func (fake *FakeServiceProvider) BindReturnsOnCall(i int, result1 map[string]interface{}, result2 error) {
	fake.bindMutex.Lock()
	defer fake.bindMutex.Unlock()
	fake.BindStub = nil
	if fake.bindReturnsOnCall == nil {
		fake.bindReturnsOnCall = make(map[int]struct {
			result1 map[string]interface{}
			result2 error
		})
	}
	fake.bindReturnsOnCall[i] = struct {
		result1 map[string]interface{}
		result2 error
	}{result1, result2}
}

func (fake *FakeServiceProvider) BuildInstanceCredentials(arg1 context.Context, arg2 map[string]interface{}, arg3 models.ServiceInstanceDetails) (*domain.Binding, error) {
	fake.buildInstanceCredentialsMutex.Lock()
	ret, specificReturn := fake.buildInstanceCredentialsReturnsOnCall[len(fake.buildInstanceCredentialsArgsForCall)]
	fake.buildInstanceCredentialsArgsForCall = append(fake.buildInstanceCredentialsArgsForCall, struct {
		arg1 context.Context
		arg2 map[string]interface{}
		arg3 models.ServiceInstanceDetails
	}{arg1, arg2, arg3})
	stub := fake.BuildInstanceCredentialsStub
	fakeReturns := fake.buildInstanceCredentialsReturns
	fake.recordInvocation("BuildInstanceCredentials", []interface{}{arg1, arg2, arg3})
	fake.buildInstanceCredentialsMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeServiceProvider) BuildInstanceCredentialsCallCount() int {
	fake.buildInstanceCredentialsMutex.RLock()
	defer fake.buildInstanceCredentialsMutex.RUnlock()
	return len(fake.buildInstanceCredentialsArgsForCall)
}

func (fake *FakeServiceProvider) BuildInstanceCredentialsCalls(stub func(context.Context, map[string]interface{}, models.ServiceInstanceDetails) (*domain.Binding, error)) {
	fake.buildInstanceCredentialsMutex.Lock()
	defer fake.buildInstanceCredentialsMutex.Unlock()
	fake.BuildInstanceCredentialsStub = stub
}

func (fake *FakeServiceProvider) BuildInstanceCredentialsArgsForCall(i int) (context.Context, map[string]interface{}, models.ServiceInstanceDetails) {
	fake.buildInstanceCredentialsMutex.RLock()
	defer fake.buildInstanceCredentialsMutex.RUnlock()
	argsForCall := fake.buildInstanceCredentialsArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeServiceProvider) BuildInstanceCredentialsReturns(result1 *domain.Binding, result2 error) {
	fake.buildInstanceCredentialsMutex.Lock()
	defer fake.buildInstanceCredentialsMutex.Unlock()
	fake.BuildInstanceCredentialsStub = nil
	fake.buildInstanceCredentialsReturns = struct {
		result1 *domain.Binding
		result2 error
	}{result1, result2}
}

func (fake *FakeServiceProvider) BuildInstanceCredentialsReturnsOnCall(i int, result1 *domain.Binding, result2 error) {
	fake.buildInstanceCredentialsMutex.Lock()
	defer fake.buildInstanceCredentialsMutex.Unlock()
	fake.BuildInstanceCredentialsStub = nil
	if fake.buildInstanceCredentialsReturnsOnCall == nil {
		fake.buildInstanceCredentialsReturnsOnCall = make(map[int]struct {
			result1 *domain.Binding
			result2 error
		})
	}
	fake.buildInstanceCredentialsReturnsOnCall[i] = struct {
		result1 *domain.Binding
		result2 error
	}{result1, result2}
}

func (fake *FakeServiceProvider) Deprovision(arg1 context.Context, arg2 models.ServiceInstanceDetails, arg3 domain.DeprovisionDetails, arg4 *varcontext.VarContext) (*string, error) {
	fake.deprovisionMutex.Lock()
	ret, specificReturn := fake.deprovisionReturnsOnCall[len(fake.deprovisionArgsForCall)]
	fake.deprovisionArgsForCall = append(fake.deprovisionArgsForCall, struct {
		arg1 context.Context
		arg2 models.ServiceInstanceDetails
		arg3 domain.DeprovisionDetails
		arg4 *varcontext.VarContext
	}{arg1, arg2, arg3, arg4})
	stub := fake.DeprovisionStub
	fakeReturns := fake.deprovisionReturns
	fake.recordInvocation("Deprovision", []interface{}{arg1, arg2, arg3, arg4})
	fake.deprovisionMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4)
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

func (fake *FakeServiceProvider) DeprovisionCalls(stub func(context.Context, models.ServiceInstanceDetails, domain.DeprovisionDetails, *varcontext.VarContext) (*string, error)) {
	fake.deprovisionMutex.Lock()
	defer fake.deprovisionMutex.Unlock()
	fake.DeprovisionStub = stub
}

func (fake *FakeServiceProvider) DeprovisionArgsForCall(i int) (context.Context, models.ServiceInstanceDetails, domain.DeprovisionDetails, *varcontext.VarContext) {
	fake.deprovisionMutex.RLock()
	defer fake.deprovisionMutex.RUnlock()
	argsForCall := fake.deprovisionArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4
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

func (fake *FakeServiceProvider) DeprovisionsAsync() bool {
	fake.deprovisionsAsyncMutex.Lock()
	ret, specificReturn := fake.deprovisionsAsyncReturnsOnCall[len(fake.deprovisionsAsyncArgsForCall)]
	fake.deprovisionsAsyncArgsForCall = append(fake.deprovisionsAsyncArgsForCall, struct {
	}{})
	stub := fake.DeprovisionsAsyncStub
	fakeReturns := fake.deprovisionsAsyncReturns
	fake.recordInvocation("DeprovisionsAsync", []interface{}{})
	fake.deprovisionsAsyncMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeServiceProvider) DeprovisionsAsyncCallCount() int {
	fake.deprovisionsAsyncMutex.RLock()
	defer fake.deprovisionsAsyncMutex.RUnlock()
	return len(fake.deprovisionsAsyncArgsForCall)
}

func (fake *FakeServiceProvider) DeprovisionsAsyncCalls(stub func() bool) {
	fake.deprovisionsAsyncMutex.Lock()
	defer fake.deprovisionsAsyncMutex.Unlock()
	fake.DeprovisionsAsyncStub = stub
}

func (fake *FakeServiceProvider) DeprovisionsAsyncReturns(result1 bool) {
	fake.deprovisionsAsyncMutex.Lock()
	defer fake.deprovisionsAsyncMutex.Unlock()
	fake.DeprovisionsAsyncStub = nil
	fake.deprovisionsAsyncReturns = struct {
		result1 bool
	}{result1}
}

func (fake *FakeServiceProvider) DeprovisionsAsyncReturnsOnCall(i int, result1 bool) {
	fake.deprovisionsAsyncMutex.Lock()
	defer fake.deprovisionsAsyncMutex.Unlock()
	fake.DeprovisionsAsyncStub = nil
	if fake.deprovisionsAsyncReturnsOnCall == nil {
		fake.deprovisionsAsyncReturnsOnCall = make(map[int]struct {
			result1 bool
		})
	}
	fake.deprovisionsAsyncReturnsOnCall[i] = struct {
		result1 bool
	}{result1}
}

func (fake *FakeServiceProvider) PollInstance(arg1 context.Context, arg2 models.ServiceInstanceDetails) (bool, string, error) {
	fake.pollInstanceMutex.Lock()
	ret, specificReturn := fake.pollInstanceReturnsOnCall[len(fake.pollInstanceArgsForCall)]
	fake.pollInstanceArgsForCall = append(fake.pollInstanceArgsForCall, struct {
		arg1 context.Context
		arg2 models.ServiceInstanceDetails
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

func (fake *FakeServiceProvider) PollInstanceCalls(stub func(context.Context, models.ServiceInstanceDetails) (bool, string, error)) {
	fake.pollInstanceMutex.Lock()
	defer fake.pollInstanceMutex.Unlock()
	fake.PollInstanceStub = stub
}

func (fake *FakeServiceProvider) PollInstanceArgsForCall(i int) (context.Context, models.ServiceInstanceDetails) {
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

func (fake *FakeServiceProvider) Provision(arg1 context.Context, arg2 *varcontext.VarContext) (models.ServiceInstanceDetails, error) {
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

func (fake *FakeServiceProvider) ProvisionCalls(stub func(context.Context, *varcontext.VarContext) (models.ServiceInstanceDetails, error)) {
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

func (fake *FakeServiceProvider) ProvisionReturns(result1 models.ServiceInstanceDetails, result2 error) {
	fake.provisionMutex.Lock()
	defer fake.provisionMutex.Unlock()
	fake.ProvisionStub = nil
	fake.provisionReturns = struct {
		result1 models.ServiceInstanceDetails
		result2 error
	}{result1, result2}
}

func (fake *FakeServiceProvider) ProvisionReturnsOnCall(i int, result1 models.ServiceInstanceDetails, result2 error) {
	fake.provisionMutex.Lock()
	defer fake.provisionMutex.Unlock()
	fake.ProvisionStub = nil
	if fake.provisionReturnsOnCall == nil {
		fake.provisionReturnsOnCall = make(map[int]struct {
			result1 models.ServiceInstanceDetails
			result2 error
		})
	}
	fake.provisionReturnsOnCall[i] = struct {
		result1 models.ServiceInstanceDetails
		result2 error
	}{result1, result2}
}

func (fake *FakeServiceProvider) ProvisionsAsync() bool {
	fake.provisionsAsyncMutex.Lock()
	ret, specificReturn := fake.provisionsAsyncReturnsOnCall[len(fake.provisionsAsyncArgsForCall)]
	fake.provisionsAsyncArgsForCall = append(fake.provisionsAsyncArgsForCall, struct {
	}{})
	stub := fake.ProvisionsAsyncStub
	fakeReturns := fake.provisionsAsyncReturns
	fake.recordInvocation("ProvisionsAsync", []interface{}{})
	fake.provisionsAsyncMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeServiceProvider) ProvisionsAsyncCallCount() int {
	fake.provisionsAsyncMutex.RLock()
	defer fake.provisionsAsyncMutex.RUnlock()
	return len(fake.provisionsAsyncArgsForCall)
}

func (fake *FakeServiceProvider) ProvisionsAsyncCalls(stub func() bool) {
	fake.provisionsAsyncMutex.Lock()
	defer fake.provisionsAsyncMutex.Unlock()
	fake.ProvisionsAsyncStub = stub
}

func (fake *FakeServiceProvider) ProvisionsAsyncReturns(result1 bool) {
	fake.provisionsAsyncMutex.Lock()
	defer fake.provisionsAsyncMutex.Unlock()
	fake.ProvisionsAsyncStub = nil
	fake.provisionsAsyncReturns = struct {
		result1 bool
	}{result1}
}

func (fake *FakeServiceProvider) ProvisionsAsyncReturnsOnCall(i int, result1 bool) {
	fake.provisionsAsyncMutex.Lock()
	defer fake.provisionsAsyncMutex.Unlock()
	fake.ProvisionsAsyncStub = nil
	if fake.provisionsAsyncReturnsOnCall == nil {
		fake.provisionsAsyncReturnsOnCall = make(map[int]struct {
			result1 bool
		})
	}
	fake.provisionsAsyncReturnsOnCall[i] = struct {
		result1 bool
	}{result1}
}

func (fake *FakeServiceProvider) Unbind(arg1 context.Context, arg2 models.ServiceInstanceDetails, arg3 string, arg4 *varcontext.VarContext) error {
	fake.unbindMutex.Lock()
	ret, specificReturn := fake.unbindReturnsOnCall[len(fake.unbindArgsForCall)]
	fake.unbindArgsForCall = append(fake.unbindArgsForCall, struct {
		arg1 context.Context
		arg2 models.ServiceInstanceDetails
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

func (fake *FakeServiceProvider) UnbindCalls(stub func(context.Context, models.ServiceInstanceDetails, string, *varcontext.VarContext) error) {
	fake.unbindMutex.Lock()
	defer fake.unbindMutex.Unlock()
	fake.UnbindStub = stub
}

func (fake *FakeServiceProvider) UnbindArgsForCall(i int) (context.Context, models.ServiceInstanceDetails, string, *varcontext.VarContext) {
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

func (fake *FakeServiceProvider) UpdateInstanceDetails(arg1 context.Context, arg2 *models.ServiceInstanceDetails) error {
	fake.updateInstanceDetailsMutex.Lock()
	ret, specificReturn := fake.updateInstanceDetailsReturnsOnCall[len(fake.updateInstanceDetailsArgsForCall)]
	fake.updateInstanceDetailsArgsForCall = append(fake.updateInstanceDetailsArgsForCall, struct {
		arg1 context.Context
		arg2 *models.ServiceInstanceDetails
	}{arg1, arg2})
	stub := fake.UpdateInstanceDetailsStub
	fakeReturns := fake.updateInstanceDetailsReturns
	fake.recordInvocation("UpdateInstanceDetails", []interface{}{arg1, arg2})
	fake.updateInstanceDetailsMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeServiceProvider) UpdateInstanceDetailsCallCount() int {
	fake.updateInstanceDetailsMutex.RLock()
	defer fake.updateInstanceDetailsMutex.RUnlock()
	return len(fake.updateInstanceDetailsArgsForCall)
}

func (fake *FakeServiceProvider) UpdateInstanceDetailsCalls(stub func(context.Context, *models.ServiceInstanceDetails) error) {
	fake.updateInstanceDetailsMutex.Lock()
	defer fake.updateInstanceDetailsMutex.Unlock()
	fake.UpdateInstanceDetailsStub = stub
}

func (fake *FakeServiceProvider) UpdateInstanceDetailsArgsForCall(i int) (context.Context, *models.ServiceInstanceDetails) {
	fake.updateInstanceDetailsMutex.RLock()
	defer fake.updateInstanceDetailsMutex.RUnlock()
	argsForCall := fake.updateInstanceDetailsArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeServiceProvider) UpdateInstanceDetailsReturns(result1 error) {
	fake.updateInstanceDetailsMutex.Lock()
	defer fake.updateInstanceDetailsMutex.Unlock()
	fake.UpdateInstanceDetailsStub = nil
	fake.updateInstanceDetailsReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeServiceProvider) UpdateInstanceDetailsReturnsOnCall(i int, result1 error) {
	fake.updateInstanceDetailsMutex.Lock()
	defer fake.updateInstanceDetailsMutex.Unlock()
	fake.UpdateInstanceDetailsStub = nil
	if fake.updateInstanceDetailsReturnsOnCall == nil {
		fake.updateInstanceDetailsReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.updateInstanceDetailsReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeServiceProvider) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.bindMutex.RLock()
	defer fake.bindMutex.RUnlock()
	fake.buildInstanceCredentialsMutex.RLock()
	defer fake.buildInstanceCredentialsMutex.RUnlock()
	fake.deprovisionMutex.RLock()
	defer fake.deprovisionMutex.RUnlock()
	fake.deprovisionsAsyncMutex.RLock()
	defer fake.deprovisionsAsyncMutex.RUnlock()
	fake.pollInstanceMutex.RLock()
	defer fake.pollInstanceMutex.RUnlock()
	fake.provisionMutex.RLock()
	defer fake.provisionMutex.RUnlock()
	fake.provisionsAsyncMutex.RLock()
	defer fake.provisionsAsyncMutex.RUnlock()
	fake.unbindMutex.RLock()
	defer fake.unbindMutex.RUnlock()
	fake.updateMutex.RLock()
	defer fake.updateMutex.RUnlock()
	fake.updateInstanceDetailsMutex.RLock()
	defer fake.updateInstanceDetailsMutex.RUnlock()
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
