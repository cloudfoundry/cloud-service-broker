package tf_test

import (
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/executor"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v10/domain"
)

var _ = Describe("Definition", func() {
	Describe("ToPlan", func() {
		var (
			plan            *tf.TfServiceDefinitionV1Plan
			maintenanceInfo *domain.MaintenanceInfo
		)

		BeforeEach(func() {
			plan = &tf.TfServiceDefinitionV1Plan{
				Name:               "test-name",
				ID:                 "test-id",
				Description:        "test-description",
				DisplayName:        "test-display-name",
				Bullets:            []string{"test-bullet"},
				Free:               false,
				Properties:         map[string]any{"test-property-key": "test-property-value"},
				ProvisionOverrides: nil,
				BindOverrides:      nil,
			}
		})

		It("returns a broker service plan", func() {
			servicePlan := plan.ToPlan(maintenanceInfo)

			Expect(servicePlan.ServicePlan.ID).To(Equal("test-id"))
			Expect(servicePlan.ServicePlan.Description).To(Equal("test-description"))
			Expect(servicePlan.ServicePlan.Name).To(Equal("test-name"))
			Expect(servicePlan.ServicePlan.Free).To(Equal(domain.FreeValue(false)))
			Expect(servicePlan.ServicePlan.Metadata).To(Equal(&domain.ServicePlanMetadata{
				DisplayName: "test-display-name",
				Bullets:     []string{"test-bullet"},
			}))
			Expect(servicePlan.ServicePlan.MaintenanceInfo).To(BeNil())
			Expect(servicePlan.ServiceProperties).To(Equal(map[string]any{"test-property-key": "test-property-value"}))
			Expect(servicePlan.ProvisionOverrides).To(BeNil())
			Expect(servicePlan.BindOverrides).To(BeNil())

		})

		When("maintenance info is not nil", func() {
			BeforeEach(func() {
				maintenanceInfo = &domain.MaintenanceInfo{
					Version:     "1.0.0",
					Description: "fake-description"}
			})

			It("returns a broker service plan with maintenance info", func() {
				servicePlan := plan.ToPlan(maintenanceInfo)

				Expect(servicePlan.ServicePlan.MaintenanceInfo).To(Equal(&domain.MaintenanceInfo{
					Version:     "1.0.0",
					Description: "fake-description",
				}))

			})
		})
	})

	Describe("ToService", func() {
		var (
			serviceOffering   *tf.TfServiceDefinitionV1
			maintenanceInfo   *domain.MaintenanceInfo
			tfBinariesContext executor.TFBinariesContext
		)

		BeforeEach(func() {
			serviceOffering = &tf.TfServiceDefinitionV1{
				Version:             1,
				Name:                "test-name",
				ID:                  "fa6334bc-5314-4b63-8a74-c0e4b638c950",
				Description:         "test-description",
				DisplayName:         "test-display-name",
				ImageURL:            "https://some-image-url",
				DocumentationURL:    "https://some-url",
				ProviderDisplayName: "company name",
				SupportURL:          "https://some-support-url",
				Tags:                []string{"Beta", "PostgreSQL"},
				Plans:               nil,
				ProvisionSettings:   tf.TfServiceDefinitionV1Action{},
				BindSettings:        tf.TfServiceDefinitionV1Action{},
				Examples:            nil,
				PlanUpdateable:      false,
				RequiredEnvVars:     nil,
			}
		})

		It("returns a broker service offering", func() {
			service, err := serviceOffering.ToService(tfBinariesContext, maintenanceInfo)
			Expect(err).NotTo(HaveOccurred())
			Expect(service.ID).To(Equal("fa6334bc-5314-4b63-8a74-c0e4b638c950"))
			Expect(service.Description).To(Equal("test-description"))
			Expect(service.Name).To(Equal("test-name"))
			Expect(service.DisplayName).To(Equal("test-display-name"))
			Expect(service.DocumentationURL).To(Equal("https://some-url"))
			Expect(service.ProviderDisplayName).To(Equal("company name"))
			Expect(service.SupportURL).To(Equal("https://some-support-url"))
			Expect(service.Tags).To(ConsistOf("Beta", "PostgreSQL"))
		})

		When("no company name is configured", func() {
			It("does not fail validation", func() {
				serviceOffering.ProviderDisplayName = ""
				service, err := serviceOffering.ToService(tfBinariesContext, maintenanceInfo)
				Expect(err).NotTo(HaveOccurred())
				Expect(service.ProviderDisplayName).To(BeEmpty())
			})
		})

		When("maintenance info is not nil", func() {
			BeforeEach(func() {
				maintenanceInfo = &domain.MaintenanceInfo{
					Version:     "1.0.0",
					Description: "fake-description"}

				plan := tf.TfServiceDefinitionV1Plan{
					Name:        "test-name",
					ID:          "fa6334bc-5314-4b63-8a74-c0e4b638c951",
					Description: "test-description",
					DisplayName: "test-display-name",
				}

				serviceOffering.Plans = []tf.TfServiceDefinitionV1Plan{plan}
			})

			It("returns a broker service plan with maintenance info", func() {
				service, err := serviceOffering.ToService(tfBinariesContext, maintenanceInfo)
				Expect(err).NotTo(HaveOccurred())
				Expect(service.Plans[0].MaintenanceInfo).To(Equal(&domain.MaintenanceInfo{
					Version:     "1.0.0",
					Description: "fake-description",
				}))

			})
		})
	})
})
