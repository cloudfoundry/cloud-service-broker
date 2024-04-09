package tf_test

import (
	"context"
	"errors"

	"github.com/cloudfoundry/cloud-service-broker/v2/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/broker"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf/executor"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf/tffakes"
	"github.com/cloudfoundry/cloud-service-broker/v2/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GetImportedProperties()", func() {
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

		result, err := tfProvider.GetImportedProperties(context.TODO(), "fakeInstanceGUID", inputVariables, nil)

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

			result, err := tfProvider.GetImportedProperties(context.TODO(), "fakeInstanceGUID", inputVariables, nil)

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

			_, err := tfProvider.GetImportedProperties(context.TODO(), "fakeInstanceGUID", inputVariables, nil)
			Expect(err).To(MatchError(`cannot find required import values for fields: azurerm_mssql_database.not-there.anything`))
		})
	})

	When("there is a tf_attribute_skip condition", func() {
		When("tf_attribute_skip evaluates to true", func() {
			It("skips looking up the attribute", func() {
				provisionDetails := storage.JSONObject{
					"existing": true,
				}

				inputVariables := []broker.BrokerVariable{
					{
						FieldName:       "field_to_replace",
						TFAttribute:     "azurerm_mssql_database.azure_sql_db.not-there",
						TFAttributeSkip: "existing",
					},
				}

				result, err := tfProvider.GetImportedProperties(context.TODO(), "fakeInstanceGUID", inputVariables, provisionDetails)
				Expect(err).NotTo(HaveOccurred())

				Expect(result).To(BeEmpty())
			})
		})

		When("tf_attribute_skip evaluates to false", func() {
			When("the attribute is found", func() {
				It("returns the value", func() {
					provisionDetails := storage.JSONObject{
						"existing": false,
					}

					inputVariables := []broker.BrokerVariable{
						{
							FieldName:       "field_to_replace",
							TFAttribute:     "azurerm_mssql_database.azure_sql_db.subsume-key",
							TFAttributeSkip: "existing",
						},
					}

					result, err := tfProvider.GetImportedProperties(context.TODO(), "fakeInstanceGUID", inputVariables, provisionDetails)
					Expect(err).NotTo(HaveOccurred())
					Expect(result).To(Equal(map[string]any{"field_to_replace": "subsume-value"}))
				})
			})

			When("the attribute is not found", func() {
				It("returns an error", func() {
					provisionDetails := storage.JSONObject{
						"existing": false,
					}

					inputVariables := []broker.BrokerVariable{
						{
							FieldName:       "field_to_replace",
							TFAttribute:     "azurerm_mssql_database.azure_sql_db.not-there",
							TFAttributeSkip: "existing",
						},
					}

					_, err := tfProvider.GetImportedProperties(context.TODO(), "fakeInstanceGUID", inputVariables, provisionDetails)
					Expect(err).NotTo(MatchError(`cannot find required subsumed values for fields: azurerm_mssql_database.azure_sql_db.not-there`))
				})
			})
		})

		When("tf_attribute_skip field is not found", func() {
			It("does not skip", func() {
				inputVariables := []broker.BrokerVariable{
					{
						FieldName:       "field_to_replace",
						TFAttribute:     "azurerm_mssql_database.azure_sql_db.subsume-key",
						TFAttributeSkip: "existing",
					},
				}

				result, err := tfProvider.GetImportedProperties(context.TODO(), "fakeInstanceGUID", inputVariables, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(map[string]any{"field_to_replace": "subsume-value"}))
			})
		})

		When("tf_attribute_skip field has wrong type", func() {
			It("does not skip", func() {
				provisionDetails := storage.JSONObject{
					"existing": "true", // string not boolean!
				}

				inputVariables := []broker.BrokerVariable{
					{
						FieldName:       "field_to_replace",
						TFAttribute:     "azurerm_mssql_database.azure_sql_db.subsume-key",
						TFAttributeSkip: "existing",
					},
				}

				result, err := tfProvider.GetImportedProperties(context.TODO(), "fakeInstanceGUID", inputVariables, provisionDetails)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(map[string]any{"field_to_replace": "subsume-value"}))
			})
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

			_, err := tfProvider.GetImportedProperties(context.TODO(), "fakeInstanceGUID", inputVariables, nil)

			Expect(err).To(MatchError("tf show failed"))
		})
	})
})
