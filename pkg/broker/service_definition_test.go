package broker_test

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/broker"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/varcontext"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v11/domain"
	"github.com/spf13/viper"
)

var _ = Describe("ServiceDefinition", func() {

	Describe("CatalogEntry", func() {

		var serviceDefinition broker.ServiceDefinition

		BeforeEach(func() {
			serviceDefinition = broker.ServiceDefinition{
				ID:                  "fa6334bc-5314-4b63-8a74-c0e4b638c950",
				Name:                "test-def",
				Description:         "test-def-desc",
				DisplayName:         "test-def-display-name",
				ImageURL:            "image-url",
				DocumentationURL:    "docs-url",
				ProviderDisplayName: "provider-display-name",
				SupportURL:          "support-url",
				Tags:                []string{"Beta", "Tag"},
				PlanUpdateable:      true,
				Plans:               nil,
			}
		})
		It("includes all metadata in the catalog", func() {
			catalogEntry := serviceDefinition.CatalogEntry()
			Expect(catalogEntry.Name).To(Equal("test-def"))
			Expect(catalogEntry.Description).To(Equal("test-def-desc"))
			Expect(catalogEntry.Metadata.DisplayName).To(Equal("test-def-display-name"))
			Expect(catalogEntry.Metadata.ImageUrl).To(Equal("image-url"))
			Expect(catalogEntry.Metadata.LongDescription).To(Equal("test-def-desc"))
			Expect(catalogEntry.Metadata.DocumentationUrl).To(Equal("docs-url"))
			Expect(catalogEntry.Metadata.ProviderDisplayName).To(Equal("provider-display-name"))
			Expect(catalogEntry.Metadata.SupportUrl).To(Equal("support-url"))
			Expect(catalogEntry.Tags).To(ConsistOf("Beta", "Tag"))
		})

		It("includes instances_retrievable: true", func() {
			catalogEntry := serviceDefinition.CatalogEntry()
			Expect(catalogEntry.InstancesRetrievable).To(BeTrue())
		})

		When("service offering is not bindable", func() {
			BeforeEach(func() {
				serviceDefinition.Bindable = false
			})
			It("includes bindable: false", func() {
				catalogEntry := serviceDefinition.CatalogEntry()
				Expect(catalogEntry.Bindable).To(BeFalse())
			})
			It("includes binding_retrievable: false", func() {
				catalogEntry := serviceDefinition.CatalogEntry()
				Expect(catalogEntry.BindingsRetrievable).To(BeFalse())
			})
		})
		When("service offering is bindable", func() {
			BeforeEach(func() {
				serviceDefinition.Bindable = true
			})
			It("includes bindable: true", func() {
				catalogEntry := serviceDefinition.CatalogEntry()
				Expect(catalogEntry.Bindable).To(BeTrue())
			})
			It("includes binding_retrievable: true", func() {
				catalogEntry := serviceDefinition.CatalogEntry()
				Expect(catalogEntry.BindingsRetrievable).To(BeTrue())
			})
		})

	})

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
			func(params map[string]any, expected bool) {
				actual := serviceDefinition.AllowedUpdate(params)
				Expect(actual).To(Equal(expected))
			},
			Entry("allowed", map[string]any{"allowed": "some_val"}, true),
			Entry("prohibited", map[string]any{"prohibited": "some_val"}, false),
			Entry("empty", nil, true),
		)
	})

	Describe("UserDefinedPlans", func() {

		const (
			fakePlanName        = "fakePlanName"
			fakePlanID          = "fakePlanID"
			fakePlanDescription = "fakePlanDescription"
			fakePlanProperty    = "fakePlanProperty"
			fakePlanGUID        = "6938cf33-0a12-4308-af4c-32134c645e87"
			defaultTFVersion    = "1.0.0"
		)

		var (
			fakeServicePlanConfig    string
			fakeServicePlanTile      string
			fakeServicePlanMissingID string
			service                  broker.ServiceDefinition
			maintenanceInfo          *domain.MaintenanceInfo
		)

		BeforeEach(func() {
			service = broker.ServiceDefinition{
				Name: "fake-service",
			}

			fakeServicePlanConfig = fmt.Sprintf(`[{"name":"%[1]s","id":"%[2]s","description":"%[3]s", "additional_property":"%[4]s"},{"name":"second-%[1]s","id":"second-%[2]s","description":"second-%[3]s", "additional_property":"second-%[4]s"} ]`,
				fakePlanName, fakePlanID, fakePlanDescription, fakePlanProperty)
			fakeServicePlanTile = fmt.Sprintf(`{"%[1]s":{"name":"%[1]s","description":"%[2]s","additional_property":"%[3]s", "guid":"%[4]s"},"second-%[1]s":{"name":"second-%[1]s","description":"second-%[2]s","additional_property":"second-%[3]s", "guid":"second-%[4]s"}}`,
				fakePlanName, fakePlanDescription, fakePlanProperty, fakePlanGUID)
		})

		AfterEach(func() {
			viper.Reset()
		})

		DescribeTable("invalid plans", func(plan any, errorMessage string) {
			service = broker.ServiceDefinition{
				Name: "fake-service",
				PlanVariables: []broker.BrokerVariable{
					{
						Required:  true,
						FieldName: "required-field",
						Type:      broker.JSONTypeString,
					},
				},
			}

			viper.Set(service.UserDefinedPlansProperty(), plan)
			defer viper.Reset()

			_, err := service.UserDefinedPlans(nil)
			Expect(err.Error()).To(ContainSubstring(errorMessage))
		},
			Entry("not valid json", `{invalid-json`, "invalid character 'i' looking for beginning of object key string"),
			Entry("missing name", `[{"id":"aaa","required-field":"present"}]`, "is missing a name"),
			Entry("missing id", `[{"name":"aaa","required-field":"present"}]`, "is missing an id"),
			Entry("missing required field", `[{"name":"aaa","id":"aaa"}]`, "is missing required property required-field"),
		)

		When("plans are set in viper configuration", func() {
			BeforeEach(func() {
				viper.Set("service.fake-service.plans", fakeServicePlanConfig)
			})

			It("should return the service plan", func() {
				actualPlans, err := service.UserDefinedPlans(maintenanceInfo)

				Expect(err).NotTo(HaveOccurred())
				Expect(actualPlans).To(HaveLen(2))
				Expect(actualPlans[0].Name).To(Equal(fakePlanName))
				Expect(actualPlans[0].ID).To(Equal(fakePlanID))
				Expect(actualPlans[0].Description).To(Equal(fakePlanDescription))
				Expect(actualPlans[0].ServiceProperties).To(Equal(map[string]any{"additional_property": fakePlanProperty}))
			})

			When("an invalid plan is provided in configuration", func() {
				BeforeEach(func() {
					fakeServicePlanConfig = `{invalid-json`
					viper.Set("service.fake-service.plans", fakeServicePlanConfig)
				})

				It("should return an error", func() {
					_, err := service.UserDefinedPlans(maintenanceInfo)
					Expect(err).To(MatchError(ContainSubstring("invalid character 'i' looking for beginning of object key string")))

				})
			})
			When("a plan is provided in configuration as an object", func() {
				BeforeEach(func() {
					// fakePlanName, fakePlanID, fakePlanDescription, fakePlanProperty
					fakeServicePlanConfigObject := []map[string]interface{}{
						{
							"name":                fakePlanName,
							"id":                  fakePlanID,
							"description":         fakePlanDescription,
							"additional_property": fakePlanProperty,
						},
						{
							"name":                fmt.Sprintf("second-%s", fakePlanName),
							"id":                  fmt.Sprintf("second-%s", fakePlanID),
							"description":         fmt.Sprintf("second-%s", fakePlanDescription),
							"additional_property": fmt.Sprintf("second-%s", fakePlanProperty),
						},
					}
					viper.Set("service.fake-service.provision.defaults", map[string]interface{}{
						"test": "value",
					})
					viper.Set("service.fake-service.plans", fakeServicePlanConfigObject)
				})
				It("should work", func() {
					plan, err := service.UserDefinedPlans(maintenanceInfo)
					Expect(err).To(Not(HaveOccurred()))
					Expect(plan).To(Not(HaveLen(0)))
					provisionOverrides, err := service.ProvisionDefaultOverrides()
					Expect(err).To(Not(HaveOccurred()))
					Expect(provisionOverrides).To(Equal(map[string]interface{}{"test": `value`}))
				})
			})

			When("a plan with provision defaults is provided in configuration as a string", func() {
				BeforeEach(func() {
					fakeServicePlanConfigObject := []map[string]interface{}{
						{
							"name":                fakePlanName,
							"id":                  fakePlanID,
							"description":         fakePlanDescription,
							"additional_property": fakePlanProperty,
						},
						{
							"name":                fmt.Sprintf("fake-string-format-%s", fakePlanName),
							"id":                  fmt.Sprintf("fake-string-format-%s", fakePlanID),
							"description":         fmt.Sprintf("fake-string-format-%s", fakePlanDescription),
							"additional_property": fmt.Sprintf("fake-string-format-%s", fakePlanProperty),
						},
					}

					viper.Set("service.fake-service.provision.defaults", `{"test": "value", "object": {"key": "value"}}`)
					bytes, err := json.Marshal(fakeServicePlanConfigObject)
					Expect(err).To(Not(HaveOccurred()))
					viper.Set("service.fake-service.plans", string(bytes))
				})

				It("should work", func() {
					plan, err := service.UserDefinedPlans(maintenanceInfo)
					Expect(err).To(Not(HaveOccurred()))
					Expect(plan[0].Name).To(Equal(fakePlanName))
					Expect(plan[1].Name).To(Equal(fmt.Sprintf("fake-string-format-%s", fakePlanName)))
					provisionOverrides, err := service.ProvisionDefaultOverrides()
					Expect(err).To(Not(HaveOccurred()))
					Expect(provisionOverrides).To(
						Equal(
							map[string]any{
								"test":   `value`,
								"object": map[string]any{"key": "value"},
							},
						),
					)
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
				actualPlans, err := service.UserDefinedPlans(maintenanceInfo)

				Expect(err).NotTo(HaveOccurred())
				Expect(actualPlans).To(HaveLen(2))

				sort.Slice(actualPlans, func(i, j int) bool {
					return actualPlans[i].Name < actualPlans[j].Name
				})
				Expect(actualPlans[0].Name).To(Equal(fakePlanName))
				Expect(actualPlans[0].ID).To(Equal(fakePlanGUID))
				Expect(actualPlans[0].Description).To(Equal(fakePlanDescription))
				Expect(actualPlans[0].ServiceProperties).To(Equal(map[string]any{"additional_property": fakePlanProperty, "guid": fakePlanGUID}))
			})

			When("an invalid plan is provided as an environment variable", func() {
				BeforeEach(func() {
					fakeServicePlanTile = `{invalid-json`
					os.Setenv(service.TileUserDefinedPlansVariable(), fakeServicePlanTile)
				})

				It("should return an error", func() {
					_, err := service.UserDefinedPlans(maintenanceInfo)
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
				_, err := service.UserDefinedPlans(maintenanceInfo)
				Expect(err).To(MatchError("fake-service custom plan {ServicePlan:{ID: Name:fakePlanName Description:fakePlanDescription Free:<nil> Bindable:<nil> Metadata:<nil> Schemas:<nil> PlanUpdatable:<nil> MaximumPollingDuration:<nil> MaintenanceInfo:<nil>} ServiceProperties:map[additional_property:fakePlanProperty] ProvisionOverrides:map[] BindOverrides:map[]} is missing an id"))
			})
		})

		When("maintenance info is not nil", func() {
			BeforeEach(func() {
				viper.Set("service.fake-service.plans", fakeServicePlanConfig)
				maintenanceInfo = &domain.MaintenanceInfo{Version: defaultTFVersion, Description: "fake-description"}

			})

			It("returns a broker service plan with same maintenance info for all plans", func() {
				actualPlans, err := service.UserDefinedPlans(maintenanceInfo)

				Expect(err).NotTo(HaveOccurred())
				Expect(actualPlans).To(HaveLen(2))
				Expect(actualPlans[0].MaintenanceInfo).To(Equal(&domain.MaintenanceInfo{Version: defaultTFVersion, Description: "fake-description"}))
				Expect(actualPlans[1].MaintenanceInfo).To(Equal(&domain.MaintenanceInfo{Version: defaultTFVersion, Description: "fake-description"}))
			})

		})
	})
})
