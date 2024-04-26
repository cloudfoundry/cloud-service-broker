// Package invoker allows different Terraform versions to be invoked
package invoker

import (
	"context"

	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/providers/tf/executor"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/providers/tf/workspace"

	"github.com/hashicorp/go-version"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . TerraformInvokerBuilder
type TerraformInvokerBuilder interface {
	VersionedTerraformInvoker(version *version.Version) TerraformInvoker
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . TerraformInvoker
type TerraformInvoker interface {
	Destroy(ctx context.Context, workspace workspace.Workspace) error
	Apply(ctx context.Context, workspace workspace.Workspace) error
	Show(ctx context.Context, workspace workspace.Workspace) (string, error)
	Plan(ctx context.Context, workspace workspace.Workspace) (executor.ExecutionOutput, error)
	Import(ctx context.Context, workspace workspace.Workspace, resources map[string]string) error
}
