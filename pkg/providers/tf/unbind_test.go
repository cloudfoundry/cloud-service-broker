package tf_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/go-version"

	"github.com/cloudfoundry/cloud-service-broker/v3/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/providers/tf/workspace"

	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/providers/tf"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/providers/tf/executor"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/providers/tf/tffakes"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/v3/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unbind", func() {
	const instanceGUID = "50d27a3f-9b85-47d7-8009-667f258ab807"
	const bindingGUID = "7d59792a-1813-4b81-8f99-1458e4267a09"
	const expectedError = "generic error"
	expectedTFID := fmt.Sprintf("tf:%s:%s", instanceGUID, bindingGUID)

	var (
		deployment            storage.TerraformDeployment
		fakeInvokerBuilder    *tffakes.FakeTerraformInvokerBuilder
		fakeDefaultInvoker    *tffakes.FakeTerraformInvoker
		fakeLogger            = utils.NewLogger("test")
		fakeServiceDefinition tf.TfServiceDefinitionV1
		fakeDeploymentManager *tffakes.FakeDeploymentManagerInterface
		unbindContext         *varcontext.VarContext
		templateVars          = map[string]any{"instance_id": instanceGUID}
	)

	BeforeEach(func() {
		var err error
		fakeInvokerBuilder = &tffakes.FakeTerraformInvokerBuilder{}
		fakeDeploymentManager = &tffakes.FakeDeploymentManagerInterface{}
		fakeDefaultInvoker = &tffakes.FakeTerraformInvoker{}

		unbindContext, err = varcontext.Builder().MergeMap(templateVars).Build()
		Expect(err).NotTo(HaveOccurred())

		template := `variable username { type = string }`

		fakeServiceDefinition = tf.TfServiceDefinitionV1{
			BindSettings: tf.TfServiceDefinitionV1Action{
				Template:  template,
				Templates: map[string]string{"first": template},
			},
		}

		deployment = storage.TerraformDeployment{
			ID: fmt.Sprintf("tf:%s:%s", instanceGUID, bindingGUID),
			Workspace: &workspace.TerraformWorkspace{
				Modules: []workspace.ModuleDefinition{
					{
						Name: "test-module-instance",
					},
				},
				Instances: []workspace.ModuleInstance{
					{
						ModuleName: "test-module-instance",
					},
				},
				State: []byte(`{"terraform_version":"1"}`),
			},
		}
	})

	JustBeforeEach(func() {
		fakeDeploymentManager.UpdateWorkspaceHCLReturns(nil)
	})

	It("destroys the binding", func() {
		fakeInvokerBuilder.VersionedTerraformInvokerReturns(fakeDefaultInvoker)
		fakeDeploymentManager.GetTerraformDeploymentReturns(deployment, nil)
		fakeDeploymentManager.OperationStatusReturns(true, "operation succeeded", nil)

		provider := tf.NewTerraformProvider(executor.TFBinariesContext{DefaultTfVersion: version.Must(version.NewVersion("1"))}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

		err := provider.Unbind(context.TODO(), instanceGUID, bindingGUID, unbindContext)
		Expect(err).NotTo(HaveOccurred())

		By("Checking the HCL was updated with correct parameters")
		actualTFID, actualBindSettings, actualContext := fakeDeploymentManager.UpdateWorkspaceHCLArgsForCall(0)
		Expect(actualTFID).To(Equal(expectedTFID))
		Expect(actualBindSettings).To(Equal(fakeServiceDefinition.BindSettings))
		Expect(actualContext["instance_id"]).To(Equal(instanceGUID))

		By("Checking destroy gets the correct deployment")
		Expect(fakeDeploymentManager.GetTerraformDeploymentCallCount()).To(Equal(1))
		Expect(fakeDeploymentManager.GetTerraformDeploymentArgsForCall(0)).To(Equal(expectedTFID))

		By("Checking that deprovision is marked as started")
		Expect(fakeDeploymentManager.MarkOperationStartedCallCount()).To(Equal(1))
		actualDeployment, actualOperationType := fakeDeploymentManager.MarkOperationStartedArgsForCall(0)
		Expect(actualDeployment).To(Equal(&deployment))
		Expect(actualOperationType).To(Equal("unbind"))

		By("checking TF destroy has been called")
		Eventually(destroyCallCount(fakeDefaultInvoker)).Should(Equal(1))
		Eventually(operationWasFinishedForDeployment(fakeDeploymentManager)).Should(Equal(deployment))
		Expect(operationWasFinishedWithError(fakeDeploymentManager)()).To(BeNil())
	})

	It("fails, when unable to update the workspace HCL", func() {
		fakeDeploymentManager.UpdateWorkspaceHCLReturns(errors.New(expectedError))

		provider := tf.NewTerraformProvider(executor.TFBinariesContext{DefaultTfVersion: version.Must(version.NewVersion("1"))}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

		err := provider.Unbind(context.TODO(), instanceGUID, bindingGUID, unbindContext)
		Expect(err).To(MatchError(expectedError))
	})

	It("fails, when unable to get the Terraform deployment", func() {
		fakeDeploymentManager.GetTerraformDeploymentReturns(storage.TerraformDeployment{}, errors.New(expectedError))

		provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

		err := provider.Unbind(context.TODO(), instanceGUID, bindingGUID, unbindContext)
		Expect(err).To(MatchError(expectedError))
	})

	It("fails, when unable to mark operation as started", func() {
		fakeDeploymentManager.GetTerraformDeploymentReturns(deployment, nil)
		fakeDeploymentManager.MarkOperationStartedReturns(errors.New(expectedError))

		provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

		err := provider.Unbind(context.TODO(), instanceGUID, bindingGUID, unbindContext)
		Expect(err).To(MatchError(expectedError))
	})

	It("returns an error if tofu destroy fails", func() {
		fakeDeploymentManager.GetTerraformDeploymentReturns(deployment, nil)
		fakeDeploymentManager.MarkOperationStartedReturns(nil)
		fakeInvokerBuilder.VersionedTerraformInvokerReturns(fakeDefaultInvoker)
		fakeDefaultInvoker.DestroyReturns(errors.New(expectedError))
		fakeDeploymentManager.OperationStatusReturns(true, "", errors.New(expectedError))

		provider := tf.NewTerraformProvider(executor.TFBinariesContext{DefaultTfVersion: version.Must(version.NewVersion("1"))}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

		err := provider.Unbind(context.TODO(), instanceGUID, bindingGUID, unbindContext)
		Expect(err).To(MatchError(expectedError))
	})

	Describe("DeleteBindingData", func() {
		var provider *tf.TerraformProvider
		BeforeEach(func() {
			fakeDeploymentManager = &tffakes.FakeDeploymentManagerInterface{}

			provider = tf.NewTerraformProvider(
				executor.TFBinariesContext{DefaultTfVersion: version.Must(version.NewVersion("1.4"))},
				fakeInvokerBuilder,
				fakeLogger,
				fakeServiceDefinition,
				fakeDeploymentManager,
			)
		})

		It("deletes binding deployment from database", func() {
			Expect(provider.DeleteBindingData(context.TODO(), instanceGUID, bindingGUID)).To(BeNil())
			Expect(fakeDeploymentManager.DeleteTerraformDeploymentCallCount()).To(Equal(1))
			Expect(fakeDeploymentManager.DeleteTerraformDeploymentArgsForCall(0)).To(Equal(fmt.Sprintf("tf:%s:%s", instanceGUID, bindingGUID)))
		})

		It("returns any errors", func() {
			fakeDeploymentManager.DeleteTerraformDeploymentReturns(fmt.Errorf("some error deleting the deployment from the database"))
			Expect(provider.DeleteBindingData(context.TODO(), instanceGUID, bindingGUID)).To(MatchError("some error deleting the deployment from the database"))
			Expect(fakeDeploymentManager.DeleteTerraformDeploymentCallCount()).To(Equal(1))
			Expect(fakeDeploymentManager.DeleteTerraformDeploymentArgsForCall(0)).To(Equal(fmt.Sprintf("tf:%s:%s", instanceGUID, bindingGUID)))
		})
	})
})
