// Package invoker allows different Terraform versions to be invoked
package invoker

import (
	"context"

	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf/executor"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf/workspace"

	"github.com/hashicorp/go-version"
)

//go:generate go tool counterfeiter -generate
//counterfeiter:generate . TerraformInvokerBuilder
type TerraformInvokerBuilder interface {
	VersionedTerraformInvoker(version *version.Version) TerraformInvoker
}

//go:generate go tool counterfeiter -generate
//counterfeiter:generate . TerraformInvoker
type TerraformInvoker interface {
	Destroy(ctx context.Context, workspace workspace.Workspace) error
	Apply(ctx context.Context, workspace workspace.Workspace) error
	Show(ctx context.Context, workspace workspace.Workspace) (string, error)
	Plan(ctx context.Context, workspace workspace.Workspace) (executor.ExecutionOutput, error)
	Import(ctx context.Context, workspace workspace.Workspace, resources map[string]string) error
}
