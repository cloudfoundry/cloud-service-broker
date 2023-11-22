package invoker

import (
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/executor"
	"github.com/hashicorp/go-version"
)

func NewTerraformInvokerFactory(executorBuilder executor.ExecutorBuilder, terraformPluginsDirectory string, pluginRenames map[string]string) TerraformInvokerBuilder {
	return TerraformInvokerFactory{executorBuilder: executorBuilder, terraformPluginsDirectory: terraformPluginsDirectory, pluginRenames: pluginRenames}
}

type TerraformInvokerFactory struct {
	executorBuilder           executor.ExecutorBuilder
	terraformPluginsDirectory string
	pluginRenames             map[string]string
}

func (factory TerraformInvokerFactory) VersionedTerraformInvoker(tfVersion *version.Version) TerraformInvoker {
	if tfVersion.LessThan(version.Must(version.NewVersion("0.13.0"))) {
		return NewTerraform012Invoker(factory.executorBuilder.VersionedExecutor(tfVersion), factory.terraformPluginsDirectory)
	}
	return NewTerraformDefaultInvoker(factory.executorBuilder.VersionedExecutor(tfVersion), factory.terraformPluginsDirectory, factory.pluginRenames)
}
