package tf

import (
	"context"

	"github.com/hashicorp/go-version"

	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/wrapper"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . TerraformInvokerBuilder
type TerraformInvokerBuilder interface {
	VersionedTerraformInvoker(version *version.Version) TerraformInvoker
}

func NewTerraformInvokerFactory(executorBuilder wrapper.ExecutorBuilder) TerraformInvokerBuilder {
	return TerraformInvokerFactory{executorBuilder: executorBuilder}
}

type TerraformInvokerFactory struct {
	executorBuilder wrapper.ExecutorBuilder
}

func (factory TerraformInvokerFactory) VersionedTerraformInvoker(tfVersion *version.Version) TerraformInvoker {
	if tfVersion.LessThan(version.Must(version.NewVersion("0.13.0"))) {
		return Terraform012Invoker{executor: factory.executorBuilder.VersionedExecutor(tfVersion)}
	} else {
		return TerraformDefaultInvoker{executor: factory.executorBuilder.VersionedExecutor(tfVersion)}
	}
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . TerraformInvoker
type TerraformInvoker interface {
	Destroy(ctx context.Context, workspace Workspace) error
	Apply(ctx context.Context, workspace Workspace) error
	Show(ctx context.Context, workspace Workspace) (string, error)
	Plan(ctx context.Context, workspace Workspace) (wrapper.ExecutionOutput, error)
	Import(ctx context.Context, workspace Workspace, resources map[string]string) error
}
