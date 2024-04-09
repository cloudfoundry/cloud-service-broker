package tf_test

import (
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf"

	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/executor"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/tffakes"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/workspace"
	"github.com/cloudfoundry/cloud-service-broker/utils"
	"github.com/hashicorp/go-version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CheckUpgradeAvailable", func() {

	const deploymentID = "fcef3bb7-5ec4-4fae-859d-0632ec1e760a"
	var tfInstanceID = "tf:" + deploymentID + ":"

	var (
		deployment              storage.TerraformDeployment
		tfBinContext            executor.TFBinariesContext
		fakeInvokerBuilder      *tffakes.FakeTerraformInvokerBuilder
		fakeLogger              = utils.NewLogger("test")
		fakeServiceDefinition   = tf.TfServiceDefinitionV1{}
		fakeDeploymentManager   *tffakes.FakeDeploymentManagerInterface
		oldTerraformVersion     *version.Version
		defaultTerraformVersion *version.Version
	)

	BeforeEach(func() {
		fakeDeploymentManager = &tffakes.FakeDeploymentManagerInterface{}

		oldTerraformVersion = version.Must(version.NewVersion("0.1.0"))
		defaultTerraformVersion = version.Must(version.NewVersion("0.2.0"))

		tfBinContext = executor.TFBinariesContext{
			DefaultTfVersion: defaultTerraformVersion,
		}

	})

	When("default tofu version greater than one in state", func() {
		BeforeEach(func() {
			deployment = storage.TerraformDeployment{
				ID: deploymentID,
				Workspace: &workspace.TerraformWorkspace{
					State: []byte(fmt.Sprintf(`{"terraform_version": "%s" }`, oldTerraformVersion.String())),
				},
			}
			fakeDeploymentManager.GetTerraformDeploymentReturns(deployment, nil)
		})

		It("returns an error", func() {
			provider := tf.NewTerraformProvider(tfBinContext, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

			err := provider.CheckUpgradeAvailable(tfInstanceID)
			Expect(err).To(MatchError("operation attempted with newer version of OpenTofu than current state, upgrade the service before retrying operation"))
		})
	})

	When("default tofu version matches version in state", func() {
		BeforeEach(func() {
			deployment = storage.TerraformDeployment{
				ID: deploymentID,
				Workspace: &workspace.TerraformWorkspace{
					State: []byte(fmt.Sprintf(`{"terraform_version": "%s" }`, defaultTerraformVersion.String())),
				},
			}
			fakeDeploymentManager.GetTerraformDeploymentReturns(deployment, nil)
		})

		It("returns nil", func() {
			provider := tf.NewTerraformProvider(tfBinContext, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

			err := provider.CheckUpgradeAvailable(tfInstanceID)
			Expect(err).NotTo(HaveOccurred())
		})

	})

	When("unable to get the tofu version from a deployment", func() {
		BeforeEach(func() {
			deployment = storage.TerraformDeployment{
				ID: deploymentID,
				Workspace: &workspace.TerraformWorkspace{
					State: []byte(`{"broken_state": "sad" }`),
				},
			}
			fakeDeploymentManager.GetTerraformDeploymentReturns(deployment, nil)
		})

		It("returns error", func() {
			provider := tf.NewTerraformProvider(tfBinContext, fakeInvokerBuilder, fakeLogger, fakeServiceDefinition, fakeDeploymentManager)

			err := provider.CheckUpgradeAvailable(tfInstanceID)
			Expect(err).To(MatchError("Malformed version: "))
		})
	})

})
