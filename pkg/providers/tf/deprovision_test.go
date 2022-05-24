package tf_test

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-version"

	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/workspace"

	"github.com/cloudfoundry/cloud-service-broker/pkg/varcontext"
	"github.com/pivotal-cf/brokerapi/v8/domain"

	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/executor"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/tffakes"
	"github.com/cloudfoundry/cloud-service-broker/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Deprovision", func() {
	const instanceGUID = "cc57a89e-8f43-48e8-9e41-c7c99d331066"
	const expectedError = "generic error"
	expectedTFID := fmt.Sprintf("tf:%s:", instanceGUID)

	var (
		deployment            storage.TerraformDeployment
		fakeInvokerBuilder    *tffakes.FakeTerraformInvokerBuilder
		fakeDefaultInvoker    *tffakes.FakeTerraformInvoker
		fakeLogger            = utils.NewLogger("test")
		fakeServiceDefinition tf.TfServiceDefinitionV1
		fakeDeploymentManager *tffakes.FakeDeploymentManagerInterface
		deprovisionContext    *varcontext.VarContext
		templateVars          = map[string]interface{}{"tf_id": instanceGUID}
	)

	BeforeEach(func() {
		var err error
		fakeInvokerBuilder = &tffakes.FakeTerraformInvokerBuilder{}
		fakeDeploymentManager = &tffakes.FakeDeploymentManagerInterface{}
		fakeDefaultInvoker = &tffakes.FakeTerraformInvoker{}

		deprovisionContext, err = varcontext.Builder().MergeMap(templateVars).Build()
		Expect(err).NotTo(HaveOccurred())

		deployment = storage.TerraformDeployment{
			ID: instanceGUID,
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

	It("deletes the instance", func() {
		fakeInvokerBuilder.VersionedTerraformInvokerReturns(fakeDefaultInvoker)
		fakeDeploymentManager.GetTerraformDeploymentReturns(deployment, nil)
		provider := tf.NewTerraformProvider(executor.TFBinariesContext{DefaultTfVersion: version.Must(version.NewVersion("1"))}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

		actualOperationId, err := provider.Deprovision(context.TODO(), instanceGUID, domain.DeprovisionDetails{}, deprovisionContext)
		Expect(err).NotTo(HaveOccurred())
		Expect(actualOperationId).To(Equal(&expectedTFID))

		By("Checking the HCL was updated with correct parameters")
		actualGUID, actualProvisionSettings, actualContext := fakeDeploymentManager.UpdateWorkspaceHCLArgsForCall(0)
		Expect(actualGUID).To(Equal(expectedTFID))
		Expect(actualProvisionSettings).To(Equal(fakeServiceDefinition.ProvisionSettings))
		Expect(actualContext["tf_id"]).To(Equal(instanceGUID))

		By("Checking destroy gets the correct deployment")
		Expect(fakeDeploymentManager.GetTerraformDeploymentCallCount()).To(Equal(1))
		Expect(fakeDeploymentManager.GetTerraformDeploymentArgsForCall(0)).To(Equal(expectedTFID))

		By("Checking that deprovision is marked as started")
		Expect(fakeDeploymentManager.MarkOperationStartedCallCount()).To(Equal(1))
		actualDeployment, actualOperationType := fakeDeploymentManager.MarkOperationStartedArgsForCall(0)
		Expect(actualDeployment).To(Equal(deployment))
		Expect(actualOperationType).To(Equal("deprovision"))

		By("checking TF apply has been called")
		Eventually(destroyCallCount(fakeDefaultInvoker)).Should(Equal(1))
		Eventually(operationWasFinishedForDeployment(fakeDeploymentManager)).Should(Equal(deployment))
		Expect(operationWasFinishedWithError(fakeDeploymentManager)()).To(BeNil())

	})

	It("fails, when unable to update the workspace HCL", func() {
		fakeDeploymentManager.UpdateWorkspaceHCLReturns(fmt.Errorf(expectedError))

		provider := tf.NewTerraformProvider(executor.TFBinariesContext{DefaultTfVersion: version.Must(version.NewVersion("1"))}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

		actualOperationId, err := provider.Deprovision(context.TODO(), instanceGUID, domain.DeprovisionDetails{}, deprovisionContext)
		Expect(err).To(MatchError(expectedError))
		Expect(actualOperationId).To(BeNil())

	})

	It("fails, when unable to get the Terraform deployment", func() {
		fakeDeploymentManager.GetTerraformDeploymentReturns(storage.TerraformDeployment{}, fmt.Errorf(expectedError))

		provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

		actualOperationId, err := provider.Deprovision(context.TODO(), instanceGUID, domain.DeprovisionDetails{}, deprovisionContext)
		Expect(err).To(MatchError(expectedError))
		Expect(actualOperationId).To(BeNil())

	})

	It("fails, when unable to mark operation as started", func() {
		fakeDeploymentManager.GetTerraformDeploymentReturns(deployment, nil)
		fakeDeploymentManager.MarkOperationStartedReturns(fmt.Errorf(expectedError))

		provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

		actualOperationId, err := provider.Deprovision(context.TODO(), instanceGUID, domain.DeprovisionDetails{}, deprovisionContext)
		Expect(err).To(MatchError(expectedError))
		Expect(actualOperationId).To(BeNil())

	})

	It("returns the error in last operation if terraform destroy fails", func() {
		fakeDeploymentManager.GetTerraformDeploymentReturns(deployment, nil)
		fakeDeploymentManager.MarkOperationStartedReturns(nil)
		fakeInvokerBuilder.VersionedTerraformInvokerReturns(fakeDefaultInvoker)
		fakeDefaultInvoker.DestroyReturns(fmt.Errorf(expectedError))

		provider := tf.NewTerraformProvider(executor.TFBinariesContext{DefaultTfVersion: version.Must(version.NewVersion("1"))}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

		actualOperationId, err := provider.Deprovision(context.TODO(), instanceGUID, domain.DeprovisionDetails{}, deprovisionContext)
		Expect(err).NotTo(HaveOccurred())
		Expect(actualOperationId).To(Equal(&expectedTFID))

		By("checking last operation updated with error")
		Eventually(operationWasFinishedForDeployment(fakeDeploymentManager)).Should(Equal(deployment))
		Expect(operationWasFinishedWithError(fakeDeploymentManager)()).To(MatchError(expectedError))
	})

})
