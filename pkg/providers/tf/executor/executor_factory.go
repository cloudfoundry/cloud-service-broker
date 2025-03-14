package executor

import (
	"path/filepath"

	"github.com/hashicorp/go-version"
)

// TFBinariesContext is used to hold information about the location of
// terraform binaries on disk along with some metadata about how
// to run them.
type TFBinariesContext struct {
	Dir              string
	DefaultTfVersion *version.Version
	Params           map[string]string

	TfUpgradePath        []*version.Version
	ProviderReplacements map[string]string
}

func NewExecutorFactory(dir string, params map[string]string, envVars map[string]string) ExecutorBuilder {
	return ExecutorFactory{
		Dir:     dir,
		Params:  params,
		EnvVars: envVars,
	}
}

type ExecutorFactory struct {
	Dir              string
	DefaultTfVersion *version.Version
	Params           map[string]string
	EnvVars          map[string]string
}

//go:generate go tool counterfeiter -generate
//counterfeiter:generate . ExecutorBuilder

type ExecutorBuilder interface {
	VersionedExecutor(tfVersion *version.Version) TerraformExecutor
}

const binaryName = "tofu"

func (executorFactory ExecutorFactory) VersionedExecutor(tfVersion *version.Version) TerraformExecutor {
	return CustomEnvironmentExecutor(executorFactory.EnvVars,
		CustomEnvironmentExecutor(
			executorFactory.Params,
			CustomTerraformExecutor(
				filepath.Join(executorFactory.Dir, "versions", tfVersion.String(), binaryName),
				executorFactory.Dir,
				tfVersion,
				DefaultExecutor(),
			),
		),
	)
}
