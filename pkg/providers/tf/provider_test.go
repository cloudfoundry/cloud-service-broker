package tf_test

import (
	"context"
	"encoding/json"

	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/broker/brokerfakes"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/providers/tf"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/providers/tf/tffakes"
	"github.com/cloudfoundry-incubator/cloud-service-broker/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = FDescribe("Provider", func() {
	Describe("AddImportedProperties", func() {
		When("instance was not subsumed", func() {
			It("should not change variables", func() {
				defaultPlanGUID := "6526a7be-8504-11ec-b558-276c48808143"
				storage := new(brokerfakes.FakeServiceProviderStorage)

				tfProvider := tf.NewTerraformProvider(
					tf.NewTfJobRunner(nil, storage),
					utils.NewLogger("test"),
					tf.TfServiceDefinitionV1{
						Plans: []tf.TfServiceDefinitionV1Plan{
							{
								Name: "default-plan",
								Id:   defaultPlanGUID,
							},
						},
					},
					storage,
				)

				provisionInput := json.RawMessage(`{"foo":"some=param"}`)

				result, err := tfProvider.AddImportedProperties(context.TODO(), defaultPlanGUID, provisionInput)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(provisionInput))
			})
		})

		When("instance was subsumed", func() {
			It("should add subsumed variables", func() {
				subsumePlanGUID := "6526a7be-8504-11ec-b558-276c48808143"
				storage := new(brokerfakes.FakeServiceProviderStorage)
				jobRunner := new(tffakes.FakeJobRunner)

				tfProvider := tf.NewTerraformProvider(
					jobRunner,
					utils.NewLogger("test"),
					tf.TfServiceDefinitionV1{
						Plans: []tf.TfServiceDefinitionV1Plan{
							{
								Name: "subsume-plan",
								Id:   subsumePlanGUID,
								Properties: map[string]interface{}{
									"subsume": true,
								},
							},
						},
					},
					storage,
				)

				provisionInput := json.RawMessage(`{"foo":"some=param"}`)

				result, err := tfProvider.AddImportedProperties(context.TODO(), subsumePlanGUID, provisionInput)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(json.RawMessage(`{"foo":"some=param","subsume-key":"subsume-value"}`)))
			})
		})
	})
})
