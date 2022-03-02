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

func NewExecutorFactoryImp(tfBinContext TFBinariesContext, envVars map[string]string) ExecutorFactory {
	return ExecutorFactoryImp{
		Dir:              tfBinContext.Dir,
		DefaultTfVersion: tfBinContext.DefaultTfVersion,
		Params:           tfBinContext.Params,
		EnvVars:          envVars,
	}
}

type ExecutorFactoryImp struct {
	Dir              string
	DefaultTfVersion *version.Version
	Params           map[string]string
	EnvVars          map[string]string
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . ExecutorFactory

type ExecutorFactory interface {
	DefaultExecutor() TerraformExecutor
	VersionedExecutor(tfVersion *version.Version) TerraformExecutor
}

func (executorFactory ExecutorFactoryImp) DefaultExecutor() TerraformExecutor {
	return executorFactory.VersionedExecutor(executorFactory.DefaultTfVersion)
}

func (executorFactory ExecutorFactoryImp) VersionedExecutor(tfVersion *version.Version) TerraformExecutor {
	return CustomEnvironmentExecutor(executorFactory.EnvVars,
		CustomEnvironmentExecutor(
			executorFactory.Params,
			CustomTerraformExecutor(
				filepath.Join(executorFactory.Dir, "versions", tfVersion.String(), "terraform"),
				executorFactory.Dir,
				tfVersion,
				DefaultExecutor,
			),
		),
	)
}
