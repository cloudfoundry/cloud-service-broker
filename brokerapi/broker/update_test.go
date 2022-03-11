package broker_test

import (
	"encoding/json"
	"errors"
	"fmt"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/cloud-service-broker/brokerapi/broker/brokerfakes"
	"github.com/cloudfoundry/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	pkgBroker "github.com/cloudfoundry/cloud-service-broker/pkg/broker"
	pkgBrokerFakes "github.com/cloudfoundry/cloud-service-broker/pkg/broker/brokerfakes"
	"github.com/cloudfoundry/cloud-service-broker/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v8/domain"
	"github.com/pivotal-cf/brokerapi/v8/middlewares"
	"github.com/spf13/viper"
	"golang.org/x/net/context"

	"github.com/cloudfoundry/cloud-service-broker/brokerapi/broker"
)

var _ = Describe("Update", func() {
	const (
		spaceID           = "test-space-id"
		orgID             = "test-org-id"
		originalPlanID    = "test-plan-id"
		offeringID        = "test-service-id"
		newPlanID         = "new-test-plan-id"
		instanceID        = "test-instance-id"
		updateOperationID = "test-operation-id"
	)

	var (
		serviceBroker *broker.ServiceBroker
		updateDetails domain.UpdateDetails

		fakeStorage         *brokerfakes.FakeStorage
		fakeServiceProvider *pkgBrokerFakes.FakeServiceProvider
	)

	BeforeEach(func() {
		fakeServiceProvider = &pkgBrokerFakes.FakeServiceProvider{}
		fakeServiceProvider.ProvisionsAsyncReturns(true)

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
								ID:            originalPlanID,
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
								ID:            newPlanID,
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
		}

		fakeStorage = &brokerfakes.FakeStorage{}
		fakeStorage.ExistsServiceInstanceDetailsReturns(true, nil)
		fakeStorage.GetServiceInstanceDetailsReturns(storage.ServiceInstanceDetails{
			GUID:             instanceID,
			ServiceGUID:      offeringID,
			PlanGUID:         originalPlanID,
			SpaceGUID:        spaceID,
			OrganizationGUID: orgID,
			OperationType:    models.ProvisionOperationType,
			OperationGUID:    "provision-operation-GUID",
		}, nil)

		var err error
		serviceBroker, err = broker.New(brokerConfig, utils.NewLogger("brokers-test"), fakeStorage)
		Expect(err).ToNot(HaveOccurred())

		updateDetails = domain.UpdateDetails{
			ServiceID: offeringID,
			PlanID:    originalPlanID,
			PreviousValues: domain.PreviousValues{
				PlanID:    originalPlanID,
				ServiceID: offeringID,
				OrgID:     orgID,
				SpaceID:   spaceID,
			},
			RawContext: json.RawMessage(fmt.Sprintf(`{"organization_guid":%q, "space_guid": %q}`, orgID, spaceID)),
		}
	})

	Describe("successful update", func() {
		When("no plan or parameter changes are requested", func() {
			BeforeEach(func() {
				fakeServiceProvider.UpdateReturns(models.ServiceInstanceDetails{
					OperationType: models.UpdateOperationType,
					OperationId:   updateOperationID,
				}, nil)
			})

			It("should complete without changing instance", func() {
				expectedHeader := "cloudfoundry eyANCiAgInVzZXJfaWQiOiAiNjgzZWE3NDgtMzA5Mi00ZmY0LWI2NTYtMzljYWNjNGQ1MzYwIg0KfQ=="
				newContext := context.WithValue(context.Background(), middlewares.OriginatingIdentityKey, expectedHeader)

				response, err := serviceBroker.Update(newContext, instanceID, updateDetails, true)
				Expect(err).ToNot(HaveOccurred())

				By("validating response")
				Expect(response.IsAsync).To(BeTrue())
				Expect(response.DashboardURL).To(BeEmpty())
				Expect(response.OperationData).To(Equal(updateOperationID))

				By("validating provider update has been called")
				Expect(fakeServiceProvider.UpdateCallCount()).To(Equal(1))
				actualContext, _ := fakeServiceProvider.UpdateArgsForCall(0)
				Expect(actualContext.Value(middlewares.OriginatingIdentityKey)).To(Equal(expectedHeader))

				By("validating SI details is not updated")
				Expect(fakeStorage.StoreServiceInstanceDetailsCallCount()).To(Equal(0))

				By("validating provision parameters storing call")
				Expect(fakeStorage.StoreProvisionRequestDetailsCallCount()).To(Equal(1))
				actualSI, actualParams := fakeStorage.StoreProvisionRequestDetailsArgsForCall(0)
				Expect(actualSI).To(Equal(instanceID))
				Expect(actualParams).To(BeEmpty())
			})
		})

		When("plan change is requested", func() {
			BeforeEach(func() {
				fakeServiceProvider.UpdateReturns(models.ServiceInstanceDetails{
					OperationType: models.UpdateOperationType,
					OperationId:   updateOperationID,
				}, nil)

				updateDetails = domain.UpdateDetails{
					ServiceID: offeringID,
					PlanID:    newPlanID,
					PreviousValues: domain.PreviousValues{
						PlanID:    originalPlanID,
						ServiceID: offeringID,
						OrgID:     orgID,
						SpaceID:   spaceID,
					},
					RawContext: json.RawMessage(fmt.Sprintf(`{"organization_guid":%q, "space_guid": %q}`, orgID, spaceID)),
				}
			})

			It("should do update async and not change planID", func() {
				response, err := serviceBroker.Update(context.TODO(), instanceID, updateDetails, true)
				Expect(err).ToNot(HaveOccurred())

				By("validating response")
				Expect(response.IsAsync).To(BeTrue())
				Expect(response.DashboardURL).To(BeEmpty())
				Expect(response.OperationData).To(Equal(updateOperationID))

				By("validating provider update has been called")
				Expect(fakeServiceProvider.UpdateCallCount()).To(Equal(1))

				By("validating SI details storing call")
				Expect(fakeStorage.StoreServiceInstanceDetailsCallCount()).To(Equal(1))
				actualSIDetails := fakeStorage.StoreServiceInstanceDetailsArgsForCall(0)
				Expect(actualSIDetails.GUID).To(Equal(instanceID))
				Expect(actualSIDetails.ServiceGUID).To(Equal(offeringID))
				Expect(actualSIDetails.PlanGUID).To(Equal(newPlanID))
				Expect(actualSIDetails.SpaceGUID).To(Equal(spaceID))
				Expect(actualSIDetails.OrganizationGUID).To(Equal(orgID))
			})
		})

		When("parameter change is requested", func() {
			BeforeEach(func() {
				fakeServiceProvider.UpdateReturns(models.ServiceInstanceDetails{
					OperationType: models.UpdateOperationType,
					OperationId:   updateOperationID,
				}, nil)

				updateDetails = domain.UpdateDetails{
					ServiceID: offeringID,
					PlanID:    originalPlanID,
					PreviousValues: domain.PreviousValues{
						PlanID:    originalPlanID,
						ServiceID: offeringID,
						OrgID:     orgID,
						SpaceID:   spaceID,
					},
					RawContext:    json.RawMessage(fmt.Sprintf(`{"organization_guid":%q, "space_guid": %q}`, orgID, spaceID)),
					RawParameters: json.RawMessage(`{"foo":"quz","guz":"muz"}`),
				}
			})

			It("should update provision params", func() {
				response, err := serviceBroker.Update(context.TODO(), instanceID, updateDetails, true)
				Expect(err).ToNot(HaveOccurred())

				By("validating response")
				Expect(response.IsAsync).To(BeTrue())
				Expect(response.DashboardURL).To(BeEmpty())
				Expect(response.OperationData).To(Equal(updateOperationID))

				By("validating provider update has been called")
				Expect(fakeServiceProvider.UpdateCallCount()).To(Equal(1))
				_, actualVars := fakeServiceProvider.UpdateArgsForCall(0)
				Expect(actualVars.GetString("foo")).To(Equal("quz"))
				Expect(actualVars.GetString("guz")).To(Equal("muz"))

				By("validating SI details is not updated")
				Expect(fakeStorage.StoreServiceInstanceDetailsCallCount()).To(Equal(0))

				By("validating provision details have been stored")
				Expect(fakeStorage.StoreProvisionRequestDetailsCallCount()).To(Equal(1))
				_, actualRequestVars := fakeStorage.StoreProvisionRequestDetailsArgsForCall(0)
				Expect(actualRequestVars).To(Equal(storage.JSONObject{"foo": "quz", "guz": "muz"}))
			})
		})
	})

	Describe("update variables", func() {
		Describe("passing variables on provision and update", func() {
			BeforeEach(func() {
				fakeStorage.GetProvisionRequestDetailsReturns(map[string]interface{}{"foo": "bar", "baz": "quz"}, nil)
			})

			It("should merge all variables", func() {
				updateDetails = domain.UpdateDetails{
					ServiceID: offeringID,
					PlanID:    originalPlanID,
					PreviousValues: domain.PreviousValues{
						PlanID:    originalPlanID,
						ServiceID: offeringID,
						OrgID:     orgID,
						SpaceID:   spaceID,
					},
					RawContext:    json.RawMessage(fmt.Sprintf(`{"organization_guid":%q, "space_guid": %q}`, orgID, spaceID)),
					RawParameters: json.RawMessage(`{"foo":"quz","guz":"muz"}`),
				}

				_, err := serviceBroker.Update(context.TODO(), instanceID, updateDetails, true)
				Expect(err).ToNot(HaveOccurred())

				By("validating provision variables are retrieved")
				Expect(fakeStorage.GetProvisionRequestDetailsCallCount()).To(Equal(1))

				By("validating provider update has been called")
				Expect(fakeServiceProvider.UpdateCallCount()).To(Equal(1))
				_, actualVars := fakeServiceProvider.UpdateArgsForCall(0)
				Expect(actualVars.GetString("foo")).To(Equal("quz"))
				Expect(actualVars.GetString("guz")).To(Equal("muz"))
				Expect(actualVars.GetString("baz")).To(Equal("quz"))

				By("validating provision details have been stored")
				Expect(fakeStorage.StoreProvisionRequestDetailsCallCount()).To(Equal(1))
				_, actualRequestVars := fakeStorage.StoreProvisionRequestDetailsArgsForCall(0)
				Expect(actualRequestVars).To(Equal(storage.JSONObject{"baz": "quz", "foo": "quz", "guz": "muz"}))
			})
		})

		Describe("passing variables on provision, import and update", func() {
			BeforeEach(func() {
				fakeStorage.GetProvisionRequestDetailsReturns(map[string]interface{}{"foo": "bar", "baz": "quz"}, nil)
				fakeServiceProvider.GetImportedPropertiesReturns(map[string]interface{}{"foo": "quz", "guz": "muz", "laz": "taz"}, nil)
			})

			It("should merge all variables", func() {
				updateDetails = domain.UpdateDetails{
					ServiceID: offeringID,
					PlanID:    originalPlanID,
					PreviousValues: domain.PreviousValues{
						PlanID:    originalPlanID,
						ServiceID: offeringID,
						OrgID:     orgID,
						SpaceID:   spaceID,
					},
					RawContext:    json.RawMessage(fmt.Sprintf(`{"organization_guid":%q, "space_guid": %q}`, orgID, spaceID)),
					RawParameters: json.RawMessage(`{"guz":"duz"}`),
				}

				_, err := serviceBroker.Update(context.TODO(), instanceID, updateDetails, true)
				Expect(err).ToNot(HaveOccurred())

				By("validating provision and import variables are retrieved")
				Expect(fakeStorage.GetProvisionRequestDetailsCallCount()).To(Equal(1))
				Expect(fakeServiceProvider.GetImportedPropertiesCallCount()).To(Equal(1))

				By("validating provider update has been called")
				Expect(fakeServiceProvider.UpdateCallCount()).To(Equal(1))
				_, actualVars := fakeServiceProvider.UpdateArgsForCall(0)
				Expect(actualVars.GetString("foo")).To(Equal("quz"))
				Expect(actualVars.GetString("guz")).To(Equal("duz"))
				Expect(actualVars.GetString("baz")).To(Equal("quz"))
				Expect(actualVars.GetString("laz")).To(Equal("taz"))

				By("validating provision details have been stored")
				Expect(fakeStorage.StoreProvisionRequestDetailsCallCount()).To(Equal(1))
				_, actualRequestVars := fakeStorage.StoreProvisionRequestDetailsArgsForCall(0)
				Expect(actualRequestVars).To(Equal(storage.JSONObject{"baz": "quz", "foo": "quz", "guz": "duz", "laz": "taz"}))
			})
		})

		Describe("using provision computed variables", func() {
			It("passes computed variables to update", func() {
				header := "cloudfoundry eyANCiAgInVzZXJfaWQiOiAiNjgzZWE3NDgtMzA5Mi00ZmY0LWI2NTYtMzljYWNjNGQ1MzYwIg0KfQ=="
				newContext := context.WithValue(context.Background(), middlewares.OriginatingIdentityKey, header)

				_, err := serviceBroker.Update(newContext, instanceID, updateDetails, true)
				Expect(err).ToNot(HaveOccurred())

				By("validating provider provision has been called with the right vars")
				Expect(fakeServiceProvider.UpdateCallCount()).To(Equal(1))
				_, actualVars := fakeServiceProvider.UpdateArgsForCall(0)
				Expect(actualVars.GetString("copyOriginatingIdentity")).To(Equal(`{"platform":"cloudfoundry","value":{"user_id":"683ea748-3092-4ff4-b656-39cacc4d5360"}}`))
				Expect(actualVars.GetString("labels")).To(Equal(`{"pcf-instance-id":"test-instance-id","pcf-organization-guid":"test-org-id","pcf-space-guid":"test-space-id"}`))
			})
		})

		Describe("updating non-updatable parameter", func() {
			It("should error", func() {
				updateDetails = domain.UpdateDetails{
					ServiceID: offeringID,
					PlanID:    originalPlanID,
					PreviousValues: domain.PreviousValues{
						PlanID:    originalPlanID,
						ServiceID: offeringID,
						OrgID:     orgID,
						SpaceID:   spaceID,
					},
					RawParameters: json.RawMessage(`{"prohibit-update-field":"test"}`),
				}

				_, err := serviceBroker.Update(context.TODO(), instanceID, updateDetails, true)
				Expect(err).To(MatchError("attempt to update parameter that may result in service instance re-creation and data loss"))
			})
		})

		Describe("updating parameter that is not defined in the service definition", func() {
			It("should error", func() {
				u := domain.UpdateDetails{
					ServiceID: offeringID,
					PlanID:    originalPlanID,
					PreviousValues: domain.PreviousValues{
						PlanID:    originalPlanID,
						ServiceID: offeringID,
						OrgID:     orgID,
						SpaceID:   spaceID,
					},
					RawParameters: json.RawMessage(`{"invalid_parameter":42,"foo":"bar","other_invalid":false}`),
				}

				_, err1 := serviceBroker.Update(context.TODO(), instanceID, u, true)
				Expect(err1).To(MatchError("additional properties are not allowed: invalid_parameter, other_invalid"))
			})
		})

		Describe("updating parameter that is specified in the plan", func() {
			It("should error", func() {
				updateDetails := domain.UpdateDetails{
					ServiceID: offeringID,
					PlanID:    newPlanID,
					PreviousValues: domain.PreviousValues{
						PlanID:    originalPlanID,
						ServiceID: offeringID,
						OrgID:     orgID,
						SpaceID:   spaceID,
					},
					RawParameters: json.RawMessage(`{"foo":"bar","new-plan-defined-key":42,"new-other-plan-defined-key":"test","other_invalid":false}`),
				}

				_, err2 := serviceBroker.Update(context.TODO(), instanceID, updateDetails, true)
				Expect(err2).To(MatchError("plan defined properties cannot be changed: new-other-plan-defined-key, new-plan-defined-key"))
			})
		})

		When("parameter validation is disabled", func() {
			It("should not error", func() {
				updateDetails = domain.UpdateDetails{
					ServiceID: offeringID,
					PlanID:    originalPlanID,
					PreviousValues: domain.PreviousValues{
						PlanID:    originalPlanID,
						ServiceID: offeringID,
						OrgID:     orgID,
						SpaceID:   spaceID,
					},
					RawParameters: json.RawMessage(`{"invalid_parameter":42,"foo":"bar","other_invalid":false,"plan-defined-key":42}`),
				}
				viper.Set(broker.DisableRequestPropertyValidation, true)

				_, err := serviceBroker.Update(context.TODO(), instanceID, updateDetails, true)
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				viper.Set(broker.DisableRequestPropertyValidation, false)
			})
		})
	})

	When("provider update errors", func() {
		BeforeEach(func() {
			fakeServiceProvider.UpdateReturns(models.ServiceInstanceDetails{}, errors.New("cannot update right now"))
		})

		It("should error and not update the provision variables", func() {
			_, err := serviceBroker.Update(context.TODO(), instanceID, updateDetails, true)
			Expect(err).To(MatchError("cannot update right now"))

			By("validate it does not update the provision requet details")
			Expect(fakeStorage.StoreProvisionRequestDetailsCallCount()).To(Equal(0))
		})
	})

	When("plan does not exists", func() {
		It("should error", func() {
			updateDetails := domain.UpdateDetails{
				PlanID: "non-existent-plan",
			}

			_, err := serviceBroker.Update(context.TODO(), instanceID, updateDetails, true)
			Expect(err).To(MatchError(`plan ID "non-existent-plan" could not be found`))
		})
	})

	When("instance does not exists", func() {
		BeforeEach(func() {
			fakeStorage.ExistsServiceInstanceDetailsReturns(false, nil)
		})

		It("should error", func() {
			_, err := serviceBroker.Update(context.TODO(), instanceID, updateDetails, true)
			Expect(err).To(MatchError("instance does not exist"))
		})
	})

	When("request json is invalid", func() {
		It("should error", func() {
			updateDetails := domain.UpdateDetails{
				PlanID:        originalPlanID,
				RawParameters: json.RawMessage("{invalid-json}"),
			}

			_, err := serviceBroker.Update(context.TODO(), instanceID, updateDetails, true)
			Expect(err).To(MatchError("User supplied parameters must be in the form of a valid JSON map."))
		})
	})

	When("client cannot accept async", func() {
		It("should error", func() {
			_, err := serviceBroker.Update(context.TODO(), instanceID, updateDetails, false)
			Expect(err).To(MatchError("This service plan requires client support for asynchronous service operations."))
		})
	})

	Describe("storage errors", func() {
		Context("storage errors when checking SI details", func() {
			BeforeEach(func() {
				fakeStorage.ExistsServiceInstanceDetailsReturns(false, errors.New("failed to check existence"))
			})

			It("should error", func() {
				_, err := serviceBroker.Update(context.TODO(), instanceID, updateDetails, true)
				Expect(err).To(MatchError("database error checking for existing instance: failed to check existence"))
			})
		})

		Context("storage errors when getting SI details", func() {
			BeforeEach(func() {
				fakeStorage.GetServiceInstanceDetailsReturns(storage.ServiceInstanceDetails{}, errors.New("failed to get SI details"))
			})

			It("should error", func() {
				_, err := serviceBroker.Update(context.TODO(), instanceID, updateDetails, true)
				Expect(err).To(MatchError("database error getting existing instance: failed to get SI details"))
			})
		})

		Context("storage errors when getting provision parameters", func() {
			BeforeEach(func() {
				fakeStorage.GetProvisionRequestDetailsReturns(nil, errors.New("failed to get provision parameters"))
			})

			It("should error", func() {
				_, err := serviceBroker.Update(context.TODO(), instanceID, updateDetails, true)
				Expect(err).To(MatchError(`error retrieving provision request details for "test-instance-id": failed to get provision parameters`))
			})
		})

		Context("storage errors when storing SI details", func() {
			BeforeEach(func() {
				fakeStorage.StoreServiceInstanceDetailsReturns(errors.New("failed to store SI details"))
			})

			It("should error", func() {
				updateDetails.PlanID = newPlanID

				_, err := serviceBroker.Update(context.TODO(), instanceID, updateDetails, true)
				Expect(err).To(MatchError("error saving instance details to database: failed to store SI details. WARNING: this instance cannot be deprovisioned through cf. Contact your operator for cleanup"))
			})
		})

		Context("storage errors when storing provision parameters", func() {
			BeforeEach(func() {
				fakeStorage.StoreProvisionRequestDetailsReturns(errors.New("failed to store provision parameters"))
			})

			It("should error", func() {
				_, err := serviceBroker.Update(context.TODO(), instanceID, updateDetails, true)
				Expect(err).To(MatchError("error saving provision request details to database: failed to store provision parameters. Services relying on async provisioning will not be able to complete provisioning"))
			})
		})
	})
})
