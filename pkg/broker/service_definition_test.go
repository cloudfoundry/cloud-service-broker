package broker_test

import (
	"fmt"
	"os"

	"github.com/cloudfoundry/cloud-service-broker/pkg/broker"
	"github.com/cloudfoundry/cloud-service-broker/pkg/varcontext"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v8/domain"
	"github.com/spf13/viper"
)

var _ = Describe("ServiceDefinition", func() {
	Describe("Validate", func() {
		Context("validate ID", func() {
			It("should fail when ID is missing", func() {
				definition := broker.ServiceDefinition{Name: "test"}

				err := definition.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("field must be a UUID: ID"))
			})

			It("should fail when ID is not valid", func() {
				definition := broker.ServiceDefinition{ID: "test", Name: "test-name"}

				err := definition.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("field must be a UUID: ID"))
			})
		})

		It("should fail when Name is missing", func() {
			definition := broker.ServiceDefinition{ID: "55ad8194-0431-11ec-948a-63ff62e94b14"}

			err := definition.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("field must match '^[a-zA-Z0-9-\\.]+$': Name"))
		})

		It("should fail when ImageURL is not a URL", func() {
			definition := broker.ServiceDefinition{
				ID:       "55ad8194-0431-11ec-948a-63ff62e94b14",
				Name:     "test-offering",
				ImageURL: "some-non-url",
			}

			err := definition.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("field must be a URL: ImageURL"))
		})

		It("should fail when DocumentationURL is not a URL", func() {
			definition := broker.ServiceDefinition{
				ID:               "55ad8194-0431-11ec-948a-63ff62e94b14",
				Name:             "test-offering",
				DocumentationURL: "some-non-url",
			}

			err := definition.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("field must be a URL: DocumentationURL"))
		})

		It("should fail when SupportURL is not a URL", func() {
			definition := broker.ServiceDefinition{
				ID:         "55ad8194-0431-11ec-948a-63ff62e94b14",
				Name:       "test-offering",
				SupportURL: "some-non-url",
			}

			err := definition.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("field must be a URL: SupportURL"))
		})

		It("should fail when ProvisionInputVariables is not a valid", func() {
			definition := broker.ServiceDefinition{
				ID:   "55ad8194-0431-11ec-948a-63ff62e94b14",
				Name: "test-offering",
				ProvisionInputVariables: []broker.BrokerVariable{
					{
						ProhibitUpdate: true,
					},
				},
			}

			err := definition.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("missing field(s): ProvisionInputVariables[0].details, ProvisionInputVariables[0].field_name"))
		})

		It("should fail when ProvisionComputedVariables is not a valid", func() {
			definition := broker.ServiceDefinition{
				ID:   "55ad8194-0431-11ec-948a-63ff62e94b14",
				Name: "test-offering",
				ProvisionComputedVariables: []varcontext.DefaultVariable{
					{
						Overwrite: true,
					},
				},
			}

			err := definition.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("missing field(s): ProvisionComputedVariables[0].default, ProvisionComputedVariables[0].name"))
		})

		It("should fail when BindInputVariables is not a valid", func() {
			definition := broker.ServiceDefinition{
				ID:   "55ad8194-0431-11ec-948a-63ff62e94b14",
				Name: "test-offering",
				BindInputVariables: []broker.BrokerVariable{
					{
						ProhibitUpdate: true,
					},
				},
			}

			err := definition.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("missing field(s): BindInputVariables[0].details, BindInputVariables[0].field_name"))
		})

		It("should fail when PlanVariables is not a valid", func() {
			definition := broker.ServiceDefinition{
				ID:   "55ad8194-0431-11ec-948a-63ff62e94b14",
				Name: "test-offering",
				PlanVariables: []broker.BrokerVariable{
					{
						ProhibitUpdate: true,
					},
				},
			}

			err := definition.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("missing field(s): PlanVariables[0].details, PlanVariables[0].field_name"))
		})

		Context("validate plans", func() {
			It("should fail when plan is missing name", func() {
				definition := broker.ServiceDefinition{
					ID:   "55ad8194-0431-11ec-948a-63ff62e94b14",
					Name: "test-offering",
					Plans: []broker.ServicePlan{
						{
							ServicePlan: domain.ServicePlan{
								ID: "c4edb10c-0434-11ec-8cbb-a777f76cf2ac",
							},
						},
					},
				}

				err := definition.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("missing field(s): Plans[0].Name"))
			})

			It("should fail when plan is missing id", func() {
				definition := broker.ServiceDefinition{
					ID:   "55ad8194-0431-11ec-948a-63ff62e94b14",
					Name: "test-offering",
					Plans: []broker.ServicePlan{
						{
							ServicePlan: domain.ServicePlan{
								Name: "test-plan",
							},
						},
					},
				}

				err := definition.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("field must be a UUID: Plans[0].ID"))
			})

			Context("plan duplication", func() {
				It("should fail when plan id is duplicated across the offering", func() {
					definition := broker.ServiceDefinition{
						ID:   "55ad8194-0431-11ec-948a-63ff62e94b14",
						Name: "test-offering",
						Plans: []broker.ServicePlan{
							{
								ServicePlan: domain.ServicePlan{
									Name: "test-plan-1",
									ID:   "c4edb10c-0434-11ec-8cbb-a777f76cf2ac",
								},
							},
							{
								ServicePlan: domain.ServicePlan{
									Name: "test-plan-2",
									ID:   "c4edb10c-0434-11ec-8cbb-a777f76cf2ac",
								},
							},
						},
					}

					err := definition.Validate()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("duplicated value, must be unique: c4edb10c-0434-11ec-8cbb-a777f76cf2ac: Plans[1].Id"))
				})

				It("should fail when plan name is duplicated across the offering", func() {
					definition := broker.ServiceDefinition{
						ID:   "55ad8194-0431-11ec-948a-63ff62e94b14",
						Name: "test-offering",
						Plans: []broker.ServicePlan{
							{
								ServicePlan: domain.ServicePlan{
									Name: "test-plan",
									ID:   "c4edb10c-0434-11ec-8cbb-a777f76cf2ac",
								},
							},
							{
								ServicePlan: domain.ServicePlan{
									Name: "test-plan",
									ID:   "f71ef2c0-0435-11ec-a1e1-17b59e2ff283",
								},
							},
						},
					}

					err := definition.Validate()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("duplicated value, must be unique: test-plan: Plans[1].Name"))
				})
			})
		})

	})

	Describe("AllowedUpdate", func() {
		serviceDefinition := broker.ServiceDefinition{
			ProvisionInputVariables: []broker.BrokerVariable{
				{
					FieldName:      "prohibited",
					ProhibitUpdate: true,
				},
				{
					FieldName: "allowed",
				},
			},
		}

		DescribeTable("returns the correct result",
			func(params map[string]interface{}, expected bool) {
				actual, err := serviceDefinition.AllowedUpdate(params)
				Expect(err).ToNot(HaveOccurred())
				Expect(actual).To(Equal(expected))
			},
			Entry("allowed", map[string]interface{}{"allowed": "some_val"}, true),
			Entry("prohibited", map[string]interface{}{"prohibited": "some_val"}, false),
			Entry("empty", nil, true),
		)
	})

	Describe("UserDefinedPlans", func() {

		const (
			fakePlanName        = "fakePlanName"
			fakePlanID          = "fakePlanID"
			fakePlanDescription = "fakePlanDescription"
			fakePlanProperty    = "fakePlanProperty"
			fakePlanGuid        = "6938cf33-0a12-4308-af4c-32134c645e87"
			defaultTFVersion    = "1.0.0"
		)

		var (
			fakeServicePlanConfig    string
			fakeServicePlanTile      string
			fakeServicePlanMissingID string
			service                  broker.ServiceDefinition
		)

		BeforeEach(func() {
			service = broker.ServiceDefinition{
				Name: "fake-service",
			}

			fakeServicePlanConfig = fmt.Sprintf(`[{"name":"%s","id":"%s","description":"%s", "additional_property":"%s"}]`,
				fakePlanName, fakePlanID, fakePlanDescription, fakePlanProperty)
			fakeServicePlanTile = fmt.Sprintf(`{"%[1]s":{"name":"%[1]s","description":"%s","additional_property":"%s", "guid":"%s"}}`,
				fakePlanName, fakePlanDescription, fakePlanProperty, fakePlanGuid)
		})

		AfterEach(func() {
			viper.Reset()
		})

		When("plans are set in viper configuration", func() {
			BeforeEach(func() {
				viper.Set("service.fake-service.plans", fakeServicePlanConfig)
			})

			It("should return the service plan", func() {
				actualPlans, err := service.UserDefinedPlans(defaultTFVersion)

				Expect(err).NotTo(HaveOccurred())
				Expect(actualPlans).To(HaveLen(1))
				Expect(actualPlans[0].Name).To(Equal(fakePlanName))
				Expect(actualPlans[0].ID).To(Equal(fakePlanID))
				Expect(actualPlans[0].Description).To(Equal(fakePlanDescription))
				Expect(actualPlans[0].ServiceProperties).To(Equal(map[string]interface{}{"additional_property": fakePlanProperty}))
			})

			When("an invalid plan is provided in configuration", func() {
				BeforeEach(func() {
					fakeServicePlanConfig = `{invalid-json`
					viper.Set("service.fake-service.plans", fakeServicePlanConfig)
				})

				It("should return an error", func() {
					_, err := service.UserDefinedPlans(defaultTFVersion)
					Expect(err).To(MatchError("invalid character 'i' looking for beginning of object key string"))

				})
			})

		})

		When("plans are set as an environment variable", func() {
			BeforeEach(func() {
				os.Setenv(service.TileUserDefinedPlansVariable(), fakeServicePlanTile)
			})

			AfterEach(func() {
				os.Unsetenv(service.TileUserDefinedPlansVariable())
			})

			It("should return the service plan", func() {
				actualPlans, err := service.UserDefinedPlans(defaultTFVersion)

				Expect(err).NotTo(HaveOccurred())
				Expect(actualPlans).To(HaveLen(1))
				Expect(actualPlans[0].Name).To(Equal(fakePlanName))
				Expect(actualPlans[0].ID).To(Equal(fakePlanGuid))
				Expect(actualPlans[0].Description).To(Equal(fakePlanDescription))
				Expect(actualPlans[0].ServiceProperties).To(Equal(map[string]interface{}{"additional_property": fakePlanProperty, "guid": fakePlanGuid}))
			})

			When("an invalid plan is provided as an environment variable", func() {
				BeforeEach(func() {
					fakeServicePlanTile = `{invalid-json`
					os.Setenv(service.TileUserDefinedPlansVariable(), fakeServicePlanTile)
				})

				It("should return an error", func() {
					_, err := service.UserDefinedPlans(defaultTFVersion)
					Expect(err).To(MatchError("invalid character 'i' looking for beginning of object key string"))

				})
			})

		})

		When("plan validation fails", func() {
			BeforeEach(func() {
				fakeServicePlanMissingID = fmt.Sprintf(`[{"name":"%s","description":"%s", "additional_property":"%s"}]`,
					fakePlanName, fakePlanDescription, fakePlanProperty)
				viper.Set("service.fake-service.plans", fakeServicePlanMissingID)
			})

			It("returns an error", func() {
				_, err := service.UserDefinedPlans(defaultTFVersion)
				Expect(err).To(MatchError("fake-service custom plan {ServicePlan:{ID: Name:fakePlanName Description:fakePlanDescription Free:<nil> Bindable:<nil> Metadata:<nil> Schemas:<nil> PlanUpdatable:<nil> MaximumPollingDuration:<nil> MaintenanceInfo:<nil>} ServiceProperties:map[additional_property:fakePlanProperty] ProvisionOverrides:map[] BindOverrides:map[]} is missing an id"))
			})
		})

		When("TFUpgrades are enabled", func() {
			BeforeEach(func() {
				viper.Set("service.fake-service.plans", fakeServicePlanConfig)
				viper.Set("brokerpak.terraform.upgrades.enabled", true)
			})

			It("returns a broker service plan with maintenance info version matching default TF version", func() {
				actualPlans, err := service.UserDefinedPlans(defaultTFVersion)

				Expect(err).NotTo(HaveOccurred())
				Expect(actualPlans).To(HaveLen(1))
				Expect(actualPlans[0].MaintenanceInfo).To(Equal(&domain.MaintenanceInfo{Version: defaultTFVersion, Description: "This upgrade provides support for Terraform version: 1.0.0. The upgrade operation will take a while. The instance and all associated bindings will be upgraded."}))
			})

		})

		When("TFUpgrades are disabled", func() {
			BeforeEach(func() {
				viper.Set("service.fake-service.plans", fakeServicePlanConfig)
			})

			It("returns a broker service plan with nil maintenance info", func() {
				actualPlans, err := service.UserDefinedPlans(defaultTFVersion)

				Expect(err).NotTo(HaveOccurred())
				Expect(actualPlans).To(HaveLen(1))
				Expect(actualPlans[0].MaintenanceInfo).To(BeNil())
			})
		})

	})
})
