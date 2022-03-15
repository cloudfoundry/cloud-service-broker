package broker_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/db_service/models"

	"github.com/cloudfoundry/cloud-service-broker/pkg/varcontext"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/cloud-service-broker/brokerapi/broker"
	"github.com/cloudfoundry/cloud-service-broker/brokerapi/broker/brokerfakes"
	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	pkgBroker "github.com/cloudfoundry/cloud-service-broker/pkg/broker"
	pkgBrokerFakes "github.com/cloudfoundry/cloud-service-broker/pkg/broker/brokerfakes"
	"github.com/cloudfoundry/cloud-service-broker/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v8/domain"
	"github.com/pivotal-cf/brokerapi/v8/middlewares"
	"github.com/spf13/viper"
)

var _ = Describe("Provision", func() {
	const (
		spaceID       = "test-space-id"
		orgID         = "test-org-id"
		planID        = "test-plan-id"
		offeringID    = "test-service-id"
		newInstanceID = "test-instance-id"
		operationID   = "test-operation-id"
	)

	var (
		serviceBroker    *broker.ServiceBroker
		provisionDetails domain.ProvisionDetails

		fakeStorage         *brokerfakes.FakeStorage
		fakeServiceProvider *pkgBrokerFakes.FakeServiceProvider
	)

	BeforeEach(func() {
		fakeServiceProvider = &pkgBrokerFakes.FakeServiceProvider{}
		fakeServiceProvider.ProvisionsAsyncReturns(true)
		fakeServiceProvider.ProvisionReturns(storage.ServiceInstanceDetails{
			OperationType: models.ProvisionOperationType,
			OperationGUID: operationID,
		}, nil)

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
					},
					ImportInputVariables: []pkgBroker.ImportVariable{
						{
							Name:       "import_field_1",
							Type:       "string",
							Details:    "fake import field",
							TfResource: "fake.tf.resource",
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
		fakeStorage.ExistsServiceInstanceDetailsReturns(false, nil)

		var err error
		serviceBroker, err = broker.New(brokerConfig, utils.NewLogger("brokers-test"), fakeStorage)
		Expect(err).ToNot(HaveOccurred())

		provisionDetails = domain.ProvisionDetails{
			ServiceID:        offeringID,
			PlanID:           planID,
			SpaceGUID:        spaceID,
			OrganizationGUID: orgID,
			RawContext:       json.RawMessage(fmt.Sprintf(`{"organization_guid":%q, "space_guid": %q}`, orgID, spaceID)),
		}
	})

	Describe("successful creation", func() {
		It("should provision without parameters", func() {
			expectedHeader := "cloudfoundry eyANCiAgInVzZXJfaWQiOiAiNjgzZWE3NDgtMzA5Mi00ZmY0LWI2NTYtMzljYWNjNGQ1MzYwIg0KfQ=="
			newContext := context.WithValue(context.Background(), middlewares.OriginatingIdentityKey, expectedHeader)

			response, err := serviceBroker.Provision(newContext, newInstanceID, provisionDetails, true)
			Expect(err).ToNot(HaveOccurred())

			By("validating response")
			Expect(response.IsAsync).To(BeTrue())
			Expect(response.DashboardURL).To(BeEmpty())
			Expect(response.OperationData).To(Equal(operationID))

			By("validating provider provision has been called")
			Expect(fakeServiceProvider.ProvisionCallCount()).To(Equal(1))
			actualContext, _ := fakeServiceProvider.ProvisionArgsForCall(0)
			Expect(actualContext.Value(middlewares.OriginatingIdentityKey)).To(Equal(expectedHeader))

			By("validating SI details storing call")
			Expect(fakeStorage.StoreServiceInstanceDetailsCallCount()).To(Equal(1))
			actualSIDetails := fakeStorage.StoreServiceInstanceDetailsArgsForCall(0)
			Expect(actualSIDetails.GUID).To(Equal(newInstanceID))
			Expect(actualSIDetails.ServiceGUID).To(Equal(offeringID))
			Expect(actualSIDetails.PlanGUID).To(Equal(planID))
			Expect(actualSIDetails.SpaceGUID).To(Equal(spaceID))
			Expect(actualSIDetails.OrganizationGUID).To(Equal(orgID))

			By("validating provision parameters storing call")
			Expect(fakeStorage.StoreProvisionRequestDetailsCallCount()).To(Equal(1))
			actualSI, actualParams := fakeStorage.StoreProvisionRequestDetailsArgsForCall(0)
			Expect(actualSI).To(Equal(newInstanceID))
			Expect(actualParams).To(BeNil())
		})

		It("should provision with parameters", func() {
			expectedParams := storage.JSONObject{"foo": "something", "import_field_1": "hello"}
			provisionDetails = domain.ProvisionDetails{
				ServiceID:     offeringID,
				PlanID:        planID,
				RawParameters: json.RawMessage(`{"foo":"something", "import_field_1":"hello"}`),
			}

			_, err := serviceBroker.Provision(context.TODO(), newInstanceID, provisionDetails, true)
			Expect(err).ToNot(HaveOccurred())

			By("validating provision has been called")
			Expect(fakeServiceProvider.ProvisionCallCount()).To(Equal(1))
			_, actualVars := fakeServiceProvider.ProvisionArgsForCall(0)
			Expect(actualVars.GetString("foo")).To(Equal("something"))
			Expect(actualVars.GetString("import_field_1")).To(Equal("hello"))

			By("validating provision parameters storing call")
			Expect(fakeStorage.StoreProvisionRequestDetailsCallCount()).To(Equal(1))
			actualSI, actualParams := fakeStorage.StoreProvisionRequestDetailsArgsForCall(0)
			Expect(actualSI).To(Equal(newInstanceID))
			Expect(actualParams).To(Equal(expectedParams))
		})

		Describe("provision variables", func() {
			It("passes plan provided service properties", func() {
				_, err := serviceBroker.Provision(context.TODO(), newInstanceID, provisionDetails, true)
				Expect(err).ToNot(HaveOccurred())

				By("validating provider provision has been called with the right vars")
				Expect(fakeServiceProvider.ProvisionCallCount()).To(Equal(1))
				_, actualVars := fakeServiceProvider.ProvisionArgsForCall(0)
				Expect(actualVars.GetString("plan-defined-key")).To(Equal("plan-defined-value"))
				Expect(actualVars.GetString("other-plan-defined-key")).To(Equal("other-plan-defined-value"))
			})

			Describe("provision computed variables", func() {
				It("passes computed variables", func() {
					header := "cloudfoundry eyANCiAgInVzZXJfaWQiOiAiNjgzZWE3NDgtMzA5Mi00ZmY0LWI2NTYtMzljYWNjNGQ1MzYwIg0KfQ=="
					newContext := context.WithValue(context.Background(), middlewares.OriginatingIdentityKey, header)

					_, err := serviceBroker.Provision(newContext, newInstanceID, provisionDetails, true)
					Expect(err).ToNot(HaveOccurred())

					By("validating provider provision has been called with the right vars")
					Expect(fakeServiceProvider.ProvisionCallCount()).To(Equal(1))
					_, actualVars := fakeServiceProvider.ProvisionArgsForCall(0)

					Expect(actualVars.GetString("copyOriginatingIdentity")).To(Equal(`{"platform":"cloudfoundry","value":{"user_id":"683ea748-3092-4ff4-b656-39cacc4d5360"}}`))
					Expect(actualVars.GetString("labels")).To(Equal(`{"pcf-instance-id":"test-instance-id","pcf-organization-guid":"test-org-id","pcf-space-guid":"test-space-id"}`))
				})
			})
		})
	})

	Describe("invalid provision parameters", func() {
		When("additional properties are passed", func() {
			It("should error", func() {
				provisionDetails = domain.ProvisionDetails{
					ServiceID:     offeringID,
					PlanID:        planID,
					RawParameters: json.RawMessage(`{"invalid_parameter":42,"foo":"bar","other_invalid":false}`),
				}

				_, err := serviceBroker.Provision(context.TODO(), "new-instance", provisionDetails, true)
				Expect(err).To(MatchError("additional properties are not allowed: invalid_parameter, other_invalid"))
			})
		})

		When("plan defined properties are passed", func() {
			It("should error", func() {
				provisionDetails = domain.ProvisionDetails{
					ServiceID:     offeringID,
					PlanID:        planID,
					RawParameters: json.RawMessage(`{"foo":"bar","plan-defined-key":42,"other-plan-defined-key":"test","other_invalid":false}`),
				}

				_, err := serviceBroker.Provision(context.TODO(), "new-instance", provisionDetails, true)
				Expect(err).To(MatchError("plan defined properties cannot be changed: other-plan-defined-key, plan-defined-key"))
			})
		})

		When("property validation is disabled", func() {
			It("should not error", func() {
				viper.Set(broker.DisableRequestPropertyValidation, true)

				provisionDetails = domain.ProvisionDetails{
					ServiceID:     offeringID,
					PlanID:        planID,
					RawParameters: json.RawMessage(`{"invalid_parameter":42,"foo":"bar","other_invalid":false,"plan-defined-key":42}`),
				}

				_, err := serviceBroker.Provision(context.TODO(), "new-instance", provisionDetails, true)
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				viper.Set(broker.DisableRequestPropertyValidation, false)
			})
		})
	})

	When("provider provision errors", func() {
		BeforeEach(func() {
			fakeServiceProvider.ProvisionReturns(storage.ServiceInstanceDetails{}, errors.New("cannot provision right now"))
		})

		It("should error", func() {
			_, err := serviceBroker.Provision(context.TODO(), "new-instance", provisionDetails, true)
			Expect(err).To(MatchError("cannot provision right now"))
		})
	})

	When("instance already exists", func() {
		BeforeEach(func() {
			fakeStorage.ExistsServiceInstanceDetailsReturns(true, nil)
		})

		It("should error", func() {
			_, err := serviceBroker.Provision(context.TODO(), "new-instance", provisionDetails, true)
			Expect(err).To(MatchError("instance already exists"))
		})
	})

	When("plan does not exists", func() {
		It("should error", func() {
			provisionDetails = domain.ProvisionDetails{
				ServiceID: offeringID,
				PlanID:    "some-non-existent-plan",
			}

			_, err := serviceBroker.Provision(context.TODO(), "new-instance", provisionDetails, true)
			Expect(err).To(MatchError(`plan ID "some-non-existent-plan" could not be found`))
		})
	})

	When("offering does not exists", func() {
		It("should error", func() {
			provisionDetails = domain.ProvisionDetails{
				ServiceID: "some-non-existent-offering",
				PlanID:    planID,
			}

			_, err := serviceBroker.Provision(context.TODO(), "new-instance", provisionDetails, true)
			Expect(err).To(MatchError(`unknown service ID: "some-non-existent-offering"`))
		})
	})

	When("request json is invalid", func() {
		It("should error", func() {
			provisionDetails = domain.ProvisionDetails{
				ServiceID:     offeringID,
				PlanID:        planID,
				RawParameters: json.RawMessage("{invalid json"),
			}

			_, err := serviceBroker.Provision(context.TODO(), "new-instance", provisionDetails, true)
			Expect(err).To(MatchError("User supplied parameters must be in the form of a valid JSON map."))
		})
	})

	When("client cannot accept async", func() {
		It("should error", func() {
			_, err := serviceBroker.Provision(context.TODO(), "new-instance", provisionDetails, false)
			Expect(err).To(MatchError("This service plan requires client support for asynchronous service operations."))
		})
	})

	Describe("storage errors", func() {
		When("storage errors when checking SI details", func() {
			BeforeEach(func() {
				fakeStorage.ExistsServiceInstanceDetailsReturns(false, errors.New("failed to check existence"))
			})

			It("should error", func() {
				_, err := serviceBroker.Provision(context.TODO(), "new-instance", provisionDetails, true)
				Expect(err).To(MatchError("database error checking for existing instance: failed to check existence"))
			})
		})

		When("storage errors when storing SI details", func() {
			BeforeEach(func() {
				fakeStorage.StoreServiceInstanceDetailsReturns(errors.New("failed to store SI details"))
			})

			It("should error", func() {
				_, err := serviceBroker.Provision(context.TODO(), "new-instance", provisionDetails, true)
				Expect(err).To(MatchError("error saving instance details to database: failed to store SI details. WARNING: this instance cannot be deprovisioned through cf. Contact your operator for cleanup"))
			})
		})

		When("storage errors when storing provision parameters", func() {
			BeforeEach(func() {
				fakeStorage.StoreProvisionRequestDetailsReturns(errors.New("failed to store provision parameters"))
			})

			It("should error", func() {
				_, err := serviceBroker.Provision(context.TODO(), "new-instance", provisionDetails, true)
				Expect(err).To(MatchError("error saving provision request details to database: failed to store provision parameters. Services relying on async provisioning will not be able to complete provisioning"))
			})
		})
	})
})
