package tf_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/v2/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf/workspace/workspacefakes"

	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf/workspace"

	"github.com/cloudfoundry/cloud-service-broker/v2/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf/executor"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf/tffakes"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/v2/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bind", func() {

	const expectedTfID = "58363e43-1f92-4174-a3fc-501dd5ee9e04"

	var (
		fakeDeploymentManager *tffakes.FakeDeploymentManagerInterface
		deployment            storage.TerraformDeployment
		fakeInvokerBuilder    *tffakes.FakeTerraformInvokerBuilder
		fakeDefaultInvoker    *tffakes.FakeTerraformInvoker
		fakeLogger            = utils.NewLogger("test")
		fakeServiceDefinition tf.TfServiceDefinitionV1
		bindContext           *varcontext.VarContext
		templateVars          = map[string]any{"tf_id": expectedTfID, "username": "some-user"}

		fakeTerraformWorkspace *workspacefakes.FakeWorkspace
	)

	BeforeEach(func() {
		fakeInvokerBuilder = &tffakes.FakeTerraformInvokerBuilder{}
		fakeTerraformWorkspace = &workspacefakes.FakeWorkspace{}

		var err error
		bindContext, err = varcontext.Builder().MergeMap(templateVars).Build()
		Expect(err).NotTo(HaveOccurred())

		template := `
				variable username { type = string }
				output username { value = var.username }
				`

		fakeServiceDefinition = tf.TfServiceDefinitionV1{
			BindSettings: tf.TfServiceDefinitionV1Action{
				Template:  template,
				Templates: map[string]string{"first": template},
			},
		}

		fakeDeploymentManager = &tffakes.FakeDeploymentManagerInterface{}
		fakeDefaultInvoker = &tffakes.FakeTerraformInvoker{}

		deployment = storage.TerraformDeployment{
			ID:        expectedTfID,
			Workspace: fakeTerraformWorkspace,
		}
	})

	When("binding successfully created", func() {
		It("returns the binding output", func() {
			fakeDeploymentManager.CreateAndSaveDeploymentReturns(deployment, nil)
			fakeDeploymentManager.GetTerraformDeploymentReturns(deployment, nil)
			fakeDeploymentManager.OperationStatusReturns(true, "operation succeeded", models.BindOperationType, nil)
			fakeInvokerBuilder.VersionedTerraformInvokerReturns(fakeDefaultInvoker)
			fakeDefaultInvoker.ApplyReturns(nil)
			fakeTerraformWorkspace.OutputsReturns(map[string]any{"username": "some-user"}, nil)

			provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

			actualBindDetails, err := provider.Bind(context.TODO(), bindContext)
			Expect(err).NotTo(HaveOccurred())
			Expect(actualBindDetails["username"]).To(Equal("some-user"))

			By("checking the new saved deployment")
			Expect(fakeDeploymentManager.CreateAndSaveDeploymentCallCount()).To(Equal(1))
			actualTfID, actualWorkspace := fakeDeploymentManager.CreateAndSaveDeploymentArgsForCall(0)
			Expect(actualTfID).To(Equal(expectedTfID))
			Expect(actualWorkspace.Modules[0].Name).To(Equal("brokertemplate"))
			Expect(actualWorkspace.Modules[0].Definition).To(Equal(fakeServiceDefinition.BindSettings.Template))
			Expect(actualWorkspace.Modules[0].Definitions).To(Equal(fakeServiceDefinition.BindSettings.Templates))
			Expect(actualWorkspace.Instances[0].Configuration).To(Equal(map[string]any{"username": "some-user"}))
			Expect(actualWorkspace.Transformer.ParameterMappings).To(Equal([]workspace.ParameterMapping{}))
			Expect(actualWorkspace.Transformer.ParametersToRemove).To(Equal([]string{}))
			Expect(actualWorkspace.Transformer.ParametersToAdd).To(Equal([]workspace.ParameterMapping{}))

			By("checking that provision is marked as started")
			Expect(fakeDeploymentManager.MarkOperationStartedCallCount()).To(Equal(1))
			actualDeployment, actualOperationType := fakeDeploymentManager.MarkOperationStartedArgsForCall(0)
			Expect(actualDeployment).To(Equal(&deployment))
			Expect(actualOperationType).To(Equal("bind"))

			By("checking TF apply has been called")
			Eventually(applyCallCount(fakeDefaultInvoker)).Should(Equal(1))
			Eventually(operationWasFinishedForDeployment(fakeDeploymentManager)).Should(Equal(deployment))
			Expect(operationWasFinishedWithError(fakeDeploymentManager)()).To(BeNil())
		})
	})

	It("fails, when tfID is not provided", func() {
		var err error
		bindContext, err = varcontext.Builder().Build()
		Expect(err).NotTo(HaveOccurred())

		provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

		_, err = provider.Bind(context.TODO(), bindContext)

		Expect(err.Error()).To(ContainSubstring(`missing value for key "tf_id"`))
	})

	It("fails, when workspace cannot be created", func() {
		fakeServiceDefinition = tf.TfServiceDefinitionV1{
			BindSettings: tf.TfServiceDefinitionV1Action{
				Template:  "invalid template",
				Templates: map[string]string{"first": "invalid template"},
			},
		}

		provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

		_, err := provider.Bind(context.TODO(), bindContext)

		Expect(err).To(MatchError(`error from provider bind: error creating workspace: :1,17-17: Invalid block definition; Either a quoted string block label or an opening brace ("{") is expected here.`))
	})

	It("fails, when it errors saving the deployment", func() {
		fakeDeploymentManager.CreateAndSaveDeploymentReturns(storage.TerraformDeployment{}, errors.New("cant save now"))
		provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

		_, err := provider.Bind(context.TODO(), bindContext)

		Expect(err).To(MatchError("error from provider bind: deployment create failed: cant save now"))
	})

	It("fails, when unable to mark the operation as started", func() {
		fakeDeploymentManager.CreateAndSaveDeploymentReturns(deployment, nil)
		fakeDeploymentManager.MarkOperationStartedReturns(errors.New("couldnt do this now"))
		provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

		_, err := provider.Bind(context.TODO(), bindContext)

		Expect(err).To(MatchError("error from provider bind: error marking job started: couldnt do this now"))
	})

	It("returns the error in last operation, if tofu apply fails", func() {
		fakeDeploymentManager.CreateAndSaveDeploymentReturns(deployment, nil)
		fakeDeploymentManager.OperationStatusReturns(true, "operation failed", models.BindOperationType, fmt.Errorf("tofu apply failed"))
		fakeDeploymentManager.MarkOperationFinishedReturns(errors.New("couldnt do this now"))
		fakeInvokerBuilder.VersionedTerraformInvokerReturns(fakeDefaultInvoker)
		fakeDefaultInvoker.ApplyReturns(errors.New("some TF issue happened"))
		provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

		_, err := provider.Bind(context.TODO(), bindContext)
		Expect(err).To(MatchError("error waiting for result: tofu apply failed"))

		By("checking TF apply has been called")
		Eventually(operationWasFinishedForDeployment(fakeDeploymentManager)).Should(Equal(deployment))
		Expect(operationWasFinishedWithError(fakeDeploymentManager)()).To(MatchError("some TF issue happened"))
	})

})
