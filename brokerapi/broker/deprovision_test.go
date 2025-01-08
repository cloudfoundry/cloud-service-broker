package broker_test

import (
	"errors"
	"fmt"

	"code.cloudfoundry.org/lager/v3"
	"github.com/cloudfoundry/cloud-service-broker/v2/brokerapi/broker"
	"github.com/cloudfoundry/cloud-service-broker/v2/brokerapi/broker/brokerfakes"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/storage"
	pkgBroker "github.com/cloudfoundry/cloud-service-broker/v2/pkg/broker"
	pkgBrokerFakes "github.com/cloudfoundry/cloud-service-broker/v2/pkg/broker/brokerfakes"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/v2/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v12/domain"
	"github.com/pivotal-cf/brokerapi/v12/domain/apiresponses"
	"golang.org/x/net/context"
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
		operationID = "test-operation-id"
		fakeServiceProvider.DeprovisionReturns(&operationID, nil)

		providerBuilder := func(logger lager.Logger, store pkgBroker.ServiceProviderStorage) pkgBroker.ServiceProvider {
			return fakeServiceProvider
		}
		brokerConfig := &broker.BrokerConfig{
			Registry: pkgBroker.BrokerRegistry{
				"test-service": &pkgBroker.ServiceDefinition{
					ID:   offeringID,
					Name: "test-service",
					Plans: []pkgBroker.ServicePlan{
						{
							ServicePlan: domain.ServicePlan{
								ID:   planID,
								Name: "test-plan",
							},
							ServiceProperties: map[string]any{
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
		}, nil)
		fakeStorage.GetTerraformDeploymentReturns(storage.TerraformDeployment{
			ID:                instanceToDeleteID,
			LastOperationType: "provision",
		}, nil)

		serviceBroker = must(broker.New(brokerConfig, fakeStorage, utils.NewLogger("brokers-test")))

		deprovisionDetails = domain.DeprovisionDetails{
			ServiceID: offeringID,
			PlanID:    planID,
		}
	})

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

		It("should return HTTP 410 as per OSBAPI spec", func() {
			_, err := serviceBroker.Deprovision(context.TODO(), instanceToDeleteID, deprovisionDetails, true)
			Expect(err).To(MatchError(apiresponses.ErrInstanceDoesNotExist))
		})
	})

	When("offering does not exists", func() {
		BeforeEach(func() {
			fakeStorage.GetServiceInstanceDetailsReturns(storage.ServiceInstanceDetails{
				ServiceGUID: "non-existent-offering",
				PlanGUID:    planID,
				GUID:        instanceToDeleteID,
			}, nil)

			fakeStorage.GetTerraformDeploymentReturns(storage.TerraformDeployment{
				LastOperationType: "provision",
				ID:                instanceToDeleteID,
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

	When("upgrade is available on instance", func() {
		It("should error", func() {
			deprovisionDetails = domain.DeprovisionDetails{
				ServiceID: offeringID,
				PlanID:    "some-non-existent-plan",
			}
			fakeServiceProvider.CheckUpgradeAvailableReturns(fmt.Errorf("generic-error"))

			_, err := serviceBroker.Deprovision(context.TODO(), instanceToDeleteID, deprovisionDetails, true)
			Expect(err).To(MatchError(`failed to delete: generic-error`))
		})
	})

	When("provision is in progress for the instance", func() {
		It("should error", func() {
			deprovisionDetails = domain.DeprovisionDetails{
				ServiceID: offeringID,
				PlanID:    "some-non-existent-plan",
			}
			fakeServiceProvider.CheckOperationConstraintsReturns(fmt.Errorf("generic-error"))

			_, err := serviceBroker.Deprovision(context.TODO(), instanceToDeleteID, deprovisionDetails, true)
			Expect(err).To(MatchError(`generic-error`))
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
	})
})
