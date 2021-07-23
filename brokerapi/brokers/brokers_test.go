// Copyright 2018 the Service Broker Project Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package brokers_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"

	"code.cloudfoundry.org/lager"
	. "github.com/cloudfoundry-incubator/cloud-service-broker/brokerapi/brokers"
	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service"
	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/broker"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/broker/brokerfakes"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/credstore"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/credstore/credstorefakes"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/providers/builtin/base"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/varcontext"
	"github.com/cloudfoundry-incubator/cloud-service-broker/utils"
	"github.com/jinzhu/gorm"
	"github.com/pborman/uuid"
	"github.com/pivotal-cf/brokerapi/v8"
	"github.com/pivotal-cf/brokerapi/v8/domain"
	"github.com/pivotal-cf/brokerapi/v8/domain/apiresponses"
	"github.com/pivotal-cf/brokerapi/v8/middlewares"
)

// InstanceState holds the lifecycle state of a provisioned service instance.
// It goes None -> Provisioned -> Bound -> Unbound -> Deprovisioned
type InstanceState int

const (
	StateNone InstanceState = iota
	StateProvisioned
	StateBound
	StateUnbound
	StateDeprovisioned
)

const (
	fakeInstanceId = "newid"
	fakeBindingId  = "newbinding"
)

// serviceStub holds a stubbed out ServiceDefinition with easy access to
// its ID, a valid plan ID, and the mock provider.
type serviceStub struct {
	ServiceId         string
	PlanId            string
	Provider          *brokerfakes.FakeServiceProvider
	ServiceDefinition *broker.ServiceDefinition
}

// ProvisionDetails creates a domain.ProvisionDetails object valid for
// the given service.
func (s *serviceStub) ProvisionDetails() domain.ProvisionDetails {
	return domain.ProvisionDetails{
		ServiceID: s.ServiceId,
		PlanID:    s.PlanId,
	}
}

// DeprovisionDetails creates a brokerapi.DeprovisionDetails object valid for
// the given service.
func (s *serviceStub) DeprovisionDetails() domain.DeprovisionDetails {
	return domain.DeprovisionDetails{
		ServiceID: s.ServiceId,
		PlanID:    s.PlanId,
	}
}

// BindDetails creates a domain.BindDetails object valid for
// the given service.
func (s *serviceStub) BindDetails() domain.BindDetails {
	return domain.BindDetails{
		ServiceID: s.ServiceId,
		PlanID:    s.PlanId,
	}
}

// UnbindDetails creates a domain.UnbindDetails object valid for
// the given service.
func (s *serviceStub) UnbindDetails() domain.UnbindDetails {
	return domain.UnbindDetails{
		ServiceID: s.ServiceId,
		PlanID:    s.PlanId,
	}
}

// UpdateDetails creates a domain.UpdateDetails object valid for
// the given service.
func (s *serviceStub) UpdateDetails() domain.UpdateDetails {
	return domain.UpdateDetails{
		ServiceID: s.ServiceId,
		PlanID:    s.PlanId,
	}
}

// fakeService creates a ServiceDefinition with a mock ServiceProvider and
// references to some important properties.
func fakeService(t *testing.T, isAsync bool) *serviceStub {
	defn := &broker.ServiceDefinition{
		Id:   uuid.New(),
		Name: "fake-service-name",
		Plans: []broker.ServicePlan{
			{
				ServicePlan: domain.ServicePlan{
					ID:   "fake-plan-id",
					Name: "fake-plan-name",
				},
			},
		},
		ProvisionComputedVariables: []varcontext.DefaultVariable{
			{Name: "labels", Default: "${json.marshal(request.default_labels)}", Overwrite: true},
			{Name: "originatingIdentity", Default: "${json.marshal(request.x_broker_api_originating_identity)}", Overwrite: true},
		},
		BindComputedVariables: []varcontext.DefaultVariable{{Name: "originatingIdentity", Default: "${json.marshal(request.x_broker_api_originating_identity)}", Overwrite: true}},
	}
	svc, err := defn.CatalogEntry()
	if err != nil {
		t.Fatal(err)
	}

	stub := serviceStub{
		ServiceId:         svc.ID,
		PlanId:            svc.Plans[0].ID,
		ServiceDefinition: defn,

		Provider: &brokerfakes.FakeServiceProvider{
			ProvisionsAsyncStub:   func() bool { return isAsync },
			DeprovisionsAsyncStub: func() bool { return isAsync },
			ProvisionStub: func(ctx context.Context, vc *varcontext.VarContext) (models.ServiceInstanceDetails, error) {
				instance := models.ServiceInstanceDetails{}
				instance.SetOtherDetails(map[string]interface{}{"mynameis": "instancename", "foo": "baz"})
				return instance, nil
			},
			BindStub: func(ctx context.Context, vc *varcontext.VarContext) (map[string]interface{}, error) {
				return map[string]interface{}{"foo": "bar"}, nil
			},
			BuildInstanceCredentialsStub: func(ctx context.Context, bc models.ServiceBindingCredentials, id models.ServiceInstanceDetails) (*domain.Binding, error) {
				mixin := base.MergedInstanceCredsMixin{}
				return mixin.BuildInstanceCredentials(ctx, bc, id)
			},
		},
	}

	stub.ServiceDefinition.ProviderBuilder = func(logger lager.Logger) broker.ServiceProvider {
		return stub.Provider
	}

	return &stub
}

