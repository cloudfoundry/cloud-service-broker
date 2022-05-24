package tf_test

import (
	"testing"

	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/tffakes"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/workspace"
	"github.com/hashicorp/go-version"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTF(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TF Suite")
}

func applyCallCount(fakeDefaultInvoker *tffakes.FakeTerraformInvoker) func() int {
	return func() int {
		return fakeDefaultInvoker.ApplyCallCount()
	}
}

func importCallCount(fakeDefaultInvoker *tffakes.FakeTerraformInvoker) func() int {
	return func() int {
		return fakeDefaultInvoker.ImportCallCount()
	}
}

func showCallCount(fakeDefaultInvoker *tffakes.FakeTerraformInvoker) func() int {
	return func() int {
		return fakeDefaultInvoker.ShowCallCount()
	}
}

func planCallCount(fakeDefaultInvoker *tffakes.FakeTerraformInvoker) func() int {
	return func() int {
		return fakeDefaultInvoker.PlanCallCount()
  }
}

func destroyCallCount(fakeDefaultInvoker *tffakes.FakeTerraformInvoker) func() int {
	return func() int {
		return fakeDefaultInvoker.DestroyCallCount()
	}
}

func getWorkspace(invoker *tffakes.FakeTerraformInvoker, pos int) workspace.Workspace {
	_, workspace := invoker.ApplyArgsForCall(pos)
	return workspace
}

func operationWasFinishedWithError(fakeDeploymentManager *tffakes.FakeDeploymentManagerInterface) func() error {
	return func() error {
		_, err := lastOperationMarkedFinished(fakeDeploymentManager)
		return err
	}
}

func operationWasFinishedForDeployment(fakeDeploymentManager *tffakes.FakeDeploymentManagerInterface) func() storage.TerraformDeployment {
	return func() storage.TerraformDeployment {
		deployment, _ := lastOperationMarkedFinished(fakeDeploymentManager)
		return deployment
	}
}

func lastOperationMarkedFinished(fakeDeploymentManager *tffakes.FakeDeploymentManagerInterface) (storage.TerraformDeployment, error) {
	callCount := fakeDeploymentManager.MarkOperationFinishedCallCount()
	if callCount == 0 {
		return storage.TerraformDeployment{}, nil
	} else {
		return fakeDeploymentManager.MarkOperationFinishedArgsForCall(callCount - 1)
	}
}

func newVersion(v string) *version.Version {
	return version.Must(version.NewVersion(v))
}
