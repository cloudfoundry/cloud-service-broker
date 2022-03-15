package broker_test

import (
	"errors"

	"github.com/cloudfoundry/cloud-service-broker/pkg/varcontext"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/cloud-service-broker/brokerapi/broker/brokerfakes"
	"github.com/cloudfoundry/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	pkgBroker "github.com/cloudfoundry/cloud-service-broker/pkg/broker"
	pkgBrokerFakes "github.com/cloudfoundry/cloud-service-broker/pkg/broker/brokerfakes"
	"github.com/cloudfoundry/cloud-service-broker/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v8/domain"
	"github.com/pivotal-cf/brokerapi/v8/middlewares"
	"golang.org/x/net/context"

	"github.com/cloudfoundry/cloud-service-broker/brokerapi/broker"
)

var _ = Describe("Deprovision", func() {
	const (
		planID             = "test-plan-id"
		offeringID         = "test-service-id"
		instanceToDeleteID = "test-instance-id"
	)

	var (
		serviceBroker      *broker.ServiceBroker
		deprovisionDetails domain.DeprovisionDetails
		operationID        string

		fakeStorage         *brokerfakes.FakeStorage
		fakeServiceProvider *pkgBrokerFakes.FakeServiceProvider
	)

	BeforeEach(func() {
		fakeServiceProvider = &pkgBrokerFakes.FakeServiceProvider{}
		fakeServiceProvider.DeprovisionsAsyncReturns(true)
		operationID = "test-operation-id"
		fakeServiceProvider.DeprovisionReturns(&operationID, nil)

		providerBuilder := func(logger lager.Logger, store pkgBroker.ServiceProviderStorage) pkgBroker.ServiceProvider {
			return fakeServiceProvider
		}
		brokerConfig := &broker.BrokerConfig{
			Registry: pkgBroker.BrokerRegistry{
				"test-service": &pkgBroker.ServiceDefinition{
					Id:   offeringID,
					Name: "test-service",
					Plans: []pkgBroker.ServicePlan{
						{
							ServicePlan: domain.ServicePlan{
								ID:   planID,
								Name: "test-plan",
							},
							ServiceProperties: map[string]interface{}{
								"plan-defined-key":       "plan-defined-value",
								"other-plan-defined-key": "other-plan-defined-value",
							},
						},
					},
					ProvisionComputedVariables: []varcontext.DefaultVariable{
						{Name: "labels", Default: "${json.marshal(request.default_labels)}", Overwrite: true},
						{Name: "copyOriginatingIdentity", Default: "${json.marshal(request.x_broker_api_originating_identity)}", Overwrite: true},
					},
					ProviderBuilder: providerBuilder,
				},
			},
		}

		fakeStorage = &brokerfakes.FakeStorage{}
		fakeStorage.ExistsServiceInstanceDetailsReturns(true, nil)
		fakeStorage.GetProvisionRequestDetailsReturns(nil, nil)
		fakeStorage.GetServiceInstanceDetailsReturns(storage.ServiceInstanceDetails{
			ServiceGUID:      offeringID,
			PlanGUID:         planID,
			GUID:             instanceToDeleteID,
			SpaceGUID:        "test-space",
			OrganizationGUID: "test-org",
			OperationType:    "provision",
			OperationGUID:    operationID,
		}, nil)

		var err error
		serviceBroker, err = broker.New(brokerConfig, utils.NewLogger("brokers-test"), fakeStorage)
		Expect(err).ToNot(HaveOccurred())

		deprovisionDetails = domain.DeprovisionDetails{
			ServiceID: offeringID,
			PlanID:    planID,
		}
	})

	Describe("successful sync deletion", func() {
		BeforeEach(func() {
			fakeServiceProvider.DeprovisionReturns(nil, nil)
		})

		It("deletes the instance", func() {
			expectedHeader := "cloudfoundry eyANCiAgInVzZXJfaWQiOiAiNjgzZWE3NDgtMzA5Mi00ZmY0LWI2NTYtMzljYWNjNGQ1MzYwIg0KfQ=="
			newContext := context.WithValue(context.Background(), middlewares.OriginatingIdentityKey, expectedHeader)

			response, err := serviceBroker.Deprovision(newContext, instanceToDeleteID, deprovisionDetails, true)
			Expect(err).ToNot(HaveOccurred())

			By("validating response")
			Expect(response.IsAsync).To(BeFalse())
			Expect(response.OperationData).To(BeEmpty())

			By("validating call to deprovision")
			Expect(fakeServiceProvider.DeprovisionCallCount()).To(Equal(1))
			actualCtx, instanceID, actualDetails, _ := fakeServiceProvider.DeprovisionArgsForCall(0)
			Expect(actualCtx.Value(middlewares.OriginatingIdentityKey)).To(Equal(expectedHeader))
			Expect(instanceID).To(Equal(instanceToDeleteID))
			Expect(actualDetails).To(Equal(deprovisionDetails))

			By("validating SI details delete call")
			Expect(fakeStorage.DeleteServiceInstanceDetailsCallCount()).To(Equal(1))
			actualInstanceID := fakeStorage.DeleteServiceInstanceDetailsArgsForCall(0)
			Expect(actualInstanceID).To(Equal(instanceToDeleteID))

			By("validating provision parameters delete call")
			Expect(fakeStorage.DeleteProvisionRequestDetailsCallCount()).To(Equal(1))
			actualInstance := fakeStorage.DeleteProvisionRequestDetailsArgsForCall(0)
			Expect(actualInstance).To(Equal(instanceToDeleteID))
		})

		Describe("deprovision variables", func() {
			When("there were provision variables during provision or update", func() {
				BeforeEach(func() {
					fakeStorage.GetProvisionRequestDetailsReturns(map[string]interface{}{"foo": "something", "import_field_1": "hello"}, nil)
				})

				It("should use the provision variables", func() {
					_, err := serviceBroker.Deprovision(context.TODO(), instanceToDeleteID, deprovisionDetails, true)
					Expect(err).ToNot(HaveOccurred())

					By("validating provider provision has been called with the right vars")
					Expect(fakeServiceProvider.DeprovisionCallCount()).To(Equal(1))
					_, _, _, actualVars := fakeServiceProvider.DeprovisionArgsForCall(0)
					Expect(actualVars.GetString("foo")).To(Equal("something"))
					Expect(actualVars.GetString("import_field_1")).To(Equal("hello"))
				})
			})

			Describe("provision computed variables", func() {
				It("passes computed variables to deprovision", func() {
					header := "cloudfoundry eyANCiAgInVzZXJfaWQiOiAiNjgzZWE3NDgtMzA5Mi00ZmY0LWI2NTYtMzljYWNjNGQ1MzYwIg0KfQ=="
					newContext := context.WithValue(context.Background(), middlewares.OriginatingIdentityKey, header)

					_, err := serviceBroker.Deprovision(newContext, instanceToDeleteID, deprovisionDetails, true)
					Expect(err).ToNot(HaveOccurred())

					By("validating provider provision has been called with the right vars")
					Expect(fakeServiceProvider.DeprovisionCallCount()).To(Equal(1))
					_, _, _, actualVars := fakeServiceProvider.DeprovisionArgsForCall(0)

					Expect(actualVars.GetString("copyOriginatingIdentity")).To(Equal(`{"platform":"cloudfoundry","value":{"user_id":"683ea748-3092-4ff4-b656-39cacc4d5360"}}`))
					Expect(actualVars.GetString("labels")).To(Equal(`{"pcf-instance-id":"test-instance-id","pcf-organization-guid":"","pcf-space-guid":""}`))
				})
			})

			It("passes plan provided service properties", func() {
				_, err := serviceBroker.Deprovision(context.TODO(), instanceToDeleteID, deprovisionDetails, true)
				Expect(err).ToNot(HaveOccurred())

				By("validating provider provision has been called with the right vars")
				Expect(fakeServiceProvider.DeprovisionCallCount()).To(Equal(1))
				_, _, _, actualVars := fakeServiceProvider.DeprovisionArgsForCall(0)
				Expect(actualVars.GetString("plan-defined-key")).To(Equal("plan-defined-value"))
				Expect(actualVars.GetString("other-plan-defined-key")).To(Equal("other-plan-defined-value"))
			})
		})
	})

	Describe("async deletion", func() {
		It("should return operationID and not remove the instance", func() {
			response, err := serviceBroker.Deprovision(context.TODO(), instanceToDeleteID, deprovisionDetails, true)
			Expect(err).ToNot(HaveOccurred())

			By("validating response")
			Expect(response.IsAsync).To(BeTrue())
			Expect(response.OperationData).To(Equal(operationID))

			By("validating call to deprovision")
			Expect(fakeServiceProvider.DeprovisionCallCount()).To(Equal(1))

			By("validating delete calls have not happen")
			Expect(fakeStorage.DeleteServiceInstanceDetailsCallCount()).To(Equal(0))
			Expect(fakeStorage.DeleteProvisionRequestDetailsCallCount()).To(Equal(0))

			By("validating SI details storing call")
			Expect(fakeStorage.StoreServiceInstanceDetailsCallCount()).To(Equal(1))
			actualSIDetails := fakeStorage.StoreServiceInstanceDetailsArgsForCall(0)
			Expect(actualSIDetails.GUID).To(Equal(instanceToDeleteID))
			Expect(actualSIDetails.OperationType).To(Equal(models.DeprovisionOperationType))
			Expect(actualSIDetails.OperationGUID).To(Equal(operationID))
		})
	})

	When("provider deprovision errors", func() {
		BeforeEach(func() {
			fakeServiceProvider.DeprovisionReturns(nil, errors.New("cannot deprovision right now"))
		})

		It("should error", func() {
			_, err := serviceBroker.Deprovision(context.TODO(), instanceToDeleteID, deprovisionDetails, true)
			Expect(err).To(MatchError("cannot deprovision right now"))
		})
	})

	When("instance does not exists", func() {
		BeforeEach(func() {
			fakeStorage.ExistsServiceInstanceDetailsReturns(false, nil)
		})

		It("should error", func() {
			_, err := serviceBroker.Deprovision(context.TODO(), instanceToDeleteID, deprovisionDetails, true)
			Expect(err).To(MatchError("instance does not exist"))
		})
	})

	When("offering does not exists", func() {
		BeforeEach(func() {
			fakeStorage.GetServiceInstanceDetailsReturns(storage.ServiceInstanceDetails{
				ServiceGUID:   "non-existent-offering",
				PlanGUID:      planID,
				GUID:          instanceToDeleteID,
				OperationType: "provision",
				OperationGUID: "opGUID",
			}, nil)
		})

		It("should error", func() {
			_, err := serviceBroker.Deprovision(context.TODO(), instanceToDeleteID, deprovisionDetails, true)
			Expect(err).To(MatchError(`unknown service ID: "non-existent-offering"`))
		})
	})

	When("plan does not exists", func() {
		It("should error", func() {
			deprovisionDetails = domain.DeprovisionDetails{
				ServiceID: offeringID,
				PlanID:    "some-non-existent-plan",
			}

			_, err := serviceBroker.Deprovision(context.TODO(), instanceToDeleteID, deprovisionDetails, true)
			Expect(err).To(MatchError(`plan ID "some-non-existent-plan" could not be found`))
		})
	})

	When("client cannot accept async", func() {
		It("should error", func() {
			_, err := serviceBroker.Deprovision(context.TODO(), instanceToDeleteID, deprovisionDetails, false)
			Expect(err).To(MatchError("This service plan requires client support for asynchronous service operations."))
		})
	})

	Describe("storage errors", func() {
		When("storage errors when checking SI existence", func() {
			BeforeEach(func() {
				fakeStorage.ExistsServiceInstanceDetailsReturns(false, errors.New("failed when checking existence"))
			})

			It("should error", func() {
				_, err := serviceBroker.Deprovision(context.TODO(), instanceToDeleteID, deprovisionDetails, true)
				Expect(err).To(MatchError("database error checking for existing instance: failed when checking existence"))
			})
		})

		When("storage errors when getting SI details", func() {
			BeforeEach(func() {
				fakeStorage.GetServiceInstanceDetailsReturns(storage.ServiceInstanceDetails{}, errors.New("failed to get SI details"))
			})

			It("should error", func() {
				_, err := serviceBroker.Deprovision(context.TODO(), instanceToDeleteID, deprovisionDetails, true)
				Expect(err).To(MatchError("database error getting existing instance: failed to get SI details"))
			})
		})

		When("storage errors when getting provision params", func() {
			BeforeEach(func() {
				fakeStorage.GetProvisionRequestDetailsReturns(nil, errors.New("failed to get SI provision params"))
			})

			It("should error", func() {
				_, err := serviceBroker.Deprovision(context.TODO(), instanceToDeleteID, deprovisionDetails, true)
				Expect(err).To(MatchError(`error retrieving provision request details for "test-instance-id": failed to get SI provision params`))
			})
		})

		When("storage errors when storing SI details", func() {
			BeforeEach(func() {
				fakeStorage.StoreServiceInstanceDetailsReturns(errors.New("failed to store SI details"))
			})

			It("should error", func() {
				_, err := serviceBroker.Deprovision(context.TODO(), instanceToDeleteID, deprovisionDetails, true)
				Expect(err).To(MatchError("error saving instance details to database: failed to store SI details. WARNING: this instance will remain visible in cf. Contact your operator for cleanup"))
			})
		})

		When("storage errors when deleting service instance details", func() {
			BeforeEach(func() {
				fakeServiceProvider.DeprovisionReturns(nil, nil)
				fakeStorage.DeleteServiceInstanceDetailsReturns(errors.New("failed to delete SI details"))
			})

			It("should error", func() {
				_, err := serviceBroker.Deprovision(context.TODO(), instanceToDeleteID, deprovisionDetails, true)
				Expect(err).To(MatchError("error deleting instance details from database: failed to delete SI details. WARNING: this instance will remain visible in cf. Contact your operator for cleanup"))
			})
		})

		When("storage errors when deleting service provision params", func() {
			BeforeEach(func() {
				fakeServiceProvider.DeprovisionReturns(nil, nil)
				fakeStorage.DeleteProvisionRequestDetailsReturns(errors.New("failed to delete provision params"))
			})

			It("should error", func() {
				_, err := serviceBroker.Deprovision(context.TODO(), instanceToDeleteID, deprovisionDetails, true)
				Expect(err).To(MatchError("error deleting provision request details from the database: failed to delete provision params"))
			})
		})
	})
})