// newStubbedBroker creates a new ServiceBroker with a dummy database for the given registry.
// It returns the broker and a callback used to clean up the database when done with it.
func newStubbedBroker(t *testing.T, registry broker.BrokerRegistry, cs credstore.CredStore) (broker *ServiceBroker, closer func()) {
	// Set up database
	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		t.Fatalf("couldn't create database: %v", err)
	}
	db_service.RunMigrations(db)
	db_service.DbConnection = db

	closer = func() {
		db.Close()
		os.Remove("test.db")
	}

	config := &BrokerConfig{
		Registry:  registry,
		Credstore: cs,
	}

	broker, err = New(config, utils.NewLogger("brokers-test"))
	if err != nil {
		t.Fatalf("couldn't create broker: %v", err)
	}

	return
}

// failIfErr is a test helper function which stops the test immediately if the
// error is set.
func failIfErr(t *testing.T, action string, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("Expected no error while %s, got: %v", action, err)
	}
}

// assertEqual does a reflect.DeepEqual on the values and if they're different
// reports the message and the values.
func assertEqual(t *testing.T, message string, expected, actual interface{}) {
	t.Helper()

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Error: %s Expected: %#v Actual: %#v", message, expected, actual)
	}
}

// assertEqual does a reflect.DeepEqual on the values and if they're different
// reports the message and the values.
func assertTrue(t *testing.T, message string, val bool) {
	t.Helper()

	if !val {
		t.Errorf("Error: %s was not true", message)
	}
}

// BrokerEndpointTestCase is the base test used for testing any
// brokerapi.ServiceBroker endpoint.
type BrokerEndpointTestCase struct {
	// The following properties are used to set up the environment for your test
	// to run in.
	AsyncService bool
	ServiceState InstanceState

	// Check is used to validate the state of the world and is where you should
	// put your test cases.
	Check     func(t *testing.T, broker *ServiceBroker, stub *serviceStub)
	Credstore credstore.CredStore
}

// BrokerEndpointTestSuite holds a set of tests for a single endpoint.
type BrokerEndpointTestSuite map[string]BrokerEndpointTestCase

// Run executes every test case, setting up a new environment for each and
// tearing it down afterward.
func (cases BrokerEndpointTestSuite) Run(t *testing.T) {
	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			stub := fakeService(t, tc.AsyncService)

			t.Log("Creating broker")
			registry := broker.BrokerRegistry{}
			registry.Register(stub.ServiceDefinition)

			broker, closer := newStubbedBroker(t, registry, tc.Credstore)
			defer closer()

			initService(t, tc.ServiceState, broker, stub)

			t.Log("Running check")
			tc.Check(t, broker, stub)
		})
	}
}

// initService creates a new service and brings it up to the lifecycle state given
// by state.
func initService(t *testing.T, state InstanceState, broker *ServiceBroker, stub *serviceStub) {
	if state >= StateProvisioned {
		_, err := broker.Provision(context.Background(), fakeInstanceId, stub.ProvisionDetails(), true)
		failIfErr(t, "provisioning", err)
	}

	if state >= StateBound {
		_, err := broker.Bind(context.Background(), fakeInstanceId, fakeBindingId, stub.BindDetails(), true)
		failIfErr(t, "binding", err)
	}

	if state >= StateUnbound {
		_, err := broker.Unbind(context.Background(), fakeInstanceId, fakeBindingId, stub.UnbindDetails(), true)
		failIfErr(t, "unbinding", err)
	}

	if state >= StateDeprovisioned {
		_, err := broker.Deprovision(context.Background(), fakeInstanceId, stub.DeprovisionDetails(), true)
		failIfErr(t, "deprovisioning", err)
	}
}

