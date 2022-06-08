package tf_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/executor"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/tffakes"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/workspace"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/workspace/workspacefakes"
	"github.com/cloudfoundry/cloud-service-broker/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/utils"
	"github.com/hashicorp/go-version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Upgrade", func() {
	const instanceDeploymentID = "tf:instance-ID:"

	var (
		fakeDeploymentManager *tffakes.FakeDeploymentManagerInterface
		fakeWorkspace         *workspacefakes.FakeWorkspace
		instanceTFDeployment  storage.TerraformDeployment
		genericError          = fmt.Errorf("genericError")
		fakeInvokerBuilder    *tffakes.FakeTerraformInvokerBuilder
		fakeDefaultInvoker    *tffakes.FakeTerraformInvoker
		fakeLogger            = utils.NewLogger("test")
		provisionAction       = tf.TfServiceDefinitionV1Action{
			Template: "provision-template",
		}
		bindAction = tf.TfServiceDefinitionV1Action{
			Template: "bind-template",
		}
		fakeServiceDefinition = tf.TfServiceDefinitionV1{
			ProvisionSettings: provisionAction,
			BindSettings:      bindAction,
		}
		instanceVarContext   *varcontext.VarContext
		instanceTemplateVars = map[string]interface{}{"tf_id": instanceDeploymentID, "var": "value"}
	)

	BeforeEach(func() {
		fakeDeploymentManager = &tffakes.FakeDeploymentManagerInterface{}
		fakeWorkspace = &workspacefakes.FakeWorkspace{}
		fakeWorkspace.ModuleInstancesReturns([]workspace.ModuleInstance{{ModuleName: "moduleName"}})
		fakeInvokerBuilder = &tffakes.FakeTerraformInvokerBuilder{}
		fakeDefaultInvoker = &tffakes.FakeTerraformInvoker{}
		instanceTFDeployment = storage.TerraformDeployment{
			ID: instanceDeploymentID,
			Workspace: &workspace.TerraformWorkspace{
				Modules: []workspace.ModuleDefinition{
					{Name: "test"},
				},
			},
		}

		var err error
		instanceVarContext, err = varcontext.Builder().MergeMap(instanceTemplateVars).Build()
		Expect(err).NotTo(HaveOccurred())
	})

	When("an upgrade path is defined", func() {
		var (
			fakeInvoker1 = &tffakes.FakeTerraformInvoker{}
			fakeInvoker2 = &tffakes.FakeTerraformInvoker{}
		)

		BeforeEach(func() {
			instanceTFDeployment.Workspace = fakeWorkspace
			fakeDeploymentManager.GetTerraformDeploymentReturns(instanceTFDeployment, nil)
			fakeInvokerBuilder.VersionedTerraformInvokerReturnsOnCall(0, fakeInvoker1)
			fakeInvokerBuilder.VersionedTerraformInvokerReturnsOnCall(1, fakeInvoker2)
			fakeInvokerBuilder.VersionedTerraformInvokerReturnsOnCall(2, fakeDefaultInvoker)

			fakeWorkspace.StateVersionReturns(newVersion("0.0.1"), nil)
			fakeWorkspace.ModuleInstancesReturns([]workspace.ModuleInstance{{ModuleName: "moduleName"}})
		})

		It("runs apply with all tf versions in the upgrade path", func() {
			tfBinContext := executor.TFBinariesContext{
				DefaultTfVersion: newVersion("0.1.0"),
				TfUpgradePath: []*version.Version{
					newVersion("0.0.1"),
					newVersion("0.0.2"),
					newVersion("0.1.0"),
				},
			}

			provider := tf.NewTerraformProvider(tfBinContext, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)
			_, err := provider.Upgrade(context.TODO(), instanceVarContext, nil)
			Expect(err).NotTo(HaveOccurred())

			Eventually(operationWasFinishedForDeployment(fakeDeploymentManager)).Should(Equal(instanceTFDeployment))
			Expect(operationWasFinishedWithError(fakeDeploymentManager)()).To(BeNil())

			Expect(fakeInvoker1.ApplyCallCount()).To(Equal(1))
			Expect(getWorkspace(fakeInvoker1, 0)).To(Equal(fakeWorkspace))

			Expect(fakeInvoker2.ApplyCallCount()).To(Equal(1))
			Expect(getWorkspace(fakeInvoker2, 0)).To(Equal(fakeWorkspace))

			Expect(fakeInvokerBuilder.VersionedTerraformInvokerCallCount()).To(Equal(2))
			Expect(fakeInvokerBuilder.VersionedTerraformInvokerArgsForCall(0)).To(Equal(newVersion("0.0.2")))
			Expect(fakeInvokerBuilder.VersionedTerraformInvokerArgsForCall(1)).To(Equal(newVersion("0.1.0")))

			Expect(fakeDeploymentManager.UpdateWorkspaceHCLCallCount()).To(Equal(1))
			actualDeploymentID, actualAction, actualUpgradeContext := fakeDeploymentManager.UpdateWorkspaceHCLArgsForCall(0)
			Expect(actualDeploymentID).To(Equal(instanceTFDeployment.ID))
			Expect(actualAction).To(Equal(provisionAction))
			Expect(actualUpgradeContext).To(Equal(instanceTemplateVars))
		})
	})

	When("instance has bindings", func() {
		var (
			firstBindingID          = "firstBindingID"
			secondBindingID         = "secondBindingID"
			firstBindingWorkspace   = &workspacefakes.FakeWorkspace{}
			firstBindingDeployment  = storage.TerraformDeployment{ID: instanceDeploymentID + firstBindingID, Workspace: firstBindingWorkspace}
			secondBindingWorkspace  = &workspacefakes.FakeWorkspace{}
			secondBindingDeployment = storage.TerraformDeployment{ID: instanceDeploymentID + secondBindingID, Workspace: secondBindingWorkspace}

			bindingDeployments = []storage.TerraformDeployment{
				firstBindingDeployment,
				secondBindingDeployment,
			}
			firstBindingVars    = map[string]interface{}{"first-binding-var": "first-binding-value"}
			secondBindingVars   = map[string]interface{}{"second-binding-var": "second-binding-value"}
			bindingsVarContexts map[string]*varcontext.VarContext

			fakeInvoker1 = &tffakes.FakeTerraformInvoker{}
			fakeInvoker2 = &tffakes.FakeTerraformInvoker{}
			fakeInvoker3 = &tffakes.FakeTerraformInvoker{}
			fakeInvoker4 = &tffakes.FakeTerraformInvoker{}
			fakeInvoker5 = &tffakes.FakeTerraformInvoker{}
			fakeInvoker6 = &tffakes.FakeTerraformInvoker{}
		)
		BeforeEach(func() {
			instanceTFDeployment.Workspace = fakeWorkspace
			fakeDeploymentManager.GetTerraformDeploymentReturns(instanceTFDeployment, nil)
			fakeInvokerBuilder.VersionedTerraformInvokerReturnsOnCall(0, fakeInvoker1)
			fakeInvokerBuilder.VersionedTerraformInvokerReturnsOnCall(1, fakeInvoker2)

			fakeWorkspace.StateVersionReturns(newVersion("0.0.1"), nil)
			fakeWorkspace.ModuleInstancesReturns([]workspace.ModuleInstance{{ModuleName: "instance-moduleName"}})

			fakeDeploymentManager.GetBindingDeploymentsReturns(bindingDeployments, nil)
			fakeInvokerBuilder.VersionedTerraformInvokerReturnsOnCall(2, fakeInvoker3)
			fakeInvokerBuilder.VersionedTerraformInvokerReturnsOnCall(3, fakeInvoker4)
			fakeInvokerBuilder.VersionedTerraformInvokerReturnsOnCall(4, fakeInvoker5)
			fakeInvokerBuilder.VersionedTerraformInvokerReturnsOnCall(5, fakeInvoker6)

			firstBindingWorkspace.StateVersionReturns(newVersion("0.0.1"), nil)
			firstBindingWorkspace.ModuleInstancesReturns([]workspace.ModuleInstance{{ModuleName: "first-binding-moduleName"}})

			secondBindingWorkspace.StateVersionReturns(newVersion("0.0.1"), nil)
			secondBindingWorkspace.ModuleInstancesReturns([]workspace.ModuleInstance{{ModuleName: "second-binding-moduleName"}})

			var err error
			firstBindingVarContext, err := varcontext.Builder().MergeMap(firstBindingVars).Build()
			Expect(err).NotTo(HaveOccurred())
			secondBindingVarContext, err := varcontext.Builder().MergeMap(secondBindingVars).Build()
			Expect(err).NotTo(HaveOccurred())

			bindingsVarContexts = map[string]*varcontext.VarContext{
				firstBindingID:  firstBindingVarContext,
				secondBindingID: secondBindingVarContext,
			}
		})

		It("upgrades all the available bindings to latest version", func() {
			tfBinContext := executor.TFBinariesContext{
				DefaultTfVersion: newVersion("0.1.0"),
				TfUpgradePath: []*version.Version{
					newVersion("0.0.1"),
					newVersion("0.0.2"),
					newVersion("0.1.0"),
				},
			}
			instanceTemplateVars = map[string]interface{}{"tf_id": instanceDeploymentID, "var": "value"}

			provider := tf.NewTerraformProvider(tfBinContext, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)
			_, err := provider.Upgrade(context.TODO(), instanceVarContext, bindingsVarContexts)
			Expect(err).NotTo(HaveOccurred())

			Eventually(operationWasFinishedForDeployment(fakeDeploymentManager)).Should(Equal(instanceTFDeployment))
			Expect(operationWasFinishedWithError(fakeDeploymentManager)()).To(BeNil())

			By("checking the binding operations were also updated")
			Expect(fakeDeploymentManager.MarkOperationFinishedCallCount()).To(Equal(3))
			actualFirstBindingDeployment, _ := fakeDeploymentManager.MarkOperationFinishedArgsForCall(0)
			Expect(actualFirstBindingDeployment.ID).To(BeElementOf([]string{firstBindingDeployment.ID, secondBindingDeployment.ID}))
			actualSecondBindingDeployment, _ := fakeDeploymentManager.MarkOperationFinishedArgsForCall(1)
			Expect(actualSecondBindingDeployment.ID).To(BeElementOf([]string{firstBindingDeployment.ID, secondBindingDeployment.ID}))

			By("checking the invoker was called for the service instance with correct workspace")
			Expect(fakeInvoker1.ApplyCallCount()).To(Equal(1))
			Expect(getWorkspace(fakeInvoker1, 0)).To(Equal(fakeWorkspace))

			Expect(fakeInvoker2.ApplyCallCount()).To(Equal(1))
			Expect(getWorkspace(fakeInvoker2, 0)).To(Equal(fakeWorkspace))

			By("checking the invoker was called for the first binding with correct workspace")
			Expect(fakeInvoker3.ApplyCallCount()).To(Equal(1))
			Expect(getWorkspace(fakeInvoker3, 0)).To(Equal(firstBindingWorkspace))

			Expect(fakeInvoker4.ApplyCallCount()).To(Equal(1))
			Expect(getWorkspace(fakeInvoker4, 0)).To(Equal(firstBindingWorkspace))

			By("checking the invoker was called for the second binding with correct workspace")
			Expect(fakeInvoker5.ApplyCallCount()).To(Equal(1))
			Expect(getWorkspace(fakeInvoker5, 0)).To(Equal(secondBindingWorkspace))

			Expect(fakeInvoker6.ApplyCallCount()).To(Equal(1))
			Expect(getWorkspace(fakeInvoker6, 0)).To(Equal(secondBindingWorkspace))

			Expect(fakeInvokerBuilder.VersionedTerraformInvokerCallCount()).To(Equal(6))
			Expect(fakeInvokerBuilder.VersionedTerraformInvokerArgsForCall(0)).To(Equal(newVersion("0.0.2")))
			Expect(fakeInvokerBuilder.VersionedTerraformInvokerArgsForCall(1)).To(Equal(newVersion("0.1.0")))
			Expect(fakeInvokerBuilder.VersionedTerraformInvokerArgsForCall(2)).To(Equal(newVersion("0.0.2")))
			Expect(fakeInvokerBuilder.VersionedTerraformInvokerArgsForCall(3)).To(Equal(newVersion("0.1.0")))
			Expect(fakeInvokerBuilder.VersionedTerraformInvokerArgsForCall(4)).To(Equal(newVersion("0.0.2")))
			Expect(fakeInvokerBuilder.VersionedTerraformInvokerArgsForCall(5)).To(Equal(newVersion("0.1.0")))

			By("ensuring the HCL was updated")
			Expect(fakeDeploymentManager.UpdateWorkspaceHCLCallCount()).To(Equal(3))
			actualDeploymentID, actualAction, actualUpgradeContext := fakeDeploymentManager.UpdateWorkspaceHCLArgsForCall(0)
			Expect(actualDeploymentID).To(Equal(instanceTFDeployment.ID))
			Expect(actualAction).To(Equal(provisionAction))
			Expect(actualUpgradeContext).To(Equal(instanceTemplateVars))

			actualFirstBindingDeploymentID, actualFirstBindingAction, actualFirstBindingUpgradeContext := fakeDeploymentManager.UpdateWorkspaceHCLArgsForCall(1)
			Expect(actualFirstBindingDeploymentID).To(BeElementOf([]string{firstBindingDeployment.ID, secondBindingDeployment.ID}))
			Expect(actualFirstBindingAction).To(Equal(bindAction))
			Expect(actualFirstBindingUpgradeContext).To(BeElementOf([]map[string]interface{}{firstBindingVars, secondBindingVars}))

			actualSecondBindingDeploymentID, actualSecondBindingAction, actualSecondBindingUpgradeContext := fakeDeploymentManager.UpdateWorkspaceHCLArgsForCall(2)
			Expect(actualSecondBindingDeploymentID).To(BeElementOf([]string{firstBindingDeployment.ID, secondBindingDeployment.ID}))
			Expect(actualSecondBindingAction).To(Equal(bindAction))
			Expect(actualSecondBindingUpgradeContext).To(BeElementOf([]map[string]interface{}{firstBindingVars, secondBindingVars}))
		})

		When("an apply fails for a binding", func() {
			var (
				fakeInvokerBind = &tffakes.FakeTerraformInvoker{}
			)

			BeforeEach(func() {
				fakeInvokerBuilder.VersionedTerraformInvokerReturnsOnCall(2, fakeInvokerBind)
				fakeInvokerBind.ApplyReturnsOnCall(0, genericError)
			})

			It("fails the upgrade, if an apply fails for a binding", func() {
				tfBinContext := executor.TFBinariesContext{
					DefaultTfVersion: newVersion("0.1.0"),
					TfUpgradePath: []*version.Version{
						newVersion("0.0.1"),
						newVersion("0.0.2"),
						newVersion("0.1.0"),
					},
				}

				provider := tf.NewTerraformProvider(tfBinContext, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

				_, err := provider.Upgrade(context.TODO(), instanceVarContext, bindingsVarContexts)
				Expect(err).NotTo(HaveOccurred())

				Eventually(operationWasFinishedForDeployment(fakeDeploymentManager)).Should(Equal(instanceTFDeployment))
				Expect(operationWasFinishedWithError(fakeDeploymentManager)()).To(MatchError(genericError))

				By("checking the binding operations were also finished with error")
				Expect(fakeDeploymentManager.MarkOperationFinishedCallCount()).To(Equal(2))
				actualFirstBindingDeployment, err := fakeDeploymentManager.MarkOperationFinishedArgsForCall(0)
				Expect(actualFirstBindingDeployment.ID).To(BeElementOf([]string{firstBindingDeployment.ID, secondBindingDeployment.ID}))
				Expect(err).To(MatchError(genericError))
			})
		})
	})

	It("fails the upgrade, if the version of statefile does not match the default tf version, and no upgrade path is specified", func() {
		tfBinContext := executor.TFBinariesContext{
			DefaultTfVersion: newVersion("0.1.0"),
		}
		instanceTFDeployment.Workspace = fakeWorkspace
		fakeDeploymentManager.GetTerraformDeploymentReturns(instanceTFDeployment, nil)
		fakeWorkspace.StateVersionReturns(newVersion("0.0.1"), nil)

		provider := tf.NewTerraformProvider(tfBinContext, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

		_, err := provider.Upgrade(context.TODO(), instanceVarContext, nil)
		Expect(err).NotTo(HaveOccurred())

		Eventually(operationWasFinishedForDeployment(fakeDeploymentManager)).Should(Equal(instanceTFDeployment))
		Expect(operationWasFinishedWithError(fakeDeploymentManager)()).To(MatchError("terraform version mismatch and no upgrade path specified"))
		Expect(fakeInvokerBuilder.VersionedTerraformInvokerCallCount()).To(Equal(0))
	})

	It("fails the upgrade, if an apply fails for the instance", func() {
		tfBinContext := executor.TFBinariesContext{
			DefaultTfVersion: newVersion("0.1.0"),
			TfUpgradePath: []*version.Version{
				newVersion("0.0.1"),
				newVersion("0.0.2"),
				newVersion("0.1.0"),
			},
		}
		fakeInvoker1 := &tffakes.FakeTerraformInvoker{}
		fakeInvoker2 := &tffakes.FakeTerraformInvoker{}

		instanceTFDeployment.Workspace = fakeWorkspace
		fakeDeploymentManager.GetTerraformDeploymentReturns(instanceTFDeployment, nil)
		fakeInvokerBuilder.VersionedTerraformInvokerReturnsOnCall(0, fakeInvoker1)
		fakeInvoker1.ApplyReturnsOnCall(0, genericError)
		fakeInvokerBuilder.VersionedTerraformInvokerReturnsOnCall(1, fakeInvoker2)

		fakeWorkspace.StateVersionReturns(newVersion("0.0.1"), nil)
		fakeWorkspace.ModuleInstancesReturns([]workspace.ModuleInstance{{ModuleName: "moduleName"}})

		provider := tf.NewTerraformProvider(tfBinContext, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)
		_, err := provider.Upgrade(context.TODO(), instanceVarContext, nil)
		Expect(err).NotTo(HaveOccurred())

		Eventually(operationWasFinishedForDeployment(fakeDeploymentManager)).Should(Equal(instanceTFDeployment))
		Expect(operationWasFinishedWithError(fakeDeploymentManager)()).To(MatchError(genericError))

		Expect(fakeInvoker1.ApplyCallCount()).To(Equal(1))
		Expect(getWorkspace(fakeInvoker1, 0)).To(Equal(fakeWorkspace))

		Expect(fakeInvoker2.ApplyCallCount()).To(Equal(0))
	})

	It("fails, if getting bindings deployments fails", func() {
		fakeDeploymentManager.GetTerraformDeploymentReturns(instanceTFDeployment, nil)
		fakeDeploymentManager.GetBindingDeploymentsReturns([]storage.TerraformDeployment{}, errors.New("error getting bindings"))

		provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)
		_, err := provider.Upgrade(context.TODO(), instanceVarContext, nil)
		Expect(err).To(MatchError("error getting bindings"))
	})

	When("it fails to upgrade the instance", func() {
		It("fails", func() {
			instanceTFDeployment.Workspace = fakeWorkspace
			fakeDeploymentManager.GetTerraformDeploymentReturns(instanceTFDeployment, nil)
			fakeWorkspace.StateVersionReturns(nil, genericError)

			provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)
			_, err := provider.Upgrade(context.TODO(), instanceVarContext, nil)
			Expect(err).NotTo(HaveOccurred())

			Eventually(operationWasFinishedForDeployment(fakeDeploymentManager)).Should(Equal(instanceTFDeployment))
			Expect(operationWasFinishedWithError(fakeDeploymentManager)()).To(MatchError(genericError))
		})
	})

	When("upgrade context is missing tf_id", func() {
		It("fails", func() {
			varContext, err := varcontext.Builder().Build()
			Expect(err).NotTo(HaveOccurred())

			provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

			_, err = provider.Upgrade(context.TODO(), varContext, nil)
			Expect(err).To(MatchError(`1 error(s) occurred: missing value for key "tf_id"`))
		})
	})

	When("updating workspace HCL errors", func() {
		It("fails", func() {
			fakeDeploymentManager.UpdateWorkspaceHCLReturns(genericError)

			provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

			_, err := provider.Upgrade(context.TODO(), instanceVarContext, nil)
			Expect(err).To(MatchError(genericError))
		})
	})

	When("getting deployment errors", func() {
		It("fails", func() {
			fakeDeploymentManager.GetTerraformDeploymentReturns(storage.TerraformDeployment{}, genericError)

			provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

			_, err := provider.Upgrade(context.TODO(), instanceVarContext, nil)
			Expect(err).To(MatchError(genericError))
		})
	})

	When("it errors while marking instance operation as started", func() {
		It("fails", func() {
			fakeDeploymentManager.MarkOperationStartedReturns(genericError)

			provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

			_, err := provider.Upgrade(context.TODO(), instanceVarContext, nil)
			Expect(err).To(MatchError(genericError))
		})
	})
})
