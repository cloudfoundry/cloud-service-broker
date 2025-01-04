package broker_test

import (
	"errors"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/lager/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v11/domain"
	"github.com/pivotal-cf/brokerapi/v11/domain/apiresponses"
	"golang.org/x/net/context"

	"github.com/cloudfoundry/cloud-service-broker/v2/brokerapi/broker/brokerfakes"
	"github.com/cloudfoundry/cloud-service-broker/v2/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/storage"
	pkgBroker "github.com/cloudfoundry/cloud-service-broker/v2/pkg/broker"
	pkgBrokerFakes "github.com/cloudfoundry/cloud-service-broker/v2/pkg/broker/brokerfakes"
	"github.com/cloudfoundry/cloud-service-broker/v2/utils"

	"github.com/cloudfoundry/cloud-service-broker/v2/brokerapi/broker"
)

var _ = Describe("GetInstance", func() {

	const (
		orgID      = "test-org-id"
		spaceID    = "test-space-id"
		planID     = "test-plan-id"
		offeringID = "test-service-id"
		instanceID = "test-instance-id"
	)

	var (
		serviceBroker *broker.ServiceBroker

		fakeStorage         *brokerfakes.FakeStorage
		fakeServiceProvider *pkgBrokerFakes.FakeServiceProvider

		brokerConfig *broker.BrokerConfig

		provisionParams *storage.JSONObject
	)

	BeforeEach(func() {
		fakeStorage = &brokerfakes.FakeStorage{}
		fakeServiceProvider = &pkgBrokerFakes.FakeServiceProvider{}

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
						},
					},
					ProviderBuilder: providerBuilder,
				},
			},
		}

		serviceBroker = must(broker.New(brokerConfig, fakeStorage, utils.NewLogger("get-instance-test")))

		provisionParams = &storage.JSONObject{
			"param1": "value1",
			"param2": 3,
			"param3": true,
			"param4": []string{"a", "b", "c"},
			"param5": map[string]string{"key1": "value", "key2": "value"},
			"param6": struct {
				A string
				B string
			}{"a", "b"},
		}

		fakeStorage.ExistsServiceInstanceDetailsReturns(true, nil)
		fakeStorage.GetServiceInstanceDetailsReturns(
			storage.ServiceInstanceDetails{
				GUID:             instanceID,
				Name:             "test-instance",
				Outputs:          storage.JSONObject{},
				ServiceGUID:      offeringID,
				PlanGUID:         planID,
				SpaceGUID:        spaceID,
				OrganizationGUID: orgID,
			}, nil)
		fakeServiceProvider.PollInstanceReturns(true, "", models.ProvisionOperationType, nil) // Operation status is provision succeeded
		fakeStorage.GetProvisionRequestDetailsReturns(*provisionParams, nil)
	})

	When("instance exists and provision succeeded", func() {
		It("returns instance details", func() {
			response, err := serviceBroker.GetInstance(context.TODO(), instanceID, domain.FetchInstanceDetails{ServiceID: offeringID, PlanID: planID})
			Expect(err).ToNot(HaveOccurred())

			By("validating response")
			Expect(response.ServiceID).To(Equal(offeringID))
			Expect(response.PlanID).To(Equal(planID))
			Expect(response.Parameters).To(BeEquivalentTo(*provisionParams))
			Expect(response.DashboardURL).To(BeEmpty()) // Broker does not set dashboard URL
			Expect(response.Metadata).To(BeZero())      // Broker does not support instance metadata

			By("validating storage is asked whether instance exists")
			Expect(fakeStorage.ExistsServiceInstanceDetailsCallCount()).To(Equal(1))

			By("validating storage is asked for instance details")
			Expect(fakeStorage.GetServiceInstanceDetailsCallCount()).To(Equal(1))

			By("validating service provider asked for instance status")
			Expect(fakeServiceProvider.PollInstanceCallCount()).To(Equal(1))

			By("validating storage is asked for provision request details")
			Expect(fakeStorage.GetProvisionRequestDetailsCallCount()).To(Equal(1))
		})
	})

	When("instance does not exist", func() {
		BeforeEach(func() {
			fakeStorage.ExistsServiceInstanceDetailsReturns(false, nil)
		})
		It("returns status code 404 (not found)", func() {
			response, err := serviceBroker.GetInstance(context.TODO(), instanceID, domain.FetchInstanceDetails{ServiceID: offeringID, PlanID: planID})

			By("validating response")
			Expect(response).To(BeZero())

			By("validating error")
			apiErr, isFailureResponse := err.(*apiresponses.FailureResponse)
			Expect(isFailureResponse).To(BeTrue())                                 // must be a failure response
			Expect(apiErr.Error()).To(Equal("not found"))                          // must contain "Not Found" error message
			Expect(apiErr.ValidatedStatusCode(nil)).To(Equal(http.StatusNotFound)) // status code must be 404

			By("validating storage is asked whether instance exists")
			Expect(fakeStorage.ExistsServiceInstanceDetailsCallCount()).To(Equal(1))

			By("validating storage is not asked for instance details")
			Expect(fakeStorage.GetServiceInstanceDetailsCallCount()).To(Equal(0))

			By("validating service provider is not asked for instance status")
			Expect(fakeServiceProvider.PollInstanceCallCount()).To(Equal(0))

			By("validating storage is not asked for provision request details")
			Expect(fakeStorage.GetProvisionRequestDetailsCallCount()).To(Equal(0))
		})
	})

	When("instance exists and provision is in progress", func() {
		BeforeEach(func() {
			fakeServiceProvider.PollInstanceReturns(false, "", models.ProvisionOperationType, nil) // Operation status is provision in progress
		})
		// According to OSB Spec, broker must return 404 in case provision is in progress
		It("returns status code 404 (not found)", func() {
			response, err := serviceBroker.GetInstance(context.TODO(), instanceID, domain.FetchInstanceDetails{ServiceID: offeringID, PlanID: planID})

			By("validating response")
			Expect(response).To(BeZero())

			By("validating error")
			apiErr, isFailureResponse := err.(*apiresponses.FailureResponse)
			Expect(isFailureResponse).To(BeTrue())                                 // must be a failure response
			Expect(apiErr.Error()).To(Equal("not found"))                          // must contain "not found" error message
			Expect(apiErr.ValidatedStatusCode(nil)).To(Equal(http.StatusNotFound)) // status code must be 404

			By("validating storage is asked whether instance exists")
			Expect(fakeStorage.ExistsServiceInstanceDetailsCallCount()).To(Equal(1))

			By("validating storage is asked for instance details")
			Expect(fakeStorage.GetServiceInstanceDetailsCallCount()).To(Equal(1))

			By("validating service provider is asked for instance status")
			Expect(fakeServiceProvider.PollInstanceCallCount()).To(Equal(1))

			By("validating storage is not asked for provision request details")
			Expect(fakeStorage.GetProvisionRequestDetailsCallCount()).To(Equal(0))
		})

	})

	When("instance exists and update is in progress", func() {
		BeforeEach(func() {
			fakeServiceProvider.PollInstanceReturns(false, "", models.UpdateOperationType, nil) // Operation status is update in progress
		})
		It("returns status code 422 (Unprocessable Entity) and error code ConcurrencyError", func() {
			response, err := serviceBroker.GetInstance(context.TODO(), instanceID, domain.FetchInstanceDetails{ServiceID: offeringID, PlanID: planID})

			By("validating response")
			Expect(response).To(BeZero())

			By("validating error")
			apiErr, isFailureResponse := err.(*apiresponses.FailureResponse)
			Expect(isFailureResponse).To(BeTrue())                                            // must be a failure response
			Expect(apiErr.Error()).To(Equal("ConcurrencyError"))                              // must contain "ConcurrencyError" error message
			Expect(apiErr.ValidatedStatusCode(nil)).To(Equal(http.StatusUnprocessableEntity)) // status code must be 404

			By("validating storage is asked whether instance exists")
			Expect(fakeStorage.ExistsServiceInstanceDetailsCallCount()).To(Equal(1))

			By("validating storage is asked for instance details")
			Expect(fakeStorage.GetServiceInstanceDetailsCallCount()).To(Equal(1))

			By("validating service provider is asked for instance status")
			Expect(fakeServiceProvider.PollInstanceCallCount()).To(Equal(1))

			By("validating storage is not asked for provision request details")
			Expect(fakeStorage.GetProvisionRequestDetailsCallCount()).To(Equal(0))
		})

	})

	When("service_id is not set", func() {
		It("ignores service_id and returns instance details", func() {
			response, err := serviceBroker.GetInstance(context.TODO(), instanceID, domain.FetchInstanceDetails{PlanID: planID})
			Expect(err).ToNot(HaveOccurred())
			Expect(response.ServiceID).To(Equal(offeringID))
		})
	})

	When("service_id does not match service for instance", func() {
		It("returns 404 (not found)", func() {
			response, err := serviceBroker.GetInstance(context.TODO(), instanceID, domain.FetchInstanceDetails{ServiceID: "otherService", PlanID: planID})

			By("validating response")
			Expect(response).To(BeZero())

			By("validating error")
			apiErr, isFailureResponse := err.(*apiresponses.FailureResponse)
			Expect(isFailureResponse).To(BeTrue())                                 // must be a failure response
			Expect(apiErr.Error()).To(Equal("not found"))                          // must contain "not found" error message
			Expect(apiErr.ValidatedStatusCode(nil)).To(Equal(http.StatusNotFound)) // status code must be 404

			By("validating storage is asked whether instance exists")
			Expect(fakeStorage.ExistsServiceInstanceDetailsCallCount()).To(Equal(1))

			By("validating storage is asked for instance details")
			Expect(fakeStorage.GetServiceInstanceDetailsCallCount()).To(Equal(1))

			By("validating service provider is asked for instance status")
			Expect(fakeServiceProvider.PollInstanceCallCount()).To(Equal(0))

			By("validating storage is not asked for provision request details")
			Expect(fakeStorage.GetProvisionRequestDetailsCallCount()).To(Equal(0))
		})
	})

	When("plan_id is not set", func() {
		It("ignores plan_id and returns instance details", func() {
			response, err := serviceBroker.GetInstance(context.TODO(), instanceID, domain.FetchInstanceDetails{ServiceID: offeringID})
			Expect(err).ToNot(HaveOccurred())
			Expect(response.ServiceID).To(Equal(offeringID))
		})
	})

	When("plan_id does not match plan for instance", func() {
		It("returns 404 (not found)", func() {
			response, err := serviceBroker.GetInstance(context.TODO(), instanceID, domain.FetchInstanceDetails{ServiceID: offeringID, PlanID: "otherPlan"})

			By("validating response")
			Expect(response).To(BeZero())

			By("validating error")
			apiErr, isFailureResponse := err.(*apiresponses.FailureResponse)
			Expect(isFailureResponse).To(BeTrue())                                 // must be a failure response
			Expect(apiErr.Error()).To(Equal("not found"))                          // must contain "not found" error message
			Expect(apiErr.ValidatedStatusCode(nil)).To(Equal(http.StatusNotFound)) // status code must be 404

			By("validating storage is asked whether instance exists")
			Expect(fakeStorage.ExistsServiceInstanceDetailsCallCount()).To(Equal(1))

			By("validating storage is asked for instance details")
			Expect(fakeStorage.GetServiceInstanceDetailsCallCount()).To(Equal(1))

			By("validating service provider is asked for instance status")
			Expect(fakeServiceProvider.PollInstanceCallCount()).To(Equal(0))

			By("validating storage is not asked for provision request details")
			Expect(fakeStorage.GetProvisionRequestDetailsCallCount()).To(Equal(0))
		})
	})

	When("fails to check for existing instance", func() {
		const (
			msg = "error-msg"
		)
		BeforeEach(func() {
			fakeStorage.ExistsServiceInstanceDetailsReturns(false, errors.New(msg))
		})
		It("should error", func() {
			_, err := serviceBroker.GetInstance(context.TODO(), instanceID, domain.FetchInstanceDetails{ServiceID: offeringID, PlanID: "otherPlan"})
			Expect(err).To(MatchError(fmt.Sprintf(`error checking for existing instance: %s`, msg)))
		})
	})
	When("fails to retrieve service instance details", func() {
		const (
			msg = "error-msg"
		)
		BeforeEach(func() {
			fakeStorage.GetServiceInstanceDetailsReturns(storage.ServiceInstanceDetails{}, errors.New(msg))
		})
		It("should error", func() {
			_, err := serviceBroker.GetInstance(context.TODO(), instanceID, domain.FetchInstanceDetails{ServiceID: offeringID, PlanID: "otherPlan"})
			Expect(err).To(MatchError(fmt.Sprintf(`error retrieving service instance details: %s`, msg)))
		})
	})
	When("fails to poll instance status", func() {
		const (
			msg = "error-msg"
		)
		BeforeEach(func() {
			fakeServiceProvider.PollInstanceReturns(false, "", "", errors.New(msg))
		})
		It("should error", func() {
			_, err := serviceBroker.GetInstance(context.TODO(), instanceID, domain.FetchInstanceDetails{ServiceID: offeringID, PlanID: planID})
			Expect(err).To(MatchError(fmt.Sprintf(`error polling instance status: %s`, msg)))
		})
	})
	When("fails to retrieve provision request details", func() {
		const (
			msg = "error-msg"
		)
		BeforeEach(func() {
			fakeStorage.GetProvisionRequestDetailsReturns(nil, errors.New(msg))
		})
		It("should error", func() {
			_, err := serviceBroker.GetInstance(context.TODO(), instanceID, domain.FetchInstanceDetails{ServiceID: offeringID, PlanID: planID})
			Expect(err).To(MatchError(fmt.Sprintf(`error retrieving provision request details: %s`, msg)))
		})
	})
})
