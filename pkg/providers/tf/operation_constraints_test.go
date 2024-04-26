package tf_test

import (
	"errors"

	"github.com/cloudfoundry/cloud-service-broker/v3/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/providers/tf"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/providers/tf/executor"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/providers/tf/tffakes"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/providers/tf/workspace"
	"github.com/cloudfoundry/cloud-service-broker/v3/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CheckOperationConstraints", func() {

	const deploymentID = "fcef3bb7-5ec4-4fae-859d-0632ec1e760a"

	var (
		deployment            storage.TerraformDeployment
		fakeInvokerBuilder    *tffakes.FakeTerraformInvokerBuilder
		fakeLogger            = utils.NewLogger("test")
		fakeServiceDefinition = tf.TfServiceDefinitionV1{}
		fakeDeploymentManager *tffakes.FakeDeploymentManagerInterface
		provider              *tf.TerraformProvider
	)

	BeforeEach(func() {
		fakeDeploymentManager = &tffakes.FakeDeploymentManagerInterface{}
		provider = tf.NewTerraformProvider(
			executor.TFBinariesContext{},
			fakeInvokerBuilder,
			fakeLogger,
			fakeServiceDefinition,
			fakeDeploymentManager,
		)
	})

	When("a provision operation is in progress", func() {
		BeforeEach(func() {
			deployment = storage.TerraformDeployment{
				ID:                 deploymentID,
				Workspace:          &workspace.TerraformWorkspace{},
				LastOperationType:  "provision",
				LastOperationState: "in progress",
			}
			fakeDeploymentManager.GetTerraformDeploymentReturns(deployment, nil)
		})

		It("returns an error", func() {

			err := provider.CheckOperationConstraints(deploymentID, "deprovision")
			Expect(err).To(
				MatchError(
					"destroy operation not allowed while provision is in progress",
				),
			)
		})
	})

	When("call from an operation which is not a deprovision", func() {
		DescribeTable(
			"does not return an error",
			func(operationType string) {

				err := provider.CheckOperationConstraints(deploymentID, operationType)
				Expect(err).NotTo(HaveOccurred())
			},
			Entry("provision", "provision"),
			Entry("update", "update"),
			Entry("upgrade", "upgrade"),
			Entry("bind", "bind"),
			Entry("unbind", "unbind"),
		)
	})

	When("getting terraform deployment errors", func() {
		BeforeEach(func() {
			deployment = storage.TerraformDeployment{}
			fakeDeploymentManager.GetTerraformDeploymentReturns(deployment, errors.New("fake-error"))
		})

		It("returns an error", func() {

			err := provider.CheckOperationConstraints(deploymentID, "deprovision")
			Expect(err).To(MatchError("fake-error"))
		})
	})

	When("the last operation was a completed provision", func() {
		BeforeEach(func() {
			deployment = storage.TerraformDeployment{
				ID:                 deploymentID,
				Workspace:          &workspace.TerraformWorkspace{},
				LastOperationType:  "provision",
				LastOperationState: "succeeded",
			}
			fakeDeploymentManager.GetTerraformDeploymentReturns(deployment, nil)
		})

		It("does not return an error", func() {

			err := provider.CheckOperationConstraints(deploymentID, "deprovision")
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
