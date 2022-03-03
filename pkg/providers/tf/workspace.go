package tf

import (
	"context"

	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/wrapper"
	"github.com/hashicorp/go-version"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . Workspace
type Workspace interface {
	StateVersion() (*version.Version, error)
	Serialize() (string, error)
	Outputs(instance string) (map[string]interface{}, error)
	Validate(ctx context.Context, executor wrapper.TerraformExecutor) error
	Apply(ctx context.Context, executor wrapper.TerraformExecutor) error
	Plan(ctx context.Context, executor wrapper.TerraformExecutor) error
	Destroy(ctx context.Context, executor wrapper.TerraformExecutor) error
	Import(ctx context.Context, executor wrapper.TerraformExecutor, resources map[string]string) error
	Show(ctx context.Context, executor wrapper.TerraformExecutor) (string, error)
	ModuleDefinitions() []wrapper.ModuleDefinition
	ModuleInstances() []wrapper.ModuleInstance
	UpdateInstanceConfiguration(vars map[string]interface{}) error
}
