package wrapper

import (
	"path/filepath"

	"github.com/cloudfoundry/cloud-service-broker/internal/brokerpak/manifest"
	"github.com/hashicorp/go-version"
)

// TFBinariesContext is used to hold information about the location of
// terraform binaries on disk along with some metadata about how
// to run them.
type TFBinariesContext struct {
	Dir              string
	DefaultTfVersion *version.Version
	Params           map[string]string

	TfUpgradePath []manifest.TerraformUpgradePath
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

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . ExecutorBuilder

type ExecutorBuilder interface {
	VersionedExecutor(tfVersion *version.Version) TerraformExecutor
}

func (executorFactory ExecutorFactory) VersionedExecutor(tfVersion *version.Version) TerraformExecutor {
	return CustomEnvironmentExecutor(executorFactory.EnvVars,
		CustomEnvironmentExecutor(
			executorFactory.Params,
			CustomTerraformExecutor(
				filepath.Join(executorFactory.Dir, "versions", tfVersion.String(), "terraform"),
				executorFactory.Dir,
				tfVersion,
				DefaultExecutor(),
			),
		),
	)
}