func TestServiceBroker_Provision(t *testing.T) {
	cases := BrokerEndpointTestSuite{
		"good-request": {
			ServiceState: StateProvisioned,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {

				assertEqual(t, "provision calls should match", 1, stub.Provider.ProvisionCallCount())
			},
		},
		"originating-header": {
			AsyncService: true,
			ServiceState: StateNone,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				header := "cloudfoundry eyANCiAgInVzZXJfaWQiOiAiNjgzZWE3NDgtMzA5Mi00ZmY0LWI2NTYtMzljYWNjNGQ1MzYwIg0KfQ=="
				newContext := context.WithValue(context.Background(), middlewares.OriginatingIdentityKey, header)
				broker.Provision(newContext, fakeInstanceId, stub.ProvisionDetails(), true)
				assertEqual(t, "provision calls should match", 1, stub.Provider.ProvisionCallCount())
				_, actualVarContext := stub.Provider.ProvisionArgsForCall(0)
				expectedOriginatingIdentityMap := `{"platform":"cloudfoundry","value":{"user_id":"683ea748-3092-4ff4-b656-39cacc4d5360"}}`

				assertEqual(t, "originatingIdentity should match", expectedOriginatingIdentityMap, actualVarContext.GetString("originatingIdentity"))
			},
		},
		"duplicate-request": {
			ServiceState: StateProvisioned,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				_, err := broker.Provision(context.Background(), fakeInstanceId, stub.ProvisionDetails(), true)
				assertEqual(t, "errors should match", apiresponses.ErrInstanceAlreadyExists, err)
			},
		},
		"requires-async": {
			AsyncService: true,
			ServiceState: StateNone,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				// false for async support
				_, err := broker.Provision(context.Background(), fakeInstanceId, stub.ProvisionDetails(), false)
				assertEqual(t, "errors should match", apiresponses.ErrAsyncRequired, err)
			},
		},
		"unknown-service-id": {
			ServiceState: StateNone,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				req := stub.ProvisionDetails()
				req.ServiceID = "bad-service-id"
				_, err := broker.Provision(context.Background(), fakeInstanceId, req, true)
				assertEqual(t, "errors should match", errors.New(`unknown service ID: "bad-service-id"`), err)
			},
		},
		"unknown-plan-id": {
			ServiceState: StateNone,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				req := stub.ProvisionDetails()
				req.PlanID = "bad-plan-id"
				_, err := broker.Provision(context.Background(), fakeInstanceId, req, true)
				assertEqual(t, "errors should match", errors.New(`plan ID "bad-plan-id" could not be found`), err)
			},
		},
		"bad-request-json": {
			ServiceState: StateNone,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				req := stub.ProvisionDetails()
				req.RawParameters = json.RawMessage("{invalid json")
				_, err := broker.Provision(context.Background(), fakeInstanceId, req, true)
				assertEqual(t, "errors should match", ErrInvalidUserInput, err)
			},
		},
	}

	cases.Run(t)
}

