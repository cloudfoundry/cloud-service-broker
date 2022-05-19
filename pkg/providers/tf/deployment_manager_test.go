package tf_test

import (
	"errors"

	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/pkg/broker"
	"github.com/cloudfoundry/cloud-service-broker/pkg/broker/brokerfakes"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/workspace"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
)

var _ = Describe("DeploymentManager", func() {

	Describe("UpdateWorkspaceHCL", func() {
		const (
			id                   = "fake-id"
			lastOperationType    = "fake operation"
			lastOperationState   = "fake operation state"
			lastOperationMessage = "fake operation message"
			terraformState       = "fake terraform state"
			template             = `
				variable resourceGroup {type = string}
	
				resource "random_string" "username" {
				  length = 16
				  special = false
				  number = false
				}

				resource "azurerm_mssql_database" "azure_sql_db" {
				  name                = "dbname"
				  resource_group_name = var.resourceGroup
				  administrator_login = random_string.username.result
				}

				output username {value = random_string.username.result}
				`
		)

		var (
			templateVars             map[string]interface{}
			store                    *brokerfakes.FakeServiceProviderStorage
			deploymentManager        *tf.DeploymentManager
			updatedProvisionSettings tf.TfServiceDefinitionV1Action
		)

		BeforeEach(func() {
			By("setting up fakes", func() {
				viper.Reset()
				store = &brokerfakes.FakeServiceProviderStorage{}
				deploymentManager = tf.NewDeploymentManager(store)
				templateVars = map[string]interface{}{}
			})

			By("creating a fake provisioned service instance", func() {
				workspace := &workspace.TerraformWorkspace{
					Modules: []workspace.ModuleDefinition{{
						Name:       "fake module name",
						Definition: "fake definition",
					}},
					Instances: []workspace.ModuleInstance{{
						ModuleName:   "fake module name",
						InstanceName: "fake instance name",
					}},
					Transformer: workspace.TfTransformer{},
					State:       []byte(terraformState),
				}

				store.GetTerraformDeploymentReturns(storage.TerraformDeployment{
					ID:                   id,
					Workspace:            workspace,
					LastOperationType:    lastOperationType,
					LastOperationState:   lastOperationState,
					LastOperationMessage: lastOperationMessage,
				}, nil)
			})

			By("having an updated service definition from the brokerpak", func() {
				updatedProvisionSettings = tf.TfServiceDefinitionV1Action{
					PlanInputs: []broker.BrokerVariable{
						{
							FieldName: "resourceGroup",
							Type:      broker.JSONTypeString,
							Details:   "The resource group name",
							Required:  true,
						},
					},
					Template: template,
					Outputs: []broker.BrokerVariable{
						{
							FieldName: "username",
							Type:      broker.JSONTypeString,
							Details:   "The administrator username",
							Required:  true,
						},
					},
				}
			})
		})

		When("brokerpak updates enabled", func() {
			BeforeEach(func() {
				viper.Set("brokerpak.updates.enabled", true)
			})

			It("updates the modules but keeps the original state", func() {
				err := deploymentManager.UpdateWorkspaceHCL(id, updatedProvisionSettings, templateVars)
				Expect(err).NotTo(HaveOccurred())

				By("checking that the right deployment is retrieved")
				Expect(store.GetTerraformDeploymentCallCount()).To(Equal(1))
				Expect(store.GetTerraformDeploymentArgsForCall(0)).To(Equal(id))

				By("checking that the updated deployment is stored")
				Expect(store.StoreTerraformDeploymentCallCount()).To(Equal(1))
				actualTerraformDeployment := store.StoreTerraformDeploymentArgsForCall(0)
				Expect(actualTerraformDeployment.ID).To(Equal(id))
				Expect(actualTerraformDeployment.LastOperationType).To(Equal("fake operation"))
				Expect(actualTerraformDeployment.LastOperationState).To(Equal("fake operation state"))
				Expect(actualTerraformDeployment.LastOperationMessage).To(Equal("fake operation message"))

				By("checking that the modules and instances are updated, but the state remains the same")
				expectedWorkspace := &workspace.TerraformWorkspace{
					Modules: []workspace.ModuleDefinition{{
						Name:       "brokertemplate",
						Definition: template,
					}},
					Instances: []workspace.ModuleInstance{{
						ModuleName:   "brokertemplate",
						InstanceName: "instance",
						Configuration: map[string]interface{}{
							"resourceGroup": nil,
						},
					}},
					Transformer: workspace.TfTransformer{
						ParameterMappings:  []workspace.ParameterMapping{},
						ParametersToRemove: []string{},
						ParametersToAdd:    []workspace.ParameterMapping{},
					},
					State: []byte(terraformState),
				}
				Expect(actualTerraformDeployment.Workspace).To(Equal(expectedWorkspace))
			})

			When("getting deployment fails", func() {
				BeforeEach(func() {
					store.GetTerraformDeploymentReturns(storage.TerraformDeployment{}, errors.New("boom"))
				})

				It("returns the error", func() {
					err := deploymentManager.UpdateWorkspaceHCL(id, updatedProvisionSettings, templateVars)
					Expect(err).To(MatchError("boom"))
				})
			})

			When("cannot create a workspace", func() {
				It("returns the error", func() {
					jammedOperationSettings := tf.TfServiceDefinitionV1Action{
						Template: `
				resource "azurerm_mssql_database" "azure_sql_db" {
				  name                = 
				}
				`,
					}
					err := deploymentManager.UpdateWorkspaceHCL(id, jammedOperationSettings, templateVars)
					Expect(err).To(MatchError(ContainSubstring("Invalid expression")))
				})
			})

			When("cannot save the deployment", func() {
				BeforeEach(func() {
					store.StoreTerraformDeploymentReturns(errors.New("fake error"))
				})

				It("returns the error", func() {
					err := deploymentManager.UpdateWorkspaceHCL(id, updatedProvisionSettings, templateVars)
					Expect(err).To(MatchError("terraform provider create failed: fake error"))
				})
			})
		})

		When("brokerpak updates disabled", func() {
			It("does not update the store", func() {
				err := deploymentManager.UpdateWorkspaceHCL(id, updatedProvisionSettings, templateVars)
				Expect(err).NotTo(HaveOccurred())

				Expect(store.StoreTerraformDeploymentCallCount()).To(BeZero())
			})
		})
	})
})
