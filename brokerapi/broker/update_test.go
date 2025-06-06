package broker_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"code.cloudfoundry.org/brokerapi/v13/domain"
	"code.cloudfoundry.org/brokerapi/v13/middlewares"
	"code.cloudfoundry.org/lager/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	"golang.org/x/net/context"

	"github.com/cloudfoundry/cloud-service-broker/v2/brokerapi/broker"
	"github.com/cloudfoundry/cloud-service-broker/v2/brokerapi/broker/brokerfakes"
	"github.com/cloudfoundry/cloud-service-broker/v2/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/storage"
	pkgBroker "github.com/cloudfoundry/cloud-service-broker/v2/pkg/broker"
	pkgBrokerFakes "github.com/cloudfoundry/cloud-service-broker/v2/pkg/broker/brokerfakes"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/featureflags"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/v2/utils"
)

var _ = Describe("Update", func() {
	const (
		spaceID           = "test-space-id"
		orgID             = "test-org-id"
		originalPlanID    = "test-plan-id"
		offeringID        = "test-service-id"
		newPlanID         = "new-test-plan-id"
		instanceID        = "test-instance-id"
		updateOperationID = "tf:test-instance-id:"
	)

	var (
		serviceBroker *broker.ServiceBroker
		updateDetails domain.UpdateDetails

		fakeStorage         *brokerfakes.FakeStorage
		fakeServiceProvider *pkgBrokerFakes.FakeServiceProvider
	)

	BeforeEach(func() {
		fakeServiceProvider = &pkgBrokerFakes.FakeServiceProvider{}

		providerBuilder := func(logger lager.Logger, store pkgBroker.ServiceProviderStorage) pkgBroker.ServiceProvider {
			return fakeServiceProvider
		}
		brokerConfig := &broker.BrokerConfig{
			Registry: pkgBroker.BrokerRegistry{
				"test-service": &pkgBroker.ServiceDefinition{
					GlobalLabels: map[string]string{"key1": "value1", "key2": "value2"},
					ID:           offeringID,
					Name:         "test-service",
					Plans: []pkgBroker.ServicePlan{
						{
							ServicePlan: domain.ServicePlan{
								ID:   originalPlanID,
								Name: "test-plan",
								MaintenanceInfo: &domain.MaintenanceInfo{
									Version: "2.0.0",
								},
							},
							ServiceProperties: map[string]any{
								"plan-defined-key":       "plan-defined-value",
								"other-plan-defined-key": "other-plan-defined-value",
							},
						},
						{
							ServicePlan: domain.ServicePlan{
								ID:   newPlanID,
								Name: "new-test-plan",
								MaintenanceInfo: &domain.MaintenanceInfo{
									Version: "2.0.0",
								},
							},
							ServiceProperties: map[string]any{
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
					BindComputedVariables: []varcontext.DefaultVariable{
						{Name: "instance_output", Default: `${instance.details["instance-provision-output"]}`, Overwrite: true},
					},
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
		}, nil)
		fakeStorage.GetTerraformDeploymentReturns(storage.TerraformDeployment{
			ID:                 updateOperationID,
			LastOperationType:  models.ProvisionOperationType,
			LastOperationState: "GOOG",
		}, nil)

		serviceBroker = must(broker.New(brokerConfig, fakeStorage, utils.NewLogger("brokers-test")))

		updateDetails = domain.UpdateDetails{
			ServiceID: offeringID,
			PlanID:    originalPlanID,
			MaintenanceInfo: &domain.MaintenanceInfo{
				Version: "2.0.0",
			},
			PreviousValues: domain.PreviousValues{
				PlanID:    originalPlanID,
				ServiceID: offeringID,
				OrgID:     orgID,
				SpaceID:   spaceID,
				MaintenanceInfo: &domain.MaintenanceInfo{
					Version: "2.0.0",
				},
			},
			RawContext: json.RawMessage(fmt.Sprintf(`{"organization_guid":%q, "space_guid": %q}`, orgID, spaceID)),
		}
	})

	Describe("update", func() {
		When("no plan or parameter changes are requested", func() {
			BeforeEach(func() {
				fakeServiceProvider.UpdateReturns(nil)
				fakeServiceProvider.PollInstanceReturns(true, "a message", models.UpdateOperationType, nil)
			})

			It("should complete changing the instance operation type", func() {
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

				By("validating SI operation type is updated")
				expectOperationTypeToBeUpdated(
					fakeStorage,
					updateOperationID,
					instanceID,
					offeringID,
					originalPlanID,
					spaceID,
					orgID,
					storage.JSONObject{},
				)
			})
		})

		When("plan change is requested", func() {
			BeforeEach(func() {
				fakeServiceProvider.UpdateReturns(nil)
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
				expectOperationTypeToBeUpdated(
					fakeStorage,
					updateOperationID,
					instanceID,
					offeringID,
					originalPlanID,
					spaceID,
					orgID,
					storage.JSONObject{},
				)
			})
		})

		When("parameter change is requested", func() {
			BeforeEach(func() {
				fakeServiceProvider.UpdateReturns(nil)

				updateDetails = domain.UpdateDetails{
					ServiceID: offeringID,
					PlanID:    originalPlanID,
					MaintenanceInfo: &domain.MaintenanceInfo{
						Version: "2.0.0",
					},
					PreviousValues: domain.PreviousValues{
						PlanID:    originalPlanID,
						ServiceID: offeringID,
						OrgID:     orgID,
						SpaceID:   spaceID,
						MaintenanceInfo: &domain.MaintenanceInfo{
							Version: "2.0.0",
						},
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

				By("validating SI details is updated")
				expectOperationTypeToBeUpdated(
					fakeStorage,
					updateOperationID,
					instanceID,
					offeringID,
					originalPlanID,
					spaceID,
					orgID,
					storage.JSONObject{"foo": "quz", "guz": "muz"},
				)
			})
		})

		When("provider update errors", func() {
			BeforeEach(func() {
				fakeServiceProvider.UpdateReturns(errors.New("cannot update right now"))
			})

			It("should error and not update the provision variables", func() {
				_, err := serviceBroker.Update(context.TODO(), instanceID, updateDetails, true)
				Expect(err).To(MatchError("cannot update right now"))

				By("validate it does not update the instance details")
				Expect(fakeStorage.StoreServiceInstanceDetailsCallCount()).To(Equal(0))

				By("validate it does not update the provision request details")
				Expect(fakeStorage.StoreProvisionRequestDetailsCallCount()).To(Equal(0))
			})
		})

		When("an upgrade should have happened", func() {
			It("fails", func() {
				fakeServiceProvider.CheckUpgradeAvailableReturns(errors.New("cannot use this tf version"))

				_, err := serviceBroker.Update(context.TODO(), instanceID, updateDetails, true)
				Expect(err).To(MatchError("tofu version check failed: cannot use this tf version"))

				By("validate it does not update")
				Expect(fakeServiceProvider.UpdateCallCount()).To(Equal(0))
			})
		})
	})

	Describe("upgrade", func() {

		var upgradeDetails domain.UpdateDetails

		BeforeEach(func() {
			upgradeDetails = domain.UpdateDetails{
				ServiceID: offeringID,
				PlanID:    originalPlanID,
				PreviousValues: domain.PreviousValues{
					PlanID:          originalPlanID,
					ServiceID:       offeringID,
					OrgID:           orgID,
					SpaceID:         spaceID,
					MaintenanceInfo: nil,
				},
				RawContext: json.RawMessage(fmt.Sprintf(`{"organization_guid":%q, "space_guid": %q}`, orgID, spaceID)),
				MaintenanceInfo: &domain.MaintenanceInfo{
					Version: "2.0.0",
				},
			}

			fakeServiceProvider.UpgradeInstanceReturns(&sync.WaitGroup{}, nil)
			fakeServiceProvider.UpgradeBindingsReturns(nil)
		})

		It("should trigger an upgrade", func() {
			expectedHeader := "cloudfoundry eyANCiAgInVzZXJfaWQiOiAiNjgzZWE3NDgtMzA5Mi00ZmY0LWI2NTYtMzljYWNjNGQ1MzYwIg0KfQ=="
			newContext := context.WithValue(context.Background(), middlewares.OriginatingIdentityKey, expectedHeader)

			response, err := serviceBroker.Update(newContext, instanceID, upgradeDetails, true)
			Expect(err).ToNot(HaveOccurred())

			By("validating response")
			Expect(response.IsAsync).To(BeTrue())
			Expect(response.DashboardURL).To(BeEmpty())
			Expect(response.OperationData).To(Equal("tf:test-instance-id:"))

			By("validating provider instance and binding upgrade has been called")
			Expect(fakeServiceProvider.UpdateCallCount()).To(Equal(0))
			Expect(fakeServiceProvider.UpgradeInstanceCallCount()).To(Equal(1))
			actualContext, _ := fakeServiceProvider.UpgradeInstanceArgsForCall(0)
			Expect(actualContext.Value(middlewares.OriginatingIdentityKey)).To(Equal(expectedHeader))
			Expect(fakeServiceProvider.UpgradeBindingsCallCount()).To(Equal(0))

			By("validating SI details and provision parameters are not updated")
			Expect(fakeStorage.StoreServiceInstanceDetailsCallCount()).To(Equal(0))
			Expect(fakeStorage.StoreProvisionRequestDetailsCallCount()).To(Equal(0))
		})

		When("provider upgrade errors", func() {
			BeforeEach(func() {
				fakeServiceProvider.UpgradeInstanceReturns(&sync.WaitGroup{}, errors.New("cannot upgrade right now"))
			})

			It("should error", func() {
				_, err := serviceBroker.Update(context.TODO(), instanceID, upgradeDetails, true)
				Expect(err).To(MatchError("cannot upgrade right now"))
			})
		})

		When("the instance has bindings", func() {
			BeforeEach(func() {
				fakeStorage.GetServiceBindingIDsForServiceInstanceReturns([]string{"firstBindingID", "secondBindingID"}, nil)

				fakeStorage.GetServiceInstanceDetailsReturns(storage.ServiceInstanceDetails{
					GUID:        instanceID,
					Outputs:     storage.JSONObject{"instance-provision-output": "admin-user-name"},
					ServiceGUID: offeringID,
					PlanGUID:    originalPlanID,
				}, nil)

				fakeServiceProvider.GetTerraformOutputsReturns(storage.JSONObject{"instance-provision-output": "admin-user-name"}, nil)

				fakeStorage.GetBindRequestDetailsReturnsOnCall(0, storage.JSONObject{"first-binding-param": "first-binding-bar"}, nil)
				fakeStorage.GetBindRequestDetailsReturnsOnCall(1, storage.JSONObject{"second-binding-param": "second-binding-bar"}, nil)
				fakeStorage.GetTerraformDeploymentReturns(storage.TerraformDeployment{LastOperationState: tf.InProgress}, nil)
			})

			It("should populate the binding contexts with binding computed output and previous bind properties", func() {
				_, err := serviceBroker.Update(context.TODO(), instanceID, upgradeDetails, true)
				Expect(err).ToNot(HaveOccurred())

				By("validating provision variables are retrieved")
				Expect(fakeStorage.GetServiceInstanceDetailsCallCount()).To(Equal(1))

				By("validating provider update has been called")
				Eventually(func(g Gomega) {
					g.Expect(fakeServiceProvider.UpgradeBindingsCallCount()).To(Equal(1))
				}).WithTimeout(time.Second).WithPolling(time.Millisecond).Should(Succeed())

				_, _, bindingVars := fakeServiceProvider.UpgradeBindingsArgsForCall(0)
				Expect(bindingVars[0].GetString("instance_output")).To(Equal("admin-user-name"))
				Expect(bindingVars[0].GetString("first-binding-param")).To(Equal("first-binding-bar"))
				Expect(bindingVars[1].GetString("instance_output")).To(Equal("admin-user-name"))
				Expect(bindingVars[1].GetString("second-binding-param")).To(Equal("second-binding-bar"))
			})

			When("getting binding credentials fails", func() {
				BeforeEach(func() {
					fakeStorage.GetServiceBindingIDsForServiceInstanceReturns([]string{}, errors.New("cant get bindings"))
				})

				It("should error", func() {
					_, err := serviceBroker.Update(context.TODO(), instanceID, upgradeDetails, true)
					Expect(err).NotTo(HaveOccurred())

					Eventually(func(g Gomega) {
						g.Expect(fakeStorage.StoreTerraformDeploymentCallCount()).Should(Equal(1))
						actualDeployment := fakeStorage.StoreTerraformDeploymentArgsForCall(0)
						g.Expect(actualDeployment.LastOperationMessage).To(ContainSubstring(`error retrieving binding for instance "test-instance-id": cant get bindings`))
					}).WithTimeout(time.Second).WithPolling(time.Millisecond).Should(Succeed())
				})
			})

			When("getting binding request details fails", func() {
				BeforeEach(func() {
					fakeStorage.GetBindRequestDetailsReturnsOnCall(1, storage.JSONObject{}, errors.New("cant get binding request details"))
				})

				It("should error", func() {
					_, err := serviceBroker.Update(context.TODO(), instanceID, upgradeDetails, true)
					Expect(err).NotTo(HaveOccurred())

					Eventually(func(g Gomega) {
						g.Expect(fakeStorage.StoreTerraformDeploymentCallCount()).Should(Equal(1))
						actualDeployment := fakeStorage.StoreTerraformDeploymentArgsForCall(0)
						g.Expect(actualDeployment.LastOperationMessage).To(ContainSubstring(`error retrieving bind request details for instance "test-instance-id": cant get binding request details`))
					}).WithTimeout(time.Second).WithPolling(time.Millisecond).Should(Succeed())
				})
			})
		})

		Context("instance context variables", func() {
			Describe("variables of previous provision or updates", func() {
				It("should populate the instance context with variables of previous provision or updates", func() {
					fakeStorage.GetProvisionRequestDetailsReturns(map[string]any{"foo": "bar", "baz": "quz"}, nil)

					_, err := serviceBroker.Update(context.TODO(), instanceID, upgradeDetails, true)
					Expect(err).ToNot(HaveOccurred())

					By("validating provider update has been called")
					Expect(fakeServiceProvider.UpgradeInstanceCallCount()).To(Equal(1))
					_, actualInstanceVars := fakeServiceProvider.UpgradeInstanceArgsForCall(0)
					Expect(actualInstanceVars.GetString("foo")).To(Equal("bar"))
					Expect(actualInstanceVars.GetString("baz")).To(Equal("quz"))
				})
			})
		})
	})

	Describe("instance context variables", func() {
		Describe("passing variables on provision and update", func() {
			BeforeEach(func() {
				fakeStorage.GetProvisionRequestDetailsReturns(map[string]any{"foo": "bar", "baz": "quz"}, nil)
			})

			It("should merge all variables", func() {
				updateDetails = domain.UpdateDetails{
					ServiceID: offeringID,
					PlanID:    originalPlanID,
					MaintenanceInfo: &domain.MaintenanceInfo{
						Version: "2.0.0",
					},
					PreviousValues: domain.PreviousValues{
						PlanID:    originalPlanID,
						ServiceID: offeringID,
						OrgID:     orgID,
						SpaceID:   spaceID,
						MaintenanceInfo: &domain.MaintenanceInfo{
							Version: "2.0.0",
						},
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
				fakeStorage.GetProvisionRequestDetailsReturns(map[string]any{"foo": "bar", "baz": "quz"}, nil)
				fakeServiceProvider.GetImportedPropertiesReturns(map[string]any{"foo": "quz", "guz": "muz", "laz": "taz"}, nil)
			})

			It("should merge all variables", func() {
				updateDetails = domain.UpdateDetails{
					ServiceID: offeringID,
					PlanID:    originalPlanID,
					MaintenanceInfo: &domain.MaintenanceInfo{
						Version: "2.0.0",
					},
					PreviousValues: domain.PreviousValues{
						PlanID:    originalPlanID,
						ServiceID: offeringID,
						OrgID:     orgID,
						SpaceID:   spaceID,
						MaintenanceInfo: &domain.MaintenanceInfo{
							Version: "2.0.0",
						},
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
				expectedLabels := `{"key1":"value1","key2":"value2","pcf-instance-id":"test-instance-id","pcf-organization-guid":"test-org-id","pcf-space-guid":"test-space-id"}`
				Expect(actualVars.GetString("labels")).To(Equal(expectedLabels))
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
					MaintenanceInfo: &domain.MaintenanceInfo{
						Version: "2.0.0",
					},
					PreviousValues: domain.PreviousValues{
						PlanID:    originalPlanID,
						ServiceID: offeringID,
						OrgID:     orgID,
						SpaceID:   spaceID,
						MaintenanceInfo: &domain.MaintenanceInfo{
							Version: "2.0.0",
						},
					},
					RawParameters: json.RawMessage(`{"invalid_parameter":42,"foo":"bar","other_invalid":false,"plan-defined-key":42}`),
				}
				viper.Set(string(featureflags.DisableRequestPropertyValidation), true)
				defer viper.Reset()

				_, err := serviceBroker.Update(context.TODO(), instanceID, updateDetails, true)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	When("specified maintenance info is invalid", func() {
		It("should error", func() {
			updateDetails = domain.UpdateDetails{
				ServiceID: offeringID,
				PlanID:    originalPlanID,
				MaintenanceInfo: &domain.MaintenanceInfo{
					Version: "1.9.0",
				},
				PreviousValues: domain.PreviousValues{
					PlanID:    originalPlanID,
					ServiceID: offeringID,
					OrgID:     orgID,
					SpaceID:   spaceID,
					MaintenanceInfo: &domain.MaintenanceInfo{
						Version: "1.0.0",
					},
				},
				RawContext: json.RawMessage(fmt.Sprintf(`{"organization_guid":%q, "space_guid": %q}`, orgID, spaceID)),
			}

			_, err := serviceBroker.Update(context.TODO(), instanceID, updateDetails, true)

			Expect(err).To(MatchError("error deciding update path: passed maintenance_info does not match the catalog maintenance_info"))
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

func expectOperationTypeToBeUpdated(
	fakeStorage *brokerfakes.FakeStorage,
	operationGUID,
	instanceID,
	offeringID,
	newPlanID,
	spaceID,
	orgID string,
	mergedDetails storage.JSONObject,
) {
	Expect(fakeStorage.StoreServiceInstanceDetailsCallCount()).To(Equal(1))
	actualServiceInstanceDetails := fakeStorage.StoreServiceInstanceDetailsArgsForCall(0)
	Expect(actualServiceInstanceDetails.GUID).To(Equal(instanceID))
	Expect(actualServiceInstanceDetails.ServiceGUID).To(Equal(offeringID))
	Expect(actualServiceInstanceDetails.PlanGUID).To(Equal(newPlanID))
	Expect(actualServiceInstanceDetails.SpaceGUID).To(Equal(spaceID))
	Expect(actualServiceInstanceDetails.OrganizationGUID).To(Equal(orgID))
	Expect(fakeStorage.StoreProvisionRequestDetailsCallCount()).To(Equal(1))
	actualSI, actualParams := fakeStorage.StoreProvisionRequestDetailsArgsForCall(0)
	Expect(actualSI).To(Equal(instanceID))
	Expect(actualParams).To(Equal(mergedDetails))
}
