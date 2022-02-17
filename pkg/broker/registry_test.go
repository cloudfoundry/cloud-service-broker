package broker_test

import (
	"fmt"

	. "github.com/cloudfoundry/cloud-service-broker/pkg/broker"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v8/domain"
	"github.com/spf13/viper"
)

var _ = Describe("Registry", func() {
	AfterEach(func() {
		defer viper.Reset()
	})

	Describe("Register", func() {
		var serviceDef ServiceDefinition
		BeforeEach(func() {
			serviceDef = ServiceDefinition{
				Id:   "b9e4332e-b42b-4680-bda5-ea1506797474",
				Name: "test-service",
				Plans: []ServicePlan{
					{
						ServicePlan: domain.ServicePlan{
							ID:          "e1d11f65-da66-46ad-977c-6d56513baf43",
							Name:        "Builtin!",
							Description: "Standard storage class",
						},
					},
				},
				IsBuiltin: true,
			}
		})

		It("fails when the service offering is already registered", func() {
			registry := BrokerRegistry{
				"test-service": &ServiceDefinition{
					Id:   "b9e4332e-b42b-4680-bda5-ea1506797474",
					Name: "test-service",
				},
			}

			err := registry.Register(&serviceDef)
			Expect(err).To(MatchError(`tried to register multiple instances of: "test-service"`))
		})

		Context("user defined plans", func() {
			It("appends user defined plans to brokerpak service plans", func() {
				const userProvidedPlan = `[{"name": "user-plan","id":"8b52a460-b246-11eb-a8f5-d349948e2480"}]`
				viper.Set("service.test-service.plans", userProvidedPlan)

				registry := BrokerRegistry{}

				err := registry.Register(&serviceDef)
				Expect(err).ToNot(HaveOccurred())

				Expect(len(registry["test-service"].Plans)).To(Equal(2))
			})

			It("errors when user defined plans have duplicate plan Id", func() {
				const userProvidedPlan = `[{"name": "user-plan","id":"e1d11f65-da66-46ad-977c-6d56513baf43"}]`
				viper.Set("service.test-service.plans", userProvidedPlan)

				registry := BrokerRegistry{}

				err := registry.Register(&serviceDef)
				Expect(err).To(MatchError(`error validating service "test-service", duplicated value, must be unique: e1d11f65-da66-46ad-977c-6d56513baf43: Plans[1].Id`))
			})

			It("errors when user defined plans have duplicate name Id", func() {
				const userProvidedPlan = `[{"name": "Builtin!","id":"8b52a460-b246-11eb-a8f5-d349948e2480"}]`
				viper.Set("service.test-service.plans", userProvidedPlan)

				registry := BrokerRegistry{}

				err := registry.Register(&serviceDef)
				Expect(err).To(MatchError(`error validating service "test-service", duplicated value, must be unique: Builtin!: Plans[1].Name`))
			})
		})

		Context("no plans defined", func() {
			It("defines a default plan", func() {
				registry := make(BrokerRegistry)

				err := registry.Register(&ServiceDefinition{
					Id:        "b9e4332e-b42b-4680-bda5-ea1506797474",
					Name:      "test-service",
					Plans:     []ServicePlan{},
					IsBuiltin: true,
				})
				Expect(err).To(MatchError(`service "test-service" has no plans defined; at least one plan must be specified in the service definition or via the environment variable "GSB_SERVICE_TEST_SERVICE_PLANS" or "TEST_SERVICE_CUSTOM_PLANS"`))
			})
		})
	})

	Describe("Validate", func() {
		It("should fail when same service ID is used in two different services", func() {
			const duplicateID = "b9e4332e-b42b-4680-bda5-ea1506797474"
			registry := BrokerRegistry{
				"test-service-1": &ServiceDefinition{
					Id:   duplicateID,
					Name: "test-service-1",
				},
				"test-service-2": &ServiceDefinition{
					Id:   duplicateID,
					Name: "test-service-2",
				},
			}

			err := registry.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(fmt.Sprintf("duplicated value, must be unique: %s: services[1].Id", duplicateID)))
		})

		It("should fail when same service name is used in two different services", func() {
			duplicateName := "test-service"
			registry := BrokerRegistry{
				"test-service-1": &ServiceDefinition{
					Id:   "b9e4332e-b42b-4680-bda5-ea1506797474",
					Name: duplicateName,
				},
				"test-service-2": &ServiceDefinition{
					Id:   "1324f91e-04cd-11ec-94ab-579a8238e388",
					Name: duplicateName,
				},
			}

			err := registry.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(fmt.Sprintf("duplicated value, must be unique: %s: services[1].Name", duplicateName)))
		})

		It("should fail when same plan ID is used in two different services", func() {
			const duplicateID = "e1d11f65-da66-46ad-977c-6d56513baf43"
			registry := BrokerRegistry{
				"test-service-1": &ServiceDefinition{
					Id:   "b9e4332e-b42b-4680-bda5-ea1506797474",
					Name: "test-service-1",
					Plans: []ServicePlan{
						{
							ServicePlan: domain.ServicePlan{
								ID:          duplicateID,
								Name:        "test-plan-1",
								Description: "Standard storage class",
							},
						},
					},
				},
				"test-service-2": &ServiceDefinition{
					Id:   "c19eb6cc-04c5-11ec-ab31-3b165292c41b",
					Name: "test-service-2",
					Plans: []ServicePlan{
						{
							ServicePlan: domain.ServicePlan{
								ID:          duplicateID,
								Name:        "test-plan-2",
								Description: "Standard storage class",
							},
						},
					},
				},
			}

			err := registry.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(fmt.Sprintf("duplicated value, must be unique: %s: services[1].Plans[0].Id", duplicateID)))
		})
	})

	Describe("GetEnabledServices", func() {
		DescribeTable("should not show offering",
			func(tag, property string) {
				serviceDef := ServiceDefinition{
					Id:   "b9e4332e-b42b-4680-bda5-ea1506797474",
					Name: "test-service",
					Tags: []string{"gcp", tag},
					Plans: []ServicePlan{
						{
							ServicePlan: domain.ServicePlan{
								ID:          "e1d11f65-da66-46ad-977c-6d56513baf43",
								Name:        "Builtin!",
								Description: "Standard storage class",
							},
						},
					},
					IsBuiltin: true,
				}

				viper.Set(property, false)
				viper.Set("compatibility.enable-builtin-services", true)

				registry := BrokerRegistry{}
				err := registry.Register(&serviceDef)
				Expect(err).ToNot(HaveOccurred())

				result, err := registry.GetEnabledServices()
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeEmpty())
			},
			Entry("when preview are disabled and build-ins are enabled", "preview", "compatibility.enable-preview-services"),
			Entry("when unmaintained are disabled and build-ins are enabled", "unmaintained", "compatibility.enable-unmaintained-services"),
			Entry("when eol are disabled and build-ins are enabled", "eol", "compatibility.enable-eol-services"),
			Entry("when beta are disabled and build-ins are enabled", "beta", "compatibility.enable-gcp-beta-services"),
			Entry("when deprecated are disabled and build-ins are enabled", "deprecated", "compatibility.enable-gcp-deprecated-services"),
		)

		DescribeTable("should show offering",
			func(tag, property string) {
				serviceDef := ServiceDefinition{
					Id:   "b9e4332e-b42b-4680-bda5-ea1506797474",
					Name: "test-service",
					Tags: []string{"gcp", tag},
					Plans: []ServicePlan{
						{
							ServicePlan: domain.ServicePlan{
								ID:          "e1d11f65-da66-46ad-977c-6d56513baf43",
								Name:        "Builtin!",
								Description: "Standard storage class",
							},
						},
					},
					IsBuiltin: true,
				}

				viper.Set(property, true)
				viper.Set("compatibility.enable-builtin-services", true)

				registry := BrokerRegistry{}
				err := registry.Register(&serviceDef)
				Expect(err).ToNot(HaveOccurred())

				result, err := registry.GetEnabledServices()
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(HaveLen(1))
			},
			Entry("when preview are enabled and build-ins are enabled", "preview", "compatibility.enable-preview-services"),
			Entry("when unmaintained are enabled and build-ins are enabled", "unmaintained", "compatibility.enable-unmaintained-services"),
			Entry("when eol are enabled and build-ins are enabled", "eol", "compatibility.enable-eol-services"),
			Entry("when beta are enabled and build-ins are enabled", "beta", "compatibility.enable-gcp-beta-services"),
			Entry("when deprecated are enabled and build-ins are enabled", "deprecated", "compatibility.enable-gcp-deprecated-services"),
		)

		DescribeTable("should not show offering",
			func(tag, property string) {
				serviceDef := ServiceDefinition{
					Id:   "b9e4332e-b42b-4680-bda5-ea1506797474",
					Name: "test-service",
					Tags: []string{"gcp", tag},
					Plans: []ServicePlan{
						{
							ServicePlan: domain.ServicePlan{
								ID:          "e1d11f65-da66-46ad-977c-6d56513baf43",
								Name:        "Builtin!",
								Description: "Standard storage class",
							},
						},
					},
					IsBuiltin: true,
				}

				viper.Set(property, true)
				viper.Set("compatibility.enable-builtin-services", false)

				registry := BrokerRegistry{}
				err := registry.Register(&serviceDef)
				Expect(err).ToNot(HaveOccurred())

				result, err := registry.GetEnabledServices()
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeEmpty())
			},
			Entry("when preview are enabled and build-ins are disabled", "preview", "compatibility.enable-preview-services"),
			Entry("when unmaintained are enabled and build-ins are disabled", "unmaintained", "compatibility.enable-unmaintained-services"),
			Entry("when eol are enabled and build-ins are disabled", "eol", "compatibility.enable-eol-services"),
			Entry("when beta are enabled and build-ins are disabled", "beta", "compatibility.enable-gcp-beta-services"),
			Entry("when deprecated are enabled and build-ins are disabled", "deprecated", "compatibility.enable-gcp-deprecated-services"),
		)
	})
})
