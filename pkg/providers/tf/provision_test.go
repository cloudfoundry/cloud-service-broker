package tf_test

import (
	"context"
	"errors"

	"github.com/cloudfoundry/cloud-service-broker/v3/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/broker"
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

	Describe("provision new resource", func() {
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
			Expect(actualWorkspace.Transformer.ParameterMappings).To(Equal([]workspace.ParameterMapping{}))
			Expect(actualWorkspace.Transformer.ParametersToRemove).To(Equal([]string{}))
			Expect(actualWorkspace.Transformer.ParametersToAdd).To(Equal([]workspace.ParameterMapping{}))

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

	Describe("provision from imported resource (aka subsume)", func() {
		var (
			provisionContext *varcontext.VarContext
			templateVars     = map[string]any{"tf_id": expectedTfID, "subsume": true, "import_input": "some_import_input", "username": "some-user"}
		)

		BeforeEach(func() {
			var err error
			provisionContext, err = varcontext.Builder().MergeMap(templateVars).Build()
			Expect(err).NotTo(HaveOccurred())

			fakeServiceDefinition = tf.TfServiceDefinitionV1{
				ProvisionSettings: tf.TfServiceDefinitionV1Action{
					PlanInputs: []broker.BrokerVariable{
						{
							FieldName: "subsume",
						},
					},
					ImportVariables: []broker.ImportVariable{
						{
							Name:       "import_input",
							TfResource: "tf_import_input",
						},
					},
					ImportParameterMappings: []tf.ImportParameterMapping{
						{
							TfVariable:    "map_this_param",
							ParameterName: "map_to_this_param",
						},
					},
					ImportParametersToAdd: []tf.ImportParameterMapping{
						{
							TfVariable:    "add_this_tf_param",
							ParameterName: "add_as_this_param",
						},
					},
					ImportParametersToDelete: []string{"remove_this_param"},
					Template:                 template,
					Templates:                map[string]string{"first": template},
				},
			}
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
			Expect(actualWorkspace.Modules[0].Definition).To(BeEmpty())
			Expect(actualWorkspace.Modules[0].Definitions).To(Equal(fakeServiceDefinition.ProvisionSettings.Templates))
			Expect(actualWorkspace.Instances[0].Configuration).To(Equal(map[string]any{"username": "some-user"}))
			Expect(actualWorkspace.Transformer.ParameterMappings).To(Equal([]workspace.ParameterMapping{{TfVariable: "map_this_param", ParameterName: "map_to_this_param"}}))
			Expect(actualWorkspace.Transformer.ParametersToRemove).To(Equal([]string{"remove_this_param"}))
			Expect(actualWorkspace.Transformer.ParametersToAdd).To(Equal([]workspace.ParameterMapping{{TfVariable: "add_this_tf_param", ParameterName: "add_as_this_param"}}))

			By("checking that provision is marked as started")
			Expect(fakeDeploymentManager.MarkOperationStartedCallCount()).To(Equal(1))
			actualDeployment, actualOperationType := fakeDeploymentManager.MarkOperationStartedArgsForCall(0)
			Expect(actualDeployment).To(Equal(&deployment))
			Expect(actualOperationType).To(Equal("provision"))

			By("checking TF import has been called")
			Eventually(importCallCount(fakeDefaultInvoker)).Should(Equal(1))
			_, _, resources := fakeDefaultInvoker.ImportArgsForCall(0)
			Expect(resources).To(Equal(map[string]string{"tf_import_input": "some_import_input"}))

			Eventually(showCallCount(fakeDefaultInvoker)).Should(Equal(1))
			Eventually(planCallCount(fakeDefaultInvoker)).Should(Equal(1))
			Eventually(applyCallCount(fakeDefaultInvoker)).Should(Equal(1))

			Eventually(operationWasFinishedForDeployment(fakeDeploymentManager)).Should(Equal(deployment))
			Expect(operationWasFinishedWithError(fakeDeploymentManager)()).To(BeNil())
		})

		It("fails, when not all import params are provided", func() {
			pc, err := varcontext.Builder().MergeMap(map[string]any{"tf_id": expectedTfID, "subsume": true, "username": "some-user"}).Build()
			Expect(err).NotTo(HaveOccurred())

			provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

			_, err = provider.Provision(context.TODO(), pc)

			Expect(err).To(MatchError("must provide values for all import parameters: import_input"))
		})

		It("fails, when tfID is not provided", func() {
			pc, err := varcontext.Builder().Build()
			Expect(err).NotTo(HaveOccurred())

			provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

			_, err = provider.Provision(context.TODO(), pc)

			Expect(err.Error()).To(ContainSubstring(`missing value for key "tf_id"`))
		})

		It("fails, when workspace cannot be created", func() {
			sd := tf.TfServiceDefinitionV1{
				ProvisionSettings: tf.TfServiceDefinitionV1Action{
					PlanInputs: []broker.BrokerVariable{
						{
							FieldName: "subsume",
						},
					},
					ImportVariables: []broker.ImportVariable{
						{
							Name:       "import_input",
							TfResource: "tf_import_input",
						},
					},
					ImportParameterMappings: []tf.ImportParameterMapping{
						{
							TfVariable:    "map_this_param",
							ParameterName: "map_to_this_param",
						},
					},
					ImportParametersToAdd: []tf.ImportParameterMapping{
						{
							TfVariable:    "add_this_tf_param",
							ParameterName: "add_as_this_param",
						},
					},
					ImportParametersToDelete: []string{"remove_this_param"},
					Templates:                map[string]string{"first": "invalid template"},
				},
			}

			provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, sd, fakeDeploymentManager)

			_, err := provider.Provision(context.TODO(), provisionContext)

			Expect(err.Error()).To(ContainSubstring(`error creating workspace: :1,17-17: Invalid block definition; Either a quoted string block label or an opening brace ("{") is expected here.`))
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

		Describe("failures on TF operations", func() {
			It("return the error in last operation, if tofu import fails", func() {
				fakeDeploymentManager.CreateAndSaveDeploymentReturns(deployment, nil)
				fakeInvokerBuilder.VersionedTerraformInvokerReturns(fakeDefaultInvoker)
				fakeDefaultInvoker.ImportReturns(errors.New("some TF import issue happened"))
				provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

				_, err := provider.Provision(context.TODO(), provisionContext)
				Expect(err).NotTo(HaveOccurred())

				By("checking TF import has been called")
				Eventually(importCallCount(fakeDefaultInvoker)).Should(Equal(1))
				Eventually(showCallCount(fakeDefaultInvoker)).Should(Equal(0))
				Eventually(planCallCount(fakeDefaultInvoker)).Should(Equal(0))
				Eventually(applyCallCount(fakeDefaultInvoker)).Should(Equal(0))
				Eventually(operationWasFinishedForDeployment(fakeDeploymentManager)).Should(Equal(deployment))
				Expect(operationWasFinishedWithError(fakeDeploymentManager)()).To(MatchError("some TF import issue happened"))
			})

			It("return the error in last operation, if tofu show fails", func() {
				fakeDeploymentManager.CreateAndSaveDeploymentReturns(deployment, nil)
				fakeInvokerBuilder.VersionedTerraformInvokerReturns(fakeDefaultInvoker)
				fakeDefaultInvoker.ShowReturns("", errors.New("some TF show issue happened"))
				provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

				_, err := provider.Provision(context.TODO(), provisionContext)
				Expect(err).NotTo(HaveOccurred())

				By("checking TF show has been called")
				Eventually(showCallCount(fakeDefaultInvoker)).Should(Equal(1))
				Eventually(planCallCount(fakeDefaultInvoker)).Should(Equal(0))
				Eventually(applyCallCount(fakeDefaultInvoker)).Should(Equal(0))
				Eventually(operationWasFinishedForDeployment(fakeDeploymentManager)).Should(Equal(deployment))
				Expect(operationWasFinishedWithError(fakeDeploymentManager)()).To(MatchError("some TF show issue happened"))
			})

			It("return the error in last operation, if tofu plan fails", func() {
				fakeDeploymentManager.CreateAndSaveDeploymentReturns(deployment, nil)
				fakeInvokerBuilder.VersionedTerraformInvokerReturns(fakeDefaultInvoker)
				fakeDefaultInvoker.PlanReturns(executor.ExecutionOutput{}, errors.New("some TF plan issue happened"))
				provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

				_, err := provider.Provision(context.TODO(), provisionContext)
				Expect(err).NotTo(HaveOccurred())

				By("checking TF plan has been called")
				Eventually(planCallCount(fakeDefaultInvoker)).Should(Equal(1))
				Eventually(applyCallCount(fakeDefaultInvoker)).Should(Equal(0))
				Eventually(operationWasFinishedForDeployment(fakeDeploymentManager)).Should(Equal(deployment))
				Expect(operationWasFinishedWithError(fakeDeploymentManager)()).To(MatchError("some TF plan issue happened"))
			})

			It("return the error in last operation, if tofu apply fails", func() {
				fakeDeploymentManager.CreateAndSaveDeploymentReturns(deployment, nil)
				fakeInvokerBuilder.VersionedTerraformInvokerReturns(fakeDefaultInvoker)
				fakeDefaultInvoker.ApplyReturns(errors.New("some TF apply issue happened"))
				provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

				_, err := provider.Provision(context.TODO(), provisionContext)
				Expect(err).NotTo(HaveOccurred())

				By("checking TF apply has been called")
				Eventually(applyCallCount(fakeDefaultInvoker)).Should(Equal(1))
				Eventually(operationWasFinishedForDeployment(fakeDeploymentManager)).Should(Equal(deployment))
				Expect(operationWasFinishedWithError(fakeDeploymentManager)()).To(MatchError("some TF apply issue happened"))
			})
		})
	})
})
