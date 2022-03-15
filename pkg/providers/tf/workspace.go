package tf

import (
	"context"

	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/wrapper"
	"github.com/hashicorp/go-version"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . Workspace
type Workspace interface {
	Serialize() (string, error)

	StateVersion() (*version.Version, error)
	Outputs(instance string) (map[string]interface{}, error)
	ModuleDefinitions() []wrapper.ModuleDefinition
	ModuleInstances() []wrapper.ModuleInstance
	UpdateInstanceConfiguration(vars map[string]interface{}) error
	Execute(ctx context.Context, executor wrapper.TerraformExecutor, commands ...wrapper.TerraformCommand) (wrapper.ExecutionOutput, error)
}
