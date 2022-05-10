package broker_test

import (
	"context"
	"fmt"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/cloud-service-broker/brokerapi/broker"
	"github.com/cloudfoundry/cloud-service-broker/brokerapi/broker/brokerfakes"
	"github.com/cloudfoundry/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	pkgBroker "github.com/cloudfoundry/cloud-service-broker/pkg/broker"
	pkgBrokerFakes "github.com/cloudfoundry/cloud-service-broker/pkg/broker/brokerfakes"
	"github.com/cloudfoundry/cloud-service-broker/pkg/credstore/credstorefakes"
	"github.com/cloudfoundry/cloud-service-broker/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v8/domain"
	"github.com/pivotal-cf/brokerapi/v8/domain/apiresponses"
	"github.com/pivotal-cf/brokerapi/v8/middlewares"
)

var _ = Describe("Unbind", func() {
	const (
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
		unbindDetails domain.UnbindDetails

		fakeStorage         *brokerfakes.FakeStorage
		fakeServiceProvider *pkgBrokerFakes.FakeServiceProvider
		fakeCredStore       *credstorefakes.FakeCredStore

		brokerConfig *broker.BrokerConfig
	)

	BeforeEach(func() {
		fakeServiceProvider = &pkgBrokerFakes.FakeServiceProvider{}
		fakeServiceProvider.UnbindReturns(nil)

		fakeStorage = &brokerfakes.FakeStorage{}
		fakeStorage.ExistsServiceBindingCredentialsReturns(true, nil)
		fakeStorage.GetServiceInstanceDetailsReturns(storage.ServiceInstanceDetails{
			GUID:             instanceID,
			ServiceGUID:      offeringID,
			PlanGUID:         planID,
			SpaceGUID:        spaceID,
			OrganizationGUID: orgID,
			OperationType:    models.ProvisionOperationType,
			OperationGUID:    "provision-operation-GUID",
		}, nil)
		fakeStorage.GetBindRequestDetailsReturns(storage.JSONObject{"foo": "bar"}, nil)

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
							ServiceProperties: map[string]interface{}{
								"plan-defined-key":       "plan-defined-value",
								"other-plan-defined-key": "other-plan-defined-value",
							},
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

		var err error
		serviceBroker, err = broker.New(brokerConfig, utils.NewLogger("unbind-test-with-credstore"), fakeStorage)
		Expect(err).ToNot(HaveOccurred())

		unbindDetails = domain.UnbindDetails{
			PlanID:    planID,
			ServiceID: serviceID,
		}
	})

	Describe("successful unbind", func() {
		BeforeEach(func() {
			fakeServiceProvider.UnbindReturns(nil)
			fakeStorage.ExistsServiceBindingCredentialsReturns(true, nil)
			fakeStorage.GetServiceInstanceDetailsReturns(storage.ServiceInstanceDetails{
				GUID:             instanceID,
				ServiceGUID:      offeringID,
				PlanGUID:         planID,
				SpaceGUID:        spaceID,
				OrganizationGUID: orgID,
				OperationType:    models.ProvisionOperationType,
				OperationGUID:    "provision-operation-GUID",
			}, nil)
			fakeStorage.GetBindRequestDetailsReturns(storage.JSONObject{"foo": "bar"}, nil)
		})

		It("should remove binding from database", func() {
			const expectedHeader = "cloudfoundry eyANCiAgInVzZXJfaWQiOiAiNjgzZWE3NDgtMzA5Mi00ZmY0LWI2NTYtMzljYWNjNGQ1MzYwIg0KfQ=="
			newContext := context.WithValue(context.Background(), middlewares.OriginatingIdentityKey, expectedHeader)

			response, err := serviceBroker.Unbind(newContext, instanceID, bindingID, unbindDetails, false)
			Expect(err).ToNot(HaveOccurred())

			By("validating response")
			Expect(response).To(Equal(domain.UnbindSpec{
				IsAsync:       false,
				OperationData: "",
			}))

			By("validating provider unbind has been called")
			Expect(fakeServiceProvider.UnbindCallCount()).To(Equal(1))
			actualContext, actualInstanceID, actualBindingID, _ := fakeServiceProvider.UnbindArgsForCall(0)
			Expect(actualContext.Value(middlewares.OriginatingIdentityKey)).To(Equal(expectedHeader))
			Expect(actualBindingID).To(Equal(bindingID))
			Expect(actualInstanceID).To(Equal(instanceID))

			By("validating credstore delete has been called")
			Expect(fakeCredStore.DeletePermissionCallCount()).To(Equal(1))
			Expect(fakeCredStore.DeleteCallCount()).To(Equal(1))

			By("validating storage is asked to delete binding credentials")
			Expect(fakeStorage.DeleteServiceBindingCredentialsCallCount()).To(Equal(1))
			actualBindingID, actualInstanceID = fakeStorage.DeleteServiceBindingCredentialsArgsForCall(0)
			Expect(actualBindingID).To(Equal(bindingID))
			Expect(actualInstanceID).To(Equal(instanceID))

			By("validating storage is asked to delete binding request details")
			Expect(fakeStorage.DeleteBindRequestDetailsCallCount()).To(Equal(1))
			actualBindingID, actualInstanceID = fakeStorage.DeleteBindRequestDetailsArgsForCall(0)
			Expect(actualBindingID).To(Equal(bindingID))
			Expect(actualInstanceID).To(Equal(instanceID))
		})

		When("credstore disabled", func() {
			BeforeEach(func() {
				brokerConfig.Credstore = nil
				var err error
				serviceBroker, err = broker.New(brokerConfig, utils.NewLogger("unbind-test-no-credstore"), fakeStorage)
				Expect(err).ToNot(HaveOccurred())
			})

			It("does not remove the credentials from the credstore", func() {
				response, err := serviceBroker.Unbind(context.TODO(), instanceID, bindingID, unbindDetails, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(response).To(BeZero())

				Expect(fakeCredStore.DeletePermissionCallCount()).To(Equal(0))
			})
		})

		Describe("unbind variables", func() {
			When("unbind variables are provided", func() {
				BeforeEach(func() {
					fakeStorage.GetBindRequestDetailsReturns(storage.JSONObject{"foo": "bar"}, nil)
				})

				It("should use the variables in unbind", func() {
					_, err := serviceBroker.Unbind(context.TODO(), instanceID, bindingID, unbindDetails, true)
					Expect(err).NotTo(HaveOccurred())

					By("validating the provider unbind has been called with correct vars")
					_, _, _, actualVars := fakeServiceProvider.UnbindArgsForCall(0)
					Expect(actualVars.GetString("foo")).To(Equal("bar"))
				})
			})

			Describe("computed variables", func() {
				It("passes computed variables to unbind", func() {
					const header = "cloudfoundry eyANCiAgInVzZXJfaWQiOiAiNjgzZWE3NDgtMzA5Mi00ZmY0LWI2NTYtMzljYWNjNGQ1MzYwIg0KfQ=="
					newContext := context.WithValue(context.Background(), middlewares.OriginatingIdentityKey, header)

					_, err := serviceBroker.Unbind(newContext, instanceID, bindingID, unbindDetails, true)
					Expect(err).NotTo(HaveOccurred())

					By("validating provider provision has been called with the right vars")
					Expect(fakeServiceProvider.UnbindCallCount()).To(Equal(1))
					_, _, _, actualVars := fakeServiceProvider.UnbindArgsForCall(0)

					Expect(actualVars.GetString("copyOriginatingIdentity")).To(Equal(`{"platform":"cloudfoundry","value":{"user_id":"683ea748-3092-4ff4-b656-39cacc4d5360"}}`))
				})
			})
		})
	})

	Describe("unsuccessful unbind", func() {
		When("error validating the service exists", func() {
			const nonExistentService = "non-existent-service"

			BeforeEach(func() {
				unbindDetails.ServiceID = nonExistentService
			})

			It("should error", func() {
				_, err := serviceBroker.Unbind(context.TODO(), instanceID, bindingID, unbindDetails, false)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(fmt.Sprintf(`unknown service ID: "%s"`, nonExistentService)))
			})
		})

		When("error validating the plan exists", func() {
			const nonExistentPlan = "non-existent-plan"

			BeforeEach(func() {
				unbindDetails.PlanID = nonExistentPlan
			})

			It("should error", func() {
				_, err := serviceBroker.Unbind(context.TODO(), instanceID, bindingID, unbindDetails, false)

				Expect(err).To(MatchError(fmt.Sprintf(`plan ID "%s" could not be found`, nonExistentPlan)))
			})
		})

		When("the service binding credentials do not exist", func() {
			BeforeEach(func() {
				fakeStorage.ExistsServiceBindingCredentialsReturns(false, nil)
			})

			It("should error", func() {
				_, err := serviceBroker.Unbind(context.TODO(), instanceID, bindingID, unbindDetails, false)

				Expect(err).To(MatchError(apiresponses.ErrBindingDoesNotExist))
			})
		})

		When("error reading binding credentials", func() {
			BeforeEach(func() {
				fakeStorage.ExistsServiceBindingCredentialsReturns(true, fmt.Errorf("error"))
			})

			It("should error", func() {
				_, err := serviceBroker.Unbind(context.TODO(), instanceID, bindingID, unbindDetails, false)

				Expect(err).To(MatchError("error locating service binding: error"))
			})
		})

		When("error retrieving service instance details", func() {
			BeforeEach(func() {
				fakeStorage.GetServiceInstanceDetailsReturns(storage.ServiceInstanceDetails{}, fmt.Errorf("error"))
			})

			It("should error", func() {
				_, err := serviceBroker.Unbind(context.TODO(), instanceID, bindingID, unbindDetails, false)

				Expect(err).To(MatchError("error retrieving service instance details: error"))
			})
		})

		When("error retrieving bind parameters", func() {
			BeforeEach(func() {
				fakeStorage.GetBindRequestDetailsReturns(storage.JSONObject{}, fmt.Errorf("error"))
			})

			It("should error", func() {
				_, err := serviceBroker.Unbind(context.TODO(), instanceID, bindingID, unbindDetails, false)

				Expect(err).To(MatchError(fmt.Sprintf(`error retrieving bind request details for %q: error`, instanceID)))
			})
		})

		When("provider unbind fails", func() {
			BeforeEach(func() {
				fakeServiceProvider.UnbindReturns(fmt.Errorf("unbind-error"))
			})

			It("should error", func() {
				_, err := serviceBroker.Unbind(context.TODO(), instanceID, bindingID, unbindDetails, false)

				Expect(err).To(MatchError(`unbind-error`))
			})
		})

		When("fails to delete service binding credentials", func() {
			const deleteError = "credential-delete-error"

			BeforeEach(func() {
				fakeStorage.DeleteServiceBindingCredentialsReturns(fmt.Errorf(deleteError))
			})

			It("should error", func() {
				_, err := serviceBroker.Unbind(context.TODO(), instanceID, bindingID, unbindDetails, false)

				Expect(err).To(MatchError(fmt.Sprintf(`error soft-deleting credentials from database: %s. WARNING: these credentials will remain visible in cf. Contact your operator for cleanup`, deleteError)))
			})
		})

		When("fails to delete binding request details", func() {
			const deleteError = "bind-details-delete-error"

			BeforeEach(func() {
				fakeStorage.DeleteBindRequestDetailsReturns(fmt.Errorf(deleteError))
			})

			It("should error", func() {
				_, err := serviceBroker.Unbind(context.TODO(), instanceID, bindingID, unbindDetails, false)

				Expect(err).To(MatchError(fmt.Sprintf(`error soft-deleting bind request details from database: %s`, deleteError)))
			})
		})

		When("credstore fails to delete key", func() {
			BeforeEach(func() {
				fakeCredStore.DeleteReturns(fmt.Errorf("credstore-error"))
			})

			It("should error", func() {
				_, err := serviceBroker.Unbind(context.TODO(), instanceID, bindingID, unbindDetails, false)

				Expect(err).To(MatchError(`credstore-error`))
			})
		})
	})
})
