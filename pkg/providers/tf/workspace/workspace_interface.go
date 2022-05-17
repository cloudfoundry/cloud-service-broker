package workspace

import (
	"context"

	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/command"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/executor"

	"github.com/hashicorp/go-version"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . Workspace
type Workspace interface {
	Serialize() (string, error)
	HasState() bool

	StateVersion() (*version.Version, error)
	Outputs(instance string) (map[string]interface{}, error)
	ModuleDefinitions() []ModuleDefinition
	ModuleInstances() []ModuleInstance
	UpdateInstanceConfiguration(vars map[string]interface{}) error
	Execute(ctx context.Context, executor executor.TerraformExecutor, commands ...command.TerraformCommand) (executor.ExecutionOutput, error)
}
