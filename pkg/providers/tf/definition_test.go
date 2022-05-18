package tf_test

import (
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/executor"
	"github.com/hashicorp/go-version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v8/domain"
	"github.com/spf13/viper"
)

var _ = Describe("Definition", func() {
	Describe("ToPlan", func() {
		var (
			plan *tf.TfServiceDefinitionV1Plan
			exec executor.TFBinariesContext
		)

		BeforeEach(func() {
			plan = &tf.TfServiceDefinitionV1Plan{
				Name:               "test-name",
				ID:                 "test-id",
				Description:        "test-description",
				DisplayName:        "test-display-name",
				Bullets:            []string{"test-bullet"},
				Free:               false,
				Properties:         map[string]interface{}{"test-property-key": "test-property-value"},
				ProvisionOverrides: nil,
				BindOverrides:      nil,
			}
			exec = executor.TFBinariesContext{}
		})

		When("a valid plan is given", func() {
			It("returns a broker service plan", func() {
				servicePlan := plan.ToPlan(exec)

				Expect(servicePlan.ServicePlan.ID).To(Equal("test-id"))
				Expect(servicePlan.ServicePlan.Description).To(Equal("test-description"))
				Expect(servicePlan.ServicePlan.Name).To(Equal("test-name"))
				Expect(servicePlan.ServicePlan.Free).To(Equal(domain.FreeValue(false)))
				Expect(servicePlan.ServicePlan.Metadata).To(Equal(&domain.ServicePlanMetadata{
					DisplayName: "test-display-name",
					Bullets:     []string{"test-bullet"},
				}))
				Expect(servicePlan.ServicePlan.MaintenanceInfo).To(BeNil())
				Expect(servicePlan.ServiceProperties).To(Equal(map[string]interface{}{"test-property-key": "test-property-value"}))
				Expect(servicePlan.ProvisionOverrides).To(BeNil())
				Expect(servicePlan.BindOverrides).To(BeNil())

			})
		})

		When("TFUpgrades are enabled", func() {
			BeforeEach(func() {
				viper.Set(tf.TfUpgradeEnabled, true)

				tfVersion := version.Must(version.NewVersion("1.0.0"))
				exec = executor.TFBinariesContext{
					DefaultTfVersion: tfVersion,
				}
			})

			AfterEach(func() {
				viper.Reset()
			})

			It("returns a broker service plan with maintenance info version matching default TF version", func() {
				expectedMaintenanceInfo := domain.MaintenanceInfo{
					Version:     "1.0.0",
					Description: "This upgrade provides support for Terraform version: 1.0.0. The upgrade operation will take a while. The instance and all associated bindings will be upgraded.",
				}

				servicePlan := plan.ToPlan(exec)

				Expect(servicePlan.ServicePlan.ID).To(Equal("test-id"))
				Expect(servicePlan.ServicePlan.Description).To(Equal("test-description"))
				Expect(servicePlan.ServicePlan.Name).To(Equal("test-name"))
				Expect(servicePlan.ServicePlan.Free).To(Equal(domain.FreeValue(false)))
				Expect(servicePlan.ServicePlan.Metadata).To(Equal(&domain.ServicePlanMetadata{
					DisplayName: "test-display-name",
					Bullets:     []string{"test-bullet"},
				}))
				Expect(servicePlan.ServicePlan.MaintenanceInfo).To(Equal(&expectedMaintenanceInfo))
				Expect(servicePlan.ServiceProperties).To(Equal(map[string]interface{}{"test-property-key": "test-property-value"}))
				Expect(servicePlan.ProvisionOverrides).To(BeNil())
				Expect(servicePlan.BindOverrides).To(BeNil())

			})
		})

		When("TFUpgrades are disabled", func() {
			BeforeEach(func() {
				tfVersion := version.Must(version.NewVersion("1.0.0"))
				exec = executor.TFBinariesContext{
					DefaultTfVersion: tfVersion,
				}
			})

			It("returns a broker service plan with nil maintenance info", func() {
				servicePlan := plan.ToPlan(exec)

				Expect(servicePlan.ServicePlan.ID).To(Equal("test-id"))
				Expect(servicePlan.ServicePlan.Description).To(Equal("test-description"))
				Expect(servicePlan.ServicePlan.Name).To(Equal("test-name"))
				Expect(servicePlan.ServicePlan.Free).To(Equal(domain.FreeValue(false)))
				Expect(servicePlan.ServicePlan.Metadata).To(Equal(&domain.ServicePlanMetadata{
					DisplayName: "test-display-name",
					Bullets:     []string{"test-bullet"},
				}))
				Expect(servicePlan.ServicePlan.MaintenanceInfo).To(BeNil())
				Expect(servicePlan.ServiceProperties).To(Equal(map[string]interface{}{"test-property-key": "test-property-value"}))
				Expect(servicePlan.ProvisionOverrides).To(BeNil())
				Expect(servicePlan.BindOverrides).To(BeNil())

			})
		})
	})
})
