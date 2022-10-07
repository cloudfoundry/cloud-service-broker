package tf_test

import (
	"context"
	"errors"

	"github.com/cloudfoundry/cloud-service-broker/internal/storage"

	"github.com/cloudfoundry/cloud-service-broker/pkg/broker"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/executor"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/tffakes"
	"github.com/cloudfoundry/cloud-service-broker/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GetImportedProperties()", func() {
	When("the offering is not subsumable", func() {
		It("should not return variables or error", func() {
			const defaultPlanGUID = "6526a7be-8504-11ec-b558-276c48808143"
			tfProvider := tf.NewTerraformProvider(
				executor.TFBinariesContext{}, nil,
				utils.NewLogger("test"),
				tf.TfServiceDefinitionV1{
					Plans: []tf.TfServiceDefinitionV1Plan{
						{
							Name: "default-plan",
							ID:   defaultPlanGUID,
						},
					},
				},
				&tffakes.FakeDeploymentManagerInterface{},
			)

			inputVariables := []broker.BrokerVariable{
				{
					FieldName:   "field_to_replace",
					TFAttribute: "azurerm_mssql_database.azure_sql_db.subsume-key",
				},
			}

			result, err := tfProvider.GetImportedProperties(context.TODO(), defaultPlanGUID, "", inputVariables)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeEmpty())
		})
	})

	When("the offering is subsumable", func() {
		var (
			tfProvider            broker.ServiceProvider
			subsumePlanGUID       string
			fakeInvokerBuilder    *tffakes.FakeTerraformInvokerBuilder
			fakeInvoker           *tffakes.FakeTerraformInvoker
			fakeDeploymentManager *tffakes.FakeDeploymentManagerInterface
		)

		BeforeEach(func() {
			fakeDeploymentManager = new(tffakes.FakeDeploymentManagerInterface)
			fakeInvoker = new(tffakes.FakeTerraformInvoker)
			fakeInvokerBuilder = new(tffakes.FakeTerraformInvokerBuilder)

			subsumePlanGUID = "6526a7be-8504-11ec-b558-276c48808143"
			fakeInvokerBuilder.VersionedTerraformInvokerReturns(fakeInvoker)
			fakeInvoker.ShowReturns("# azurerm_mssql_database.azure_sql_db:\nresource \"azurerm_mssql_database\" \"azure_sql_db\" {\nsubsume-key = \"subsume-value\"\n}\nOutputs:\nname = \"test-name\"", nil)

			tfProvider = tf.NewTerraformProvider(
				executor.TFBinariesContext{},
				fakeInvokerBuilder,
				utils.NewLogger("test"),
				tf.TfServiceDefinitionV1{
					Plans: []tf.TfServiceDefinitionV1Plan{
						{
							Name: "subsume-plan",
							ID:   subsumePlanGUID,
							Properties: map[string]any{
								"subsume": true,
							},
						},
					},
				},
				fakeDeploymentManager,
			)
		})

		It("returns subsumed variables", func() {
			fakeDeploymentManager.GetTerraformDeploymentReturns(storage.TerraformDeployment{}, nil)

			inputVariables := []broker.BrokerVariable{
				{
					FieldName:   "field_to_replace",
					TFAttribute: "azurerm_mssql_database.azure_sql_db.subsume-key",
				},
			}

			result, err := tfProvider.GetImportedProperties(context.TODO(), subsumePlanGUID, "fakeInstanceGUID", inputVariables)

			actualTfID := fakeDeploymentManager.GetTerraformDeploymentArgsForCall(0)
			Expect(actualTfID).To(Equal("tf:fakeInstanceGUID:"))
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(map[string]any{"field_to_replace": "subsume-value"}))
		})

		When("no replace vars are defined", func() {
			It("returns empty list and no error", func() {
				inputVariables := []broker.BrokerVariable{
					{
						FieldName: "field_to_replace",
					},
				}

				result, err := tfProvider.GetImportedProperties(context.TODO(), subsumePlanGUID, "fakeInstanceGUID", inputVariables)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeEmpty())
				Expect(fakeDeploymentManager.GetTerraformDeploymentCallCount()).To(BeZero())
				Expect(fakeInvoker.ShowCallCount()).To(BeZero())
			})
		})

		When("the tf_attribute lookup fails", func() {
			It("returns an error", func() {
				inputVariables := []broker.BrokerVariable{
					{
						FieldName:   "field_to_replace",
						TFAttribute: "azurerm_mssql_database.not-there.anything",
					},
				}

				_, err := tfProvider.GetImportedProperties(context.TODO(), subsumePlanGUID, "fakeInstanceGUID", inputVariables)
				Expect(err).To(MatchError(`cannot find required subsumed values for fields: azurerm_mssql_database.not-there.anything`))
			})
		})

		When("tf show fails", func() {
			It("returns an error", func() {
				fakeInvoker.ShowReturns("", errors.New("tf show failed"))
				fakeDeploymentManager.GetTerraformDeploymentReturns(storage.TerraformDeployment{}, nil)

				inputVariables := []broker.BrokerVariable{
					{
						FieldName:   "field_to_replace",
						TFAttribute: "resourc.name.attribute",
					},
				}

				_, err := tfProvider.GetImportedProperties(context.TODO(), subsumePlanGUID, "fakeInstanceGUID", inputVariables)

				Expect(err).To(MatchError("tf show failed"))
			})
		})
	})
})
