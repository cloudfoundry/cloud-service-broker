package tf_test

import (
	"context"
	"errors"

	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/executor"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/tffakes"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/workspace/workspacefakes"
	"github.com/cloudfoundry/cloud-service-broker/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Outputs", func() {
	Describe("GetTerraformOutputs", func() {
		var (
			fakeDeploymentManager *tffakes.FakeDeploymentManagerInterface
			fakeInvokerBuilder    *tffakes.FakeTerraformInvokerBuilder
			fakeLogger            = utils.NewLogger("test")
			fakeServiceDefinition tf.TfServiceDefinitionV1
			fakeWorkspace         *workspacefakes.FakeWorkspace
		)

		BeforeEach(func() {
			fakeInvokerBuilder = &tffakes.FakeTerraformInvokerBuilder{}
			fakeDeploymentManager = &tffakes.FakeDeploymentManagerInterface{}
			fakeWorkspace = &workspacefakes.FakeWorkspace{}
		})

		It("returns workspace outputs", func() {
			fakeDeploymentManager.GetTerraformDeploymentReturns(storage.TerraformDeployment{
				Workspace: fakeWorkspace,
			}, nil)
			fakeWorkspace.OutputsReturns(map[string]interface{}{"out": "foo"}, nil)

			provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

			output, err := provider.GetTerraformOutputs(context.TODO(), "instance-guid")

			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal(storage.JSONObject{"out": "foo"}))

			Expect(fakeDeploymentManager.GetTerraformDeploymentCallCount()).To(Equal(1))
			Expect(fakeDeploymentManager.GetTerraformDeploymentArgsForCall(0)).To(Equal("tf:instance-guid:"))

			Expect(fakeWorkspace.OutputsCallCount()).To(Equal(1))
		})

		It("fails, when it cant get the terraform deployment", func() {
			fakeDeploymentManager.GetTerraformDeploymentReturns(storage.TerraformDeployment{}, errors.New("cant get deployment now"))

			provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

			_, err := provider.GetTerraformOutputs(context.TODO(), "instance-guid")

			Expect(err).To(MatchError("error getting TF deployment: cant get deployment now"))
		})

		It("fails, when it cant get workspace output", func() {
			fakeDeploymentManager.GetTerraformDeploymentReturns(storage.TerraformDeployment{
				Workspace: fakeWorkspace,
			}, nil)
			fakeWorkspace.OutputsReturns(map[string]interface{}{}, errors.New("cant get outputs"))

			provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

			_, err := provider.GetTerraformOutputs(context.TODO(), "instance-guid")

			Expect(err).To(MatchError("cant get outputs"))
		})
	})
})
