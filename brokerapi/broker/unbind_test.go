package broker_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/pkg/credstore/credstorefakes"
	"github.com/cloudfoundry/cloud-service-broker/pkg/varcontext"

	"code.cloudfoundry.org/lager"
	pkgBroker "github.com/cloudfoundry/cloud-service-broker/pkg/broker"
	"github.com/cloudfoundry/cloud-service-broker/utils"

	"github.com/cloudfoundry/cloud-service-broker/brokerapi/broker"
	"github.com/cloudfoundry/cloud-service-broker/brokerapi/broker/brokerfakes"
	pkgBrokerFakes "github.com/cloudfoundry/cloud-service-broker/pkg/broker/brokerfakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v8/domain"
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
	)

	BeforeEach(func() {
		fakeServiceProvider = &pkgBrokerFakes.FakeServiceProvider{}
		fakeCredStore = &credstorefakes.FakeCredStore{}

		providerBuilder := func(logger lager.Logger, store pkgBroker.ServiceProviderStorage) pkgBroker.ServiceProvider {
			return fakeServiceProvider
		}
		planUpdatable := true

		brokerConfig := &broker.BrokerConfig{
			Registry: pkgBroker.BrokerRegistry{
				"test-service": &pkgBroker.ServiceDefinition{
					Id:   offeringID,
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
						{
							ServicePlan: domain.ServicePlan{
								ID:            planID,
								Name:          "new-test-plan",
								PlanUpdatable: &planUpdatable,
							},
							ServiceProperties: map[string]interface{}{
								"new-plan-defined-key":       "plan-defined-value",
								"new-other-plan-defined-key": "other-plan-defined-value",
							},
						},
					},
					ProvisionInputVariables: []pkgBroker.BrokerVariable{
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
						{
							FieldName:      "prohibit-update-field",
							Type:           "string",
							Details:        "fake field name",
							ProhibitUpdate: true,
						},
					},
					ProvisionComputedVariables: []varcontext.DefaultVariable{
						{Name: "labels", Default: "${json.marshal(request.default_labels)}", Overwrite: true},
						{Name: "copyOriginatingIdentity", Default: "${json.marshal(request.x_broker_api_originating_identity)}", Overwrite: true},
					},
					ProviderBuilder: providerBuilder,
				},
			},
			Credstore: fakeCredStore,
		}

		fakeStorage = &brokerfakes.FakeStorage{}

		var err error
		serviceBroker, err = broker.New(brokerConfig, utils.NewLogger("brokers-test"), fakeStorage)
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
			fakeStorage.GetBindRequestDetailsReturns(json.RawMessage(`{"decrypted":{"foo":"bar"}}`), nil)
		})

		It("should remove binding from database", func() {
			expectedHeader := "cloudfoundry eyANCiAgInVzZXJfaWQiOiAiNjgzZWE3NDgtMzA5Mi00ZmY0LWI2NTYtMzljYWNjNGQ1MzYwIg0KfQ=="
			newContext := context.WithValue(context.Background(), middlewares.OriginatingIdentityKey, expectedHeader)

			response, err := serviceBroker.Unbind(newContext, instanceID, bindingID, unbindDetails, false)
			Expect(err).ToNot(HaveOccurred())

			By("validating response")
			Expect(response).To(BeZero())

			By("validating provider unbind has been called")
			Expect(fakeServiceProvider.UnbindCallCount()).To(Equal(1))
			actualContext, _, _, _ := fakeServiceProvider.UnbindArgsForCall(0)
			Expect(actualContext.Value(middlewares.OriginatingIdentityKey)).To(Equal(expectedHeader))

			By("validating DeleteServiceBindingCredentials has been called")
			Expect(fakeStorage.DeleteServiceBindingCredentialsCallCount()).To(Equal(1))
			actualBindingId, actualInstanceID := fakeStorage.DeleteServiceBindingCredentialsArgsForCall(0)
			Expect(actualBindingId).To(Equal(bindingID))
			Expect(actualInstanceID).To(Equal(instanceID))

			By("validating DeleteBindRequestDetails has been called")
			Expect(fakeStorage.DeleteBindRequestDetailsCallCount()).To(Equal(1))
			actualBindingId, actualInstanceID = fakeStorage.DeleteBindRequestDetailsArgsForCall(0)
			Expect(actualBindingId).To(Equal(bindingID))
			Expect(actualInstanceID).To(Equal(instanceID))
		})

	})

	Describe("unsuccessful unbind", func() {
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
			fakeStorage.GetBindRequestDetailsReturns(json.RawMessage(`{"decrypted":{"foo":"bar"}}`), nil)
		})

		When("error validating the service exists", func() {
			const nonExistentService = "non-existent-service"

			BeforeEach(func() {
				unbindDetails = domain.UnbindDetails{
					PlanID:    planID,
					ServiceID: nonExistentService,
				}
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
				unbindDetails = domain.UnbindDetails{
					PlanID:    nonExistentPlan,
					ServiceID: serviceID,
				}
			})
			It("should error", func() {
				_, err := serviceBroker.Unbind(context.TODO(), instanceID, bindingID, unbindDetails, false)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(fmt.Sprintf(`plan ID "%s" could not be found`, nonExistentPlan)))
			})
		})
		When("the service binding credentials do not exist", func() {
			BeforeEach(func() {
				fakeStorage.ExistsServiceBindingCredentialsReturns(false, nil)
			})
			It("should error", func() {
				_, err := serviceBroker.Unbind(context.TODO(), instanceID, bindingID, unbindDetails, false)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("binding does not exist"))
			})
		})
		When("error reading binding credentials", func() {
			BeforeEach(func() {
				fakeStorage.ExistsServiceBindingCredentialsReturns(true, fmt.Errorf("error"))
			})
			It("should error", func() {
				_, err := serviceBroker.Unbind(context.TODO(), instanceID, bindingID, unbindDetails, false)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("error locating service binding: error"))
			})
		})
		When("error retrieving service instance details", func() {
			BeforeEach(func() {
				fakeStorage.GetServiceInstanceDetailsReturns(storage.ServiceInstanceDetails{}, fmt.Errorf("error"))
			})
			It("should error", func() {
				_, err := serviceBroker.Unbind(context.TODO(), instanceID, bindingID, unbindDetails, false)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("error retrieving service instance details: error"))
			})
		})
		When("error retrieving bind parameters", func() {
			BeforeEach(func() {
				fakeStorage.GetBindRequestDetailsReturns(json.RawMessage{}, fmt.Errorf("error"))
			})
			It("should error", func() {
				_, err := serviceBroker.Unbind(context.TODO(), instanceID, bindingID, unbindDetails, false)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(fmt.Sprintf(`error retrieving bind request details for %q: error`, instanceID)))
			})
		})
		When("provider unbind fails", func() {
			BeforeEach(func() {
				fakeServiceProvider.UnbindReturns(fmt.Errorf("unbind-error"))
			})
			It("should error", func() {
				_, err := serviceBroker.Unbind(context.TODO(), instanceID, bindingID, unbindDetails, false)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(fmt.Sprintf(`unbind-error`)))
			})
		})
		When("fails to delete service binding credentials", func() {
			deleteError := "credential-delete-error"
			BeforeEach(func() {
				fakeStorage.DeleteServiceBindingCredentialsReturns(fmt.Errorf(deleteError))
			})
			It("should error", func() {
				_, err := serviceBroker.Unbind(context.TODO(), instanceID, bindingID, unbindDetails, false)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(fmt.Sprintf(`error soft-deleting credentials from database: %s. WARNING: these credentials will remain visible in cf. Contact your operator for cleanup`, deleteError)))
			})
		})
		When("fails to delete binding request details", func() {
			deleteError := "bind-details-delete-error"
			BeforeEach(func() {
				fakeStorage.DeleteBindRequestDetailsReturns(fmt.Errorf(deleteError))
			})
			It("should error", func() {
				_, err := serviceBroker.Unbind(context.TODO(), instanceID, bindingID, unbindDetails, false)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(fmt.Sprintf(`error soft-deleting bind request details from database: %s`, deleteError)))
			})
		})
		When("credstore fails to delete key", func() {
			BeforeEach(func() {
				fakeCredStore.DeleteReturns(fmt.Errorf("credstore-error"))
			})
			It("should error", func() {
				_, err := serviceBroker.Unbind(context.TODO(), instanceID, bindingID, unbindDetails, false)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(`credstore-error`))
			})
		})

	})

})