func TestServiceBroker_Deprovision(t *testing.T) {
	cases := BrokerEndpointTestSuite{
		"good-request": {
			ServiceState: StateProvisioned,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				_, err := broker.Deprovision(context.Background(), fakeInstanceId, stub.DeprovisionDetails(), true)
				failIfErr(t, "deprovisioning", err)

				assertEqual(t, "deprovision calls should match", 1, stub.Provider.DeprovisionCallCount())
			},
		},
		"originating-header": {
			AsyncService: true,
			ServiceState: StateProvisioned,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				header := "cloudfoundry eyANCiAgInVzZXJfaWQiOiAiNjgzZWE3NDgtMzA5Mi00ZmY0LWI2NTYtMzljYWNjNGQ1MzYwIg0KfQ=="
				newContext := context.WithValue(context.Background(), middlewares.OriginatingIdentityKey, header)
				_, err := broker.Deprovision(newContext, fakeInstanceId, stub.DeprovisionDetails(), true)
				failIfErr(t, "deprovisioning", err)

				assertEqual(t, "deprovision calls should match", 1, stub.Provider.DeprovisionCallCount())
				_, _, _, actualVarContext := stub.Provider.DeprovisionArgsForCall(0)
				expectedOriginatingIdentityMap := `{"platform":"cloudfoundry","value":{"user_id":"683ea748-3092-4ff4-b656-39cacc4d5360"}}`

				assertEqual(t, "originatingIdentity should match", expectedOriginatingIdentityMap, actualVarContext.GetString("originatingIdentity"))
			},
		},
		"duplicate-deprovision": {
			ServiceState: StateDeprovisioned,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				_, err := broker.Deprovision(context.Background(), fakeInstanceId, stub.DeprovisionDetails(), true)
				assertEqual(t, "duplicate deprovision should lead to DNE", brokerapi.ErrInstanceDoesNotExist, err)
			},
		},
		"instance-does-not-exist": {
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				_, err := broker.Deprovision(context.Background(), fakeInstanceId, stub.DeprovisionDetails(), true)
				assertEqual(t, "instance does not exist should be set", brokerapi.ErrInstanceDoesNotExist, err)
			},
		},
		"async-required": {
			AsyncService: true,
			ServiceState: StateProvisioned,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				_, err := broker.Deprovision(context.Background(), fakeInstanceId, stub.DeprovisionDetails(), false)
				assertEqual(t, "async required should be returned if not supported", brokerapi.ErrAsyncRequired, err)
			},
		},
		"async-deprovision-returns-operation": {
			AsyncService: true,
			ServiceState: StateProvisioned,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				operationId := "my-operation-id"
				stub.Provider.DeprovisionReturns(&operationId, nil)
				resp, err := broker.Deprovision(context.Background(), fakeInstanceId, stub.DeprovisionDetails(), true)
				failIfErr(t, "deprovisioning", err)

				assertEqual(t, "operationid should be set as the data", operationId, resp.OperationData)
				assertEqual(t, "IsAsync should be set", true, resp.IsAsync)
			},
		},

		"async-deprovision-updates-db": {
			AsyncService: true,
			ServiceState: StateProvisioned,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				operationId := "my-operation-id"
				stub.Provider.DeprovisionReturns(&operationId, nil)
				_, err := broker.Deprovision(context.Background(), fakeInstanceId, stub.DeprovisionDetails(), true)
				failIfErr(t, "deprovisioning", err)

				details, err := db_service.GetServiceInstanceDetailsById(context.Background(), fakeInstanceId)
				failIfErr(t, "looking up details", err)

				assertEqual(t, "OperationId should be set as the data", operationId, details.OperationId)
				assertEqual(t, "OperationType should be set as Deprovision", models.DeprovisionOperationType, details.OperationType)
			},
		},
	}

	cases.Run(t)
}

