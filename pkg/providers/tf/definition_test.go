package tf_test

import (
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v8/domain"
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
				Properties:         map[string]interface{}{"test-property-key": "test-property-value"},
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
			Expect(servicePlan.ServiceProperties).To(Equal(map[string]interface{}{"test-property-key": "test-property-value"}))
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
})
