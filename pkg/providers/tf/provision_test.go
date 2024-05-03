package tf_test

import (
	"context"
	"errors"

	"github.com/cloudfoundry/cloud-service-broker/v3/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/providers/tf"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/providers/tf/executor"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/providers/tf/tffakes"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/providers/tf/workspace"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/v3/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Provision", func() {
	const expectedTfID = "567c6af0-d68a-11ec-a5b6-367dda7ea869"

	var (
		fakeDeploymentManager *tffakes.FakeDeploymentManagerInterface
		deployment            storage.TerraformDeployment
		fakeInvokerBuilder    *tffakes.FakeTerraformInvokerBuilder
		fakeDefaultInvoker    *tffakes.FakeTerraformInvoker
		fakeLogger            = utils.NewLogger("test")
		fakeServiceDefinition tf.TfServiceDefinitionV1
		template              = `variable username {type = string}`
	)

	BeforeEach(func() {
		fakeInvokerBuilder = &tffakes.FakeTerraformInvokerBuilder{}

		fakeServiceDefinition = tf.TfServiceDefinitionV1{
			ProvisionSettings: tf.TfServiceDefinitionV1Action{
				Template:  template,
				Templates: map[string]string{"first": template},
			},
		}
		fakeDeploymentManager = &tffakes.FakeDeploymentManagerInterface{}
		fakeDefaultInvoker = &tffakes.FakeTerraformInvoker{}
		deployment = storage.TerraformDeployment{
			ID: expectedTfID,
			Workspace: &workspace.TerraformWorkspace{
				Modules: []workspace.ModuleDefinition{
					{Name: "brokertemplate"},
				},
			},
		}
	})

	var (
		provisionContext *varcontext.VarContext
		templateVars     = map[string]any{"tf_id": expectedTfID, "username": "some-user"}
	)

	BeforeEach(func() {
		var err error
		provisionContext, err = varcontext.Builder().MergeMap(templateVars).Build()
		Expect(err).NotTo(HaveOccurred())
	})

	It("creates a provision deployment", func() {
		fakeDeploymentManager.CreateAndSaveDeploymentReturns(deployment, nil)
		fakeInvokerBuilder.VersionedTerraformInvokerReturns(fakeDefaultInvoker)
		provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

		actualInstanceDetails, err := provider.Provision(context.TODO(), provisionContext)

		Expect(err).NotTo(HaveOccurred())
		Expect(actualInstanceDetails.OperationGUID).To(Equal(expectedTfID))
		Expect(actualInstanceDetails.OperationType).To(Equal("provision"))

		By("checking the new saved deployment")
		Expect(fakeDeploymentManager.CreateAndSaveDeploymentCallCount()).To(Equal(1))
		actualTfID, actualWorkspace := fakeDeploymentManager.CreateAndSaveDeploymentArgsForCall(0)
		Expect(actualTfID).To(Equal(expectedTfID))
		Expect(actualWorkspace.Modules[0].Name).To(Equal("brokertemplate"))
		Expect(actualWorkspace.Modules[0].Definition).To(Equal(fakeServiceDefinition.ProvisionSettings.Template))
		Expect(actualWorkspace.Modules[0].Definitions).To(Equal(fakeServiceDefinition.ProvisionSettings.Templates))
		Expect(actualWorkspace.Instances[0].Configuration).To(Equal(map[string]any{"username": "some-user"}))

		By("checking that provision is marked as started")
		Expect(fakeDeploymentManager.MarkOperationStartedCallCount()).To(Equal(1))
		actualDeployment, actualOperationType := fakeDeploymentManager.MarkOperationStartedArgsForCall(0)
		Expect(actualDeployment).To(Equal(&deployment))
		Expect(actualOperationType).To(Equal("provision"))

		By("checking TF apply has been called")
		Eventually(applyCallCount(fakeDefaultInvoker)).Should(Equal(1))
		Eventually(operationWasFinishedForDeployment(fakeDeploymentManager)).Should(Equal(deployment))
		Expect(operationWasFinishedWithError(fakeDeploymentManager)()).To(BeNil())
	})

	It("fails, when tfID is not provided", func() {
		var err error
		provisionContext, err = varcontext.Builder().Build()
		Expect(err).NotTo(HaveOccurred())

		provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

		_, err = provider.Provision(context.TODO(), provisionContext)

		Expect(err.Error()).To(ContainSubstring(`missing value for key "tf_id"`))
	})

	It("fails, when workspace cannot be created", func() {
		fakeServiceDefinition = tf.TfServiceDefinitionV1{
			ProvisionSettings: tf.TfServiceDefinitionV1Action{
				Template:  "invalid template",
				Templates: map[string]string{"first": "invalid template"},
			},
		}

		provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

		_, err := provider.Provision(context.TODO(), provisionContext)

		Expect(err).To(MatchError(`error creating workspace: :1,17-17: Invalid block definition; Either a quoted string block label or an opening brace ("{") is expected here.`))
	})

	It("fails, when it errors saving the deployment", func() {
		fakeDeploymentManager.CreateAndSaveDeploymentReturns(storage.TerraformDeployment{}, errors.New("cant save now"))
		provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

		_, err := provider.Provision(context.TODO(), provisionContext)

		Expect(err).To(MatchError("deployment create failed: cant save now"))
	})

	It("fails, when it errors marking the operation as started", func() {
		fakeDeploymentManager.CreateAndSaveDeploymentReturns(deployment, nil)
		fakeDeploymentManager.MarkOperationStartedReturns(errors.New("couldnt do this now"))
		provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

		_, err := provider.Provision(context.TODO(), provisionContext)

		Expect(err).To(MatchError("error marking job started: couldnt do this now"))
	})

	It("return the error in last operation, if tofu apply fails", func() {
		fakeDeploymentManager.CreateAndSaveDeploymentReturns(deployment, nil)
		fakeInvokerBuilder.VersionedTerraformInvokerReturns(fakeDefaultInvoker)
		fakeDefaultInvoker.ApplyReturns(errors.New("some TF issue happened"))
		provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

		_, err := provider.Provision(context.TODO(), provisionContext)
		Expect(err).NotTo(HaveOccurred())

		By("checking TF apply has been called")
		Eventually(operationWasFinishedForDeployment(fakeDeploymentManager)).Should(Equal(deployment))
		Expect(operationWasFinishedWithError(fakeDeploymentManager)()).To(MatchError("some TF issue happened"))
	})
})