func TestServiceBroker_Bind(t *testing.T) {
	cases := BrokerEndpointTestSuite{
		"good-request": {
			ServiceState: StateBound,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				assertEqual(t, "BindCallCount should match", 1, stub.Provider.BindCallCount())
				assertEqual(t, "BuildInstanceCredentialsCallCount should match", 1, stub.Provider.BuildInstanceCredentialsCallCount())
			},
		},
		"good-request-with-credstore": {
			ServiceState: StateBound,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				assertEqual(t, "BindCallCount should match", 1, stub.Provider.BindCallCount())
				assertEqual(t, "BuildInstanceCredentialsCallCount should match", 1, stub.Provider.BuildInstanceCredentialsCallCount())
				fcs := broker.Credstore.(*credstorefakes.FakeCredStore)
				assertEqual(t, "Credstore Put call count should match", 1, fcs.PutCallCount())
				assertEqual(t, "Credstore AddPermission call count should match", 1, fcs.AddPermissionCallCount())
			},
			Credstore: &credstorefakes.FakeCredStore{},
		},
		"originating-header": {
			ServiceState: StateProvisioned,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				header := "cloudfoundry eyANCiAgInVzZXJfaWQiOiAiNjgzZWE3NDgtMzA5Mi00ZmY0LWI2NTYtMzljYWNjNGQ1MzYwIg0KfQ=="
				newContext := context.WithValue(context.Background(), middlewares.OriginatingIdentityKey, header)
				broker.Bind(newContext, fakeInstanceId, fakeBindingId, stub.BindDetails(), true)
				assertEqual(t, "bind calls should match", 1, stub.Provider.BindCallCount())
				_, actualVarContext := stub.Provider.BindArgsForCall(0)
				expectedOriginatingIdentityMap := `{"platform":"cloudfoundry","value":{"user_id":"683ea748-3092-4ff4-b656-39cacc4d5360"}}`

				assertEqual(t, "originatingIdentity should match", expectedOriginatingIdentityMap, actualVarContext.GetString("originatingIdentity"))
			},
		},
		"duplicate-request": {
			ServiceState: StateBound,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				_, err := broker.Bind(context.Background(), fakeInstanceId, fakeBindingId, stub.BindDetails(), true)
				assertEqual(t, "errors should match", brokerapi.ErrBindingAlreadyExists, err)
			},
		},
		"bad-bind-call": {
			ServiceState: StateProvisioned,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				req := stub.BindDetails()
				stub.Provider.BindStub = func(ctx context.Context, vc *varcontext.VarContext) (map[string]interface{}, error) {
					return nil, errors.New("fake error")
				}

				_, err := broker.Bind(context.Background(), fakeInstanceId, "bad-bind-call", req, true)
				assertEqual(t, "errors should match", "error performing bind: fake error", err.Error())
			},
		},
		"bad-request-json": {
			ServiceState: StateProvisioned,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				req := stub.BindDetails()
				req.RawParameters = json.RawMessage("{invalid json")
				_, err := broker.Bind(context.Background(), fakeInstanceId, fakeBindingId, req, true)
				assertEqual(t, "errors should match", ErrInvalidUserInput, err)
			},
		},
		"bind-variables-override-instance-variables": {
			ServiceState: StateProvisioned,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				req := stub.BindDetails()
				bindResult, err := broker.Bind(context.Background(), fakeInstanceId, "override-params", req, true)
				failIfErr(t, "binding", err)
				credMap, ok := bindResult.Credentials.(map[string]interface{})
				assertTrue(t, "bind result credentials should be a map", ok)
				assertEqual(t, "credential overridden", "bar", credMap["foo"].(string))
			},
		},
		"bind-returns-credhub-ref": {
			ServiceState: StateProvisioned,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				req := stub.BindDetails()
				bindResult, err := broker.Bind(context.Background(), fakeInstanceId, "override-params", req, true)
				failIfErr(t, "binding", err)
				credMap, ok := bindResult.Credentials.(map[string]interface{})
				assertTrue(t, "bind result credentials should be a map", ok)
				assertTrue(t, "value foo missing", credMap["foo"] == nil)
				assertTrue(t, "cred-hub ref exists", credMap["credhub-ref"] != nil)
				assertEqual(t, "cred-hub ref has correct value", "/c/csb/fake-service-name/override-params/secrets-and-services", credMap["credhub-ref"].(string))
			},
			Credstore: &credstorefakes.FakeCredStore{},
		},
	}

	cases.Run(t)
}

func TestServiceBroker_Unbind(t *testing.T) {
	cases := BrokerEndpointTestSuite{
		"good-request": {
			ServiceState: StateBound,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				_, err := broker.Unbind(context.Background(), fakeInstanceId, fakeBindingId, stub.UnbindDetails(), true)
				failIfErr(t, "unbinding", err)
			},
		},
		"good-request-with-credhub": {
			ServiceState: StateBound,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				_, err := broker.Unbind(context.Background(), fakeInstanceId, fakeBindingId, stub.UnbindDetails(), true)
				failIfErr(t, "unbinding", err)
				fcs := broker.Credstore.(*credstorefakes.FakeCredStore)
				assertEqual(t, "Credstore DeletePermission call count should match", 1, fcs.DeletePermissionCallCount())
				assertEqual(t, "Credstore Delete call count should match", 1, fcs.DeleteCallCount())
			},
			Credstore: &credstorefakes.FakeCredStore{},
		},
		"originating-header": {
			ServiceState: StateBound,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				header := "cloudfoundry eyANCiAgInVzZXJfaWQiOiAiNjgzZWE3NDgtMzA5Mi00ZmY0LWI2NTYtMzljYWNjNGQ1MzYwIg0KfQ=="
				newContext := context.WithValue(context.Background(), middlewares.OriginatingIdentityKey, header)
				broker.Unbind(newContext, fakeInstanceId, fakeBindingId, stub.UnbindDetails(), true)
				assertEqual(t, "unbind calls should match", 1, stub.Provider.UnbindCallCount())
				_, _, _, actualVarContext := stub.Provider.UnbindArgsForCall(0)
				expectedOriginatingIdentityMap := `{"platform":"cloudfoundry","value":{"user_id":"683ea748-3092-4ff4-b656-39cacc4d5360"}}`

				assertEqual(t, "originatingIdentity should match", expectedOriginatingIdentityMap, actualVarContext.GetString("originatingIdentity"))
			},
		},
		"multiple-unbinds": {
			ServiceState: StateUnbound,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				_, err := broker.Unbind(context.Background(), fakeInstanceId, fakeBindingId, stub.UnbindDetails(), true)
				assertEqual(t, "errors should match", brokerapi.ErrBindingDoesNotExist, err)
			},
		},
	}

	cases.Run(t)
}

