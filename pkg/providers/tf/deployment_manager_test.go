package tf_test

import (
	"errors"

	"code.cloudfoundry.org/lager/v3/lagertest"

	"github.com/cloudfoundry/cloud-service-broker/v3/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/broker"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/broker/brokerfakes"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/featureflags"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/providers/tf"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/providers/tf/workspace"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/providers/tf/workspace/workspacefakes"
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
			deploymentManager = tf.NewDeploymentManager(&fakeStore, lagertest.NewTestLogger("test"))
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
			Expect(actualDeployment.LastOperationType).To(BeEmpty())
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
				Expect(actualDeployment.LastOperationType).To(Equal("provision"))
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
			fakeStore          brokerfakes.FakeServiceProviderStorage
			deploymentManager  *tf.DeploymentManager
			existingDeployment storage.TerraformDeployment
		)

		BeforeEach(func() {
			fakeStore = brokerfakes.FakeServiceProviderStorage{}
			deploymentManager = tf.NewDeploymentManager(&fakeStore, lagertest.NewTestLogger("test"))
			existingDeployment = storage.TerraformDeployment{
				ID: "tf:instance:binding",
				Workspace: &workspace.TerraformWorkspace{
					Modules: []workspace.ModuleDefinition{{
						Name: "existing module",
					}},
				},
				LastOperationType: "validation",
			}
		})

		It("updates last operation to in progress", func() {
			err := deploymentManager.MarkOperationStarted(&existingDeployment, "provision")

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeStore.StoreTerraformDeploymentCallCount()).To(Equal(1))
			storedDeployment := fakeStore.StoreTerraformDeploymentArgsForCall(0)

			Expect(storedDeployment.ID).To(Equal(existingDeployment.ID))
			Expect(storedDeployment.Workspace).To(Equal(existingDeployment.Workspace))
			Expect(storedDeployment.LastOperationType).To(Equal("provision"))
			Expect(storedDeployment.LastOperationState).To(Equal("in progress"))
			Expect(storedDeployment.LastOperationMessage).To(Equal("provision in progress"))
		})

		It("fails, when storing deployment fails", func() {
			fakeStore.StoreTerraformDeploymentReturns(errors.New("couldn't store deployment"))

			err := deploymentManager.MarkOperationStarted(&existingDeployment, "provision")

			Expect(err).To(MatchError("couldn't store deployment"))
		})
	})

	Describe("MarkOperationFinished", func() {
		var (
			fakeStore          brokerfakes.FakeServiceProviderStorage
			deploymentManager  *tf.DeploymentManager
			fakeLogger         *lagertest.TestLogger
			existingDeployment storage.TerraformDeployment
			fakeWorkspace      *workspacefakes.FakeWorkspace
		)

		BeforeEach(func() {
			fakeWorkspace = &workspacefakes.FakeWorkspace{}
			fakeWorkspace.ModuleInstancesReturns([]workspace.ModuleInstance{{InstanceName: "test-name"}})
			fakeWorkspace.OutputsReturns(map[string]any{}, nil)
			existingDeployment = storage.TerraformDeployment{
				ID:                   "deploymentID",
				Workspace:            fakeWorkspace,
				LastOperationType:    "provision",
				LastOperationState:   "in progress",
				LastOperationMessage: "test",
			}
			fakeStore = brokerfakes.FakeServiceProviderStorage{}
			fakeLogger = lagertest.NewTestLogger("broker")
			deploymentManager = tf.NewDeploymentManager(&fakeStore, fakeLogger)
		})

		When("operation finished successfully", func() {
			It("sets operation state to succeeded", func() {
				err := deploymentManager.MarkOperationFinished(&existingDeployment, nil)

				Expect(err).NotTo(HaveOccurred())

				Expect(fakeStore.StoreTerraformDeploymentCallCount()).To(Equal(1))
				storedDeployment := fakeStore.StoreTerraformDeploymentArgsForCall(0)
				Expect(storedDeployment.ID).To(Equal(existingDeployment.ID))
				Expect(storedDeployment.Workspace).To(Equal(existingDeployment.Workspace))
				Expect(storedDeployment.LastOperationType).To(Equal(existingDeployment.LastOperationType))
				Expect(storedDeployment.LastOperationState).To(Equal("succeeded"))
				Expect(storedDeployment.LastOperationMessage).To(Equal("provision succeeded"))
				Expect(fakeLogger.Errors).To(BeEmpty())
			})

			It("sets the last operation message from the TF output status", func() {
				fakeWorkspace.OutputsReturns(map[string]any{"status": "apply completed successfully"}, nil)

				err := deploymentManager.MarkOperationFinished(&existingDeployment, nil)

				Expect(err).NotTo(HaveOccurred())

				Expect(fakeStore.StoreTerraformDeploymentCallCount()).To(Equal(1))
				storedDeployment := fakeStore.StoreTerraformDeploymentArgsForCall(0)
				Expect(storedDeployment.LastOperationState).To(Equal("succeeded"))
				Expect(storedDeployment.LastOperationMessage).To(Equal("provision succeeded: apply completed successfully"))
				Expect(fakeLogger.Logs()).To(HaveLen(1))
				Expect(fakeLogger.Logs()[0].Message).To(Equal("broker.successfully stored state for deploymentID"))
			})
		})

		When("operation finished with an error", func() {
			It("sets operation state to failed, logs and stores the error", func() {
				err := deploymentManager.MarkOperationFinished(&existingDeployment, errors.New("operation failed dramatically"))

				Expect(err).NotTo(HaveOccurred())

				Expect(fakeStore.StoreTerraformDeploymentCallCount()).To(Equal(1))
				storedDeployment := fakeStore.StoreTerraformDeploymentArgsForCall(0)
				Expect(storedDeployment.ID).To(Equal(existingDeployment.ID))
				Expect(storedDeployment.Workspace).To(Equal(existingDeployment.Workspace))
				Expect(storedDeployment.LastOperationType).To(Equal(existingDeployment.LastOperationType))
				Expect(storedDeployment.LastOperationState).To(Equal("failed"))
				Expect(storedDeployment.LastOperationMessage).To(Equal("provision failed: operation failed dramatically"))
				Expect(fakeLogger.Logs()).To(HaveLen(2))
				Expect(fakeLogger.Logs()[0].Message).To(ContainSubstring("operation-failed"))
				Expect(fakeLogger.Logs()[0].Data).To(HaveKeyWithValue("error", Equal("operation failed dramatically")))
				Expect(fakeLogger.Logs()[0].Data).To(HaveKeyWithValue("message", Equal("provision failed: operation failed dramatically")))
				Expect(fakeLogger.Logs()[0].Data).To(HaveKeyWithValue("deploymentID", Equal(existingDeployment.ID)))
				Expect(fakeLogger.Logs()[1].Message).To(Equal("broker.successfully stored state for deploymentID"))
			})
		})
	})

	Describe("OperationStatus", func() {
		var (
			fakeStore          brokerfakes.FakeServiceProviderStorage
			deploymentManager  *tf.DeploymentManager
			existingDeployment storage.TerraformDeployment
		)

		const existingDeploymentID = "tf:instance:binding"

		BeforeEach(func() {
			fakeStore = brokerfakes.FakeServiceProviderStorage{}
			deploymentManager = tf.NewDeploymentManager(&fakeStore, lagertest.NewTestLogger("test"))
		})

		When("last operation has succeeded", func() {
			It("reports completion and last operation message", func() {
				existingDeployment = storage.TerraformDeployment{
					ID:                   existingDeploymentID,
					LastOperationType:    "update",
					LastOperationState:   "succeeded",
					LastOperationMessage: "great update",
				}
				fakeStore.GetTerraformDeploymentReturns(existingDeployment, nil)

				completed, lastOpMessage, err := deploymentManager.OperationStatus(existingDeploymentID)

				Expect(err).NotTo(HaveOccurred())
				Expect(completed).To(BeTrue())
				Expect(lastOpMessage).To(Equal("great update"))
			})
		})

		When("last operation has failed", func() {
			It("reports completion and errors", func() {
				existingDeployment = storage.TerraformDeployment{
					ID:                   existingDeploymentID,
					LastOperationType:    "update",
					LastOperationState:   "failed",
					LastOperationMessage: "not so great update",
				}
				fakeStore.GetTerraformDeploymentReturns(existingDeployment, nil)

				completed, lastOpMessage, err := deploymentManager.OperationStatus(existingDeploymentID)

				Expect(err).To(MatchError("not so great update"))
				Expect(completed).To(BeTrue())
				Expect(lastOpMessage).To(Equal("not so great update"))
			})
		})

		When("last operation is in progress", func() {
			It("reports in progress and last operation message", func() {
				existingDeployment = storage.TerraformDeployment{
					ID:                   existingDeploymentID,
					LastOperationType:    "update",
					LastOperationMessage: "still doing stuff",
				}
				fakeStore.GetTerraformDeploymentReturns(existingDeployment, nil)

				completed, lastOpMessage, err := deploymentManager.OperationStatus(existingDeploymentID)

				Expect(err).NotTo(HaveOccurred())
				Expect(completed).To(BeFalse())
				Expect(lastOpMessage).To(Equal("still doing stuff"))
			})
		})

		It("fails, when it errors getting the deployment", func() {
			fakeStore.GetTerraformDeploymentReturns(storage.TerraformDeployment{}, errors.New("cant get it now"))

			_, _, err := deploymentManager.OperationStatus(existingDeploymentID)

			Expect(err).To(MatchError("cant get it now"))
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
			templateVars             map[string]any
			store                    *brokerfakes.FakeServiceProviderStorage
			deploymentManager        *tf.DeploymentManager
			updatedProvisionSettings tf.TfServiceDefinitionV1Action
		)

		BeforeEach(func() {
			By("setting up fakes", func() {
				viper.Reset()
				store = &brokerfakes.FakeServiceProviderStorage{}
				deploymentManager = tf.NewDeploymentManager(store, lagertest.NewTestLogger("test"))
				templateVars = map[string]any{}
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
				viper.Set(string(featureflags.DynamicHCLEnabled), true)
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
						Configuration: map[string]any{
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
					Expect(err).To(MatchError("deployment store failed: fake error"))
				})
			})
		})

		When("terraform upgrades enabled", func() {
			BeforeEach(func() {
				viper.Set(string(featureflags.TfUpgradeEnabled), true)
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
						Configuration: map[string]any{
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
					Expect(err).To(MatchError("deployment store failed: fake error"))
				})
			})
		})

		When("brokerpak updates and terraform upgrades disabled", func() {
			It("does not update the store", func() {
				err := deploymentManager.UpdateWorkspaceHCL(id, updatedProvisionSettings, templateVars)
				Expect(err).NotTo(HaveOccurred())

				Expect(store.StoreTerraformDeploymentCallCount()).To(BeZero())
			})
		})
	})

	Describe("GetTerraformDeployment", func() {
		var (
			fakeStore          brokerfakes.FakeServiceProviderStorage
			deploymentManager  *tf.DeploymentManager
			existingDeployment storage.TerraformDeployment
		)

		const existingDeploymentID = "tf:instance:binding"

		BeforeEach(func() {
			fakeStore = brokerfakes.FakeServiceProviderStorage{}
			deploymentManager = tf.NewDeploymentManager(&fakeStore, lagertest.NewTestLogger("test"))
			existingDeployment = storage.TerraformDeployment{
				ID:                existingDeploymentID,
				LastOperationType: "validation",
			}
		})

		It("get the terraform deployment", func() {
			fakeStore.GetTerraformDeploymentReturns(existingDeployment, nil)

			deployment, err := deploymentManager.GetTerraformDeployment(existingDeploymentID)

			Expect(err).NotTo(HaveOccurred())
			Expect(deployment).To(Equal(existingDeployment))
		})

		It("fails, when getting terraform deployment fails", func() {
			fakeStore.GetTerraformDeploymentReturns(storage.TerraformDeployment{}, errors.New("cant get it now"))

			_, err := deploymentManager.GetTerraformDeployment(existingDeploymentID)

			Expect(err).To(MatchError("cant get it now"))
		})
	})

	Describe("GetBindingDeployments", func() {
		var (
			fakeStore         brokerfakes.FakeServiceProviderStorage
			deploymentManager *tf.DeploymentManager
		)

		const existingInstanceDeploymentID = "tf:instance-guid:"

		BeforeEach(func() {
			fakeStore = brokerfakes.FakeServiceProviderStorage{}
			deploymentManager = tf.NewDeploymentManager(&fakeStore, lagertest.NewTestLogger("test"))
		})

		It("gets all binding deployments for a service instance", func() {
			fakeStore.GetServiceBindingIDsForServiceInstanceReturns([]string{"first-binding-guid", "second-binding-guid"}, nil)
			firstBindingDeployment := storage.TerraformDeployment{
				ID: "tf:instance-guid:first-binding-guid",
			}
			secondBindingDeployment := storage.TerraformDeployment{
				ID: "tf:instance-guid:second-binding-guid",
			}
			fakeStore.GetTerraformDeploymentReturnsOnCall(0, firstBindingDeployment, nil)
			fakeStore.GetTerraformDeploymentReturnsOnCall(1, secondBindingDeployment, nil)

			deployments, err := deploymentManager.GetBindingDeployments(existingInstanceDeploymentID)

			Expect(err).NotTo(HaveOccurred())
			Expect(len(deployments)).To(Equal(2))
			Expect(deployments[0]).To(Equal(firstBindingDeployment))
			Expect(deployments[1]).To(Equal(secondBindingDeployment))

			Expect(fakeStore.GetTerraformDeploymentCallCount()).To(Equal(2))
			Expect(fakeStore.GetTerraformDeploymentArgsForCall(0)).To(Equal("tf:instance-guid:first-binding-guid"))
			Expect(fakeStore.GetTerraformDeploymentArgsForCall(1)).To(Equal("tf:instance-guid:second-binding-guid"))
		})

		It("fails, when getting service bindings fails", func() {
			fakeStore.GetServiceBindingIDsForServiceInstanceReturns([]string{}, errors.New("cant get it now"))

			_, err := deploymentManager.GetBindingDeployments(existingInstanceDeploymentID)

			Expect(err).To(MatchError("cant get it now"))
		})

		It("fails, when getting a terraform deployment fails", func() {
			fakeStore.GetServiceBindingIDsForServiceInstanceReturns([]string{"first-binding-guid", "second-binding-guid"}, nil)
			fakeStore.GetTerraformDeploymentReturns(storage.TerraformDeployment{}, errors.New("cant get it now"))

			_, err := deploymentManager.GetBindingDeployments(existingInstanceDeploymentID)

			Expect(err).To(MatchError("cant get it now"))
		})
	})
})
