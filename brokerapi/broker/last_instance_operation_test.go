package broker_test

import (
	"errors"

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

var _ = Describe("LastInstanceOperation", func() {
	const (
		spaceID     = "test-space-id"
		orgID       = "test-org-id"
		planID      = "test-plan-id"
		offeringID  = "test-service-id"
		instanceID  = "test-instance-id"
		operationID = "test-operation-id"
	)

	var (
		serviceBroker *broker.ServiceBroker
		pollDetails   domain.PollDetails

		fakeStorage         *brokerfakes.FakeStorage
		fakeServiceProvider *pkgBrokerFakes.FakeServiceProvider
		expectedTFOutput    storage.JSONObject
	)

	BeforeEach(func() {
		fakeServiceProvider = &pkgBrokerFakes.FakeServiceProvider{}
		fakeServiceProvider.ProvisionsAsyncReturns(true)
		fakeServiceProvider.DeprovisionsAsyncReturns(true)
		expectedTFOutput = storage.JSONObject{"output": "value"}
		fakeServiceProvider.GetTerraformOutputsReturns(expectedTFOutput, nil)
		fakeServiceProvider.PollInstanceReturns(true, "operation complete", nil)

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
						},
					},
					ProviderBuilder: providerBuilder,
				},
			},
		}

		fakeStorage = &brokerfakes.FakeStorage{}
		fakeStorage.GetServiceInstanceDetailsReturns(
			storage.ServiceInstanceDetails{
				GUID:             instanceID,
				OperationType:    models.ProvisionOperationType,
				OperationGUID:    operationID,
				PlanGUID:         planID,
				ServiceGUID:      offeringID,
				SpaceGUID:        spaceID,
				OrganizationGUID: orgID,
			}, nil)

		var err error
		serviceBroker, err = broker.New(brokerConfig, utils.NewLogger("brokers-test"), fakeStorage)
		Expect(err).ToNot(HaveOccurred())

		pollDetails = domain.PollDetails{
			ServiceID:     offeringID,
			PlanID:        planID,
			OperationData: operationID,
		}
	})

	Describe("operation complete", func() {
		Describe("provision", func() {
			BeforeEach(func() {
				fakeServiceProvider.PollInstanceReturns(true, "operation complete", nil)
			})

			It("should complete provision", func() {
				expectedHeader := "cloudfoundry eyANCiAgInVzZXJfaWQiOiAiNjgzZWE3NDgtMzA5Mi00ZmY0LWI2NTYtMzljYWNjNGQ1MzYwIg0KfQ=="
				newContext := context.WithValue(context.Background(), middlewares.OriginatingIdentityKey, expectedHeader)

				response, err := serviceBroker.LastOperation(newContext, instanceID, pollDetails)
				Expect(err).ToNot(HaveOccurred())

				By("validating response")
				Expect(response.State).To(Equal(domain.Succeeded))
				Expect(response.Description).To(Equal("operation complete"))

				By("validating that provider polling has occurred")
				Expect(fakeServiceProvider.PollInstanceCallCount()).To(Equal(1))
				actualContext, actualInstanceID := fakeServiceProvider.PollInstanceArgsForCall(0)
				Expect(actualContext.Value(middlewares.OriginatingIdentityKey)).To(Equal(expectedHeader))
				Expect(actualInstanceID).To(Equal(instanceID))

				By("validating that new instance details are stored")
				Expect(fakeStorage.StoreServiceInstanceDetailsCallCount()).To(Equal(1))
				actualSIDetails := fakeStorage.StoreServiceInstanceDetailsArgsForCall(0)
				Expect(actualSIDetails.GUID).To(Equal(instanceID))
				Expect(actualSIDetails.Outputs).To(Equal(expectedTFOutput))
				Expect(actualSIDetails.OperationGUID).To(BeEmpty())
				Expect(actualSIDetails.OperationType).To(BeEmpty())
				Expect(actualSIDetails.ServiceGUID).To(Equal(offeringID))
				Expect(actualSIDetails.PlanGUID).To(Equal(planID))
				Expect(actualSIDetails.SpaceGUID).To(Equal(spaceID))
				Expect(actualSIDetails.OrganizationGUID).To(Equal(orgID))
			})
		})

		Describe("deprovision", func() {
			BeforeEach(func() {
				fakeStorage.GetServiceInstanceDetailsReturns(
					storage.ServiceInstanceDetails{
						GUID:             instanceID,
						OperationType:    models.DeprovisionOperationType,
						OperationGUID:    operationID,
						PlanGUID:         planID,
						ServiceGUID:      offeringID,
						SpaceGUID:        spaceID,
						OrganizationGUID: orgID,
					}, nil)

				fakeServiceProvider.PollInstanceReturns(true, "operation complete", nil)
			})

			It("should complete deprovision", func() {
				expectedHeader := "cloudfoundry eyANCiAgInVzZXJfaWQiOiAiNjgzZWE3NDgtMzA5Mi00ZmY0LWI2NTYtMzljYWNjNGQ1MzYwIg0KfQ=="
				newContext := context.WithValue(context.Background(), middlewares.OriginatingIdentityKey, expectedHeader)

				response, err := serviceBroker.LastOperation(newContext, instanceID, pollDetails)
				Expect(err).ToNot(HaveOccurred())

				By("validating response")
				Expect(response.State).To(Equal(domain.Succeeded))
				Expect(response.Description).To(Equal("operation complete"))

				By("validating that provider polling has occurred")
				Expect(fakeServiceProvider.PollInstanceCallCount()).To(Equal(1))
				actualContext, actualInstanceID := fakeServiceProvider.PollInstanceArgsForCall(0)
				Expect(actualContext.Value(middlewares.OriginatingIdentityKey)).To(Equal(expectedHeader))
				Expect(actualInstanceID).To(Equal(instanceID))

				By("validating that instance details are removed")
				Expect(fakeStorage.DeleteServiceInstanceDetailsCallCount()).To(Equal(1))
				actualSI := fakeStorage.DeleteServiceInstanceDetailsArgsForCall(0)
				Expect(actualSI).To(Equal(instanceID))

				By("validating that provision request parameters are removed")
				Expect(fakeStorage.DeleteProvisionRequestDetailsCallCount()).To(Equal(1))
				actualSIID := fakeStorage.DeleteProvisionRequestDetailsArgsForCall(0)
				Expect(actualSIID).To(Equal(instanceID))

				Expect(fakeStorage.StoreServiceInstanceDetailsCallCount()).To(Equal(0))
			})
		})
	})

	Describe("operation in progress", func() {
		BeforeEach(func() {
			fakeServiceProvider.PollInstanceReturns(false, "operation in progress still", nil)
		})

		It("should not update the service instance", func() {
			response, err := serviceBroker.LastOperation(context.TODO(), instanceID, pollDetails)
			Expect(err).ToNot(HaveOccurred())

			By("validating response")
			Expect(response.State).To(Equal(domain.InProgress))
			Expect(response.Description).To(Equal("operation in progress still"))

			By("validating that provider polling has occurred")
			Expect(fakeServiceProvider.PollInstanceCallCount()).To(Equal(1))
			_, actualInstanceID := fakeServiceProvider.PollInstanceArgsForCall(0)
			Expect(actualInstanceID).To(Equal(instanceID))

			By("validating that SI has not been updated")
			Expect(fakeStorage.StoreServiceInstanceDetailsCallCount()).To(Equal(0))
			Expect(fakeStorage.DeleteProvisionRequestDetailsCallCount()).To(Equal(0))
			Expect(fakeStorage.DeleteProvisionRequestDetailsCallCount()).To(Equal(0))
		})
	})

	Describe("operation failed", func() {
		BeforeEach(func() {
			fakeServiceProvider.PollInstanceReturns(false, "there was an error", errors.New("some error happened"))
		})

		It("should set operation to failed", func() {
			response, err := serviceBroker.LastOperation(context.TODO(), instanceID, pollDetails)
			Expect(err).ToNot(HaveOccurred())

			By("validating response")
			Expect(response.State).To(Equal(domain.Failed))
			Expect(response.Description).To(Equal("some error happened"))

			By("validating that provider polling has occurred")
			Expect(fakeServiceProvider.PollInstanceCallCount()).To(Equal(1))
			_, actualInstanceID := fakeServiceProvider.PollInstanceArgsForCall(0)
			Expect(actualInstanceID).To(Equal(instanceID))

			By("validating that SI has not been updated")
			Expect(fakeStorage.StoreServiceInstanceDetailsCallCount()).To(Equal(0))
			Expect(fakeStorage.DeleteProvisionRequestDetailsCallCount()).To(Equal(0))
			Expect(fakeStorage.DeleteProvisionRequestDetailsCallCount()).To(Equal(0))
		})
	})

	Describe("sync operations", func() {
		When("async provision and deprovision is not supported", func() {
			BeforeEach(func() {
				fakeServiceProvider.ProvisionsAsyncReturns(false)
				fakeServiceProvider.DeprovisionsAsyncReturns(false)
			})

			It("should error", func() {
				_, err := serviceBroker.LastOperation(context.TODO(), instanceID, pollDetails)
				Expect(err).To(MatchError("This service plan requires client support for asynchronous service operations."))
			})
		})
	})

	Describe("storage errors", func() {
		Context("storage errors when getting SI details", func() {
			BeforeEach(func() {
				fakeStorage.GetServiceInstanceDetailsReturns(storage.ServiceInstanceDetails{}, errors.New("failed to get SI details"))
			})

			It("should error", func() {
				_, err := serviceBroker.LastOperation(context.TODO(), instanceID, pollDetails)
				Expect(err).To(MatchError("instance does not exist"))
			})
		})

		Context("storage errors when storing SI details", func() {
			BeforeEach(func() {
				fakeStorage.StoreServiceInstanceDetailsReturns(errors.New("failed to store SI details"))
			})

			It("should error", func() {
				result, err := serviceBroker.LastOperation(context.TODO(), instanceID, pollDetails)
				Expect(err).To(MatchError("error saving instance details to database failed to store SI details"))
				Expect(result.State).To(Equal(domain.Succeeded))
				Expect(result.Description).To(Equal("operation complete"))
			})
		})

		Context("storage errors when deleting service instance details", func() {
			BeforeEach(func() {
				fakeStorage.GetServiceInstanceDetailsReturns(
					storage.ServiceInstanceDetails{
						GUID:          instanceID,
						OperationType: models.DeprovisionOperationType,
						ServiceGUID:   offeringID,
					}, nil)
				fakeStorage.DeleteServiceInstanceDetailsReturns(errors.New("failed to delete SI details"))
			})

			It("should error", func() {
				result, err := serviceBroker.LastOperation(context.TODO(), instanceID, pollDetails)
				Expect(err).To(MatchError("error deleting instance details from database: failed to delete SI details. WARNING: this instance will remain visible in cf. Contact your operator for cleanup"))
				Expect(result.State).To(Equal(domain.Succeeded))
				Expect(result.Description).To(Equal("operation complete"))
			})
		})

		Context("storage errors when deleting service provision params", func() {
			BeforeEach(func() {
				fakeStorage.GetServiceInstanceDetailsReturns(
					storage.ServiceInstanceDetails{
						GUID:          instanceID,
						OperationType: models.DeprovisionOperationType,
						ServiceGUID:   offeringID,
					}, nil)
				fakeStorage.DeleteProvisionRequestDetailsReturns(errors.New("failed to delete provision params"))
			})

			It("should error", func() {
				result, err := serviceBroker.LastOperation(context.TODO(), instanceID, pollDetails)
				Expect(err).To(MatchError("error deleting provision request details from the database: failed to delete provision params"))
				Expect(result.State).To(Equal(domain.Succeeded))
				Expect(result.Description).To(Equal("operation complete"))
			})
		})
	})
})
