package tf_test

import (
	"context"
	"os/exec"

	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/tffakes"

	"github.com/cloudfoundry/cloud-service-broker/brokerapi/broker/brokerfakes"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/wrapper"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/wrapper/wrapperfakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = XDescribe("TfJobRunner", func() {

	XDescribe("Update", func() {
		It("does stuff", func() {
			fakeStore := &brokerfakes.FakeStorage{}
			fakeExecutorFactory := &wrapperfakes.FakeExecutorFactory{}
			tfBinContext := wrapper.TFBinariesContext{}
			workspaceFactory := &tffakes.FakeWorkspaceFactory{}
			fakeWorkspace := &tffakes.FakeWorkspace{}

			runner := tf.NewTfJobRunner(fakeStore, fakeExecutorFactory, tfBinContext, workspaceFactory)

			var id string
			var templateVars map[string]interface{}

			deployment := storage.TerraformDeployment{
				ID: id,
			}
			fakeStore.GetTerraformDeploymentReturns(deployment, nil)
			workspaceFactory.CreateWorkspaceReturns(fakeWorkspace, nil)
			fakeExecutorFactory.VersionedExecutorReturns(func(context.Context, *exec.Cmd) (wrapper.ExecutionOutput, error) {

			})

			runner.Update(context.TODO(), id, templateVars)

			Expect(fakeStore.GetTerraformDeploymentCallCount()).To(Equal(1))
			Expect(workspaceFactory.CreateWorkspaceCallCount()).To(Equal(1))

			Expect(fakeExecutorFactory.DefaultExecutorCallCount()).To(Equal(1))

			Expect(true).To(BeTrue())
		})
	})

	XIt("ïf GetTerraformDeployment errors")
})
