package tf_test

import (
	"context"

	"github.com/cloudfoundry/cloud-service-broker/internal/brokerpak/manifest"
	"github.com/hashicorp/go-version"

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
			tfBinContext := wrapper.TFBinariesContext{
				DefaultTfVersion: version.Must(version.NewVersion("0.1.0")),
				TfUpgradePath: []manifest.TerraformUpgradePath{
					{Version: "0.0.1"},
					{Version: "0.0.2"},
					{Version: "0.1.0"},
				},
			}
			workspaceFactory := &tffakes.FakeWorkspaceFactory{}
			fakeWorkspace := &tffakes.FakeWorkspace{}
			fakeExecutorDefault := &wrapperfakes.FakeTerraformExecutor{}
			fakeExecutor1 := &wrapperfakes.FakeTerraformExecutor{}
			fakeExecutor2 := &wrapperfakes.FakeTerraformExecutor{}
			runner := tf.NewTfJobRunner(fakeStore, fakeExecutorFactory, tfBinContext, workspaceFactory)

			var id string
			deployment := storage.TerraformDeployment{
				ID: id,
			}
			fakeStore.GetTerraformDeploymentReturns(deployment, nil)
			workspaceFactory.CreateWorkspaceReturns(fakeWorkspace, nil)
			fakeExecutorFactory.DefaultExecutorReturnsOnCall(0, fakeExecutorDefault)
			fakeExecutorFactory.VersionedExecutorReturnsOnCall(0, fakeExecutor1)
			fakeExecutorFactory.VersionedExecutorReturnsOnCall(1, fakeExecutor2)

			fakeWorkspace.StateVersionReturns(version.Must(version.NewVersion("0.0.1")), nil)

			var templateVars map[string]interface{}
			runner.Update(context.TODO(), id, templateVars)

			Expect(fakeWorkspace.ApplyCallCount()).To(Equal(3))
			//Expect(fakeWorkspace.ApplyArgsForCall(0)).To(Equal())

			Expect(fakeStore.GetTerraformDeploymentCallCount()).To(Equal(1))
			Expect(workspaceFactory.CreateWorkspaceCallCount()).To(Equal(1))
			Expect(fakeExecutorFactory.VersionedExecutorCallCount()).To(Equal(2))
			Expect(fakeExecutorFactory.VersionedExecutorArgsForCall(0)).To(Equal(version.Must(version.NewVersion("0.0.1"))))
			Expect(fakeExecutorFactory.VersionedExecutorArgsForCall(1)).To(Equal(version.Must(version.NewVersion("0.0.2"))))
			Expect(fakeExecutorFactory.DefaultExecutorCallCount()).To(Equal(1))

			Expect(true).To(BeTrue())
		})
	})

	XIt("ïf GetTerraformDeployment errors")
})
