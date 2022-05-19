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
	Describe("CreateAndSaveDeployment", func() {
		var (
			fakeStore         brokerfakes.FakeServiceProviderStorage
			deploymentManager *tf.DeploymentManager
			ws                *workspace.TerraformWorkspace
			deploymentID      string
		)

		BeforeEach(func() {
			fakeStore = brokerfakes.FakeServiceProviderStorage{}
			deploymentManager = tf.NewDeploymentManager(&fakeStore)
			ws = &workspace.TerraformWorkspace{
				Modules: []workspace.ModuleDefinition{{
					Name:       "fake module name",
					Definition: "fake definition",
				}},
			}
			deploymentID = "tf:instance:binding"
		})

		It("stores a new deployment", func() {
			actualDeployment, err := deploymentManager.CreateAndSaveDeployment(deploymentID, ws)

			By("checking the deployment object is correct")
			Expect(err).NotTo(HaveOccurred())
			Expect(actualDeployment.ID).To(Equal(deploymentID))
			Expect(actualDeployment.Workspace).To(Equal(ws))
			Expect(actualDeployment.LastOperationType).To(Equal("validation"))
			Expect(actualDeployment.LastOperationState).To(BeEmpty())
			Expect(actualDeployment.LastOperationMessage).To(BeEmpty())

			By("validating a call to store was made")
			Expect(fakeStore.StoreTerraformDeploymentCallCount()).To(Equal(1))
			storedDeployment := fakeStore.StoreTerraformDeploymentArgsForCall(0)
			Expect(storedDeployment).To(Equal(actualDeployment))
		})

		When("deployment exists", func() {
			var existingDeployment storage.TerraformDeployment
			BeforeEach(func() {
				existingDeployment = storage.TerraformDeployment{
					ID: deploymentID,
					Workspace: &workspace.TerraformWorkspace{
						Modules: []workspace.ModuleDefinition{{
							Name: "existing module",
						}},
					},
					LastOperationType:    "provision",
					LastOperationState:   "in progress",
					LastOperationMessage: "test",
				}
				fakeStore.ExistsTerraformDeploymentReturns(true, nil)
				fakeStore.GetTerraformDeploymentReturns(existingDeployment, nil)
			})

			It("updates workspace if deployment exists", func() {
				actualDeployment, err := deploymentManager.CreateAndSaveDeployment(deploymentID, ws)

				Expect(err).NotTo(HaveOccurred())
				Expect(actualDeployment.ID).To(Equal(deploymentID))
				Expect(actualDeployment.Workspace).To(Equal(ws))
				Expect(actualDeployment.LastOperationType).To(Equal("validation"))
				Expect(actualDeployment.LastOperationState).To(Equal("in progress"))
				Expect(actualDeployment.LastOperationMessage).To(Equal("test"))

				By("validating a call to store was made")
				Expect(fakeStore.StoreTerraformDeploymentCallCount()).To(Equal(1))
				storedDeployment := fakeStore.StoreTerraformDeploymentArgsForCall(0)
				Expect(storedDeployment).To(Equal(actualDeployment))
			})
		})

		It("fails, when checking if deployment exists fails", func() {
			fakeStore.ExistsTerraformDeploymentReturns(true, errors.New("failed to check"))

			_, err := deploymentManager.CreateAndSaveDeployment(deploymentID, ws)

			Expect(err).To(MatchError("failed to check"))
		})

		It("fails, when getting deployment fails", func() {
			fakeStore.ExistsTerraformDeploymentReturns(true, nil)
			fakeStore.GetTerraformDeploymentReturns(storage.TerraformDeployment{}, errors.New("failed to get deployment"))

			_, err := deploymentManager.CreateAndSaveDeployment(deploymentID, ws)

			Expect(err).To(MatchError("failed to get deployment"))
		})

	})

	Describe("MarkOperationStarted", func() {
		var (
			fakeStore         brokerfakes.FakeServiceProviderStorage
			deploymentManager *tf.DeploymentManager
			deployment        storage.TerraformDeployment
		)

		BeforeEach(func() {
			fakeStore = brokerfakes.FakeServiceProviderStorage{}
			deploymentManager = tf.NewDeploymentManager(&fakeStore)
			deployment = storage.TerraformDeployment{
				ID:                "tf:instance:binding",
				LastOperationType: "validation",
			}
		})

		It("updates last operation to in progress", func() {
			err := deploymentManager.MarkOperationStarted(deployment, "provision")

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeStore.StoreTerraformDeploymentCallCount()).To(Equal(1))
			storedDeployment := fakeStore.StoreTerraformDeploymentArgsForCall(0)

			Expect(storedDeployment.ID).To(Equal("tf:instance:binding"))
			Expect(storedDeployment.LastOperationType).To(Equal("provision"))
			Expect(storedDeployment.LastOperationState).To(Equal("in progress"))
			Expect(storedDeployment.LastOperationMessage).To(Equal(""))
		})

		It("fails, when storing deployment fails", func() {
			fakeStore.StoreTerraformDeploymentReturns(errors.New("couldn't store deployment"))

			err := deploymentManager.MarkOperationStarted(deployment, "provision")

			Expect(err).To(MatchError("couldn't store deployment"))
		})
	})

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