func TestServiceBroker_LastOperation(t *testing.T) {
	cases := BrokerEndpointTestSuite{
		"missing-instance": {
			ServiceState: StateProvisioned,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				_, err := broker.LastOperation(context.Background(), "invalid-instance-id", domain.PollDetails{OperationData: "operationtoken"})
				assertEqual(t, "errors should match", apiresponses.ErrInstanceDoesNotExist, err)
			},
		},
		"called-on-synchronous-service": {
			ServiceState: StateProvisioned,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				_, err := broker.LastOperation(context.Background(), fakeInstanceId, domain.PollDetails{OperationData: "operationtoken"})
				assertEqual(t, "errors should match", apiresponses.ErrAsyncRequired, err)
			},
		},
		"called-on-async-service": {
			AsyncService: true,
			ServiceState: StateProvisioned,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				_, err := broker.LastOperation(context.Background(), fakeInstanceId, domain.PollDetails{OperationData: "operationtoken"})
				failIfErr(t, "shouldn't be called on async service", err)

				assertEqual(t, "PollInstanceCallCount should match", 1, stub.Provider.PollInstanceCallCount())
			},
		},
		"poll-returns-failure": {
			AsyncService: true,
			ServiceState: StateProvisioned,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				stub.Provider.PollInstanceReturns(false, "", errors.New("not-retryable"))
				status, err := broker.LastOperation(context.Background(), fakeInstanceId, domain.PollDetails{OperationData: "operationtoken"})
				failIfErr(t, "checking last operation", err)
				assertEqual(t, "non-retryable errors should result in a failure state", domain.Failed, status.State)
				assertEqual(t, "description should be error string", "not-retryable", status.Description)
			},
		},
		"poll-returns-not-done": {
			AsyncService: true,
			ServiceState: StateProvisioned,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				stub.Provider.PollInstanceReturns(false, "", nil)
				status, err := broker.LastOperation(context.Background(), fakeInstanceId, domain.PollDetails{OperationData: "operationtoken"})
				failIfErr(t, "checking last operation", err)
				assertEqual(t, "polls that return no error should result in an in-progress state", domain.InProgress, status.State)
			},
		},
		"poll-returns-success": {
			AsyncService: true,
			ServiceState: StateProvisioned,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				stub.Provider.PollInstanceReturns(true, "message", nil)
				status, err := broker.LastOperation(context.Background(), fakeInstanceId, domain.PollDetails{OperationData: "operationtoken"})
				failIfErr(t, "checking last operation", err)
				assertEqual(t, "polls that return finished should result in a succeeded state", domain.Succeeded, status.State)
				assertEqual(t, "polls that return finished should have status message", "message", status.Description)
			},
		},
	}

	cases.Run(t)
}

func TestServiceBroker_GetBinding(t *testing.T) {
	cases := BrokerEndpointTestSuite{
		"called-on-bound": {
			ServiceState: StateBound,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				_, err := broker.GetBinding(context.Background(), fakeInstanceId, fakeBindingId, domain.FetchBindingDetails{})

				assertEqual(t, "expect get binding not supported err", ErrGetBindingsUnsupported, err)
			},
		},
	}

	cases.Run(t)
}

