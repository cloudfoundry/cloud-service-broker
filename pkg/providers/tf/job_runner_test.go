package tf_test

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-version"
	"github.com/spf13/viper"

	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/tffakes"

	"github.com/cloudfoundry/cloud-service-broker/brokerapi/broker/brokerfakes"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/wrapper"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/wrapper/wrapperfakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TfJobRunner", func() {
	var (
		fakeStore           *brokerfakes.FakeStorage
		fakeExecutorFactory *wrapperfakes.FakeExecutorBuilder
		workspaceFactory    *tffakes.FakeWorkspaceBuilder
		fakeWorkspace       *tffakes.FakeWorkspace
		fakeExecutorDefault *wrapperfakes.FakeTerraformExecutor
		deploymentId        string
		deployment          storage.TerraformDeployment
		genericError        = fmt.Errorf("genericError")
	)
	BeforeEach(func() {
		fakeStore = &brokerfakes.FakeStorage{}
		fakeExecutorFactory = &wrapperfakes.FakeExecutorBuilder{}
		workspaceFactory = &tffakes.FakeWorkspaceBuilder{}
		fakeWorkspace = &tffakes.FakeWorkspace{}
		fakeWorkspace.ModuleInstancesReturns([]wrapper.ModuleInstance{{ModuleName: "moduleName"}})
		fakeExecutorDefault = &wrapperfakes.FakeTerraformExecutor{}
		deploymentId = "deploymentID"
		deployment = storage.TerraformDeployment{
			ID: deploymentId,
		}
		viper.Reset()
	})

	Describe("Update", func() {
		It("fails, when deployment can't be found", func() {
			fakeStore.GetTerraformDeploymentReturns(storage.TerraformDeployment{}, genericError)
			runner := tf.NewTfJobRunner(fakeStore, fakeExecutorFactory, wrapper.TFBinariesContext{}, workspaceFactory)
			Expect(runner.Update(context.TODO(), deploymentId, nil)).To(MatchError(genericError))
		})

		It("fails, when workspace can't be created", func() {
			fakeStore.GetTerraformDeploymentReturns(deployment, nil)
			workspaceFactory.CreateWorkspaceReturns(nil, genericError)
			runner := tf.NewTfJobRunner(fakeStore, fakeExecutorFactory, wrapper.TFBinariesContext{}, workspaceFactory)
			Expect(runner.Update(context.TODO(), deploymentId, nil)).To(MatchError(genericError))
		})

		It("fails, when store cant be updated", func() {
			fakeStore.GetTerraformDeploymentReturns(deployment, nil)
			workspaceFactory.CreateWorkspaceReturns(fakeWorkspace, nil)
			fakeStore.StoreTerraformDeploymentReturns(genericError)
			runner := tf.NewTfJobRunner(fakeStore, fakeExecutorFactory, wrapper.TFBinariesContext{}, workspaceFactory)
			Expect(runner.Update(context.TODO(), deploymentId, nil)).To(MatchError(genericError))
		})

		It("fails, when cant get version from state", func() {
			fakeStore.GetTerraformDeploymentReturns(deployment, nil)
			workspaceFactory.CreateWorkspaceReturns(fakeWorkspace, nil)
			fakeWorkspace.StateVersionReturns(nil, genericError)

			runner := tf.NewTfJobRunner(fakeStore, fakeExecutorFactory, wrapper.TFBinariesContext{}, workspaceFactory)
			Expect(runner.Update(context.TODO(), deploymentId, nil)).To(Succeed())
			Eventually(lastStoredLastOperation(fakeStore)).Should(Or(Equal(tf.Succeeded), Equal(tf.Failed)))

			Expect(lastStoredLastOperation(fakeStore)()).To(Equal(tf.Failed))
			Expect(lastStoredDeployment(fakeStore)().LastOperationMessage).To(ContainSubstring(genericError.Error()))
		})

		It("fails, when cant update instance configuration", func() {
			fakeStore.GetTerraformDeploymentReturns(deployment, nil)
			workspaceFactory.CreateWorkspaceReturns(fakeWorkspace, nil)
			tfVersion := "1.1"
			fakeWorkspace.StateVersionReturns(newVersion(tfVersion), nil)
			fakeWorkspace.UpdateInstanceConfigurationReturns(genericError)

			runner := tf.NewTfJobRunner(fakeStore, fakeExecutorFactory, wrapper.TFBinariesContext{DefaultTfVersion: newVersion(tfVersion)}, workspaceFactory)
			Expect(runner.Update(context.TODO(), deploymentId, nil)).To(Succeed())
			Eventually(lastStoredLastOperation(fakeStore)).Should(Or(Equal(tf.Succeeded), Equal(tf.Failed)))

			Expect(lastStoredLastOperation(fakeStore)()).To(Equal(tf.Failed))
			Expect(lastStoredDeployment(fakeStore)().LastOperationMessage).To(ContainSubstring(genericError.Error()))
		})

		It("updates templates before applying", func() {
			fakeStore.GetTerraformDeploymentReturns(deployment, nil)
			workspaceFactory.CreateWorkspaceReturns(fakeWorkspace, nil)
			tfVersion := "1.1"
			fakeWorkspace.StateVersionReturns(newVersion(tfVersion), nil)

			runner := tf.NewTfJobRunner(fakeStore, fakeExecutorFactory, wrapper.TFBinariesContext{DefaultTfVersion: newVersion(tfVersion)}, workspaceFactory)
			templateVars := map[string]interface{}{"var": "value"}

			Expect(runner.Update(context.TODO(), deploymentId, templateVars)).To(Succeed())
			Eventually(lastStoredLastOperation(fakeStore)).Should(Or(Equal(tf.Succeeded), Equal(tf.Failed)))
			Expect(lastStoredLastOperation(fakeStore)()).To(Equal(tf.Succeeded))

			Expect(fakeWorkspace.UpdateInstanceConfigurationCallCount()).To(Equal(1))
			Expect(fakeWorkspace.UpdateInstanceConfigurationArgsForCall(0)).To(Equal(templateVars))
		})

		It("updates the last operation on success, with the status from terraform", func() {
			fakeStore.GetTerraformDeploymentReturns(deployment, nil)
			workspaceFactory.CreateWorkspaceReturns(fakeWorkspace, nil)
			tfVersion := "1.1"
			fakeWorkspace.StateVersionReturns(newVersion(tfVersion), nil)
			fakeWorkspace.OutputsReturns(map[string]interface{}{"status": "status from terraform"}, nil)

			runner := tf.NewTfJobRunner(fakeStore, fakeExecutorFactory, wrapper.TFBinariesContext{DefaultTfVersion: newVersion(tfVersion)}, workspaceFactory)

			Expect(runner.Update(context.TODO(), deploymentId, nil)).To(Succeed())
			Eventually(lastStoredLastOperation(fakeStore)).Should(Or(Equal(tf.Succeeded), Equal(tf.Failed)))
			Expect(lastStoredLastOperation(fakeStore)()).To(Equal(tf.Succeeded))

			Expect(lastStoredDeployment(fakeStore)().LastOperationMessage).To(Equal("status from terraform"))
		})

		It("return the error in last operation, if terraform apply fails", func() {
			fakeStore.GetTerraformDeploymentReturns(deployment, nil)
			workspaceFactory.CreateWorkspaceReturns(fakeWorkspace, nil)
			tfVersion := "1.1"
			fakeWorkspace.StateVersionReturns(newVersion(tfVersion), nil)
			fakeWorkspace.ApplyReturns(genericError)

			runner := tf.NewTfJobRunner(fakeStore, fakeExecutorFactory, wrapper.TFBinariesContext{DefaultTfVersion: newVersion(tfVersion)}, workspaceFactory)

			Expect(runner.Update(context.TODO(), deploymentId, nil)).To(Succeed())
			Eventually(lastStoredLastOperation(fakeStore)).Should(Or(Equal(tf.Succeeded), Equal(tf.Failed)))
			Expect(lastStoredLastOperation(fakeStore)()).To(Equal(tf.Failed))
			Expect(lastStoredDeployment(fakeStore)().LastOperationMessage).To(ContainSubstring(genericError.Error()))
		})

		Context("when tfUpgrades are enabled", func() {
			BeforeEach(func() {
				viper.Set(tf.TfUpgradeEnabled, true)
			})
			It("runs apply with all tf versions in the upgrade path", func() {
				tfBinContext := wrapper.TFBinariesContext{
					DefaultTfVersion: newVersion("0.1.0"),
					TfUpgradePath: []*version.Version{
						newVersion("0.0.1"),
						newVersion("0.0.2"),
						newVersion("0.1.0"),
					},
				}
				fakeExecutor1 := &wrapperfakes.FakeTerraformExecutor{}
				fakeExecutor2 := &wrapperfakes.FakeTerraformExecutor{}

				fakeStore.GetTerraformDeploymentReturns(deployment, nil)
				workspaceFactory.CreateWorkspaceReturns(fakeWorkspace, nil)
				fakeExecutorFactory.VersionedExecutorReturnsOnCall(0, fakeExecutor1)
				fakeExecutorFactory.VersionedExecutorReturnsOnCall(1, fakeExecutor2)
				fakeExecutorFactory.VersionedExecutorReturnsOnCall(2, fakeExecutorDefault)

				fakeWorkspace.StateVersionReturns(newVersion("0.0.1"), nil)
				fakeWorkspace.ModuleInstancesReturns([]wrapper.ModuleInstance{{ModuleName: "moduleName"}})

				runner := tf.NewTfJobRunner(fakeStore, fakeExecutorFactory, tfBinContext, workspaceFactory)
				runner.Update(context.TODO(), deploymentId, nil)

				Eventually(lastStoredLastOperation(fakeStore)).Should(Or(Equal(tf.Succeeded), Equal(tf.Failed)))
				Expect(lastStoredLastOperation(fakeStore)()).To(Equal(tf.Succeeded))

				Expect(fakeWorkspace.ApplyCallCount()).To(Equal(3))

				Expect(getExecutor(fakeWorkspace, 0)).To(Equal(fakeExecutor1))
				Expect(getExecutor(fakeWorkspace, 1)).To(Equal(fakeExecutor2))
				Expect(getExecutor(fakeWorkspace, 1)).To(Equal(fakeExecutorDefault))

				Expect(fakeExecutorFactory.VersionedExecutorCallCount()).To(Equal(3))
				Expect(fakeExecutorFactory.VersionedExecutorArgsForCall(0)).To(Equal(newVersion("0.0.2")))
				Expect(fakeExecutorFactory.VersionedExecutorArgsForCall(1)).To(Equal(newVersion("0.1.0")))
				Expect(fakeExecutorFactory.VersionedExecutorArgsForCall(2)).To(Equal(newVersion("0.1.0")))
			})

			It("fails the update, if the version of version of statefile does not match the default tf version, and no upgrade path is specified", func() {
				tfBinContext := wrapper.TFBinariesContext{
					DefaultTfVersion: newVersion("0.1.0"),
				}

				fakeStore.GetTerraformDeploymentReturns(deployment, nil)
				workspaceFactory.CreateWorkspaceReturns(fakeWorkspace, nil)
				fakeWorkspace.StateVersionReturns(newVersion("0.0.1"), nil)

				runner := tf.NewTfJobRunner(fakeStore, fakeExecutorFactory, tfBinContext, workspaceFactory)
				runner.Update(context.TODO(), deploymentId, nil)

				Eventually(lastStoredLastOperation(fakeStore)).Should(Or(Equal(tf.Succeeded), Equal(tf.Failed)))
				Expect(lastStoredLastOperation(fakeStore)()).To(Equal(tf.Failed))
				Expect(fakeExecutorFactory.VersionedExecutorCallCount()).To(Equal(0))
			})
		})

		Context("when tfUpgrades are disabled", func() {
			BeforeEach(func() {
				viper.Set(tf.TfUpgradeEnabled, false)
			})

			It("fails the update, if the version of version of statefile does not match the default tf version", func() {
				tfBinContext := wrapper.TFBinariesContext{
					DefaultTfVersion: newVersion("0.1.0"),
				}

				fakeStore.GetTerraformDeploymentReturns(deployment, nil)
				workspaceFactory.CreateWorkspaceReturns(fakeWorkspace, nil)
				fakeWorkspace.StateVersionReturns(newVersion("0.0.1"), nil)

				runner := tf.NewTfJobRunner(fakeStore, fakeExecutorFactory, tfBinContext, workspaceFactory)
				runner.Update(context.TODO(), deploymentId, nil)

				Eventually(lastStoredLastOperation(fakeStore)).Should(Or(Equal(tf.Succeeded), Equal(tf.Failed)))
				Expect(lastStoredLastOperation(fakeStore)()).To(Equal(tf.Failed))
				Expect(fakeExecutorFactory.VersionedExecutorCallCount()).To(Equal(0))
			})

			It("performs the update, default tf version matches instance", func() {
				tfBinContext := wrapper.TFBinariesContext{
					DefaultTfVersion: newVersion("0.1.0"),
				}

				fakeStore.GetTerraformDeploymentReturns(deployment, nil)
				workspaceFactory.CreateWorkspaceReturns(fakeWorkspace, nil)
				fakeExecutorFactory.VersionedExecutorReturnsOnCall(0, fakeExecutorDefault)
				fakeWorkspace.StateVersionReturns(newVersion("0.1.0"), nil)

				var templateVars map[string]interface{}

				runner := tf.NewTfJobRunner(fakeStore, fakeExecutorFactory, tfBinContext, workspaceFactory)
				runner.Update(context.TODO(), deploymentId, templateVars)

				Eventually(lastStoredLastOperation(fakeStore)).Should(Or(Equal(tf.Succeeded), Equal(tf.Failed)))

				Expect(lastStoredLastOperation(fakeStore)()).To(Equal(tf.Succeeded))

				Expect(fakeWorkspace.ApplyCallCount()).To(Equal(1))
				_, executor := fakeWorkspace.ApplyArgsForCall(0)
				Expect(executor).To(Equal(fakeExecutorDefault))
			})
		})

	})

})

func getExecutor(workspace *tffakes.FakeWorkspace, pos int) wrapper.TerraformExecutor {
	_, executor := workspace.ApplyArgsForCall(pos)
	return executor
}

func lastStoredDeployment(fakeStore *brokerfakes.FakeStorage) func() storage.TerraformDeployment {
	return func() storage.TerraformDeployment {
		callCount := fakeStore.StoreTerraformDeploymentCallCount()
		if callCount == 0 {
			return storage.TerraformDeployment{}
		} else {
			return fakeStore.StoreTerraformDeploymentArgsForCall(callCount - 1)
		}
	}
}
func lastStoredLastOperation(fakeStore *brokerfakes.FakeStorage) func() string {
	return func() string {
		return lastStoredDeployment(fakeStore)().LastOperationState
	}
}

func newVersion(v string) *version.Version {
	return version.Must(version.NewVersion(v))
}
