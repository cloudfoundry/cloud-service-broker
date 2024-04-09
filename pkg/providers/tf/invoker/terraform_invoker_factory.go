package invoker

import (
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf/executor"
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
	return NewTerraformDefaultInvoker(factory.executorBuilder.VersionedExecutor(tfVersion), factory.terraformPluginsDirectory, factory.pluginRenames)
}
