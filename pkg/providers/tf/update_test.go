package tf_test

import (
	"context"
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/pkg/broker"

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
	"github.com/spf13/viper"

	. "github.com/onsi/gomega"
)

var _ = Describe("Update", func() {
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

		viper.Reset()
	})

	When("update called on subsume plan", func() {
		It("fails", func() {
			varContext, err := varcontext.Builder().MergeMap(map[string]interface{}{"tf_id": "567c6af0-d68a-11ec-a5b6-367dda7ea869", "var": "value", "subsume": true}).Build()
			Expect(err).NotTo(HaveOccurred())
			fakeServiceDefinition.ProvisionSettings.PlanInputs = []broker.BrokerVariable{
				{
					FieldName: "subsume",
				},
			}

			provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

			_, err = provider.Update(context.TODO(), varContext)
			Expect(err).To(MatchError("cannot update to subsume plan\n\nFor OpsMan Tile users see documentation here: https://via.vmw.com/ENs4\n\nFor Open Source users deployed via 'cf push' see documentation here:  https://via.vmw.com/ENw4"))
		})
	})

	When("there is an error in provision context", func() {
		It("fails", func() {
			varContext, err := varcontext.Builder().MergeMap(map[string]interface{}{}).Build()
			Expect(err).NotTo(HaveOccurred())

			provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

			_, err = provider.Update(context.TODO(), varContext)
			Expect(err).To(MatchError(`1 error(s) occurred: missing value for key "tf_id"`))
		})
	})

	When("unable to update workspace HCL", func() {
		It("fails", func() {
			fakeDeploymentManager.UpdateWorkspaceHCLReturns(genericError)

			provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

			_, err := provider.Update(context.TODO(), varContext)
			Expect(err).To(MatchError(genericError))
		})
	})

	When("deployment can't be found", func() {
		It("fails", func() {
			fakeDeploymentManager.GetTerraformDeploymentReturns(storage.TerraformDeployment{}, genericError)
			provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

			_, err := provider.Update(context.TODO(), varContext)
			Expect(err).To(MatchError(genericError))
		})
	})

	When("job can't be marked as started", func() {
		It("fails", func() {
			fakeDeploymentManager.GetTerraformDeploymentReturns(deployment, nil)
			fakeDeploymentManager.MarkOperationStartedReturns(genericError)
			runner := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

			_, err := runner.Update(context.TODO(), varContext)

			Expect(err).To(MatchError(genericError))
		})
	})

	When("can't get terraform version from state", func() {
		It("fails", func() {
			deployment.Workspace = fakeWorkspace
			fakeDeploymentManager.GetTerraformDeploymentReturns(deployment, nil)
			fakeWorkspace.StateVersionReturns(nil, genericError)

			runner := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)
			_, err := runner.Update(context.TODO(), varContext)
			Expect(err).NotTo(HaveOccurred())

			Eventually(operationWasFinishedForDeployment(fakeDeploymentManager)).Should(Equal(deployment))
			Expect(operationWasFinishedWithError(fakeDeploymentManager)()).To(MatchError(genericError))
		})
	})

	When("unable to update instance configuration", func() {
		It("fails", func() {
			deployment.Workspace = fakeWorkspace
			fakeDeploymentManager.GetTerraformDeploymentReturns(deployment, nil)
			tfVersion := "1.1"
			fakeWorkspace.StateVersionReturns(newVersion(tfVersion), nil)
			fakeWorkspace.UpdateInstanceConfigurationReturns(genericError)

			runner := tf.NewTerraformProvider(executor.TFBinariesContext{DefaultTfVersion: newVersion(tfVersion)}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

			_, err := runner.Update(context.TODO(), varContext)
			Expect(err).NotTo(HaveOccurred())

			Eventually(operationWasFinishedForDeployment(fakeDeploymentManager)).Should(Equal(deployment))
			Expect(operationWasFinishedWithError(fakeDeploymentManager)()).To(MatchError(genericError))
		})
	})

	It("updates templates before applying", func() {
		deployment.Workspace = fakeWorkspace
		fakeDeploymentManager.GetTerraformDeploymentReturns(deployment, nil)
		tfVersion := "1.1"
		fakeWorkspace.StateVersionReturns(newVersion(tfVersion), nil)
		fakeInvokerBuilder.VersionedTerraformInvokerReturns(fakeDefaultInvoker)

		runner := tf.NewTerraformProvider(executor.TFBinariesContext{DefaultTfVersion: newVersion(tfVersion)}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

		_, err := runner.Update(context.TODO(), varContext)
		Expect(err).NotTo(HaveOccurred())
		Eventually(operationWasFinishedForDeployment(fakeDeploymentManager)).Should(Equal(deployment))
		Expect(operationWasFinishedWithError(fakeDeploymentManager)()).To(BeNil())

		Expect(fakeWorkspace.UpdateInstanceConfigurationCallCount()).To(Equal(1))
		Expect(fakeWorkspace.UpdateInstanceConfigurationArgsForCall(0)).To(Equal(templateVars))
	})

	It("updates the last operation on success", func() {
		deployment.Workspace = fakeWorkspace
		fakeDeploymentManager.GetTerraformDeploymentReturns(deployment, nil)
		tfVersion := "1.1"
		fakeWorkspace.StateVersionReturns(newVersion(tfVersion), nil)
		fakeInvokerBuilder.VersionedTerraformInvokerReturns(fakeDefaultInvoker)
		fakeWorkspace.OutputsReturns(map[string]interface{}{"status": "status from terraform"}, nil)

		runner := tf.NewTerraformProvider(executor.TFBinariesContext{DefaultTfVersion: newVersion(tfVersion)}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)
		_, err := runner.Update(context.TODO(), varContext)
		Expect(err).To(Succeed())
		Eventually(operationWasFinishedForDeployment(fakeDeploymentManager)).Should(Equal(deployment))
		Expect(operationWasFinishedWithError(fakeDeploymentManager)()).To(BeNil())
	})

	It("returns the error in last operation, if terraform apply fails", func() {
		deployment.Workspace = fakeWorkspace
		fakeDeploymentManager.GetTerraformDeploymentReturns(deployment, nil)
		tfVersion := "1.1"
		fakeInvokerBuilder.VersionedTerraformInvokerReturns(fakeDefaultInvoker)
		fakeWorkspace.StateVersionReturns(newVersion(tfVersion), nil)
		fakeDefaultInvoker.ApplyReturns(genericError)

		runner := tf.NewTerraformProvider(executor.TFBinariesContext{DefaultTfVersion: newVersion(tfVersion)}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)
		_, err := runner.Update(context.TODO(), varContext)
		Expect(err).To(Succeed())

		Eventually(operationWasFinishedForDeployment(fakeDeploymentManager)).Should(Equal(deployment))
		Expect(operationWasFinishedWithError(fakeDeploymentManager)()).To(MatchError(genericError))
	})

	Context("when tfUpgrades are enabled", func() {
		BeforeEach(func() {
			viper.Set(tf.TfUpgradeEnabled, true)
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
			fakeInvoker1 := &tffakes.FakeTerraformInvoker{}
			fakeInvoker2 := &tffakes.FakeTerraformInvoker{}

			deployment.Workspace = fakeWorkspace
			fakeDeploymentManager.GetTerraformDeploymentReturns(deployment, nil)
			fakeInvokerBuilder.VersionedTerraformInvokerReturnsOnCall(0, fakeInvoker1)
			fakeInvokerBuilder.VersionedTerraformInvokerReturnsOnCall(1, fakeInvoker2)
			fakeInvokerBuilder.VersionedTerraformInvokerReturnsOnCall(2, fakeDefaultInvoker)

			fakeWorkspace.StateVersionReturns(newVersion("0.0.1"), nil)
			fakeWorkspace.ModuleInstancesReturns([]workspace.ModuleInstance{{ModuleName: "moduleName"}})

			runner := tf.NewTerraformProvider(tfBinContext, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)
			_, err := runner.Update(context.TODO(), varContext)
			Expect(err).NotTo(HaveOccurred())

			Eventually(operationWasFinishedForDeployment(fakeDeploymentManager)).Should(Equal(deployment))
			Expect(operationWasFinishedWithError(fakeDeploymentManager)()).To(BeNil())

			Expect(fakeInvoker1.ApplyCallCount()).To(Equal(1))
			Expect(getWorkspace(fakeInvoker1, 0)).To(Equal(fakeWorkspace))

			Expect(fakeInvoker2.ApplyCallCount()).To(Equal(1))
			Expect(getWorkspace(fakeInvoker2, 0)).To(Equal(fakeWorkspace))

			Expect(fakeDefaultInvoker.ApplyCallCount()).To(Equal(1))
			Expect(getWorkspace(fakeDefaultInvoker, 0)).To(Equal(fakeWorkspace))

			Expect(fakeInvokerBuilder.VersionedTerraformInvokerCallCount()).To(Equal(3))
			Expect(fakeInvokerBuilder.VersionedTerraformInvokerArgsForCall(0)).To(Equal(newVersion("0.0.2")))
			Expect(fakeInvokerBuilder.VersionedTerraformInvokerArgsForCall(1)).To(Equal(newVersion("0.1.0")))
			Expect(fakeInvokerBuilder.VersionedTerraformInvokerArgsForCall(2)).To(Equal(newVersion("0.1.0")))
		})

		It("fails the update, if the version of statefile does not match the default tf version, and no upgrade path is specified", func() {
			tfBinContext := executor.TFBinariesContext{
				DefaultTfVersion: newVersion("0.1.0"),
			}
			deployment.Workspace = fakeWorkspace
			fakeDeploymentManager.GetTerraformDeploymentReturns(deployment, nil)
			fakeWorkspace.StateVersionReturns(newVersion("0.0.1"), nil)

			runner := tf.NewTerraformProvider(tfBinContext, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

			_, err := runner.Update(context.TODO(), varContext)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeInvokerBuilder.VersionedTerraformInvokerCallCount()).To(Equal(0))
		})
	})

	Context("when tfUpgrades are disabled", func() {
		BeforeEach(func() {
			viper.Set(tf.TfUpgradeEnabled, false)
		})

		It("fails the update, if the version of statefile does not match the default tf version", func() {
			tfBinContext := executor.TFBinariesContext{
				DefaultTfVersion: newVersion("0.1.0"),
			}
			deployment.Workspace = fakeWorkspace
			fakeDeploymentManager.GetTerraformDeploymentReturns(deployment, nil)

			fakeWorkspace.StateVersionReturns(newVersion("0.0.1"), nil)

			runner := tf.NewTerraformProvider(tfBinContext, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

			_, err := runner.Update(context.TODO(), varContext)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeInvokerBuilder.VersionedTerraformInvokerCallCount()).To(Equal(0))
		})

		It("performs the update, default tf version matches instance", func() {
			tfBinContext := executor.TFBinariesContext{
				DefaultTfVersion: newVersion("0.1.0"),
			}
			deployment.Workspace = fakeWorkspace
			fakeDeploymentManager.GetTerraformDeploymentReturns(deployment, nil)
			fakeInvokerBuilder.VersionedTerraformInvokerReturnsOnCall(0, fakeDefaultInvoker)
			fakeWorkspace.StateVersionReturns(newVersion("0.1.0"), nil)

			runner := tf.NewTerraformProvider(tfBinContext, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

			_, err := runner.Update(context.TODO(), varContext)
			Expect(err).NotTo(HaveOccurred())

			Eventually(operationWasFinishedForDeployment(fakeDeploymentManager)).Should(Equal(deployment))
			Expect(operationWasFinishedWithError(fakeDeploymentManager)()).To(BeNil())

			Expect(fakeDefaultInvoker.ApplyCallCount()).To(Equal(1))
			_, workspace := fakeDefaultInvoker.ApplyArgsForCall(0)
			Expect(workspace).To(Equal(fakeWorkspace))
		})
	})
},
)
