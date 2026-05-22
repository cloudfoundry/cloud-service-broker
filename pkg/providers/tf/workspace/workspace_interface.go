package workspace

import (
	"context"

	"github.gwd.broadcom.net/TNZ/cloud-service-broker/v2/pkg/providers/tf/command"
	"github.gwd.broadcom.net/TNZ/cloud-service-broker/v2/pkg/providers/tf/executor"

	"github.com/hashicorp/go-version"
)

//go:generate go tool counterfeiter -generate
//counterfeiter:generate . Workspace
type Workspace interface {
	Serialize() (string, error)
	HasState() bool

	StateTFVersion() (*version.Version, error)
	Outputs(instance string) (map[string]any, error)
	ModuleDefinitions() []ModuleDefinition
	ModuleInstances() []ModuleInstance
	UpdateInstanceConfiguration(vars map[string]any) error
	Execute(ctx context.Context, executor executor.TerraformExecutor, commands ...command.TerraformCommand) (executor.ExecutionOutput, error)
}
