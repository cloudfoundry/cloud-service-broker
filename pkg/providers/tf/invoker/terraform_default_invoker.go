package invoker

import (
	"context"

	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/executor"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/workspace"

	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/command"
)

func NewTerraformDefaultInvoker(executor executor.TerraformExecutor, pluginDirectory string, pluginRenames map[string]string) TerraformInvoker {
	return TerraformDefaultInvoker{executor: executor, pluginDirectory: pluginDirectory, providerReplaceGenerator: pluginRenames}
}

type TerraformDefaultInvoker struct {
	executor        executor.TerraformExecutor
	pluginDirectory string
	providerReplaceGenerator
}

func (cmd TerraformDefaultInvoker) Apply(ctx context.Context, workspace workspace.Workspace) error {
	var commands []command.TerraformCommand
	if workspace.HasState() {
		commands = cmd.ReplacementCommands()
	}
	commands = append(commands, command.NewInit(cmd.pluginDirectory), command.NewApply())

	_, err := workspace.Execute(ctx, cmd.executor, commands...)
	return err
}

func (cmd TerraformDefaultInvoker) Show(ctx context.Context, workspace workspace.Workspace) (string, error) {
	output, err := workspace.Execute(ctx, cmd.executor,
		append(
			cmd.ReplacementCommands(),
			command.NewInit(cmd.pluginDirectory),
			command.NewShow(),
		)...)
	return output.StdOut, err
}

func (cmd TerraformDefaultInvoker) Destroy(ctx context.Context, workspace workspace.Workspace) error {
	_, err := workspace.Execute(ctx, cmd.executor,
		append(
			cmd.ReplacementCommands(),
			command.NewInit(cmd.pluginDirectory),
			command.NewDestroy(),
		)...)
	return err
}

func (cmd TerraformDefaultInvoker) Plan(ctx context.Context, workspace workspace.Workspace) (executor.ExecutionOutput, error) {
	return workspace.Execute(ctx, cmd.executor,
		command.NewInit(cmd.pluginDirectory),
		command.NewPlan())
}

func (cmd TerraformDefaultInvoker) Import(ctx context.Context, workspace workspace.Workspace, resources map[string]string) error {
	commands := []command.TerraformCommand{
		command.NewInit(cmd.pluginDirectory),
	}
	for resource, id := range resources {
		commands = append(commands, command.NewImport(resource, id))
	}

	_, err := workspace.Execute(ctx, cmd.executor, commands...)
	return err
}

type providerReplaceGenerator map[string]string

func (replace providerReplaceGenerator) ReplacementCommands() []command.TerraformCommand {
	var commands []command.TerraformCommand

	for oldValue, newValue := range replace {
		commands = append(commands, command.NewRenameProvider(oldValue, newValue))
	}

	return commands
}
