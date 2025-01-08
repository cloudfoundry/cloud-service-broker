package broker_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"code.cloudfoundry.org/lager/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v12/domain"
	"github.com/pivotal-cf/brokerapi/v12/domain/apiresponses"
	"github.com/pivotal-cf/brokerapi/v12/middlewares"

	"github.com/cloudfoundry/cloud-service-broker/v2/brokerapi/broker"
	"github.com/cloudfoundry/cloud-service-broker/v2/brokerapi/broker/brokerfakes"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/storage"
	pkgBroker "github.com/cloudfoundry/cloud-service-broker/v2/pkg/broker"
	pkgBrokerFakes "github.com/cloudfoundry/cloud-service-broker/v2/pkg/broker/brokerfakes"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/credstore/credstorefakes"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/v2/utils"
)

var _ = Describe("Bind", func() {
	const (
		appGUID    = "test-app-guid"
		orgID      = "test-org-id"
		spaceID    = "test-space-id"
		planID     = "test-plan-id"
		serviceID  = "test-service-id"
		offeringID = "test-service-id"
		instanceID = "test-instance-id"
		bindingID  = "test-binding-id"
	)

	var (
		serviceBroker *broker.ServiceBroker
		bindDetails   domain.BindDetails

		fakeStorage         *brokerfakes.FakeStorage
		fakeServiceProvider *pkgBrokerFakes.FakeServiceProvider
		fakeCredStore       *credstorefakes.FakeCredStore

		brokerConfig *broker.BrokerConfig
	)

	BeforeEach(func() {
		fakeServiceProvider = &pkgBrokerFakes.FakeServiceProvider{}
		fakeServiceProvider.BindReturns(map[string]any{
			"fakeOutput": "fakeValue",
		}, nil)

		fakeStorage = &brokerfakes.FakeStorage{}
		fakeStorage.ExistsServiceBindingCredentialsReturns(false, nil)
		fakeStorage.GetServiceInstanceDetailsReturns(storage.ServiceInstanceDetails{
			GUID:             instanceID,
			ServiceGUID:      offeringID,
			PlanGUID:         planID,
			SpaceGUID:        spaceID,
			OrganizationGUID: orgID,
			Outputs:          map[string]any{"fakeInstanceOutput": "fakeInstanceValue"},
		}, nil)

		fakeCredStore = &credstorefakes.FakeCredStore{}

		providerBuilder := func(logger lager.Logger, store pkgBroker.ServiceProviderStorage) pkgBroker.ServiceProvider {
			return fakeServiceProvider
		}

		planUpdatable := true

		brokerConfig = &broker.BrokerConfig{
			Registry: pkgBroker.BrokerRegistry{
				"test-service": &pkgBroker.ServiceDefinition{
					ID:   offeringID,
					Name: "test-service",
					Plans: []pkgBroker.ServicePlan{
						{
							ServicePlan: domain.ServicePlan{
								ID:            planID,
								Name:          "test-plan",
								PlanUpdatable: &planUpdatable,
							},
							ServiceProperties: map[string]any{
								"plan-defined-key":       "plan-defined-value",
								"other-plan-defined-key": "other-plan-defined-value",
							},
						},
					},
					BindInputVariables: []pkgBroker.BrokerVariable{
						{
							FieldName: "bind_field_1",
							Type:      "string",
							Details:   "fake bind field",
						},
					},
					BindComputedVariables: []varcontext.DefaultVariable{
						{Name: "copyOriginatingIdentity", Default: "${json.marshal(request.x_broker_api_originating_identity)}", Overwrite: true},
					},
					ProviderBuilder: providerBuilder,
				},
			},
			Credstore: fakeCredStore,
		}

		serviceBroker = must(broker.New(brokerConfig, fakeStorage, utils.NewLogger("bind-test-with-credstore")))

		bindDetails = domain.BindDetails{
			AppGUID:       appGUID,
			PlanID:        planID,
			ServiceID:     serviceID,
			RawParameters: json.RawMessage(`{"bind_field_1":"bind_value_1"}`),
		}
	})

	Describe("successful bind", func() {
		It("should create a binding in the database", func() {
			const expectedHeader = "cloudfoundry eyANCiAgInVzZXJfaWQiOiAiNjgzZWE3NDgtMzA5Mi00ZmY0LWI2NTYtMzljYWNjNGQ1MzYwIg0KfQ=="
			newContext := context.WithValue(context.Background(), middlewares.OriginatingIdentityKey, expectedHeader)

			response, err := serviceBroker.Bind(newContext, instanceID, bindingID, bindDetails, false)
			Expect(err).ToNot(HaveOccurred())

			By("validating response")
			Expect(response).To(Equal(domain.Binding{
				IsAsync:       false,
				AlreadyExists: false,
				Credentials: map[string]any{
					"credhub-ref": "/c/csb/test-service/test-binding-id/secrets-and-services",
				},
			}))

			By("validating provider bind has been called")
			Expect(fakeServiceProvider.BindCallCount()).To(Equal(1))
			actualContext, _ := fakeServiceProvider.BindArgsForCall(0)
			Expect(actualContext.Value(middlewares.OriginatingIdentityKey)).To(Equal(expectedHeader))

			By("validating credstore delete has been called")
			Expect(fakeCredStore.PutCallCount()).To(Equal(1))
			Expect(fakeCredStore.AddPermissionCallCount()).To(Equal(1))

			By("validating storage is asked to store binding credentials")
			Expect(fakeStorage.StoreBindRequestDetailsCallCount()).To(Equal(1))
			actualBindRequest := fakeStorage.StoreBindRequestDetailsArgsForCall(0)
			Expect(actualBindRequest).To(Equal(storage.BindRequestDetails{
				ServiceInstanceGUID: instanceID,
				ServiceBindingGUID:  bindingID,
				RequestDetails: map[string]any{
					"bind_field_1": "bind_value_1",
				},
			}))
		})

		When("credstore disabled", func() {
			BeforeEach(func() {
				brokerConfig.Credstore = nil
				serviceBroker = must(broker.New(brokerConfig, fakeStorage, utils.NewLogger("bind-test-no-credstore")))
			})

			It("does not add the credentials to the credstore", func() {
				response, err := serviceBroker.Bind(context.TODO(), instanceID, bindingID, bindDetails, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(response).To(Equal(domain.Binding{
					IsAsync:       false,
					AlreadyExists: false,
					Credentials:   map[string]any{"fakeInstanceOutput": "fakeInstanceValue", "fakeOutput": "fakeValue"},
				}))

				Expect(fakeCredStore.PutCallCount()).To(Equal(0))
				Expect(fakeCredStore.AddPermissionCallCount()).To(Equal(0))
			})
		})

		Describe("bind variables", func() {
			When("bind variables are provided", func() {
				It("should use the variables in bind", func() {
					_, err := serviceBroker.Bind(context.TODO(), instanceID, bindingID, bindDetails, true)
					Expect(err).NotTo(HaveOccurred())

					By("validating the provider bind has been called with correct vars")
					_, actualVars := fakeServiceProvider.BindArgsForCall(0)
					Expect(actualVars.GetString("bind_field_1")).To(Equal("bind_value_1"))
				})
			})
		})

		Describe("computed variables", func() {
			It("passes computed variables to bind", func() {
				const header = "cloudfoundry eyANCiAgInVzZXJfaWQiOiAiNjgzZWE3NDgtMzA5Mi00ZmY0LWI2NTYtMzljYWNjNGQ1MzYwIg0KfQ=="
				newContext := context.WithValue(context.Background(), middlewares.OriginatingIdentityKey, header)

				_, err := serviceBroker.Bind(newContext, instanceID, bindingID, bindDetails, true)
				Expect(err).NotTo(HaveOccurred())

				By("validating provider bind has been called with the right vars")
				Expect(fakeServiceProvider.BindCallCount()).To(Equal(1))
				_, actualVars := fakeServiceProvider.BindArgsForCall(0)

				Expect(actualVars.GetString("copyOriginatingIdentity")).To(Equal(`{"platform":"cloudfoundry","value":{"user_id":"683ea748-3092-4ff4-b656-39cacc4d5360"}}`))
			})
		})

	})

	Describe("unsuccessful bind", func() {
		When("error reading binding credentials", func() {
			BeforeEach(func() {
				fakeStorage.ExistsServiceBindingCredentialsReturns(true, fmt.Errorf("error"))
			})

			It("should error", func() {
				_, err := serviceBroker.Bind(context.TODO(), instanceID, bindingID, bindDetails, false)

				Expect(err).To(MatchError("error checking for existing binding: error"))
			})
		})

		When("the service binding credentials already exist", func() {
			BeforeEach(func() {
				fakeStorage.ExistsServiceBindingCredentialsReturns(true, nil)
			})

			It("should return HTTP 409 as per OSBAPI spec", func() {
				_, err := serviceBroker.Bind(context.TODO(), instanceID, bindingID, bindDetails, false)

				Expect(err).To(MatchError(apiresponses.ErrBindingAlreadyExists))
			})
		})

		When("error retrieving service instance details", func() {
			BeforeEach(func() {
				fakeStorage.GetServiceInstanceDetailsReturns(storage.ServiceInstanceDetails{}, fmt.Errorf("error"))
			})

			It("should error", func() {
				_, err := serviceBroker.Bind(context.TODO(), instanceID, bindingID, bindDetails, false)

				Expect(err).To(MatchError("error retrieving service instance details: error"))
			})
		})

		When("error validating the service exists", func() {
			const nonExistentService = "non-existent-service"

			BeforeEach(func() {
				fakeStorage.GetServiceInstanceDetailsReturns(storage.ServiceInstanceDetails{ServiceGUID: nonExistentService}, nil)
			})

			It("should error", func() {
				_, err := serviceBroker.Bind(context.TODO(), instanceID, bindingID, bindDetails, false)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(fmt.Sprintf(`error retrieving service definition: unknown service ID: "%s"`, nonExistentService)))
			})
		})

		When("upgrade is available on instance", func() {
			It("should error", func() {
				fakeServiceProvider.CheckUpgradeAvailableReturns(fmt.Errorf("generic-error"))

				_, err := serviceBroker.Bind(context.TODO(), instanceID, bindingID, bindDetails, false)
				Expect(err).To(MatchError(`failed to bind: generic-error`))
			})
		})

		When("error parsing bind details", func() {
			BeforeEach(func() {
				bindDetails.RawParameters = json.RawMessage(`sadfasdf`)
			})

			It("should error", func() {
				_, err := serviceBroker.Bind(context.TODO(), instanceID, bindingID, bindDetails, false)

				Expect(err).To(MatchError(broker.ErrInvalidUserInput))
			})
		})

		When("error validating the plan exists", func() {
			const nonExistentPlan = "non-existent-plan"

			BeforeEach(func() {
				bindDetails.PlanID = nonExistentPlan
			})

			It("should error", func() {
				_, err := serviceBroker.Bind(context.TODO(), instanceID, bindingID, bindDetails, false)

				Expect(err).To(MatchError(fmt.Sprintf(`error getting service plan: plan ID "%s" could not be found`, nonExistentPlan)))
			})
		})

		When("error validating bind parameters", func() {

			BeforeEach(func() {
				bindDetails.RawParameters = json.RawMessage(`{"invalid_bind_field":"bind_value_1"}`)
			})

			It("should error", func() {
				_, err := serviceBroker.Bind(context.TODO(), instanceID, bindingID, bindDetails, false)

				Expect(err).To(MatchError(`additional properties are not allowed: invalid_bind_field`))
			})
		})

		When("provider bind fails", func() {
			BeforeEach(func() {
				fakeServiceProvider.BindReturns(nil, fmt.Errorf("bind-error"))
			})

			It("should error", func() {
				_, err := serviceBroker.Bind(context.TODO(), instanceID, bindingID, bindDetails, false)

				Expect(err).To(MatchError(`error performing bind: bind-error`))
			})
		})

		When("fails to store service binding credentials", func() {
			const saveError = "credential-save-error"

			BeforeEach(func() {
				fakeStorage.CreateServiceBindingCredentialsReturns(errors.New(saveError))
			})

			It("should error", func() {
				_, err := serviceBroker.Bind(context.TODO(), instanceID, bindingID, bindDetails, false)

				Expect(err).To(MatchError(fmt.Sprintf(`error saving credentials to database: %s. WARNING: these credentials cannot be unbound through cf. Please contact your operator for cleanup`, saveError)))
			})
		})

		When("fails to store service binding request details", func() {
			const saveBindRequestError = "bind-request-save-error"

			BeforeEach(func() {
				fakeStorage.StoreBindRequestDetailsReturns(errors.New(saveBindRequestError))
			})

			It("should error", func() {
				_, err := serviceBroker.Bind(context.TODO(), instanceID, bindingID, bindDetails, false)

				Expect(err).To(MatchError(fmt.Sprintf(`error saving bind request details to database: %s. Unbind operations will not be able to complete`, saveBindRequestError)))
			})
		})

		When("credstore fails to save key", func() {
			const credstoreError = "credstore-error"

			BeforeEach(func() {
				fakeCredStore.PutReturns(nil, errors.New(credstoreError))
			})

			It("should error", func() {
				_, err := serviceBroker.Bind(context.TODO(), instanceID, bindingID, bindDetails, false)

				Expect(err).To(MatchError(fmt.Sprintf(`bind failure: unable to put credentials in Credstore: %s`, credstoreError)))
			})
		})

		When("credstore fails to add permissions to app", func() {
			const credstorePermissionError = "credstore-error-permissions"

			BeforeEach(func() {
				fakeCredStore.AddPermissionReturns(nil, errors.New(credstorePermissionError))
			})

			It("should error", func() {
				_, err := serviceBroker.Bind(context.TODO(), instanceID, bindingID, bindDetails, false)

				Expect(err).To(MatchError(fmt.Sprintf(`bind failure: unable to add Credstore permissions to app: %s`, credstorePermissionError)))
			})
		})
	})
})
