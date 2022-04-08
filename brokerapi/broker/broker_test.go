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

package broker_test

import (
	"context"
	"os"
	"reflect"
	"testing"

	"code.cloudfoundry.org/lager"
	. "github.com/cloudfoundry/cloud-service-broker/brokerapi/broker"
	"github.com/cloudfoundry/cloud-service-broker/db_service"
	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/internal/storage/storagefakes"
	"github.com/cloudfoundry/cloud-service-broker/pkg/broker"
	"github.com/cloudfoundry/cloud-service-broker/pkg/broker/brokerfakes"
	"github.com/cloudfoundry/cloud-service-broker/pkg/credstore"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/builtin/base"
	"github.com/cloudfoundry/cloud-service-broker/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/utils"
	"github.com/pborman/uuid"
	"github.com/pivotal-cf/brokerapi/v8/domain"
	"github.com/spf13/viper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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
					ID:   "3dcc45e8-0020-11ec-8ae1-f7abb5d2e742",
					Name: "fake-plan-name",
				},
				ServiceProperties: map[string]interface{}{
					"plan-defined-key":       "plan-defined-value",
					"other-plan-defined-key": "other-plan-defined-value",
				},
			},
		},
		ProvisionComputedVariables: []varcontext.DefaultVariable{
			{Name: "labels", Default: "${json.marshal(request.default_labels)}", Overwrite: true},
			{Name: "originatingIdentity", Default: "${json.marshal(request.x_broker_api_originating_identity)}", Overwrite: true},
		},
		BindComputedVariables: []varcontext.DefaultVariable{{Name: "originatingIdentity", Default: "${json.marshal(request.x_broker_api_originating_identity)}", Overwrite: true}},
		ProvisionInputVariables: []broker.BrokerVariable{
			{
				FieldName: "foo",
				Type:      "string",
				Details:   "fake field name",
			},
			{
				FieldName: "baz",
				Type:      "string",
				Details:   "other fake field name",
			},
			{
				FieldName: "guz",
				Type:      "string",
				Details:   "yet another fake field name",
			},
		},
		ImportInputVariables: []broker.ImportVariable{
			{
				Name:       "import_field_1",
				Type:       "string",
				Details:    "fake import field",
				TfResource: "fake.tf.resource",
			},
		},
		BindInputVariables: []broker.BrokerVariable{
			{
				FieldName: "valid_bind_parameter",
				Type:      "string",
				Details:   "fake valid field",
			},
		},
	}
	svc := defn.CatalogEntry()

	stub := serviceStub{
		ServiceId:         svc.ID,
		PlanId:            svc.Plans[0].ID,
		ServiceDefinition: defn,

		Provider: &brokerfakes.FakeServiceProvider{
			ProvisionsAsyncStub:   func() bool { return isAsync },
			DeprovisionsAsyncStub: func() bool { return isAsync },
			GetImportedPropertiesStub: func(ctx context.Context, planGUID string, tfID string, inputVariables []broker.BrokerVariable) (map[string]interface{}, error) {
				return map[string]interface{}{}, nil
			},
			ProvisionStub: func(ctx context.Context, vc *varcontext.VarContext) (storage.ServiceInstanceDetails, error) {
				return storage.ServiceInstanceDetails{
					Outputs: map[string]interface{}{"mynameis": "instancename", "foo": "baz"},
				}, nil
			},
			BindStub: func(ctx context.Context, vc *varcontext.VarContext) (map[string]interface{}, error) {
				return map[string]interface{}{"foo": "bar"}, nil
			},
			BuildInstanceCredentialsStub: func(ctx context.Context, creds map[string]interface{}, outs storage.JSONObject) (*domain.Binding, error) {
				mixin := base.MergedInstanceCredsMixin{}
				return mixin.BuildInstanceCredentials(ctx, creds, outs)
			},
		},
	}

	stub.ServiceDefinition.ProviderBuilder = func(logger lager.Logger, store broker.ServiceProviderStorage) broker.ServiceProvider {
		return stub.Provider
	}

	return &stub
}

// newStubbedBroker creates a new ServiceBroker with a dummy database for the given registry.
// It returns the broker and a callback used to clean up the database when done with it.
func newStubbedBroker(t *testing.T, registry broker.BrokerRegistry, cs credstore.CredStore, encryptor *storagefakes.FakeEncryptor) (broker *ServiceBroker, closer func()) {
	// Set up database
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		t.Fatalf("couldn't create database: %v", err)
	}
	db_service.RunMigrations(db)

	closer = func() {
		os.Remove("test.db")
	}

	config := &BrokerConfig{
		Registry:  registry,
		Credstore: cs,
	}

	broker, err = New(config, utils.NewLogger("brokers-test"), storage.New(db, encryptor))
	if err != nil {
		t.Fatalf("couldn't create broker: %v", err)
	}

	return broker, closer
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
	Check     func(t *testing.T, broker *ServiceBroker, stub *serviceStub, encryptor *storagefakes.FakeEncryptor)
	Credstore credstore.CredStore
}

// BrokerEndpointTestSuite holds a set of tests for a single endpoint.
type BrokerEndpointTestSuite map[string]BrokerEndpointTestCase

// Run executes every test case, setting up a new environment for each and
// tearing it down afterward.
func (cases BrokerEndpointTestSuite) Run(t *testing.T) {
	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			viper.Reset()
			stub := fakeService(t, tc.AsyncService)

			t.Log("Creating broker")
			registry := broker.BrokerRegistry{}
			registry.Register(stub.ServiceDefinition)

			encryptor := &storagefakes.FakeEncryptor{
				DecryptStub: func(bytes []byte) ([]byte, error) { return bytes, nil },
				EncryptStub: func(bytes []byte) ([]byte, error) { return bytes, nil },
			}
			broker, closer := newStubbedBroker(t, registry, tc.Credstore, encryptor)
			defer closer()

			initService(t, tc.ServiceState, broker, stub)

			t.Log("Running check")
			tc.Check(t, broker, stub, encryptor)
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
