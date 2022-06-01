package tf_test

import (
	"context"
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
	var (
		fakeDeploymentManager *tffakes.FakeDeploymentManagerInterface
		fakeWorkspace         *workspacefakes.FakeWorkspace
		deploymentID          string
		deployment            storage.TerraformDeployment
		genericError          = fmt.Errorf("genericError")
		fakeInvokerBuilder    *tffakes.FakeTerraformInvokerBuilder
		fakeDefaultInvoker    *tffakes.FakeTerraformInvoker
		fakeLogger            = utils.NewLogger("test")
		fakeServiceDefinition = tf.TfServiceDefinitionV1{}
		varContext            *varcontext.VarContext
		templateVars          = map[string]interface{}{"tf_id": "567c6af0-d68a-11ec-a5b6-367dda7ea869", "var": "value"}
	)

	BeforeEach(func() {
		fakeDeploymentManager = &tffakes.FakeDeploymentManagerInterface{}
		fakeWorkspace = &workspacefakes.FakeWorkspace{}
		fakeWorkspace.ModuleInstancesReturns([]workspace.ModuleInstance{{ModuleName: "moduleName"}})
		fakeInvokerBuilder = &tffakes.FakeTerraformInvokerBuilder{}
		fakeDefaultInvoker = &tffakes.FakeTerraformInvoker{}
		deploymentID = "deploymentID"
		deployment = storage.TerraformDeployment{
			ID: deploymentID,
			Workspace: &workspace.TerraformWorkspace{
				Modules: []workspace.ModuleDefinition{
					{Name: "test"},
				},
			},
		}

		var err error
		varContext, err = varcontext.Builder().MergeMap(templateVars).Build()
		Expect(err).NotTo(HaveOccurred())
	})

	When("an upgrade path is defined", func() {
		var (
			fakeInvoker1 = &tffakes.FakeTerraformInvoker{}
			fakeInvoker2 = &tffakes.FakeTerraformInvoker{}
		)

		BeforeEach(func() {
			deployment.Workspace = fakeWorkspace
			fakeDeploymentManager.GetTerraformDeploymentReturns(deployment, nil)
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
			_, err := provider.Upgrade(context.TODO(), varContext)
			Expect(err).NotTo(HaveOccurred())

			Eventually(operationWasFinishedForDeployment(fakeDeploymentManager)).Should(Equal(deployment))
			Expect(operationWasFinishedWithError(fakeDeploymentManager)()).To(BeNil())

			Expect(fakeInvoker1.ApplyCallCount()).To(Equal(1))
			Expect(getWorkspace(fakeInvoker1, 0)).To(Equal(fakeWorkspace))

			Expect(fakeInvoker2.ApplyCallCount()).To(Equal(1))
			Expect(getWorkspace(fakeInvoker2, 0)).To(Equal(fakeWorkspace))

			Expect(fakeInvokerBuilder.VersionedTerraformInvokerCallCount()).To(Equal(2))
			Expect(fakeInvokerBuilder.VersionedTerraformInvokerArgsForCall(0)).To(Equal(newVersion("0.0.2")))
			Expect(fakeInvokerBuilder.VersionedTerraformInvokerArgsForCall(1)).To(Equal(newVersion("0.1.0")))
		})
	})

	When("instance has bindings", func() {
		var (
			firstBindingWorkspace   = &workspacefakes.FakeWorkspace{}
			firstBindingDeployment  = storage.TerraformDeployment{ID: "first-bindingID", Workspace: firstBindingWorkspace}
			secondBindingWorkspace  = &workspacefakes.FakeWorkspace{}
			secondBindingDeployment = storage.TerraformDeployment{ID: "second-bindingID", Workspace: secondBindingWorkspace}
			bindingDeployments      = []storage.TerraformDeployment{
				firstBindingDeployment,
				secondBindingDeployment,
			}
			fakeInvoker1 = &tffakes.FakeTerraformInvoker{}
			fakeInvoker2 = &tffakes.FakeTerraformInvoker{}
			fakeInvoker3 = &tffakes.FakeTerraformInvoker{}
			fakeInvoker4 = &tffakes.FakeTerraformInvoker{}
			fakeInvoker5 = &tffakes.FakeTerraformInvoker{}
			fakeInvoker6 = &tffakes.FakeTerraformInvoker{}
		)

		BeforeEach(func() {
			deployment.Workspace = fakeWorkspace
			fakeDeploymentManager.GetTerraformDeploymentReturns(deployment, nil)
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

			provider := tf.NewTerraformProvider(tfBinContext, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)
			_, err := provider.Upgrade(context.TODO(), varContext)
			Expect(err).NotTo(HaveOccurred())

			Eventually(operationWasFinishedForDeployment(fakeDeploymentManager)).Should(Equal(deployment))
			Expect(operationWasFinishedWithError(fakeDeploymentManager)()).To(BeNil())

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
		})
	})

	It("fails the upgrade, if the version of statefile does not match the default tf version, and no upgrade path is specified", func() {
		tfBinContext := executor.TFBinariesContext{
			DefaultTfVersion: newVersion("0.1.0"),
		}
		deployment.Workspace = fakeWorkspace
		fakeDeploymentManager.GetTerraformDeploymentReturns(deployment, nil)
		fakeWorkspace.StateVersionReturns(newVersion("0.0.1"), nil)

		provider := tf.NewTerraformProvider(tfBinContext, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

		_, err := provider.Upgrade(context.TODO(), varContext)
		Expect(err).NotTo(HaveOccurred())

		Eventually(operationWasFinishedForDeployment(fakeDeploymentManager)).Should(Equal(deployment))
		Expect(operationWasFinishedWithError(fakeDeploymentManager)()).To(MatchError("terraform version mismatch and no upgrade path specified"))
		Expect(fakeInvokerBuilder.VersionedTerraformInvokerCallCount()).To(Equal(0))
	})

	It("fails the upgrade, if an apply fails", func() {
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

		deployment.Workspace = fakeWorkspace
		fakeDeploymentManager.GetTerraformDeploymentReturns(deployment, nil)
		fakeInvokerBuilder.VersionedTerraformInvokerReturnsOnCall(0, fakeInvoker1)
		fakeInvoker1.ApplyReturns(genericError)
		fakeInvokerBuilder.VersionedTerraformInvokerReturnsOnCall(1, fakeInvoker2)

		fakeWorkspace.StateVersionReturns(newVersion("0.0.1"), nil)
		fakeWorkspace.ModuleInstancesReturns([]workspace.ModuleInstance{{ModuleName: "moduleName"}})

		provider := tf.NewTerraformProvider(tfBinContext, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)
		_, err := provider.Upgrade(context.TODO(), varContext)
		Expect(err).NotTo(HaveOccurred())

		Eventually(operationWasFinishedForDeployment(fakeDeploymentManager)).Should(Equal(deployment))
		Expect(operationWasFinishedWithError(fakeDeploymentManager)()).To(MatchError(genericError))

		Expect(fakeInvoker1.ApplyCallCount()).To(Equal(1))
		Expect(getWorkspace(fakeInvoker1, 0)).To(Equal(fakeWorkspace))

		Expect(fakeInvoker2.ApplyCallCount()).To(Equal(0))
	})

	When("can't get terraform version from state", func() {
		It("fails", func() {
			deployment.Workspace = fakeWorkspace
			fakeDeploymentManager.GetTerraformDeploymentReturns(deployment, nil)
			fakeWorkspace.StateVersionReturns(nil, genericError)

			provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)
			_, err := provider.Upgrade(context.TODO(), varContext)
			Expect(err).NotTo(HaveOccurred())

			Eventually(operationWasFinishedForDeployment(fakeDeploymentManager)).Should(Equal(deployment))
			Expect(operationWasFinishedWithError(fakeDeploymentManager)()).To(MatchError(genericError))
		})
	})

	When("upgrade context is missing tf_id", func() {
		It("fails", func() {
			varContext, err := varcontext.Builder().Build()
			Expect(err).NotTo(HaveOccurred())

			provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

			_, err = provider.Upgrade(context.TODO(), varContext)
			Expect(err).To(MatchError(`1 error(s) occurred: missing value for key "tf_id"`))
		})
	})

	When("updating workspace HCL errors", func() {
		It("fails", func() {
			fakeDeploymentManager.UpdateWorkspaceHCLReturns(genericError)

			provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

			_, err := provider.Upgrade(context.TODO(), varContext)
			Expect(err).To(MatchError(genericError))
		})
	})

	When("getting deployment errors", func() {
		It("fails", func() {
			fakeDeploymentManager.GetTerraformDeploymentReturns(storage.TerraformDeployment{}, genericError)

			provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

			_, err := provider.Upgrade(context.TODO(), varContext)
			Expect(err).To(MatchError(genericError))
		})
	})

	When("it errors while marking operation as started", func() {
		It("fails", func() {
			fakeDeploymentManager.MarkOperationStartedReturns(genericError)

			provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

			_, err := provider.Upgrade(context.TODO(), varContext)
			Expect(err).To(MatchError(genericError))
		})
	})
})