func TestServiceBroker_GetInstance(t *testing.T) {
	cases := BrokerEndpointTestSuite{
		"called-while-provisioned": {
			ServiceState: StateProvisioned,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				_, err := broker.GetInstance(context.Background(), fakeInstanceId, domain.FetchInstanceDetails{})

				assertEqual(t, "expect get instances not supported err", ErrGetInstancesUnsupported, err)
			},
		},
	}

	cases.Run(t)
}

func TestServiceBroker_LastBindingOperation(t *testing.T) {
	cases := BrokerEndpointTestSuite{
		"called-while-bound": {
			ServiceState: StateProvisioned,
			AsyncService: true,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				_, err := broker.LastBindingOperation(context.Background(), fakeInstanceId, fakeBindingId, domain.PollDetails{})

				assertEqual(t, "expect last binding to return async required", apiresponses.ErrAsyncRequired, err)
			},
		},
	}

	cases.Run(t)
}

func TestServiceBroker_Update(t *testing.T) {
	cases := BrokerEndpointTestSuite{
		"good-request": {
			ServiceState: StateProvisioned,
			AsyncService: true,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				_, err := broker.Update(context.Background(), fakeInstanceId, stub.UpdateDetails(), true)

				failIfErr(t, "update", err)
			},
		},
		"originating-header": {
			ServiceState: StateProvisioned,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				header := "cloudfoundry eyANCiAgInVzZXJfaWQiOiAiNjgzZWE3NDgtMzA5Mi00ZmY0LWI2NTYtMzljYWNjNGQ1MzYwIg0KfQ=="
				newContext := context.WithValue(context.Background(), middlewares.OriginatingIdentityKey, header)
				broker.Update(newContext, fakeInstanceId, stub.UpdateDetails(), true)
				assertEqual(t, "update calls should match", 1, stub.Provider.UpdateCallCount())
				_, actualVarContext := stub.Provider.UpdateArgsForCall(0)
				expectedOriginatingIdentityMap := `{"platform":"cloudfoundry","value":{"user_id":"683ea748-3092-4ff4-b656-39cacc4d5360"}}`

				assertEqual(t, "originatingIdentity should match", expectedOriginatingIdentityMap, actualVarContext.GetString("originatingIdentity"))
			},
		},
		"missing-instance": {
			ServiceState: StateProvisioned,
			AsyncService: true,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				_, err := broker.Update(context.Background(), "bogus-id", stub.UpdateDetails(), true)

				assertEqual(t, "expect update error to be instance does not exist", brokerapi.ErrInstanceDoesNotExist, err)
			},
		},
		"requires-async": {
			AsyncService: true,
			ServiceState: StateProvisioned,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				// false for async support
				_, err := broker.Update(context.Background(), fakeInstanceId, stub.UpdateDetails(), false)
				assertEqual(t, "errors should match", brokerapi.ErrAsyncRequired, err)
			},
		},
		"instance-does-not-exist": {
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				_, err := broker.Update(context.Background(), fakeInstanceId, stub.UpdateDetails(), true)
				assertEqual(t, "instance does not exist should be set", brokerapi.ErrInstanceDoesNotExist, err)
			},
		},
		"bad-request-json": {
			AsyncService: true,
			ServiceState: StateProvisioned,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				req := stub.UpdateDetails()
				req.RawParameters = json.RawMessage("{invalid json")
				_, err := broker.Update(context.Background(), fakeInstanceId, req, true)
				assertEqual(t, "errors should match", ErrInvalidUserInput, err)
			},
		},
		"attempt-to-update-non-updatable-parameter": {
			AsyncService: true,
			ServiceState: StateProvisioned,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				req := stub.UpdateDetails()
				stub.Provider.UpdateStub = func(ctx context.Context, varContext *varcontext.VarContext) (models.ServiceInstanceDetails, error) {
					return models.ServiceInstanceDetails{}, ErrNonUpdatableParameter
				}

				_, err := broker.Update(context.Background(), fakeInstanceId, req, true)
				assertEqual(t, "errors should match", ErrNonUpdatableParameter, err)
			},
		},
		"good-request-valid-parameter": {
			ServiceState: StateProvisioned,
			AsyncService: true,
			Check: func(t *testing.T, broker *ServiceBroker, stub *serviceStub) {
				req := stub.UpdateDetails()
				req.RawParameters = json.RawMessage(`{"force_delete":"false"}`)
				_, err := broker.Update(context.Background(), fakeInstanceId, req, true)

				failIfErr(t, "update", err)
			},
		},
	}

	cases.Run(t)
}
